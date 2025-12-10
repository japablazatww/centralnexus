package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

//go:embed registry.json
var registryData []byte

// --- Structs ---

type FunctionMetadata struct {
	Name           string
	Params         []Param
	Returns        []string
	RequestStruct  string
	ResponseStruct string
	Comment        string
}

type Param struct {
	Name      string
	Type      string
	JSONTag   string
	FieldName string // PascalCase for struct
}

type Catalog struct {
	Services []ServiceEntry `json:"services"`
}

type ServiceEntry struct {
	Namespace   string          `json:"namespace"`
	Method      string          `json:"method"`
	Description string          `json:"description"`
	Inputs      []ParamMetadata `json:"inputs"`
	Outputs     []ParamMetadata `json:"outputs"`
}

type ParamMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SearchResult struct {
	Namespace    string
	Method       string
	MatchedParam string
	ParamType    string // "Input" or "Output"
}

// --- Main ---

func main() {
	// Subcommands
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchParam := searchCmd.String("search-param", "", "Search service by parameter name")

	if len(os.Args) < 2 {
		fmt.Println("Usage: nexus-cli <command> [arguments]")
		fmt.Println("Commands: build, search")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		runBuild()
	case "search":
		searchCmd.Parse(os.Args[2:])
		runSearch(*searchParam)
	default:
		// Smart-Run: If argument looks like a flag for search, just search
		if strings.HasPrefix(os.Args[1], "-") {
			searchCmd.Parse(os.Args[1:])
			runSearch(*searchParam)
		} else {
			fmt.Println("Expected 'build' or 'search' subcommands.")
			os.Exit(1)
		}
	}
}

// --- Search Logic ---

func runSearch(query string) {
	// 1. Resolve Catalog Path
	catalogPath := resolveDefaultCatalog()

	// 2. Auto-Discovery Check
	data, err := os.ReadFile(catalogPath)
	if err != nil {
		fmt.Println("Catalog not found. Running auto-discovery...")
		runBuild()
		// Re-read
		data, err = os.ReadFile(catalogPath)
		if err != nil {
			fmt.Printf("Error: Could not build catalog: %v\n", err)
			os.Exit(1)
		}
	}

	// 3. Parse Catalog
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		fmt.Printf("Error parsing catalog: %v\n", err)
		os.Exit(1)
	}

	// 4. Search Execution
	if query != "" {
		results := searchByParam(catalog, query)
		if len(results) == 0 {
			fmt.Println("No services found with that parameter.")
		} else {
			fmt.Printf("Found %d services with parameter '%s':\n", len(results), query)
			for _, res := range results {
				fmt.Printf("- %s.%s\n", res.Namespace, res.Method)
				fmt.Printf("  Match: %s (%s)\n", res.MatchedParam, res.ParamType)
			}
		}
	} else {
		// List all by default
		fmt.Println("Available Services:")
		for _, s := range catalog.Services {
			fmt.Printf("- %s.%s\n  %s\n", s.Namespace, s.Method, s.Description)
			if len(s.Inputs) > 0 {
				fmt.Println("  Inputs:")
				for _, in := range s.Inputs {
					fmt.Printf("    - %s (%s)\n", in.Name, in.Type)
				}
			}
			if len(s.Outputs) > 0 {
				fmt.Println("  Outputs:")
				for _, out := range s.Outputs {
					fmt.Printf("    - %s (%s)\n", out.Name, out.Type)
				}
			}
		}
	}
}

func searchByParam(catalog Catalog, query string) []SearchResult {
	var results []SearchResult
	normalizedQuery := normalize(query)

	for _, svc := range catalog.Services {
		// Check Inputs
		for _, param := range svc.Inputs {
			if normalize(param.Name) == normalizedQuery {
				results = append(results, SearchResult{
					Namespace:    svc.Namespace,
					Method:       svc.Method,
					MatchedParam: param.Name,
					ParamType:    "Input",
				})
			}
		}
		// Check Outputs
		for _, param := range svc.Outputs {
			if normalize(param.Name) == normalizedQuery {
				results = append(results, SearchResult{
					Namespace:    svc.Namespace,
					Method:       svc.Method,
					MatchedParam: param.Name,
					ParamType:    "Output",
				})
			}
		}
	}
	return results
}

func normalize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func resolveDefaultCatalog() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".nexus", "catalog.json")
	}
	return "catalog.json"
}

// --- Build / Index Logic ---

func runBuild() {
	fmt.Println("Starting Nexus Library Discovery...")

	// Create Temp Dir for safe go get execution
	tempDir, err := os.MkdirTemp("", "nexus-build")
	if err != nil {
		log.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// init temp module
	execCmd(tempDir, "go", "mod", "init", "nexus-temp-builder")

	var libraries []string
	if err := json.Unmarshal(registryData, &libraries); err != nil {
		log.Fatalf("Error parsing internal registry: %v", err)
	}

	var allMetadata []FunctionMetadata
	var catalog Catalog

	for _, lib := range libraries {
		fmt.Printf("Checking library: %s ... ", lib)

		// 1. Ensure Installed (in temp module context)
		if err := ensureLibraryInstalled(tempDir, lib); err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		// 2. Resolve Path (using go list in temp context)
		path, err := resolvePackagePath(tempDir, lib)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			continue
		}
		fmt.Println("OK")

		// 3. Parse AST
		meta, entries := parseLibrary(path, lib)
		allMetadata = append(allMetadata, meta...)
		catalog.Services = append(catalog.Services, entries...)
	}

	updateGlobalCatalog(catalog)
}

func execCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	// cmd.Stdout = os.Stdout // Uncomment for verbose debug
	// cmd.Stderr = os.Stderr // Uncomment for verbose debug
	return cmd.Run()
}

func ensureLibraryInstalled(widthDir string, pkg string) error {
	// go get pkg@latest
	// stderr capture for better error reporting
	cmd := exec.Command("go", "get", pkg+"@latest")
	cmd.Dir = widthDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running go get: %s\nOutput: %s", err, string(output))
	}
	return nil
}

func resolvePackagePath(withDir string, pkg string) (string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
	cmd.Dir = withDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func parseLibrary(path string, namespace string) ([]FunctionMetadata, []ServiceEntry) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Warning: error parsing %s: %v", path, err)
		return nil, nil
	}

	var metadata []FunctionMetadata
	var entries []ServiceEntry

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if !fn.Name.IsExported() {
						continue
					}

					fname := fn.Name.Name

					// Inputs
					inputs := []ParamMetadata{}
					params := []Param{}
					for _, field := range fn.Type.Params.List {
						typeExpr := typeToString(field.Type)
						for _, name := range field.Names {
							pName := name.Name
							// Add to internal params (for server gen compat if needed later)
							params = append(params, Param{
								Name:      pName,
								Type:      typeExpr,
								JSONTag:   toSnakeCase(pName),
								FieldName: toPascalCase(pName),
							})
							// Add to Catalog Inputs
							inputs = append(inputs, ParamMetadata{
								Name: toSnakeCase(pName),
								Type: typeExpr,
							})
						}
					}

					// Outputs
					returns := []string{}
					outputs := []ParamMetadata{}
					if fn.Type.Results != nil {
						for i, field := range fn.Type.Results.List {
							typeExpr := typeToString(field.Type)
							// Return values often don't have names, or share types
							// We make best effort to label them if multiple
							// If named returns, we use them. Else "ret0", "ret1" or request user spec?
							// For PoC: just show types.
							name := ""
							if len(field.Names) > 0 {
								for _, n := range field.Names {
									name = n.Name
									outputs = append(outputs, ParamMetadata{Name: name, Type: typeExpr})
								}
							} else {
								// Unnamed return
								name = fmt.Sprintf("result_%d", i)
								outputs = append(outputs, ParamMetadata{Name: name, Type: typeExpr})
							}
							returns = append(returns, typeExpr)
						}
					}

					meta := FunctionMetadata{
						Name:          fname,
						Params:        params,
						Returns:       returns,
						RequestStruct: fname + "Request",
						Comment:       fn.Doc.Text(),
					}
					metadata = append(metadata, meta)

					entries = append(entries, ServiceEntry{
						Namespace:   strings.TrimPrefix(namespace, "github.com/japablazatww/"),
						Method:      fname,
						Description: strings.TrimSpace(fn.Doc.Text()),
						Inputs:      inputs,
						Outputs:     outputs,
					})
				}
			}
		}
	}
	return metadata, entries
}

func updateGlobalCatalog(cat Catalog) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	globalDir := filepath.Join(home, ".nexus")
	os.MkdirAll(globalDir, 0755)

	fGlobal, err := os.Create(filepath.Join(globalDir, "catalog.json"))
	if err != nil {
		log.Fatal(err)
	}
	defer fGlobal.Close()

	encGlobal := json.NewEncoder(fGlobal)
	encGlobal.SetIndent("", "  ")
	encGlobal.Encode(cat)
	fmt.Printf("Success. Catalog updated: %s\n", filepath.Join(globalDir, "catalog.json"))
}

// --- Helpers ---

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}

func toSnakeCase(str string) string {
	var matchFirstCap = unicode.IsUpper
	var result strings.Builder
	for i, r := range str {
		if matchFirstCap(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func toPascalCase(str string) string {
	if len(str) == 0 {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

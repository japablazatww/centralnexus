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
	// Global Debug Flag? No, flag parsing is per subcommand.
	// We'll add --debug to each.

	// 1. Build
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	buildDebug := buildCmd.Bool("debug", false, "Enable verbose output")

	// 2. Search
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchParam := searchCmd.String("search-param", "", "Search service by parameter name")
	searchDebug := searchCmd.Bool("debug", false, "Enable verbose output")

	// 3. Dump
	dumpCmd := flag.NewFlagSet("dump-catalog", flag.ExitOnError)
	dumpDebug := dumpCmd.Bool("debug", false, "Enable verbose output")

	if len(os.Args) < 2 {
		fmt.Println("Usage: nexus-cli <command> [arguments]")
		fmt.Println("Commands: build, search, dump-catalog")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		buildCmd.Parse(os.Args[2:])
		runBuild(*buildDebug)
	case "search":
		searchCmd.Parse(os.Args[2:])
		runSearch(*searchParam, *searchDebug)
	case "dump-catalog":
		dumpCmd.Parse(os.Args[2:])
		runDump(*dumpDebug)
	default:
		// Smart-Run search?
		if strings.HasPrefix(os.Args[1], "-") {
			searchCmd.Parse(os.Args[1:])
			runSearch(*searchParam, *searchDebug)
		} else {
			fmt.Println("Unknown command. Expected 'build', 'search', or 'dump-catalog'.")
			os.Exit(1)
		}
	}
}

func runDump(debug bool) {
	path := resolveDefaultCatalog()
	if debug {
		fmt.Printf("Reading catalog from: %s\n", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading catalog: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// --- Search Logic ---

func runSearch(query string, debug bool) {
	// 1. Resolve Catalog Path
	catalogPath := resolveDefaultCatalog()
	if debug {
		fmt.Printf("DEBUG: Using catalog path: %s\n", catalogPath)
	}

	// 2. Auto-Discovery Check
	data, err := os.ReadFile(catalogPath)
	if err != nil {
		fmt.Println("Catalog not found. Running auto-discovery...")
		runBuild(debug) // Propagate debug
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
		// If data exists but is bad invalid json, maybe print it in debug
		if debug {
			fmt.Printf("DEBUG: Invalid JSON content:\n%s\n", string(data))
		}
		os.Exit(1)
	}

	if debug {
		fmt.Printf("DEBUG: Catalog loaded. %d services found.\n", len(catalog.Services))
	}

	// 4. Search Execution
	if query != "" {
		if debug {
			fmt.Printf("DEBUG: Searching for param '%s'...\n", query)
		}
		results := searchByParam(catalog, query)
		// ... (rest is same logic, just keeping signatures consistent)
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

func runBuild(debug bool) {
	fmt.Println("Starting Nexus Library Discovery...")

	// Create Temp Dir for safe go get execution
	tempDir, err := os.MkdirTemp("", "nexus-build")
	if err != nil {
		log.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	if debug {
		fmt.Printf("DEBUG: Temp build dir: %s\n", tempDir)
	}

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
		if err := ensureLibraryInstalled(tempDir, lib, debug); err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		// 2. Resolve Path (using go list in temp context)
		path, err := resolvePackagePath(tempDir, lib, debug)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			continue
		}
		if debug {
			fmt.Printf("DEBUG: Resolved path for %s: %s\n", lib, path)
		} else {
			fmt.Println("OK")
		}

		// 3. Parse AST
		meta, entries := parseLibrary(path, lib, debug)
		if debug {
			fmt.Printf("DEBUG: Parsed %d functions from %s\n", len(entries), lib)
		}
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

func ensureLibraryInstalled(widthDir string, pkg string, debug bool) error {
	// go get pkg@latest
	// stderr capture for better error reporting
	cmd := exec.Command("go", "get", pkg+"@latest")
	cmd.Dir = widthDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Always return output details
		return fmt.Errorf("error running go get: %s\nOutput: %s", err, string(output))
	}
	if debug {
		fmt.Printf("\nDEBUG: go get output:\n%s\n", string(output))
	}
	return nil
}

func resolvePackagePath(withDir string, pkg string, debug bool) (string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
	cmd.Dir = withDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(string(output))
	if debug {
		fmt.Printf("DEBUG: Raw path bytes: %x\n", path)
	}
	return path, nil
}

func parseLibrary(path string, namespace string, debug bool) ([]FunctionMetadata, []ServiceEntry) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Warning: error parsing %s: %v", path, err)
		return nil, nil
	}
	if debug {
		fmt.Printf("DEBUG: ParseDir found %d packages in %s\n", len(pkgs), path)
	}

	var metadata []FunctionMetadata
	var entries []ServiceEntry

	for _, pkg := range pkgs {
		if debug {
			fmt.Printf("DEBUG: Visiting package %s\n", pkg.Name)
		}
		for _, file := range pkg.Files {
			if debug {
				fmt.Printf("DEBUG: Visiting file in %s\n", pkg.Name)
			}
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if !fn.Name.IsExported() {
						continue
					}
					if debug {
						fmt.Printf("DEBUG: Found exported func %s\n", fn.Name.Name)
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

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
	Parameters  []ParamMetadata `json:"parameters"`
}

type ParamMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SearchResult struct {
	Namespace    string
	Method       string
	MatchedParam string
}

// --- Main ---

func main() {
	// Subcommands
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchParam := searchCmd.String("search-param", "", "Search service by parameter name")

	if len(os.Args) < 2 {
		// Default behavior: Auto-build and then wait for commands or just exit?
		// For PoC ease of use: "If no args, just list services from default catalog"
		fmt.Println("Expected 'build' or 'search' subcommands.")
		fmt.Println("Example: nexus-cli search --search-param user_id")
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
	// 1. Ensure Catalog Exists (Auto-indexing check could go here, for now assumes build run once)
	catalogPath := resolveDefaultCatalog()

	// If catalog doesn't exist, try to auto-build first?
	// Optimistic: Try to read, if fail, warn.
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

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		fmt.Printf("Error parsing catalog: %v\n", err)
		os.Exit(1)
	}

	if query != "" {
		results := searchByParam(catalog, query)
		if len(results) == 0 {
			fmt.Println("No services found with that parameter.")
		} else {
			fmt.Printf("Found %d services with parameter '%s':\n", len(results), query)
			for _, res := range results {
				fmt.Printf("- %s.%s (Found matching param: %s)\n", res.Namespace, res.Method, res.MatchedParam)
			}
		}
	} else {
		// List all by default
		fmt.Println("Available Services:")
		for _, s := range catalog.Services {
			fmt.Printf("- %s.%s: %s\n", s.Namespace, s.Method, s.Description)
		}
	}
}

func searchByParam(catalog Catalog, query string) []SearchResult {
	var results []SearchResult
	normalizedQuery := normalize(query)

	for _, svc := range catalog.Services {
		for _, param := range svc.Parameters {
			if normalize(param.Name) == normalizedQuery {
				results = append(results, SearchResult{
					Namespace:    svc.Namespace,
					Method:       svc.Method,
					MatchedParam: param.Name,
				})
				break
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

	var libraries []string
	if err := json.Unmarshal(registryData, &libraries); err != nil {
		log.Fatalf("Error parsing internal registry: %v", err)
	}

	var allMetadata []FunctionMetadata
	var catalog Catalog

	for _, lib := range libraries {
		fmt.Printf("Checking library: %s ... ", lib)

		// 1. Ensure Installed
		if err := ensureLibraryInstalled(lib); err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		// 2. Resolve Path
		path, err := resolvePackagePath(lib)
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

	// 4. Generate Output (Assuming running in project root or relative)
	// For global CLI, we mainly care about the Catalog.
	// However, if run inside a project, we might want to generate code.
	// For this PoC, we update the Global Catalog mostly.
	// NOTE: Code generation (SDK/Server) usually happens in the repo that consumes it.
	// If running `nexus-cli build` is intended to scaffold the current repo, we need an output flag.
	// For "search" functionality, we only need the catalog.

	updateGlobalCatalog(catalog)
}

func ensureLibraryInstalled(pkg string) error {
	cmd := exec.Command("go", "get", pkg+"@latest")
	return cmd.Run()
}

func resolvePackagePath(pkg string) (string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
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
					params := []Param{}

					// Parse Params
					for _, field := range fn.Type.Params.List {
						typeExpr := typeToString(field.Type)
						for _, name := range field.Names {
							pName := name.Name
							params = append(params, Param{
								Name:      pName,
								Type:      typeExpr,
								JSONTag:   toSnakeCase(pName),
								FieldName: toPascalCase(pName),
							})
						}
					}

					// Return types
					returns := []string{}
					if fn.Type.Results != nil {
						for _, field := range fn.Type.Results.List {
							returns = append(returns, typeToString(field.Type))
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

					// Catalog Entry
					catParams := []ParamMetadata{}
					for _, p := range params {
						catParams = append(catParams, ParamMetadata{Name: p.JSONTag, Type: p.Type})
					}
					entries = append(entries, ServiceEntry{
						Namespace:   strings.TrimPrefix(namespace, "github.com/japablazatww/"), // Simplify namespace name
						Method:      fname,
						Description: strings.TrimSpace(fn.Doc.Text()),
						Parameters:  catParams,
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

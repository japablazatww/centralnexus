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

type LibConfig struct {
	HasNestedDomains bool     `json:"hasNestedDomains"`
	Domains          []string `json:"domains"`
	IsDomain         bool     `json:"isDomain"`
}

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
	buildDebug := buildCmd.Bool("debug", false, "Enable verbose output")

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchParam := searchCmd.String("search-param", "", "Search service by parameter name")
	searchDebug := searchCmd.Bool("debug", false, "Enable verbose output")

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

// --- Build / Crawler Logic ---

func runBuild(debug bool) {
	fmt.Println("Starting Nexus Library Discovery (DDD Mode)...")

	// Create Temp Dir
	tempDir, err := os.MkdirTemp("", "nexus-build")
	if err != nil {
		log.Fatalf("Error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if debug {
		fmt.Printf("DEBUG: Temp build dir: %s\n", tempDir)
	}

	execCmd(tempDir, "go", "mod", "init", "nexus-temp-builder")

	var libraries []string
	if err := json.Unmarshal(registryData, &libraries); err != nil {
		log.Fatalf("Error parsing internal registry: %v", err)
	}

	var catalog Catalog

	for _, lib := range libraries {
		fmt.Printf("Checking library: %s (@develop) ... ", lib)

		// 1. Ensure Installed (FORCE @develop)
		// NOTE: In production this should come from registry.json metadata
		if err := ensureLibraryInstalled(tempDir, lib, "develop", debug); err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		// 2. Resolve Root Path
		rootPath, err := resolvePackagePath(tempDir, lib, debug)
		if err != nil {
			fmt.Printf("Error resolving path: %v\n", err)
			continue
		}
		if debug {
			fmt.Printf("DEBUG: Root path for %s: %s\n", lib, rootPath)
		} else {
			fmt.Println("OK")
		}

		// 3. Crawl Recursively
		// Simplify namespace: github.com/japablazatww/libreria-a -> libreria-a
		baseNamespace := filepath.Base(lib)
		crawlLibrary(rootPath, baseNamespace, &catalog, debug)
	}

	updateGlobalCatalog(catalog)
}

func crawlLibrary(currentPath string, currentNamespace string, catalog *Catalog, debug bool) {
	if debug {
		fmt.Printf("DEBUG: Crawling %s (NS: %s)\n", currentPath, currentNamespace)
	}

	// 1. Read lib_config.json
	configFile := filepath.Join(currentPath, "lib_config.json")
	configData, err := os.ReadFile(configFile)
	if err != nil {
		if debug {
			fmt.Printf("DEBUG: No lib_config.json in %s\n", currentPath)
		}
		return
	}

	var config LibConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		if debug {
			fmt.Printf("DEBUG: Invalid lib_config.json in %s: %v\n", currentPath, err)
		}
		return
	}

	// 2. If it is a domain with functions, parse them
	if config.IsDomain {
		if debug {
			fmt.Printf("DEBUG: Found Domain at %s. Parsing functions...\n", currentNamespace)
		}
		_, entries := parseLibrary(currentPath, currentNamespace, debug)
		catalog.Services = append(catalog.Services, entries...)
	}

	// 3. If it has nested domains, recurse
	if config.HasNestedDomains {
		for _, domain := range config.Domains {
			subPath := filepath.Join(currentPath, domain)
			// Construct nested namespace: libreria-a.transfers.national
			subNamespace := fmt.Sprintf("%s.%s", currentNamespace, domain)
			crawlLibrary(subPath, subNamespace, catalog, debug)
		}
	}
}

func execCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}

func ensureLibraryInstalled(widthDir string, pkg string, version string, debug bool) error {
	// usage: go get pkg@version
	target := fmt.Sprintf("%s@%s", pkg, version)
	cmd := exec.Command("go", "get", target)
	cmd.Dir = widthDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running go get: %s\nOutput: %s", err, string(output))
	}
	if debug {
		fmt.Printf("\nDEBUG: go get output:\n%s\n", string(output))
	}

	return nil
}

func resolvePackagePath(withDir string, pkg string, debug bool) (string, error) {
	// Use -m to resolve the Module Root, as the root might not be a package anymore (no .go files)
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", pkg)
	cmd.Dir = withDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if debug {
			fmt.Printf("DEBUG: go list error output:\n%s\n", string(output))
		}
		return "", fmt.Errorf("go list failed: %v", err)
	}
	path := strings.TrimSpace(string(output))
	if debug {
		fmt.Printf("DEBUG: Raw path bytes: %x\n", path)
	}
	return path, nil
}

func parseLibrary(path string, namespace string, debug bool) ([]FunctionMetadata, []ServiceEntry) {
	fset := token.NewFileSet()
	// Parse only .go files in this directory
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
					// Check convention: Files containing functions usually named 'functions.go'
					// But we parse all for now.

					fname := fn.Name.Name

					// Inputs
					inputs := []ParamMetadata{}
					params := []Param{}
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
							name := ""
							if len(field.Names) > 0 {
								for _, n := range field.Names {
									name = n.Name
									outputs = append(outputs, ParamMetadata{Name: name, Type: typeExpr})
								}
							} else {
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
						Namespace:   namespace, // Namespace is passed from crawler now
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
	var result strings.Builder
	runes := []rune(str)
	length := len(runes)

	for i := 0; i < length; i++ {
		r := runes[i]
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				result.WriteRune('_')
			} else if i+1 < length && unicode.IsLower(runes[i+1]) {
				result.WriteRune('_')
			}
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

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
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

type StructMetadata struct {
	Name      string
	JsonName  string // Snake case of struct name for potential usage
	Fields    []StructField
	Namespace string
}

type StructField struct {
	Name    string
	Type    string
	JSONTag string
}

type Catalog struct {
	Services []ServiceEntry   `json:"services"`
	Structs  []StructMetadata `json:"structs"`
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
	buildOutput := buildCmd.String("output", "", "Path to the 'nexus/generated' directory")

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
		runBuild(*buildDebug, *buildOutput)
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
		runBuild(debug, "") // Propagate debug, no output override
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

// --- Path Resolution Logic (FIXED) ---

func resolveOutputDir(flagPath string) (string, error) {
	// 1. Explicit Flag
	if flagPath != "" {
		if _, err := os.Stat(flagPath); os.IsNotExist(err) {
			if err := os.MkdirAll(flagPath, 0755); err != nil {
				return "", fmt.Errorf("could not create output dir %s: %v", flagPath, err)
			}
		}
		path, _ := filepath.Abs(flagPath)
		return path, nil
	}

	// 2. Default: Create 'generated' in Current Working Directory
	cwd, _ := os.Getwd()
	target := filepath.Join(cwd, "generated")

	if _, err := os.Stat(target); os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			return "", fmt.Errorf("could not create output dir %s: %v", target, err)
		}
	}

	return target, nil
}

// --- Build / Crawler Logic ---

func runBuild(debug bool, outputFlag string) {
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
	var allMetadata []FunctionMetadata

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
		crawlLibrary(rootPath, baseNamespace, &catalog, &allMetadata, debug)
	}

	updateGlobalCatalog(catalog)

	// 4. Generate Code
	outputDir, err := resolveOutputDir(outputFlag)
	if err != nil {
		fmt.Printf("Error resolving output directory: %v\n", err)
		fmt.Println("Tip: Use --output <path> to specify the 'nexus/generated' folder.")
		return
	}
	fmt.Printf("Writing generated code to: %s\n", outputDir)

	// Dump Local Catalog
	writeLocalCatalog(catalog, outputDir)

	if err := generateServer(catalog, allMetadata, outputDir); err != nil {
		fmt.Printf("Error generating server: %v\n", err)
	} else {
		fmt.Println("Server code generated.")
	}

	if err := generateSDK(catalog, outputDir); err != nil {
		fmt.Printf("Error generating SDK: %v\n", err)
	} else {
		fmt.Println("SDK code generated.")
	}

	if err := generateTypes(catalog, outputDir); err != nil {
		fmt.Printf("Error generating Types: %v\n", err)
	} else {
		fmt.Println("Types code generated.")
	}
}

// --- Code Generation ---

func generateServer(catalog Catalog, metadata []FunctionMetadata, outputDir string) error {
	imports := make(map[string]string) // path -> alias

	type HandlerData struct {
		Route     string
		FuncAlias string
		FuncName  string
		Inputs    []ParamMetadata
		Outputs   []ParamMetadata // For signature
	}

	handlers := []HandlerData{}

	for _, svc := range catalog.Services {
		validPath := strings.ReplaceAll(svc.Namespace, ".", "/")
		importPath := "github.com/japablazatww/" + validPath

		alias := strings.ReplaceAll(svc.Namespace, ".", "_")
		alias = strings.ReplaceAll(alias, "-", "_")

		imports[importPath] = alias

		handlers = append(handlers, HandlerData{
			Route:     svc.Namespace + "." + svc.Method,
			FuncAlias: alias,
			FuncName:  svc.Method,
			Inputs:    svc.Inputs,
			Outputs:   svc.Outputs,
		})
	}

	// Template
	tmpl := `package generated

import (
	"encoding/json"
	"net/http"
	"strings"
    
	{{range $path, $alias := .Imports}}
	{{$alias}} "{{$path}}"
	{{end}}
)

func RegisterHandlers(mux *http.ServeMux) {
	{{range .Handlers}}
	mux.HandleFunc("/{{.Route}}", handle{{.FuncAlias}}_{{.FuncName}})
	{{end}}
}

{{range .Handlers}}
func handle{{.FuncAlias}}_{{.FuncName}}(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	{{if .Outputs}}resp, err := {{else}}{{end}}wrapper{{.FuncAlias}}_{{.FuncName}}(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	{{if .Outputs}}
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	{{else}}
	w.WriteHeader(http.StatusOK)
	{{end}}
}

func wrapper{{.FuncAlias}}_{{.FuncName}}(params map[string]interface{}) ({{if .Outputs}}interface{}, error{{else}}{{end}}) {
    // Inputs: {{range .Inputs}}{{.Name}}({{.Type}}), {{end}}
    
    {{$alias := .FuncAlias}}
    {{range .Inputs}}
    
    // Determine Type string (Primitive vs Complex)
    {{if or (eq .Type "string") (eq .Type "float64") (eq .Type "int") (eq .Type "bool")}}
    var val_{{.Name}} {{.Type}}
    {{else}}
    var val_{{.Name}} {{$alias}}.{{.Type}}
    {{end}}

    // Fuzzy Match Logic
    found_{{.Name}} := false
    target_{{.Name}} := strings.ToLower(strings.ReplaceAll("{{.Name}}", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_{{.Name}} {
            {{if eq .Type "string"}}
            val_{{.Name}}, _ = v.(string)
            {{else if eq .Type "float64"}}
            val_{{.Name}}, _ = v.(float64)
            {{else if eq .Type "int"}}
             // JSON numbers are float64
             if fVal, ok := v.(float64); ok {
                 val_{{.Name}} = int(fVal)
             }
            {{else if eq .Type "bool"}}
            val_{{.Name}}, _ = v.(bool)
            {{else}}
            // Complex Type: Convert map -> json -> struct
            jsonBody, _ := json.Marshal(v)
            json.Unmarshal(jsonBody, &val_{{.Name}})
            {{end}}
            found_{{.Name}} = true
            break
        }
    }
    
    if !found_{{.Name}} {
       // Optional: Log or Error if required param missing?
    }
    {{end}}

    // Call
    {{if .Outputs}}ret0, ret1 := {{end}}{{.FuncAlias}}.{{.FuncName}}({{range .Inputs}}val_{{.Name}}, {{end}})
    
    {{if .Outputs}}
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    {{else}}
    return nil, nil // void
    {{end}}
}
{{end}}
`

	f, err := os.Create(filepath.Join(outputDir, "server_gen.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	return executeTemplate(f, tmpl, map[string]interface{}{
		"Imports":  imports,
		"Handlers": handlers,
	})
}

func generateSDK(catalog Catalog, outputDir string) error {
	// We need to build a hierarchy.
	// Root -> LibreriaA -> System
	//                   -> Transfers -> National
	//                                -> International

	// Tree structure
	type Node struct {
		Name     string // e.g. "System"
		Children map[string]*Node
		Methods  []ServiceEntry
	}

	root := &Node{Name: "Client", Children: make(map[string]*Node)}

	for _, svc := range catalog.Services {
		// Split namespace: libreria-a.transfers.national
		parts := strings.Split(svc.Namespace, ".")

		current := root
		for _, p := range parts {
			// Normalize PascalCase for Struct fields
			p = toPascalCase(strings.ReplaceAll(p, "-", "")) // libreria-a -> LibreriaA

			if _, exists := current.Children[p]; !exists {
				current.Children[p] = &Node{Name: p, Children: make(map[string]*Node)}
			}
			current = current.Children[p]
		}
		current.Methods = append(current.Methods, svc)
	}

	// Flatten tree to generate structs
	type StructDef struct {
		Name    string
		Fields  []string // "System *LibreriaASystemClient"
		Methods []ServiceEntry
	}

	var structs []StructDef

	// BFS or DFS to traverse and build structs
	var traverse func(n *Node, prefix string) string // returns TypeName
	traverse = func(n *Node, prefix string) string {
		var typeName string
		// Special case for Root
		if n == root {
			typeName = "Client"
		} else {
			typeName = prefix + n.Name + "Client"
		}

		myStruct := StructDef{Name: typeName}

		// Compute next prefix for children
		var nextPrefix string
		if n == root {
			nextPrefix = ""
		} else {
			nextPrefix = prefix + n.Name
		}

		for childName, childNode := range n.Children {
			childType := traverse(childNode, nextPrefix)
			myStruct.Fields = append(myStruct.Fields, fmt.Sprintf("%s *%s", childName, childType))
		}

		myStruct.Methods = n.Methods
		structs = append(structs, myStruct)
		return typeName
	}

	traverse(root, "")

	// --- Dynamic Init Code Generation ---
	var initLines []string
	var genInit func(n *Node, accessPath string, typePrefix string)
	genInit = func(n *Node, accessPath string, typePrefix string) {
		// Sorted keys to ensure deterministic output
		// (optional but good for consistency)

		for childName, childNode := range n.Children {
			childAccess := accessPath + "." + childName
			childType := typePrefix + childName + "Client"

			// Line: c.Libreriaa = &LibreriaaClient{transport: t}
			// Line: c.Libreriaa = &LibreriaaClient{transport: t}

			// Needs "c." prefix? No, accessPath is "c" initially.
			// But childAccess acts as next accessPath.
			// wait, if I am at "c", child is "Libreriaa".
			// generated line: c.Libreriaa = &LibreriaaClient...

			// Logic check:
			// Root (n=Client), accessPath="c", typePrefix=""
			// Child "Libreriaa". childAccess="c.Libreriaa". childType="LibreriaaClient".
			// Line: c.Libreriaa = &LibreriaaClient{transport: t} -> YES.

			// Next recursion: accessPath="c.Libreriaa", typePrefix="Libreriaa"
			// Child "System". childAccess="c.Libreriaa.System". childType="LibreriaaSystemClient".
			// Line: c.Libreriaa.System = &LibreriaaSystemClient{transport: t} -> YES.

			// Use the calculated full access path for the assignment
			initLines = append(initLines, fmt.Sprintf("\t%s = &%s{transport: t}", childAccess, childType))

			genInit(childNode, childAccess, typePrefix+childName)
		}
	}

	genInit(root, "c", "")

	initCode := strings.Join(initLines, "\n")

	f, err := os.Create(filepath.Join(outputDir, "sdk_gen.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	return executeSDKTemplate(f, structs, initCode)
}

func generateTypes(catalog Catalog, outputDir string) error {
	tmpl := `package generated

// --- Shared Types ---

type GenericRequest struct {
	Params map[string]interface{} ` + "`" + `json:"params"` + "`" + `
}

{{range .Structs}}
type {{.Name}} struct {
{{range .Fields}}
    {{.Name}} {{.Type}} ` + "`" + `json:"{{.JSONTag}}"` + "`" + `
{{end}}
}
{{end}}
`
	f, err := os.Create(filepath.Join(outputDir, "types_gen.go"))
	if err != nil {
		return err
	}
	defer f.Close()
	return executeTemplate(f, tmpl, catalog)
}

func executeTemplate(w io.Writer, tmplStr string, data interface{}) error {
	t, err := template.New("gen").Parse(tmplStr)
	if err != nil {
		return err
	}
	return t.Execute(w, data)
}

func executeSDKTemplate(w io.Writer, structs interface{}, initCode string) error {
	const tmpl = `package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)



type Transport interface {
	Call(method string, req GenericRequest) (interface{}, error)
}

type httpTransport struct {
	BaseURL string
	Client  *http.Client
}

func (t *httpTransport) Call(method string, req GenericRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := t.Client.Post(t.BaseURL + "/" + method, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %s", resp.Status)
	}
	
	var result interface{}
	// Decode logic... for now just simple
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Structs ---

{{range $struct := .Structs}}
type {{$struct.Name}} struct {
	transport Transport
	{{range .Fields}}
	{{.}}
	{{end}}
}

{{range .Methods}}
func (c *{{$struct.Name}}) {{.Method}}(req GenericRequest) (interface{}, error) {
	return c.transport.Call("{{.Namespace}}.{{.Method}}", req)
}
{{end}}
{{end}}

func NewClient(baseURL string) *Client {
	t := &httpTransport{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
	c := &Client{transport: t}
	
	// Dynamic Init
{{.InitCode}}

	return c
}
`
	t, err := template.New("sdk").Parse(tmpl)
	if err != nil {
		return err
	}

	data := struct {
		Structs  interface{}
		InitCode string
	}{
		Structs:  structs,
		InitCode: initCode,
	}

	return t.Execute(w, data)
}

func crawlLibrary(currentPath string, currentNamespace string, catalog *Catalog, allMetadata *[]FunctionMetadata, debug bool) {
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
		meta, entries, structs := parseLibrary(currentPath, currentNamespace, debug)
		catalog.Services = append(catalog.Services, entries...)
		catalog.Structs = append(catalog.Structs, structs...)
		*allMetadata = append(*allMetadata, meta...)
	}

	// 3. If it has nested domains, recurse
	if config.HasNestedDomains {
		for _, domain := range config.Domains {
			subPath := filepath.Join(currentPath, domain)
			// Construct nested namespace: libreria-a.transfers.national
			subNamespace := fmt.Sprintf("%s.%s", currentNamespace, domain)
			crawlLibrary(subPath, subNamespace, catalog, allMetadata, debug)
		}
	}
}

func parseLibrary(path string, namespace string, debug bool) ([]FunctionMetadata, []ServiceEntry, []StructMetadata) {
	fset := token.NewFileSet()
	// Parse only .go files in this directory
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Printf("Warning: error parsing %s: %v", path, err)
		return nil, nil, nil
	}

	var metadata []FunctionMetadata
	var entries []ServiceEntry
	var structs []StructMetadata

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				// 1. Structs
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							if structType, ok := typeSpec.Type.(*ast.StructType); ok {
								if !typeSpec.Name.IsExported() {
									continue
								}
								structName := typeSpec.Name.Name
								var fields []StructField
								for _, field := range structType.Fields.List {
									fType := typeToString(field.Type)
									// Parse Tag
									tag := ""
									if field.Tag != nil {
										tag = strings.ReplaceAll(field.Tag.Value, "`", "")
										tag = strings.ReplaceAll(tag, "json:\"", "")
										tag = strings.ReplaceAll(tag, "\"", "")
									}
									for _, name := range field.Names {
										fields = append(fields, StructField{
											Name:    name.Name,
											Type:    fType,
											JSONTag: tag,
										})
									}
								}
								structs = append(structs, StructMetadata{
									Name:      structName,
									JsonName:  toSnakeCase(structName),
									Fields:    fields,
									Namespace: namespace,
								})
							}
						}
					}
				}

				// 2. Functions
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
	return metadata, entries, structs
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

func writeLocalCatalog(cat Catalog, outputDir string) {
	f, err := os.Create(filepath.Join(outputDir, "catalog.json"))
	if err != nil {
		fmt.Printf("Warning: Could not write catalog.json: %v\n", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(cat)
	fmt.Println("Catalog saved to generated folder.")
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

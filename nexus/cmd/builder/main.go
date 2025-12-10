package main

import (
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
	"text/template"
	"unicode"
)

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

var (
	inputPkg   = flag.String("package", "github.com/japablazatww/libreria-a", "Go package to analyze")
	outputPath = flag.String("output", "../generated", "Path to output generation")
)

func main() {
	flag.Parse()

	inputPath, err := resolvePackagePath(*inputPkg)
	if err != nil {
		log.Fatalf("Error resolving package %s: %v", *inputPkg, err)
	}
	fmt.Printf("Analyzing package at: %s\n", inputPath)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, inputPath, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing directory: %v", err)
	}

	var metadata []FunctionMetadata
	var catalog Catalog

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

					// Return types (simplified for PoC)
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
					catalog.Services = append(catalog.Services, ServiceEntry{
						Namespace:   "libreria-a",
						Method:      fname,
						Description: strings.TrimSpace(fn.Doc.Text()),
						Parameters:  catParams,
					})
				}
			}
		}
	}

	// Ensure output dir exists
	os.MkdirAll(*outputPath, 0755)

	// Generators
	generateTypes(metadata, *outputPath)
	generateServer(metadata, *outputPath)
	generateSDK(metadata, *outputPath)
	generateCatalog(catalog, *outputPath)
}

func resolvePackagePath(pkg string) (string, error) {
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

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

// --- Templates ---

const serverTemplate = `package generated

import (
	"encoding/json"
	"net/http"
	"github.com/japablazatww/libreria-a"
	"fmt"
	"strings"
	"unicode"
)

func RegisterHandlers(mux *http.ServeMux) {
	{{ range . }}
	mux.HandleFunc("/liba/{{ .Name }}", handle{{ .Name }})
	{{ end }}
}

func getParam(params map[string]interface{}, name string) (interface{}, error) {
	// 1. Try exact match
	if v, ok := params[name]; ok { return v, nil }

	// 2. Case-Insensitive Match
	// Create a normalized map where keys are lowercased (without underscores for fuzzy matching might be better, but let's stick to lower case first)
	// For performance in a real app this should be done once per request, but for PoC this function is fine.
	target := strings.ToLower(name)
	targetNoUnderscore := strings.ReplaceAll(target, "_", "")

	for k, v := range params {
		kLower := strings.ToLower(k)
		if kLower == target { return v, nil }
		
		// 3. Fuzzy match (ignoring underscores) e.g. "user_id" vs "userid"
		kNoUnderscore := strings.ReplaceAll(kLower, "_", "")
		if kNoUnderscore == targetNoUnderscore { return v, nil }
	}

	return nil, fmt.Errorf("param %s not found in request params", name)
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
	if len(str) == 0 { return "" }
	return strings.ToUpper(str[:1]) + str[1:]
}



{{ range . }}
func handle{{ .Name }}(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := req.Params
	if params == nil {
		params = make(map[string]interface{})
	}

	// Dynamic Parameter Extraction
	{{ range .Params }}
	val_{{ .Name }}, err := getParam(params, "{{ .Name }}")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_{{ .Name }} {{ .Type }}
	
	switch v := val_{{ .Name }}.(type) {
	case {{ .Type }}:
		arg_{{ .Name }} = v
	{{ if (or (eq .Type "int") (and (eq .Type "float64") false)) }}
	case float64:
		// JSON numbers are float64
		{{ if eq .Type "int" }}arg_{{ .Name }} = int(v){{ end }}
	{{ end }}
	{{ if ne .Type "string" }}
	case string:
		// Try to handle string if needed, currently empty for strict types but avoided duplicate case
	{{ end }}
	default:
		_ = v
	}
	{{ end }}

	// Call underlying library
	{{ if gt (len .Returns) 0 }}res, err := {{ end }}liba.{{ .Name }}(
		{{ range .Params }}arg_{{ .Name }},
		{{ end }}
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": res})
}
{{ end }}
`

const sdkTemplate = `package generated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL    string
	HTTP       *http.Client
	LibreriaA  *LibreriaAClient
}

func NewClient(baseURL string) *Client {
	c := &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}
	c.LibreriaA = &LibreriaAClient{client: c}
	return c
}

type LibreriaAClient struct {
	client *Client
}

{{ range . }}
func (c *LibreriaAClient) {{ .Name }}(req GenericRequest) (interface{}, error) {
	body, _ := json.Marshal(req)
	resp, err := c.client.HTTP.Post(c.client.BaseURL+"/liba/{{ .Name }}", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result["result"], nil
}
{{ end }}
`

func generateServer(meta []FunctionMetadata, outDir string) {
	f, err := os.Create(filepath.Join(outDir, "server_gen.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("server").Parse(serverTemplate))
	tmpl.Execute(f, meta)
}

func generateSDK(meta []FunctionMetadata, outDir string) {
	f, err := os.Create(filepath.Join(outDir, "sdk_gen.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("sdk").Parse(sdkTemplate))
	tmpl.Execute(f, meta)
}

func generateCatalog(cat Catalog, outDir string) {
	// Write to local output
	f, err := os.Create(filepath.Join(outDir, "catalog.json"))
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(cat)
	f.Close()

	// Write to global ~/.nexus/catalog.json
	home, err := os.UserHomeDir()
	if err == nil {
		globalDir := filepath.Join(home, ".nexus")
		os.MkdirAll(globalDir, 0755)
		fGlobal, err := os.Create(filepath.Join(globalDir, "catalog.json"))
		if err == nil {
			encGlobal := json.NewEncoder(fGlobal)
			encGlobal.SetIndent("", "  ")
			encGlobal.Encode(cat)
			fGlobal.Close()
			fmt.Printf("Global catalog updated at: %s\n", filepath.Join(globalDir, "catalog.json"))
		}
	}
}

const typesTemplate = `package generated

// GenericRequest is the standard request envelope
type GenericRequest struct {
	Params map[string]interface{} ` + "`json:\"params\"`" + `
}
`

func generateTypes(meta []FunctionMetadata, outDir string) {
	f, err := os.Create(filepath.Join(outDir, "types_gen.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	tmpl := template.Must(template.New("types").Parse(typesTemplate))
	tmpl.Execute(f, meta)
}

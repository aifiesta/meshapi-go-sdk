// Command dumpsurface prints the SDK's typed surface as JSON for the
// meshapi-contract drift gate. Run it from the repository root.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// SDK type name -> spec component name, where the two differ.
var specNames = map[string]string{
	"ModelInfo":                    "ModelOut",
	"ChatCompletionParams":         "ChatCompletionRequest",
	"ChatMessage":                  "Message",
	"ChatCompletionChoice":         "Choice",
	"CompareParams":                "CompareRequest",
	"CreateBatchParams":            "CreateBatchRequest",
	"CreateTemplateParams":         "CreateTemplateRequest",
	"UpdateTemplateParams":         "UpdateTemplateRequest",
	"EmbeddingsParams":             "EmbeddingsRequest",
	"ImageGenerationParams":        "ImageGenerationRequest",
	"ModerationParams":             "ModerationRequest",
	"ResponsesParams":              "ResponsesRequest",
	"RouterSelectParams":           "RouterSelectRequest",
	"SpeechParams":                 "SpeechRequest",
	"VideoGenerationParams":        "VideoGenerationRequest",
	"WebSearchParams":              "WebSearchRequest",
	"TranscriptionParams":          "Body_create_transcription",
	"AudioTranslationParams":       "Body_create_translation",
	"TranscriptionTranslateParams": "Body_create_transcription_translate",
	"RagFileStatus":                "FileStatusResponse",
	"RagFileListResponse":          "FileListResponse",
	"ImageURL":                     "ImageUrl",
	"VideoURL":                     "VideoUrl",
	"ModerationImageURL":           "ModerationImageUrl",
	"Voice":                        "VoiceInfo",
	"ChatCompletionMessage":        "Message",
}

var verbMethods = map[string]string{
	"get": "GET", "post": "POST", "patch": "PATCH", "delete": "DELETE",
	"getBytes": "GET", "postBytes": "POST", "postMultipart": "POST", "stream": "POST",
}

type surface struct {
	SchemaVersion string      `json:"schema_version"`
	SpecVersion   string      `json:"spec_version"`
	ModelFields   [][2]string `json:"model_fields"`
	Params        []string    `json:"params"`
}

func main() {
	version, err := os.ReadFile(".meshapi-spec-version")
	if err != nil {
		fail(err)
	}
	types := map[string]reflect.Type{}
	rows := map[[2]string]bool{}
	for _, v := range registry {
		t := reflect.TypeOf(v)
		types[t.Name()] = t
		collect(t.Name(), t, rows, map[visit]bool{})
	}
	params, err := sourceParams(".", types)
	if err != nil {
		fail(err)
	}
	out := surface{
		SchemaVersion: "1",
		SpecVersion:   strings.TrimSpace(string(version)),
		ModelFields:   make([][2]string, 0, len(rows)),
		Params:        params,
	}
	for r := range rows {
		out.ModelFields = append(out.ModelFields, r)
	}
	sort.Slice(out.ModelFields, func(i, j int) bool {
		a, b := out.ModelFields[i], out.ModelFields[j]
		return a[0] < b[0] || (a[0] == b[0] && a[1] < b[1])
	})
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	if err := enc.Encode(out); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "dumpsurface:", err)
	os.Exit(1)
}

func structOf(t reflect.Type) (reflect.Type, bool) {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	return t, t.Kind() == reflect.Struct
}

// jsonName returns the wire name of a field, or "" when it is not serialized.
func jsonName(f reflect.StructField) string {
	tag, ok := f.Tag.Lookup("json")
	if !ok {
		return f.Name
	}
	name := strings.Split(tag, ",")[0]
	if name == "-" {
		return ""
	}
	if name == "" {
		return f.Name
	}
	return name
}

func owners(goName string) []string {
	if spec, ok := specNames[goName]; ok {
		return []string{goName, spec}
	}
	return []string{goName}
}

type visit struct {
	name string
	t    reflect.Type
}

func collect(name string, t reflect.Type, rows map[[2]string]bool, seen map[visit]bool) {
	key := visit{name, t}
	if seen[key] {
		return
	}
	seen[key] = true
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		inner, isStruct := structOf(f.Type)
		_, tagged := f.Tag.Lookup("json")
		// encoding/json promotes exported fields of an untagged embedded struct,
		// even when the embedded type itself is unexported.
		if f.Anonymous && isStruct && !tagged {
			collect(name, inner, rows, seen)
			continue
		}
		if !f.IsExported() {
			continue
		}
		if n := jsonName(f); n != "" {
			for _, o := range owners(name) {
				rows[[2]string{o, n}] = true
			}
		}
		if isStruct && inner.Name() != "" && inner.PkgPath() == t.PkgPath() {
			collect(inner.Name(), inner, rows, seen)
		}
	}
}

var verbs = func() map[string]bool {
	m := map[string]bool{}
	for k := range verbMethods {
		m[k] = true
	}
	return m
}()

var camel = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func snake(s string) string {
	s = strings.ReplaceAll(s, "ID", "Id")
	return strings.ToLower(camel.ReplaceAllString(s, "${1}_${2}"))
}

func sourceParams(dir string, types map[string]reflect.Type) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	found := map[string]bool{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
					funcParams(fn, types, found)
				}
			}
		}
	}
	out := make([]string, 0, len(found))
	for p := range found {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

func funcParams(fn *ast.FuncDecl, types map[string]reflect.Type, found map[string]bool) {
	names := map[string]bool{}
	for _, field := range fn.Type.Params.List {
		for _, id := range field.Names {
			if id.Name != "ctx" && id.Name != "_" {
				names[snake(id.Name)] = true
			}
		}
		if t, ok := types[typeName(field.Type)]; ok {
			addStructFields(t, names)
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IndexExpr:
			addLiteral(x.Index, names)
		case *ast.KeyValueExpr:
			addLiteral(x.Key, names)
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Set" && len(x.Args) == 2 {
				addLiteral(x.Args[0], names)
			}
		}
		return true
	})
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !verbs[sel.Sel.Name] || !isHTTPField(sel.X) {
			return true
		}
		path, vars := pathTemplate(call.Args[1])
		if !strings.HasPrefix(path, "/") {
			return true
		}
		scope := verbMethods[sel.Sel.Name] + " " + path
		for n := range names {
			found[scope+":"+n] = true
		}
		for _, v := range vars {
			found[scope+":"+v] = true
		}
		return true
	})
}

func isHTTPField(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "http"
}

func typeName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return typeName(x.X)
	case *ast.ArrayType:
		return typeName(x.Elt)
	}
	return ""
}

func addStructFields(t reflect.Type, names map[string]bool) {
	rows := map[[2]string]bool{}
	collect("", t, rows, map[visit]bool{})
	for r := range rows {
		names[r[1]] = true
	}
}

func addLiteral(e ast.Expr, names map[string]bool) {
	if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		if s, err := strconv.Unquote(lit.Value); err == nil && s != "" {
			names[s] = true
		}
	}
}

// pathTemplate renders a request path expression as "/v1/x/{name}", naming
// each dynamic segment after its Go identifier in snake_case.
func pathTemplate(e ast.Expr) (string, []string) {
	var vars []string
	var render func(ast.Expr) string
	render = func(e ast.Expr) string {
		switch x := e.(type) {
		case *ast.BasicLit:
			s, _ := strconv.Unquote(x.Value)
			return s
		case *ast.BinaryExpr:
			return render(x.X) + render(x.Y)
		case *ast.Ident:
			vars = append(vars, snake(x.Name))
			return "{" + snake(x.Name) + "}"
		case *ast.CallExpr:
			sel, _ := x.Fun.(*ast.SelectorExpr)
			if sel != nil && sel.Sel.Name == "Sprintf" && len(x.Args) > 0 {
				format := render(x.Args[0])
				args := x.Args[1:]
				i := 0
				return regexp.MustCompile(`%[sdv]`).ReplaceAllStringFunc(format, func(string) string {
					if i >= len(args) {
						return ""
					}
					i++
					return render(args[i-1])
				})
			}
			if len(x.Args) == 1 {
				return render(x.Args[0])
			}
		}
		return "{param}"
	}
	return render(e), vars
}

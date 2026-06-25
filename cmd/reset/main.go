package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

const genResetMarker = "generate:reset"

const outFile = "reset.gen.go"

type targetStruct struct {
	Name      string   // struct name
	SmallName string   // struct small (one-letter) name
	Body      []string // string of struct body
}

type templateData struct {
	Package string         // package name
	Structs []targetStruct // structs in package
}

var fileTemplate = template.Must(template.New("reset").Parse(`// Generated code; DO NOT EDIT.

package {{ .Package }}

{{ range .Structs }}
// Reset reset fields {{ .Name }} to their initial values.
func ({{ .SmallName }} *{{ .Name }}) Reset() {
{{- range .Body }}
	{{ . }}
{{- end }}
}
{{ end }}`))

func main() {
	gofile := os.Getenv("GOFILE")
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pkgInfo, err := build.ImportDir(dir, 0)
	if err != nil {
		panic(err)
	}

	if err := processPackage(gofile, dir, pkgInfo); err != nil {
		log.Fatalf("genreset: %v", err)
	}
}

func processPackage(gofile string, dir string, pkgInfo *build.Package) error {
	// temp objects for structs which have generate mark
	resettable := map[string]bool{}
	type pending struct {
		name string
		st   *ast.StructType
	}
	var marked []pending

	// loop through go files in package
	fset := token.NewFileSet()
	markedFiles := []string{}
	for _, filename := range pkgInfo.GoFiles {
		file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		// find generated markers in files
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					if hasMarker(d.Doc) || hasMarker(ts.Doc) {
						resettable[ts.Name.Name] = true
						marked = append(marked, pending{name: ts.Name.Name, st: st})
						markedFiles = append(markedFiles, filename)
					}
				}
			}
		}
	}

	// run processPackage only once per package
	slices.Sort(markedFiles)
	if markedFiles[0] != gofile {
		return nil
	}

	if len(marked) == 0 {
		return nil
	}

	// collect structs templates
	structs := make([]targetStruct, 0, len(marked))
	for _, p := range marked {
		smallName := structSmallName(p.name)
		body := buildBody(smallName, p.st, resettable)
		structs = append(structs, targetStruct{Name: p.name, SmallName: smallName, Body: body})
	}

	// sort structs by name
	sort.Slice(structs, func(i, j int) bool { return structs[i].Name < structs[j].Name })

	var buf bytes.Buffer
	if err := fileTemplate.Execute(&buf, templateData{Package: pkgInfo.Name, Structs: structs}); err != nil {
		return fmt.Errorf("template execution: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting generated code: %w", err)
	}

	outPath := filepath.Join(dir, outFile)
	if err := os.WriteFile(outPath, formatted, os.ModePerm); err != nil {
		return fmt.Errorf("writing generated code %q: %w", outPath, err)
	}

	log.Printf("reset generation: generated %d method(s) in %s", len(structs), outPath)
	return nil
}

func hasMarker(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, s := range strings.Split(doc.Text(), "\n") {
		if strings.HasPrefix(s, genResetMarker) {
			return true
		}
	}
	return false
}

func structSmallName(name string) string {
	nameRunes := []rune(name)
	return string(unicode.ToLower(nameRunes[0]))
}

func buildBody(smallName string, st *ast.StructType, resettable map[string]bool) []string {
	var body []string
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue // встроенные (анонимные) поля для простоты пропускаем
		}
		for _, name := range field.Names {
			target := smallName + "." + name.Name
			body = append(body, fieldReset(target, field.Type, resettable)...)
		}
	}
	return body
}

func fieldReset(target string, typ ast.Expr, resettable map[string]bool) []string {
	switch t := typ.(type) {
	case *ast.ArrayType:
		if t.Len == nil {
			// cut slice by len
			return []string{fmt.Sprintf("%s = %s[:0]", target, target)}
		}
		// slice of fixed length: reset to zero value
		return []string{fmt.Sprintf("%s = %s{}", target, exprString(typ))}
	case *ast.MapType:
		// clear map
		return []string{fmt.Sprintf("clear(%s)", target)}
	case *ast.StarExpr:
		// pointer: reset only if not nil or if has Reset() func - call it
		inner := fieldReset("(*"+target+")", t.X, resettable)
		if id, ok := t.X.(*ast.Ident); ok && resettable[id.Name] {
			inner = []string{fmt.Sprintf("%s.Reset()", target)}
		}
		if len(inner) == 0 {
			return nil
		}
		out := []string{fmt.Sprintf("if %s != nil {", target)}
		for _, line := range inner {
			out = append(out, "\t"+line)
		}
		out = append(out, "}")
		return out
	case *ast.Ident:
		if resettable[t.Name] {
			// inner struct with Reset() field: call it
			return []string{fmt.Sprintf("%s.Reset()", target)}
		}
		if zero, ok := zeroValue(t.Name); ok {
			// assign zero value
			return []string{fmt.Sprintf("%s = %s", target, zero)}
		}
		return nil

	default:
		return nil
	}
}

func exprString(expr ast.Expr) string {
	var buf bytes.Buffer
	_ = format.Node(&buf, token.NewFileSet(), expr)
	return buf.String()
}

func zeroValue(name string) (string, bool) {
	switch name {
	case "string":
		return `""`, true
	case "bool":
		return "false", true
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune",
		"float32", "float64",
		"complex64", "complex128":
		return "0", true
	}
	return "", false
}

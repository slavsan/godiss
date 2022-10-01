package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var currentFile string

type Struct struct {
	Name    string
	Fields  []*Field
	Methods []*Method
}

type Method struct {
	Signature string
}

type Field struct {
	Name string
	Type string
}

type File struct {
	BuildConstraint string
	Path            string
	Structs         []*Struct
}

type Package struct {
	Name  string
	Path  string
	Files []*File
}

func ParsePackage(pkg *Package) error {
	path := pkg.Path

	fset := token.NewFileSet()
	pkgMap, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for pkgName, astPkg := range pkgMap {
		if strings.HasSuffix(pkgName, "_test") {
			continue
		}

		if hasBuldConstraint(astPkg) {
			continue
		}

		pkg.Name = pkgName

		var files []*File

		for fileName, astFile := range astPkg.Files {
			f := &File{}
			f.Path = fileName
			structs := []*Struct{}

			currentFile = fileName

			methods := map[string][]*Method{}

			for _, node := range astFile.Decls {
				switch v := node.(type) {
				case *ast.GenDecl:
					for _, spec := range v.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							s := extractStruct(ts)
							if s == nil {
								continue
							}
							structs = append(structs, s)
						}
					}
				case *ast.FuncDecl:
					if v.Recv == nil {
						continue
					}

					var receiver string
					switch t := v.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						receiver = t.X.(*ast.Ident).Name
					case *ast.Ident:
						receiver = t.Name
					}
					method := v.Name.Name
					params := getCommaSeparated(v.Type.Params)
					results := getCommaSeparated(v.Type.Results)

					signature := fmt.Sprintf(
						"%s(%s) %s",
						method,
						params,
						results,
					)

					if _, ok := methods[receiver]; !ok {
						methods[receiver] = []*Method{}
					}

					methods[receiver] = append(methods[receiver], &Method{
						Signature: signature,
					})

				default:
					panic(fmt.Sprintf("unknown decl: %v", node))
				}
			}

			f.Structs = structs
			files = append(files, f)

			for receiver, m := range methods {
				var s *Struct
				for _, found := range f.Structs {
					if x.Name == receiver {
						s = found
						break
					}
				}

				if s != nil {
					s.Methods = append(s.Methods, m...)
				}
			}
		}

		pkg.Files = files
	}

	return nil
}

func LoadPackages(path string) (map[string]*Package, error) {
	res := map[string]*Package{}

	err := filepath.Walk(
		path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				panic(err)
			}
			if info.IsDir() {
				if strings.Contains(p, "vendor") {
					return nil
				}
				res[p] = &Package{
					Path: p,
				}
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func LoadStructs(path string) ([]*Struct, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var structs []*Struct

	for _, node := range file.Decls {
		switch v := node.(type) {
		case *ast.GenDecl:
			for _, spec := range v.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					s := extractStruct(ts)
					if s == nil {
						continue
					}
					structs = append(structs, s)
				}
			}
		}
	}

	return structs, nil
}

func extractStruct(n *ast.TypeSpec) *Struct {
	s := &Struct{}

	st, ok := n.Type.(*ast.StructType)
	if !ok {
		return nil
	}

	s.Name = n.Name.Name

	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			s.Fields = append(s.Fields, &Field{
				Name: "",
				Type: getType(f.Type),
			})
			continue
		}
		for _, n := range f.Names {
			s.Fields = append(s.Fields, &Field{
				Name: n.Name,
				Type: getType(f.Type),
			})
		}
	}

	return s
}

func getType(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", getType(v.Elt))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", getType(v.X))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getType(v.X), getType(v.Sel))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getType(v.Key), getType(v.Value))
	case *ast.FuncType:
		res := fmt.Sprintf("func(%s) (%s)", getCommaSeparated(v.Params), getCommaSeparated(v.Results))
		if v.Results == nil {
			res = fmt.Sprintf("func(%s)", getCommaSeparated(v.Params))
		}
		return res
	case *ast.InterfaceType:
		return fmt.Sprintf("interface{}")
	case *ast.ChanType:
		if v.Arrow == token.NoPos {
			return fmt.Sprintf("chan %s", getType(v.Value))
		}
		if v.Dir == ast.SEND {
			return fmt.Sprintf("chan<- %s", getType(v.Value))
		}
		if v.Dir == ast.RECV {
			return fmt.Sprintf("<-chan %s", getType(v.Value))
		}
		panic("invalid channel def")
	case *ast.StructType:
		return fmt.Sprintf("struct{ %s }", getStructFields(v.Fields))
	case *ast.IndexExpr:
		return "TODO: INDEX EXPR"
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", getType(v.Elt))
	case *ast.ParenExpr:
		return "TODO: PAREN EXPR"
	default:
		panic(fmt.Sprintf("unhandled type: %#v", e))
	}
}

func getCommaSeparated(fl *ast.FieldList) string {
	fields := []string{}

	if fl == nil {
		return ""
	}

	for _, f := range fl.List {
		fields = append(fields, getType(f.Type))
	}

	return strings.Join(fields, ", ")
}

func getStructFields(fl *ast.FieldList) string {
	fields := []string{}

	if fl == nil {
		return ""
	}

	for _, f := range fl.List {
		if len(f.Names) == 0 {
			if _, ok := f.Type.(*ast.SelectorExpr); ok {
				fields = append(fields, fmt.Sprintf("%s %s", "", getType(f.Type)))
				continue
			}
		}
		fields = append(fields, fmt.Sprintf("%s %s", getType(f.Names[0]), getType(f.Type)))
	}

	return strings.Join(fields, ", ")
}

func FormatPackages(packages map[string]*Package) string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString(`	graph [
		label = "Basic git concepts and operations\n\n"
		labelloc = t
		fontname = "Helvetica,Arial,sans-serif"
		fontsize = 20
		layout = dot
		rankdir = LR
		newrank = true
	]

	node [
		style=filled
		shape=rect
		pencolor="#00000044" // frames color
		fontname="Helvetica,Arial,sans-serif"
		shape=plaintext
	]`)

	sortedPackages := []*Package{}

	for _, p := range packages {
		sortedPackages = append(sortedPackages, p)
	}

	sort.SliceStable(sortedPackages, func(i, j int) bool {
		if sortedPackages[i].Name == sortedPackages[j].Name {
			return sortedPackages[i].Path < sortedPackages[j].Path
		}
		return sortedPackages[i].Name < sortedPackages[j].Name
	})

	for _, pkg := range sortedPackages {
		pkgName := pkg.Path
		sb.WriteString(fmt.Sprintf("\nsubgraph cluster_%s {", normalizePackageName(pkgName)))
		sb.WriteString(fmt.Sprintf(`label = "%s"`, pkgName))
		sb.WriteString("\n\n")
		for _, f := range pkg.Files {
			for _, s := range f.Structs {
				formatStruct(&sb, s)
			}
		}
		sb.WriteString("}\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

func Format(structs []*Struct) string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString(`	graph [
		label = "Basic git concepts and operations\n\n"
		labelloc = t
		fontname = "Helvetica,Arial,sans-serif"
		fontsize = 20
		layout = dot
		rankdir = LR
		newrank = true
	]

	node [
		style=filled
		shape=rect
		pencolor="#00000044" // frames color
		fontname="Helvetica,Arial,sans-serif"
		shape=plaintext
	]`)
	sb.WriteString("\n\n")
	for _, s := range structs {
		formatStruct(&sb, s)
	}

	sb.WriteString("}\n")

	return sb.String()
}

func formatStruct(sb *strings.Builder, s *Struct) {
	sb.WriteString(fmt.Sprintf(`
	"%s" [
		fillcolor="#88ff0022"
		label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
			<tr> <td port="push" sides="ltr"><b>%s</b></td> </tr>
			<tr> <td port="switch" align="left">
				%s
			</td> </tr>
			<tr> <td port="switch" align="left">
				%s
			</td> </tr>
		</table>>
		shape=plain
	]`, s.Name, s.Name, formatStructFields(s), formatStructMethods(s)))
	sb.WriteString("\n")
}

func formatStructFields(s *Struct) string {
	var sb strings.Builder

	for _, f := range s.Fields {
		sb.WriteString(fmt.Sprintf(
			"%s %s<br/>\n",
			strings.ReplaceAll(strings.ReplaceAll(f.Name, "<", "&lt;"), ">", "&gt;"),
			strings.ReplaceAll(strings.ReplaceAll(f.Type, "<", "&lt;"), ">", "&gt;"),
		))
	}

	return sb.String()
}

func formatStructMethods(s *Struct) string {
	var sb strings.Builder

	for _, m := range s.Methods {
		sb.WriteString(fmt.Sprintf(
			"%s<br/>\n",
			strings.ReplaceAll(strings.ReplaceAll(m.Signature, "<", "&lt;"), ">", "&gt;"),
		))
	}

	return sb.String()
}

func hasBuldConstraint(pkg *ast.Package) bool {
	for _, f := range pkg.Files {
		for _, c := range f.Comments {
			for _, l := range c.List {
				if strings.HasPrefix(l.Text, "//go:build") {
					return true
				}
			}
		}
	}
	return false
}

func normalizePackageName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), ".", "_"), "-", "_")
}

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

const (
	NoColor = "\033[0m"
	Red     = "\033[0;31m"
	Green   = "\033[0;32m"
	Yellow  = "\033[0;33m"
	Blue    = "\033[0;34m"
	Purple  = "\033[0;35m"
	Cyan    = "\033[0;36m"
)

type Struct struct {
	Name    string
	Fields  []*Field
	Methods []*Method
}

type Method struct {
	Signature string
}

type Import struct {
	Name   string
	Path   string
	StdLib bool
}

type Field struct {
	Name string
	Type string
}

type File struct {
	Path             string
	BuildConstraints []string
	Structs          []*Struct
	Imports          []*Import
}

type Directory struct {
	Path     string
	Packages map[string]*Package
}

type Package struct {
	Name       string
	Path       string
	ModulePath string
	Files      []*File
}

func ParsePackage(directory *Directory, module, target string) error {
	path := directory.Path

	fset := token.NewFileSet()
	pkgMap, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for pkgName, astPkg := range pkgMap {
		pkg := &Package{}

		if strings.HasSuffix(pkgName, "_test") {
			continue
		}

		modulePath := fmt.Sprintf(
			"%s%s",
			module, strings.TrimPrefix(directory.Path, target),
		)
		pkg.ModulePath = modulePath

		pkg.Name = pkgName

		var files []*File

		for fileName, astFile := range astPkg.Files {
			f := &File{
				Imports: []*Import{},
			}
			f.Path = fileName
			structs := []*Struct{}
			currentFile = fileName
			methods := map[string][]*Method{}

			if bc, ok := buildConstraints(astFile); ok {
				f.BuildConstraints = bc
			}

			for _, node := range astFile.Imports {
				name := ""
				if node.Name != nil {
					name = node.Name.Name
				}
				path := strings.ReplaceAll(node.Path.Value, "\"", "")
				f.Imports = append(f.Imports, &Import{Name: name, Path: path, StdLib: isStdLib(path)})
			}

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
						if indexExpr, ok := t.X.(*ast.IndexExpr); ok {
							receiver = indexExpr.X.(*ast.Ident).Name
							continue
						}

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
					if found.Name == receiver {
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
		directory.Packages[pkgName] = pkg
	}

	return nil
}

func LoadPackages(path, module, target string) (map[string]*Directory, error) {
	res := map[string]*Directory{}

	err := filepath.Walk(
		path,
		func(p string, info os.FileInfo, err error) error {
			// TODO: don't follow symlinks
			if err != nil {
				panic(err)
			}
			if info.IsDir() {
				if strings.Contains(p, "vendor") {
					return nil
				}
				if strings.Contains(p, ".git") {
					return nil
				}
				res[p] = &Directory{
					Path:     p,
					Packages: map[string]*Package{},
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

func FormatPackages(directories map[string]*Directory) string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString(`    graph [
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
    ]
`)

	for _, directory := range DirectoryMap(directories).SortedDirectories() {
		for _, pkg := range PackagesMap(directory.Packages).SortedPackages() {
			sb.WriteString(fmt.Sprintf("\n    subgraph cluster_%s {", normalizePackageName(directory.Path)))
			sb.WriteString(fmt.Sprintf("\n        label = \"%s\"", directory.Path))
			sb.WriteString("\n")
			for _, f := range pkg.Files {
				for _, s := range f.Structs {
					formatStruct(&sb, s, true)
				}
			}
			sb.WriteString("    }\n")
		}
	}

	sb.WriteString("}\n")

	return sb.String()
}

func Format(structs []*Struct) string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString(`    graph [
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
	sb.WriteString("\n")
	for _, s := range structs {
		formatStruct(&sb, s, false)
	}

	sb.WriteString("}\n")

	return sb.String()
}

func pad(indent bool, x string) string {
	if !indent {
		return x
	}
	lines := strings.Split(x, "\n")
	for i, l := range lines {
		if l == "" {
			continue
		}
		lines[i] = fmt.Sprintf("%s%s", tab, l)
	}
	return strings.Join(lines, "\n")
}

func formatStruct(sb *strings.Builder, s *Struct, indent bool) {
	sb.WriteString(pad(indent, fmt.Sprintf(`
    "%s" [
        fillcolor="#88ff0022"
        label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="3">
            <tr><td port="push" sides="ltr"><b>%s</b></td></tr>
            <tr><td port="switch" align="left">%s
            </td></tr>
            <tr><td port="switch" align="left">%s
            </td></tr>
        </table>>
        shape=plain
    ]`, s.Name, s.Name, formatStructFields(s), formatStructMethods(s))))
	sb.WriteString("\n")
}

const tab = "    "

func escape(v string) string {
	if v == "" {
		return v
	}
	return strings.ReplaceAll(strings.ReplaceAll(v, "<", "&lt;"), ">", "&gt;")
}

func formatStructFields(s *Struct) string {
	var sb strings.Builder

	for _, f := range s.Fields {
		if f.Name == "" {
			sb.WriteString(fmt.Sprintf(
				"\n%s%s<br/>",
				strings.Repeat(tab, 4),
				escape(f.Type),
			))
			continue
		}
		sb.WriteString(fmt.Sprintf(
			"\n%s%s %s<br/>",
			strings.Repeat(tab, 4),
			escape(f.Name),
			escape(f.Type),
		))
	}

	return sb.String()
}

func formatStructMethods(s *Struct) string {
	var sb strings.Builder

	for _, m := range s.Methods {
		sb.WriteString(fmt.Sprintf(
			"\n%s%s<br/>",
			strings.Repeat(tab, 4),
			strings.ReplaceAll(strings.ReplaceAll(m.Signature, "<", "&lt;"), ">", "&gt;"),
		))
	}

	return sb.String()
}

func buildConstraints(f *ast.File) ([]string, bool) {
	var found []string
	for _, c := range f.Comments {
		for _, l := range c.List {
			if strings.HasPrefix(l.Text, "//go:build") {
				found = append(found, strings.ReplaceAll(l.Text, "//go:build ", ""))
			}
		}
	}
	return found, len(found) > 0
}

func normalizePackageName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), ".", "_"), "-", "_")
}

func FormatImports(directories map[string]*Directory) string {
	var sb strings.Builder
	sb.WriteString("digraph {\n")
	sb.WriteString("    rankdir=\"LR\"\n\n")

	for _, directory := range DirectoryMap(directories).SortedDirectories() {
		for _, pkg := range PackagesMap(directory.Packages).SortedPackages() {
			if isFake(pkg.Name) || isMock(pkg.Name) || isTest(pkg.Name) {
				continue
			}

			unique := map[string]struct{}{}
			for _, f := range pkg.Files {
				for _, i := range f.Imports {
					if _, ok := unique[i.Path]; !ok {
						unique[i.Path] = struct{}{}
					}
				}
			}
			sortedUnique := make([]string, 0, len(unique))
			for i, _ := range unique {
				sortedUnique = append(sortedUnique, i)
			}
			sort.Strings(sortedUnique)

			for _, i := range sortedUnique {
				if isStdLib(strings.ReplaceAll(i, "\"", "")) {
					continue
				}

				sb.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\"\n", pkg.ModulePath, i))
			}
		}
	}
	sb.WriteString("}\n")
	return sb.String()
}

func FormatImportsTable(directories map[string]*Directory, module string) string {
	var sb strings.Builder

	type Stat struct {
		Path  string
		Count int
	}

	stats := map[string]Stat{}

	for _, directory := range directories {
		for _, pkg := range directory.Packages {
			unique := map[string]struct{}{}
			for _, f := range pkg.Files {
				for _, i := range f.Imports {
					if _, ok := unique[i.Path]; !ok {
						unique[i.Path] = struct{}{}
					}
				}
			}

			for p, _ := range unique {
				if _, ok := stats[p]; ok {
					stat := stats[p]
					stat.Count++
					stats[p] = stat
				} else {
					stats[p] = Stat{Path: p, Count: 1}
				}
			}
		}
	}

	sortedStats := make([]Stat, 0, len(stats))

	max := 0

	for _, stat := range stats {
		if stat.Count > max {
			max = stat.Count
		}
		sortedStats = append(sortedStats, stat)
	}

	sort.SliceStable(sortedStats, func(i, j int) bool {
		if sortedStats[i].Count == sortedStats[j].Count {
			return sortedStats[i].Path < sortedStats[j].Path
		}
		return sortedStats[i].Count > sortedStats[j].Count
	})

	for _, stat := range sortedStats {
		// if isStdLib(stat.Path) {
		// 	continue
		// }
		sb.WriteString(fmt.Sprintf(
			"%*d %s\n",
			digitsCount(max), stat.Count, colorize(stat.Path, module),
		))
	}

	return sb.String()
}

func formatStructFieldVisibility(f *Field) string {
	if len(f.Name) > 0 {
		return formatTokenVisibility(f.Name)
	}
	return formatTokenVisibility(f.Type)
}

func formatTokenVisibility(token string) string {
	if ast.IsExported(token) {
		return fmt.Sprintf("%s+%s ", Green, NoColor)
	}
	return fmt.Sprintf("%s-%s ", Red, NoColor)
}

func maybeAddBuildConstraint(f *File) string {
	if len(f.BuildConstraints) > 0 {
		return fmt.Sprintf(" %s%s%s", Red, strings.Join(f.BuildConstraints, ","), NoColor)
	}
	return ""
}

func formatStructForConsole(s *Struct, f *File) string {
	var sb strings.Builder

	// TODO: sort fields and methods by visibility (or alphabetically, or do no sorting optionally)
	// TODO: move visibility logic to fields and methods parsing
	sb.WriteString(fmt.Sprintf(
		"%stype %s {%s\n",
		formatTokenVisibility(s.Name), s.Name,
		maybeAddBuildConstraint(f),
	))
	for _, f := range s.Fields {
		if f.Name == "" {
			sb.WriteString(fmt.Sprintf("    %s%s\n", formatStructFieldVisibility(f), f.Type))
		} else {
			sb.WriteString(fmt.Sprintf("    %s%s %s\n", formatStructFieldVisibility(f), f.Name, f.Type))
		}
	}
	for _, m := range s.Methods {
		sb.WriteString(fmt.Sprintf("    %s%s\n", formatTokenVisibility(m.Signature), m.Signature))
	}
	sb.WriteString("}\n")

	return sb.String()
}

func FormatEntrypoints(directories map[string]*Directory, module string) string {
	var sb strings.Builder

	for _, directory := range DirectoryMap(directories).SortedDirectories() {
		for _, pkg := range PackagesMap(directory.Packages).SortedPackages() {
			for _, f := range pkg.Files {
				if strings.HasSuffix(f.Path, "/main.go") {
					sb.WriteString(fmt.Sprintf("%s %s%s%s\n", pkg.ModulePath, Red, strings.Join(f.BuildConstraints, ","), NoColor))
				}
			}
		}
	}

	return sb.String()
}

func FormatTypes(directories map[string]*Directory, module string) string {
	var sb strings.Builder

	for _, directory := range DirectoryMap(directories).SortedDirectories() {
		for _, pkg := range PackagesMap(directory.Packages).SortedPackages() {
			sb.WriteString(fmt.Sprintf("%s%s%s\n", Yellow, pkg.ModulePath, NoColor))

			pkgEmpty := true

			sort.Sort(ByFilePath(pkg.Files))

			for _, f := range pkg.Files {
				sort.Sort(ByStructName(f.Structs))

				for _, s := range f.Structs {
					sb.WriteString(fmt.Sprintf("\n%s", formatStructForConsole(s, f)))
					pkgEmpty = false
				}
			}

			if !pkgEmpty {
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

func digitsCount(num int) int {
	var count int
	for num > 0 {
		num = num / 10
		count++
	}
	return count
}

func colorize(path, module string) string {
	color := NoColor
	if strings.HasPrefix(path, module) {
		color = Green
	} else if isStdLib(path) {
		color = Yellow
	}
	return fmt.Sprintf("%s%s%s", color, path, NoColor)
}

var stdLib = map[string]struct{}{
	"bufio":           {},
	"bytes":           {},
	"context":         {},
	"crypto/md5":      {},
	"crypto/sha256":   {},
	"crypto/tls":      {},
	"crypto/x509":     {},
	"embed":           {},
	"encoding/binary": {},
	"encoding/json":   {},
	"encoding/xml":    {},
	"errors":          {},
	"flag":            {},
	"fmt":             {},
	"io":              {},
	"io/fs":           {},
	"io/ioutil":       {},
	"log":             {},
	"math":            {},
	"math/rand":       {},
	"net":             {},
	"net/http":        {},
	"net/http/pprof":  {},
	"net/url":         {},
	"os":              {},
	"os/exec":         {},
	"os/signal":       {},
	"os/user":         {},
	"path":            {},
	"path/filepath":   {},
	"reflect":         {},
	"regexp":          {},
	"runtime":         {},
	"runtime/debug":   {},
	"sort":            {},
	"strconv":         {},
	"strings":         {},
	"sync":            {},
	"sync/atomic":     {},
	"syscall":         {},
	"testing":         {},
	"text/tabwriter":  {},
	"text/template":   {},
	"time":            {},
	"unicode":         {},
	"unsafe":          {},
}

func isStdLib(path string) bool {
	if _, ok := stdLib[path]; ok {
		return true
	}
	if strings.HasPrefix(path, "go/") {
		return true
	}
	return false
}

func isFake(name string) bool {
	return name == "fake"
}

func isMock(name string) bool {
	return name == "mock"
}

func isTest(name string) bool {
	return name == "test"
}

type DirectoryMap map[string]*Directory

func (dirsMap DirectoryMap) SortedDirectories() []*Directory {
	sortedDirectories := make([]*Directory, 0, len(dirsMap))

	for _, p := range dirsMap {
		sortedDirectories = append(sortedDirectories, p)
	}

	sort.SliceStable(sortedDirectories, func(i, j int) bool {
		return sortedDirectories[i].Path < sortedDirectories[j].Path
	})

	return sortedDirectories
}

type PackagesMap map[string]*Package

func (pkgMap PackagesMap) SortedPackages() []*Package {
	sortedPackages := make([]*Package, 0, len(pkgMap))

	for _, p := range pkgMap {
		sortedPackages = append(sortedPackages, p)
	}

	sort.SliceStable(sortedPackages, func(i, j int) bool {
		if sortedPackages[i].Name == sortedPackages[j].Name {
			return sortedPackages[i].Path < sortedPackages[j].Path
		}
		return sortedPackages[i].Name < sortedPackages[j].Name
	})

	return sortedPackages
}

type ByFilePath []*File

func (f ByFilePath) Len() int           { return len(f) }
func (f ByFilePath) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByFilePath) Less(i, j int) bool { return f[i].Path < f[j].Path }

type ByStructName []*Struct

func (s ByStructName) Len() int           { return len(s) }
func (s ByStructName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByStructName) Less(i, j int) bool { return s[i].Name < s[j].Name }

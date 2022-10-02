package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/slavsan/gog/internal"
)

func main() {
	if len(os.Args) < 3 {
		panic("invalid arguments")
	}

	target, err := filepath.Abs(os.Args[2])
	if err != nil {
		panic(err)
	}

	bytes, err := ioutil.ReadFile(path.Join(target, "go.mod"))
	if err != nil {
		panic(err)
	}

	module, err := getModule(string(bytes))
	if err != nil {
		panic(err)
	}

	var structs []*internal.Struct
	var packages map[string]*internal.Package

	switch os.Args[1] {
	case "structs":
		structs, err = internal.LoadStructs(target)
		if err != nil {
			panic(err)
		}
		fmt.Println(internal.Format(structs))
	case "packages":
		packages, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, p := range packages {
			internal.ParsePackage(p)
		}
		fmt.Println(internal.FormatPackages(packages))
	case "imports":
		packages, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, p := range packages {
			internal.ParsePackage(p)
		}

		fmt.Printf("%s\n", internal.FormatImports(packages))
	case "imports_table":
		packages, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, p := range packages {
			internal.ParsePackage(p)
		}

		fmt.Printf("%s", internal.FormatImportsTable(packages, module))
	default:
		panic(fmt.Sprintf("unknown subcommand: %s", os.Args[1]))
	}
}

func getModule(content string) (string, error) {
	lines := strings.Split(content, "\n")
	return strings.ReplaceAll(lines[0], "module ", ""), nil
}

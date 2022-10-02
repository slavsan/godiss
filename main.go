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
	var directories map[string]*internal.Directory

	switch os.Args[1] {
	case "structs":
		structs, err = internal.LoadStructs(target)
		if err != nil {
			panic(err)
		}
		fmt.Println(internal.Format(structs))
	case "packages":
		directories, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, directory := range directories {
			internal.ParsePackage(directory, module, target)
		}
		fmt.Println(internal.FormatPackages(directories))
	case "imports":
		directories, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, directory := range directories {
			internal.ParsePackage(directory, module, target)
		}

		fmt.Printf("%s\n", internal.FormatImports(directories))
	case "imports_table":
		directories, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, directory := range directories {
			internal.ParsePackage(directory, module, target)
		}

		fmt.Printf("%s", internal.FormatImportsTable(directories, module))
	case "types":
		directories, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, p := range directories {
			internal.ParsePackage(p, module, target)
		}

		fmt.Printf("%s", internal.FormatTypes(directories, module))
	case "entrypoints":
		directories, err = internal.LoadPackages(target, module, target)
		if err != nil {
			panic(err)
		}
		for _, directory := range directories {
			internal.ParsePackage(directory, module, target)
		}

		fmt.Printf("%s", internal.FormatEntrypoints(directories, module))
	default:
		panic(fmt.Sprintf("unknown subcommand: %s", os.Args[1]))
	}
}

func getModule(content string) (string, error) {
	lines := strings.Split(content, "\n")
	return strings.ReplaceAll(lines[0], "module ", ""), nil
}

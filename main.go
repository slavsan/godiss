package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/slavsan/gog/internal"
)

func main() {
	if len(os.Args) < 3 {
		panic("invalid arguments")
	}

	path, err := filepath.Abs(os.Args[2])
	if err != nil {
		panic(err)
	}

	var structs []*internal.Struct
	var packages map[string]*internal.Package

	switch os.Args[1] {
	case "structs":
		structs, err = internal.LoadStructs(path)
		if err != nil {
			panic(err)
		}
		fmt.Println(internal.Format(structs))
	default:
		packages, err = internal.LoadPackages(path)
		if err != nil {
			panic(err)
		}
		for _, p := range packages {
			internal.ParsePackage(p)
		}
		fmt.Println(internal.FormatPackages(packages))
	}
}

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/slavsan/godiss/internal"
)

func packages() *Command {
	return &Command{
		Name:        "packages",
		Description: "Display packages in a project",
		DefaultArg:  ".",
		Run: func(args []string) error {
			var target string
			var module string
			var err error
			var directories map[string]*internal.Directory

			target, err = filepath.Abs(args[0])
			if err != nil {
				return err
			}

			module, err = getModule(target)
			if err != nil {
				return err
			}

			directories, err = internal.LoadPackages(target, module, target)
			if err != nil {
				return err
			}

			for _, directory := range directories {
				internal.ParsePackage(directory, module, target, &internal.Config{})
			}

			fmt.Println(internal.FormatPackages(directories))

			return nil
		},
	}
}

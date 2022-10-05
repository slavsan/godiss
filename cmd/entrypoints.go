package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/slavsan/godiss/internal"
)

func entrypoints() *Command {
	command := &Command{
		Name:        "entrypoints",
		Description: "Display entrypoints",
		Subcommands: map[string]*Command{},
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

			fmt.Printf("%s", internal.FormatEntrypoints(directories, module))

			return nil
		},
	}

	return command
}

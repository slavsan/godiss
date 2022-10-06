package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/slavsan/godiss/internal"
)

func stats() *Command {
	command := &Command{
		Name:        "stats",
		Description: "Display stats",
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

			config := &internal.Config{
				IncludeTests: true,
			}

			for _, directory := range directories {
				internal.ParsePackage(directory, module, target, config)
			}

			fmt.Printf("%s", internal.FormatStats(directories, module))

			return nil
		},
	}

	return command
}

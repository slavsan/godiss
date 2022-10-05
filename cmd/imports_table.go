package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/slavsan/godiss/internal"
)

func imports_table() *Command {
	var command *Command
	command = &Command{
		Name:        "imports_table",
		Description: "Display imports (in a table)",
		DefaultArg:  ".",
		Flags: map[string]*Flag{
			"nostdlib": {"n", false, "exclude stdlib packages"},
			"select":   {"s", "", "select packages"},
		},
		Run: func(args []string) error {
			var target string
			var module string
			var err error
			var directories map[string]*internal.Directory

			excludeStdLib := command.Flags["nostdlib"].Value.(bool)
			selected := command.Flags["select"].Value.(string)

			target, err = filepath.Abs(args[0])
			if err != nil {
				panic(err)
			}

			module, err = getModule(target)
			if err != nil {
				panic(err)
			}

			directories, err = internal.LoadPackages(target, module, target)
			if err != nil {
				panic(err)
			}

			config := &internal.Config{
				ExcludeStdLib: excludeStdLib,
				Select:        createSet(selected),
			}

			for _, directory := range directories {
				internal.ParsePackage(directory, module, target, config)
			}

			fmt.Printf("%s", internal.FormatImportsTable(directories, module, config))

			return nil
		},
	}
	return command
}

func createSet(value string) map[string]struct{} {
	items := strings.Split(value, ",")
	set := make(map[string]struct{}, len(items))
	for _, i := range items {
		if i == "" {
			continue
		}
		set[i] = struct{}{}
	}
	return set
}

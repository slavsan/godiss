package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/slavsan/godiss/internal"
)

func types() *Command {
	var command *Command
	command = &Command{
		Name:        "types",
		Description: "Display defined types",
		DefaultArg:  ".",
		Flags: map[string]*Flag{
			"exclude":      {"e", "", "exclude packages"},
			"select-exact": {"E", "", "select exact packages"},
			"select":       {"s", "", "select packages"},
		},
		Run: func(args []string) error {
			var target string
			var module string
			var err error
			var directories map[string]*internal.Directory

			exclude := command.Flags["exclude"].Value.(string)
			selectExact := command.Flags["select-exact"].Value.(string)
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
				Exclude:     createSet(exclude),
				SelectExact: createSet(selectExact),
				Select:      createSet(selected),
			}

			for _, p := range directories {
				internal.ParsePackage(p, module, target, config)
			}

			fmt.Printf("%s", internal.FormatTypes(directories, module))

			return nil
		},
	}
	return command
}

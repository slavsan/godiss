package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/slavsan/godiss/internal"
)

func structs() *Command {
	return &Command{
		Name:        "structs",
		Description: "Display structs defined in a file",
		Run: func(args []string) error {
			var target string
			var err error
			var structs []*internal.Struct

			target, err = filepath.Abs(args[0])
			if err != nil {
				return err
			}

			structs, err = internal.LoadStructs(target)
			if err != nil {
				return err
			}

			fmt.Println(internal.Format(structs))

			return nil
		},
	}
}

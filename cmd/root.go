package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
)

type Command struct {
	Name        string
	Description string
	Subcommands map[string]*Command
	Flags       map[string]*Flag
	DefaultArg  string
	Run         func(args []string) error
}

type Flag struct {
	Short string
	Value any
	Usage string
}

func (c *Command) Add(sub *Command) {
	if _, ok := c.Subcommands[sub.Name]; ok {
		panic(fmt.Sprintf("subcommand already exists: %s", sub.Name))
	}
	c.Subcommands[sub.Name] = sub
}

type Executor struct {
	Arg int
}

func NewExecutor() *Executor {
	return &Executor{
		Arg: 0,
	}
}

func (e *Executor) Usage(c *Command) {
	if c.Subcommands == nil {
		return
	}

	sorted := make([]string, 0, len(c.Subcommands))
	max := 0

	for _, sub := range c.Subcommands {
		if len(sub.Name) > max {
			max = len(sub.Name)
		}

		sorted = append(sorted, sub.Name)
	}

	sort.Strings(sorted)

	fmt.Printf("Subcommands:\n")
	for _, sub := range sorted {
		command := c.Subcommands[sub]
		fmt.Printf("\t%-*s\t%s\n", max, command.Name, command.Description)
	}
}

func showHelp(c *Command) {
	max := 0
	for k := range c.Flags {
		if len(k) > max {
			max = len(k)
		}
	}

	fmt.Printf("Usage:\n")
	for k, f := range c.Flags {
		fmt.Printf(
			"\t-%s, --%s%s  %s\n",
			f.Short, k, strings.Repeat(" ", max-len(k)), f.Usage,
		)
	}
}

func (e *Executor) Execute(c *Command) {
	e.Arg++
	if len(c.Subcommands) > 0 {
		var arg string
		if len(os.Args) > e.Arg {
			arg = os.Args[e.Arg]
		}

		if arg == "help" {
			e.Arg++
			if len(os.Args) > e.Arg {
				arg = os.Args[e.Arg]
			}

			if command, ok := c.Subcommands[arg]; ok {
				showHelp(command)
				return
			}

			fmt.Printf("unknown subcommand: %s\n", arg)
			os.Exit(1)
		}

		if command, ok := c.Subcommands[arg]; ok {
			e.Execute(command)
			return
		}

		e.Usage(c)
		return
	}

	flagSets := map[string]any{}
	flagSet := flag.NewFlagSet(c.Name, flag.ContinueOnError)

	for k, f := range c.Flags {
		switch v := f.Value.(type) {
		case bool:
			flagSets[k] = flagSet.Bool(k, v, f.Usage)
			flagSets[k] = flagSet.Bool(f.Short, v, f.Usage)
		case string:
			flagSets[k] = flagSet.String(k, v, f.Usage)
			flagSets[k] = flagSet.String(f.Short, v, f.Usage)
		default:
			panic(fmt.Sprintf("unhandled type: %#v - (%#v)", reflect.TypeOf(v), f))
		}
	}

	flagSet.SetOutput(&bytes.Buffer{})

	err := flagSet.Parse(os.Args[e.Arg:])
	if err != nil {
		fmt.Printf("%s\n\n", err.Error())

		showHelp(c)
		os.Exit(1)
	}

	for k, v := range flagSets {
		switch f := v.(type) {
		case *string:
			(*c).Flags[k].Value = *f
		case *bool:
			(*c).Flags[k].Value = *f
		default:
			panic(fmt.Sprintf("unhandled type: %#v\n", f))
		}
	}

	args := flagSet.Args()

	if len(args) == 0 {
		if c.DefaultArg == "" {
			fmt.Printf("emtpy args\n")
			e.Usage(c)
			os.Exit(1)
		}
		args = append(args, c.DefaultArg)
	}

	err = c.Run(args)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

func root() *Command {
	command := &Command{
		Name:        "root",
		Description: "",
		Subcommands: map[string]*Command{},
	}

	command.Add(structs())
	command.Add(packages())
	command.Add(imports())
	command.Add(imports_table())
	command.Add(types())
	command.Add(entrypoints())

	return command
}

func Execute() {
	NewExecutor().Execute(root())
}

func getModule(target string) (string, error) {
	bytes, err := ioutil.ReadFile(path.Join(target, "go.mod"))
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(bytes), "\n")
	return strings.ReplaceAll(lines[0], "module ", ""), nil
}

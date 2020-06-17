package flags

import (
	"fmt"
	"sort"
	"strings"
)

// Function defines the type signature of an executable function.
type Function func(ctx *Context) error

// Command represents a pair of a Function and its Description.
type Command struct {
	Desc string
	Func Function
}

// CommandSet is a map of Commands and its names.
type CommandSet map[string]Command

// Register a Function with the given name and description.
func (set CommandSet) Register(name, desc string, f Function) {
	set[name] = Command{desc, f}
}

// Help lists the names and descriptions of the commands registered.
func (set CommandSet) Help() string {
	names := make([]string, len(set))
	i := 0
	for name := range set {
		names[i] = name
		i++
	}
	sort.Strings(names)
	builder := strings.Builder{}
	builder.WriteString("available commands:")
	for _, name := range names {
		cmd := set[name]
		builder.WriteString("\n" + formatHelp(name, cmd.Desc))
	}
	return builder.String()
}

// Compile the CommandSet into a single Function.
func (set CommandSet) Compile() Function {
	return func(ctx *Context) error {
		if len(ctx.Args) == 0 {
			return fmt.Errorf("%s expected a command.\n\n%s", ctx.Name, set.Help())
		}
		head, tail := shift(ctx.Args)
		if (strings.HasPrefix(head, "-") && strings.Contains(head, "h")) || head == "--help" {
			return fmt.Errorf("%s: %s\n\n%s", ctx.Name, ctx.Desc, set.Help())
		}
		if head == "--version" {
			fmt.Println(ctx.Version.String())
			return nil
		}
		cmd, ok := set[head]
		if !ok {
			return fmt.Errorf("unknown command name `%s`", head)
		}
		name := fmt.Sprintf("%s %s", ctx.Name, head)
		return cmd.Func(&Context{name, cmd.Desc, ctx.Version, tail, ctx.Ctx})
	}
}

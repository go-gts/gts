package flags

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/go-wrap/wrap"
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

// Ronn creates a manpage markdown template for ronn.
func (set CommandSet) Ronn(ctx *Context) error {
	name, desc := ctx.Name, ctx.Desc
	usage := fmt.Sprintf("usage: %s [--version] [-h | --help] <command> [<args>]", ctx.Name)
	name = strings.ReplaceAll(name, " ", "-")
	filename := fmt.Sprintf("%s.1.md", name)

	f, err := os.Create(filename)
	if err != nil {
		return ctx.Raise(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	parts := []string{
		fmt.Sprintf("# %s -- %s", name, desc),
		"## SYNOPSIS",
		usage,
		"## DESCRIPTION",
		sentencify(desc),
		"## COMMANDS",
	}

	commands := []string{}
	seealso := []string{}

	cmdNames := make([]string, len(set))

	i := 0
	for name := range set {
		cmdNames[i] = name
		i++
	}

	sort.Strings(cmdNames)

	for _, cmdName := range cmdNames {
		cmd := set[cmdName]
		s := fmt.Sprintf("  * `%s-%s(1)`:\n    %s", name, cmdName, sentencify(cmd.Desc))
		commands = append(commands, s)
		s = fmt.Sprintf("%s-%s(1)", name, cmdName)
		seealso = append(seealso, s)
	}

	parts = append(parts, commands...)
	parts = append(parts, []string{
		"## BUGS",
		fmt.Sprintf("**%s** currently has no known bugs.", name),
		"## AUTHORS",
		fmt.Sprintf("**%s** is written and maintained by @AUTHOR@.", name),
		"## SEE ALSO",
		strings.Join(seealso, ", "),
	}...)

	s := strings.Join(parts, "\n\n")
	s = wrap.Space(s, 80)
	if n, err := io.WriteString(w, s); err != nil || n != len(s) {
		if n != len(s) {
			return ctx.Raise(fmt.Errorf("wrote %d of %d bytes", n, len(s)))
		}
		return ctx.Raise(err)
	}

	return w.Flush()
}

// Help lists the names and descriptions of the commands registered.
func (set CommandSet) Help(ctx *Context) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("usage: %s [--version] [-h | --help] <command> [<args>]\n\n", ctx.Name))

	names := make([]string, len(set))

	i := 0
	for name := range set {
		names[i] = name
		i++
	}

	sort.Strings(names)

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
			return fmt.Errorf("%s expected a command.\n\n%s", ctx.Name, set.Help(ctx))
		}

		head, tail := shift(ctx.Args)
		if (strings.HasPrefix(head, "-") && strings.Contains(head, "h")) || head == "--help" {
			return fmt.Errorf("%s: %s\n\n%s", ctx.Name, ctx.Desc, set.Help(ctx))
		}

		if head == "--ronn" {
			if err := set.Ronn(ctx); err != nil {
				return err
			}
			return ctx.Raise(errors.New("created ronn template"))
		}

		cmd, ok := set[head]
		if !ok {
			return fmt.Errorf("unknown command name `%s`", head)
		}

		name := fmt.Sprintf("%s %s", ctx.Name, head)
		return cmd.Func(&Context{name, cmd.Desc, tail, ctx.Ctx})
	}
}

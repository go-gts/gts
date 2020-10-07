package flags

import (
	"bufio"
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

// Commands returns the list of command names in alphabetical order.
func (set CommandSet) Commands() []string {
	names := make([]string, len(set))
	i := 0
	for name := range set {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

// Ronn creates a manpage markdown template for ronn.
func (set CommandSet) Ronn(ctx *Context) error {
	usage := fmt.Sprintf("usage: %s [--version] [-h | --help] <command> [<args>]", ctx.JoinedName())
	name := strings.Join(ctx.Name, "-")
	filename := fmt.Sprintf("%s.1.ronn", name)

	f, err := os.Create(filename)
	if err != nil {
		return ctx.Raise(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	parts := []string{
		fmt.Sprintf("# %s -- %s", name, ctx.Desc),
		"## SYNOPSIS",
		usage,
		"## DESCRIPTION",
		sentencify(ctx.Desc),
		"## COMMANDS",
	}

	commands := []string{}
	seealso := []string{}

	cmdNames := set.Commands()

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

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	return nil
}

func (set CommandSet) compBash(ctx *Context) error {
	funcName := strings.Join(ctx.Name, "_")

	cmdNames := set.Commands()
	cmdFuncs := make([]string, len(cmdNames))

	for i, cmdName := range cmdNames {
		cmdFuncs[i] = fmt.Sprintf("%[1]s) &_%[2]s_%[1]s ;;", cmdName, funcName)
	}

	comps := strings.Join(cmdNames, " ")
	funcs := alignLines(strings.Join(cmdFuncs, "\n"), '&')
	funcs = strings.ReplaceAll(funcs, "\n", "\n        ")

	comp := fmt.Sprintf(compSetBashFormat, funcName, comps, funcs)

	filename := fmt.Sprintf("%s-completion.bash", ctx.Name[0])
	return fileAppend(filename, comp)
}

func (set CommandSet) compZsh(ctx *Context) error {
	funcName := strings.Join(ctx.Name, "_")

	cmdNames := set.Commands()
	cmdList := make([]string, len(cmdNames))
	cmdFuncs := make([]string, len(cmdNames))

	for i, cmdName := range cmdNames {
		cmdList[i] = fmt.Sprintf("'%s:%s'", cmdName, set[cmdName].Desc)
		cmdFuncs[i] = fmt.Sprintf("%[1]s) &_%[2]s_%[1]s ;;", cmdName, funcName)
	}

	list := strings.Join(cmdList, "\n            ")
	funcs := alignLines(strings.Join(cmdFuncs, "\n"), '&')
	funcs = strings.ReplaceAll(funcs, "\n", "\n        ")

	comp := fmt.Sprintf(compSetZshFormat, funcName, list, funcs)

	filename := fmt.Sprintf("%s-completion.zsh", ctx.Name[0])
	return fileAppend(filename, comp)
}

// Comp creates a completion script.
func (set CommandSet) Comp(ctx *Context) error {
	if err := set.compBash(ctx); err != nil {
		return ctx.Raise(err)
	}

	if err := set.compZsh(ctx); err != nil {
		return ctx.Raise(err)
	}

	return nil
}

// Help lists the names and descriptions of the commands registered.
func (set CommandSet) Help(ctx *Context) string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("usage: %s [--version] [-h | --help] <command> [<args>]\n\n", ctx.JoinedName()))

	names := make([]string, len(set))

	i := 0
	for name := range set {
		names[i] = name
		i++
	}

	sort.Strings(names)

	b.WriteString("available commands:")
	for _, name := range names {
		cmd := set[name]
		b.WriteString("\n" + formatHelp(name, cmd.Desc))
	}
	return b.String()
}

// Compile the CommandSet into a single Function.
func (set CommandSet) Compile() Function {
	return func(ctx *Context) error {
		if len(ctx.Args) == 0 {
			return fmt.Errorf("%s expected a command.\n\n%s", ctx.JoinedName(), set.Help(ctx))
		}

		head, tail := shift(ctx.Args)
		if (strings.HasPrefix(head, "-") && strings.Contains(head, "h")) || head == "--help" {
			return fmt.Errorf("%s: %s\n\n%s", ctx.JoinedName(), ctx.Desc, set.Help(ctx))
		}

		switch head {
		case "generate-ronn-templates":
			if err := set.Ronn(ctx); err != nil {
				return fmt.Errorf("while generating ronn file for %s: %v", ctx.JoinedName(), err)
			}
			for name, cmd := range set {
				child := &Context{append(ctx.Name, name), cmd.Desc, ctx.Args, ctx.Ctx}
				if err := cmd.Func(child); err != errRonn {
					return fmt.Errorf("while generating ronn file for %s: %v", name, err)
				}
			}
			return nil

		case "generate-completions":
			if len(ctx.Name) == 1 {
				bash := fmt.Sprintf("%s-completion.bash", ctx.Name[0])
				if err := touch(bash); err != nil {
					return fmt.Errorf("while generating completion for %s: %v", ctx.Name[0], err)
				}

				zsh := fmt.Sprintf("%s-completion.zsh", ctx.Name[0])
				if err := touch(zsh); err != nil {
					return fmt.Errorf("while generating completion for %s: %v", ctx.Name[0], err)
				}

				zcomp := fmt.Sprintf("#compdef %s\n\n", ctx.Name[0])
				if err := fileAppend(zsh, zcomp); err != nil {
					return fmt.Errorf("while generating completion for %s: %v", ctx.JoinedName(), err)
				}
			}

			names := set.Commands()
			for _, name := range names {
				cmd := set[name]
				child := &Context{append(ctx.Name, name), cmd.Desc, ctx.Args, ctx.Ctx}
				if err := cmd.Func(child); err != errComp {
					return fmt.Errorf("while generating completion for %s: %v", name, err)
				}
			}

			if err := set.Comp(ctx); err != nil {
				return fmt.Errorf("while generating completion for %s: %v", ctx.JoinedName(), err)
			}

			if len(ctx.Name) == 1 {
				bash := fmt.Sprintf("%s-completion.bash", ctx.Name[0])
				bcomp := fmt.Sprintf("complete -F _%[1]s %[1]s", ctx.Name[0])
				if err := fileAppend(bash, bcomp); err != nil {
					return fmt.Errorf("while generating completion for %s: %v", ctx.JoinedName(), err)
				}
			}

			return nil
		}

		cmd, ok := set[head]
		if !ok {
			return fmt.Errorf("unknown command name `%s`", head)
		}

		return cmd.Func(&Context{append(ctx.Name, head), cmd.Desc, tail, ctx.Ctx})
	}
}

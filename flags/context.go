package flags

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-wrap/wrap"
)

// Context is an implementation of context.Context which contains extra data
// needed to parse command line arguments and print command details.
type Context struct {
	Name []string
	Desc string
	Args []string
	Ctx  context.Context
}

// JoinedName returns the name joined by a whitespace.
func (ctx Context) JoinedName() string {
	return strings.Join(ctx.Name, " ")
}

// Done implements the context.Context.Done method.
func (ctx Context) Done() <-chan struct{} {
	return ctx.Ctx.Done()
}

// Err implements the context.Context.Err method.
func (ctx Context) Err() error {
	return ctx.Ctx.Err()
}

// Value implements the context.Context.Value method.
func (ctx Context) Value(key interface{}) interface{} {
	return ctx.Ctx.Value(key)
}

// Parse will parse the Context arguments based on the given positional and
// optional argument definition objects.
func (ctx *Context) Parse(pos *Positional, opt *Optional) error {
	args, err := Parse(pos, opt, ctx.Args)
	if err != nil {
		b := strings.Builder{}
		name := ctx.JoinedName()
		usage := wrap.Space(Usage(pos, opt), 72-len(name))

		switch err {
		case errHelp:
			b.WriteString(fmt.Sprintf("%s: %s\n\n", name, ctx.Desc))
			b.WriteString(fmt.Sprintf("usage: %s %s\n", name, usage))
			b.WriteString(Help(pos, opt))

		case errRonn:
			if err := Ronn(ctx, pos, opt); err != nil {
				return ctx.Raise(err)
			}
			return errRonn

		case errComp:
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

			if err := Comp(ctx, pos, opt); err != nil {
				return ctx.Raise(err)
			}

			if len(ctx.Name) == 1 {
				bash := fmt.Sprintf("%s-completion.bash", ctx.Name[0])
				bcomp := fmt.Sprintf("complete -F _%[1]s %[1]s", ctx.Name[0])
				if err := fileAppend(bash, bcomp); err != nil {
					return fmt.Errorf("while generating completion for %s: %v", ctx.JoinedName(), err)
				}
			}

			return errComp

		default:
			b.WriteString(fmt.Sprintf("%v\n\nusage: %s %s", err, name, usage))
		}

		return errors.New(b.String())
	}
	ctx.Args = args
	return nil
}

// Raise creates an error with the current context.
func (ctx Context) Raise(err error) error {
	if err != nil {
		return fmt.Errorf("%s: %v", ctx.JoinedName(), err)
	}
	return nil
}

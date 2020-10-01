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
	Name string
	Desc string
	Args []string
	Ctx  context.Context
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
		name := ctx.Name
		usage := wrap.Space(Usage(pos, opt), 72-len(name))

		switch err {
		case errHelp:
			b.WriteString(fmt.Sprintf("%s: %s\n\n", name, ctx.Desc))
			b.WriteString(fmt.Sprintf("usage: %s %s\n", name, usage))
			b.WriteString(Help(pos, opt))

		case errRonn:
			return Ronn(ctx, pos, opt)

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
		return fmt.Errorf("%s: %v", ctx.Name, err)
	}
	return nil
}

package flags

import (
	"context"
	"fmt"

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
func (ctx Context) Parse(pos *Positional, opt *Optional) error {
	args, err := Parse(pos, opt, ctx.Args)
	if err != nil {
		name := ctx.Name
		usage := wrap.Space(Usage(pos, opt), 72-len(name))
		if err == errHelp {
			return fmt.Errorf("usage: %s %s\n%s", ctx.Name, usage, Help(pos, opt))
		}
		return fmt.Errorf("%v\nusage: %s %s", err, ctx.Name, usage)
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

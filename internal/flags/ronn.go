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

// Ronn creates a manpage markdown template for ronn.
func Ronn(ctx *Context, pos *Positional, opt *Optional) error {
	usage := wrap.Space(Usage(pos, opt), 72-len(ctx.JoinedName()))
	name := strings.Join(ctx.Name, "-")
	filename := fmt.Sprintf("%s.1.ronn", name)

	f, err := os.Create(filename)
	if err != nil {
		return ctx.Raise(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	parts := []string{
		fmt.Sprintf("# %s(1) -- %s", name, ctx.Desc),
		"## SYNOPSIS",
		name + " " + usage,
		"## DESCRIPTION",
		sentencify(ctx.Desc),
		"## OPTIONS",
	}

	options := []string{}

	for _, name := range pos.Order {
		arg := pos.Args[name]
		usage := wrap.Space(sentencify(arg.Usage), 76)
		usage = strings.ReplaceAll(usage, "\n", "    \n")
		switch pos.Args[name].Value.(type) {
		case *StringSliceValue:
			options = append(options, fmt.Sprintf("  * `<%s>...`:\n    %s", name, usage))
		default:
			options = append(options, fmt.Sprintf("  * `<%s>`:\n    %s", name, usage))
		}
	}

	optNames := []optionalName{}
	for long := range opt.Args {
		optName := optionalName{0, long}
		for short := range opt.Alias {
			if opt.Alias[short] == long {
				optName.Short = short
			}
		}
		optNames = append(optNames, optName)
	}

	sort.Sort(byShort(optNames))

	for _, optName := range optNames {
		short, long := optName.Short, optName.Long
		arg := opt.Args[long]
		usage := wrap.Space(sentencify(arg.Usage), 76)
		usage = strings.ReplaceAll(usage, "\n", "    \n")
		var flag string

		switch arg.Value.(type) {
		case *BoolValue:
			switch short {
			case 0:
				flag = fmt.Sprintf("  * `--%s`:\n", long)
			default:
				flag = fmt.Sprintf("  * `-%c`, `--%s`:\n", short, long)
			}
		default:
			switch short {
			case 0:
				flag = fmt.Sprintf("  * `--%[1]s=<%[1]s>`:\n", long)
			default:
				flag = fmt.Sprintf("  * `-%[1]c <%[2]s>`, `--%[2]s=<%[2]s>`:\n", short, long)
			}
		}

		options = append(options, flag+"    "+usage)
	}

	parts = append(parts, options...)
	parts = append(parts, []string{
		"## BUGS",
		fmt.Sprintf("**%s** currently has no known bugs.", name),
		"## AUTHORS",
		fmt.Sprintf("**%s** is written and maintained by @AUTHOR@.", name),
		"## SEE ALSO",
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

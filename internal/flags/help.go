package flags

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-wrap/wrap"
)

func formatHelp(name, desc string) string {
	desc = wrap.Space(desc, 55)
	desc = strings.Replace(desc, "\n", "\n                        ", -1)
	if len(name) < 22 {
		return "  " + name + strings.Repeat(" ", 22-len(name)) + desc
	}
	return "  " + name + "\n                        " + desc
}

// Usage creates a usage string for the given argument definitions.
func Usage(pos *Positional, opt *Optional) string {
	b := strings.Builder{}
	b.WriteString("[--version] [-h | --help]")
	if opt != nil && len(opt.Args) > 0 {
		b.WriteString(" [<args>]")
	}
	if pos != nil {
		for _, name := range pos.Order {
			switch pos.Args[name].Value.(type) {
			case *StringSliceValue:
				b.WriteString(fmt.Sprintf(" <%s>...", name))
			default:
				b.WriteString(fmt.Sprintf(" <%s>", name))
			}
		}
	}
	return b.String()
}

// Help creaes a help string for the given argument definitions.
func Help(pos *Positional, opt *Optional) string {
	parts := []string{}
	if pos != nil {
		parts = append(parts, "\npositional arguments:")
		for _, name := range pos.Order {
			usage := pos.Args[name].Usage
			switch pos.Args[name].Value.(type) {
			case *StringSliceValue:
				name = fmt.Sprintf("<%s>...", name)
			default:
				name = fmt.Sprintf("<%s>", name)
			}
			parts = append(parts, formatHelp(name, usage))
		}
	}

	if opt != nil {
		parts = append(parts, "\noptional arguments:")

		names := []optionalName{}
		for long := range opt.Args {
			name := optionalName{0, long}
			for short := range opt.Alias {
				if opt.Alias[short] == long {
					name.Short = short
				}
			}
			names = append(names, name)
		}

		sort.Sort(byShort(names))

		for _, name := range names {
			long, short := name.Long, name.Short
			arg := opt.Args[long]
			usage := arg.Usage
			var flag string

			switch arg.Value.(type) {
			case *BoolValue:
				switch short {
				case 0:
					flag = "--" + long
				default:
					flag = fmt.Sprintf("-%c, --%s", short, long)
				}
			case SliceValue:
				switch short {
				case 0:
					flag = fmt.Sprintf("--%[1]s=<%[1]s> [--%[1]s=<%[1]s> ...]", long)
				default:
					flag = fmt.Sprintf("-%[1]c <%[2]s> [-%[1]c <%[2]s> ...]", short, long)
				}
			default:
				switch short {
				case 0:
					flag = fmt.Sprintf("--%[1]s <%[1]s>", long)
				default:
					flag = fmt.Sprintf("-%c <%[2]s>, --%[2]s=<%[2]s>", short, long)
				}
			}
			parts = append(parts, formatHelp(flag, usage))
		}
	}
	return strings.Join(parts, "\n")
}

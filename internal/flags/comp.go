package flags

import (
	"fmt"
	"sort"
	"strings"
)

var compSetBashFormat = strings.Join([]string{
	"_%s()",
	"{",
	"    cmds=\"-h --help --version %s\"",
	"    local i=0 cmd",
	"",
	"    while [[ \"$i\" -lt \"$COMP_CWORD\" ]]",
	"    do",
	"        local s=\"${COMP_WORDS[$i]}\"",
	"        case \"$s\" in",
	"            gts)",
	"                (( i++ ))",
	"                break",
	"                ;;",
	"        esac",
	"        (( i++ ))",
	"    done",
	"",
	"    while [[ \"$i\" -lt \"$COMP_CWORD\" ]]",
	"    do",
	"        local s=\"${COMP_WORDS[$i]}\"",
	"        case \"$s\" in",
	"            -*) ;;",
	"            *)",
	"                cmd=\"$s\"",
	"                break",
	"                ;;",
	"        esac",
	"        (( i++ ))",
	"    done",
	"",
	"    if [[ \"$i\" -eq \"$COMP_CWORD\" ]]",
	"    then",
	"        local cur=\"${COMP_WORDS[$COMP_CWORD]}\"",
	"        COMPREPLY=()",
	"        while IFS='' read -r line",
	"        do",
	"            COMPREPLY+=(\"$line\")",
	"        done < <(compgen -W \"$cmds\" -- \"$cur\")",
	"        return",
	"    fi",
	"",
	"    case \"$cmd\" in",
	"        %s",
	"        *) ;;",
	"    esac",
	"}",
	"",
	"",
}, "\n")

var compFuncBashFormat = strings.Join([]string{
	"_%s()",
	"{",
	"    opts=\"-h --help --version %s\"",
	"    local cur=\"${COMP_WORDS[$COMP_CWORD]}\"",
	"    case \"$cur\" in",
	"        -*)",
	"            COMPREPLY=()",
	"            while IFS='' read -r line",
	"            do",
	"                COMPREPLY+=(\"$line\")",
	"            done < <(compgen -W \"$opts\" -- \"$cur\")",
	"            ;;",
	"        *)",
	"            COMPREPLY=()",
	"            while IFS='' read -r line",
	"            do ",
	"                COMPREPLY+=(\"$line\")",
	"            done < <(compgen -f -- \"$cur\")",
	"            ;;",
	"    esac",
	"}",
	"",
	"",
}, "\n")

var compSetZshFormat = strings.Join([]string{
	"function _%s {",
	"    local line",
	"",
	"    function _commands {",
	"        local -a commands",
	"        commands=(",
	"            %s",
	"        )",
	"        _describe 'command' commands",
	"    }",
	"",
	"    _arguments -C \\",
	"        \"-h[show help]\" \\",
	"        \"--help[show help]\" \\",
	"        \"--version[print the version number]\" \\",
	"        \"1: :_commands\" \\",
	"        \"*::arg:->args\"",
	"",
	"    case $line[1] in",
	"        %s",
	"        *) ;;",
	"    esac",
	"}",
	"",
	"",
}, "\n")

var compFuncZshFormat = strings.Join([]string{
	"function _%s {",
	"    _arguments \\",
	"        \"-h[show help]\" \\",
	"        \"--help[show help]\" \\",
	"        \"--version[print the version number]\" \\",
	"        %s \\",
	"        \"*::files:_files\"",
	"}",
	"",
	"",
}, "\n")

func compBash(ctx *Context, pos *Positional, opt *Optional) error {
	funcName := strings.Join(ctx.Name, "_")

	optNames := []optionalName{}
	for long := range opt.Args {
		name := optionalName{0, long}
		for short := range opt.Alias {
			if opt.Alias[short] == long {
				name.Short = short
			}
		}
		optNames = append(optNames, name)
	}

	sort.Sort(byShort(optNames))

	optFlags := []string{}
	for _, optName := range optNames {
		short, long := optName.Short, optName.Long
		if short != 0 {
			optFlags = append(optFlags, fmt.Sprintf("-%c", short))
		}
		optFlags = append(optFlags, fmt.Sprintf("--%s", long))
	}

	opts := strings.Join(optFlags, " ")

	comp := fmt.Sprintf(compFuncBashFormat, funcName, opts)

	filename := fmt.Sprintf("%s-completion.bash", ctx.Name[0])
	return fileAppend(filename, comp)
}

func compZsh(ctx *Context, pos *Positional, opt *Optional) error {
	funcName := strings.Join(ctx.Name, "_")

	optNames := []optionalName{}
	for long := range opt.Args {
		name := optionalName{0, long}
		for short := range opt.Alias {
			if opt.Alias[short] == long {
				name.Short = short
			}
		}
		optNames = append(optNames, name)
	}

	sort.Sort(byShort(optNames))

	optFlags := []string{}
	for _, optName := range optNames {
		short, long := optName.Short, optName.Long
		usage := opt.Args[long].Usage
		if short != 0 {
			optFlags = append(optFlags, fmt.Sprintf("\"-%c[%s]\"", short, usage))
		}
		optFlags = append(optFlags, fmt.Sprintf("\"--%s[%s]\"", long, usage))
	}

	opts := strings.Join(optFlags, " \\\n        ")

	comp := fmt.Sprintf(compFuncZshFormat, funcName, opts)

	filename := fmt.Sprintf("%s-completion.zsh", ctx.Name[0])
	return fileAppend(filename, comp)
}

// Comp creates a bash completion script.
func Comp(ctx *Context, pos *Positional, opt *Optional) error {
	if err := compBash(ctx, pos, opt); err != nil {
		return ctx.Raise(err)
	}

	if err := compZsh(ctx, pos, opt); err != nil {
		return ctx.Raise(err)
	}

	return nil
}

package flags

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

func spaces(n int) string { return strings.Repeat(" ", n) }

var shortKeys = []byte("#%123456789AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz")

func isLong(s string) bool {
	return strings.HasPrefix(s, "--") && s != "--"
}

func isShort(s string) bool {
	return strings.HasPrefix(s, "-") && s != "-" && !isLong(s)
}

func isName(s string) bool {
	return (isShort(s) || isLong(s))
}

type CommandFunc func(*Command, []string) error

type Subcommand struct {
	Func CommandFunc
	Desc string
}

type commandInfo struct {
	Name string
	Desc string
}

type commandByName []commandInfo

func (c commandByName) Len() int {
	return len(c)
}

func (c commandByName) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c commandByName) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}

type Command struct {
	Prog         string
	Desc         string
	Values       map[string]Value
	Usages       map[string]string
	Aliases      map[byte]string
	Extras       []string
	Commands     map[string]Subcommand
	InfileValue  *FileValue
	InfileUsage  string
	OutfileValue *FileValue
	OutfileUsage string
	Positional   *Positional
}

func NewCommand(prog, desc string) *Command {
	return &Command{
		Prog:         prog,
		Desc:         desc,
		Values:       make(map[string]Value),
		Usages:       make(map[string]string),
		Aliases:      make(map[byte]string),
		Extras:       make([]string, 0),
		Commands:     make(map[string]Subcommand),
		InfileValue:  nil,
		OutfileValue: nil,
		Positional:   NewPositional(),
	}
}

func (command *Command) Command(name, desc string, f CommandFunc) {
	command.Commands[name] = Subcommand{f, desc}
}

func (command Command) hasValue(name string) bool {
	_, ok := command.Values[name]
	return ok
}

func (command Command) hasAlias(short byte) bool {
	_, ok := command.Aliases[short]
	return ok
}

func (command *Command) addAlias(short byte, long string) {
	if short == 0 {
		return
	}

	if command.hasAlias(short) {
		panic(fmt.Errorf("argument with alias `%c` already exists", short))
	}

	command.Aliases[short] = long
}

func (command Command) hasPositional() bool {
	if command.Positional.Len() > 0 {
		return true
	}
	if command.InfileValue != nil {
		return true
	}
	if command.OutfileValue != nil {
		return true
	}
	return false
}

func (command *Command) Register(short byte, long string, value Value, usage string) {
	if command.hasValue(long) {
		panic(fmt.Errorf("argument with name `%s` already exists", long))
	}

	command.addAlias(short, long)
	command.Values[long] = value
	command.Usages[long] = usage
}

func (command *Command) Switch(short byte, long string, usage string) *bool {
	value := NewBoolValue(false)
	command.Register(short, long, value, usage)
	return (*bool)(value)
}

func (command *Command) Int(short byte, long string, init int, usage string) *int {
	value := NewIntValue(init)
	command.Register(short, long, value, usage)
	return (*int)(value)
}

func (command *Command) String(short byte, long string, init string, usage string) *string {
	value := NewStringValue(init)
	command.Register(short, long, value, usage)
	return (*string)(value)
}

func (command *Command) Choice(short byte, long string, usage string, choices ...string) *int {
	value := NewChoiceValue(choices, 0)
	usage = fmt.Sprintf("%s (available values: [%s])", usage, strings.Join(choices, ", "))
	command.Register(short, long, value, usage)
	return value.Chosen
}

func (command *Command) Strings(short byte, long string, usage string) *[]string {
	value := NewStringsValue(make([]string, 0))
	command.Register(short, long, value, usage)
	return (*[]string)(value)
}

func (command *Command) Open(short byte, long string, usage string) *os.File {
	return command.File(short, long, os.O_RDONLY, 0, usage)
}

func (command *Command) Create(short byte, long string, usage string) *os.File {
	return command.File(short, long, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666, usage)
}

func (command *Command) File(short byte, long string, flag int, perm os.FileMode, usage string) *os.File {
	value := NewFileValue(flag, perm)
	command.Register(short, long, value, usage)
	return value.File
}

func (command *Command) Infile(usage string) *os.File {
	if command.InfileValue != nil {
		panic("only one positional input file is allowed")
	}
	command.InfileUsage = usage
	if IsTerminal(os.Stdin.Fd()) {
		command.InfileValue = NewFileValue(os.O_RDONLY, 0)
		return (*os.File)(command.InfileValue.File)
	}
	command.InfileValue = NewFileValueWithFile(os.Stdin)
	return command.InfileValue.File
}

func (command *Command) Outfile(usage string) *os.File {
	if command.OutfileValue != nil {
		panic("only one positional output file is allowed")
	}
	command.OutfileUsage = usage
	if IsTerminal(os.Stdout.Fd()) {
		command.OutfileValue = NewFileValue(os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		return (*os.File)(command.OutfileValue.File)
	}
	command.OutfileValue = NewFileValueWithFile(os.Stdout)
	return command.OutfileValue.File
}

func splitEqual(s string) (string, string) {
	for i, c := range s {
		if c == '=' {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

func (command Command) handleValue(value Value, args []string) ([]string, error) {
	switch p := value.(type) {
	case *BoolValue:
		*p = BoolValue(true)
		return args, nil
	case SliceValue:
		count := command.Positional.Len() - len(command.Extras)
		if command.infileMissing() {
			count++
		}
		if command.outfileMissing() {
			count++
		}
		n := len(args) - count
		if n < 0 {
			return nil, fmt.Errorf("not enough arguments")
		}
		tail, err := processSlice(p, args[:n])
		if err != nil {
			return nil, err
		}
		return append(tail, args[n:]...), nil
	default:
		return processValue(p, args)
	}
}

func (command Command) handleLong(long string, args []string) ([]string, error) {
	if long == "help" {
		return nil, HelpError(command.Help())
	}

	name, arg := splitEqual(long)
	if len(arg) > 0 {
		value, ok := command.Values[name]
		if !ok {
			return nil, fmt.Errorf("unknown argument name `--%s`", name)
		}
		if err := value.Set(arg); err != nil {
			return nil, fmt.Errorf("%s for argument `--%s`", err, name)
		}
		return args, nil
	}

	value, ok := command.Values[name]
	if !ok {
		return nil, fmt.Errorf("unknown argument name `--%s`", name)
	}

	args, err := command.handleValue(value, args)
	if err != nil {
		return nil, fmt.Errorf("%s for argument `--%s`", err, name)
	}

	return args, nil
}

func (command Command) getShortValue(short byte) (Value, error) {
	name, ok := command.Aliases[short]
	if !ok {
		return nil, fmt.Errorf("unknown argument alias `%c`", short)
	}

	value, ok := command.Values[name]
	if !ok {
		// This shouldn't happen under normal circumstances.
		panic(fmt.Errorf("value for `%s` with alias `%c` not found", name, short))
	}

	return value, nil
}

func (command Command) handleShortHead(short byte) error {
	value, err := command.getShortValue(short)
	if err != nil {
		return err
	}

	p, ok := value.(*BoolValue)
	if !ok {
		return fmt.Errorf("missing value for argument alias `%c`", short)
	}

	*p = BoolValue(true)
	return nil
}

func (command Command) handleShortTail(short byte, args []string) ([]string, error) {
	value, err := command.getShortValue(short)
	if err != nil {
		return nil, err
	}

	args, err = command.handleValue(value, args)
	if err != nil {
		return nil, fmt.Errorf("%s for argument alias `%c`", err, short)
	}

	return args, nil
}

func (command Command) handleShort(group string, args []string) ([]string, error) {
	if strings.ContainsRune(group, 'h') {
		return nil, HelpError(command.Help())
	}

	p := []byte(group)
	head, tail := p[:len(p)-1], p[len(p)-1]
	for _, short := range head {
		if err := command.handleShortHead(short); err != nil {
			return nil, err
		}
	}
	return command.handleShortTail(tail, args)
}

func (command Command) infileMissing() bool {
	return command.InfileValue != nil && command.InfileValue.Empty
}

func (command Command) outfileMissing() bool {
	return command.OutfileValue != nil && command.OutfileValue.Empty
}

func (command *Command) setStdin() {
	if command.InfileValue != nil {
		*(command.InfileValue.File) = *os.Stdin
	}
}

func (command *Command) setStdout() {
	if command.OutfileValue != nil {
		*(command.OutfileValue.File) = *os.Stdout
	}
}

func (command *Command) handleOne(args []string) ([]string, error) {
	head, tail := shift(args)

	if isLong(head) {
		long := strings.TrimPrefix(head, "--")
		return command.handleLong(long, tail)
	}

	if isShort(head) {
		group := strings.TrimPrefix(head, "-")
		return command.handleShort(group, tail)
	}

	if sub, ok := command.Commands[head]; ok {
		f, desc := sub.Func, sub.Desc
		name := fmt.Sprintf("%s %s", command.Prog, head)
		command.Commands = make(map[string]Subcommand)
		return nil, f(NewCommand(name, desc), tail)
	}

	command.Extras = append(command.Extras, head)
	return tail, nil
}

func (command *Command) handleLastArg(arg string) error {
	if arg == "-" {
		if command.infileMissing() {
			command.setStdin()
			command.setStdout()
			return nil
		}

		if command.outfileMissing() {
			command.setStdout()
			return nil
		}

		return nil
	}

	if command.infileMissing() {
		command.setStdout()
		return command.InfileValue.Set(arg)
	}

	if command.outfileMissing() {
		return command.OutfileValue.Set(arg)
	}

	return errors.New("too many arguments")
}

func (command *Command) handleLastArgs(first, second string) error {
	if first == "-" {
		command.setStdin()
	} else {
		if err := command.InfileValue.Set(first); err != nil {
			return err
		}
	}

	if second == "-" {
		command.setStdout()
	} else {
		if err := command.OutfileValue.Set(second); err != nil {
			return err
		}
	}

	return nil
}

func (command *Command) handleArgs(args []string) (err error) {
	for len(args) != 0 {
		if args, err = command.handleOne(args); err != nil {
			return err
		}
	}

	if len(command.Commands) != 0 {
		return fmt.Errorf("command not specified")
	}

	extras, err := command.Positional.Handle(command.Extras)
	if err != nil {
		return err
	}

	switch len(extras) {
	case 0:
		if command.infileMissing() {
			return fmt.Errorf("missing input file")
		}
		if command.outfileMissing() {
			command.setStdout()
			return nil
		}
	case 1:
		return command.handleLastArg(extras[0])
	case 2:
		return command.handleLastArgs(extras[0], extras[1])
	default:
		return fmt.Errorf("too many arguments")
	}

	return nil
}

func (command *Command) Run(args []string, fs ...func() error) error {
	if err := command.handleArgs(args); err != nil {
		switch err.(type) {
		case HelpError, UsageError:
			return err
		default:
			return UsageError(fmt.Sprintf("%s\n%s", err, command.Usage()))
		}
	}

	for _, f := range fs {
		if err := f(); err != nil {
			switch err.(type) {
			case HelpError, UsageError:
				return err
			default:
				return UsageError(fmt.Sprintf("%s\n%s", err, command.Usage()))
			}
		}
	}

	return nil
}

func (command Command) findAlias(long string) byte {
	for k, v := range command.Aliases {
		if v == long {
			return k
		}
	}
	return 0
}

func (command Command) listNames() []string {
	names := make([]string, 0)

	for k := range command.Values {
		if command.findAlias(k) == 0 {
			names = append(names, k)
		}
	}

	sort.Strings(names)

	ret := make([]string, 0)

	i := 0
	for _, short := range shortKeys {
		if long, ok := command.Aliases[short]; ok {
			ret = append(ret, long)
		} else {
			for i < len(names) && names[i][0] == short {
				ret = append(ret, names[i])
				i++
			}
		}
	}

	return ret
}

func appendWrap(s, t string, d int) string {
	if strings.ContainsRune(s, '\n') {
		i := strings.LastIndexByte(s, '\n') + 1
		return s[:i] + appendWrap(s[i:], t, d)
	}

	if len(s)+len(t) < 78 {
		return s + " " + t
	}

	return s + "\n" + spaces(d) + t
}

func wrap(s string, d int) string {
	if strings.ContainsRune(s, '\n') {
		lines := strings.Split(s, "\n")
		ret := make([]string, len(lines))
		for i, line := range lines {
			ret[i] = wrap(line, d)
		}
		return strings.Join(ret, "\n")
	}

	if len(s) < 80 {
		return s
	}

	if len(s) < 80 {
		return s
	}

	i := 79
	for i >= 0 && s[i] != ' ' {
		i--
	}

	if i == 0 {
		i = 79
		for i < len(s) && s[i] != ' ' {
			i++
		}
	}

	if i == len(s) {
		return s
	}

	t := s[:i]
	r := spaces(d-1) + s[i:]

	return t + "\n" + wrap(r, d)

}

func (command Command) Usage() string {
	usage := fmt.Sprintf("usage: %s", command.Prog)
	depth := len(usage) + 1
	usage = appendWrap(usage, "[-h | --help]", depth)

	for _, short := range shortKeys {
		if long, ok := command.Aliases[short]; ok {
			switch command.Values[long].(type) {
			case *BoolValue:
				tmp := fmt.Sprintf("[-%c | --%s]", short, long)
				usage = appendWrap(usage, tmp, depth)
			case SliceValue:
				tmp := fmt.Sprintf("[-%c <%s> [<%s> ...]]", short, long, long)
				usage = appendWrap(usage, tmp, depth)
			default:
				tmp := fmt.Sprintf("[-%c <%s>]", short, long)
				usage = appendWrap(usage, tmp, depth)
			}
		}
	}

	for _, name := range command.Positional.Order {
		usage = appendWrap(usage, fmt.Sprintf("<%s>", name), depth)
	}

	if command.InfileValue != nil {
		usage = appendWrap(usage, "<infile>", depth)
	}

	if command.OutfileValue != nil {
		usage = appendWrap(usage, "<outfile>", depth)
	}

	if len(command.Commands) > 0 {
		usage = appendWrap(usage, "<command> [<args>]", depth)
	}

	return usage
}

func (command Command) switchSyntax(short byte, long string) string {
	if short == 0 {
		return fmt.Sprintf("  --%s", long)
	}
	return fmt.Sprintf("  -%c, --%s", short, long)
}

func (command Command) sliceSyntax(short byte, long string) string {
	if short == 0 {
		return strings.Join([]string{
			fmt.Sprintf("  --%s <%s> [<%s> ...]", long, long, long),
			fmt.Sprintf("  --%s <%s> [--%s <%s> ...]", long, long, long, long),
			fmt.Sprintf("  --%s=<%s> [--%s=<%s> ...]", long, long, long, long),
		}, ",\n")
	}
	return strings.Join([]string{
		fmt.Sprintf("  -%c <%s> [<%s> ...]", short, long, long),
		fmt.Sprintf("  -%c <%s> [-%c <%s> ...]", short, long, short, long),
		fmt.Sprintf("  --%s <%s> [--%s <%s> ...]", long, long, long, long),
		fmt.Sprintf("  --%s=<%s> [--%s=<%s> ...]", long, long, long, long),
	}, ",\n")
}

func (command Command) valueSyntax(short byte, long string) string {
	if short == 0 {
		return fmt.Sprintf("  --%s <%s>, --%s=<%s>", long, long, long, long)
	}
	return fmt.Sprintf("  -%c <%s>, --%s <%s>, --%s=<%s>", short, long, long, long, long, long)
}

func (command Command) syntax(long string) string {
	short := command.findAlias(long)
	switch command.Values[long].(type) {
	case *BoolValue:
		return command.switchSyntax(short, long)
	case SliceValue:
		return "\n" + command.sliceSyntax(short, long)
	default:
		return command.valueSyntax(short, long)
	}
}

func (command Command) listCommands() []commandInfo {
	commands := make([]commandInfo, 0)
	for key, value := range command.Commands {
		commands = append(commands, commandInfo{key, value.Desc})
	}
	sort.Sort(commandByName(commands))
	return commands
}

func (command Command) Help() string {
	parts := []string{
		command.Usage(),
		"",
		"description:",
		wrap(fmt.Sprintf("  %s", command.Desc), 2),
	}

	commands := command.listCommands()
	if len(commands) > 0 {
		parts = append(parts, "", "available commands:")
		for _, info := range commands {
			part := fmt.Sprintf("  %s\n                        %s", info.Name, info.Desc)
			if len(info.Name) < 22 {
				padding := spaces(22 - len(info.Name))
				part = fmt.Sprintf("  %s%s%s", info.Name, padding, info.Desc)
			}
			parts = append(parts, wrap(part, 24))
		}
	}

	if command.hasPositional() {
		parts = append(parts, "", "positional arguments:")

		for _, name := range command.Positional.Order {
			usage := command.Positional.Usages[name]
			padding := spaces(20 - len(name))
			part := fmt.Sprintf("  <%s>%s%s", name, padding, usage)
			parts = append(parts, wrap(part, 24))
		}

		if command.InfileValue != nil {
			part := fmt.Sprintf("  <infile>              %s", command.InfileUsage)
			parts = append(parts, wrap(part, 24))
		}

		if command.OutfileValue != nil {
			part := fmt.Sprintf("  <outfile>             %s", command.OutfileUsage)
			parts = append(parts, wrap(part, 24))
		}
	}

	parts = append(parts, "", "optional arguments:",
		"  -h, --help            show this help message and exit")

	for _, name := range command.listNames() {
		syntax := command.syntax(name)
		usage := command.Usages[name]
		value := command.Values[name].Format()
		padding := "\n                        "
		if len(syntax) < 23 {
			padding = spaces(24 - len(syntax))
		}
		part := wrap(syntax+padding+usage, 24)
		full := appendWrap(part, fmt.Sprintf("(value: %s)", value), 24)
		parts = append(parts, full)
	}

	help := strings.Join(parts, "\n")
	return help
}

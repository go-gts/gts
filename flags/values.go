package flags

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Value interface {
	Set(value string) error
	Format() string
}

type SliceValue interface {
	Value
	Len() int
}

type BoolValue bool

func NewBoolValue(init bool) *BoolValue {
	p := new(bool)
	*p = init
	return (*BoolValue)(p)
}

func (p *BoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("value `%s` cannot be interpreted as bool", s)
	}
	*p = BoolValue(v)
	return nil
}

func (p BoolValue) Format() string {
	return strconv.FormatBool(bool(p))
}

type IntValue int

func NewIntValue(init int) *IntValue {
	p := new(int)
	*p = init
	return (*IntValue)(p)
}

func (p *IntValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("value `%s` cannot be interpreted as int", s)
	}
	*p = IntValue(v)
	return nil
}

func (p IntValue) Format() string {
	return strconv.Itoa(int(p))
}

type StringValue string

func NewStringValue(init string) *StringValue {
	p := new(string)
	*p = init
	return (*StringValue)(p)
}

func (p *StringValue) Set(s string) error {
	*p = StringValue(s)
	return nil
}

func (p StringValue) Format() string {
	return string(p)
}

type StringsValue []string

func NewStringsValue(init []string) *StringsValue {
	p := new([]string)
	*p = init
	return (*StringsValue)(p)
}

func (p *StringsValue) Set(s string) error {
	ss := []string(*p)
	ss = append(ss, s)
	*p = StringsValue(ss)
	return nil
}

func (p StringsValue) Format() string {
	return strings.Join([]string(p), ", ")
}

func (p StringsValue) Len() int {
	return len(p)
}

// os.File contains a pointer to an OS specific file struct.
type FileValue struct {
	File  *os.File
	Flag  int
	Perm  os.FileMode
	Empty bool
}

func NewFileValue(flag int, perm os.FileMode) *FileValue {
	p := &FileValue{new(os.File), flag, perm, true}
	return (*FileValue)(p)
}

func (p *FileValue) Set(s string) error {
	f, err := os.OpenFile(s, p.Flag, p.Perm)
	if err != nil {
		return err
	}
	*(p.File) = *f
	p.Empty = false
	return nil
}

func (p *FileValue) Format() string {
	return p.File.Name()
}

type Values map[string]Value

func (v Values) Get(name string) Value {
	return v[name]
}

func (v Values) Bool(name string) bool {
	p := v.Get(name).(*BoolValue)
	return bool(*p)
}

func (v Values) Int(name string) int {
	p := v.Get(name).(*IntValue)
	return int(*p)
}

func (v Values) String(name string) string {
	p := v.Get(name).(*StringValue)
	return string(*p)
}

func shift(ss []string) (string, []string) {
	if len(ss) == 0 {
		panic("shift on empty list")
	}
	return ss[0], ss[1:]
}

func unshift(ss []string, s string) []string {
	r := make([]string, len(ss)+1)
	copy(r[1:], ss)
	r[0] = s
	return r
}

func pop(ss []string) ([]string, string) {
	if len(ss) == 0 {
		panic("pop on empty list")
	}
	return ss[:len(ss)-1], ss[len(ss)-1]
}

func processValue(value Value, args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("missing value")
	}

	head, args := shift(args)
	if err := value.Set(head); err != nil {
		return nil, err
	}

	return args, nil
}

func processSlice(value SliceValue, args []string) ([]string, error) {
	if len(args) == 0 {
		if value.Len() == 0 {
			return nil, fmt.Errorf("missing value")
		}
		return args, nil
	}

	head, args := shift(args)
	if isLong(head) || isShort(head) {
		if value.Len() == 0 {
			return nil, fmt.Errorf("missing value")
		}
		return unshift(args, head), nil
	}

	if err := value.Set(head); err != nil {
		if value.Len() == 0 {
			return nil, err
		}
		return unshift(args, head), nil
	}

	return processSlice(value, args)
}

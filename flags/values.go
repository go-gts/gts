package flags

import (
	"strconv"
	"strings"
)

type Value interface {
	Set(value string) error
	Format() string
}

type BoolValue bool

func NewBoolValue(value bool) *BoolValue {
	p := new(bool)
	*p = value
	return (*BoolValue)(p)
}

func (p *BoolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*p = BoolValue(v)
	return nil
}

func (p BoolValue) Format() string {
	return strconv.FormatBool(bool(p))
}

type IntValue int

func NewIntValue(value int) *IntValue {
	p := new(int)
	*p = value
	return (*IntValue)(p)
}

func (p *IntValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*p = IntValue(v)
	return nil
}

func (p IntValue) Format() string {
	return strconv.Itoa(int(p))
}

type StringValue string

func NewStringValue(value string) *StringValue {
	p := new(string)
	*p = value
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

func NewStringsValue(value []string) *StringsValue {
	p := new([]string)
	*p = value
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

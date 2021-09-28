package mango

import (
	"fmt"
	"io"
)

type Section struct {
	Name string `yaml:"name"`
	Body string `yaml:"body"`
}

type Manpage struct {
	Name string    `yaml:"name"`
	Info string    `yaml:"info"`
	Chpt int       `yaml:"chpt"`
	Sect []Section `yaml:"sect"`
}

type Command struct {
	Name string    `yaml:"name"`
	Info string    `yaml:"info"`
	Desc string    `yaml:"desc"`
	Cmds []Command `yaml:"command"`
	Sect []Section `yaml:"sect"`
	Docs []Manpage `yaml:"docs"`
}

func tryAll(ff ...func() error) error {
	for _, f := range ff {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (cmd Command) Roff(w io.Writer) error {
	roff := RoffWriter{w}

	return tryAll(
		func() error { return roff.Comment("generated with ManGo/0.0.1") },
		roff.Newline,
		func() error { return roff.Section("NAME") },
		roff.Newline,
		func() error { return roff.Bold(cmd.Name) },
		roff.Newline,
		func() error { return roff.Text(fmt.Sprintf(" - %s", cmd.Info)) },
		roff.Newline,
	)
}

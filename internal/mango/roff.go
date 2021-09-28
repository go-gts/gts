package mango

import (
	"fmt"
	"io"
	"time"
)

type RoffWriter struct {
	w io.Writer
}

func (w RoffWriter) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w RoffWriter) Comment(s string) error {
	_, err := io.WriteString(w, fmt.Sprintf(".\\\"%s\n", s))
	return err
}

func (w RoffWriter) Newline() error {
	_, err := io.WriteString(w, ".\n")
	return err
}

func (w RoffWriter) Title(title string, section int, t time.Time, source, manual string) error {
	_, err := io.WriteString(w, fmt.Sprintf(
		".TH %q %q %q %q %q\n",
		title,
		section,
		t.Format("Jan 2006"),
		source,
		manual,
	))
	return err
}

func (w RoffWriter) Section(text string) error {
	_, err := io.WriteString(w, fmt.Sprintf(".SH %s\n", text))
	return err
}

func (w RoffWriter) Bold(text string) error {
	_, err := io.WriteString(w, fmt.Sprintf(".B %s\n", text))
	return err
}

func (w RoffWriter) Text(text string) error {
	_, err := io.WriteString(w, text)
	return err
}

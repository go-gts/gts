package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/go-ascii/ascii"
	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("summary", "report a brief summary of the sequence(s)", summaryFunc)
}

type pairStringInt struct {
	Key   string
	Value int
}

type byValue []pairStringInt

func (pp byValue) Len() int {
	return len(pp)
}

func (pp byValue) Less(i, j int) bool {
	if pp[i].Value > pp[j].Value {
		return true
	}
	if pp[j].Value > pp[i].Value {
		return false
	}
	return pp[i].Key < pp[j].Key
}

func (pp byValue) Swap(i, j int) {
	pp[i], pp[j] = pp[j], pp[i]
}

func summaryFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	outPath := opt.String('o', "output", "-", "output file (specifying `-` will force standard output)")
	nofeature := opt.Switch('F', "no-feature", "suppress feature summary")
	noqualifier := opt.Switch('Q', "no-qualifier", "suppress qualifier summary")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	d, err := newIODelegate(h, *seqinPath, *outPath)
	if err != nil {
		return ctx.Raise(err)
	}
	defer d.Close()

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"nofeature", *nofeature},
			{"noqualifier", noqualifier},
		})

		ok, err := d.Cache(data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	w := bufio.NewWriter(d)

	scanner := seqio.NewAutoScanner(d)
	i := 0
	for scanner.Scan() {
		seq := scanner.Value()

		b := strings.Builder{}
		switch info := seq.Info().(type) {
		case string:
			b.WriteString(fmt.Sprintln(info))
		case fmt.Stringer:
			b.WriteString(fmt.Sprintln(info.String()))
		default:
			b.WriteString(fmt.Sprintf("Sequence %d\n", i+1))
		}

		basemap := make(map[byte]int)
		bases := []pairStringInt{}
		for _, c := range bytes.ToUpper(seq.Bytes()) {
			basemap[c]++
		}
		for _, c := range ascii.Graphic {
			if n, ok := basemap[c]; ok {
				bases = append(bases, pairStringInt{fmt.Sprintf("%c", c), n})
			}
		}
		for _, c := range ascii.Control {
			if n, ok := basemap[c]; ok {
				bases = append(bases, pairStringInt{fmt.Sprintf("%X", c), n})
			}
		}

		ff := seq.Features()
		keymap := make(map[string]int)
		qfsmap := make(map[string]int)
		for _, f := range ff {
			keymap[f.Key]++
			for name, values := range f.Qualifiers {
				qfsmap[name] += len(values)
			}
		}

		keys := []pairStringInt{}
		for key, value := range keymap {
			keys = append(keys, pairStringInt{key, value})
		}
		sort.Sort(byValue(keys))

		qfs := []pairStringInt{}
		for key, value := range qfsmap {
			qfs = append(qfs, pairStringInt{key, value})
		}
		sort.Sort(byValue(qfs))

		longest := 0
		for _, p := range bases {
			if n := len(p.Key); n > longest {
				longest = n
			}
		}
		for _, p := range keys {
			if n := len(p.Key); n > longest {
				longest = n
			}
		}
		for _, p := range qfs {
			if n := len(p.Key); n > longest {
				longest = n
			}
		}

		format := fmt.Sprintf("%%%ds:\t%%s\n", longest)

		b.WriteString("Sequence Summary\n")
		b.WriteString(fmt.Sprintf(format, "Length", humanize.Comma(int64(gts.Len(seq)))))
		for _, p := range bases {
			b.WriteString(fmt.Sprintf(format, p.Key, humanize.Comma(int64(p.Value))))
		}

		if !*nofeature {
			b.WriteString("Feature Summary\n")
			b.WriteString(fmt.Sprintf(format, "Features", humanize.Comma(int64(len(ff)))))
			for _, p := range keys {
				b.WriteString(fmt.Sprintf(format, p.Key, humanize.Comma(int64(p.Value))))
			}
		}

		if !*noqualifier {
			b.WriteString("Qualifier Summary\n")
			for _, p := range qfs {
				b.WriteString(fmt.Sprintf(format, p.Key, humanize.Comma(int64(p.Value))))
			}
			b.WriteString("//\n")
		}

		if _, err := io.WriteString(w, b.String()); err != nil {
			return ctx.Raise(err)
		}

		if err := w.Flush(); err != nil {
			return ctx.Raise(err)
		}

		i++
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

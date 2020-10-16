package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("split", "split the sequence at the provided locations", splitFunc)
}

func splitFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([modifier|selector|point|range][@modifier])")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
	}

	d, err := newIODelegate(h, *seqinPath, *seqoutPath)
	if err != nil {
		return ctx.Raise(err)
	}
	defer d.Close()

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"locator", *locstr},
			{"filetype", filetype},
		})

		ok, err := d.Cache(data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	w := bufio.NewWriter(d)

	scanner := seqio.NewAutoScanner(d)
	for scanner.Scan() {
		seq := scanner.Value()
		rr := locate(seq)

		top := gts.Linear
		switch v := seq.(type) {
		case seqio.GenBank:
			top = v.Fields.Topology
		}

		switch {
		case len(rr) == 0:
			formatter := seqio.NewFormatter(seq, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
			}

		case len(rr) == 1 && top == gts.Circular:
			seq = gts.Rotate(seq, -rr.Head())
			seq = gts.WithTopology(seq, gts.Linear)
			formatter := seqio.NewFormatter(seq, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
			}

		default:
			unique := make(map[int]interface{})
			for _, r := range rr {
				head, tail := r.Head(), r.Tail()
				if tail < head {
					head = tail
				}
				unique[head] = nil
			}

			heads := make([]int, len(unique))
			i := 0
			for head := range unique {
				heads[i] = head
				i++
			}

			sort.Ints(heads)

			splits := make([]int, len(heads)+2)
			if top == gts.Circular {
				splits[0] = heads[len(heads)-1]
				splits = splits[:len(splits)-1]
			} else {
				splits[len(splits)-1] = gts.Len(seq)
			}
			for i, head := range heads {
				splits[i+1] = head
			}

			for i, tail := range splits[1:] {
				head := splits[i]
				fmt.Fprintln(os.Stderr, head, tail)
				sub := gts.Slice(seq, head, tail)
				fmt.Fprintln(os.Stderr, gts.Len(sub))
				sub = gts.WithTopology(sub, gts.Linear)
				formatter := seqio.NewFormatter(sub, filetype)
				if _, err := formatter.WriteTo(w); err != nil {
					return ctx.Raise(err)
				}
			}
		}

		if err := w.Flush(); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

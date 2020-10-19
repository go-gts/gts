package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gts/flags"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("pick", "pick sequence(s) from multiple sequences", pickFunc)
}

type picker func(i int) bool

func pickAll(pickers ...picker) picker {
	return func(i int) bool {
		for _, pick := range pickers {
			if !pick(i) {
				return false
			}
		}
		return true
	}
}

func pickAny(pickers ...picker) picker {
	return func(i int) bool {
		for _, pick := range pickers {
			if pick(i) {
				return true
			}
		}
		return false
	}
}

func pickAfter(m int) picker {
	return func(i int) bool {
		return m <= i
	}
}

func pickBefore(n int) picker {
	return func(i int) bool {
		return i <= n
	}
}

func pickBetween(m, n int) picker {
	return pickAll(pickAfter(m), pickBefore(n))
}

func pickOne(n int) picker {
	return func(i int) bool {
		return n == i
	}
}

func mustAtoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return n
}

func asPicker(list string) picker {
	blocks := strings.Split(list, ",")
	pickers := make([]picker, len(blocks))
	for i, block := range blocks {
		index := strings.IndexByte(block, '-')
		switch index {
		case -1:
			n := mustAtoi(block)
			pickers[i] = pickOne(n)
		case 0:
			n := mustAtoi(block[index+1:])
			pickers[i] = pickBefore(n)
		case len(block) - 1:
			m := mustAtoi(block[:index])
			pickers[i] = pickAfter(m)
		default:
			n := mustAtoi(block[index+1:])
			m := mustAtoi(block[:index])
			pickers[i] = pickBetween(m, n)
		}
	}
	return pickAny(pickers...)
}

func pickFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	list := pos.String("list", "list of sequences to pick (identical to the list option in cut)")

	seqinPath := new(string)
	*seqinPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	feature := opt.Switch('f', "feature", "pick features instead of sequences")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	pick := asPicker(*list)

	d, err := newIODelegate(*seqinPath, *seqoutPath)
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
			{"list", *list},
			{"feature", *feature},
			{"filetype", filetype},
		})

		ok, err := d.TryCache(h, data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	w := bufio.NewWriter(d)

	scanner := seqio.NewAutoScanner(d)
	i := 0
	for scanner.Scan() {
		seq := scanner.Value()
		i++

		if pick(i) || *feature {
			if *feature {
				ff := seq.Features()
				indices := make([]int, 0, len(ff))
				for j := range ff {
					if pick(j) {
						indices = append(indices, j)
					}
				}
				gg := make([]gts.Feature, len(indices))
				for j, k := range indices {
					gg[j] = ff[k]
				}
				seq = gts.WithFeatures(seq, ff)
			}

			formatter := seqio.NewFormatter(seq, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
			}

			if err := w.Flush(); err != nil {
				return ctx.Raise(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

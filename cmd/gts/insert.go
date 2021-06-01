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
	flags.Register("insert", "insert guest sequence(s) into the input sequence(s)", insertFunc)
}

func insertFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([specifier][@modifier])")
	guestPath := pos.String("guest", "guest sequence file (will be interpreted literally if preceded with @)")

	hostPath := new(string)
	*hostPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		hostPath = pos.String("host", "input sequence file (may be omitted if standard input is provided)")
	}

	nocache := opt.Switch(0, "no-cache", "do not use or create cache")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	embed := opt.Switch('e', "embed", "extend existing feature locations when inserting instead of splitting them")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
	}

	guests := []gts.Sequence{}
	guestBytes := []byte(*guestPath)

	h.Reset()
	switch guestBytes[0] {
	case '@':
		h.Write(guestBytes)
		guest := gts.New(nil, nil, guestBytes[1:])
		guests = append(guests, guest)

	default:
		f, err := os.Open(*guestPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file: %q: %v", *guestPath, err))
		}
		defer f.Close()

		r := attach(h, f)
		scanner := seqio.NewAutoScanner(r)
		for scanner.Scan() {
			guests = append(guests, scanner.Value())
		}
		if len(guests) == 0 {
			ctx.Raise(fmt.Errorf("guest sequence file %q does not contain a sequence", *guestPath))
		}
	}
	guestSum := h.Sum(nil)

	d, err := newIODelegate(*hostPath, *seqoutPath)
	if err != nil {
		return ctx.Raise(err)
	}
	defer d.Close()

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	insert := gts.Insert
	if *embed {
		insert = gts.Embed
	}

	if !*nocache {
		data := encodePayload([]tuple{
			{"command", strings.Join(ctx.Name, "-")},
			{"version", gts.Version.String()},
			{"locator", *locstr},
			{"guest", guestSum},
			{"embed", *embed},
			{"filetype", filetype},
		})

		ok, err := d.TryCache(h, data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	scanner := seqio.NewAutoScanner(d)
	buffer := bufio.NewWriter(d)
	writer := seqio.NewWriter(buffer, filetype)

	for scanner.Scan() {
		host := scanner.Value()

		rr := locate(host)
		indices := make([]int, len(rr))
		for i, r := range rr {
			indices[i] = r.Head()
		}
		sort.Sort(sort.Reverse(sort.IntSlice(indices)))

		for _, guest := range guests {
			out := gts.Sequence(gts.Copy(host))
			for _, index := range indices {
				out = insert(out, index, guest)
			}

			if _, err := writer.WriteSeq(out); err != nil {
				return ctx.Raise(err)
			}

			if err := buffer.Flush(); err != nil {
				return ctx.Raise(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

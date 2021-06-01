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
	flags.Register("infix", "infix input sequence(s) into the host sequence(s)", infixFunc)
}

func infixFunc(ctx *flags.Context) error {
	h := newHash()
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([modifier|selector|point|range][@modifier])")
	hostPath := pos.String("host", "host sequence")

	guestPath := new(string)
	*guestPath = "-"
	if cmd.IsTerminal(os.Stdin.Fd()) {
		guestPath = pos.String("guest", "input sequence file (may be omitted if standard input is provided)")
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

	hosts := []gts.Sequence{}

	f, err := os.Open(*hostPath)
	if err != nil {
		return ctx.Raise(fmt.Errorf("failed to open file: %q: %v", *hostPath, err))
	}
	defer f.Close()

	h.Reset()
	r := attach(h, f)
	scanner := seqio.NewAutoScanner(r)
	for scanner.Scan() {
		hosts = append(hosts, scanner.Value())
	}
	if len(hosts) == 0 {
		ctx.Raise(fmt.Errorf("host sequence file %q does not contain a sequence", *hostPath))
	}
	hostSum := h.Sum(nil)

	d, err := newIODelegate(*guestPath, *seqoutPath)
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
			{"host", hostSum},
			{"embed", *embed},
			{"filetype", filetype},
		})

		ok, err := d.TryCache(h, data)
		if ok || err != nil {
			return ctx.Raise(err)
		}
	}

	scanner = seqio.NewAutoScanner(d)
	buffer := bufio.NewWriter(d)
	writer := seqio.NewWriter(buffer, filetype)

	for scanner.Scan() {
		seq := scanner.Value()

		for _, host := range hosts {
			rr := locate(host)
			indices := make([]int, len(rr))
			for i, r := range rr {
				indices[i] = r.Head()
			}
			sort.Sort(sort.Reverse(sort.IntSlice(indices)))

			out := gts.Sequence(gts.Copy(host))
			for _, index := range indices {
				out = insert(out, index, seq)
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

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/go-flip/flip"
	"github.com/go-gts/gts"
	"github.com/go-gts/gts/cmd"
	"github.com/go-gts/gts/flags"
	"github.com/go-gts/gts/seqio"
)

func init() {
	flags.Register("length", "report the length of the sequence(s)", sequenceLength)
	flags.Register("insert", "insert a sequence into another sequence(s)", sequenceInsert)
	flags.Register("delete", "delete a region of the given sequence(s)", sequenceDelete)
	flags.Register("reverse", "reverse order of the given sequence(s)", sequenceReverse)
	flags.Register("complement", "compute the complement of the given sequence(s)", sequenceComplement)
	flags.Register("rotate", "shift the coordinates of a circular sequence", sequenceRotate)
}

func sequenceLength(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	outPath := opt.String('o', "output", "-", "output table file (specifying `-` will force standard output)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	outFile := os.Stdout
	if *outPath != "-" {
		f, err := os.Create(*outPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *outPath, err))
		}
		outFile = f
		defer outFile.Close()
	}

	w := bufio.NewWriter(outFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		_, err := io.WriteString(w, fmt.Sprintf("%d\n", gts.Len(seq)))
		if err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

func sequenceInsert(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([Select|Point|Range][@Modifier])")
	guestPath := pos.String("guest", "guest sequence file")

	var hostPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		hostPath = pos.String("host", "host sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")
	embed := opt.Switch('e', "embed", "extend existing feature locations when inserting instead of splitting them")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
	}

	hostFile := os.Stdin
	if hostPath != nil && *hostPath != "-" {
		f, err := os.Open(*hostPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *hostPath, err))
		}
		hostFile = f
		defer hostFile.Close()
	}

	guestFile, err := os.Open(*guestPath)
	if err != nil {
		return ctx.Raise(fmt.Errorf("failed to open file: %q: %v", *guestPath, err))
	}
	defer guestFile.Close()

	seqoutFile := os.Stdout
	if *seqoutPath != "-" {
		f, err := os.Create(*seqoutPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *seqoutPath, err))
		}
		seqoutFile = f
		defer seqoutFile.Close()
	}

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	scanner := seqio.NewAutoScanner(guestFile)
	guests := []gts.Sequence{}
	for scanner.Scan() {
		guests = append(guests, scanner.Value())
	}
	if len(guests) == 0 {
		ctx.Raise(fmt.Errorf("guest sequence file %q does not contain a sequence", *guestPath))
	}

	w := bufio.NewWriter(seqoutFile)

	insert := gts.Insert
	if *embed {
		insert = gts.Embed
	}

	scanner = seqio.NewAutoScanner(hostFile)
	for scanner.Scan() {
		host := scanner.Value()

		rr := locate(host.Features())
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
			formatter := seqio.NewFormatter(out, filetype)
			if _, err := formatter.WriteTo(w); err != nil {
				return ctx.Raise(err)
			}
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

func sequenceDelete(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	locstr := pos.String("locator", "a locator string ([Select|Point|Range][@Modifier])")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	locate, err := gts.AsLocator(*locstr)
	if err != nil {
		return ctx.Raise(err)
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	seqoutFile := os.Stdout
	if *seqoutPath != "-" {
		f, err := os.Create(*seqoutPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *seqoutPath, err))
		}
		seqoutFile = f
		defer seqoutFile.Close()
	}

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		ss := gts.Minimize(locate(seq.Features()))
		flip.Flip(gts.BySegment(ss))
		for _, s := range ss {
			i, n := s.Head(), s.Len()
			seq = gts.Delete(seq, i, n)
		}
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

func sequenceReverse(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	seqoutFile := os.Stdout
	if *seqoutPath != "-" {
		f, err := os.Create(*seqoutPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *seqoutPath, err))
		}
		seqoutFile = f
		defer seqoutFile.Close()
	}

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		seq = gts.Reverse(seq)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

func sequenceRotate(ctx *flags.Context) error {
	pos, opt := flags.Flags()

	n := pos.Int("amount", "the amount to rotate the sequence by")

	var seqinPath *string
	if cmd.IsTerminal(os.Stdin.Fd()) {
		seqinPath = pos.String("input", "input sequence file (may be omitted if standard input is provided)")
	}

	seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
	backward := opt.Switch('v', "backward", "rotate the sequence backwards (equivalent to a negative amount)")
	format := opt.String('F', "format", "", "output file format (defaults to same as input)")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	if *backward {
		*n = -*n
	}

	seqinFile := os.Stdin
	if seqinPath != nil && *seqinPath != "-" {
		f, err := os.Open(*seqinPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to open file %q: %v", *seqinPath, err))
		}
		seqinFile = f
		defer seqinFile.Close()
	}

	seqoutFile := os.Stdout
	if *seqoutPath != "-" {
		f, err := os.Create(*seqoutPath)
		if err != nil {
			return ctx.Raise(fmt.Errorf("failed to create file %q: %v", *seqoutPath, err))
		}
		seqoutFile = f
		defer seqoutFile.Close()
	}

	filetype := seqio.Detect(*seqoutPath)
	if *format != "" {
		filetype = seqio.ToFileType(*format)
	}

	w := bufio.NewWriter(seqoutFile)

	scanner := seqio.NewAutoScanner(seqinFile)
	for scanner.Scan() {
		seq := scanner.Value()
		seq = gts.Rotate(seq, *n)
		formatter := seqio.NewFormatter(seq, filetype)
		if _, err := formatter.WriteTo(w); err != nil {
			return ctx.Raise(err)
		}
	}

	if err := w.Flush(); err != nil {
		return ctx.Raise(err)
	}

	if err := scanner.Err(); err != nil {
		return ctx.Raise(fmt.Errorf("encountered error in scanner: %v", err))
	}

	return nil
}

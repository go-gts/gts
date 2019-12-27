package main

import (
	"bytes"

	"github.com/ktnyt/gts"
	flags "gopkg.in/ktnyt/flags.v1"
)

func init() {
	prog := flags.NewProgram()
	prog.Add("select", "select features by feature keys", FeatureSelect)
	flags.Add("feature", "feature manipulation commands", prog.Compile())
}

func FeatureSelect(ctx *flags.Context) error {
	pos, opt := flags.Args()

	infile := pos.Input("input record file")
	outfile := pos.Output("output record file")
	mainKey := pos.String("key", "feature key to select")

	invert := opt.Switch('v', "invert-match", "select feature that do not match the given criteria")
	extraKeys := opt.StringSlice(0, "and", nil, "additional feature key(s) to select")

	if err := ctx.Parse(pos, opt); err != nil {
		return err
	}

	keys := append([]string{*mainKey}, (*extraKeys)...)
	ss := make([]gts.FeatureSelector, len(keys))
	for i, key := range keys {
		ss[i] = gts.Key(key)
	}
	sel := gts.Or(ss...)
	if *invert {
		sel = gts.Not(sel)
	}

	scanner := gts.NewRecordScanner(infile)
	buffer := &bytes.Buffer{}

	for scanner.Scan() {
		in := scanner.Record()
		ff := in.Select(sel)
		out := gts.NewRecord(in.Metadata(), ff, in.Bytes())
		gts.DefaultFormatter(out).WriteTo(buffer)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if _, err := buffer.WriteTo(outfile); err != nil {
		return err
	}

	return nil
}

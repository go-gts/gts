package main

// import (
// 	"os"
// 	"reflect"

// 	"github.com/go-gts/gts"
// 	"github.com/go-gts/gts/cmd"
// 	"github.com/go-gts/gts/internal/flags"
// 	"github.com/go-gts/gts/seqio"
// )

// func init() {
// 	flags.Register("extract", "extract the sequences referenced by the features", extractFunc)
// }

// func containsRegion(rr []gts.Region, r gts.Region) bool {
// 	for i := range rr {
// 		if reflect.DeepEqual(rr[i], r) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func extractFunc(ctx *flags.Context) error {
// 	pos, opt := flags.Flags()

// 	locstrs := pos.Extra("locator", "a locator string ([specifier][@modifier])")

// 	seqinPath := new(string)
// 	*seqinPath = "-"
// 	if cmd.IsTerminal(os.Stdin.Fd()) {
// 		seqinPath = pos.String("seqin", "input sequence file (may be omitted if standard input is provided)")
// 	}

// 	// nocache := opt.Switch(0, "no-cache", "do not use or create cache")
// 	// seqoutPath := opt.String('o', "output", "-", "output sequence file (specifying `-` will force standard output)")
// 	invert := opt.Switch('v', "invert-region", "extract the sequences that are not referenced by the features")

// 	if err := ctx.Parse(pos, opt); err != nil {
// 		return err
// 	}

// 	if len(*locstrs) == 0 {
// 		*locstrs = append(*locstrs, "@^..$")
// 	}

// 	locators := make([]gts.Locator, len(*locstrs))

// 	for i, locstr := range *locstrs {
// 		locator, err := gts.AsLocator(locstr)
// 		if err != nil {
// 			return ctx.Raise(err)
// 		}
// 		locators[i] = locator
// 	}

// 	/*
// 		if !*nocache {
// 			data := encodePayload([]tuple{
// 				{"command", strings.Join(ctx.Name, "-")},
// 				{"version", gts.Version.String()},
// 				{"locators", *locstrs},
// 				{"filetype", filetype},
// 			})

// 		}
// 	*/

// 	var stream seqio.IOStream

// 	err := stream.ForEach(func(i int, header interface{}, ff gts.Features) (seqio.Callback, error) {
// 		return func(seq gts.Sequence) error {
// 			rr := make([]gts.Region, 0)
// 			for _, locate := range locators {
// 				for _, r := range locate(ff, seq) {
// 					if !containsRegion(rr, r) {
// 						rr = append(rr, r)
// 					}
// 				}
// 			}

// 			if *invert {
// 				// Support linear inversion only as topology is not well defined.
// 				rr = gts.InvertLinear(gts.Regions(rr), gts.Len(seq))
// 			}

// 			for _, region := range rr {
// 				if len(rr) == 1 || region.Len() != gts.Len(seq) {
// 					out := region.Locate(seq)
// 					gg := region.Crop(ff)
// 					if err := stream.PushHeader(header); err != nil {
// 						return err
// 					}
// 					if err := stream.PushFeatures(gg); err != nil {
// 						return err
// 					}
// 					if err := stream.PushSequence(out); err != nil {
// 						return err
// 					}
// 				}
// 			}

// 			return nil
// 		}, nil
// 	})

// 	return ctx.Raise(err)
// }

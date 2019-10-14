package main

import (
	"bufio"
	"image/color"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ktnyt/gt1/flags"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func init() {
	register("plot", "basic plotting routines", plotFunc)
}

func mustFloat64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func loadXYData(r io.Reader, hasHeader bool, delim string, xIndex, yIndex int) ([]string, plotter.XYs) {
	xys := make([]plotter.XY, 0)
	scanner := bufio.NewScanner(r)
	header := []string{"", ""}
	if hasHeader && scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, delim)
		header = []string{fields[xIndex], fields[yIndex]}
	}
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, delim)
		x := mustFloat64(fields[xIndex])
		y := mustFloat64(fields[yIndex])
		xys = append(xys, plotter.XY{x, y})
	}
	return header, plotter.XYs(xys)
}

func plotLineFunc(command *flags.Command, args []string) error {
	infile := command.Infile("input data file")
	outfile := command.Outfile("output image file")
	noHeader := command.Switch(0, "no-header", "do not interpret the first line as a header")
	delim := command.String('d', "delimiter", "\t", "string to split input file lines by")
	xIndex := command.Int('x', "x-index", 0, "column index to use as x axis value")
	yIndex := command.Int('y', "y-index", 1, "column index to use as y axis value")
	width := command.Int('w', "width", 640, "output image width")
	height := command.Int('h', "height", 480, "output image height")
	title := command.String(0, "title", "", "plot title")

	return command.Run(args, func() error {
		header, data := loadXYData(infile, !(*noHeader), *delim, *xIndex, *yIndex)

		p, err := plot.New()
		if err != nil {
			return err
		}

		if *title != "" {
			p.Title.Text = *title
		}
		p.X.Label.Text = header[0]
		p.Y.Label.Text = header[1]

		p.Add(plotter.NewGrid())

		line, err := plotter.NewLine(data)
		if err != nil {
			return err
		}

		line.LineStyle.Color = color.RGBA{R: 255, G: 128, B: 128, A: 255}

		p.Add(line)

		w := vg.Length(*width)
		h := vg.Length(*height)
		format := filepath.Ext(outfile.Name())[1:]

		writerTo, err := p.WriterTo(w, h, format)
		if err != nil {
			return err
		}

		_, err = writerTo.WriteTo(outfile)

		return err
	})
}

func plotFunc(command *flags.Command, args []string) error {
	command.Command("line", "create a line plot", plotLineFunc)
	return command.Run(args)
}

package lib

import (
	"io"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"github.com/gonum/plot/vg/vgsvg"
)

const dpi = 96

func PlotGraph(title string, timestamps []int64, w io.Writer) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = title
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Number of stars"
	p.Y.Min = 0

	points := make(plotter.XYs, len(timestamps))
	for i, timestamp := range timestamps {
		points[i].X = float64(timestamp)
		points[i].Y = float64(i + 1)
	}
	plotutil.AddLinePoints(p, "Stars", points)

	c := vgsvg.New(4*vg.Inch, 4*vg.Inch)

	p.Draw(draw.New(c))

	if _, err := c.WriteTo(w); err != nil {
		return err
	}
	return nil
}

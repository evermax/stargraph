package lib

import (
	"io"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
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

	c := vgimg.New(4*vg.Inch, 4*vg.Inch)
	cpng := vgimg.PngCanvas{Canvas: c}

	p.Draw(draw.New(cpng))

	if _, err := cpng.WriteTo(w); err != nil {
		return err
	}
	return nil
}

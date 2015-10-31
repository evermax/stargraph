package main

import (
	"encoding/json"
	"os"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

func persistData(timestamps []int64) error {
	canvasDBstars := make([]canvasDBstar, len(timestamps))
	jqplotDBstars := make([][]int64, len(timestamps))
	for i, timestamp := range timestamps {
		canvasDBstars[i] = canvasDBstar{Timestamp: timestamp, Count: i + 1}
		jqplotDBstars[i] = []int64{timestamp, int64(i + 1)}
	}
	if err := writeCanvasDB(canvasDBstars); err != nil {
		return err
	}
	if err := writeJQplotDB(jqplotDBstars); err != nil {
		return err
	}
	return nil
}

func writeCanvasDB(canvasDBstars []canvasDBstar) error {
	bytes, err := json.Marshal(canvasDBstars)
	if err != nil {
		return err
	}
	file, err := os.OpenFile("canvasDB.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

type canvasDBstar struct {
	Timestamp int64 `json:"x"`
	Count     int   `json:"y"`
}

func writeJQplotDB(values [][]int64) error {
	bytes, err := json.Marshal(values)
	if err != nil {
		return err
	}
	file, err := os.OpenFile("jqplotDB.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_, err = file.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func plotGraph(timestamps []int64) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = "Graph of the stars of the project over time"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Number of stars"
	p.Y.Min = 0

	points := make(plotter.XYs, len(timestamps))
	for i, timestamp := range timestamps {
		points[i].X = float64(timestamp)
		points[i].Y = float64(i + 1)
	}
	plotutil.AddLinePoints(p, "Stars", points)
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "graph.png"); err != nil {
		return err
	}
	return nil
}

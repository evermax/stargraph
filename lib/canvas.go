package lib

import (
	"encoding/json"
	"io"
)

type CanvasData struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

func WriteCanvasJS(timestamps []int64, w io.Writer) error {
	canvasData := make([]CanvasData, len(timestamps))
	for i, timestamp := range timestamps {
		canvasData[i] = CanvasData{X: timestamp * 1000, Y: int64(i + 1)}
	}
	bytes, err := json.Marshal(canvasData)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

package lib

import (
	"encoding/json"
	"io"

	"github.com/evermax/stargraph/github"
)

type CanvasData struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type CanvasJSON struct {
	CreatedDate string       `json:"created_at"`
	Data        []CanvasData `json:"data"`
}

func WriteCanvasJS(timestamps []int64, info github.RepoInfo, w io.Writer) error {
	canvasData := make([]CanvasData, len(timestamps))
	for i, timestamp := range timestamps {
		canvasData[i] = CanvasData{X: timestamp * 1000, Y: int64(i + 1)}
	}

	bytes, err := json.Marshal(CanvasJSON{CreatedDate: info.CreationDate, Data: canvasData})
	if err != nil {
		return err
	}

	_, err = w.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

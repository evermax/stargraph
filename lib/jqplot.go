package lib

import (
	"encoding/json"
	"io"
)

func WriteJQPlot(timestamps []int64, w io.Writer) error {
	jqplots := make([][]int64, len(timestamps))
	for i, timestamp := range timestamps {
		jqplots[i] = []int64{timestamp, int64(i + 1)}
	}
	bytes, err := json.Marshal(jqplots)
	if err != nil {
		return err
	}
	_, err = w.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

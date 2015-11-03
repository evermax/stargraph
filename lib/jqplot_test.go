package lib

import (
	"bytes"
	"testing"
)

func TestWriteJQPlot(t *testing.T) {
	timestamps := []int64{1234, 5678, 91011, 121314}
	expected := "[[1234,1],[5678,2],[91011,3],[121314,4]]"
	var buff bytes.Buffer
	if err := WriteJQPlot(timestamps, &buff); err != nil {
		t.Fatalf("An error occured when writting JQ Plot: %v", err)
	}
	str := buff.String()
	if str != expected {
		t.Fatalf("Expected %s, got %s", expected, str)
	}

}

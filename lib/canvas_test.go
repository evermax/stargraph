package lib

import (
	"bytes"
	"testing"
)

func TestWriteCanvas(t *testing.T) {
	timestamps := []int64{1234, 5678, 91011, 121314}
	expected := `[{"x":1234000,"y":1},{"x":5678000,"y":2},{"x":91011000,"y":3},{"x":121314000,"y":4}]`
	var buff bytes.Buffer
	if err := WriteCanvasJS(timestamps, &buff); err != nil {
		t.Fatalf("An error when writing the writer: %v", err)
	}
	str := buff.String()
	if str != expected {
		t.Fatalf("Expected %s, got %s", expected, str)
	}
}

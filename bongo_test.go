package bongo

import (
	"io"
	"testing"
)

type testFront struct{}

func (f testFront) Parse(in io.Reader) (map[string]interface{}, string, error) {
	return nil, "yoyo", nil
}
func TestApp(t *testing.T) {
	app := NewApp()
	app.Run("testdata/sample")
}

func TestSet(t *testing.T) {
	app := NewApp()
	app.Set(testFront{})
}

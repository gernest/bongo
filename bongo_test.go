package bongo

import (
	"testing"
)

func TestApp(t *testing.T) {
	app := NewApp()
	app.Run("testdata/sample")
}

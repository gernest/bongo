package bongo

import "testing"

func TestApp(t *testing.T) {
	app := New()
	err := app.Run("testdata/sample")
	if err != nil {
		t.Error(err)
	}
}

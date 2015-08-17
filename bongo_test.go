package bongo

import (
	"errors"
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
	testFileLoader := func(in string) ([]string, error) {
		return []string{in}, nil
	}
	testRenderer := func(pages PageList, opts ...interface{}) error {
		return errors.New("whacko")
	}
	app := NewApp()
	app.Set(testFront{})
	app.Set(testFileLoader)
	app.Set(testRenderer)

	if foo, _ := app.fileLoader("foo"); foo[0] != "foo" {
		t.Errorf("expected foo got %s", foo)
	}
	if err := app.rendr(nil); err.Error() != "whacko" {
		t.Errorf("expected whcko got %v", err)
	}
}

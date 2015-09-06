package bongo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	//ErrIsEmpty is an error indicating no front matter was found
	ErrIsEmpty = errors.New("an empty file")

	//ErrUnknownDelim is returned when the delimiters are not known by the
	//FrontMatter implementation.
	ErrUnknownDelim = errors.New("unknown delim")

	defaultDelim = "---"
)

type (
	//HandlerFunc is an interface for a function that process front matter text.
	HandlerFunc func(string) (map[string]interface{}, error)
)

//Matter is all what matters here.
type Matter struct {
	handlers  map[string]HandlerFunc
	delim     string
	lastDelim bool
	lastIndex int
}

func newMatter() *Matter {
	return &Matter{handlers: make(map[string]HandlerFunc)}
}

//NewYAML returns a new FrontMatter implementation with support for yaml frontmatter.
// default delimiters is ---
func NewYAML(opts ...string) *Matter {
	delim := defaultDelim
	if len(opts) > 0 {
		delim = opts[0]
	}
	m := newMatter()
	m.Handle(delim, YAMLHandler)
	return m
}

//NewJSON returns a new FrontMatter implementation with support for json frontmatter.
// default delimiters is ---
func NewJSON(opts ...string) *Matter {
	delim := defaultDelim
	if len(opts) > 0 {
		delim = opts[0]
	}
	m := newMatter()
	m.Handle(delim, JSONHandler)
	return m
}

//Handle registers a handler for the given frontmatter delimiter
func (m *Matter) Handle(delim string, fn HandlerFunc) {
	m.handlers[delim] = fn
}

// Parse parses the input and extract the frontmatter
func (m *Matter) Parse(input io.Reader) (front map[string]interface{}, body io.Reader, err error) {
	return m.parse(input)
}
func (m *Matter) parse(input io.Reader) (front map[string]interface{}, body io.Reader, err error) {
	var getFront = func(f string) string {
		return strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(f, m.delim), m.delim))
	}
	f, body, err := m.splitFront(input)
	if err != nil {
		return nil, nil, err
	}

	h := m.handlers[f[:3]]
	front, err = h(getFront(f))
	if err != nil {
		return nil, nil, err
	}

	return front, body, nil

}
func sniffDelim(input []byte) (string, error) {
	if len(input) < 4 {
		return "", ErrIsEmpty
	}
	return string(input[:3]), nil
}

func (m *Matter) splitFront(input io.Reader) (front string, body io.Reader, err error) {
	s := bufio.NewScanner(input)
	s.Split(m.split)
	var (
		f string
		b string
	)
	n := 0
	for s.Scan() {
		if n == 0 {
			f = s.Text()
		} else {
			b = b + s.Text()
		}
		n++

	}
	return f, strings.NewReader(b), s.Err()
}

//split implements bufio.SplitFunc for spliting fron matter from the body text.
func (m *Matter) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if m.delim == "" {
		delim, err := sniffDelim(data)
		if err != nil {
			return 0, nil, err
		}
		m.delim = delim
	}
	if _, ok := m.handlers[m.delim]; !ok {
		return 0, nil, ErrUnknownDelim
	}
	if x := bytes.Index(data, []byte(m.delim)); x >= 0 {
		// check the next delim index
		if next := bytes.Index(data[x+len(m.delim):], []byte(m.delim)); next > 0 {
			if !m.lastDelim {
				m.lastDelim = true
				m.lastIndex = next + len(m.delim)
				return next + len(m.delim)*2, dropSpace(data[x : next+len(m.delim)]), nil
			}
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func dropSpace(d []byte) []byte {
	return bytes.TrimSpace(d)
}

//JSONHandler implements HandlerFunc interface. It extracts front matter data from the given
// string argument by interpreting it as a json string.
func JSONHandler(front string) (map[string]interface{}, error) {
	var rst interface{}
	err := json.Unmarshal([]byte(front), &rst)
	if err != nil {
		return nil, err
	}
	return rst.(map[string]interface{}), nil
}

//YAMLHandler decodes ymal string into a go map[string]interface{}
func YAMLHandler(front string) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(front), out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

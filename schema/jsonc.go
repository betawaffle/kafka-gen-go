package schema

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

func (m *MessageData) readJSONC(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := &jsoncReader{reader: f}
	defer func() { m.Comments = r.comments }()

	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	return d.Decode(m)
}

type jsoncState int

const (
	jsoncStart jsoncState = iota
	jsoncSlash
	jsoncComment
)

type jsoncReader struct {
	reader   io.Reader
	state    jsoncState
	comments []byte
}

func (jr *jsoncReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	var i int
	if jr.state == jsoncSlash {
		p[0], i = '/', 1
	}
	j, err := jr.reader.Read(p[i:])
	j += i
	return jr.filter(p[:j]), err
}

func (jr *jsoncReader) filter(p []byte) int {
	var r, w int
	for r < len(p) {
		var i, n int
		switch jr.state {
		case jsoncStart:
			if i := bytes.IndexByte(p[r:], '/'); i != -1 {
				jr.state, n = jsoncSlash, i
			} else {
				n = len(p) - r
			}
		case jsoncSlash:
			if r+1 == len(p) {
				return w
			} else if p[r+1] != '/' {
				jr.state, n = jsoncStart, 2
			} else {
				jr.state, i = jsoncComment, 2
			}
		case jsoncComment:
			if j := bytes.IndexByte(p[r:], '\n'); j != -1 {
				jr.state, i = jsoncStart, j+1
			} else {
				i = len(p) - r
			}
		default:
			panic("invalid jsoncState")
		}
		i += r
		n += i
		if i > r {
			jr.comments = append(jr.comments, p[r:i]...)
		}
		if n > i {
			w += copy(p[w:], p[i:n])
		}
		r = n
	}
	return w
}

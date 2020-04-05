package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"go/scanner"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type File struct {
	name string
	data []byte
	tabs int
}

func NewFile(name string) *File {
	f := &File{name: name}
	f.WriteString("// Code generated by kafka-gen-go. DO NOT EDIT.\n\n")
	f.WriteString("package ")
	f.WriteString(f.pkgName())
	f.WriteString("\n\n")
	return f
}

func (f *File) Flush() error {
	b, err := format.Source(f.data)
	if err != nil {
		return goFormatError(err, f.name, f.data)
	}
	return ioutil.WriteFile(f.name, b, 0644)
}

func (f *File) Write(v []byte) (int, error) {
	var i int
	for {
		j := bytes.IndexByte(v[i:], '\n')
		if j == -1 {
			break
		}
		j += i
		j++
		f.data = append(f.data, v[i:j]...)
		f.updateTabs()
		i = j
	}
	f.data = append(f.data, v[i:]...)
	return len(v), nil
}

func (f *File) Writef(format string, args ...interface{}) {
	fmt.Fprintf(f, format, args...)
}

func (f *File) WriteBool(v bool) {
	f.data = strconv.AppendBool(f.data, v)
}

func (f *File) WriteByte(v byte) error {
	if v == '\n' {
		f.data = append(f.data, '\n')
		f.updateTabs()
		return nil
	}
	f.data = append(f.data, v)
	return nil
}

func (f *File) WriteFloat(v float64, fmt byte, prec, bitSize int) {
	f.data = strconv.AppendFloat(f.data, v, fmt, prec, bitSize)
}

func (f *File) WriteInt(v int64, base int) {
	f.data = strconv.AppendInt(f.data, v, base)
}

func (f *File) WriteQuoted(v string) {
	f.data = strconv.AppendQuote(f.data, v)
}

func (f *File) WriteRune(v rune) (int, error) {
	if v == '\n' {
		f.data = append(f.data, '\n')
		f.updateTabs()
		return 1, nil
	}
	b := make([]byte, utf8.UTFMax)
	n := utf8.EncodeRune(b, v)
	f.data = append(f.data, b...)
	return n, nil
}

func (f *File) WriteString(v string) (int, error) {
	var i int
	for {
		j := strings.IndexByte(v[i:], '\n')
		if j == -1 {
			break
		}
		j += i
		j++
		f.data = append(f.data, v[i:j]...)
		f.updateTabs()
		i = j
	}
	f.data = append(f.data, v[i:]...)
	return len(v), nil
}

func (f *File) WriteUint(v uint64, base int) {
	f.data = strconv.AppendUint(f.data, v, base)
}

func (f *File) pkgName() string {
	return filepath.Base(filepath.Dir(f.name))
}

func (f *File) updateTabs() {
	j := len(f.data) - 1
	i := bytes.LastIndexByte(f.data[:j], '\n') + 1
	if i == j {
		return
	}
	switch f.data[j-1] {
	case '{':
		f.tabs++
	case '}':
		if k := j - 2; k != -1 && f.data[k] == '{' {
			break
		}
		f.tabs--
	}
	for range make([]struct{}, f.tabs) {
		f.data = append(f.data, '\t')
	}
}

func goFormatError(err error, filename string, data []byte) error {
	es, ok := err.(scanner.ErrorList)
	if !ok {
		panic(fmt.Errorf("expected scanner.ErrorList, got %T: %[1]v", err))
	}
	var buf strings.Builder
	for _, e := range es {
		if e.Pos.IsValid() {
			e.Pos.Filename = filename
			buf.WriteString(e.Error())
			writeContext(&buf, e, data)
		}
	}
	return errors.New(buf.String())
}

func writeContext(w *strings.Builder, e *scanner.Error, b []byte) {
	zc := e.Pos.Column - 1
	if zc == -1 {
		return
	}
	w.WriteString("\n\t")
	i := e.Pos.Offset - zc
	if j := bytes.LastIndexByte(b[:i-1], '\n'); j != -1 {
		j++
		w.Write(b[j:i])
		w.WriteByte('\t')
	}
	var d int
	if j := bytes.IndexByte(b[i:], '\n'); j != -1 {
		j++
		j += i
		w.Write(b[i:j])
		d = countTabs(b[i:j])
	} else {
		w.Write(b[i:])
		w.WriteByte('\n')
		d = countTabs(b[i:])
	}
	for range make([]struct{}, d+1) {
		w.WriteByte('\t')
	}
	for range make([]struct{}, zc-d) {
		w.WriteByte(' ')
	}
	w.WriteString("^\n\n")
}

func countTabs(b []byte) int {
	for i, c := range b {
		if c != '\t' {
			return i
		}
	}
	return len(b)
}
package schema

import (
	"bytes"
	"strings"

	"github.com/iancoleman/strcase"
)

func (m *MessageData) applyHacks() {
	applyCommentsHack(m)
	switch {
	case strings.HasPrefix(m.Name, "IncrementalAlterConfigs"):
		for _, f := range m.Fields {
			applyIncrementalHack(f)
		}
	case strings.HasPrefix(m.Name, "CreateTopicsRequest"):
		for _, f := range m.Fields {
			applyExportHack(f)
		}
	}
}

func applyExportHack(f *Field) {
	if c := f.Name[0]; c >= 'a' && c <= 'z' {
		f.Name = strcase.ToCamel(f.Name)
	}
	for _, sf := range f.Fields {
		applyExportHack(sf)
	}
}

func applyIncrementalHack(f *Field) {
	switch t := f.Type.Elem; t {
	case "AlterConfigsResource", "AlterConfigsResourceResponse":
		f.Type.Elem = "Incremental" + t
	case "AlterableConfig":
		f.Type.Elem = "Incrementally" + t
	}
	for _, sf := range f.Fields {
		applyIncrementalHack(sf)
	}
}

func applyCommentsHack(m *MessageData) {
	m.Comments = trimEmptyComments(trimLines(m.Comments, 14))
}

func trimLines(b []byte, n int) []byte {
	var i int
	for n > 0 {
		j := bytes.IndexByte(b[i:], '\n')
		if j == -1 {
			return nil
		}
		i += j
		i++
		n--
	}
	return b[i:]
}

func trimEmptyComments(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	var more bool
	if b[2] == '\n' {
		b, more = b[3:], true
	} else if b[3] == '\n' {
		b, more = b[4:], true
	}
	if j := len(b) - 2; b[j] == ' ' {
		b, more = b[:j-2], true
	} else if b[j] == '/' {
		b, more = b[:j-1], true
	}
	if more {
		b = trimEmptyComments(b)
	}
	return b
}

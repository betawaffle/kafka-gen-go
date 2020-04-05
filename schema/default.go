package schema

import (
	"bytes"
	"errors"
	"strconv"
)

var (
	errDefaultNotBoolean = errors.New("default value is not a boolean")
	errDefaultNotInteger = errors.New("default value is not an integer")
	errDefaultNotString  = errors.New("default value is no a string")
)

type jsonType int

const (
	jsonNull jsonType = iota
	jsonFalse
	jsonTrue
	jsonNumber
	jsonString
	jsonArray
	jsonObject
)

type Default struct {
	typ jsonType
	str string
	set bool
}

func (v *Default) Boolean() bool {
	if !v.set || v.typ == jsonFalse {
		return false
	}
	if v.typ == jsonTrue {
		return true
	}
	if v.typ != jsonString {
		panic(errDefaultNotBoolean)
	}
	b, err := strconv.ParseBool(v.str)
	if err != nil {
		panic(err)
	}
	return b
}

func (v *Default) Integer(bitSize int) int64 {
	if !v.set || v.str == "null" {
		return 0
	}
	i, err := strconv.ParseInt(v.str, 0, bitSize)
	if err != nil {
		panic(err)
	}
	return i
}

func (v *Default) String() string {
	if !v.set {
		return ""
	}
	if v.typ == jsonString {
		return v.str
	}
	panic(errDefaultNotString)
}

func (v *Default) UnmarshalJSON(b []byte) error {
	if v.set {
		panic("BUG: reused jsoncValue")
	}
	v.set = true
	switch b[0] {
	case 'n':
		panic("BUG: unhandled null default")
	case 'f':
		v.typ = jsonFalse
	case 't':
		v.typ = jsonTrue
	case '"':
		i, j := 1, len(b)-1
		v.typ = jsonString
		switch {
		case i == j:
			v.str = ""
		case bytes.IndexByte(b[i:j], '\\') == -1:
			v.str = string(b[i:j])
		default:
			panic("complex json string unquoting not implemented")
		}
	case '[':
		panic("BUG: unhandled array default")
	case '{':
		panic("BUG: unhandled object default")
	default:
		v.typ = jsonNumber
		v.str = string(b)
	}
	return nil
}

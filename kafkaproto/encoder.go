package kafkaproto

import "encoding/binary"

type encoder []byte

func (e *encoder) encodeBool(v bool) {
	var b byte
	if v {
		b = 1
	} else {
		b = 0
	}
	*e = append(*e, b)
}

func (e *encoder) encodeInt8(v int8) {
	*e = append(*e, byte(v))
}

func (e *encoder) encodeInt16(v int16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	*e = append(*e, b...)
}

func (e *encoder) encodeInt32(v int32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	*e = append(*e, b...)
}

func (e *encoder) encodeInt64(v int64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	*e = append(*e, b...)
}

func (e *encoder) encodeString(v string) {
	e.encodeInt16(int16(len(v)))
	*e = append(*e, v...)
}

func (e *encoder) encodeBytes(v []byte) {
	e.encodeInt32(int32(len(v)))
	*e = append(*e, v...)
}

package kafkaproto

import "encoding/binary"

type decoder []byte

func (d *decoder) decodeBool() bool {
	b := *d
	*d = b[1:]
	return b[0] != 0
}

func (d *decoder) decodeInt8() int8 {
	b := *d
	*d = b[1:]
	return int8(b[0])
}

func (d *decoder) decodeInt16() int16 {
	b := *d
	*d = b[2:]
	return int16(binary.BigEndian.Uint16(b[:2]))
}

func (d *decoder) decodeInt32() int32 {
	b := *d
	*d = b[4:]
	return int32(binary.BigEndian.Uint32(b[:4]))
}

func (d *decoder) decodeInt64() int64 {
	b := *d
	*d = b[8:]
	return int64(binary.BigEndian.Uint64(b[:8]))
}

func (d *decoder) decodeString() string {
	n := int(d.decodeInt16())
	b := *d
	*d = b[n:]
	return string(b[:n])
}

func (d *decoder) decodeBytes() []byte {
	n := int(d.decodeInt32())
	b := *d
	*d = b[n:]
	return b[:n]
}

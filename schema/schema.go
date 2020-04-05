package schema

import (
	"bytes"
	"strconv"
)

type CommonStruct struct {
	Name     string        `json:"name"`
	Versions *VersionRange `json:"versions"`
	Fields   []*Field      `json:"fields"`
}

type Field struct {
	Name             string        `json:"name"`
	Desc             string        `json:"about"`
	Type             FieldType     `json:"type"`
	Versions         *VersionRange `json:"versions"`
	TaggedVersions   *VersionRange `json:"taggedVersions"`
	NullableVersions *VersionRange `json:"nullableVersions"`
	FlexibleVersions *VersionRange `json:"flexibleVersions"`
	Fields           []*Field      `json:"fields"`
	Default          Default       `json:"default"`
	Tag              *int32        `json:"tag"`
	EntityType       string        `json:"entityType"`
	MapKey           bool          `json:"mapKey"`
	Ignorable        bool          `json:"ignorable"`
}

type FieldType struct {
	Elem  string
	Array bool
}

func (t *FieldType) UnmarshalText(b []byte) error {
	if len(b) > 2 && string(b[:2]) == "[]" {
		t.Array, b = true, b[2:]
	}
	t.Elem = string(b)
	return nil
}

type MessageData struct {
	ApiKey           *int16          `json:"apiKey"`
	Type             string          `json:"type"`
	Name             string          `json:"name"`
	ValidVersions    *VersionRange   `json:"validVersions"`
	FlexibleVersions *VersionRange   `json:"flexibleVersions"`
	Fields           []*Field        `json:"fields"`
	CommonStructs    []*CommonStruct `json:"commonStructs"`
	Comments         []byte          `json:"-"`
}

func ReadMessageData(filename string) (*MessageData, error) {
	msg := new(MessageData)
	err := msg.readJSONC(filename)
	if err != nil {
		return nil, err
	}
	msg.applyHacks()
	return msg, nil
}

type VersionRange struct {
	Min int16
	Max int16
}

func (v *VersionRange) UnmarshalText(b []byte) error {
	// Closed Range
	if i := bytes.IndexByte(b, '-'); i != -1 {
		min, err := parseVersion(b[:i])
		if err != nil {
			return err
		}
		max, err := parseVersion(b[i+1:])
		if err != nil {
			return err
		}
		*v = VersionRange{Min: min, Max: max}
		return nil
	}

	// Open Range
	if i := len(b) - 1; b[i] == '+' {
		min, err := parseVersion(b[:i])
		if err != nil {
			return err
		}
		*v = VersionRange{Min: min, Max: -1}
		return nil
	}

	// Empty Range
	if string(b) == "none" {
		*v = VersionRange{Min: -1, Max: -1}
		return nil
	}

	// Single Version
	min, err := parseVersion(b)
	if err != nil {
		return err
	}
	*v = VersionRange{Min: min, Max: min}
	return nil
}

func parseVersion(b []byte) (int16, error) {
	v, err := strconv.ParseInt(string(b), 10, 15)
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}

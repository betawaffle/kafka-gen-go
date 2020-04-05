package kafkaproto

import "errors"

var (
	errVersion = errors.New("unsupported message version")
)

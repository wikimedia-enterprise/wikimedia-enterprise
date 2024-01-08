package schema

import (
	"encoding/binary"
	"errors"

	"github.com/hamba/avro/v2"
)

// ErrUnknownMagicByte if first byte of the message is not zero.
var ErrUnknownMagicByte = errors.New("unknown magic byte")

// ErrEmptyData data slice length is less than 6 bytes.
var ErrEmptyData = errors.New("empty data slice")

// Unmarshal decode message in wire format.
func Unmarshal(sch avro.Schema, data []byte, v interface{}, api avro.API) error {
	if len(data) < 6 {
		return ErrEmptyData
	}

	if byte(0) != data[0] {
		return ErrUnknownMagicByte
	}

	if api == nil {
		api = avro.DefaultConfig
	}

	return api.Unmarshal(sch, data[5:], v)
}

// GetID gets schema id from the key or value of the message.
func GetID(data []byte) int {
	if len(data) >= 5 {
		return int(binary.BigEndian.Uint32(data[1:5]))
	}

	return 0
}

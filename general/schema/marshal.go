package schema

import (
	"encoding/binary"

	"github.com/hamba/avro/v2"
)

// Marshal the message in wire format to make it schema registry compatible.
// See https://docs.confluent.io/5.0.0/schema-registry/docs/serializer-formatter.html?_ga=2.159130106.1161873717.1636281013-1282765040.1635954506#wire-format for more info.
func Marshal(id int, sch avro.Schema, v interface{}) ([]byte, error) {
	data, err := avro.Marshal(sch, v)

	if err != nil {
		return []byte{}, err
	}

	idb := make([]byte, 4)
	binary.BigEndian.PutUint32(idb, uint32(id)) // #nosec G115: False positive — uint32 cast is safe as input values are within uint32 range.

	return append(append(append([]byte{}, byte(0)), idb...), data...), nil
}

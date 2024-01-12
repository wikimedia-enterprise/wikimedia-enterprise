package ksqldb

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

// ConvString convert interface to string.
func ConvString(val interface{}) string {
	switch val := val.(type) {
	case string:
		return val
	}

	return ""
}

// ConvInt convert interface to integer.
func ConvInt(val interface{}) int {
	switch val := val.(type) {
	case int:
		return val
	case float64:
		return int(val)
	}

	return 0
}

// ConvTime convert interface to time.
func ConvTime(val interface{}) *time.Time {
	switch val := val.(type) {
	case float64:
		ts := time.UnixMicro(int64(val))
		return &ts
	}

	ts := time.UnixMicro(0)
	return &ts
}

// ConvBool convert interface to boolean.
func ConvBool(val interface{}) bool {
	switch val := val.(type) {
	case bool:
		return val
	}

	return false
}

// ConvFloat64 convert interface to floating point number.
func ConvFloat64(val interface{}) float64 {
	switch val := val.(type) {
	case float32:
		return float64(val)
	case float64:
		return val
	}

	return 0
}

// ConvMap convert interface to map.
func ConvMap(val interface{}) Map {
	res := Map{}

	switch val := val.(type) {
	case map[string]interface{}:
		return Map(val)
	}

	return res
}

// ConvRow convert interface to row.
func ConvRow(val interface{}) Row {
	res := Row{}

	switch val := val.(type) {
	case []interface{}:
		return Row(val)
	}

	return res
}

// ConvPtr convert reflect pointer value to struct or time.
func ConvPtr(val interface{}, value reflect.Value) error {
	if value.CanSet() {
		switch value.Type().String() {
		case "*time.Time":
			value.Set(reflect.ValueOf(ConvTime(val)))
		default:
			switch val.(type) {
			case map[string]interface{}:
				if value.IsNil() {
					value.Set(reflect.New(value.Type().Elem()))
				}

				dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
					TagName:  "avro",
					Result:   value.Interface(),
					Metadata: nil,
					DecodeHook: func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
						switch to.String() {
						case "*time.Time":
							return ConvTime(v), nil
						default:
							return v, nil
						}
					},
				})

				if err != nil {
					return err
				}

				return dec.Decode(val)
			}
		}
	}

	return nil
}

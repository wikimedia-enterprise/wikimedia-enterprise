package ksqldb

import (
	"reflect"
	"time"
)

// Row ksqlDB row representation.
type Row []interface{}

// String get string value by row index.
func (r Row) String(idx int) string {
	if idx >= 0 && idx < len(r) {
		return ConvString(r[idx])
	}

	return ""
}

// Int get integer value by row index.
func (r Row) Int(idx int) int {
	if idx >= 0 && idx < len(r) {
		return ConvInt(r[idx])
	}

	return 0
}

// Bool get boolean value by row index.
func (r Row) Bool(idx int) bool {
	if idx >= 0 && idx < len(r) {
		return ConvBool(r[idx])
	}

	return false
}

// Float64 get floating point value by row index.
func (r Row) Float64(idx int) float64 {
	if idx >= 0 && idx < len(r) {
		return ConvFloat64(r[idx])
	}

	return 0
}

// Time get time (timestamp) value by row index.
func (r Row) Time(idx int) *time.Time {
	if idx >= 0 && idx < len(r) {
		return ConvTime(r[idx])
	}

	ts := time.UnixMicro(0)
	return &ts
}

// Row get full new row of values value by row index.
func (r Row) Row(idx int) Row {
	if idx >= 0 && idx < len(r) {
		return ConvRow(r[idx])
	}

	return Row{}
}

// Map get full new map values by row index.
func (r Row) Map(idx int) Map {
	if idx >= 0 && idx < len(r) {
		return ConvMap(r[idx])
	}

	return Map{}
}

// Value converts row value to reflection value.
func (r Row) Value(idx int, value reflect.Value) {
	if idx >= 0 && idx < len(r) {
		if value.CanSet() {
			switch value.Kind() {
			case reflect.String:
				value.SetString(r.String(idx))
			case reflect.Int:
				value.SetInt(int64(r.Int(idx)))
			case reflect.Bool:
				value.SetBool(r.Bool(idx))
			case reflect.Float64:
				value.SetFloat(r.Float64(idx))
			}
		}
	}
}

// ForEach loops through the row and returns row and it's index.
func (r Row) ForEach(cb func(r Row, i int) error) error {
	for i := 0; i < len(r); i++ {
		if err := cb(r, i); err != nil {
			return err
		}
	}

	return nil
}

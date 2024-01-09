package ksqldb

import (
	"strings"
	"time"
)

// Map ksqlDB struct representation.
type Map map[string]interface{}

// String get string value by map key.
func (m Map) String(idx string) string {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvString(m[idxup])
	}

	return ""
}

// Int get integer value by map key.
func (m Map) Int(idx string) int {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvInt(m[idxup])
	}

	return 0
}

// Bool get boolean value by map key.
func (m Map) Bool(idx string) bool {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvBool(m[idxup])
	}

	return false
}

// Float64 get floating point value by map key.
func (m Map) Float64(idx string) float64 {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvFloat64(m[idxup])
	}

	return 0
}

// Time get time (timestamp) value by map key.
func (m Map) Time(idx string) *time.Time {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvTime(m[idxup])
	}

	ts := time.UnixMicro(0)
	return &ts
}

// Row get full new row of values value by map key.
func (m Map) Row(idx string) Row {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvRow(m[idxup])
	}

	return Row{}
}

// Map get full new map values by map key.
func (m Map) Map(idx string) Map {
	idxup := strings.ToUpper(idx)

	if _, ok := m[idxup]; ok {
		return ConvMap(m[idxup])
	}

	return Map{}
}

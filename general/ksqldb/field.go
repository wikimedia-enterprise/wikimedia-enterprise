package ksqldb

import (
	"fmt"
	"reflect"
	"strings"
)

// NewField creates new ksqldb field instance.
func NewField(table string, value reflect.Value, str reflect.StructField) *Field {
	name := str.Tag.Get("avro")

	return &Field{
		Table:  table,
		Name:   name,
		Join:   str.Tag.Get("ksql"),
		Value:  value,
		Str:    str,
		Column: strings.TrimPrefix(fmt.Sprintf("%s_%s", table, name), "_"),
	}
}

// Field creates a wrapper on top of reflection to simplify work for decoder.
type Field struct {
	Name   string
	Table  string
	Join   string
	Column string
	Value  reflect.Value
	Str    reflect.StructField
}

// IsValid check if field contains avro tag.
func (f *Field) IsValid() bool {
	return len(f.Name) > 0
}

// IsJoin checks if this is a join field.
func (f *Field) IsJoin() bool {
	return len(f.Join) > 0
}

// NewValue check if field value is nil and assign new pointer value.
func (f *Field) NewValue() {
	if f.Value.IsNil() {
		f.Value.Set(reflect.New(f.Value.Type().Elem()))
	}
}

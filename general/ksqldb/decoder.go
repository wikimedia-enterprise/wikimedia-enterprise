package ksqldb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrTableNameIsEmpty indicates that table name field not found in the struct
// or that field has no `ksql` tag that indicates the name.
var ErrTableNameIsEmpty = errors.New("TableName field not found or `ksql` tag is empty")

// NewDecoder creates new instance of ksql avro decoder.
func NewDecoder(row Row, columns []string, ops ...func(*Decoder)) *Decoder {
	dec := &Decoder{
		row:     row,
		columns: map[string]int{},
	}

	for _, opt := range ops {
		opt(dec)
	}

	for i, name := range columns {
		dec.columns[strings.ToLower(name)] = i
	}

	return dec
}

// Decoder decodes ksqldb row to struct.
// Maps fields using `avro` tag.
type Decoder struct {
	row      Row
	columns  map[string]int
	HasJoins bool
}

// Decode decodes row into struct.
// Automatically resolves joins if table was provided as a constructor argument.
func (d *Decoder) Decode(v interface{}) error {
	return d.decode(v, "")
}

func (d *Decoder) decode(v interface{}, table string) error {
	var rv reflect.Value

	switch v := v.(type) {
	case reflect.Value:
		rv = v
	default:
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("non pointer reference for '%s'", reflect.TypeOf(v).String())
	}

	values := rv.Elem()

	if d.HasJoins && len(table) == 0 {
		field, ok := values.Type().FieldByName("TableName")

		if !ok {
			return ErrTableNameIsEmpty
		}

		table = field.Tag.Get("ksql")

		if len(table) == 0 {
			return ErrTableNameIsEmpty
		}
	}

	for i, field := range reflect.VisibleFields(values.Type()) {
		fld := NewField(table, values.Field(i), field)

		if !fld.IsValid() {
			continue
		}

		if d.HasJoins && fld.IsJoin() && fld.Value.Kind() == reflect.Ptr {
			for name := range d.columns {
				if strings.HasPrefix(name, fld.Join) {
					fld.NewValue()

					if err := d.decode(fld.Value, fld.Join); err != nil {
						return err
					}

					break
				}
			}

			continue
		}

		j, ok := d.columns[fld.Column]

		if !ok {
			continue
		}

		switch fld.Value.Kind() {
		case reflect.Slice:
			slc := reflect.MakeSlice(fld.Value.Type(), len(d.row.Row(j)), cap(d.row.Row(j)))

			for k := 0; k < slc.Len(); k++ {
				idx := slc.Index(k)

				if idx.Kind() == reflect.Ptr {
					if err := ConvPtr(d.row.Row(j)[k], idx); err != nil {
						return err
					}
				} else {
					d.row.Row(j).Value(k, idx)
				}
			}

			if slc.Len() > 0 {
				fld.Value.Set(slc)
			}
		case reflect.Ptr:
			if err := ConvPtr(d.row[j], fld.Value); err != nil {
				return err
			}
		default:
			d.row.Value(j, fld.Value)
		}
	}

	return nil
}

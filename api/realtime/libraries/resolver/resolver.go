// Package resolver is intended to simplify field related queries and mapping
// inside firehose api.
package resolver

import (
	"errors"
	"fmt"
	"reflect"
	"wikimedia-enterprise/api/realtime/submodules/log"
	"wikimedia-enterprise/api/realtime/submodules/schema"
)

var (
	// ErrNoKSQLTag happens if you are trying to resolve field with empty `ksql` tag in TableName field.
	ErrNoKSQLTag = errors.New("no ksql tag in TableName")

	// ErrNoTableName happens when to TableName field was specified for the struct.
	ErrNoTableName = errors.New("no TableName field found")

	// ErrSchemaTypeNotDefined happens when resolver is unaware of this schema type.
	ErrSchemaTypeNotDefined = errors.New("resolver is unaware of this schema type")
)

// NewResolvers returns multiple schema resolver.
// Example: {"article": *Resolver, "entity": *Resolver, "structured": *Resolver}
func NewResolvers(schemas map[string]interface{}, ops ...func(r *Resolver)) (Resolvers, error) {
	resolvers := make(Resolvers)

	for key, schema := range schemas {
		r, err := New(schema, ops...)

		if err != nil {
			return nil, err
		}

		resolvers[key] = r
	}

	return resolvers, nil
}

// Resolvers maintain multiple schema resolver.
// Example: {"article": *Resolver, "entity": *Resolver, "structured": *Resolver}
type Resolvers map[string]*Resolver

// GetSchema returns schema corresponding to the schema key.
func (rs Resolvers) GetSchema(key string) any {
	if _, ok := rs[key]; !ok {
		log.Error(ErrSchemaTypeNotDefined)
		return nil
	}

	// Update this to include structured, entity
	switch key {
	case schema.KeyTypeArticle:
		return new(schema.Article)
	}

	return nil
}

// New creates new resolver instance.
func New(v interface{}, ops ...func(r *Resolver)) (*Resolver, error) {
	r := &Resolver{
		Struct:  NewStruct(nil, nil),
		Fields:  map[string]*Field{},
		Structs: map[string]*Struct{},
		Slices:  map[string]*Slice{},
	}

	for _, opt := range ops {
		if opt != nil {
			opt(r)
		}
	}

	rv := reflect.ValueOf(v)

	if err := r.validate(rv); err != nil {
		return nil, err
	}

	if r.HasJoins {
		fld, ok := rv.Elem().Type().FieldByName("TableName")

		if !ok {
			return nil, ErrNoTableName
		}

		r.Struct.Table = fld.Tag.Get("ksql")

		if len(r.Struct.Table) == 0 {
			return nil, ErrNoKSQLTag
		}
	}

	if err := r.resolve(rv, r.Struct); err != nil {
		return nil, err
	}

	return r, nil
}

// Resolver gets a struct and resolves field mapping between API
// and SQL using `avro` and `ksqldb` tags.
type Resolver struct {
	Struct   *Struct
	Fields   map[string]*Field
	Structs  map[string]*Struct
	Slices   map[string]*Slice
	Keywords map[string]string
	HasJoins bool
}

// GetField returns a field struct for a particular name.
func (r *Resolver) GetField(name string) *Field {
	return r.Fields[name]
}

// HasField check if this resolver has this particular field.
func (r *Resolver) HasField(name string) bool {
	_, ok := r.Fields[name]
	return ok
}

// HasStruct check if this resolver has this particular struct.
func (r *Resolver) HasStruct(name string) bool {
	_, ok := r.Structs[name]
	return ok
}

// HasSlice check if this resolver has this particular slice.
func (r *Resolver) HasSlice(name string) bool {
	_, ok := r.Slices[name]
	return ok
}

func (r *Resolver) validate(rv reflect.Value) error {
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("non pointer reference for '%s'", rv.Type().String())
	}

	return nil
}

func (r *Resolver) resolve(rv reflect.Value, par *Struct) error {
	if err := r.validate(rv); err != nil {
		return err
	}

	for i, rf := range reflect.VisibleFields(rv.Type().Elem()) {
		if name := rf.Tag.Get("avro"); len(name) > 0 {
			fld := NewField(name, rf.Name, par, rv.Elem().Field(i), r.Keywords)
			r.Fields[fld.FullName] = fld

			if rf.Type.Kind() == reflect.Ptr && rf.Type.Elem().Kind() == reflect.Struct &&
				fld.Value.Type().String() != "*time.Time" {
				str := NewStruct(fld, par)

				par.Structs[fld.Name] = str
				r.Structs[fld.FullName] = str

				if fld.Value.IsNil() {
					fld.Value.Set(reflect.New(fld.Value.Type().Elem()))
				}

				if err := r.resolve(fld.Value, str); err != nil {
					return err
				}
			} else if rf.Type.Kind() == reflect.Slice && rf.Type.Elem().Kind() == reflect.Ptr &&
				rf.Type.Elem().Elem().Kind() == reflect.Struct && fld.Value.Type().String() != "*time.Time" {
				str := NewStruct(fld, par)
				slc := NewSlice(str)

				par.Slices[fld.Name] = slc
				r.Slices[fld.FullName] = slc

				val := reflect.MakeSlice(fld.Value.Type(), 1, 1).Index(0)

				if val.IsNil() {
					val.Set(reflect.New(val.Type().Elem()))
				}

				if err := r.resolve(val, str); err != nil {
					return err
				}
			} else {
				par.Fields[fld.Name] = fld
			}
		}
	}

	return nil
}

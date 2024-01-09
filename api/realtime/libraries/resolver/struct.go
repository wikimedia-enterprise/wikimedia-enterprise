package resolver

import (
	"fmt"
	"strings"
)

// NewStruct creates new struct instance.
func NewStruct(field *Field, par *Struct) *Struct {
	str := &Struct{
		Field:   field,
		Parent:  par,
		Fields:  map[string]*Field{},
		Slices:  map[string]*Slice{},
		Structs: map[string]*Struct{},
		IsRoot:  par != nil && par.Field == nil,
	}

	if par != nil {
		str.Table = par.Table
	}

	return str
}

// Struct holds all struct field relations for the resolver,
// including it's own children.
type Struct struct {
	Table   string
	Field   *Field
	Parent  *Struct
	Fields  map[string]*Field
	Slices  map[string]*Slice
	Structs map[string]*Struct
	IsRoot  bool
}

// GetSql get generated SQL statement for struct. Gives us
// the ability to dynamically resolve and specify fields.
// TODO - Add support for slices inside low level structs.
func (s *Struct) GetSql(filter Filter) string {
	sql := ""

	for name, fld := range s.Fields {
		if _, ok := s.Structs[name]; !ok && filter(fld) {
			sql += fmt.Sprintf("%s := %s, ", fld.Name, fld.Path)
		}
	}

	for _, str := range s.Structs {
		if filter(str.Field) {
			if val := str.GetSql(filter); len(val) > 0 {
				sql += fmt.Sprintf("%s := %s, ", str.Field.Name, val)
			}
		}
	}

	if len(sql) == 0 {
		return ""
	}

	if s.IsRoot {
		return fmt.Sprintf("STRUCT(%s) as %s", strings.TrimSuffix(sql, ", "), s.Field.Path)
	}

	return fmt.Sprintf("STRUCT(%s)", strings.TrimSuffix(sql, ", "))
}

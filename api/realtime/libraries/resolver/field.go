package resolver

import (
	"fmt"
	"reflect"
	"strings"
)

// NewField create new field resolver. For ksqldb handler, we need to handle some keywords passed via argument keywords.
func NewField(name string, goName string, par *Struct, value reflect.Value, keywords map[string]string) *Field {
	parent := ""

	if par.Field != nil {
		parent = par.Field.FullName
	}

	// this is a special case for direct use of keywords
	// in case when field belongs to a table this is not needed
	// cuz there is no direct use of the keyword
	if len(par.Table) == 0 {
		if rpl, ok := keywords[name]; ok {
			name = rpl
		}
	}

	fullName := strings.TrimPrefix(fmt.Sprintf("%s.%s", parent, name), ".")

	return &Field{
		Name:     name,
		GoName:   goName,
		FullName: fullName,
		Path:     strings.TrimPrefix(fmt.Sprintf("%s_%s", par.Table, strings.ReplaceAll(fullName, ".", "->")), "_"),
		Value:    value,
	}
}

// Field keeps track and maps ksqldb to API field names.
type Field struct {
	Name     string // from json (or avro tag) e.g., identifier
	GoName   string // struct field name e.g., Identifier
	FullName string // with dot notation e.g, version.editor.identifier
	Path     string // with -> notation e.g., version->editor->identifier
	Value    reflect.Value
}

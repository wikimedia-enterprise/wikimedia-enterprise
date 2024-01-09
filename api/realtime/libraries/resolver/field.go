package resolver

import (
	"fmt"
	"reflect"
	"strings"
)

// Keywords check if field is not keyword.
// TODO: Replace this logic, we need to make sure
// that braces and capital letters are being put
// on all field names, but that's complicated solution
// so will put that in backlog for now.
var Keywords = map[string]string{
	"namespace": "`NAMESPACE`",
}

// NewField create new field resolver.
func NewField(name string, par *Struct, value reflect.Value) *Field {
	parent := ""

	if par.Field != nil {
		parent = par.Field.FullName
	}

	// this is a special case for direct use of keywords
	// in case when field belongs to a table this is not needed
	// cuz there is no direct use of the keyword
	if len(par.Table) == 0 {
		if rpl, ok := Keywords[name]; ok {
			name = rpl
		}
	}

	fullName := strings.TrimPrefix(fmt.Sprintf("%s.%s", parent, name), ".")

	return &Field{
		Name:     name,
		FullName: fullName,
		Path:     strings.TrimPrefix(fmt.Sprintf("%s_%s", par.Table, strings.ReplaceAll(fullName, ".", "->")), "_"),
		Value:    value,
	}
}

// Field keeps track and maps ksqldb to API field names.
type Field struct {
	Name     string
	FullName string
	Path     string
	Value    reflect.Value
}

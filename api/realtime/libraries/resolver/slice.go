package resolver

import (
	"fmt"
	"strings"
)

// NewSlice creates new slice instance for the resolver.
func NewSlice(str *Struct) *Slice {
	return &Slice{
		Struct: str,
	}
}

// Slice slice representation for the resolver.
// Used for mapping golang slice representation to ksqldb query.
type Slice struct {
	Struct *Struct
}

// GetSql generates SQL query (for SELECT statement) for mapping struct
// fields to ksqldb query.
func (s *Slice) GetSql(filter Filter) string {
	sql := ""

	for name, fld := range s.Struct.Fields {
		if _, ok := s.Struct.Structs[name]; !ok && filter(fld) {
			sql += fmt.Sprintf("%s := %s, ", fld.Name, strings.ReplaceAll(fld.FullName, ".", "->"))
		}
	}

	if len(sql) == 0 {
		return ""
	}

	return fmt.Sprintf("TRANSFORM(%[1]s, x => STRUCT(%[2]s)) as %[3]s",
		s.Struct.Field.Path,
		strings.ReplaceAll(strings.TrimSuffix(sql, ", "), s.Struct.Field.Path, "x"),
		s.Struct.Field.Name,
	)
}

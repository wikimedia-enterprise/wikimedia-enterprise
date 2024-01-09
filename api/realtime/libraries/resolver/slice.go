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

	if s.Struct.IsRoot {
		return fmt.Sprintf("TRANSFORM(%[1]s, (%[2]s) => STRUCT(%[3]s)) as %[4]s",
			s.Struct.Field.Path,
			s.Struct.Field.Name,
			strings.TrimSuffix(sql, ", "),
			s.Struct.Field.Path)
	}

	return fmt.Sprintf("TRANSFORM(%[1]s, (%[2]s) => STRUCT(%[3]s))",
		s.Struct.Field.Path,
		s.Struct.Field.Name,
		strings.TrimSuffix(sql, ", "))
}

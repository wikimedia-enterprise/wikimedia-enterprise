package resolver

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// NewValueFilter returns a ValueFilter based on RequestFilter directives.
//
// This is achieved by writing annotations to the base tree.
// For example, with the following request filters:
//
//	fltr := []realtime.Filter{
//	  { Field: "is_part_of.identifier", Value: "enwiki"    },
//	  { Field: "event.type",           Value: "update"    },
//	  { Field: "version.noindex",      Value: false       },
//	  { Field: "namespace.identifier", Value: 0           },
//	}
//
// Annotations:
// Resolver.Struct (root)
// └─ Article (Field=nil)
//
//	├─ FilterFields: { is_part_of, event, version, namespace }
//	├─ ScalarFilters: { }
//	├─ Fields:
//	│     • name
//	│     • identifier
//	├─ Structs:
//	│
//	│   ├─ is_part_of
//	│   │    ├─ FilterFields: { identifier }
//	│   │    ├─ ScalarFilters: { identifier == "enwiki" }
//	│   │    └─ (no deeper Structs/Slices annotated)
//	│   │
//	│   ├─ event
//	│   │    ├─ FilterFields: { type }
//	│   │    ├─ ScalarFilters: { type == "update" }
//	│   │    └─ (no deeper Structs/Slices annotated)
//	│   │
//	│   ├─ version
//	│   │    ├─ FilterFields: { noindex }
//	│   │    ├─ ScalarFilters: { noindex == "false" }
//	│   │    ├─ Fields:
//	│   │    │    • identifier   (unfiltered)
//	│   │    │    • noindex      (filtered)
//	│   │    └─ Structs:
//	│   │         └─ editor
//	│   │             ├─ (FilterFields: none – because no filter on version.editor.*)
//	│   │             └─ Fields:
//	│   │                  • identifier
//	│   │
//	│   └─ namespace
//	│        ├─ FilterFields: { identifier }
//	│        ├─ ScalarFilters: { identifier == "0" }
//	│        └─ Fields:
//	│             • identifier   (filtered)
//	│
//	└─ Slices:
//	     └─ templates
//	          ├─ (FilterFields: none – no filter on templates.*)
//	          └─ Struct: Template
//	               └─ Fields:
//	                    • name
//	                    • url
func (r *Resolver) NewValueFilter(req []RequestFilter) ValueFilter {
	fmap := make(map[string]string, len(req))

	for _, f := range req {
		fmap[f.Field] = stringify(f.Value)
	}

	root := r.Struct.Copy()
	annotateFilter(root, fmap)

	return func(msg interface{}) bool {
		v := reflect.ValueOf(msg)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if !v.IsValid() {
			return false
		}

		return root.ShouldKeep(v)
	}
}

// ValueFilter receives a schema instance and tells whether or not to keep this instance as per user filtering directives.
type ValueFilter func(interface{}) bool

// RequestFilter model for the filter field and value passed as an API call argument.
type RequestFilter struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

// annotateFilter marks scalar filters and FilterFields on the cloned tree.
func annotateFilter(node *Struct, filters map[string]string) {
	for path, exp := range filters {
		parts := strings.Split(path, ".")
		annotateFilterPath(node, parts, exp)
	}
}

func annotateFilterPath(node *Struct, path []string, exp string) {
	if len(path) == 0 {
		return
	}
	name := path[0]

	if len(path) == 1 {
		// scalar filter
		if _, ok := node.Fields[name]; ok {
			node.ScalarFilters[name] = makeFilterFn(exp)
			node.FilterFields[name] = true
		}
		return
	}

	// nested
	if child, ok := node.Structs[name]; ok {
		node.FilterFields[name] = true
		annotateFilterPath(child, path[1:], exp)
	} else if sl, ok := node.Slices[name]; ok {
		node.FilterFields[name] = true
		annotateFilterPath(sl.Struct, path[1:], exp)
	}
}

// ShouldKeep applies all scalar filters and recurses into children.
func (s *Struct) ShouldKeep(v reflect.Value) bool {
	// scalar
	for name, fn := range s.ScalarFilters {
		fld, ok := s.Fields[name]
		if !ok {
			return false
		}
		fv := v.FieldByName(fld.GoName)
		if !fv.IsValid() || !fn(fv) {
			return false
		}
	}

	// nested structs
	for name, child := range s.Structs {
		if !s.FilterFields[name] {
			continue
		}
		fld := child.Field
		fv := v.FieldByName(fld.GoName)
		if !fv.IsValid() {
			return false
		}
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				return false
			}
			if !child.ShouldKeep(fv.Elem()) {
				return false
			}
		} else {
			if !child.ShouldKeep(fv) {
				return false
			}
		}
	}

	// slices
	for name, slc := range s.Slices {
		if !s.FilterFields[name] {
			continue
		}
		fld := slc.Struct.Field
		fv := v.FieldByName(fld.GoName)
		if !fv.IsValid() {
			return false
		}
		for i := 0; i < fv.Len(); i++ {
			e := fv.Index(i)
			if e.Kind() == reflect.Ptr {
				if e.IsNil() || !slc.Struct.ShouldKeep(e.Elem()) {
					return false
				}
			} else {
				if !slc.Struct.ShouldKeep(e) {
					return false
				}
			}
		}
	}

	return true
}

// makeFilterFn returns a func matching a reflect.Value to the expected string
func makeFilterFn(expected string) func(reflect.Value) bool {
	return func(v reflect.Value) bool {
		switch v.Kind() {
		case reflect.String:
			return v.String() == expected
		case reflect.Int, reflect.Int64, reflect.Int32:
			i, err := strconv.ParseInt(expected, 10, 64)
			if err != nil {
				return false
			}
			return v.Int() == i
		case reflect.Bool:
			b, err := strconv.ParseBool(expected)
			if err != nil {
				return false
			}
			return v.Bool() == b
		}
		return false
	}
}

// stringify normalizes any value to a string
func stringify(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case int, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(x).Int(), 10)
	default:
		return fmt.Sprint(x)
	}
}

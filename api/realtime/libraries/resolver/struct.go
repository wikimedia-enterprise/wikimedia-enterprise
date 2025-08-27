package resolver

import (
	"reflect"
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

// Struct represents one node in the schema tree, with child fields, nested structs, and slices.
// It also carries per-request annotations for modify (prune) and filter.
//
// Tree structure for resolver.Struct for the following simplistic schema.
//
//	type Article struct {
//		Name                   string           `json:"name,omitempty" avro:"name"`
//		Identifier             int              `json:"identifier,omitempty" avro:"identifier"`
//		Version                *Version         `json:"version,omitempty" avro:"version"`
//		Templates              []*Template      `json:"templates,omitempty" avro:"templates"`
//	}
//
//	type Version struct {
//		Identifier          int              `json:"identifier,omitempty" avro:"identifier"`
//		Editor              *Editor          `json:"editor,omitempty" avro:"editor"`
//		Noindex             bool             `json:"noindex,omitempty" avro:"noindex"`
//	}
//
//	type Editor struct {
//		Identifier        int        `json:"identifier,omitempty" avro:"identifier"`
//	}
//
//	type Template struct {
//		Name string `json:"name,omitempty" avro:"name"`
//		URL  string `json:"url,omitempty" avro:"url"`
//	}
//
// Resolver.Struct (root)
// └─ Article (Field=nil)
//
//	├─ Fields:
//	│   ├─ name         (GoName: Name)
//	│   └─ identifier   (GoName: Identifier)
//	├─ Structs:
//	│   └─ version      (Field.FullName: "version")
//	│       ├─ Fields:
//	│       │   ├─ identifier   (GoName: Identifier)
//	│       │   └─ noindex      (GoName: Noindex)
//	│       └─ Structs:
//	│           └─ editor    (Field.FullName: "version.editor")
//	│               └─ Fields:
//	│                   └─ identifier   (GoName: Identifier)
//	└─ Slices:
//	    └─ templates    (Field.FullName: "templates")
//	        └─ Struct: Template
//	            └─ Fields:
//	                ├─ name      (GoName: Name)
//	                └─ url       (GoName: URL)
type Struct struct {
	// static metadata
	Table   string
	Field   *Field
	Parent  *Struct
	Fields  map[string]*Field
	Slices  map[string]*Slice
	Structs map[string]*Struct
	IsRoot  bool

	// per-request modify annotations
	DiscardSelf bool            // if true, zero entire struct
	KeepFields  map[string]bool // children names to keep
	KeepAll     bool            // true if all children kept

	// per-request filter annotations
	ScalarFilters map[string]func(reflect.Value) bool // match functions for scalar fields
	FilterFields  map[string]bool                     // children names to apply filter
}

// Copy makes a deep copy of the schema tree, clearing all per-request annotations.
func (s *Struct) Copy() *Struct {
	out := &Struct{
		Table:         s.Table,
		Parent:        nil,
		Field:         s.Field,
		Fields:        s.Fields,
		Structs:       make(map[string]*Struct, len(s.Structs)),
		Slices:        make(map[string]*Slice, len(s.Slices)),
		IsRoot:        s.IsRoot,
		DiscardSelf:   false,
		KeepFields:    make(map[string]bool),
		ScalarFilters: make(map[string]func(reflect.Value) bool),
		FilterFields:  make(map[string]bool),
	}

	// clone nested structs
	for name, child := range s.Structs {
		cl := child.Copy()
		out.Structs[name] = cl
	}

	// clone slices
	for name, sl := range s.Slices {
		cl := sl.Struct.Copy()
		out.Slices[name] = &Slice{Struct: cl}
	}

	return out
}

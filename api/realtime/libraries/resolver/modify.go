package resolver

import (
	"reflect"
	"strings"
)

// NewModifier returns a new modifier that modifies (prunes out) unrequested fields in-place.
//
// This is achieved by annotating the base tree with per-request annotations.
// Example modify annotation for the request fields NewModifier([]string{"name", "templates.*", "version.tags.*", "categories.url"})
//
// Resolver.Struct (root)
// └─ Article (Field=nil)
//
//	├─ DiscardSelf: false
//	├─ KeepFields: map[string]bool{
//	│     "name": true,
//	│     "templates": true,
//	│     "version": true,
//	│     "categories": true,
//	│   }
//	├─ Fields:
//	│     name         (GoName: Name)         ← kept
//	│     identifier   (GoName: Identifier)   ← dropped
//	│     abstract     (GoName: Abstract)     ← dropped
//	│     … (all others dropped)
//	├─ Structs:
//	│
//	│   "version" → &Struct{
//	│       DiscardSelf: false
//	│       KeepFields: map[string]bool{
//	│         "tags": true,
//	│       }
//	│       Fields:
//	│         identifier   (GoName: Identifier) ← dropped
//	│         tags         (GoName: Tags)       ← kept
//	│         noindex      (GoName: Noindex)    ← dropped
//	│         … (other version.* fields dropped)
//	│       Structs: {}    // no nested structs under tags
//	│       Slices:  {}
//	│   }
//	└─ Slices:
//	    │
//	    ├─ "templates" → &Slice{
//	    │     Struct: &Struct{
//	    │       DiscardSelf: false
//	    │       KeepFields: map[string]bool{
//	    │         "name": true,
//	    │         "url":  true,
//	    │       }
//	    │       Fields:
//	    │         name  (kept)
//	    │         url   (kept)
//	    │       Structs: {}
//	    │       Slices:  {}
//	    │     }
//	    │   }
//	    │
//	    └─ "categories" → &Slice{
//	          Struct: &Struct{
//	            DiscardSelf: false
//	            KeepFields: map[string]bool{
//	              "url": true,
//	            }
//	            Fields:
//	              name  (dropped)
//	              url   (kept)
//	            Structs: {}
//	            Slices:  {}
//	          }
//	        }
func (r *Resolver) NewModifier(fields []string) Modifier {
	paths := expandPaths(fields)

	root := r.Struct.Copy()
	annotateModify(root, paths)

	return func(msg interface{}) {
		v := reflect.ValueOf(msg)
		if v.Kind() != reflect.Ptr || v.IsNil() {
			return
		}
		v = v.Elem()
		root.ApplyModify(v) // mutates in‐place
	}
}

// Modifier receives a schema instance and modifies it as per user field selection directives.
// It modifies the article by zeroing out the fields that are not requested.
type Modifier func(interface{})

// annotateModify marks the cloned tree with DiscardSelf / KeepFields based on requested paths.
func annotateModify(node *Struct, paths [][]string) {
	// wildcard or empty path => keep the subtree
	for _, p := range paths {
		if len(p) == 0 || (len(p) == 1 && p[0] == "*") {
			node.DiscardSelf = false
			node.KeepAll = true

			for name := range node.Fields {
				node.KeepFields[name] = true
			}

			for name, child := range node.Structs {
				node.KeepFields[name] = true
				annotateModify(child, [][]string{{"*"}})
			}

			for name, sl := range node.Slices {
				node.KeepFields[name] = true
				annotateModify(sl.Struct, [][]string{{"*"}})
			}

			return
		}
	}

	// no wildcard: only keep branches that match a path
	for _, p := range paths {
		if len(p) > 0 {
			name := p[0]
			sub := filterSubPaths(paths, name)
			node.KeepFields[name] = true

			if child, ok := node.Structs[name]; ok {
				annotateModify(child, sub)
			} else if sl, ok := node.Slices[name]; ok {
				annotateModify(sl.Struct, sub)
			}
		}
	}
}

// ApplyModify walks the struct value and zeros out unkept branches.
func (s *Struct) ApplyModify(v reflect.Value) {
	if s.DiscardSelf {
		v.Set(reflect.Zero(v.Type()))
		return
	}

	if s.KeepAll {
		return
	}

	// scalar fields
	for name, fld := range s.Fields {
		if !s.KeepFields[name] {
			v.FieldByName(fld.GoName).Set(reflect.Zero(fld.Value.Type()))
		}
	}

	// nested structs
	for name, child := range s.Structs {
		fv := v.FieldByName(child.Field.GoName)
		if s.KeepFields[name] {
			if fv.Kind() == reflect.Ptr && !fv.IsNil() {
				child.ApplyModify(fv.Elem())
			}
		} else {
			fv.Set(reflect.Zero(fv.Type()))
		}
	}

	// slices of pointers
	for name, slc := range s.Slices {
		fv := v.FieldByName(slc.Struct.Field.GoName)
		if s.KeepFields[name] {
			for i := 0; i < fv.Len(); i++ {
				e := fv.Index(i)
				if e.Kind() == reflect.Ptr && !e.IsNil() {
					slc.Struct.ApplyModify(e.Elem())
				}
			}
		} else {
			fv.Set(reflect.Zero(fv.Type()))
		}
	}
}

// expandPaths splits dot-notation into string slices.
// Example: ["templates.url","name"] → [][]string{ {"templates","url"}, {"name"} }
func expandPaths(fields []string) [][]string {
	out := make([][]string, len(fields))
	for i, f := range fields {
		out[i] = strings.Split(f, ".")
	}

	return out
}

// filterSubPaths pulls out only those paths whose first segment == name,
// then strips off that segment.
func filterSubPaths(paths [][]string, name string) [][]string {
	var out [][]string

	for _, p := range paths {
		if len(p) > 0 && p[0] == name {
			out = append(out, p[1:])
		}
	}

	return out
}

package resolver_test

import (
	"strings"
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type structTestSuite struct {
	suite.Suite
	fr          resolver.Filter
	contains    []string // should be an array of strings cuz we can't predict the order of Fields and Structs map
	notContains []string
	fld         *resolver.Field
	par         *resolver.Struct
	flds        map[string]*resolver.Field
	strs        map[string]*resolver.Struct
}

func (s *structTestSuite) TestNew() {
	str := resolver.NewStruct(s.fld, s.par)
	s.Assert().NotNil(str)
	s.Assert().Equal(s.fld, str.Field)
	s.Assert().Equal(s.par != nil && s.par.Field == nil, str.IsRoot)

	if s.par != nil {
		s.Assert().Equal(s.par.Table, str.Table)
	}
}

func (s *structTestSuite) TestGetSql() {
	str := resolver.NewStruct(s.fld, s.par)

	for name, fld := range s.flds {
		str.Fields[name] = fld
	}

	for name, st := range s.strs {
		str.Structs[name] = st
	}

	sql := str.GetSql(s.fr)

	for _, part := range s.contains {
		s.Assert().Contains(sql, part)
	}

	for _, part := range s.notContains {
		s.Assert().NotContains(sql, part)
	}

	if len(s.contains) == 0 && len(s.notContains) == 0 {
		s.Assert().Equal("", sql)
	}
}

func TestStruct(t *testing.T) {
	for _, testCase := range []*structTestSuite{
		{
			fld: &resolver.Field{},
		},
		{
			fld: &resolver.Field{},
			par: &resolver.Struct{
				Table: "table",
				Field: &resolver.Field{},
			},
		},
		{
			fld: &resolver.Field{},
			par: &resolver.Struct{
				Table: "table",
			},
		},
		{
			contains: []string{"STRUCT(identifier := versions->identifier, editor := STRUCT(identifier := versions->editor->identifier)) as version"},
			flds: map[string]*resolver.Field{
				"identifier": {
					Name: "identifier",
					Path: "versions->identifier",
				},
			},
			strs: map[string]*resolver.Struct{
				"editor": {
					Field: &resolver.Field{
						Name: "editor",
					},
					Structs: map[string]*resolver.Struct{},
					Fields: map[string]*resolver.Field{
						"identifier": {
							Name: "identifier",
							Path: "versions->editor->identifier",
						},
					},
					IsRoot: false,
				},
			},
			fld: &resolver.Field{
				Path: "version",
			},
			par: &resolver.Struct{
				Table: "versions",
			},
			fr: func(fld *resolver.Field) bool {
				return true
			},
		},
		{
			contains: []string{"STRUCT(identifier := versions->identifier, editor := STRUCT(identifier := versions->editor->identifier))"},
			flds: map[string]*resolver.Field{
				"identifier": {
					Name: "identifier",
					Path: "versions->identifier",
				},
			},
			strs: map[string]*resolver.Struct{
				"editor": {
					Field: &resolver.Field{
						Name: "editor",
					},
					Structs: map[string]*resolver.Struct{},
					Fields: map[string]*resolver.Field{
						"identifier": {
							Name: "identifier",
							Path: "versions->editor->identifier",
						},
					},
				},
			},
			fld: &resolver.Field{
				Path: "version",
			},
			par: &resolver.Struct{
				Table: "versions",
				Field: &resolver.Field{},
			},
			fr: func(fld *resolver.Field) bool {
				return true
			},
		},
		{
			contains:    []string{"STRUCT(editor := STRUCT(identifier := versions->editor->identifier))"},
			notContains: []string{"identifier := versions->identifier"},
			flds: map[string]*resolver.Field{
				"identifier": {
					Name: "identifier",
					Path: "versions->identifier",
				},
			},
			strs: map[string]*resolver.Struct{
				"editor": {
					Field: &resolver.Field{
						Name: "editor",
					},
					Structs: map[string]*resolver.Struct{},
					Fields: map[string]*resolver.Field{
						"identifier": {
							Name: "identifier",
							Path: "versions->editor->identifier",
						},
					},
				},
			},
			fld: &resolver.Field{
				Path: "version",
			},
			par: &resolver.Struct{
				Table: "versions",
				Field: &resolver.Field{},
			},
			fr: func(fld *resolver.Field) bool {
				return fld.Path != "versions->identifier"
			},
		},
		{
			contains:    []string{"STRUCT(identifier := versions->identifier)"},
			notContains: []string{"editor := STRUCT(identifier := versions->editor->identifier)"},
			flds: map[string]*resolver.Field{
				"identifier": {
					Name: "identifier",
					Path: "versions->identifier",
				},
			},
			strs: map[string]*resolver.Struct{
				"editor": {
					Field: &resolver.Field{
						Name: "editor",
					},
					Structs: map[string]*resolver.Struct{},
					Fields: map[string]*resolver.Field{
						"identifier": {
							Name: "identifier",
							Path: "versions->editor->identifier",
						},
					},
				},
			},
			fld: &resolver.Field{
				Path: "version",
			},
			par: &resolver.Struct{
				Table: "versions",
				Field: &resolver.Field{},
			},
			fr: func(fld *resolver.Field) bool {
				return !strings.HasPrefix(fld.Path, "versions->editor")
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type sliceTestSuite struct {
	suite.Suite
	contains    []string // should be an array of strings cuz we can't predict the order of Fields map
	notContains []string
	str         *resolver.Struct
	fr          resolver.Filter
}

func (s *sliceTestSuite) TestNewSlice() {
	slc := resolver.NewSlice(s.str)
	s.Assert().NotNil(slc)
	s.Assert().Equal(slc.Struct, s.str)
}

func (s *sliceTestSuite) TestGetSql() {
	sql := resolver.NewSlice(s.str).GetSql(s.fr)

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

func TestSlice(t *testing.T) {
	for _, testCase := range []*sliceTestSuite{
		{
			contains: []string{},
			str: &resolver.Struct{
				Field: &resolver.Field{},
			},
			fr: func(fld *resolver.Field) bool {
				return true
			},
		},
		{
			contains: []string{
				"TRANSFORM(articles_protection, (protection) =>",
				"name := protection->name",
				"type := protection->type",
				"))",
			},
			str: &resolver.Struct{
				Structs: map[string]*resolver.Struct{},
				Field: &resolver.Field{
					Path: "articles_protection",
					Name: "protection",
				},
				Fields: map[string]*resolver.Field{
					"name": {
						Name:     "name",
						FullName: "protection->name",
					},
					"protection": {
						Name:     "type",
						FullName: "protection->type",
					},
				},
			},
			fr: func(fld *resolver.Field) bool {
				return true
			},
		},
		{
			contains: []string{
				"TRANSFORM(articles_protection, (protection) => STRUCT(type := protection->type))",
			},
			notContains: []string{
				"name := protection->name",
			},
			str: &resolver.Struct{
				Structs: map[string]*resolver.Struct{},
				Field: &resolver.Field{
					Path: "articles_protection",
					Name: "protection",
				},
				Fields: map[string]*resolver.Field{
					"name": {
						Name:     "name",
						FullName: "protection->name",
					},
					"protection": {
						Name:     "type",
						FullName: "protection->type",
					},
				},
			},
			fr: func(fld *resolver.Field) bool {
				return fld.FullName != "protection->name"
			},
		},
		{
			contains: []string{
				"TRANSFORM(articles_protection, (protection) =>",
				"name := protection->name",
				"type := protection->type",
				")) as articles_protection",
			},
			str: &resolver.Struct{
				Structs: map[string]*resolver.Struct{},
				IsRoot:  true,
				Field: &resolver.Field{
					Path: "articles_protection",
					Name: "protection",
				},
				Fields: map[string]*resolver.Field{
					"name": {
						Name:     "name",
						FullName: "protection->name",
					},
					"protection": {
						Name:     "type",
						FullName: "protection->type",
					},
				},
			},
			fr: func(fld *resolver.Field) bool {
				return true
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

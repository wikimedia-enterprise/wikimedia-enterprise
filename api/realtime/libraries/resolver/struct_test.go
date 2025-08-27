package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type structTestSuite struct {
	suite.Suite
	fld *resolver.Field
	par *resolver.Struct
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

func (s *structTestSuite) TestCopy() {
	str := resolver.NewStruct(s.fld, s.par)
	copy := str.Copy()

	s.Assert().NotNil(copy)
	s.Assert().Equal(s.fld, copy.Field)
	s.Assert().Equal(s.par != nil && s.par.Field == nil, copy.IsRoot)
	s.Assert().Equal(len(str.Fields), len(copy.Fields))
	s.Assert().Equal(len(str.Structs), len(copy.Structs))
	s.Assert().Equal(len(str.Slices), len(copy.Slices))
	s.Assert().False(copy.DiscardSelf)
	s.Assert().Empty(copy.KeepFields)
	s.Assert().Empty(copy.FilterFields)
	s.Assert().Empty(copy.ScalarFilters)

	if s.par != nil {
		s.Assert().Equal(s.par.Table, copy.Table)
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
			fld: &resolver.Field{},
			par: &resolver.Struct{
				Table: "table",
				Fields: map[string]*resolver.Field{
					"identifier": {
						Name: "identifier",
					},
				},
				Structs: map[string]*resolver.Struct{
					"version": {
						Fields: map[string]*resolver.Field{
							"identifier": {
								Name: "identifier",
								Path: "versions->identifier",
							},
						},
					},
				},
				Slices: map[string]*resolver.Slice{
					"categories": {
						Struct: &resolver.Struct{
							Fields: map[string]*resolver.Field{},
						},
					},
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

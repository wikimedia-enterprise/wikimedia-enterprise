package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type sliceTestSuite struct {
	suite.Suite
	str *resolver.Struct
}

func (s *sliceTestSuite) TestNewSlice() {
	slc := resolver.NewSlice(s.str)
	s.Assert().NotNil(slc)
	s.Assert().Equal(slc.Struct, s.str)
}

func TestSlice(t *testing.T) {
	for _, testCase := range []*sliceTestSuite{
		{
			str: &resolver.Struct{
				Field: &resolver.Field{},
			},
		},
		{
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
		},
	} {
		suite.Run(t, testCase)
	}
}

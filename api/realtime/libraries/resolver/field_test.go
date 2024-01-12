package resolver_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

	"github.com/stretchr/testify/suite"
)

type fieldTestSuite struct {
	suite.Suite
	name  string
	value reflect.Value
	par   *resolver.Struct
}

func (s *fieldTestSuite) TestNewField() {
	fld := resolver.NewField(s.name, s.par, s.value)

	s.Assert().NotNil(fld)
	s.Assert().Equal(s.name, fld.Name)
	s.Assert().Equal(s.value, fld.Value)

	if s.par != nil && s.par.Field != nil {
		s.Assert().Equal(fmt.Sprintf("%s_%s->%s", s.par.Table, s.par.Field.FullName, s.name), fld.Path)
		s.Assert().Equal(fmt.Sprintf("%s.%s", s.par.Field.FullName, s.name), fld.FullName)
	} else {
		s.Assert().Equal(strings.TrimPrefix(fmt.Sprintf("%s_%s", s.par.Table, s.name), "_"), fld.Path)
		s.Assert().Equal(s.name, fld.FullName)
	}
}

func TestField(t *testing.T) {
	for _, testCase := range []*fieldTestSuite{
		{
			name:  "test",
			value: reflect.Value{},
			par: &resolver.Struct{
				Table: "main",
				Field: &resolver.Field{
					FullName: "secondary",
				},
			},
		},
		{
			name:  "test",
			value: reflect.Value{},
			par:   &resolver.Struct{},
		},
	} {
		suite.Run(t, testCase)
	}
}

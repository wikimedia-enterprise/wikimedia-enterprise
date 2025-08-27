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
	name     string
	goName   string
	value    reflect.Value
	par      *resolver.Struct
	keywords map[string]string
}

func (s *fieldTestSuite) TestNewField() {
	fld := resolver.NewField(s.name, s.goName, s.par, s.value, s.keywords)

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
			name:   "test",
			goName: "Test",
			value:  reflect.Value{},
			par: &resolver.Struct{
				Table: "main",
				Field: &resolver.Field{
					FullName: "secondary",
				},
			},
		},
		{
			name:   "test",
			goName: "Test",
			value:  reflect.Value{},
			par:    &resolver.Struct{},
		},
		{
			name:  "test",
			value: reflect.Value{},
			keywords: map[string]string{
				"keyword": "`keyword`",
			},
			par: &resolver.Struct{
				Table: "main",
				Field: &resolver.Field{
					FullName: "secondary",
				},
			},
		},
		{
			name: "test",
			keywords: map[string]string{
				"keyword": "`keyword`",
			},
			value: reflect.Value{},
			par:   &resolver.Struct{},
		},
	} {
		suite.Run(t, testCase)
	}
}

package ksqldb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type fieldTestSuite struct {
	suite.Suite
	table   string
	val     interface{}
	isValid bool
	isJoin  bool
	field   *Field
}

func (s *fieldTestSuite) SetupSuite() {
	value := reflect.ValueOf(s.val).Elem()
	field, _ := value.Type().FieldByName("Test")
	s.field = NewField(s.table, value.FieldByName("Test"), field)
}

func (s *fieldTestSuite) TestNewField() {
	s.Assert().Equal(fmt.Sprintf("%s_%s", s.table, s.field.Name), s.field.Column)
}

func (s *fieldTestSuite) TestIsValid() {
	s.Assert().Equal(s.isValid, s.field.IsValid())
}

func (s *fieldTestSuite) TestIsJoin() {
	s.Assert().Equal(s.isJoin, s.field.IsJoin())
}

func (s *fieldTestSuite) TestNewValue() {
	if s.field.Value.IsNil() {
		s.field.NewValue()
		s.Assert().False(s.field.Value.IsNil())
	}
}

func TestField(t *testing.T) {
	for _, testCase := range []*fieldTestSuite{
		{
			table: "articles",
			val: &struct {
				Test *struct{} `avro:"test" ksql:"versions"`
			}{},
			isValid: true,
			isJoin:  true,
		},
		{
			table: "articles",
			val: &struct {
				Test *struct{} `ksql:"versions"`
			}{},
			isValid: false,
			isJoin:  true,
		},
		{
			table: "articles",
			val: &struct {
				Test *struct{} `avro:"test"`
			}{},
			isValid: true,
			isJoin:  false,
		},
		{
			table: "articles",
			val: &struct {
				Test *struct{}
			}{},
			isValid: false,
			isJoin:  false,
		},
	} {
		suite.Run(t, testCase)
	}
}

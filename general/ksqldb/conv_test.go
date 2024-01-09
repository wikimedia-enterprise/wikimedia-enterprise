package ksqldb

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type convStringTestSuite struct {
	suite.Suite
	val    interface{}
	result string
}

// TestConvString method tests ConvString correct behaviour.
func (s *convStringTestSuite) TestConvString() {
	s.Assert().Equal(s.result, ConvString(s.val))
}

// TestConvString test runner.
func TestConvString(t *testing.T) {
	for _, testCase := range []*convStringTestSuite{
		{
			val:    "Hello world",
			result: "Hello world",
		},
		{
			val:    10,
			result: "",
		},
		{
			val:    3.16,
			result: "",
		},
		{
			val:    map[string]string{"test": "here"},
			result: "",
		},
	} {
		suite.Run(t, testCase)
	}
}

type convIntTestSuite struct {
	suite.Suite
	val    interface{}
	result int
}

// TestConvInt method tests ConvInt correct behaviour.
func (s *convIntTestSuite) TestConvInt() {
	s.Assert().Equal(s.result, ConvInt(s.val))
}

// TestConvInt test runner.
func TestConvInt(t *testing.T) {
	for _, testCase := range []*convIntTestSuite{
		{
			val:    10,
			result: 10,
		},
		{
			val:    "",
			result: 0,
		},
		{
			val:    3.16,
			result: 3,
		},
	} {
		suite.Run(t, testCase)
	}
}

type convTimeTestSuite struct {
	suite.Suite
	val    interface{}
	result time.Time
}

// TestConvTime method tests ConvTime correct behaviour.
func (s *convTimeTestSuite) TestConvTime() {
	s.Assert().Equal(s.result, *ConvTime(s.val))
}

// TestConvTime test runner.
func TestConvTime(t *testing.T) {
	for _, testCase := range []*convTimeTestSuite{
		{
			val:    "asdfad",
			result: time.UnixMicro(0),
		},
		{
			val:    0,
			result: time.UnixMicro(0),
		},
		{
			val:    float64(1636371200000000),
			result: time.UnixMicro(int64(1636371200000000)),
		},
	} {
		suite.Run(t, testCase)
	}
}

type convBoolTesSuite struct {
	suite.Suite
	val    interface{}
	result bool
}

// TestConvBool method tests ConvBool correct behaviour.
func (s *convBoolTesSuite) TestConvBool() {
	s.Assert().Equal(s.result, ConvBool(s.val))
}

// TestConvBool test runner.
func TestConvBool(t *testing.T) {
	for _, testCase := range []*convBoolTesSuite{
		{
			val:    false,
			result: false,
		},
		{
			val:    true,
			result: true,
		},
		{
			val:    "true",
			result: false,
		},
		{
			val:    1,
			result: false,
		},
		{
			val:    3.16,
			result: false,
		},
	} {
		suite.Run(t, testCase)
	}
}

type convFloat64TestSuite struct {
	suite.Suite
	val    interface{}
	result float64
}

// TestConvFloat64  method tests ConvFloat64 correct behaviour.
func (s *convFloat64TestSuite) TestConvFloat64() {
	s.Assert().Equal(s.result, ConvFloat64(s.val))
}

// TestConvFloat64 test runner.
func TestConvFloat64(t *testing.T) {
	for _, testCase := range []*convFloat64TestSuite{
		{
			val:    float32(3.00),
			result: float64(3.00),
		},
		{
			val:    float64(1),
			result: float64(1),
		},
		{
			val:    "1",
			result: float64(0),
		},
		{
			val:    map[string]string{"hello": "world"},
			result: float64(0),
		},
	} {
		suite.Run(t, testCase)
	}
}

type convMapTestSuite struct {
	suite.Suite
	val    interface{}
	result Map
}

// TestConvMap method tests ConvMap correct behaviour.
func (s *convMapTestSuite) TestConvMap() {
	s.Assert().Equal(s.result, ConvMap(s.val))
}

// TestConvMap test runner.
func TestConvMap(t *testing.T) {
	for _, testCase := range []*convMapTestSuite{
		{
			val:    map[string]interface{}{"hello": "world"},
			result: Map{"hello": "world"},
		},
		{
			val:    "",
			result: Map{},
		},
		{
			val:    map[string]interface{}{"number": 18},
			result: Map{"number": 18},
		},
	} {
		suite.Run(t, testCase)
	}
}

type convRowTestSuite struct {
	suite.Suite
	val    interface{}
	result Row
}

// TestConvRow method tests ConvRow correct behaviour.
func (s *convRowTestSuite) TestConvRow() {
	s.Assert().Equal(s.result, ConvRow(s.val))
}

// TestConvRow test runner.
func TestConvRow(t *testing.T) {
	for _, testCase := range []*convRowTestSuite{
		{
			val:    map[string]interface{}{"hello": "world"},
			result: Row{},
		},
		{
			val:    []interface{}{map[string]interface{}{"hello": "world"}, 1},
			result: Row{map[string]interface{}{"hello": "world"}, 1},
		},
		{
			val:    []interface{}{"test", "1", 2, 3.16},
			result: Row{"test", "1", 2, 3.16},
		},
	} {
		suite.Run(t, testCase)
	}
}

type convPtrTestSuite struct {
	suite.Suite
	val    interface{}
	result interface{}
}

func (s *convPtrTestSuite) TestConvPtr() {
	rv := reflect.New(reflect.TypeOf(s.result))
	s.Assert().NoError(ConvPtr(s.val, rv.Elem()))

	switch v := rv.Elem().Interface().(type) {
	case *time.Time:
		s.Assert().Equal(s.result.(*time.Time).Format(time.RFC1123), v.Format(time.RFC1123))
	default:
		s.Assert().Equal(s.result, rv.Elem().Interface())
	}
}

// TestConvPtr method tests ConvPtr correct behaviour.
func TestConvPtr(t *testing.T) {
	dateNow := time.Now()
	user := struct {
		ID       int
		FullName string `avro:"full_name"`
	}{
		10,
		"John Doe",
	}

	for _, testCase := range []*convPtrTestSuite{
		{
			val:    float64(dateNow.UnixMicro()),
			result: &dateNow,
		},
		{
			val: map[string]interface{}{
				"id":        user.ID,
				"full_name": user.FullName,
			},
			result: &user,
		},
	} {
		suite.Run(t, testCase)
	}
}

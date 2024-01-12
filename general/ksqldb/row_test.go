package ksqldb

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type rowTestSuite struct {
	suite.Suite
}

// TestRow tests retriving diffirent data methods from a row.
func (s *rowTestSuite) TestRow() {
	rows := Row{
		"test",
		2,
		3.16,
		true,
		float64(1636371200000000),
		map[string]interface{}{"string": "value"},
		[]interface{}{"test"},
	}

	s.Assert().Equal("test", rows.String(0))
	s.Assert().Equal(2, rows.Int(1))
	s.Assert().Equal(3.16, rows.Float64(2))
	s.Assert().Equal(true, rows.Bool(3))
	s.Assert().Equal(time.UnixMicro(int64(1636371200000000)), *rows.Time(4))
	s.Assert().Equal(Map{"string": "value"}, rows.Map(5))
	s.Assert().Equal(Row{"test"}, rows.Row(6))

	length := 0
	s.Assert().NoError(rows.ForEach(func(r Row, i int) error {
		s.Assert().Equal(rows[i], r[i])
		length++
		return nil
	}))

	s.Assert().Equal(len(rows), length)

	for i, v := range []interface{}{
		"test",
		2,
		3.16,
		true,
		float64(1636371200000000),
	} {
		ref := reflect.New(reflect.TypeOf(v))
		rows.Value(i, ref.Elem())
		s.Assert().Equal(v, ref.Elem().Interface())
	}
}

func TestRow(t *testing.T) {
	suite.Run(t, new(rowTestSuite))
}

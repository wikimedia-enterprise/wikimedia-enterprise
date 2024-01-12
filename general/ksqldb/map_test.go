package ksqldb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type mapTestSuite struct {
	suite.Suite
}

// TestMap tests methods of retriving diffirent data types from a map.
func (s *mapTestSuite) TestMap() {
	mp := &Map{
		"STRING": "value",
		"INT":    10,
		"FLOAT":  1.0,
		"TIME":   float64(1636371200000000),
		"BOOL":   true,
		"ROW":    []interface{}{"value"},
		"MAP":    map[string]interface{}{"string": "value"},
	}

	s.Assert().Equal("value", mp.String("string"))
	s.Assert().Equal(10, mp.Int("int"))
	s.Assert().Equal(1.0, mp.Float64("float"))
	s.Assert().Equal(time.UnixMicro(int64(1636371200000000)), *mp.Time("time"))
	s.Assert().Equal(true, mp.Bool("bool"))
	s.Assert().Equal(Row(Row{"value"}), mp.Row("row"))
	s.Assert().Equal(Map(Map{"string": "value"}), mp.Map("map"))
}

func TestMapString(t *testing.T) {
	suite.Run(t, new(mapTestSuite))
}

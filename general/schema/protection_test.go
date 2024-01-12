package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type protectionTestSuite struct {
	suite.Suite
	protection *Protection
}

func (s *protectionTestSuite) SetupTest() {
	s.protection = &Protection{
		Type:   "general",
		Expiry: "infinite",
		Level:  "max",
	}
}

func (s *protectionTestSuite) TestNewProtectionSchema() {
	sch, err := NewProtectionSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.protection)
	s.Assert().NoError(err)

	protection := new(Protection)
	s.Assert().NoError(avro.Unmarshal(sch, data, protection))
	s.Assert().Equal(s.protection.Type, protection.Type)
	s.Assert().Equal(s.protection.Expiry, protection.Expiry)
	s.Assert().Equal(s.protection.Level, protection.Level)
}

func TestProtection(t *testing.T) {
	suite.Run(t, new(protectionTestSuite))
}

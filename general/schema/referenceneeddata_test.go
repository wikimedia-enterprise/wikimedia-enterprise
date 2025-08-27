package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type referenceNeedTestSuite struct {
	suite.Suite
	referenceNeed *ReferenceNeedData
}

func (s *referenceNeedTestSuite) SetupTest() {
	s.referenceNeed = &ReferenceNeedData{
		ReferenceNeedScore: 0.85,
	}
}

func (s *referenceNeedTestSuite) TestNewReferenceNeedSchema() {
	sch, err := NewReferenceNeedData()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.referenceNeed)
	s.Assert().NoError(err)

	decoded := new(ReferenceNeedData)
	s.Assert().NoError(avro.Unmarshal(sch, data, decoded))
	s.Assert().Equal(s.referenceNeed.ReferenceNeedScore, decoded.ReferenceNeedScore)
}

func TestReferenceNeed(t *testing.T) {
	suite.Run(t, new(referenceNeedTestSuite))
}

package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type referenceDetailsTestSuite struct {
	suite.Suite
	referenceDetails *ReferenceDetails
}

func (s *referenceDetailsTestSuite) SetupTest() {
	psLabelLocal := "reliable"
	psLabelEnwiki := "mostly_reliable"
	s.referenceDetails = &ReferenceDetails{
		URL:        "https://example.com",
		DomainName: "example.com",
		DomainMetadata: &DomainMetadata{
			PsLabelLocal:  &psLabelLocal,
			PsLabelEnwiki: &psLabelEnwiki,
			SurvivalRatio: 0.75,
			PageCount:     120,
			EditorsCount:  45,
		},
	}
}

func (s *referenceDetailsTestSuite) TestNewReferenceDetailsSchema() {
	sch, err := NewReferenceDetails()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.referenceDetails)
	s.Assert().NoError(err)

	decoded := new(ReferenceDetails)
	s.Assert().NoError(avro.Unmarshal(sch, data, decoded))
	s.Assert().Equal(s.referenceDetails.URL, decoded.URL)
	s.Assert().Equal(s.referenceDetails.DomainName, decoded.DomainName)
	s.Assert().Equal(s.referenceDetails.DomainMetadata, decoded.DomainMetadata)
}

func TestReferenceDetails(t *testing.T) {
	suite.Run(t, new(referenceDetailsTestSuite))
}

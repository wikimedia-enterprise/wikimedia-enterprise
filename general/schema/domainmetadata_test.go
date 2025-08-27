package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type domainMetadataTestSuite struct {
	suite.Suite
	domainMetadata *DomainMetadata
}

func (s *domainMetadataTestSuite) SetupTest() {
	psLabelLocal := "reliable"
	psLabelEnwiki := "mostly_reliable"
	s.domainMetadata = &DomainMetadata{
		PsLabelLocal:  &psLabelLocal,
		PsLabelEnwiki: &psLabelEnwiki,
		SurvivalRatio: 0.75,
		PageCount:     120,
		EditorsCount:  45,
	}
}

func (s *domainMetadataTestSuite) TestNewDomainMetadataSchema() {
	sch, err := NewDomainMetadata()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.domainMetadata)
	s.Assert().NoError(err)

	decoded := new(DomainMetadata)
	s.Assert().NoError(avro.Unmarshal(sch, data, decoded))
	s.Assert().Equal(s.domainMetadata.PsLabelLocal, decoded.PsLabelLocal)
	s.Assert().Equal(s.domainMetadata.PsLabelEnwiki, decoded.PsLabelEnwiki)
	s.Assert().Equal(s.domainMetadata.SurvivalRatio, decoded.SurvivalRatio)
	s.Assert().Equal(s.domainMetadata.PageCount, decoded.PageCount)
	s.Assert().Equal(s.domainMetadata.EditorsCount, decoded.EditorsCount)
}

func TestDomainMetadata(t *testing.T) {
	suite.Run(t, new(domainMetadataTestSuite))
}

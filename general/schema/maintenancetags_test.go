package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type maintenanceTagsTestSuite struct {
	suite.Suite
	maintenanceTags *MaintenanceTags
}

func (s *maintenanceTagsTestSuite) SetupTest() {
	s.maintenanceTags = &MaintenanceTags{
		CitationNeededCount:      1,
		PovCount:                 2,
		ClarificationNeededCount: 3,
		UpdateCount:              4,
	}
}

func (s *maintenanceTagsTestSuite) TestNewMaintenanceTagsSchema() {
	sch, err := NewMaintenanceTagsSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.maintenanceTags)
	s.Assert().NoError(err)

	maintenanceTags := new(MaintenanceTags)
	s.Assert().NoError(avro.Unmarshal(sch, data, maintenanceTags))
	s.Assert().Equal(s.maintenanceTags.CitationNeededCount, maintenanceTags.CitationNeededCount)
	s.Assert().Equal(s.maintenanceTags.PovCount, maintenanceTags.PovCount)
	s.Assert().Equal(s.maintenanceTags.ClarificationNeededCount, maintenanceTags.ClarificationNeededCount)
	s.Assert().Equal(s.maintenanceTags.UpdateCount, maintenanceTags.UpdateCount)
}

func TestMaintenanceTags(t *testing.T) {
	suite.Run(t, new(maintenanceTagsTestSuite))
}

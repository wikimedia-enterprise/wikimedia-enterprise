package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type projectTestSuite struct {
	suite.Suite
	project *Project
}

func (s *projectTestSuite) SetupTest() {
	dateModified := time.Now().UTC()
	s.project = &Project{
		Event:          NewEvent(EventTypeCreate),
		Name:           "Wikipedia",
		AdditionalType: "wikipedia",
		Identifier:     "enwiki",
		URL:            "http://enwiki.org",
		Version:        "072328a8d11be67279ade5d6288fc17f",
		DateModified:   &dateModified,
		Namespace: &Namespace{
			Identifier:    14,
			AlternateName: "File",
			Name:          "Файл",
		},
		InLanguage: &Language{
			Identifier:    "en",
			Name:          "English",
			AlternateName: "English",
		},
		Size: &Size{
			Value:    50,
			UnitText: "MB",
		},
	}
}

func (s *projectTestSuite) TestNewProjectSchema() {
	sch, err := NewProjectSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.project)
	s.Assert().NoError(err)

	project := new(Project)
	s.Assert().NoError(avro.Unmarshal(sch, data, project))
	s.Assert().Equal(s.project.Name, project.Name)
	s.Assert().Equal(s.project.Identifier, project.Identifier)
	s.Assert().Equal(s.project.URL, project.URL)
	s.Assert().Equal(s.project.Version, project.Version)
	s.Assert().Equal(s.project.DateModified.Format(time.RFC1123), project.DateModified.Format(time.RFC1123))
	s.Assert().Equal(s.project.Size.UnitText, project.Size.UnitText)
	s.Assert().Equal(s.project.Size.Value, project.Size.Value)
	s.Assert().Equal(s.project.Namespace.Identifier, project.Namespace.Identifier)
	s.Assert().Equal(s.project.Namespace.AlternateName, project.Namespace.AlternateName)
	s.Assert().Equal(s.project.Namespace.Name, project.Namespace.Name)
	s.Assert().Equal(s.project.InLanguage.Identifier, project.InLanguage.Identifier)
	s.Assert().Equal(s.project.InLanguage.Name, project.InLanguage.Name)
	s.Assert().Equal(s.project.InLanguage.AlternateName, project.InLanguage.AlternateName)
	s.Assert().Equal(s.project.AdditionalType, project.AdditionalType)

	// Event fields
	s.Assert().Equal(s.project.Event.Identifier, project.Event.Identifier)
	s.Assert().Equal(s.project.Event.Type, project.Event.Type)
	s.Assert().Equal(s.project.Event.DateCreated.Format(time.RFC1123), project.Event.DateCreated.Format(time.RFC1123))
}

func TestProject(t *testing.T) {
	suite.Run(t, new(projectTestSuite))
}

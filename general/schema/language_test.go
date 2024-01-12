package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type languageTestSuite struct {
	suite.Suite
	language *Language
}

func (s *languageTestSuite) SetupTest() {
	s.language = &Language{
		Event:         NewEvent(EventTypeCreate),
		Identifier:    "uk",
		Name:          "Українська",
		AlternateName: "Ukrainian",
		Direction:     "ltr",
	}
}

func (s *languageTestSuite) TestNewLanguageSchema() {
	sch, err := NewLanguageSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.language)
	s.Assert().NoError(err)

	language := new(Language)
	s.Assert().NoError(avro.Unmarshal(sch, data, language))
	s.Assert().Equal(s.language.Identifier, language.Identifier)
	s.Assert().Equal(s.language.Name, language.Name)
	s.Assert().Equal(s.language.AlternateName, language.AlternateName)
	s.Assert().Equal(s.language.Direction, language.Direction)

	// Event fields
	s.Assert().Equal(s.language.Event.Identifier, language.Event.Identifier)
	s.Assert().Equal(s.language.Event.Type, language.Event.Type)
	s.Assert().Equal(s.language.Event.DateCreated.Format(time.RFC1123), language.Event.DateCreated.Format(time.RFC1123))
}

func TestLanguage(t *testing.T) {
	suite.Run(t, new(languageTestSuite))
}

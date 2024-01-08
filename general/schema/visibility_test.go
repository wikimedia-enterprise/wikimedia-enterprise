package schema

import (
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type visibilityTestSuite struct {
	suite.Suite
	visibility *Visibility
}

func (s *visibilityTestSuite) SetupTest() {
	s.visibility = &Visibility{
		Text:    true,
		Comment: true,
		Editor:  true,
	}
}

func (s *visibilityTestSuite) TestNewVisibilitySchema() {
	sch, err := NewVisibilitySchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.visibility)
	s.Assert().NoError(err)

	visibility := new(Visibility)
	s.Assert().NoError(avro.Unmarshal(sch, data, visibility))
	s.Assert().Equal(s.visibility.Text, visibility.Text)
	s.Assert().Equal(s.visibility.Comment, visibility.Comment)
	s.Assert().Equal(s.visibility.Editor, visibility.Editor)
}

func TestVisibility(t *testing.T) {
	suite.Run(t, new(visibilityTestSuite))
}

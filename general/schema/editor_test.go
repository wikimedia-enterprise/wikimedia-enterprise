package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type editorTestSuite struct {
	suite.Suite
	editor *Editor
}

func (s *editorTestSuite) SetupTest() {
	dateStarted := time.Now().UTC()
	s.editor = &Editor{
		Identifier:        100,
		Name:              "Ninja",
		EditCount:         99,
		Groups:            []string{"bot", "admin"},
		IsBot:             true,
		IsAnonymous:       true,
		IsAdmin:           true,
		IsPatroller:       true,
		HasAdvancedRights: true,
		DateStarted:       &dateStarted,
	}
}

func (s *editorTestSuite) TestNewEditorSchema() {
	sch, err := NewEditorSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.editor)
	s.Assert().NoError(err)

	editor := new(Editor)
	s.Assert().NoError(avro.Unmarshal(sch, data, editor))
	s.Assert().Equal(s.editor.Identifier, editor.Identifier)
	s.Assert().Equal(s.editor.Name, editor.Name)
	s.Assert().Equal(s.editor.EditCount, editor.EditCount)
	s.Assert().Equal(s.editor.Groups, editor.Groups)
	s.Assert().Equal(s.editor.IsBot, editor.IsBot)
	s.Assert().Equal(s.editor.IsAnonymous, editor.IsAnonymous)
	s.Assert().Equal(s.editor.IsAdmin, editor.IsAdmin)
	s.Assert().Equal(s.editor.IsPatroller, editor.IsPatroller)
	s.Assert().Equal(s.editor.HasAdvancedRights, editor.HasAdvancedRights)
	s.Assert().Equal(s.editor.DateStarted.Format(time.RFC1123), editor.DateStarted.Format(time.RFC1123))
}

func TestEditor(t *testing.T) {
	suite.Run(t, new(editorTestSuite))
}

package builder_test

import (
	"math/rand"
	"testing"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/packages/builder"

	"github.com/stretchr/testify/suite"
)

type editorBuilderTestSuite struct {
	suite.Suite
	builder *builder.EditorBuilder
	rdm     *rand.Rand
}

func (s *editorBuilderTestSuite) SetupTest() {
	s.rdm = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.builder = builder.NewEditorBuilder()
}

func (s *editorBuilderTestSuite) TestNewEditorBuilder() {
	eb := builder.NewEditorBuilder()
	s.Assert().NotNil(eb)

	s.Assert().NotNil(eb.Build())
}

func (s *editorBuilderTestSuite) TestIdentifier() {
	id := s.rdm.Int()
	editor := s.builder.Identifier(id).Build()

	s.Assert().Equal(id, editor.Identifier)
}

func (s *editorBuilderTestSuite) TestName() {
	name := "Citation Bot"
	editor := s.builder.Name(name).Build()

	s.Assert().Equal(name, editor.Name)
}

func (s *editorBuilderTestSuite) TestIsAnonymous() {
	isAnonymous := s.rdm.Float32() < 0.5
	editor := s.builder.IsAnonymous(isAnonymous).Build()

	s.Assert().Equal(isAnonymous, editor.IsAnonymous)
}

func (s *editorBuilderTestSuite) TestDateStarted() {
	dateStarted := time.Now()
	editor := s.builder.DateStarted(&dateStarted).Build()

	s.Assert().Equal(&dateStarted, editor.DateStarted)
}

func (s *editorBuilderTestSuite) TestDateStartedNilDate() {
	editor := s.builder.DateStarted(nil).Build()

	s.Assert().Nil(editor.DateStarted)
}

func (s *editorBuilderTestSuite) TestEditCount() {
	editCount := s.rdm.Int()
	editor := s.builder.EditCount(editCount).Build()

	s.Assert().Equal(editCount, editor.EditCount)
}

func (s *editorBuilderTestSuite) TestGroups() {
	groups := []string{"example 1", "example 2"}
	editor := s.builder.Groups(groups).Build()

	s.Assert().Equal(groups, editor.Groups)
}

func (s *editorBuilderTestSuite) TestIsBot() {
	editor := s.builder.IsBot(false).Build()
	s.Assert().False(editor.IsBot)

	editor = s.builder.IsBot(true).Build()
	s.Assert().True(editor.IsBot)
}

func (s *editorBuilderTestSuite) TestIsAdmin() {
	groups := []string{"group 1", "group 2"}
	editor := s.builder.IsAdmin(groups).Build()
	s.Assert().False(editor.IsAdmin)

	groups = append(groups, "sysop")
	editor = s.builder.IsAdmin(groups).Build()
	s.Assert().True(editor.IsAdmin)
}

func (s *editorBuilderTestSuite) TestIsPatroller() {
	groups := []string{"group 1", "group 2"}
	editor := s.builder.IsPatroller(groups).Build()
	s.Assert().False(editor.IsPatroller)

	for role := range builder.PatrollerRoles {
		editor = s.builder.IsPatroller(append(groups, role)).Build()
		s.Assert().True(editor.IsPatroller)
	}
}

func (s *editorBuilderTestSuite) TestHasAdvancedRights() {
	groups := []string{"group 1", "group 2"}
	editor := s.builder.HasAdvancedRights(groups).Build()
	s.Assert().False(editor.HasAdvancedRights)

	for role := range builder.AdvancedRoles {
		editor = s.builder.HasAdvancedRights(append(groups, role)).Build()
		s.Assert().True(editor.HasAdvancedRights)
	}
}
func TestEditorBuilderSuite(t *testing.T) {
	suite.Run(t, new(editorBuilderTestSuite))
}

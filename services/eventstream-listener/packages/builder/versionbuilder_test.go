package builder_test

import (
	"math/rand"
	"testing"
	"time"
	"wikimedia-enterprise/services/eventstream-listener/packages/builder"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	"github.com/stretchr/testify/suite"
)

// versionBuilderTestSuite is the test suite for the VersionBuilder
type versionBuilderTestSuite struct {
	suite.Suite
	rdm     *rand.Rand
	builder *builder.VersionBuilder
}

func (s *versionBuilderTestSuite) SetupTest() {
	// Seed random generator
	s.rdm = rand.New(rand.NewSource(time.Now().UnixNano()))

	s.builder = builder.NewVersionBuilder()
}

// TestNewVersionBuilder ensures the NewVersionBuilder function returns a working version builder.
func (s *versionBuilderTestSuite) TestNewVersionBuilder() {
	vb := builder.NewVersionBuilder()
	s.Assert().NotNil(vb)

	version := vb.Build()

	s.Assert().NotNil(version)
}

func (s *versionBuilderTestSuite) TestIdentifier() {
	id := s.rdm.Int()
	version := s.builder.Identifier(id).Build()

	s.Assert().Equal(id, version.Identifier)
}

func (s *versionBuilderTestSuite) TestComment() {
	comment := "this is a comment..."
	version := s.builder.Comment(comment).Build()

	s.Assert().Equal(comment, version.Comment)
}

func (s *versionBuilderTestSuite) TestTags() {
	tags := []string{"tag 1", "tag 2", "tag 3"}

	version := s.builder.Tags(tags).Build()
	s.Assert().Equal(tags, version.Tags)
}

func (s *versionBuilderTestSuite) TestIsMinorEdit() {
	version := s.builder.IsMinorEdit(false).Build()
	s.Assert().Equal(false, version.IsMinorEdit)

	version = s.builder.IsMinorEdit(true).Build()
	s.Assert().Equal(true, version.IsMinorEdit)
}

func (s *versionBuilderTestSuite) TestIsFlaggedStable() {
	version := s.builder.IsFlaggedStable(false).Build()
	s.Assert().Equal(false, version.IsFlaggedStable)

	version = s.builder.IsFlaggedStable(true).Build()
	s.Assert().Equal(true, version.IsFlaggedStable)
}

func (s *versionBuilderTestSuite) TestScores() {
	scores := &schema.Scores{
		Damaging: &schema.ProbabilityScore{
			Prediction: false,
			Probability: &schema.Probability{
				False: 0.9,
				True:  0.1,
			},
		},
		GoodFaith: &schema.ProbabilityScore{
			Prediction: true,
			Probability: &schema.Probability{
				False: 0.1,
				True:  0.9,
			},
		},
	}

	version := s.builder.Scores(scores).Build()
	s.Assert().Equal(scores, version.Scores)
}

func (s *versionBuilderTestSuite) TestEditor() {
	editor := &schema.Editor{
		Identifier: s.rdm.Int(),
		Name:       "Citation Bot",
		EditCount:  s.rdm.Int(),
	}

	version := s.builder.Editor(editor).Build()
	s.Assert().Equal(editor, version.Editor)
}
func (s *versionBuilderTestSuite) TestNumberOfCharacters() {
	content := "some content"
	version := s.builder.NumberOfCharacters(content).Build()

	s.Assert().Equal(len([]rune(content)), version.NumberOfCharacters)
}

func (s *versionBuilderTestSuite) TestSize() {
	testCases := []int{
		6,
		5,
	}

	for _, tc := range testCases {
		version := s.builder.Size(tc).Build()
		s.Assert().Equal("B", version.Size.UnitText)
		s.Assert().Equal(float64(tc), version.Size.Value)
	}
}

func TestVersionBuilder(t *testing.T) {
	suite.Run(t, new(versionBuilderTestSuite))
}

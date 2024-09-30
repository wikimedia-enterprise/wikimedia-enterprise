package builder_test

import (
	"math/rand"
	"testing"
	"time"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/structured-data/packages/builder"

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
	testCases := []string{
		"Hello World!",
		"Hello, 世界", // Unicode characters for multi-byte rune testing
		`First line
		Second line
		This is a multiline string test`,
	}

	for _, tc := range testCases {
		version := s.builder.Size(tc).Build()
		s.Assert().Equal("B", version.Size.UnitText)
		s.Assert().Equal(float64(len([]byte(tc))), version.Size.Value)
	}
}

func (s *versionBuilderTestSuite) TestIsBreakingNews() {
	testCases := []struct {
		ibn bool
		res bool
	}{
		{true, true},
		{false, false},
	}

	for _, tc := range testCases {
		version := s.builder.IsBreakingNews(tc.ibn).Build()
		s.Assert().Equal(tc.res, version.IsBreakingNews)
	}
}

func (s *versionBuilderTestSuite) TestNoindex() {
	testCases := []struct {
		ndx bool
		res bool
	}{
		{true, true},
		{false, false},
	}

	for _, tc := range testCases {
		version := s.builder.Noindex(tc.ndx).Build()
		s.Assert().Equal(tc.res, version.Noindex)
	}
}

func (s *versionBuilderTestSuite) TestDiff() {
	diff := &schema.Diff{
		LongestNewRepeatedCharacter: 5,
		Size: &schema.Size{
			Value:    10,
			UnitText: "MB",
		},
	}

	version := s.builder.Diff(diff).Build()
	s.Assert().NotNil(version.Diff)
	s.Assert().Equal(diff, version.Diff)
}

func (s *versionBuilderTestSuite) TestMaintenanceTags() {
	mt := &schema.MaintenanceTags{
		CitationNeededCount: 5,
		PovCount:            10,
	}

	version := s.builder.MaintenanceTags(mt).Build()
	s.Assert().NotNil(version.MaintenanceTags)
	s.Assert().Equal(mt, version.MaintenanceTags)
}

func TestVersionBuilder(t *testing.T) {
	suite.Run(t, new(versionBuilderTestSuite))
}

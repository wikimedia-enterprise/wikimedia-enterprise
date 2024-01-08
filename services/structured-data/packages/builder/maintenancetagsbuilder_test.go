package builder

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type maintenanceTagsBuilderTestSuite struct {
	suite.Suite
	builder *MaintenanceTagsBuilder
}

func (s *maintenanceTagsBuilderTestSuite) SetupTest() {
	s.builder = NewMaintenanceTagsBuilder()
}

func (s *maintenanceTagsBuilderTestSuite) TestHasUpdate() {
	testCases := []struct {
		str string
		tot int
	}{
		{"{{update|}} {{update|}}", 2},
		{"{{Update|}} {{updated}", 1},
		{"n{{Update|}} lorem ipsum", 1},
		{"{{update|reason=test|date=November 2023}} and {{Update|}} {{Update}}", 3},
		{"{{update|date=December 2021}}", 1},
		{"{{update|reason=%_/|date=December 2021}}", 1},
		{
			`Multiline test case
			this article has the {{update}} tag
			`,
			1,
		},
		{"{{updated}", 0},
		{"{updated|}", 0},
		{"update", 0},
		{"{{}}", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		mts := s.builder.Update(tc.str).Build()
		s.Assert().Equal(tc.tot, mts.UpdateCount)
	}
}

func (s *maintenanceTagsBuilderTestSuite) TestHasClarify() {
	testCases := []struct {
		str string
		tot int
	}{
		{"{{clarify|}} {{clarify|}}", 2},
		{"{{Clarify|}} {{clarified}", 1},
		{"n{{Clarify|}} lorem ipsum", 1},
		{"{{clarify|reason=test|date=November 2023}} and {{Clarify|}} {{Clarify}}", 3},
		{"{{clarify|date=December 2021}}", 1},
		{"{{clarify|reason=%_/|date=December 2021}}", 1},
		{
			`Multiline test case
			this article has the {{clarify}} tag
			`,
			1,
		},
		{"{{clarified}", 0},
		{"{clarified}}", 0},
		{"{clarified|}", 0},
		{"clarify", 0},
		{"{{}}", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		mts := s.builder.ClarificationNeeded(tc.str).Build()
		s.Assert().Equal(tc.tot, mts.ClarificationNeededCount)
	}
}

func (s *maintenanceTagsBuilderTestSuite) TestHasPOV() {
	testCases := []struct {
		str string
		tot int
	}{
		{"{{POV|}} {{POV|}}", 2},
		{"{{POV|}} {{POVed}", 1},
		{"n{{POV|}} lorem ipsum", 1},
		{"{{POV|reason=test|date=November 2023}} and {{POV|}} {{POV}}", 3},
		{"{{POV|date=December 2021}}", 1},
		{"{{POV|reason=%_/|date=December 2021}}", 1},
		{
			`Multiline test case
			this article has the {{POV}} tag
			`,
			1,
		},
		{"{{POVed}", 0},
		{"{POVed|}", 0},
		{"POV", 0},
		{"{{}}", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		mts := s.builder.PovIdentified(tc.str).Build()
		s.Assert().Equal(tc.tot, mts.PovCount)
	}
}

func (s *maintenanceTagsBuilderTestSuite) TestHasCitationNeeded() {
	testCases := []struct {
		str string
		tot int
	}{
		{"{{Citation needed|jhh}} {{Citation needed|}}", 2},
		{"{{Citation needed|}} {{Citation needed|}}", 2},
		{"n{{Citation needed|}} lorem ipsum", 1},
		{"{{Citation needed|reason=test|date=November 2023}} and {{Citation needed|}} {{Citation needed}}", 3},
		{"{{Citation needed|date=December 2021}}", 1},
		{"{{Citation needed|reason=%_/|date=December 2021}}", 1},
		{
			`Multiline test case
			this article has the {{Citation needed}} tag
			`,
			1,
		},
		{"{{Citation needed}}", 1},
		{"{Citation needed|}", 0},
		{"Citation needed", 0},
		{"{{}}", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		mts := s.builder.CitationNeeded(tc.str).Build()
		s.Assert().Equal(tc.tot, mts.CitationNeededCount)
	}
}

func TestMaintenanceTagsBuilderSuite(t *testing.T) {
	suite.Run(t, new(maintenanceTagsBuilderTestSuite))
}

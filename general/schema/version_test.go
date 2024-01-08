package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type versionTestSuite struct {
	suite.Suite
	version *Version
}

func (s *versionTestSuite) SetupTest() {
	dateStarted := time.Now().UTC()
	dateCreated := time.Now().UTC()
	s.version = &Version{
		Identifier:      88,
		Comment:         "...comment goes here...",
		Tags:            []string{"stable", "new"},
		IsMinorEdit:     true,
		IsFlaggedStable: true,
		Diff: &Diff{
			Size: &Size{
				UnitText: "MB",
				Value:    500,
			},
		},
		Scores: &Scores{
			Damaging: &ProbabilityScore{
				Prediction: true,
				Probability: &Probability{
					True:  0.1,
					False: 0.9,
				},
			},
			GoodFaith: &ProbabilityScore{
				Prediction: true,
				Probability: &Probability{
					True:  0.1,
					False: 0.9,
				},
			},
		},
		Editor: &Editor{
			Identifier:  100,
			Name:        "Ninja",
			EditCount:   99,
			Groups:      []string{"bot", "admin"},
			IsBot:       true,
			IsAnonymous: true,
			DateStarted: &dateStarted,
		},
		Event: &Event{
			Identifier:  "ec369574-bf5e-4325-a720-c78d36a80cdb",
			Type:        "delete",
			DateCreated: &dateCreated,
		},
		NumberOfCharacters:  5000,
		HasTagNeedsCitation: true,
		Size: &Size{
			UnitText: "B",
			Value:    100,
		},
		Noindex: true,
	}
}

func (s *versionTestSuite) TestNewVersionSchema() {
	sch, err := NewVersionSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.version)
	s.Assert().NoError(err)

	version := new(Version)
	s.Assert().NoError(avro.Unmarshal(sch, data, version))
	s.Assert().Equal(s.version.Identifier, version.Identifier)
	s.Assert().Equal(s.version.Comment, version.Comment)
	s.Assert().Equal(s.version.Tags, version.Tags)
	s.Assert().Equal(s.version.IsMinorEdit, version.IsMinorEdit)
	s.Assert().Equal(s.version.IsFlaggedStable, version.IsFlaggedStable)
	s.Assert().Equal(s.version.Editor.Identifier, version.Editor.Identifier)
	s.Assert().Equal(s.version.Editor.Name, version.Editor.Name)
	s.Assert().Equal(s.version.Editor.EditCount, version.Editor.EditCount)
	s.Assert().Equal(s.version.Editor.Groups, version.Editor.Groups)
	s.Assert().Equal(s.version.Editor.IsBot, version.Editor.IsBot)
	s.Assert().Equal(s.version.Editor.IsAnonymous, version.Editor.IsAnonymous)
	s.Assert().Equal(s.version.Editor.DateStarted.Format(time.RFC1123), version.Editor.DateStarted.Format(time.RFC1123))
	s.Assert().Equal(s.version.Scores.Damaging.Prediction, version.Scores.Damaging.Prediction)
	s.Assert().Equal(s.version.Scores.Damaging.Probability.False, version.Scores.Damaging.Probability.False)
	s.Assert().Equal(s.version.Scores.Damaging.Probability.True, version.Scores.Damaging.Probability.True)
	s.Assert().Equal(s.version.Scores.GoodFaith.Prediction, version.Scores.GoodFaith.Prediction)
	s.Assert().Equal(s.version.Scores.GoodFaith.Probability.False, version.Scores.GoodFaith.Probability.False)
	s.Assert().Equal(s.version.Scores.GoodFaith.Probability.True, version.Scores.GoodFaith.Probability.True)
	s.Assert().Equal(s.version.HasTagNeedsCitation, version.HasTagNeedsCitation)
	s.Assert().Equal(s.version.NumberOfCharacters, version.NumberOfCharacters)
	s.Assert().Equal(s.version.Size.UnitText, version.Size.UnitText)
	s.Assert().Equal(s.version.Size.Value, version.Size.Value)
	s.Assert().Equal(s.version.Diff.Size.Value, version.Diff.Size.Value)
	s.Assert().Equal(s.version.Diff.Size.UnitText, version.Diff.Size.UnitText)
	s.Assert().Equal(s.version.Noindex, version.Noindex)

	// event fields
	// s.Assert().Equal(s.version.Event.Identifier, version.Event.Identifier)
	// s.Assert().Equal(s.version.Event.Type, version.Event.Type)
	// s.Assert().Equal(s.version.Event.DateCreated.Format(time.RFC1123), version.Event.DateCreated.Format(time.RFC1123))
}

func TestVersion(t *testing.T) {
	suite.Run(t, new(versionTestSuite))
}

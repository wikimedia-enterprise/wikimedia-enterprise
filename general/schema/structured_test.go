package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type structuredTestSuite struct {
	suite.Suite
	astr *AvroStructured
}

func (s *structuredTestSuite) SetupTest() {

	pfc := 25
	lrd := "06/25"

	hasParts := &HasParts{
		Type:     "update2",
		Name:     "test2",
		Value:    "value2",
		HasParts: []*HasParts{},
	}

	hasParts2 := &HasParts{
		Type:     "update3",
		Name:     "test3",
		Value:    "value3",
		HasParts: []*HasParts{hasParts},
	}

	avroHasParts := &AvroHasParts{
		Type:     "update",
		Name:     "test",
		Value:    "value",
		Values:   []string{"val1"},
		HasParts: to_struct_ptr([]*HasParts{hasParts2}),
		Images: []*Image{{
			ContentUrl: "https://localhost",
			Width:      100,
			Height:     100,
			Thumbnail: &Thumbnail{
				ContentUrl: "https://localhost",
				Width:      10,
				Height:     10,
			},
			AlternativeText: "altText",
			Caption:         "Caption",
		}},
		Links: []*Link{{
			URL:  "url",
			Text: "text",
			Images: []*Image{{
				ContentUrl: "https://localhost",
				Width:      100,
				Height:     100,
				Thumbnail: &Thumbnail{
					ContentUrl: "https://localhost",
					Width:      10,
					Height:     10,
				},
				AlternativeText: "altText",
				Caption:         "Caption",
			}},
		}},
	}

	metadata := map[string]string{}

	metadata["a"] = "b"

	referenceLink := &Link{
		URL:  "url",
		Text: "text",
		Images: []*Image{{
			ContentUrl: "https://localhost",
			Width:      100,
			Height:     100,
			Thumbnail: &Thumbnail{
				ContentUrl: "https://localhost",
				Width:      10,
				Height:     10,
			},
			AlternativeText: "altText",
			Caption:         "Caption",
		},
		},
	}

	referenceText := &StructuredReferenceText{
		Value: "value",
		Links: []*Link{referenceLink},
	}

	reference := &StructuredReference{
		Identifier: "id",
		Group:      "group",
		Type:       "type",
		Metadata:   metadata,
		Text:       referenceText}

	dateModified := time.Now().UTC()
	dateCreated := time.Now().UTC()
	dateStarted := time.Now().UTC()

	s.astr = &AvroStructured{
		Name:       "Name",
		Identifier: 1,
		Abstract:   "Abstract",
		Version: &Version{
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
		},
		Event: &Event{
			Identifier:  "ec369574-bf5e-4325-a720-c78d36a80cdb",
			Type:        "create",
			DateCreated: &dateCreated,
		},
		URL:          "url",
		DateCreated:  &dateCreated,
		DateModified: &dateModified,
		MainEntity: &Entity{
			Identifier: "Q1",
			URL:        "http://test.com/Q1",
			Aspects:    []string{"A", "B", "C"},
		},
		IsPartOf: &Project{
			Name:       "Wikipedia",
			Identifier: "enwiki",
		},
		AdditionalEntities: []*Entity{&Entity{
			Identifier: "Q1",
			URL:        "http://test.com/Q1",
			Aspects:    []string{"A", "B", "C"},
		},
		},
		InLanguage: &Language{
			Identifier: "uk",
			Name:       "Українська",
		},
		Image: &Image{
			ContentUrl: "https://localhost",
			Width:      100,
			Height:     100,
			Thumbnail: &Thumbnail{
				ContentUrl: "https://localhost",
				Width:      10,
				Height:     10,
			},
			AlternativeText: "altText",
			Caption:         "Caption",
		},
		Visibility: &Visibility{
			Text:    true,
			Comment: true,
			Editor:  true,
		},
		License: []*License{
			{
				Name:       "Creative Commons Attribution Share Alike 3.0 Unported",
				Identifier: "CC-BY-SA-3.0",
				URL:        "https://creativecommons.org/licenses/by-sa/3.0/",
			},
		},
		Description: "description",
		Infoboxes:   []*AvroHasParts{avroHasParts},
		Sections:    []*AvroHasParts{avroHasParts},
		References:  []*StructuredReference{reference},
		Attribution: &Attribution{
			EditorsCount: &EditorsCount{
				AnonymousCount:  "<100",
				RegisteredCount: ">500",
			},
			ParsedReferencesCount: pfc,
			LastRevised:           lrd,
		},
	}
}

func (s *structuredTestSuite) TestStructuredSchema() {
	sch, err := NewAvroStructuredSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.astr)
	s.Assert().NoError(err)

	hasparts := new(AvroStructured)
	s.Assert().NoError(avro.Unmarshal(sch, data, hasparts))

	hp1, _ := hasparts.ToJsonStruct()
	hp2, _ := s.astr.ToJsonStruct()

	php := hp1.Infoboxes[0]
	ohp := hp2.Infoboxes[0]

	s.Assert().Equal(ohp.Type, php.Type)
	s.Assert().Equal(ohp.Name, php.Name)
	s.Assert().Equal(ohp.Value, php.Value)

	s.Assert().Equal(ohp.Values, php.Values)

	s.Assert().Equal(ohp.HasParts[0].Type, php.HasParts[0].Type)
	s.Assert().Equal(ohp.HasParts[0].Name, php.HasParts[0].Name)
	s.Assert().Equal(ohp.HasParts[0].Value, php.HasParts[0].Value)

	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Type, php.HasParts[0].HasParts[0].Type)
	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Name, php.HasParts[0].HasParts[0].Name)
	s.Assert().Equal(ohp.HasParts[0].HasParts[0].Value, php.HasParts[0].HasParts[0].Value)

	s.Assert().Equal(ohp.Images[0].ContentUrl, php.Images[0].ContentUrl)
	s.Assert().Equal(ohp.Images[0].Thumbnail, php.Images[0].Thumbnail)
	s.Assert().Equal(ohp.Images[0].Width, php.Images[0].Width)
	s.Assert().Equal(ohp.Images[0].Height, php.Images[0].Height)
	s.Assert().Equal(ohp.Images[0].AlternativeText, php.Images[0].AlternativeText)
	s.Assert().Equal(ohp.Images[0].Caption, php.Images[0].Caption)

	s.Assert().Equal(ohp.Links[0].URL, php.Links[0].URL)
	s.Assert().Equal(ohp.Links[0].Text, php.Links[0].Text)
	s.Assert().Equal(ohp.Links[0].Images[0].ContentUrl, php.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(ohp.Links[0].Images[0].Thumbnail, php.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(ohp.Links[0].Images[0].Width, php.Links[0].Images[0].Width)
	s.Assert().Equal(ohp.Links[0].Images[0].Height, php.Links[0].Images[0].Height)
	s.Assert().Equal(ohp.Links[0].Images[0].AlternativeText, php.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(ohp.Links[0].Images[0].Caption, php.Links[0].Images[0].Caption)

	// structured fields
	s.Assert().Equal(hp1.Name, hp2.Name)
	s.Assert().Equal(hp1.Identifier, hp2.Identifier)
	s.Assert().Equal(hp1.Abstract, hp2.Abstract)

	s.Assert().Equal(hp1.License, hp2.License)
	s.Assert().Equal(hp1.DateCreated.Format(time.RFC1123), hp2.DateCreated.Format(time.RFC1123))
	s.Assert().Equal(hp1.DateModified.Format(time.RFC1123), hp2.DateModified.Format(time.RFC1123))

	// version fields
	s.Assert().Equal(hp1.Version.Identifier, hp2.Version.Identifier)
	s.Assert().Equal(hp1.Version.Comment, hp2.Version.Comment)
	s.Assert().Equal(hp1.Version.Tags, hp2.Version.Tags)
	s.Assert().Equal(hp1.Version.IsMinorEdit, hp2.Version.IsMinorEdit)
	s.Assert().Equal(hp1.Version.IsFlaggedStable, hp2.Version.IsFlaggedStable)
	s.Assert().Equal(hp1.Version.HasTagNeedsCitation, hp2.Version.HasTagNeedsCitation)
	s.Assert().Equal(hp1.Version.NumberOfCharacters, hp2.Version.NumberOfCharacters)
	s.Assert().Equal(hp1.Version.Editor.Identifier, hp2.Version.Editor.Identifier)
	s.Assert().Equal(hp1.Version.Editor.Name, hp2.Version.Editor.Name)
	s.Assert().Equal(hp1.Version.Editor.EditCount, hp2.Version.Editor.EditCount)
	s.Assert().Equal(hp1.Version.Editor.Groups, hp2.Version.Editor.Groups)
	s.Assert().Equal(hp1.Version.Editor.IsBot, hp2.Version.Editor.IsBot)
	s.Assert().Equal(hp1.Version.Editor.IsAnonymous, hp2.Version.Editor.IsAnonymous)
	s.Assert().Equal(hp1.Version.Editor.DateStarted.Format(time.RFC1123), hp2.Version.Editor.DateStarted.Format(time.RFC1123))
	s.Assert().Equal(hp1.Version.Scores.Damaging.Prediction, hp2.Version.Scores.Damaging.Prediction)
	s.Assert().Equal(hp1.Version.Scores.Damaging.Probability.False, hp2.Version.Scores.Damaging.Probability.False)
	s.Assert().Equal(hp1.Version.Scores.Damaging.Probability.True, hp2.Version.Scores.Damaging.Probability.True)
	s.Assert().Equal(hp1.Version.Scores.GoodFaith.Prediction, hp2.Version.Scores.GoodFaith.Prediction)
	s.Assert().Equal(hp1.Version.Scores.GoodFaith.Probability.False, hp2.Version.Scores.GoodFaith.Probability.False)
	s.Assert().Equal(hp1.Version.Scores.GoodFaith.Probability.True, hp2.Version.Scores.GoodFaith.Probability.True)
	s.Assert().Equal(hp1.Version.Size.Value, hp2.Version.Size.Value)
	s.Assert().Equal(hp1.Version.Size.UnitText, hp2.Version.Size.UnitText)
	s.Assert().Equal(hp1.Version.Diff.Size.Value, hp2.Version.Diff.Size.Value)
	s.Assert().Equal(hp1.Version.Diff.Size.UnitText, hp2.Version.Diff.Size.UnitText)

	// language fields
	s.Assert().Equal(hp1.InLanguage.Identifier, hp2.InLanguage.Identifier)
	s.Assert().Equal(hp1.InLanguage.Name, hp2.InLanguage.Name)

	// entity fields
	s.Assert().Equal(hp1.MainEntity.Identifier, hp2.MainEntity.Identifier)
	s.Assert().Equal(hp1.MainEntity.URL, hp2.MainEntity.URL)
	s.Assert().Equal(hp1.MainEntity.Aspects, hp2.MainEntity.Aspects)

	// project fields
	s.Assert().Equal(hp1.IsPartOf.Name, hp2.IsPartOf.Name)
	s.Assert().Equal(hp1.IsPartOf.Identifier, hp2.IsPartOf.Identifier)

	// visibility fields
	s.Assert().Equal(hp1.Visibility.Text, hp2.Visibility.Text)
	s.Assert().Equal(hp1.Visibility.Comment, hp2.Visibility.Comment)
	s.Assert().Equal(hp1.Visibility.Editor, hp2.Visibility.Editor)

	// event fields
	s.Assert().Equal(hp1.Event.Identifier, hp2.Event.Identifier)
	s.Assert().Equal(hp1.Event.Type, hp2.Event.Type)
	s.Assert().Equal(hp1.Event.DateCreated.Format(time.RFC1123), hp2.Event.DateCreated.Format(time.RFC1123))

	s.Assert().Equal(hp1.Image.ContentUrl, hp2.Image.ContentUrl)
	s.Assert().Equal(hp1.Image.Width, hp2.Image.Width)
	s.Assert().Equal(hp1.Image.Height, hp2.Image.Height)
	//s.Assert().Equal(hp1.Image.Thumbnail.ContentUrl, hp2.Image.Thumbnail.ContentUrl)
	//s.Assert().Equal(hp1.Image.Thumbnail.Width, hp2.Image.Thumbnail.Width)
	//s.Assert().Equal(hp1.Image.Thumbnail.Height, hp2.Image.Thumbnail.Height)

	s.Assert().Equal(hp1.References[0].Identifier, hp2.References[0].Identifier)
	s.Assert().Equal(hp1.References[0].Group, hp2.References[0].Group)
	s.Assert().Equal(hp1.References[0].Type, hp2.References[0].Type)
	s.Assert().Equal(hp1.References[0].Metadata, hp2.References[0].Metadata)
	s.Assert().Equal(hp1.References[0].Text.Value, hp2.References[0].Text.Value)

	s.Assert().Equal(hp1.References[0].Text.Value, hp2.References[0].Text.Value)

	s.Assert().Equal(hp1.References[0].Text.Links[0].URL, hp2.References[0].Text.Links[0].URL)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Text, hp2.References[0].Text.Links[0].Text)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].ContentUrl, hp2.References[0].Text.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Thumbnail, hp2.References[0].Text.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Width, hp2.References[0].Text.Links[0].Images[0].Width)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Height, hp2.References[0].Text.Links[0].Images[0].Height)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].AlternativeText, hp2.References[0].Text.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Caption, hp2.References[0].Text.Links[0].Images[0].Caption)

	// Attribution fields
	s.Assert().Equal(hp1.Attribution.EditorsCount.AnonymousCount, hp2.Attribution.EditorsCount.AnonymousCount)
	s.Assert().Equal(hp1.Attribution.EditorsCount.RegisteredCount, hp2.Attribution.EditorsCount.RegisteredCount)
	s.Assert().Equal(hp1.Attribution.ParsedReferencesCount, hp2.Attribution.ParsedReferencesCount)
	s.Assert().Equal(hp1.Attribution.LastRevised, hp2.Attribution.LastRevised)
}

func (s *structuredTestSuite) TestAvroStructuredSchema() {
	sch, err := NewAvroStructuredSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.astr)
	s.Assert().NoError(err)

	hasparts := new(AvroStructured)
	s.Assert().NoError(avro.Unmarshal(sch, data, hasparts))

	p, _ := hasparts.ToJsonStruct()
	hp1, _ := p.ToAvroStruct()

	p2, _ := s.astr.ToJsonStruct()
	hp2, _ := p2.ToAvroStruct()

	php := hp1.Infoboxes[0]
	ohp := hp2.Infoboxes[0]

	s.Assert().Equal(ohp.Type, php.Type)
	s.Assert().Equal(ohp.Name, php.Name)
	s.Assert().Equal(ohp.Value, php.Value)

	s.Assert().Equal(ohp.Values, php.Values)

	s.Assert().Equal(ohp.HasParts, php.HasParts)

	s.Assert().Equal(ohp.Images[0].ContentUrl, php.Images[0].ContentUrl)
	s.Assert().Equal(ohp.Images[0].Thumbnail, php.Images[0].Thumbnail)
	s.Assert().Equal(ohp.Images[0].Width, php.Images[0].Width)
	s.Assert().Equal(ohp.Images[0].Height, php.Images[0].Height)
	s.Assert().Equal(ohp.Images[0].AlternativeText, php.Images[0].AlternativeText)
	s.Assert().Equal(ohp.Images[0].Caption, php.Images[0].Caption)

	s.Assert().Equal(ohp.Links[0].URL, php.Links[0].URL)
	s.Assert().Equal(ohp.Links[0].Text, php.Links[0].Text)
	s.Assert().Equal(ohp.Links[0].Images[0].ContentUrl, php.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(ohp.Links[0].Images[0].Thumbnail, php.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(ohp.Links[0].Images[0].Width, php.Links[0].Images[0].Width)
	s.Assert().Equal(ohp.Links[0].Images[0].Height, php.Links[0].Images[0].Height)
	s.Assert().Equal(ohp.Links[0].Images[0].AlternativeText, php.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(ohp.Links[0].Images[0].Caption, php.Links[0].Images[0].Caption)

	// structured fields
	s.Assert().Equal(hp1.Name, hp2.Name)
	s.Assert().Equal(hp1.Identifier, hp2.Identifier)
	s.Assert().Equal(hp1.Abstract, hp2.Abstract)

	s.Assert().Equal(hp1.License, hp2.License)
	s.Assert().Equal(hp1.DateCreated.Format(time.RFC1123), hp2.DateCreated.Format(time.RFC1123))
	s.Assert().Equal(hp1.DateModified.Format(time.RFC1123), hp2.DateModified.Format(time.RFC1123))

	// version fields
	s.Assert().Equal(hp1.Version.Identifier, hp2.Version.Identifier)
	s.Assert().Equal(hp1.Version.Comment, hp2.Version.Comment)
	s.Assert().Equal(hp1.Version.Tags, hp2.Version.Tags)
	s.Assert().Equal(hp1.Version.IsMinorEdit, hp2.Version.IsMinorEdit)
	s.Assert().Equal(hp1.Version.IsFlaggedStable, hp2.Version.IsFlaggedStable)
	s.Assert().Equal(hp1.Version.HasTagNeedsCitation, hp2.Version.HasTagNeedsCitation)
	s.Assert().Equal(hp1.Version.NumberOfCharacters, hp2.Version.NumberOfCharacters)
	s.Assert().Equal(hp1.Version.Editor.Identifier, hp2.Version.Editor.Identifier)
	s.Assert().Equal(hp1.Version.Editor.Name, hp2.Version.Editor.Name)
	s.Assert().Equal(hp1.Version.Editor.EditCount, hp2.Version.Editor.EditCount)
	s.Assert().Equal(hp1.Version.Editor.Groups, hp2.Version.Editor.Groups)
	s.Assert().Equal(hp1.Version.Editor.IsBot, hp2.Version.Editor.IsBot)
	s.Assert().Equal(hp1.Version.Editor.IsAnonymous, hp2.Version.Editor.IsAnonymous)
	s.Assert().Equal(hp1.Version.Editor.DateStarted.Format(time.RFC1123), hp2.Version.Editor.DateStarted.Format(time.RFC1123))
	s.Assert().Equal(hp1.Version.Scores.Damaging.Prediction, hp2.Version.Scores.Damaging.Prediction)
	s.Assert().Equal(hp1.Version.Scores.Damaging.Probability.False, hp2.Version.Scores.Damaging.Probability.False)
	s.Assert().Equal(hp1.Version.Scores.Damaging.Probability.True, hp2.Version.Scores.Damaging.Probability.True)
	s.Assert().Equal(hp1.Version.Scores.GoodFaith.Prediction, hp2.Version.Scores.GoodFaith.Prediction)
	s.Assert().Equal(hp1.Version.Scores.GoodFaith.Probability.False, hp2.Version.Scores.GoodFaith.Probability.False)
	s.Assert().Equal(hp1.Version.Scores.GoodFaith.Probability.True, hp2.Version.Scores.GoodFaith.Probability.True)
	s.Assert().Equal(hp1.Version.Size.Value, hp2.Version.Size.Value)
	s.Assert().Equal(hp1.Version.Size.UnitText, hp2.Version.Size.UnitText)
	s.Assert().Equal(hp1.Version.Diff.Size.Value, hp2.Version.Diff.Size.Value)
	s.Assert().Equal(hp1.Version.Diff.Size.UnitText, hp2.Version.Diff.Size.UnitText)

	// language fields
	s.Assert().Equal(hp1.InLanguage.Identifier, hp2.InLanguage.Identifier)
	s.Assert().Equal(hp1.InLanguage.Name, hp2.InLanguage.Name)

	// entity fields
	s.Assert().Equal(hp1.MainEntity.Identifier, hp2.MainEntity.Identifier)
	s.Assert().Equal(hp1.MainEntity.URL, hp2.MainEntity.URL)
	s.Assert().Equal(hp1.MainEntity.Aspects, hp2.MainEntity.Aspects)

	// project fields
	s.Assert().Equal(hp1.IsPartOf.Name, hp2.IsPartOf.Name)
	s.Assert().Equal(hp1.IsPartOf.Identifier, hp2.IsPartOf.Identifier)

	// visibility fields
	s.Assert().Equal(hp1.Visibility.Text, hp2.Visibility.Text)
	s.Assert().Equal(hp1.Visibility.Comment, hp2.Visibility.Comment)
	s.Assert().Equal(hp1.Visibility.Editor, hp2.Visibility.Editor)

	// event fields
	s.Assert().Equal(hp1.Event.Identifier, hp2.Event.Identifier)
	s.Assert().Equal(hp1.Event.Type, hp2.Event.Type)
	s.Assert().Equal(hp1.Event.DateCreated.Format(time.RFC1123), hp2.Event.DateCreated.Format(time.RFC1123))

	s.Assert().Equal(hp1.Image.ContentUrl, hp2.Image.ContentUrl)
	s.Assert().Equal(hp1.Image.Width, hp2.Image.Width)
	s.Assert().Equal(hp1.Image.Height, hp2.Image.Height)
	// s.Assert().Equal(hp1.Image.Thumbnail.ContentUrl, hp2.Image.Thumbnail.ContentUrl)
	// s.Assert().Equal(hp1.Image.Thumbnail.Width, hp2.Image.Thumbnail.Width)
	// s.Assert().Equal(hp1.Image.Thumbnail.Height, hp2.Image.Thumbnail.Height)

	s.Assert().Equal(hp1.References[0].Identifier, hp2.References[0].Identifier)
	s.Assert().Equal(hp1.References[0].Group, hp2.References[0].Group)
	s.Assert().Equal(hp1.References[0].Type, hp2.References[0].Type)
	s.Assert().Equal(hp1.References[0].Metadata, hp2.References[0].Metadata)
	s.Assert().Equal(hp1.References[0].Text.Value, hp2.References[0].Text.Value)

	s.Assert().Equal(hp1.References[0].Text.Value, hp2.References[0].Text.Value)

	s.Assert().Equal(hp1.References[0].Text.Links[0].URL, hp2.References[0].Text.Links[0].URL)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Text, hp2.References[0].Text.Links[0].Text)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].ContentUrl, hp2.References[0].Text.Links[0].Images[0].ContentUrl)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Thumbnail, hp2.References[0].Text.Links[0].Images[0].Thumbnail)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Width, hp2.References[0].Text.Links[0].Images[0].Width)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Height, hp2.References[0].Text.Links[0].Images[0].Height)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].AlternativeText, hp2.References[0].Text.Links[0].Images[0].AlternativeText)
	s.Assert().Equal(hp1.References[0].Text.Links[0].Images[0].Caption, hp2.References[0].Text.Links[0].Images[0].Caption)

	// Attribution fields
	s.Assert().Equal(hp1.Attribution.EditorsCount.AnonymousCount, hp2.Attribution.EditorsCount.AnonymousCount)
	s.Assert().Equal(hp1.Attribution.EditorsCount.RegisteredCount, hp2.Attribution.EditorsCount.RegisteredCount)
	s.Assert().Equal(hp1.Attribution.ParsedReferencesCount, hp2.Attribution.ParsedReferencesCount)
	s.Assert().Equal(hp1.Attribution.LastRevised, hp2.Attribution.LastRevised)
}

func TestStructured(t *testing.T) {
	suite.Run(t, new(structuredTestSuite))
}

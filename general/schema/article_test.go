package schema

import (
	"testing"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

type articleTestSuite struct {
	suite.Suite
	article *Article
}

func (s *articleTestSuite) SetupTest() {
	datePreviouslyModified := time.Now().UTC()
	dateModified := time.Now().UTC()
	dateStarted := time.Now().UTC()
	dateCreated := time.Now().UTC()
	s.article = &Article{
		Name:                   "Earth",
		Identifier:             100,
		Abstract:               "...short description of the article...",
		WatchersCount:          111,
		DateCreated:            &dateCreated,
		DateModified:           &dateModified,
		DatePreviouslyModified: &datePreviouslyModified,
		URL:                    "http://enwiki.org",
		Version: &Version{
			Identifier:          88,
			Comment:             "...comment goes here...",
			Tags:                []string{"stable", "new"},
			IsMinorEdit:         true,
			IsFlaggedStable:     true,
			HasTagNeedsCitation: true,
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
			NumberOfCharacters: 11000,
			Size: &Size{
				Value:    50,
				UnitText: "MB",
			},
			Diff: &Diff{
				Size: &Size{
					Value:    500,
					UnitText: "MB",
				},
			},
		},
		PreviousVersion: &PreviousVersion{
			Identifier:         87,
			NumberOfCharacters: 10000,
		},
		VersionIdentifier: "/enwiki/100",
		Namespace: &Namespace{
			Name:       "File",
			Identifier: 6,
		},
		InLanguage: &Language{
			Identifier: "uk",
			Name:       "Українська",
		},
		MainEntity: &Entity{
			Identifier: "Q1",
			URL:        "http://test.com/Q1",
			Aspects:    []string{"A", "B", "C"},
		},
		IsPartOf: &Project{
			Name:       "Wikipedia",
			Identifier: "enwiki",
		},
		Redirects: []*Redirect{
			{
				Name: "Planet Earth",
				URL:  "http://enwiki.org/Planet_Earth",
			},
		},
		Templates: []*Template{
			{
				Name: "Info Box",
				URL:  "http://enwiki.org/Template:Info_box",
			},
		},
		Categories: []*Category{
			{
				Name: "Category:Planet",
				URL:  "http://enwiki.org/Category:Planet",
			},
		},
		Protection: []*Protection{
			{
				Type:   "general",
				Expiry: "infinite",
				Level:  "max",
			},
		},
		ArticleBody: &ArticleBody{
			HTML:     "...article body goes here...",
			WikiText: "...article body wikitext goes here...",
		},
		License: []*License{
			{
				Name:       "Creative Commons Attribution Share Alike 3.0 Unported",
				Identifier: "CC-BY-SA-3.0",
				URL:        "https://creativecommons.org/licenses/by-sa/3.0/",
			},
		},
		Visibility: &Visibility{
			Text:    true,
			Comment: true,
			Editor:  true,
		},
		Event: &Event{
			Identifier:  "ec369574-bf5e-4325-a720-c78d36a80cdb",
			Type:        "create",
			DateCreated: &dateCreated,
		},
		Image: &Image{
			ContentUrl: "https://upload.wikimedia.org/wikipedia/commons/3/3e/Einstein_1921_by_F_Schmutzer_-_restoration.jpg",
			Width:      2523,
			Height:     3313,
			Thumbnail: &Thumbnail{
				ContentUrl: "https://upload.wikimedia.org/wikipedia/commons/3/3e/Einstein_1921_by_F_Schmutzer_-_restoration.jpg",
				Width:      100,
				Height:     100,
			},
		},
	}
}

func (s *articleTestSuite) TestNewArticleSchema() {
	sch, err := NewArticleSchema()
	s.Assert().NoError(err)

	data, err := avro.Marshal(sch, s.article)
	s.Assert().NoError(err)

	article := new(Article)
	s.Assert().NoError(avro.Unmarshal(sch, data, article))

	// article fields
	s.Assert().Equal(s.article.Name, article.Name)
	s.Assert().Equal(s.article.WatchersCount, article.WatchersCount)
	s.Assert().Equal(s.article.Identifier, article.Identifier)
	s.Assert().Equal(s.article.Abstract, article.Abstract)
	s.Assert().Equal(s.article.ArticleBody, article.ArticleBody)
	s.Assert().Equal(s.article.Protection, article.Protection)
	s.Assert().Equal(s.article.Templates, article.Templates)
	s.Assert().Equal(s.article.Redirects, article.Redirects)
	s.Assert().Equal(s.article.Categories, article.Categories)
	s.Assert().Equal(s.article.License, article.License)
	s.Assert().Equal(s.article.DateCreated.Format(time.RFC1123), article.DateCreated.Format(time.RFC1123))
	s.Assert().Equal(s.article.DateModified.Format(time.RFC1123), article.DateModified.Format(time.RFC1123))
	s.Assert().Equal(s.article.DatePreviouslyModified.Format(time.RFC1123), article.DatePreviouslyModified.Format(time.RFC1123))

	// version fields
	s.Assert().Equal(s.article.Version.Identifier, article.Version.Identifier)
	s.Assert().Equal(s.article.Version.Comment, article.Version.Comment)
	s.Assert().Equal(s.article.Version.Tags, article.Version.Tags)
	s.Assert().Equal(s.article.Version.IsMinorEdit, article.Version.IsMinorEdit)
	s.Assert().Equal(s.article.Version.IsFlaggedStable, article.Version.IsFlaggedStable)
	s.Assert().Equal(s.article.Version.HasTagNeedsCitation, article.Version.HasTagNeedsCitation)
	s.Assert().Equal(s.article.Version.NumberOfCharacters, article.Version.NumberOfCharacters)
	s.Assert().Equal(s.article.Version.Editor.Identifier, article.Version.Editor.Identifier)
	s.Assert().Equal(s.article.Version.Editor.Name, article.Version.Editor.Name)
	s.Assert().Equal(s.article.Version.Editor.EditCount, article.Version.Editor.EditCount)
	s.Assert().Equal(s.article.Version.Editor.Groups, article.Version.Editor.Groups)
	s.Assert().Equal(s.article.Version.Editor.IsBot, article.Version.Editor.IsBot)
	s.Assert().Equal(s.article.Version.Editor.IsAnonymous, article.Version.Editor.IsAnonymous)
	s.Assert().Equal(s.article.Version.Editor.DateStarted.Format(time.RFC1123), article.Version.Editor.DateStarted.Format(time.RFC1123))
	s.Assert().Equal(s.article.Version.Scores.Damaging.Prediction, article.Version.Scores.Damaging.Prediction)
	s.Assert().Equal(s.article.Version.Scores.Damaging.Probability.False, article.Version.Scores.Damaging.Probability.False)
	s.Assert().Equal(s.article.Version.Scores.Damaging.Probability.True, article.Version.Scores.Damaging.Probability.True)
	s.Assert().Equal(s.article.Version.Scores.GoodFaith.Prediction, article.Version.Scores.GoodFaith.Prediction)
	s.Assert().Equal(s.article.Version.Scores.GoodFaith.Probability.False, article.Version.Scores.GoodFaith.Probability.False)
	s.Assert().Equal(s.article.Version.Scores.GoodFaith.Probability.True, article.Version.Scores.GoodFaith.Probability.True)
	s.Assert().Equal(s.article.Version.Size.Value, article.Version.Size.Value)
	s.Assert().Equal(s.article.Version.Size.UnitText, article.Version.Size.UnitText)
	s.Assert().Equal(s.article.Version.Diff.Size.Value, article.Version.Diff.Size.Value)
	s.Assert().Equal(s.article.Version.Diff.Size.UnitText, article.Version.Diff.Size.UnitText)

	s.Assert().Equal(*s.article.PreviousVersion, *article.PreviousVersion)
	s.Assert().Equal(s.article.VersionIdentifier, article.VersionIdentifier)

	// namespace fields
	s.Assert().Equal(s.article.Namespace.Identifier, article.Namespace.Identifier)
	s.Assert().Equal(s.article.Namespace.Name, article.Namespace.Name)

	// language fields
	s.Assert().Equal(s.article.InLanguage.Identifier, article.InLanguage.Identifier)
	s.Assert().Equal(s.article.InLanguage.Name, article.InLanguage.Name)

	// entity fields
	s.Assert().Equal(s.article.MainEntity.Identifier, article.MainEntity.Identifier)
	s.Assert().Equal(s.article.MainEntity.URL, article.MainEntity.URL)
	s.Assert().Equal(s.article.MainEntity.Aspects, article.MainEntity.Aspects)

	// project fields
	s.Assert().Equal(s.article.IsPartOf.Name, article.IsPartOf.Name)
	s.Assert().Equal(s.article.IsPartOf.Identifier, article.IsPartOf.Identifier)

	// visibility fields
	s.Assert().Equal(s.article.Visibility.Text, article.Visibility.Text)
	s.Assert().Equal(s.article.Visibility.Comment, article.Visibility.Comment)
	s.Assert().Equal(s.article.Visibility.Editor, article.Visibility.Editor)

	// event fields
	s.Assert().Equal(s.article.Event.Identifier, article.Event.Identifier)
	s.Assert().Equal(s.article.Event.Type, article.Event.Type)
	s.Assert().Equal(s.article.Event.DateCreated.Format(time.RFC1123), article.Event.DateCreated.Format(time.RFC1123))

	s.Assert().Equal(s.article.Image.ContentUrl, article.Image.ContentUrl)
	s.Assert().Equal(s.article.Image.Width, article.Image.Width)
	s.Assert().Equal(s.article.Image.Height, article.Image.Height)
	s.Assert().Equal(s.article.Image.Thumbnail.ContentUrl, article.Image.Thumbnail.ContentUrl)
	s.Assert().Equal(s.article.Image.Thumbnail.Width, article.Image.Thumbnail.Width)
	s.Assert().Equal(s.article.Image.Thumbnail.Height, article.Image.Thumbnail.Height)

}

func TestArticle(t *testing.T) {
	suite.Run(t, new(articleTestSuite))
}

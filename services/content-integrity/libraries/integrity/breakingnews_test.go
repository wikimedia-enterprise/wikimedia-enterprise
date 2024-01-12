package integrity_test

import (
	"context"
	"testing"
	"time"
	"wikimedia-enterprise/services/content-integrity/config/env"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"
	"wikimedia-enterprise/services/content-integrity/libraries/integrity"

	"github.com/stretchr/testify/suite"
)

func (c *collectorMock) GetIsBreakingNews(ctx context.Context, prj string, idr int) (bool, error) {
	return true, nil
}

// SetIsBreakingNews method for setting if an article is already breaking news
func (c *collectorMock) SetIsBreakingNews(ctx context.Context, prj string, idr int) error {
	return nil
}

type breakingNewsTestSuite struct {
	suite.Suite
	ctx context.Context
	bkn *integrity.BreakingNews
	art *integrity.Article
	aps *integrity.ArticleParams
	clm collectorMock
	err error
	ibn bool
}

func (s *breakingNewsTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.bkn = &integrity.BreakingNews{
		Env: &env.Environment{
			BreakingNewsMandatoryTemplates: []string{"Template:Cite news"},
			BreakingNewsTemplatesPrefix:    []string{"Template:Current"},
			BreakingNewsCreatedHours:       24,
			BreakingNewsMovedHours:         2,
			BreakingNewsUniqueEditors:      1,
			BreakingNewsUniqueEditorsHours: 1,
			BreakingNewsEdits:              5,
		},
		Collector: &s.clm,
	}
}

func (s *breakingNewsTestSuite) TestCheckArticle() {
	err := s.bkn.CheckArticle(s.ctx, s.art, s.aps)

	if s.err != nil {
		s.Equal(s.err, err)
	} else {
		s.NoError(err)
	}

	s.Equal(s.ibn, s.art.GetIsBreakingNews())
}

func TestBreakingNews(t *testing.T) {
	tmn := func() *time.Time {
		tmn := time.Now()
		return &tmn
	}
	dtn := tmn()

	tmt := func(dt string) *time.Time {
		dte, _ := time.Parse(time.RFC3339Nano, dt)

		return &dte
	}

	for _, testCase := range []*breakingNewsTestSuite{
		{
			art: new(integrity.Article),
			aps: new(integrity.ArticleParams),
			err: integrity.ErrEmptyDate,
			ibn: false,
		},
		{
			art: &integrity.Article{
				DateCreated: tmn(),
			},
			aps: &integrity.ArticleParams{
				Project:      "enwiki",
				Identifier:   100,
				Templates:    []string{"Template:Cite news", "Template:Current Test1"},
				DateModified: *tmn(),
			},
			ibn: true,
		},
		{
			art: &integrity.Article{
				DateCreated: tmn(),
				Versions: []*collector.Version{
					{
						Identifier:  1,
						Editor:      "Editor1",
						DateCreated: tmn(),
					},
					{
						Identifier:  2,
						Editor:      "Editor1",
						DateCreated: tmn(),
					},
					{
						Identifier:  3,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
					{
						Identifier:  4,
						Editor:      "Editor1",
						DateCreated: tmn(),
					},
					{
						Identifier:  5,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
				},
			},
			aps: &integrity.ArticleParams{
				Project:           "enwiki",
				Identifier:        100,
				VersionIdentifier: 4,
				Templates:         []string{"Template:Cite news"},
				DateModified:      *tmn(),
			},
			ibn: true,
		},
		{
			art: &integrity.Article{
				DateCreated: tmn(),
				Versions: []*collector.Version{
					{
						Identifier:  1,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
					{
						Identifier:  2,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
					{
						Identifier:  3,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
					{
						Identifier:  4,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
					{
						Identifier:  5,
						Editor:      "Editor2",
						DateCreated: tmn(),
					},
				},
				DateNamespaceMoved: tmn(),
			},
			aps: &integrity.ArticleParams{
				Project:           "enwiki",
				Identifier:        100,
				VersionIdentifier: 1,
				Templates:         []string{"Template:Cite news"},
				DateModified:      *tmn(),
			},
			ibn: true,
		},
		{
			art: &integrity.Article{
				DateCreated: tmn(),
				Versions: collector.Versions{
					{
						Identifier:  1,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
					{
						Identifier:  2,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
					{
						Identifier:  3,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
					{
						Identifier:  4,
						Editor:      "Editor2",
						DateCreated: dtn,
					},

					{
						Identifier:  5,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
				},
			},
			aps: &integrity.ArticleParams{
				Project:           "enwiki",
				Identifier:        100,
				VersionIdentifier: 5,
				Templates:         []string{"Template:Cite news"},
				DateModified:      *tmn(),
			},
		},
		{
			art: &integrity.Article{
				DateCreated: tmn(),
				Versions: collector.Versions{
					{
						Identifier:  1,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
					{
						Identifier:  2,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
					{
						Identifier:  3,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
					{
						Identifier:  4,
						Editor:      "Editor2",
						DateCreated: dtn,
					},

					{
						Identifier:  5,
						Editor:      "Editor2",
						DateCreated: dtn,
					},
				},
			},
			aps: &integrity.ArticleParams{
				Project:           "enwiki",
				Identifier:        100,
				VersionIdentifier: 5,
				Templates:         []string{"Template:Cite news"},
				DateModified:      *tmn(),
			},
		},
		{
			art: &integrity.Article{
				DateNamespaceMoved: tmt("2023-06-01T16:10:16Z"),
				Versions: collector.Versions{
					{
						Identifier:  1158042592,
						Editor:      "Editor1",
						DateCreated: tmt("2023-06-01T16:23:33Z"),
					},
					{
						Identifier:  1158041086,
						Editor:      "Editor2",
						DateCreated: tmt("2023-06-01T16:11:14Z"),
					},
				},
			},
			aps: &integrity.ArticleParams{
				Project:           "enwiki",
				Identifier:        13231,
				VersionIdentifier: 1158042592,
				Templates:         []string{"Template:Cite news"},
				DateModified:      *tmt("2023-06-01T16:23:33Z"),
			},
			ibn: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

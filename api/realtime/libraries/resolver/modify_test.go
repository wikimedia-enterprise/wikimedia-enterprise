package resolver_test

import (
	"testing"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type modifierTestSuite struct {
	suite.Suite
	res    *resolver.Resolver
	input  *schema.Article
	output *schema.Article
	fields []string
}

func (s *modifierTestSuite) SetupTest() {
	var err error
	s.res, err = resolver.New(new(schema.Article))

	s.Assert().NoError(err)
	s.Assert().NotNil(s.res)
}

func (s *modifierTestSuite) TestNewModifier() {
	s.Assert().NotNil(s.res.NewModifier(s.fields))
}

func (s *modifierTestSuite) TestApplyModify() {
	mod := s.res.NewModifier(s.fields)
	mod(s.input)
	s.Assert().Equal(s.output, s.input)
}

func TestModify(t *testing.T) {
	for _, testCase := range []*modifierTestSuite{
		{
			fields: []string{"name"},
			input: &schema.Article{
				Name:      "new_article",
				IsPartOf:  &schema.Project{Identifier: "enwiki"},
				Event:     &schema.Event{Type: "update"},
				Version:   &schema.Version{Noindex: false},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			output: &schema.Article{
				Name: "new_article",
			},
		},
		{
			fields: []string{},
			input: &schema.Article{
				Name:      "new_article",
				IsPartOf:  &schema.Project{Identifier: "enwiki"},
				Event:     &schema.Event{Type: "update"},
				Version:   &schema.Version{Noindex: false},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			output: &schema.Article{},
		},
		{
			fields: []string{"version.identifier"},
			input: &schema.Article{
				Name:      "new_article",
				IsPartOf:  &schema.Project{Identifier: "enwiki"},
				Event:     &schema.Event{Type: "update"},
				Version:   &schema.Version{Noindex: false},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			output: &schema.Article{
				Version: &schema.Version{},
			},
		},
		{
			fields: []string{"version.editor.identifier"},
			input: &schema.Article{
				Name:     "new_article",
				IsPartOf: &schema.Project{Identifier: "enwiki"},
				Event:    &schema.Event{Type: "update"},
				Version: &schema.Version{Noindex: false,
					Editor: &schema.Editor{Identifier: 99}},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			output: &schema.Article{
				Version: &schema.Version{
					Editor: &schema.Editor{
						Identifier: 99,
					},
				},
			},
		},
		{
			fields: []string{"is_part_of.identifier", "event.*"},
			input: &schema.Article{
				Name:     "new_article",
				IsPartOf: &schema.Project{Identifier: "enwiki"},
				Event:    &schema.Event{Type: "update"},
				Version: &schema.Version{Noindex: false,
					Editor: &schema.Editor{Identifier: 99}},
				Namespace: &schema.Namespace{Identifier: 0},
			},
			output: &schema.Article{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Event: &schema.Event{
					Type: "update",
				},
			},
		},
		{
			fields: []string{"is_part_of.identifier", "event.*", "categories.url"},
			input: &schema.Article{
				Name:     "new_article",
				IsPartOf: &schema.Project{Identifier: "enwiki"},
				Event:    &schema.Event{Type: "update"},
				Version: &schema.Version{Noindex: false,
					Editor: &schema.Editor{Identifier: 99}},
				Namespace: &schema.Namespace{Identifier: 0},
				Categories: []*schema.Category{
					{
						Name: "cat1",
						URL:  "http:/cat1.url",
					},
					{
						Name: "cat2",
						URL:  "http:/cat2.url",
					},
				},
			},
			output: &schema.Article{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Event: &schema.Event{
					Type: "update",
				},
				Categories: []*schema.Category{
					{
						URL: "http:/cat1.url",
					},
					{
						URL: "http:/cat2.url",
					},
				},
			},
		},
		{
			fields: []string{"is_part_of.identifier", "event.*", "categories.*"},
			input: &schema.Article{
				Name:     "new_article",
				IsPartOf: &schema.Project{Identifier: "enwiki"},
				Event:    &schema.Event{Type: "update"},
				Version: &schema.Version{Noindex: false,
					Editor: &schema.Editor{Identifier: 99}},
				Namespace: &schema.Namespace{Identifier: 0},
				Categories: []*schema.Category{
					{
						Name: "cat1",
						URL:  "http:/cat1.url",
					},
					{
						Name: "cat2",
						URL:  "http:/cat2.url",
					},
				},
			},
			output: &schema.Article{
				IsPartOf: &schema.Project{
					Identifier: "enwiki",
				},
				Event: &schema.Event{
					Type: "update",
				},
				Categories: []*schema.Category{
					{
						Name: "cat1",
						URL:  "http:/cat1.url",
					},
					{
						Name: "cat2",
						URL:  "http:/cat2.url",
					},
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

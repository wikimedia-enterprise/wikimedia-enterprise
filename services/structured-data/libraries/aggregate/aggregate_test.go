package aggregate_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type aggregationPageTestSuite struct {
	suite.Suite
	agg *aggregate.Aggregation
	pge *wmf.Page
}

func (s *aggregationPageTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		Page: s.pge,
	}
}

func (s *aggregationPageTestSuite) TestGetPage() {
	s.Assert().Equal(s.pge, s.agg.GetPage())
}

func (s *aggregationPageTestSuite) TestGetPageMissing() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Missing, s.agg.GetPageMissing())
	} else {
		s.Assert().Equal(true, s.agg.GetPageMissing())
	}
}

func (s *aggregationPageTestSuite) TestGetPageTitle() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Title, s.agg.GetPageTitle())
	} else {
		s.Assert().Equal("", s.agg.GetPageTitle())
	}
}

func (s *aggregationPageTestSuite) TestGetPageLanguage() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.PageLanguage, s.agg.GetPageLanguage())
	} else {
		s.Assert().Equal("", s.agg.GetPageLanguage())
	}
}

func (s *aggregationPageTestSuite) TestGetPageNs() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Ns, s.agg.GetPageNs())
	} else {
		s.Assert().Equal(0, s.agg.GetPageNs())
	}
}

func (s *aggregationPageTestSuite) TestGetPageWatchers() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Watchers, s.agg.GetPageWatchers())
	} else {
		s.Assert().Equal(0, s.agg.GetPageWatchers())
	}
}

func (s *aggregationPageTestSuite) TestGetPageID() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.PageID, s.agg.GetPageID())
	} else {
		s.Assert().Equal(0, s.agg.GetPageID())
	}
}

func (s *aggregationPageTestSuite) TestGetPageLastRevID() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.LastRevID, s.agg.GetPageLastRevID())
	} else {
		s.Assert().Equal(0, s.agg.GetPageLastRevID())
	}
}

func (s *aggregationPageTestSuite) TestGetPageFlagged() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Flagged, s.agg.GetPageFlagged())
	} else {
		s.Assert().Nil(s.agg.GetPageFlagged())
	}
}

func (s *aggregationPageTestSuite) TestGetPageFlaggedStableRevID() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Flagged.StableRevID, s.agg.GetPageFlaggedStableRevID())
	} else {
		s.Assert().Equal(0, s.agg.GetPageFlaggedStableRevID())
	}
}

func (s *aggregationPageTestSuite) TestGetPageCanonicalURL() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.CanonicalURL, s.agg.GetPageCanonicalURL())
	} else {
		s.Assert().Equal("", s.agg.GetPageCanonicalURL())
	}
}

func (s *aggregationPageTestSuite) TestGetPagePropsWikiBaseItem() {
	if s.pge != nil && s.pge.PageProps != nil {
		s.Assert().Equal(s.pge.PageProps.WikiBaseItem, s.agg.GetPagePropsWikiBaseItem())
	} else {
		s.Assert().Equal("", s.agg.GetPagePropsWikiBaseItem())
	}
}

func (s *aggregationPageTestSuite) TestGetPageWbEntityUsage() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.WbEntityUsage, s.agg.GetPageWbEntityUsage())
	} else {
		s.Assert().Nil(s.agg.GetPageWbEntityUsage())
	}
}

func (s *aggregationPageTestSuite) TestGetPageProtection() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Protection, s.agg.GetPageProtection())
	} else {
		s.Assert().Nil(s.agg.GetPageProtection())
	}
}

func (s *aggregationPageTestSuite) TestGetPageCategories() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Categories, s.agg.GetPageCategories())
	} else {
		s.Assert().Nil(s.agg.GetPageCategories())
	}
}

func (s *aggregationPageTestSuite) TestGetPageTemplates() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Templates, s.agg.GetPageTemplates())
	} else {
		s.Assert().Nil(s.agg.GetPageTemplates())
	}
}

func (s *aggregationPageTestSuite) TestGetPageRedirects() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Redirects, s.agg.GetPageRedirects())
	} else {
		s.Assert().Nil(s.agg.GetPageRedirects())
	}
}

func (s *aggregationPageTestSuite) TestGetPageOriginalImage() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Original, s.agg.GetPageOriginalImage())
	} else {
		s.Assert().Nil(s.agg.GetPageOriginalImage())
	}
}

func (s *aggregationPageTestSuite) TestGetPageThumbnailImage() {
	if s.pge != nil {
		s.Assert().Equal(s.pge.Thumbnail, s.agg.GetPageThumbnailImage())
	} else {
		s.Assert().Nil(s.agg.GetPageThumbnailImage())
	}
}

func TestPageAggregation(t *testing.T) {
	for _, testCase := range []*aggregationPageTestSuite{
		{
			pge: nil,
		},
		{
			pge: &wmf.Page{
				Title:        "Test title",
				PageLanguage: "en",
				Missing:      false,
				Watchers:     100,
				PageID:       99,
				LastRevID:    77,
				Ns:           10,
				Flagged: &wmf.Flagged{
					StableRevID: 76,
				},
				CanonicalURL: "http://localhost:8080/hello",
				PageProps: &wmf.PageProps{
					WikiBaseItem: "Q2",
				},
				Protection: []*wmf.Protection{
					{
						Type:   "main",
						Level:  "max",
						Expiry: "soon",
					},
				},
				Categories: []*wmf.Category{
					{
						Title: "Hello",
					},
				},
				Templates: []*wmf.Template{
					{
						Title: "Hello",
					},
				},
				Redirects: []*wmf.Redirect{
					{
						Title: "Hello",
					},
				},
				Original: &wmf.Image{
					Source: "https://localhost/original.jpg",
					Width:  1000,
					Height: 1000,
				},
				Thumbnail: &wmf.Image{
					Source: "https://localhost/thumb.jpg",
					Width:  100,
					Height: 100,
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationPageHTMLTestSuite struct {
	suite.Suite
	phl *wmf.PageHTML
	agg *aggregate.Aggregation
}

func (s *aggregationPageHTMLTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		PageHTML: s.phl,
	}
}

func (s *aggregationPageHTMLTestSuite) TestGetPageHTML() {
	if s.phl != nil {
		s.Assert().Equal(s.phl, s.agg.GetPageHTML())
	} else {
		s.Assert().Nil(s.agg.GetPageHTML())
	}
}

func (s *aggregationPageHTMLTestSuite) TestGetPageHTMLContent() {
	if s.phl != nil {
		s.Assert().Equal(s.phl.Content, s.agg.GetPageHTMLContent())
	} else {
		s.Assert().Equal("", s.agg.GetPageHTMLContent())
	}
}

func (s *aggregationPageHTMLTestSuite) TestGetPageHTMLError() {
	if s.phl != nil {
		s.Assert().Equal(s.phl.Error, s.agg.GetPageHTMLError())
	} else {
		s.Assert().Nil(s.agg.GetPageHTMLError())
	}
}

func TestPageHTMLAggregation(t *testing.T) {
	for _, testCase := range []*aggregationPageHTMLTestSuite{
		{
			phl: nil,
		},
		{
			phl: &wmf.PageHTML{
				Title:   "New",
				Content: "<p>...html goes here...</p>",
				Error:   errors.New("test error"),
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationUserTestSuite struct {
	suite.Suite
	usr *wmf.User
	agg *aggregate.Aggregation
}

func (s *aggregationUserTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		User: s.usr,
	}
}

func (s *aggregationUserTestSuite) TestGetUser() {
	if s.usr != nil {
		s.Assert().Equal(s.usr, s.agg.GetUser())
	} else {
		s.Assert().Nil(s.agg.GetUser())
	}
}

func (s *aggregationUserTestSuite) TestGetUserRegistration() {
	if s.usr != nil {
		s.Assert().Equal(s.usr.Registration, s.agg.GetUserRegistration())
	} else {
		s.Assert().Nil(s.agg.GetUserRegistration())
	}
}

func (s *aggregationUserTestSuite) TestGetUserEditCount() {
	if s.usr != nil {
		s.Assert().Equal(s.usr.EditCount, s.agg.GetUserEditCount())
	} else {
		s.Assert().Equal(0, s.agg.GetUserEditCount())
	}
}

func (s *aggregationUserTestSuite) TestGetUserGroups() {
	if s.usr != nil {
		s.Assert().Equal(s.usr.Groups, s.agg.GetUserGroups())
	} else {
		s.Assert().Nil(s.agg.GetUserGroups())
	}
}

func TestUserAggregation(t *testing.T) {
	for _, testCase := range []*aggregationUserTestSuite{
		{
			usr: nil,
		},
		{
			usr: &wmf.User{
				Registration: &time.Time{},
				EditCount:    10,
				Groups:       []string{"admin"},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationScoreTestSuite struct {
	suite.Suite
	scr *wmf.Score
	agg *aggregate.Aggregation
}

func (s *aggregationScoreTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		Score: s.scr,
	}
}

func (s *aggregationScoreTestSuite) TestGetScore() {
	if s.scr != nil {
		s.Assert().Equal(s.scr, s.agg.GetScore())
	} else {
		s.Assert().Nil(s.agg.GetScore())
	}
}

func (s *aggregationScoreTestSuite) TestRevertScore() {
	if s.scr != nil && s.scr.Output != nil {
		s.Assert().Equal(s.scr, s.agg.GetScore())
	} else {
		s.Assert().Nil(s.agg.GetScore())
	}
}

func TestScoreAggregation(t *testing.T) {
	for _, testCase := range []*aggregationScoreTestSuite{
		{
			scr: nil,
		},
		{
			scr: &wmf.Score{
				Output: &wmf.LiftWingScore{
					Prediction: false,
					Probability: &wmf.BooleanProbability{
						True:  0.3,
						False: 0.9,
					},
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationFirstRevisionTestSuite struct {
	suite.Suite
	rvs *wmf.Revision
	agg *aggregate.Aggregation
}

func (s *aggregationFirstRevisionTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		Revision: s.rvs,
	}
}

func (s *aggregationFirstRevisionTestSuite) TestGetFirstRevision() {
	if s.rvs != nil {
		s.Assert().Equal(s.rvs, s.agg.GetFirstRevision())
	} else {
		s.Assert().Nil(s.agg.GetFirstRevision())
	}
}

func (s *aggregationFirstRevisionTestSuite) TestGetFirstRevisionTimestamp() {
	if s.rvs != nil {
		s.Assert().Equal(s.rvs.Timestamp, s.agg.GetFirstRevisionTimestamp())
	} else {
		s.Assert().Nil(s.agg.GetFirstRevisionTimestamp())
	}
}

func (s *aggregationFirstRevisionTestSuite) TestGetFirstRevisionContent() {
	if s.rvs != nil && s.rvs.Slots != nil {
		s.Assert().Equal(s.rvs.Slots.Main.Content, s.agg.GetFirstRevisionContent())
	} else {
		s.Assert().Equal("", s.agg.GetFirstRevisionContent())
	}
}

func TestFirstRevisionAggregation(t *testing.T) {
	for _, testCase := range []*aggregationFirstRevisionTestSuite{
		{
			rvs: nil,
		},
		{
			rvs: &wmf.Revision{
				Timestamp: &time.Time{},
				Slots: &wmf.Slots{
					Main: &wmf.Main{
						Content: "...wikitext goes here...",
					},
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationCurrentRevisionTestSuite struct {
	suite.Suite
	rvs []*wmf.Revision
	agg *aggregate.Aggregation
}

func (s *aggregationCurrentRevisionTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		Page: &wmf.Page{
			Revisions: s.rvs,
		},
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevision() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0], s.agg.GetCurrentRevision())
	} else {
		s.Assert().Nil(s.agg.GetCurrentRevision())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionTags() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].Tags, s.agg.GetCurrentRevisionTags())
	} else {
		s.Assert().Nil(s.agg.GetCurrentRevisionTags())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionMinor() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].Minor, s.agg.GetCurrentRevisionMinor())
	} else {
		s.Assert().False(s.agg.GetCurrentRevisionMinor())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionComment() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].Comment, s.agg.GetCurrentRevisionComment())
	} else {
		s.Assert().Equal("", s.agg.GetCurrentRevisionComment())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionUser() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].User, s.agg.GetCurrentRevisionUser())
	} else {
		s.Assert().Equal("", s.agg.GetCurrentRevisionUser())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionUserID() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].UserID, s.agg.GetCurrentRevisionUserID())
	} else {
		s.Assert().Equal(0, s.agg.GetCurrentRevisionUserID())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionTimestamp() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].Timestamp, s.agg.GetCurrentRevisionTimestamp())
	} else {
		s.Assert().Nil(s.agg.GetCurrentRevisionTimestamp())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionContent() {
	if len(s.rvs) > 0 && s.rvs[0].Slots != nil {
		s.Assert().Equal(s.rvs[0].Slots.Main.Content, s.agg.GetCurrentRevisionContent())
	} else {
		s.Assert().Equal("", s.agg.GetCurrentRevisionContent())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionContentModel() {
	if len(s.rvs) > 0 && s.rvs[0].Slots != nil {
		s.Assert().Equal(s.rvs[0].Slots.Main.ContentModel, s.agg.GetCurrentRevisionContentModel())
	} else {
		s.Assert().Equal("", s.agg.GetCurrentRevisionContentModel())
	}
}

func (s *aggregationCurrentRevisionTestSuite) TestGetCurrentRevisionParentID() {
	if len(s.rvs) > 0 {
		s.Assert().Equal(s.rvs[0].ParentID, s.agg.GetCurrentRevisionParentID())
	} else {
		s.Assert().Equal(0, s.agg.GetCurrentRevisionParentID())
	}
}

func TestCurrentRevisionAggregation(t *testing.T) {
	for _, testCase := range []*aggregationCurrentRevisionTestSuite{
		{
			rvs: nil,
		},
		{
			rvs: []*wmf.Revision{
				{
					Tags:      []string{"hello", "world"},
					Minor:     true,
					Comment:   "...comment goes here...",
					User:      "unknown",
					UserID:    96,
					Timestamp: &time.Time{},
					Slots: &wmf.Slots{
						Main: &wmf.Main{
							Content:      "...content goes here...",
							ContentModel: "wikitext",
						},
					},
					ParentID: 65,
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregationGetPreviousRevisionTestSuite struct {
	suite.Suite
	rvs []*wmf.Revision
	agg *aggregate.Aggregation
}

func (s *aggregationGetPreviousRevisionTestSuite) SetupTest() {
	s.agg = &aggregate.Aggregation{
		Page: &wmf.Page{
			Revisions: s.rvs,
		},
	}
}

func (s *aggregationGetPreviousRevisionTestSuite) TestGetPreviousRevision() {
	if len(s.rvs) > 1 {
		s.Assert().Equal(s.rvs[1], s.agg.GetPreviousRevision())
	} else {
		s.Assert().Nil(s.agg.GetPreviousRevision())
	}
}

func (s *aggregationGetPreviousRevisionTestSuite) TestGetPreviousRevisionID() {
	if len(s.rvs) > 1 {
		s.Assert().Equal(s.rvs[1].RevID, s.agg.GetPreviousRevisionID())
	} else {
		s.Assert().Equal(0, s.agg.GetPreviousRevisionID())
	}
}

func (s *aggregationGetPreviousRevisionTestSuite) TestGetPreviousRevisionContent() {
	if len(s.rvs) > 1 && s.rvs[1].Slots != nil {
		s.Assert().Equal(s.rvs[1].Slots.Main.Content, s.agg.GetPreviousRevisionContent())
	} else {
		s.Assert().Equal("", s.agg.GetPreviousRevisionContent())
	}
}

func (s *aggregationGetPreviousRevisionTestSuite) TestGetPreviousRevisionTimestamp() {
	if len(s.rvs) > 1 {
		s.Assert().Equal(s.rvs[1].Timestamp, s.agg.GetPreviousRevisionTimestamp())
	} else {
		s.Assert().Nil(s.agg.GetPreviousRevisionTimestamp())
	}
}

func TestPreviousRevisionAggregation(t *testing.T) {
	for _, testCase := range []*aggregationGetPreviousRevisionTestSuite{
		{
			rvs: nil,
		},
		{
			rvs: []*wmf.Revision{
				{},
				{
					Tags:      []string{"hello", "world"},
					Minor:     true,
					Comment:   "...comment goes here...",
					User:      "unknown",
					UserID:    96,
					Timestamp: &time.Time{},
					Slots: &wmf.Slots{
						Main: &wmf.Main{
							Content: "...content goes here...",
						},
					},
					ParentID: 65,
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type aggregateAPIMock struct {
	wmf.API
}

type aggregateGetterMock struct {
	mock.Mock
}

func (m *aggregateGetterMock) SetAPI(api wmf.API) {
	m.Called(api)
}

func (m *aggregateGetterMock) GetData(_ context.Context, dtb string, tls []string) (interface{}, error) {
	ags := m.Called(dtb, tls)
	return ags.Get(0), ags.Error(1)
}

type aggregateTestSuite struct {
	suite.Suite
	ctx context.Context
	agg *aggregate.Aggregate
	gtr *aggregateGetterMock
	dtb string
	ttl string
	tls []string
	rsp interface{}
	err error
}

func (s *aggregateTestSuite) SetupTest() {
	amc := new(aggregateAPIMock)

	s.ctx = context.Background()
	s.gtr = new(aggregateGetterMock)
	s.agg = &aggregate.Aggregate{
		API: amc,
	}

	s.gtr.On("SetAPI", amc).Return()
}

func (s *aggregateTestSuite) TestGetAggregation() {
	s.gtr.On("GetData", s.dtb, []string{s.ttl}).Return(s.rsp, s.err)

	agr := new(aggregate.Aggregation)
	err := s.agg.GetAggregation(s.ctx, s.dtb, s.ttl, agr, s.gtr)

	if s.err != nil {
		s.Assert().Equal(s.err, err)
		s.Assert().Nil(agr.Page)
		s.Assert().Nil(agr.PageHTML)
		s.Assert().Nil(agr.Revision)
		s.Assert().Nil(agr.User)
		s.Assert().Nil(agr.Score)
	} else {
		s.Assert().NoError(err)
	}
}

func (s *aggregateTestSuite) TestGetAggregations() {
	s.gtr.On("GetData", s.dtb, s.tls).Return(s.rsp, s.err)

	ars := map[string]*aggregate.Aggregation{}
	err := s.agg.GetAggregations(s.ctx, s.dtb, s.tls, ars, s.gtr)

	if s.err != nil {
		s.Assert().Equal(s.err, err)

		for _, agr := range ars {
			s.Assert().Nil(agr.Page)
			s.Assert().Nil(agr.PageHTML)
			s.Assert().Nil(agr.Revision)
			s.Assert().Nil(agr.User)
			s.Assert().Nil(agr.Score)
		}
	} else {
		s.Assert().NoError(err)
	}
}

func TestAggregate(t *testing.T) {
	for _, testCase := range []*aggregateTestSuite{
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.Page{
				"Earth": {},
				"Ninja": {},
			},
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.Page{
				"Earth": {},
				"Ninja": {},
			},
			err: errors.New("this is test err"),
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.PageHTML{
				"Earth": {},
				"Ninja": {},
			},
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.PageHTML{
				"Earth": {},
				"Ninja": {},
			},
			err: errors.New("this is test err"),
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.Revision{
				"Earth": {},
				"Ninja": {},
			},
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.Revision{
				"Earth": {},
				"Ninja": {},
			},
			err: errors.New("this is test err"),
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.User{
				"Earth": {},
				"Ninja": {},
			},
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.User{
				"Earth": {},
				"Ninja": {},
			},
			err: errors.New("this is test err"),
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.Score{
				"Earth": {},
				"Ninja": {},
			},
		},
		{
			dtb: "enwiki",
			ttl: "Earth",
			tls: []string{"Earth", "Ninja"},
			rsp: map[string]*wmf.Score{
				"Earth": {},
				"Ninja": {},
			},
			err: errors.New("this is test err"),
		},
	} {
		suite.Run(t, testCase)
	}
}

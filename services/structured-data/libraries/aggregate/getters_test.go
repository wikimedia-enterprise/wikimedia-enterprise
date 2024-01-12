package aggregate_test

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"testing"
	"time"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type defaultGetterTestSuite struct {
	suite.Suite
	dgr *aggregate.DefaultGetter
	api wmf.API
}

func (s *defaultGetterTestSuite) SetupSuite() {
	s.dgr = new(aggregate.DefaultGetter)
}

func (s *defaultGetterTestSuite) TestSetAPI() {
	s.dgr.SetAPI(s.api)
	s.Assert().Equal(s.api, s.dgr.API)
}

func TestDefaultGetter(t *testing.T) {
	for _, testCase := range []*defaultGetterTestSuite{
		{
			api: wmf.NewAPI(),
		},
		{
			api: nil,
		},
	} {
		suite.Run(t, testCase)
	}
}

type wmfAPIMock struct {
	wmf.API
	mock.Mock
}

func (m *wmfAPIMock) GetPages(_ context.Context, dtb string, tls []string, _ ...func(*url.Values)) (map[string]*wmf.Page, error) {
	ags := m.Called(dtb, tls)

	return ags.Get(0).(map[string]*wmf.Page), ags.Error(1)
}

func (m *wmfAPIMock) GetPage(_ context.Context, dtb string, ttl string, _ ...func(*url.Values)) (*wmf.Page, error) {
	ags := m.Called(dtb, ttl)

	return ags.Get(0).(*wmf.Page), ags.Error(1)
}

func (m *wmfAPIMock) GetPagesHTML(_ context.Context, dtb string, tls []string, mxc int, _ ...func(*url.Values)) map[string]*wmf.PageHTML {
	ags := m.Called(dtb, tls, mxc)

	return ags.Get(0).(map[string]*wmf.PageHTML)
}

func (m *wmfAPIMock) GetUsers(_ context.Context, dtb string, ids []int, _ ...func(*url.Values)) (map[int]*wmf.User, error) {
	ags := m.Called(dtb, ids)

	return ags.Get(0).(map[int]*wmf.User), ags.Error(1)
}

func (m *wmfAPIMock) GetScore(_ context.Context, rvd int, lng string, prj string, mdl string) (*wmf.Score, error) {
	ags := m.Called(rvd, lng, mdl)
	return ags.Get(0).(*wmf.Score), ags.Error(1)
}

type pagesGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	pgg *aggregate.PagesGetter
	amc *wmfAPIMock
	tls []string
	dtb string
	pgs map[string]*wmf.Page
	rvl int
	err error
}

func (s *pagesGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pgg = &aggregate.PagesGetter{
		RevisionsLimit: s.rvl,
		MaxConcurrency: 5,
	}
	s.amc = new(wmfAPIMock)

	if s.rvl == 1 {
		s.amc.On("GetPages", s.dtb, s.tls).Return(s.pgs, s.err)
	} else {
		for _, ttl := range s.tls {
			s.amc.On("GetPage", s.dtb, ttl).Return(s.pgs[ttl], s.err)
		}
	}

	s.pgg.SetAPI(s.amc)
}

func (s *pagesGetterTestSuite) TestGetData() {
	dta, err := s.pgg.GetData(s.ctx, s.dtb, s.tls)

	if s.err != nil {
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().Equal(s.pgs, dta)
		s.Assert().NoError(err)
	}
}

func TestPagesGetter(t *testing.T) {
	for _, testCase := range []*pagesGetterTestSuite{
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: nil,
			rvl: 1,
			err: errors.New("test error"),
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: nil,
			rvl: 2,
			err: errors.New("test error"),
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: map[string]*wmf.Page{
				"Earth": {
					Title: "Earth",
				},
				"Ninja": {
					Title: "Ninja",
				},
			},
			rvl: 1,
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: map[string]*wmf.Page{
				"Earth": {
					Title: "Earth",
				},
				"Ninja": {
					Title: "Ninja",
				},
			},
			rvl: 2,
		},
	} {
		suite.Run(t, testCase)
	}
}

type withPagesTestSuite struct {
	suite.Suite
	rvl int
	mxc int
}

func (s *withPagesTestSuite) TestWithPages() {
	pgt := aggregate.WithPages(s.rvl, s.mxc)

	s.Assert().Equal(s.rvl, pgt.RevisionsLimit)
	s.Assert().Equal(s.mxc, pgt.MaxConcurrency)
}

func TestWithPages(t *testing.T) {
	for _, testCase := range []*withPagesTestSuite{
		{
			rvl: 100,
			mxc: 100,
		},
		{
			rvl: 0,
			mxc: 0,
		},
	} {
		suite.Run(t, testCase)
	}
}

type pagesHTMlGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	phg *aggregate.PagesHTMLGetter
	amc *wmfAPIMock
	phs map[string]*wmf.PageHTML
	tls []string
	dtb string
	mxc int
	hse bool
}

func (s *pagesHTMlGetterTestSuite) SetupSuite() {
	s.phg = &aggregate.PagesHTMLGetter{
		MaxConcurrency: s.mxc,
	}
	s.amc = new(wmfAPIMock)
	s.amc.On("GetPagesHTML", s.dtb, s.tls, s.mxc).Return(s.phs)

	s.phg.SetAPI(s.amc)
}

func (s *pagesHTMlGetterTestSuite) TestGetData() {
	dta, err := s.phg.GetData(s.ctx, s.dtb, s.tls)

	if s.hse {
		s.Assert().Error(err)
		s.Assert().Nil(dta)
	} else {
		s.Assert().NoError(err)
		s.Assert().Equal(s.phs, dta)
	}
}

func TestPagesHTMLGetter(t *testing.T) {
	for _, testCase := range []*pagesHTMlGetterTestSuite{
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			phs: map[string]*wmf.PageHTML{
				"Earth": {},
				"Ninja": {},
			},
			hse: false,
			mxc: 1,
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			phs: map[string]*wmf.PageHTML{
				"Earth": {
					Error: errors.New("earth error"),
				},
				"Ninja": {},
			},
			hse: false,
			mxc: 2,
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			phs: map[string]*wmf.PageHTML{
				"Earth": {
					Error: errors.New("earth error"),
				},
				"Ninja": {
					Error: errors.New("Ninja error"),
				},
			},
			mxc: 3,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type withPagesHTMLTestSuite struct {
	suite.Suite
	mxc int
}

func (s *withPagesHTMLTestSuite) TestWithPagesHTML() {
	phg := aggregate.WithPagesHTML(s.mxc)

	s.Assert().Equal(s.mxc, phg.MaxConcurrency)
}

func TestWithPagesHTML(t *testing.T) {
	for _, testCase := range []*withPagesHTMLTestSuite{
		{
			mxc: 100,
		},
		{
			mxc: 0,
		},
	} {
		suite.Run(t, testCase)
	}
}

type revisionsGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	rvg *aggregate.RevisionsGetter
	amc *wmfAPIMock
	tls []string
	dtb string
	pgs map[string]*wmf.Page
	mxc int
	err error
}

func (s *revisionsGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.rvg = &aggregate.RevisionsGetter{
		MaxConcurrency: s.mxc,
	}
	s.amc = new(wmfAPIMock)

	for _, ttl := range s.tls {
		s.amc.On("GetPage", s.dtb, ttl).Return(s.pgs[ttl], s.err)
	}

	s.rvg.SetAPI(s.amc)
}

func (s *revisionsGetterTestSuite) TestGetData() {
	dta, err := s.rvg.GetData(s.ctx, s.dtb, s.tls)

	if s.err != nil {
		s.Assert().Nil(dta)
		s.Assert().Contains(err.Error(), s.err.Error())
		return
	}

	rvs := map[string]*wmf.Revision{}

	for ttl, pge := range s.pgs {
		if len(pge.Revisions) > 0 {
			rvs[ttl] = pge.Revisions[0]
		} else {
			rvs[ttl] = nil
		}
	}

	s.Assert().Equal(rvs, dta)
}

func TestRevisionsGetter(t *testing.T) {
	for _, testCase := range []*revisionsGetterTestSuite{
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: nil,
			mxc: 1,
			err: errors.New("test error"),
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: map[string]*wmf.Page{
				"Earth": {
					Title: "Earth",
					Revisions: []*wmf.Revision{
						{},
					},
				},
				"Ninja": {
					Title: "Ninja",
					Revisions: []*wmf.Revision{
						{},
					},
				},
			},
			mxc: 1,
		},
		{
			tls: []string{"Earth", "Ninja"},
			dtb: "enwiki",
			pgs: map[string]*wmf.Page{
				"Earth": {
					Title: "Earth",
					Revisions: []*wmf.Revision{
						{},
					},
				},
				"Ninja": {
					Title:     "Ninja",
					Revisions: []*wmf.Revision{},
				},
			},
			mxc: 2,
		},
	} {
		suite.Run(t, testCase)
	}
}

type withRevisionsTestSuite struct {
	suite.Suite
	mxc int
}

func (s *withRevisionsTestSuite) TestWithRevisions() {
	rgr := aggregate.WithRevisions(s.mxc)

	s.Assert().Equal(s.mxc, rgr.MaxConcurrency)
}

func TestWithRevisions(t *testing.T) {
	for _, testCase := range []*withRevisionsTestSuite{
		{
			mxc: 1,
		},
		{
			mxc: 100,
		},
	} {
		suite.Run(t, testCase)
	}
}

type usersGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	ugr *aggregate.UsersGetter
	amc *wmfAPIMock
	dtb string
	tls []string
	urs map[string]int
	rur map[int]*wmf.User
	err error
}

func (s *usersGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.amc = new(wmfAPIMock)
	s.ugr = &aggregate.UsersGetter{
		Users: s.urs,
	}

	uis := []int{}

	for _, uid := range s.urs {
		uis = append(uis, uid)
	}

	sort.Ints(uis)

	s.amc.On("GetUsers", s.dtb, uis).Return(s.rur, s.err)

	s.ugr.SetAPI(s.amc)
}

func (s *usersGetterTestSuite) TestGetData() {
	dta, err := s.ugr.GetData(s.ctx, s.dtb, s.tls)

	if s.err != nil {
		s.Assert().Equal(s.err, err)
		s.Assert().Nil(dta)
	} else {
		rur := map[string]*wmf.User{}

		for _, ttl := range s.tls {
			rur[ttl] = s.rur[s.urs[ttl]]
		}

		s.Assert().Equal(dta, rur)
		s.Assert().NoError(err)
	}
}

func TestUsersGetter(t *testing.T) {
	for _, testCase := range []*usersGetterTestSuite{
		{
			dtb: "enwiki",
			tls: []string{"Earth", "Ninja"},
			urs: map[string]int{
				"Earth": 10,
				"Ninja": 11,
			},
			rur: map[int]*wmf.User{
				10: {},
				11: {},
			},
		},
		{
			err: errors.New("test user error"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type withUserTestSuite struct {
	suite.Suite
	ttl string
	uid int
}

func (s *withUserTestSuite) TestWithUser() {
	ugr := aggregate.WithUser(s.ttl, s.uid)

	s.Assert().Equal(map[string]int{s.ttl: s.uid}, ugr.Users)
}

func TestWithUser(t *testing.T) {
	for _, testCase := range []*withUserTestSuite{
		{
			ttl: "Earth",
			uid: 100,
		},
	} {
		suite.Run(t, testCase)
	}
}

type withUsersTestSuite struct {
	suite.Suite
	urs map[string]int
}

func (s *withUsersTestSuite) TestGetData() {
	ugr := aggregate.WithUsers(s.urs)

	s.Assert().Equal(s.urs, ugr.Users)
}

func TestWithUsers(t *testing.T) {
	for _, testCase := range []*withUsersTestSuite{
		{
			urs: map[string]int{
				"Earth": 100,
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type scoresGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	agg *aggregate.ScoreGetter
	amc *wmfAPIMock
	lng string
	rvd int
	mdl string
	tls []string
	scr *wmf.Score
	ser error
}

func (s *scoresGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.amc = new(wmfAPIMock)
	s.agg = &aggregate.ScoreGetter{
		Language:       s.lng,
		Revision:       map[string]int{s.tls[0]: s.rvd},
		Model:          s.mdl,
		RequestTimeout: time.Millisecond * 500,
	}

	s.amc.On("GetScore", s.rvd, s.lng, s.mdl).Return(s.scr, s.ser)
	s.agg.SetAPI(s.amc)
}

func (s *scoresGetterTestSuite) TestGetData() {
	scs, err := s.agg.GetData(s.ctx, s.lng, s.tls)

	if s.ser != nil {
		s.Assert().Equal(s.ser, err)
		s.Assert().Nil(scs)
	} else if scs == nil && s.ser == nil {
		s.Assert().Equal(s.ser, err)
		s.Assert().Nil(scs)
	} else {
		rsc := map[string]*wmf.Score{}

		for _, ttl := range s.tls {
			rsc[ttl] = s.scr
		}

		s.Assert().Equal(rsc, scs)
	}
}

func TestScoresGetter(t *testing.T) {
	for _, testCase := range []*scoresGetterTestSuite{
		{
			mdl: "revertrisk",
			lng: "en",
			rvd: 100,
			ser: nil,
			tls: []string{"Earth"},
			scr: &wmf.Score{
				Output: &wmf.LiftWingScore{
					Prediction: false,
					Probability: &wmf.BooleanProbability{
						True:  0.1,
						False: 0.9,
					},
				},
			},
		},
		// We want GetData to not return an error for now, not errors. No retries.
		{
			mdl: "revertrisk",
			lng: "en",
			rvd: 100,
			ser: nil, // We want GetData to not return an error for now. No retries.
			scr: nil,
			tls: []string{"Earth"},
		},
	} {
		suite.Run(t, testCase)
	}
}

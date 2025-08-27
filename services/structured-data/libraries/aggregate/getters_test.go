package aggregate_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
	"wikimedia-enterprise/services/structured-data/libraries/aggregate"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"

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

func (m *wmfAPIMock) GetRevisionsHTML(_ context.Context, dtb string, rvs []string, mxc int, _ ...func(*url.Values)) map[string]*wmf.PageHTML {
	ags := m.Called(dtb, rvs, mxc)

	return ags.Get(0).(map[string]*wmf.PageHTML)
}

func (m *wmfAPIMock) GetUsers(_ context.Context, dtb string, ids []int, _ ...func(*url.Values)) (map[int]*wmf.User, error) {
	ags := m.Called(dtb, ids)

	return ags.Get(0).(map[int]*wmf.User), ags.Error(1)
}

func (m *wmfAPIMock) GetScore(ctx context.Context, rev int, lng, prj, mdl string) (*wmf.Score, error) {
	ags := m.Called(ctx, rev, lng, prj, mdl)

	var scr *wmf.Score
	if result := ags.Get(0); result != nil {
		scr, _ = result.(*wmf.Score)
	}

	return scr, ags.Error(1)
}

func (m *wmfAPIMock) GetReferenceRiskScore(ctx context.Context, rev int, lng, prj string) (*wmf.ReferenceRiskScore, error) {
	args := m.Called(ctx, rev, lng, prj)
	return args.Get(0).(*wmf.ReferenceRiskScore), args.Error(1)
}

func (m *wmfAPIMock) GetReferenceNeedScore(ctx context.Context, rev int, lng, prj string) (*wmf.ReferenceNeedScore, error) {
	args := m.Called(ctx, rev, lng, prj)
	return args.Get(0).(*wmf.ReferenceNeedScore), args.Error(1)
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

type revisionsHTMlGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	phg *aggregate.RevisionsHTMLGetter
	amc *wmfAPIMock
	phs map[string]*wmf.PageHTML
	rvs []string
	dtb string
	mxc int
	hse bool
}

func (s *revisionsHTMlGetterTestSuite) SetupSuite() {
	s.phg = &aggregate.RevisionsHTMLGetter{
		MaxConcurrency: s.mxc,
	}
	s.amc = new(wmfAPIMock)
	s.amc.On("GetRevisionsHTML", s.dtb, s.rvs, s.mxc).Return(s.phs)

	s.phg.SetAPI(s.amc)
}

func (s *revisionsHTMlGetterTestSuite) TestGetData() {
	dta, err := s.phg.GetData(s.ctx, s.dtb, s.rvs)

	if s.hse {
		s.Assert().Error(err)
		s.Assert().Nil(dta)
	} else {
		s.Assert().NoError(err)
		s.Assert().Equal(s.phs, dta)
	}
}

func TestRevisionsHTMLGetter(t *testing.T) {
	for _, testCase := range []*revisionsHTMlGetterTestSuite{
		{
			rvs: []string{"1234", "4567"},
			dtb: "enwiki",
			phs: map[string]*wmf.PageHTML{
				"Earth": {},
				"Ninja": {},
			},
			hse: false,
			mxc: 1,
		},
		{
			rvs: []string{"1234", "4567"},
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
			rvs: []string{"1234", "4567"},
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

type withRevisionsHTMLTestSuite struct {
	suite.Suite
	mxc int
}

func (s *withRevisionsHTMLTestSuite) TestWithRevisionsHTML() {
	phg := aggregate.WithRevisionsHTML(s.mxc)

	s.Assert().Equal(s.mxc, phg.MaxConcurrency)
}

func TestWithRevisionsHTML(t *testing.T) {
	for _, testCase := range []*withRevisionsHTMLTestSuite{
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

type scoreGetterTestSuite struct {
	suite.Suite
	ctx       context.Context
	sg        *aggregate.ScoreGetter
	amc       *wmfAPIMock
	lng       string
	mdl       string
	prj       string
	revisions map[string]int
	expected  map[string]*wmf.Score
	ser       error
}

func (s *scoreGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.amc = new(wmfAPIMock)
	s.sg = &aggregate.ScoreGetter{
		Language:       s.lng,
		Revision:       s.revisions,
		Model:          s.mdl,
		Project:        s.prj,
		RequestTimeout: time.Millisecond * 500,
	}
	s.sg.API = s.amc
}

func (s *scoreGetterTestSuite) TeardownSuite() {
	s.amc.ExpectedCalls = nil
}

func (s *scoreGetterTestSuite) TestGetData() {
	// Set up mock expectations for each revision
	for title, rid := range s.revisions {
		expectedScore, hasScore := s.expected[title]
		if hasScore {
			s.amc.On("GetScore", mock.MatchedBy(func(ctx context.Context) bool { return true }), rid, s.lng, s.prj, s.mdl).Return(expectedScore, nil)
		} else {
			s.amc.On("GetScore", mock.MatchedBy(func(ctx context.Context) bool { return true }), rid, s.lng, s.prj, s.mdl).Return(nil, s.ser)
		}
	}

	scs, err := s.sg.GetData(s.ctx, s.lng, nil)

	// If there were expected scores, check the results
	if s.expected != nil {
		s.Assert().Equal(s.expected, scs, "Expected scores map does not match actual result")
	}

	// Verify the error handling
	if s.ser != nil {
		s.Assert().Error(err, "Expected an error but got nil")
		s.Assert().Contains(err.Error(), s.ser.Error(), "Expected error message to match")
	} else {
		s.Assert().NoError(err, "Expected no error but got one")
	}

	// Ensure all expectations are met
	s.amc.AssertExpectations(s.T())
}

func (s *scoreGetterTestSuite) TestGetDataWithTimeout() {
	var cancel context.CancelFunc
	s.ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond) // Very short timeout
	defer cancel()

	for title, rid := range s.revisions {
		if title == "Mars" { // Simulate timeout for "Mars"
			s.amc.On("GetScore", mock.MatchedBy(func(ctx context.Context) bool { return true }), rid, s.lng, s.prj, s.mdl).
				Run(func(args mock.Arguments) { time.Sleep(10 * time.Millisecond) }). // Delay execution
				Return(nil, context.DeadlineExceeded)
		} else { // Return a valid response for "Earth"
			expectedScore := s.expected[title]
			s.amc.On("GetScore", mock.MatchedBy(func(ctx context.Context) bool { return true }), rid, s.lng, s.prj, s.mdl).
				Return(expectedScore, nil)
		}
	}

	scs, err := s.sg.GetData(s.ctx, s.lng, nil)

	if s.expected != nil {
		s.Assert().Equal(s.expected, scs, "Expected scores map does not match actual result")
	}

	if err != nil && !strings.Contains(err.Error(), "API failure") {
		s.Assert().Contains(err.Error(), context.DeadlineExceeded.Error(), "Expected error to be a deadline exceeded error")
	} else if err != nil {
		s.Assert().Contains(err.Error(), "API failure", "Expected error to be an API failure")
	}

	s.amc.AssertExpectations(s.T())
}

func TestScoreGetter(t *testing.T) {
	for _, testCase := range []*scoreGetterTestSuite{
		{
			mdl:       "revertrisk",
			lng:       "en",
			prj:       "wikipedia",
			revisions: map[string]int{"Earth": 102, "Mars": 101},
			expected: map[string]*wmf.Score{
				"Earth": {
					Output: &wmf.LiftWingScore{
						Prediction: false,
						Probability: &wmf.BooleanProbability{
							True:  0.1,
							False: 0.9,
						},
					},
				},
				"Mars": {
					Output: &wmf.LiftWingScore{
						Prediction: false,
						Probability: &wmf.BooleanProbability{
							True:  0.1,
							False: 0.9,
						},
					},
				},
			},
			ser: nil,
		},
		{
			mdl:       "revertrisk",
			lng:       "en",
			prj:       "wikipedia",
			revisions: map[string]int{"Earth": 102, "Mars": 101},
			expected:  nil,
			ser:       context.DeadlineExceeded,
		},
		{
			mdl:       "revertrisk",
			lng:       "en",
			prj:       "wikipedia",
			revisions: map[string]int{"Earth": 102, "Mars": 101},
			expected:  nil,
			ser:       errors.New("score: API failure"),
		},
		{
			mdl:       "revertrisk",
			lng:       "en",
			prj:       "wikipedia",
			revisions: map[string]int{"Earth": 103, "Mars": 104},
			expected: map[string]*wmf.Score{
				"Earth": {
					Output: &wmf.LiftWingScore{
						Prediction: false,
						Probability: &wmf.BooleanProbability{
							True:  0.3,
							False: 0.7,
						},
					},
				},
			},
			ser: context.DeadlineExceeded,
		},
		{
			mdl:       "revertrisk",
			lng:       "en",
			prj:       "wikipedia",
			revisions: map[string]int{"Earth": 103, "Mars": 104},
			expected: map[string]*wmf.Score{
				"Mars": {
					Output: &wmf.LiftWingScore{
						Prediction: false,
						Probability: &wmf.BooleanProbability{
							True:  0.5,
							False: 0.5,
						},
					},
				},
			},
			ser: errors.New("score: API failure"),
		},
	} {
		suite.Run(t, testCase)
	}
}

type referenceNeedScoreGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	agg *aggregate.ReferenceNeedScoreGetter
	amc *wmfAPIMock
	lng string
	rvd int
	prj string
	tls []string
	scr *wmf.ReferenceNeedScore
	ser error
}

func (s *referenceNeedScoreGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.amc = new(wmfAPIMock)
	// Create the Revision map using the first title in tls.
	s.agg = &aggregate.ReferenceNeedScoreGetter{
		Language:       s.lng,
		Revision:       map[string]int{s.tls[0]: s.rvd},
		Project:        s.prj,
		RequestTimeout: time.Millisecond * 500,
	}
	// The actual call is: GetReferenceNeedScore(ctx, s.rvd, s.lng, s.prj)
	s.amc.On("GetReferenceNeedScore", mock.Anything, s.rvd, s.lng, s.prj).Return(s.scr, s.ser)
	s.agg.SetAPI(s.amc)
}

func (s *referenceNeedScoreGetterTestSuite) TestGetData() {
	scs, err := s.agg.GetData(s.ctx, "", nil)
	if s.ser != nil {
		s.Assert().Contains(err.Error(), s.ser.Error())
		s.Assert().Nil(scs)
	} else {
		expected := make(map[string]*wmf.ReferenceNeedScore)
		for _, ttl := range s.tls {
			expected[ttl] = s.scr
		}
		s.Assert().Equal(expected, scs)
	}
}

func TestReferenceNeedScoreGetter(t *testing.T) {
	for _, testCase := range []*referenceNeedScoreGetterTestSuite{
		{
			lng: "en",
			rvd: 100,
			prj: "wiki",
			tls: []string{"Earth"},
			ser: nil,
			scr: &wmf.ReferenceNeedScore{},
		},
		{
			lng: "fr",
			rvd: 101,
			prj: "wiktionary",
			tls: []string{"Earth"},
			ser: fmt.Errorf("API error"),
			scr: nil,
		},
	} {
		suite.Run(t, testCase)
	}
}

type referenceRiskScoreGetterTestSuite struct {
	suite.Suite
	ctx context.Context
	agg *aggregate.ReferenceRiskScoreGetter
	amc *wmfAPIMock
	lng string
	rvd int
	tls []string
	prj string
	scr *wmf.ReferenceRiskScore
	ser error
}

func (s *referenceRiskScoreGetterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.amc = new(wmfAPIMock)
	s.agg = &aggregate.ReferenceRiskScoreGetter{
		Language:       s.lng,
		Revision:       map[string]int{s.tls[0]: s.rvd},
		Project:        s.prj,
		RequestTimeout: time.Millisecond * 500,
	}
	s.amc.On("GetReferenceRiskScore", mock.Anything, s.rvd, s.lng, s.prj).Return(s.scr, s.ser)
	s.agg.SetAPI(s.amc)
}

func (s *referenceRiskScoreGetterTestSuite) TestGetData() {
	scs, err := s.agg.GetData(s.ctx, "", s.tls)
	if s.ser != nil {
		s.Assert().Contains(err.Error(), s.ser.Error())
		s.Assert().Nil(scs)
	} else {
		expected := make(map[string]*wmf.ReferenceRiskScore)
		for _, ttl := range s.tls {
			expected[ttl] = s.scr
		}
		s.Assert().Equal(expected, scs)
	}
}

func TestReferenceRiskScoreGetter(t *testing.T) {
	for _, testCase := range []*referenceRiskScoreGetterTestSuite{
		{
			lng: "en",
			rvd: 102,
			prj: "wiki",
			tls: []string{"Earth"},
			ser: nil,
			scr: &wmf.ReferenceRiskScore{},
		},
		{
			lng: "es",
			rvd: 103,
			prj: "wikiquote",
			tls: []string{"Earth"},
			ser: fmt.Errorf("API timeout"),
			scr: nil,
		},
	} {
		suite.Run(t, testCase)
	}
}

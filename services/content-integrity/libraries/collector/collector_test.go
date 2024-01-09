package collector_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
	"wikimedia-enterprise/services/content-integrity/config/env"
	"wikimedia-enterprise/services/content-integrity/libraries/collector"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type versionsTestSuite struct {
	suite.Suite
	vrs collector.Versions
	cir int
	ect int
}

func (s *versionsTestSuite) TestGetCurrentIdentifier() {
	s.Equal(s.cir, s.vrs.GetCurrentIdentifier())
}

func (s *versionsTestSuite) TestGetUniqueEditorsCount() {
	s.Equal(s.ect, s.vrs.GetUniqueEditorsCount())
}

func TestVersions(t *testing.T) {
	dte := time.Now()

	for _, testCase := range []*versionsTestSuite{
		{
			vrs: []*collector.Version{
				{Identifier: 1, Editor: "Test Editor 1", DateCreated: &dte},
				{Identifier: 2, Editor: "Test Editor 2", DateCreated: &dte},
				{Identifier: 3, Editor: "Test Editor 3", DateCreated: &dte},
			},
			cir: 1,
			ect: 3,
		},
		{
			vrs: []*collector.Version{},
			cir: 0,
			ect: 0,
		},
		{
			vrs: []*collector.Version{
				{Identifier: 1, Editor: "Test Editor 2", DateCreated: &dte},
				{Identifier: 2, Editor: "Test Editor 2", DateCreated: &dte},
				{Identifier: 3, Editor: "Test Editor 3", DateCreated: &dte},
			},
			cir: 1,
			ect: 2,
		},
	} {
		suite.Run(t, testCase)
	}
}

type redisMock struct {
	redis.Cmdable
	mock.Mock
}

func (m *redisMock) Get(ctx context.Context, key string) *redis.StringCmd {
	ags := m.Called(key)

	cmd := redis.NewStringCmd(ctx)
	cmd.SetVal(ags.String(0))
	cmd.SetErr(ags.Error(1))

	return cmd
}

func (m *redisMock) Set(ctx context.Context, key string, val interface{}, exp time.Duration) *redis.StatusCmd {
	ags := m.Called(key, val, exp)

	cmd := redis.NewStatusCmd(ctx)
	cmd.SetErr(ags.Error(0))

	return cmd
}

func (m *redisMock) LRange(ctx context.Context, key string, str, stp int64) *redis.StringSliceCmd {
	ags := m.Called(key, str, stp)

	cmd := redis.NewStringSliceCmd(ctx)
	cmd.SetVal(ags.Get(0).([]string))
	cmd.SetErr(ags.Error(1))

	return cmd
}

func (m *redisMock) LPush(ctx context.Context, key string, val ...interface{}) *redis.IntCmd {
	ags := m.Called(key, val[0])

	cmd := redis.NewIntCmd(ctx)
	cmd.SetErr(ags.Error(0))

	return cmd
}

func (m *redisMock) ExpireNX(ctx context.Context, key string, exp time.Duration) *redis.BoolCmd {
	ags := m.Called(key, exp)

	cmd := redis.NewBoolCmd(ctx)
	cmd.SetVal(ags.Bool(0))
	cmd.SetErr(ags.Error(1))

	return cmd
}

type articleCollectorTestSuite struct {
	suite.Suite
	env *env.Environment
	rdm *redisMock
	art collector.ArticleAPI
	vrs collector.Versions
	ctx context.Context
	idr int
	prj string
	dte *time.Time
	vsk string
	dck string
	nmk string
	ack string
	exp time.Duration
	ttl string
	nak string
}

func (s *articleCollectorTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.env = new(env.Environment)
	s.idr = 100
	s.prj = "enwiki"
	s.vsk = fmt.Sprintf("article:%s:%d:versions", s.prj, s.idr)
	s.dck = fmt.Sprintf("article:%s:%d:date_created", s.prj, s.idr)
	s.nmk = fmt.Sprintf("article:%s:%d:date_namespace_moved", s.prj, s.idr)
	s.nak = fmt.Sprintf("article:%s:%d:name", s.prj, s.idr)
	s.ack = fmt.Sprintf("article:%s:%d:is_being_tracked", s.prj, s.idr)
	s.exp = time.Hour * time.Duration(s.env.BreakingNewsKeysExpiration)
}

func (s *articleCollectorTestSuite) SetupTest() {
	now := time.Now()
	s.dte = &now
	s.rdm = new(redisMock)
	s.art = collector.NewArticle(s.rdm, s.env)
	s.vrs = []*collector.Version{
		{Identifier: 1, Editor: "Test Editor 1", DateCreated: s.dte},
		{Identifier: 2, Editor: "Test Editor 2", DateCreated: s.dte},
		{Identifier: 3, Editor: "Test Editor 3", DateCreated: s.dte},
	}
	s.ttl = "Squirrel"
}

func (s *articleCollectorTestSuite) TestGetDateCreated() {
	s.rdm.On("Get", s.dck).Return(s.dte.Format(time.RFC3339), nil)
	dcr, err := s.art.GetDateCreated(s.ctx, s.prj, s.idr)

	s.Assert().NoError(err)
	s.Assert().NotNil(dcr)
	s.Assert().Equal(s.dte.Format(time.RFC3339), dcr.Format(time.RFC3339))
}

func (s *articleCollectorTestSuite) TestGetDateCreatedError() {
	der := errors.New("GetDateCreated error")
	s.rdm.On("Get", s.dck).Return("", der)

	dcr, err := s.art.GetDateCreated(s.ctx, s.prj, s.idr)

	s.Assert().Error(err)
	s.Assert().Nil(dcr)
	s.Assert().Equal(der, err)
}

func (s *articleCollectorTestSuite) TestGetVersions() {
	vrj := []string{}

	for _, ver := range s.vrs {
		cnt, err := json.Marshal(ver)
		s.Assert().NoError(err)
		vrj = append(vrj, string(cnt))
	}

	s.rdm.On("LRange", s.vsk, int64(0), int64(-1)).Return(vrj, nil)

	cnt, err := s.art.GetVersions(s.ctx, s.prj, s.idr)

	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)

	for i, ver := range cnt {
		s.Assert().Equal(s.vrs[i].Identifier, ver.Identifier)
		s.Assert().Equal(s.vrs[i].Editor, ver.Editor)
		s.Assert().Equal(s.vrs[i].DateCreated.Unix(), ver.DateCreated.Unix())
	}
}

func (s *articleCollectorTestSuite) TestGetVersionsError() {
	gvr := errors.New("GetDateCreated error")

	s.rdm.On("LRange", s.vsk, int64(0), int64(-1)).Return([]string{}, gvr)

	cnt, err := s.art.GetVersions(s.ctx, s.prj, s.idr)

	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(gvr, err)
}

func (s *articleCollectorTestSuite) TestGetDateNamespaceMoved() {
	s.rdm.On("Get", s.nmk).Return(s.dte.Format(time.RFC3339Nano), nil)
	dnm, err := s.art.GetDateNamespaceMoved(s.ctx, s.prj, s.idr)

	s.Assert().NoError(err)
	s.Assert().NotNil(dnm)
	s.Assert().Equal(s.dte.Format(time.RFC3339Nano), dnm.Format(time.RFC3339Nano))
}

func (s *articleCollectorTestSuite) TestGetDateNamespaceMovedError() {
	nmr := errors.New("GetDateCreated error")
	s.rdm.On("Get", s.nmk).Return("", nmr)

	dcr, err := s.art.GetDateNamespaceMoved(s.ctx, s.prj, s.idr)

	s.Assert().Error(err)
	s.Assert().Nil(dcr)
	s.Assert().Equal(nmr, err)
}

func (s *articleCollectorTestSuite) TestPrependVersion() {
	nvr := &collector.Version{
		Identifier:  3,
		Editor:      "Test Editor 3",
		DateCreated: s.dte,
	}

	mrv, err := json.Marshal(nvr)
	s.Assert().NoError(err)

	s.vrs = append([]*collector.Version{nvr}, s.vrs...)

	vrj := []string{}

	for _, ver := range s.vrs {
		cnt, err := json.Marshal(ver)
		s.Assert().NoError(err)
		vrj = append(vrj, string(cnt))
	}

	s.rdm.On("Get", s.ack).Return("true", nil)
	s.rdm.On("LPush", s.vsk, mrv).Return(nil)
	s.rdm.On("ExpireNX", s.vsk, s.exp).Return(true, nil)
	s.rdm.On("LRange", s.vsk, int64(0), int64(-1)).Return(vrj, nil)

	cnt, err := s.art.PrependVersion(s.ctx, s.prj, s.idr, nvr)

	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)
	s.Assert().Equal(4, len(cnt))

	for i, ver := range cnt {
		s.Assert().Equal(s.vrs[i].Identifier, ver.Identifier)
		s.Assert().Equal(s.vrs[i].Editor, ver.Editor)
		s.Assert().Equal(s.vrs[i].DateCreated.Unix(), ver.DateCreated.Unix())
	}
}

func (s *articleCollectorTestSuite) TestPrependVersionError() {
	nvr := &collector.Version{
		Identifier:  3,
		Editor:      "Test Editor 3",
		DateCreated: s.dte,
	}
	// Get error.
	pvr := errors.New("PrependVersionError Get error")
	s.rdm.On("Get", s.ack).Return("", pvr)

	cnt, err := s.art.PrependVersion(s.ctx, s.prj, s.idr, nvr)
	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(pvr, err)

	// Lpush error.
	pvr = errors.New("PrependVersionError LPush error")
	mrv, err := json.Marshal(nvr)
	s.Assert().NoError(err)

	s.rdm = new(redisMock)
	s.art = collector.NewArticle(s.rdm, s.env)

	s.rdm.On("Get", s.ack).Return("true", nil)
	s.rdm.On("LPush", s.vsk, mrv).Return(pvr)

	cnt, err = s.art.PrependVersion(s.ctx, s.prj, s.idr, nvr)
	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(pvr, err)

	// Expire error.
	pvr = errors.New("PrependVersionError Expire error")
	s.rdm = new(redisMock)
	s.art = collector.NewArticle(s.rdm, s.env)
	s.rdm.On("Get", s.ack).Return("true", nil)
	s.rdm.On("LPush", s.vsk, mrv).Return(nil)
	s.rdm.On("ExpireNX", s.vsk, s.exp).Return(false, pvr)

	cnt, err = s.art.PrependVersion(s.ctx, s.prj, s.idr, nvr)
	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(pvr, err)

	// LRange error.
	pvr = errors.New("PrependVersionError LRange error")
	s.rdm = new(redisMock)
	s.art = collector.NewArticle(s.rdm, s.env)
	s.rdm.On("Get", s.ack).Return("true", nil)
	s.rdm.On("LPush", s.vsk, mrv).Return(nil)
	s.rdm.On("ExpireNX", s.vsk, s.exp).Return(true, nil)
	s.rdm.On("LRange", s.vsk, int64(0), int64(-1)).Return([]string{}, pvr)

	cnt, err = s.art.PrependVersion(s.ctx, s.prj, s.idr, nvr)
	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(pvr, err)
}

func (s *articleCollectorTestSuite) TestSetDateCreated() {
	s.rdm.On("Set", s.dck, s.dte.Format(time.RFC3339Nano), s.exp).Return(nil)
	s.rdm.On("Set", s.ack, true, s.exp).Return(nil)

	cnt, err := s.art.SetDateCreated(s.ctx, s.prj, s.idr, s.dte)

	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)
	s.Assert().Equal(s.formatDate(s.dte), s.formatDate(cnt))
}

func (s *articleCollectorTestSuite) TestSetDateCreatedError() {
	ser := errors.New("set datecreated error")
	acr := errors.New("set isbeingtracked error")

	s.rdm.On("Set", s.dck, s.dte.Format(time.RFC3339Nano), s.exp).Return(ser)

	cnt, err := s.art.SetDateCreated(s.ctx, s.prj, s.idr, s.dte)

	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(ser, err)

	s.rdm = new(redisMock)
	s.art = collector.NewArticle(s.rdm, s.env)
	s.rdm.On("Set", s.dck, s.dte.Format(time.RFC3339Nano), s.exp).Return(nil)
	s.rdm.On("Set", s.ack, true, s.exp).Return(acr)

	cnt, err = s.art.SetDateCreated(s.ctx, s.prj, s.idr, s.dte)

	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(acr, err)
}
func (s *articleCollectorTestSuite) TestSetDateNamespaceMoved() {
	s.rdm.On("Set", s.nmk, s.dte.Format(time.RFC3339Nano), s.exp).Return(nil)
	s.rdm.On("Set", s.ack, true, s.exp).Return(nil)

	cnt, err := s.art.SetDateNamespaceMoved(s.ctx, s.prj, s.idr, s.dte)

	s.Assert().NoError(err)
	s.Assert().NotNil(cnt)
	s.Assert().Equal(s.formatDate(s.dte), s.formatDate(cnt))
}

func (s *articleCollectorTestSuite) TestSetDateNamespaceMovedError() {
	ser := errors.New("set datecreated error")
	acr := errors.New("set isbeingtracked error")

	s.rdm.On("Set", s.nmk, s.dte.Format(time.RFC3339Nano), s.exp).Return(ser)

	cnt, err := s.art.SetDateNamespaceMoved(s.ctx, s.prj, s.idr, s.dte)

	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(ser, err)

	s.rdm = new(redisMock)
	s.art = collector.NewArticle(s.rdm, s.env)
	s.rdm.On("Set", s.nmk, s.dte.Format(time.RFC3339Nano), s.exp).Return(nil)
	s.rdm.On("Set", s.ack, true, s.exp).Return(acr)

	cnt, err = s.art.SetDateNamespaceMoved(s.ctx, s.prj, s.idr, s.dte)

	s.Assert().Error(err)
	s.Assert().Nil(cnt)
	s.Assert().Equal(acr, err)
}

func (s *articleCollectorTestSuite) TestGetIsBeingTrackedFalse() {
	s.rdm.On("Get", s.ack).Return("", nil)

	bol, err := s.art.GetIsBeingTracked(s.ctx, s.prj, s.idr)

	s.Assert().Error(err)
	s.Assert().False(bol)
}

func (s *articleCollectorTestSuite) TestGetIsBeingTrackedTrue() {
	s.rdm.On("Get", s.ack).Return("1", nil)

	bol, err := s.art.GetIsBeingTracked(s.ctx, s.prj, s.idr)

	s.Assert().Nil(err)
	s.Assert().True(bol)
}

func (s *articleCollectorTestSuite) TestGetIsBeingTrackedError() {
	acr := errors.New("set isbeingtracked error")

	s.rdm.On("Get", s.ack).Return("", acr)

	bol, err := s.art.GetIsBeingTracked(s.ctx, s.prj, s.idr)

	s.Assert().Error(err)
	s.Assert().False(bol)
}

func (s *articleCollectorTestSuite) formatDate(dat *time.Time) string {
	return dat.Format(time.RFC3339)
}

func (s *articleCollectorTestSuite) TestSetName() {
	s.rdm.On("Set", s.nak, s.ttl, s.exp).Return(nil)
	err := s.art.SetName(s.ctx, s.prj, s.idr, s.ttl)

	s.Assert().NoError(err)
}

func (s *articleCollectorTestSuite) TestSetNameError() {
	ser := errors.New("set name error")
	s.rdm.On("Set", s.nak, s.ttl, s.exp).Return(ser)
	err := s.art.SetName(s.ctx, s.prj, s.idr, s.ttl)

	s.Assert().Error(err)
	s.Assert().Equal(ser, err)
}

func (s *articleCollectorTestSuite) TestGetName() {
	s.rdm.On("Get", s.nak).Return(s.ttl, nil)
	nam, err := s.art.GetName(s.ctx, s.prj, s.idr)

	s.Assert().NoError(err)
	s.Assert().Equal(s.ttl, nam)
}

func (s *articleCollectorTestSuite) TestGetNameError() {
	ger := errors.New("get name error")
	s.rdm.On("Get", s.nak).Return("", ger)
	nam, err := s.art.GetName(s.ctx, s.prj, s.idr)

	s.Assert().Error(err)
	s.Assert().Equal(ger, err)
	s.Assert().Equal("", nam)
}

func TestArticleCollector(t *testing.T) {
	suite.Run(t, new(articleCollectorTestSuite))
}

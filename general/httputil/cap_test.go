package httputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type redisMock struct {
	redis.Cmdable
	mock.Mock
}

func (m *redisMock) Get(_ context.Context, key string) *redis.StringCmd {
	ags := m.Called(key)

	return redis.NewStringResult(
		strconv.Itoa(ags.Int(0)),
		ags.Error(1),
	)
}

func (m *redisMock) Set(_ context.Context, key string, val interface{}, exp time.Duration) *redis.StatusCmd {
	ags := m.Called(key, val, int(exp))

	return redis.NewStatusResult("", ags.Error(0))
}

func (m *redisMock) Incr(_ context.Context, key string) *redis.IntCmd {
	ags := m.Called(key)

	return redis.NewIntResult(0, ags.Error(0))
}

type capByRedisTestSuite struct {
	suite.Suite
	ctx context.Context
	cbr *CapByRedis
	idr string
	lmt int
	cnt int
	alw bool
	err error
	ger error
	ser error
	ier error
}

func (s *capByRedisTestSuite) SetupSuite() {
	key := fmt.Sprintf("%s:count", s.idr)

	cmd := new(redisMock)
	cmd.On("Get", key).Return(s.cnt, s.ger)
	cmd.On("Set", key, 1, 0).Return(s.ser)
	cmd.On("Incr", key).Return(s.ier)

	s.cbr = &CapByRedis{
		Redis: cmd,
	}
	s.ctx = context.Background()
}

func (s *capByRedisTestSuite) TestCheck() {
	alw, err := s.cbr.Check(s.ctx, s.idr, s.lmt)

	s.Assert().Equal(s.alw, alw)
	s.Assert().Equal(s.err, err)
}

func TestCapByRedis(t *testing.T) {
	ter := errors.New("error for testing")

	for _, testCase := range []*capByRedisTestSuite{
		{
			idr: "test",
			alw: true,
			cnt: 1,
			lmt: 1,
		},
		{
			idr: "test",
			alw: false,
			cnt: 11,
			lmt: 10,
		},
		{
			idr: "test",
			cnt: 0,
			lmt: 1,
			ger: redis.Nil,
			alw: true,
		},
		{
			idr: "test",
			cnt: 1,
			lmt: 2,
			ger: ter,
			err: ter,
		},
		{
			idr: "test",
			cnt: 1,
			lmt: 2,
			ger: redis.Nil,
			ser: ter,
			err: ter,
		},
		{
			idr: "test",
			cnt: 1,
			lmt: 2,
			ier: ter,
			err: ter,
		},
	} {
		suite.Run(t, testCase)
	}
}

type capMock struct {
	mock.Mock
}

func (m *capMock) Check(_ context.Context, idr string, lmt int) (bool, error) {
	ags := m.Called(idr, lmt)

	return ags.Bool(0), ags.Error(1)
}

type capTestSuite struct {
	suite.Suite
	sts int
	idr string
	lmt int
	alw bool
	err error
	cfg *CapConfig
	usr *User
	cmk *capMock
	srv *httptest.Server
}

func (s *capTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.Use(func(gcx *gin.Context) {
		gcx.Set("user", s.usr)
	})
	rtr.GET("/test", Cap(s.cmk, s.cfg), func(gcx *gin.Context) {
		gcx.Status(http.StatusOK)
	})

	return rtr
}

func (s *capTestSuite) SetupSuite() {
	s.cmk = new(capMock)
	s.cmk.On("Check", s.idr, s.lmt).Return(s.alw, s.err)

	s.srv = httptest.NewServer(s.createServer())
}

func (s *capTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *capTestSuite) TestCap() {
	res, err := http.Get(fmt.Sprintf("%s/test", s.srv.URL))
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(s.sts, res.StatusCode)

	if s.err != nil {
		dta, err := io.ReadAll(res.Body)
		s.Assert().NoError(err)
		s.Assert().Contains(string(dta), s.err.Error())
	}
}

func TestCap(t *testing.T) {
	for _, testCase := range []*capTestSuite{
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: &CapConfig{
				Limit:  10,
				Groups: []string{"group_1"},
			},
			idr: "user:test",
			lmt: 10,
		},
		{
			sts: http.StatusOK,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: &CapConfig{
				Limit:  10,
				Groups: []string{"group_2"},
			},
			idr: "user:test",
			lmt: 10,
			alw: true,
		},
		{
			sts: http.StatusUnauthorized,
			usr: nil,
			cfg: &CapConfig{
				Limit:  10,
				Groups: []string{"group_2"},
			},
			idr: "user:test",
			lmt: 10,
			alw: true,
		},
		{
			sts: http.StatusTooManyRequests,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: &CapConfig{
				Limit:  10,
				Groups: []string{"group_2"},
			},
			idr: "user:test",
			lmt: 10,
			alw: false,
		},
		{
			sts: http.StatusInternalServerError,
			usr: &User{
				Username: "test",
				Groups:   []string{"group_2"},
			},
			cfg: &CapConfig{
				Limit:  10,
				Groups: []string{"group_2"},
			},
			idr: "user:test",
			lmt: 10,
			alw: false,
			err: errors.New("this is a test error"),
		},
	} {
		suite.Run(t, testCase)
	}
}

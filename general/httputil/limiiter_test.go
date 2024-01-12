package httputil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type limiterTestSuite struct {
	suite.Suite
	lmr *Limiter
	kys map[string]int
	dta string
}

func (s *limiterTestSuite) SetupTest() {
	s.lmr = new(Limiter)
	s.kys = map[string]int{
		"group_1": 100,
		"group_2": 200,
	}
	s.dta = `{"group_1":100,"group_2":200}`
}

func (s *limiterTestSuite) TestGetSet() {
	for key, lim := range s.kys {
		s.lmr.Set(key, lim)

		lmr := s.lmr.Get(key)
		s.Assert().NotNil(lmr)
		s.Assert().Equal(lmr.GetMax(), float64(lim))
	}
}

func (s *limiterTestSuite) TestUnmarshalEnvironmentValue() {
	s.Assert().NoError(
		s.lmr.UnmarshalEnvironmentValue(s.dta),
	)

	for key, lim := range s.kys {
		lmr := s.lmr.Get(key)
		s.Assert().NotNil(lmr)
		s.Assert().Equal(lmr.GetMax(), float64(lim))
	}
}

func TestLimiter(t *testing.T) {
	suite.Run(t, new(limiterTestSuite))
}

type limitTestSuite struct {
	suite.Suite
	srv *httptest.Server
	usr *User
	lmr *Limiter
	gps map[string]int
	rqn int
	sts int
}

func (s *limitTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.Use(func(gcx *gin.Context) {
		gcx.Set("user", s.usr)
	})
	rtr.Use(Limit(s.lmr))
	rtr.GET("/test", func(gcx *gin.Context) {
		gcx.Status(http.StatusOK)
	})

	return rtr
}

func (s *limitTestSuite) SetupSuite() {
	s.lmr = new(Limiter)

	for grp, lmt := range s.gps {
		s.lmr.Set(grp, lmt)
	}

	s.srv = httptest.NewServer(s.createServer())
}

func (s *limitTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *limitTestSuite) TestLimit() {
	sts := http.StatusOK

	for i := 0; i < s.rqn; i++ {
		res, err := http.Get(fmt.Sprintf("%s/test", s.srv.URL))
		s.Assert().NoError(err)
		sts = res.StatusCode
	}

	s.Assert().Equal(s.sts, sts)
}

func TestLimit(t *testing.T) {
	for _, testCase := range []*limitTestSuite{
		{
			gps: map[string]int{
				"group_1": 50,
				"group_2": 100,
			},
			usr: &User{
				Username: "john",
				Groups:   []string{"group_2", "group_1"},
			},
			rqn: 101,
			sts: http.StatusTooManyRequests,
		},
		{
			gps: map[string]int{
				"group_1": 50,
				"group_2": 100,
			},
			usr: &User{
				Username: "john",
				Groups:   []string{"group_2", "group_1"},
			},
			rqn: 100,
			sts: http.StatusOK,
		},
		{
			gps: map[string]int{
				"group_1": 50,
				"group_2": 100,
			},
			usr: &User{
				Username: "john",
				Groups:   []string{"group_3"},
			},
			rqn: 1000,
			sts: http.StatusOK,
		},
		{
			sts: http.StatusInternalServerError,
			rqn: 1,
		},
	} {
		suite.Run(t, testCase)
	}
}

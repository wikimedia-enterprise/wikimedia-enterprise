package httputil

import (
	"fmt"
	"math"
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
	srv                      *httptest.Server
	usr                      *User
	lmr                      *Limiter
	gps                      map[string]int
	rqn                      int
	sts                      int
	statusPercentExpectation float64
	statusPercentEpsilon     float64
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

func (s *limitTestSuite) SetupTest() {
	s.lmr = new(Limiter)

	for grp, lmt := range s.gps {
		s.lmr.Set(grp, lmt)
	}

	s.srv = httptest.NewServer(s.createServer())
}

func (s *limitTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *limitTestSuite) TestLimit() {
	stsCounts := map[int]int{}

	for i := 0; i < s.rqn; i++ {
		res, err := http.Get(fmt.Sprintf("%s/test", s.srv.URL))
		s.Assert().NoError(err)
		if _, exists := stsCounts[res.StatusCode]; !exists {
			stsCounts[res.StatusCode] = 0
		}
		stsCounts[res.StatusCode]++
	}

	percentObserved := float64(stsCounts[s.sts]) / float64(s.rqn)

	epsilon := .1
	s.Assert().LessOrEqual(
		math.Abs(percentObserved-s.statusPercentExpectation),
		epsilon,
		"Expected %.2f%% of responses to have status %d, but got %d/%d (%.2f%%, +/-%.2f%%)",
		s.statusPercentExpectation*100, s.sts, stsCounts[s.sts], s.rqn, percentObserved*100, math.Abs(percentObserved-s.statusPercentExpectation)*100)
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
			rqn:                      200,
			sts:                      http.StatusTooManyRequests,
			statusPercentExpectation: .5,
			// Use a very generous epsilon to account for differences in test runtime.
			// For example, if the test takes longer, then more quota is available for the same amount of requests.
			statusPercentEpsilon: .1,
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
			rqn:                      100,
			sts:                      http.StatusOK,
			statusPercentExpectation: 1,
			statusPercentEpsilon:     0,
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
			rqn:                      1000,
			sts:                      http.StatusOK,
			statusPercentExpectation: 1,
			statusPercentEpsilon:     0,
		},
		{
			sts:                      http.StatusInternalServerError,
			rqn:                      1,
			statusPercentExpectation: 1,
			statusPercentEpsilon:     0,
		},
	} {
		suite.Run(t, testCase)
	}
}

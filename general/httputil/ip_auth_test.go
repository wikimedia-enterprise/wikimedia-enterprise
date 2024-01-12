package httputil

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type ipAuthTestSuite struct {
	suite.Suite
	pth string
	unm string
	ugr string
	ips string
	ipe string
	srv *httptest.Server
}

func (s *ipAuthTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.Use(IPAuth([]*IPUser{
		{
			IPRange: &IPRange{
				Start: net.ParseIP(s.ips),
				End:   net.ParseIP(s.ipe),
			},
			User: &User{
				Username: s.unm,
				Groups:   []string{s.ugr},
			},
		},
	}))

	rtr.GET(s.pth, func(gcx *gin.Context) {
		user, _ := gcx.Get("user")

		if user == nil {
			Unauthorized(gcx)
			gcx.Abort()
			return
		}

		s.Assert().Equal(s.unm, user.(*User).GetUsername())
		s.Assert().Contains(user.(*User).GetGroups(), s.ugr)
		gcx.Status(http.StatusOK)
	})

	return rtr
}

func (s *ipAuthTestSuite) SetupTest() {
	s.pth = "/login"
	s.srv = httptest.NewServer(s.createServer())
}

func (s *ipAuthTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *ipAuthTestSuite) TestAuthIPSuccess() {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", s.srv.URL, s.pth), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-For", "192.168.0.4")

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)

	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *ipAuthTestSuite) TestAuthIPError() {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", s.srv.URL, s.pth), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)

	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
}

func TestIPAuth(t *testing.T) {
	for _, testCase := range []*ipAuthTestSuite{
		{
			unm: "john_doe",
			ugr: "gr_1",
			ips: "192.168.0.1",
			ipe: "192.168.0.10",
		},
	} {
		suite.Run(t, testCase)
	}
}

package httputil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

func createTMPFile(dta string, nme string) *os.File {
	fle, _ := os.CreateTemp("/tmp", nme)
	_, _ = fle.Write([]byte(dta))
	return fle
}

func setupUser(unm string, groups ...string) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		user := new(User)
		user.SetUsername(unm)
		user.SetGroups(groups)

		gcx.Set("user", user)
	}
}

type rbacCasbinTestSuite struct {
	suite.Suite
	mdl string
	plc string
	pth string
	unm string
	gps []string
	hdl gin.HandlerFunc
	sts int
	enf *casbin.Enforcer
	srv *httptest.Server
}

func (s *rbacCasbinTestSuite) SetupSuite() {
	pfl := createTMPFile(s.plc, "policy")
	mfl := createTMPFile(s.mdl, "model")

	var err error
	s.enf, err = casbin.NewEnforcer(mfl.Name(), pfl.Name())
	s.Assert().NoError(err)

	gin.SetMode(gin.TestMode)
	rtr := gin.New()
	rtr.Use(setupUser(s.unm, s.gps...))
	rtr.Use(RBAC(CasbinRBACAuthorizer(s.enf)))
	rtr.Use(s.hdl)
	s.srv = httptest.NewServer(rtr)
}

func (s *rbacCasbinTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *rbacCasbinTestSuite) TestAuthorizer() {
	res, err := http.Get(fmt.Sprintf("%s/%s", s.srv.URL, s.pth))
	s.Assert().NoError(err)
	s.Assert().Equal(s.sts, res.StatusCode)
}

func TestRBACCasbin(t *testing.T) {
	mdl := `
	[request_definition]
	r = sub, obj, act
	[policy_definition]
	p = sub, obj, act
	[role_definition]
	g = _, _
	[policy_effect]
	e = some(where (p.eft == allow))
	[matchers]
	m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
	`
	plc := `
	p, group_1, /group-1, GET
	p, group_2, /group-2, GET
	p, group_3, /group-3, GET
	g, group_2, group_1
	g, group_3, group_2
	`
	hdl := func(gcx *gin.Context) {
		gcx.Status(http.StatusTeapot)
	}

	for _, testCase := range []*rbacCasbinTestSuite{
		{
			mdl: mdl,
			plc: plc,
			unm: "john",
			gps: []string{"group_3"},
			pth: "group-1",
			sts: http.StatusTeapot,
			hdl: hdl,
		},
		{
			mdl: mdl,
			plc: plc,
			unm: "john",
			gps: []string{"group_1"},
			pth: "group-3",
			sts: http.StatusForbidden,
			hdl: hdl,
		},
		{
			mdl: mdl,
			plc: plc,
			unm: "john",
			gps: []string{},
			pth: "group-1",
			sts: http.StatusForbidden,
			hdl: hdl,
		},
	} {
		suite.Run(t, testCase)
	}
}

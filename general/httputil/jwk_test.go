package httputil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type jwkTestSuite struct {
	suite.Suite
	cid string
	kid string
	wid string
	prk *rsa.PrivateKey
	jsv *httptest.Server
	jwn *JWKWellKnown
}

func (s *jwkTestSuite) createJWKServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, JWKWellKnown{
			Keys: []*JWK{
				{
					KID: s.kid,
					E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.prk.PublicKey.E)).Bytes()),
					N:   base64.RawURLEncoding.EncodeToString(s.prk.PublicKey.N.Bytes()),
				},
			},
		})
	})

	return rtr
}

func (s *jwkTestSuite) SetupTest() {
	s.prk, _ = rsa.GenerateKey(rand.Reader, 2048)
	s.jsv = httptest.NewServer(s.createJWKServer())
}

func (s *jwkTestSuite) SetupSuite() {
	s.jwn = new(JWKWellKnown)
}

func (s *jwkTestSuite) TearDownTest() {
	s.jsv.Close()
}

func (s *jwkTestSuite) TestJWKFetchSuccess() {
	err := s.jwn.Fetch(s.jsv.URL)
	s.Assert().NoError(err)
}

func (s *jwkTestSuite) TestJWKFetchError() {
	err := s.jwn.Fetch("")
	s.Assert().Error(err)
}

func (s *jwkTestSuite) TestJWKFindSuccess() {
	err := s.jwn.Fetch(s.jsv.URL)
	s.Assert().NoError(err)

	_, err = s.jwn.Find(s.kid)
	s.Assert().NoError(err)
}

func (s *jwkTestSuite) TestJWKFindError() {
	err := s.jwn.Fetch(s.jsv.URL)
	s.Assert().NoError(err)

	_, err = s.jwn.Find(s.wid)
	s.Assert().Error(err)
	s.Assert().Contains("key not found", err.Error())
}

func (s *jwkTestSuite) TestJWKSuccess() {
	err := s.jwn.Fetch(s.jsv.URL)
	s.Assert().NoError(err)

	jwk, err := s.jwn.Find(s.kid)
	s.Assert().NoError(err)

	_, err = jwk.RSA256()
	s.Assert().NoError(err)
}

func TestJWK(t *testing.T) {
	for _, testCase := range []*jwkTestSuite{
		{
			cid: "client-id",
			kid: "1234example",
			wid: "1234wrongexample",
		},
	} {
		suite.Run(t, testCase)
	}
}

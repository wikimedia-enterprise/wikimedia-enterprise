package httputil

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type authProviderMock struct {
	mock.Mock
	AuthProvider
}

func (c *authProviderMock) GetUser(tkn string) (*User, error) {
	arg := c.Called(tkn)

	return arg.Get(0).(*User), arg.Error(1)
}

type JWKProviderMock struct {
	mock.Mock
	JWKFetcher
	JWKFinder
}

func (c *JWKProviderMock) Fetch(iss interface{}) error {
	arg := c.Called(iss)

	return arg.Error(0)
}

func (c *JWKProviderMock) Find(kid string) (*JWK, error) {
	arg := c.Called(kid)

	return arg.Get(0).(*JWK), arg.Error(1)
}

type cognitoAuthTestSuite struct {
	suite.Suite
	kid string
	cid string
	csc string
	wci string
	unm string
	ugr string
	pth string
	iss string
	ugs []string
	prk *rsa.PrivateKey
	cmd *redis.Client
	apm *authProviderMock
	jwp *JWKProviderMock
	exp time.Duration
	jwk *JWK
	mrd *miniredis.Miniredis
}

func (s *cognitoAuthTestSuite) getJWTToken() (string, error) {
	token := jwt.
		NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"client_id":      s.cid,
			"iss":            s.iss,
			"cognito:groups": []string{s.ugr},
			"sub":            "provided-sub",
		})

	token.Header["kid"] = s.kid

	return token.SignedString(s.prk)
}

func (s *cognitoAuthTestSuite) getWrongJWTToken() (string, error) {
	token := jwt.
		NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"client_id":      s.wci,
			"iss":            s.iss,
			"cognito:groups": []string{s.ugr},
			"sub":            "provided-sub",
		})

	token.Header["kid"] = s.kid

	return token.SignedString(s.prk)
}

func (s *cognitoAuthTestSuite) createServer(mwr ...gin.HandlerFunc) http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()
	rtr.Use(mwr...)

	rtr.Use(Auth(&AuthParams{
		Provider: s.apm,
		Cache:    s.cmd,
		ClientID: s.cid,
		Expire:   s.exp,
		JWK:      s.jwp,
	}))

	rtr.GET(s.pth, func(gcx *gin.Context) {
		user, _ := gcx.Get("user")
		s.Assert().Equal(s.unm, user.(*User).GetUsername())
		s.Assert().Contains(user.(*User).GetGroups(), s.ugr)
		s.Assert().Equal("provided-sub", user.(*User).Sub)
		gcx.Status(http.StatusOK)
	})

	return rtr
}

func (s *cognitoAuthTestSuite) createCache() *redis.Client {
	cmdable := redis.NewClient(&redis.Options{
		Addr: s.mrd.Addr(),
	})

	return cmdable
}

func (s *cognitoAuthTestSuite) SetupTest() {
	s.pth = "/login"
	s.ugs = []string{s.ugr}
	s.mrd, _ = miniredis.Run()
	s.apm = new(authProviderMock)
	s.jwp = new(JWKProviderMock)
	s.cmd = s.createCache()
}

func (s *cognitoAuthTestSuite) SetupSuite() {
	s.prk, _ = rsa.GenerateKey(rand.Reader, 2048)
	s.jwk = &JWK{
		KID: s.kid,
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.prk.PublicKey.E)).Bytes()),
		N:   base64.RawURLEncoding.EncodeToString(s.prk.PublicKey.N.Bytes()),
	}
}

func (s *cognitoAuthTestSuite) TearDownTest() {
	s.mrd.Close()
}

func (s *cognitoAuthTestSuite) TestAuthUnauthorizedError() {
	srv := httptest.NewServer(s.createServer())
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s%s", srv.URL, s.pth))
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	s.apm.AssertNumberOfCalls(s.T(), "GetUser", 0)
}

func (s *cognitoAuthTestSuite) TestAuthSuccess() {
	srv := httptest.NewServer(s.createServer())
	defer srv.Close()

	token, err := s.getJWTToken()
	s.Assert().NoError(err)

	s.jwp.On("Fetch", s.iss).Return(nil)
	s.jwp.On("Find", s.kid).Return(s.jwk, nil)

	s.apm.On("GetUser", token).Return(
		&User{
			Username: s.unm,
		},
		nil,
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, s.pth), nil)
	s.Assert().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)

	s.Assert().Equal(http.StatusOK, res.StatusCode)

	s.jwp.AssertNumberOfCalls(s.T(), "Fetch", 1)
	s.jwp.AssertNumberOfCalls(s.T(), "Find", 1)
	s.apm.AssertNumberOfCalls(s.T(), "GetUser", 1)
}

func (s *cognitoAuthTestSuite) TestAuthUserSetSuccess() {
	mwr := func(gcx *gin.Context) {
		gcx.Set("user", &User{
			Username: s.unm,
			Groups:   []string{s.ugr},
			Sub:      "provided-sub",
		})
	}

	srv := httptest.NewServer(s.createServer(mwr))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, s.pth), nil)
	s.Assert().NoError(err)

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)

	s.Assert().Equal(http.StatusOK, res.StatusCode)
}

func (s *cognitoAuthTestSuite) TestAuthCacheSuccess() {
	srv := httptest.NewServer(s.createServer())
	defer srv.Close()

	token, err := s.getJWTToken()
	s.Assert().NoError(err)

	s.jwp.On("Fetch", s.iss).Return(nil)
	s.jwp.On("Find", s.kid).Return(s.jwk, nil)

	s.apm.On("GetUser", token).Return(
		&User{
			Username: s.unm,
		},
		nil,
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, s.pth), nil)
	s.Assert().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	for i := 0; i < 10; i++ {
		res, err := http.DefaultClient.Do(req)
		s.Assert().NoError(err)

		s.Assert().Equal(http.StatusOK, res.StatusCode)
	}

	s.apm.AssertNumberOfCalls(s.T(), "GetUser", 1)
}

func (s *cognitoAuthTestSuite) TestAuthWrongTokenClientError() {
	srv := httptest.NewServer(s.createServer())
	defer srv.Close()

	token, err := s.getWrongJWTToken()
	s.Assert().NoError(err)

	s.jwp.On("Fetch", s.iss).Return(nil)
	s.jwp.On("Find", s.kid).Return(s.jwk, nil)

	s.apm.On("GetUser", token).Return(
		&User{
			Username: s.unm,
		},
		nil,
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, s.pth), nil)
	s.Assert().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), "incorrect client id")

	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	s.jwp.AssertNumberOfCalls(s.T(), "Fetch", 0)
	s.jwp.AssertNumberOfCalls(s.T(), "Find", 0)
	s.apm.AssertNumberOfCalls(s.T(), "GetUser", 0)
}

func (s *cognitoAuthTestSuite) TestAuthTokenInvalidError() {
	srv := httptest.NewServer(s.createServer())
	defer srv.Close()

	token, err := s.getJWTToken()
	s.Assert().NoError(err)

	s.jwp.On("Fetch", s.iss).Return(nil)
	s.jwp.On("Find", s.kid).Return(s.jwk, nil)

	s.apm.On("GetUser", token).Return(
		&User{},
		errors.New("token is not valid"),
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", srv.URL, s.pth), nil)
	s.Assert().NoError(err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	s.Assert().NoError(err)

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), "token is not valid")
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
}

func TestAuth(t *testing.T) {
	for _, testCase := range []*cognitoAuthTestSuite{
		{
			kid: "pqZ9xSMr5rtwrPG2LRM9v",
			unm: "john_doe",
			ugr: "admin",
			cid: "jN4Ag4CEL2TQtrqk",
			csc: "secret",
			wci: "VnFAL5ke9hK8v6bT",
			iss: "iss",
			exp: time.Minute * 1,
		},
	} {
		suite.Run(t, testCase)
	}
}

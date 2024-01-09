package tokenrevoke_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/handlers/v1/tokenrevoke"

	"github.com/alicebob/miniredis"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const tokenrevokeTestURL = "/token-revoke"
const tokenrevokeTestErrMsg = "service_error"

func createServer(p *tokenrevoke.Parameters) http.Handler {
	router := gin.New()
	router.POST(tokenrevokeTestURL, tokenrevoke.NewHandler(p))
	return router
}

type tokenrevokeCognitoAPIMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (m *tokenrevokeCognitoAPIMock) RevokeTokenWithContext(_ aws.Context, in *cognitoidentityprovider.RevokeTokenInput, _ ...request.Option) (*cognitoidentityprovider.RevokeTokenOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.RevokeTokenOutput), args.Error(1)
}

type tokenRevokeRedis struct {
	server *miniredis.Miniredis
	client *redis.Client
}

type tokenrevokeTestSuite struct {
	suite.Suite
	env   *env.Environment
	cog   *tokenrevokeCognitoAPIMock
	mod   *tokenrevoke.Model
	inp   *cognitoidentityprovider.RevokeTokenInput
	srv   *httptest.Server
	redis *tokenRevokeRedis
}

func (s *tokenrevokeTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.env = new(env.Environment)
	s.env.CognitoClientID = "id"
	s.env.CognitoSecret = "supersecretkey"
	s.mod = &tokenrevoke.Model{
		RefreshToken: "refresh_token",
	}

	s.inp = &cognitoidentityprovider.RevokeTokenInput{
		ClientId:     aws.String(s.env.CognitoClientID),
		ClientSecret: aws.String(s.env.CognitoSecret),
		Token:        aws.String(s.mod.RefreshToken),
	}

}

func (s *tokenrevokeTestSuite) SetupTest() {
	var err error

	s.redis = new(tokenRevokeRedis)
	s.redis.server, err = miniredis.Run()
	s.Assert().NoError(err)
	s.redis.client = redis.NewClient(&redis.Options{
		Addr: s.redis.server.Addr(),
	})

	s.cog = new(tokenrevokeCognitoAPIMock)
	s.srv = httptest.NewServer(createServer(&tokenrevoke.Parameters{
		Env:     s.env,
		Cognito: s.cog,
		Redis:   s.redis.client,
	}))
}

func (s *tokenrevokeTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *tokenrevokeTestSuite) TestRevokeHandler() {
	out := &cognitoidentityprovider.RevokeTokenOutput{}
	s.cog.On("RevokeTokenWithContext", s.inp).Return(out, nil)

	key := fmt.Sprintf("refresh_token:username:%s:access_tokens", s.mod.RefreshToken)
	tokens := []string{key}

	for i := 0; i < 10; i++ {
		accessToken := fmt.Sprintf("some_access_token_%d", i)
		accessTokenKey := fmt.Sprintf("access_token:%s", accessToken)
		tokens = append(tokens, accessToken)
		s.redis.client.SAdd(context.Background(), key, accessToken)
		s.redis.client.Set(context.Background(), accessTokenKey, "some_value", 1*time.Minute)
	}

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrevokeTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusNoContent, res.StatusCode)

	for _, t := range tokens {
		exists, err := s.redis.client.Exists(context.Background(), t).Result()
		s.Assert().NoError(err)
		s.Assert().Zero(exists)
	}
}

func (s *tokenrevokeTestSuite) TestRevokeValidation() {
	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(&tokenrevoke.Model{}))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrevokeTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *tokenrevokeTestSuite) TestRevokeError() {
	out := &cognitoidentityprovider.RevokeTokenOutput{}
	s.cog.On("RevokeTokenWithContext", s.inp).Return(out, errors.New(tokenrevokeTestErrMsg))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrevokeTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)

	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), tokenrevokeTestErrMsg)
}

func (s *tokenrevokeTestSuite) TestTokenRevokeRedisError() {
	out := &cognitoidentityprovider.RevokeTokenOutput{}
	s.cog.On("RevokeTokenWithContext", s.inp).Return(out, nil)
	s.redis.server.Close()

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrevokeTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(tokenrevokeTestSuite))
}

package tokenrefresh_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/handlers/v1/tokenrefresh"

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

const tokenrefreshTestURL = "/refresh-token"

func createServer(p *tokenrefresh.Parameters) http.Handler {
	router := gin.New()
	router.POST(tokenrefreshTestURL, tokenrefresh.NewHandler(p))
	return router
}

type refreshTokenCognitoAPIMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (m *refreshTokenCognitoAPIMock) InitiateAuthWithContext(_ aws.Context, in *cognitoidentityprovider.InitiateAuthInput, _ ...request.Option) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.InitiateAuthOutput), args.Error(1)
}

type tokenRefreshRedis struct {
	server *miniredis.Miniredis
	client *redis.Client
}

type refreshTokenTestSuite struct {
	suite.Suite
	env   *env.Environment
	cog   *refreshTokenCognitoAPIMock
	inp   *cognitoidentityprovider.InitiateAuthInput
	srv   *httptest.Server
	req   *tokenrefresh.Request
	redis *tokenRefreshRedis
}

func (s *refreshTokenTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.env = new(env.Environment)
	s.env = &env.Environment{
		CognitoClientID:      "client_id",
		CognitoSecret:        "secret",
		MaxAccessTokens:      10,
		AccessTokensExpHours: 240,
	}

	s.req = &tokenrefresh.Request{
		Username:     "test_username",
		RefreshToken: "refresh_token",
	}

	hmac := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, err := hmac.Write([]byte(fmt.Sprintf("%s%s", s.req.Username, s.env.CognitoClientID)))
	s.Assert().NoError(err)

	s.inp = &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(s.env.CognitoClientID),
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(s.req.RefreshToken),
			"SECRET_HASH":   aws.String(base64.StdEncoding.EncodeToString(hmac.Sum(nil))),
		},
	}
}

func (s *refreshTokenTestSuite) TearDownTest() {
	s.srv.Close()
	s.redis.client.Close()
	s.redis.server.Close()
}

func (s *refreshTokenTestSuite) SetupTest() {
	var err error

	s.redis = new(tokenRefreshRedis)
	s.redis.server, err = miniredis.Run()
	s.Assert().NoError(err)
	s.redis.client = redis.NewClient(&redis.Options{
		Addr: s.redis.server.Addr(),
	})

	s.cog = new(refreshTokenCognitoAPIMock)
	s.srv = httptest.NewServer(createServer(&tokenrefresh.Parameters{
		Env:     s.env,
		Cognito: s.cog,
		Redis:   s.redis.client,
	}))

}

func (s *refreshTokenTestSuite) TestRefreshTokenHandler() {
	out := &cognitoidentityprovider.InitiateAuthOutput{
		AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
			AccessToken: aws.String("token"),
			IdToken:     aws.String("id_token"),
			ExpiresIn:   aws.Int64(300),
		},
	}

	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrefreshTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	result := new(tokenrefresh.Response)
	s.Assert().NoError(json.NewDecoder(res.Body).Decode(result))

	s.Assert().Equal(*out.AuthenticationResult.AccessToken, result.AccessToken)
	s.Assert().Equal(*out.AuthenticationResult.IdToken, result.IdToken)
	s.Assert().Equal(int(*out.AuthenticationResult.ExpiresIn), result.ExpiresIn)

	members, err := s.redis.client.SMembers(
		context.Background(),
		fmt.Sprintf("refresh_token:%s:%s:access_tokens", s.req.Username, s.req.RefreshToken),
	).Result()
	s.Assert().NoError(err)
	s.Assert().Contains(members, *out.AuthenticationResult.AccessToken)
}

func (s *refreshTokenTestSuite) TestRefreshTokenErrors() {
	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(&tokenrefresh.Request{}))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrefreshTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *refreshTokenTestSuite) TestRefreshTokenEmptyResponse() {
	out := &cognitoidentityprovider.InitiateAuthOutput{}

	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrefreshTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *refreshTokenTestSuite) TestRefreshTokenErrorMaxTokens() {
	out := &cognitoidentityprovider.InitiateAuthOutput{
		AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
			AccessToken: aws.String("token"),
			IdToken:     aws.String("id_token"),
			ExpiresIn:   aws.Int64(300),
		},
	}

	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, nil)

	var tokens []interface{}

	for i := 0; i < int(s.env.MaxAccessTokens); i++ {
		tokens = append(tokens, fmt.Sprintf("access_token_%d", i))
	}

	s.redis.client.SAdd(
		context.Background(),
		fmt.Sprintf("refresh_token:%s:%s:access_tokens", s.req.Username, s.req.RefreshToken),
		tokens...,
	)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrefreshTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(data), "limit has been exceeded")
	s.Assert().Equal(http.StatusTooManyRequests, res.StatusCode)
}

func (s *refreshTokenTestSuite) TestRefreshTokenRedisError() {
	out := &cognitoidentityprovider.InitiateAuthOutput{
		AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
			AccessToken: aws.String("token"),
			IdToken:     aws.String("id_token"),
			ExpiresIn:   aws.Int64(300),
		},
	}

	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, nil)
	s.redis.server.Close()

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, tokenrefreshTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(refreshTokenTestSuite))
}

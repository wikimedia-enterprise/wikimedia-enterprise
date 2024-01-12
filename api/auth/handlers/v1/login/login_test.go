package login_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/handlers/v1/login"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const loginTestURL = "/login"

func createServer(p *login.Parameters) http.Handler {
	router := gin.New()
	router.POST(loginTestURL, login.NewHandler(p))
	return router
}

type loginCognitoAPIMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (m *loginCognitoAPIMock) InitiateAuthWithContext(_ aws.Context, in *cognitoidentityprovider.InitiateAuthInput, _ ...request.Option) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.InitiateAuthOutput), args.Error(1)
}

type loginTestSuite struct {
	suite.Suite
	env *env.Environment
	cog *loginCognitoAPIMock
	mod *login.Model
	inp *cognitoidentityprovider.InitiateAuthInput
	srv *httptest.Server
}

func (s *loginTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.env = new(env.Environment)
	s.env.CognitoClientID = "id"
	s.env.CognitoSecret = "supersecretkey"
	s.mod = &login.Model{
		Username: "test",
		Password: "sql",
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, _ = h.Write([]byte(fmt.Sprintf("%s%s", s.mod.Username, s.env.CognitoClientID)))
	s.inp = &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(s.env.CognitoClientID),
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME":    aws.String(s.mod.Username),
			"PASSWORD":    aws.String(s.mod.Password),
			"SECRET_HASH": aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
		},
	}
}

func (s *loginTestSuite) SetupTest() {
	s.cog = new(loginCognitoAPIMock)
	s.srv = httptest.NewServer(createServer(&login.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))
}

func (s *loginTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *loginTestSuite) TestHandlerTokens() {
	out := &cognitoidentityprovider.InitiateAuthOutput{
		AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
			AccessToken:  aws.String("token"),
			IdToken:      aws.String("id_token"),
			RefreshToken: aws.String("refersh_token"),
			ExpiresIn:    aws.Int64(20),
		},
	}
	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, loginTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	result := new(login.Response)
	s.Assert().NoError(json.NewDecoder(res.Body).Decode(result))
	s.Assert().Equal(*out.AuthenticationResult.AccessToken, result.AccessToken)
	s.Assert().Equal(*out.AuthenticationResult.IdToken, result.IDToken)
	s.Assert().Equal(*out.AuthenticationResult.RefreshToken, result.RefreshToken)
	s.Assert().Equal(int(*out.AuthenticationResult.ExpiresIn), result.ExpiresIn)
}

func (s *loginTestSuite) TestHandlerChallenge() {
	out := &cognitoidentityprovider.InitiateAuthOutput{
		ChallengeName: aws.String("challenge_name"),
		Session:       aws.String("session"),
	}

	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, loginTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	result := new(login.Response)
	s.Assert().NoError(json.NewDecoder(res.Body).Decode(result))
	s.Assert().Equal(*out.Session, result.Session)
	s.Assert().Equal(*out.ChallengeName, result.ChallengeName)
}

func (s *loginTestSuite) TestHandlerValidationErr() {
	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(&login.Model{}))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, loginTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *loginTestSuite) TestHandlerUnauthorizedErr() {
	out := &cognitoidentityprovider.InitiateAuthOutput{}
	s.cog.On("InitiateAuthWithContext", s.inp).Return(out, errors.New("test_error"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, loginTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(loginTestSuite))
}

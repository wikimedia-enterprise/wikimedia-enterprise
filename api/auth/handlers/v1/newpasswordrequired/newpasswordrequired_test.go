package newpasswordrequired_test

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
	"wikimedia-enterprise/api/auth/handlers/v1/newpasswordrequired"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const newPasswordRequiredTestURL = "/new-password-required"

type newPasswordRequiredCognitoAPIMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (m *newPasswordRequiredCognitoAPIMock) RespondToAuthChallengeWithContext(_ aws.Context, in *cognitoidentityprovider.RespondToAuthChallengeInput, _ ...request.Option) (*cognitoidentityprovider.RespondToAuthChallengeOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.RespondToAuthChallengeOutput), args.Error(1)
}

type newPasswordRequiredTestSuite struct {
	suite.Suite
	env *env.Environment
	cog *newPasswordRequiredCognitoAPIMock
	mod *newpasswordrequired.Model
	inp *cognitoidentityprovider.RespondToAuthChallengeInput
	srv *httptest.Server
}

func (s *newPasswordRequiredTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.env = new(env.Environment)
	s.env.CognitoClientID = "id"
	s.env.CognitoSecret = "supersecretkey"
	s.mod = &newpasswordrequired.Model{
		Username:    "test",
		NewPassword: "sql1234",
		Session:     "sessionkey",
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, _ = h.Write([]byte(fmt.Sprintf("%s%s", s.mod.Username, s.env.CognitoClientID)))
	s.inp = &cognitoidentityprovider.RespondToAuthChallengeInput{
		Session:       aws.String(s.mod.Session),
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
		ClientId:      aws.String(s.env.CognitoClientID),
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(s.mod.Username),
			"NEW_PASSWORD": aws.String(s.mod.NewPassword),
			"SECRET_HASH":  aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
		},
	}
}

func (s *newPasswordRequiredTestSuite) createServer(p *newpasswordrequired.Parameters) http.Handler {
	r := gin.New()
	r.POST(newPasswordRequiredTestURL, newpasswordrequired.NewHandler(p))
	return r
}

func (s *newPasswordRequiredTestSuite) SetupTest() {
	s.cog = new(newPasswordRequiredCognitoAPIMock)
	s.srv = httptest.NewServer(s.createServer(&newpasswordrequired.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))
}

func (s *newPasswordRequiredTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *newPasswordRequiredTestSuite) TestHandlerTokens() {
	out := &cognitoidentityprovider.RespondToAuthChallengeOutput{
		AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
			AccessToken:  aws.String("token"),
			IdToken:      aws.String("id_token"),
			RefreshToken: aws.String("refersh_token"),
			ExpiresIn:    aws.Int64(20),
		},
	}
	s.cog.On("RespondToAuthChallengeWithContext", s.inp).Return(out, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, newPasswordRequiredTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	result := new(newpasswordrequired.Response)
	s.Assert().NoError(json.NewDecoder(res.Body).Decode(result))
	s.Assert().Equal(*out.AuthenticationResult.AccessToken, result.AccessToken)
	s.Assert().Equal(*out.AuthenticationResult.IdToken, result.IDToken)
	s.Assert().Equal(*out.AuthenticationResult.RefreshToken, result.RefreshToken)
	s.Assert().Equal(int(*out.AuthenticationResult.ExpiresIn), result.ExpiresIn)

	s.cog.AssertNumberOfCalls(s.T(), "RespondToAuthChallengeWithContext", 1)
}

func (s *newPasswordRequiredTestSuite) TestHandlerTokensErr() {
	out := &cognitoidentityprovider.RespondToAuthChallengeOutput{}
	s.cog.On("RespondToAuthChallengeWithContext", s.inp).Return(out, errors.New("session is not valid"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, newPasswordRequiredTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)

	s.cog.AssertNumberOfCalls(s.T(), "RespondToAuthChallengeWithContext", 1)
}

func (s *newPasswordRequiredTestSuite) TestHandlerValidationErr() {
	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(&newpasswordrequired.Model{}))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, newPasswordRequiredTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)

	s.cog.AssertNotCalled(s.T(), "RespondToAuthChallengeWithContext")
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(newPasswordRequiredTestSuite))
}

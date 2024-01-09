package forgotpassword_test

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
	"wikimedia-enterprise/api/auth/handlers/v1/forgotpassword"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const forgotPasswordTestURL = "/forgot-password"

type forgotPasswordCognitoAPIMock struct {
	mock.Mock
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (m *forgotPasswordCognitoAPIMock) ForgotPasswordWithContext(_ aws.Context, in *cognitoidentityprovider.ForgotPasswordInput, _ ...request.Option) (*cognitoidentityprovider.ForgotPasswordOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ForgotPasswordOutput), args.Error(1)
}

type forgotPasswordTestSuite struct {
	suite.Suite
	env *env.Environment
	req *forgotpassword.Request
	cog *forgotPasswordCognitoAPIMock
	fpi *cognitoidentityprovider.ForgotPasswordInput
	srv *httptest.Server
}

func (s *forgotPasswordTestSuite) createServer(p *forgotpassword.Parameters) http.Handler {
	r := gin.New()
	r.POST(forgotPasswordTestURL, forgotpassword.NewHandler(p))
	return r
}

func (s *forgotPasswordTestSuite) SetupTest() {
	s.cog = new(forgotPasswordCognitoAPIMock)
	s.srv = httptest.NewServer(s.createServer(&forgotpassword.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))
}

func (s *forgotPasswordTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *forgotPasswordTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.req = &forgotpassword.Request{Username: "username"}
	s.env = &env.Environment{
		CognitoClientID: "client_id",
		CognitoSecret:   "secret",
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, err := h.Write([]byte(fmt.Sprintf("%s%s", s.req.Username, s.env.CognitoClientID)))
	s.Assert().NoError(err)

	s.fpi = &cognitoidentityprovider.ForgotPasswordInput{
		ClientId:   aws.String(s.env.CognitoClientID),
		SecretHash: aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
		Username:   aws.String(s.req.Username),
	}
}

func (s *forgotPasswordTestSuite) TestForgotPasswordHandler() {
	out := new(cognitoidentityprovider.ForgotPasswordOutput)
	s.cog.On("ForgotPasswordWithContext", s.fpi).Return(out, nil).Once()

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusNoContent, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ForgotPasswordWithContext", 1)
}

func (s *forgotPasswordTestSuite) TestValidationError() {
	out := new(cognitoidentityprovider.ForgotPasswordOutput)
	s.cog.On("ForgotPasswordWithContext", s.fpi).Return(out, nil).Once()

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordTestURL), "application/json", nil)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNotCalled(s.T(), "ForgotPasswordWithContext")
}

func (s *forgotPasswordTestSuite) TestUnauthorizedError() {
	out := new(cognitoidentityprovider.ForgotPasswordOutput)
	s.cog.On("ForgotPasswordWithContext", s.fpi).Return(out, errors.New("test_error")).Once()

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ForgotPasswordWithContext", 1)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(forgotPasswordTestSuite))
}

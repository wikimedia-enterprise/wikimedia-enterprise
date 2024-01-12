package forgotpasswordconfirm_test

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
	"wikimedia-enterprise/api/auth/handlers/v1/forgotpasswordconfirm"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const forgotPasswordConfirmTestURL = "/confirm-forgot-password"

type forgotPasswordConfirmCognitoAPIMock struct {
	mock.Mock
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (m *forgotPasswordConfirmCognitoAPIMock) ConfirmForgotPasswordWithContext(_ aws.Context, in *cognitoidentityprovider.ConfirmForgotPasswordInput, _ ...request.Option) (*cognitoidentityprovider.ConfirmForgotPasswordOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ConfirmForgotPasswordOutput), args.Error(1)
}

type forgotPasswordConfirmTestSuite struct {
	suite.Suite
	env  *env.Environment
	req  *forgotpasswordconfirm.Request
	cog  *forgotPasswordConfirmCognitoAPIMock
	cfpi *cognitoidentityprovider.ConfirmForgotPasswordInput
	srv  *httptest.Server
}

func (s *forgotPasswordConfirmTestSuite) createServer(p *forgotpasswordconfirm.Parameters) http.Handler {
	r := gin.New()
	r.POST(forgotPasswordConfirmTestURL, forgotpasswordconfirm.NewHandler(p))
	return r
}

func (s *forgotPasswordConfirmTestSuite) SetupTest() {
	s.cog = new(forgotPasswordConfirmCognitoAPIMock)
	s.srv = httptest.NewServer(s.createServer(&forgotpasswordconfirm.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))
}

func (s *forgotPasswordConfirmTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *forgotPasswordConfirmTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.req = &forgotpasswordconfirm.Request{
		Username:         "username",
		Password:         "new_password",
		ConfirmationCode: "123456",
	}
	s.env = &env.Environment{
		CognitoClientID: "client_id",
		CognitoSecret:   "secret",
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, err := h.Write([]byte(fmt.Sprintf("%s%s", s.req.Username, s.env.CognitoClientID)))
	s.Assert().NoError(err)

	s.cfpi = &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(s.env.CognitoClientID),
		ConfirmationCode: aws.String(s.req.ConfirmationCode),
		Password:         aws.String(s.req.Password),
		Username:         aws.String(s.req.Username),
		SecretHash:       aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
	}
}

func (s *forgotPasswordConfirmTestSuite) TestConfirmForgotPasswordHandler() {
	out := new(cognitoidentityprovider.ConfirmForgotPasswordOutput)
	s.cog.On("ConfirmForgotPasswordWithContext", s.cfpi).Return(out, nil).Once()

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordConfirmTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusNoContent, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ConfirmForgotPasswordWithContext", 1)
}

func (s *forgotPasswordConfirmTestSuite) TestValidationError() {
	out := new(cognitoidentityprovider.ConfirmForgotPasswordOutput)
	s.cog.On("ConfirmForgotPasswordWithContext", s.cfpi).Return(out, nil).Once()

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordConfirmTestURL), "application/json", nil)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNotCalled(s.T(), "ConfirmForgotPasswordWithContext")
}

func (s *forgotPasswordConfirmTestSuite) TestUnauthorizedError() {
	out := new(cognitoidentityprovider.ConfirmForgotPasswordOutput)
	s.cog.On("ConfirmForgotPasswordWithContext", s.cfpi).Return(out, errors.New("test_error")).Once()

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordConfirmTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ConfirmForgotPasswordWithContext", 1)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(forgotPasswordConfirmTestSuite))
}

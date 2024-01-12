package resendconfirm_test

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
	"wikimedia-enterprise/api/auth/handlers/v1/resendconfirm"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const resendConfirmationTestURL = "/resend-confirmation"

type resendConfirmationCognitoAPIMock struct {
	mock.Mock
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (m *resendConfirmationCognitoAPIMock) ResendConfirmationCodeWithContext(_ aws.Context, in *cognitoidentityprovider.ResendConfirmationCodeInput, _ ...request.Option) (*cognitoidentityprovider.ResendConfirmationCodeOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ResendConfirmationCodeOutput), args.Error(1)
}

type resendConfirmationTestSuite struct {
	suite.Suite
	env *env.Environment
	req *resendconfirm.Request
	cog *resendConfirmationCognitoAPIMock
	fpi *cognitoidentityprovider.ResendConfirmationCodeInput
	srv *httptest.Server
}

func (s *resendConfirmationTestSuite) createServer(p *resendconfirm.Parameters) http.Handler {
	r := gin.New()
	r.POST(resendConfirmationTestURL, resendconfirm.NewHandler(p))
	return r
}

func (s *resendConfirmationTestSuite) SetupTest() {
	s.cog = new(resendConfirmationCognitoAPIMock)
	s.srv = httptest.NewServer(s.createServer(&resendconfirm.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))
}

func (s *resendConfirmationTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *resendConfirmationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.req = &resendconfirm.Request{Username: "username"}
	s.env = &env.Environment{
		CognitoClientID: "client_id",
		CognitoSecret:   "secret",
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, err := h.Write([]byte(fmt.Sprintf("%s%s", s.req.Username, s.env.CognitoClientID)))
	s.Assert().NoError(err)

	s.fpi = &cognitoidentityprovider.ResendConfirmationCodeInput{
		ClientId:   aws.String(s.env.CognitoClientID),
		SecretHash: aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
		Username:   aws.String(s.req.Username),
	}
}

func (s *resendConfirmationTestSuite) TestResendConfirmationHandler() {
	out := new(cognitoidentityprovider.ResendConfirmationCodeOutput)
	s.cog.On("ResendConfirmationCodeWithContext", s.fpi).Return(out, nil).Once()

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, resendConfirmationTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusNoContent, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ResendConfirmationCodeWithContext", 1)
}

func (s *resendConfirmationTestSuite) TestValidationError() {
	out := new(cognitoidentityprovider.ResendConfirmationCodeOutput)
	s.cog.On("ResendConfirmationCodeWithContext", s.fpi).Return(out, nil).Once()

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, resendConfirmationTestURL), "application/json", nil)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNotCalled(s.T(), "ResendConfirmationCodeWithContext")
}

func (s *resendConfirmationTestSuite) TestUnauthorizedError() {
	out := new(cognitoidentityprovider.ResendConfirmationCodeOutput)
	s.cog.On("ResendConfirmationCodeWithContext", s.fpi).Return(out, errors.New("test_error")).Once()

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, resendConfirmationTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ResendConfirmationCodeWithContext", 1)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(resendConfirmationTestSuite))
}

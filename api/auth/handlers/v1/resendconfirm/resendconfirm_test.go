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
	"wikimedia-enterprise/api/auth/libraries/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const resendConfirmationTestURL = "/resend-confirmation"

type cognitoAPIMock struct {
	mock.Mock
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (m *cognitoAPIMock) ResendConfirmationCodeWithContext(_ aws.Context, in *cognitoidentityprovider.ResendConfirmationCodeInput, _ ...request.Option) (*cognitoidentityprovider.ResendConfirmationCodeOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ResendConfirmationCodeOutput), args.Error(1)
}

func (m *cognitoAPIMock) AdminGetUser(input *cognitoidentityprovider.AdminGetUserInput) (*cognitoidentityprovider.AdminGetUserOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cognitoidentityprovider.AdminGetUserOutput), args.Error(1)
}

func (m *cognitoAPIMock) AdminListGroupsForUser(input *cognitoidentityprovider.AdminListGroupsForUserInput) (*cognitoidentityprovider.AdminListGroupsForUserOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cognitoidentityprovider.AdminListGroupsForUserOutput), args.Error(1)
}

func (m *cognitoAPIMock) ListUsers(input *cognitoidentityprovider.ListUsersInput) (*cognitoidentityprovider.ListUsersOutput, error) {
	args := m.Called(input)
	if args.Get(0) != nil {
		return args.Get(0).(*cognitoidentityprovider.ListUsersOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

type resendConfirmationTestSuite struct {
	suite.Suite
	env    *env.Environment
	req    *resendconfirm.Request
	cog    *cognitoAPIMock
	params *resendconfirm.Parameters
	utl    *utils.Parameters
	fpi    *cognitoidentityprovider.ResendConfirmationCodeInput
	srv    *httptest.Server
}

// Fake output data for mocked functions
var (
	fakeEmailOutput = &cognitoidentityprovider.AdminGetUserOutput{
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String("test@abc.com"),
			},
		},
	}

	fakeGroupsOutput = &cognitoidentityprovider.AdminListGroupsForUserOutput{
		Groups: []*cognitoidentityprovider.GroupType{
			{
				GroupName: aws.String("group_2"),
			},
			{
				GroupName: aws.String("group_1"),
			},
		},
	}
)

func (s *resendConfirmationTestSuite) createServer(p *resendconfirm.Parameters) http.Handler {
	r := gin.New()
	r.POST(resendConfirmationTestURL, resendconfirm.NewHandler(p))
	return r
}

func (s *resendConfirmationTestSuite) SetupTest() {
	s.cog = new(cognitoAPIMock)

	s.env = &env.Environment{
		CognitoClientID:   "client_id",
		CognitoSecret:     "secret",
		CognitoUserPoolID: "us-east-1_test",
	}

	s.srv = httptest.NewServer(s.createServer(&resendconfirm.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))

	s.params = &resendconfirm.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}

	s.utl = &utils.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}
}

func (s *resendConfirmationTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *resendConfirmationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.req = &resendconfirm.Request{Username: "testuser"}
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
		ClientMetadata: map[string]*string{
			"username": aws.String(s.req.Username),
			"group":    aws.String(utils.ZendeskTicketId["group_2"]),
		},
	}
}

func (s *resendConfirmationTestSuite) TestResendConfirmationHandler() {
	out := new(cognitoidentityprovider.ResendConfirmationCodeOutput)
	s.cog.On("ResendConfirmationCodeWithContext", s.fpi).Return(out, nil).Once()
	s.cog.On("AdminGetUser", mock.Anything).Return(fakeEmailOutput, nil)
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

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
	s.cog.On("ResendConfirmationCodeWithContext", mock.Anything).Return(out, errors.New("test_error")).Once()
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

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

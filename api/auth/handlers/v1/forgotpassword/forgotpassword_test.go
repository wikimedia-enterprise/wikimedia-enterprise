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

	"wikimedia-enterprise/api/auth/libraries/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const forgotPasswordTestURL = "/forgot-password"

type cognitoAPIMock struct {
	mock.Mock
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (m *cognitoAPIMock) ForgotPasswordWithContext(_ aws.Context, in *cognitoidentityprovider.ForgotPasswordInput, _ ...request.Option) (*cognitoidentityprovider.ForgotPasswordOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ForgotPasswordOutput), args.Error(1)
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

type forgotPasswordTestSuite struct {
	suite.Suite
	env    *env.Environment
	req    *forgotpassword.Request
	cog    *cognitoAPIMock
	fpi    *cognitoidentityprovider.ForgotPasswordInput
	srv    *httptest.Server
	params *forgotpassword.Parameters
	agui   *cognitoidentityprovider.AdminGetUserInput
	utl    *utils.Parameters
}

func (s *forgotPasswordTestSuite) createServer(p *forgotpassword.Parameters) http.Handler {
	r := gin.New()

	r.POST(forgotPasswordTestURL, forgotpassword.NewHandler(p))
	return r
}

func (s *forgotPasswordTestSuite) SetupTest() {
	s.cog = new(cognitoAPIMock)
	s.srv = httptest.NewServer(s.createServer(&forgotpassword.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))

	s.params = &forgotpassword.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}

	s.utl = &utils.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}
}

func (s *forgotPasswordTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *forgotPasswordTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.req = &forgotpassword.Request{
		Username: "testuser",
	}

	s.env = &env.Environment{
		CognitoClientID:   "client_id",
		CognitoSecret:     "secret",
		CognitoUserPoolID: "us-east-1_test",
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	_, err := h.Write([]byte(fmt.Sprintf("%s%s", s.req.Username, s.env.CognitoClientID)))
	s.Assert().NoError(err)

	s.fpi = &cognitoidentityprovider.ForgotPasswordInput{
		ClientId:   aws.String(s.env.CognitoClientID),
		SecretHash: aws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
		Username:   aws.String(s.req.Username),
		ClientMetadata: map[string]*string{
			"group":    aws.String(utils.ZendeskTicketId["group_2"]),
			"username": aws.String(s.req.Username),
		},
	}

	s.agui = &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(s.env.CognitoUserPoolID),
		Username:   aws.String("testuser"),
	}
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

	fakeForgotPasswordOutput = new(cognitoidentityprovider.ForgotPasswordOutput)

	fakeListUsersOutput = &cognitoidentityprovider.ListUsersOutput{
		Users: []*cognitoidentityprovider.UserType{
			{
				Username: aws.String("testuser"),
			},
		},
	}

	fakeEmptyListUsersOutput = &cognitoidentityprovider.ListUsersOutput{
		Users: []*cognitoidentityprovider.UserType{},
	}
)

func (s *forgotPasswordTestSuite) TestForgotPasswordHandler() {
	s.cog.On("ForgotPasswordWithContext", s.fpi).Return(fakeForgotPasswordOutput, nil).Once()
	s.cog.On("AdminGetUser", s.agui).Return(fakeEmailOutput, nil)
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusNoContent, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ForgotPasswordWithContext", 1)
}

func (s *forgotPasswordTestSuite) TestValidationError() {
	s.cog.On("ForgotPasswordWithContext", s.fpi).Return(fakeForgotPasswordOutput, nil).Once()
	s.cog.On("AdminGetUser", s.agui).Return(fakeEmailOutput, nil)
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordTestURL), "application/json", nil)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNotCalled(s.T(), "ForgotPasswordWithContext")
}

func (s *forgotPasswordTestSuite) TestUnauthorizedError() {
	s.cog.On("ForgotPasswordWithContext", mock.Anything).Return(fakeForgotPasswordOutput, errors.New("test_error")).Once()
	s.cog.On("AdminGetUser", s.agui).Return(fakeEmailOutput, nil)
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

	body := new(bytes.Buffer)
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, forgotPasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	defer res.Body.Close()

	s.cog.AssertNumberOfCalls(s.T(), "ForgotPasswordWithContext", 1)
}

func (s *forgotPasswordTestSuite) TestGetUserGroup() {
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

	group, err := s.utl.GetUserGroup("testuser")
	s.Assert().NoError(err)
	s.Assert().Equal(utils.ZendeskTicketId["group_2"], group)
}

func (s *forgotPasswordTestSuite) TestGetUserGroupEmail() {
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)
	s.cog.On("AdminGetUser", mock.Anything).Return(fakeEmailOutput, nil)

	group, err := s.utl.GetUserGroup("testuser@wikimedia.org")
	s.Assert().NoError(err)
	s.Assert().Equal(utils.ZendeskTicketId["group_2"], group)
}

func (s *forgotPasswordTestSuite) TestGetUserGroupBadUser() {
	fakeErr := errors.New("test error")
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, fakeErr)

	group, err := s.utl.GetUserGroup("bad_user")
	s.Assert().ErrorContains(err, "test error")
	s.Assert().Equal("", group)
}

func (s *forgotPasswordTestSuite) TestGetUserGroupNoUser() {
	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutput, nil)

	group, err := s.utl.GetUserGroup("testuser")
	s.Assert().NoError(err)
	s.Assert().Equal(utils.ZendeskTicketId["group_2"], group)
}

func (s *forgotPasswordTestSuite) TestGetUsernameByEmail() {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(s.env.CognitoUserPoolID),
		Filter:     aws.String(fmt.Sprintf("email = \"%s\"", "test@abc.com")),
	}

	s.cog.On("ListUsers", input).Return(fakeListUsersOutput, nil).Once()

	username, err := s.utl.GetUsernameByEmail("test@abc.com")
	s.Assert().NoError(err)
	s.Assert().Equal("testuser", username)
}

func (s *forgotPasswordTestSuite) TestGetUsernameByEmailNotFound() {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(s.env.CognitoUserPoolID),
		Filter:     aws.String(fmt.Sprintf("email = \"%s\"", "notfound@abc.com")),
	}

	s.cog.On("ListUsers", input).Return(fakeEmptyListUsersOutput, nil).Once()

	username, err := s.utl.GetUsernameByEmail("notfound@abc.com")
	s.Assert().Error(err)
	s.Assert().Equal("", username)
}

func (s *forgotPasswordTestSuite) TestGetUsernameByEmailError() {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(s.env.CognitoUserPoolID),
		Filter:     aws.String(fmt.Sprintf("email = \"%s\"", "error@abc.com")),
	}

	s.cog.On("ListUsers", input).Return(nil, errors.New("test_error")).Once()

	username, err := s.utl.GetUsernameByEmail("error@abc.com")
	s.Assert().Error(err)
	s.Assert().Equal("", username)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(forgotPasswordTestSuite))
}

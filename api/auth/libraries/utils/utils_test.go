package utils_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
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
	params *forgotpassword.Parameters
	agui   *cognitoidentityprovider.AdminGetUserInput
	utl    *utils.Parameters
}

func (s *forgotPasswordTestSuite) SetupTest() {
	s.cog = new(cognitoAPIMock)

	s.params = &forgotpassword.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}

	s.utl = &utils.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}
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
			"group":    aws.String("group_1"),
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

	fakeGroupsOutputZendesk = &cognitoidentityprovider.AdminListGroupsForUserOutput{
		Groups: []*cognitoidentityprovider.GroupType{
			{
				GroupName: aws.String("zendesk_1"),
			},
		},
	}

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

func (s *forgotPasswordTestSuite) TestGetUserGroupNoZendeskUser() {
	user := "nouser"

	s.cog.On("AdminListGroupsForUser", mock.Anything).Return(fakeGroupsOutputZendesk, nil)
	username, err := s.utl.GetUserGroup(user)
	s.Assert().ErrorContains(err, "failed to find ticket group for user group")
	s.Assert().Equal("", username)
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

func (s *forgotPasswordTestSuite) TestGetUsernameByEmailNoZendesk() {
	email := "nouser@abc.com"
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(s.env.CognitoUserPoolID),
		Filter:     aws.String(fmt.Sprintf("email = \"%s\"", email)),
	}

	s.cog.On("ListUsers", input).Return(fakeEmptyListUsersOutput, nil)
	username, err := s.utl.GetUsernameByEmail(email)
	s.Assert().ErrorContains(err, "no user found with email")
	s.Assert().Equal("", username)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(forgotPasswordTestSuite))
}

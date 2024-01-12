package createuser_test

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
	"time"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/handlers/v1/createuser"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const createUserTestURL = "/create-user"

type cognitoAPIMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (c *cognitoAPIMock) SignUpWithContext(_ aws.Context, in *cognitoidentityprovider.SignUpInput, _ ...request.Option) (*cognitoidentityprovider.SignUpOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*cognitoidentityprovider.SignUpOutput), args.Error(1)
}

func (c *cognitoAPIMock) AdminAddUserToGroupWithContext(_ aws.Context, in *cognitoidentityprovider.AdminAddUserToGroupInput, _ ...request.Option) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*cognitoidentityprovider.AdminAddUserToGroupOutput), args.Error(1)
}

func (c *cognitoAPIMock) AdminGetUserWithContext(_ aws.Context, in *cognitoidentityprovider.AdminGetUserInput, _ ...request.Option) (*cognitoidentityprovider.AdminGetUserOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*cognitoidentityprovider.AdminGetUserOutput), args.Error(1)
}

func (c *cognitoAPIMock) GetGroupWithContext(_ aws.Context, in *cognitoidentityprovider.GetGroupInput, _ ...request.Option) (*cognitoidentityprovider.GetGroupOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*cognitoidentityprovider.GetGroupOutput), args.Error(1)
}

type createUserTestSuite struct {
	suite.Suite
	env  *env.Environment
	cog  *cognitoAPIMock
	mod  *createuser.Model
	srv  *httptest.Server
	cmd  redis.Cmdable
	mcmd redismock.ClientMock
	sgi  *cognitoidentityprovider.SignUpInput
	agi  *cognitoidentityprovider.AdminAddUserToGroupInput
	gui  *cognitoidentityprovider.AdminGetUserInput
	ggi  *cognitoidentityprovider.GetGroupInput
}

func (s *createUserTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.env = new(env.Environment)
	s.env.CognitoClientID = "id"
	s.env.CognitoSecret = "supersecretkey"
	s.env.CognitoUserPoolID = "userpool"
	s.env.CognitoUserGroup = "freetiergroup"
	s.mod = &createuser.Model{
		Username:         "test",
		Email:            "test@test.com",
		Password:         "randompassword",
		CaptchaID:        uniuri.NewLen(20),
		CaptchaSolution:  uniuri.NewLenChars(6, []byte("0123456789")),
		PolicyVersion:    "v1.1",
		PolicyDateAccept: time.Now(),
		MarketingEmails:  "true",
	}
}

func (s *createUserTestSuite) SetupTest() {
	s.cog = new(cognitoAPIMock)

	s.cmd, s.mcmd = redismock.NewClientMock()

	s.sgi = new(cognitoidentityprovider.SignUpInput)
	s.sgi.SetClientId(s.env.CognitoClientID)
	s.sgi.SetUsername(s.mod.Username)
	s.sgi.SetPassword(s.mod.Password)
	s.sgi.SetUserAttributes([]*cognitoidentityprovider.AttributeType{
		{
			Name:  aws.String("email"),
			Value: aws.String(s.mod.Email),
		},
		{
			Name:  aws.String("custom:policy_version"),
			Value: aws.String(s.mod.PolicyVersion),
		},
		{
			Name:  aws.String("custom:policy_date_accept"),
			Value: aws.String(s.mod.PolicyDateAccept.Format(time.RFC3339)),
		},
		{
			Name:  aws.String("custom:marketing_emails"),
			Value: aws.String(s.mod.MarketingEmails),
		},
		{
			Name:  aws.String("custom:username"),
			Value: aws.String(s.mod.Username),
		},
	})

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))

	_, err := h.Write([]byte(fmt.Sprintf("%s%s", s.mod.Username, s.env.CognitoClientID)))
	s.Assert().NoError(err)

	s.sgi.SetSecretHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	s.agi = new(cognitoidentityprovider.AdminAddUserToGroupInput)
	s.agi.SetGroupName(s.env.CognitoUserGroup)
	s.agi.SetUserPoolId(s.env.CognitoUserPoolID)
	s.agi.SetUsername(s.mod.Username)

	s.gui = new(cognitoidentityprovider.AdminGetUserInput)
	s.gui.SetUsername(s.mod.Email)
	s.gui.SetUserPoolId(s.env.CognitoUserPoolID)

	s.ggi = new(cognitoidentityprovider.GetGroupInput)
	s.ggi.SetGroupName(s.env.CognitoUserGroup)
	s.ggi.SetUserPoolId(s.env.CognitoUserPoolID)

	s.srv = httptest.NewServer(s.createServer())
}

func (c *createUserTestSuite) TearDownTest() {
	c.srv.Close()
}

func (c *createUserTestSuite) createServer() http.Handler {
	router := gin.New()
	router.POST(createUserTestURL, createuser.NewHandler(&createuser.Parameters{
		Env:     c.env,
		Cognito: c.cog,
		Redis:   c.cmd,
	}))
	return router
}

func (s *createUserTestSuite) TestCreateUserHandler() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeUserNotFoundException, "", nil))
	s.cog.On("GetGroupWithContext", s.ggi).Return(&cognitoidentityprovider.GetGroupOutput{Group: &cognitoidentityprovider.GroupType{GroupName: &s.env.CognitoUserGroup}}, nil)
	s.cog.On("SignUpWithContext", s.sgi).Return(&cognitoidentityprovider.SignUpOutput{}, nil)
	s.cog.On("AdminAddUserToGroupWithContext", s.agi).Return(&cognitoidentityprovider.AdminAddUserToGroupOutput{}, nil)

	s.mcmd.ExpectGet(s.mod.CaptchaID).SetVal(s.mod.CaptchaSolution)
	s.mcmd.ExpectDel(s.mod.CaptchaID).SetVal(1)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusNoContent, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerGetUserFails() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeAliasExistsException, "", nil))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerGetUserExists() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerGetGroupFails() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeUserNotFoundException, "", nil))
	s.cog.On("GetGroupWithContext", s.ggi).Return(&cognitoidentityprovider.GetGroupOutput{Group: &cognitoidentityprovider.GroupType{GroupName: &s.env.CognitoUserGroup}}, errors.New("test error"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerGetFails() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeUserNotFoundException, "", nil))
	s.cog.On("GetGroupWithContext", s.ggi).Return(&cognitoidentityprovider.GetGroupOutput{Group: &cognitoidentityprovider.GroupType{GroupName: &s.env.CognitoUserGroup}}, nil)

	s.mcmd.ExpectGet(s.mod.CaptchaID).SetVal("wrong")

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusBadRequest, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerDelFails() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeUserNotFoundException, "", nil))
	s.cog.On("GetGroupWithContext", s.ggi).Return(&cognitoidentityprovider.GetGroupOutput{Group: &cognitoidentityprovider.GroupType{GroupName: &s.env.CognitoUserGroup}}, nil)

	s.mcmd.ExpectGet(s.mod.CaptchaID).SetVal(s.mod.CaptchaSolution)
	s.mcmd.ExpectDel(s.mod.CaptchaID).SetErr(errors.New("some error"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerSignUpWithContextFails() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeUserNotFoundException, "", nil))
	s.cog.On("GetGroupWithContext", s.ggi).Return(&cognitoidentityprovider.GetGroupOutput{Group: &cognitoidentityprovider.GroupType{GroupName: &s.env.CognitoUserGroup}}, nil)

	s.mcmd.ExpectGet(s.mod.CaptchaID).SetVal(s.mod.CaptchaSolution)
	s.mcmd.ExpectDel(s.mod.CaptchaID).SetVal(1)

	s.cog.On("SignUpWithContext", s.sgi).Return(&cognitoidentityprovider.SignUpOutput{}, errors.New("some error"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *createUserTestSuite) TestCreateUserHandlerAdminAddUserToGroupWithContextFails() {
	s.cog.On("AdminGetUserWithContext", s.gui).Return(&cognitoidentityprovider.AdminGetUserOutput{}, awserr.New(cognitoidentityprovider.ErrCodeUserNotFoundException, "", nil))
	s.cog.On("GetGroupWithContext", s.ggi).Return(&cognitoidentityprovider.GetGroupOutput{Group: &cognitoidentityprovider.GroupType{GroupName: &s.env.CognitoUserGroup}}, nil)

	s.mcmd.ExpectGet(s.mod.CaptchaID).SetVal(s.mod.CaptchaSolution)
	s.mcmd.ExpectDel(s.mod.CaptchaID).SetVal(1)

	s.cog.On("SignUpWithContext", s.sgi).Return(&cognitoidentityprovider.SignUpOutput{}, nil)
	s.cog.On("AdminAddUserToGroupWithContext", s.agi).Return(&cognitoidentityprovider.AdminAddUserToGroupOutput{}, errors.New("some error"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, createUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(createUserTestSuite))
}

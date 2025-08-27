package changepassword_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/handlers/v1/changepassword"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const changepasswordTestURL = "/change-password"

type changePasswordAPIMock struct {
	mock.Mock
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (m *changePasswordAPIMock) ChangePasswordWithContext(_ aws.Context, in *cognitoidentityprovider.ChangePasswordInput, _ ...request.Option) (*cognitoidentityprovider.ChangePasswordOutput, error) {
	args := m.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ChangePasswordOutput), args.Error(1)
}

type changePasswordTestSuite struct {
	suite.Suite
	env    *env.Environment
	cog    *changePasswordAPIMock
	params *changepassword.Parameters
	srv    *httptest.Server
	req    *changepassword.Request
	in     *cognitoidentityprovider.ChangePasswordInput
	out    *cognitoidentityprovider.ChangePasswordOutput
}

func (s *changePasswordTestSuite) createServer() http.Handler {
	router := gin.New()
	router.POST(changepasswordTestURL, changepassword.NewHandler(s.params))
	return router
}

func (s *changePasswordTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	s.env = new(env.Environment)
	s.out = new(cognitoidentityprovider.ChangePasswordOutput)
}

func (s *changePasswordTestSuite) SetupTest() {
	s.cog = new(changePasswordAPIMock)
	s.params = &changepassword.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}
	s.req = &changepassword.Request{
		AccessToken:      "test_token",
		PreviousPassword: "old_password",
		ProposedPassword: "new_password",
	}
	s.in = &cognitoidentityprovider.ChangePasswordInput{
		AccessToken:      aws.String(s.req.AccessToken),
		PreviousPassword: aws.String(s.req.PreviousPassword),
		ProposedPassword: aws.String(s.req.ProposedPassword),
	}
	s.srv = httptest.NewServer(s.createServer())
}

func (s *changePasswordTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *changePasswordTestSuite) TestChangePasswordHandler() {
	s.cog.On("ChangePasswordWithContext", s.in).Return(s.out, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, changepasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusNoContent, res.StatusCode)

	s.cog.AssertNumberOfCalls(s.T(), "ChangePasswordWithContext", 1)
}

func (s *changePasswordTestSuite) TestInputValidationError() {
	s.cog.On("ChangePasswordWithContext", s.in).Return(s.out, nil)

	body := bytes.NewBuffer([]byte{})
	s.req.ProposedPassword = "new"
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, changepasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)

	s.cog.AssertNotCalled(s.T(), "ChangePasswordWithContext")
}

func (s *changePasswordTestSuite) TestAPIError() {
	s.cog.On("ChangePasswordWithContext", s.in).Return(s.out, awserr.New(cognitoidentityprovider.ErrCodeNotAuthorizedException, "<AWS Message>", nil))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.req))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, changepasswordTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusUnauthorized, res.StatusCode)
	b, err := io.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Contains(string(b), "Incorrect username or password.")
	s.Assert().NotContains(string(b), "<AWS Message>")

	s.cog.AssertNumberOfCalls(s.T(), "ChangePasswordWithContext", 1)
}

func TestChangePassword(t *testing.T) {
	suite.Run(t, new(changePasswordTestSuite))
}

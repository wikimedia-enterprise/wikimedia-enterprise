package confirmuser_test

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
	"wikimedia-enterprise/api/auth/handlers/v1/confirmuser"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const confirmUserTestURL = "/confirm-user"

var userIdKey = "sub"
var userEmailKey = "email"

type confirmUserTestSuite struct {
	suite.Suite
	uid string
	uml string
	unm string
	ccd string
	env *env.Environment
	cog *confirmUserCognitoAPIMock
	mod *confirmuser.Model
	srv *httptest.Server
	in  *cognitoidentityprovider.ConfirmSignUpInput
	out *cognitoidentityprovider.ConfirmSignUpOutput
	gui *cognitoidentityprovider.AdminGetUserInput
	guo *cognitoidentityprovider.AdminGetUserOutput
}

type confirmUserCognitoAPIMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (c *confirmUserCognitoAPIMock) ConfirmSignUpWithContext(_ aws.Context, in *cognitoidentityprovider.ConfirmSignUpInput, _ ...request.Option) (*cognitoidentityprovider.ConfirmSignUpOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*cognitoidentityprovider.ConfirmSignUpOutput), args.Error(1)
}

func (c *confirmUserCognitoAPIMock) AdminGetUserWithContext(_ aws.Context, in *cognitoidentityprovider.AdminGetUserInput, _ ...request.Option) (*cognitoidentityprovider.AdminGetUserOutput, error) {
	args := c.Called(in)
	return args.Get(0).(*cognitoidentityprovider.AdminGetUserOutput), args.Error(1)
}

func (s *confirmUserTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	s.in = new(cognitoidentityprovider.ConfirmSignUpInput)
	s.out = new(cognitoidentityprovider.ConfirmSignUpOutput)

	s.mod = &confirmuser.Model{
		Username:         s.unm,
		ConfirmationCode: s.ccd,
	}

	s.in.SetUsername(s.mod.Username)
	s.in.SetClientId(s.env.CognitoClientID)
	s.in.SetConfirmationCode(s.mod.ConfirmationCode)
	s.in.ClientMetadata = map[string]*string{
		"username": aws.String(s.unm),
	}

	h := hmac.New(sha256.New, []byte(s.env.CognitoSecret))
	if _, err := h.Write([]byte(fmt.Sprintf("%s%s", s.mod.Username, s.env.CognitoClientID))); err != nil {
		return
	}

	s.in.SetSecretHash(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	s.gui = new(cognitoidentityprovider.AdminGetUserInput)
	s.gui.SetUsername(s.mod.Username)
	s.gui.SetUserPoolId(s.env.CognitoUserPoolID)

	s.guo = &cognitoidentityprovider.AdminGetUserOutput{
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  &userIdKey,
				Value: &s.uid,
			},
			{
				Name:  &userEmailKey,
				Value: &s.uml,
			},
		},
	}
}

func (s *confirmUserTestSuite) SetupTest() {
	s.cog = new(confirmUserCognitoAPIMock)
	s.srv = httptest.NewServer(s.createServer(&confirmuser.Parameters{
		Env:     s.env,
		Cognito: s.cog,
	}))
}

func (s *confirmUserTestSuite) TearDownTest() {
	s.srv.Close()
}

func (s *confirmUserTestSuite) createServer(p *confirmuser.Parameters) http.Handler {
	r := gin.New()
	r.POST(confirmUserTestURL, confirmuser.NewHandler(p))
	return r
}

func (s *confirmUserTestSuite) TestConfirmUser() {
	s.cog.On("ConfirmSignUpWithContext", s.in).Return(s.out, nil)
	s.cog.On("AdminGetUserWithContext", s.gui).Return(s.guo, nil)

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, confirmUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()
	s.Assert().Equal(http.StatusOK, res.StatusCode)

	result := new(confirmuser.Response)
	s.Assert().NoError(json.NewDecoder(res.Body).Decode(result))
	s.Assert().Equal(result.UserId, s.uid)
	s.Assert().Equal(result.Email, s.uml)
}

func (s *confirmUserTestSuite) TestConfirmUserValidationError() {
	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(confirmuser.Model{}))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, confirmUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *confirmUserTestSuite) TestConfirmUserSignUpError() {
	s.cog.On("ConfirmSignUpWithContext", s.in).Return(s.out, errors.New("Internal server error!"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, confirmUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusInternalServerError, res.StatusCode)
}

func (s *confirmUserTestSuite) TestConfirmGetUserError() {
	s.cog.On("ConfirmSignUpWithContext", s.in).Return(s.out, nil)
	s.cog.On("AdminGetUserWithContext", s.gui).Return(s.guo, errors.New("Internal server error!"))

	body := bytes.NewBuffer([]byte{})
	s.Assert().NoError(json.NewEncoder(body).Encode(s.mod))

	res, err := http.Post(fmt.Sprintf("%s%s", s.srv.URL, confirmUserTestURL), "application/json", body)
	s.Assert().NoError(err)
	defer res.Body.Close()

	s.Assert().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func TestHandler(t *testing.T) {
	for _, testCase := range []*confirmUserTestSuite{
		{
			uid: "userId",
			uml: "userEmail",
			unm: "test_name",
			ccd: "123456",
			env: &env.Environment{
				CognitoClientID: "someid",
				CognitoSecret:   "supersecretkey",
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

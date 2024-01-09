package httputil

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type cognitoIdentityProviderClientMock struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	mock.Mock
}

func (c *cognitoIdentityProviderClientMock) GetUser(input *cognitoidentityprovider.GetUserInput) (*cognitoidentityprovider.GetUserOutput, error) {
	args := c.Called(input)

	return args.Get(0).(*cognitoidentityprovider.GetUserOutput), args.Error(1)
}

type cognitoTestSuite struct {
	suite.Suite
	cip *cognitoIdentityProviderClientMock
	unm string
	tkn string
	cid string
	csc string
}

func (s *cognitoTestSuite) SetupSuite() {
	s.cip = new(cognitoIdentityProviderClientMock)
}

func (s *cognitoTestSuite) TestGetUserSuccess() {
	s.cip.On("GetUser", &cognitoidentityprovider.GetUserInput{AccessToken: &s.tkn}).
		Return(
			&cognitoidentityprovider.GetUserOutput{
				Username: &s.unm,
			},
			nil,
		)
	ap := &CognitoAuthProvider{
		CIP: s.cip,
	}

	usr, err := ap.GetUser(s.tkn)
	s.Assert().NoError(err)
	s.Assert().Equal(s.unm, usr.GetUsername())
}

func TestCognitoAuthProvider(t *testing.T) {
	for _, testCase := range []*cognitoTestSuite{
		{
			unm: "john_doe",
			tkn: "some-token",
			cid: "client-id",
			csc: "client-secret",
		},
	} {
		suite.Run(t, testCase)
	}
}

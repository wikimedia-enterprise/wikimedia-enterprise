package httputil

import (
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

// CognitoAuthProvider represents structured of auth provider.
type CognitoAuthProvider struct {
	CIP cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

// GetUser returns User from cognito.
func (ap *CognitoAuthProvider) GetUser(token string) (*User, error) {
	res, err := ap.CIP.GetUser(&cognitoidentityprovider.GetUserInput{AccessToken: &token})

	if err != nil {
		return nil, err
	}

	usr := new(User)
	usr.SetUsername(*res.Username)

	return usr, nil
}

package utils

import (
	"fmt"
	"strconv"
	"strings"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"go.uber.org/dig"
)

var ZendeskTicketId = map[string]string{
	"group_1": "8358891019803",
	"group_2": "5906748593435",
	"group_3": "4742304188059",
}

type Parameters struct {
	dig.In
	Cognito cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Env     *env.Environment
}

func NewParameters(cog cognitoidentityprovideriface.CognitoIdentityProviderAPI, env *env.Environment) *Parameters {
	return &Parameters{
		Cognito: cog,
		Env:     env,
	}
}

// GetUserGroup returns the group of the user.
func (p *Parameters) GetUserGroup(username string) (string, error) {
	gfu, err := p.Cognito.AdminListGroupsForUser(&cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(p.Env.CognitoUserPoolID),
		Username:   aws.String(username),
	})
	if err != nil {
		log.Error(err, log.Tip("failed to list user groups from cognito"))
		return "", err
	}

	var hgp string
	hgh := -1
	for _, grp := range gfu.Groups {
		pts := strings.Split(*grp.GroupName, "_")
		if len(pts) == 2 {
			num, err := strconv.Atoi(pts[1])
			if err == nil && num > hgh {
				hgh = num
				hgp = *grp.GroupName
			}
		}
	}

	if len(hgp) == 0 {
		err = fmt.Errorf("user is not a member of any groups")
		log.Error(err, log.Tip("user has no groups in cognito"))
		return "", err
	}

	zid, ok := ZendeskTicketId[hgp]
	if !ok {
		err = fmt.Errorf("failed to find ticket group for user group %s", hgp)
		log.Error(err)
		return "", err
	}

	return zid, nil
}

// GetUsernameByEmail returns the username associated with the given email.
func (p *Parameters) GetUsernameByEmail(email string) (string, error) {
	input := &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(p.Env.CognitoUserPoolID),
		Filter:     aws.String(fmt.Sprintf("email = \"%s\"", email)),
	}

	result, err := p.Cognito.ListUsers(input)
	if err != nil {
		return "", err
	}

	if len(result.Users) == 0 {
		return "", fmt.Errorf("no user found with email %s", email)
	}

	return *result.Users[0].Username, nil
}

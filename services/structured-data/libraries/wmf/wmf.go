package wmf

import (
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/config/env"
)

// NewAPI create new wmf API with custom configuration.
func NewAPI(env *env.Environment) wmf.API {
	fwc := func(c *wmf.Client) {
		c.OAuthToken = env.OauthToken
	}

	return wmf.NewAPI(fwc)
}

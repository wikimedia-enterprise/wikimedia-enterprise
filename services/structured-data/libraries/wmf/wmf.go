package wmf

import (
	"wikimedia-enterprise/general/tracing"
	"wikimedia-enterprise/general/wmf"
	"wikimedia-enterprise/services/structured-data/config/env"
)

// NewAPI create new wmf API with custom configuration.
func NewAPI(env *env.Environment, trc tracing.Tracer) wmf.API {
	fwc := func(c *wmf.Client) {
		c.OAuthToken = env.OauthToken
		c.Tracer = tracing.InjectTrace(trc)
	}

	return wmf.NewAPI(fwc)
}

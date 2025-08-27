package wmf

import (
	"wikimedia-enterprise/services/structured-data/config/env"
	"wikimedia-enterprise/services/structured-data/submodules/tracing"
	"wikimedia-enterprise/services/structured-data/submodules/wmf"
)

// NewAPI create new wmf API with custom configuration.
func NewAPI(env *env.Environment, trc tracing.Tracer) wmf.API {
	fwc := func(c *wmf.Client) {
		c.OAuthToken = env.OauthToken
		c.Tracer = tracing.InjectTrace(trc)
		c.EnableRetryAfter = false
	}

	return wmf.NewAPI(fwc)
}

package wmf

import (
	"wikimedia-enterprise/services/structured-contents/config/env"
	"wikimedia-enterprise/services/structured-contents/submodules/tracing"
	"wikimedia-enterprise/services/structured-contents/submodules/wmf"
)

// NewAPI create new wmf API with custom configuration.
func NewAPI(env *env.Environment, trc tracing.Tracer) wmf.API {
	fwc := func(c *wmf.Client) {
		c.Tracer = tracing.InjectTrace(trc)
	}

	return wmf.NewAPI(fwc)
}

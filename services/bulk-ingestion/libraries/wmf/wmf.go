// Package tracing dependency injection provider for wmf API.
package wmf

import (
	"wikimedia-enterprise/services/bulk-ingestion/config/env"

	"github.com/wikimedia-enterprise/wmf"
)

// NewAPI create new wmf API with custom configuration.
func NewAPI(env *env.Environment) wmf.API {
	fwc := func(c *wmf.Client) {
		c.DefaultURL = env.DefaultURL
		c.DefaultDatabase = env.DefaultProject
	}

	return wmf.NewAPI(fwc)
}

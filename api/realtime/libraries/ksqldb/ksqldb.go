package ksqldb

import (
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/submodules/ksqldb"
)

// New creates new ksqldb client instance with credentials.
func New(env *env.Environment) ksqldb.PushPuller {
	return ksqldb.NewClient(env.KSQLURL, func(c *ksqldb.Client) {
		if len(env.KSQLUsername) > 0 && len(env.KSQLPassword) > 0 {
			c.BasicAuth = &ksqldb.BasicAuth{
				Username: env.KSQLUsername,
				Password: env.KSQLPassword,
			}
		}
	})
}

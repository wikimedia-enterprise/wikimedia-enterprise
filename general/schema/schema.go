// Package schema provides messages and responses schema for the system.
// Tries to be compliant with https://schema.org/ whenever possible.
package schema

import "github.com/hamba/avro/v2"

// New create new avro schema from configuration.
// Recursively resolves dependencies from configuration.
func New(c *Config) (avro.Schema, error) {
	for _, cfg := range c.ResolveReferences() {
		if _, err := avro.Parse(cfg.Schema); err != nil {
			return nil, err
		}
	}

	return avro.Parse(c.Schema)
}

// Package stream dependency injection provider for schema unmarshaler.
package stream

import (
	"fmt"
	"strings"
	"wikimedia-enterprise/api/realtime/config/env"
	"wikimedia-enterprise/api/realtime/submodules/schema"
)

// New creates new schema helper for dependency injection.
func New(shr schema.GetterCreator) schema.Unmarshaler {
	return schema.NewHelper(shr, nil)
}

func NewGetterCreator(env *env.Environment) schema.GetterCreator {
	url := env.SchemaRegistryURL

	if !strings.HasPrefix(url, "http") {
		if env.SchemaRegistryCreds != nil {
			url = fmt.Sprintf("https://%s", env.SchemaRegistryURL)
		} else {
			url = fmt.Sprintf("http://%s", env.SchemaRegistryURL)
		}
	}

	shr := schema.NewRegistry(url, func(cl *schema.Registry) {
		if env.SchemaRegistryCreds != nil {
			cl.BasicAuth = &schema.BasicAuth{
				Username: env.SchemaRegistryCreds.Username,
				Password: env.SchemaRegistryCreds.Password,
			}
		}
	})

	return shr
}

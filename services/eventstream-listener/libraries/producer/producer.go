// Package producer dependency injection provider for schema producer.
package producer

import (
	"fmt"
	"strings"
	"wikimedia-enterprise/services/eventstream-listener/config/env"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// New create new schema helper for dependency injection.
func New(env *env.Environment, prod *kafka.Producer) schema.Producer {
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

	return schema.NewHelper(shr, prod)
}

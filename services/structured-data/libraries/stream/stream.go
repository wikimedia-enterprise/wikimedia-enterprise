// Package stream dependency injection provider for schema unmarshal - producer.
package stream

import (
	"fmt"
	"strings"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/structured-data/config/env"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// New creates new schema helper for dependency injection.
func New(env *env.Environment, prod *kafka.Producer) schema.UnmarshalProducer {
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

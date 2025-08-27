// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/services/eventstream-listener/submodules/schema"

	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

// Credentials set of credentials (JSON encoded).
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal JSON value.
func (c *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), c)
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers       string         `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaCreds                  *Credentials   `env:"KAFKA_CREDS"`
	RedisAddr                   string         `env:"REDIS_ADDR,required=true"`
	RedisPassword               string         `env:"REDIS_PASSWORD"`
	SchemaRegistryURL           string         `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds         *Credentials   `env:"SCHEMA_REGISTRY_CREDS"`
	PrometheusPort              int            `env:"PROMETHEUS_PORT,defaut=12411"`
	OTELCollectorAddr           string         `env:"OTEL_COLLECTOR_ADDR,default=collector:4317"`
	TracingSamplingRate         float64        `env:"TRACING_SAMPLING_RATE,default=0.1"`
	ServiceName                 string         `env:"SERVICE_NAME,default=event-bridge.service"`
	OutputTopics                *schema.Topics `env:"OUTPUT_TOPICS,default={}"`
	HealthChecksPort            int            `env:"HEALTH_CHECKS_PORT,default=8082"`
	HealthChecksAsyncIntervalMs int            `env:"HEALTH_CHECKS_ASYNC_INTERVAL_MS,default=10000"`
}

// New initialize the environment.
func New() (*Environment, error) {
	var (
		_, b, _, _ = runtime.Caller(0)
		base       = filepath.Dir(b)
		_          = godotenv.Load(fmt.Sprintf("%s/../../.env", base))
		config     = new(Environment)
	)

	_, err := env.UnmarshalFromEnviron(config)
	return config, err
}

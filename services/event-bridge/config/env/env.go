// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

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
	KafkaBootstrapServers        string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaCreds                   *Credentials `env:"KAFKA_CREDS"`
	RedisAddr                    string       `env:"REDIS_ADDR,required=true"`
	RedisPassword                string       `env:"REDIS_PASSWORD"`
	SchemaRegistryURL            string       `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds          *Credentials `env:"SCHEMA_REGISTRY_CREDS"`
	TopicArticleDelete           string       `env:"TOPIC_ARTICLE_DELETE,default=aws.event-bridge.article-delete.v1"`
	TopicArticleUpdate           string       `env:"TOPIC_ARTICLE_UPDATE,default=aws.event-bridge.article-update.v1"`
	TopicArticleCreate           string       `env:"TOPIC_ARTICLE_CREATE,default=aws.event-bridge.article-create.v1"`
	TopicArticleMove             string       `env:"TOPIC_ARTICLE_MOVE,default=aws.event-bridge.article-move.v1"`
	TopicArticleVisibilityChange string       `env:"TOPIC_ARTICLE_VISIBILITY_CHANGE,default=aws.event-bridge.article-visibility-change.v1"`
	PrometheusPort               int          `env:"PROMETHEUS_PORT,defaut=12411"`
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

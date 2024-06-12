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

// Credentials kafka SASL/SSL credentials.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (c *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), c)
}

// Topics list of topics to subscribe to.
type Topics []string

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (t *Topics) UnmarshalEnvironmentValue(data string) error {
	_ = json.Unmarshal([]byte(data), t)

	if len(*t) == 0 {
		*t = []string{data}
	}

	return nil
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaConsumerGroupID  string       `env:"KAFKA_CONSUMER_GROUP_ID,required=true"`
	KafkaCreds            *Credentials `env:"KAFKA_CREDS"`
	SchemaRegistryURL     string       `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds   *Credentials `env:"SCHEMA_REGISTRY_CREDS"`
	TopicArticles         Topics       `env:"TOPIC_ARTICLES,default=aws.structured-data.articles.v1"`
	AWSURL                string       `env:"AWS_URL"`
	AWSRegion             string       `env:"AWS_REGION"`
	AWSBucket             string       `env:"AWS_BUCKET"`
	AWSKey                string       `env:"AWS_KEY"`
	AWSID                 string       `env:"AWS_ID"`
	NumberOfWorkers       int          `env:"NUMBER_OF_WORKERS,default=15"`
	EventChannelSize      int          `env:"EVENT_CHANNEL_SIZE,default=1000000"`
	ArticleKeyTypeSuffix  string       `env:"KEY_TYPE_SUFFIX"`
}

// New initialize the environment
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

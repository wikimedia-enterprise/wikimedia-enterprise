// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/services/structured-contents/submodules/schema"

	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

// Credentials set of credentials (JSON encoded).
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (c *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), c)
}

// List of Strings from environment variables (JSON Encoded).
type List []string

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (l *List) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaConsumerGroupID  string       `env:"KAFKA_CONSUMER_GROUP_ID,required=true"`
	KafkaCreds            *Credentials `env:"KAFKA_CREDS"`
	KafkaAutoOffsetReset  string       `env:"KAFKA_AUTO_OFFSET_RESET"`
	KafkaMaxPollInterval  int          `env:"KAFKA_MAX_POLL_INTERVAL"` // Max ms interval allowed between two polls by a consumer
	SchemaRegistryURL     string       `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds   *Credentials `env:"SCHEMA_REGISTRY_CREDS"`
	NumberOfWorkers       int          `env:"NUMBER_OF_WORKERS,default=5"`
	EventChannelSize      int          `env:"EVENT_CHANNEL_SIZE,default=1000000"`
	// INPUT_TOPICS is what differentiates the structured-update service from the structured-bulk service
	// For structured-update, the input topic is just one - aws.structured-data.articles.v1
	// For structured-bulk, create several services using this image and passing chunks of aws.structured-data.{project}-articles-compacted.v2 topics
	InputTopics           List           `env:"INPUT_TOPICS,default=[\"aws.structured-data.articles.v1\"]"`
	OutputCompactedTopics *schema.Topics `env:"OUTPUT_COMPACTED_TOPICS,default={}"`

	// For structured-bulk, do not publish to non compacted topic
	OutputNoncompactedTopic string `env:"OUTPUT_NONCOMPACTED_TOPIC,default=aws.structured-contents.articles.v1"`

	// Error, deadletter topics for sturctured-update and structured-bulk service
	// For structured-update service, set this to aws.structured-contents.update-dead-letter.v1 from IaC
	// For structured-bulk service, set this to aws.structured-contents.bulk-dead-letter.v1 from IaC
	DeadLetterTopic string `env:"DEAD_LETTER_TOPIC"`
	MaxFailCount    int    `env:"MAX_FAIL_COUNT,default=0"`

	TracingGRPCHost     string  `env:"OTEL_COLLECTOR_ADDR,default=collector"`
	TracingSamplingRate float64 `env:"TRACING_SAMPLING_RATE,default=0.1"`
	ServiceName         string  `env:"SERVICE_NAME,default=structured-contents.service"`
	PrometheusPort      int     `env:"PROMETHEUS_PORT,defaut=12411"`
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

	// Override service-name for structured contents
	config.OutputCompactedTopics.ServiceName = "structured-contents"

	return config, err
}

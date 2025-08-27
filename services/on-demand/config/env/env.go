// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	conf "wikimedia-enterprise/services/on-demand/submodules/config"
	"wikimedia-enterprise/services/on-demand/submodules/log"

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

// List of Strings from environment variables (JSON Encoded).
type List []string

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (l *List) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), l)
}

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (t *Topics) UnmarshalEnvironmentValue(data string) error {

	if len(data) == 0 {
		*t = []string{data}
		return nil
	}

	log.Info("data output", log.Any("data", data))
	if err := json.Unmarshal([]byte(data), t); err != nil {
		log.Info("failed to unmarshal topics", log.Any("error", err))
		return err
	}

	if len(*t) == 0 {
		*t = []string{data}
	}

	return nil
}

type ConfigMap map[string]string

func (m *ConfigMap) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), m)
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers      string                   `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaConsumerGroupID       string                   `env:"KAFKA_CONSUMER_GROUP_ID,required=true"`
	KafkaCreds                 *Credentials             `env:"KAFKA_CREDS"`
	KafkaAutoOffsetReset       string                   `env:"KAFKA_AUTO_OFFSET_RESET"`
	KafkaExtraConfig           ConfigMap                `env:"KAFKA_EXTRA_CONFIG,default={}"`
	SchemaRegistryURL          string                   `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds        *Credentials             `env:"SCHEMA_REGISTRY_CREDS"`
	TopicArticles              Topics                   `env:"TOPIC_ARTICLES"`
	Topics                     *conf.TopicArticleConfig `env:"TOPICS"`
	AWSURL                     string                   `env:"AWS_URL"`
	AWSRegion                  string                   `env:"AWS_REGION"`
	AWSBucket                  string                   `env:"AWS_BUCKET"`
	AWSKey                     string                   `env:"AWS_KEY"`
	AWSID                      string                   `env:"AWS_ID"`
	NumberOfWorkers            int                      `env:"NUMBER_OF_WORKERS,default=15"`
	EventChannelSize           int                      `env:"EVENT_CHANNEL_SIZE,default=1000000"`
	ArticleKeyTypeSuffix       string                   `env:"KEY_TYPE_SUFFIX"`
	TracingGrpcHost            string                   `env:"OTEL_COLLECTOR_ADDR,default=collector"`
	TracingGrpcPort            string                   `env:"TRACING_GRPC_PORT,default=4317"`
	TracingSamplingRate        float64                  `env:"TRACING_SAMPLING_RATE,default=0.1"`
	ServiceName                string                   `env:"SERVICE_NAME,default=on-demand.service"`
	PrometheusPort             int                      `env:"PROMETHEUS_PORT,default=12411"`
	KafkaHealthCheckIntervalMs int                      `env:"KAFKA_HEALTH_CHECK_INTERVAL_MS,default=10000"`
	KafkaHealthCheckLagNMS     int                      `env:"KAFKA_HEALTH_CHECK_LAG_NMS,default=1000"`
	HealthCheckPort            int                      `env:"HEALTH_CHECK_PORT,default=8855"`
	HealthCheckMaxRetries      int                      `env:"HEALTH_CHECK_MAX_RETRIES,default=3"`
	HealthChecksConsumerTopics List                     `env:"HEALTH_CHECKS_CONSUMER_TOPIC,default=[\"aws.structured-data.enwiki-articles-compacted.v2\"]"`
	UseHashedPrefixes          bool                     `env:"USE_HASHED_PREFIXES,default=false"`
	MaxMsgsPerSecond           int                      `env:"MAX_MSGS_PER_SECOND,default=1000"`
	WorkerBufferSize           int                      `env:"WORKER_BUFFER_SIZE,default=5"`
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
	if err != nil {
		return nil, err
	}

	con, err := conf.New()
	if err != nil {
		return nil, err
	}

	if config.Topics != nil && len(config.TopicArticles) == 0 {
		topics, err := con.GetArticleTopics(config.Topics)
		if err != nil {
			log.Info("failed to load TopicArticles from TOPICS", log.Any("error", err))
			return config, nil
		}

		config.TopicArticles = topics
	}

	return config, nil
}

// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/services/snapshots/submodules/schema"

	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

// Credentials SASL/SSL credentials.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (kc *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), kc)
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaCreds            *Credentials `env:"KAFKA_CREDS"`
	KafkaAutoOffsetReset  string       `env:"KAFKA_AUTO_OFFSET_RESET"`
	KafkaDebugOptions     string       `env:"KAFKA_DEBUG_OPTIONS"`
	KafkaLogWatermark     bool         `env:"KAFKA_LOG_WATERMARK"`
	// If set, consumer timeouts will be retried a few times. Consumers will use the high watermark offset
	// to know when to stop.
	KafkaUseWatermark   bool           `env:"KAFKA_USE_WATERMARK"`
	Topics              *schema.Topics `env:"TOPICS,default={}"`
	SchemaRegistryURL   string         `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds *Credentials   `env:"SCHEMA_REGISTRY_CREDS"`
	ServerPort          string         `env:"SERVER_PORT,default=5050"`
	AWSURL              string         `env:"AWS_URL"`
	AWSRegion           string         `env:"AWS_REGION"`
	AWSBucket           string         `env:"AWS_BUCKET"`
	AWSBucketCommons    string         `env:"AWS_BUCKET_COMMONS"`
	AWSKey              string         `env:"AWS_KEY"`
	AWSID               string         `env:"AWS_ID"`
	Prefix              string         `env:"PREFIX,default=snapshots"`
	InternalRootCAPem   string         `env:"INTERNAL_ROOT_CA_PEM"`
	TLSCertPem          string         `env:"TLS_CERT_PEM"`
	TLSPrivateKeyPem    string         `env:"TLS_PRIVATE_KEY_PEM"`
	PartitionConfig     string         `env:"PARTITION_CONFIG"`
	BufferSize          int            `env:"BUFFER_SIZE,default=100000000"`
	PrometheusPort      int            `env:"PROMETHEUS_PORT,default=12411"`
	MessagesChannelCap  int            `env:"MESSAGES_CHANNEL_CAP,default=100"`
	PipeBufferSize      int            `env:"PIPE_BUFFER_SIZE,default=100000000"`
	UploadPartSize      int            `env:"UPLOAD_PART_SIZE,default=100000000"`
	UploadConcurrency   int            `env:"UPLOAD_CONCURRENCY,default=10"`
	// FreeTierGroup is a group to create monthly snapshots copy for.
	FreeTierGroup                    string  `env:"FREE_TIER_GROUP,default=group_1"`
	OTELCollectorAddr                string  `env:"OTEL_COLLECTOR_ADDR,default=collector:4317"`
	TracingSamplingRate              float64 `env:"TRACING_SAMPLING_RATE,default=0.1"`
	ServiceName                      string  `env:"SERVICE_NAME,default=snapshot.service"`
	ChunkWorkers                     int     `env:"CHUNK_WORKERS,default=4"`
	LineLimit                        int     `env:"LINE_LIMIT,default=5000"`
	LogInterval                      int     `env:"LOG_INTERVAL,default=30"` // aggregatecommons logs progress every x mins
	CommonsAggregationWorkercount    int     `env:"COMMONS_AGGREGATION_WORKER_COUNT,default=100"`
	CommonsAggregationKeyChannelSize int     `env:"COMMONS_AGGREGATION_KEY_CHANNEL_SIZE,default=1000000"`
}

// TLSEnabled checks if TLS enabled in current environment.
func (env *Environment) TLSEnabled() bool {
	return false
	// return len(env.InternalRootCAPem) > 0 &&
	// 	len(env.TLSCertPem) > 0 &&
	// 	len(env.TLSPrivateKeyPem) > 0
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

	if len(config.AWSBucketCommons) == 0 {
		config.AWSBucketCommons = config.AWSBucket
	}

	return config, err
}

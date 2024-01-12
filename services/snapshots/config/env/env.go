// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/general/schema"

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
	KafkaBootstrapServers string         `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaCreds            *Credentials   `env:"KAFKA_CREDS"`
	Topics                *schema.Topics `env:"TOPICS,default={}"`
	SchemaRegistryURL     string         `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds   *Credentials   `env:"SCHEMA_REGISTRY_CREDS"`
	ServerPort            string         `env:"SERVER_PORT,default=5050"`
	AWSURL                string         `env:"AWS_URL"`
	AWSRegion             string         `env:"AWS_REGION"`
	AWSBucket             string         `env:"AWS_BUCKET"`
	AWSKey                string         `env:"AWS_KEY"`
	AWSID                 string         `env:"AWS_ID"`
	Prefix                string         `env:"PREFIX,default=snapshots"`
	InternalRootCAPem     string         `env:"INTERNAL_ROOT_CA_PEM"`
	TLSCertPem            string         `env:"TLS_CERT_PEM"`
	TLSPrivateKeyPem      string         `env:"TLS_PRIVATE_KEY_PEM"`
	PartitionConfig       string         `env:"PARTITION_CONFIG"`
	BufferSize            int            `env:"BUFFER_SIZE,default=100000000"`
	PrometheusPort        int            `env:"PROMETHEUS_PORT,default=12411"`
	MessagesChannelCap    int            `env:"MESSAGES_CHANNEL_CAP,default=100"`
	PipeBufferSize        int            `env:"PIPE_BUFFER_SIZE,default=100000000"`
	UploadPartSize        int            `env:"UPLOAD_PART_SIZE,default=100000000"`
	UploadConcurrency     int            `env:"UPLOAD_CONCURRENCY,default=10"`
	// FreeTierGroup is a group to create monthly snapshots copy for.
	FreeTierGroup string `env:"FREE_TIER_GROUP,default=group_1"`
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
	return config, err
}

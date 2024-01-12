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

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value
func (kc *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), kc)
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	ServerPort            string       `env:"SERVER_PORT,default=50051"`
	KafkaCreds            *Credentials `env:"KAFKA_CREDS"`
	SchemaRegistryURL     string       `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds   *Credentials `env:"SCHEMA_REGISTRY_CREDS"`
	MediawikiClientURL    string       `env:"MEDIAWIKI_CLIENT_URL,default=https://en.wikipedia.org/"`
	TopicProjects         string       `env:"TOPIC_PROJECTS,default=aws.bulk-ingestion.projects.v1"`
	TopicLanguages        string       `env:"TOPIC_LANGUAGES,default=aws.bulk-ingestion.languages.v1"`
	TopicNamespaces       string       `env:"TOPIC_NAMESPACES,default=aws.bulk-ingestion.namespaces.v1"`
	TopicArticles         string       `env:"TOPIC_ARTICLES,default=aws.bulk-ingestion.article-names.v1"`
	DefaultProject        string       `env:"DEFAULT_PROJECT,default=enwiki"`
	InternalRootCAPem     string       `env:"INTERNAL_ROOT_CA_PEM"`
	TLSCertPem            string       `env:"TLS_CERT_PEM"`
	TLSPrivateKeyPem      string       `env:"TLS_PRIVATE_KEY_PEM"`
	AWSURL                string       `env:"AWS_URL"`
	AWSRegion             string       `env:"AWS_REGION"`
	AWSBucket             string       `env:"AWS_BUCKET"`
	AWSKey                string       `env:"AWS_KEY"`
	AWSID                 string       `env:"AWS_ID"`
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

// Package config holds a configuration variables for the service.
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

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (c *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), c)
}

// List of Strings from environment variables ( JSON Encoded ).
type List []string

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (t *List) UnmarshalEnvironmentValue(data string) error {
	_ = json.Unmarshal([]byte(data), t)

	if len(*t) == 0 {
		*t = []string{data}
	}

	return nil
}

// Environment environment variables configuration.
type Environment struct {
	ServerPort                        string       `env:"SERVER_PORT,default=5050"`
	PrometheusPort                    int          `env:"PROMETHEUS_PORT,default=12411"`
	RedisAddr                         string       `env:"REDIS_ADDR,required=true"`
	RedisPassword                     string       `env:"REDIS_PASSWORD"`
	KafkaBootstrapServers             string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaConsumerGroupID              string       `env:"KAFKA_CONSUMER_GROUP_ID,default=content-integrity"`
	KafkaCreds                        *Credentials `env:"KAFKA_CREDS"`
	NumberOfWorkers                   int          `env:"NUMBER_OF_WORKERS,default=15"`
	EventChannelSize                  int          `env:"EVENT_CHANNEL_SIZE,default=1000000"`
	SchemaRegistryCreds               *Credentials `env:"SCHEMA_REGISTRY_CREDS"`
	SchemaRegistryURL                 string       `env:"SCHEMA_REGISTRY_URL,required=true"`
	TopicArticleCreate                string       `env:"TOPIC_ARTICLE_CREATE,default=aws.event-bridge.article-create.v1"`
	TopicArticleUpdate                string       `env:"TOPIC_ARTICLE_UPDATE,default=aws.event-bridge.article-update.v1"`
	TopicArticleMove                  string       `env:"TOPIC_ARTICLE_MOVE,default=aws.event-bridge.article-move.v1"`
	BreakingNewsKeysExpiration        int          `env:"BREAKING_NEWS_KEYS_EXPIRATION,default=48"`
	BreakingNewsMandatoryTemplates    List         `env:"BREAKING_NEWS_TEMPLATES,default=[\"Template:Cite news\"]"`
	BreakingNewsTemplatesPrefix       List         `env:"BREAKING_NEWS_TEMPLATES_PREFIX,default=[\"Template:Current\"]"`
	BreakingNewsTemplatesPrefixIgnore List         `env:"BREAKING_NEWS_TEMPLATES_PREFIX_IGNORE,default=[]"`
	BreakingNewsCreatedHours          int          `env:"BREAKING_NEWS_CREATED_HOURS,default=24"`
	BreakingNewsMovedHours            int          `env:"BREAKING_NEWS_MOVED_HOURS,default=24"`
	BreakingNewsUniqueEditors         int          `env:"BREAKING_NEWS_UNIQUE_EDITORS,default=2"`
	BreakingNewsUniqueEditorsHours    int          `env:"BREAKING_NEWS_UNIQUE_EDITORS_HOURS,default=1"`
	BreakingNewsEditsHours            int          `env:"BREAKING_NEWS_EDITS_HOURS,default=1"`
	BreakingNewsEdits                 int          `env:"BREAKING_NEWS_EDITS,default=10"`
}

// New initialize new environment configuration.
func New() (*Environment, error) {
	var (
		_, b, _, _ = runtime.Caller(0)
		base       = filepath.Dir(b)
		_          = godotenv.Load(fmt.Sprintf("%s/../../.env", base))
		cfg        = new(Environment)
	)

	_, err := env.UnmarshalFromEnviron(cfg)
	return cfg, err
}

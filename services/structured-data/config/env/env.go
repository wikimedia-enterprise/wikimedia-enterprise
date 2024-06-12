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
	KafkaBootstrapServers        string         `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaConsumerGroupID         string         `env:"KAFKA_CONSUMER_GROUP_ID,required=true"`
	KafkaCreds                   *Credentials   `env:"KAFKA_CREDS"`
	KafkaAutoOffsetReset         string         `env:"KAFKA_AUTO_OFFSET_RESET"`
	KafkaMaxPollInterval         int            `env:"KAFKA_MAX_POLL_INTERVAL"` // Max ms interval allowed between two polls by a consumer
	SchemaRegistryURL            string         `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds          *Credentials   `env:"SCHEMA_REGISTRY_CREDS"`
	Topics                       *schema.Topics `env:"TOPICS,default={}"`
	TextProcessorURL             string         `env:"TEXT_PROCESSOR_URL,default=textprocessor:5050"`
	EnableDiffs                  bool           `env:"ENABLE_DIFFS"`
	ContentIntegrityURL          string         `env:"CONTENT_INTEGRITY_URL,required=true"`
	BreakingNewsEnabled          bool           `env:"BREAKING_NEWS_ENABLED,default=false"`
	TopicArticles                string         `env:"TOPIC_ARTICLES,default=aws.structured-data.articles.v1"`
	TopicArticleVisibilityChange string         `env:"TOPIC_ARTICLE_VISIBILITY_CHANGE,default=aws.event-bridge.article-visibility-change.v1"`
	TopicArticleDelete           string         `env:"TOPIC_ARTICLE_DELETE,default=aws.event-bridge.article-delete.v1"`
	TopicArticleDeleteError      string         `env:"TOPIC_ARTICLE_DELETE_ERROR,default=aws.structured-data.article-delete-error.v1"`
	TopicArticleDeleteDeadLetter string         `env:"TOPIC_ARTICLE_DELETE_DEAD_LETTER,default=aws.structured-data.article-delete-dead-letter.v1"`
	TopicArticleUpdate           string         `env:"TOPIC_ARTICLE_UPDATE,default=aws.event-bridge.article-update.v1"`
	TopicArticleUpdateError      string         `env:"TOPIC_ARTICLE_UPDATE_ERROR,default=aws.structured-data.article-update-error.v1"`
	TopicArticleUpdateDeadLetter string         `env:"TOPIC_ARTICLE_UPDATE_DEAD_LETTER,default=aws.structured-data.article-update-dead-letter.v1"`
	TopicArticleBulk             string         `env:"TOPIC_ARTICLE_BULK,default=aws.bulk-ingestion.article-names.v1"`
	TopicArticleBulkError        string         `env:"TOPIC_ARTICLE_BULK_ERROR,default=aws.structured-data.article-bulk-error.v1"`
	TopicArticleBulkDeadLetter   string         `env:"TOPIC_ARTICLE_BULK_DEAD_LETTER,default=aws.structured-data.article-bulk-dead-letter.v1"`
	MaxFailCount                 int            `env:"MAX_FAIL_COUNT,default=10"`
	NumberOfWorkers              int            `env:"NUMBER_OF_WORKERS,default=5"`
	EventChannelSize             int            `env:"EVENT_CHANNEL_SIZE,default=1000000"`
	BackOffBase                  int            `env:"BACK_OFF_BASE,default=2"`
	OauthToken                   string         `env:"WMF_OAUTH_TOKEN"`
	NoindexTemplatePatterns      List           `env:"NOINDEX_TEMPLATE_PATTERNS,default=[]"`
	NoindexCategoryPatterns      List           `env:"NOINDEX_CATEGORY_PATTERNS,default=[]"`
	LatencyThresholdMS           int64          `env:"LATENCY_THRESHOLD_MS,default=500"`
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

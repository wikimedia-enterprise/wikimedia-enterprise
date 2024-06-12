package env_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/services/structured-data/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey        string
	kafkaBootstrapServers           string
	kafkaConsumerGroupID            string
	kafkaConsumerGroupIDKey         string
	kafkaCreds                      string
	kafkaCredsKey                   string
	kafkaAutoOffsetReset            string
	kafkaAutoOffsetResetKey         string
	kafkaMaxPollIntervalKey         string
	kafkaMaxPollInterval            int
	maxFailCount                    string
	maxFailCountKey                 string
	topicArticleDelete              string
	topicArticleDeleteKey           string
	topicArticleDeleteError         string
	topicArticleDeleteErrorKey      string
	topicArticleDeleteDeadLetter    string
	topicArticleDeleteDeadLetterKey string
	topicArticleUpdate              string
	topicArticleUpdateKey           string
	topicArticleUpdateError         string
	topicArticleUpdateErrorKey      string
	topicArticleUpdateDeadLetter    string
	topicArticleUpdateDeadLetterKey string
	topicArticleBulk                string
	topicArticleBulkKey             string
	topicArticleBulkError           string
	topicArticleBulkErrorKey        string
	topicArticleBulkDeadLetter      string
	topicArticleBulkDeadLetterKey   string
	schemaRegistryURLKey            string
	schemaRegistryURL               string
	schemaRegistryCredsKey          string
	schemaRegistryCreds             string
	topics                          string
	topicsKey                       string
	topicArticlesKey                string
	topicArticles                   string
	topicArticleVisibilityChangeKey string
	topicArticleVisibilityChange    string
	textProcessorURLKey             string
	textProcessorURL                string
	enableDiffsKey                  string
	enableDiffs                     bool
	numberOfWorkersKey              string
	numberOfWorkers                 int
	contentintegrityURLKey          string
	contentintegrityURL             string
	breakingNewsEnabledKey          string
	breakingNewsEnabled             bool
	oauthToken                      string
	oauthTokenKey                   string
	noindexTemplatePatternsKey      string
	noindexTemplatePatterns         env.List
	noindexCategoryPatternsKey      string
	noindexCategoryPatterns         env.List
	LatencyThresholdMS              int64
	LatencyThresholdKey             string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.kafkaConsumerGroupIDKey = "KAFKA_CONSUMER_GROUP_ID"
	s.kafkaConsumerGroupID = "unique"
	s.kafkaAutoOffsetReset = "earliest"
	s.kafkaAutoOffsetResetKey = "KAFKA_AUTO_OFFSET_RESET"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.kafkaMaxPollIntervalKey = "KAFKA_MAX_POLL_INTERVAL"
	s.kafkaMaxPollInterval = 0
	s.maxFailCountKey = "MAX_FAIL_COUNT"
	s.maxFailCount = "1"
	s.topicArticleDeleteKey = "TOPIC_ARTICLE_DELETE"
	s.topicArticleDelete = "local.event-bridge.article-delete.v1"
	s.topicArticleDeleteErrorKey = "TOPIC_ARTICLE_DELETE_ERROR"
	s.topicArticleDeleteError = "local.structured-data.article-delete-error.v1"
	s.topicArticleDeleteDeadLetterKey = "TOPIC_ARTICLE_DELETE_DEAD_LETTER"
	s.topicArticleDeleteDeadLetter = "local.structured-data.article-delete-dead-letter.v1"
	s.topicArticleUpdateKey = "TOPIC_ARTICLE_UPDATE"
	s.topicArticleUpdate = "local.event-bridge.article-update.v1"
	s.topicArticleUpdateErrorKey = "TOPIC_ARTICLE_UPDATE_ERROR"
	s.topicArticleUpdateError = "local.structured-data.article-update-error.v1"
	s.topicArticleUpdateDeadLetterKey = "TOPIC_ARTICLE_UPDATE_DEAD_LETTER"
	s.topicArticleUpdateDeadLetter = "local.structured-data.article-update-dead-letter.v1"

	s.topics = `{"version":["v1"],"service_name":"structured-data","location":"local"}`
	s.topicsKey = "TOPICS"
	s.topicArticleBulkKey = "TOPIC_ARTICLE_BULK"
	s.topicArticleBulk = "local.bulk-ingestion.articles.v1"
	s.topicArticleBulkErrorKey = "TOPIC_ARTICLE_bulk_ERROR"
	s.topicArticleBulkError = "local.structured-data.article-bulk-error.v1"
	s.topicArticleBulkDeadLetterKey = "TOPIC_ARTICLE_BULK_DEAD_LETTER"
	s.topicArticleBulkDeadLetter = "local.structured-data.article-bulk-dead-letter.v1"

	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "localhost:2020"
	s.schemaRegistryCredsKey = "SCHEMA_REGISTRY_CREDS"
	s.schemaRegistryCreds = `{"username":"reg_admin","password":"654321"}`
	s.topicArticlesKey = "TOPIC_ARTICLES"
	s.topicArticles = "local.structured-data.versions.v1"
	s.topicArticleVisibilityChangeKey = "TOPIC_ARTICLE_VISIBILITY_CHANGE"
	s.topicArticleVisibilityChange = "local.event-bridge.article-visibility-change.v1"
	s.textProcessorURLKey = "TEXT_PROCESSOR_URL"
	s.textProcessorURL = "localhost:5050"

	s.enableDiffsKey = "ENABLE_DIFFS"
	s.enableDiffs = true

	s.numberOfWorkersKey = "NUMBER_OF_WORKERS"
	s.numberOfWorkers = 10

	s.breakingNewsEnabledKey = "BREAKING_NEWS_ENABLED"
	s.breakingNewsEnabled = true
	s.contentintegrityURLKey = "CONTENT_INTEGRITY_URL"
	s.contentintegrityURL = "localhost:5020"
	s.oauthTokenKey = "WMF_OAUTH_TOKEN"
	s.oauthToken = "o-auth-token-testing"

	s.noindexTemplatePatternsKey = "NOINDEX_TEMPLATE_PATTERNS"
	s.noindexTemplatePatterns = env.List{"noindex"}
	s.noindexCategoryPatternsKey = "NOINDEX_CATEGORY_PATTERNS"
	s.noindexCategoryPatterns = env.List{"noindex"}
	s.LatencyThresholdKey = "LATENCY_THRESHOLD_MS"
	s.LatencyThresholdMS = 500
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaConsumerGroupIDKey, s.kafkaConsumerGroupID)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.kafkaAutoOffsetResetKey, s.kafkaAutoOffsetReset)
	os.Setenv(s.maxFailCountKey, s.maxFailCount)

	os.Setenv(s.topicsKey, s.topics)
	os.Setenv(s.topicArticleDeleteKey, s.topicArticleDelete)
	os.Setenv(s.topicArticleDeleteErrorKey, s.topicArticleDeleteError)
	os.Setenv(s.topicArticleDeleteDeadLetterKey, s.topicArticleDeleteDeadLetter)
	os.Setenv(s.topicArticleUpdateKey, s.topicArticleUpdate)
	os.Setenv(s.topicArticleUpdateErrorKey, s.topicArticleUpdateError)
	os.Setenv(s.topicArticleUpdateDeadLetterKey, s.topicArticleUpdateDeadLetter)
	os.Setenv(s.topicArticleBulkKey, s.topicArticleBulk)
	os.Setenv(s.topicArticleBulkErrorKey, s.topicArticleBulkError)
	os.Setenv(s.topicArticleBulkDeadLetterKey, s.topicArticleBulkDeadLetter)

	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.schemaRegistryCredsKey, s.schemaRegistryCreds)
	os.Setenv(s.topicArticlesKey, s.topicArticles)
	os.Setenv(s.topicArticleVisibilityChangeKey, s.topicArticleVisibilityChange)
	os.Setenv(s.textProcessorURLKey, s.textProcessorURL)
	os.Setenv(s.enableDiffsKey, strconv.FormatBool(s.enableDiffs))

	os.Setenv(s.numberOfWorkersKey, strconv.Itoa(s.numberOfWorkers))
	os.Setenv(s.contentintegrityURLKey, s.contentintegrityURL)
	os.Setenv(s.breakingNewsEnabledKey, strconv.FormatBool(s.breakingNewsEnabled))
	os.Setenv(s.oauthTokenKey, s.oauthToken)
	os.Setenv(s.LatencyThresholdKey, strconv.FormatInt(s.LatencyThresholdMS, 10))

	tmp, err := json.Marshal(s.noindexTemplatePatterns)
	s.Assert().NoError(err)
	os.Setenv(s.noindexTemplatePatternsKey, string(tmp))

	ctg, err := json.Marshal(s.noindexCategoryPatterns)
	s.Assert().NoError(err)
	os.Setenv(s.noindexCategoryPatternsKey, string(ctg))
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaAutoOffsetReset, env.KafkaAutoOffsetReset)
	s.Assert().Equal(s.kafkaMaxPollInterval, env.KafkaMaxPollInterval)

	max, err := strconv.Atoi(s.maxFailCount)
	s.Assert().NoError(err)
	s.Assert().Equal(max, env.MaxFailCount)
	s.Assert().Equal(s.topicArticleDelete, env.TopicArticleDelete)
	s.Assert().Equal(s.topicArticleDeleteError, env.TopicArticleDeleteError)
	s.Assert().Equal(s.topicArticleDeleteDeadLetter, env.TopicArticleDeleteDeadLetter)
	s.Assert().Equal(s.topicArticleUpdate, env.TopicArticleUpdate)
	s.Assert().Equal(s.topicArticleUpdateError, env.TopicArticleUpdateError)
	s.Assert().Equal(s.topicArticleUpdateDeadLetter, env.TopicArticleUpdateDeadLetter)
	s.Assert().Equal(s.kafkaConsumerGroupID, env.KafkaConsumerGroupID)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)
	s.Assert().Equal(s.topicArticles, env.TopicArticles)
	s.Assert().Equal(s.topicArticleVisibilityChange, env.TopicArticleVisibilityChange)
	s.Assert().Equal(s.textProcessorURL, env.TextProcessorURL)
	s.Assert().Equal(s.enableDiffs, env.EnableDiffs)
	s.Assert().Equal(s.numberOfWorkers, env.NumberOfWorkers)
	s.Assert().Equal(s.LatencyThresholdMS, env.LatencyThresholdMS)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	creds, err = json.Marshal(env.SchemaRegistryCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.schemaRegistryCreds, string(creds))

	tps, err := json.Marshal(env.Topics)
	s.Assert().NoError(err)
	s.Assert().Equal(s.topics, string(tps))

	s.Assert().Equal(s.contentintegrityURL, env.ContentIntegrityURL)
	s.Assert().Equal(s.breakingNewsEnabled, env.BreakingNewsEnabled)
	s.Assert().Equal(s.oauthToken, env.OauthToken)

	s.Assert().Equal(s.noindexTemplatePatterns, env.NoindexTemplatePatterns)
	s.Assert().Equal(s.noindexCategoryPatterns, env.NoindexCategoryPatterns)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

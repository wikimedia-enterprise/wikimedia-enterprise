package env_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/submodules/config"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey      string
	kafkaBootstrapServers         string
	kafkaConsumerGroupID          string
	kafkaConsumerGroupIDKey       string
	kafkaCreds                    string
	kafkaCredsKey                 string
	schemaRegistryURLKey          string
	schemaRegistryURL             string
	schemaRegistryCredsKey        string
	schemaRegistryCreds           string
	topicArticlesKey              string
	topicArticles                 string
	AWSURLKey                     string
	AWSURL                        string
	AWSRegionKey                  string
	AWSRegion                     string
	AWSBucketKey                  string
	AWSBucket                     string
	AWSKeyKey                     string
	AWSKey                        string
	AWSIDKey                      string
	AWSID                         string
	ArticleKeyTypeSuffixKey       string
	ArticleKeyTypeSuffix          string
	TopicsKey                     string
	Topics                        string
	TopicArticlesKey              string
	TopicArticles                 string
	TracingGrpcHostKey            string
	TracingGrpcHost               string
	TracingGrpcPortKey            string
	TracingGrpcPort               string
	TracingSamplingRateKey        string
	TracingSamplingRate           float64
	ServiceNameKey                string
	ServiceName                   string
	PrometheusPort                int
	PrometheusPortKey             string
	KafkaHealthCheckIntervalMsKey string
	KafkaHealthCheckIntervalMs    int
	KafkaHealthCheckLagMsKey      string
	KafkaHealthCheckLagMs         int
	KafkaHealthCheckPortKey       string
	KafkaHealthCheckPort          int
	UseHashedPrefixesKey          string
	UseHashedPrefixes             bool
	KafkaAutoOffsetResetKey       string
	KafkaAutoOffsetReset          string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.kafkaConsumerGroupIDKey = "KAFKA_CONSUMER_GROUP_ID"
	s.kafkaConsumerGroupID = "unique"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "localhost:2020"
	s.schemaRegistryCredsKey = "SCHEMA_REGISTRY_CREDS"
	s.schemaRegistryCreds = `{"username":"reg_admin","password":"654321"}`
	s.topicArticlesKey = "TOPIC_ARTICLES"
	s.topicArticles = ""
	s.AWSURLKey = "AWS_URL"
	s.AWSURL = "http://host.com:9090"
	s.AWSRegionKey = "AWS_REGION"
	s.AWSRegion = "us-east-2"
	s.AWSBucketKey = "AWS_BUCKET"
	s.AWSBucket = "diff_bucket"
	s.AWSKeyKey = "AWS_KEY"
	s.AWSKey = "diff_key"
	s.AWSIDKey = "AWS_ID"
	s.AWSID = "diff_id"
	s.ArticleKeyTypeSuffixKey = "KEY_TYPE_SUFFIX"
	s.ArticleKeyTypeSuffix = "v2"
	s.TracingGrpcHostKey = "OTEL_COLLECTOR_ADDR"
	s.TracingGrpcHost = "collector"
	s.TracingGrpcPortKey = "TRACING_GRPC_PORT"
	s.TracingGrpcPort = "4317"
	s.TracingSamplingRateKey = "TRACING_SAMPLING_RATE"
	s.TracingSamplingRate = 0.1
	s.ServiceNameKey = "SERVICE_NAME"
	s.ServiceName = "on-demand.service"
	s.PrometheusPortKey = "PROMETHEUS_PORT"
	s.PrometheusPort = 12411
	s.TopicsKey = "TOPICS"
	s.Topics = `{"location":"local","service_name":"structured-data","namespace":"articles","version":["v1"]}`
	s.KafkaHealthCheckIntervalMsKey = "KAFKA_HEALTH_CHECK_INTERVAL_MS"
	s.KafkaHealthCheckIntervalMs = 5000
	s.KafkaHealthCheckLagMsKey = "KAFKA_HEALTH_CHECK_LAG_NMS"
	s.KafkaHealthCheckLagMs = 500
	s.KafkaHealthCheckPortKey = "HEALTH_CHECK_PORT"
	s.KafkaHealthCheckPort = 9999
	s.UseHashedPrefixesKey = "USE_HASHED_PREFIXES"
	s.UseHashedPrefixes = false
	s.KafkaAutoOffsetResetKey = "KAFKA_AUTO_OFFSET_RESET"
	s.KafkaAutoOffsetReset = "earliest"
}
func (s *envTestSuite) TearDownSuite() {
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaConsumerGroupIDKey, s.kafkaConsumerGroupID)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.schemaRegistryCredsKey, s.schemaRegistryCreds)
	os.Setenv(s.topicArticlesKey, s.topicArticles)
	os.Setenv(s.AWSURLKey, s.AWSURL)
	os.Setenv(s.AWSRegionKey, s.AWSRegion)
	os.Setenv(s.AWSBucketKey, s.AWSBucket)
	os.Setenv(s.AWSKeyKey, s.AWSKey)
	os.Setenv(s.AWSIDKey, s.AWSID)
	os.Setenv(s.ArticleKeyTypeSuffixKey, s.ArticleKeyTypeSuffix)
	os.Setenv(s.TracingGrpcHostKey, s.TracingGrpcHost)
	os.Setenv(s.TracingSamplingRateKey, strconv.FormatFloat(s.TracingSamplingRate, 'f', -1, 64))
	os.Setenv(s.TracingGrpcPortKey, s.TracingGrpcPort)
	os.Setenv(s.ServiceNameKey, s.ServiceName)
	os.Setenv(s.PrometheusPortKey, strconv.Itoa(s.PrometheusPort))
	os.Setenv(s.TopicsKey, s.Topics)
	os.Setenv(s.KafkaHealthCheckIntervalMsKey, strconv.Itoa(s.KafkaHealthCheckIntervalMs))
	os.Setenv(s.KafkaHealthCheckLagMsKey, strconv.Itoa(s.KafkaHealthCheckLagMs))
	os.Setenv(s.KafkaHealthCheckPortKey, strconv.Itoa(s.KafkaHealthCheckPort))
	os.Setenv(s.UseHashedPrefixesKey, strconv.FormatBool(s.UseHashedPrefixes))
	os.Setenv(s.KafkaAutoOffsetResetKey, s.KafkaAutoOffsetReset)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaConsumerGroupID, env.KafkaConsumerGroupID)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)
	s.Assert().Equal(s.AWSURL, env.AWSURL)
	s.Assert().Equal(s.AWSRegion, env.AWSRegion)
	s.Assert().Equal(s.AWSBucket, env.AWSBucket)
	s.Assert().Equal(s.AWSID, env.AWSID)
	s.Assert().Equal(s.AWSKey, env.AWSKey)
	s.Assert().Equal(s.TracingGrpcHost, env.TracingGrpcHost)
	s.Assert().Equal(s.TracingGrpcPort, env.TracingGrpcPort)
	s.Assert().Equal(s.TracingSamplingRate, env.TracingSamplingRate)
	s.Assert().Equal(s.ServiceName, env.ServiceName)
	s.Assert().Equal(s.PrometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.ArticleKeyTypeSuffix, env.ArticleKeyTypeSuffix)
	s.Assert().Equal(s.KafkaHealthCheckIntervalMs, env.KafkaHealthCheckIntervalMs)
	s.Assert().Equal(s.KafkaHealthCheckLagMs, env.KafkaHealthCheckLagNMS)
	s.Assert().Equal(s.KafkaHealthCheckPort, env.HealthCheckPort)
	s.Assert().Equal(s.UseHashedPrefixes, env.UseHashedPrefixes)
	s.Assert().Equal(s.KafkaAutoOffsetReset, env.KafkaAutoOffsetReset)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	creds, err = json.Marshal(env.SchemaRegistryCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.schemaRegistryCreds, string(creds))

	cfg, err := json.Marshal(env.Topics)
	s.Assert().NoError(err)
	s.Assert().Equal(s.Topics, string(cfg))

}

func (s *envTestSuite) TestMainCodeTopicHandling() {
	tests := []struct {
		name           string
		config         *env.Environment
		expected       []string
		notExpected    []string
		shouldCallFunc bool
		errMsg         string
	}{
		{
			name: "Topics is nil",
			config: &env.Environment{
				Topics:        nil,
				TopicArticles: []string{},
			},
			expected: []string{
				"aws.structured-contents.enwiki-articles-compacted.v1",
				"aws.structured-contents.itwiki-articles-compacted.v1",
				"aws.structured-contents.itwiki-files-compacted.v1",
				"aws.structured-contents.itwiki-templates-compacted.v1",
				"aws.structured-contents.itwiki-categories-compacted.v1",
				"aws.structured-contents.dewiki-articles-compacted.v1",
			},
			notExpected: []string{
				"aws.structured-data.dewiki-articles-compacted.v1",
			},
			shouldCallFunc: false,
			errMsg:         "",
		},
		{
			name: "Topics is nil",
			config: &env.Environment{
				Topics: nil,
				TopicArticles: []string{
					"aws.structured-contents.enwiki-articles-compacted.v1",
					"aws.structured-contents.itwiki-articles-compacted.v1",
				},
			},
			expected: []string{
				"aws.structured-contents.enwiki-articles-compacted.v1",
				"aws.structured-contents.itwiki-articles-compacted.v1",
			},
			notExpected: []string{
				"aws.structured-contents.frwiki-articles-compacted.v1",
			},
			shouldCallFunc: false,
			errMsg:         "",
		},
		{
			name: "Topics is not nil and TopicArticles is empty",
			config: &env.Environment{
				Topics: &config.TopicArticleConfig{
					Location:  "aws",
					Service:   "structured-contents",
					Namespace: "articles",
					Version:   []string{"v1"},
				},
				TopicArticles: []string{},
			},
			expected: []string{
				"aws.structured-contents.xhwiki-articles-compacted.v1",
				"aws.structured-contents.biwiki-articles-compacted.v1",
			},

			shouldCallFunc: true,
			errMsg:         "",
		},
		{
			name: "Topics is not nil and TopicArticles has empty string",
			config: &env.Environment{
				Topics: &config.TopicArticleConfig{
					Location:  "aws",
					Service:   "structured-contents",
					Namespace: "articles",
					Version:   []string{"v1"},
				},
				TopicArticles: []string{""},
			},
			expected: []string{
				"aws.structured-contents.enwiki-articles-compacted.v1",
				"aws.structured-contents.itwiki-articles-compacted.v1",
			},
			notExpected: []string{
				"aws.structured-contents.enwiki-files-compacted.v1",
				"aws.structured-contents.enwiki-categories-compacted.v1",
				"aws.structured-contents.enwiki-templates-compacted.v1",
			},
			shouldCallFunc: true,
			errMsg:         "",
		},
		{
			name: "Topics is not nil and TopicArticles is empty",
			config: &env.Environment{
				Topics: &config.TopicArticleConfig{
					Location:  "aws",
					Service:   "structured-contents",
					Namespace: "articles",
					Version:   []string{"v1"},
				},
				TopicArticles: []string{},
			},
			expected: []string{
				"aws.structured-contents.enwiki-articles-compacted.v1",
				"aws.structured-contents.itwiki-articles-compacted.v1",
			},
			notExpected: []string{
				"aws.structured-contents.enwiki-files-compacted.v1",
				"aws.structured-contents.enwiki-categories-compacted.v1",
				"aws.structured-contents.enwiki-templates-compacted.v1",
			},
			shouldCallFunc: true,
			errMsg:         "",
		},
		{
			name: "Topics is not nil and TopicArticles is empty",
			config: &env.Environment{
				Topics: &config.TopicArticleConfig{
					Location:  "aws",
					Service:   "structured-contents",
					Namespace: "articles",
					Version:   []string{"v1"},
				},
				TopicArticles: nil,
			},
			expected: []string{
				"aws.structured-contents.enwiki-articles-compacted.v1",
				"aws.structured-contents.itwiki-articles-compacted.v1",
			},
			notExpected: []string{
				"aws.structured-contents.enwiki-files-compacted.v1",
				"aws.structured-contents.enwiki-categories-compacted.v1",
				"aws.structured-contents.enwiki-templates-compacted.v1",
			},
			shouldCallFunc: true,
			errMsg:         "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c, err := config.New()
			s.Assert().NoError(err)

			if tt.config.Topics != nil {
				topics, err := c.GetArticleTopics(tt.config.Topics)
				if tt.shouldCallFunc {
					s.Assert().NoError(err)
					for _, topic := range tt.expected {
						s.Assert().Contains(topics, topic)
					}

					for _, topic := range tt.notExpected {
						s.Assert().NotContains(topics, topic)
					}
				}
			}
		})
	}
}

func TestEnvTestSuite(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

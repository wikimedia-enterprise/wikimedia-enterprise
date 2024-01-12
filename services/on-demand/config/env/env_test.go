package env_test

import (
	"encoding/json"
	"os"
	"testing"
	"wikimedia-enterprise/services/on-demand/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey string
	kafkaBootstrapServers    string
	kafkaConsumerGroupID     string
	kafkaConsumerGroupIDKey  string
	kafkaCreds               string
	kafkaCredsKey            string
	schemaRegistryURLKey     string
	schemaRegistryURL        string
	schemaRegistryCredsKey   string
	schemaRegistryCreds      string
	topicArticlesKey         string
	topicArticles            string
	AWSURLKey                string
	AWSURL                   string
	AWSRegionKey             string
	AWSRegion                string
	AWSBucketKey             string
	AWSBucket                string
	AWSKeyKey                string
	AWSKey                   string
	AWSIDKey                 string
	AWSID                    string
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
	s.topicArticles = "local.structured-data.articles.v1"
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
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.kafkaConsumerGroupID, env.KafkaConsumerGroupID)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)
	s.Assert().Equal(s.topicArticles, env.TopicArticles[0])
	s.Assert().Equal(s.AWSURL, env.AWSURL)
	s.Assert().Equal(s.AWSRegion, env.AWSRegion)
	s.Assert().Equal(s.AWSBucket, env.AWSBucket)
	s.Assert().Equal(s.AWSKey, env.AWSKey)
	s.Assert().Equal(s.AWSID, env.AWSID)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	creds, err = json.Marshal(env.SchemaRegistryCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.schemaRegistryCreds, string(creds))
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}
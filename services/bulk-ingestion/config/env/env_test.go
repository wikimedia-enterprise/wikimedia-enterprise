package env_test

import (
	"encoding/json"
	"os"
	"testing"
	"wikimedia-enterprise/services/bulk-ingestion/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	kafkaBootstrapServersKey string
	kafkaBootstrapServers    string
	kafkaCredsKey            string
	kafkaCreds               string
	serverPortKey            string
	serverPort               string
	topicProjectsKey         string
	topicProjects            string
	topicLanguagesKey        string
	topicLanguages           string
	topicNamespaces          string
	topicNamespacesKey       string
	schemaRegistryURL        string
	schemaRegistryCreds      string
	schemaRegistryCredsKey   string
	schemaRegistryURLKey     string
	numberOfArticles         int
	numberOfArticlesKey      string
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
	defaultProjectKey        string
	defaultProject           string
	defaultURLKey            string
	defaultURL               string
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.topicProjectsKey = "TOPIC_PROJECTS"
	s.topicProjects = "local.bulk-ingestion.projects.v1"
	s.topicLanguagesKey = "TOPIC_LANGUAGES"
	s.topicLanguages = "local.bulk-ingestion.languages.v1"
	s.topicNamespacesKey = "TOPIC_NAMESPACES"
	s.topicNamespaces = "local.bulk-ingestion.namespace.v1"
	s.kafkaCredsKey = "KAFKA_CREDS"
	s.kafkaCreds = `{"username":"admin","password":"123456"}`
	s.schemaRegistryURLKey = "SCHEMA_REGISTRY_URL"
	s.schemaRegistryURL = "localhost:2020"
	s.schemaRegistryCredsKey = "SCHEMA_REGISTRY_CREDS"
	s.schemaRegistryCreds = `{"username":"reg_admin","password":"654321"}`
	s.numberOfArticlesKey = "NUMBER_OF_ARTICLES"
	s.numberOfArticles = 10
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
	s.defaultProjectKey = "DEFAULT_PROJECT"
	s.defaultProject = "enwiki"
	s.defaultURLKey = "DEFAULT_URL"
	s.defaultURL = "https://en.wikipedia.org"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.topicProjectsKey, s.topicProjects)
	os.Setenv(s.topicLanguagesKey, s.topicLanguages)
	os.Setenv(s.topicNamespacesKey, s.topicNamespaces)
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.schemaRegistryCredsKey, s.schemaRegistryCreds)
	os.Setenv(s.AWSURLKey, s.AWSURL)
	os.Setenv(s.AWSRegionKey, s.AWSRegion)
	os.Setenv(s.AWSBucketKey, s.AWSBucket)
	os.Setenv(s.AWSKeyKey, s.AWSKey)
	os.Setenv(s.AWSIDKey, s.AWSID)
	os.Setenv(s.defaultProjectKey, s.defaultProject)
	os.Setenv(s.defaultURLKey, s.defaultURL)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.topicProjects, env.TopicProjects)
	s.Assert().Equal(s.topicLanguages, env.TopicLanguages)
	s.Assert().Equal(s.topicNamespaces, env.TopicNamespaces)
	s.Assert().Equal(s.schemaRegistryURL, env.SchemaRegistryURL)
	s.Assert().Equal(s.numberOfArticles, env.NumberOfArticles)

	creds, err := json.Marshal(env.KafkaCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.kafkaCreds, string(creds))

	creds, err = json.Marshal(env.SchemaRegistryCreds)
	s.Assert().NoError(err)
	s.Assert().Equal(s.schemaRegistryCreds, string(creds))

	s.Assert().Equal(s.AWSURL, env.AWSURL)
	s.Assert().Equal(s.AWSRegion, env.AWSRegion)
	s.Assert().Equal(s.AWSBucket, env.AWSBucket)
	s.Assert().Equal(s.AWSID, env.AWSID)
	s.Assert().Equal(s.AWSKey, env.AWSKey)
	s.Assert().Equal(s.defaultProject, env.DefaultProject)
	s.Assert().Equal(s.defaultURL, env.DefaultURL)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

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
	mediawikiClientURLKey    string
	mediawikiClientURL       string
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
}

func (s *envTestSuite) SetupSuite() {
	s.kafkaBootstrapServersKey = "KAFKA_BOOTSTRAP_SERVERS"
	s.kafkaBootstrapServers = "localhost:8001"
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.mediawikiClientURLKey = "MEDIAWIKI_CLIENT_URL"
	s.mediawikiClientURL = "https://en.wikipedia.org/"
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
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.kafkaBootstrapServersKey, s.kafkaBootstrapServers)
	os.Setenv(s.kafkaCredsKey, s.kafkaCreds)
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.mediawikiClientURLKey, s.mediawikiClientURL)
	os.Setenv(s.topicProjectsKey, s.topicProjects)
	os.Setenv(s.topicLanguagesKey, s.topicLanguages)
	os.Setenv(s.topicNamespacesKey, s.topicNamespaces)
	os.Setenv(s.schemaRegistryURLKey, s.schemaRegistryURL)
	os.Setenv(s.schemaRegistryCredsKey, s.schemaRegistryCreds)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()
	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.kafkaBootstrapServers, env.KafkaBootstrapServers)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.mediawikiClientURL, env.MediawikiClientURL)
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
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

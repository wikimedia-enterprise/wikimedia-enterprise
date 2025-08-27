package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
	cfg API
}

func (s *configTestSuite) SetupSuite() {
	cfg, err := New()
	s.Assert().NoError(err)

	s.cfg = cfg
}

func (s *configTestSuite) TestGetProjects() {
	s.Assert().NotZero(len(s.cfg.GetProjects()))
}

func (s *configTestSuite) TestGetStructuredProjects() {
	s.Assert().NotZero(len(s.cfg.GetStructuredProjects()))
}

func (s *configTestSuite) TestGetLanguage() {
	s.Assert().Equal("en", s.cfg.GetLanguage("enwiki"))
}

func (s *configTestSuite) TestGetNamespaces() {
	s.Assert().NotEmpty(s.cfg.GetNamespaces())
}

func (s *configTestSuite) TestGetNamespacesMetadata() {
	var namespacesMetadata = s.cfg.GetNamespacesMetadata()
	for _, namespace := range s.cfg.GetNamespaces() {
		s.Assert().Contains(namespacesMetadata, namespace)
		s.Assert().NotEmpty(namespacesMetadata[namespace].Description)
	}
}

func (s *configTestSuite) TestGetPartitions() {
	s.Assert().Equal([]int{0}, s.cfg.GetPartitions("afwikibooks", 0))
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}
func (s *configTestSuite) TestGetArticleTopics() {
	// Define test cases
	tests := []struct {
		name     string
		config   TopicArticleConfig
		expected []string
		errMsg   string
	}{
		{
			name: "Valid pattern with 1 version",
			config: TopicArticleConfig{
				Location: "aws",
				Service:  "structured-data",
				Version:  []string{"v1"},
			},
			expected: []string{
				"aws.structured-data.enwiki-templates-compacted.v1",
				"aws.structured-data.enwiki-categories-compacted.v1",
				"aws.structured-data.enwiki-articles-compacted.v1",
				"aws.structured-data.enwiki-files-compacted.v1",
				"aws.structured-data.frwiki-templates-compacted.v1",
				"aws.structured-data.frwiki-categories-compacted.v1",
				"aws.structured-data.frwiki-articles-compacted.v1",
				"aws.structured-data.frwiki-files-compacted.v1",
			},
			errMsg: "",
		},
		{
			name: "Valid pattern with many versions",
			config: TopicArticleConfig{
				Location: "aws",
				Service:  "structured-data",
				Version:  []string{"v1", "v2", "v4"},
			},
			expected: []string{
				"aws.structured-data.enwiki-templates-compacted.v1",
				"aws.structured-data.enwiki-categories-compacted.v1",
				"aws.structured-data.enwiki-articles-compacted.v1",
				"aws.structured-data.enwiki-files-compacted.v1",
				"aws.structured-data.frwiki-templates-compacted.v1",
				"aws.structured-data.frwiki-categories-compacted.v1",
				"aws.structured-data.frwiki-articles-compacted.v1",
				"aws.structured-data.frwiki-files-compacted.v1",
				"aws.structured-data.enwiki-templates-compacted.v2",
				"aws.structured-data.enwiki-categories-compacted.v2",
				"aws.structured-data.enwiki-articles-compacted.v2",
				"aws.structured-data.enwiki-files-compacted.v2",
				"aws.structured-data.frwiki-templates-compacted.v2",
				"aws.structured-data.frwiki-categories-compacted.v2",
				"aws.structured-data.frwiki-articles-compacted.v2",
				"aws.structured-data.frwiki-files-compacted.v2",
				"aws.structured-data.enwiki-templates-compacted.v4",
				"aws.structured-data.enwiki-categories-compacted.v4",
				"aws.structured-data.enwiki-articles-compacted.v4",
				"aws.structured-data.enwiki-files-compacted.v4",
				"aws.structured-data.frwiki-templates-compacted.v4",
				"aws.structured-data.frwiki-categories-compacted.v4",
				"aws.structured-data.frwiki-articles-compacted.v4",
				"aws.structured-data.frwiki-files-compacted.v4",
			},
			errMsg: "",
		},
		{
			name: "Valid TopicArticles with 1 version and 1 namespace",
			config: TopicArticleConfig{
				Location:  "aws",
				Service:   "structured-data",
				Namespace: "articles",
				Version:   []string{"v1"},
			},
			expected: []string{
				"aws.structured-data.enwiki-articles-compacted.v1",
				"aws.structured-data.frwiki-articles-compacted.v1",
			},
			errMsg: "",
		},
		{
			name: "TopicArticles with no service",
			config: TopicArticleConfig{
				Location:  "aws",
				Service:   "",
				Namespace: "articles",
				Version:   []string{"v1"},
			},
			errMsg: "service is required",
		},
		{
			name: "TopicArticles with no location",
			config: TopicArticleConfig{
				Service:   "structured-data",
				Location:  "",
				Namespace: "articles",
				Version:   []string{"v1"},
			},
			errMsg: "location is required",
		},
		{
			name:     "Invalid TopicArticles",
			config:   TopicArticleConfig{},
			expected: nil,
			errMsg:   "location is required",
		},
	}

	// Run test cases
	for _, tt := range tests {
		s.Run(tt.name, func() {

			topics, err := s.cfg.GetArticleTopics(&tt.config)

			// Validate error scenarios
			if tt.errMsg != "" {
				s.Assert().Error(err)
				s.Assert().Contains(err.Error(), tt.errMsg)
				s.Assert().Nil(topics)
			} else {
				// Validate successful execution
				s.Assert().NoError(err)
				s.Assert().NotNil(topics)
				s.Assert().Greater(len(topics), 0)

				// Validate the expected topics
				for _, expected := range tt.expected {
					s.Assert().Contains(topics, expected)
				}
			}
		})
	}
}

type Environment struct {
	TopicArticleConfig *TopicArticleConfig
}

func (s *configTestSuite) TestEnvFileLoadingAndUnmarshalling() {
	// Simulate loading the `.env` file configuration
	envConfig := &Environment{
		TopicArticleConfig: &TopicArticleConfig{
			Location:  "aws",
			Service:   "structured-data",
			Version:   []string{"v1"},
			Namespace: "articles",
		},
	}

	// Marshal the loaded TopicArticleConfig to JSON
	tac, err := json.Marshal(envConfig.TopicArticleConfig)
	s.Assert().NoError(err)

	// Unmarshal the JSON back into a TopicArticleConfig instance
	var cfg TopicArticleConfig
	err = cfg.UnmarshalEnvironmentValue(string(tac))
	s.Assert().NoError(err)

	// Define the expected configuration
	expectedCfg := TopicArticleConfig{
		Location:  "aws",
		Service:   "structured-data",
		Version:   []string{"v1"},
		Namespace: "articles",
	}

	// Validate that the unmarshalled configuration matches the expected configuration
	s.Assert().Equal(expectedCfg, cfg)
}

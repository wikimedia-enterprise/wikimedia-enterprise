package env_test

import (
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/api/main/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	serverPortKey           string
	serverPort              string
	serverModeKey           string
	serverMode              string
	AWSRegionKey            string
	AWSRegion               string
	AWSIDKey                string
	AWSID                   string
	AWSKeyKey               string
	AWSKey                  string
	AWSURLKey               string
	AWSURL                  string
	AWSBucketKey            string
	AWSBucket               string
	CognitoClientIdKey      string
	CognitoClientId         string
	CognitoClientSecretKey  string
	CognitoClientSecret     string
	RedisAddr               string
	RedisAddrKey            string
	RedisPassword           string
	RedisPasswordKey        string
	AccessModel             string
	AccessModelKey          string
	AccessPolicy            string
	AccessPolicyKey         string
	IPAllowList             string
	IPAllowListKey          string
	FreeTierGroup           string
	FreeTierGroupKey        string
	PrometheusPortKey       string
	PrometheusPort          int
	DescriptionEnabledKey   string
	DescriptionEnabled      bool
	SectionsEnabledKey      string
	SectionsEnabled         bool
	ArticleKeyTypeSuffix    string
	ArticleKeyTypeSuffixKey string
}

func (s *envTestSuite) SetupSuite() {
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.serverModeKey = "SERVER_MODE"
	s.serverMode = "debug"
	s.AWSRegionKey = "AWS_REGION"
	s.AWSRegion = "us-east-1"
	s.AWSIDKey = "AWS_ID"
	s.AWSID = "some_AWS_id"
	s.AWSKeyKey = "AWS_KEY"
	s.AWSKey = "some_AWS_key"
	s.AWSBucketKey = "AWS_BUCKET"
	s.AWSBucket = "wme-data-bk"
	s.AWSURLKey = "AWS_URL"
	s.AWSURL = "http://minio:9000"
	s.CognitoClientId = "clientId"
	s.CognitoClientIdKey = "COGNITO_CLIENT_ID"
	s.CognitoClientSecret = "secret"
	s.CognitoClientSecretKey = "COGNITO_CLIENT_SECRET"
	s.RedisAddr = "addr"
	s.RedisAddrKey = "REDIS_ADDR"
	s.RedisPassword = "pass"
	s.RedisPasswordKey = "REDIS_PASSWORD"
	s.AccessModelKey = "ACCESS_MODEL"
	s.AccessModel = "model content"
	s.AccessPolicyKey = "ACCESS_POLICY"
	s.AccessPolicy = "policy content"
	s.IPAllowListKey = "IP_ALLOW_LIST"
	s.IPAllowList = "[{\"ip_range\": {\"start\": \"192.168.0.1\", \"end\": \"192.168.0.10\"}, \"user\": {\"username\": \"user1\", \"groups\": [\"group_1\"]}}, {\"ip_range\": {\"start\": \"192.168.1.1\", \"end\": \"192.168.1.10\"}, \"user\": {\"username\": \"user2\", \"groups\": [\"group_2\"]}}]"
	s.FreeTierGroupKey = "FREE_TIER_GROUP"
	s.FreeTierGroup = "group_1"
	s.PrometheusPortKey = "PROMETHEUS_PORT"
	s.PrometheusPort = 101
	s.DescriptionEnabledKey = "DESCRIPTION_ENABLED"
	s.DescriptionEnabled = true
	s.SectionsEnabledKey = "SECTIONS_ENABLED"
	s.SectionsEnabled = true
	s.ArticleKeyTypeSuffixKey = "KEY_TYPE_SUFFIX"
	s.ArticleKeyTypeSuffix = "v2"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.serverModeKey, s.serverMode)
	os.Setenv(s.AWSRegionKey, s.AWSRegion)
	os.Setenv(s.AWSIDKey, s.AWSID)
	os.Setenv(s.AWSKeyKey, s.AWSKey)
	os.Setenv(s.AWSURLKey, s.AWSURL)
	os.Setenv(s.AWSBucketKey, s.AWSBucket)
	os.Setenv(s.CognitoClientIdKey, s.CognitoClientId)
	os.Setenv(s.CognitoClientSecretKey, s.CognitoClientSecret)
	os.Setenv(s.RedisAddrKey, s.RedisAddr)
	os.Setenv(s.RedisPasswordKey, s.RedisPassword)
	os.Setenv(s.AccessModelKey, s.AccessModel)
	os.Setenv(s.AccessPolicyKey, s.AccessPolicy)
	os.Setenv(s.IPAllowListKey, s.IPAllowList)
	os.Setenv(s.FreeTierGroupKey, s.FreeTierGroup)
	os.Setenv(s.PrometheusPortKey, strconv.Itoa(s.PrometheusPort))
	os.Setenv(s.ArticleKeyTypeSuffixKey, s.ArticleKeyTypeSuffix)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()

	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.serverMode, env.ServerMode)
	s.Assert().Equal(s.AWSRegion, env.AWSRegion)
	s.Assert().Equal(s.AWSID, env.AWSID)
	s.Assert().Equal(s.AWSKey, env.AWSKey)
	s.Assert().Equal(s.AWSBucket, env.AWSBucket)
	s.Assert().Equal(s.AWSURL, env.AWSURL)
	s.Assert().Equal(s.CognitoClientId, env.CognitoClientID)
	s.Assert().Equal(s.CognitoClientSecret, env.CognitoClientSecret)
	s.Assert().Equal(s.RedisAddr, env.RedisAddr)
	s.Assert().Equal(s.RedisPassword, env.RedisPassword)
	s.Assert().Equal(s.FreeTierGroup, env.FreeTierGroup)
	s.Assert().Equal(s.PrometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.DescriptionEnabled, env.DescriptionEnabled)
	s.Assert().Equal(s.SectionsEnabled, env.SectionsEnabled)
	s.Assert().NotNil(env.AccessModel.Path)
	s.Assert().NotNil(env.AccessPolicy.Path)
	s.Assert().NotNil(env.IPAllowList)
	s.Assert().Equal(s.ArticleKeyTypeSuffix, env.ArticleKeyTypeSuffix)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

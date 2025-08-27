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
	serverPortKey                      string
	serverPort                         string
	serverModeKey                      string
	serverMode                         string
	AWSRegionKey                       string
	AWSRegion                          string
	AWSIDKey                           string
	AWSID                              string
	AWSKeyKey                          string
	AWSKey                             string
	AWSURLKey                          string
	AWSURL                             string
	AWSBucketKey                       string
	AWSBucket                          string
	AWSBucketCommonsKey                string
	AWSBucketCommons                   string
	CognitoClientIdKey                 string
	CognitoClientId                    string
	CognitoClientSecretKey             string
	CognitoClientSecret                string
	RedisAddr                          string
	RedisAddrKey                       string
	RedisPassword                      string
	RedisPasswordKey                   string
	AccessModel                        string
	AccessModelKey                     string
	AccessPolicy                       string
	AccessPolicyKey                    string
	CapConfig                          string
	CapConfigKey                       string
	IPAllowList                        string
	IPAllowListKey                     string
	FreeTierGroup                      string
	FreeTierGroupKey                   string
	PrometheusPortKey                  string
	PrometheusPort                     int
	MetricsReadTimeOutSecondsKey       string
	MetricsReadTimeOutSeconds          int
	MetricsReadHeaderTimeOutSecondsKey string
	MetricsReadHeaderTimeOutSeconds    int
	MetricsWriteTimeoutSecondsKey      string
	MetricsWriteTimeoutSeconds         int
	DescriptionEnabledKey              string
	DescriptionEnabled                 bool
	SectionsEnabledKey                 string
	SectionsEnabled                    bool
	ArticleKeyTypeSuffix               string
	ArticleKeyTypeSuffixKey            string
	HealthChecksTimeoutMsKey           string
	HealthChecksTimeoutMs              int
	HealthCheckMaxRetries              int
	HealthCheckMaxRetriesKey           string
	EnableAttribution                  bool
	EnableAttributionKey               string
	UseHashedPrefixesKey               string
	UseHashedPrefixes                  bool
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
	s.AWSBucket = "wme-primary-bk"
	s.AWSBucketCommonsKey = "AWS_BUCKET_COMMONS"
	s.AWSBucketCommons = "wme-data-bk"
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
	s.CapConfigKey = "CAP_CONFIGURATION"
	s.CapConfig = "[{\"groups\":[\"test_group\"],\"limit\":100,\"products\":[\"test_product\"],\"prefix_group\":\"cap:test\"}]"
	s.IPAllowListKey = "IP_ALLOW_LIST"
	s.IPAllowList = "[{\"ip_range\": {\"start\": \"192.168.0.1\", \"end\": \"192.168.0.10\"}, \"user\": {\"username\": \"user1\", \"groups\": [\"group_1\"]}}, {\"ip_range\": {\"start\": \"192.168.1.1\", \"end\": \"192.168.1.10\"}, \"user\": {\"username\": \"user2\", \"groups\": [\"group_2\"]}}]"
	s.FreeTierGroupKey = "FREE_TIER_GROUP"
	s.FreeTierGroup = "group_1"
	s.PrometheusPortKey = "PROMETHEUS_PORT"
	s.PrometheusPort = 101
	s.MetricsReadTimeOutSecondsKey = "METRICS_READ_TIMEOUT_SECONDS"
	s.MetricsReadTimeOutSeconds = 15
	s.MetricsReadHeaderTimeOutSecondsKey = "METRICS_READ_HEADER_TIMEOUT_SECONDS"
	s.MetricsReadHeaderTimeOutSeconds = 20
	s.MetricsWriteTimeoutSecondsKey = "METRICS_WRITE_TIMEOUT_SECONDS"
	s.MetricsWriteTimeoutSeconds = 25
	s.DescriptionEnabledKey = "DESCRIPTION_ENABLED"
	s.DescriptionEnabled = true
	s.SectionsEnabledKey = "SECTIONS_ENABLED"
	s.SectionsEnabled = true
	s.ArticleKeyTypeSuffixKey = "KEY_TYPE_SUFFIX"
	s.ArticleKeyTypeSuffix = "v2"
	s.HealthChecksTimeoutMsKey = "HEALTH_CHECKS_TIMEOUT_MS"
	s.HealthChecksTimeoutMs = 5000
	s.HealthCheckMaxRetries = 3
	s.HealthCheckMaxRetriesKey = "HEALTH_CHECK_MAX_RETRIES"
	s.EnableAttribution = false
	s.EnableAttributionKey = "ENABLE_ATTRIBUTION"
	s.UseHashedPrefixesKey = "USE_HASHED_PREFIXES"
	s.UseHashedPrefixes = false
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.serverModeKey, s.serverMode)
	os.Setenv(s.AWSRegionKey, s.AWSRegion)
	os.Setenv(s.AWSIDKey, s.AWSID)
	os.Setenv(s.AWSKeyKey, s.AWSKey)
	os.Setenv(s.AWSURLKey, s.AWSURL)
	os.Setenv(s.AWSBucketKey, s.AWSBucket)
	os.Setenv(s.AWSBucketCommonsKey, s.AWSBucketCommons)
	os.Setenv(s.CognitoClientIdKey, s.CognitoClientId)
	os.Setenv(s.CognitoClientSecretKey, s.CognitoClientSecret)
	os.Setenv(s.RedisAddrKey, s.RedisAddr)
	os.Setenv(s.RedisPasswordKey, s.RedisPassword)
	os.Setenv(s.AccessModelKey, s.AccessModel)
	os.Setenv(s.AccessPolicyKey, s.AccessPolicy)
	os.Setenv(s.CapConfigKey, s.CapConfig)
	os.Setenv(s.IPAllowListKey, s.IPAllowList)
	os.Setenv(s.FreeTierGroupKey, s.FreeTierGroup)
	os.Setenv(s.PrometheusPortKey, strconv.Itoa(s.PrometheusPort))
	os.Setenv(s.MetricsReadTimeOutSecondsKey, strconv.Itoa(s.MetricsReadTimeOutSeconds))
	os.Setenv(s.MetricsReadHeaderTimeOutSecondsKey, strconv.Itoa(s.MetricsReadHeaderTimeOutSeconds))
	os.Setenv(s.MetricsWriteTimeoutSecondsKey, strconv.Itoa(s.MetricsWriteTimeoutSeconds))
	os.Setenv(s.ArticleKeyTypeSuffixKey, s.ArticleKeyTypeSuffix)
	os.Setenv(s.HealthChecksTimeoutMsKey, strconv.Itoa(s.HealthChecksTimeoutMs))
	os.Setenv(s.HealthCheckMaxRetriesKey, strconv.Itoa(s.HealthCheckMaxRetries))
	os.Setenv(s.EnableAttributionKey, strconv.FormatBool(s.EnableAttribution))
	os.Setenv(s.UseHashedPrefixesKey, strconv.FormatBool(s.UseHashedPrefixes))
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
	s.Assert().Equal(s.AWSBucketCommons, env.AWSBucketCommons)
	s.Assert().Equal(s.AWSURL, env.AWSURL)
	s.Assert().Equal(s.CognitoClientId, env.CognitoClientID)
	s.Assert().Equal(s.CognitoClientSecret, env.CognitoClientSecret)
	s.Assert().Equal(s.RedisAddr, env.RedisAddr)
	s.Assert().Equal(s.RedisPassword, env.RedisPassword)
	s.Assert().Equal(s.FreeTierGroup, env.FreeTierGroup)
	s.Assert().Equal(s.PrometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.MetricsReadTimeOutSeconds, env.MetricsReadTimeOutSeconds)
	s.Assert().Equal(s.MetricsReadHeaderTimeOutSeconds, env.MetricsReadHeaderTimeOutSeconds)
	s.Assert().Equal(s.MetricsWriteTimeoutSeconds, env.MetricsWriteTimeoutSeconds)
	s.Assert().Equal(s.DescriptionEnabled, env.DescriptionEnabled)
	s.Assert().Equal(s.SectionsEnabled, env.SectionsEnabled)
	s.Assert().NotNil(env.CapConfig)
	s.Assert().NotNil(env.AccessModel.Path)
	s.Assert().NotNil(env.AccessPolicy.Path)
	s.Assert().NotNil(env.IPAllowList)
	s.Assert().Equal(s.ArticleKeyTypeSuffix, env.ArticleKeyTypeSuffix)
	s.Assert().Equal(s.HealthChecksTimeoutMs, env.HealthChecksTimeoutMs)
	s.Assert().Equal(s.EnableAttribution, env.EnableAttribution)
	s.Assert().Equal(s.UseHashedPrefixes, env.UseHashedPrefixes)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

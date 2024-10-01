package env_test

import (
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/api/auth/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	prometheusPortKey       string
	prometheusPort          int
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
	cognitoClientIDKey      string
	cognitoClientID         string
	cognitoSecretKey        string
	cognitoSecret           string
	cognitoUserPoolIDKey    string
	cognitoUserPoolID       string
	cognitoUserGroupKey     string
	cognitoUserGroup        string
	redisAddr               string
	redisAddrKey            string
	redisPassword           string
	redisPasswordKey        string
	maxAccessTokensKey      string
	maxAccessTokens         int64
	accessTokensExpHoursKey string
	accessTokensExpHours    int64
	cognitoCacheExpKey      string
	cognitoCacheExp         int
	ipRangeKey              string
	ipRange                 string
	ondemandLimit           string
	ondemandLimitKey        string
	snapshotLimit           string
	snapshotLimitKey        string
	accessPolicy            string
	accessPolicyKey         string
	domainDenyList          env.List
	domainDenyListKey       string
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
	s.cognitoClientIDKey = "COGNITO_CLIENT_ID"
	s.cognitoClientID = "some_client"
	s.cognitoSecretKey = "COGNITO_SECRET"
	s.cognitoSecret = "some_cognito_secret"
	s.cognitoUserPoolIDKey = "COGNITO_USER_POOL_ID"
	s.cognitoUserPoolID = "cognito_pool_id"
	s.cognitoUserGroupKey = "COGNITO_USER_GROUP"
	s.cognitoUserGroup = "free"
	s.redisAddrKey = "REDIS_ADDR"
	s.redisAddr = "some_addr:6379"
	s.redisPasswordKey = "REDIS_PASSWORD"
	s.redisPassword = "some_password"
	s.maxAccessTokensKey = "MAX_ACCESS_TOKENS"
	s.maxAccessTokens = 90
	s.accessTokensExpHoursKey = "ACCESS_TOKENS_EXPIRATION"
	s.accessTokensExpHours = 2160
	s.cognitoCacheExpKey = "COGNITO_CACHE_EXPIRATION"
	s.cognitoCacheExp = 300
	s.ipRangeKey = "IP_RANGE"
	s.ipRange = "127.0.0.0-127.0.0.1"
	s.ondemandLimitKey = "ONDEMAND_LIMIT"
	s.ondemandLimit = "5000"
	s.snapshotLimitKey = "SNAPSHOT_LIMIT"
	s.snapshotLimit = "15"
	s.accessPolicyKey = "ACCESS_POLICY"
	s.accessPolicy = ""
	s.prometheusPortKey = "PROMETHEUS_PORT"
	s.prometheusPort = 101
	s.domainDenyListKey = "DOMAIN_DENY_LIST"
	s.domainDenyList = env.List{"example.com", "example.org"}
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.serverModeKey, s.serverMode)
	os.Setenv(s.AWSRegionKey, s.AWSRegion)
	os.Setenv(s.AWSIDKey, s.AWSID)
	os.Setenv(s.AWSKeyKey, s.AWSKey)
	os.Setenv(s.cognitoClientIDKey, s.cognitoClientID)
	os.Setenv(s.cognitoSecretKey, s.cognitoSecret)
	os.Setenv(s.cognitoClientIDKey, s.cognitoClientID)
	os.Setenv(s.cognitoUserPoolIDKey, s.cognitoUserPoolID)
	os.Setenv(s.cognitoUserGroupKey, s.cognitoUserGroup)
	os.Setenv(s.redisAddrKey, s.redisAddr)
	os.Setenv(s.redisPasswordKey, s.redisPassword)
	os.Setenv(s.maxAccessTokensKey, strconv.FormatInt(s.maxAccessTokens, 10))
	os.Setenv(s.accessTokensExpHoursKey, strconv.FormatInt(s.accessTokensExpHours, 10))
	os.Setenv(s.cognitoCacheExpKey, strconv.Itoa(s.cognitoCacheExp))
	os.Setenv(s.ipRangeKey, s.ipRange)
	os.Setenv(s.ondemandLimitKey, s.ondemandLimit)
	os.Setenv(s.snapshotLimitKey, s.snapshotLimit)
	os.Setenv(s.accessPolicyKey, s.accessPolicy)
	os.Setenv(s.prometheusPortKey, strconv.Itoa(s.prometheusPort))
	os.Setenv(s.domainDenyListKey, "example.com,example.org")
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
	s.Assert().Equal(s.cognitoClientID, env.CognitoClientID)
	s.Assert().Equal(s.cognitoSecret, env.CognitoSecret)
	s.Assert().Equal(s.redisAddr, env.RedisAddr)
	s.Assert().Equal(s.redisPassword, env.RedisPassword)
	s.Assert().Equal(s.maxAccessTokens, env.MaxAccessTokens)
	s.Assert().Equal(s.accessTokensExpHours, env.AccessTokensExpHours)
	s.Assert().Equal(s.ondemandLimit, env.OndemandLimit)
	s.Assert().Equal(s.snapshotLimit, env.SnapshotLimit)
	s.Assert().Equal(s.prometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.domainDenyList, env.DomainDenyList)
	s.Assert().NotNil(env.AccessPolicy.Map)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

package env_test

import (
	"os"
	"strconv"
	"testing"
	"wikimedia-enterprise/api/realtime/config/env"

	"github.com/stretchr/testify/suite"
)

type envTestSuite struct {
	suite.Suite
	serverPortKey     string
	serverPort        string
	serverModeKey     string
	serverMode        string
	ksqlURLkey        string
	ksqlURL           string
	KSQLUsernameKey   string
	KSQLUsername      string
	KSQLPasswordKey   string
	KSQLPassword      string
	awsRegionKey      string
	awsRegion         string
	awsIDKey          string
	awsID             string
	awsKeyKey         string
	awsKey            string
	clientIDKey       string
	clientID          string
	clientSecretKey   string
	clientSecret      string
	redisAddrKey      string
	redisAddr         string
	redisPasswordKey  string
	redisPassword     string
	prometheusPortKey string
	prometheusPort    int
	partitionsKey     string
	partitions        int
	maxPartsKey       string
	maxParts          int
	articlesStreamKey string
	articlesStream    string
	IPAllowList       string
	IPAllowListKey    string
}

func (s *envTestSuite) SetupSuite() {
	s.serverPortKey = "SERVER_PORT"
	s.serverPort = "5052"
	s.serverModeKey = "SERVER_MODE"
	s.serverMode = "debug"
	s.ksqlURLkey = "KSQL_URL"
	s.ksqlURL = "localhost:2020"
	s.KSQLUsernameKey = "KSQL_USERNAME"
	s.KSQLUsername = "foo"
	s.KSQLPasswordKey = "KSQL_PASSWORD"
	s.KSQLPassword = "bar"
	s.awsRegionKey = "AWS_REGION"
	s.awsRegion = "region-1"
	s.awsIDKey = "AWS_ID"
	s.awsID = "foo"
	s.awsKeyKey = "AWS_KEY"
	s.awsKey = "bar"
	s.clientIDKey = "COGNITO_CLIENT_ID"
	s.clientID = "client_id"
	s.clientSecretKey = "COGNITO_CLIENT_SECRET"
	s.clientSecret = "secret"
	s.redisAddrKey = "REDIS_ADDR"
	s.redisAddr = "redis:2020"
	s.redisPasswordKey = "REDIS_PASSWORD"
	s.redisPassword = "secret"
	s.prometheusPortKey = "PROMETHEUS_PORT"
	s.prometheusPort = 101
	s.partitionsKey = "PARTITIONS"
	s.partitions = 100
	s.maxPartsKey = "MAX_PARTS"
	s.maxParts = 10
	s.articlesStreamKey = "ARTICLES_STREAM"
	s.articlesStream = "articles_str"
	s.IPAllowListKey = "IP_ALLOW_LIST"
	s.IPAllowList = "[{\"ip_range\": {\"start\": \"192.168.0.1\", \"end\": \"192.168.0.10\"}, \"user\": {\"username\": \"user1\", \"groups\": [\"group_1\"]}}, {\"ip_range\": {\"start\": \"192.168.1.1\", \"end\": \"192.168.1.10\"}, \"user\": {\"username\": \"user2\", \"groups\": [\"group_2\"]}}]"
}

func (s *envTestSuite) SetupTest() {
	os.Setenv(s.serverPortKey, s.serverPort)
	os.Setenv(s.serverModeKey, s.serverMode)
	os.Setenv(s.ksqlURLkey, s.ksqlURL)
	os.Setenv(s.KSQLUsernameKey, s.KSQLUsername)
	os.Setenv(s.KSQLPasswordKey, s.KSQLPassword)
	os.Setenv(s.awsRegionKey, s.awsRegion)
	os.Setenv(s.awsIDKey, s.awsID)
	os.Setenv(s.awsKeyKey, s.awsKey)
	os.Setenv(s.clientIDKey, s.clientID)
	os.Setenv(s.clientSecretKey, s.clientSecret)
	os.Setenv(s.redisAddrKey, s.redisAddr)
	os.Setenv(s.redisPasswordKey, s.redisPassword)
	os.Setenv(s.prometheusPortKey, strconv.Itoa(s.prometheusPort))
	os.Setenv(s.partitionsKey, strconv.Itoa(s.partitions))
	os.Setenv(s.maxPartsKey, strconv.Itoa(s.maxParts))
	os.Setenv(s.articlesStreamKey, s.articlesStream)
	os.Setenv(s.IPAllowListKey, s.IPAllowList)
}

func (s *envTestSuite) TestNew() {
	env, err := env.New()

	s.Assert().NoError(err)
	s.Assert().NotNil(env)
	s.Assert().Equal(s.serverPort, env.ServerPort)
	s.Assert().Equal(s.serverMode, env.ServerMode)
	s.Assert().Equal(s.ksqlURL, env.KSQLURL)
	s.Assert().Equal(s.KSQLUsername, env.KSQLUsername)
	s.Assert().Equal(s.KSQLPassword, env.KSQLPassword)
	s.Assert().Equal(s.awsRegion, env.AWSRegion)
	s.Assert().Equal(s.awsID, env.AWSID)
	s.Assert().Equal(s.awsKey, env.AWSKey)
	s.Assert().Equal(s.clientID, env.CognitoClientID)
	s.Assert().Equal(s.clientSecret, env.CognitoClientSecret)
	s.Assert().Equal(s.redisAddr, env.RedisAddr)
	s.Assert().Equal(s.redisPassword, env.RedisPassword)
	s.Assert().Equal(s.prometheusPort, env.PrometheusPort)
	s.Assert().Equal(s.partitions, env.Partitions)
	s.Assert().Equal(s.maxParts, env.MaxParts)
	s.Assert().Equal(s.articlesStream, env.ArticlesStream)
	s.Assert().NotNil(env.IPAllowList)
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envTestSuite))
}

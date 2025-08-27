// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/api/realtime/submodules/httputil"

	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

const defaultModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (g(r.sub, p.sub) || keyMatch(r.sub, p.sub)) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
`
const defaultPolicy = `p, realtime, /v2/articles, GET
p, realtime, /v2/articles, POST
g, group_3, realtime`

func unmarshalAccess(dta string, fnm string, cnt string) (string, error) {
	f, err := os.CreateTemp("/tmp", fnm)

	if err != nil {
		return "", fmt.Errorf("error creating tmp file %s: %w", fnm, err)
	}

	if len(dta) > 0 {
		_, err = f.Write([]byte(dta))
	} else {
		_, err = f.Write([]byte(cnt))
	}

	if err != nil {
		return "", fmt.Errorf("error writing to tmp file %s: %w", fnm, err)
	}

	return f.Name(), nil
}

// AccessModel structure wraps Path to the access model created file.
type AccessModel struct {
	Path string
}

// UnmarshalEnvironmentValue called by env package on initialization to create a temporary file with AccessModel for casbin middleware.
func (m *AccessModel) UnmarshalEnvironmentValue(data string) (err error) {
	m.Path, err = unmarshalAccess(data, "model.conf", defaultModel)
	return err
}

// AccessPolicy structure wraps Path to the access policy created file.
type AccessPolicy struct {
	Path string
}

// UnmarshalEnvironmentValue called by env package on initialization to create a temporary file with AccessPolicy for casbin middleware.
func (m *AccessPolicy) UnmarshalEnvironmentValue(data string) (err error) {
	m.Path, err = unmarshalAccess(data, "policy.csv", defaultPolicy)
	return err
}

// Credentials SASL/SSL credentials.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (kc *Credentials) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), kc)
}

type ConfigMap map[string]string

func (m *ConfigMap) UnmarshalEnvironmentValue(data string) error {
	return json.Unmarshal([]byte(data), m)
}

// Environment environment variables configuration.
type Environment struct {
	KafkaBootstrapServers string       `env:"KAFKA_BOOTSTRAP_SERVERS,required=true"`
	KafkaCreds            *Credentials `env:"KAFKA_CREDS"`
	KafkaExtraConfig      ConfigMap    `env:"KAFKA_EXTRA_CONFIG,default={}"`
	SchemaRegistryURL     string       `env:"SCHEMA_REGISTRY_URL,required=true"`
	SchemaRegistryCreds   *Credentials `env:"SCHEMA_REGISTRY_CREDS"`

	ServerMode          string               `env:"SERVER_MODE,default=release"`
	ServerPort          string               `env:"SERVER_PORT,default=4040"`
	AWSRegion           string               `env:"AWS_REGION"`
	AWSID               string               `env:"AWS_ID,default=foo"`
	AWSKey              string               `env:"AWS_KEY,default=bar"`
	CognitoClientID     string               `env:"COGNITO_CLIENT_ID,required=true"`
	CognitoClientSecret string               `env:"COGNITO_CLIENT_SECRET,required=true"`
	RedisAddr           string               `env:"REDIS_ADDR,required=true"`
	RedisPassword       string               `env:"REDIS_PASSWORD,required=true"`
	AccessModel         *AccessModel         `env:"ACCESS_MODEL,required=true"`
	AccessPolicy        *AccessPolicy        `env:"ACCESS_POLICY,required=true"`
	PrometheusPort      int                  `env:"PROMETHEUS_PORT,default=12411"`
	IPAllowList         httputil.IPAllowList `env:"IP_ALLOW_LIST"`
	// Partitions number of partitions for structured-data.articles.v1 topic
	Partitions int `env:"PARTITIONS,default=50"`
	// MaxParts max parts a query can be sent with. This is also the max number of parallel connections allowed to realtime API.
	// e.g., if max_parts is 10 and partitions is 50, each part represents (50/10) 5 consecutive partitions.
	// So, query with parts [0,1] will point to partitions [0,1,2,3,4,5,6,7,8,9]
	MaxParts int `env:"MAX_PARTS,default=10"`

	// Workers number of parallel consumers from kafka that processes and writes events.
	Workers int `env:"WORKERS,default=1"`
	// ThrottlingMsgsPerSecond limits the write rate.
	ThrottlingMsgsPerSecond int `env:"THROTTLING_MSGS_PER_SECOND,default=50"`
	// If true, THROTTLING_MSGS_PER_SECOND will be the max throughput for a request for 10 parts, and requests
	// for fewer parts will have lower limits. e.g. a request for 1 part will be limited to 0.1 *  THROTTLING_MSGS_PER_SECOND msg/s.
	PerPartThrottling bool `env:"PER_PART_THROTTLING,default=true"`

	ArticlesTopic string `env:"ARTICLES_TOPIC,default=aws.structured-data.articles.v1"`
	// Populate these for wikidata and structured-contents support
	StructuredTopic string `env:"STRUCTURED_TOPIC"`
	EntityTopic     string `env:"ENTITY_TOPIC"`

	// This roughly translates to per-request memory cap
	ArticleChannelSize int `env:"ARTICLE_CHANNEL_SIZE,default=50"`

	// Remove all the following env - marked for deprecation
	KSQLURL      string `env:"KSQL_URL,default=http://ksqldb:8088"`
	KSQLUsername string `env:"KSQL_USERNAME"`
	KSQLPassword string `env:"KSQL_PASSWORD"`
	// ArticlesStream is the name of the ksqldb stream for articles data.
	// Don't forget to update this after new migration.
	ArticlesStream                 string `env:"ARTICLES_STREAM,default=rlt_articles_str_v8"`
	KSQLHealthCheckTimeoutSeconds  int    `env:"KSQL_HEALTH_TIMEOUT,default=5"`
	KSQLHealthCheckIntervalSeconds int    `env:"KSQL_HEALTH_INTERVAL,default=10"`
	RedisHealthCheckTimeoutSeconds int    `env:"REDIS_HEALTH_TIMEOUT,default=2"`
	// If true, /articlesdirect won't be registered.
	DisableDirectHandler bool `env:"DISABLE_DIRECT_HANDLER"`
	// If true, the second version of the struct filter will be used in the direct realtime handler.
	// Two versions have been implemented, respectively: one that checks every field of every message to decide whether to keep or drop,
	// and one that builds a data structure at the beginning of the requesst and uses it on each message to only check
	// the fields that need to be dropped.
	UseDirectHandlerFilterV2 bool `env:"USE_DIRECT_HANDLER_FILTER_V2,default=true"`
	// For switching to old ksqldb-based handler
	UseKsqldb bool `env:"USE_KSQLDB,default=false"`
	// How many messages between transfer rate checks.
	ThrottlingBatchSize int `env:"THROTTLING_BATCH_SIZE,default=32"`
}

// New initialize the environment
func New() (*Environment, error) {
	var (
		_, b, _, _ = runtime.Caller(0)
		base       = filepath.Dir(b)
		_          = godotenv.Load(fmt.Sprintf("%s/../../.env", base))
		config     = new(Environment)
	)

	_, err := env.UnmarshalFromEnviron(config)
	return config, err
}

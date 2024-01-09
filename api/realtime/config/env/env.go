// Package env provides the ability to validate and parse environment variables.
package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/general/httputil"

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
const defaultPolicy = `p, realtime, /v1/page-update, GET
p, realtime, /v1/page-delete, GET
p, realtime, /v1/page-visibility, GET
p, realtime, /v2/articles, GET
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

// Environment environment variables configuration.
type Environment struct {
	KSQLURL             string               `env:"KSQL_URL,default=http://ksqldb:8088"`
	KSQLUsername        string               `env:"KSQL_USERNAME"`
	KSQLPassword        string               `env:"KSQL_PASSWORD"`
	ServerMode          string               `env:"SERVER_MODE,default=release"`
	ServerPort          string               `env:"SERVER_PORT,default=4040"`
	AWSRegion           string               `env:"AWS_REGION"`
	AWSID               string               `env:"AWS_ID,default=foo"`
	AWSKey              string               `env:"AWS_KEY,default=bar"`
	CognitoClientID     string               `env:"COGNITO_CLIENT_ID,required=true"`
	CognitoClientSecret string               `env:"COGNITO_CLIENT_SECRET,required=true"`
	RedisAddr           string               `env:"REDIS_ADDR,required=true"`
	RedisPassword       string               `env:"REDIS_PASSWORD,required=true"`
	AccessModel         *AccessModel         `env:"ACCESS_MODEL"`
	AccessPolicy        *AccessPolicy        `env:"ACCESS_POLICY"`
	PrometheusPort      int                  `env:"PROMETHEUS_PORT,default=12411"`
	IPAllowList         httputil.IPAllowList `env:"IP_ALLOW_LIST"`
	// Partitions number of partitions for structured-data.articles.v1 topic
	Partitions int `env:"PARTITIONS,default=50"`
	// MaxParts max parts a query can be sent with. This is also the max number of parallel connections allowed to realtime API.
	// e.g., if max_parts is 10 and partitions is 50, each part represents (50/10) 5 consecutive partitions.
	// So, query with parts [0,1] will point to partitions [0,1,2,3,4,5,6,7,8,9]
	MaxParts int `env:"MAX_PARTS,default=10"`

	// ArticlesStream is the name of the ksqldb stream for articles data.
	// Don't forget to update this after new migration.
	ArticlesStream string `env:"ARTICLES_STREAM,default=rlt_articles_str_v3"`
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

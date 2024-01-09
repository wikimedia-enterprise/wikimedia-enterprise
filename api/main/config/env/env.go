// Package env provides the ability to validate and parse environment variables.
package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

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
const defaultPolicy = `p, snapshots, /v1/exports/download/*, GET
p, snapshots, /v1/exports/download/*, GET
p, snapshots, /v1/exports/download/*, HEAD
p, snapshots, /v1/exports/meta/*, GET
p, snapshots, /v2/snapshots, GET
p, snapshots, /v2/snapshots, POST
p, snapshots, /v2/snapshots/:identifier, GET
p, snapshots, /v2/snapshots/:identifier, POST
p, snapshots, /v2/snapshots/:identifier/download, GET
p, snapshots, /v2/snapshots/:identifier/download, HEAD
g, group_1, snapshots
g, group_2, snapshots
g, group_3, snapshots

p, batches, /v1/diffs/download/*, GET
p, batches, /v1/diffs/download/*, HEAD
p, batches, /v1/diffs/meta/*, GET
p, batches, /v2/batches/:date, GET
p, batches, /v2/batches/:date, POST
p, batches, /v2/batches/:date/:identifier, GET
p, batches, /v2/batches/:date/:identifier, POST
p, batches, /v2/batches/:date/:identifier/download, GET
p, batches, /v2/batches/:date/:identifier/download, HEAD
g, group_2, batches
g, group_3, batches

p, beta, /v2/structured-contents/:name, GET
p, beta, /v2/structured-contents/:name, POST
g, group_2, beta
g, group_3, beta

p, meta, /v1/projects, GET
p, meta, /v1/namespaces, GET
p, meta, /v1/pages/meta/*, GET
p, meta, /v2/codes, GET
p, meta, /v2/codes, POST
p, meta, /v2/codes/:identifier, GET
p, meta, /v2/codes/:identifier, POST
p, meta, /v2/languages, GET
p, meta, /v2/languages, POST
p, meta, /v2/languages/:identifier, GET
p, meta, /v2/languages/:identifier, POST
p, meta, /v2/projects, GET
p, meta, /v2/projects, POST
p, meta, /v2/projects/:identifier, GET
p, meta, /v2/projects/:identifier, POST
p, meta, /v2/namespaces, GET
p, meta, /v2/namespaces, POST
p, meta, /v2/namespaces/:identifier, GET
p, meta, /v2/namespaces/:identifier, POST
p, meta, /v2/articles/:name, GET
p, meta, /v2/articles/:name, POST
g, group_1, meta
g, group_2, meta
g, group_3, meta`

func unmarshalAccess(dta string, fnm string, cnt string) (string, error) {
	f, err := os.CreateTemp("/tmp", fnm)

	if err != nil {
		aer := fmt.Errorf("error creating tmp file %s: %w", fnm, err)
		log.Error(aer)
		return "", aer
	}

	if len(dta) > 0 {
		_, err = f.Write([]byte(dta))
	} else {
		_, err = f.Write([]byte(cnt))
	}

	if err != nil {
		aer := fmt.Errorf("error writing to tmp file %s: %w", fnm, err)
		log.Error(aer)
		return "", aer
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
	AWSRegion           string               `env:"AWS_REGION,required=true"`
	AWSID               string               `env:"AWS_ID,required=true"`
	AWSKey              string               `env:"AWS_KEY,required=true"`
	ServerMode          string               `env:"SERVER_MODE,default=release"`
	ServerPort          string               `env:"SERVER_PORT,default=4060"`
	AWSURL              string               `env:"AWS_URL"`
	AWSBucket           string               `env:"AWS_BUCKET,required=true"`
	CognitoClientID     string               `env:"COGNITO_CLIENT_ID,required=true"`
	CognitoClientSecret string               `env:"COGNITO_CLIENT_SECRET,required=true"`
	RedisAddr           string               `env:"REDIS_ADDR,required=true"`
	RedisPassword       string               `env:"REDIS_PASSWORD,required=true"`
	AccessModel         *AccessModel         `env:"ACCESS_MODEL,required=true"`
	AccessPolicy        *AccessPolicy        `env:"ACCESS_POLICY,required=true"`
	RateLimitsByGroup   *httputil.Limiter    `env:"RATE_LIMITS_BY_GROUP"`
	CapConfig           *httputil.CapConfig  `env:"CAP_CONFIGURATION"`
	IPAllowList         httputil.IPAllowList `env:"IP_ALLOW_LIST"`
	FreeTierGroup       string               `env:"FREE_TIER_GROUP,default=group_1"`
	PrometheusPort      int                  `env:"PROMETHEUS_PORT,default=12411"`
	DescriptionEnabled  bool                 `env:"DESCRIPTION_ENABLED,default=true"`
	SectionsEnabled     bool                 `env:"SECTIONS_ENABLED,default=true"`
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

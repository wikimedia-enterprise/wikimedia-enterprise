// Package env provides the ability to validate and parse environment variables.
package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"wikimedia-enterprise/api/main/submodules/httputil"
	"wikimedia-enterprise/api/main/submodules/log"

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
const defaultPolicy = `p, snapshots, /v2/snapshots, GET
p, snapshots, /v2/snapshots, GET
p, snapshots, /v2/snapshots, POST
p, snapshots, /v2/snapshots/:identifier, GET
p, snapshots, /v2/snapshots/:identifier, POST
p, snapshots, /v2/snapshots/:identifier/download, GET
p, snapshots, /v2/snapshots/:identifier/download, HEAD
g, group_1, snapshots
g, group_2, snapshots
g, group_3, snapshots

p, structured-snapshots, /v2/snapshots/structured-contents, GET
p, structured-snapshots, /v2/snapshots/structured-contents, POST
p, structured-snapshots, /v2/snapshots/structured-contents/:identifier, GET
p, structured-snapshots, /v2/snapshots/structured-contents/:identifier, POST
p, structured-snapshots, /v2/snapshots/structured-contents/:identifier/download, GET
p, structured-snapshots, /v2/snapshots/structured-contents/:identifier/download, HEAD
g, group_2, structured-snapshots
g, group_3, structured-snapshots

p, batches, /v2/batches/:date/:hour, GET
p, batches, /v2/batches/:date/:hour, POST
p, batches, /v2/batches/:date/:hour/:identifier, GET
p, batches, /v2/batches/:date/:hour/:identifier, POST
p, batches, /v2/batches/:date/:hour/:identifier/download, GET
p, batches, /v2/batches/:date/:hour/:identifier/download, HEAD
g, group_2, batches
g, group_3, batches

p, beta, /v2/structured-contents/*, GET
p, beta, /v2/structured-contents/*, POST
g, group_2, beta
g, group_3, beta

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
p, meta, /v2/articles/*, GET
p, meta, /v2/articles/*, POST
g, group_1, meta
g, group_2, meta
g, group_3, meta

p, chunks, /v2/snapshots/:identifier/chunks, GET
p, chunks, /v2/snapshots/:identifier/chunks, POST
p, chunks, /v2/snapshots/:identifier/chunks/:chunkIdentifier, GET
p, chunks, /v2/snapshots/:identifier/chunks/:chunkIdentifier, POST
p, chunks, /v2/snapshots/:identifier/chunks/:chunkIdentifier/download, GET
p, chunks, /v2/snapshots/:identifier/chunks/:chunkIdentifier/download, HEAD
g, group_1, chunks
g, group_2, chunks
g, group_3, chunks

p, files, /v2/files/:filename, GET
p, files, /v2/files/:filename, POST
p, files, /v2/files/:filename/download, GET
g, group_3, files`

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
// Note that TEST_ONLY_GROUP is only effective when using TEST_ONLY_SKIP_AUTH.
type Environment struct {
	AWSRegion                       string                     `env:"AWS_REGION,required=true"`
	AWSID                           string                     `env:"AWS_ID,required=true"`
	AWSKey                          string                     `env:"AWS_KEY,required=true"`
	ServerMode                      string                     `env:"SERVER_MODE,default=release"`
	ServerPort                      string                     `env:"SERVER_PORT,default=4060"`
	AWSURL                          string                     `env:"AWS_URL"`
	AWSBucket                       string                     `env:"AWS_BUCKET,required=true"`
	AWSBucketCommons                string                     `env:"AWS_BUCKET_COMMONS,required=false"`
	CognitoClientID                 string                     `env:"COGNITO_CLIENT_ID,required=true"`
	CognitoClientSecret             string                     `env:"COGNITO_CLIENT_SECRET,required=true"`
	RedisAddr                       string                     `env:"REDIS_ADDR,required=true"`
	RedisPassword                   string                     `env:"REDIS_PASSWORD,required=true"`
	AccessModel                     *AccessModel               `env:"ACCESS_MODEL,required=true"`
	AccessPolicy                    *AccessPolicy              `env:"ACCESS_POLICY,required=true"`
	RateLimitsByGroup               *httputil.Limiter          `env:"RATE_LIMITS_BY_GROUP"`
	CapConfig                       *httputil.CapConfigWrapper `env:"CAP_CONFIGURATION,default=[]"`
	IPAllowList                     *httputil.IPAllowList      `env:"IP_ALLOW_LIST,default=[]"`
	FreeTierGroup                   string                     `env:"FREE_TIER_GROUP,default=group_1"`
	PrometheusPort                  int                        `env:"PROMETHEUS_PORT,default=12411"`
	MetricsReadTimeOutSeconds       int                        `env:"METRICS_READ_TIMEOUT_SECONDS,default=10"`
	MetricsReadHeaderTimeOutSeconds int                        `env:"METRICS_READ_HEADER_TIMEOUT_SECONDS,default=10"`
	MetricsWriteTimeoutSeconds      int                        `env:"METRICS_WRITE_TIMEOUT_SECONDS,default=10"`
	DescriptionEnabled              bool                       `env:"DESCRIPTION_ENABLED,default=true"`
	SectionsEnabled                 bool                       `env:"SECTIONS_ENABLED,default=true"`
	ArticleKeyTypeSuffix            string                     `env:"KEY_TYPE_SUFFIX"`
	LogLevel                        string                     `env:"LOG_LEVEL,default=info"`
	HealthChecksTimeoutMs           int                        `env:"HEALTH_CHECKS_TIMEOUT_MS,default=5000"`
	TestOnlySkipAuth                bool                       `env:"TEST_ONLY_SKIP_AUTH,default=false"`
	TestOnlyGroup                   string                     `env:"TEST_ONLY_GROUP"`
	HealthCheckMaxRetries           int                        `env:"HEALTH_CHECK_MAX_RETRIES,default=3"`
	EnablePProf                     bool                       `env:"ENABLE_PPROF,default=false"`
	EnableAttribution               bool                       `env:"ENABLE_ATTRIBUTION,default=true"`
	EnableTables                    bool                       `env:"ENABLE_TABLES,default=true"`
	UseHashedPrefixes               bool                       `env:"USE_HASHED_PREFIXES,default=false"`
	TablesConfidenceThreshold       float64                    `env:"TABLES_CONFIDENCE_THRESHOLD,default=0.5"`
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

	// Set default value if optional Commons bucket field is missing
	if len(config.AWSBucketCommons) == 0 {
		config.AWSBucketCommons = config.AWSBucket
	}

	return config, err
}

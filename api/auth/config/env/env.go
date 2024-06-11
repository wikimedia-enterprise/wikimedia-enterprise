// Package env provides the ability to validate and parse environment variables.
package env

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"wikimedia-enterprise/general/log"

	env "github.com/Netflix/go-env"
	"github.com/gocarina/gocsv"
	"github.com/joho/godotenv"
)

const defaultPolicy = `p, exports, /v1/exports/download/:namespace/:project, GET
p, exports, /v1/exports/download/:namespace/:project, HEAD
p, exports, /v1/exports/meta/*, GET
p, exports, /v1/exports/meta/*, HEAD
g, group_1, exports
g, group_2, exports
g, group_3, exports
p, diffs, /v1/diffs/download/*, GET
p, diffs, /v1/diffs/download/*, HEAD
p, diffs, /v1/diffs/meta/*, GET
p, diffs, /v1/diffs/meta/*, HEAD
g, group_2, diffs
g, group_3, diffs
p, real_time, /v1/page-update, GET
p, real_time, /v1/page-delete, GET
p, real_time, /v1/page-visibility, GET
g, group_3, real_time
p, meta, /v1/projects, GET
p, meta, /v1/namespaces, GET
p, meta, /v1/pages/meta/*, GET
g, group_1, meta
g, group_2, meta
g, group_3, meta
p, *, /v1/docs/*, GET
p, *, /v1/status, GET

`

// AccessPath structure represents access policy csv row.
type AccessPath struct {
	Marker string `csv:"marker" json:"-"`
	Name   string `csv:"name" json:"-"`
	Path   string `csv:"path" json:"path"`
	Method string `csv:"method" json:"method"`
}

// AccessPolicy structure wrapps map of groups with acceptable paths.
type AccessPolicy struct {
	Map map[string][]AccessPath
}

// UnmarshalEnvironmentValue called by env package on initialization to create a map of groups with acceptable paths.
func (m *AccessPolicy) UnmarshalEnvironmentValue(data string) (err error) {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)

		r.TrimLeadingSpace = true
		r.FieldsPerRecord = -1

		return r
	})

	ama := []AccessPath{}

	if len(data) > 0 {
		if err != gocsv.UnmarshalWithoutHeaders(strings.NewReader(data), &ama) {
			log.Error(err, log.Tip("problem read environment values in CSV"))
			return fmt.Errorf("error reading csv: %w", err)
		}
	} else {
		if err != gocsv.UnmarshalWithoutHeaders(strings.NewReader(defaultPolicy), &ama) {
			log.Error(err, log.Tip("problem read environment values in CSV. No data"))
			return fmt.Errorf("error reading csv: %w", err)
		}
	}

	pbg := make(map[string][]string)
	pbn := make(map[string][]AccessPath)

	for _, amr := range ama {
		if amr.Marker == "p" {
			pbn[amr.Name] = append(pbn[amr.Name], amr)
		}
		if amr.Marker == "g" {
			pbg[amr.Name] = append(pbg[amr.Name], amr.Path)
		}
	}

	egp := make(map[string][]AccessPath)

	for grp, grs := range pbg {
		for _, grn := range grs {
			egp[grp] = append(egp[grp], pbn[grn]...)
		}
	}

	m.Map = egp

	return nil
}

// List of Strings from environment variables (JSON Encoded).
type List []string

// UnmarshalEnvironmentValue called by env package on initialization to unmarshal json value.
func (t *List) UnmarshalEnvironmentValue(data string) error {
	_ = json.Unmarshal([]byte(data), t)

	if len(*t) == 0 {
		*t = strings.Split(data, ",")
	}

	return nil
}

// Environment environment variables configuration.
type Environment struct {
	AWSRegion              string        `env:"AWS_REGION,required=true"`
	AWSID                  string        `env:"AWS_ID,required=true"`
	AWSKey                 string        `env:"AWS_KEY,required=true"`
	CognitoClientID        string        `env:"COGNITO_CLIENT_ID,required=true"`
	CognitoSecret          string        `env:"COGNITO_SECRET,required=true"`
	CognitoUserPoolID      string        `env:"COGNITO_USER_POOL_ID,required=true"`
	CognitoUserGroup       string        `env:"COGNITO_USER_GROUP,required=true"`
	ServerMode             string        `env:"SERVER_MODE,default=release"`
	ServerPort             string        `env:"SERVER_PORT,default=4050"`
	RedisAddr              string        `env:"REDIS_ADDR,default=redis:6379"`
	RedisPassword          string        `env:"REDIS_PASSWORD"`
	MaxAccessTokens        int64         `env:"MAX_ACCESS_TOKENS,default=90"`
	AccessTokensExpHours   int64         `env:"ACCESS_TOKENS_EXPIRATION_HOURS,default=2160"`
	IPRange                string        `env:"IP_RANGE"`
	CognitoCacheExpiration int           `env:"COGNITO_CACHE_EXPIRATION,default=300"`
	AccessPolicy           *AccessPolicy `env:"ACCESS_POLICY,required=true"`
	GroupDownloadLimit     string        `env:"GROUP_DOWNLOAD_LIMIT,required=true"`
	PrometheusPort         int           `env:"PROMETHEUS_PORT,default=12411"`
	DomainDenyList         List          `env:"DOMAIN_DENY_LIST,default="`
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

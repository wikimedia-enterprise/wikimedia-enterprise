package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/hamba/avro/v2"
)

// ErrNoAVRO avro field is not set for schema.
var ErrNoAVRO = errors.New("AVRO schema not set")

// Supported schema types.
const (
	SchemaTypeAVRO = "AVRO"
)

// Parser wrapper to make `avro.Parse` function thread safe.
// More info here https://pkg.go.dev/github.com/hamba/avro/v2#Parser.
type Parser struct {
	mutex sync.Mutex
}

// Parse schema string into avro schema.
// Safe to call from multiple goroutines,
// blocks if another thread is parsing the schema.
func (p *Parser) Parse(sch string) (avro.Schema, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return avro.Parse(sch)
}

// This is an instance of the `Parser`.
// Used from multiple goroutines due to the non-thread safety of the `avro.Parse` function.
// P.S. This is not an ideal solution, but it's simple and it works.
var parser = new(Parser)

// SubjectCreator is the interface to wrap default CreateSubject method for unit testing.
type SubjectCreator interface {
	CreateSubject(ctx context.Context, name string, subject *Subject) (*Schema, error)
}

// BySubjectGetter is the interface to wrap default GetBySubject method for unit testing.
type BySubjectGetter interface {
	GetBySubject(ctx context.Context, name string, versions ...int) (*Schema, error)
}

// ByIdGetter is the interface to wrap default GetByID method for unit testing.
type ByIdGetter interface {
	GetByID(ctx context.Context, id int) (*Schema, error)
}

// GetterCreator wraps all methods of the schema under single interface for unit testing.
type GetterCreator interface {
	SubjectCreator
	ByIdGetter
	BySubjectGetter
}

// BasicAuth struct to pass authentication to schema registry.
type BasicAuth struct {
	Username string
	Password string
}

// Reference schema references that need to resolved inside the schema.
// More info here https://docs.confluent.io/platform/current/schema-registry/serdes-develop/index.html#schema-references.
type Reference struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Version int    `json:"version"`
}

// Subject refers to the name under which the schema is registered.
// More info https://docs.confluent.io/platform/current/schema-registry/develop/api.html#subjects.
type Subject struct {
	Schema     string       `json:"schema"`
	SchemaType string       `json:"schemaType"`
	References []*Reference `json:"references"`
}

// Schema value from the registry.
// More info https://docs.confluent.io/platform/current/schema-registry/develop/api.html#subjects.
type Schema struct {
	ID         int          `json:"id"`
	Subject    string       `json:"subject"`
	Version    int          `json:"version"`
	Schema     string       `json:"schema"`
	References []*Reference `json:"references"`
	AVRO       avro.Schema  `json:"-"`
}

// Parse parse current registry schema into avro.
func (s *Schema) Parse() error {
	var err error
	s.AVRO, err = parser.Parse(s.Schema)
	return err
}

// Marshal encode value using this schema.
func (s *Schema) Marshal(v interface{}) ([]byte, error) {
	if s.AVRO == nil {
		return nil, ErrNoAVRO
	}

	return Marshal(s.ID, s.AVRO, v)
}

// Unmarshal decode value using this schema.
func (s *Schema) Unmarshal(data []byte, v interface{}, api avro.API) error {
	if s.AVRO == nil {
		return ErrNoAVRO
	}

	return Unmarshal(s.AVRO, data, v, api)
}

// NewRegistry create new schema registry client.
func NewRegistry(url string, opts ...func(cl *Registry)) *Registry {
	cl := &Registry{
		url:        url,
		HTTPClient: new(http.Client),
	}

	for _, opt := range opts {
		opt(cl)
	}

	return cl
}

// Client schema registry API client.
type Registry struct {
	url        string
	HTTPClient *http.Client
	BasicAuth  *BasicAuth
}

func (r *Registry) req(ctx context.Context, method string, endpoint string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", r.url, endpoint), bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	res, err := r.HTTPClient.Do(req)

	if err != nil {
		return res, err
	}

	if res.StatusCode != http.StatusOK {
		data, err := io.ReadAll(res.Body)

		if err != nil {
			return res, err
		}

		_ = res.Body.Close()
		return res, fmt.Errorf("%s:%s", res.Status, string(data))
	}

	return res, nil
}

// CreateSubject create new schema subject by name.
func (r *Registry) CreateSubject(ctx context.Context, name string, subject *Subject) (*Schema, error) {
	res, err := r.req(ctx, http.MethodPost, fmt.Sprintf("subjects/%s/versions", name), subject)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	sch := new(Schema)

	if err := json.NewDecoder(res.Body).Decode(sch); err != nil {
		return nil, err
	}

	return sch, nil
}

// GetBySubject get particular version of the schema by subject and version.
// If no version argument is provided then gets the latest version.
func (r *Registry) GetBySubject(ctx context.Context, name string, versions ...int) (*Schema, error) {
	version := "latest"

	if len(versions) > 0 {
		version = strconv.Itoa(versions[0])
	}

	res, err := r.req(ctx, http.MethodGet, fmt.Sprintf("subjects/%s/versions/%s", name, version), nil)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	sch := new(Schema)

	if err := json.NewDecoder(res.Body).Decode(sch); err != nil {
		return nil, err
	}

	return sch, nil
}

// GetByID get schema independent from subject (by id).
func (r *Registry) GetByID(ctx context.Context, id int) (*Schema, error) {
	res, err := r.req(ctx, http.MethodGet, fmt.Sprintf("schemas/ids/%d", id), nil)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	sch := &Schema{
		ID: id,
	}

	if err := json.NewDecoder(res.Body).Decode(sch); err != nil {
		return nil, err
	}

	return sch, nil
}

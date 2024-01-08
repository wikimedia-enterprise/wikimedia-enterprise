package schema

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hamba/avro/v2"
	"github.com/stretchr/testify/suite"
)

func createServer(id int, subject string, version int, schema string) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc(fmt.Sprintf("/subjects/%s/versions", subject), func(rw http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(rw, `{"id":%d}`, id)
	})

	router.HandleFunc(fmt.Sprintf("/subjects/%s/versions/%d", subject, version), func(rw http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(rw, `{"subject":"%s", "version": %d, "id": %d,"schema": "%s"}`, subject, version, id, strings.ReplaceAll(schema, `"`, `\"`))
	})

	router.HandleFunc(fmt.Sprintf("/schemas/ids/%d", id), func(rw http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(rw, `{"schema": "%s"}`, strings.ReplaceAll(schema, `"`, `\"`))
	})

	return router
}

type clientTestSuite struct {
	suite.Suite
	ctx     context.Context
	srv     *httptest.Server
	id      int
	version int
	subject string
	schema  string
	reg     *Registry
}

func (s *clientTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.srv = httptest.NewServer(createServer(s.id, s.subject, s.version, s.schema))
	s.reg = NewRegistry(s.srv.URL)
}

func (s *clientTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *clientTestSuite) TestCreateSubject() {
	sch, err := s.reg.CreateSubject(s.ctx, s.subject, &Subject{
		Schema:     s.schema,
		SchemaType: SchemaTypeAVRO,
	})

	s.Assert().NoError(err)
	s.Assert().Equal(s.id, sch.ID)
}

func (s *clientTestSuite) TestGetByID() {
	sch, err := s.reg.GetByID(s.ctx, s.id)

	s.Assert().NoError(err)
	s.Assert().Equal(s.schema, sch.Schema)
}

func (s *clientTestSuite) TestGetBySubject() {
	sch, err := s.reg.GetBySubject(s.ctx, s.subject, s.version)

	s.Assert().NoError(err)
	s.Assert().Equal(s.id, sch.ID)
	s.Assert().Equal(s.subject, sch.Subject)
	s.Assert().Equal(s.version, sch.Version)
	s.Assert().Equal(s.schema, sch.Schema)
}

func TestClient(t *testing.T) {
	for _, testCase := range []*clientTestSuite{
		{
			id:      10,
			subject: "key",
			version: 1,
			schema:  `{"type":"record","name":"Key","namespace":"wikimedia_enterprise.general.schema","fields":[{"name":"identifier","type":"string"},{"name":"type","type":"string"}]}`,
		},
	} {
		suite.Run(t, testCase)
	}
}

type schemaTestValue struct {
	Identifier string `json:"identifier" avro:"identifier"`
	Type       string `json:"type" avro:"type"`
}

type schemaTestSuite struct {
	suite.Suite
	val *schemaTestValue
	sch *Schema
}

func (s *schemaTestSuite) SetupSuite() {
	s.val = &schemaTestValue{
		Identifier: "unique",
		Type:       "article",
	}
}

func (s *schemaTestSuite) SetupTest() {
	s.sch = &Schema{
		ID:     10,
		Schema: `{"type":"record","name":"Key","namespace":"wikimedia_enterprise.general.schema","fields":[{"name":"identifier","type":"string"},{"name":"type","type":"string"}]}`,
	}
}

func (s *schemaTestSuite) TestMarshal() {
	s.Assert().NoError(s.sch.Parse())

	data, err := s.sch.Marshal(s.val)
	s.Assert().NoError(err)

	val := new(schemaTestValue)
	s.Assert().NoError(avro.Unmarshal(s.sch.AVRO, data[5:], val))
	s.Assert().Equal(s.val, val)
}

func (s *schemaTestSuite) TestMarshalErr() {
	_, err := new(Schema).Marshal(s.val)
	s.Assert().Equal(ErrNoAVRO, err)
}

func (s *schemaTestSuite) TestUnmarshal() {
	s.Assert().NoError(s.sch.Parse())

	data, err := avro.Marshal(s.sch.AVRO, s.val)
	s.Assert().NoError(err)

	val := new(schemaTestValue)
	s.Assert().NoError(s.sch.Unmarshal(append(make([]byte, 5), data...), val, nil))
	s.Assert().Equal(s.val, val)
}

func (s *schemaTestSuite) TestUnmarshalErr() {
	s.Assert().Equal(ErrNoAVRO, new(Schema).Unmarshal([]byte{}, new(schemaTestValue), nil))
}

func (s *schemaTestSuite) TestParse() {
	s.Assert().NoError(s.sch.Parse())
	s.Assert().NotNil(s.sch.AVRO)
}

func (s *schemaTestSuite) TestParseErr() {
	s.Assert().Error(new(Schema).Parse())
}

func TestSchema(t *testing.T) {
	suite.Run(t, new(schemaTestSuite))
}

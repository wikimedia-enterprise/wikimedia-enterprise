package resolver_test

import (
	"reflect"
	"testing"
	"time"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"github.com/stretchr/testify/suite"
)

type resolverEmptyModelMock struct {
	TableName struct{} `json:"-"`
}

type resolverSubModelMock struct {
	Identifier int    `json:"identifier,omitempty" avro:"identifier"`
	InLanguage string `json:"in_language,omitempty" avro:"in_language"`
	Event      string `json:"event,omitempty" avro:"event"`
}

type resolverModelMock struct {
	TableName        struct{}                `json:"-" ksql:"model"`
	Identifier       int                     `json:"identifier,omitempty" avro:"identifier"`
	ParentIdentifier int                     `json:"parent_identifier,omitempty" avro:"parent_identifier"`
	Comment          string                  `json:"comment,omitempty" avro:"comment"`
	SubModel         *resolverSubModelMock   `json:"sub_model,omitempty" avro:"sub_model"`
	DateStarted      *time.Time              `json:"date_started,omitempty" avro:"date_started"`
	SubModels        []*resolverSubModelMock `json:"sub_models,omitempty" avro:"sub_models"`
}

type newResolverTestSuite struct {
	suite.Suite
	model interface{}
}

func (s *newResolverTestSuite) TestNew() {
	mrv := reflect.ValueOf(s.model)
	rvr, err := resolver.New(s.model, func(r *resolver.Resolver) {
		r.HasJoins = true
	})

	if mrv.Kind() != reflect.Ptr || mrv.IsNil() {
		s.Assert().Contains(err.Error(), "non pointer reference")
	} else if _, ok := mrv.Elem().Type().FieldByName("TableName"); !ok {
		s.Assert().Equal(resolver.ErrNoTableName, err)
	} else if fld, ok := mrv.Elem().Type().FieldByName("TableName"); ok && len(fld.Tag.Get("ksql")) <= 0 {
		s.Assert().Equal(resolver.ErrNoKSQLTag, err)
	} else {
		s.Assert().NotNil(rvr)
		s.Assert().NoError(err)
	}
}

func TestNew(t *testing.T) {
	for _, testCase := range []*newResolverTestSuite{
		{
			model: new(resolverModelMock),
		},
		{
			model: new(resolverSubModelMock),
		},
		{
			model: new(resolverEmptyModelMock),
		},
		{
			model: resolverModelMock{},
		},
	} {
		suite.Run(t, testCase)
	}
}

type newResolversTestSuite struct {
	suite.Suite
	models map[string]interface{}
	fn     func(r *resolver.Resolver)
}

func (s *newResolversTestSuite) TestNewResolvers() {
	resolvers, err := resolver.NewResolvers(s.models, s.fn)
	s.Assert().NoError(err)
	s.Assert().Equal(len(resolvers), len(s.models))
}

func TestNewResolvers(t *testing.T) {
	for _, testCase := range []*newResolversTestSuite{
		{
			models: map[string]interface{}{
				"article": new(resolverModelMock),
				"entity":  new(resolverModelMock),
			},
		},
		{
			models: map[string]interface{}{},
		},
		{
			models: map[string]interface{}{
				"article":    new(resolverModelMock),
				"entity":     new(resolverModelMock),
				"structured": new(resolverModelMock),
			},
			fn: func(r *resolver.Resolver) {
				r.Keywords = map[string]string{
					"keyword": "`keyword`",
				}
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

type getSchemaTestSuite struct {
	suite.Suite
}

func (s *getSchemaTestSuite) TestGetSchema() {
	resolvers, err := resolver.NewResolvers(map[string]interface{}{
		schema.KeyTypeArticle: new(schema.Article),
	})

	s.Assert().NoError(err)

	out := resolvers.GetSchema("articles")
	s.Assert().IsType(new(schema.Article), out)

	out = resolvers.GetSchema("unknown")
	s.Assert().Nil(out)
}

func TestGetSchema(t *testing.T) {
	suite.Run(t, new(getSchemaTestSuite))
}

type resolverTestSuite struct {
	suite.Suite
	model             interface{}
	rvr               *resolver.Resolver
	fields            []string
	structs           []string
	slices            []string
	filters           []resolver.Filter
	fieldsContain     []string
	structsContain    []string
	structsNotContain []string
	slicesContain     []string
	slicesNotContain  []string
}

func (s *resolverTestSuite) SetupSuite() {
	var err error
	s.rvr, err = resolver.New(s.model, func(r *resolver.Resolver) {
		r.HasJoins = true
	})
	s.Assert().NoError(err)
	s.Assert().NotNil(s.rvr)
}

func (s *resolverTestSuite) TestHasField() {
	for _, fld := range s.fields {
		s.Assert().True(s.rvr.HasField(fld))
	}

	for _, str := range s.structs {
		s.Assert().True(s.rvr.HasField(str))
	}

	for _, slc := range s.slices {
		s.Assert().True(s.rvr.HasField(slc))
	}
}

func (s *resolverTestSuite) TestHasSlice() {
	for _, str := range s.structs {
		s.Assert().False(s.rvr.HasSlice(str))
	}

	for _, slc := range s.slices {
		s.Assert().True(s.rvr.HasSlice(slc))
	}
}

func (s *resolverTestSuite) TestHasStruct() {
	for _, str := range s.structs {
		s.Assert().True(s.rvr.HasStruct(str))
	}

	for _, slc := range s.slices {
		s.Assert().False(s.rvr.HasStruct(slc))
	}
}

func TestResolver(t *testing.T) {
	for _, testCase := range []*resolverTestSuite{
		{
			model: new(resolverModelMock),
			fields: []string{
				"identifier",
				"parent_identifier",
				"comment",
				"date_started",
			},
			fieldsContain: []string{
				"model_identifier",
				"model_parent_identifier",
				"model_comment",
				"model_date_started",
			},
			structs: []string{
				"sub_model",
			},
			structsContain: []string{
				"STRUCT(",
				"identifier := model_sub_model->identifier",
				"in_language := model_sub_model->in_language",
				"event := model_sub_model->event",
			},
			slices: []string{
				"sub_models",
			},
			slicesContain: []string{
				"TRANSFORM(model_sub_models, x => ",
				"STRUCT(",
				"identifier := sub_models->identifier",
				"in_language := sub_models->in_language",
				"event := sub_models->event",
			},
			filters: []resolver.Filter{
				func(fld *resolver.Field) bool {
					return true
				},
			},
		},
		{
			model: new(resolverModelMock),
			fields: []string{
				"identifier",
				"parent_identifier",
				"comment",
				"date_started",
			},
			fieldsContain: []string{
				"model_identifier",
				"model_parent_identifier",
				"model_comment",
				"model_date_started",
			},
			structs: []string{
				"sub_model",
			},
			structsContain: []string{
				"STRUCT(",
				"identifier := model_sub_model->identifier",
				") as model_sub_model",
			},
			structsNotContain: []string{
				"in_language := model_sub_model->in_language",
				"event := model_sub_model->event",
			},
			slices: []string{
				"sub_models",
			},
			slicesContain: []string{
				"TRANSFORM(model_sub_models, x => ",
				"STRUCT(",
				"identifier := sub_models->identifier",
			},
			slicesNotContain: []string{
				"in_language := model_sub_model->in_language",
				"event := model_sub_model->event",
			},
			filters: []resolver.Filter{
				func(fld *resolver.Field) bool {
					return fld.Name != "in_language" && fld.Name != "event"
				},
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

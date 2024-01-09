package resolver_test

import (
	"reflect"
	"testing"
	"time"
	"wikimedia-enterprise/api/realtime/libraries/resolver"

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
	TableName        struct{}              `json:"-" ksql:"model"`
	Identifier       int                   `json:"identifier,omitempty" avro:"identifier"`
	ParentIdentifier int                   `json:"parent_identifier,omitempty" avro:"parent_identifier"`
	Comment          string                `json:"comment,omitempty" avro:"comment"`
	SubModel         *selectSubModelMock   `json:"sub_model,omitempty" avro:"sub_model"`
	DateStarted      *time.Time            `json:"date_started,omitempty" avro:"date_started"`
	SubModels        []*selectSubModelMock `json:"sub_models,omitempty" avro:"sub_models"`
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

type resolverTestSuite struct {
	suite.Suite
	model             interface{}
	rvr               *resolver.Resolver
	fields            []string
	structs           []string
	slices            []string
	filters           []resolver.Filter
	fieldsContain     []string
	fieldsNotContain  []string
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

func (s *resolverTestSuite) TestGetFieldsSql() {
	sql := s.rvr.GetFieldsSql(s.filters...)

	for _, statement := range s.fieldsContain {
		s.Assert().Contains(sql, statement)
	}

	for _, statement := range s.fieldsNotContain {
		s.Assert().NotContains(sql, statement)
	}
}

func (s *resolverTestSuite) TestGetStructsSql() {
	sql := s.rvr.GetSlicesSql(s.filters...)

	for _, statement := range s.slicesContain {
		s.Assert().Contains(sql, statement)
	}

	for _, statement := range s.slicesNotContain {
		s.Assert().NotContains(sql, statement)
	}
}

func (s *resolverTestSuite) TestGetSlicesSql() {
	sql := s.rvr.GetStructsSql(s.filters...)

	for _, statement := range s.structsContain {
		s.Assert().Contains(sql, statement)
	}

	for _, statement := range s.structsNotContain {
		s.Assert().NotContains(sql, statement)
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
				") as model_sub_model",
			},
			slices: []string{
				"sub_models",
			},
			slicesContain: []string{
				"TRANSFORM(model_sub_models, (sub_models) => ",
				"STRUCT(",
				"identifier := sub_models->identifier",
				"in_language := sub_models->in_language",
				"event := sub_models->event",
				")) as model_sub_models",
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
				"TRANSFORM(model_sub_models, (sub_models) => ",
				"STRUCT(",
				"identifier := sub_models->identifier",
				")) as model_sub_models",
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

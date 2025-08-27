package articles_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	articles "wikimedia-enterprise/api/realtime/handlers/v2/articles"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/schema"
)

type modelTestSuite struct {
	suite.Suite
	maxParts   int
	partitions int
	res        *resolver.Resolver
	model      *articles.Model
	gcx        *gin.Context
	err        string
}

func (s *modelTestSuite) SetupSuite() {
	s.maxParts = 10
	s.partitions = 50

	var err error
	s.res, err = resolver.New(new(schema.Article))
	s.Assert().NoError(err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	s.gcx, _ = gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/dummy", nil)
	s.gcx.Request = req
}

func (s *modelTestSuite) TestNewModel() {
	m := articles.NewModel(s.maxParts, s.partitions)
	s.Assert().NotNil(m)
}

func (s *modelTestSuite) TestSetParse() {
	err := s.model.Parse(s.gcx, s.res, s.maxParts, s.partitions)
	if len(s.err) > 0 {
		s.Assert().Contains(err.Error(), s.err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestModel(t *testing.T) {
	filter := resolver.RequestFilter{
		Field: "is_part_of.identifier",
		Value: "enwiki",
	}

	invalidFilter := []resolver.RequestFilter{}

	for i := 0; i <= 256; i++ {
		invalidFilter = append(invalidFilter, filter)
	}

	for _, tc := range []*modelTestSuite{
		{
			model: &articles.Model{Fields: []string{"name", "version.identifier"},
				Filters: []resolver.RequestFilter{{Field: "is_part_of.identifier", Value: "enwiki"}}},
		},
		{
			model: &articles.Model{Fields: []string{strings.Repeat("a", 256)}},
			err:   "Model.Fields",
		},
		{
			model: &articles.Model{Filters: invalidFilter},
			err:   "Model.Filters",
		},
		{
			model: &articles.Model{Fields: []string{"not_found"}},
			err:   "not found",
		},
		{
			model: &articles.Model{Fields: []string{"version"}},
			err:   "use wildcard",
		},
		{
			model: &articles.Model{Filters: []resolver.RequestFilter{{Field: "is_part_of", Value: "enwiki"}}},
			err:   "can't filter",
		},
		{
			model: &articles.Model{Filters: []resolver.RequestFilter{{Field: "is_par_of.identifier", Value: ""}}},
			err:   "no such field",
		},
		{
			model: &articles.Model{Parts: []int{1, 1, 1, 1, 1, 1, 1, 11, 1, 1}},
			err:   "part value",
		},
		{
			model: &articles.Model{Parts: []int{0}, Since: time.Now()},
			err:   "specify either offsets or time-offsets",
		},
		{
			model: &articles.Model{Parts: []int{0}, SincePerPartition: map[int]time.Time{0: time.Now()}, Offsets: map[int]int64{0: 0}},
			err:   "not both",
		},
		{
			model: &articles.Model{Parts: []int{0}, SincePerPartition: map[int]time.Time{0: time.Now().Add(1 * time.Minute)}},
			err:   "future",
		},
		{
			model: &articles.Model{Since: time.Now().Add(1 * time.Minute)},
			err:   "future",
		},
	} {
		suite.Run(t, tc)
	}
}

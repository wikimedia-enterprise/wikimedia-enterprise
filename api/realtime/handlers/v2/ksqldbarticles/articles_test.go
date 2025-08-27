package ksqldbarticles_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"wikimedia-enterprise/api/realtime/config/env"
	articles "wikimedia-enterprise/api/realtime/handlers/v2/ksqldbarticles"
	"wikimedia-enterprise/api/realtime/libraries/resolver"
	"wikimedia-enterprise/api/realtime/submodules/ksqldb"
	"wikimedia-enterprise/api/realtime/submodules/schema"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ksqldbMock struct {
	mock.Mock
	ksqldb.PushPuller

	Channel     chan any
	returnValue error
}

func (s *ksqldbMock) Push(_ context.Context, req *ksqldb.QueryRequest, callback func(qrw *ksqldb.HeaderRow, row ksqldb.Row) error) error {
	s.Called(req, callback)
	return s.returnValue
}

type modelTestSuite struct {
	suite.Suite
	mdl *articles.Model
}

func (s *modelTestSuite) SetupSuite() {
	s.mdl = new(articles.Model)
	s.mdl.Parts = []int{0, 1}
	s.mdl.SetPartitions(50, 10)
}

func (s *modelTestSuite) TestGetPartitions() {
	s.NotEmpty(s.mdl.GetPartitions())
	s.Equal(
		[]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		s.mdl.GetPartitions())
}

func (s *modelTestSuite) TestPartitionsContain() {
	s.True(s.mdl.PartitionsContain(0))
	s.True(s.mdl.PartitionsContain(1))
	s.True(s.mdl.PartitionsContain(2))
	s.True(s.mdl.PartitionsContain(4))
	s.True(s.mdl.PartitionsContain(9))
}

func (s *modelTestSuite) TestGetFormattedPartitions() {
	s.Equal("0,1,2,3,4,5,6,7,8,9", s.mdl.GetFormattedPartitions())
}

func TestModel(t *testing.T) {
	suite.Run(t, new(modelTestSuite))
}

type articlesTestSuite struct {
	suite.Suite
	pms              *articles.Parameters
	ctx              context.Context
	srv              *httptest.Server
	pld              string
	sts              int
	ksqldb           *ksqldbMock
	ksqldbResults    []map[string]any
	expectedArticles []schema.Article
	queryPredicate   func(string) error
	ker              error
	ksqldbBlocker    chan struct{}
}

func (s *articlesTestSuite) createServer() http.Handler {
	gin.SetMode(gin.TestMode)
	rtr := gin.New()

	rtr.POST("/articles", articles.NewHandler(s.ctx, s.pms))

	return rtr
}

func getOrDefault(m map[string]any, key string, def any) any {
	if val, ok := m[key]; ok {
		return val
	}
	return def
}

func SendRow(mock *ksqldbMock, row map[string]any) error {
	cb := mock.Calls[len(mock.Calls)-1].Arguments[1].(func(*ksqldb.HeaderRow, ksqldb.Row) error)

	rowPartition := getOrDefault(row, "ROWPARTITION", 1)
	rowOffset := getOrDefault(row, "ROWOFFSET", 2.0)
	rowTime := getOrDefault(row, "ROWTIME", 123456789.0)

	columns := []string{}
	values := []any{}
	delete(row, "ROWPARTITION")
	delete(row, "ROWOFFSET")
	delete(row, "ROWTIME")
	for c, v := range row {
		columns = append(columns, c)
		values = append(values, v)
	}

	// ROWPARTITION, ROWOFFSET, ROWTIME always come at the end, in this order.
	columns = append(columns, "ROWPARTITION", "ROWOFFSET", "ROWTIME")
	values = append(values, rowPartition, rowOffset, rowTime)

	// Missing columns are ignored, so we can provide just what's convenient for the test.
	return cb(&ksqldb.HeaderRow{
		ColumnNames: columns,
	}, values)
}

func (s *articlesTestSuite) SetupTest() {
	rvr, err := resolver.NewResolvers(map[string]interface{}{
		schema.KeyTypeArticle: new(schema.Article),
	})
	s.Assert().NoError(err)

	s.ctx = context.Background()

	s.ksqldbBlocker = make(chan struct{})
	s.ksqldb = new(ksqldbMock)
	s.ksqldb.returnValue = s.ker
	s.ksqldb.On("Push", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		if s.ksqldb.returnValue != nil {
			return
		}

		// We need to respond to the KSQLDB callback immediately, otherwise the HTTP client won't unblock in the test.
		err = SendRow(s.ksqldb, map[string]any{"name": "Josephine_Baker", "event": map[string]any{}})
		s.NoError(err)

		// Hang until done "returning results", otherwise we'll close the HTTP connection before we finish the test.
		<-s.ksqldbBlocker
	})

	s.pms = &articles.Parameters{
		Resolvers: rvr,
		KSQLDB:    s.ksqldb,
		Env: &env.Environment{
			MaxParts:       2,
			Partitions:     4,
			ArticlesStream: "KSQLDB_article_stream",
		},
	}

	s.srv = httptest.NewServer(s.createServer())
}

func (s *articlesTestSuite) TearDownTest() {
	s.srv.Close()
}

type Reader struct {
	scanner  *bufio.Scanner
	articles chan *schema.Article
	quit     chan struct{}
}

func (r *Reader) StartRead() {
	go func() {
		for r.scanner.Scan() {
			text := r.scanner.Text()
			if len(strings.TrimSpace(text)) == 0 {
				continue
			}
			art := schema.Article{}
			err := json.Unmarshal([]byte(text), &art)
			if err != nil {
				log.Panic(err)
			}
			r.articles <- &art
		}
	}()
}

func (s *articlesTestSuite) TestNewHandler() {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/articles", s.srv.URL), strings.NewReader(s.pld))
	s.Assert().NoError(err)
	client := &http.Client{}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	res, err := client.Do(req)
	s.Assert().NoError(err)
	defer res.Body.Close()

	if s.queryPredicate != nil {
		req := s.ksqldb.Calls[0].Arguments[0].(*ksqldb.QueryRequest)
		s.NoError(s.queryPredicate(req.SQL))
	}

	if s.ksqldb.returnValue != nil {
		s.Assert().Equal(s.sts, res.StatusCode)
		dta, err := io.ReadAll(res.Body)
		s.Assert().NoError(err)
		s.Assert().Contains(string(dta), s.ksqldb.returnValue.Error())

		s.Empty(s.ksqldbResults, "no rows may be returned when there's KSQLDB errors")
		s.Empty(s.expectedArticles, "no articles may be expected when there's KSQLDB errors")
		return
	} else {
		s.Assert().Equal(s.sts, res.StatusCode)
		if s.sts != http.StatusOK {
			// ksqldbResults may still be valid, since we may fail while processing them.
			s.Empty(s.expectedArticles, "no articles may be expected when errors are returned")
			return
		}
	}

	reader := Reader{scanner: bufio.NewScanner(res.Body), articles: make(chan *schema.Article), quit: make(chan struct{})}
	reader.StartRead()

	// Sentinel value, always sent in this test.
	article := <-reader.articles
	s.Equal("Josephine_Baker", article.Name)

	for _, result := range s.ksqldbResults {
		err = SendRow(s.ksqldb, result)
		s.NoError(err)
	}

	getEventPublished := func(art schema.Article) int64 {
		if art.Event == nil {
			return 0
		}
		return art.Event.DatePublished.UnixMilli()
	}
	for _, expected := range s.expectedArticles {
		article = <-reader.articles
		// Compare event.date_published separately, as there are some inconsistencies when running on the pipeline.
		s.Equal(getEventPublished(expected), getEventPublished(*article))
		if expected.Event != nil {
			expected.Event.DatePublished = nil
		}
		if (*article).Event != nil {
			(*article.Event).DatePublished = nil
		}
		s.Equal(expected, *article)
	}

	// No more lines expected
	select {
	case article = <-reader.articles:
		s.Fail("Did not expect more articles from realtime", "Got: %s", article)
	default:
		// All good
	}

	// Release KSQLDB query, if it's still open.
	close(s.ksqldbBlocker)
}

func TestNewHandler(t *testing.T) {
	// Metadata returned by ksqldb.
	eventPartition := 1
	var eventOffset int64 = 2
	eventPublished := time.UnixMilli(123456789.0)

	for _, testCase := range []*articlesTestSuite{
		{
			sts: http.StatusOK,
			// Important: sub-structs must have type map[string]any or they won't be decoded.
			ksqldbResults: []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"identifier": "my-identifier"}}},
			expectedArticles: []schema.Article{{Name: "Albert_Einstein",
				Event: &schema.Event{Identifier: "my-identifier", Partition: &eventPartition, Offset: &eventOffset, DatePublished: &eventPublished}}},
		},
		{
			// If Event is not requested, it shouldn't be in the output.
			// This is a special case. Other fields are filtered out by ksqldb, but we always request Event
			// to discard internal events.
			sts:              http.StatusOK,
			pld:              `{"fields":["name"]}`,
			ksqldbResults:    []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"identifier": "my-identifier"}}},
			expectedArticles: []schema.Article{{Name: "Albert_Einstein"}},
			queryPredicate: func(query string) error {
				fieldsString := strings.Split(strings.TrimPrefix(query, "SELECT "), " FROM ")[0]

				if !strings.Contains(fieldsString, " event") {
					return fmt.Errorf("'event' must always be requested, query: %s", query)
				}

				return nil
			},
		},
		{
			sts:           http.StatusOK,
			pld:           `{"fields":["name", "event.*"]}`,
			ksqldbResults: []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"identifier": "my-identifier"}}},
			expectedArticles: []schema.Article{{Name: "Albert_Einstein",
				Event: &schema.Event{Identifier: "my-identifier", Partition: &eventPartition, Offset: &eventOffset, DatePublished: &eventPublished}}},
		},
		{
			sts:              http.StatusOK,
			ksqldbResults:    []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"is_internal_message": true}}},
			expectedArticles: []schema.Article{},
		},
		{
			sts:              http.StatusOK,
			pld:              `{"fields":["name"]}`,
			ksqldbResults:    []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"is_internal_message": true}}},
			expectedArticles: []schema.Article{},
		},
		{
			sts:              http.StatusOK,
			pld:              `{"fields":["name", "event.*"]}`,
			ksqldbResults:    []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"is_internal_message": true}}},
			expectedArticles: []schema.Article{},
		},
		{
			sts:              http.StatusOK,
			pld:              `{"fields":["name", "event.identifier"]}`,
			ksqldbResults:    []map[string]any{{"name": "Albert_Einstein", "event": map[string]any{"is_internal_message": true}}},
			expectedArticles: []schema.Article{},
		},
		{
			sts: http.StatusInternalServerError,
			ker: errors.New("ksqldb not available"),
		},
		{
			pld: `{"since":"string"}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"fields":["string"]}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"filters":[{"field":"string","value":1}]}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"filters":[{"field":"name","value":1}]}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"filters":[{"field":"identifier","value":1}]}`,
			sts: http.StatusOK,
		},
		{
			pld: `{"parts":[0,1], "offsets":{"0":100, "1":101, "2":100, "3":50}}`,
			sts: http.StatusOK,
		},
		{
			pld: `{"parts":[1], "since_per_partition":{"0":"2023-06-19T19:43:44.52Z"}}`,
			sts: http.StatusOK,
		},
		{
			pld: `{"parts":[1], "since":"2023-06-19T19:43:44.52Z"}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0], "since_per_partition":{"0":"2023-06-19T19:43:44.52Z"}, "offsets":{"0":100, "1":101, "2":100, "3":50}}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0,2], "offsets":{"0":100, "1":101, "2":100, "3":50}}`,
			sts: http.StatusUnprocessableEntity,
		},
		{
			pld: `{"parts":[0,1], "offsets":{"0":100, "1":101, "2":100, "3":50, "4":50}}`,
			sts: http.StatusUnprocessableEntity,
		},
	} {
		suite.Run(t, testCase)
	}
}

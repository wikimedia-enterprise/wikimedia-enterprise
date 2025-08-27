package ksqldb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func createServer(sql string, hr string, rows []string, sleep int) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/query", func(rw http.ResponseWriter, r *http.Request) {
		req := new(QueryRequest)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.SQL != sql {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		fl := rw.(http.Flusher)

		_, _ = rw.Write([]byte(fmt.Sprintf("%s\n", hr)))
		fl.Flush()

		for _, row := range rows {
			if r.Context().Err() != context.Canceled {
				time.Sleep(time.Millisecond * time.Duration(sleep))
				_, _ = rw.Write([]byte(fmt.Sprintf("%s\n", row)))
				fl.Flush()
			}
		}
	})

	return router
}

type ksqldbTestSuite struct {
	suite.Suite
	ctx     context.Context
	sql     string
	queryId string
	hr      string
	rows    []string
}

// SetupTest.
func (s *ksqldbTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.sql = "SELECT * FROM articles;"
	s.queryId = "aee9f5fe-e98f-4bd2-876c-9ca09eefbf28"
	s.hr = fmt.Sprintf(`{"queryId":"%s","columnNames":["ROWKEY","NAME","IDENTIFIER","DATE_MODIFIED","VERSION","URL"],"columnTypes":["STRUCT<IDENTIFIER STRING, TYPE STRING>","STRING","INTEGER","BIGINT","STRUCT<IDENTIFIER INTEGER, PARENT_IDENTIFIER INTEGER, COMMENT STRING>","STRING"]}`, s.queryId)
	s.rows = []string{
		`[{"IDENTIFIER":"/articles/enwiki/Wong_(Marvel_Cinematic_Universe)","TYPE":"version"},"Wong_(Marvel_Cinematic_Universe)",68044232,1636371202000000,{"IDENTIFIER":1054155861,"PARENT_IDENTIFIER":0,"COMMENT":""},"https://en.wikipedia.org/wiki/Wong_(Marvel_Cinematic_Universe)"]`,
		`[{"IDENTIFIER":"/articles/enwiki/Princess_Elizabeth_Land","TYPE":"version"},"Princess_Elizabeth_Land",2774809,1636371202000000,{"IDENTIFIER":1054155862,"PARENT_IDENTIFIER":0,"COMMENT":""},"https://en.wikipedia.org/wiki/Princess_Elizabeth_Land"]`,
		`[{"IDENTIFIER":"/articles/enwiki/Pixie_(film)","TYPE":"version"},"Pixie_(film)",61555016,1636371200000000,{"IDENTIFIER":1054155860,"PARENT_IDENTIFIER":0,"COMMENT":""},"https://en.wikipedia.org/wiki/Pixie_(film)"]`,
		`[{"IDENTIFIER":"/articles/enwiki/Talk:Bayside_(album)","TYPE":"version"},"Talk:Bayside_(album)",4509726,1636371204000000,{"IDENTIFIER":1054155863,"PARENT_IDENTIFIER":0,"COMMENT":""},"https://en.wikipedia.org/wiki/Talk:Bayside_(album)"]`,
		`[{"IDENTIFIER":"/articles/enwiki/Template:Mark","TYPE":"version"},"Template:Mark",25695872,1636371204000000,{"IDENTIFIER":1054155864,"PARENT_IDENTIFIER":0,"COMMENT":""},"https://en.wikipedia.org/wiki/Template:Mark"]`,
		`[{"IDENTIFIER":"/articles/enwiki/Portal:Bangladesh/Did_you_know","TYPE":"version"},"Portal:Bangladesh/Did_you_know",2066721,1636371201000000,{"IDENTIFIER":1054155865,"PARENT_IDENTIFIER":0,"COMMENT":""},"https://en.wikipedia.org/wiki/Portal:Bangladesh/Did_you_know"]`,
	}
}

// TestPush performes the Push method testing.
func (s *ksqldbTestSuite) TestPush() {
	srv := httptest.NewServer(createServer(s.sql, s.hr, s.rows, 0))
	defer srv.Close()

	client := NewClient(srv.URL, func(c *Client) {
		c.HTTPClient = &http.Client{}
	})

	var header *HeaderRow
	idx := 0
	err := client.Push(s.ctx, &QueryRequest{SQL: s.sql}, func(hr *HeaderRow, row Row) error {
		header = hr
		s.Assert().Contains(s.rows[idx], row.String(2))
		idx++
		return nil
	})

	s.Assert().NoError(err)
	s.Assert().NotNil(header)
}

// TestPushCtx performes the Push method cancelation test by timeout.
func (s *ksqldbTestSuite) TestPushCtx() {
	srv := httptest.NewServer(createServer(s.sql, s.hr, s.rows, 500))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(s.ctx, time.Millisecond*500)
	defer cancel()

	client := NewClient(srv.URL, func(c *Client) {
		c.HTTPClient = &http.Client{}
	})

	err := client.Push(ctx, &QueryRequest{SQL: s.sql}, func(hr *HeaderRow, row Row) error { return nil })
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), context.DeadlineExceeded.Error())
}

// TestPull performes Pull method testing.
func (s *ksqldbTestSuite) TestPull() {
	srv := httptest.NewServer(createServer(s.sql, s.hr, s.rows, 0))
	defer srv.Close()

	client := NewClient(srv.URL, func(c *Client) {
		c.HTTPClient = &http.Client{}
	})

	hr, rows, err := client.Pull(s.ctx, &QueryRequest{SQL: s.sql})

	s.Assert().NoError(err)
	s.Assert().NotNil(hr)
	s.Assert().NotEmpty(rows)

	for idx, row := range rows {
		s.Assert().Contains(s.rows[idx], row.String(2))
	}
}

// TestPull performes Pull method cancelation testing by timeout.
func (s *ksqldbTestSuite) TestPullCtx() {
	srv := httptest.NewServer(createServer(s.sql, s.hr, s.rows, 500))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(s.ctx, time.Millisecond*500)
	defer cancel()

	client := NewClient(srv.URL, func(c *Client) {
		c.HTTPClient = &http.Client{}
	})

	_, _, err := client.Pull(ctx, &QueryRequest{SQL: s.sql})
	s.Assert().Error(err)
	s.Assert().Equal(context.DeadlineExceeded, err)
}

func TestKqsqlDBClient(t *testing.T) {
	suite.Run(t, new(ksqldbTestSuite))
}

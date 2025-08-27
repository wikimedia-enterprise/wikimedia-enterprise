// Package ksqldb provides the bare-bones integration with ksqlDB.
// It supports both pull and push queries.
package ksqldb

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/http2"
)

const (

	// Default response content type for pull & push queries
	// In the case of a successful query, if the content type is application/vnd.ksqlapi.delimited.v1,
	// the results are returned as a header JSON object followed by zero or more JSON arrays that are delimited by newlines.
	ContentTypeDelim = "application/vnd.ksqlapi.delimited.v1; charset=utf-8"

	// Default serialization format for requests and responses.
	ContentTypeDefault = "application/vnd.ksql.v1+json; charset=utf-8"

	// EndpointRunStreamQuery is used to run push and pull queries.
	// These endpoints are only available when using HTTP 2.
	EndpointRunStreamQuery string = "query"
	// EndpointCloseQuery used to terminates a push query.
	EndpointCloseQuery string = "close-query"
	// CloseQueryTimeout is the default timeout for closing a query.
	CloseQueryTimeout = 5 * time.Second
)

// Pusher is an interface to wrap default ksqldb Push method for unit testing.
type Pusher interface {
	Push(ctx context.Context, q *QueryRequest, cb func(qr *HeaderRow, row Row) error) error
}

// Puller is an interface to wrap default ksqld Pull method for unit testing.
type Puller interface {
	Pull(ctx context.Context, q *QueryRequest) (*HeaderRow, []Row, error)
}

// PusherPuller wraps entire client into single interface for unit testing.
type PushPuller interface {
	Pusher
	Puller
}

// BasicAuth struct to pass authentication to ksqlDB.
type BasicAuth struct {
	Username string
	Password string
}

// QueryRequest payload for database request.
type QueryRequest struct {
	SQL        string            `json:"ksql"`
	Properties map[string]string `json:"streamsProperties"`
}

// HeaderRow leading row of the query response that contains information about columns, types and query identifier.
type HeaderRow struct {
	QueryID     string   `json:"queryID"`
	ColumnNames []string `json:"columnNames"`
	ColumnTypes []string `json:"columnTypes"`
}

// NewClient creates a new ksqlDB client.
func NewClient(url string, options ...func(*Client)) *Client {
	client := &Client{
		url:             url,
		MessageMaxBytes: 20971520,
		HTTPClient: &http.Client{
			// In go, the standard http.Client is used for HTTP/2 requests as well.
			// The only difference is the usage of http2.Transport instead of http.Transport in the client’s Transport field
			Transport: &http2.Transport{
				AllowHTTP: true,
				// Pretend we are dialing a TLS endpoint.
				// Note, we ignore the passed tls.Config
				DialTLS: func(network string, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
	}

	if strings.HasPrefix(url, "https://") {
		client.HTTPClient = &http.Client{}
	}

	for _, opt := range options {
		opt(client)
	}

	return client
}

// Client ksqlDB database query client.
type Client struct {
	url             string
	HTTPClient      *http.Client
	BasicAuth       *BasicAuth
	MessageMaxBytes int
}

func (c *Client) req(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)

	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s", c.url, endpoint), bytes.NewBuffer(body))

	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	switch endpoint {
	case EndpointRunStreamQuery:
		req.Header.Set("Content-Type", ContentTypeDelim)
		req.Header.Set("Accept-Encoding", "identity")
		req.Header.Set("Accept", ContentTypeDelim)
	default:
		req.Header.Set("Content-Type", ContentTypeDefault)
		req.Header.Set("Accept-Encoding", "identity")
		req.Header.Set("Accept", ContentTypeDefault)
	}

	if c.BasicAuth != nil {
		req.SetBasicAuth(c.BasicAuth.Username, c.BasicAuth.Password)
	}

	res, err := c.HTTPClient.Do(req)

	if err != nil {
		return res, errors.Wrap(err, "failed to do request")
	}

	if res.StatusCode != http.StatusOK {
		data, err := io.ReadAll(res.Body)

		if err != nil {
			return res, errors.Wrap(err, "failed to read response body")
		}

		return res, fmt.Errorf("%s:%s", http.StatusText(res.StatusCode), string(data))
	}

	return res, nil
}

// Pull query to pull data from ksqlDB, returns list of rows that was retrieved from ksqldb and leading row with metadata.
func (c *Client) Pull(ctx context.Context, q *QueryRequest) (*HeaderRow, []Row, error) {
	res, err := c.req(ctx, EndpointRunStreamQuery, q)
	hr := new(HeaderRow)
	rows := []Row{}

	if err != nil {
		return hr, rows, err
	}

	defer res.Body.Close()

	scn := bufio.NewScanner(res.Body)
	scn.Buffer([]byte{}, c.MessageMaxBytes)

	for scn.Scan() {
		if len(hr.ColumnNames) <= 0 {
			if err := json.Unmarshal([]byte(scn.Text()), hr); err != nil {
				log.Printf("failed to unmarshal header row: %v\n", err)
			}
			continue
		}

		row := Row{}

		if err := json.Unmarshal([]byte(scn.Text()), &row); err != nil {
			log.Printf("failed to unmarshal row: %v\n", err)
		} else {
			rows = append(rows, row)
		}
	}

	if err := scn.Err(); err != io.EOF && err != nil {
		return hr, rows, err
	}

	return hr, rows, nil
}

// Push query to receive streaming data in real time, every time new record gets returned call back
// function gets called with header row and the actual row.
func (c *Client) Push(ctx context.Context, q *QueryRequest, cb func(qr *HeaderRow, row Row) error) error {
	res, err := c.req(ctx, EndpointRunStreamQuery, q)

	if err != nil {
		return err
	}

	closeQueryDeferred := false
	defer res.Body.Close()

	hr := new(HeaderRow)
	scn := bufio.NewScanner(res.Body)
	scn.Buffer([]byte{}, c.MessageMaxBytes)

	for scn.Scan() {
		if len(hr.ColumnNames) <= 0 {
			if err := json.Unmarshal([]byte(scn.Text()), hr); err != nil {
				log.Printf("failed to unmarshal header row: %v\n", err)
			}

			if !closeQueryDeferred && hr.QueryID != "" {
				queryID := hr.QueryID
				defer func() {
					var closeCtx context.Context
					var cancel context.CancelFunc

					// if the ctx context is already done or cancelled, use a new context with timeout
					if ctx.Err() != nil {
						closeCtx, cancel = context.WithTimeout(context.Background(), CloseQueryTimeout)
					} else {
						closeCtx, cancel = context.WithTimeout(ctx, CloseQueryTimeout)
					}
					defer cancel()

					if err := c.CloseQuery(closeCtx, queryID); err != nil {
						log.Printf("failed to close query %s: %v", queryID, err)
					}
				}()

				closeQueryDeferred = true
			}

			continue
		}

		row := Row{}

		if err := json.Unmarshal([]byte(scn.Text()), &row); err != nil {
			log.Printf("failed to unmarshal row: %v\n", err)
			continue
		}

		if err := cb(hr, row); err != nil {
			return errors.Wrap(err, "failed to call callback function")
		}
	}

	if err := scn.Err(); err != io.EOF && err != nil {
		return errors.Wrap(err, "failed to scan response body, non-EOF error")
	}

	return nil
}

// CloseQuery terminates a query initiated by yhe push method.
func (c *Client) CloseQuery(ctx context.Context, queryId string) error {
	_, err := c.req(ctx, EndpointCloseQuery, map[string]string{"queryId": queryId})
	return err
}

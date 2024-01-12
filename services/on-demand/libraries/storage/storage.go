// Package storage works as a encapsulation for the logic of updating s3 storage.
package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/on-demand/config/env"

	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"go.uber.org/dig"
)

// New creates new instance of the storage for dependency injection.
func New(p Storage) Updater {
	return &p
}

// Updater interface that hides implementation for unit testing.
type Updater interface {
	Update(ctx context.Context, kdt []byte, etp string, v interface{}) error
}

// Storage helper that consumes messages from kafka and puts them into s3.
type Storage struct {
	dig.In
	Stream schema.UnmarshalProducer
	Env    *env.Environment
	S3     s3iface.S3API
}

// Update this method will update or delete the specific key using the event type.
// Uses s3API and PUT and DELETE methods to do so.
func (s *Storage) Update(ctx context.Context, kdt []byte, etp string, val interface{}) error {
	key := new(schema.Key)

	if err := s.Stream.Unmarshal(ctx, kdt, key); err != nil {
		return err
	}

	dta, err := json.Marshal(val)

	if err != nil {
		return err
	}

	loc := fmt.Sprintf("%s%s.json", key.Type, key.Identifier)

	switch etp {
	case schema.EventTypeCreate, schema.EventTypeUpdate:
		pin := &s3.PutObjectInput{
			Bucket:             aws.String(s.Env.AWSBucket),
			Key:                aws.String(loc),
			Body:               bytes.NewReader(dta),
			ContentType:        aws.String("application/json"),
			ContentLength:      aws.Int64(int64(len(dta))),
			ContentDisposition: aws.String("attachment"),
		}

		pof := func() (err error) {
			_, err = s.S3.PutObjectWithContext(ctx, pin)
			return err
		}

		if err := retry.Do(pof); err != nil {
			return err.(retry.Error)[0]
		}
	case schema.EventTypeDelete:
		pin := &s3.DeleteObjectInput{
			Bucket: aws.String(s.Env.AWSBucket),
			Key:    aws.String(loc),
		}

		dof := func() (err error) {
			_, err = s.S3.DeleteObjectWithContext(ctx, pin)
			return err
		}

		if err := retry.Do(dof); err != nil {
			return err.(retry.Error)[0]
		}
	}

	return nil
}

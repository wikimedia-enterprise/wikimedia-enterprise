package storage_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"wikimedia-enterprise/general/schema"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/libraries/storage"

	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type s3APIMock struct {
	mock.Mock
	s3iface.S3API
}

func (s *s3APIMock) PutObjectWithContext(_ aws.Context, inp *s3.PutObjectInput, _ ...request.Option) (*s3.PutObjectOutput, error) {
	return nil, s.
		Called(*inp.Bucket, *inp.Key).
		Error(0)
}

func (s *s3APIMock) DeleteObjectWithContext(_ aws.Context, inp *s3.DeleteObjectInput, _ ...request.Option) (*s3.DeleteObjectOutput, error) {
	return nil, s.
		Called(*inp.Bucket, *inp.Key).
		Error(0)
}

type streamMock struct {
	mock.Mock
	schema.UnmarshalProducer
}

func (s *streamMock) Unmarshal(_ context.Context, data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}

	return s.
		Called(data, v).
		Error(0)
}

type storageTestSuite struct {
	suite.Suite
	sto *storage.Storage
	env *env.Environment
	str *streamMock
	s3a *s3APIMock
	ctx context.Context
	key *schema.Key
	kdt []byte
	etp string
	val interface{}
	uer error
	der error
	ser error
}

func (s *storageTestSuite) SetupSuite() {
	s.env = &env.Environment{
		AWSBucket: "bucket",
	}
	s.key = &schema.Key{
		Identifier: "/enwiki/Earth",
		Type:       "articles",
	}
	s.val = &schema.Article{
		Name: "Earth",
	}

	s.kdt, _ = json.Marshal(s.key)

	s.str = new(streamMock)
	s.str.On("Unmarshal", s.kdt, s.key).Return(s.ser)

	loc := fmt.Sprintf("%s%s.json", s.key.Type, s.key.Identifier)
	s.s3a = new(s3APIMock)
	s.s3a.On("PutObjectWithContext", s.env.AWSBucket, loc).Return(s.uer)
	s.s3a.On("DeleteObjectWithContext", s.env.AWSBucket, loc).Return(s.der)

	s.ctx = context.Background()
	s.sto = &storage.Storage{
		Stream: s.str,
		Env:    s.env,
		S3:     s.s3a,
	}
}

func (s *storageTestSuite) TestUpdate() {
	err := s.sto.Update(s.ctx, s.kdt, s.etp, s.val)

	if s.ser != nil {
		s.Assert().Equal(s.ser, err)
	} else if s.uer != nil {
		s.Assert().Equal(s.uer, err)
	} else if s.der != nil {
		s.Assert().Equal(s.der, err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestStorage(t *testing.T) {
	retry.DefaultAttempts = 1 // for testing

	for _, testCase := range []*storageTestSuite{
		{
			etp: schema.EventTypeUpdate,
			ser: errors.New("unmarshal error"),
		},
		{
			etp: schema.EventTypeUpdate,
			uer: nil,
		},
		{
			etp: schema.EventTypeCreate,
			uer: nil,
		},
		{
			uer: errors.New("update error"),
			etp: schema.EventTypeUpdate,
		},
		{
			uer: errors.New("create error"),
			etp: schema.EventTypeCreate,
		},
		{
			der: nil,
			etp: schema.EventTypeDelete,
		},
		{
			der: errors.New("delete error"),
			etp: schema.EventTypeDelete,
		},
	} {
		suite.Run(t, testCase)
	}
}

package storage_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"wikimedia-enterprise/services/on-demand/config/env"
	"wikimedia-enterprise/services/on-demand/libraries/storage"
	"wikimedia-enterprise/services/on-demand/submodules/schema"

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

	keyTypeSuffix      string
	expectedPutPath    string
	expectedDeletePath string
	useHashedPrefixes  bool
}

func (s *storageTestSuite) SetupTest() {
	s.env = &env.Environment{
		AWSBucket:            "bucket",
		ArticleKeyTypeSuffix: s.keyTypeSuffix,
		UseHashedPrefixes:    s.useHashedPrefixes,
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

	s.s3a = new(s3APIMock)
	if s.expectedPutPath != "" {
		s.s3a.On("PutObjectWithContext", s.env.AWSBucket, s.expectedPutPath).Return(s.uer)
	}
	if s.expectedDeletePath != "" {
		s.s3a.On("DeleteObjectWithContext", s.env.AWSBucket, s.expectedDeletePath).Return(s.der)
	}

	s.ctx = context.Background()
	s.sto = &storage.Storage{
		Stream: s.str,
		Env:    s.env,
		S3:     s.s3a,
	}
}

func (s *storageTestSuite) TearDownTest() {
	s.s3a.AssertExpectations(s.T())
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
			etp:             schema.EventTypeUpdate,
			uer:             nil,
			expectedPutPath: "articles/enwiki/Earth.json",
		},
		{
			etp:             schema.EventTypeCreate,
			uer:             nil,
			expectedPutPath: "articles/enwiki/Earth.json",
		},
		{
			uer:             errors.New("update error"),
			etp:             schema.EventTypeUpdate,
			expectedPutPath: "articles/enwiki/Earth.json",
		},
		{
			uer:             errors.New("create error"),
			etp:             schema.EventTypeCreate,
			expectedPutPath: "articles/enwiki/Earth.json",
		},
		{
			der:                nil,
			etp:                schema.EventTypeDelete,
			keyTypeSuffix:      "v2",
			expectedDeletePath: "articles_v2/enwiki/Earth.json",
		},
		{
			der:                errors.New("delete error"),
			etp:                schema.EventTypeDelete,
			expectedDeletePath: "articles/enwiki/Earth.json",
		},
		{
			useHashedPrefixes: true,
			uer:               errors.New("create error"),
			etp:               schema.EventTypeCreate,
			expectedPutPath:   "articles/enwiki/5/5c/Earth.json",
		},
		{
			useHashedPrefixes:  true,
			der:                nil,
			etp:                schema.EventTypeDelete,
			keyTypeSuffix:      "v2",
			expectedDeletePath: "articles_v2/enwiki/5/5c/Earth.json",
		},
	} {
		suite.Run(t, testCase)
	}
}

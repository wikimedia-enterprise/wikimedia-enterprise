package s3tracerproxy

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	s3API "github.com/aws/aws-sdk-go/service/s3/s3iface"

	trc "wikimedia-enterprise/general/tracing"
)

// S3TracerProxy is an interface that defines the methods that are to be traced for s3.
type S3TracerProxy interface {
	GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
	DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error)
	PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error)
	HeadObjectWithContext(context.Context, *s3.HeadObjectInput, ...request.Option) (*s3.HeadObjectOutput, error)
	CompleteMultipartUploadWithContext(context.Context, *s3.CompleteMultipartUploadInput, ...request.Option) (*s3.CompleteMultipartUploadOutput, error)
	UploadPartCopyWithContext(context.Context, *s3.UploadPartCopyInput, ...request.Option) (*s3.UploadPartCopyOutput, error)
	CreateMultipartUploadWithContext(context.Context, *s3.CreateMultipartUploadInput, ...request.Option) (*s3.CreateMultipartUploadOutput, error)
	CopyObjectWithContext(context.Context, *s3.CopyObjectInput, ...request.Option) (*s3.CopyObjectOutput, error)
	ListObjectsV2PagesWithContext(context.Context, *s3.ListObjectsV2Input, func(*s3.ListObjectsV2Output, bool) bool, ...request.Option) error
}

// StorageTrace is a struct that wraps the S3API interface and the TracerInterface interface.
type StorageTrace struct {
	s3tracerproxy s3API.S3API
	tracer        *trc.Tracer
}

// GetObjectWithContext is a method that wraps the GetObjectWithContext method of the S3API interface.
func (tr *StorageTrace) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "GetObjectWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished GetObjectWithContext")
	}

	return tr.s3tracerproxy.GetObjectWithContext(ctt, input, opts...)
}

// DeleteObjectWithContext is a method that wraps the DeleteObjectWithContext method of the S3API interface.
func (tr *StorageTrace) DeleteObjectWithContext(ctx aws.Context, input *s3.DeleteObjectInput, opts ...request.Option) (*s3.DeleteObjectOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "DeleteObjectWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished DeleteObjectWithContext")
	}

	return tr.s3tracerproxy.DeleteObjectWithContext(ctt, input, opts...)
}

// PutObjectWithContext is a method that wraps the PutObjectWithContext method of the S3API interface.
func (tr *StorageTrace) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "PutObjectWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished PutObjectWithContext")
	}

	return tr.s3tracerproxy.PutObjectWithContext(ctt, input, opts...)
}

// HeadObjectWithContext is a method that wraps the HeadObjectWithContext method of the S3API interface.
func (tr *StorageTrace) HeadObjectWithContext(ctx context.Context, input *s3.HeadObjectInput, opt ...request.Option) (*s3.HeadObjectOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "HeadObjectWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished HeadObjectWithContext")
	}

	return tr.s3tracerproxy.HeadObjectWithContext(ctt, input)
}

// CompleteMultipartUploadWithContext wraps the CompleteMultipartUploadWithContext method of the S3API interface.
func (tr *StorageTrace) CompleteMultipartUploadWithContext(ctx context.Context, input *s3.CompleteMultipartUploadInput, opts ...request.Option) (*s3.CompleteMultipartUploadOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "CompleteMultipartUploadWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished CompleteMultipartUploadWithContext")
	}

	return tr.s3tracerproxy.CompleteMultipartUploadWithContext(ctt, input, opts...)
}

// UploadPartCopyWithContext wraps the UploadPartCopyWithContext method of the S3API interface.
func (tr *StorageTrace) UploadPartCopyWithContext(ctx context.Context, input *s3.UploadPartCopyInput, opts ...request.Option) (*s3.UploadPartCopyOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "UploadPartCopyWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished UploadPartCopyWithContext")
	}

	return tr.s3tracerproxy.UploadPartCopyWithContext(ctt, input, opts...)
}

// CreateMultipartUploadWithContext wraps the CreateMultipartUploadWithContext method of the S3API interface.
func (tr *StorageTrace) CreateMultipartUploadWithContext(ctx context.Context, input *s3.CreateMultipartUploadInput, opts ...request.Option) (*s3.CreateMultipartUploadOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "CreateMultipartUploadWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished CreateMultipartUploadWithContext")
	}

	return tr.s3tracerproxy.CreateMultipartUploadWithContext(ctt, input, opts...)
}

// CopyObjectWithContext wraps the CopyObjectWithContext method of the S3API interface.
func (tr *StorageTrace) CopyObjectWithContext(ctx context.Context, input *s3.CopyObjectInput, opts ...request.Option) (*s3.CopyObjectOutput, error) {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "CopyObjectWithContext", "bucket": *input.Bucket, "key": *input.Key})
		defer end(nil, "finished CopyObjectWithContext")
	}

	return tr.s3tracerproxy.CopyObjectWithContext(ctt, input, opts...)
}

// ListObjectsV2PagesWithContext wraps the ListObjectsV2PagesWithContext method of the S3API interface.
func (tr *StorageTrace) ListObjectsV2PagesWithContext(ctx context.Context, input *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool, opts ...request.Option) error {
	var (
		ctt context.Context
		end func(error, string)
	)

	if tr.tracer != nil {
		end, ctt = (*tr.tracer).Trace(ctx, map[string]string{"method": "ListObjectsV2PagesWithContext", "bucket": *input.Bucket})
		defer end(nil, "finished ListObjectsV2PagesWithContext")
	}

	return tr.s3tracerproxy.ListObjectsV2PagesWithContext(ctt, input, fn, opts...)
}

// NewStorageTrace is a constructor function that returns a new instance of the StorageTrace struct.
func NewStorageTrace(api s3API.S3API, tracer trc.Tracer) S3TracerProxy {
	return &StorageTrace{
		s3tracerproxy: api,
		tracer:        &tracer,
	}
}

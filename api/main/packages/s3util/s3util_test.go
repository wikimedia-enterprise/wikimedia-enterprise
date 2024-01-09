package s3util_test

import (
	"net/http/httptest"
	"testing"
	"time"
	"wikimedia-enterprise/api/main/packages/s3util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type s3Reader struct {
	mock.Mock
	s3.SelectObjectContentEventStreamReader
}

func (s *s3Reader) Events() <-chan s3.SelectObjectContentEventStreamEvent {
	ags := s.Called()

	ch := make(chan s3.SelectObjectContentEventStreamEvent, 1)
	ch <- &s3.RecordsEvent{
		Payload: []byte(ags.String(0)),
	}
	close(ch)

	return ch
}

type eventContentTestSuite struct {
	suite.Suite
	s3r *s3Reader
	pld string
	res string
}

func (s *eventContentTestSuite) SetupSuite() {
	s.s3r = new(s3Reader)
	s.s3r.On("Events").Return(s.pld)
}

func (s *eventContentTestSuite) TestEventContent() {
	evt := s3.NewSelectObjectContentEventStream(func(soces *s3.SelectObjectContentEventStream) {
		soces.Reader = s.s3r
	})

	s.Assert().Equal(s.res, s3util.EventContent(evt))
}

func TestEventContent(t *testing.T) {
	for _, testCase := range []*eventContentTestSuite{
		{
			pld: "test string,",
			res: "test string",
		},
		{
			pld: "test string\n",
			res: "test string",
		},
	} {
		suite.Run(t, testCase)
	}
}

type writeHeadersTestSuite struct {
	suite.Suite
	out *s3.HeadObjectOutput
	hds map[string]string
}

func (s *writeHeadersTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (s *writeHeadersTestSuite) TestWriteHeaders() {
	rdr := httptest.NewRecorder()
	gcx, _ := gin.CreateTestContext(rdr)

	s3util.WriteHeaders(gcx, s.out)

	for key, val := range s.hds {
		s.Assert().Equal(val, rdr.Header().Get(key))
	}
}

func TestWriteHeaders(t *testing.T) {
	tmn := time.Now().UTC()

	for _, testCase := range []*writeHeadersTestSuite{
		{
			out: &s3.HeadObjectOutput{
				AcceptRanges:       aws.String("bytes"),
				ContentLength:      aws.Int64(1),
				LastModified:       aws.Time(tmn),
				ETag:               aws.String("v1"),
				CacheControl:       aws.String("1923"),
				ContentDisposition: aws.String("dpn"),
				ContentEncoding:    aws.String("enc"),
				ContentType:        aws.String("ctp"),
				Expires:            aws.String("exp"),
			},
			hds: map[string]string{
				"Accept-Ranges":       "bytes",
				"Last-Modified":       tmn.Format(time.RFC1123),
				"Content-Length":      "1",
				"ETag":                "v1",
				"Cache-Control":       "1923",
				"Content-Disposition": "dpn",
				"Content-Encoding":    "enc",
				"Content-Type":        "ctp",
				"Expires":             "exp",
			},
		},
	} {
		suite.Run(t, testCase)
	}
}

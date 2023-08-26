package boss

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"go-common/library/net/trace"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

const (
	_defaultRegion        = "sh"
	_defaultComponentName = "library/database/boss"
)

// Config boss config
type Config struct {
	Host            string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// Boss is
type Boss struct {
	conf *Config
	*s3.S3
}

func traceHandler() func(*request.Request) {
	return func(r *request.Request) {
		ctx := r.Context()
		if ctx == nil {
			return
		}
		t, ok := trace.FromContext(ctx)
		if !ok {
			return
		}
		t = t.Fork("", r.Operation.Name)
		t.SetTag(trace.SpanKindClientTag)
		t.SetTag(trace.TagString(trace.TagComponent, _defaultComponentName))
		t.SetTag(trace.TagString(trace.TagHTTPMethod, r.Operation.HTTPMethod))
		t.SetTag(trace.TagString(trace.TagSpanKind, "client"))
		t.SetTag(trace.TagString(trace.TagPeerService, r.ClientInfo.ServiceName))
		trace.Inject(t, trace.HTTPFormat, r.HTTPRequest.Header)

		r.Handlers.Send.PushFront(func(req *request.Request) {
			t.SetTag(trace.TagString(trace.TagHTTPURL, r.HTTPRequest.URL.String()))
		})
		r.Handlers.Complete.PushBack(func(req *request.Request) {
			if req.HTTPResponse != nil {
				t.SetTag(trace.TagInt64(trace.TagHTTPStatusCode, int64(req.HTTPResponse.StatusCode)))
			} else {
				t.SetTag(trace.TagBool(trace.TagError, true))
			}
			t.Finish(&req.Error)
		})
		r.Handlers.Retry.PushBack(func(req *request.Request) {
			t.SetLog(trace.Log(trace.LogEvent, "retry"))
		})
	}
}

const (
	Bucket = "activity"
)

var (
	Client *Boss
)

func NewClient(conf *Config) {
	Client = New(conf)
}

// New is
func New(conf *Config) *Boss {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(_defaultRegion),
			Credentials: credentials.NewStaticCredentials(
				conf.AccessKeyID,
				conf.SecretAccessKey,
				conf.SessionToken,
			),
		},
		SharedConfigState: session.SharedConfigDisable,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create boss client: %+v", err))
	}
	cfg := aws.NewConfig().
		WithEndpoint(conf.Host).
		WithRegion(_defaultRegion).
		WithS3ForcePathStyle(true)
	s3client := s3.New(sess, cfg)
	s3client.Handlers.Build.PushFront(traceHandler())
	boss := &Boss{
		conf: conf,
		S3:   s3client,
	}
	return boss
}

// PutObject is
func (b *Boss) PutObject(ctx context.Context, bucket, path string, payload []byte) (*s3.PutObjectOutput, error) {
	body := bytes.NewReader(payload)
	req := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          body,
		ContentLength: aws.Int64(int64(len(payload))),
		ContentType:   aws.String(http.DetectContentType(payload)),
	}
	reply, err := b.PutObjectWithContext(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

// UploadObject ...
func (b *Boss) UploadObject(ctx context.Context, bucket, path string, reader io.Reader) (string, error) {
	uploader := s3manager.NewUploaderWithClient(b.S3)
	result, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   reader,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}
	return result.Location, nil
}

// HeadObject is
func (b *Boss) HeadObject(ctx context.Context, bucket, path string) (*s3.HeadObjectOutput, error) {
	req := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	reply, err := b.HeadObjectWithContext(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

// GetObject is
func (b *Boss) GetObject(ctx context.Context, bucket, path string) (*s3.GetObjectOutput, error) {
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	reply, err := b.GetObjectWithContext(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

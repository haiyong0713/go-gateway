package boss

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go-common/library/log"

	"go-gateway/app/web-svr/web/interface/conf"
)

type Dao struct {
	cfg  *conf.Boss
	sess *session.Session
}

func NewDao(cfg *conf.Boss) *Dao {
	sess := session.Must(session.NewSession(&aws.Config{
		DisableSSL:                aws.Bool(true),
		Endpoint:                  aws.String(cfg.EndPoint),
		Region:                    aws.String(cfg.Region),
		DisableEndpointHostPrefix: aws.Bool(true),
		DisableComputeChecksums:   aws.Bool(true),
		S3ForcePathStyle:          aws.Bool(true),
		S3Disable100Continue:      aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	}))
	return &Dao{cfg: cfg, sess: sess}
}

func (d *Dao) Upload(key string, body io.Reader) (*s3manager.UploadOutput, error) {
	uploader := s3manager.NewUploader(d.sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		log.Error("Fail to upload to boss, bucket=%s key=%s error=%+v", d.cfg.Bucket, key, err)
		return nil, err
	}
	return result, nil
}

package boss

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	cfg    *conf.BossCfg
	sess   *session.Session
	client *s3.S3
}

func NewDao(cfg *conf.BossCfg) *Dao {
	sess := session.Must(session.NewSession(&aws.Config{
		DisableSSL:                aws.Bool(true),
		Endpoint:                  aws.String(cfg.EntryPoint),
		Region:                    aws.String(cfg.Region),
		DisableEndpointHostPrefix: aws.Bool(true),
		DisableComputeChecksums:   aws.Bool(true),
		S3ForcePathStyle:          aws.Bool(true),
		S3Disable100Continue:      aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
	}))
	return &Dao{cfg: cfg, sess: sess, client: s3.New(sess)}
}

func (d *Dao) DownloadBuffer(key string) ([]byte, error) {
	writeAtBuffer := aws.NewWriteAtBuffer([]byte{})
	downloader := s3manager.NewDownloader(d.sess)
	if _, err := downloader.Download(writeAtBuffer, &s3.GetObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	}); err != nil {
		log.Error("Fail to download from boss to buffer, bucket=%s key=%s error=%+v", d.cfg.Bucket, key, err)
		return nil, err
	}
	return writeAtBuffer.Bytes(), nil
}

func (d *Dao) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	}
	if _, err := d.client.DeleteObject(input); err != nil {
		log.Error("Fail to delete boss object, key=%+v error=%+v", key, err)
		return err
	}
	return nil
}

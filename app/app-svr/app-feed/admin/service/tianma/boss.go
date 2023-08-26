package tianma

import (
	"fmt"
	"os"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type SignedInfo struct {
	Location  string `json:"-"`
	SignedUrl string `json:"signed_url"`
	Key       string `json:"key"`
	Expire    int64  `json:"expire"`
}

// 获取上传到boss的预签名url
func (s *Service) BossSignedUploadUrl(username string) (signedInfo *SignedInfo, err error) {
	// if username == "" {
	// 	err = ecode.Error(ecode.NoLogin, "未登录")
	// 	return
	// }

	svc := s3.New(s.boss)

	key := uuid.New().String()

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.c.Boss.Bucket),
		Key:    aws.String(key),
	})

	//nolint:gomnd
	expire := 24 * time.Hour
	url, err := req.Presign(expire)
	if err != nil {
		return
	}

	signedInfo = &SignedInfo{
		SignedUrl: url,
		Key:       key,
		Expire:    int64(expire.Round(time.Second).Seconds()),
	}

	return
}

// 通过key获取到下载用的预签名url，有有效期
func (s *Service) BossSignedDownloadUrl(key string, username string) (signedInfo *SignedInfo, err error) {
	// if username == "" {
	// 	err = ecode.Error(ecode.NoLogin, "未登录")
	// 	return
	// }

	svc := s3.New(s.boss)

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.c.Boss.Bucket),
		Key:    aws.String(key),
	})

	//nolint:gomnd
	expire := 24 * time.Hour
	url, err := req.Presign(expire)

	signedInfo = &SignedInfo{
		SignedUrl: url,
		Key:       key,
		Expire:    int64(expire.Round(time.Second).Seconds()),
	}

	return
}

// 上传本地文件到boss
func (s *Service) BossUploadLocalFile(filePath string) (signedInfo *SignedInfo, err error) {
	uploader := s3manager.NewUploader(s.boss)

	fp, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer fp.Close()

	key := uuid.New().String()
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.c.Boss.Bucket),
		Key:    aws.String(key),
		Body:   fp,
	})

	if err != nil {
		log.Error("Boss file uploaded error, %s\n", err)
		return
	}

	log.Info("Boss file uploaded to, %s\n", result.Location)

	signedInfo = &SignedInfo{
		Location: result.Location,
		Key:      key,
	}

	return
}

// 下载 boss 上的文件到本地
func (s *Service) BossDownloadLocalFile(key string, target string) (err error) {
	fp, err := os.Create(target)
	if err != nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("Boss file uploaded error, %s\n", err))
		log.Error("Boss file uploaded error, %s\n", err)
		return
	}
	defer fp.Close()

	downloader := s3manager.NewDownloader(s.boss)
	_, err = downloader.Download(fp, &s3.GetObjectInput{
		Bucket: aws.String(s.c.Boss.Bucket),
		Key:    aws.String(key),
	})
	return
}

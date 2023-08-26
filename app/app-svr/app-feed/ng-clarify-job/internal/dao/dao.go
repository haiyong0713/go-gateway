package dao

import (
	"context"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/queue/databus"
	"go-common/library/queue/databus/databusutil"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/internal/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, NewKV)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	SaveSession(context.Context, *model.IndexSession) error
	DownloadURL(string, time.Duration) (string, error)
	ScanArchvieIndex(context.Context, int64, int64, string) (*model.ScanArchiveIndexReply, error)
}

// dao dao.
type dao struct {
	taishan         *Taishan
	s3              *s3.S3
	sessionDatabus  *databus.Databus
	sessionGroup    *databusutil.Group
	sessionRecorder *sessionRecorder
}

// New new a dao and return.
func New(taishan *Taishan) (Dao, func(), error) {
	return newDao(taishan)
}

type s3Config struct {
	Endpoint   string
	Region     string
	Bucket     string
	CredID     string
	CredSecret string
	CredToken  string
}

func newS3Client(s3Cfg *s3Config) *s3.S3 {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(s3Cfg.Region),
			Credentials: credentials.NewStaticCredentials(s3Cfg.CredID, s3Cfg.CredSecret, s3Cfg.CredToken),
		},
		SharedConfigState: session.SharedConfigDisable,
	})
	if err != nil {
		panic(err)
	}
	cfg := aws.NewConfig().
		WithEndpoint(s3Cfg.Endpoint).
		WithRegion(s3Cfg.Region).
		WithS3ForcePathStyle(true)
	return s3.New(sess, cfg)
}

func newDao(taishan *Taishan) (*dao, func(), error) {
	var cfg struct {
		SessionDatabus     *databus.Config
		SessionConsumeUtil *databusutil.Config
		S3                 *s3Config
		SessionRecorder    struct {
			MaxBufferSize int
		}
	}
	if err := paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	d := &dao{
		taishan:        taishan,
		s3:             newS3Client(cfg.S3),
		sessionDatabus: databus.New(cfg.SessionDatabus),
	}
	d.sessionRecorder = newSessionRecorder(taishan, d.s3, cfg.S3.Bucket, cfg.SessionRecorder.MaxBufferSize)
	d.sessionGroup = databusutil.NewGroup(cfg.SessionConsumeUtil, d.sessionDatabus.Messages())
	d.startSessionGroup()
	cf := d.Close
	return d, cf, nil
}

func (d *dao) startSessionGroup() {
	d.sessionGroup.New = parseSession
	d.sessionGroup.Split = shardingSession
	d.sessionGroup.Do = d.processSession
	d.sessionGroup.Start()
}

// Close close the resource.
func (d *dao) Close() {
	d.sessionGroup.Close()
	d.sessionDatabus.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func (d *dao) SaveSession(ctx context.Context, session *model.IndexSession) error {
	return d.sessionRecorder.saveSession(session)
}

func (d *dao) DownloadURL(archivePath string, duration time.Duration) (string, error) {
	return d.sessionRecorder.presignedURL(archivePath, duration)
}

func (d *dao) ScanArchvieIndex(ctx context.Context, startTS, endTS int64, lastKey string) (*model.ScanArchiveIndexReply, error) {
	return d.sessionRecorder.scanArchvieIndex(ctx, startTS, endTS, lastKey)
}

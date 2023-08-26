package tianma

import (
	"os"
	"path"

	"go-common/library/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/tianma"
)

const (
	_midFileNetDir  = "mid-file/net"
	_midFileBossDir = "mid-file/boss"
)

// Service is tianma service
type Service struct {
	dao  *tianma.Dao
	boss *session.Session
	c    *conf.Config
}

// New new a tianma service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:  tianma.New(c),
		boss: NewSession(c),
		c:    c,
	}

	netFilePath := path.Join(c.Boss.LocalDir, _midFileNetDir)
	if ext := path.Ext(netFilePath); ext == "" {
		err := os.MkdirAll(netFilePath, os.ModePerm)
		if err != nil {
			log.Error("os.Mkdir error %s", err)
		}
	}
	bossFilePath := path.Join(c.Boss.LocalDir, _midFileBossDir)
	if ext := path.Ext(bossFilePath); ext == "" {
		err := os.MkdirAll(bossFilePath, os.ModePerm)
		if err != nil {
			log.Error("os.Mkdir error %s", err)
		}
	}

	// 监控是否有推荐对应的mid人群包没有被下载或者上传
	//nolint:errcheck,biligowordcheck
	go s.MidFileMonitor()

	return
}

// 创建新的 boss session
func NewSession(c *conf.Config) *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		DisableSSL:                aws.Bool(true),
		Endpoint:                  aws.String(c.Boss.EntryPoint),
		Region:                    aws.String(c.Boss.Region),
		DisableEndpointHostPrefix: aws.Bool(true),
		DisableComputeChecksums:   aws.Bool(true),
		S3ForcePathStyle:          aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			c.Boss.AccessKey,
			c.Boss.SecretKey,
			"",
		),
	}))
	return sess
}

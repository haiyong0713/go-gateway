package app

import (
	"context"

	"go-common/library/database/bfs"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	ossdao "go-gateway/app/app-svr/fawkes/service/dao/oss"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service struct.
type Service struct {
	c          *conf.Config
	fkDao      *fkdao.Dao
	httpClient *bm.Client
	hotfixChan chan func()
	// bfs client
	bfsCli *bfs.BFS
	ossDao *ossdao.Dao
	gitSvr *gitSvr.Service
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		fkDao:      fkdao.New(c),
		httpClient: bm.NewClient(c.HTTPClient),
		hotfixChan: make(chan func(), 512),
		// bfs
		bfsCli: bfs.New(c.BFS),
		ossDao: ossdao.New(c),
		gitSvr: gitSvr.New(c),
	}
	// nolint:biligowordcheck
	go s.hotfixproc()
	return
}

// AddHotfixProc add hotfix proc
func (s *Service) AddHotfixProc(f func()) {
	select {
	case s.hotfixChan <- f:
	default:
		log.Warn("addHotfix chan full")
	}
}

func (s *Service) hotfixproc() {
	for {
		f, ok := <-s.hotfixChan
		if !ok {
			log.Warn("hotfix proc exit")
			return
		}
		f()
	}
}

// Upload .
func (s *Service) Upload(c context.Context, bucket, fileName, contentType string, file []byte) (url string, err error) {
	if url, err = s.bfsCli.Upload(c, &bfs.Request{
		Bucket:      bucket,
		ContentType: contentType,
		Filename:    fileName,
		File:        file,
	}); err != nil {
		log.Error("Upload(err:%v)", err)
	}
	return
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.fkDao.Close()
}

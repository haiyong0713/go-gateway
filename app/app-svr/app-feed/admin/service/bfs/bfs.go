package bfs

import (
	"bytes"
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	bfsdao "go-gateway/app/app-svr/app-feed/admin/dao/bfs"
)

// Service bfs service.
type Service struct {
	dao        *bfsdao.Dao
	BfsMaxSize int
}

// New new a bfs service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:        bfsdao.New(c),
		BfsMaxSize: c.Bfs.MaxFileSize,
	}
	return
}

// ClientUpCover client upload cover.
func (s *Service) ClientUpCover(c context.Context, fileType string, body []byte) (url string, err error) {
	if len(body) == 0 {
		err = ecode.FileNotExists
		return
	}
	if len(body) > s.BfsMaxSize {
		err = ecode.FileTooLarge
		return
	}
	url, err = s.dao.Upload(c, fileType, bytes.NewReader(body))
	if err != nil {
		log.Error("s.bfs.Upload error(%v)", err)
	}
	return
}

// FileMd5 is used for calculating file md5.
func (s *Service) FileMd5(content []byte) (md5Str string, err error) {
	return s.dao.FileMd5(content)
}

// ValidGif .
func (s *Service) ValidGif(c context.Context, frame string, contents []byte) (err error) {
	return s.dao.ValidGif(c, frame, contents)
}

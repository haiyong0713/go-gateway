package service

import (
	"context"
	"strings"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
)

// BfsProxy .
func (s *Service) BfsProxy(c context.Context, url string) (location string) {
	var (
		bs   []byte
		err  error
		name string
	)
	if env.DeployEnv == env.DeployEnvUat {
		return url
	}
	if url == "" || (strings.Index(url, "https://") == _imageExits && strings.Index(url, _scoreImgage) != _imageExits) {
		location = url
		return
	}
	for i := 0; i < _tryTimes; i++ {
		if bs, err = s.dao.ThirdGet(context.Background(), url); err != nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	if err != nil {
		log.Error("s.dao.BfsPicture url(%s) error(%+v)", url, err)
		return
	}
	if strings.Index(url, _scoreImgage) == _imageExits {
		name = strings.Replace(url, _scoreImgage, "", 1)
	} else {
		name = strings.Replace(url, s.c.Leidata.IP, "", 1)
	}
	if location, err = s.dao.BfsUpload(c, bs, name); err != nil {
		log.Error("s.dao.BfsUpload url(%s) error(%v)", url, err)
		time.Sleep(time.Millisecond * 100)
	}
	time.Sleep(time.Millisecond * 10)
	return
}

package service

import (
	"context"

	"go-gateway/app/app-svr/kvo/interface/conf"
	"go-gateway/app/app-svr/kvo/interface/dao"
	"go-gateway/app/app-svr/kvo/interface/model"

	"go-common/library/database/taishan"
	"go-common/library/sync/pipeline/fanout"

	infoc2 "go-common/library/log/infoc.v2"
)

// Service kvo main service
type Service struct {
	cfg            *conf.Config
	da             *dao.Dao
	docLimit       int
	localCache     *model.LRUCache
	taishan        taishan.TaishanProxyClient
	infocLogStream infoc2.Infoc
	cacheLog       *fanout.Fanout
}

// New get a kvo service
func New(c *conf.Config) *Service {
	var err error
	s := &Service{
		cfg:        c,
		da:         dao.New(c),
		docLimit:   c.Rule.DocLimit,
		localCache: model.New(int64(c.Localcache.BucketSize), c.Localcache.Max),
		cacheLog:   fanout.New("bi_log", fanout.Worker(1), fanout.Buffer(1024)),
	}
	s.infocLogStream, _ = infoc2.New(c.InfocLogStream)
	if s.taishan, err = taishan.NewClient(c.TaishanRPC); err != nil {
		panic(err)
	}
	return s
}

// Ping kvo service check
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.da.Ping(ctx)
}

func (s *Service) Close() {
	if s.infocLogStream != nil {
		_ = s.infocLogStream.Close()
	}
	if s.cacheLog != nil {
		_ = s.cacheLog.Close()
	}
}

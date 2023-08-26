package operation

import (
	"context"
	"strconv"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-show/interface/conf"
	"go-gateway/app/web-svr/web-show/interface/dao/operation"
	opdml "go-gateway/app/web-svr/web-show/interface/model/operation"

	"github.com/robfig/cron"
)

const (
	_rankCacheLen = 20
)

// Service struct
type Service struct {
	dao   *operation.Dao
	cache map[string]map[int][]*opdml.Operation
	// archive grpc
	arcGRPC     arcgrpc.ArchiveClient
	cron        *cron.Cron
	operRunning bool
}

// New init
func New(c *conf.Config) (s *Service) {
	s = &Service{
		cache: make(map[string]map[int][]*opdml.Operation, len(opdml.Types)),
	}
	s.dao = operation.New(c)
	s.cron = cron.New()
	var err error
	if s.arcGRPC, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	s.reload()
	if err = s.loadCron(c); err != nil {
		panic(err)
	}
	return
}

func (s *Service) loadCron(c *conf.Config) error {
	err := s.cron.AddFunc(c.Cron.Notice, s.loadproc)
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

// Notice return notice info
func (s *Service) operation(tp string, rank, num int) map[string][]*opdml.Operation {
	tmp, ok := s.cache[tp]
	if !ok {
		return nil
	}
	res := map[string][]*opdml.Operation{}
	if rank != 0 {
		if ns := tmp[rank]; ns != nil {
			if len(ns) < num || num < 0 {
				num = len(ns)
			}
			ns = ns[:num]
			res[strconv.FormatInt(int64(rank), 10)] = ns
		}
		return res
	}
	for rk, ns := range tmp {
		if ns != nil {
			if len(ns) < num || num < 0 {
				num = len(ns)
			}
			ns = ns[:num]
			res[strconv.FormatInt(int64(rk), 10)] = ns
		}
	}
	return res
}

// reload Service
func (s *Service) reload() {
	var (
		tmpT = make(map[string]map[int][]*opdml.Operation)
	)
	ops, err := s.dao.Operation(context.Background())
	if err != nil {
		log.Error("s.reloadNotice error(%v)", err)
		return
	}
	for _, op := range ops {
		tmp, ok := tmpT[op.Type]
		if !ok {
			tmp = make(map[int][]*opdml.Operation)
		}
		if len(tmp) > _rankCacheLen {
			continue
		}
		tmp[op.Rank] = append(tmp[op.Rank], op)
		tmpT[op.Type] = tmp
	}
	s.cache = tmpT
}

// loadproc Service
func (s *Service) loadproc() {
	if s.operRunning {
		return
	}
	s.operRunning = true
	defer func() {
		s.operRunning = false
	}()
	s.reload()
}

// Close Service
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}

// Ping Service
func (s *Service) Ping(c context.Context) (err error) {
	err = s.dao.Ping(c)
	return
}

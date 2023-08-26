package ranklist

import (
	"context"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/conf"
	ranklistdao "go-gateway/app/app-svr/app-job/job/dao/rank-list"

	"github.com/robfig/cron"
)

// Service is
type Service struct {
	c    *conf.Config
	cron *cron.Cron
	dao  *ranklistdao.Dao
}

// New is
func New(c *conf.Config) *Service {
	s := &Service{
		c:    c,
		cron: cron.New(),
		dao:  ranklistdao.New(c),
	}
	if err := s.loadRankMeta(context.Background()); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc("@every 5m", func() {
		if err := s.loadRankMeta(context.Background()); err != nil {
			log.Error("日志告警 榜单元数据加载失败: %+v", err)
		}
	}); err != nil {
		panic(err)
	}
	s.cron.Start()
	return s
}

func (s *Service) loadRankMeta(ctx context.Context) error {
	page := int64(1)
	size := int64(20)

	metaPage, err := s.dao.ScanRankMeta(ctx, size, page)
	if err != nil {
		return err
	}
	if err := s.dao.SetCacheRankMeta(ctx, metaPage.List...); err != nil {
		return err
	}

	steps := metaPage.Page.Total / size
	if (metaPage.Page.Total % size) > 0 {
		steps += 1
	}
	for i := int64(0); i < steps; i++ {
		metaPage, err := s.dao.ScanRankMeta(ctx, size, page+i+1)
		if err != nil {
			return err
		}
		if err := s.dao.SetCacheRankMeta(ctx, metaPage.List...); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 20)
	}
	return nil
}

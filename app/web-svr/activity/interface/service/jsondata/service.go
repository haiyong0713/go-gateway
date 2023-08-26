package jsondata

import (
	"context"
	"time"

	"go-common/library/log"

	"go-gateway/app/web-svr/activity/interface/conf"
	dao "go-gateway/app/web-svr/activity/interface/dao/jsondata"
	mdl "go-gateway/app/web-svr/activity/interface/model/jsondata"
)

// Service ...
type Service struct {
	c          *conf.Config
	dao        *dao.Dao
	summerGift []*mdl.SummerGift
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
	}

	ctx := context.Background()
	err := s.initSummerGift(ctx)
	if err != nil {
		log.Errorc(ctx, "New init task error: %v", err)
		panic(err)
	}

	go s.updateSummerGiftLoop()
	return s
}

func (s *Service) updateSummerGiftLoop() {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		err := s.initSummerGift(ctx)
		if err != nil {
			continue
		}
	}
}

// Close ...
func (s *Service) Close() {
}

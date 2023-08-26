package service

import (
	"context"
	"go-common/library/log"
	"time"
)

func (s *Service) loadCustomConfig() {
	ctx := context.Background()
	now := time.Now()
	list, err := s.dao.GetModifiedCCFromDB(ctx)
	if err != nil {
		return
	}
	if len(list) == 0 {
		log.Error("loadCustomConfig success at %s, data length is %v", now.Format("2006-01-02 15:04:05"), 0)
		return
	}
	err = s.dao.SetModifiedCCIntoRedis(ctx, list)
	if err == nil {
		log.Error("loadCustomConfig success at %s, data length is %v", now.Format("2006-01-02 15:04:05"), len(list))
	}
}

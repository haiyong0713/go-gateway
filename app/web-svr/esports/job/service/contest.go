package service

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
)

const (
	cacheKey4MaxContestID = "contest:max_id"
)

func (s *Service) ASyncResetMaxContestID(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			s.resetMaxContestID()
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) resetMaxContestID() {
	if d, err := s.dao.FetchMaxContestID(context.Background()); err == nil {
		_ = resetMaxContestIdInCache(d)
	}
}

func resetMaxContestIdInCache(maxID int64) (err error) {
	item := &memcache.Item{Key: cacheKey4MaxContestID, Object: maxID, Expiration: 3600, Flags: memcache.FlagJSON}

	return globalMemcache.Set(context.Background(), item)
}

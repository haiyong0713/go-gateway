package service

import (
	"context"
	"testing"
	xtime "time"

	"go-common/library/cache/memcache"
	"go-common/library/container/pool"
	"go-common/library/time"
)

// go test -v --count=1 archive.go auto_subscribe.go bfs.go contest.go contest_analysis.go contest_test.go ftp.go leidata.go live_s10.go match.go pointdata.go s10ranking.go score.go season_biz.go season_notify.go service.go
func TestContestBiz(t *testing.T) {
	cfg := &memcache.Config{
		Name:  "local",
		Proto: "tcp",
		Addr:  "127.0.0.1:11211",
		Config: &pool.Config{
			IdleTimeout: time.Duration(10 * xtime.Second),
			Idle:        2,
			Active:      8,
		},
		WriteTimeout: time.Duration(10 * xtime.Second),
		ReadTimeout:  time.Duration(10 * xtime.Second),
		DialTimeout:  time.Duration(10 * xtime.Second),
	}
	globalMemcache = memcache.New(cfg)
	t.Run("test reset max contestID", resetMaxContestID)
}

func resetMaxContestID(t *testing.T) {
	tmpContestID := int64(88888888)
	if err := resetMaxContestIdInCache(tmpContestID); err != nil {
		t.Error(err)

		return
	}

	var maxContestID int64
	if err := globalMemcache.Get(context.Background(), cacheKey4MaxContestID).Scan(&maxContestID); err != nil || maxContestID != tmpContestID {
		t.Error(err, maxContestID)

		return
	}
}

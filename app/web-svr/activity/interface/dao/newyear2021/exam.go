package newyear2021

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/tool"

	"go-common/library/cache/redis"
)

const (
	cacheKey4ExamStats        = "bnj:2021:live:exam:stats"
	cacheKey4UserCommitDetail = "bnj:2021:live:exam:%v:commits"
)

func FetchExamStats(ctx context.Context) (m map[string]int64, err error) {
	m = make(map[string]int64, 0)
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	m, err = redis.Int64Map(conn.Do("HGETALL", cacheKey4ExamStats))

	return
}

func MultiUpdateExamStats(ctx context.Context, stats *api.ExamStatsReq) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	args := []interface{}{cacheKey4ExamStats}
	for _, v := range stats.Stats {
		args = append(args, fmt.Sprintf("%v_%v", v.Id, v.OptID))
		args = append(args, v.Total)
	}
	_, err = conn.Do("HMSET", args...)

	return
}

func userExamCommitCacheKey(mid int64) string {
	return fmt.Sprintf(cacheKey4UserCommitDetail, mid)
}

func FetchUserCommitDetail(ctx context.Context, mid int64) (m map[string]int64, err error) {
	m = make(map[string]int64, 0)
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	m, err = redis.Int64Map(conn.Do("HGETALL", userExamCommitCacheKey(mid)))

	return
}

func CommitUserOption(ctx context.Context, mid, itemID, optID int64) (affect bool, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var exists bool
	cacheKey := userExamCommitCacheKey(mid)
	exists, err = redis.Bool(conn.Do("HEXISTS", cacheKey, itemID))
	if err == nil && !exists {
		_, err = conn.Do("HSET", cacheKey, itemID, optID)
		_, _ = conn.Do("EXPIRE", cacheKey, tool.CalculateExpiredSeconds(1))
		if err == nil {
			affect = true
		}
	}

	return
}

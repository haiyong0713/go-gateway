package newyear2021

import (
	"context"
	"encoding/json"
	"fmt"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/newyear2021"

	"go-common/library/cache/redis"
)

const (
	bizNameOfLiveLotteryList = "bnj_live_lottery_list"

	cacheKey4BnjLotteryRecordOfUser    = "bnj2021:unReceived:reward:%v:live"
	cacheKey4BnjLotteryRecordOfReserve = "bnj2021:unReceived:reward:%v:reserve"

	SceneID4NiuDan   = 1
	SceneID4LiveView = 2
	SceneID4Reserve  = 3
	SceneID4ARDraw   = 4
)

func cacheKey4UserLotteryOfLive(mid int64) string {
	return fmt.Sprintf(cacheKey4BnjLotteryRecordOfUser, mid)
}

func cacheKey4UserLotteryOfReserve(mid int64) string {
	return fmt.Sprintf(cacheKey4BnjLotteryRecordOfReserve, mid)
}

func PopUserRewardBySceneID(ctx context.Context, mid, sceneID int64) (
	reward *newyear2021.UserRewardInLiveRoom, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var bs []byte
	reward = new(newyear2021.UserRewardInLiveRoom)
	switch sceneID {
	case SceneID4Reserve:
		cacheKey := cacheKey4UserLotteryOfReserve(mid)
		bs, err = redis.Bytes(conn.Do("GET", cacheKey))
		if err == nil {
			_, _ = conn.Do("DEL", cacheKey)
		}
	case SceneID4LiveView:
		cacheKey := cacheKey4UserLotteryOfLive(mid)
		bs, err = redis.Bytes(conn.Do("LPOP", cacheKey))
	}

	if err != nil && err != redis.ErrNil {
		err = ecode.BNJTooManyUser

		return
	}

	if err == redis.ErrNil {
		err = ecode.BNJNoEnoughCoupon2Draw

		return
	}

	_ = json.Unmarshal(bs, reward)

	return
}

func FetchLiveLotteryQuota(ctx context.Context, mid, sceneID int64) (quota int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	switch sceneID {
	case SceneID4Reserve:
		var reserveRewardExists bool
		cacheKey := cacheKey4UserLotteryOfReserve(mid)
		reserveRewardExists, err = redis.Bool(conn.Do("EXISTS", cacheKey))
		if err == nil && reserveRewardExists {
			quota = 1
		}
	case SceneID4LiveView:
		cacheKey := cacheKey4UserLotteryOfLive(mid)
		quota, err = redis.Int64(conn.Do("LLEN", cacheKey))
	}

	return
}

func FetchUserLiveLotteryList(ctx context.Context, mid int64) (list []*newyear2021.LiveRewardDetail, err error) {
	list = make([]*newyear2021.LiveRewardDetail, 0)
	if mid == 0 {
		return
	}

	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var (
		reserveRewardExists bool
		liveRewardCount     int64
	)
	// fetch reserve reward
	cacheKey := cacheKey4UserLotteryOfReserve(mid)
	reserveRewardExists, err = redis.Bool(conn.Do("EXISTS", cacheKey))
	if err == nil || err == redis.ErrNil {
		tmp := new(newyear2021.LiveRewardDetail)
		{
			tmp.SceneID = SceneID4Reserve
			tmp.Quota = 0
		}

		if reserveRewardExists {
			tmp.Quota = 1
		}

		list = append(list, tmp)
	}

	cacheKey = cacheKey4UserLotteryOfLive(mid)
	liveRewardCount, err = redis.Int64(conn.Do("LLEN", cacheKey))
	if err == nil || err == redis.ErrNil {
		tmp := new(newyear2021.LiveRewardDetail)
		{
			tmp.SceneID = SceneID4LiveView
			tmp.Quota = liveRewardCount
		}

		list = append(list, tmp)
	}

	c, err := FetchUserCoupon(ctx, mid)
	if err == nil {
		tmp := new(newyear2021.LiveRewardDetail)
		{
			tmp.SceneID = SceneID4NiuDan
			tmp.Quota = c.ND
		}

		list = append(list, tmp)
	}

	if err == redis.ErrNil {
		err = nil
	}

	return
}

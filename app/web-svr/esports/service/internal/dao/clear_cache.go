package dao

import (
	"context"
	"encoding/json"
	actmdl "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/service/internal/model"

	"go-common/library/log"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

const (
	cacheKey4AutoSubscribeList = "auto_subscribe"
)

func (d *dao) clearCacheWhenUpdateContest(ctx context.Context, param *model.ContestModel, preData *model.ContestModel) {
	d.cache.Do(ctx, func(c context.Context) {
		_ = d.ClearESportCacheByType(v1.ClearCacheType_CONTEST, []int64{param.ID})
		if e := d.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: param.Sid, ContestID: param.ID, ContestHome: param.HomeID, ContestAway: param.AwayID}); e != nil {
			log.Errorc(c, "[ClearComponentContestCacheByGRPC]Update, sid:%d, contestId:%d, %+v", param.Sid, param.ID, e)
		}
		if _, e := d.espClient.RefreshContestDataPageCache(context.Background(), &v1.RefreshContestDataPageCacheRequest{
			Cids: []int64{param.ID},
		}); e != nil {
			log.Errorc(c, "s.espClient.RefreshContestDataPageCache  param(%+v) error(%v)", param.ID, e)
		}
		if preData.GuessType > 0 && (preData.Stime != param.Stime || preData.Etime != param.Etime) {
			editReq := &actmdl.GuessEditReq{Business: int64(actmdl.GuessBusiness_esportsType), Oid: param.ID, Stime: param.Stime, Etime: param.Etime}
			if _, e := d.activityClient.GuessEdit(context.Background(), editReq); e != nil {
				log.Error("s.actClient.GuessEdit  param(%+v) error(%v)", editReq, e)
				return
			}
		}
		if e := d.RefreshContestSeriesExtraInfo(context.Background(), param); e != nil {
			return
		}
		if err := d.DeleteContestCache(ctx, param.ID); err != nil {
			log.Errorc(c, "[ClearCacheWhenAddContest][DeleteContestCache] param(%+v) error(%v)", param.ID, err)
		}
	})
}

func (d *dao) clearCacheWhenAddContest(ctx context.Context, param *model.ContestModel) {
	// 删除赛程组件缓存.
	d.cache.Do(ctx, func(c context.Context) {
		_ = d.ClearESportCacheByType(v1.ClearCacheType_CONTEST, []int64{param.ID})
		if e := d.ClearComponentContestCacheByGRPC(&v1.ClearComponentContestCacheRequest{SeasonID: param.Sid, ContestID: param.ID, ContestHome: param.HomeID, ContestAway: param.AwayID}); e != nil {
			log.Errorc(c, "[ClearComponentContestCacheByGRPC]Add, sid:%d, contestId:%d, %+v", param.Sid, param.ID, e)
		}
		autoSubK4Home := model.AutoSubscribeDetail{
			SeasonID:  param.Sid,
			TeamId:    param.HomeID,
			ContestID: param.ID,
		}
		bs4Home, _ := json.Marshal(autoSubK4Home)
		autoSub4Away := model.AutoSubscribeDetail{
			SeasonID:  param.Sid,
			TeamId:    param.AwayID,
			ContestID: param.ID,
		}
		bs4Away, _ := json.Marshal(autoSub4Away)
		if _, pushErr := d.redis.Do(ctx, "LPUSH", cacheKey4AutoSubscribeList, string(bs4Home), string(bs4Away)); pushErr != nil {
			log.Errorc(ctx, "[ClearCache][AutoSubscribeDetail][LPUSH]")
		}
		if _, err := d.espClient.RefreshContestDataPageCache(c, &v1.RefreshContestDataPageCacheRequest{
			Cids: []int64{param.ID},
		}); err != nil {
			log.Errorc(c, "s.espClient.RefreshContestDataPageCache  param(%+v) error(%v)", param.ID, err)
		}
		if err := d.DeleteContestCache(ctx, param.ID); err != nil {
			log.Errorc(c, "[ClearCacheWhenAddContest][DeleteContestCache] param(%+v) error(%v)", param.ID, err)
		}
	})

}

func (d *dao) RefreshContestSeriesExtraInfo(ctx context.Context, contest *model.ContestModel) (err error) {
	cs, err := d.GetSeriesById(ctx, contest.SeriesId)
	if err != nil {
		log.Errorc(ctx, "[ClearCache][RefreshContestSeriesExtraInfo], err:%+v", err)
		return
	}
	switch cs.Type {
	case model.SeriesTypPoint:
		_, err = d.espClient.RefreshSeriesPointMatchInfo(ctx, &v1.RefreshSeriesPointMatchInfoReq{
			SeriesId: cs.ID,
		})
	case model.SeriesTypKnockout:
		_, err = d.espClient.RefreshSeriesKnockoutMatchInfo(ctx, &v1.RefreshSeriesKnockoutMatchInfoReq{
			SeriesId: cs.ID,
		})
	}
	if err != nil {
		log.Errorc(ctx, "[ClearCache][RefreshContestSeriesExtraInfo][RefreshSeries], err:%+v", err)
	}
	return
}

func (d *dao) ClearESportCacheByType(cacheType v1.ClearCacheType, list []int64) (err error) {
	if len(list) == 0 {
		return
	}
	req := new(v1.ClearCacheRequest)
	{
		req.CacheType = cacheType
		req.CacheKeys = list
	}
	reqBs, _ := json.Marshal(req)
	for i := 1; i <= 3; i++ {
		_, err = d.espClient.ClearCache(context.Background(), req)
		if err == nil {
			break
		}
		log.Error("[Dao][ClearCache] occur err: %v, req: %v, try times: %v", err, string(reqBs), i)
	}
	return
}

func (d *dao) ClearComponentContestCacheByGRPC(param *v1.ClearComponentContestCacheRequest) (err error) {
	if _, err = d.espClient.ClearComponentContestCache(context.Background(), param); err != nil {
		log.Error("contest component ClearComponentContestCacheGRPC param(%+v) error(%+v)", param, err)
	}
	return
}

func (d *dao) ClearMatchSeasonsCacheByGRPC(param *v1.ClearMatchSeasonsCacheRequest) (err error) {
	if _, err = d.espClient.ClearMatchSeasonsCache(context.Background(), param); err != nil {
		log.Error("MatchSeasonsInfo ClearMatchSeasonsCacheByGRPC param(%+v) error(%+v)", param, err)
	}
	return
}

func (d *dao) ClearVideoListCacheByGRPC(id int64) (err error) {
	if id == 0 {
		return
	}
	ctx := context.Background()
	arg := &v1.ClearTopicVideoListRequest{ID: id}
	if _, err = d.espClient.ClearTopicVideoListCache(ctx, arg); err != nil {
		log.Errorc(ctx, "contest component ClearTopicVideoListCache param(%+v) error(%+v)", arg, err)
	}
	return
}

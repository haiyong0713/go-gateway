package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"strconv"
	"time"
)

/**********************************缓存Key控制**********************************/
//redisRankExternalWithInfoCacheVersionKey: 控制当前RankCache的版本, 实现刷新DSG后快速切换Cache版本功能
//curr* -> 获取当前最新版本的Cache Key
func redisRankExternalWithInfoCacheVersionKey(activityId int64) string {
	return fmt.Sprintf("vote_rank_2CInfo_version_%v", activityId)
}

func (d *Dao) mustGetCurrRedisExternalRankWithInfoCacheVersion(ctx context.Context, activityId int64) (version int64) {
	var err error
	err = retry.Infinite(ctx, "mustGetCurrRedisExternalRankWithInfoCacheVersion", netutil.DefaultBackoffConfig, func(c context.Context) error {
		version, err = redis.Int64(d.redis.Do(ctx, "GET", redisRankExternalWithInfoCacheVersionKey(activityId)))
		if err == redis.ErrNil {
			err = nil
		}
		return err
	})
	if err != nil {
		log.Errorc(ctx, "mustGetCurrRedisExternalRankWithInfoCacheVersion error: %v", err)
	}
	return
}

func (d *Dao) redisRankExternalWithInfoCacheKeyByVersion(sourceGroupId, version int64) (key string) {
	return fmt.Sprintf("vote_rank_2CInfo_list_%v_v%v", sourceGroupId, version)
}

/**********************************排名数据计算/刷新**********************************/

// CalcDSGRankExternal: 计算外部数据组投票排名,外部使用(用户返回).不包含敏感信息, 如干预信息,拉黑配置
// 总票数更新根据投票规则
func (d *Dao) CalcDSGRankExternal(ctx context.Context, DSG *api.VoteDataSourceGroupItem) (res []*model.RankInfo, err error) {
	res = make([]*model.RankInfo, 0)
	dsI, ok := d.datasourceMap[DSG.SourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	//从投票sorted set中获取前N
	topN, err := d.rawExternalDSGVoteInfoById(ctx, DSG.GroupId)
	if err != nil {
		return
	}
	res, err = d.innerCalcDSGRank(ctx, DSG, dsI, topN)
	return
}

// RefreshDSGRankExternal: 计算并更新数据组投票排名缓存.不包含敏感信息, 如干预信息,拉黑配置
func (d *Dao) RefreshDSGRankExternal(ctx context.Context, DSG *api.VoteDataSourceGroupItem, newVersion, expireTime int64) (err error) {
	rankWithInfoListCacheKey := d.redisRankExternalWithInfoCacheKeyByVersion(DSG.GroupId, newVersion)
	var res []*model.RankInfo
	//1.计算数据组下的投票排名
	{
		res, err = d.CalcDSGRankExternal(ctx, DSG)
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityRankExternal d.CalcDSGRankExternal for DSG (%+v) error: %v", DSG, err)
			return
		}
	}

	//2.将结果集存入缓存
	{
		for _, info := range res {
			bs, _ := json.Marshal(info)
			err = retry.WithAttempts(ctx, "RefreshVoteActivityRankExternal_RPUSH", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "RPUSH", rankWithInfoListCacheKey, bs)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RefreshVoteActivityRankExternal RPUSH for DSG (%+v) error: %v", DSG, err)
				return
			}
		}
	}
	//3.设置过期时间
	{
		err = retry.WithAttempts(ctx, "RefreshVoteActivityRankExternal_SetTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "EXPIRE", rankWithInfoListCacheKey, expireTime)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityRankExternal SetNewRankTTL for DSG (%+v) error: %v", DSG, err)
			return
		}
	}
	return
}

// RefreshVoteActivityRankExternal: 刷新整个投票活动的外部投票排名.
func (d *Dao) RefreshVoteActivityRankExternal(ctx context.Context, activityId int64) (err error) {
	defer func() {
		if err != nil {
			if err1 := d.RenewCurrentExternalRankWithInfoCacheTTL(ctx, activityId); err1 != nil {
				err = ecode.ActivityVoteRefreshAndRenewDSGItemsFail
			}
		}
	}()
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "RefreshVoteActivityRankExternal d.CacheActivity error: %v", err)
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	if activity.Rule == nil {
		err = ecode.ActivityVoteRuleNotConfig
		return
	}
	//不同类型的刷新方式, 使用不同的TTL
	expireTime := d.getExternalVoteWithInfoExpireTime(activity)
	DSGs, err := d.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: activityId})
	if err != nil {
		return
	}

	//1.刷新所有DSG的投票排名
	newVersion, _ := strconv.ParseInt(time.Now().Format(cacheVersionFormat), 10, 64)
	eg := errgroup.WithContext(ctx)
	for _, DSG := range DSGs.Groups {
		tmpDsg := DSG
		eg.Go(func(ctx context.Context) error {
			return d.RefreshDSGRankExternal(ctx, tmpDsg, newVersion, expireTime)
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}

	//2.更新缓存版本号
	{
		err = retry.WithAttempts(ctx, "RefreshAllRank_IncrVersion", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "SET", redisRankExternalWithInfoCacheVersionKey(activityId), newVersion)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems IncrVersion error: %v", err)
		}
	}
	//3.更新DB刷新时间
	{
		err = d.updateActivityRankRefreshTime(ctx, activityId)
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems updateActivityRankRefreshTime error: %v", err)
		}
	}
	return
}

// GetDSGRankExternal: 用户请求投票排名(票数排名).
func (d *Dao) GetDSGRankExternal(ctx context.Context, params *model.InnerRankParams) (res *model.RankResultExternal, err error) {
	res, err = d.innerGetDSGRank(ctx, params, rankTypeExternal)
	return
}

// RenewCurrentExternalRankWithInfoCacheTTL: 延长当前版本排行榜的有效期(刷新失败时使用, 避免
func (d *Dao) RenewCurrentExternalRankWithInfoCacheTTL(ctx context.Context, activityId int64) (err error) {
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "RenewCurrentExternalRankWithInfoCacheTTL d.CacheActivity error: %v", err)
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	currVersion := d.mustGetCurrRedisExternalRankWithInfoCacheVersion(ctx, activityId)
	if currVersion == 0 {
		err = ecode.ActivityVoteItemNotFound
		return
	}
	dataSources, err := d.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: activityId})
	if err != nil {
		return
	}
	expireTime := d.getExternalVoteWithInfoExpireTime(activity)
	var eg errgroup.Group
	for _, dsg := range dataSources.Groups {
		tmpDsg := dsg
		eg.Go(func(ctx context.Context) (err error) {
			redisDSItemsMapCacheKey := d.redisRankExternalWithInfoCacheKeyByVersion(tmpDsg.GroupId, currVersion)
			err = retry.WithAttempts(ctx, "RenewCurrentRankWithInfoCacheTTL_SetNewItemTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "EXPIRE", redisDSItemsMapCacheKey, expireTime)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RenewCurrentExternalRankWithInfoCacheTTL SetNewItemTTL error: %v", err)
				return
			}
			return
		})
	}
	err = eg.Wait()
	return
}

// RefreshVoteActivityRankZset: 手动/定时更新外部活动排名使用
func (d *Dao) RefreshVoteActivityRankZset(ctx context.Context, req *api.RefreshVoteActivityRankZsetReq) (err error) {
	activity, err := d.RawActivity(ctx, req.ActivityId)
	if err != nil {
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	sqlStr := sql4UpdateDSTotalVoteItemVoteWithRiskByMainId
	if !activity.Rule.DisplayRiskVote {
		sqlStr = sql4UpdateDSTotalVoteItemVoteWithoutRiskByMainId
	}
	err = retry.WithAttempts(ctx, "RefreshVoteActivityRankZset", 3, netutil.DefaultBackoffConfig,
		func(c context.Context) (err error) {
			_, err = d.db.Exec(ctx, sqlStr, req.ActivityId)
			return
		})

	if err != nil {
		log.Errorc(ctx, "RefreshVoteActivityRankZset sql4UpdateDSTotalVote for activityId: %v exec error: %v", req.ActivityId, err)
		return
	}
	eg := errgroup.WithContext(ctx)
	DSGs, err := d.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: req.ActivityId})
	if err != nil {
		return
	}
	for _, DSG := range DSGs.Groups {
		id := DSG.GroupId
		eg.Go(func(ctx context.Context) error {
			return d.RebuildDSGVoteCountCache(ctx, id)
		})
	}
	err = eg.Wait()
	return
}

// GetDSGRankExternalOrder: 用户请求投票排名(其他排名).
func (d *Dao) GetDSGRankExternalOrder(ctx context.Context, rand bool, params *model.InnerRankParams) (res *model.RankResultExternal, err error) {
	res, err = d.innerGetDSGRankOrder(ctx, rand, params, rankTypeExternal)
	return
}

func (d *Dao) getExternalVoteWithInfoExpireTime(activity *api.VoteActivity) (expireTime int64) {
	expireTime = int64(0)
	//结束90天内的活动每天刷新一次排名, 使用单独的TTL
	if activity.EndTime < time.Now().Unix() && activity.EndTime > time.Now().AddDate(0, 0, -90).Unix() {
		expireTime = d.outdatedVoteRankWithInfoExpire
		return
	}
	switch activity.Rule.VoteUpdateRule {
	case int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime):
		expireTime = d.realTimeVoteRankWithInfoExpire
	case int64(api.VoteCountUpdateRule_VoteCountUpdateRuleManual):
		expireTime = d.manualVoteRankWithInfoExpire
	case int64(api.VoteCountUpdateRule_VoteCountUpdateRuleOnTime):
		expireTime = d.onTimeVoteRankWithInfoExpire
	default:
		expireTime = d.onTimeVoteRankWithInfoExpire
	}
	return expireTime
}

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

const (
	maxRankSize = 20000
)

/**********************************缓存Key控制**********************************/
//redisRankInternalWithInfoCacheVersionKey: 控制当前RankCache的版本, 实现刷新DSG后快速切换Cache版本功能
//curr* -> 获取当前最新版本的Cache Key
func redisRankInternalWithInfoCacheVersionKey(activityId int64) string {
	return fmt.Sprintf("vote_rank_2AInfo_version_%v", activityId)
}

func (d *Dao) mustGetCurrRedisInternalRankWithInfoCacheVersion(ctx context.Context, activityId int64) (version int64) {
	var err error
	err = retry.Infinite(ctx, "mustGetCurrRedisInternalRankWithInfoCacheVersion", netutil.DefaultBackoffConfig, func(c context.Context) error {
		version, err = redis.Int64(d.redis.Do(ctx, "GET", redisRankInternalWithInfoCacheVersionKey(activityId)))
		if err == redis.ErrNil {
			err = nil
		}
		return err
	})
	if err != nil {
		log.Errorc(ctx, "mustGetCurrRedisInternalRankWithInfoCacheVersion error: %v", err)
	}
	return
}

func (d *Dao) redisRankInternalWithInfoCacheKeyByVersion(sourceGroupId, version int64) (key string) {
	return fmt.Sprintf("vote_rank_2AInfo_list_%v_v%v", sourceGroupId, version)
}

/**********************************排名数据计算/刷新**********************************/

// CalcDSGRankInternal: 计算内部数据组投票排名,内部使用.票数从DB获取
// 总票数实时更新
func (d *Dao) CalcDSGRankInternal(ctx context.Context, DSG *api.VoteDataSourceGroupItem) (res []*model.RankInfo, err error) {
	res = make([]*model.RankInfo, 0)
	dsI, ok := d.datasourceMap[DSG.SourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	activity, err := d.Activity(ctx, DSG.ActivityId)
	if err != nil {
		return
	}

	//从DB中中获取前N
	dbN, err := d.rawInternalDSGVoteInfoById(ctx, DSG.GroupId, activity.Rule.DisplayRiskVote)
	if err != nil {
		return
	}
	topN := make([]*model.DataSourceItemVoteInfo, 0, len(dbN))
	for _, t := range dbN {
		topN = append(topN, &model.DataSourceItemVoteInfo{
			SourceItemId:   t.SourceItemId,
			TotalVoteCount: t.TotalVoteCount,
		})
	}
	res, err = d.innerCalcDSGRank(ctx, DSG, dsI, topN)
	return
}

// RefreshDSGRankInternal: 计算并更新数据组内部投票排名缓存.
func (d *Dao) RefreshDSGRankInternal(ctx context.Context, DSG *api.VoteDataSourceGroupItem, newVersion, expireTime int64) (err error) {
	rankWithInfoListCacheKey := d.redisRankInternalWithInfoCacheKeyByVersion(DSG.GroupId, newVersion)
	var res []*model.RankInfo
	//1.计算数据组下的投票排名
	{
		res, err = d.CalcDSGRankInternal(ctx, DSG)
		if err != nil {
			log.Errorc(ctx, "RefreshDSGRankInternal d.CalcDSGRankExternal for DSG (%+v) error: %v", DSG, err)
			return
		}
	}

	//2.将结果集存入缓存
	{
		for _, info := range res {
			bs, _ := json.Marshal(info)
			err = retry.WithAttempts(ctx, "RefreshDSGRankInternal_RPUSH", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "RPUSH", rankWithInfoListCacheKey, bs)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RefreshDSGRankInternal RPUSH for DSG (%+v) error: %v", DSG, err)
				return
			}
		}
	}
	//3.设置过期时间
	{
		err = retry.WithAttempts(ctx, "RefreshDSGRankInternal_SetTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "EXPIRE", rankWithInfoListCacheKey, expireTime)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshDSGRankInternal SetNewRankTTL for DSG (%+v) error: %v", DSG, err)
			return
		}
	}
	return
}

// RefreshVoteActivityRankExternal: 刷新整个投票活动的内部投票排名.
func (d *Dao) RefreshVoteActivityRankInternal(ctx context.Context, activityId int64) (err error) {
	defer func() {
		if err != nil {
			if err1 := d.RenewCurrentExternalRankWithInfoCacheTTL(ctx, activityId); err1 != nil {
				err = ecode.ActivityVoteRefreshAndRenewDSGItemsFail
			}
		}
	}()
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "RefreshVoteActivityRankInternal d.CacheActivity error: %v", err)
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
	expireTime := d.getInternalVoteWithInfoExpireTime(activity)
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
			return d.RefreshDSGRankInternal(ctx, tmpDsg, newVersion, expireTime)
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}

	//2.更新缓存版本号
	{
		err = retry.WithAttempts(ctx, "RefreshAllRank_IncrVersion", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "SET", redisRankInternalWithInfoCacheVersionKey(activityId), newVersion)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityRankInternal IncrVersion error: %v", err)
		}
	}
	return
}

// GetDSGRankInternal: 内部计算投票排名,内部使用,包含干预信息,拉黑配置.不能直接返回给用户
// 复用外部排名逻辑, 附加上内部信息
func (d *Dao) GetDSGRankInternal(ctx context.Context, req *api.GetVoteActivityRankInternalReq) (res *api.GetVoteActivityRankInternalResp, err error) {
	DSG, err := d.DataSourceGroup(ctx, req.SourceGroupId)
	if err != nil {
		log.Errorc(ctx, "GetDSGRankExternal d.DataSourceGroup error: %v", err)
		return
	}
	if DSG == nil {
		err = ecode.ActivityVoteDSGNotFound
		return
	}
	activity, err := d.Activity(ctx, DSG.ActivityId)
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
	//1.获取外部排名
	externalRes := &model.RankResultExternal{
		List: make([]*model.RankInfo, 0),
		Page: &model.Page{},
	}
	if req.SourceItemId != 0 { //查询单个id
		item, err1 := d.GetDSGItem(ctx, DSG, req.SourceItemId)
		if err1 != nil {
			err = err1
			return
		}
		externalRes.List = append(externalRes.List, &model.RankInfo{
			DataSourceGroupId:  DSG.GroupId,
			DataSourceItemId:   item.GetId(),
			DataSourceItemName: item.GetName(),
		})
		externalRes.Page.Total = 1
		return d.RankExternal2Internal(ctx, req, DSG, externalRes, activity.Rule.DisplayRiskVote)

	}

	//查询列表
	params := &model.InnerRankParams{
		Mid:               0,
		ActivityId:        DSG.ActivityId,
		DataSourceGroupId: DSG.GroupId,
		Version:           0,
		Pn:                req.Pn,
		Ps:                req.Ps,
	}
	if !req.OnlyBlackList {
		switch req.Sort {
		case 0: //票数排序
			externalRes, err = d.innerGetDSGRank(ctx, params, rankTypeInternal)
		case 1: //时间排序
			externalRes, err = d.innerGetDSGRankOrder(ctx, false, params, rankTypeInternal)
		}
	} else {
		externalRes.List, err = d.GetBlackListItemInfo(ctx, DSG)
		externalRes.Page.Total = int64(len(externalRes.List))
	}

	if err != nil {
		return
	}
	return d.RankExternal2Internal(ctx, req, DSG, externalRes, activity.Rule.DisplayRiskVote)
}

func (d *Dao) RankExternal2Internal(ctx context.Context, req *api.GetVoteActivityRankInternalReq, DSG *api.VoteDataSourceGroupItem, externalRes *model.RankResultExternal, displayRiskVote bool) (res *api.GetVoteActivityRankInternalResp, err error) {
	res = &api.GetVoteActivityRankInternalResp{
		Rank: make([]*api.InternalRankInfo, 0),
		Page: &api.VotePage{
			Num:   req.Pn,
			Ps:    req.Ps,
			Total: 0,
		},
	}
	//1.获取内部配置
	internalRes, err := d.rawInternalDSGVoteInfoById(ctx, DSG.GroupId, displayRiskVote)
	if err != nil {
		return
	}
	//2.合并
	internalResMap := make(map[int64]*api.InternalRankInfo)
	for _, r := range internalRes {
		tmpR := r
		internalResMap[tmpR.SourceItemId] = tmpR
	}
	for _, r := range externalRes.List {
		tmpR := &api.InternalRankInfo{
			ActivityId:     DSG.ActivityId,
			SourceGroupId:  DSG.GroupId,
			SourceItemId:   r.DataSourceItemId,
			SourceItemName: r.DataSourceItemName,
		}
		if internal, ok := internalResMap[r.DataSourceItemId]; ok {
			tmpR.Id = internal.Id
			tmpR.InterveneVoteCount = internal.InterveneVoteCount
			tmpR.UserVoteCount = internal.UserVoteCount
			tmpR.RiskVoteCount = internal.RiskVoteCount
			tmpR.TotalVoteCount = internal.TotalVoteCount
			tmpR.TotalVoteMtime = internal.TotalVoteMtime
			tmpR.InBlackList = internal.InBlackList
			tmpR.Ctime = internal.Ctime
			tmpR.Mtime = internal.Mtime
		}
		res.Rank = append(res.Rank, tmpR)
	}
	res.Page.Total = externalRes.Page.Total
	return
}

func (d *Dao) RenewCurrentInternalRankWithInfoCacheTTL(ctx context.Context, activityId int64) (err error) {
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "RenewCurrentInternalRankWithInfoCacheTTL d.CacheActivity error: %v", err)
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

func (d *Dao) getInternalVoteWithInfoExpireTime(activity *api.VoteActivity) (expireTime int64) {
	expireTime = int64(0)
	//结束90天内的活动每天刷新一次排名, 使用单独的TTL
	if activity.EndTime < time.Now().Unix() && activity.EndTime > time.Now().AddDate(0, 0, -90).Unix() {
		expireTime = d.outdatedVoteRankWithInfoExpire
		return
	}
	expireTime = d.adminVoteRankWithInfoExpire
	return expireTime
}

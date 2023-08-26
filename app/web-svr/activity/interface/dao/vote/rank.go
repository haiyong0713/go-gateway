package vote

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
)

type rankType int

const (
	//外部排序, 外部用户使用
	rankTypeExternal rankType = iota
	//内部排序, 管理后台使用
	rankTypeInternal
	esDataSourceGroupIdFieldName = "datasource_group_id"
	esDataSourceGroupDataVersion = "data_version"
)

var esSearchFields = []string{"search_field_1", "search_field_2", "search_field_3"}

func (d *Dao) innerCalcDSGRankTop(ctx context.Context, DSG *api.VoteDataSourceGroupItem, dsI DataSource, topN []*model.DataSourceItemVoteInfo, version int64) (res []*model.RankInfo) {
	res = make([]*model.RankInfo, 0)
	itemIds := make([]int64, 0)
	for _, v := range topN {
		itemIds = append(itemIds, v.SourceItemId)
	}
	itemInfos, err := d.innerCacheBatchGetDSGItemByConn(ctx, dsI, DSG.GroupId, itemIds, version)
	if err != nil {
		return
	}
	for _, v := range topN {
		inBlack, err := d.innerBlackListCheckByConn(ctx, DSG.GroupId, v.SourceItemId)
		if err == nil && inBlack {
			continue
		}

		item, ok := itemInfos[v.SourceItemId]
		if !ok { //底层数据源中的稿件已经删除
			//TODO: 根据ID去底层数据源再次确认一次(避免稿件过多导致排名无法展示)
			continue
		}
		res = append(res, &model.RankInfo{
			Data:               item,
			DataSourceGroupId:  DSG.GroupId,
			DataSourceItemId:   v.SourceItemId,
			DataSourceItemName: item.GetName(),
			Vote:               v.TotalVoteCount,
		})
	}
	return
}

func (d *Dao) innerCalcDSGRank(ctx context.Context, DSG *api.VoteDataSourceGroupItem, dsI DataSource, topN []*model.DataSourceItemVoteInfo) (res []*model.RankInfo, err error) {
	version := d.mustRedisGetCurrentDSCacheVersion(ctx, DSG.ActivityId)
	res = d.innerCalcDSGRankTop(ctx, DSG, dsI, topN, version)
	if len(res) >= maxRankSize {
		return
	}

	start := int64(0)
	end := int64(maxRankSize / 4)
	var nextN []int64
OUTER:
	for len(res) <= maxRankSize {
		nextN, err = d.innerDSGOriginalRankById(ctx, DSG.GroupId, start, end, version)
		if err != nil {
			return
		}
		if len(nextN) == 0 { //DataSource已经取完, 提前退出
			break OUTER
		}
		itemInfos, err := d.innerCacheBatchGetDSGItemByConn(ctx, dsI, DSG.GroupId, nextN, version)
		if err != nil {
			log.Errorc(ctx, "CalcAllRankExternal CacheGetDataSourceItem error: %v", err)
			continue
		}
		for _, itemId := range nextN {
			//黑名单检查
			inBlack, err := d.innerBlackListCheckByConn(ctx, DSG.GroupId, itemId)
			if err == nil && inBlack {
				continue
			}
			//此ID已在Top N中存在, 跳过
			if exist, _, _ := d.GetDSGItemVoteCount(ctx, DSG.GroupId, itemId); exist {
				continue
			}
			item, ok := itemInfos[itemId]
			if !ok {
				continue
			}
			res = append(res, &model.RankInfo{
				Data:               item,
				DataSourceGroupId:  DSG.GroupId,
				DataSourceItemId:   itemId,
				DataSourceItemName: item.GetName(),
				Vote:               0,
			})
			if len(res) >= maxRankSize {
				break OUTER
			}
		}

		start = end + 1
		end = start + int64(maxRankSize/10)
	}

	return
}

func (d *Dao) getRankVersionAndCacheKey(ctx context.Context, activityId, sourceGroupId, oldVersion int64, rankTyp rankType) (newVersion int64, cacheKey string, err error) {
	if oldVersion != 0 { //主动指定了版本, 直接返回
		newVersion = oldVersion
	} else {
		switch rankTyp {
		case rankTypeInternal:
			newVersion = d.mustGetCurrRedisInternalRankWithInfoCacheVersion(ctx, activityId)
		case rankTypeExternal:
			newVersion = d.mustGetCurrRedisExternalRankWithInfoCacheVersion(ctx, activityId)
		default:
			err = ecode.ActivityVoteRankExpired
			return
		}
	}

	if newVersion == 0 {
		err = ecode.ActivityVoteRankExpired
		return
	}

	switch rankTyp {
	case rankTypeInternal:
		cacheKey = d.redisRankInternalWithInfoCacheKeyByVersion(sourceGroupId, newVersion)
	case rankTypeExternal:
		cacheKey = d.redisRankExternalWithInfoCacheKeyByVersion(sourceGroupId, newVersion)
	default:
		err = ecode.ActivityVoteRankExpired
		return
	}
	exists, err := redis.Bool(d.redis.Do(ctx, "EXISTS", cacheKey))
	if err != nil {
		return
	}
	if !exists {
		err = ecode.ActivityVoteRankExpired
		return
	}
	return
}

func (d *Dao) getDSVersionAndCacheKey(ctx context.Context, activityId, sourceGroupId, oldVersion int64) (newVersion int64, zsetKey, setKey string, err error) {
	if oldVersion == 0 {
		newVersion = d.mustRedisGetCurrentDSCacheVersion(ctx, activityId)
	} else {
		newVersion = oldVersion
	}
	if newVersion == 0 {
		err = ecode.ActivityVoteRankExpired
		return
	}

	zsetKey = d.redisDSGSortZsetKeyByVersion(sourceGroupId, newVersion)
	setKey = d.redisDSGSetKeyByVersion(sourceGroupId, newVersion)
	exists, err := redis.Bool(d.redis.Do(ctx, "EXISTS", zsetKey))
	if err != nil {
		return
	}
	if !exists {
		err = ecode.ActivityVoteRankExpired
		return
	}
	return
}

// GetDSGRankExternal: 投票排名(票数排名).
func (d *Dao) innerGetDSGRank(ctx context.Context, params *model.InnerRankParams, rankTyp rankType) (res *model.RankResultExternal, err error) {
	res = &model.RankResultExternal{
		VoteRankVersion:    0,
		UserAvailVoteCount: 0,
		DataSourceType:     "",
		DataSourceGroupId:  0,
		List:               make([]*model.RankInfo, 0),
		Page: &model.Page{
			Pn: params.Pn,
			Ps: params.Ps,
		},
	}
	DSG, activity, err := d.voteListCheckDSGAndActivity(ctx, params.DataSourceGroupId, params.ActivityId)
	if err != nil {
		return
	}
	res.VoteRankType = activity.Rule.VoteUpdateRule
	res.DataSourceType = DSG.SourceType
	res.DataSourceGroupId = DSG.GroupId
	version, cacheKey, err := d.getRankVersionAndCacheKey(ctx, activity.Id, params.DataSourceGroupId, params.Version, rankTyp)
	if err != nil {
		return
	}
	params.Version = version
	res.VoteRankVersion = version

	rankCount, err := redis.Int64(d.redis.Do(ctx, "LLEN", cacheKey))
	if err != nil {
		return
	}
	res.Page.Total = rankCount
	start := (params.Pn - 1) * params.Ps
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "LRANGE", cacheKey, start, start+params.Ps-1))
	if err != nil {
		return
	}

	var exist bool
	var vote int64
	for _, bs := range bss {
		tmp := &model.RankInfo{}
		err = json.Unmarshal(bs, &tmp)
		if err != nil {
			log.Errorc(ctx, "GetDSGRankExternal json.Unmarshal %v error: %v", bs, err)
		}
		if err == nil {
			//票数外显逻辑
			if rankTyp == rankTypeInternal || activity.Rule.DisplayVoteCount { //显示票数, 或者是来自admin后台的请求
				exist, vote, err = d.GetDSGItemVoteCount(ctx, DSG.GroupId, tmp.DataSourceItemId) //从缓存中重新获取一次票数
				if exist && err == nil {
					tmp.Vote = vote
				}
				if tmp.Vote < 0 {
					tmp.Vote = 0
				}
			} else {
				//不显示票数, 置为-1
				tmp.Vote = -1
			}
			res.List = append(res.List, tmp)
		}
	}
	if params.Mid != 0 {
		var availCount int64
		var extraCount int64
		availCount, extraCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, params.ActivityId, params.Mid)
		if err != nil {
			return
		}
		res.UserAvailVoteCount = availCount
		res.UserExtraAvailVoteCount = extraCount
		err = d.batchGetUserVoteCountForDSGItem(ctx, activity.Rule, params.DataSourceGroupId, params.Mid, res.List)
	}

	return
}

// innerGetDSGRankOrder: 投票排名(时间排序/随机排序).
func (d *Dao) innerGetDSGRankOrder(ctx context.Context, rand bool, params *model.InnerRankParams, rankTyp rankType) (res *model.RankResultExternal, err error) {
	res = &model.RankResultExternal{
		VoteRankVersion:    0,
		UserAvailVoteCount: 0,
		DataSourceType:     "",
		DataSourceGroupId:  0,
		List:               make([]*model.RankInfo, 0),
		Page: &model.Page{
			Pn: params.Pn,
			Ps: params.Ps,
		},
	}
	DSG, activity, err := d.voteListCheckDSGAndActivity(ctx, params.DataSourceGroupId, params.ActivityId)
	if err != nil {
		log.Errorc(ctx, "GetDSGRankExternal d.voteCheckDSGAndActivity error: %v", err)
		return
	}
	dsI, ok := d.datasourceMap[DSG.SourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	res.VoteRankType = activity.Rule.VoteUpdateRule
	res.DataSourceType = DSG.SourceType
	res.DataSourceGroupId = DSG.GroupId
	newVersion, zsetKey, setKey, err := d.getDSVersionAndCacheKey(ctx, activity.Id, params.DataSourceGroupId, params.Version)
	if err != nil {
		return
	}
	res.VoteRankVersion = newVersion

	rankCount, err := redis.Int64(d.redis.Do(ctx, "ZCOUNT", zsetKey, "-inf", "+inf"))
	if err != nil {
		return
	}
	res.Page.Total = rankCount
	start := (params.Pn - 1) * params.Ps
	var ids []int64
	if !rand {
		ids, err = redis.Int64s(d.redis.Do(ctx, "ZRANGE", zsetKey, start, start+params.Ps-1))
	} else {
		ids, err = redis.Int64s(d.redis.Do(ctx, "SRANDMEMBER", setKey, params.Ps))
	}

	if err != nil {
		return
	}
	var (
		exist bool
		vote  int64
	)

	itemInfos, err := d.innerCacheBatchGetDSGItemByConn(ctx, dsI, DSG.GroupId, ids, newVersion)
	if err != nil {
		return
	}
	for _, id := range ids {
		tmp, ok := itemInfos[id]
		if ok {
			t := &model.RankInfo{
				Data:               tmp,
				DataSourceGroupId:  DSG.GroupId,
				DataSourceItemId:   id,
				DataSourceItemName: tmp.GetName(),
				Vote:               0,
			}
			//票数外显逻辑
			if rankTyp == rankTypeInternal || activity.Rule.DisplayVoteCount { //显示票数, 或者是来自admin后台的请求
				exist, vote, err = d.GetDSGItemVoteCount(ctx, DSG.GroupId, id)
				if exist && err == nil {
					t.Vote = vote
				}
				if t.Vote < 0 {
					t.Vote = 0
				}
			} else {
				//不显示票数, 置为-1
				t.Vote = -1
			}

			res.List = append(res.List, t)
		}
	}
	if params.Mid != 0 {
		var availCount int64
		var extraCount int64
		availCount, extraCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, params.ActivityId, params.Mid)
		if err != nil {
			return
		}
		res.UserAvailVoteCount = availCount
		res.UserExtraAvailVoteCount = extraCount
		err = d.batchGetUserVoteCountForDSGItem(ctx, activity.Rule, params.DataSourceGroupId, params.Mid, res.List)
	}

	return
}

func (d *Dao) Search(ctx context.Context, mid int64, req *model.RankSearchParams) (res *model.RankSearchResultExternal, err error) {
	res = &model.RankSearchResultExternal{
		VoteRankVersion:    0,
		UserAvailVoteCount: 0,
		DataSourceType:     "",
		DataSourceGroupId:  req.DataSourceGroupId,
		List:               make([]*model.RankInfo, 0),
	}

	DSG, err := d.DataSourceGroup(ctx, req.DataSourceGroupId)
	if err != nil {
		return
	}
	dsI, ok := d.datasourceMap[DSG.SourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	activity, err := d.Activity(ctx, DSG.ActivityId)
	if err != nil {
		return
	}
	version := d.mustRedisGetCurrentDSCacheVersion(ctx, DSG.ActivityId)
	res.VoteRankVersion = version
	res.DataSourceType = DSG.SourceType
	res.VoteRankType = activity.Rule.VoteUpdateRule

	esRes := &model.EsDataSourceSearchResult{}
	err = d.esClient.NewRequest(ESBizName).Index(ESIndexName).
		WhereEq(esDataSourceGroupIdFieldName, req.DataSourceGroupId).
		WhereEq(esDataSourceGroupDataVersion, version).
		WhereLike(esSearchFields, []string{req.KeyWord}, true, elastic.LikeLevelLow).
		Ps(req.Limit).Scan(ctx, esRes)
	if err != nil {
		return
	}
	for _, e := range esRes.Result {
		tmp, err := d.innerCacheGetDSGItemByConn(ctx, dsI, DSG.GroupId, e.DataSourceItemId, version)
		if err != nil {
			log.Errorc(ctx, "CalcAllRankExternal CacheGetDataSourceItem error: %v", err)
			continue
		}
		t := &model.RankInfo{
			Data:               tmp,
			DataSourceGroupId:  DSG.GroupId,
			DataSourceItemId:   e.DataSourceItemId,
			DataSourceItemName: tmp.GetName(),
			Vote:               0,
		}
		//票数外显逻辑
		if activity.Rule.DisplayVoteCount {
			exist, vote, err := d.GetDSGItemVoteCount(ctx, DSG.GroupId, e.DataSourceItemId)
			if exist && err == nil {
				t.Vote = vote
			}
			if t.Vote < 0 {
				t.Vote = 0
			}
		} else {
			//不显示票数, 置为-1
			t.Vote = -1
		}

		res.List = append(res.List, t)
	}
	if mid != 0 {
		var availCount int64
		var extraCount int64
		availCount, extraCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, DSG.ActivityId, mid)
		if err != nil {
			return
		}
		res.UserAvailVoteCount = availCount
		res.UserExtraAvailVoteCount = extraCount
		err = d.batchGetUserVoteCountForDSGItem(ctx, activity.Rule, req.DataSourceGroupId, mid, res.List)
	}
	return
}

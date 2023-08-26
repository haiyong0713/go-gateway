package vote

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
)

// SQL
const (
	sql4RawDSGBlackListById = `
SELECT /*master*/ source_item_id
FROM act_vote_data_source_items
WHERE source_group_id = ?
  AND in_blacklist= 1
`

	sql4AddBlackList = `
INSERT INTO act_vote_data_source_items(main_id, source_group_id, source_item_id, in_blacklist)
VALUES(?,?,?,1) ON DUPLICATE KEY
UPDATE in_blacklist = 1
`

	sql4DelBlackList = `
UPDATE act_vote_data_source_items
SET in_blacklist = 0
WHERE main_id=?
  AND source_group_id=?
  AND source_item_id=?
`

	sql4UpdateItemIntervene = `
INSERT INTO act_vote_data_source_items(main_id, source_group_id, source_item_id, intervene_votes)
VALUES(?,?,?,?) ON DUPLICATE KEY
UPDATE intervene_votes = ?
`
)

/**********************************缓存Key控制**********************************/
func redisDSGBlackListSetCacheKey(sourceGroupId int64) string {
	return fmt.Sprintf("vote_DSG_bl_Sets_%v", sourceGroupId)
}

/**********************************黑名单校验**********************************/
func (d *Dao) rawDSGBlackListById(ctx context.Context, sourceGroupId int64) (res []int64, err error) {
	res = make([]int64, 0)
	rows, err := d.db.Query(ctx, sql4RawDSGBlackListById, sourceGroupId)
	if err != nil {
		log.Errorc(ctx, "rawDSGBlackListById query error: %v", err)
		return
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		sourceItemId := int64(0)
		err = rows.Scan(&sourceItemId)
		if err != nil {
			log.Errorc(ctx, "rawDSGBlackListById scan error: %v", err)
			return
		}
		res = append(res, sourceItemId)
	}
	err = rows.Err()
	return
}

func (d *Dao) delDSGBlackListCacheById(ctx context.Context, sourceGroupId int64) (err error) {
	cacheKey := redisDSGBlackListSetCacheKey(sourceGroupId)
	err = retry.WithAttempts(ctx, "delDSGBlackListCacheById", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "DEL", cacheKey)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "delDSGBlackListCacheById SADD key (%v) error: %v", cacheKey, err)
		return
	}
	return
}

func (d *Dao) rebuildDSGBlackListCacheById(ctx context.Context, sourceGroupId int64) (err error) {
	var items []int64
	items, err = d.rawDSGBlackListById(ctx, sourceGroupId)
	if err != nil {
		return
	}
	return d.resetDSGBlackListCacheById(ctx, sourceGroupId, items)
}

func (d *Dao) resetDSGBlackListCacheById(ctx context.Context, sourceGroupId int64, blackList []int64) (err error) {
	cacheKey := redisDSGBlackListSetCacheKey(sourceGroupId)
	args := make([]interface{}, 0)
	args = append(args, cacheKey, DummyInt64Key)
	for _, s := range blackList {
		args = append(args, s)
	}
	err = retry.WithAttempts(ctx, "RebuildBlackListCache", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "SADD", args...)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "resetDSGBlackListCacheById SADD key (%v) error: %v", cacheKey, err)
		return
	}

	err = retry.WithAttempts(ctx, "innerRebuildDataSourceItemVoteInfosCacheByActivityId_EXPIRE", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "EXPIRE", cacheKey, d.blackListCacheExpire)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "resetDSGBlackListCacheById EXPIRE key (%v) error: %v", cacheKey, err)
		return
	}

	return
}

func (d *Dao) BlackListCheck(ctx context.Context, sourceGroupId, sourceItemId int64) (inBlackList bool, err error) {
	return d.innerBlackListCheckByConn(ctx, sourceGroupId, sourceItemId)
}

func (d *Dao) innerBlackListCheckByConn(ctx context.Context, sourceGroupId, sourceItemId int64) (inBlackList bool, err error) {
	setsCacheKey := redisDSGBlackListSetCacheKey(sourceGroupId)
	exist, err := redis.Bool(d.redis.Do(ctx, "EXISTS", setsCacheKey))
	if err != nil {
		log.Errorc(ctx, "innerBlackListCheckByConn EXISTS (%v) error: %v", setsCacheKey, err)
		return
	}
	if !exist {
		log.Infoc(ctx, "BlackListCheck cache for sourceGroupId:%v empty, try back to source", sourceGroupId)
		err = d.rebuildDSGBlackListCacheById(ctx, sourceGroupId)
		if err != nil {
			log.Errorc(ctx, "innerBlackListCheckByConn d.rebuildDSGBlackListCacheById(%v) error: %v", sourceGroupId, err)
			return
		}
	}
	inBlackList, err = redis.Bool(d.redis.Do(ctx, "SISMEMBER", setsCacheKey, sourceItemId))
	if err == redis.ErrNil {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "innerBlackListCheckByConn SISMEMBER (%v) (%v) error: %v", setsCacheKey, sourceItemId, err)
	}
	return
}

/**********************************黑名单CURD**********************************/

func (d *Dao) AddVoteActivityBlackList(ctx context.Context, req *api.AddVoteActivityBlackListReq) (err error) {
	activity, err := d.Activity(ctx, req.ActivityId)
	if err != nil {
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	_, err = d.db.Exec(ctx, sql4AddBlackList, req.ActivityId, req.SourceGroupId, req.SourceItemId)
	if err == nil {
		err = d.delDSGBlackListCacheById(ctx, req.SourceGroupId)
	}
	return
}

func (d *Dao) DelVoteActivityBlackList(ctx context.Context, req *api.DelVoteActivityBlackListReq) (err error) {
	activity, err := d.Activity(ctx, req.ActivityId)
	if err != nil {
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	res, err := d.db.Exec(ctx, sql4DelBlackList, req.ActivityId, req.SourceGroupId, req.SourceItemId)
	if err != nil {
		return
	}
	af, err := res.RowsAffected()
	if err != nil {
		return
	}
	if af == 0 {
		err = ecode.ActivityVoteItemNotFound
		return
	}
	err = d.delDSGBlackListCacheById(ctx, req.SourceGroupId)
	return
}

func (d *Dao) UpdateVoteActivityInterveneVoteCount(ctx context.Context, req *api.UpdateVoteActivityInterveneVoteCountReq) (err error) {
	activity, err := d.Activity(ctx, req.ActivityId)
	if err != nil {
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	_, err = d.db.Exec(ctx, sql4UpdateItemIntervene, req.ActivityId, req.SourceGroupId, req.SourceItemId, req.InterveneVoteCount, req.InterveneVoteCount)
	if err == nil {
		//实时更新票数, 刷新一次total
		if activity.Rule.VoteUpdateRule == int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime) {
			err = d.RefreshVoteActivityRankZset(ctx, &api.RefreshVoteActivityRankZsetReq{ActivityId: req.ActivityId})
		}
	}
	return
}

// GetBlackListItemInfo: 获取黑名单中的稿件信息(内部使用,使用前请确认,部分信息不完整)
func (d *Dao) GetBlackListItemInfo(ctx context.Context, DSG *api.VoteDataSourceGroupItem) (res []*model.RankInfo, err error) {
	res = make([]*model.RankInfo, 0)
	blackItemIds, err := d.rawDSGBlackListById(ctx, DSG.GroupId)
	if err != nil {
		return
	}
	dsI, ok := d.datasourceMap[DSG.SourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	version := d.mustRedisGetCurrentDSCacheVersion(ctx, DSG.ActivityId)
	itemInfos, err := d.innerCacheBatchGetDSGItemByConn(ctx, dsI, DSG.GroupId, blackItemIds, version)
	if err != nil {
		return
	}
	for _, id := range blackItemIds {
		item, ok := itemInfos[id]
		if ok {
			res = append(res, &model.RankInfo{
				Data:               item,
				DataSourceGroupId:  DSG.GroupId,
				DataSourceItemId:   id,
				DataSourceItemName: item.GetName(),
			})
		}
	}
	return
}

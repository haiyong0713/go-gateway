package vote

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"sort"
)

const (
	sql4RawInternalDSGVoteInfosById = `
SELECT id,
       main_id,
       source_group_id,
       source_item_id,
       votes,
       intervene_votes,
       risk_votes,
       total_votes,
       total_votes_mtime,
       in_blacklist,
       ctime,
       mtime
FROM act_vote_data_source_items
WHERE source_group_id = ?
ORDER BY total_votes DESC, source_item_id DESC
`

	sql4RawExternalDSGVoteInfosById = `
SELECT 
       source_item_id,
       total_votes
FROM act_vote_data_source_items
WHERE source_group_id = ?
ORDER BY total_votes DESC, source_item_id DESC
`
)

func (d *Dao) redisDSGItemVoteCountKey(sourceGroupId, sourceItemId int64) string {
	return fmt.Sprintf("vote_DSI_voteC_%v_%v", sourceGroupId, sourceItemId)
}

func (d *Dao) RebuildDSGVoteCountCache(ctx context.Context, sourceGroupId int64) (err error) {
	var items []*model.DataSourceItemVoteInfo
	items, err = d.rawExternalDSGVoteInfoById(ctx, sourceGroupId)
	if err != nil {
		return
	}
	return d.resetDSGItemVoteCountCache(ctx, sourceGroupId, items)
}

// rawExternalDSGVoteInfoById: 获取稿件投票信息(外部使用)
func (d *Dao) rawExternalDSGVoteInfoById(ctx context.Context, sourceGroupId int64) (res []*model.DataSourceItemVoteInfo, err error) {
	res = make([]*model.DataSourceItemVoteInfo, 0)
	rows, err := d.db.Query(ctx, sql4RawExternalDSGVoteInfosById, sourceGroupId)
	if err != nil {
		log.Errorc(ctx, "rawExternalDSGVoteInfoById for sourceGroupId %v error: %v", sourceGroupId, err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		tmp := &model.DataSourceItemVoteInfo{}
		err = rows.Scan(&tmp.SourceItemId, &tmp.TotalVoteCount)
		if err != nil {
			log.Errorc(ctx, "RawInternalDataSourceItemInfoByActivityId scan error: %v", err)
			return
		}
		res = append(res, tmp)
	}
	err = rows.Err()
	return
}

// rawInternalDSGVoteInfoById: 获取稿件投票信息(内部使用)
func (d *Dao) rawInternalDSGVoteInfoById(ctx context.Context, sourceGroupId int64, displayRiskVotes bool) (res []*api.InternalRankInfo, err error) {
	res = make([]*api.InternalRankInfo, 0)
	rows, err := d.db.Query(ctx, sql4RawInternalDSGVoteInfosById, sourceGroupId)
	if err != nil {
		log.Errorc(ctx, "rawExternalDSGVoteInfoById for sourceGroupId %v error: %v", sourceGroupId, err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		tmp := &api.InternalRankInfo{}
		err = rows.Scan(&tmp.Id, &tmp.ActivityId, &tmp.SourceGroupId, &tmp.SourceItemId,
			&tmp.UserVoteCount, &tmp.InterveneVoteCount, &tmp.RiskVoteCount,
			&tmp.TotalVoteCount, &tmp.TotalVoteMtime, &tmp.InBlackList,
			&tmp.Ctime, &tmp.Mtime)
		if err != nil {
			log.Errorc(ctx, "RawInternalDataSourceItemInfoByActivityId scan error: %v", err)
			return
		}
		//db中的total_votes有可能不是实时更新, 但后台需要看到实时数据.
		//所以此处重新计算并排序
		if displayRiskVotes {
			tmp.TotalVoteCount = tmp.UserVoteCount + tmp.InterveneVoteCount + tmp.RiskVoteCount
		} else {
			tmp.TotalVoteCount = tmp.UserVoteCount + tmp.InterveneVoteCount
		}
		res = append(res, tmp)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].TotalVoteCount != res[j].TotalVoteCount {
			return res[i].TotalVoteCount > res[j].TotalVoteCount
		}
		return res[i].SourceItemId > res[j].SourceItemId
	})
	return
}

func (d *Dao) resetDSGItemVoteCountCache(ctx context.Context, dataSourceGroupId int64,
	items []*model.DataSourceItemVoteInfo) (err error) {
	if len(items) == 0 {
		return
	}
	//1.写入缓存
	for _, item := range items {
		cacheKey := d.redisDSGItemVoteCountKey(dataSourceGroupId, item.SourceItemId)
		err = retry.WithAttempts(ctx, "innerRebuildDataSourceItemVoteInfosCacheByActivityId_ZADD", 5, netutil.DefaultBackoffConfig,
			func(c context.Context) (err error) {
				_, err = d.redis.Do(ctx, "SETEX", cacheKey, d.voteRankZsetExpire, item.TotalVoteCount)
				return err
			})
		if err != nil {
			log.Errorc(ctx, "resetDSGItemVoteCountCache SETEX key (%v) %v error: %v", cacheKey, item.SourceItemId, err)
			return
		}
	}
	return
}

func (d *Dao) IncrDSGItemVoteCountCache(ctx context.Context, sourceGroupId, sourceItemId, incr int64) {
	var err error
	cacheKey := d.redisDSGItemVoteCountKey(sourceGroupId, sourceItemId)
	err = retry.WithAttempts(ctx, "UpdateDataSourceItemVoteCache_ZINCRBY", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "INCRBY", cacheKey, incr)
		if err == redis.ErrNil {
			err = nil
		}
		return err
	})
	if err != nil {
		log.Errorc(ctx, "IncrDSGItemVoteCountCache for sourceGroupId: %v, sourceItemId: %v, error: %v", sourceGroupId, sourceItemId, err)
	}
}

func (d *Dao) GetDSGItemVoteCount(ctx context.Context, sourceGroupId, sourceItemId int64) (exist bool, vote int64, err error) {
	cacheKey := d.redisDSGItemVoteCountKey(sourceGroupId, sourceItemId)
	exist = true
	vote, err = redis.Int64(d.redis.Do(ctx, "GET", cacheKey))
	if err == redis.ErrNil {
		err = nil
		exist = false
	}
	if err != nil {
		log.Errorc(ctx, "GetDSGItemVoteCount conn.Do error: %v", err)
	}
	return
}

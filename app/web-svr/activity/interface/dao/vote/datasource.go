package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"strconv"
	"time"

	"go-common/library/sync/errgroup.v2"
)

// ES
const (
	ESBizName   = "vote_new"
	ESIndexName = "vote_new"
)

// SQL
const (
	sql4UpdateDataSourceVoteItemVote = `
INSERT INTO act_vote_data_source_items(main_id, source_group_id, source_item_id, votes)
VALUES(?,?,?,?)ON duplicate KEY
UPDATE votes=votes+?
`
	sql4UpdateDataSourceVoteItemRiskVote = `
INSERT INTO act_vote_data_source_items(main_id, source_group_id, source_item_id, risk_votes)
VALUES(?,?,?,?)ON duplicate KEY
UPDATE risk_votes=risk_votes+?
`

	sql4UpdateDataSourceTotalVoteItemVoteWithoutRisk = `
UPDATE act_vote_data_source_items
SET total_votes=votes+intervene_votes
WHERE source_group_id=?
  AND source_item_id=?
`

	sql4UpdateDataSourceTotalVoteItemVoteWithRisk = `
UPDATE act_vote_data_source_items
SET total_votes=votes+intervene_votes+risk_votes
WHERE source_group_id=?
  AND source_item_id=?
`

	sql4UpdateDSTotalVoteItemVoteWithoutRiskByMainId = `
UPDATE act_vote_data_source_items
SET total_votes=votes+intervene_votes
WHERE main_id=?
`

	sql4UpdateDSTotalVoteItemVoteWithRiskByMainId = `
UPDATE act_vote_data_source_items
SET total_votes=votes+intervene_votes+risk_votes
WHERE main_id=?
`
)

// dummy
const (
	DummyInt64Key = -888
)

type DataSource interface {
	//TODO: 最多返回N条
	ListAllItems(ctx context.Context, sourceId int64) ([]DataSourceItem, error)
	NewEmptyItem() DataSourceItem
}

type DataSourceItem interface {
	GetName() string
	GetId() int64
	GetSearchField1() string
	GetSearchField2() string
	GetSearchField3() string
}

//部分缩写:
//DSG -> DataSourceGroup, 数据组

/**********************************缓存Key控制**********************************/
//redisDSCacheVersionKey: 控制当前DSGCache的版本, 实现刷新DSG后快速切换Cache版本功能
//curr* -> 获取当前最新版本的Cache Key
func redisDSCacheVersionKey(activityId int64) string {
	return fmt.Sprintf("vote_DSGCache_version_%v", activityId)
}

func (d *Dao) mustRedisGetCurrentDSCacheVersion(ctx context.Context, activityId int64) (version int64) {
	var err error
	err = retry.Infinite(ctx, "mustRedisGetCurrentDSCacheVersion", netutil.DefaultBackoffConfig, func(c context.Context) error {
		version, err = redis.Int64(d.redis.Do(ctx, "GET", redisDSCacheVersionKey(activityId)))
		if err == redis.ErrNil {
			err = nil
		}
		return err
	})
	if err != nil {
		log.Errorc(ctx, "mustRedisGetCurrentDSCacheVersion error: %v", err)
	}
	return
}

func (d *Dao) currRedisDSGItemsMapCacheKey(ctx context.Context, sourceGroupId, activityId int64) string {
	return d.redisDSGItemsMapCacheKeyByVersion(sourceGroupId, d.mustRedisGetCurrentDSCacheVersion(ctx, activityId))
}

func (d *Dao) redisDSGItemsMapCacheKeyByVersion(dataSourceGroupId, version int64) string {
	return fmt.Sprintf("vote_DSGI_hashes_%v_v%v", dataSourceGroupId, version)
}

func (d *Dao) redisDSGSortZsetKeyByVersion(sourceGroupId, version int64) string {
	return fmt.Sprintf("vote_DSGI_zset_%v_v%v", sourceGroupId, version)
}

func (d *Dao) redisDSGSetKeyByVersion(sourceGroupId, version int64) string {
	return fmt.Sprintf("vote_DSGI_set_%v_v%v", sourceGroupId, version)
}

/**********************************数据源数据同步**********************************/
//getDSItemsExpireTime: 计算底层稿件数据的缓存过期时间
func (d *Dao) getDSItemsExpireTime(activity *api.VoteActivity) (expireTime int64) {
	//结束90天内的活动每天刷新一次排名, 使用单独的TTL
	if activity.EndTime < time.Now().Unix() && activity.EndTime > time.Now().AddDate(0, 0, -90).Unix() {
		expireTime = d.outdatedDataSourceItemsInfoCacheExpire
		return
	}
	return d.dataSourceItemsInfoCacheExpire
}

// RefreshVoteActivityDSItems: 同步投票活动的底层稿件数据到投票层.
// 投票层包含的数据:
// 1.所有底层数据源内的稿件排序信息, []int64(int64为稿件ID) -> Redis Sorted Set(score为数组下标)
// 2.所有底层数据源内的稿件详情, map[int64]interface{} -> Redis Hashes
func (d *Dao) RefreshVoteActivityDSItems(ctx context.Context, activityId int64) (err error) {
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "RefreshVoteActivityDSItems d.CacheActivity error: %v", err)
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}

	dataSources, err := d.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: activityId})
	if err != nil {
		return
	}
	if len(dataSources.Groups) == 0 {
		return
	}
	defer func() {
		if err != nil {
			err = ecode.ActivityVoteRefreshDSGItemsFail
			if err1 := d.RenewCurrentDSGItemCacheTTL(ctx, activityId); err1 != nil {
				err = ecode.ActivityVoteRefreshAndRenewDSGItemsFail
			}
		}
	}()
	//1.获取新版本号(当前缓存版本号+1)
	newCacheVersion, _ := strconv.ParseInt(time.Now().Format(cacheVersionFormat), 10, 64)
	//2.刷新所有DSG
	cacheExpireTime := d.getDSItemsExpireTime(activity)
	eg := errgroup.WithContext(ctx)
	for _, dsg := range dataSources.Groups {
		//1.根据数据源类型获取底层数据源的交互函数
		tmpDsg := dsg
		dsI, ok := d.datasourceMap[dsg.SourceType]
		if !ok {
			err = ecode.ActivityVoteSourceTypeUnknown
			return
		}
		eg.Go(func(ctx context.Context) error {
			return d.innerRefreshDSGItemsCache(ctx, tmpDsg, dsI, newCacheVersion, cacheExpireTime)
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}
	//3.更新缓存版本号
	{
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_IncrVersion", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "SET", redisDSCacheVersionKey(activityId), newCacheVersion)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems IncrVersion error: %v", err)
		}
	}
	return
}

// innerRefreshDSGItemsCache: 同步特定数据组的稿件数据到指定版本的Cache中
func (d *Dao) innerRefreshDSGItemsCache(ctx context.Context, dsg *api.VoteDataSourceGroupItem, dsI DataSource, newCacheVersion, cacheExpireTime int64) (err error) {

	var (
		itemSortCacheKey   = d.redisDSGSortZsetKeyByVersion(dsg.GroupId, newCacheVersion)
		newSortingZsetArgs = []interface{}{itemSortCacheKey}
		items              []DataSourceItem
		itemsSort          []int64
	)
	//1.拉取底层数据源的稿件信息
	{
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_ListAllItems", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			items, err = dsI.ListAllItems(ctx, dsg.SourceId)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems ds.ListAllItems sourceType %v, sourceId %v, error: %v", dsg.SourceType, dsg.SourceId, err)
			return
		}
	}
	if len(items) == 0 {
		log.Infoc(ctx, "RefreshVoteActivityDSItems return because datasource is empty")
		return
	}
	//2.更新稿件信息到Hashes(快速获取稿件信息用)
	{
		redisDSItemsMapCacheKey := d.redisDSGItemsMapCacheKeyByVersion(dsg.GroupId, newCacheVersion)
		var bs []byte
		var inBlackList bool
		for _, item := range items {
			inBlackList, err = d.BlackListCheck(ctx, dsg.GroupId, item.GetId())
			if err == nil && !inBlackList {
				//在黑名单中的稿件,只同步信息到hashes,不放到排序Set中(防止按时间排序看到拉黑的稿件)
				itemsSort = append(itemsSort, item.GetId())
				newSortingZsetArgs = append(newSortingZsetArgs, len(newSortingZsetArgs), item.GetId())
			} else {
				log.Infoc(ctx, "RefreshVoteActivityDSItems skip item: %v because err =%v, and inBlackList=%v", item.GetId(), err, inBlackList)
			}

			bs, err = json.Marshal(item)
			if err != nil {
				log.Errorc(ctx, "RefreshVoteActivityDSItems json.Marshal %+v error: %v", item, err)
				continue
			}
			err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_SetNewItem", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "HSET", redisDSItemsMapCacheKey, item.GetId(), bs)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewItem error: %v", err)
				return
			}
		}
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_SetNewItemTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "EXPIRE", redisDSItemsMapCacheKey, cacheExpireTime)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewItemTTL error: %v", err)
			return
		}
	}

	//3.更新稿件原始排序到Sorted Set
	{
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_SetNewRank", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "ZADD", newSortingZsetArgs...)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewRank args: %v, error: %v", newSortingZsetArgs, err)
			return
		}
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_SetNewRankTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "EXPIRE", itemSortCacheKey, cacheExpireTime)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewRankTTL error: %v", err)
			return
		}

	}

	//3.更新所有稿件ID到Set(随机排名用)
	{
		setCacheKey := d.redisDSGSetKeyByVersion(dsg.GroupId, newCacheVersion)
		newSet := []interface{}{setCacheKey}
		for _, itemId := range itemsSort {
			newSet = append(newSet, itemId)
		}
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_SetNewSet", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "SADD", newSet...)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewSet error: %v", err)
			return
		}
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_SetNewSetTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
			_, err = d.redis.Do(ctx, "EXPIRE", setCacheKey, cacheExpireTime)
			return err
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewSetTTL error: %v", err)
			return
		}

	}

	//4.写入所有稿件信息到ES(搜索用)
	{
		upt := d.esClient.NewUpdate(ESBizName)
		for _, item := range items {
			upt.AddData(ESIndexName, &model.EsDataSourceItem{
				ID:                fmt.Sprintf("%v-%v-%v", dsg.GroupId, item.GetId(), newCacheVersion),
				DataSourceGroupId: dsg.GroupId,
				DataSourceItemId:  item.GetId(),
				SearchField1:      item.GetSearchField1(),
				SearchField2:      item.GetSearchField2(),
				SearchField3:      item.GetSearchField3(),
				DataVersion:       newCacheVersion,
				WriteTime:         time.Now().Format("2006-01-02 15:04:05"),
			})
		}
		err = retry.WithAttempts(ctx, "RefreshCacheDatasourceItem_WriteEs", 5, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
			err = upt.Insert().Do(ctx)
			return
		})
		if err != nil {
			log.Errorc(ctx, "RefreshVoteActivityDSItems write es req: %v error: %v", upt.Params(), err)
			return
		}

	}
	return
}

// GetDSGItem: 获取指定数据组内的指定稿件信息
func (d *Dao) GetDSGItem(ctx context.Context, DSG *api.VoteDataSourceGroupItem, sourceItemId int64) (item DataSourceItem, err error) {
	dsI, ok := d.datasourceMap[DSG.SourceType]
	if !ok {
		err = ecode.ActivityVoteSourceTypeUnknown
		return
	}
	return d.innerCacheGetDSGItemByConn(ctx, dsI, DSG.GroupId, sourceItemId, d.mustRedisGetCurrentDSCacheVersion(ctx, DSG.ActivityId))
}

// innerCacheGetDSGItemByConn: 获取指定数据组内的指定稿件信息
// 抽取inner和conn作为参数的目的: redis连接复用, 减少刷新排名时的连接建立成本
func (d *Dao) innerCacheGetDSGItemByConn(ctx context.Context, dsI DataSource, sourceGroupId, sourceItemId, version int64) (item DataSourceItem, err error) {
	key := d.redisDSGItemsMapCacheKeyByVersion(sourceGroupId, version)
	item = dsI.NewEmptyItem()
	bs, err := redis.Bytes(d.redis.Do(ctx, "HGET", key, sourceItemId))
	if err == redis.ErrNil {
		err = ecode.ActivityVoteItemNotFound
		return
	}
	if err != nil {
		log.Errorc(ctx, "innerCacheGetDSGItemByConn conn.Do error: %v", err)
		return
	}
	err = json.Unmarshal(bs, item)
	if err != nil {
		log.Errorc(ctx, "innerCacheGetDSGItemByConn json.Unmarshal error: %v", err)
	}
	return
}

// innerCacheBatchGetDSGItemByConn: 批量获取指定数据组内的指定稿件信息
// 抽取inner和conn作为参数的目的: redis连接复用, 减少刷新排名时的连接建立成本
func (d *Dao) innerCacheBatchGetDSGItemByConn(ctx context.Context, dsI DataSource, sourceGroupId int64, sourceItemIds []int64, version int64) (items map[int64]DataSourceItem, err error) {
	items = make(map[int64]DataSourceItem)
	if len(sourceItemIds) == 0 {
		return
	}
	key := d.redisDSGItemsMapCacheKeyByVersion(sourceGroupId, version)
	args := redis.Args{}.Add(key).AddFlat(sourceItemIds)
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "HMGET", args...))
	if err == redis.ErrNil {
		err = ecode.ActivityVoteItemNotFound
		return
	}
	if err != nil {
		log.Errorc(ctx, "innerCacheGetDSGItemByConn conn.Do error: %v", err)
		return
	}
	for _, bs := range bss {
		if len(bs) == 0 {
			continue
		}
		item := dsI.NewEmptyItem()
		err = json.Unmarshal(bs, item)
		if err != nil {
			log.Errorc(ctx, "innerCacheGetDSGItemByConn json.Unmarshal error: %v", err)
		}
		items[item.GetId()] = item
	}
	return
}

// innerDSGOriginalRankById: 获取数据组内的原始稿件排序, 作为没有票数时的排序顺序
func (d *Dao) innerDSGOriginalRankById(ctx context.Context, sourceGroupId, start, end, version int64) (res []int64, err error) {
	zsetKey := d.redisDSGSortZsetKeyByVersion(sourceGroupId, version)
	res, err = redis.Int64s(d.redis.Do(ctx, "ZRANGE", zsetKey, start, end))
	if err == redis.ErrNil {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "innerDSGOriginalRankById conn.Do error: %v", err)
	}
	return
}

// IsDSItemExists: 判断特定数据组下的特定稿件是否存在
func (d *Dao) IsDSItemExists(ctx context.Context, activityId, sourceGroupId, sourceItemId int64) (exist bool, err error) {
	zsetCacheKey := d.currRedisDSGItemsMapCacheKey(ctx, sourceGroupId, activityId)
	exist, err = redis.Bool(d.redis.Do(ctx, "HEXISTS", zsetCacheKey, sourceItemId))
	if err != nil {
		log.Errorc(ctx, "IsDSItemExists error: %v", err)
	}
	return
}

func (d *Dao) RenewCurrentDSGItemCacheTTL(ctx context.Context, activityId int64) (err error) {
	activity, err := d.Activity(ctx, activityId)
	if err != nil {
		log.Errorc(ctx, "RefreshVoteActivityDSItems d.CacheActivity error: %v", err)
		return
	}
	if activity == nil {
		err = ecode.ActivityVoteNotFound
		return
	}
	currVersion := d.mustRedisGetCurrentDSCacheVersion(ctx, activityId)
	if currVersion == 0 {
		err = ecode.ActivityVoteItemNotFound
		return
	}
	dataSources, err := d.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: activityId})
	if err != nil {
		return
	}
	var eg errgroup.Group
	for _, dsg := range dataSources.Groups {
		tmpDsg := dsg
		eg.Go(func(ctx context.Context) (err error) {
			redisDSItemsMapCacheKey := d.redisDSGItemsMapCacheKeyByVersion(tmpDsg.GroupId, currVersion)
			err = retry.WithAttempts(ctx, "RenewCurrentDSGItemCacheTTL_SetNewItemTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "EXPIRE", redisDSItemsMapCacheKey, d.dataSourceItemsInfoCacheExpire)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewItemTTL error: %v", err)
				return
			}

			itemSortCacheKey := d.redisDSGSortZsetKeyByVersion(tmpDsg.GroupId, currVersion)
			err = retry.WithAttempts(ctx, "RenewCurrentDSGItemCacheTTL_SetNewRankTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "EXPIRE", itemSortCacheKey, d.dataSourceItemsInfoCacheExpire)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewRankTTL error: %v", err)
				return
			}

			setCacheKey := d.redisDSGSetKeyByVersion(tmpDsg.GroupId, currVersion)
			err = retry.WithAttempts(ctx, "RenewCurrentDSGItemCacheTTL_SetNewSetTTL", 5, netutil.DefaultBackoffConfig, func(c context.Context) error {
				_, err = d.redis.Do(ctx, "EXPIRE", setCacheKey, d.dataSourceItemsInfoCacheExpire)
				return err
			})
			if err != nil {
				log.Errorc(ctx, "RefreshVoteActivityDSItems SetNewSetTTL error: %v", err)
				return
			}
			return
		})
	}
	err = eg.Wait()
	return
}

package vote

import (
	"context"
	xsql "database/sql"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	riskModel "go-gateway/app/web-svr/activity/interface/model/risk"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"time"

	"github.com/pkg/errors"
)

// SQL
const (
	sql4GetUserVoteCountInActivity = `
SELECT votes
FROM act_vote_user_summary_%v
WHERE main_id = ?
  AND mid = ?
`

	sql4InsertUserVoteSummary = `
INSERT INTO act_vote_user_summary_%v (main_id, mid)
VALUES(?,?) 
`
	sql4GetUserTodayVoteCountForItem = `
SELECT COALESCE(sum(votes),0)
FROM act_vote_user_action_%v FORCE INDEX (ix_sgid_mid)
WHERE mid = ?
  AND source_group_id = ?
  AND is_undo = 0
  AND source_item_id = ?
  AND DATE(ctime) = CURRENT_DATE()
`

	sql4GetUserVoteCountForItem = `
SELECT COALESCE(sum(votes),0)
FROM act_vote_user_action_%v FORCE INDEX (ix_sgid_mid)
WHERE mid = ?
  AND source_group_id = ?
  AND is_undo = 0
  AND source_item_id = ?
`

	sql4UpdateUserVoteCountInActivity = `
UPDATE act_vote_user_summary_%v
SET votes = votes + ?
WHERE main_id=?
  AND mid = ?
`
	sql4LockUserVoteSummary = `
SELECT votes
FROM act_vote_user_summary_%v
WHERE main_id = ?
  AND mid = ?
  FOR
  UPDATE
`

	sql4InsertUserVoteRecord = `
INSERT INTO act_vote_user_action_%v 
(main_id, source_group_id, source_item_id, mid, had_risk, votes, times_type)
VALUES(?,?,?,?,?,?,?)
`
	sql4BatchGetUserVoteCountForDSGItems = `
SELECT source_item_id,
       sum(votes)
FROM act_vote_user_action_%v FORCE INDEX (ix_sgid_mid)
WHERE source_group_id=?
  AND mid = ?
  AND is_undo = 0
GROUP BY source_group_id,
         source_item_id
`

	sql4BatchGetUserTodayVoteCountForDSGItems = `
SELECT source_item_id,
       sum(votes)
FROM act_vote_user_action_%v FORCE INDEX (ix_sgid_mid)
WHERE source_group_id=?
  AND mid = ?
  AND is_undo = 0
  AND DATE(ctime) = CURRENT_DATE()
GROUP BY source_group_id,
         source_item_id
`

	sql4UndoUserTodayVoteForDSG = `
UPDATE act_vote_user_action_%v
SET is_undo=1
WHERE source_group_id=?
  AND source_item_id=?
  AND mid = ?
  AND is_undo = 0
  AND DATE(ctime) = CURRENT_DATE()
`

	sql4GetUserTodayVoteCountForDSGForUndo = `
SELECT had_risk,
       sum(votes)
FROM act_vote_user_action_%v
WHERE source_group_id=?
  AND source_item_id=?
  AND mid = ?
  AND is_undo = 0
  AND DATE(ctime) = CURRENT_DATE()
GROUP BY had_risk
`
	sql4GetUserTodayTimesCountForDSGForUndo = `
SELECT times_type,
       sum(votes)
FROM act_vote_user_action_%v
WHERE source_group_id=?
  AND source_item_id=?
  AND mid = ?
  AND is_undo = 0
  AND DATE(ctime) = CURRENT_DATE()
GROUP BY times_type
`
)

func tableIdx(id int64) string {
	return fmt.Sprintf("%02d", id%100)
}

// redisUserVoteCountCacheKeyForActivity: 用户在某活动下的总投票次数
func redisUserVoteCountCacheKeyForActivity(activityId, mid int64) string {
	return fmt.Sprintf("user_vote_count_act_%v_%v", activityId, mid)
}

// redisUserTodayVoteCountCacheKeyForActivity: 用户当天内在某活动下的总投票次数
func redisUserTodayVoteCountCacheKeyForActivity(activityId, mid int64) string {
	return fmt.Sprintf("user_vote_count_%v_act_%v_%v", time.Now().Format("20060102"), activityId, mid)
}

// redisUserTodayBaseVoteCountCacheKeyForActivity: 用户当天内在某活动下消耗的基础投票次数
func redisUserTodayBaseVoteCountCacheKeyForActivity(activityId, mid int64) string {
	return fmt.Sprintf("user_vote_count_base_%v_act_%v_%v", time.Now().Format("20060102"), activityId, mid)
}

// redisUserVoteCountCacheKeyForDSGItem: 用户对某个投票项的总投票次数
func redisUserVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid int64) string {
	return fmt.Sprintf("user_vote_count_DSGItem_%v_%v_%v", sourceGroupId, sourceItemId, mid)
}

// redisUserTodayVoteCountCacheKeyForDSGItem: 用户对某个投票项的当日投票次数
func redisUserTodayVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid int64) string {
	return fmt.Sprintf("user_vote_count_%v_DSGItem_%v_%v_%v", time.Now().Format("20060102"), sourceGroupId, sourceItemId, mid)
}

// RawUserVoteCountForActivity: 从DB中获取某用户在活动下累计投票次数
func (d *Dao) RawUserVoteCountForActivity(ctx context.Context, activityId, mid int64) (count int64, err error) {
	err = d.db.QueryRow(ctx, fmt.Sprintf(sql4GetUserVoteCountInActivity, tableIdx(mid)), activityId, mid).Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "RawUserVoteCountForActivity db.QueryRow mid:%v, activity:%v error: %v", mid, activityId, err)
	}
	return
}

// CacheGetUserVoteCountForActivity: 获取用户在整个活动下的投票总次数
func (d *Dao) CacheGetUserVoteCountForActivity(ctx context.Context, activityId, mid int64) (count int64, err error) {
	cacheKey := redisUserVoteCountCacheKeyForActivity(activityId, mid)
	shouldUpdateCache := false
	defer func() {
		if shouldUpdateCache {
			_, err1 := d.redis.Do(ctx, "SETEX", cacheKey, d.userVoteCountExpire, count)
			if err1 != nil {
				log.Errorc(ctx, "CacheGetUserVoteCountForActivity error: %v", err1)
			}
		}
	}()
	count, err = redis.Int64(d.redis.Do(ctx, "GET", cacheKey))
	if err == nil {
		return
	}
	count, err = d.RawUserVoteCountForActivity(ctx, activityId, mid)
	shouldUpdateCache = err == nil
	return
}

// CacheGetUserTodayVoteCountForActivity: 获取用户在整个活动下的当日投票总次数
func (d *Dao) CacheGetUserTodayVoteCountForActivity(ctx context.Context, activityId, mid int64) (baseCount, total int64, err error) {
	cacheKey := redisUserTodayVoteCountCacheKeyForActivity(activityId, mid)
	baseCacheKey := redisUserTodayBaseVoteCountCacheKeyForActivity(activityId, mid)
	shouldUpdateCache := false
	defer func() {
		if shouldUpdateCache {
			_, err = d.redis.Do(ctx, "SETEX", cacheKey, d.userVoteCountExpire, total)
			if err != nil {
				log.Errorc(ctx, "CacheGetUserVoteCountForActivity set cache error: %v", err)
			}

		}
	}()
	total, err = redis.Int64(d.redis.Do(ctx, "GET", cacheKey))
	if err == nil {
		baseCount, err = redis.Int64(d.redis.Do(ctx, "GET", baseCacheKey))
		if err == nil {
			return
		}
	}
	baseCount, _, _, total, err = d.RawUserTodayVoteCountForActivity(ctx, nil, activityId, mid)
	shouldUpdateCache = err == nil
	return
}

// CacheIncrUserVoteCountForActivity: 增加缓存中用户在整个活动下的投票总次数
func (d *Dao) CacheIncrUserVoteCountForActivity(ctx context.Context, activityId, mid, incr int64) {
	var err error
	cacheKey := redisUserVoteCountCacheKeyForActivity(activityId, mid)
	err = retry.WithAttempts(ctx, "CacheIncrUserVoteCountForActivity", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "INCRBY", cacheKey, incr))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserVoteCountForActivity for activityId: %v,mid: %v,incr: %v error: %v", activityId, mid, incr, err)
	}
}

// CacheDelUserVoteCountForActivity: 删除缓存中用户在整个活动下的投票总次数
func (d *Dao) CacheDelUserVoteCountForActivity(ctx context.Context, activityId, mid int64) {
	var err error
	cacheKey := redisUserVoteCountCacheKeyForActivity(activityId, mid)
	err = retry.WithAttempts(ctx, "CacheDelUserVoteCountForActivity", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "DEL", cacheKey))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheDelUserVoteCountForActivity for activityId: %v,mid: %v error: %v", activityId, mid, err)
	}
}

// CacheIncrUserTodayVoteCountForActivity: 增加缓存中用户在整个活动下的投票总次数
func (d *Dao) CacheIncrUserTodayVoteCountForActivity(ctx context.Context, activityId, mid, baseIncr, totalIncr int64) {
	var err error
	cacheKey := redisUserTodayVoteCountCacheKeyForActivity(activityId, mid)
	baseCacheKey := redisUserTodayBaseVoteCountCacheKeyForActivity(activityId, mid)
	err = retry.WithAttempts(ctx, "CacheIncrUserTodayVoteCountForActivity", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "INCRBY", cacheKey, totalIncr))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserTodayVoteCountForActivity for key: %v,incr: %v error: %v", cacheKey, totalIncr, err)
	}

	err = retry.WithAttempts(ctx, "CacheIncrUserTodayVoteCountForActivity", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "INCRBY", baseCacheKey, baseIncr))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserTodayVoteCountForActivity for key: %v,incr: %v error: %v", baseCacheKey, totalIncr, err)
	}
}

// CacheDelUserTodayVoteCountForActivity: 增加缓存中用户在整个活动下的投票总次数
func (d *Dao) CacheDelUserTodayVoteCountForActivity(ctx context.Context, activityId, mid int64) {
	var err error
	cacheKey := redisUserTodayVoteCountCacheKeyForActivity(activityId, mid)
	err = retry.WithAttempts(ctx, "CacheDelUserTodayVoteCountForActivity", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "DEL", cacheKey))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheDelUserTodayVoteCountForActivity for activityId: %v,mid: %v error: %v", activityId, mid, err)
	}
}

// DoVote: 用户进行投票动作
func (d *Dao) DoVote(ctx context.Context, mid int64, risk *riskModel.Base, req *model.DoVoteParams) (resp *api.VoteUserDoResp, err error) {
	var (
		DSG              *api.VoteDataSourceGroupItem
		activity         *api.VoteActivity
		DSGItem          DataSourceItem
		haveRiskBool     bool
		haveRisk         int64 //0: 未被风控, 1: 被风控
		tx               *sql.Tx
		baseAlreadyUsed  int64
		gotBaseCount     int64
		gotDECount       int64
		gotNECount       int64
		currentItemTotal int64
	)
	resp = &api.VoteUserDoResp{}
	//1.数据组/活动检查
	{
		DSG, activity, err = d.voteCheckDSGAndActivity(ctx, req.DataSourceGroupId, req.ActivityId)
		if err != nil {
			return
		}
	}
	//2.稿件检查
	{
		DSGItem, err = d.voteCheckDSItem(ctx, DSG, req.DataSourceItemId)
		if err != nil {
			return
		}
	}

	//3.投票次数校验(缓存过滤)
	{
		err = d.voteCheckCacheUserVoteCount(ctx, activity, mid, req)
		if err != nil {
			return
		}
	}

	//4.风控检查
	{
		if risk != nil {
			riskPrams := &riskModel.VoteNew{
				Base:        *risk,
				ActivityUID: riskModel.ActionVoteNew,
				TargetId:    req.DataSourceItemId,
				TargetType:  d.getRiskType(DSG.SourceType),
				TargetName:  DSGItem.GetName(),
				Id:          req.ActivityId,
				Score:       req.Vote,
			}
			haveRiskBool, err = d.RuleCheck(ctx, activity.Rule.RiskControlRule, riskPrams)
			if err != nil {
				log.Errorc(ctx, "DoVote RuleCheck risk: %+v, req: %+v error: %v", risk, req, err)
			}
			if haveRiskBool {
				haveRisk = 1
				resp.HadRisk = true
			}
		}
	}

	tx, err = d.db.Begin(ctx)
	if err != nil {
		log.Errorc(ctx, "DoVote db.Begin error: %v", err)
		err = ecode.ActivityVoteError
		return
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(ctx, "rollback error, err: %v, err1: %v", err, err1)
			} else {
				log.Errorc(ctx, "do vote error: %+v rollback success", err)
			}
		}
	}()
	//5.进行投票动作
	{
		//1.再次检查投票次数资格
		{
			baseAlreadyUsed, err = d.voteCheckDBUserVoteCount(ctx, activity, tx, mid, req.Vote)
			if err != nil {
				//可能出现DB和Cache不一致, 主动删除cache
				d.CacheDelUserVoteCache(ctx, activity.Id, mid, req.DataSourceGroupId, req.DataSourceItemId)
				return
			}
		}

		//2.单选项投票上限检查
		{
			err = d.voteCheckDBSingleOption(ctx, activity, tx, mid, req)
			if err != nil {
				d.CacheDelUserVoteCache(ctx, activity.Id, mid, req.DataSourceGroupId, req.DataSourceItemId)
				return
			}
		}

		//3.进行投票
		{
			gotBaseCount, err = d.tryDecrBaseVote(activity.Rule, baseAlreadyUsed, req.Vote)
			if err != nil {
				return
			}
			gotDECount, err = d.tryDecrExtraDEVote(ctx, tx, activity.Id, mid, req.Vote-gotBaseCount)
			if err != nil {
				return
			}
			gotNECount, err = d.tryDecrExtraNEVote(ctx, tx, activity.Id, mid, req.Vote-gotBaseCount-gotDECount)
			if err != nil {
				return
			}

			_, err = tx.Exec(fmt.Sprintf(sql4UpdateUserVoteCountInActivity, tableIdx(mid)), req.Vote, activity.Id, mid)
			if err != nil {
				log.Errorc(ctx, "DoVote tx.Exec sql4UpdateUserVoteCountInActivity error: %v", err)
				err = ecode.ActivityVoteError
				return
			}
			if gotBaseCount != 0 {
				_, err = tx.Exec(fmt.Sprintf(sql4InsertUserVoteRecord, tableIdx(mid)), activity.Id, req.DataSourceGroupId, req.DataSourceItemId, mid, haveRisk, gotBaseCount, timesTypeBase)
				if err != nil {
					log.Errorc(ctx, "DoVote tx.Exec sql4InsertUserVoteRecord error: %v", err)
					err = ecode.ActivityVoteError
					return
				}
			}
			if gotDECount != 0 {
				_, err = tx.Exec(fmt.Sprintf(sql4InsertUserVoteRecord, tableIdx(mid)), activity.Id, req.DataSourceGroupId, req.DataSourceItemId, mid, haveRisk, gotDECount, timesTypeExtraDailyExpire)
				if err != nil {
					log.Errorc(ctx, "DoVote tx.Exec sql4InsertUserVoteRecord error: %v", err)
					err = ecode.ActivityVoteError
					return
				}
			}
			if gotNECount != 0 {
				_, err = tx.Exec(fmt.Sprintf(sql4InsertUserVoteRecord, tableIdx(mid)), activity.Id, req.DataSourceGroupId, req.DataSourceItemId, mid, haveRisk, gotNECount, timesTypeExtraNotExpire)
				if err != nil {
					log.Errorc(ctx, "DoVote tx.Exec sql4InsertUserVoteRecord error: %v", err)
					err = ecode.ActivityVoteError
					return
				}
			}

			//TODO: 稿件等冲突较大的行锁使用异步监听binlog, 尝试合并写入
			if haveRiskBool {
				_, err = tx.Exec(sql4UpdateDataSourceVoteItemRiskVote, activity.Id, req.DataSourceGroupId, req.DataSourceItemId, req.Vote, req.Vote)
			} else {
				_, err = tx.Exec(sql4UpdateDataSourceVoteItemVote, activity.Id, req.DataSourceGroupId, req.DataSourceItemId, req.Vote, req.Vote)
			}
			if err != nil {
				log.Errorc(ctx, "DoVote tx.Exec sql4UpdateDataSourceVoteItemVote error: %v", err)
				err = ecode.ActivityVoteError
				return
			}
			currentItemTotal, err = d.incrUserItemContribRank(tx, activity.Id, req.DataSourceGroupId, req.DataSourceItemId, mid, req.Vote, haveRiskBool, activity.Rule.DisplayRiskVote)
			if err != nil {
				err = errors.Wrap(err, "d.incrUserItemContribRank")
				log.Errorc(ctx, "DoVote tx.Exec sql4UpdateDataSourceVoteItemVote error: %+v", err)
				err = ecode.ActivityVoteError
				return
			}
			if activity.Rule.VoteUpdateRule == int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime) {
				sqlStr := sql4UpdateDataSourceTotalVoteItemVoteWithoutRisk
				if activity.Rule.DisplayRiskVote {
					sqlStr = sql4UpdateDataSourceTotalVoteItemVoteWithRisk
				}
				_, err = tx.Exec(sqlStr, req.DataSourceGroupId, req.DataSourceItemId)
				if err != nil {
					log.Errorc(ctx, "DoVote tx.Exec sql4UpdateDataSourceTotalVoteItemVote error: %v", err)
					err = ecode.ActivityVoteError
					return
				}
			}
		}

	}
	err = tx.Commit()
	if err != nil {
		log.Errorc(ctx, "tx.Commit error: %v", err)
		return
	}
	//8.刷新缓存
	{
		//只有在实时刷新票数时才更新投票Zset
		if activity.Rule.VoteUpdateRule == int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime) {
			if !(haveRiskBool && !activity.Rule.DisplayRiskVote) { //风控判断
				d.IncrDSGItemVoteCountCache(ctx, req.DataSourceGroupId, req.DataSourceItemId, req.Vote)
				_ = d.AddUserToItemContribRankCache(ctx, req.DataSourceGroupId, req.DataSourceItemId, mid, currentItemTotal, time.Now().Unix())
			}
		}
		d.CacheIncrUserVoteCountForActivity(ctx, activity.Id, mid, req.Vote)
		d.CacheIncrUserTodayVoteCountForActivity(ctx, activity.Id, mid, gotBaseCount, req.Vote)
		d.CacheIncrUserVoteCountForDSGItem(ctx, req.DataSourceGroupId, req.DataSourceItemId, mid, req.Vote)
		d.CacheIncrUserExtraVoteTimesDE(ctx, activity.Id, mid, -gotDECount)
		d.CacheIncrUserExtraVoteTimesNE(ctx, activity.Id, mid, -gotNECount)
	}
	resp.UserAvailVoteCount, resp.UserAvailTmpVoteCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, activity.Id, mid)
	if err == nil {
		resp.UserCanVoteCountForItem, err = d.GetUserAvailVoteCountForDSGItem(ctx, activity.Rule, req.DataSourceGroupId, req.DataSourceItemId, mid)
	}

	return
}

// rawBatchGetUserVoteCountForDSGItem: 获取用户对某数据组下所有稿件的投票次数
func (d *Dao) rawBatchGetUserVoteCountForDSGItem(ctx context.Context, mid, sourceGroupId int64) (res map[int64] /*itemId*/ int64 /*投票次数*/, err error) {
	res = make(map[int64]int64)
	rows, err := d.db.Query(ctx, fmt.Sprintf(sql4BatchGetUserVoteCountForDSGItems, tableIdx(mid)), sourceGroupId, mid)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err == nil {
			err = rows.Err()
		}
	}()
	for rows.Next() {
		var (
			dataSourceItemId int64
			voteCount        int64
		)
		err = rows.Scan(&dataSourceItemId, &voteCount)
		if err != nil {
			return
		}
		res[dataSourceItemId] = voteCount
	}
	return
}

// rawBatchGetUserTodayVoteCountForDSGItem: 获取用户今天对某数据组下所有稿件的投票次数
func (d *Dao) rawBatchGetUserTodayVoteCountForDSGItem(ctx context.Context, mid, sourceGroupId int64) (res map[int64] /*itemId*/ int64 /*投票次数*/, err error) {
	res = make(map[int64]int64, 0)
	rows, err := d.db.Query(ctx, fmt.Sprintf(sql4BatchGetUserTodayVoteCountForDSGItems, tableIdx(mid)), sourceGroupId, mid)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err == nil {
			err = rows.Err()
		}
	}()
	for rows.Next() {
		var (
			dataSourceItemId int64
			voteCount        int64
		)
		err = rows.Scan(&dataSourceItemId, &voteCount)
		if err != nil {
			return
		}
		res[dataSourceItemId] = voteCount
	}
	return
}

// cacheGetUserVoteCountForDSGItem: 从缓存中获取用户对数据组下稿件的投票次数, 如果缓存没有则进行回源
func (d *Dao) cacheGetUserVoteCountForDSGItem(ctx context.Context, mid, sourceGroupId, sourceItemId int64) (count int64, miss bool) {
	if sourceGroupId == 0 || sourceItemId == 0 {
		return 0, false
	}
	miss = true
	count, err := redis.Int64(d.redis.Do(ctx, "GET", redisUserVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid)))
	if err == nil {
		miss = false
	}
	return
}

// cacheGetUserTodayVoteCountForDSGItem: 从缓存中获取用户对数据组下稿件的投票次数
func (d *Dao) cacheGetUserTodayVoteCountForDSGItem(ctx context.Context, mid, sourceGroupId, sourceItemId int64) (count int64, miss bool) {
	if sourceGroupId == 0 || sourceItemId == 0 {
		return 0, false
	}
	miss = true
	count, err := redis.Int64(d.redis.Do(ctx, "GET", redisUserTodayVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid)))
	if err == nil {
		miss = false
	}
	return
}

func (d *Dao) cacheSetUserVoteCountForDataSourceItem(ctx context.Context, mid, sourceGroupId, sourceItemId, votes int64) (err error) {
	err = retry.WithAttempts(ctx, "cacheSetUserVoteCountForDataSourceItem", 2, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "SETEX", redisUserVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid), d.userVoteCountExpire, votes)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "cacheSetUserVoteCountForDataSourceItem mid: %v, group: %v, item: %v, votes: %v error: %v", mid, sourceGroupId, sourceItemId, votes, err)
	}
	return
}

func (d *Dao) cacheSetUserTodayVoteCountForDataSourceItem(ctx context.Context, mid, sourceGroupId, sourceItemId, votes int64) (err error) {
	err = retry.WithAttempts(ctx, "cacheSetUserTodayVoteCountForDataSourceItem", 2, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "SETEX", redisUserTodayVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid), d.userVoteCountExpire, votes)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "cacheSetUserTodayVoteCountForDataSourceItem mid: %v, group: %v, item: %v, votes: %v error: %v", mid, sourceGroupId, sourceItemId, votes, err)
	}
	return
}

func (d *Dao) CacheIncrUserVoteCountForDSGItem(ctx context.Context, sourceGroupId, sourceItemId, mid, incr int64) {
	var err error
	err = retry.WithAttempts(ctx, "CacheIncrUserVoteCountForDSGItem", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "INCRBY", redisUserVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid), incr)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserVoteCountForDSGItem for sourceGroupId: %v,sourceItemId: %v, mid: %v, error: %v", sourceGroupId, sourceItemId, mid, err)
	}

	err = retry.WithAttempts(ctx, "CacheIncrUserTodayVoteCountForDSGItem", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "INCRBY", redisUserTodayVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid), incr)
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserTodayVoteCountForDSGItem for sourceGroupId: %v,sourceItemId: %v, mid: %v, error: %v", sourceGroupId, sourceItemId, mid, err)
	}
}

func (d *Dao) CacheDelUserVoteCountForDSGItem(ctx context.Context, sourceGroupId, sourceItemId, mid int64) {
	var err error
	err = retry.WithAttempts(ctx, "CacheDelUserVoteCountForDSGItem", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "DEL", redisUserVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheDelUserVoteCountForDSGItem for sourceGroupId: %v,sourceItemId: %v, mid: %v, error: %v", sourceGroupId, sourceItemId, mid, err)
	}

	err = retry.WithAttempts(ctx, "CacheDelUserTodayVoteCountForDSGItem", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = d.redis.Do(ctx, "DEL", redisUserTodayVoteCountCacheKeyForDSGItem(sourceGroupId, sourceItemId, mid))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheDelUserTodayVoteCountForDSGItem for sourceGroupId: %v,sourceItemId: %v, mid: %v, error: %v", sourceGroupId, sourceItemId, mid, err)
	}
}

func (d *Dao) CacheDelUserVoteCache(ctx context.Context, activityId, mid, sourceGroupId, sourceItemId int64) {
	d.CacheDelUserVoteCountForActivity(ctx, activityId, mid)
	d.CacheDelUserTodayVoteCountForActivity(ctx, activityId, mid)
	d.CacheDelUserVoteCountForDSGItem(ctx, sourceGroupId, sourceItemId, mid)
	d.CacheDelUserExtraVoteTimesDE(ctx, activityId, mid)
	d.CacheDelUserExtraVoteTimesNE(ctx, activityId, mid)
}

func (d *Dao) GetUserAvailVoteCountForDSGItem(ctx context.Context, voteRule *api.VoteActivityRule, sourceGroupId, sourceItemId, mid int64) (availCount int64, err error) {
	fake := []*model.RankInfo{{
		DataSourceGroupId: sourceGroupId,
		DataSourceItemId:  sourceItemId,
	}}
	err = d.batchGetUserVoteCountForDSGItem(ctx, voteRule, sourceGroupId, mid, fake)
	availCount = fake[0].UserCanVoteCount
	return
}

func (d *Dao) batchGetUserVoteCountForDSGItem(ctx context.Context, voteRule *api.VoteActivityRule, sourceGroupId, mid int64, res []*model.RankInfo) (err error) {
	var missCount int64
	for _, r := range res {
		count, miss := d.cacheGetUserVoteCountForDSGItem(ctx, mid, r.DataSourceGroupId, r.DataSourceItemId)
		if miss {
			missCount++
			break
		} else {
			r.UserVoteCount = count
		}
		todayCount, miss := d.cacheGetUserTodayVoteCountForDSGItem(ctx, mid, r.DataSourceGroupId, r.DataSourceItemId)
		if miss {
			missCount++
			break
		} else {
			r.UserVoteCountToday = todayCount
		}
		switch voteRule.SingleOptionBehavior {
		case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorUnlimited):
			r.UserCanVoteCount = 99
		case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorDayOnce):
			r.UserCanVoteCount = int64(1) - r.UserVoteCountToday
		case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorTotalOnce):
			r.UserCanVoteCount = int64(1) - r.UserVoteCount
		default:
		}
		if r.UserCanVoteCount < 0 {
			r.UserCanVoteCount = 0
		}
	}

	if missCount > 0 {
		var rawCount map[int64]int64
		var rawTodayVoteCount map[int64]int64
		rawCount, err = d.rawBatchGetUserVoteCountForDSGItem(ctx, mid, sourceGroupId)
		if err != nil {
			log.Errorc(ctx, "batchGetUserVoteCountForDSGItem rawBatchGetUserVoteCountForDSGItem sourceGroupId: %v, mid: %v, error: %v", sourceGroupId, mid, err)
			return
		}
		rawTodayVoteCount, err = d.rawBatchGetUserTodayVoteCountForDSGItem(ctx, mid, sourceGroupId)
		if err != nil {
			log.Errorc(ctx, "batchGetUserVoteCountForDSGItem rawBatchGetUserTodayVoteCountForDSGItem sourceGroupId: %v, mid: %v, error: %v", sourceGroupId, mid, err)
			return
		}
		for _, r := range res {
			r.UserVoteCount = rawCount[r.DataSourceItemId]
			r.UserVoteCountToday = rawTodayVoteCount[r.DataSourceItemId]
			switch voteRule.SingleOptionBehavior {
			case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorUnlimited):
				r.UserCanVoteCount = 99
			case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorDayOnce):
				r.UserCanVoteCount = int64(1) - r.UserVoteCountToday
			case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorTotalOnce):
				r.UserCanVoteCount = int64(1) - r.UserVoteCount
			default:
			}
			if r.UserCanVoteCount < 0 {
				r.UserCanVoteCount = 0
			}
			_ = d.cacheSetUserVoteCountForDataSourceItem(ctx, mid, r.DataSourceGroupId, r.DataSourceItemId, r.UserVoteCount)
			_ = d.cacheSetUserTodayVoteCountForDataSourceItem(ctx, mid, r.DataSourceGroupId, r.DataSourceItemId, r.UserVoteCountToday)
		}
	}
	return
}

// UndoVote: 用户进行取消投票动作
func (d *Dao) UndoVote(ctx context.Context, mid int64, req *model.UndoVoteParams) (resp *api.VoteUserUndoResp, err error) {
	var (
		activity            *api.VoteActivity
		undoRecordCount     int64
		totalUndoCount      int64
		tx                  *sql.Tx
		res                 xsql.Result
		rows                *sql.Rows
		baseTimesCount      int64
		extraDETimesCount   int64
		extraNETimesCount   int64
		currentContribCount int64
		riskVoteCount       int64
		nonRiskVoteCount    int64
	)
	resp = &api.VoteUserUndoResp{}
	//1.数据组校验
	_, activity, err = d.voteCheckDSGAndActivity(ctx, req.DataSourceGroupId, req.ActivityId)
	if err != nil {
		return
	}
	//1.1 重建次数缓存
	_, _, err = d.GetUserAvailVoteCount(ctx, activity.Rule, req.ActivityId, mid)
	if err != nil {
		return
	}
	//2.缓存次数校验
	//检查当日投票次数
	//检查稿件是否已投票, 不是的话报错退出
	{
		todayCount, miss := d.cacheGetUserTodayVoteCountForDSGItem(ctx, mid, req.DataSourceGroupId, req.DataSourceItemId)
		if !miss && todayCount == 0 {
			err = ecode.ActivityVoteNoHistory
			return
		}
	}
	//3.DB中进行撤销动作
	{
		tx, err = d.db.Begin(ctx)
		if err != nil {
			return
		}
		defer func() {
			log.Infoc(ctx, "UndoVote success, baseTimesCount: %v, extraDETimesCount: %v,"+
				"extraNETimesCount: %v, currentContribCount: %v", baseTimesCount, extraDETimesCount,
				extraNETimesCount, currentContribCount)
			if err != nil {
				log.Infoc(ctx, "UndoVote fail: %v, baseTimesCount: %v, extraDETimesCount: %v,"+
					"extraNETimesCount: %v, currentContribCount: %v", err, baseTimesCount, extraDETimesCount,
					extraNETimesCount, currentContribCount)
				err1 := tx.Rollback()
				if err1 != nil {
					log.Errorc(ctx, "UndoVote tx.Rollback error: %v", err)
				}
			}
		}()

		//3.1 用户+活动级别上锁
		{
			currCount := int64(0)
			err = tx.QueryRow(fmt.Sprintf(sql4LockUserVoteSummary, tableIdx(mid)), activity.Id, mid).Scan(&currCount)
			if err == sql.ErrNoRows || currCount == 0 {
				err = ecode.ActivityVoteNoHistory
				d.CacheDelUserVoteCache(ctx, activity.Id, mid, req.DataSourceGroupId, req.DataSourceItemId)
				return
			}
			if err != nil {
				log.Errorc(ctx, "UndoVote try to lock error: %v", err)
				err = ecode.ActivityVoteError
				return
			}
		}

		//获取此用户今天的风控票数和非风控票数
		{
			rows, err = tx.Query(fmt.Sprintf(sql4GetUserTodayVoteCountForDSGForUndo, tableIdx(mid)), req.DataSourceGroupId, req.DataSourceItemId, mid)
			if err != nil {
				log.Errorc(ctx, "UndoVote tx.Query error: %v", err)
				return
			}
			defer func() {
				_ = rows.Close()
				if err == nil {
					err = rows.Err()
				}
			}()
			for rows.Next() {
				hadRisk := int64(0)
				tmpCount := int64(0)
				err = rows.Scan(&hadRisk, &tmpCount)
				if err != nil {
					log.Errorc(ctx, "UndoVote rows.Scan error: %v", err)
					return
				}
				switch hadRisk {
				case 0:
					nonRiskVoteCount = tmpCount
				case 1:
					riskVoteCount = tmpCount
				default:
					log.Errorc(ctx, "UndoVote got unexpected hadRisk: %v", hadRisk)
					return
				}
			}
			if err != nil {
				return
			}
		}

		//获取各类型次数的消耗情况
		{
			rows, err = tx.Query(fmt.Sprintf(sql4GetUserTodayTimesCountForDSGForUndo, tableIdx(mid)), req.DataSourceGroupId, req.DataSourceItemId, mid)
			if err != nil {
				log.Errorc(ctx, "UndoVote tx.Query error: %v", err)
				return
			}
			defer func() {
				_ = rows.Close()
				if err == nil {
					err = rows.Err()
				}
			}()
			for rows.Next() {
				timesType := int64(0)
				tmpCount := int64(0)
				err = rows.Scan(&timesType, &tmpCount)
				if err != nil {
					log.Errorc(ctx, "UndoVote rows.Scan error: %v", err)
					return
				}
				totalUndoCount = totalUndoCount + tmpCount
				switch timesType {
				case timesTypeBase:
					baseTimesCount = tmpCount
				case timesTypeExtraNotExpire:
					extraNETimesCount = tmpCount
				case timesTypeExtraDailyExpire:
					extraDETimesCount = tmpCount
				default:
					log.Errorc(ctx, "UndoVote got unexpected timesType: %v", timesType)
					return
				}
			}
			if err != nil {
				return
			}
		}

		res, err = tx.Exec(fmt.Sprintf(sql4UndoUserTodayVoteForDSG, tableIdx(mid)), req.DataSourceGroupId, req.DataSourceItemId, mid)
		if err != nil {
			log.Errorc(ctx, "UndoVote tx.Exec error: %v", err)
			return
		}
		undoRecordCount, err = res.RowsAffected()
		if err != nil {
			log.Errorc(ctx, "UndoVote res.RowsAffected error: %v", err)
			return
		}
		if undoRecordCount == 0 {
			err = ecode.ActivityVoteNoHistory
			d.CacheDelUserVoteCache(ctx, activity.Id, mid, req.DataSourceGroupId, req.DataSourceItemId)
			return
		}
		_, err = tx.Exec(fmt.Sprintf(sql4UpdateUserVoteCountInActivity, tableIdx(mid)), -totalUndoCount, req.ActivityId, mid)
		if err != nil {
			return
		}

		//回滚票数(非风控)
		if nonRiskVoteCount > 0 {
			_, err = tx.Exec(sql4UpdateDataSourceVoteItemVote, req.ActivityId, req.DataSourceGroupId, req.DataSourceItemId, -nonRiskVoteCount, -nonRiskVoteCount)
			if err != nil {
				return
			}
		}

		//回滚票数(风控)
		if riskVoteCount > 0 {
			_, err = tx.Exec(sql4UpdateDataSourceVoteItemRiskVote, req.ActivityId, req.DataSourceGroupId, req.DataSourceItemId, -riskVoteCount, -riskVoteCount)
			if err != nil {
				return
			}
		}

		//更新总票数
		if activity.Rule.VoteUpdateRule == int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime) {
			sqlStr := sql4UpdateDataSourceTotalVoteItemVoteWithoutRisk
			if activity.Rule.DisplayRiskVote {
				sqlStr = sql4UpdateDataSourceTotalVoteItemVoteWithRisk
			}
			_, err = tx.Exec(sqlStr, req.DataSourceGroupId, req.DataSourceItemId)
			if err != nil {
				log.Errorc(ctx, "DoVote tx.Exec sql4UpdateDataSourceTotalVoteItemVote error: %v", err)
				err = ecode.ActivityVoteError
				return
			}
		}

		if extraDETimesCount != 0 {
			err = d.incrUserExtraDECount(ctx, tx, req.ActivityId, mid, extraDETimesCount)
			if err != nil {
				return
			}
		}

		if extraNETimesCount != 0 {
			err = d.incrUserExtraNECount(ctx, tx, req.ActivityId, mid, extraNETimesCount)
			if err != nil {
				return
			}
		}
		//TODO: total_votes根据投票规则进行修改(binlog监听)
		currentContribCount, err = d.decrUserItemContribRank(tx, activity.Id, req.DataSourceGroupId, req.DataSourceItemId, mid, nonRiskVoteCount, riskVoteCount, activity.Rule.DisplayRiskVote)
		if err != nil {
			return
		}
		err = tx.Commit()
		if err != nil {
			return
		}
	}

	//3.更新缓存
	{
		//只有在实时刷新票数时才更新投票Zset
		if activity.Rule.VoteUpdateRule == int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime) {
			d.IncrDSGItemVoteCountCache(ctx, req.DataSourceGroupId, req.DataSourceItemId, -nonRiskVoteCount)
			if activity.Rule.DisplayRiskVote {
				d.IncrDSGItemVoteCountCache(ctx, req.DataSourceGroupId, req.DataSourceItemId, -riskVoteCount)
			}
		}
		d.CacheIncrUserVoteCountForActivity(ctx, activity.Id, mid, -totalUndoCount)
		d.CacheIncrUserTodayVoteCountForActivity(ctx, activity.Id, mid, -baseTimesCount, -totalUndoCount)
		d.CacheIncrUserVoteCountForDSGItem(ctx, req.DataSourceGroupId, req.DataSourceItemId, mid, -totalUndoCount)
		d.CacheIncrUserExtraVoteTimesDE(ctx, activity.Id, mid, extraDETimesCount)
		d.CacheIncrUserExtraVoteTimesNE(ctx, activity.Id, mid, extraNETimesCount)
		_ = d.AddUserToItemContribRankCache(ctx, req.DataSourceGroupId, req.DataSourceItemId, mid, currentContribCount, time.Now().Unix())
	}
	resp.UserAvailVoteCount, resp.UserAvailTmpVoteCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, activity.Id, mid)
	if err == nil {
		resp.UserCanVoteCountForItem, err = d.GetUserAvailVoteCountForDSGItem(ctx, activity.Rule, req.DataSourceGroupId, req.DataSourceItemId, mid)
	}
	return
}

func (d *Dao) existCheckDSGAndActivity(ctx context.Context, sourceGroupId, activityId int64) (DSG *api.VoteDataSourceGroupItem, activity *api.VoteActivity, err error) {
	//1.数据组校验
	{
		DSG, err = d.DataSourceGroup(ctx, sourceGroupId)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		if DSG == nil {
			err = ecode.ActivityVoteDSGNotFound
			return
		}
	}
	//2.活动校验
	{
		activity, err = d.Activity(ctx, activityId)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		if activity == nil {
			err = ecode.ActivityVoteNotFound
			return
		}
	}
	return
}

// voteCheckDSGAndActivity: 用户进行投票的检查, 必须是正在进行中的活动
func (d *Dao) voteCheckDSGAndActivity(ctx context.Context, sourceGroupId, activityId int64) (DSG *api.VoteDataSourceGroupItem, activity *api.VoteActivity, err error) {
	DSG, activity, err = d.existCheckDSGAndActivity(ctx, sourceGroupId, activityId)
	if err != nil {
		return
	}
	now := time.Now().Unix()
	if activity.StartTime >= now {
		err = ecode.ActivityVoteNotStarted
		return
	}
	if activity.EndTime <= now {
		err = ecode.ActivityVoteFinished
		return
	}
	if activity.Rule.TotalLimit == 0 || activity.Rule.SingleDayLimit == 0 {
		err = ecode.ActivityVoteOverLimit
		return
	}
	return
}

// voteListCheckDSGAndActivity: 用户请求投票排行时的检查, 结束后90天内的排行仍允许查看
func (d *Dao) voteListCheckDSGAndActivity(ctx context.Context, sourceGroupId, activityId int64) (DSG *api.VoteDataSourceGroupItem, activity *api.VoteActivity, err error) {
	DSG, activity, err = d.existCheckDSGAndActivity(ctx, sourceGroupId, activityId)
	if err != nil {
		return
	}
	now := time.Now().Unix()
	if activity.StartTime >= now {
		err = ecode.ActivityVoteNotStarted
		return
	}
	if activity.EndTime <= time.Now().AddDate(0, 0, -90).Unix() {
		err = ecode.ActivityVoteFinished
		return
	}
	return
}

func (d *Dao) voteCheckDSItem(ctx context.Context, DSG *api.VoteDataSourceGroupItem, sourceItemId int64) (DSGItem DataSourceItem, err error) {
	//3.检查数据源下的指定稿件是否存在
	{
		DSGItem, err = d.GetDSGItem(ctx, DSG, sourceItemId)
		if err != nil {
			return
		}
	}
	//5.黑名单校验
	{
		var inBlack bool
		inBlack, err = d.BlackListCheck(ctx, DSG.GroupId, sourceItemId)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		if inBlack {
			err = ecode.ActivityVoteItemNotFound
			return
		}
	}
	return
}

func (d *Dao) voteCheckCacheUserVoteCount(ctx context.Context, activity *api.VoteActivity, mid int64, req *model.DoVoteParams) (err error) {
	//总数
	if activity.Rule != nil {
		var currTotalCount int64
		currTotalCount, err = d.CacheGetUserVoteCountForActivity(ctx, activity.Id, mid)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		if currTotalCount+req.Vote > activity.Rule.TotalLimit {
			err = ecode.ActivityVoteExceed
			return
		}
	}
	//当日
	if activity.Rule != nil {
		var currTodayCount int64
		var baseTodayCount int64
		var extraDETimes, extraNETimes int64
		baseTodayCount, currTodayCount, err = d.CacheGetUserTodayVoteCountForActivity(ctx, activity.Id, mid)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		if currTodayCount+req.Vote > activity.Rule.SingleDayLimit { //当日投票上限
			err = ecode.ActivityVoteExceed
			return
		}

		extraDETimes, err = d.CacheGetUserExtraVoteTimesDE(ctx, activity.Id, mid)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		extraNETimes, err = d.CacheGetUserExtraVoteTimesNE(ctx, activity.Id, mid)
		if err != nil {
			err = ecode.ActivityVoteError
			return
		}
		if req.Vote > activity.Rule.BaseTimes-baseTodayCount+extraDETimes+extraNETimes { //检验可用次数是否足够
			err = ecode.ActivityVoteOverLimit
			return
		}

		currItemTodayCount, miss := d.cacheGetUserTodayVoteCountForDSGItem(ctx, mid, req.DataSourceGroupId, req.DataSourceItemId)
		if !miss && currItemTodayCount > 0 && activity.Rule.SingleOptionBehavior != int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorUnlimited) {
			err = ecode.ActivityVoteItemVoted
			return
		}
	}
	return
}

func (d *Dao) voteCheckDBUserVoteCount(ctx context.Context, activity *api.VoteActivity, tx *sql.Tx, mid, voteCount int64) (baseUsed int64, err error) {
	//用户投票总数, 开启用户级别的锁定.
	{
		currCount := int64(0)
		err = tx.QueryRow(fmt.Sprintf(sql4LockUserVoteSummary, tableIdx(mid)), activity.Id, mid).Scan(&currCount)
		if err == sql.ErrNoRows {
			err = nil
			//记录不存在, 进行插入
			_, err = tx.Exec(fmt.Sprintf(sql4InsertUserVoteSummary, tableIdx(mid)), activity.Id, mid)
			if err != nil {
				log.Errorc(ctx, "DoVote insert vote summary error: %v", err)
				err = ecode.ActivityVoteError
				return
			}
			//重新获取锁
			err = tx.QueryRow(fmt.Sprintf(sql4LockUserVoteSummary, tableIdx(mid)), activity.Id, mid).Scan(&currCount)
		}
		if err != nil {
			log.Errorc(ctx, "DoVote try to lock error: %v", err)
			err = ecode.ActivityVoteError
			return
		}
		if activity.Rule != nil {
			if currCount+voteCount > activity.Rule.TotalLimit {
				err = ecode.ActivityVoteOverLimit
				return
			}
		}
	}

	//投票次数校验
	{
		if activity.Rule != nil {
			var total int64
			//配置了单选项限制, incr不能大于1.
			if voteCount > activity.Rule.SingleDayLimit {
				err = ecode.ActivityVoteExceed
				return
			}
			baseUsed, _, _, total, err = d.RawUserTodayVoteCountForActivity(ctx, tx, activity.Id, mid)
			if total+voteCount > activity.Rule.SingleDayLimit {
				err = ecode.ActivityVoteExceed
				return
			}
			var extraTimesDE int64
			err = tx.QueryRow(fmt.Sprintf(sql4GetUserDETotalExtraTimes, tableIdx(mid)), activity.Id, mid).Scan(&extraTimesDE)
			if err == sql.ErrNoRows {
				err = nil
			}
			if err != nil {
				log.Errorc(ctx, "DoVote get extraTimesDE error: %v", err)
				err = ecode.ActivityVoteError
				return
			}

			var extraTimesNE int64
			err = tx.QueryRow(fmt.Sprintf(sql4GetUserNETotalExtraTimes, tableIdx(mid)), activity.Id, mid).Scan(&extraTimesNE)
			if err == sql.ErrNoRows {
				err = nil
			}
			if err != nil {
				log.Errorc(ctx, "DoVote get extraTimesNE error: %v", err)
				err = ecode.ActivityVoteError
				return
			}
			if voteCount > activity.Rule.BaseTimes-baseUsed+extraTimesDE+extraTimesNE {
				err = ecode.ActivityVoteOverLimit
				return
			}

		}
	}

	return
}

func (d *Dao) voteCheckDBSingleOption(ctx context.Context, activity *api.VoteActivity, tx *sql.Tx, mid int64, req *model.DoVoteParams) (err error) {
	if activity.Rule != nil && activity.Rule.SingleOptionBehavior != int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorUnlimited) {
		//配置了单选项限制, incr不能大于1.
		if req.Vote > 1 {
			err = ecode.ActivityVoteExceed
			return
		}
		sqlStr := ""
		itemVoteCount := 0
		switch activity.Rule.SingleOptionBehavior {
		case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorDayOnce):
			sqlStr = sql4GetUserTodayVoteCountForItem
		case int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorTotalOnce):
			sqlStr = sql4GetUserVoteCountForItem
		default:
			sqlStr = sql4GetUserVoteCountForItem
		}
		err = tx.QueryRow(fmt.Sprintf(sqlStr, tableIdx(mid)), mid, req.DataSourceGroupId, req.DataSourceItemId).Scan(&itemVoteCount)
		if err == sql.ErrNoRows {
			err = nil
		}
		if err != nil {
			log.Errorc(ctx, "DoVote get item vote count error: %v", err)
			err = ecode.ActivityVoteError
			return
		}
		if itemVoteCount >= 1 {
			err = ecode.ActivityVoteItemVoted
			return
		}
	}
	return
}

package vote

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/api"
	"time"
)

const (
	timesTypeExtraNotExpire   = 0
	timesTypeExtraDailyExpire = 1
	timesTypeBase             = 2
)

const (
	sql4GetUserDETotalExtraTimes = `
SELECT times
FROM   act_vote_user_times_%v
WHERE  main_id = ?
AND    mid = ?
AND	   times_type = 1
AND    times_date = CURRENT_DATE()
`

	sql4DecrUserDETotalExtraTimes = `
UPDATE act_vote_user_times_%v
SET times = times - ?
WHERE  main_id = ?
AND    mid = ?
AND	   times_type = 1
AND    times_date = CURRENT_DATE()
`

	sql4IncrUserDETotalExtraTimes = `
INSERT INTO  act_vote_user_times_%v (main_id, mid, times_type,times_date,times) VALUES(?,?,?,current_date(),?) on duplicate key update times=times+?
`

	sql4GetUserNETotalExtraTimes = `
SELECT times
FROM   act_vote_user_times_%v
WHERE  main_id = ?
AND    mid = ?
AND	   times_type = 0
AND    times_date = '9999-12-31'
`

	sql4DecrUserNETotalExtraTimes = `
UPDATE act_vote_user_times_%v
SET times = times - ?
WHERE  main_id = ?
AND    mid = ?
AND	   times_type = 0
AND    times_date = '9999-12-31'
`

	sql4IncrUserNETotalExtraTimes = `
INSERT INTO  act_vote_user_times_%v (main_id, mid, times_type,times_date,times) VALUES(?,?,?,'9999-12-31',?) on duplicate key update times=times+?

`

	sql4GetUserTodayVoteCountForActivityGroupByType = `
SELECT times_type,COALESCE(sum(votes),0)
FROM act_vote_user_action_%v FORCE INDEX (ix_sgid_mid)
WHERE mid = ?
  AND main_id = ?
  AND is_undo = 0
  AND DATE(ctime) = CURRENT_DATE()
GROUP BY times_type
`
)

/**********************************缓存Key控制**********************************/
//redisUserExtraVoteTimesCacheKeyDE: 用户每日过期的票数
func redisUserExtraVoteTimesCacheKeyDE(activityId, mid int64) string {
	return fmt.Sprintf("vote_act_times_de_%v_%v_%v", activityId, mid, time.Now().Format("20060102"))
}

// redisUserExtraVoteTimesCacheKeyNE: 用户不过期的票数
func redisUserExtraVoteTimesCacheKeyNE(activityId, mid int64) string {
	return fmt.Sprintf("vote_act_times_ne_%v_%v", activityId, mid)
}

// CacheDelUserExtraVoteTimesDE: 删除用户的当日过期票数缓存
func (d *Dao) CacheDelUserExtraVoteTimesDE(ctx context.Context, activityId, mid int64) {
	var err error
	cacheKey := redisUserExtraVoteTimesCacheKeyDE(activityId, mid)
	err = retry.WithAttempts(ctx, "CacheDelUserExtraVoteTimesDE", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "DEL", cacheKey))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheDelUserExtraVoteTimesDE for activityId: %v,mid: %v error: %v", activityId, mid, err)
	}
}

// CacheGetUserExtraVoteTimesDE: 获取用户的当日过期票数
func (d *Dao) CacheGetUserExtraVoteTimesDE(ctx context.Context, activityId, mid int64) (tmpTimes int64, err error) {
	cacheKey := redisUserExtraVoteTimesCacheKeyDE(activityId, mid)
	tmpTimes, err = redis.Int64(d.redis.Do(ctx, "GET", cacheKey))
	if err == nil {
		return
	}
	err = d.db.QueryRow(ctx, fmt.Sprintf(sql4GetUserDETotalExtraTimes, tableIdx(mid)), activityId, mid).Scan(&tmpTimes)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err == nil {
		_, _ = d.redis.Do(ctx, "SETEX", cacheKey, d.userVoteCountExpire, tmpTimes)
	}
	return
}

// CacheIncrUserExtraVoteTimesDE: 修改用户的当日过期票数缓存
func (d *Dao) CacheIncrUserExtraVoteTimesDE(ctx context.Context, activityId, mid int64, incr int64) {
	var err error
	cacheKey := redisUserExtraVoteTimesCacheKeyDE(activityId, mid)
	_, err = d.redis.Do(ctx, "INCRBY", cacheKey, incr)
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserExtraVoteTimesDE for activityId: %v,mid: %v error: %v", activityId, mid, err)
	}
}

// CacheDelUserExtraVoteTimesNE: 删除用户的不过期票数缓存
func (d *Dao) CacheDelUserExtraVoteTimesNE(ctx context.Context, activityId, mid int64) {
	var err error
	cacheKey := redisUserExtraVoteTimesCacheKeyNE(activityId, mid)
	err = retry.WithAttempts(ctx, "CacheDelUserExtraVoteTimesNE", 1, netutil.DefaultBackoffConfig, func(c context.Context) error {
		_, err = redis.Int64(d.redis.Do(ctx, "DEL", cacheKey))
		return err
	})
	if err != nil {
		log.Errorc(ctx, "CacheDelUserExtraVoteTimesNE for activityId: %v,mid: %v error: %v", activityId, mid, err)
	}
}

// CacheGetUserExtraVoteTimesNE: 获取用户的不过期票数
func (d *Dao) CacheGetUserExtraVoteTimesNE(ctx context.Context, activityId, mid int64) (tmpTimes int64, err error) {
	cacheKey := redisUserExtraVoteTimesCacheKeyNE(activityId, mid)
	tmpTimes, err = redis.Int64(d.redis.Do(ctx, "GET", cacheKey))
	if err == nil {
		return
	}
	err = d.db.QueryRow(ctx, fmt.Sprintf(sql4GetUserNETotalExtraTimes, tableIdx(mid)), activityId, mid).Scan(&tmpTimes)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err == nil {
		_, _ = d.redis.Do(ctx, "SETEX", cacheKey, d.userVoteCountExpire, tmpTimes)
	}
	return
}

// CacheIncrUserExtraVoteTimesNE: 修改用户的当日过期票数缓存
func (d *Dao) CacheIncrUserExtraVoteTimesNE(ctx context.Context, activityId, mid int64, incr int64) {
	var err error
	cacheKey := redisUserExtraVoteTimesCacheKeyNE(activityId, mid)
	_, err = d.redis.Do(ctx, "INCRBY", cacheKey, incr)
	if err != nil {
		log.Errorc(ctx, "CacheIncrUserExtraVoteTimesNE for activityId: %v,mid: %v error: %v", activityId, mid, err)
	}
}

func (d *Dao) AddUserTimes(ctx context.Context, input *api.VoteUserAddTimesReq) (res *api.NoReply, err error) {
	res = &api.NoReply{}
	switch input.VoteTimesExpireType {
	case api.VoteTimesExpireType_Daily:
		err = d.incrUserExtraDECount(ctx, nil, input.ActivityId, input.Mid, input.Times)
		if err == nil {
			d.CacheDelUserExtraVoteTimesDE(ctx, input.ActivityId, input.Mid)
		}
	case api.VoteTimesExpireType_NotExpire:
		err = d.incrUserExtraNECount(ctx, nil, input.ActivityId, input.Mid, input.Times)
		if err == nil {
			d.CacheDelUserExtraVoteTimesNE(ctx, input.ActivityId, input.Mid)
		}
	default:
		err = ecode.RequestErr
	}
	return
}

func (d *Dao) GetUserAvailVoteCount(ctx context.Context, rule *api.VoteActivityRule, activityId, mid int64) (availCount int64, extraAvailCount int64, err error) {
	if mid == 0 {
		return
	}
	var extraAvailCountDE int64
	var extraAvailCountNE int64
	var baseVoteCount int64
	baseVoteCount, _, err = d.CacheGetUserTodayVoteCountForActivity(ctx, activityId, mid)
	if err != nil {
		return
	}
	extraAvailCountDE, err = d.CacheGetUserExtraVoteTimesDE(ctx, activityId, mid)
	if err != nil {
		return
	}
	extraAvailCountNE, err = d.CacheGetUserExtraVoteTimesNE(ctx, activityId, mid)
	if err != nil {
		return
	}
	extraAvailCount = extraAvailCountDE + extraAvailCountNE
	availCount = rule.BaseTimes - baseVoteCount

	if availCount < 0 {
		availCount = 0
	}
	if extraAvailCount < 0 {
		extraAvailCount = 0
	}
	return
}

// tryDecrExtraDEVote: 尝试扣减当日可用次数
func (d *Dao) tryDecrBaseVote(rule *api.VoteActivityRule, baseAlreadyUsed, count int64) (got int64, err error) {
	if count <= 0 || baseAlreadyUsed >= rule.BaseTimes {
		return
	}
	baseLeft := rule.BaseTimes - baseAlreadyUsed
	if baseLeft <= count {
		got = baseLeft
	} else {
		got = count
	}
	return
}

// tryDecrExtraDEVote: 尝试扣减当日可用次数
func (d *Dao) tryDecrExtraDEVote(ctx context.Context, tx *sql.Tx, activityId, mid, count int64) (got int64, err error) {
	if count <= 0 {
		return
	}
	err = tx.QueryRow(fmt.Sprintf(sql4GetUserDETotalExtraTimes, tableIdx(mid)), activityId, mid).Scan(&got)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "tryDecrExtraDEVote error: %v", err)
		return
	}
	if got == 0 {
		return
	}
	if got > count {
		got = count
	}
	_, err = tx.Exec(fmt.Sprintf(sql4DecrUserDETotalExtraTimes, tableIdx(mid)), got, activityId, mid)
	return
}

// tryDecrExtraNEVote: 尝试扣减永久可用次数
func (d *Dao) tryDecrExtraNEVote(ctx context.Context, tx *sql.Tx, activityId, mid, count int64) (got int64, err error) {
	if count <= 0 {
		return
	}
	err = tx.QueryRow(fmt.Sprintf(sql4GetUserNETotalExtraTimes, tableIdx(mid)), activityId, mid).Scan(&got)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "tryDecrExtraNEVote error: %v", err)
		return
	}
	if got == 0 {
		return
	}
	if got > count {
		got = count
	}
	_, err = tx.Exec(fmt.Sprintf(sql4DecrUserNETotalExtraTimes, tableIdx(mid)), got, activityId, mid)
	return
}

func (d *Dao) incrUserExtraDECount(ctx context.Context, tx *sql.Tx, activityId, mid, count int64) (err error) {
	if tx == nil {
		_, err = d.db.Exec(ctx, fmt.Sprintf(sql4IncrUserDETotalExtraTimes, tableIdx(mid)), activityId, mid, timesTypeExtraDailyExpire, count, count)
	} else {
		_, err = tx.Exec(fmt.Sprintf(sql4IncrUserDETotalExtraTimes, tableIdx(mid)), activityId, mid, timesTypeExtraDailyExpire, count, count)
	}

	return
}

func (d *Dao) incrUserExtraNECount(ctx context.Context, tx *sql.Tx, activityId, mid, count int64) (err error) {
	if tx == nil {
		_, err = d.db.Exec(ctx, fmt.Sprintf(sql4IncrUserNETotalExtraTimes, tableIdx(mid)), activityId, mid, timesTypeExtraNotExpire, count, count)
	} else {
		_, err = tx.Exec(fmt.Sprintf(sql4IncrUserNETotalExtraTimes, tableIdx(mid)), activityId, mid, timesTypeExtraNotExpire, count, count)
	}
	return
}

// RawUserTodayVoteCountForActivity: 从DB中获取某用户在活动的当日投票次数
func (d *Dao) RawUserTodayVoteCountForActivity(ctx context.Context, tx *sql.Tx, activityId, mid int64) (baseCount, extraDECount, extraNECount, total int64, err error) {
	var rows *sql.Rows
	if tx == nil {
		rows, err = d.db.Query(ctx, fmt.Sprintf(sql4GetUserTodayVoteCountForActivityGroupByType, tableIdx(mid)), mid, activityId)
	} else {
		rows, err = tx.Query(fmt.Sprintf(sql4GetUserTodayVoteCountForActivityGroupByType, tableIdx(mid)), mid, activityId)
	}
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	if err != nil {
		log.Errorc(ctx, "RawUserTodayVoteCountForActivity db.QueryRow mid:%v, activity:%v error: %v", mid, activityId, err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var typ, c int64
		err = rows.Scan(&typ, &c)
		if err != nil {
			log.Errorc(ctx, "RawUserTodayVoteCountForActivity rows.Scan mid:%v, activity:%v error: %v", mid, activityId, err)
			return
		}
		total = total + c
		switch typ {
		case timesTypeExtraNotExpire:
			extraNECount = c
		case timesTypeExtraDailyExpire:
			extraDECount = c
		case timesTypeBase:
			baseCount = c
		}
	}
	err = rows.Err()
	return
}

func (d *Dao) UserGetTimes(ctx context.Context, input *api.VoteUserGetTimesReq) (res *api.VoteUserGetTimesResp, err error) {
	res = &api.VoteUserGetTimesResp{
		UserAvailVoteCount:    0,
		UserAvailTmpVoteCount: 0,
	}
	activity, err := d.Activity(ctx, input.ActivityId)
	if err != nil {
		return
	}
	res.UserAvailVoteCount, res.UserAvailTmpVoteCount, err = d.GetUserAvailVoteCount(ctx, activity.Rule, activity.Id, input.Mid)
	return
}

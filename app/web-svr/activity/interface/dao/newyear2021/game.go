package newyear2021

import (
	"context"
	"crypto/md5"
	xsql "database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"

	xecode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const (
	bizNameOfARCommitTimes     = "ar_commit_times"
	bizNameOfARCommitRequestID = "ar_request_id"
	bizNameOfARLogInsert       = "ar_log_insert"
	bizNameOfARScore           = "ar_user_score"
	bizNameOfARCoupon          = "ar_user_coupon"

	cacheKey4WebViewData = "bnj2021:pc:web_view:data"

	cacheKey4GameCommitTimes          = "bnj2021:gameCommit:%v:times:%v"
	cacheKey4GameCommitRequestID      = "bnj2021:gameCommit:%v:requestID:%v"
	cacheKey4GameCommitRequestID4User = "bnj2021:%v:gameCommit:requestID"
	cacheKey4GameScoreOfUser          = "bnj2021:%v:gameScore"
	cacheKey4CouponOfUser             = "bnj2021:%v:coupon"
	cacheKey4LastARRewardOfUser       = "bnj2021:%v:lastARData"
	cacheKey4LastPlayedInAR           = "bnj2021:%v:lastARPlay"
	cacheKey4LastDrawAwardOfUser      = "bnj2021:%v:lastDrawAward"

	bizKey4ARScoreOfDBRestore       = "db_restore_AR_score"
	bizKey4ARCouponOfDBRestore      = "db_restore_AR_coupon"
	bizKey4ARCommitTimesOfDBRestore = "db_restore_AR_commit_times"
)

const (
	sql4CountByMIDAndDate = `
SELECT count(1)
FROM bnj_ar_log_%v
WHERE mid = ?
	AND date_str = ?
`
	sql4InsertARGameLog = `
INSERT INTO bnj_ar_log_%v (mid, score, date_str, log_index)
VALUES (?, ?, ?, ?)
`
	sql4UserGameScore = `
SELECT sum(score)
FROM bnj_ar_log_%v
WHERE mid = ?
`
	sql4CouponCountByMID = `
SELECT type_code, sum(num)
FROM bnj_ar_coupon_%v
WHERE mid = ?
GROUP BY type_code
`
	sql4InsertCouponLog = `
INSERT INTO bnj_ar_coupon_%v (mid, type_code, comment, num)
VALUES (?, ?, ?, ?)
`
	sql4CouponSummaryByMID = `
SELECT num
FROM bnj_ar_coupon_summary_%v
WHERE mid = ?
`
	sql4DecreaseCoupon = `
UPDATE bnj_ar_coupon_summary_%v
SET num = num - ?
WHERE mid = ?
	AND num = ?
`
	sql4UpdateCouponSummary = `
INSERT INTO bnj_ar_coupon_summary_%v (mid, num)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE num = num + ?; 
`
	sql4FetchARSetting = `
SELECT setting
FROM bnj_ar_setting
ORDER BY id DESC
LIMIT 1
`
	sql4FetchScore2CouponRuleList = `
SELECT score, coupon
FROM bnj_ar_exchange_rule
WHERE is_deleted = 0
ORDER BY score ASC;
`
)

func gameCommitTimesCacheKey(mid int64, dateStr string) string {
	return fmt.Sprintf(cacheKey4GameCommitTimes, mid, dateStr)
}

func gameCommitRequestIDCacheKey(mid int64, dateStr string) string {
	filledStr := fmt.Sprintf("%02d", mid%_userSub)

	return fmt.Sprintf(cacheKey4GameCommitRequestID, dateStr, filledStr)
}

func gameCommitRequestIDCacheKey4User(mid int64) string {
	return fmt.Sprintf(cacheKey4GameCommitRequestID4User, mid)
}

func gameScoreCacheKey4User(mid int64) string {
	return fmt.Sprintf(cacheKey4GameScoreOfUser, mid)
}

func couponCacheKey4User(mid int64) string {
	return fmt.Sprintf(cacheKey4CouponOfUser, mid)
}

func lastARRewardCacheKey4User(mid int64) string {
	return fmt.Sprintf(cacheKey4LastARRewardOfUser, mid)
}

func lastDrawAwardCacheKey4User(mid int64) string {
	return fmt.Sprintf(cacheKey4LastDrawAwardOfUser, mid)
}

func lastARPlayedCacheKey4User(mid int64) string {
	return fmt.Sprintf(cacheKey4LastPlayedInAR, mid)
}

func ARConfirmInH5(ctx context.Context, mid int64) (confirmed int64) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := lastARPlayedCacheKey4User(mid)
	if exists, err := redis.Bool(conn.Do("EXISTS", cacheKey)); err == nil && exists {
		confirmed = 1
		_, _ = conn.Do("DEL", cacheKey)
	}

	return
}

func UpdateUserLastARIdentity(ctx context.Context, mid int64) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var bs []byte
	_, err = conn.Do("SETEX", lastARPlayedCacheKey4User(mid), tool.CalculateExpiredSeconds(30), bs)

	return
}

func FetchARExchangeRuleList(ctx context.Context) (list []*model.Score2Coupon, err error) {
	list = make([]*model.Score2Coupon, 0)
	var rows *sql.Rows

	rows, err = component.GlobalBnjDB.Query(ctx, sql4FetchScore2CouponRuleList)
	if err != nil {
		return
	}

	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		tmp := new(model.Score2Coupon)
		if tmpErr := rows.Scan(&tmp.Score, &tmp.Coupon); tmpErr != nil {
			continue
		}

		list = append(list, tmp)
	}

	err = rows.Err()

	return
}

func ResetLastDrawAward(ctx context.Context, mid int64, name string) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Do("SETEX", lastDrawAwardCacheKey4User(mid), 3600, []byte(name))

	return
}

func GetLastDrawAward(ctx context.Context, mid int64) (name string, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var bs []byte
	cacheKey := lastDrawAwardCacheKey4User(mid)
	bs, err = redis.Bytes(conn.Do("GET", cacheKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorc(ctx, "GeLastDrawAward(%v) err: %v", cacheKey, err)
		}

		return
	}

	name = string(bs)

	return
}

func ResetLastARReward(ctx context.Context, mid int64, info model.Score2Coupon) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	bs, _ := json.Marshal(info)
	_, err = conn.Do("SETEX", lastARRewardCacheKey4User(mid), 3600, bs)

	return
}

func GetLastARReward(ctx context.Context, mid int64) (reward *model.Score2Coupon, err error) {
	reward = new(model.Score2Coupon)
	{
		reward.Score = 0
		reward.Coupon = 0
	}

	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var bs []byte
	cacheKey := lastARRewardCacheKey4User(mid)
	bs, err = redis.Bytes(conn.Do("GET", cacheKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorc(ctx, "GetLastARReward(%v) err: %v", cacheKey, err)
		}

		err = nil

		return
	}

	_ = json.Unmarshal(bs, reward)

	return
}

func FetchWebViewResource4PC(ctx context.Context) (m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var bs []byte
	bs, err = redis.Bytes(conn.Do("GET", cacheKey4WebViewData))
	if err != nil {
		return
	}

	_ = json.Unmarshal(bs, &m)

	return
}

func md5RequestID(requestID, mid int64) string {
	hasher := md5.New()
	requestIDStr := strconv.FormatInt(requestID, 10)
	sum := fmt.Sprintf("%02d", mid%_userSub)
	hasher.Write([]byte(requestIDStr))

	return hex.EncodeToString(hasher.Sum([]byte(sum)))
}

func GenGameCommitRequestID(ctx context.Context, mid int64, dateStr string) (requestID string, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var incrID int64
	cacheKey := gameCommitRequestIDCacheKey(mid, dateStr)
	incrID, err = redis.Int64(conn.Do("INCR", cacheKey))
	if err != nil {
		return
	}

	expiredSec := tool.CalculateExpiredSeconds(0)
	if incrID == 1 {
		if _, cacheErr := conn.Do("EXPIRE", cacheKey, expiredSec); cacheErr != nil {
			log.Errorc(ctx, "expire_dist_requestID(%v) err: %v", cacheKey, cacheErr)
		}
	}

	requestID = md5RequestID(incrID, mid)
	cacheKey4User := gameCommitRequestIDCacheKey4User(mid)
	if _, cacheErr := conn.Do("SETEX", cacheKey4User, 600, []byte(requestID)); cacheErr != nil {
		err = cacheErr
		tool.IncrCacheResetMetric(bizNameOfARCommitRequestID, tool.StatusOfFailed)
		log.Errorc(ctx, "expire_user_requestID(%v) err: %v", cacheKey4User, cacheErr)
	}

	return
}

func IsRequestIDValid(ctx context.Context, mid int64, requestID string) (isValid bool, err error) {
	cacheKey := gameCommitRequestIDCacheKey4User(mid)
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var requestIDInServer string
	requestIDInServer, err = redis.String(conn.Do("GET", cacheKey))
	if err != nil {
		return
	}

	isValid = requestID == requestIDInServer
	if _, cacheErr := conn.Do("DEL", cacheKey); cacheErr != nil {
		tool.IncrCacheResetMetric(bizNameOfARCommitRequestID, tool.StatusOfFailed)
		log.Errorc(ctx, "del_user_requestID(%v) err: %v", cacheKey, cacheErr)
	}

	return
}

func delUserGameScoreCache(ctx context.Context, mid int64) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := gameScoreCacheKey4User(mid)
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		tool.IncrCacheResetMetric(bizNameOfARScore, tool.StatusOfFailed)
		log.Errorc(ctx, "del_user_requestID(%v) err: %v", cacheKey, err)
	}
}

func delUserCommitTimesAndScoreCache(ctx context.Context, mid int64, dateStr string) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := gameCommitTimesCacheKey(mid, dateStr)
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		tool.IncrCacheResetMetric(bizNameOfARCommitTimes, tool.StatusOfFailed)
		log.Errorc(ctx, "del_user_commitTimes(%v) err: %v", cacheKey, err)
	}

	cacheKey = gameScoreCacheKey4User(mid)
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		tool.IncrCacheResetMetric(bizNameOfARScore, tool.StatusOfFailed)
		log.Errorc(ctx, "del_user_score(%v) err: %v", cacheKey, err)
	}

	cacheKey = couponCacheKey4User(mid)
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		tool.IncrCacheResetMetric(bizNameOfARCoupon, tool.StatusOfFailed)
		log.Errorc(ctx, "del_user_coupon(%v) err: %v", cacheKey, err)
	}
}

func delUserCouponCache(ctx context.Context, mid int64) (err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := couponCacheKey4User(mid)
	for i := 0; i < 3; i++ {
		_, err = conn.Do("DEL", cacheKey)
		if err != nil {
			tool.IncrCacheResetMetric(bizNameOfARCoupon, tool.StatusOfFailed)
			log.Errorc(ctx, "del_user_coupon(%v) err: %v", cacheKey, err)

			continue
		}

		break
	}

	return
}

func delUserCommitTimesCache(ctx context.Context, mid int64, dateStr string) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := gameCommitTimesCacheKey(mid, dateStr)
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		tool.IncrCacheResetMetric(bizNameOfARCommitTimes, tool.StatusOfFailed)
		log.Errorc(ctx, "del_user_commitTimes(%v) err: %v", cacheKey, err)
	}
}

func FetchUserCoupon(ctx context.Context, mid int64) (coupon *model.UserCoupon, err error) {
	coupon = new(model.UserCoupon)
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	var bsInCache []byte
	cacheKey := couponCacheKey4User(mid)
	bsInCache, err = redis.Bytes(conn.Do("GET", cacheKey))
	if err != nil && err != redis.ErrNil {
		err = ecode.ServerErr

		return
	}

	if err == nil {
		_ = json.Unmarshal(bsInCache, coupon)

		return
	}

	if err == redis.ErrNil {
		err = nil
		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizKey4ARCouponOfDBRestore) {
			coupon, err = fetchUserCouponFromDB(ctx, mid)
			if err == nil {
				bs, _ := json.Marshal(coupon)
				// expiredSec := tool.CalculateExpiredSeconds(0)
				if _, cacheErr := conn.Do("SETEX", cacheKey, 1800, bs); cacheErr != nil {
					err = cacheErr
					tool.IncrCacheResetMetric(bizNameOfARCommitTimes, tool.StatusOfFailed)
					log.Errorc(ctx, "expire_user_game_score(%v) err: %v", cacheKey, cacheErr)
				}
			}
		} else {
			// TODO
		}
	}

	return
}

func fetchUserCouponFromDB(ctx context.Context, mid int64) (coupon *model.UserCoupon, err error) {
	coupon = new(model.UserCoupon)
	sqlStr := fmt.Sprintf(sql4CouponSummaryByMID, fmt.Sprintf("%02d", mid%_userSub))

	row := component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid)
	err = row.Scan(&coupon.ND)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func FetchUserGameScore(ctx context.Context, mid int64) (score int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := gameScoreCacheKey4User(mid)
	score, err = redis.Int64(conn.Do("GET", cacheKey))
	if err != nil && err != redis.ErrNil {
		err = ecode.ServerErr

		return
	}

	if err == redis.ErrNil {
		err = nil
		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizKey4ARScoreOfDBRestore) {
			score, err = fetchUserGameScoreFromDB(ctx, mid)
			if err == nil {
				scoreStr := strconv.FormatInt(score, 10)
				expiredSec := tool.CalculateExpiredSeconds(0)
				if _, cacheErr := conn.Do("SETEX", cacheKey, expiredSec, []byte(scoreStr)); cacheErr != nil {
					err = cacheErr
					tool.IncrCacheResetMetric(bizNameOfARCommitTimes, tool.StatusOfFailed)
					log.Errorc(ctx, "expire_user_game_score(%v) err: %v", cacheKey, cacheErr)
				}
			}
		} else {
			// TODO
		}
	}

	return
}

func fetchUserGameScoreFromDB(ctx context.Context, mid int64) (score int64, err error) {
	sqlStr := fmt.Sprintf(sql4UserGameScore, fmt.Sprintf("%02d", mid%_userSub))
	row := component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid)
	err = row.Scan(&score)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func FetchGameCommitTimes(ctx context.Context, mid int64, dateStr string) (times int64, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := gameCommitTimesCacheKey(mid, dateStr)
	times, err = redis.Int64(conn.Do("GET", cacheKey))
	if err != nil && err != redis.ErrNil {
		err = ecode.ServerErr

		return
	}

	if err == redis.ErrNil {
		err = nil
		if tool.IsLimiterAllowedByUniqBizKey(tool.BizLimitKey4DBRestoreOfLow, bizKey4ARCommitTimesOfDBRestore) {
			times, err = fetchGameCommitTimesFromDB(ctx, mid, dateStr)
			if err == nil {
				timesStr := strconv.FormatInt(times, 10)
				expiredSec := tool.CalculateExpiredSeconds(0)
				if _, cacheErr := conn.Do("SETEX", cacheKey, expiredSec, []byte(timesStr)); cacheErr != nil {
					err = cacheErr
					tool.IncrCacheResetMetric(bizNameOfARCommitTimes, tool.StatusOfFailed)
					log.Errorc(ctx, "expire_user_commit_times(%v) err: %v", cacheKey, cacheErr)
				}
			}
		}
	}

	return
}

func fetchGameCommitTimesFromDB(ctx context.Context, mid int64, dateStr string) (times int64, err error) {
	defer func() {
		if err != nil {
			tool.AddDBErrMetrics(bizNameOfARCommitTimes)
		}
	}()

	sqlStr := fmt.Sprintf(sql4CountByMIDAndDate, fmt.Sprintf("%02d", mid%_userSub))
	row := component.GlobalBnjDB.QueryRow(ctx, sqlStr, mid, dateStr)
	err = row.Scan(&times)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func UpsertARCoupon(ctx context.Context, arLog *model.ARGameLog) (err error) {
	var tx *sql.Tx
	tx, err = component.GlobalBnjDB.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			tool.AddDBErrMetrics(bizNameOfARLogInsert)
			if txErr := tx.Rollback(); txErr != nil {
				log.Errorc(ctx, "InsertARGameLog rollback err: %v", txErr)
			}

			return
		}

		err = tx.Commit()
		if err != nil {
			log.Errorc(ctx, "InsertARGameLog commit err: %v", err)

			return
		}

		delUserCommitTimesAndScoreCache(ctx, arLog.MID, arLog.Date)
	}()

	suffix := fmt.Sprintf("%02d", arLog.MID%_userSub)
	sql4GameLog := fmt.Sprintf(sql4InsertARGameLog, suffix)
	_, err = tx.Exec(sql4GameLog, arLog.MID, arLog.Score, arLog.Date, arLog.Index)
	if err != nil {
		return
	}

	if arLog.Coupon > 0 {
		sql4Coupon := fmt.Sprintf(sql4UpdateCouponSummary, suffix)
		_, err = tx.Exec(sql4Coupon, arLog.MID, arLog.Coupon, arLog.Coupon)
	}

	return
}

func IncrARCoupon(ctx context.Context, mid, count int64) (err error) {
	defer func() {
		if err != nil {
			tool.AddDBErrMetrics(bizNameOfARLogInsert)
			return
		}
		_ = delUserCouponCache(context.Background(), mid)
	}()
	suffix := fmt.Sprintf("%02d", mid%_userSub)
	if count > 0 {
		sql4Coupon := fmt.Sprintf(sql4UpdateCouponSummary, suffix)
		_, err = component.GlobalBnjDB.Exec(ctx, sql4Coupon, mid, count, count)
	}
	return
}

func decreaseARCouponByCAS(ctx context.Context, mid, decr, oldNum int64) (affectRows int64, err error) {
	var result xsql.Result
	suffix := fmt.Sprintf("%02d", mid%_userSub)
	query := fmt.Sprintf(sql4DecreaseCoupon, suffix)
	result, err = component.GlobalBnjDB.Exec(ctx, query, decr, mid, oldNum)
	if err == nil {
		affectRows, err = result.RowsAffected()
		if err == nil {
			_ = delUserCouponCache(ctx, mid)
		}
	}

	if err != nil {
		err = ecode.ServerErr
	}

	return
}

func DecreaseARCouponByCAS(ctx context.Context, mid, num int64) (err error) {
	coupon := new(model.UserCoupon)
	coupon, err = FetchUserCoupon(ctx, mid)
	if err != nil {
		return
	}

	if coupon.ND < num {
		err = xecode.BNJNoEnoughCoupon2Draw

		return
	}

	var affectRows int64
	affectRows, err = decreaseARCouponByCAS(ctx, mid, num, coupon.ND)
	if err == nil {
		if affectRows == 0 {
			err = xecode.BNJNoEnoughCoupon2Draw
		}
	}

	return
}

package bnj

import (
	"context"
	xsql "database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/bnj"

	"go-common/library/database/sql"
	"go-common/library/log"
)

const (
	expiredSecondsOf30Day = 30 * 86400

	bizNameOfDelUserLiveLotteryReceive = "bnj_live_lottery_receive_del"

	cacheKey4WebViewData = "bnj2021:pc:web_view:data"

	cacheKey4BnjLotteryRecordOfUser    = "bnj2021:unReceived:reward:%v:live"
	cacheKey4BnjLotteryRecordOfReserve = "bnj2021:unReceived:reward:%v:reserve"
	cacheKey4BnjLotteryReceiveOfUser   = "bnj2021:live:lottery:%v:receive"

	sql4Bnj2021LotteryRuleList = `
SELECT duration, UNIX_TIMESTAMP(start_time), UNIX_TIMESTAMP(end_time)
FROM bnj_live_lottery_rule
WHERE end_time > now();
`
	sql4Bnj20212InsertUnReceivedUser = `
INSERT INTO bnj_live_user_%v (mid, duration, unique_id)
VALUES (?, ?, ?);
`
	sql4Bnj2021UnReceivedUserListWithID = `
SELECT /*master*/ id, mid
FROM bnj_live_user_%v FORCE INDEX (PRIMARY)
WHERE received = 0
	AND duration = ?
	AND id > ?
ORDER BY id ASC
LIMIT ?
`
	sql4LiveUserRecords = `
SELECT duration, received
FROM bnj_live_user_%v
WHERE mid = ?
`
	sql4Bnj2021Lottery2MarkAsReceived = `
UPDATE bnj_live_user_%v
SET received = 2
WHERE mid = ?
	AND duration = ?
`
	sql4Bnj2021LiveCouponLottery2UpdateReward = `
UPDATE bnj_ar_draw_log_%v
SET received = 1, reward = ?
WHERE mid = ?
	AND rec_unix = ?
	AND no = ?
ORDER BY ctime DESC
LIMIT 1
`
	sql4Bnj2021LiveCouponLottery2MarkAsReceived = `
UPDATE bnj_ar_draw_log_%v
SET received = 2
WHERE mid = ?
	AND rec_unix = ?
	AND no = ?
ORDER BY ctime DESC
LIMIT 1
`
	sql4Bnj2021LiveCouponLottery2BatchInsert = `
INSERT INTO bnj_ar_draw_log_%v (mid, rec_unix, no, received, reward)
values%v
`
	sql4Bnj2021Lottery2UpdateReward = `
UPDATE bnj_live_user_%v
SET reward = ?, received = 1
WHERE mid = ?
	AND duration = ?
`
	sql4Bnj2021LotteryLastID2Update = `
INSERT INTO bnj_live_lottery_receive_last_id (duration, suffix, last_received_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE last_received_id = ?; 
`
	sql4Bnj2021LotteryLastID2Select = `
SELECT /*master*/ last_received_id
FROM bnj_live_lottery_receive_last_id
WHERE duration = ?
	AND suffix = ?;
`
)

func ResetWebViewData(m map[string]interface{}) (err error) {
	bs, _ := json.Marshal(m)
	_, err = component.GlobalCache.Do(
		context.Background(),
		"SETEX",
		cacheKey4WebViewData,
		expiredSecondsOf30Day,
		string(bs))

	return
}

func FetchLastReceiveIDByDurationAndSuffix(ctx context.Context, duration int64, suffix string) (lastID int64, err error) {
	for i := 0; i < 10; i++ {
		err = component.GlobalBnjDB.QueryRow(ctx, sql4Bnj2021LotteryLastID2Select, duration, suffix).Scan(&lastID)
		if err == sql.ErrNoRows {
			err = nil
		}

		if err == nil {
			break
		}
	}

	return
}

func UpdateLastIDByDurationAndSuffix(ctx context.Context, duration, lastID int64, suffix string) (err error) {
	_, err = component.GlobalBnjDB.Exec(
		ctx,
		sql4Bnj2021LotteryLastID2Update,
		duration,
		suffix,
		lastID,
		lastID)

	return
}

func cacheKey4UserLotteryReceive(mid int64) string {
	return fmt.Sprintf(cacheKey4BnjLotteryReceiveOfUser, mid)
}

func cacheKey4UserLotteryOfLive(mid int64) string {
	return fmt.Sprintf(cacheKey4BnjLotteryRecordOfUser, mid)
}

func cacheKey4UserLotteryOfReserve(mid int64) string {
	return fmt.Sprintf(cacheKey4BnjLotteryRecordOfReserve, mid)
}

func RPushUserUnReceivedRewardInLiveDraw(ctx context.Context, reward *bnj.UserRewardInLiveRoom) (err error) {
	bs, _ := json.Marshal(reward)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(50*i) * time.Millisecond)
		switch reward.SceneID {
		case bnj.SceneID4Reserve:
			cacheKey := cacheKey4UserLotteryOfReserve(reward.MID)
			_, err = component.GlobalCache.Do(ctx, "SETEX", cacheKey, expiredSecondsOf30Day, string(bs))
		default:
			cacheKey := cacheKey4UserLotteryOfLive(reward.MID)
			_, err = component.GlobalCache.Do(ctx, "RPUSH", cacheKey, string(bs))
			if err == nil {
				_, _ = component.GlobalCache.Do(ctx, "EXPIRE", cacheKey, expiredSecondsOf30Day)
			} else {
				log.Error("RPushUserUnReceivedRewardInLiveDraw_err:", string(bs), time.Now())

				return
			}
		}

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Error("RPushUserUnReceivedRewardInLiveDraw failed, err: %v, info: %v", err, string(bs))
	}

	return
}

func UpdateUserRewardInLive(ctx context.Context, mid, duration int64, reward string) (affectRows int64, err error) {
	query := fmt.Sprintf(sql4Bnj2021Lottery2UpdateReward, fmt.Sprintf("%02d", mid%100))
	var result xsql.Result
	result, err = component.GlobalBnjDB.Exec(ctx, query, reward, mid, duration)
	if err != nil {
		log.Error(
			"SetUserBnjLotteryInfo failed, mid(%v), duration(%v), reward(%v), db err: %v",
			mid,
			duration,
			reward,
			err)

		return
	}

	affectRows, err = result.RowsAffected()

	return
}

func InsertUnReceivedUserInfo(ctx context.Context, info *bnj.UserInLiveRoomFor2021) (err error) {
	query := fmt.Sprintf(sql4Bnj20212InsertUnReceivedUser, fmt.Sprintf("%02d", info.MID%100))
	_, err = component.GlobalBnjDB.Exec(ctx, query, info.MID, info.Duration, info.UniqueID)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			err = nil

			return
		}

		bs, _ := json.Marshal(info)
		log.Error("InsertUnReceivedUserInfo failed, detail: %v, err: %v", string(bs), err)
	}

	return
}

func BatchInsertLiveUserCouponLog(ctx context.Context, list []*bnj.UserCouponLogInLiveRoom) (err error) {
	var values string
	valuesFormat := "(?, ?, ?, ?, ?)"
	args := make([]interface{}, 0)

	for _, v := range list {
		if values == "" {
			values = valuesFormat
		} else {
			values = fmt.Sprintf("%v, %v", values, valuesFormat)
		}

		args = append(args, v.MID, v.ReceiveUnix, v.No, 0, "")
	}

	query := fmt.Sprintf(sql4Bnj2021LiveCouponLottery2BatchInsert, fmt.Sprintf("%02d", list[0].MID%100), values)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(50*i) * time.Millisecond)
		_, err = component.GlobalBnjDB.Exec(ctx, query, args...)
		if err == nil {
			break
		}
	}

	return
}

func UpdateUserLiveCouponLotteryReward(ctx context.Context, reward *bnj.UserRewardInLiveRoom) (
	affectRows int64, err error) {
	query := fmt.Sprintf(sql4Bnj2021LiveCouponLottery2UpdateReward, fmt.Sprintf("%02d", reward.MID%100))
	bs, _ := json.Marshal(reward.Reward)
	var result xsql.Result
	result, err = component.GlobalBnjDB.Exec(ctx, query, string(bs), reward.MID, reward.ReceiveUnix, reward.No)
	if err == nil {
		affectRows, err = result.RowsAffected()
	}

	return
}

func MarkBnjLiveUserCouponLotteryReceived(ctx context.Context, reward *bnj.UserRewardInLiveRoom) (
	affectRows int64, err error) {
	query := fmt.Sprintf(sql4Bnj2021LiveCouponLottery2MarkAsReceived, fmt.Sprintf("%02d", reward.MID%100))
	var result xsql.Result
	result, err = component.GlobalBnjDB.Exec(ctx, query, reward.MID, reward.ReceiveUnix, reward.No)
	if err == nil {
		affectRows, err = result.RowsAffected()
	}

	return
}

func MarkBnjLiveUserLotteryReceived(ctx context.Context, mid, duration int64) (affectRows int64, err error) {
	query := fmt.Sprintf(sql4Bnj2021Lottery2MarkAsReceived, fmt.Sprintf("%02d", mid%100))
	var result xsql.Result
	result, err = component.GlobalBnjDB.Exec(ctx, query, mid, duration)
	if err == nil {
		affectRows, err = result.RowsAffected()
	}

	return
}

func FetchBnjLiveUserRecordList(ctx context.Context, mid int64) (
	list []map[string]interface{}, err error) {
	list = make([]map[string]interface{}, 0)

	var rows *sql.Rows
	query := fmt.Sprintf(sql4LiveUserRecords, fmt.Sprintf("%02d", mid%100))
	rows, err = component.GlobalBnjDB.Query(ctx, query, mid)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var duration, received int64
		tmpErr := rows.Scan(
			&duration,
			&received)
		if tmpErr == nil {
			m := make(map[string]interface{}, 0)
			{
				m["mid"] = mid
				m["count"] = duration
				switch received {
				case 0:
					m["received"] = "未预抽"
				case 1:
					m["received"] = "未抽奖"
				case 2:
					m["received"] = "已领取"
				default:
					m["received"] = "未知状态"
				}
			}
			list = append(list, m)
		}
	}

	err = rows.Err()

	return
}

func FetchBnjUnReceivedUserList(ctx context.Context, suffix string, lastID, duration, limit int64) (
	list []*bnj.UserInLiveRoomFor2021, err error) {
	list = make([]*bnj.UserInLiveRoomFor2021, 0)

	var rows *sql.Rows
	query := fmt.Sprintf(sql4Bnj2021UnReceivedUserListWithID, suffix)
	rows, err = component.GlobalBnjDB.Query(ctx, query, duration, lastID, limit)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		user := new(bnj.UserInLiveRoomFor2021)
		tmpErr := rows.Scan(
			&user.ID,
			&user.MID)
		if tmpErr == nil {
			list = append(list, user)
		}
	}

	err = rows.Err()

	return
}

func FetchBnjLotteryRuleFor2021(ctx context.Context) (rules []*bnj.LotteryRuleFor2021, err error) {
	rules = make([]*bnj.LotteryRuleFor2021, 0)
	var rows *sql.Rows
	rows, err = component.GlobalBnjDB.Query(ctx, sql4Bnj2021LotteryRuleList)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		rule := new(bnj.LotteryRuleFor2021)
		tmpErr := rows.Scan(
			&rule.Duration,
			&rule.StartTime,
			&rule.EndTime)
		if tmpErr == nil {
			rules = append(rules, rule)
		}
	}

	err = rows.Err()

	return
}

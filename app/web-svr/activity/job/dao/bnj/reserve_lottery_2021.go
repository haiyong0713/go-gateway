package bnj

import (
	"context"
	xsql "database/sql"
	"fmt"
	"time"

	"go-common/library/log"

	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/bnj"

	"go-common/library/database/sql"
)

const (
	sql4Bnj2021ReserveLotteryRuleList = `
SELECT count, UNIX_TIMESTAMP(start_time), UNIX_TIMESTAMP(end_time), reward_id, activity_id
FROM bnj_reserve_reward_rule
WHERE end_time > now();
`
	sql4Bnj2021ReserveRewardLastID2Select = `
SELECT /*master*/ last_received_id
FROM bnj_reserve_reward_receive_last_id
WHERE count = ?
	AND suffix = ?;
`
	sql4Bnj2021ReserveReward2UpdateLastID = `
INSERT INTO bnj_reserve_reward_receive_last_id (count, suffix, last_received_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE last_received_id = ?;
`
	sql4FetchReservedMIDList = `
SELECT /*master*/ id, mid
FROM act_reserve_%v
WHERE sid = ?
	AND id > ?
ORDER BY id ASC
LIMIT 1000
`
	sql4UpsertReserveRewardLog = `
INSERT INTO bnj_reserve_reward_%v (mid, count, received, reward)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE received = ?, reward = ?;
`
	sql4UpdateReserveDrawRewardAsReceived = `
UPDATE bnj_reserve_reward_%v
SET received = 2
WHERE mid = ?
	AND count = 2;
`
)

func UpdateReserveDrawRewardAsReceived(ctx context.Context, mid int64) (affectedRows int64, err error) {
	defer func() {
		if err != nil {
			logStr := "UpdateReserveDrawRewardAsReceived failed, reserveCount: %v, mid: %v, err: %v"
			log.Error(logStr, 0, mid, err)
		}
	}()

	query := fmt.Sprintf(sql4UpdateReserveDrawRewardAsReceived, fmt.Sprintf("%02d", mid%100))
	var result xsql.Result
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(50*i) * time.Millisecond)
		result, err = component.GlobalBnjDB.Exec(
			ctx,
			query,
			mid)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	affectedRows, err = result.RowsAffected()

	return
}

func UpsertReserveRewardLog(ctx context.Context, mid, count, received int64, reward string) (affectedRows int64, err error) {
	defer func() {
		if err != nil {
			logStr := "UpsertReserveRewardLog failed, reserveCount: %v, mid: %v, reward: %v, err: %v"
			log.Error(logStr, count, mid, reward, err)
		}
	}()

	query := fmt.Sprintf(sql4UpsertReserveRewardLog, fmt.Sprintf("%02d", mid%100))
	var result xsql.Result
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(50*i) * time.Millisecond)
		result, err = component.GlobalBnjDB.Exec(
			ctx,
			query,
			mid,
			count,
			received,
			reward,
			bnj.RewardTypeOfReceived,
			reward)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	affectedRows, err = result.RowsAffected()

	return
}

func ResetBnj2021ReserveRewardLastRecID(ctx context.Context, lastID, count int64, suffix string) (err error) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(50*i) * time.Millisecond)
		_, err = component.GlobalBnjDB.Exec(
			ctx,
			sql4Bnj2021ReserveReward2UpdateLastID,
			count,
			suffix,
			lastID,
			lastID)
		if err == nil {
			break
		}
	}

	return
}

func FetchReservedMID(ctx context.Context, activityID, lastID int64, suffix string) (list []*bnj.ReservedUser, err error) {
	list = make([]*bnj.ReservedUser, 0)
	query := fmt.Sprintf(sql4FetchReservedMIDList, suffix)

	var rows *sql.Rows
	rows, err = component.GlobalDBOfRead.Query(ctx, query, activityID, lastID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		tmp := new(bnj.ReservedUser)
		if tmpErr := rows.Scan(&tmp.ID, &tmp.MID); tmpErr == nil {
			list = append(list, tmp)
		}
	}

	err = rows.Err()

	return
}

func FetchLastReceiveIDByCountAndSuffix(ctx context.Context, count int64, suffix string) (lastID int64, err error) {
	err = component.GlobalBnjDB.QueryRow(ctx, sql4Bnj2021ReserveRewardLastID2Select, count, suffix).Scan(&lastID)
	if err == sql.ErrNoRows {
		err = nil
	}

	return
}

func FetchBnjReserveRewardRuleFor2021(ctx context.Context) (rules map[int64]*bnj.ReserveRewardRuleFor2021, err error) {
	rules = make(map[int64]*bnj.ReserveRewardRuleFor2021, 0)
	var rows *sql.Rows
	rows, err = component.GlobalBnjDB.Query(ctx, sql4Bnj2021ReserveLotteryRuleList)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		rule := new(bnj.ReserveRewardRuleFor2021)
		tmpErr := rows.Scan(
			&rule.Count,
			&rule.StartTime,
			&rule.EndTime,
			&rule.RewardID,
			&rule.ActivityID)
		if tmpErr == nil {
			rules[rule.Count] = rule
		}
	}

	err = rows.Err()

	return
}

package mission

import (
	"context"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/model/mission"
)

const (
	updateReceiveFinishStateSql = `
update
  act_user_task_record_%s_%s
set
  task_rewards_status = ?,
  reason = ?
where
  mid = ?
  and task_id = ?
  and complete_period = ?
`

	updateReceiveStartingStateSql = `
update
  act_user_task_record_%s_%s
set
  task_rewards_status = ?,
  serial_num = ?,
  stock_no = ?
where
  mid = ?
  and task_id = ?
  and complete_period = ?
`

	checkStockStateSql = `
select count(*) from
  act_user_task_record_%s_%s
where
  mid = ?
  and task_id = ?
  and serial_num = ?
  and stock_no = ?
`
)

func (d *Dao) UpdateUserReceiveRecordToFinish(ctx context.Context, record *mission.UserCompleteRecord, period int64) (err error) {
	receiveState := mission.TaskRewardStatusSuccess
	tablePrefixFirst, tablePrefixSecond := getUserTable(record.ActId, record.Mid)
	err = retry.WithAttempts(ctx, "UpdateUserReceiveRecord", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		_, err = d.db.Exec(ctx, fmt.Sprintf(updateReceiveFinishStateSql, tablePrefixFirst, tablePrefixSecond), receiveState, "", record.Mid, record.TaskId, period)
		return
	})
	record.TaskRewardsStatus = int64(receiveState)
	return
}

func (d *Dao) UpdateUserReceiveRecordToStarting(ctx context.Context, record *mission.UserCompleteRecord, period int64, uniqueId string, stockNo string) (err error) {
	receiveState := mission.TaskRewardStatusIn
	tablePrefixFirst, tablePrefixSecond := getUserTable(record.ActId, record.Mid)
	err = retry.WithAttempts(ctx, "UpdateUserReceiveRecord", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		_, err = d.db.Exec(ctx,
			fmt.Sprintf(updateReceiveStartingStateSql,
				tablePrefixFirst,
				tablePrefixSecond),
			receiveState,
			uniqueId,
			stockNo,
			record.Mid,
			record.TaskId,
			period,
		)
		return
	})
	record.TaskRewardsStatus = int64(receiveState)
	return
}

func (d *Dao) CheckStock(ctx context.Context, mid, actId, taskId int64, serialNum, stockNo string) (exist bool, err error) {
	tablePrefixFirst, tablePrefixSecond := getUserTable(actId, mid)
	var count int64
	err = retry.WithAttempts(ctx, "UpdateUserReceiveRecord", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		err = d.db.QueryRow(ctx,
			fmt.Sprintf(checkStockStateSql,
				tablePrefixFirst,
				tablePrefixSecond),
			mid,
			taskId,
			serialNum,
			stockNo,
		).Scan(&count)
		return
	})
	exist = count > 0
	return
}

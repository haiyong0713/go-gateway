package guess

import (
	"context"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/guess"
	"go-gateway/app/web-svr/activity/job/tool"
	"time"
)

const (
	sqlOfInsertFailedDetail = `
INSERT INTO act_finish_error_guess (main_id, result_id, business, oid, table_index, odds)
VALUES (?, ?, ?, ?, ?, ?) `
	sqlOfMarkFailedDetailAsDone = `
UPDATE act_finish_error_guess
SET retry_status = 1
WHERE id = ?`
	sqlOfListAllFailedDetail = `
SELECT id, main_id, result_id, business, oid
	, table_index, odds, ctime
FROM act_finish_error_guess
WHERE retry_status = 0`
	notification4SettlementOfRetryGetListError = "结算失败自动重试: 检测到 %v 个超过12小时未结算成功的任务"
)

// AddFinishGuessFailTask: add fail task to retry pool, will be a worker to retry doing this task.
func (d *Dao) AddFinishGuessFailTask(ctx context.Context, task guess.FinishGuessFailTask) error {
	_, err := d.db.Exec(ctx, sqlOfInsertFailedDetail, task.MainID, task.ResultID, task.Business, task.Oid, task.TableIndex, task.Odds)
	if err != nil {
		log.Errorc(ctx, "AddFinishGuessFailTask save to db error: %v", err)
	}
	return err
}

// MarkFinishGuessFailTaskAsDone: used for worker retry task successfully, it should mark this task as done.
func (d *Dao) MarkFinishGuessFailTaskAsDone(ctx context.Context, primaryKey int64) error {
	_, err := d.db.Exec(ctx, sqlOfMarkFailedDetailAsDone, primaryKey)
	if err != nil {
		log.Errorc(ctx, "MarkFinishGuessFailTaskAsDone update db error: %v", err)
	}
	return err
}

// GetAllFinishGuessFailTask: get all task that is not done
func (d *Dao) GetAllFinishGuessFailTask(ctx context.Context) (res []*guess.FinishGuessFailTask, err error) {
	var rows *sql.Rows
	for i := 0; i < 3; i++ {
		rows, err = d.db.Query(ctx, sqlOfListAllFailedDetail)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Errorc(ctx, "GetAllFinishGuessFailTask query db error: %v", err)
		return
	}
	defer rows.Close()
	res = make([]*guess.FinishGuessFailTask, 0)
	oldTasks := 0
	for rows.Next() {
		s := &guess.FinishGuessFailTask{}
		if err = rows.Scan(&s.Id, &s.MainID, &s.ResultID, &s.Business, &s.Oid, &s.TableIndex, &s.Odds, &s.CreateTime); err != nil {
			log.Errorc(ctx, "rows.Scan error(%v)", err)
			return
		}
		//check if there had very old fail tasks.
		if time.Now().After(s.CreateTime.Time().Add(12 * time.Hour)) {
			log.Errorc(ctx, "found old fail task :id(%v), MainId(%v), ResultId(%v), Business(%v), Oid(%v), TableIndex(%v), Odds(%v), CreateTime(%v)",
				s.Id, s.MainID, s.ResultID, s.Business, s.Oid, s.TableIndex, s.Odds, s.CreateTime.Time().Format(time.RFC3339))
			oldTasks++
		}
		res = append(res, s)
	}
	if oldTasks > 0 {
		if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, fmt.Sprintf(notification4SettlementOfRetryGetListError, oldTasks), true); err == nil {
			log.Errorc(ctx, " GetAllFinishGuessFailTask: send message via WeChat error: %v", tool.SendCorpWeChatRobotAlarm(bs))
		}
	}
	err = rows.Err()
	return res, err
}

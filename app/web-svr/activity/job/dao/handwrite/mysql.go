package handwrite

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/handwrite"
	"strings"

	"github.com/pkg/errors"
)

const (
	handWriteName     = "act_handwrite_mid"
	handWriteTaskName = "act_handwrite_task"
)

const (
	_midListDistinctSQL = "SELECT DISTINCT(mid) FROM %s WHERE mid IN (%s)"
	_midTaskSQL         = "SELECT mid,task_detail,task_type,finish_count,finish_time FROM %s WHERE mid IN (%s) and task_type =?"
	_midTaskAllSQL      = "SELECT mid,task_detail,task_type,finish_count,finish_time FROM %s  LIMIT ?,?"
	_midTaskAdd         = "INSERT INTO %s (`mid`,`task_type`,`task_detail`,`finish_count`,`finish_time`) VALUES %s ON DUPLICATE KEY UPDATE mid=VALUES(mid), task_type=VALUES(task_type), task_detail=VALUES(task_detail), finish_count=VALUES(finish_count), finish_time=VALUES(finish_time)"
)

// MidListDistinct get mid lit
func (d *dao) MidListDistinct(ctx context.Context, mids []int64) (rs []*handwrite.Mid, err error) {
	rs = []*handwrite.Mid{}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_midListDistinctSQL, handWriteName, xstr.JoinInts(mids)))
	if err != nil {
		err = errors.Wrap(err, "MidListDistinct:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &handwrite.Mid{}
		err = rows.Scan(&r.Mid)
		if err != nil {
			err = errors.Wrap(err, "MidListDistinct:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "MidListDistinct:rows.Err")
	}
	return
}

// MidTask get mid lit
func (d *dao) MidTask(ctx context.Context, mids []int64, taskType int) (rs []*handwrite.MidTaskDB, err error) {
	rs = []*handwrite.MidTaskDB{}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_midTaskSQL, handWriteTaskName, xstr.JoinInts(mids)), taskType)
	if err != nil {
		err = errors.Wrap(err, "MidTask:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &handwrite.MidTaskDB{}
		err = rows.Scan(&r.Mid, &r.TaskDetail, &r.TaskType, &r.FinishCount, &r.FinishTime)
		if err != nil {
			err = errors.Wrap(err, "MidTask:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "MidTask:rows.Err")
	}
	return
}

// BatchAddTask batch add task
func (d *dao) BatchAddTask(c context.Context, tasks []*handwrite.MidTaskDB) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, r := range tasks {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?)")
		rows = append(rows, r.Mid, r.TaskType, xstr.JoinInts(r.TaskDetailStruct), r.FinishCount, r.FinishTime)
	}
	sql := fmt.Sprintf(_midTaskAdd, handWriteTaskName, strings.Join(rowsTmp, ","))
	if _, err = d.db.Exec(c, sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddTask: d.db.Exec")
	}
	return
}

// GetAllMidTask get mid lit
func (d *dao) GetAllMidTask(ctx context.Context, offset, limit int64) (rs []*handwrite.MidTaskDB, err error) {
	rs = []*handwrite.MidTaskDB{}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_midTaskAllSQL, handWriteTaskName), offset, limit)
	if err != nil {
		err = errors.Wrap(err, "GetAllMidTask:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &handwrite.MidTaskDB{}
		err = rows.Scan(&r.Mid, &r.TaskDetail, &r.TaskType, &r.FinishCount, &r.FinishTime)
		if err != nil {
			err = errors.Wrap(err, "GetAllMidTask:rows.Scan error")
			return
		}

		if r.TaskDetail != "" {
			taskDetail, err := xstr.SplitInts(r.TaskDetail)
			if err != nil {
				log.Errorc(ctx, "task detail turn to ints error (%v)", err)
			} else {
				r.TaskDetailStruct = taskDetail
			}
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetAllMidTask:rows.Err")
	}
	return
}

package task

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"
)

const (
	_taskUserSQL         = "SELECT id,mid,business_id,task_id,foreign_id,round,cnt,finish,award,ctime,mtime,round_count FROM task_user_state_%02d WHERE mid = ? AND task_id IN (%s)"
	_taskUserStateAddSQL = "INSERT INTO task_user_state_%02d (mid,business_id,task_id,foreign_id,round,cnt,finish,award,round_count) VALUES(?,?,?,?,?,?,?,?,?)"
	_taskUserStateUpSQL  = "UPDATE task_user_state_%02d SET cnt = ?,finish = ?,round_count = ? WHERE mid = ? AND task_id = ? AND round = ?"
	_taskUserAwardSQL    = "UPDATE task_user_state_%02d SET award = ? WHERE mid = ? AND task_id = ? AND round = ?"
	_taskUserAddLogSQL   = "INSERT INTO task_user_log_%02d (mid,business_id,task_id,foreign_id,round) VALUES(?,?,?,?,?)"
)

// RawTaskUserState user task state.
func (d *Dao) RawTaskUserState(c context.Context, taskIDs []int64, mid, businessID, foreignID, nowTs int64) (list map[string]*taskmdl.UserTask, err error) {
	var (
		ids  []int64
		rows *xsql.Rows
	)
	rows, err = d.db.Query(c, fmt.Sprintf(_taskUserSQL, foreignID%100, xstr.JoinInts(taskIDs)), mid)
	if err != nil {
		log.Error("RawTaskUserState:d.db.Query(%v,%d,%d) error(%v)", ids, mid, foreignID, err)
		return
	}
	defer rows.Close()
	list = make(map[string]*taskmdl.UserTask)
	for rows.Next() {
		n := new(taskmdl.UserTask)
		if err = rows.Scan(&n.ID, &n.Mid, &n.BusinessID, &n.TaskID, &n.ForeignID, &n.Round, &n.Count, &n.Finish, &n.Award, &n.Ctime, &n.Mtime, &n.RoundCount); err != nil {
			log.Error("RawUserTaskState:rows.Scan() error(%v)", err)
			return
		}
		list[fmt.Sprintf("%d_%d", n.TaskID, n.Round)] = n
	}
	if err = rows.Err(); err != nil {
		log.Error("RawUserTaskState:rows.Err() error(%v)", err)
	}
	return
}

// TaskUserStateAdd add user task state.
func (d *Dao) TaskUserStateAdd(c context.Context, mid, businessID, taskID, foreignID, round, count, finish, award, roundCount int64) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_taskUserStateAddSQL, foreignID%100), mid, businessID, taskID, foreignID, round, count, finish, award, roundCount); err != nil {
		log.Error("TaskUserStateAdd error d.db.Exec(%d,%d,%d,%d,%d,%d,%d,%d,%d) error(%v)", mid, businessID, taskID, foreignID, round, count, finish, award, foreignID, err)
	}
	return
}

// TaskUserStateUp update user task state.
func (d *Dao) TaskUserStateUp(c context.Context, mid, taskID, round, count, finish, award, foreignID, roundCount int64) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_taskUserStateUpSQL, foreignID%100), count, finish, roundCount, mid, taskID, round); err != nil {
		log.Error("TaskUserStateUp error d.db.Exec(%d,%d,%d,%d,%d,%d,%d) error(%v)", mid, taskID, round, count, finish, roundCount, foreignID, err)
	}
	return
}

// TaskUserAward update award user task.
func (d *Dao) TaskUserAward(c context.Context, mid, taskID, round int64, num int64, foreignID int64) (err error) {
	var rly sql.Result
	if rly, err = d.db.Exec(c, fmt.Sprintf(_taskUserAwardSQL, foreignID%100), num, mid, taskID, round); err != nil {
		log.Error("UserTaskAward error d.db.Exec(%d,%d,%d,%d) error(%v)", mid, taskID, round, foreignID, err)
		return
	}
	count, _ := rly.RowsAffected()
	if count == 0 {
		err = ecode.ActivityTaskAwardFailed
	}
	return
}

// TaskUserLogAdd add user task log.
func (d *Dao) TaskUserLogAdd(c context.Context, mid, businessID, taskID, foreignID, round int64) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_taskUserAddLogSQL, foreignID%100), mid, businessID, taskID, foreignID, round); err != nil {
		log.Error("AddUserTaskLog error d.db.Exec(%d,%d,%d,%d,%d) error(%v)", mid, businessID, taskID, foreignID, round, err)
	}
	return
}

package task

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"

	"github.com/pkg/errors"
)

const (
	_taskIDsSQL          = "SELECT id FROM task WHERE business_id = ? AND foreign_id = ? AND state = 1 ORDER BY rank ASC,id ASC"
	_taskSQL             = "SELECT id,`name`,business_id,foreign_id,rank,finish_count,attribute,cycle_duration,stime,etime,award_type,award_id,award_count,award_expire FROM task WHERE id = ? AND state = 1"
	_tasksSQL            = "SELECT id,`name`,business_id,foreign_id,rank,finish_count,attribute,cycle_duration,stime,etime,award_type,award_id,award_count,award_expire FROM task WHERE id IN (%s) AND state = 1"
	_taskAddSQL          = "INSERT INTO task(`name`, `business_id`, `foreign_id`, `attribute`) VALUES(?, ?, ?, ?)"
	_userStateSQL        = "SELECT id,mid,business_id,task_id,foreign_id,round,count,finish,award,ctime FROM %s WHERE mid = ? AND task_id IN (%s)"
	_userTaskStateAddSQL = "INSERT INTO %s (mid,business_id,task_id,foreign_id,round,count,finish,award) VALUES(?,?,?,?,?,?,?,?)"
	_userTaskStateUpSQL  = "UPDATE %s SET count = ?,finish = ? WHERE mid = ? AND task_id = ? AND round = ?"
	_userTaskAwardSQL    = "UPDATE %s SET award = ? WHERE mid = ? AND task_id = ? AND round = ?"
	_userTaskLogSQL      = "SELECT id,mid,business_id,task_id,foreign_id,round,ctime FROM %s WHERE mid = ? AND task_id = ?"
	_userTaskAddSQL      = "INSERT INTO %s(mid,business_id,task_id,foreign_id,round) VALUES(?,?,?,?,?)"
	_taskRuleSQL         = "SELECT id,task_id,pre_task,level,ctime,mtime FROM task_rule WHERE task_id = ?"
)

func (d *Dao) userStateTableName(foreignID int64) string {
	if _, ok := d.c.Image.NewTask[strconv.FormatInt(foreignID, 10)]; ok {
		return fmt.Sprintf("task_user_state_%d", foreignID)
	}
	return "task_user_state"
}

func (d *Dao) userLogTableName(foreignID int64) string {
	if _, ok := d.c.Image.NewTask[strconv.FormatInt(foreignID, 10)]; ok {
		return fmt.Sprintf("task_user_log_%d", foreignID)
	}
	return "task_user_log"
}

// RawTaskIDs get raw task ids.
func (d *Dao) RawTaskIDs(c context.Context, businessID, foreignID int64) (ids []int64, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _taskIDsSQL, businessID, foreignID)
	if err != nil {
		log.Error("RawTaskIDs:d.db.Query(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			log.Error("RawTaskIDs:rows.Scan() error(%v)", err)
			return
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawTasks:rows.Err() error(%v)", err)
	}
	return
}

// RawTasks get tasks by task ids.
func (d *Dao) RawTasks(c context.Context, ids []int64) (list map[int64]*taskmdl.Task, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_tasksSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("RawTasks:d.db.Query(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	list = make(map[int64]*taskmdl.Task, len(ids))
	for rows.Next() {
		n := new(taskmdl.Task)
		if err = rows.Scan(&n.ID, &n.Name, &n.BusinessID, &n.ForeignID, &n.Rank, &n.FinishCount, &n.Attribute, &n.CycleDuration, &n.Stime, &n.Etime, &n.AwardType, &n.AwardID, &n.AwardCount, &n.AwardExpire); err != nil {
			log.Error("RawTasks:rows.Scan() error(%v)", err)
			return
		}
		list[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		log.Error("RawTasks:rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) AddTask(c context.Context, name string, businessID, foreignID, attribute int64) (t *taskmdl.Task, err error) {
	var r sql.Result
	if r, err = d.db.Exec(c, _taskAddSQL, name, businessID, foreignID, attribute); err != nil {
		log.Error("AddTask error d.db.Exec(%s,%d,%d,%d) error(%v)", name, businessID, foreignID, attribute, err)
		return
	}
	var id int64
	id, err = r.LastInsertId()
	t = &taskmdl.Task{
		ID:         id,
		Name:       name,
		BusinessID: businessID,
		ForeignID:  foreignID,
		Attribute:  attribute,
	}
	return
}

// RawTask get raw task .
func (d *Dao) RawTask(c context.Context, id int64) (t *taskmdl.Task, err error) {
	t = new(taskmdl.Task)
	row := d.db.QueryRow(c, _taskSQL, id)
	if err = row.Scan(&t.ID, &t.Name, &t.BusinessID, &t.ForeignID, &t.Rank, &t.FinishCount, &t.Attribute, &t.CycleDuration, &t.Stime, &t.Etime, &t.AwardType, &t.AwardID, &t.AwardCount, &t.AwardExpire); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawTask:QueryRow")
		}
	}
	return
}

// RawUserTaskState user task state.
func (d *Dao) RawUserTaskState(c context.Context, taskIDs []int64, mid, businessID, foreignID, nowTs int64) (list map[string]*taskmdl.UserTask, err error) {
	var (
		ids   []int64
		rows  *xsql.Rows
		tName = d.userStateTableName(foreignID)
	)
	rows, err = d.db.Query(c, fmt.Sprintf(_userStateSQL, tName, xstr.JoinInts(taskIDs)), mid)
	if err != nil {
		log.Error("RawUserTaskState:d.db.Query(%v,%d,%s) error(%v)", ids, mid, tName, err)
		return
	}
	defer rows.Close()
	list = make(map[string]*taskmdl.UserTask)
	for rows.Next() {
		n := new(taskmdl.UserTask)
		if err = rows.Scan(&n.ID, &n.Mid, &n.BusinessID, &n.TaskID, &n.ForeignID, &n.Round, &n.Count, &n.Finish, &n.Award, &n.Ctime); err != nil {
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

// RawUserTaskLog user task log data.
func (d *Dao) RawUserTaskLog(c context.Context, mid, taskID, foreignID int64) (list []*taskmdl.UserTaskLog, err error) {
	tName := d.userLogTableName(foreignID)
	var rows *xsql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_userTaskLogSQL, tName), mid, taskID)
	if err != nil {
		log.Error("RawUserTaskLog:d.db.Query mid(%d) taskID(%d) error(%v)", mid, taskID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(taskmdl.UserTaskLog)
		if err = rows.Scan(&n.ID, &n.Mid, &n.BusinessID, &n.TaskID, &n.ForeignID, &n.Round, &n.Ctime); err != nil {
			log.Error("RawUserTaskLog:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawUserTaskLog:rows.Err() error(%v)", err)
	}
	return
}

// RawUserRule get pretask by task_id
func (d *Dao) RawTaskRule(c context.Context, taskID int64) (res *taskmdl.TaskRule, err error) {
	rows := d.db.QueryRow(c, _taskRuleSQL, taskID)
	var n = &taskmdl.TaskRule{}
	if err = rows.Scan(&n.ID, &n.TaskID, &n.PreTask, &n.Level, &n.Ctime, &n.Mtime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawTaskRule:QueryRow")
		}
		return
	}
	res = n
	return
}

// AddUserTaskState add user task state.
func (d *Dao) AddUserTaskState(c context.Context, mid, businessID, taskID, foreignID, round, count, finish, award int64) (err error) {
	tName := d.userStateTableName(foreignID)
	if _, err = d.db.Exec(c, fmt.Sprintf(_userTaskStateAddSQL, tName), mid, businessID, taskID, foreignID, round, count, finish, award); err != nil {
		log.Error("AddUserTaskState error d.db.Exec(%d,%d,%d,%d,%d,%d,%d,%d,%s) error(%v)", mid, businessID, taskID, foreignID, round, count, finish, award, tName, err)
	}
	return
}

// UpUserTaskState update user task state.
func (d *Dao) UpUserTaskState(c context.Context, mid, taskID, round, count, finish, award, foreignID int64) (err error) {
	tName := d.userStateTableName(foreignID)
	if _, err = d.db.Exec(c, fmt.Sprintf(_userTaskStateUpSQL, tName), count, finish, mid, taskID, round); err != nil {
		log.Error("UpUserTaskState error d.db.Exec(%d,%d,%d,%d,%d,%s) error(%v)", mid, taskID, round, count, finish, tName, err)
	}
	return
}

// UserTaskAward update award user task.
func (d *Dao) UserTaskAward(c context.Context, mid, taskID, round int64, num int64, foreignID int64) (err error) {
	var rly sql.Result
	tName := d.userStateTableName(foreignID)
	if rly, err = d.db.Exec(c, fmt.Sprintf(_userTaskAwardSQL, tName), num, mid, taskID, round); err != nil {
		log.Error("UserTaskAward error d.db.Exec(%d,%d,%d,%s) error(%v)", mid, taskID, round, tName, err)
		return
	}
	count, _ := rly.RowsAffected()
	if count == 0 {
		err = ecode.ActivityTaskAwardFailed
	}
	return
}

// AddUserTaskLog add user task log.
func (d *Dao) AddUserTaskLog(c context.Context, mid, businessID, taskID, foreignID, round int64) (err error) {
	tName := d.userLogTableName(foreignID)
	if _, err = d.db.Exec(c, fmt.Sprintf(_userTaskAddSQL, tName), mid, businessID, taskID, foreignID, round); err != nil {
		log.Error("AddUserTaskLog error d.db.Exec(%d,%d,%d,%d,%d) error(%v)", mid, businessID, taskID, foreignID, round, err)
	}
	return
}

const _taskStatsSQL = "SELECT task_id,num FROM task_state_%02d WHERE foreign_id=? AND business_id=? AND task_id IN (%s)"

func (d *Dao) RawTaskStats(ctx context.Context, taskIDs []int64, foreignID, businessID int64) (map[int64]int64, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_taskStatsSQL, foreignID%100, xstr.JoinInts(taskIDs)), foreignID, businessID)
	if err != nil {
		return nil, errors.Wrap(err, "TaskState Query")
	}
	defer rows.Close()
	taskStat := make(map[int64]int64)
	for rows.Next() {
		var taskID, num int64
		if err = rows.Scan(&taskID, &num); err != nil {
			return nil, errors.Wrap(err, "TaskState Scan")
		}
		taskStat[taskID] = num
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "TaskState rows.Err")
	}
	res := make(map[int64]int64)
	for _, id := range taskIDs {
		if stat, ok := taskStat[id]; ok {
			res[id] = stat
		} else {
			res[id] = 0
		}
	}
	return res, nil
}

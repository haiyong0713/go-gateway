package like

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/job/component"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/task"

	"github.com/pkg/errors"
)

const (
	_doTaskURI        = "/x/internal/activity/task/do"
	taskTableName     = "task"
	taskRuleTableName = "task_rule"
	countKey          = "count"
	midRuleKey        = "midRule"
)

func taskStateKey(sid, businessID, taskID int64) string {
	return fmt.Sprintf("task_state_%d_%d_%d", sid, businessID, taskID)
}

// DoTask .
func (d *Dao) DoTask(c context.Context, taskID, mid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	if err = d.httpClient.Post(c, d.doTaskURL, "", params, &res); err != nil {
		log.Errorc(c, "DoTask:d.httpClient.Post mid(%d) taskID(%d) error(%v)", mid, taskID, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.doTaskURL+"?"+params.Encode())
	}
	return
}

const _taskStateDecrSQL = "UPDATE task_state_%02d SET num=num-1 WHERE foreign_id = ? AND business_id=? AND task_id=? AND num > 0"

// TaskStateDecr ...
func (d *Dao) TaskStateDecr(ctx context.Context, sid, businessID, taskID int64) (int64, error) {
	row, err := d.db.Exec(ctx, fmt.Sprintf(_taskStateDecrSQL, sid%100), sid, businessID, taskID)
	if err != nil {
		return 0, errors.Wrap(err, "TaskStateDecr Exec")
	}
	return row.RowsAffected()
}

const _taskStateAddSQL = "INSERT INTO task_state_%02d (foreign_id,business_id,task_id,num) VALUES (?,?,?,?)"

// TaskStateAdd ...
func (d *Dao) TaskStateAdd(ctx context.Context, sid, businessID, taskID, num int64) error {
	if _, err := d.db.Exec(ctx, fmt.Sprintf(_taskStateAddSQL, sid%100), sid, businessID, taskID, num); err != nil {
		return errors.Wrap(err, "TaskStateAdd Exec")
	}
	return nil
}

const _taskStateIncrSQL = "UPDATE task_state_%02d SET num=num+1 WHERE foreign_id = ? AND business_id=? AND task_id=?"

// TaskStateIncr ...
func (d *Dao) TaskStateIncr(ctx context.Context, sid, businessID, taskID int64) (int64, error) {
	row, err := d.db.Exec(ctx, fmt.Sprintf(_taskStateIncrSQL, sid%100), sid, businessID, taskID)
	if err != nil {
		return 0, errors.Wrap(err, "TaskStateIncr Exec")
	}
	return row.RowsAffected()
}

// UpTaskState ...
func (d *Dao) UpTaskState(ctx context.Context, sid, businessID, taskID int64) error {
	affected, err := d.TaskStateDecr(ctx, sid, businessID, taskID)
	if err != nil {
		return err
	}
	if affected > 0 {
		// update cache
		if err = d.TaskStateCacheDecr(ctx, sid, businessID, taskID); err != nil {
			return err
		}
	}
	return nil
}

// TaskStateCacheDecr ...
func (d *Dao) TaskStateCacheDecr(ctx context.Context, sid, businessID, taskID int64) error {
	conn := component.GlobalRedisStore.Conn(ctx)
	defer conn.Close()
	key := taskStateKey(sid, businessID, taskID)
	ok, err := redis.Bool(conn.Do("EXPIRE", key, d.taskStateExpire))
	if err != nil {
		return errors.Wrap(err, "TaskStateCacheDecr EXPIRE")
	}
	if ok {
		if _, err = conn.Do("DECR", key); err != nil {
			return errors.Wrap(err, "TaskStateCacheDecr DECR")
		}
	}
	return nil
}

// TaskStateCacheIncr ...
func (d *Dao) TaskStateCacheIncr(ctx context.Context, sid, businessID, taskID int64) error {
	conn := component.GlobalRedisStore.Conn(ctx)
	defer conn.Close()
	key := taskStateKey(sid, businessID, taskID)
	ok, err := redis.Bool(conn.Do("EXPIRE", key, d.taskStateExpire))
	if err != nil {
		return errors.Wrap(err, "TaskStateCacheIncr EXPIRE")
	}
	if ok {
		if _, err = conn.Do("INCR", key); err != nil {
			return errors.Wrap(err, "TaskStateCacheIncr INCR")
		}
	}
	return nil
}

const _taskByForeignIDSQL = "SELECT id,name,business_id,foreign_id,finish_count,attribute,cycle_duration,stime,etime,award_type,award_id,award_count FROM %s WHERE foreign_id = ? AND business_id = ? limit 1"

// GetTaskByForeignID 根据关联活动获取task
func (d *Dao) GetTaskByForeignID(ctx context.Context, foreignID, businessID int64) (r *task.Task, err error) {
	r = &task.Task{}
	row := d.db.QueryRow(ctx, fmt.Sprintf(_taskByForeignIDSQL, taskTableName), foreignID, businessID)
	if err = row.Scan(&r.ID, &r.Name, &r.BusinessID, &r.ForeignID, &r.FinishCount, &r.Attribute, &r.CycleDuration, &r.Stime, &r.Etime, &r.AwardCount, &r.AwardID, &r.AwardCount); err != nil {
		err = errors.Wrap(err, "GetTaskByForeignID:row.Scan error")
		return
	}
	return

}

const _taskByTaskIDSQL = "SELECT id,name,business_id,foreign_id,finish_count,attribute,cycle_duration,stime,etime,award_type,award_id,award_count FROM %s WHERE id = ? limit 1"

// GetTaskByTaskID 根据关联活动获取task
func (d *Dao) GetTaskByTaskID(ctx context.Context, taskID int64) (r *task.Task, err error) {
	r = &task.Task{}
	row := d.db.QueryRow(ctx, fmt.Sprintf(_taskByTaskIDSQL, taskTableName), taskID)
	if err = row.Scan(&r.ID, &r.Name, &r.BusinessID, &r.ForeignID, &r.FinishCount, &r.Attribute, &r.CycleDuration, &r.Stime, &r.Etime, &r.AwardCount, &r.AwardID, &r.AwardCount); err != nil {
		err = errors.Wrap(err, "GetTaskByTaskID:row.Scan error")
		return
	}
	return

}

const _childTaskByTaskIDSQL = "SELECT id,task_id,pre_task,object,count,count_type,level,ctime,mtime FROM %s WHERE task_id = ?"

// GetChildTask 获取子任务
func (d *Dao) GetChildTask(ctx context.Context, taskID int64) (rs []*task.Rule, err error) {
	rs = []*task.Rule{}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_childTaskByTaskIDSQL, taskRuleTableName), taskID)
	if err != nil {
		err = errors.Wrap(err, "GetChildTask:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &task.Rule{}
		err = rows.Scan(&r.ID, &r.TaskID, &r.PreTask, &r.Object, &r.Count, &r.CountType, &r.Level, &r.Ctime, &r.Mtime)
		if err != nil {
			err = errors.Wrap(err, "GetChildTask:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetChildTask:rows.Err")
	}
	return
}

// SetTaskCount 任务完成总人数统计
func (d *Dao) SetTaskCount(c context.Context, id int64, count int64) (err error) {
	var (
		bs   []byte
		key  = buildKey(countKey, id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs = []byte(strconv.FormatInt(count, 10))
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Errorc(c, "SetTaskCount conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// GetTaskCount 任务完成总人数获取
func (d *Dao) GetTaskCount(c context.Context, id int64) (count int64, err error) {
	var (
		key  = buildKey(countKey, id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if count, err = redis.Int64(conn.Do("GET", key)); err != nil {
		log.Errorc(c, "SetTaskCount conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// ActivityTaskMidStatus 活动用户任务状态
func (d *Dao) ActivityTaskMidStatus(c context.Context, id int64, midRule map[int64][]*task.MidRule) (err error) {
	if len(midRule) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	keys := make([]string, 0)
	for k, v := range midRule {
		var bs []byte
		if bs, err = json.Marshal(v); err != nil {
			log.Error("ActivityTaskMidStatus json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(buildKey(midRuleKey, id, k)).Add(bs)
		keys = append(keys, buildKey(midRuleKey, id, k))
	}
	if err = conn.Send("MSET", args...); err != nil {
		err = errors.Wrap(err, "ActivityTaskMidStatus conn.Do(MSET)")
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.dataExpire); err != nil {
			log.Errorc(c, "ActivityTaskMidStatus conn.Send(Expire, %s, %d) error(%v)", v, d.dataExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "ActivityTaskMidStatus Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(c, "ActivityTaskMidStatus conn.Receive() error(%v)", err)
			return err
		}
	}
	return
}

// GetActivityTaskMidStatus 获得用户任务完成情况
func (d *Dao) GetActivityTaskMidStatus(c context.Context, id int64, mids []int64) (res map[int64][]*task.MidRule, err error) {
	res = map[int64][]*task.MidRule{}
	if len(mids) == 0 {
		return
	}
	var (
		bss  [][]byte
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range mids {
		args = args.Add(buildKey(midRuleKey, id, v))
	}
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Errorc(c, "conn.Do(MGET,%s) error(%v)", args, err)
		return nil, err
	}
	for _, bs := range bss {
		rule := []*task.MidRule{}
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, &rule); err != nil {
			log.Errorc(c, "json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		if len(rule) > 0 {
			mid := rule[0].MID
			res[mid] = rule
		}
	}
	return
}

const insertTaskUserStateSQL = "INSERT INTO task_user_state_%02d (mid,business_id,task_id,foreign_id,round,cnt,finish,award) VALUES %s"

// TaskUserStateUp 用户获奖日志
func (d *Dao) TaskUserStateUp(c context.Context, foreignID int64, userState []*task.UserState) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, r := range userState {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?,?)")
		rows = append(rows, r.MID, r.BusinessID, r.TaskID, r.ForeignID, r.Round, r.Count, r.Finish, r.Award)
	}
	sql := fmt.Sprintf(insertTaskUserStateSQL, foreignID%100, strings.Join(rowsTmp, ","))
	if _, err = d.db.Exec(c, sql, rows...); err != nil {
		err = errors.Wrap(err, "TaskUserStateUp: d.db.Exec")
	}
	return
}

const _userstateTaskIDSQL = "SELECT id,mid,business_id,task_id,foreign_id,round,cnt,finish,award FROM task_user_state_%02d WHERE task_id = ? and mid in (%s) "

// GetUserTaskState 获取任务情况
func (d *Dao) GetUserTaskState(ctx context.Context, taskID int64, foreignID int64, mids []int64) (rs []*task.UserState, err error) {
	rs = []*task.UserState{}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_userstateTaskIDSQL, foreignID%100, xstr.JoinInts(mids)), taskID)
	if err != nil {
		err = errors.Wrap(err, "GetChildTask:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &task.UserState{}
		err = rows.Scan(&r.ID, &r.MID, &r.BusinessID, &r.TaskID, &r.ForeignID, &r.Round, &r.Count, &r.Finish, &r.Award)
		if err != nil {
			err = errors.Wrap(err, "GetUserTaskState:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetUserTaskState:rows.Err")
	}
	return
}

package bws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/pkg/errors"
)

const _taskSQL = "SELECT id,title,cate,finish_count,rule_ids,order_num,ctime,mtime FROM act_bws_task WHERE id=? AND state=1"

func (d *Dao) RawTask(ctx context.Context, taskID int64) (*bwsmdl.Task, error) {
	row := d.db.QueryRow(ctx, _taskSQL, taskID)
	data := new(bwsmdl.Task)
	var ruleIDStr string
	if err := row.Scan(&data.ID, &data.Title, &data.Cate, &data.FinishCount, &ruleIDStr, &data.OrderNum, &data.Ctime, &data.Mtime); err != nil {
		return nil, errors.Wrap(err, "RawTask:QueryRow")
	}
	if ruleIDs, err := xstr.SplitInts(ruleIDStr); err == nil {
		data.RuleIds = ruleIDs
	}
	return data, nil
}

const _taskListSQL = "SELECT id,title,cate,finish_count,rule_ids,order_num,ctime,mtime FROM act_bws_task WHERE state=1"

func (d *Dao) RawTaskList(ctx context.Context) (map[int64]*bwsmdl.Task, error) {
	rows, err := d.db.Query(ctx, _taskListSQL)
	if err != nil {
		return nil, errors.Wrap(err, "RawTaskList Query")
	}
	defer rows.Close()
	data := make(map[int64]*bwsmdl.Task)
	for rows.Next() {
		r := new(bwsmdl.Task)
		var (
			ruleIDStr string
			ruleIDs   []int64
		)
		if err = rows.Scan(&r.ID, &r.Title, &r.Cate, &r.FinishCount, &ruleIDStr, &r.OrderNum, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawTaskByIDs Scan")
		}
		if ruleIDs, err = xstr.SplitInts(ruleIDStr); err == nil {
			r.RuleIds = ruleIDs
		}
		data[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawTaskList rows")
	}
	return data, nil
}

const _userTasksSQL = "SELECT task_id,now_count,finish,award,task_day FROM act_bws_task_user_state WHERE user_token=? AND task_day=?"

func (d *Dao) RawUserTasks(ctx context.Context, userToken string, taskDay int64) ([]*bwsmdl.UserTask, error) {
	rows, err := d.db.Query(ctx, _userTasksSQL, userToken, taskDay)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserTasks Query")
	}
	defer rows.Close()
	var data []*bwsmdl.UserTask
	for rows.Next() {
		r := new(bwsmdl.UserTask)
		if err = rows.Scan(&r.TaskID, &r.NowCount, &r.UserState, &r.AwardState, &r.TaskDay); err != nil {
			return nil, errors.Wrap(err, "RawUserTasks Scan")
		}
		data = append(data, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserTasks rows")
	}
	return data, nil
}

func userTaskKey(userToken string, day int64) string {
	return fmt.Sprintf("bws20_user_task_%s_%d", userToken, day)
}

func (d *Dao) CacheUserTasks(ctx context.Context, userToken string, day int64) ([]*bwsmdl.UserTask, error) {
	key := userTaskKey(userToken, day)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bytes, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "CacheUserTasks:redis GET")
	}
	var data []*bwsmdl.UserTask
	if err = json.Unmarshal(bytes, &data); err != nil {
		return nil, errors.Wrap(err, "CacheUserTasks json.Unmarshal")
	}
	return data, nil
}

func (d *Dao) AddCacheUserTasks(ctx context.Context, userToken string, data []*bwsmdl.UserTask, day int64) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "AddCacheUserTasks json.Marshal")
	}
	key := userTaskKey(userToken, day)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, d.bwsUserExpire, bytes); err != nil {
		return errors.Wrapf(err, "AddCacheUserTasks SETEX key:%s", key)
	}
	return nil
}

func (d *Dao) DelCacheUserTasks(ctx context.Context, userToken string, day int64) error {
	key := userTaskKey(userToken, day)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelCacheUserTasks conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _allUserTasksSQL = "SELECT task_id,now_count,finish,award,task_day FROM act_bws_task_user_state WHERE user_token=?"

func (d *Dao) RawLastUserTasks(ctx context.Context, userToken string) ([]*bwsmdl.UserTask, int64, error) {
	rows, err := d.db.Query(ctx, _allUserTasksSQL, userToken)
	if err != nil {
		return nil, 0, errors.Wrap(err, "RawLastUserTasks Query")
	}
	defer rows.Close()
	data := make(map[int64][]*bwsmdl.UserTask)
	var maxTaskDay int64
	for rows.Next() {
		r := new(bwsmdl.UserTask)
		if err = rows.Scan(&r.TaskID, &r.NowCount, &r.UserState, &r.AwardState, &r.TaskDay); err != nil {
			return nil, 0, errors.Wrap(err, "RawLastUserTasks Scan")
		}
		data[r.TaskDay] = append(data[r.TaskDay], r)
		if r.TaskDay > maxTaskDay {
			maxTaskDay = r.TaskDay
		}
	}
	if err = rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "RawLastUserTasks rows")
	}
	return data[maxTaskDay], maxTaskDay, nil
}

const _addUserTaskSQL = "INSERT INTO act_bws_task_user_state(task_id,user_token,task_day,now_count,finish,award) VALUES %s"

func (d *Dao) AddUserTask(ctx context.Context, taskAdds []*bwsmdl.UserTask, day int64, userToken string) (int64, error) {
	var (
		rowPlaces []string
		args      []interface{}
	)
	for _, v := range taskAdds {
		rowPlaces = append(rowPlaces, "(?,?,?,?,?,?)")
		args = append(args, v.TaskID, userToken, day, v.NowCount, v.UserState, v.AwardState)
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_addUserTaskSQL, strings.Join(rowPlaces, ",")), args...)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserTask")
	}
	return row.RowsAffected()
}

const _upUserTaskSQL = "UPDATE act_bws_task_user_state SET now_count=now_count+?,finish=? WHERE user_token=? AND task_id=? AND task_day=?"

func (d *Dao) UpUserTask(ctx context.Context, userToken string, taskID, taskDay, count, finish int64) (int64, error) {
	row, err := d.db.Exec(ctx, _upUserTaskSQL, count, finish, userToken, taskID, taskDay)
	if err != nil {
		return 0, errors.Wrap(err, "UpUserTask")
	}
	return row.RowsAffected()
}

const _awardUserTaskSQL = "UPDATE act_bws_task_user_state SET award=1 WHERE user_token=? AND task_id=? AND task_day=? AND award=0"

func (d *Dao) AwardUserTask(ctx context.Context, userToken string, taskID, taskDay int64) (int64, error) {
	row, err := d.db.Exec(ctx, _awardUserTaskSQL, userToken, taskID, taskDay)
	if err != nil {
		return 0, errors.Wrap(err, "AwardUserTask")
	}
	return row.RowsAffected()
}

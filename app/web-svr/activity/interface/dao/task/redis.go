package task

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"

	"github.com/pkg/errors"
)

const (
	_userTaskStateFmt = "task_s_t_%d_%d_%d"
)

func taskRoundKey(taskID, round int64) string {
	return fmt.Sprintf("%d_%d", taskID, round)
}

func userTaskStateKey(mid, businessID, foreignID int64) string {
	return fmt.Sprintf(_userTaskStateFmt, mid, businessID, foreignID)
}

func taskStateKey(sid, businessID, taskID int64) string {
	return fmt.Sprintf("task_state_%d_%d_%d", sid, businessID, taskID)
}

// CacheUserTaskState get user task cache.
func (d *Dao) CacheUserTaskState(c context.Context, tasks map[int64]*taskmdl.Task, mid, businessID, foreignID int64) (list map[string]*taskmdl.UserTask, err error) {
	if len(tasks) == 0 {
		return
	}
	key := userTaskStateKey(mid, businessID, foreignID)
	args := redis.Args{}.Add(key)
	nowTs := time.Now().Unix()
	var taskList []string
	for _, v := range tasks {
		if v.IsCycle() {
			for i := v.Round(nowTs); i >= 0; i-- {
				args = args.Add(taskRoundKey(v.ID, i))
				taskList = append(taskList, taskRoundKey(v.ID, i))
			}
		} else {
			args = args.Add(taskRoundKey(v.ID, 0))
			taskList = append(taskList, taskRoundKey(v.ID, 0))
		}
	}
	var values [][]byte
	if values, err = redis.ByteSlices(component.GlobalRedis.Do(c, "HMGET", args...)); err != nil {
		if err != redis.ErrNil {
			log.Error("CacheUserTaskState redis.ByteSlices(%s,%v) error(%v)", key, tasks, err)
			return
		}
	}
	list = make(map[string]*taskmdl.UserTask)
	for index, v := range taskList {
		if values[index] == nil {
			continue
		}
		userTask := new(taskmdl.UserTask)
		if e := json.Unmarshal(values[index], userTask); e != nil {
			log.Warn("CacheUserTaskState json.Unmarshal(%x) error(%v)", values[index], err)
			continue
		}
		list[v] = userTask
	}
	return
}

// AddCacheUserTaskState add user task state cache.
func (d *Dao) AddCacheUserTaskState(c context.Context, missData map[string]*taskmdl.UserTask, mid, businessID, foreignID int64) (err error) {
	var (
		bs []byte
	)
	if len(missData) == 0 {
		return
	}
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	key := userTaskStateKey(mid, businessID, foreignID)
	args := redis.Args{}.Add(key)
	for k, v := range missData {
		if bs, err = json.Marshal(v); err != nil {
			log.Warn("AddCacheUserTaskState json.Marshal(%v) error(%v)", v, err)
			continue
		}
		args = args.Add(k).Add(bs)
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Error("AddCacheUserTaskState conn.Send(HMSET, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.userTaskExpire); err != nil {
		log.Error("AddCacheUserTaskState conn.Send(Expire, %s, %d) error(%v)", key, d.userTaskExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUserTaskState conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUserTaskState conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// SetCacheUserTaskState set one cache user task state
func (d *Dao) SetCacheUserTaskState(c context.Context, data *taskmdl.UserTask, mid, businessID, foreignID int64) (err error) {
	var (
		bs []byte
	)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	key := userTaskStateKey(mid, businessID, foreignID)
	args := redis.Args{}.Add(key)
	if bs, err = json.Marshal(data); err != nil {
		log.Warn("SetCacheUserTaskState json.Marshal(%v) error(%v)", data, err)
		return
	}
	args = args.Add(taskRoundKey(data.TaskID, data.Round)).Add(bs)
	if err = conn.Send("HSET", args...); err != nil {
		log.Error("SetCacheUserTaskState conn.Send(HSET, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.userTaskExpire); err != nil {
		log.Error("SetCacheUserTaskState conn.Send(Expire, %s, %d) error(%v)", key, d.userTaskExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetCacheUserTaskState conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetCacheUserTaskState conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) CacheTaskStats(ctx context.Context, taskIDs []int64, sid, businessID int64) (map[int64]int64, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	var (
		args = redis.Args{}
		ss   []int64
	)
	for _, taskID := range taskIDs {
		args = args.Add(taskStateKey(sid, businessID, taskID))
	}
	ss, err := redis.Int64s(component.GlobalRedisStore.Do(ctx, "MGET", args...))
	if err != nil {
		return nil, errors.Wrap(err, "CacheTaskStats MGET")
	}
	res := make(map[int64]int64, len(taskIDs))
	for key, val := range ss {
		if val == 0 {
			continue
		}
		res[taskIDs[key]] = val
	}
	return res, nil
}

func (d *Dao) AddCacheTaskStats(ctx context.Context, taskStates map[int64]int64, sid int64, businessID int64) error {
	if len(taskStates) == 0 {
		return nil
	}
	conn := component.GlobalRedisStore.Conn(ctx)
	defer conn.Close()
	var reserveKey []string
	args := redis.Args{}
	for k, v := range taskStates {
		keyStr := taskStateKey(sid, businessID, k)
		args = args.Add(keyStr).Add(v)
		reserveKey = append(reserveKey, keyStr)
	}
	var count int
	if err := conn.Send("MSET", args...); err != nil {
		return errors.Wrap(err, "AddCacheTaskStats MSET")
	}
	count++
	for _, v := range reserveKey {
		if err := conn.Send("EXPIRE", v, d.taskStateExpire); err != nil {
			return errors.Wrap(err, "AddCacheTaskStats EXPIRE")
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "AddCacheTaskStats EXPIRE")
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			return errors.Wrap(err, "AddCacheTaskStats Receive")
		}
	}
	return nil
}

// GetActivityTaskMidStatus 获得用户任务完成情况
func (d *Dao) GetActivityTaskMidStatus(c context.Context, id int64, mid int64) (res []*taskmdl.MidRule, err error) {

	var (
		bs   []byte
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	res = []*taskmdl.MidRule{}
	if bs, err = redis.Bytes(conn.Do("GET", buildKey(midRuleKey, id, mid))); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", buildKey(midRuleKey, mid), err)
		}
		return
	}
	if bs != nil {
		if err = json.Unmarshal(bs, &res); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
		}
	}
	return
}

// GetActityTaskCount 任务完成总人数统计
func (d *Dao) GetActityTaskCount(c context.Context, id int64) (res int64, err error) {

	var (
		bs   []byte
		key  = buildKey(countKey, id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

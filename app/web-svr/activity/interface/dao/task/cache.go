package task

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/interface/model/task"
)

const (
	_taskFmt        = "task_%d"
	_taskIdsFmt     = "task_ids_%d_%d"
	_taskRuleFmt    = "task_rule_%d"
	_userTaskFinFmt = "user_task_fin_%d_%d"
)

func taskKey(id int64) string {
	return fmt.Sprintf(_taskFmt, id)
}

func taskIDsKey(businessID, foreignID int64) string {
	return fmt.Sprintf(_taskIdsFmt, businessID, foreignID)
}

func taskRuleKey(taskID int64) string {
	return fmt.Sprintf(_taskRuleFmt, taskID)
}

func userTaskFinKey(mid, taskID int64) string {
	return fmt.Sprintf(_userTaskFinFmt, mid, taskID)
}

//go:generate kratos tool btsgen
type _bts interface {
	// get a foreign business task id data.
	// bts:-struct_name=Dao
	TaskIDs(c context.Context, businessID int64, foreignID int64) ([]int64, error)
	// get task data by id.
	// bts:-struct_name=Dao
	Task(c context.Context, id int64) (*task.Task, error)
	// get tasks data by ids.
	// bts:-struct_name=Dao
	Tasks(c context.Context, ids []int64) (map[int64]*task.Task, error)
	// bts:-struct_name=Dao
	TaskStats(ctx context.Context, taskIDs []int64, sid int64, businessID int64) (map[int64]int64, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=taskKey -struct_name=Dao
	CacheTask(c context.Context, id int64) (*task.Task, error)
	// mc: -key=taskKey -expire=d.mcTaskExpire -encode=pb -struct_name=Dao
	AddCacheTask(c context.Context, id int64, data *task.Task) error
	// mc: -key=taskKey -struct_name=Dao
	DelCacheTask(c context.Context, id int64) error
	// mc: -key=taskKey -struct_name=Dao
	CacheTasks(c context.Context, ids []int64) (map[int64]*task.Task, error)
	// mc: -key=taskKey -expire=d.mcTaskExpire -encode=pb -struct_name=Dao
	AddCacheTasks(c context.Context, data map[int64]*task.Task) error
	// mc: -key=taskIDsKey -struct_name=Dao
	CacheTaskIDs(c context.Context, businessID int64, foreignID int64) ([]int64, error)
	// mc: -key=taskIDsKey -expire=d.mcTaskExpire -struct_name=Dao
	AddCacheTaskIDs(c context.Context, businessID int64, taskIDs []int64, foreignID int64) error
	// mc: -key=taskIDsKey -struct_name=Dao
	DelCacheTaskIDs(c context.Context, businessID int64, foreignID int64) error
	// mc: -key=taskRuleKey -struct_name=Dao
	CacheTaskRule(c context.Context, taskID int64) (res *task.TaskRule, err error)
	// mc: -key=taskRuleKey -expire=d.mcTaskExpire -struct_name=Dao
	AddCacheTaskRule(c context.Context, taskID int64, val *task.TaskRule) error
}

// UserTaskState get user task state by task ids.
func (d *Dao) UserTaskState(c context.Context, tasks map[int64]*task.Task, mid int64, businessID int64, foreignID, nowTs int64) (res map[string]*task.UserTask, err error) {
	if len(tasks) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheUserTaskState(c, tasks, mid, businessID, foreignID); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	//miss := make(map[int64]*task.Task)
	var miss []int64
	for _, key := range tasks {
		if key.IsCycle() {
			for i := key.Round(nowTs); i >= 0; i-- {
				if (res == nil) || (res[fmt.Sprintf("%d_%d", key.ID, i)] == nil) {
					miss = append(miss, key.ID)
				}
			}
		} else {
			if (res == nil) || (res[fmt.Sprintf("%d_%d", key.ID, 0)] == nil) {
				miss = append(miss, key.ID)
			}
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var (
		missData   map[string]*task.UserTask
		isNewTable bool
	)
	for _, v := range tasks {
		if v != nil && v.IsNewTable() {
			isNewTable = true
		}
	}
	if isNewTable {
		missData, err = d.RawTaskUserState(c, miss, mid, businessID, foreignID, nowTs)
	} else {
		missData, err = d.RawUserTaskState(c, miss, mid, businessID, foreignID, nowTs)
	}
	if res == nil {
		res = make(map[string]*task.UserTask, len(tasks))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheUserTaskState(c, missData, mid, businessID, foreignID)
	})
	return
}

// TaskRule get data from cache if miss will call source method, then add to cache.
func (d *Dao) TaskRule(c context.Context, taskID int64) (res *task.TaskRule, err error) {
	addCache := true
	res, err = d.CacheTaskRule(c, taskID)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		return
	}
	res, err = d.RawTaskRule(c, taskID)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheTaskRule(c, taskID, miss)
	})
	return
}

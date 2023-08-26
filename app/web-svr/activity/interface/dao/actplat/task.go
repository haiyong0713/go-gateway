package actplat

import (
	"context"
	"go-common/library/log"
	task "go-gateway/app/web-svr/activity/interface/model/task"
)

// GetTaskDetail ...
func (d *Dao) GetTaskDetail(c context.Context, activity string) (res []*task.Detail, err error) {
	taskDetail, err := d.GetTaskCache(c, activity)
	if err != nil {
		log.Errorc(c, "d.MidCardDetail err(%v)", err)
	}
	if taskDetail != nil && err == nil {
		return taskDetail, nil
	}
	taskDetailDb, err := d.RawTaskList(c, activity)
	if err != nil {
		log.Errorc(c, "d.MidNums(c, %s) err(%v)", activity, err)
		return nil, err
	}

	err = d.SetTaskCache(c, activity, taskDetailDb)
	if err != nil {
		log.Errorc(c, " d.SetTaskCache err(%v)", err)
	}
	return taskDetailDb, nil
}

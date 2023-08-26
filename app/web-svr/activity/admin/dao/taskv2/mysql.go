package taskv2

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
)

const (
	taskName = "act_task"
)

const (
	updateTask = "INSERT INTO %s (`id`,`task_name`,`order_id`,`activity`,`counter`,`task_desc`,`link`,`finish_times`,`state`,`link_name`,`activity_id`,`risk_level`,`risk_operation`,`is_fe`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE task_name=VALUES(task_name),order_id=VALUES(order_id),activity=VALUES(activity),counter=VALUES(counter),task_desc=VALUES(task_desc),link=VALUES(link),finish_times=VALUES(finish_times),link_name=VALUES(link_name),activity_id=VALUES(activity_id),state=VALUES(state),risk_level=VALUES(risk_level),risk_operation=VALUES(risk_operation),is_fe=VALUES(is_fe)"
)

// TaskInsertOrUpdate ...
func (d *Dao) TaskInsertOrUpdate(c context.Context, task *model.ActTask) (err error) {
	if task != nil {
		if _, err = d.db.Exec(c, fmt.Sprintf(updateTask, taskName), task.ID, task.TaskName, task.OrderID, task.Activity,
			task.Counter, task.TaskDesc,
			task.Link, task.FinishTimes, task.State, task.LinkName, task.ActivityID, task.RiskLevel, task.RiskOperation, task.IsFe); err != nil {
			log.Errorc(c, "task@TaskInsertOrUpdate d.db.Exec() failed. error(%v)", err)
		}
	}
	return
}

package actplat

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	task "go-gateway/app/web-svr/activity/interface/model/task"

	"github.com/pkg/errors"
)

const (
	_getAllTaskSQL = "SELECT id,activity_id,task_name,link_name,order_id,activity,counter,task_desc,link,finish_times,state,risk_level,risk_operation,is_fe FROM act_task WHERE activity = ? and state = 1 order by order_id asc "
)

// RawTaskList ...
func (d *Dao) RawTaskList(c context.Context, activity string) (res []*task.Detail, err error) {
	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, _getAllTaskSQL, activity); err != nil {
		err = errors.Wrap(err, "RawTaskList:dao.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*task.Detail, 0)
	for rows.Next() {
		l := &task.Detail{}
		if err = rows.Scan(&l.ID, &l.ActivityID, &l.TaskName, &l.LinkName, &l.OrderID, &l.Activity, &l.Counter, &l.Desc, &l.Link, &l.FinishTimes, &l.State, &l.RiskLevel, &l.RiskOperation, &l.IsFe); err != nil {
			err = errors.Wrap(err, "RawTaskList:rows.Scan()")
			return
		}
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawTaskList:rows.Err()")
	}
	return
}

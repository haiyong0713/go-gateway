package task

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/task"

	"github.com/pkg/errors"
)

const (
	_saveTaskSQL     = "UPDATE task SET `name` = ?,business_id = ?,foreign_id = ?,rank = ?,finish_count = ?,attribute = ?,cycle_duration = ?,award_id = ?,award_count = ?,state = ?,stime = ?,etime = ? WHERE id = ?"
	_saveTaskRuleSQL = "UPDATE task_rule SET pre_task = ?,`level` = ? WHERE id = ?"
	_addAwardURI     = "/x/internal/activity/task/add/award"
)

// SaveTask save task data.
func (d *Dao) SaveTask(c context.Context, arg *task.SaveArg, preData *task.Item) (err error) {
	tx := d.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("SaveTask d.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&task.Task{}).Exec(_saveTaskSQL, arg.Name, arg.BusinessID, arg.ForeignID, arg.Rank, arg.FinishCount, arg.Attribute, arg.CycleDuration, arg.AwardID, arg.AwardCount, arg.State, arg.Stime, arg.Etime, arg.ID).Error; err != nil {
		log.Error("SaveTask task Update(%+v) error(%v)", arg, err)
		err = tx.Rollback().Error
		return
	}
	taskRule := &task.Rule{
		TaskID:  arg.ID,
		PreTask: arg.PreTask,
		Level:   arg.Level,
	}
	if preData.Rule != nil {
		if arg.PreTask != preData.Rule.PreTask || arg.Level != preData.Rule.Level {
			// update
			taskRule.ID = preData.Rule.ID
			if err = tx.Model(&task.Rule{}).Exec(_saveTaskRuleSQL, arg.PreTask, arg.Level, arg.ID).Error; err != nil {
				log.Error("SaveTask rule Exec(%+v) error(%v)", taskRule, err)
				err = tx.Rollback().Error
				return
			}
		}
	} else {
		if arg.PreTask != "" || arg.Level != 0 {
			// insert
			if err = tx.Model(&task.Rule{}).Create(taskRule).Error; err != nil {
				log.Error("SaveTask rule Create(%+v) error(%v)", taskRule, err)
				err = tx.Rollback().Error
				return
			}
		}
	}
	err = tx.Commit().Error
	return
}

// AddAward .
func (d *Dao) AddAward(c context.Context, taskID, mid, award int64) (err error) {
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("award", strconv.FormatInt(award, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Post(c, d.addAwardURL, "", params, &res); err != nil {
		err = errors.Wrapf(err, "AddAward d.client.Post(%s)", d.addAwardURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

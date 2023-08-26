package question

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/question"
)

const (
	_saveBaseSQL           = "UPDATE question_base SET business_id = ?,foreign_id = ?,count = ?,one_ts = ?,retry_ts = ?,stime = ?,etime = ?,name = ?,`separator` = ?,`distribute_type` = ?  WHERE id = ?"
	_saveDetailSQL         = "UPDATE question_detail SET base_id = ?,`name` = ?,right_answer = ?,wrong_answer = ?,state = ?,pic = ? , attribute = ? WHERE id = ?"
	_batchDetailSQL        = "INSERT INTO question_detail(base_id,`name`,right_answer,wrong_answer,attribute,state,pic) VALUES %s"
	_userLogCreateSQL      = "CREATE TABLE IF NOT EXISTS question_user_log_%d LIKE question_user_log"
	_updateSingleDetailSQL = "UPDATE question_detail SET  state = ?  WHERE id = ? and  state != 2"
)

// SaveBase save question base data.
func (d *Dao) SaveBase(c context.Context, arg *question.SaveBaseArg) (err error) {
	if err = d.DB.Model(&question.Base{}).Exec(_saveBaseSQL, arg.BusinessID, arg.ForeignID, arg.Count, arg.OneTs, arg.RetryTs, arg.Stime, arg.Etime, arg.Name, arg.Separator, arg.DistributeType, arg.ID).Error; err != nil {
		log.Error("SaveDetail task Update(%+v) error(%v)", arg, err)
	}
	return
}

// SaveDetail save question detail data.
func (d *Dao) SaveDetail(c context.Context, arg *question.SaveDetailArg) (err error) {
	if err = d.DB.Model(&question.Detail{}).Exec(_saveDetailSQL, arg.BaseID, arg.Name, arg.RightAnswer, arg.WrongAnswer, arg.State, arg.Pic, arg.Attribute, arg.ID).Error; err != nil {
		log.Error("SaveDetail Update(%+v) error(%v)", arg, err)
	}
	return
}

func (d *Dao) UpdateDetailState(c context.Context, id int64, state int) (err error) {
	if err = d.DB.Model(&question.Detail{}).Exec(_updateSingleDetailSQL, state, id).Error; err != nil {
		log.Errorc(c, "DelDetail args(%v) error(%+v)", id, err)
	}
	return
}

// BatchAddDetail batch add detail.
func (d *Dao) BatchAddDetail(c context.Context, list []*question.AddDetailArg) (err error) {
	rowStr := make([]string, 0, len(list))
	args := make([]interface{}, 0)
	for _, v := range list {
		rowStr = append(rowStr, "(?,?,?,?,?,?,?)")
		args = append(args, v.BaseID, v.Name, v.RightAnswer, v.WrongAnswer, v.Attribute, v.State, v.Pic)
	}
	sql := fmt.Sprintf(_batchDetailSQL, strings.Join(rowStr, ","))
	if err = d.DB.Model(&question.Detail{}).Exec(sql, args...).Error; err != nil {
		log.Errorc(c, "BatchAddDetail Exec(%s) error(%v)", sql, err)
	}
	return
}

// UserLogCreate create user log table.
func (d *Dao) UserLogCreate(c context.Context, id int64) (err error) {
	if err = d.DB.Exec(fmt.Sprintf(_userLogCreateSQL, id)).Error; err != nil {
		log.Error("UserLogCreate id(%d) error(%v)", id, err)
	}
	return
}

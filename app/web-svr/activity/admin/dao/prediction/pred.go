package prediction

import (
	"context"
	"fmt"
	"strings"

	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"

	"github.com/pkg/errors"
)

const (
	_insertPreSQL = "INSERT INTO prediction(`sid`,`min`,`max`,`pid`,`name`,`type`,`state`) VALUES %s"
	_upPreSQL     = "UPDATE prediction SET `min` = ?,`max` = ? ,`name` = ?,`type` = ?,`state` = ? where `id` = ?"
)

// BatchAdd .
func (d *Dao) BatchAdd(c context.Context, list []*premdl.Prediction) (err error) {
	rowStr := make([]string, 0, len(list))
	args := make([]interface{}, 0)
	for _, v := range list {
		rowStr = append(rowStr, "(?,?,?,?,?,?,?)")
		args = append(args, v.Sid, v.Min, v.Max, v.Pid, v.Name, v.Type, v.State)
	}
	tx := d.DB
	if err = tx.Model(&premdl.Prediction{}).Exec(fmt.Sprintf(_insertPreSQL, strings.Join(rowStr, ",")), args...).Error; err != nil {
		err = errors.Wrap(err, "premdl.Prediction")
	}
	return
}

// Search .
func (d *Dao) Search(c context.Context, ser *premdl.PredSearch) (SearchRes *premdl.SearchRes, err error) {
	var (
		count int64
		list  []*premdl.Prediction
	)
	tx := d.DB
	if ser.ID > 0 {
		tx = tx.Where("id = ?", ser.ID)
	}
	if ser.Sid > 0 {
		tx = tx.Where("sid = ?", ser.Sid)
	}
	if ser.Pid != -1 {
		tx = tx.Where("pid = ?", ser.Pid)
	}
	if ser.Type != -1 {
		tx = tx.Where("type = ?", ser.Type)
	}
	if ser.State != -1 {
		tx = tx.Where("state = ?", ser.State)
	}
	pn := (ser.Pn - 1) * ser.Ps
	if err = tx.Order("id desc").Limit(ser.Ps).Offset(pn).Find(&list).Error; err != nil {
		err = errors.Wrap(err, "find err")
		return
	}
	if err = tx.Model(&premdl.Prediction{}).Count(&count).Error; err != nil {
		err = errors.Wrapf(err, "count error")
		return
	}
	SearchRes = &premdl.SearchRes{
		List: list,
		Pages: premdl.Pages{
			Total: count,
			Size:  ser.Ps,
			Num:   ser.Pn,
		},
	}
	return
}

// PresUp sid,pid can not modify.
func (d *Dao) PresUp(c context.Context, up *premdl.PresUp) (err error) {
	tx := d.DB
	if err = tx.Model(&premdl.Prediction{}).Exec(_upPreSQL, up.Min, up.Max, up.Name, up.Type, up.State, up.ID).Error; err != nil {
		err = errors.Wrapf(err, "%s", _upPreSQL)
	}
	return
}

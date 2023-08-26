package dao

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/up"

	"github.com/pkg/errors"
)

const (
	_tableUpAct      = "act_up"
	_createUserState = "CREATE TABLE IF NOT EXISTS act_up_user_state_%d LIKE act_up_user_state"
)

func (d *Dao) SearchUpList(c context.Context, uid, state, pn, ps int64) (res *up.ListReply, err error) {
	db := d.DB.Model(&up.UpAct{}).Where("state = ?", state)
	if uid > 0 {
		db = db.Where("mid like '%" + strconv.FormatInt(uid, 10) + "%'")
	}
	count := int64(0)
	if err = db.Count(&count).Error; err != nil {
		err = errors.Wrap(err, "db.Count()")
		return
	}
	if count == 0 {
		return
	}
	offset := (pn - 1) * ps
	list := make([]*up.UpAct, 0)
	if err = db.Offset(offset).Limit(ps).Order("id desc").Find(&list).Error; err != nil {
		err = errors.Wrap(err, "db.Offset.Find(list)")
		return
	}
	res = &up.ListReply{List: list, Count: count}
	return
}

func (d *Dao) UpActEdit(c context.Context, id, state, suffix int64) (err error) {
	db := d.DB
	if err = db.Table(_tableUpAct).Where("id=?", id).Update(map[string]interface{}{"state": state, "suffix": suffix, "finish_count": 1}).Error; err != nil {
		err = errors.Wrapf(err, "UpActEdit db.Update(%d,%d,%d)", id, state, suffix)
		return
	}
	if err = db.Exec(fmt.Sprintf(_createUserState, suffix)).Error; err != nil {
		err = errors.Wrapf(err, "UpActEdit db.Exec(%d) create", suffix)
	}
	return
}

func (d *Dao) UpActOffline(c context.Context, id, state int64) (err error) {
	if err = d.DB.Table(_tableUpAct).Where("id=?", id).Update(map[string]interface{}{"offline": state}).Error; err != nil {
		log.Error("ModifyPage d.DB.Table(%d,%d) error(%v)", id, state, err)
	}
	return
}

package dao

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/up_reserve"
	"strconv"
	"strings"
)

const upReserveTable = "up_act_reserve_relation"
const upReserveHangTable = "up_act_reserve_hang"
const upReserveHangLogTable = "up_act_reserve_hang_log"

// GetUpReserve
func (d *Dao) GetUpReserve(c context.Context, sid, mid int64, pn, ps int) (rly *up_reserve.UpReserveListReply, err error) {
	var (
		count int64
		list  []*up_reserve.UpReserveList
	)
	rly = &up_reserve.UpReserveListReply{}
	db := d.DB.Table(upReserveTable)
	if sid > 0 {
		db = db.Where("sid = ?", sid)
	}
	if mid > 0 {
		db = db.Where("mid = ?", mid)
	}
	if err = db.Count(&count).Error; err != nil {
		log.Errorc(c, "GetUpReserve count error(%v)", err)
		return
	}
	if count == 0 {
		return
	}
	offset := (pn - 1) * ps
	err = db.Offset(offset).Limit(ps).Order("id desc").Find(&list).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Errorc(c, "GetUpReserve list error(%v)", err)
			return
		}
	}
	rly.Total = count
	rly.List = list
	rly.Size = ps
	rly.Num = pn
	return
}

func (d *Dao) BatchInsertUpReserveHangAndLog(ctx context.Context, sid int64, mids []string, operator string) (err error) {
	tx := d.DB.Begin()

	defer func() {
		if err != nil {
			if err = tx.Rollback().Error; err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
		}
	}()

	var buffer bytes.Buffer
	sql := "insert into " + upReserveHangTable + " (`pub_mid`,`sid`) values"
	if _, err = buffer.WriteString(sql); err != nil {
		err = errors.Wrap(err, "buffer.WriteString(sql) err")
		return
	}

	for i, v := range mids {
		var mid int64
		if mid, err = strconv.ParseInt(v, 10, 64); err != nil {
			err = errors.Wrapf(err, "strconv.ParseInt(v, 10, 64) err v(%+v)", v)
			return
		}
		if mid == 0 {
			err = fmt.Errorf("reserve_id must not equal 0")
			return
		}
		if i == len(mids)-1 {
			buffer.WriteString(fmt.Sprintf("(%v,%v);", mid, sid))
		} else {
			buffer.WriteString(fmt.Sprintf("(%v,%v),", mid, sid))
		}
	}

	if err = tx.Exec(buffer.String()).Error; err != nil {
		err = errors.Wrap(err, "tx.Exec(buffer.String()) err")
		return
	}

	if err = tx.Create(&up_reserve.CreateHangLog{
		Operator: operator,
		Type:     1,
		Detail:   fmt.Sprintf("sid:%d, mid:%s", sid, strings.Join(mids, ",")),
		Result:   "成功",
		Remark:   "",
		Sid:      sid,
	}).Error; err != nil {
		err = errors.Wrap(err, "tx.Create err")
		return
	}

	if err = tx.Commit().Error; err != nil {
		err = errors.Wrap(err, "tx.Commit() err")
		return
	}

	return
}

func (d *Dao) GetUpReserveHangLogList(ctx context.Context, sid, pn, ps int64) (res *up_reserve.HangLogListReply, err error) {
	res = &up_reserve.HangLogListReply{
		List: make([]*up_reserve.HangLogItem, 0),
		Pager: &up_reserve.Pager{
			Num: pn, Size: ps,
		},
	}
	db := d.DB.Table(upReserveHangLogTable)
	if err = db.Count(&res.Pager.Total).Error; err != nil {
		err = errors.Wrap(err, "db.Count(&count) err")
		return
	}
	if res.Pager.Total == 0 {
		return
	}
	offset := (pn - 1) * ps
	err = db.Where("sid = ?", sid).Offset(offset).Limit(ps).Order("id desc").Find(&res.List).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		err = errors.Wrap(err, "db.Offset(offset).Limit(ps) err")
		return
	}
	return
}

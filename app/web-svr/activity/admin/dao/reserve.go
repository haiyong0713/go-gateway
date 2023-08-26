package dao

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/activity/admin/model/reserve"
)

const (
	_reserveTable    = "act_reserve_%02d"
	_reserveTableNum = 100
)

// reserveName .[00,99]
func reserveName(sid int64) string {
	return fmt.Sprintf(_reserveTable, sid%_reserveTableNum)
}

func (d *Dao) BwsReserveGift(c context.Context, mid int64) (err error) {
	var (
		param = url.Values{}
		res   struct {
			Code int `json:"code"`
		}
	)
	param.Set("mid", strconv.FormatInt(mid, 10))
	if err = d.client.Get(c, d.actBwsReserveGiftURL, "", param, &res); err != nil {
		log.Error("AddReserve d.client.Post %s error(%v)", d.actReserveURL+"?"+param.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
		log.Error("[AddReserve] error(%v)", err)
	}
	return
}

// AddReserve .
func (d *Dao) AddReserve(c context.Context, sid, mid int64, num int) (err error) {
	var (
		param = url.Values{}
		res   struct {
			Code int `json:"code"`
		}
	)
	param.Set("sid", strconv.FormatInt(sid, 10))
	param.Set("mid", strconv.FormatInt(mid, 10))
	param.Set("num", strconv.Itoa(num))
	if err = d.client.Post(c, d.actReserveURL, "", param, &res); err != nil {
		log.Error("AddReserve d.client.Post %s error(%v)", d.actReserveURL+"?"+param.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
		log.Error("[AddReserve] error(%v)", err)
	}
	return
}

// SearchReserve .
func (d *Dao) SearchReserve(c context.Context, sid, mid int64, pn, ps int) (rly *reserve.ListReply, err error) {
	var (
		count int64
		list  []*reserve.ActReserve
	)
	rly = &reserve.ListReply{}
	db := d.DB.Table(reserveName(sid)).Where("sid = ?", sid)
	if mid > 0 {
		db = db.Where("mid = ?", mid)
	}
	if err = db.Count(&count).Error; err != nil {
		log.Error("SearchList count error(%v)", err)
		return
	}
	if count == 0 {
		return
	}
	offset := (pn - 1) * ps
	if err = db.Offset(offset).Limit(ps).Order("id desc").Find(&list).Error; err != nil {
		log.Error("SearchList list error(%v)", err)
		return
	}
	rly.Count = count
	rly.List = list
	return
}

// GetReserve ...
func (d *Dao) GetReserve(c context.Context, sid, mid int64) (rly *reserve.ActReserve, err error) {
	rly = new(reserve.ActReserve)
	if err = d.DB.Table(reserveName(sid)).Where("sid = ?", sid).Where("mid = ?", mid).Last(&rly).Error; err != nil {
		return
	}
	return rly, nil
}

// UpdateReserve ...
func (d *Dao) UpdateReserve(c context.Context, r *reserve.ActReserve) (err error) {
	return d.DB.Table(reserveName(r.Sid)).Update(&r).Error
}

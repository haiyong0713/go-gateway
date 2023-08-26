package fawkes

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	fkappmdl "go-gateway/app/app-svr/fawkes/service/model/app"

	"github.com/pkg/errors"
)

const (
	_opt = "1007"

	_laserAll           = "/x/admin/fawkes/business/laser/all"
	_laserReport        = "/x/admin/fawkes/business/laser/report"
	_laserReportSilence = "/x/admin/fawkes/business/laser/report/silence"
	_broadcastPushAll   = "/x/internal/broadcast/push/all"
	_laserAllSilence    = "/x/admin/fawkes/business/laser/all/silence"
)

// LaserReportBroadCast report laser broadcaset status and url.
func (d *Dao) LaserReportBroadCast(c context.Context, taskID int64, status int, logURL string) (err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	params.Set("status", strconv.Itoa(status))
	params.Set("url", logURL)
	var re struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = d.client.Post(c, d.laserReport, ip, params, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserReport)
		log.Error("%v", err)
	}
	err = nil
	return
}

// LaserAll get all laser.
func (d *Dao) LaserAll(c context.Context) (res []*fkappmdl.Laser, err error) {
	var re struct {
		Code int               `json:"code"`
		Data []*fkappmdl.Laser `json:"data"`
	}
	if err = d.client.Get(c, d.laserAll, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserAll)
		return
	}
	for _, v := range re.Data {
		if v == nil {
			err = errors.New("list struct is nil")
			return
		}
	}
	res = re.Data
	return
}

// PushAll broadcast push all.
func (d *Dao) PushAll(c context.Context, msg, filter string) (err error) {
	params := url.Values{}
	params.Set("operation", _opt)
	params.Set("message", msg)
	params.Set("expired", strconv.FormatInt(70*3600+time.Now().Unix(), 10))
	params.Set("ack_kind", "2")
	if filter != "" {
		params.Set("filter", filter)
	}
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Post(c, d.broadcastPushAll, "", params, &res); err != nil {
		log.Error("PushAll url(%s) error(%v)", d.broadcastPushAll+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return
}

// LaserReport report laser push status and url.
func (d *Dao) LaserReportSilence(c context.Context, taskID int64, status int, logURL string) (err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	params.Set("status", strconv.Itoa(status))
	params.Set("url", logURL)
	var re struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = d.client.Post(c, d.laserReportSilence, ip, params, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserReportSilence)
		log.Error("%v", err)
	}
	err = nil
	return
}

// LaserAll get all laser.
func (d *Dao) LaserAllSilence(c context.Context) (res []*fkappmdl.Laser, err error) {
	var re struct {
		Code int               `json:"code"`
		Data []*fkappmdl.Laser `json:"data"`
	}
	if err = d.client.Get(c, d.laserAllSilence, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserAllSilence)
		return
	}
	for _, v := range re.Data {
		if v == nil {
			err = errors.New("list struct is nil")
			return
		}
	}
	res = re.Data
	return
}

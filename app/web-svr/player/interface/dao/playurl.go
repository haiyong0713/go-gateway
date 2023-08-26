package dao

import (
	"context"

	"go-common/library/ecode"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/web-svr/player/interface/model"

	"github.com/pkg/errors"
)

func (d *Dao) formatHlsCommon(params *model.ParamHls, mid int64) *v2.HlsCommonReq {
	req := &v2.HlsCommonReq{
		Aid:         params.AID,
		Cid:         params.CID,
		Qn:          params.Qn,
		Fnver:       params.Fnver,
		Fnval:       params.Fnval,
		Mid:         mid,
		BackupNum:   2, //客户端请求默认2个
		ForceHost:   params.ForceHost,
		RequestType: v2.RequestType(params.RequestType),
		DeviceType:  params.DeviceType,
		Buvid:       params.Buvid,
		Business:    v2.Business_UGC,
		QnCategory:  v2.QnCategory(params.QnCategory),
		Dolby:       params.Dolby,
		Platform:    params.Platform,
	}
	return req
}

// HlsScheduler is
func (d *Dao) HlsScheduler(c context.Context, params *model.ParamHls, mid int64) (*v2.HlsSchedulerReply, error) {
	req := d.formatHlsCommon(params, mid)
	res, err := d.playURLRPCV2.HlsScheduler(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.HlsScheduler args(%v)", req)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	return res, nil
}

// MasterScheduler is
func (d *Dao) MasterScheduler(c context.Context, params *model.ParamHls, mid int64) (*v2.MasterSchedulerReply, error) {
	req := d.formatHlsCommon(params, mid)
	res, err := d.playURLRPCV2.MasterScheduler(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.HlsScheduler args(%v)", req)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	return res, nil
}

func (d *Dao) M3U8Scheduler(c context.Context, params *model.ParamHls, mid int64) (*v2.M3U8SchedulerReply, error) {
	req := d.formatHlsCommon(params, mid)
	res, err := d.playURLRPCV2.M3U8Scheduler(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.M3U8Scheduler args(%v)", req)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	return res, nil
}

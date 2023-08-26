package dynamic

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"

	"github.com/pkg/errors"
)

const (
	epListURL     = "/pgc/internal/dynamic/v2/ep/list"
	batchInfoURL  = "/pugv/internal/dynamic/batch"
	seasonInfoURL = "/pugv/internal/dynamic/season"
	_fromDynamic  = 1
)

func (d *Dao) EpList(c context.Context, epids []int64, mobiApp, platform, device, ip string, build, fnver, fnval int) (map[int64]*dynmdl.PGCInfo, error) {
	epidStr := xstr.JoinInts(epids)
	params := url.Values{}
	params.Set("ep_ids", epidStr)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("ip", ip)
	params.Set("fnver", strconv.Itoa(fnver))
	params.Set("fnval", strconv.Itoa(fnval))
	epListURL := d.pgcInfo
	var ret struct {
		Code int                       `json:"code"`
		Msg  string                    `json:"msg"`
		Res  map[int64]*dynmdl.PGCInfo `json:"result"`
	}
	if err := d.client.Get(c, epListURL, "", params, &ret); err != nil {
		log.Errorc(c, "EpList http GET(%s) failed, params:(%s), error(%+v)", epListURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "EpList http GET(%s) failed, params:(%s), code: %v, msg: %v", epListURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "EpList url(%v) code(%v) msg(%v)", epListURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Res, nil
}

func (d *Dao) PGCSeason(c context.Context, season []int64) (map[int64]*dynmdl.PGCSeason, error) {
	params := url.Values{}
	params.Set("season_ids", xstr.JoinInts(season))
	pgcSeason := d.pgcSeason
	var ret struct {
		Code int                         `json:"code"`
		Msg  string                      `json:"message"`
		Data map[int64]*dynmdl.PGCSeason `json:"data"`
	}
	if err := d.client.Get(c, pgcSeason, "", params, &ret); err != nil {
		log.Errorc(c, "PGCSeason http GET(%s) failed, params:(%s), error(%+v)", pgcSeason, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "PGCSeason http GET(%s) failed, params:(%s), code: %v, msg: %v", pgcSeason, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "PGCSeason url(%v) code(%v) msg(%v)", pgcSeason, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) PGCBatch(c context.Context, batch []int64) (map[int64]*dynmdl.PGCBatch, error) {
	params := url.Values{}
	params.Set("batch_ids", xstr.JoinInts(batch))
	pgcBatch := d.pgcBatch
	var ret struct {
		Code int                        `json:"code"`
		Msg  string                     `json:"message"`
		Data map[int64]*dynmdl.PGCBatch `json:"data"`
	}
	if err := d.client.Get(c, pgcBatch, "", params, &ret); err != nil {
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), error(%+v)", pgcBatch, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), code: %v, msg: %v", pgcBatch, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "PGCBatch url(%v) code(%v) msg(%v)", pgcBatch, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) MyFollows(c context.Context, mid int64) (*pgcAppGrpc.FollowReply, error) {
	in := &pgcAppGrpc.FollowReq{
		Mid:  mid,
		From: _fromDynamic,
	}
	rsp, err := d.pgcAppGRPC.MyFollows(c, in)
	if err != nil {
		log.Errorc(c, "MyFollows(params:%+v) failed. error(%+v)", in, err)
		return nil, err
	}
	return rsp, nil
}

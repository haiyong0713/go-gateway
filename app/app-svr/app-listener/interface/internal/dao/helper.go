package dao

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	lismeta "git.bilibili.co/bapis/bapis-go/dynamic/common/metadata"
	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
	"github.com/pkg/errors"
)

func wrapDaoError(err error, funcName string, req interface{}) error {
	return errors.WithMessagef(errors.WithStack(err), "dao.%s failed req(%+v)", funcName, req)
}

func wrapHttpError(err error, uri string, req interface{}) error {
	return errors.WithMessagef(errors.WithStack(err), "bmHTTP failed url(%s) req(%+v)", uri, req)
}

func (d *dao) resolvePlayItem(_ context.Context, item *v1.PlayItem) (aid, cid int64) {
	switch item.ItemType {
	case model.PlayItemUGC, model.PlayItemAudio:
		aid = item.Oid
		if len(item.SubId) > 0 {
			cid = item.SubId[0]
		}
		return
	case model.PlayItemOGV:
		// TODO
	}
	return
}

func (d *dao) resolveFavItem(_ context.Context, itemType int32, oid int64) (aid int64) {
	switch itemType {
	case model.FavTypeVideo, model.FavTypeAudio:
		return oid
	case model.FavTypeOgv:
		return oid
	}
	return
}

type reportType int64
type reportAction int64

//nolint:deadcode,varcheck
const (
	// 上报类型
	_ reportType = iota
	_play
	_fav
	_share
	_thumb
	_coin
	_comment

	// 上报动作 通用
	_actDo     reportAction = 1
	_actCancel reportAction = 2

	// 上报特殊
	_playVod  reportAction = 1 // 点播
	_playRcmd reportAction = 2 // 连播
	_coinOne  reportAction = 1
	_coinTwo  reportAction = 2
)

type reportOpt struct {
	Typ       reportType
	Act       reportAction
	Mid       int64
	Buvid     string
	ArcType   int32
	Aid, Cid  int64
	Scene     string
	Device    *device.Device
	Network   *network.Network
	FromSpmId string
}

func (d *dao) reportAction(ctx context.Context, opt reportOpt) error {
	req := &listenerSvc.ReportPlayActionReq{
		Mid: opt.Mid, Buvid: opt.Buvid, Type: int64(opt.ArcType),
		Action: int64(opt.Typ), Detail: int64(opt.Act),
		Aid: opt.Aid, Cid: opt.Cid, Scene: opt.Scene,
		Device: toListenSvrDevMeta(opt.Device), Network: toListenSvrNetMeta(opt.Network),
		Spmid: opt.FromSpmId,
	}
	resp, err := d.listenerGRPC.ReportPlayAction(ctx, req)
	if err != nil {
		return wrapDaoError(err, "listenerGRPC.ReportPlayAction", req)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("report action type(%d) act(%d) failed", opt.Typ, opt.Act)
	}
	return nil
}

func toListenSvrDevMeta(dev *device.Device) *lismeta.Device {
	if dev == nil {
		return &lismeta.Device{}
	}
	return &lismeta.Device{
		AppId:    1, // 粉板
		Build:    int32(dev.Build),
		Buvid:    dev.Buvid,
		MobiApp:  dev.RawMobiApp,
		Platform: dev.RawPlatform,
		Device:   dev.Device,
		Channel:  dev.Channel, Brand: dev.Brand, Model: dev.Model,
		Osver: dev.Osver, VersionName: dev.VersionName,
		FpLocal: dev.FpLocal, FpRemote: dev.FpRemote,
	}
}

func toListenSvrNetMeta(net *network.Network) *lismeta.Network {
	if net == nil {
		return &lismeta.Network{}
	}
	return &lismeta.Network{
		Type: lismeta.NetworkType(net.Type),
		Tf:   lismeta.TFType(net.TF),
		Oid:  net.Operator,
	}
}

func toHistoryDevType(d *device.Device) int64 {
	switch d.Plat() {
	case device.PlatIPhone, device.PlatIPhoneI, device.PlatIPhoneB:
		return model.HistoryDeviceIPhone
	case device.PlatIPad, device.PlatIPadI:
		return model.HistoryDeviceIPad
	case device.PlatAndroid, device.PlatAndroidI, device.PlatAndroidB, device.PlatAndroidG:
		return model.HistoryDeviceAndroid
	case device.PlatAndroidHD:
		return model.HistoryDeviceAndroidPad
	case device.PlatAndroidTV, device.PlatAndroidTVYST:
		return model.HistoryDeviceAndroidTV
	case device.PlatWeb:
		return model.HistoryDevicePC
	case device.PlatWPhone:
		return model.HistoryDeviceWP8
	default:
		return model.HistoryDeviceUnknown
	}
}

func int64IDM2Slc(ids interface{}) ([]string, []int64) {
	var sidSlc []string
	var sids []int64
	switch sid := ids.(type) {
	case []int64:
		sidSlc = make([]string, 0, len(sid))
		for _, s := range sid {
			sidSlc = append(sidSlc, strconv.Itoa(int(s)))
		}
		sids = sid
	case map[int64]struct{}:
		sidSlc = make([]string, 0, len(sid))
		for s := range sid {
			sidSlc = append(sidSlc, strconv.Itoa(int(s)))
			sids = append(sids, s)
		}
	default:
		panic(fmt.Sprintf("programmer error: unknown type %T", ids))
	}
	return sidSlc, sids
}

func cloneUrlParam(in url.Values) url.Values {
	out := url.Values{}
	for k, v := range in {
		for _, vv := range v {
			out.Add(k, vv)
		}
	}
	return out
}

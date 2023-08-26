package dao

import (
	"context"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

const (
	_epPlayer = "/pgc/internal/dynamic/v3/ep/list"
)

type bangumiDao struct {
	host       string
	httpClient *bm.Client
}

func (d *bangumiDao) EpPlayer(c context.Context, epIDs []int64, dev *device.Device, playArg *arcgrpc.BatchPlayArg) (map[int64]*model.EpPlayer, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("ep_ids", xstr.JoinInts(epIDs))
	params.Set("mobi_app", dev.RawMobiApp)
	params.Set("platform", dev.RawPlatform)
	params.Set("device", dev.Device)
	params.Set("build", strconv.Itoa(int(dev.Build)))
	params.Set("ip", ip)
	if playArg != nil {
		params.Set("fnval", strconv.Itoa(int(playArg.Fnval)))
		params.Set("fnver", strconv.Itoa(int(playArg.Fnver)))
	}
	var res struct {
		Code   int                       `json:"code"`
		Result map[int64]*model.EpPlayer `json:"result"`
	}
	if err := d.httpClient.Get(c, d.host+_epPlayer, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.host+_epPlayer+"?"+params.Encode())
	}
	return res.Result, nil
}

package playurl

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/playurl"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"

	"github.com/pkg/errors"
)

type Dao struct {
	c            *conf.Config
	playURLRPCV2 v2.PlayURLClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	d.playURLRPCV2, err = v2.NewClient(c.PlayURLClient)
	if err != nil {
		panic(fmt.Sprintf("player v2 NewClient error(%v)", err))
	}
	return d
}

func (d *Dao) PlayURL(c context.Context, buvid string, mid int64, params *playurl.Param) (*v2.ResponseMsg, error) {
	fh := int32(1)
	if params.ForceHost > 0 {
		fh = int32(params.ForceHost)
	}
	req := &v2.PlayURLReq{
		Aid:      params.Oid,
		Cid:      params.Cid,
		Qn:       params.Qn,
		Platform: params.Platform,
		Fnver:    int32(params.Fnver),
		//增加dolby请求
		Fnval:        int32(params.Fnval),
		ForceHost:    fh,
		Mid:          mid,
		Fourk:        true,
		Device:       params.Device,
		MobiApp:      params.MobiApp,
		BackupNum:    d.c.Custom.BackupNum,
		Build:        int32(params.Build),
		Buvid:        buvid,
		NetType:      v2.NetworkType(params.NetType),
		TfType:       v2.TFType(params.TfType),
		VerifyVip:    1,
		H5Hq:         true,
		IsDazhongcar: params.IsDazhongcar,
	}
	player, err := d.playURLRPCV2.PlayURL(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.playURLRPCV2.PlayURL args(%v)", req)
		return nil, err
	}
	return player.GetPlayurl(), nil
}

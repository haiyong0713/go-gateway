package fission

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	fissiMdl "go-gateway/app/app-svr/app-resource/interface/model/fission"

	fissiGrpc "git.bilibili.co/bapis/bapis-go/account/service/fission"
)

// Dao is fission dao.
type Dao struct {
	// grpc
	fissiGRPC fissiGrpc.FissionClient
}

// New new a fission dao.
func New(c *conf.Config) (d *Dao) {
	g, err := fissiGrpc.NewClient(c.FissionGRPC)
	if err != nil {
		panic(err)
	}
	d = &Dao{
		// grpc
		fissiGRPC: g,
	}
	return
}

// CheckNew fission check new.
func (d *Dao) CheckNew(c context.Context, param *fissiMdl.ParamCheck) (rs *fissiGrpc.CheckNewResp, err error) {
	arg := &fissiGrpc.CheckNewReq{
		Mid:      param.Mid,
		Buvid:    param.Buvid,
		MobiApp:  param.MobiApp,
		Device:   param.Device,
		Platform: param.Platform,
		Build:    param.Build,
	}
	if rs, err = d.fissiGRPC.CheckNew(c, arg); err != nil {
		log.Error("d.fissiGRPC.CheckNew arg(%+v) error(%v)", arg, err)
		rs = &fissiGrpc.CheckNewResp{}
	}
	return
}

// CheckDevice fission check device.
func (d *Dao) CheckDevice(c context.Context, param *fissiMdl.ParamCheck) (rs *fissiGrpc.CheckNewResp, err error) {
	arg := &fissiGrpc.CheckDeviceReq{
		Buvid:    param.Buvid,
		MobiApp:  param.MobiApp,
		Device:   param.Device,
		Platform: param.Platform,
		Build:    param.Build,
	}
	if rs, err = d.fissiGRPC.CheckDevice(c, arg); err != nil {
		log.Error("d.fissiGRPC.CheckDevice arg(%+v) error(%v)", arg, err)
		rs = &fissiGrpc.CheckNewResp{}
	}
	return
}

// CheckDevice fission check device.
func (d *Dao) Entrance(c context.Context, mid, build int64, buvid, mobiApp, device, platform, ua string) (rs *fissiGrpc.EntranceReply, err error) {
	arg := &fissiGrpc.EntranceReq{
		Mid:      mid,
		Buvid:    buvid,
		MobiApp:  mobiApp,
		Device:   device,
		Platform: platform,
		Build:    build,
		Ip:       metadata.String(c, metadata.RemoteIP),
		Ua:       ua,
	}
	if rs, err = d.fissiGRPC.Entrance(c, arg); err != nil {
		return nil, err
	}
	return rs, nil
}

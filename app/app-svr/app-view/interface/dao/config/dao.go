package config

import (
	"context"
	"fmt"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"

	appConf "git.bilibili.co/bapis/bapis-go/community/service/appconfig"
)

type Dao struct {
	confClient appConf.AppConfigClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.confClient, err = appConf.NewClient(c.AppConfClient)
	if err != nil {
		panic(fmt.Sprintf("appconf NewClient error(%v)", err))
	}
	return
}

// VideoGuide .
func (d *Dao) VideoGuide(c context.Context, aid, cid, mid int64, dev device.Device) (*appConf.PlayerCardsReply, error) {
	res, err := d.confClient.PlayerCards(c, &appConf.PlayerCardsReq{
		Aid: aid,
		Cid: cid,
		Mid: mid,
		Common: &appConf.CommonParam{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
			Ip:       metadata.String(c, metadata.RemoteIP),
		}})
	if err != nil {
		log.Error("PlayerCards err(%+v) aid(%d) cid(%d)", err, aid, cid)
		return nil, err
	}
	return res, nil
}

// VideoGuide .
func (d *Dao) ClickPlayerCard(c context.Context, arg *viewApi.ClickPlayerCardReq, mid int64, dev device.Device) error {
	req := &appConf.ClickPlayerCardReq{
		Id:      arg.Id,
		Mid:     mid,
		OidType: appConf.OidTypeUGC,
		Oid:     arg.Cid,
		Pid:     arg.Aid,
		Common: &appConf.CommonParam{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Spmid:    arg.Spmid,
		},
		Action: arg.Action,
	}
	_, err := d.confClient.ClickPlayerCard(c, req)
	if err != nil {
		log.Error("PlayerCards err(%+v) req(%v)", err, req)
		return err
	}
	return nil
}

func (d *Dao) ClickPlayerCardV2(c context.Context, arg *viewApi.ClickPlayerCardReq, mid int64, dev device.Device) (*appConf.ClickPlayerCardResp, error) {
	req := &appConf.ClickPlayerCardReq{
		Id:      arg.Id,
		Mid:     mid,
		OidType: appConf.OidTypeUGC,
		Oid:     arg.Cid,
		Pid:     arg.Aid,
		Common: &appConf.CommonParam{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Spmid:    arg.Spmid,
		},
		Action: arg.Action,
	}
	res, err := d.confClient.ClickPlayerCard(c, req)
	if err != nil {
		log.Error("PlayerCards err(%+v) req(%v)", err, req)
		return nil, err
	}
	return res, nil
}

// ExposePlayerCard .
func (d *Dao) ExposePlayerCard(c context.Context, arg *viewApi.ExposePlayerCardReq, mid int64, dev device.Device) error {
	req := &appConf.ExposePlayerCardReq{
		CardType: appConf.PlayerCardType(arg.CardType),
		Mid:      mid,
		Cid:      arg.Cid,
		Aid:      arg.Aid,
		Common: &appConf.CommonParam{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Spmid:    arg.Spmid,
		},
	}
	_, err := d.confClient.ExposePlayerCard(c, req)
	if err != nil {
		log.Error("ExposePlayerCard err(%+v) req(%v)", err, req)
		return err
	}
	return nil
}

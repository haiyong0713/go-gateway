package listener

import (
	"context"
	"fmt"

	api "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
)

// Dao is space dao
type Dao struct {
	listenerRPCClient listenerSvc.ListenerSvrClient
}

type ListenerSwitchOpt struct {
	Mid   int64
	Aid   int64
	Spmid string
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.listenerRPCClient, err = listenerSvc.NewClient(c.ListenerClient); err != nil {
		panic(fmt.Sprintf("listenerSvc newListenerClient error (%+v)", err))
	}
	return
}

func (d *Dao) ListenerConfig(ctx context.Context, opt ListenerSwitchOpt) (ret *api.ListenerConfig, err error) {
	source := spmid2Source(opt.Spmid)

	req := &listenerSvc.ListenerSwitchReq{
		Uid:    opt.Mid,
		Aid:    opt.Aid,
		Source: source,
	}
	ret = new(api.ListenerConfig)

	resp, err := d.listenerRPCClient.ListenerSwitch(ctx, req)
	if err != nil {
		return
	}
	if resp.GetStatus() == 0 {
		return nil, nil
	}

	if resp.Guidebar != nil {
		ret.GuideBar = &api.ListenerGuideBar{
			ShowStrategy:   resp.Guidebar.GetStatus(),
			Icon:           resp.Guidebar.GetIconUrl(),
			Text:           resp.Guidebar.GetText(),
			BtnText:        resp.Guidebar.GetButtonText(),
			ShowTime:       resp.Guidebar.GetCountdown(),
			BackgroundTime: resp.Guidebar.GetBackgroundTime(),
		}
	}

	ret.JumpStyle = resp.Goto

	return ret, err
}

func spmid2Source(spmid string) int64 {
	const (
		// 1: 视频详情页 2: 播单的视频详情页 3: 合集\多P的详情页
		_ugcView       = 1
		_medialistView = 2
		//_dualPartView = 3
	)
	if spmid == "playlist.playlist-video-detail.0.0" {
		return _medialistView
	} else { // main.ugc-video-detail.0.0 以及其他
		return _ugcView
	}
}

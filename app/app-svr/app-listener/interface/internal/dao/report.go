package dao

import (
	"context"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
)

type PlayActionReportOpt struct {
	Mid       int64
	Buvid     string
	Item      *v1.PlayItem
	Device    *device.Device
	Network   *network.Network
	FromSpmId string
}

func (d *dao) PlayActionReport(ctx context.Context, opt PlayActionReportOpt) error {
	avid, cid := d.resolvePlayItem(ctx, opt.Item)
	if avid <= 0 {
		return nil
	}
	return d.reportAction(ctx, reportOpt{
		Typ:     _play, // act 目前默认不传 不区分点播连播
		ArcType: opt.Item.ItemType,
		Mid:     opt.Mid, Buvid: opt.Buvid,
		Aid: avid, Cid: cid,
		Scene:  opt.Item.GetEt().GetOperator(),
		Device: opt.Device, Network: opt.Network,
		FromSpmId: opt.FromSpmId,
	})
}

//nolint:biligowordcheck
func (d *dao) favActReport(_ context.Context, dev *device.Device, net *network.Network, mid, aid int64, arcType int32, act reportAction) {
	if aid <= 0 {
		return
	}
	go func() {
		err := d.reportAction(context.TODO(), reportOpt{
			Aid: aid, ArcType: arcType,
			Typ: _fav, Act: act,
			Mid: mid, Buvid: dev.Buvid,
			Device: dev, Network: net,
		})
		if err != nil {
			actText := "FAV_ADD"
			if act == _actCancel {
				actText = "FAV_DEL"
			}
			log.Warn("reportAction failed to report %s action aid(%d) mid(%d): %v", actText, aid, mid, err)
		}
		if arcType == model.PlayItemAudio {
			err = d.MusicClickReport(context.TODO(), MusicClickReportOpt{
				ClickTyp: MusicClickFav, SongId: aid, AddMetric: act == _actDo,
			})
			if err != nil {
				log.Error("MusicClickReport failed to report fav click: %v", err)
			}
		}
	}()
}

//nolint:biligowordcheck
func (d *dao) coinAddReport(_ context.Context, dev *device.Device, net *network.Network, mid, aid int64, arcType int32, coinNum int64) {
	if aid <= 0 {
		return
	}
	go func() {
		act := _coinOne
		if coinNum > 1 {
			act = _coinTwo
		}
		err := d.reportAction(context.TODO(), reportOpt{
			Aid: aid, ArcType: arcType,
			Typ: _coin, Act: act,
			Mid: mid, Buvid: dev.Buvid,
			Device: dev, Network: net,
		})
		if err != nil {
			log.Warn("reportAction failed to report COIN_ADD action aid(%d) mid(%d): %v", aid, mid, err)
		}
	}()
}

//nolint:biligowordcheck
func (d *dao) thumbReport(_ context.Context, dev *device.Device, net *network.Network, mid, aid int64, arcType int32, act v1.ThumbUpReq_ThumbType) {
	if aid <= 0 {
		return
	}
	var thumbAct reportAction
	switch act {
	case v1.ThumbUpReq_LIKE:
		thumbAct = _actDo
	case v1.ThumbUpReq_CANCEL_LIKE:
		thumbAct = _actCancel
	default:
		return
	}
	go func() {
		err := d.reportAction(context.TODO(), reportOpt{
			Aid: aid, ArcType: arcType,
			Typ: _thumb, Act: thumbAct,
			Mid: mid, Buvid: dev.Buvid,
			Device: dev, Network: net,
		})
		if err != nil {
			actText := "THUMB_LIKE"
			if thumbAct == _actCancel {
				actText = "THUMB_LIKE_CANCEL"
			}
			log.Warn("reportAction failed to report %s action aid(%d) mid(%d): %v", actText, aid, mid, err)
		}
	}()
}

type GuideBarShowReportOpt struct {
	Mid  int64
	Type int64
	Oid  int64
}

func (d *dao) GuideBarShowReport(ctx context.Context, opt GuideBarShowReportOpt) (success bool, err error) {
	req := &listenerSvc.ReportGuideBarShowReq{
		Uid:  opt.Mid,
		Type: opt.Type,
		Oid:  opt.Oid,
	}

	resp, err := d.listenerGRPC.ReportGuideBarShow(ctx, req)

	return resp.GetSuccess(), err
}

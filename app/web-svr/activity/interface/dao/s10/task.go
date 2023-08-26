package s10

import (
	"context"
	"strconv"

	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/s10"
	"go-gateway/app/web-svr/activity/interface/tool"

	task "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"
)

const (
	Path4ActPlatformOfCounterRes    = "GetCounterRes"
	Path4ActPlatformOfFormulaResult = "GetFormulaResult"
)

func (d *Dao) TaskPubDataBus(ctx context.Context, mid, timestamp int64, act string) (err error) {
	pack := &s10.S10MainDtb{Mid: mid, Timestamp: timestamp, Act: act}
	if err = d.signedDataBusPub.Send(ctx, strconv.FormatInt(mid, 10), pack); err != nil {
		tool.Metric4PubDatabus.WithLabelValues("SignPubDataBus").Inc()
		log.Errorc(ctx, "s10 d.dao.SignedPubDataBus(mid:%d) error(%v)", mid, err)
	}
	return
}

func fetchActPlatformCounterRes(ctx context.Context, req interface{}) (interface{}, error) {
	return client.ActPlatClient.GetCounterRes(ctx, req.(*task.GetCounterResReq))
}

func (d *Dao) GetCounterRes(ctx context.Context, mid, time int64, counter, act string) (int32, error) {
	req := &task.GetCounterResReq{
		Counter:  counter,
		Activity: act,
		Mid:      mid,
		Time:     time,
		Start:    nil,
	}
	reply, err := client.FetchResourceFromActPlatform(ctx, Path4ActPlatformOfCounterRes, fetchActPlatformCounterRes, req)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.GetCounterRes(mid:%d,time:%d,counter:%s,activity:%s) error:%v", mid, time, counter, act, err)
		return 0, err
	}
	var progress int32
	res := reply.(*task.GetCounterResResp)
	for _, v := range res.CounterList {
		if v.Val > 0 {
			progress += int32(v.Val)
		}
	}

	return progress, nil
}

func fetchActPlatformTotalPoints(ctx context.Context, req interface{}) (interface{}, error) {
	return client.ActPlatClient.GetFormulaTotal(ctx, req.(*task.GetFormulaTotalReq))
}

func (d *Dao) TotalPoints(ctx context.Context, mid int64, act string) (int32, error) {
	req := &task.GetFormulaTotalReq{
		Activity: act,
		Formula:  "total",
		Mid:      mid,
	}
	reply, err := client.FetchResourceFromActPlatform(ctx, Path4ActPlatformOfFormulaResult, fetchActPlatformTotalPoints, req)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.TotalPoints(%d) error(%v)", mid, err)
		return 0, err
	}
	return int32(reply.(*task.GetFormulaResultResp).Result), nil
}

package actplat

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"

	model "go-gateway/app/web-svr/activity/interface/model/actplat"
	"time"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

// AddFilter mid加入活动
func (d *Dao) AddFilter(ctx context.Context, mid int64, activityID, counter, filter string) (err error) {
	_, err = client.ActPlatClient.AddFilterMemberInt(ctx, &actplatapi.SetFilterMemberIntReq{
		Activity: activityID,
		Counter:  counter,
		Filter:   filter,
		Values:   []*actplatapi.FilterMemberInt{{Value: mid}},
	})
	if err != nil {
		log.Errorc(ctx, "client.ActPlatClient.AddFilterMemberInt mid(%d) err(%v)", mid, err)
		return err
	}
	return nil
}

// AddFilterSet mid加入活动
func (d *Dao) AddFilterSet(ctx context.Context, mid int64, activityID, filter string) (err error) {
	_, err = client.ActPlatClient.AddSetMemberInt(ctx, &actplatapi.SetMemberIntReq{
		Activity: activityID,
		Name:     filter,
		Values:   []*actplatapi.SetMemberInt{{Value: mid}},
	})
	if err != nil {
		log.Errorc(ctx, "client.ActPlatClient.AddSetMemberInt mid(%d) err(%v)", mid, err)
		return err
	}
	return nil
}

// Send 自定义流
func (d *Dao) Send(ctx context.Context, mid int64, data *model.ActivityPoints) (err error) {
	midStr := fmt.Sprintf("%d", mid)

	b, _ := json.Marshal(*data)
	if err = component.ActPlatProducer.Send(
		ctx,
		fmt.Sprintf(midStr),
		b); err != nil {
		log.Errorc(ctx, "SendPoint sync failed:%v", err)
	}
	log.Infoc(ctx, "SendPoint: d.actPlatPub.Send(%d,%v)", mid, *data)
	return err
}

// GetCounter ...
func (d *Dao) GetCounter(c context.Context, mid int64, activity, counter string) (int64, error) {
	resp, err := client.ActPlatClient.GetCounterRes(c, &actplatapi.GetCounterResReq{
		Counter:  counter,
		Activity: activity,
		Mid:      mid,
		Time:     time.Now().Unix(),
	})
	if err != nil {
		log.Errorc(c, "s.actPlatClient.GetCounterRes(%v) counter(%s) error(%v)", mid, counter, err)
		return 0, err
	}
	if resp == nil || len(resp.CounterList) != 1 {
		log.Errorc(c, "s.actPlatClient.GetCounterRes(%v) error(%v)", mid, err)
		return 0, err
	}
	res := resp.CounterList[0]
	return res.Val, nil
}

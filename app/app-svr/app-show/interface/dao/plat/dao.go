package plat

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-show/interface/conf"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

// Dao is activity dao.
type Dao struct {
	platRPC actplatapi.ActPlatClient
}

// New new a activity dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.platRPC, err = actplatapi.NewClient(c.PlatGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return d
}

// GetTotalRes 累计的积分统计.
func (d *Dao) GetTotalRes(c context.Context, counter, activity string, mid int64) (int64, error) {
	rly, err := d.platRPC.GetTotalRes(c, &actplatapi.GetTotalResReq{Counter: counter, Activity: activity, Mid: mid})
	if err != nil {
		log.Error("d.platRPC.GetTotalRes(%s,%s,%d) error(%v)", counter, activity, mid, err)
		return 0, err
	}
	if rly == nil {
		return 0, ecode.NothingFound
	}
	return rly.Total, nil
}

// GetCounterRes 单日的积分统计.
func (d *Dao) GetCounterRes(c context.Context, counter, activity string, mid int64) (int64, error) {
	rly, err := d.platRPC.GetCounterRes(c, &actplatapi.GetCounterResReq{Counter: counter, Activity: activity, Mid: mid, Time: time.Now().Unix()})
	if err != nil {
		log.Error("d.platRPC.GetTotalRes(%s,%s,%d) error(%v)", counter, activity, mid, err)
		return 0, err
	}
	if rly == nil || len(rly.CounterList) == 0 {
		return 0, ecode.NothingFound
	}
	return rly.CounterList[0].Val, nil
}

func (d *Dao) GetHistory(c context.Context, activity, counter string, mid int64, start []byte) (*actplatapi.GetHistoryResp, error) {
	req := &actplatapi.GetHistoryReq{
		Activity: activity,
		Counter:  counter,
		Mid:      mid,
		Start:    start,
	}
	rly, err := d.platRPC.GetHistory(c, req)
	if err != nil {
		log.Error("Fail to request actplatapi.GetHistory, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

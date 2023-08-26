package dao

import (
	"context"
	"fmt"
	"time"

	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"

	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
)

type activityCfg struct {
	Client *warden.ClientConfig
}

type activityDao struct {
	client actGRPC.ActivityClient
}

func newActivityDao(cfg *activityCfg) *activityDao {
	d := &activityDao{}
	var err error
	if d.client, err = actGRPC.NewClient(cfg.Client); err != nil {
		panic(fmt.Sprintf("Fail to new activityClient, config=%+v error=%+v", cfg.Client, err))
	}
	return d
}

func (d *activityDao) GetReserveProgress(c context.Context, sid, mid, ruleID, typ, dataType int64, dimension actGRPC.GetReserveProgressDimension) (int64, error) {
	if sid == 0 {
		log.Errorc(c, "Sid is empty")
		return 0, errors.New("sid is empty")
	}
	req := &actGRPC.GetReserveProgressReq{
		Sid: sid,
		Mid: mid,
		Rules: []*actGRPC.ReserveProgressRule{
			{Dimension: dimension, RuleId: ruleID, Type: typ, DataType: dataType},
		},
	}
	rly, err := d.client.GetReserveProgress(c, req)
	if err != nil {
		log.Errorc(c, "Fail to get reserveProgress, req=%+v error=%+v", req, err)
		return 0, err
	}
	for _, v := range rly.Data {
		if v == nil || v.Rule == nil {
			continue
		}
		if v.Rule.Dimension == dimension && v.Rule.RuleId == ruleID && v.Rule.Type == typ && v.Rule.DataType == dataType {
			return v.Progress, nil
		}
	}
	log.Errorc(c, "Progress not found in result, req=%+v", req)
	return 0, errors.New("Progress not found in result")
}

func (d *activityDao) ActivityProgress(c context.Context, sid, typ, mid int64, gids []int64) (*actGRPC.ActivityProgressReply, error) {
	req := &actGRPC.ActivityProgressReq{Sid: sid, Gids: gids, Type: typ, Mid: mid, Time: time.Now().Unix()}
	rly, err := d.client.ActivityProgress(c, req)
	if err != nil {
		log.Error("Fail to request ActivityProgress, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

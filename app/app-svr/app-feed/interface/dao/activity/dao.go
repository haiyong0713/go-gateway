package activity

import (
	"context"
	"fmt"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	appResourcegrpc "go-gateway/app/app-svr/app-resource/interface/api/v1"

	"github.com/pkg/errors"
)

var (
	ErrNoLiveReservation = errors.New("story-reservation: no live reservation")
)

type Dao struct {
	actClient         activitygrpc.ActivityClient
	appResourceClient appResourcegrpc.AppResourceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.actClient, err = activitygrpc.NewClient(c.ActivityClient); err != nil {
		panic(fmt.Sprintf("activityClient NewClient error(%v)", err))
	}
	if d.appResourceClient, err = appResourcegrpc.NewClient(c.AppResourceClient); err != nil {
		panic(fmt.Sprintf("appResourceClient NewClient error(%v)", err))
	}
	return
}

func (d *Dao) StoryLiveReserveCard(c context.Context, arg *activitygrpc.UpActReserveRelationInfo4LiveReq) (*activitygrpc.UpActReserveRelationInfo, error) {
	reply, err := d.actClient.UpActReserveRelationInfo4Live(c, arg)
	if err != nil {
		return nil, err
	}
	if len(reply.List) == 0 {
		return nil, ErrNoLiveReservation
	}
	return reply.List[0], nil
}

func (d *Dao) StoryLiveReserveKeyExists(c context.Context, req *appResourcegrpc.CheckEntranceInfocRequest) (bool, error) {
	reply, err := d.appResourceClient.CheckEntranceInfoc(c, req)
	if err != nil {
		return false, err
	}
	return reply.IsExisted, nil
}

func (d *Dao) ActReserveCard(ctx context.Context, mid int64, rids []int64) (map[int64]*activitygrpc.UpActReserveRelationInfo, error) {
	reply, err := d.actClient.UpActReserveRelationInfo(ctx, &activitygrpc.UpActReserveRelationInfoReq{
		Mid:  mid,
		Sids: rids,
	})
	if err != nil {
		return nil, err
	}
	if len(reply.List) == 0 {
		return nil, errors.Errorf("no reservation: %+v", rids)
	}
	return reply.List, nil
}

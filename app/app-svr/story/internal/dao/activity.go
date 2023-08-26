package dao

import (
	"context"

	appResourcegrpc "go-gateway/app/app-svr/app-resource/interface/api/v1"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"github.com/pkg/errors"
)

var (
	ErrNoLiveReservation = errors.New("story-reservation: no live reservation")
)

func (d *dao) StoryLiveReserveCard(c context.Context, arg *activitygrpc.UpActReserveRelationInfo4LiveReq) (*activitygrpc.UpActReserveRelationInfo, error) {
	reply, err := d.actClient.UpActReserveRelationInfo4Live(c, arg)
	if err != nil {
		return nil, err
	}
	if len(reply.List) == 0 {
		return nil, ErrNoLiveReservation
	}
	return reply.List[0], nil
}

func (d *dao) StoryLiveReserveKeyExists(c context.Context, req *appResourcegrpc.CheckEntranceInfocRequest) (bool, error) {
	reply, err := d.appResourceClient.CheckEntranceInfoc(c, req)
	if err != nil {
		return false, err
	}
	return reply.IsExisted, nil
}

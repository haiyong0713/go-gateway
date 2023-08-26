package dao

import (
	"context"

	"go-common/library/log"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
)

// ActProtocol get act subject & protocol
func (d *Dao) ActProtocol(c context.Context, messionID int64) (*actgrpc.ActSubProtocolReply, error) {
	rly, err := d.ActivityClient.ActSubProtocol(c, &actgrpc.ActSubProtocolReq{Sid: messionID})
	if err != nil {
		log.Errorc(c, "Fail to request activity.ActSubProtocol, sid=%+v error=%+v", messionID, err)
		return nil, err
	}
	return rly, nil
}

// 通过稿件ID查询首映预约ID
func (d *Dao) GetPremiereSidByAid(c context.Context, aid int64) (*actgrpc.GetPremiereSidByAidReply, error) {
	res, err := d.ActivityClient.GetPremiereSidByAid(c, &actgrpc.GetPremiereSidByAidReq{Aid: aid})
	if err != nil {
		log.Error("d.GetPremiereSidByAid error:%+v, aid:%d", err, aid)
		return nil, err
	}
	return res, nil
}

package native

import (
	"context"

	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/log"
)

// UpActReserveRelationInfo.
func (d *Dao) UpActReserveRelationInfo(c context.Context, mid int64, sids []int64) (map[int64]*actGRPC.UpActReserveRelationInfo, error) {
	req := &actGRPC.UpActReserveRelationInfoReq{Sids: sids, Mid: mid}
	rly, err := d.actClient.UpActReserveRelationInfo(c, req)
	if err != nil {
		log.Error("Fail to request UpActReserveRelationInfo, req=%+v error=%+v", req, err)
		return nil, err
	}
	if rly == nil {
		return make(map[int64]*actGRPC.UpActReserveRelationInfo), nil
	}
	return rly.List, nil
}

func (d *Dao) ActSubject(c context.Context, sid int64) (*actGRPC.Subject, error) {
	rly, err := d.actClient.ActSubject(c, &actGRPC.ActSubjectReq{Sid: sid})
	if err != nil {
		log.Error("Fail to reqeust actGRPC.ActSubject, sid=%d error=%+v", sid, err)
		return nil, err
	}
	if rly == nil {
		return nil, nil
	}
	return rly.Subject, nil
}

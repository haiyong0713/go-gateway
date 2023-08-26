package dao

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	"go-common/library/log"
)

func (d *Dao) IsUpActUid(c context.Context, mid int64) (bool, error) {
	if mid == 0 {
		return false, nil
	}
	rly, err := d.naPageClient.IsUpActUid(c, &api.IsUpActUidReq{Mid: mid})
	if err != nil {
		log.Error("Fail to get isUpActUid, mid=%+v error=%+v", mid, err)
		return false, err
	}
	return rly.Match, nil
}

package dao

import (
	"context"
	activityGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	xecode "go-common/library/ecode"
	xtime "go-common/library/time"
	"time"
)

// CheckIfAuthorizedByMID 从活动平台获取用户是否已授权
func (d *Dao) CheckIfAuthorizedByMID(sid int64, mid int64) (authorized bool, authorizationTime xtime.Time, err error) {
	if sid == 0 || mid == 0 {
		return false, xtime.Time(time.Time{}.Unix()), xecode.RequestErr
	}
	req := &activityGRPC.ReserveFollowingReq{
		Sid: sid,
		Mid: mid,
	}
	res, _err := d.activityGRPCClient.ReserveFollowing(context.Background(), req)
	if _err != nil {
		return false, xtime.Time(time.Time{}.Unix()), _err
	}
	if res.IsFollow {
		authorized = true
		authorizationTime = res.Mtime
	}

	return
}

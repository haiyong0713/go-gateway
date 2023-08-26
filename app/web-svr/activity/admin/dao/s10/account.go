package s10

import (
	"context"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	blackList "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"

	"go-common/library/log"
)

func (d *Dao) Profile(ctx context.Context, mid int64) (*account.Profile, error) {
	reply, err := d.accountClient.Profile3(ctx, &account.MidReq{Mid: mid})
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.Profile(mid:%d) error:%v", mid, err)
		return nil, err
	}
	return reply.Profile, nil
}

func (d *Dao) InBackList(ctx context.Context, mid int64, action []string) (bool, error) {
	reply, err := d.backListClient.IsAllowedToDo(ctx, &blackList.IsAllowedToDoReq{Mid: mid, ControlAction: action})
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.InBackList(mid:%d,action:%v) error:%v", mid, action, err)
		return false, err
	}
	return !reply.AllAllowed, nil
}

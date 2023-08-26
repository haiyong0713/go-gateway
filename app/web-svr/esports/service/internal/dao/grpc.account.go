package dao

import (
	"context"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
)

func (d *dao) GetAccountInfos(ctx context.Context, mids []int64) (userInfoMap *accapi.InfosReply, err error) {
	userArg := &accapi.MidsReq{Mids: mids}
	if userInfoMap, err = d.accountClient.Infos3(ctx, userArg); err != nil {
		log.Errorc(ctx, "ContestReplyWall s.accClient.Infos3() mids(%+v) error(%+v)", mids, err)
		return
	}
	return
}

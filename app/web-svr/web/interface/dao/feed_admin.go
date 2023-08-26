package dao

import (
	"context"

	feedadmingrpc "git.bilibili.co/bapis/bapis-go/platform/admin/app-feed"
	"go-common/library/log"
)

func (d *Dao) CreatePwdAppeal(c context.Context, req *feedadmingrpc.CreatePwdAppealReq) (*feedadmingrpc.CreatePwdAppealRly, error) {
	rly, err := d.feedAdminClient.CreatePwdAppeal(c, req)
	if err != nil {
		log.Errorc(c, "Fail to request feedadmingrpc.CreatePwdAppeal, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

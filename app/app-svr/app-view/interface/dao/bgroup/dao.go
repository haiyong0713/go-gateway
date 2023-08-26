package bgroup

import (
	"context"
	"fmt"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
)

type Dao struct {
	bGroupClient api.BGroupServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.bGroupClient, err = api.NewClient(c.BGroupClient); err != nil {
		panic(fmt.Sprintf("bgroup NewClient not found err(%v)", err))
	}
	return
}

// 是否在某个人群包中
func (d *Dao) GetMidExists(c context.Context, req *api.MemberInReq) ([]*api.MemberInReply_MemberInReplySingle, error) {
	res, err := d.bGroupClient.MemberIn(c, req)
	if err != nil {
		log.Error("d.bGroupClient.MemberIn err:%+v", err)
		return nil, err
	}
	return res.Results, nil
}

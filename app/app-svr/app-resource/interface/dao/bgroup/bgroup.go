package bgroup

import (
	"context"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
)

type Dao struct {
	bgroup bGroup.BGroupServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.bgroup, err = bGroup.NewClient(c.BGroupClient); err != nil {
		panic(err)
	}
	return
}

// LoadingUserEquip .
func (d *Dao) MemberIn(ctx context.Context, in *bGroup.MemberInReq) (*bGroup.MemberInReply, error) {
	return d.bgroup.MemberIn(ctx, in)
}

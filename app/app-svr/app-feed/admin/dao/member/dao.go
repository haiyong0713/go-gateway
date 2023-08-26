package member

import (
	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	client membergrpc.MemberClient
}

func NewDao(cfg *conf.Config) *Dao {
	client, err := membergrpc.NewClient(cfg.MemberClient)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

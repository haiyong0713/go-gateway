package fawkes

import (
	"context"

	"go-gateway/app/app-svr/fawkes/service/conf"

	fissiGrpc "git.bilibili.co/bapis/bapis-go/account/service/fission"
)

func NewFisson(c *conf.Config) fissiGrpc.FissionClient {
	g, err := fissiGrpc.NewClient(c.FissionGRPC)
	if err != nil {
		panic(err)
	}
	return g
}

// DomainStatus fission check domain status.
func (d *Dao) DomainStatus(c context.Context, domain []string) (rs *fissiGrpc.DomainStatustReply, err error) {
	arg := &fissiGrpc.DomainStatusReq{
		Domain: domain,
	}
	if rs, err = d.fission.DomainStatus(c, arg); err != nil {
		return nil, err
	}
	return rs, nil
}

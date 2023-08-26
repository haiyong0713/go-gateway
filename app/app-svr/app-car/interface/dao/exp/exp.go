package exp

import (
	"context"

	ab "git.bilibili.co/bapis/bapis-go/ott/ab"
	"go-gateway/app/app-svr/app-car/interface/conf"
)

type Dao struct {
	c         *conf.Config
	expClient ab.ABClient
}

func New(c *conf.Config) *Dao {
	expCli, err := ab.NewClientAB(c.ABClientCfg)
	if err != nil {
		panic(err)
	}
	return &Dao{
		c:         c,
		expClient: expCli,
	}
}

func (d *Dao) ExpGroupMatch(c context.Context, req *ab.ExpGroupMatchReq) (*ab.ExpGroupMatchReply, error) {
	match, err := d.expClient.ExpGroupMatch(c, req)
	if err != nil {
		return nil, err
	}
	return match, nil
}

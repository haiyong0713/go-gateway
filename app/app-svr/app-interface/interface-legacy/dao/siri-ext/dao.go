package siri_ext

import (
	"context"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/siri-ext/service/api"
)

type Dao struct {
	c             *conf.Config
	siriExtClient api.SiriExtClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.siriExtClient, err = api.NewClient(c.SiriExtClieng); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) ResolveCommand(ctx context.Context, req *api.ResolveCommandReq) (*api.ResolveCommandReply, error) {
	return d.siriExtClient.ResolveCommand(ctx, req)
}

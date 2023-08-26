package nativePage

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	nativePagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

type Dao struct {
	c                *conf.Config
	nativePageClient nativePagegrpc.NaPageClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.nativePageClient, err = nativePagegrpc.NewClient(c.NativePageGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) IsUpActUID(c context.Context, mid int64) (bool, error) {
	resTmp, err := d.nativePageClient.IsUpActUid(c, &nativePagegrpc.IsUpActUidReq{Mid: mid})
	if err != nil {
		return false, err
	}
	return resTmp.Match, nil
}

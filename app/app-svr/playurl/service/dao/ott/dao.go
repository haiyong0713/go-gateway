package ott

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/playurl/service/conf"

	api "git.bilibili.co/bapis/bapis-go/tv/interface"

	"github.com/pkg/errors"
)

// Dao is ott dao.
type Dao struct {
	ottClient api.TVInterfaceClient
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.ottClient, err = api.NewClient(c.OTTClient)
	if err != nil {
		panic(fmt.Sprintf("ott NewClient error(%v)", err))
	}
	return
}

// VideoAuthUgc is
func (d *Dao) VideoAuthUgc(c context.Context, aid, cid int64) (bool, error) {
	reply, err := d.ottClient.VideoAuthUgc(c, &api.VideoAuthUgcReq{Aid: aid, Cid: cid})
	if err != nil || reply == nil {
		err = errors.Wrapf(err, "VideoAuthUgc arg(%+v)", &api.VideoAuthUgcReq{Aid: aid, Cid: cid})
		return false, err
	}
	return reply.CanPlay, nil
}

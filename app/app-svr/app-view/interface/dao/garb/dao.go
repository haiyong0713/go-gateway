package garb

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-view/interface/conf"

	garbmdl "git.bilibili.co/bapis/bapis-go/garb/model"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
)

type Dao struct {
	garbClient garb.GarbClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.garbClient, err = garb.NewClient(c.GarbClient); err != nil {
		panic(err)
	}
	return
}

// ThumbupUserEquip .
func (d *Dao) ThumbupUserEquip(c context.Context, mid int64) (*garbmdl.UserThumbup, error) {
	rly, err := d.garbClient.ThumbupUserEquip(c, &garb.ThumbupUserEquipReq{Mid: mid})
	if err != nil {
		log.Error("d.garbClient.ThumbupUserEquip(%d) error(%v)", mid, err)
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	return rly, nil
}

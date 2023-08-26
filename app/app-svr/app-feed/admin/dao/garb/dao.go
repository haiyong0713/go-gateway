package garb

import (
	"context"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	api "git.bilibili.co/bapis/bapis-go/garb/service"
)

// Dao is space dao
type Dao struct {
	c          *conf.Config
	garbClient api.GarbClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.garbClient, err = api.NewClient(c.GarbClient); err != nil {
		panic(err)
	}
	return
}

// SkinInfo .
func (d *Dao) SkinInfos(c context.Context, sids []int64) (rly *api.SkinListReply, err error) {
	return d.garbClient.SkinList(c, &api.SkinListReq{IDs: sids})
}

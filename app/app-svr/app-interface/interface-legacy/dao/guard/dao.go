package guard

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	api "git.bilibili.co/bapis/bapis-go/live/xuser/v1"
)

// Dao is space dao
type Dao struct {
	c           *conf.Config
	guardClient api.GuardClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.guardClient, err = api.NewClient(c.GuardGRPC); err != nil {
		panic(err)
	}
	return
}

// GetTopListGuardAttr .
func (d *Dao) GetTopListGuardAttr(c context.Context, mid, pn, ps int64, sortAttr []string) (rly *api.GetTopListGuardAttrResp, err error) {
	if rly, err = d.guardClient.GetTopListGuardAttr(c, &api.GetTopListGuardAttrReq{Targetid: mid, Page: pn, PageSize: ps, SortAttr: sortAttr}); err != nil {
		log.Error("d.guardClient.GetTopListGuardAttr(%d,%v) error(%v)", mid, sortAttr, err)
	}
	return
}

package manager

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"git.bilibili.co/bapis/bapis-go/manager/service/active"
)

// Dao dao
type Dao struct {
	c            *conf.Config
	mngActClient api.CommonActiveClient
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
	}
	var err error
	if dao.mngActClient, err = api.NewClient(c.MngClient); err != nil {
		panic(fmt.Sprintf("manager active newClient panic(%+v)", err))
	}
	return
}

// CommonActivity .
func (d *Dao) CommonActivity(c context.Context, sid, mid int64, asPlat int32) (*api.CommonActivityResp, error) {
	req := &api.CommonActivityReq{SeasonId: sid, Plat: asPlat, Mid: mid}
	return d.mngActClient.CommonActivity(c, req)
}

// CommonActivities .
func (d *Dao) CommonActivities(c context.Context) (*api.CommonActivitiesResp, error) {
	req := &api.CommonActivitiesReq{}
	return d.mngActClient.CommonActivities(c, req)
}

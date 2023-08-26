package manager

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/archive/service/conf"
	"go-gateway/app/app-svr/archive/service/model"

	"git.bilibili.co/bapis/bapis-go/manager/service/active"
	"github.com/pkg/errors"
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
		panic(fmt.Sprintf("manager newClient panic(%+v)", err))
	}
	return
}

// ActSeasonColor .
func (d *Dao) ActSeasonColor(c context.Context, sids []int64, mid int64, mobiApp, device string) (map[int64]*api.Color, error) {
	plat := model.Plat(mobiApp, device)
	asPlat := int32(1)
	if model.IsIPad(plat) {
		asPlat = 2
	}
	req := &api.BatchColorReq{SeasonId: sids, Mid: mid, Plat: asPlat}
	res, err := d.mngActClient.BatchColor(c, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("mng activity is nil")
	}
	return res.BatchColor, nil
}

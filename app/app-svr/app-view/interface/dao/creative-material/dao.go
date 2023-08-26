package creative_material

import (
	"context"
	"fmt"

	"go-common/library/ecode"

	"github.com/pkg/errors"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/material/creative/interface/v1"
)

type Dao struct {
	creativeMaterial api.MaterialCreativeClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.creativeMaterial, err = api.NewClientMaterialCreative(c.CreativeMaterialClient); err != nil {
		panic(fmt.Sprintf("creativeMaterial NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) GetArcMaterialListTag(ctx context.Context, req *api.ArcMaterialListReq) (*api.PlayPageMaterialTag, error) {
	res, err := d.creativeMaterial.GetArcMaterialList(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, " d.creativeMaterial.GetArcMaterialList is err %+v %+v", req, err)
	}
	if res.PlayPageTag == nil {
		return nil, ecode.NothingFound
	}
	return res.PlayPageTag, nil
}

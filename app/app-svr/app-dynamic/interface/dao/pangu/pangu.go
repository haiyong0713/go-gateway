package pangu

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	pangugrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

type Dao struct {
	c         *conf.Config
	panguGRPC pangugrpc.GalleryServiceClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.panguGRPC, err = pangugrpc.NewClient(c.PanguGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) GetNFTRegion(c context.Context, ids []string) (map[string]*pangugrpc.NFTRegion, error) {
	arg := &pangugrpc.GetNFTRegionReq{
		NftId: ids,
	}
	reply, err := d.panguGRPC.GetNFTRegion(c, arg)
	if err != nil {
		return nil, err
	}
	return reply.Region, nil
}

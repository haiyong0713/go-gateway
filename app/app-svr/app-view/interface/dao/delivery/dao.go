package delivery

import (
	"context"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-view/interface/conf"

	"google.golang.org/grpc"

	api "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
)

type Dao struct {
	deliveryClient api.DeliveryClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.deliveryClient, err = deliveryNewClient(c.DeliveryClient); err != nil {
		panic(fmt.Sprintf("delivery NewClient not found err(%v)", err))
	}
	return
}

// NewClient new a grpc client
func deliveryNewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (api.DeliveryClient, error) {
	const (
		_appID = "ogv.operation.servant"
	)
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+_appID)
	if err != nil {
		return nil, err
	}
	return api.NewDeliveryClient(conn), nil
}

// 批量查询ep物料
func (d *Dao) GetBatchEpMaterial(ctx context.Context, materialEpId map[int64]int32) (map[int64]*api.EpMaterial, error) {
	epMaterial := []*api.EpMaterialReq{}
	for materialId, epId := range materialEpId {
		epMaterial = append(epMaterial, &api.EpMaterialReq{
			MaterialNo: materialId,
			Epid:       epId,
		})
	}
	req := &api.BatchEpMaterialReq{
		Reqs:     epMaterial,
		BizScene: 0,
	}
	res, err := d.deliveryClient.BatchEpMaterial(ctx, req)
	if err != nil {
		return nil, err
	}
	if res.GetMaterialMap() == nil {
		return map[int64]*api.EpMaterial{}, ecode.NothingFound
	}
	return res.MaterialMap, nil
}

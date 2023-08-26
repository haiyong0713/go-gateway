package subscription

import (
	"context"

	"go-common/library/log"

	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
)

func (d *Dao) Tunnel(c context.Context, ids []int64, general *mdlv2.GeneralParam) (map[int64]*tunnelgrpc.DynamicCardMaterial, error) {
	resTmp, err := d.tunnelClient.DynamicCardMaterial(c, &tunnelgrpc.DynamicCardMaterialReq{
		Mid:      general.Mid,
		Oids:     ids,
		Platform: general.GetPlatform(),
		MobiApp:  general.GetMobiApp(),
	})
	if err != nil {
		log.Error("DynamicCardMaterial err %v", err)
		return nil, err
	}
	var res = make(map[int64]*tunnelgrpc.DynamicCardMaterial)
	for _, t := range resTmp.GetCards() {
		if t == nil {
			continue
		}
		res[t.Oid] = t
	}
	return res, err
}

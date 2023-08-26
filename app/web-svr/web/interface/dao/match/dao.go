package match

import (
	"context"

	"go-gateway/app/web-svr/web/interface/conf"

	esportConfGRPC "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	esportGRPC "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
)

type Dao struct {
	c              *conf.Config
	esportgrpc     esportGRPC.EsportsServiceClient
	esportConfgrpc esportConfGRPC.SearchMgrInterfaceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.esportgrpc, err = esportGRPC.NewClient(c.ESportsGRPC); err != nil {
		panic(err)
	}
	if d.esportConfgrpc, err = esportConfGRPC.NewClientSearchMgrInterface(c.ESportsConfGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) LiveContests(c context.Context, mid int64, matchIDs []int64) (map[int64]*esportGRPC.ContestDetail, error) {
	args := &esportGRPC.GetContestsRequest{Mid: mid, Cids: matchIDs}
	reply, err := d.esportgrpc.GetContests(c, args)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*esportGRPC.ContestDetail)
	for _, match := range reply.GetContests() {
		if match == nil || match.ID == 0 {
			continue
		}
		res[match.ID] = match
	}
	return res, nil
}

func (d *Dao) GetEsportConfigs(c context.Context, esportIDs []int64) (map[int64]*esportConfGRPC.EsportConfigInfo, error) {
	args := &esportConfGRPC.GetEsportConfigsReq{EsportIds: esportIDs, Plat: 30}
	reply, err := d.esportConfgrpc.GetEsportConfigs(c, args)
	if err != nil {
		return nil, err
	}
	return reply.GetConfigs(), nil
}

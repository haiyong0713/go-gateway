package esports

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
)

type Dao struct {
	esportsClient api.EsportsServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.esportsClient, err = api.NewClient(c.ESportsClient); err != nil {
		panic(fmt.Sprintf("esports NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) GetContestInfo(c context.Context, req *api.GetContestInfoByBvIdRequest) (*api.GetContestInfoByBvIdResponse, error) {
	res, err := d.esportsClient.GetContestInfoByBvId(c, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) GetCapsuleCard(ctx context.Context, tagIds []int64, mid int64) (*api.GetCapsuleCardByTagIdsResponse, error) {
	req := &api.GetCapsuleCardByTagIdsReq{
		TagIds: tagIds,
		Mid:    mid,
	}
	res, err := d.esportsClient.GetCapsuleCardByTagIds(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

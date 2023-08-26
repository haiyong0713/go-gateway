package uprating

import (
	"context"

	upratingGRPC "git.bilibili.co/bapis/bapis-go/crm/service/uprating"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	client upratingGRPC.UpRatingClient
}

func New(c *conf.Config) *Dao {
	client, err := upratingGRPC.NewClient(c.UpRatingClient)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

func (d *Dao) Rating(c context.Context, mid int64) (*upratingGRPC.RatingReply, error) {
	rly, err := d.client.Rating(c, &upratingGRPC.MidReq{Mid: mid})
	if err != nil {
		log.Error("Fail to request upratingGRPC.Rating(), mid=%+v error=%+v", mid, err)
		return nil, err
	}
	return rly, nil
}

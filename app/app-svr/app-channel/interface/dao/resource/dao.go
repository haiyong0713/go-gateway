package resource

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-channel/interface/conf"
	resApi "go-gateway/app/app-svr/resource/service/api/v1"
)

// Dao is resource dao.
type Dao struct {
	resGRPC resApi.ResourceClient
}

// New new a location dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.resGRPC, err = resApi.NewClient(c.ResourceGRPC); err != nil {
		panic(fmt.Sprintf("resApi.NewClient err(%+v)", err))
	}
	return
}

// EntrancesIsHidden is
func (d *Dao) EntrancesIsHidden(ctx context.Context, oids []int64, build int, plat int8, channel string) (*resApi.EntrancesIsHiddenReply, error) {
	req := &resApi.EntrancesIsHiddenRequest{
		Oids:    oids,
		Otype:   1, //分区入口对应rid,
		Build:   int64(build),
		Plat:    int32(plat),
		Channel: channel,
	}
	res, err := d.resGRPC.EntrancesIsHidden(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

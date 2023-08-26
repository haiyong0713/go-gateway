package up

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

// Dao is up dao
type Dao struct {
	upClient upgrpc.UpArchiveClient
}

// New initial up dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.upClient, err = upgrpc.NewClient(c.UpGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) ArcsPassedTotal(c context.Context, mids []int64) (map[int64]int64, error) {
	reply, err := d.upClient.ArcsPassedTotal(c, &upgrpc.ArcsPassedTotalReq{Mids: mids})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.Total, nil
}

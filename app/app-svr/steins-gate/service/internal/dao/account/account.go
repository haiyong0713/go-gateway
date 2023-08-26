package account

import (
	"context"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/steins-gate/service/conf"
)

// Dao is
type Dao struct {
	c   *conf.Config
	acc accountapi.AccountClient
}

// New new a dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	acc, err := accountapi.NewClient(c.Account)
	if err != nil {
		panic(err)
	}
	d.acc = acc
	return d
}

// Infos3 is
func (d *Dao) Infos3(ctx context.Context, mids []int64) (map[int64]*accountapi.Info, error) {
	reply, err := d.acc.Infos3(ctx, &accountapi.MidsReq{
		Mids:   mids,
		RealIp: metadata.String(ctx, metadata.RemoteIP),
	})
	if err != nil {
		return nil, err
	}
	return reply.Infos, nil

}

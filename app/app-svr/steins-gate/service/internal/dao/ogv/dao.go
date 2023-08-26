package ogv

import (
	ogvgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/pay"

	"go-gateway/app/app-svr/steins-gate/service/conf"
)

// Dao dao.
type Dao struct {
	c            *conf.Config
	ogvpayClient ogvgrpc.PayServiceClient
}

// New new a dao and return.
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
	}
	var err error
	if dao.ogvpayClient, err = ogvgrpc.NewClient(c.OgvPay); err != nil {
		panic(err)
	}
	return

}

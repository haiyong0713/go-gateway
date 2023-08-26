package favorite

import (
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-gateway/app/app-svr/app-show/interface/conf"
)

// Dao is rpc dao.
type Dao struct {
	favClient favgrpc.FavoriteClient
	conf      *conf.Config
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.favClient, err = favgrpc.NewClient(c.FavoriteGRPC); err != nil {
		panic(err)
	}
	return
}

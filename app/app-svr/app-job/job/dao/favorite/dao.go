package favorite

import (
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-gateway/app/app-svr/app-job/job/conf"
)

// Dao is account dao.
type Dao struct {
	favClient favgrpc.FavoriteClient
	conf      *conf.Config
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.favClient, err = favgrpc.NewClient(c.FavClient); err != nil {
		panic(err)
	}
	return
}

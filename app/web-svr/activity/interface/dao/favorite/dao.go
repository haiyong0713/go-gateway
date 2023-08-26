package favorite

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	favgrpc "go-main/app/community/favorite/service/api"
)

// Dao is rpc dao.
type Dao struct {
	FavClient favgrpc.FavoriteClient
	conf      *conf.Config
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf: c,
	}
	var err error
	if d.FavClient, err = favgrpc.New(c.FavoriteClient); err != nil {
		panic(err)
	}
	return
}

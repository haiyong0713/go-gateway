package favorite

import (
	"git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

// Dao struct user of Dao.
type Dao struct {
	c         *conf.Config
	favClient api.FavoriteClient
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.favClient, err = api.NewClient(nil); err != nil {
		panic(err)
	}
	return
}

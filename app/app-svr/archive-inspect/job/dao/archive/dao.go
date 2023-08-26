package archive

import (
	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-gateway/app/app-svr/archive-inspect/job/conf"
)

// Dao is redis dao.
type Dao struct {
	c            *conf.Config
	creativeGRPC creativeAPI.VideoUpOpenClient
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.creativeGRPC, err = creativeAPI.NewClient(c.CreativeClient); err != nil {
		panic(err)
	}
	return d
}

// Close dao
func (d *Dao) Close() {
}

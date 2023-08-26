package archive

import (
	"go-gateway/app/app-svr/ugc-season/job/conf"

	videoUpOpen "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

// Dao is redis dao.
type Dao struct {
	c                 *conf.Config
	videoUpOpenClient videoUpOpen.VideoUpOpenClient
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.videoUpOpenClient, err = videoUpOpen.NewClient(c.VideoUpOpenClient); err != nil {
		panic(err)
	}
	return d
}

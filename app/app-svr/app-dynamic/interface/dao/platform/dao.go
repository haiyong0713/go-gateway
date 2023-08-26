package platform

import (
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	grpcShortURL "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"
)

type Dao struct {
	c                  *conf.Config
	grpcClientShortURL grpcShortURL.ShortUrlClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClientShortURL, err = grpcShortURL.NewClient(c.ShortURLGRPC); err != nil {
		panic(err)
	}
	return d
}

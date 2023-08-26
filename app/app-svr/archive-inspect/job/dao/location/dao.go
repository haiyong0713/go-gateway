package location

import (
	"context"
	"fmt"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive-inspect/job/conf"
)

type Dao struct {
	c         *conf.Config
	locClient location.LocationClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.locClient, err = location.NewClient(c.LocationClient); err != nil {
		panic(fmt.Sprintf("locationNewClient error (%+v)", err))
	}
	return d
}

func (d *Dao) Info2Special(c context.Context, ip string) (string, error) {
	req := &location.AddrReq{Addr: ip}
	reply, err := d.locClient.Info2Special(c, req)
	if err != nil {
		log.Error("Info2WithRetry d.locClient.Info2Special error req(%+v) err(%+v)", req, err)
		return "", err
	}
	return reply.GetInfo().GetShow(), nil
}

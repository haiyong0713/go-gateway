package location

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"

	"go-gateway/app/app-svr/archive/job/conf"
)

const _maxRetryTime = 3

type Dao struct {
	c         *conf.Config
	locClient location.LocationClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.locClient, err = locationNewClient(c.LocationClient); err != nil {
		panic(fmt.Sprintf("locationNewClient error (%+v)", err))
	}
	return d
}

// NewClient new a grpc client
func locationNewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (location.LocationClient, error) {
	const (
		_appID = "location.service"
	)
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+_appID)
	if err != nil {
		return nil, err
	}
	return location.NewLocationClient(conn), nil
}

func (d *Dao) Info2WithRetry(c context.Context, ip string) (*location.InfoComplete, error) {
	for i := 0; i < _maxRetryTime; i++ {
		req := &location.AddrReq{Addr: ip}
		reply, err := d.locClient.Info2Special(c, req)
		if err != nil {
			log.Error("Info2WithRetry d.locClient.Info2Special error req(%+v) err(%+v)", req, err)
			continue
		}
		return reply.GetInfo(), nil
	}
	log.Error("日志告警 Info2WithRetry exceed max retry times d.locClient.Info2Special ip(%s) error", ip)
	return nil, nil
}

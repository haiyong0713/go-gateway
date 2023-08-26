package serial

import (
	"go-gateway/app/app-svr/app-car/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/serial/service"
)

type Dao struct {
	serialCli api.SerialServiceClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.serialCli, err = api.NewClientSerialService(c.SerialClient); err != nil {
		panic(err)
	}
	return d
}

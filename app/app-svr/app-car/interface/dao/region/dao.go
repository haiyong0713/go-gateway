package region

import (
	"fmt"

	"go-gateway/app/app-svr/app-car/interface/conf"
	dynGRPC "go-gateway/app/web-svr/dynamic/service/api/v1"
)

type Dao struct {
	dynClient dynGRPC.DynamicClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.dynClient, err = dynGRPC.NewClient(nil); err != nil {
		panic(fmt.Sprintf("dynGRPC.NewClient error (%+v)", err))
	}
	return d
}

package newyear2021

import (
	"fmt"

	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const _userSub = 100

type Dao struct {
	newyear2021DataBusPub *databus.Databus
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		newyear2021DataBusPub: databus.New(c.DataBus.NewYear2021Pub),
	}

	return d
}

func userHit(mid int64) string {
	return fmt.Sprintf("%02d", mid%_userSub)
}

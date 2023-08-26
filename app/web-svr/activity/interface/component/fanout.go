package component

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
)

var (
	ReserveFanout *fanout.Fanout
)

func initFanout(cfg *conf.Config) {
	ReserveFanout = fanout.New("reserve", fanout.Worker(5), fanout.Buffer(256))
}

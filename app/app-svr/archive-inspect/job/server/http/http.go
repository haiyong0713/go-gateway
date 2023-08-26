package http

import (
	"fmt"

	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive-inspect/job/conf"
)

// Init init http router.
func Init(c *conf.Config) {
	// init internal router
	en := bm.DefaultServer(c.BM)
	en.Ping(ping)
	en.Register(register)
	// init internal server
	if err := en.Start(); err != nil {
		panic(fmt.Sprintf("en.Start error(%+v)", err))
	}
}

// ping check server ok.
func ping(c *bm.Context) {
	// Do nothing because of check server health
}

func register(c *bm.Context) {
	c.JSON(nil, nil)
}

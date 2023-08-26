package http

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

// ping check server ok.
func ping(c *bm.Context) {
	if err := resSvc.Ping(c); err != nil {
		log.Error("resource service ping error(%v)", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

// version check server version.
func version(c *bm.Context) {
	c.JSONMap(resSvc.VersionMap(c), nil)
}

// register for discovery
func register(c *bm.Context) {
	c.JSON(nil, nil)
}

// monitor for monitorURL
func monitor(c *bm.Context) {
	resSvc.Monitor(c)
	c.JSON(nil, nil)
}

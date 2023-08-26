package http

import (
	"net/http"
	"strconv"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/web-show/interface/conf"
)

// ping check server ok.
func ping(c *bm.Context) {
	if err := resSvc.Ping(c); err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

// version check server version.
func version(c *bm.Context) {
	c.JSON(map[string]interface{}{
		"version": conf.Conf.Version,
	}, nil)
}

func grayRate(c *bm.Context) {
	params := c.Request.Form
	rateStr := params.Get("rate")
	whiteStr := params.Get("white")
	swtStr := params.Get("swt")
	if rateStr == "" && whiteStr == "" {
		res := map[string]interface{}{}
		res["rate"], res["white"], res["swt"] = resSvc.GrayRate(c)
		c.JSON(res, nil)
		return
	}
	rate, _ := strconv.ParseInt(rateStr, 10, 64)
	if rate < 0 || rate > 100 {
		rate = 0
	}
	swt, _ := strconv.ParseBool(swtStr)
	white, _ := xstr.SplitInts(whiteStr)
	resSvc.SetGrayRate(c, swt, rate, white)
}

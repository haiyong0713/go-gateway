package http

import (
	"io/ioutil"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

// fingerprint post to hakase/v1/profile.
func fingerprint(c *bm.Context) {
	var (
		params          = c.Request.Form
		header          = c.Request.Header
		mid             int64
		platform, buvid string
		err             error
	)
	if platform = params.Get("platform"); platform == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	buvid = header.Get("Buvid")
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	c.JSON(fingerPrintSvc.Fingerprint(c, platform, buvid, mid, body))
}

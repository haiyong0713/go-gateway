package http

import (
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"

	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
)

func feedInfo(c *bm.Context) {
	var (
		params                           = c.Request.Form
		aidsTmp, aids, artidsTmp, artids []int64
		epids                            []int32
		bvids                            []string
		mid                              int64
		mobiApp, device                  string
		err                              error
	)
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device = params.Get("device")
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if params.Get("aids") != "" {
		aidsTmp, err = xstr.SplitInts(params.Get("aids"))
		if err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		for _, aid := range aidsTmp {
			if aid != 0 {
				aids = append(aids, aid)
			}
		}
		if len(aids) == 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if params.Get("bvids") != "" {
		for _, bvid := range strings.Split(params.Get("bvids"), ",") {
			if bvid != "" {
				bvids = append(bvids, bvid)
			}
		}
		if len(bvids) == 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if params.Get("article_ids") != "" {
		artidsTmp, err = xstr.SplitInts(params.Get("article_ids"))
		if err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		for _, artID := range artidsTmp {
			if artID != 0 {
				artids = append(artids, artID)
			}
		}
		if len(artids) == 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if params.Get("ep_ids") != "" {
		epidsTmp, err := xstr.SplitInts(params.Get("ep_ids"))
		if err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		for _, epid := range epidsTmp {
			if epid != 0 {
				epids = append(epids, int32(epid))
			}
		}
		if len(epids) == 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	if len(aids) == 0 && len(bvids) == 0 && len(artids) == 0 && len(epids) == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(dynamicSvc.FeedInfo(c, aids, artids, epids, bvids, mobiApp, device, mid))
}

func geoCoder(c *bm.Context) {
	var req = new(dynmdl.GeoCoderReq)
	if err := c.Bind(req); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(dynamicSvc.GeoCoder(c, req.Lat, req.Lng, req.From))
}

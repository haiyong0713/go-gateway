package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"strconv"
)

const (
	_headerBuvid = "Buvid"
)

func userOperatorIP(c *bm.Context) {
	res := map[string]interface{}{}
	params := c.Request.Form
	operator := params.Get("operator")
	if operator == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	var ispOperator string
	switch operator {
	case "unicom", "mobile", "telecom":
		res["data"], ispOperator = unicomSvc.UserIPOrder(c, ip, operator)
	default:
		c.JSON(nil, ecode.RequestErr)
		return
	}
	localResult := params.Get("local_result")
	if localResult != "" {
		var mid int64
		if midInter, ok := c.Get("mid"); ok {
			mid = midInter.(int64)
		}
		header := c.Request.Header
		buvid := header.Get(_headerBuvid)
		mobiApp := params.Get("mobi_app")
		device := params.Get("device")
		platform := params.Get("platform")
		build, _ := strconv.Atoi(params.Get("build"))
		pip := params.Get("pip")
		localInfo := params.Get("local_info")
		unicomSvc.InfocIP(c, mid, buvid, mobiApp, device, platform, build, localInfo, ip, pip, localResult, ispOperator)
	}
	returnDataJSON(c, res, nil)
}

func mOperatorIP(c *bm.Context) {
	params := c.Request.Form
	ipStr := params.Get("ip")
	c.JSON(unicomSvc.UserIPInfoV2(c, ipStr), nil)
}

func operatorIPInfo(c *bm.Context) {
	ipStr := metadata.String(c, metadata.RemoteIP)
	c.JSON(unicomSvc.UserIPInfo(c, ipStr), nil)
}

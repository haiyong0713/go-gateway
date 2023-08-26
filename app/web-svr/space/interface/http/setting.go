package http

import (
	"strconv"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/space/interface/model"
)

func settingInfo(c *bm.Context) {
	var (
		mid int64
		err error
	)
	midStr := c.Request.Form.Get("mid")
	if mid, err = strconv.ParseInt(midStr, 10, 64); err != nil || mid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(spcSvc.SettingInfo(c, mid))
}

// privacySetting .
func privacySetting(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(spcSvc.PrivacySetting(c, mid), nil)
}

func privacyModify(c *bm.Context) {
	var (
		mid   int64
		field string
		value int
		err   error
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	if field = params.Get("field"); field == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	valueStr := params.Get("value")
	if value, err = strconv.Atoi(valueStr); err != nil || (value != 0 && value != 1) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, spcSvc.PrivacyModify(c, mid, field, value))
}

func privacyBatchModify(c *bm.Context) {
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	batch := make(map[string]int, len(model.PrivacyFields))
	for _, v := range model.PrivacyFields {
		if valueStr := params.Get(v); valueStr != "" {
			if value, err := strconv.Atoi(valueStr); err == nil && (value == 0 || value == 1) {
				batch[v] = value
			}
		}
	}
	outerBatch := make(map[string]int, len(model.OuterPrivacyFields))
	for _, v := range model.OuterPrivacyFields {
		if valueStr := params.Get(v); valueStr != "" {
			if value, err := strconv.Atoi(params.Get(v)); err == nil && (value == 0 || value == 1) {
				outerBatch[v] = value
			}
		}
	}
	c.JSON(nil, spcSvc.PrivacyBatchModify(c, mid, batch, outerBatch))
}

func indexOrderModify(c *bm.Context) {
	var (
		mid           int64
		indexOrderStr string
		indexOrder    []string
	)
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	if indexOrderStr = params.Get("index_order"); indexOrderStr == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	indexOrders := strings.Split(indexOrderStr, ",")
	for _, v := range indexOrders {
		i, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		if i == model.IndexOrderOfficialEvents {
			continue
		}
		if _, ok := model.IndexOrderMap[i]; !ok {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		indexOrder = append(indexOrder, v)
	}
	if len(indexOrder) != len(model.DefaultIndexOrder) {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, spcSvc.IndexOrderModify(c, mid, indexOrder))
}

func appSetting(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	req := new(struct {
		MobiApp string `form:"mobi_app"`
		Device  string `form:"device"`
	})
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(spcSvc.AppSetting(c, mid, req.MobiApp, req.Device))
}

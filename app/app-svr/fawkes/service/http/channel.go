package http

import (
	"strconv"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

func channelList(c *bm.Context) {
	var (
		size, sizeEr = strconv.Atoi(c.Request.Form.Get("ps"))
		page, _      = strconv.Atoi(c.Request.Form.Get("pn"))
		filterKey    = c.Request.Form.Get("filter_key")
	)
	if sizeEr != nil || size <= 0 || size > 20 {
		size = 20
	}
	if page < 1 {
		page = 1
	}
	c.JSON(s.AppSvr.ChannelList(c, size, page, filterKey))
}

func channelAdd(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		code, name, plate string
		isSync            bool
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if code = params.Get("code"); code == "" {
		res["message"] = "code异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if plate = params.Get("plate"); plate == "" {
		res["message"] = "plate异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	isSync, _ = strconv.ParseBool(params.Get("is_sync"))
	c.JSON(nil, s.AppSvr.ChannelAdd(c, code, name, plate, userName, appmdl.ChannelStatic, isSync))
}

func channelDelete(c *bm.Context) {
	var (
		params    = c.Request.Form
		res       = map[string]interface{}{}
		channelID int64
		err       error
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if channelID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil {
		res["message"] = "channel_id为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.ChannelDelete(c, channelID, userName))
}

// app channel list pagination version api
func appChannelListV2(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, filterKey string
		groupID           int64
		pn, ps            int
		err               error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	groupIDStr := params.Get("group_id")
	if groupIDStr == "" {
		groupID = -1
	} else if groupID, err = strconv.ParseInt(groupIDStr, 10, 64); err != nil {
		res["message"] = "groupID 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil {
		pn = 1
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		ps = 20
	}
	filterKey = params.Get("filter_key")
	c.JSON(s.AppSvr.AppChannelListV2(c, appKey, filterKey, groupID, pn, ps))
}

func appChannelList(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, filterKey string
		groupID           int64
		err               error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	groupIDStr := params.Get("group_id")
	if groupIDStr == "" {
		groupID = -1
	} else if groupID, err = strconv.ParseInt(groupIDStr, 10, 64); err != nil {
		res["message"] = "groupID 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	filterKey = params.Get("filter_key")
	c.JSON(s.AppSvr.AppChannelList(c, appKey, filterKey, groupID))
}

func appChannelAdd(c *bm.Context) {
	var (
		params                    = c.Request.Form
		res                       = map[string]interface{}{}
		channelID, groupID        int64
		chType                    int
		code, name, plate, appKey string
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	chType, err := strconv.Atoi(params.Get("type"))
	if err != nil {
		res["message"] = "type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// nolint:gomnd
	if chType == 1 {
		if channelID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil || channelID < 1 {
			res["message"] = "channel_id异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else if chType == 2 {
		if code = params.Get("code"); code == "" {
			res["message"] = "code异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if name = params.Get("name"); code == "" || name == "" {
			res["message"] = "name异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if plate = params.Get("plate"); plate == "" {
			res["message"] = "plate异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if groupID, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
			res["message"] = "group_id异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		res["message"] = "type异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelAdd(c, chType, channelID, groupID, code, name, plate, userName, appKey))
}

func appChannelDelete(c *bm.Context) {
	var (
		params    = c.Request.Form
		res       = map[string]interface{}{}
		appKey    string
		channelID int64
		err       error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channelID, err = strconv.ParseInt(params.Get("channel_id"), 10, 64); err != nil {
		res["message"] = "channel_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelDelete(c, appKey, channelID))
}
func appChannelGroupRelate(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		groudID          int64
		appChannelIDs    []int64
		appChannelIDsStr string
		err              error
	)
	if appChannelIDsStr = params.Get("app_channel_ids"); appChannelIDsStr == "" {
		res["message"] = "app_channel_ids 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for _, acID := range strings.Split(appChannelIDsStr, ",") {
		appChannelID, err := strconv.ParseInt(acID, 10, 64)
		if err != nil {
			res["message"] = "app_channel_ids 异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		appChannelIDs = append(appChannelIDs, appChannelID)
	}
	if groudID, err = strconv.ParseInt(params.Get("group_id"), 10, 64); err != nil {
		res["message"] = "group_id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSONMap(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelGroupRelate(c, appChannelIDs, groudID, userName))
}

func appChannelGroupList(c *bm.Context) {
	var (
		params            = c.Request.Form
		res               = map[string]interface{}{}
		appKey, filterKey string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	filterKey = params.Get("filter_key")
	c.JSON(s.AppSvr.AppChannelGroupList(c, appKey, filterKey))
}

func appChannelGroupAdd(c *bm.Context) {
	var (
		params                                          = c.Request.Form
		res                                             = map[string]interface{}{}
		appKey, name, description, qaOwner, marketOwner string
		autoPushCdn, isAutoGen, priority                int64
		err                                             error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name不能为空"
		return
	}
	if autoPushCdn, err = strconv.ParseInt(params.Get("auto_push_cdn"), 10, 64); err != nil {
		res["message"] = "auto_push_cdn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isAutoGen, err = strconv.ParseInt(params.Get("is_auto_gen"), 10, 64); err != nil {
		res["message"] = "is_auto_gen异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	qaOwner = params.Get("qa_owner")
	marketOwner = params.Get("market_owner")
	if isAutoGen == 1 {
		if len(qaOwner) == 0 || len(marketOwner) == 0 {
			res["message"] = "选择自动生成渠道包，需要提供市场负责人和测试负责人姓名"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if p := params.Get("priority"); len(p) != 0 {
		if priority, err = strconv.ParseInt(p, 10, 64); err != nil {
			res["message"] = "priority异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	description = params.Get("description")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSONMap(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelGroupAdd(c, appKey, name, description, userName, autoPushCdn, isAutoGen, qaOwner, marketOwner, priority))
}

func appChannelGroupUpdate(c *bm.Context) {
	var (
		params                                  = c.Request.Form
		res                                     = map[string]interface{}{}
		id, autoPushCdn, isAutoGen, priority    int64
		name, description, qaOwner, marketOwner string
		err                                     error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "name不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if autoPushCdn, err = strconv.ParseInt(params.Get("auto_push_cdn"), 10, 64); err != nil {
		res["message"] = "auto_push_cdn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if isAutoGen, err = strconv.ParseInt(params.Get("is_auto_gen"), 10, 64); err != nil {
		res["message"] = "is_auto_gen异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	qaOwner = params.Get("qa_owner")
	marketOwner = params.Get("market_owner")
	if isAutoGen == 1 {
		if len(qaOwner) == 0 || len(marketOwner) == 0 {
			res["message"] = "选择自动生成渠道包，需要提供市场负责人和测试负责人姓名"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if p := params.Get("priority"); len(p) != 0 {
		if priority, err = strconv.ParseInt(p, 10, 64); err != nil {
			res["message"] = "priority异常"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	description = params.Get("description")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || username == "" {
		c.JSONMap(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelGroupUpdate(c, id, name, description, userName, autoPushCdn, isAutoGen, qaOwner, marketOwner, priority))
}

func appChannelGroupDel(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		id     int64
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || username == "" {
		c.JSONMap(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.AppSvr.AppChannelGroupDel(c, id, userName))
}

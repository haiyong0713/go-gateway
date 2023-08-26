package http

import (
	"strconv"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func list(c *bm.Context) {
	var (
		params          = c.Request.Form
		hotWord, filter string
		state           int
		err             error
	)
	hotWord = params.Get("hot_word")
	if params.Get("state") == "" {
		state = -1
	} else {
		if state, err = strconv.Atoi(params.Get("state")); err != nil {
			state = -1
		}
	}
	if filter = params.Get("filter"); filter == "all" {
		filter = ""
	}
	c.JSON(aggSvc2.List(c, hotWord, filter, state))
}

func operate(c *bm.Context) {
	var (
		params        = c.Request.Form
		res           = map[string]interface{}{}
		plat, hotword string
		state         int
		err           error
	)
	if plat = params.Get("plat"); plat == "" {
		res["message"] = "plat不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if hotword = params.Get("hot_word"); hotword == "" {
		res["message"] = "hot_word不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if state, err = strconv.Atoi(params.Get("state")); err != nil {
		res["message"] = "state"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, aggSvc2.Operate(c, plat, hotword, state))
}

func add(c *bm.Context) {
	var (
		params  = c.Request.Form
		res     = map[string]interface{}{}
		hotword string
	)
	if hotword = params.Get("hot_word"); hotword == "" {
		res["message"] = "hot_word不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, aggSvc2.Add(c, hotword))
}

func save(c *bm.Context) {
	var (
		params                 = c.Request.Form
		res                    = map[string]interface{}{}
		id                     int64
		title, subTitle, cover string
		err                    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if title = params.Get("title"); title == "" {
		res["message"] = "入口文案不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if subTitle = params.Get("sub_title"); subTitle == "" {
		res["message"] = "副标题不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if cover = params.Get("cover"); cover == "" {
		res["message"] = "页面头图不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, aggSvc2.Save(c, id, title, subTitle, cover))
}

func view(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		id          int64
		sort, order string
		err         error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	sort = params.Get("sort")
	order = params.Get("order")
	c.JSON(aggSvc2.Materiels(c, id, sort, order))
}

func viewAdd(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		hotID  int64
		oids   []string
		err    error
	)
	oidParam := params.Get("oids")
	if oidParam == "" {
		res["message"] = "oids异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	oids = strings.Split(oidParam, ",")
	if hotID, err = strconv.ParseInt(params.Get("hot_id"), 10, 64); err != nil {
		res["message"] = "hot_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, aggSvc2.ViewAdd(c, hotID, oids))
}

func viewOperate(c *bm.Context) {
	var (
		params     = c.Request.Form
		res        = map[string]interface{}{}
		hotID, oid int64
		state      int
		source     string
		err        error
	)
	if hotID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "hot_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if oid, err = strconv.ParseInt(params.Get("oid"), 10, 64); err != nil {
		res["message"] = "oid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if state, err = strconv.Atoi(params.Get("state")); err != nil {
		res["message"] = "state异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if source = params.Get("source"); source == "" {
		res["message"] = "source不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, aggSvc2.ViewOperate(c, hotID, oid, state, source))
}

func tagAdd(c *bm.Context) {
	var (
		params       = c.Request.Form
		res          = map[string]interface{}{}
		hotID, tagID int64
		err          error
	)
	if hotID, err = strconv.ParseInt(params.Get("hot_id"), 10, 64); err != nil {
		res["message"] = "hot_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if tagID, err = strconv.ParseInt(params.Get("tag_id"), 10, 64); err != nil {
		res["message"] = "tag_id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, aggSvc2.TagAdd(c, hotID, tagID))
}

func tagDel(c *bm.Context) {
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
	c.JSON(nil, aggSvc2.TagDel(c, id))
}

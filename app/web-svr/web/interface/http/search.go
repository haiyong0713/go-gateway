package http

import (
	"go-common/library/log"
	"net/http"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web/interface/model/search"
)

const (
	_searchForbidCode   = -110
	_searchNoResultCode = -111
)

func searchAll(c *bm.Context) {
	var (
		mid   int64
		buvid string
		err   error
	)
	v := new(search.SearchAllArg)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Pn <= 0 {
		v.Pn = 1
	}
	singleColumnStr := c.Request.Form.Get("single_column")
	if v.SingleColumn, err = strconv.Atoi(singleColumnStr); err != nil {
		v.SingleColumn = -1
	}
	var ck *http.Cookie
	if ck, err = c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data, err := webSvc.SearchAll(c, mid, v, buvid, c.Request.Header.Get("User-Agent"), "")
	if err != nil {
		if ecode.Cause(err).Code() != _searchForbidCode && ecode.Cause(err).Code() != _searchNoResultCode {
			log.Info("日志告警 SearchAll降级 data:%+v,err:%+v", v, err)
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)

}

func searchAllV2(c *bm.Context) {
	v := new(search.SearchAllArg)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.Pn <= 0 {
		v.Pn = 1
	}
	singleColumnStr := c.Request.Form.Get("single_column")
	var err error
	if v.SingleColumn, err = strconv.Atoi(singleColumnStr); err != nil {
		v.SingleColumn = -1
	}
	var buvid string
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	data, err := webSvc.SearchAllV2(c, mid, v, buvid, c.Request.Header.Get("User-Agent"), "")
	if err != nil {
		if ecode.Cause(err).Code() != _searchForbidCode && ecode.Cause(err).Code() != _searchNoResultCode {
			log.Info("日志告警 SearchAllV2降级 data:%+v,err:%+v", v, err)
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func searchByType(c *bm.Context) {
	var (
		mid   int64
		buvid string
		err   error
	)
	v := new(search.SearchTypeArg)
	if err = c.Bind(v); err != nil {
		return
	}
	// 注意：兼容逻辑！
	switch v.Platform {
	case "pc", "":
		v.Platform = "web"
	}
	singleColumnStr := c.Request.Form.Get("single_column")
	if v.SingleColumn, err = strconv.Atoi(singleColumnStr); err != nil {
		v.SingleColumn = -1
	}
	var ck *http.Cookie
	if ck, err = c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ua := c.Request.Header.Get("User-Agent")
	var data interface{}
	switch v.SearchType {
	case search.SearchTypeBangumi, search.SearchTypeMovie:
		data, err = webSvc.SearchPGC(c, mid, v, buvid, ua)
	case search.SearchTypeVideo:
		data, err = webSvc.SearchVideo(c, mid, v, buvid, ua)
	case search.SearchTypeUser:
		data, err = webSvc.SearchUser(c, mid, v, buvid, ua)
	default:
		data, err = webSvc.SearchByType(c, mid, v, buvid, ua)
	}
	if err != nil {
		if ecode.Cause(err).Code() != _searchForbidCode && ecode.Cause(err).Code() != _searchNoResultCode {
			log.Info("日志告警 SearchByType2降级 data:%+v,err:%+v", v, err)
			err = ecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func searchRec(c *bm.Context) {
	var (
		mid   int64
		buvid string
		err   error
	)
	v := new(struct {
		Pn         int    `form:"page" default:"1" validate:"min=1"`
		Ps         int    `form:"pagesize" default:"5" validate:"min=1"`
		Keyword    string `form:"keyword"`
		FromSource string `form:"from_source"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.SearchRec(c, mid, v.Pn, v.Ps, v.Keyword, v.FromSource, buvid, c.Request.Header.Get("User-Agent")))
}

func searchDefault(c *bm.Context) {
	var (
		mid   int64
		buvid string
		err   error
	)
	v := new(struct {
		FromSource string `form:"from_source"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(webSvc.SearchDefault(c, mid, v.FromSource, buvid, c.Request.Header.Get("User-Agent")))
}

func upRec(c *bm.Context) {
	v := new(search.SearchUpRecArg)
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if v.Buvid == "" {
		if ck, err := c.Request.Cookie("buvid3"); err == nil {
			v.Buvid = ck.Value
		}
	}
	c.JSON(webSvc.UpRec(c, mid, v))
}

func searchEgg(c *bm.Context) {
	v := new(struct {
		EggID int64 `form:"egg_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(webSvc.SearchEgg(c, v.EggID))
}

func searchGameInfo(c *bm.Context) {
	v := new(struct {
		GameBaseID int64 `form:"game_base_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(webSvc.SearchGameInfo(c, v.GameBaseID))
}

func searchSquare(c *bm.Context) {
	v := &search.SquareArg{}
	if err := c.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	var buvid string
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	c.JSON(webSvc.SearchSquare(c, mid, buvid, v.Limit, v.IsInner, v.Platform))
}

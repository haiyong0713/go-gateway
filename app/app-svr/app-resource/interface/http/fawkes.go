package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	fkmdl "go-gateway/app/app-svr/app-resource/interface/model/fawkes"
)

func laserReport(c *bm.Context) {
	var (
		params = c.Request.Form
		taskID int64
		status int
		url    string
		err    error
	)
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	url = params.Get("url")
	if status == fkmdl.StatusUpSuccess && url == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	errMsg := params.Get("error_msg")
	mobiApp := params.Get("mobi_app")
	build := params.Get("build")
	md5 := params.Get("md5")
	rawUposUri := params.Get("raw_upos_uri")
	c.JSON(nil, fkSvc.LaserReport(c, taskID, status, url, errMsg, mobiApp, build, md5, rawUposUri))
}

func laserReport2(c *bm.Context) {
	var (
		params             = c.Request.Form
		mid, mid2, taskID  int64
		appkey, buvid, url string
		status             int
	)
	if appkey = params.Get("app_key"); appkey == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	// 从请求参数中读取mid
	mid2, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	if mid == 0 && mid2 != 0 {
		mid = mid2
	}
	buvid = params.Get("buvid")
	if mid == 0 && buvid == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	url = params.Get("url")
	errMsg := params.Get("error_msg")
	mobiApp := params.Get("mobi_app")
	build := params.Get("build")
	status, _ = strconv.Atoi(params.Get("status"))
	taskID, _ = strconv.ParseInt(params.Get("task_id"), 10, 64)
	md5 := params.Get("md5")
	rawUposUri := params.Get("raw_upos_uri")
	c.JSON(fkSvc.LaserReport2(c, appkey, buvid, url, errMsg, mobiApp, build, md5, rawUposUri, mid, taskID, status))
}

func laserReportSilence(c *bm.Context) {
	var (
		params = c.Request.Form
		taskID int64
		status int
		url    string
		err    error
	)
	if taskID, err = strconv.ParseInt(params.Get("task_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if status, err = strconv.Atoi(params.Get("status")); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	url = params.Get("url")
	if status == fkmdl.StatusUpSuccess && url == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	errMsg := params.Get("error_msg")
	mobiApp := params.Get("mobi_app")
	build := params.Get("build")
	c.JSON(nil, fkSvc.LaserReportSilence(c, taskID, status, url, errMsg, mobiApp, build))
}

func laserCmdReport(c *bm.Context) {
	v := new(struct {
		TaskID     int64  `form:"task_id"`
		Status     int    `form:"status"`
		MobiApp    string `form:"mobi_app"`
		Build      string `form:"build"`
		URL        string `form:"url"`
		ErrorMsg   string `form:"error_msg"`
		Result     string `form:"result"`
		Md5        string `form:"md5"`
		RawUposUri string `form:"raw_upos_uri"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, fkSvc.LaserCmdReport(c, v.TaskID, v.Status, v.MobiApp, v.Build, v.URL, v.ErrorMsg, v.Result, v.Md5, v.RawUposUri))
}

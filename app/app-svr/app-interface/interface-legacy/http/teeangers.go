package http

import (
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	rpt "go-common/library/queue/databus/report"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"
)

func teenagersPwd(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	v := new(space.TeenagersPwdParam)
	if err := c.Bind(v); err != nil {
		return
	}
	if err := rpt.User(&rpt.UserInfo{
		Mid:      mid,
		Business: 92,
		Action:   "teenagers_pwd_set",
		Ctime:    time.Now(),
		IP:       metadata.String(c, metadata.RemoteIP),
		Buvid:    c.Request.Header.Get("Buvid"),
		Content: map[string]interface{}{
			"device_model": v.DeviceModel,
			"password":     v.Pwd,
		},
	}); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

func teenagersStatus(ctx *bm.Context) {
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ctx.JSON(teenSvr.Status(ctx, mid))
}

// nolint:biligowordcheck
func teenagersUpdate(ctx *bm.Context) {
	v := new(usermodel.UpdateTeenagerReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if mid == 0 && (v.MobiApp == "" || v.DeviceToken == "") {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "未登录时mobi_app和device_token不能为空"))
		return
	}
	if v.Sync && (mid == 0 || v.MobiApp == "" || v.DeviceToken == "") {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "同步时mid和mobi_app和device_token不能为空"))
		return
	}
	// 网关强制端上更新，端上再同步时只更新设备状态
	if v.From == usermodel.PwdFromForcedSync {
		mid = 0
		v.Sync = false
	}
	err := teenSvr.UpdateTeenager(ctx, v, mid)
	ctx.JSON(nil, err)
}

func modeStatus(ctx *bm.Context) {
	var v struct {
		MobiApp     string `form:"mobi_app"`
		DeviceToken string `form:"device_token"`
	}
	if err := ctx.Bind(&v); err != nil {
		return
	}
	switch v.MobiApp {
	case "android_comic", "iphone_comic", "ipad_comic":
		data := []*usermodel.UserModel{
			{Mode: "teenagers", Status: 0},
			{Mode: "lessons", Status: 0},
		}
		ctx.JSON(data, nil)
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if mid == 0 && v.MobiApp == "" {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "未登录时mobi_app不能为空"))
		return
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	buvid := ctx.Request.Header.Get("Buvid")
	ctx.JSON(teenSvr.UserModel(ctx, v.MobiApp, v.DeviceToken, mid, ip, buvid))
}

// nolint:biligowordcheck
func lessonsUpdate(ctx *bm.Context) {
	var param struct {
		Pwd           string `form:"pwd"`
		LessonsStatus int    `form:"lessons_status" validate:"min=0,max=1"`
		From          string `form:"from"`
		DeviceModel   string `form:"device_model"`
	}
	if err := ctx.Bind(&param); err != nil {
		return
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	var operation int
	switch param.LessonsStatus {
	case 0:
		if op, ok := usermodel.From2OpOfQuit[param.From]; ok {
			operation = op
		}
	case 1:
		if op, ok := usermodel.From2OpOfOpen[param.From]; ok {
			operation = op
		}
	}
	_, err := teenSvr.UserModelUpdate(ctx, "", "", mid, param.Pwd, "", param.LessonsStatus, false, usermodel.LessonsModel, operation, usermodel.PwdTypeSelf)
	rptFunc := func() (err error) {
		if err = rpt.User(&rpt.UserInfo{
			Mid:      mid,
			Business: 92,
			Action:   "lessons_pwd_set",
			Ctime:    time.Now(),
			IP:       metadata.String(ctx, metadata.RemoteIP),
			Buvid:    ctx.Request.Header.Get("Buvid"),
			Content: map[string]interface{}{
				"device_model": param.DeviceModel,
				"password":     param.Pwd,
			},
		}); err != nil {
			log.Error("%+v", err)
		}
		return
	}
	if err == nil && param.LessonsStatus == 1 {
		if err := rptFunc(); err != nil {
			go func() {
				for i := 0; i < 3; i++ {
					if err := rptFunc(); err == nil {
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()
		}
	}
	ctx.JSON(nil, err)
}

//nolint:gosimple
func setTimer(ctx *bm.Context) {
	param := usermodel.AntiAddiction{}
	if err := ctx.Bind(&param); err != nil {
		return
	}
	if param.UseTime < 0 {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "设置时间不能小于0"))
		return
	}
	if midInter, ok := ctx.Get("mid"); ok {
		param.MID = midInter.(int64)
	}
	param.Day = int64(time.Now().YearDay())
	ctx.JSON(nil, teenSvr.SetAntiAddictionTime(ctx, &param))
	return
}

//nolint:gosimple
func getTimer(ctx *bm.Context) {
	param := usermodel.AntiAddiction{}
	if err := ctx.Bind(&param); err != nil {
		return
	}
	if midInter, ok := ctx.Get("mid"); ok {
		param.MID = midInter.(int64)
	}
	param.Day = int64(time.Now().YearDay())
	res, err := teenSvr.GetAntiAddictionTime(ctx, &param)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	data := map[string]interface{}{}
	data["time"] = res
	ctx.JSON(data, nil)
	return
}

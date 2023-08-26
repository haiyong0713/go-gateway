package http

import (
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	mdl "go-gateway/app/app-svr/app-resource/interface/model/module"
)

const (
	_md5Header = "Params_md5"
)

func list(c *bm.Context) {
	header := c.Request.Header
	params := c.Request.Form
	buvid := header.Get("Buvid")
	fawkesAppKey := header.Get("App-key")
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	envStr := params.Get("env")
	if envStr != mdl.EnvRelease && envStr != mdl.EnvTest && envStr != mdl.EnvDefault {
		envStr = mdl.EnvRelease
	}
	build, _ := strconv.Atoi(params.Get("build"))
	sysver, _ := strconv.Atoi(params.Get("sysver"))
	level, _ := strconv.Atoi(params.Get("level"))
	scale, _ := strconv.Atoi(params.Get("scale"))
	arch, _ := strconv.Atoi(params.Get("arch"))
	ts, _ := strconv.Atoi(params.Get("param1"))
	rPoolName := params.Get("resource_pool_name")
	// params
	verlist := params.Get("verlist")
	var versions []*mdl.Versions
	if verlist != "" {
		if err := json.Unmarshal([]byte(verlist), &versions); err != nil {
			log.Error("http list() json.Unmarshal(%s) mobile_app(%s) device(%s) build(%d) error(%v)", verlist, mobiApp, device, build, err)
		}
	}
	now := time.Now()
	md5 := mdl.ModuleMd5(mobiApp, device, build, ts, c.Request.Header.Get(_headerBuvid), c.Request.PostForm)
	c.Writer.Header().Set(_md5Header, md5)
	data := modSvc.HTTPList(c, fawkesAppKey, buvid, mid, device, rPoolName, envStr, build, sysver, level, scale, arch, versions, now)
	c.JSON(data, nil)
}

func module(c *bm.Context) {
	header := c.Request.Header
	params := c.Request.Form
	buvid := header.Get("Buvid")
	fawkesAppKey := header.Get("App-key")
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	mobiApp := params.Get("mobi_app")
	// 目前rpc接口无需所app-key的兼容，但是由于历史原因http接口需要作app-key的兼容
	// 1. 针对较早版本在header中没有app-key的情况，统一配置默认值，具体如下
	// mobi_app = android, app_key = android
	// mobi_app = iphone, app_key = iphone
	// mobi_app = ipad, app_key = ipad2
	// 2. 针对HD1，若在header中传递了app_key参数，强制改为ipad2。换句话说mobi_app = ipad， app_key只能为ipad2
	// 3. 老版国际版是从年初的6.0版本从粉板切出的，所以相关http接口参数条件是完备的，不用兼容
	// 4. 新国际版暂不考虑，因为更换了mobi_app
	// 5. android_b暂时忽略
	if fawkesAppKey == "" {
		switch mobiApp {
		case "android":
			fawkesAppKey = "android"
		case "iphone":
			fawkesAppKey = "iphone"
		}
	}
	if mobiApp == "ipad" {
		fawkesAppKey = "ipad2"
	}
	device := params.Get("device")
	rPoolName := params.Get("resource_pool_name")
	rName := params.Get("resource_name")
	envStr := params.Get("env")
	if envStr != mdl.EnvRelease && envStr != mdl.EnvTest && envStr != mdl.EnvDefault {
		envStr = mdl.EnvRelease
	}
	build, _ := strconv.Atoi(params.Get("build"))
	sysver, _ := strconv.Atoi(params.Get("sysver"))
	level, _ := strconv.Atoi(params.Get("level"))
	scale, _ := strconv.Atoi(params.Get("scale"))
	arch, _ := strconv.Atoi(params.Get("arch"))
	ver, _ := strconv.Atoi(params.Get("ver"))
	ts, _ := strconv.Atoi(params.Get("param1"))
	if rPoolName == "" || rName == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	now := time.Now()
	md5 := mdl.ModuleMd5(mobiApp, device, build, ts, c.Request.Header.Get(_headerBuvid), c.Request.PostForm)
	c.Writer.Header().Set(_md5Header, md5)
	data, err := modSvc.HTTPModule(c, fawkesAppKey, buvid, mid, device, rPoolName, rName, envStr, ver, build, sysver, level, scale, arch, now)
	c.JSON(data, err)
}

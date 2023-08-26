package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	fkmdl "go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	"go-gateway/app/app-svr/app-resource/interface/model/version"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"google.golang.org/protobuf/proto"
)

// getVersion get version
func getVersion(c *bm.Context) {
	params := c.Request.Form
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	data, err := verSvc.Version(plat)
	c.JSON(data, err)
}

// versionUpdate get versionUpdate
func versionUpdate(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
	)
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	buildStr := params.Get("build")
	channel := params.Get("channel")
	sdkStr := params.Get("sdkint")
	platModel := params.Get("model")
	oldID := params.Get("old_id")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	buvid := header.Get(_headerBuvid)
	// mobiApp not equal to android or mobiApp is null,default android
	if !model.IsAndroid(plat) {
		plat = model.PlatAndroid
	}
	if feature.GetBuildLimit(c, config.Feature.FeatureBuildLimit.UpdateAndroidB, &feature.OriginResutl{
		MobiApp:    mobiApp,
		Device:     device,
		Build:      int64(build),
		BuildLimit: (plat == model.PlatAndroid) && (build >= 591000 && build <= 599000),
	}) {
		plat = model.PlatAndroidB
	}
	data, err := verSvc.VersionUpdate(build, plat, buvid, sdkStr, channel, platModel, oldID)
	if err != nil {
		c.JSON(nil, ecode.NotModified)
		return
	}
	c.JSON(data, nil)
}

// versionUpdate get versionUpdate
func versionUpdatePb(c *bm.Context) {
	var (
		params = c.Request.Form
		header = c.Request.Header
	)
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	buildStr := params.Get("build")
	channel := params.Get("channel")
	sdkStr := params.Get("sdkint")
	platModel := params.Get("model")
	oldID := params.Get("old_id")
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	buvid := header.Get(_headerBuvid)
	// mobiApp not equal to android or mobiApp is null,default android
	if !model.IsAndroid(plat) {
		plat = model.PlatAndroid
	}
	data, err := verSvc.VersionUpdate(build, plat, buvid, sdkStr, channel, platModel, oldID)
	if err != nil {
		c.JSON(nil, ecode.NotModified)
		return
	}
	//nolint:gosec
	size, _ := strconv.Atoi(data.Size)
	resPb := &version.VerUpdate{
		Ver:     *proto.String(data.Version),
		Build:   *proto.Int32(int32(data.Build)),
		Info:    *proto.String(data.Desc),
		Size:    *proto.Int32(int32(size)),
		Url:     *proto.String(data.Url),
		Hash:    *proto.String(data.MD5),
		Policy:  *proto.Int32(int32(data.Policy)),
		IsForce: *proto.Int32(int32(data.IsForce)),
		Mtime:   *proto.Int64(data.Mtime.Time().Unix()),
	}
	c.JSON(resPb, nil)
}

func versionSo(c *bm.Context) {
	params := c.Request.Form
	name := params.Get("name")
	seedStr := params.Get("seed")
	buildStr := params.Get("build")
	sdkStr := params.Get("sdkint")
	model := params.Get("model")
	seed, err := strconv.Atoi(seedStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	sdkint, _ := strconv.Atoi(sdkStr)
	data, err := verSvc.VersionSo(build, seed, sdkint, name, model)
	if err != nil {
		c.JSON(nil, ecode.NotModified)
		return
	}
	c.JSON(data, nil)
}

// versionRn get versionUpdate
func versionRn(c *bm.Context) {
	params := c.Request.Form
	deploymentKey := params.Get("deployment_key")
	bundleID := params.Get("bundle_id")
	version := params.Get("base_version")
	data, err := verSvc.VersionRn(version, deploymentKey, bundleID)
	c.JSON(data, err)
}

// fawkesUpgrade fawkes upgrade
func fawkesUpgrade(c *bm.Context) {
	var (
		params                                                              = c.Request.Form
		header                                                              = c.Request.Header
		res                                                                 = map[string]interface{}{}
		fawkesAppKey, fawkesEnv, mobiApp, version, network, channel, system string
		build, buildID                                                      int64
		data                                                                *fawkes.Item
		err                                                                 error
	)
	fawkesAppKey = header.Get("App-key")
	if fawkesEnv = header.Get("Env"); fawkesEnv == "" {
		fawkesEnv = "prod"
	}
	buvid := header.Get("Buvid")
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		res["message"] = "mobi_app异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fawkesAppKey != "" {
		mobiApp = fawkesAppKey
	}
	if version = params.Get("vn"); version == "" {
		res["message"] = "vn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if build, err = strconv.ParseInt(params.Get("build"), 10, 64); err != nil {
		res["message"] = "build异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("sn"), 10, 64); err != nil {
		res["message"] = "sn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if network = params.Get("nt"); network == "" {
		res["message"] = "nt异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if channel = params.Get("channel"); channel == "" {
		res["message"] = "channel异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if system = params.Get("ov"); system == "" {
		res["message"] = "ov异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if data, err = fkSvc.Upgrade(c, mobiApp, fawkesEnv, version, build, buildID, buvid, network, channel, system); err != nil || data == nil {
		c.JSON(nil, ecode.NotModified)
		return
	}
	c.JSON(data, err)
}

func fawkesUpgradeIOS(ctx *bm.Context) {
	v := &fawkes.UpgradeIOSParam{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	header := ctx.Request.Header
	v.FawkesAppKey = header.Get("App-key")
	if v.FawkesAppKey == "" {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	fawkesEnv := header.Get("Env")
	if fawkesEnv == "" {
		fawkesEnv = "prod"
	}
	v.FawkesEnv = fawkesEnv
	v.Buvid = header.Get("Buvid")
	v.IP = metadata.String(ctx, metadata.RemoteIP)
	data, err := fkSvc.UpgradeIOS(ctx, v)
	if err != nil {
		log.Error("%+v", err)
		ctx.JSON(nil, ecode.Cause(err))
		return
	}
	ctx.JSON(data, nil)
}

func fawkesHfUpgrade(c *bm.Context) {
	var (
		params                                             = c.Request.Form
		header                                             = c.Request.Header
		mobiApp, env, vn, deviceID, channel, md5, fkAppKey string
		build, sn, ov                                      int64
		resp                                               *fkmdl.HfUpgradeInfo
		err                                                error
	)
	md5 = params.Get("md5")
	fkAppKey = header.Get("App-key")
	if env = header.Get("Env"); env == "" || env != "prod" {
		//nolint:ineffassign
		env = "test"
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "mobi_app异常"))
		return
	}
	if fkAppKey != "" {
		mobiApp = fkAppKey
	}
	if env = params.Get("env"); env == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "env异常"))
		return
	}
	if vn = params.Get("vn"); vn == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "vn异常"))
		return
	}
	if deviceID = params.Get("deviceid"); deviceID == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "deviceid异常"))
		return
	}
	if channel = params.Get("channel"); channel == "" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "channel异常"))
		return
	}
	if build, err = strconv.ParseInt(params.Get("build"), 10, 64); err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "build异常"))
		return
	}
	if sn, err = strconv.ParseInt(params.Get("sn"), 10, 64); err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "sn异常"))
		return
	}
	if ov, err = strconv.ParseInt(params.Get("ov"), 10, 64); err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "ov异常"))
		return
	}
	if fkAppKey == "android" && (ov == 21 || ov == 22) && sn == 4190269 {
		resp = &fawkes.HfUpgradeInfo{
			Version:     "6.9.0",
			VersionCode: 4190269,
			PatchURL:    "https://dl.hdslb.com/mobile/pack/android/4190269/hotfix/562/patch_signed.apk",
			PatchMd5:    "dfb5e75b1103cbf7857d8db1e76205e7",
		}
		c.JSON(resp, nil)
	} else if fkAppKey == "android64" && (ov == 21 || ov == 22) && sn == 4190272 {
		resp = &fawkes.HfUpgradeInfo{
			Version:     "6.9.0",
			VersionCode: 4190272,
			PatchURL:    "https://dl.hdslb.com/mobile/pack/android64/4190272/hotfix/563/patch_signed.apk",
			PatchMd5:    "622404e5cb52ad2993f5c10dc5f3dde2",
		}
		c.JSON(resp, nil)
	} else {
		if resp, err = fkSvc.HfUpgrade(c, mobiApp, env, vn, deviceID, channel, md5, build, sn, ov); err != nil {
			c.JSON(nil, ecode.NotModified)
			return
		}
		c.JSON(resp, nil)
	}
}

func apkList(c *bm.Context) {
	header := c.Request.Header
	param := &fkmdl.ApkListParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	fkAppKey := header.Get("App-key")
	if fkAppKey == "" {
		res := map[string]interface{}{}
		res["message"] = "App-key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	param.AppKey = fkAppKey
	param.Buvid = header.Get("Buvid")
	env := header.Get("Env")
	if env == "" {
		env = "prod"
	}
	param.Env = env
	res, err := fkSvc.ApkList(c, param)
	if err != nil {
		c.JSON(nil, ecode.NotModified)
		return
	}
	c.JSON(res, nil)
}

func tribeList(c *bm.Context) {
	header := c.Request.Header
	param := &fkmdl.TribeListParam{}
	if err := c.Bind(param); err != nil {
		log.Error("TribeList ParamFormat error(%v)", err)
		return
	}
	fkAppKey := header.Get("App-key")
	if fkAppKey == "" {
		log.Error("TribeList app_key is empty")
		res := map[string]interface{}{}
		res["message"] = "App-key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	param.AppKey = fkAppKey
	param.Buvid = header.Get("Buvid")
	env := header.Get("Env")
	if env == "" {
		env = "prod"
	}
	param.Env = env
	res, err := fkSvc.TribeAllList(c, param)
	if err != nil {
		log.Error("fkSvc.TribeAllList error(%v)", err)
		c.JSON(nil, ecode.NotModified)
		return
	}
	c.JSON(res, nil)
}

func testFlight(c *bm.Context) {
	param := &fkmdl.TestFlightParam{}
	if err := c.Bind(param); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	header := c.Request.Header
	env := header.Get("Env")
	if env == "" {
		env = "prod"
	}
	param.Buvid = header.Get("Buvid")
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(fkSvc.TestFlight(param, env, mid))
}

func fawkesTinyUpgrade(c *bm.Context) {
	var (
		params                                = c.Request.Form
		header                                = c.Request.Header
		res                                   = map[string]interface{}{}
		fawkesAppKey, fawkesEnv, mobiApp, abi string
		data                                  *fawkes.Item
		err                                   error
	)
	fawkesAppKey = header.Get("App-key")
	if fawkesEnv = header.Get("Env"); fawkesEnv == "" {
		fawkesEnv = "prod"
	}
	if mobiApp = params.Get("mobi_app"); mobiApp == "" {
		res["message"] = "mobi_app异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fawkesAppKey != "" {
		mobiApp = fawkesAppKey
	}
	if abi = params.Get("abi"); abi == "" {
		res["message"] = "abi不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if data, err = fkSvc.UpgradeTiny(c, mobiApp, fawkesEnv, abi); err != nil || data == nil {
		c.JSON(nil, ecode.NotModified)
		return
	}
	c.JSON(data, err)
}

package http

import (
	"encoding/json"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/archive"
)

const _maxAids = 50

// arcInfo write the archive data.
func arcInfo(c *bm.Context) {
	var (
		err error
		aid int64
	)
	params := c.Request.Form
	aidStr := params.Get("aid")
	// check params
	aid, err = strconv.ParseInt(aidStr, 10, 64)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(arcSvc.ArcWithStat(c, &api.ArcRequest{Aid: aid}))
}

// archives write the archives data.
func archives(c *bm.Context) {
	params := c.Request.Form
	aidsStr := params.Get("aids")
	mid, _ := strconv.ParseInt(params.Get("mid"), 10, 64)
	mobiApp := params.Get("mobi_app")
	device := params.Get("device")
	// check params
	aids, err := xstr.SplitInts(aidsStr)
	if err != nil {
		log.Error("query aids(%s) split error(%v)", aidsStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if params.Get("appkey") == "fb06a25c6338edbc" && len(aids) > 50 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if len(aids) > _maxAids {
		log.Error("Too many Args aids(%d) caller(%s)", len(aids), params.Get("appkey"))
	}
	c.JSON(arcSvc.Archives3(c, aids, mid, mobiApp, device))
}

// archivesWithPlayer write the archives data.
func archivesWithPlayer(c *bm.Context) {
	params := c.Request.Form
	aidsStr := params.Get("aids")
	qnStr := params.Get("qn")
	qn, _ := strconv.ParseInt(qnStr, 10, 64)
	pt := params.Get("platform")
	ip := params.Get("ip")
	fnver, _ := strconv.ParseInt(params.Get("fnver"), 10, 64)
	fnval, _ := strconv.ParseInt(params.Get("fnval"), 10, 64)
	forceHost, _ := strconv.ParseInt(params.Get("force_host"), 10, 64)
	session := params.Get("session")
	containsPGC, _ := strconv.Atoi(params.Get("contains_pgc"))
	build, _ := strconv.ParseInt(params.Get("build"), 10, 64)
	device := params.Get("device")
	mid, _ := strconv.ParseInt(params.Get("mid"), 10, 64)
	fourk, _ := strconv.ParseInt(params.Get("fourk"), 10, 64)
	buvid := params.Get("buvid")
	from := params.Get("from")
	// check params
	aids, err := xstr.SplitInts(aidsStr)
	if err != nil {
		log.Error("query aids(%s) split error(%v)", aidsStr, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if len(aids) > _maxAids {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	netType, tfType := model.TrafficFree(params.Get("tf_isp"))
	var batchPlayArg *api.BatchPlayArg
	batchArg := params.Get("batch_play_arg")
	if batchArg == "" {
		// 记录日志观察除动态外是否有业务调用
		log.Warn("no batch_play_arg appkey(%s) platform(%s) build(%d)", params.Get("appkey"), pt, build)
		batchPlayArg = &api.BatchPlayArg{
			Build:          build,
			Device:         device,
			NetType:        api.NetworkType(netType),
			Qn:             qn,
			MobiApp:        pt,
			Fnver:          fnver,
			Fnval:          fnval,
			Ip:             ip,
			Session:        session,
			ForceHost:      forceHost,
			Buvid:          buvid,
			Mid:            mid,
			Fourk:          fourk,
			TfType:         api.TFType(tfType),
			From:           from,
			ShowPgcPlayurl: containsPGC == 1,
		}
	} else {
		if err := json.Unmarshal([]byte(batchArg), &batchPlayArg); err != nil {
			log.Error("BatchPlayArg json.Unmarshal err(%+v) arg(%s)", err, batchArg)
		}
	}
	// 由于动态秒开层级修改在archive做版本判断
	plat := model.Plat(batchPlayArg.GetMobiApp(), batchPlayArg.GetDevice())
	if model.PlayerInfoNew(plat, batchPlayArg.GetBuild()) {
		c.JSON(arcSvc.ArcsWithPlayurl(c, &api.ArcsWithPlayurlRequest{
			Aids:         aids,
			BatchPlayArg: batchPlayArg,
		}))
	} else {
		c.JSON(arcSvc.ArchivesWithPlayer(c, &archive.ArgPlayer{
			Aids:      aids,
			Qn:        batchPlayArg.GetQn(),
			Platform:  batchPlayArg.GetMobiApp(),
			RealIP:    batchPlayArg.GetIp(),
			Build:     batchPlayArg.GetBuild(),
			Fnval:     batchPlayArg.GetFnval(),
			Fnver:     batchPlayArg.GetFnver(),
			Session:   batchPlayArg.GetSession(),
			ForceHost: batchPlayArg.GetForceHost(),
			Mid:       batchPlayArg.GetMid(),
			Buvid:     batchPlayArg.GetBuvid(),
		}, containsPGC == 1))
	}
}

func typelist(c *bm.Context) {
	c.JSON(arcSvc.AllTypes(c), nil)
}

package middleware

import (
	"context"
	"fmt"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/archive/service/api"

	archivev1 "go-gateway/app/app-svr/archive/middleware/v1"
)

type batchPlayKey struct{}

func trafficFree(xTfIsp string) (api.NetworkType, api.TFType) {
	switch xTfIsp {
	case "ct":
		return api.NetworkType_CELLULAR, api.TFType_T_CARD
	case "cu":
		return api.NetworkType_CELLULAR, api.TFType_U_CARD
	case "cm":
		return api.NetworkType_CELLULAR, api.TFType_C_CARD
	}
	return 0, 0
}

func disableFnval(arg *api.BatchPlayArg) bool {
	if arg == nil {
		return false
	}
	if arg.MobiApp == "android" && arg.Build <= 5325000 {
		return true
	}
	if arg.MobiApp == "iphone" && arg.Build <= 8160 {
		return true
	}
	return false
}

func NewContext(ctx context.Context, arg *api.BatchPlayArg) context.Context {
	return context.WithValue(ctx, batchPlayKey{}, arg)
}

func FromContext(ctx context.Context) (*api.BatchPlayArg, bool) {
	arg, ok := ctx.Value(batchPlayKey{}).(*api.BatchPlayArg)
	return arg, ok
}

// 设置秒开所需参数，注意需要放置在登录认证的 Handler 之后
func BatchPlayArgs() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		args := new(api.BatchPlayArg)
		if err := ctx.Bind(args); err != nil {
			log.Error("bind BatchPlayArg(%s) err(%+v)", ctx.Request.Form.Encode(), err)
			return
		}
		if midInter, ok := ctx.Get("mid"); ok {
			args.Mid = midInter.(int64)
		}
		if disableFnval(args) {
			args.Fnval = 0
		}
		args.Buvid = ctx.Request.Header.Get("Buvid")
		// From 为非请求参数，由业务主动填入
		args.From = ""
		args.Ip = metadata.String(ctx, metadata.RemoteIP)
		args.NetType, args.TfType = trafficFree(ctx.Request.Header.Get("X-Tf-Isp"))
		playerNetStr := ctx.Request.Form.Get("player_net")
		playerNet, _ := strconv.ParseInt(playerNetStr, 10, 64)
		if playerNet > 0 {
			args.NetType = api.NetworkType(playerNet)
		}
		ctx.Context = NewContext(ctx.Context, args)
	}
}

func MossBatchPlayArgs(params *archivev1.PlayerArgs, dev device.Device, network network.Network, mid int64) *api.BatchPlayArg {
	batchArg := &api.BatchPlayArg{}
	if params != nil {
		batchArg.Fnval = params.Fnval
		batchArg.Fnver = params.Fnver
		batchArg.ForceHost = params.ForceHost
		batchArg.Qn = params.Qn
		batchArg.VoiceBalance = params.VoiceBalance
	}
	batchArg.Ip = network.RemoteIP
	batchArg.Buvid = dev.Buvid
	batchArg.MobiApp = dev.RawMobiApp
	batchArg.NetType = api.NetworkType(network.Type)
	batchArg.TfType = api.TFType(network.TF)
	batchArg.Build = dev.Build
	batchArg.Mid = mid
	batchArg.Device = dev.Device
	return batchArg
}

//省份及大区映射先保留
//东北：'吉林4259840','辽宁4292608','黑龙江4440064'
//华北：'北京4243456','天津4276224','河北4554752','山西4456448','内蒙古4325376','山东4423680'
//西北：'新疆4653056','陕西4472832','甘肃4620288','青海4669440','宁夏4358144'
//华中：'湖北4587520','湖南4702208','河南4505600','江西4685824'
//华东：'上海4308992','江苏4390912','浙江4374528','安徽4407296'
//西南：'四川4538368','重庆4521984','西藏4603904','云南4571136','贵州4341760'
//华南：'广东4227072','广西4489216','海南4636672','福建4210688'
//const (
//	jiLin        = 4259840
//	liaoNing     = 4292608
//	heiLongJiang = 4440064
//	beiJing      = 4243456
//	tianJing     = 4276224
//	heBei        = 4554752
//	shanXi1      = 4456448
//	neiMengGu    = 4325376
//	shanDong     = 4423680
//	xinJiang     = 4653056
//	shanXi4      = 4472832
//	ganSu        = 4620288
//	qingHai      = 4669440
//	ningXia      = 4358144
//	huBei        = 4587520
//	huNan        = 4702208
//	heNan        = 4505600
//	jiangXi      = 4685824
//	shangHai     = 4308992
//	jiangSu      = 4390912
//	zheJiang     = 4374528
//	anHui        = 4407296
//	siChuan      = 4538368
//	chongQing    = 4521984
//	xiZang       = 4603904
//	yunNan       = 4571136
//	guiZhou      = 4341760
//	guangDong    = 4227072
//	guangXi      = 4489216
//	haiNan       = 4636672
//	fuJian       = 4210688
//)
//
//func ProvinceToRegion(provinceID int64) string {
//	switch provinceID {
//	case jiLin, liaoNing, heiLongJiang:
//		return "northeast"
//	case beiJing, tianJing, heBei, shanXi1, neiMengGu, shanDong:
//		return "north"
//	case xinJiang, shanXi4, ganSu, qingHai, ningXia:
//		return "northwest"
//	case huBei, huNan, heNan, jiangXi:
//		return "central"
//	case shangHai, jiangSu, zheJiang, anHui:
//		return "east"
//	case siChuan, chongQing, xiZang, yunNan, guiZhou:
//		return "southwest"
//	case guangDong, guangXi, haiNan, fuJian:
//		return "south"
//	default:
//		return ""
//	}
//}

//var RegionMap = map[string][]int64{
//	"northeast": {jiLin, liaoNing, heiLongJiang},
//	"north":     {beiJing, tianJing, heBei, shanXi1, neiMengGu, shanDong},
//	"northwest": {xinJiang, shanXi4, ganSu, qingHai, ningXia},
//	"central":   {huBei, huNan, heNan, jiangXi},
//	"east":      {shangHai, jiangSu, zheJiang, anHui},
//	"southwest": {siChuan, chongQing, xiZang, yunNan, guiZhou},
//	"south":     {guangDong, guangXi, haiNan, fuJian},
//}

func CdnZoneKey(domain, isp string, zoneID int64) string {
	return fmt.Sprintf("%s_%d_%s_new", domain, zoneID, isp)
}

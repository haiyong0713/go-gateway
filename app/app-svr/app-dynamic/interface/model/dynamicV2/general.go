package dynamicV2

import (
	"context"
	"strconv"
	"sync"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	xmetadata "go-common/library/net/metadata"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"

	topicV2Cmn "git.bilibili.co/bapis/bapis-go/topic/common"

	dynamicCommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

// GeneralParam 动态grpcV2通用参数
type GeneralParam struct {
	Restriction *restriction.Restriction
	Device      *device.Device
	Mid         int64
	IP          string
	LocalTime   int32
	// AD card_status
	Pattern       string
	ShareID       string
	ShareMode     int32
	From          string
	AdFrom        string
	CloseAutoPlay bool
	DynFrom       string
	Plat          int8
	Network       *network.Network
	AbTest        AbTest
	Config        *api.Config
}

type AbTest struct {
}

func NewGeneralParamFromCtx(ctx context.Context) *GeneralParam {
	au, _ := auth.FromContext(ctx)
	dev, _ := device.FromContext(ctx)
	limit, _ := restriction.FromContext(ctx)
	// 获取网络信息
	nw, _ := network.FromContext(ctx)
	general := &GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
		Plat:        model.Plat(dev.MobiApp(), dev.Device),
		Network:     &nw,
		LocalTime:   8,
	}
	return general
}

func (g *GeneralParam) SetLocalTime(lo int32) {
	g.LocalTime = lo
	if g.LocalTime < -12 || g.LocalTime > 14 {
		g.LocalTime = 8
	}
}

func (g *GeneralParam) GetTeenagerInt() int {
	if g.Restriction == nil {
		return 0
	}
	if g.Restriction.IsTeenagers {
		return 1
	}
	return 0
}

func (g *GeneralParam) GetDisableRcmdInt() int {
	if g.Restriction == nil {
		return 0
	}
	if g.Restriction.DisableRcmd {
		return 1
	}
	return 0
}

func (g *GeneralParam) GetDisableRcmd() bool {
	if g.Restriction == nil {
		return false
	}
	if g.Restriction.DisableRcmd {
		return true
	}
	return false
}

func (g *GeneralParam) GetBuvid() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Buvid
}

func (g *GeneralParam) GetNetWork() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Network
}

func (g *GeneralParam) GetMobiApp() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.RawMobiApp
}

func (g *GeneralParam) GetPlat() int8 {
	return model.Plat(g.GetMobiApp(), g.GetDevice())
}

func (g *GeneralParam) GetBuild() int64 {
	if g.Device == nil {
		return 0
	}
	return g.Device.Build
}

func (g *GeneralParam) GetChannel() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Channel
}

func (g *GeneralParam) GetBuildStr() string {
	build := g.GetBuild()
	if build != 0 {
		return strconv.FormatInt(build, 10)
	}
	return ""
}

func (g *GeneralParam) GetPlatform() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.RawPlatform
}

func (g *GeneralParam) GetDevice() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Device
}

func (g *GeneralParam) GetVersion() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.VersionName
}

func (g *GeneralParam) GetNetWorkTypeStr() string {
	if g.Network == nil {
		return ""
	}
	switch g.Network.Type {
	case network.TypeWIFI:
		return "wifi"
	case network.TypeCellular:
		return "mobile"
	default:
		return "other"
	}
}

func (g *GeneralParam) GetRemoteIP() string {
	if g.Network == nil {
		return ""
	}
	return g.Network.RemoteIP
}

// 安装 iosHD的ipad
func (g *GeneralParam) IsPadHD() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawMobiApp == "ipad"
}

// 安装 iphone粉版的ipad
func (g *GeneralParam) IsPad() bool {
	return g.GetMobiApp() == "iphone" && g.GetDevice() == "pad"
}

// 安装 android粉版的设备
func (g *GeneralParam) IsAndroidPick() bool {
	return g.GetMobiApp() == "android"
}

// 安装 iphone粉版的手机
func (g *GeneralParam) IsIPhonePick() bool {
	return g.GetMobiApp() == "iphone" && g.GetDevice() == "phone"
}

// 安装 androidHD的设备
func (g *GeneralParam) IsAndroidHD() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawMobiApp == "android_hd"
}

// 是否是国际版
func (g *GeneralParam) IsOverseas() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawMobiApp == "android_i" || g.Device.RawMobiApp == "iphone_i" || g.Device.RawMobiApp == "ipad_i"
}

// 是否是ios平台
func (g *GeneralParam) IsIOSPlatform() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawPlatform == "ios"
}

func (g *GeneralParam) IsAndroidPlatform() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawPlatform == "android"
}

func (g *GeneralParam) ToDynCmnMetaData() *dynamicCommon.CmnMetaData {
	return &dynamicCommon.CmnMetaData{
		Build:     g.GetBuildStr(),
		Platform:  g.GetPlatform(),
		MobiApp:   g.GetMobiApp(),
		Device:    g.GetDevice(),
		FromSpmid: "",
		Version:   g.GetVersion(),
		Buvid:     g.GetBuvid(),
	}
}

func (g *GeneralParam) ToDynMetaDataCtrl(fn func(md *dynamicCommon.MetaDataCtrl)) *dynamicCommon.MetaDataCtrl {
	if g == nil {
		return &dynamicCommon.MetaDataCtrl{}
	}
	ret := &dynamicCommon.MetaDataCtrl{
		Platform:     g.GetPlatform(),
		Build:        g.GetBuildStr(),
		MobiApp:      g.GetMobiApp(),
		Buvid:        g.GetBuvid(),
		Device:       g.GetDevice(),
		From:         g.From,
		TeenagerMode: int32(g.GetTeenagerInt()),
		Version:      g.GetVersion(),
		Network:      int32(g.Network.Type),
		Ip:           g.Network.RemoteIP,
		Port:         g.Network.RemotePort,
		Ua:           g.Device.UserAgent,
	}
	if fn != nil {
		fn(ret)
	}
	return ret
}

func (g *GeneralParam) ToDynVersionCtrlMeta(modFn ...func(m *dynamicCommon.VersionCtrlMeta)) *dynamicCommon.VersionCtrlMeta {
	if g == nil {
		return &dynamicCommon.VersionCtrlMeta{}
	}
	ret := &dynamicCommon.VersionCtrlMeta{
		Build:    g.GetBuildStr(),
		Platform: g.GetPlatform(),
		MobiApp:  g.GetMobiApp(),
		Buvid:    g.GetBuvid(),
		Device:   g.GetDevice(),
		Ip:       g.IP,
		From:     g.From,
		Version:  g.GetVersion(),
		Network:  g.Device.NetworkType,
	}
	if g.Device != nil {
		ret.Network = g.Device.NetworkType
	}
	if g.Restriction != nil {
		if g.Restriction.DisableRcmd {
			ret.CloseRcmd = 1
		}
		if g.Restriction.IsTeenagers {
			ret.TeenagerMode = 1
		}
	}
	for _, fn := range modFn {
		fn(ret)
	}
	return ret
}

func (g *GeneralParam) ToTopicCmnMetaData() *topicV2Cmn.MetaDataCtrl {
	if g == nil {
		return nil
	}
	return &topicV2Cmn.MetaDataCtrl{
		Platform:     g.GetPlatform(),
		Build:        g.GetBuildStr(),
		MobiApp:      g.GetMobiApp(),
		Buvid:        g.GetBuvid(),
		Device:       g.GetDevice(),
		FromSpmid:    "",
		From:         "",
		TraceId:      "",
		TeenagerMode: int32(g.GetTeenagerInt()),
		ColdStart:    0,
		Version:      g.GetVersion(),
		Network:      g.Device.NetworkType,
		Ip:           g.IP,
		Port:         "",
	}
}

type BuildLimitOperator int8

const (
	Less BuildLimitOperator = iota
	LessOrEqual
	Equal
	GreaterOrEqual
	Greater
)

func (g *GeneralParam) IsMobileBuildLimitMet(operator BuildLimitOperator, requiredAndroidBuild, requiredIosBuild int64) bool {
	var compare func(cur, required int64) bool
	switch operator {
	case Less:
		compare = func(cur, required int64) bool {
			return cur < required
		}
	case LessOrEqual:
		compare = func(cur, required int64) bool {
			return cur <= required
		}
	case Equal:
		compare = func(cur, required int64) bool {
			return cur == required
		}
	case GreaterOrEqual:
		compare = func(cur, required int64) bool {
			return cur >= required
		}
	case Greater:
		compare = func(cur, required int64) bool {
			return cur > required
		}
	}
	return (g.IsAndroidPick() && compare(g.GetBuild(), requiredAndroidBuild)) ||
		(g.IsIPhonePick() && compare(g.GetBuild(), requiredIosBuild))
}

// 当前请求的context里 某些功能的开关情况
// 通常由后台配置 然后应用读取
type FeatureStatus struct {
	// 不出游戏附加卡
	NoGameAttach *FeatureLazySwitch
}

type FeatureLazySwitch struct {
	once sync.Once
	// 用于获取结果的function
	Fn func(ctx context.Context) bool
	// 结果缓存
	result bool
}

// 开关是否打开
func (ls *FeatureLazySwitch) IsOn(ctx context.Context) bool {
	if ls == nil || ls.Fn == nil {
		return false
	}
	ls.once.Do(func() {
		ls.result = ls.Fn(ctx)
	})
	return ls.result
}

type _featureStatusKey struct{}

func FeatureStatusFromCtx(ctx context.Context) *FeatureStatus {
	if fs, ok := ctx.Value(_featureStatusKey{}).(*FeatureStatus); ok && fs != nil {
		return fs
	}
	return &FeatureStatus{}
}

func NewFeatureStatusCtx(ctx context.Context, status *FeatureStatus) context.Context {
	return context.WithValue(ctx, _featureStatusKey{}, status)
}

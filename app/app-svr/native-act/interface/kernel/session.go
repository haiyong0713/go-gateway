package kernel

import (
	"context"
	"reflect"
	"strconv"

	api "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/trace"

	"go-gateway/app/app-svr/native-act/interface/middleware/webdevice"
)

const (
	_platformH5     = "h5"
	_platformWeb    = "web"
	_mobiAppIphone  = "iphone"
	_mobiAppAndroid = "android"
)

// 卡片构建会话数据
type Session struct {
	ctx context.Context
	// 延迟加载，必须调用RawXXX()来获取数据
	auth        *auth.Auth
	device      *device.Device
	restriction *restriction.Restriction
	network     *network.Network
	trace       trace.Trace
	webdevice   *webdevice.Webdevice
	// 请求参数
	ReqFrom     string
	ShareReq    *ShareReq
	FromSpmid   string
	IsColdStart bool
	LocalTime   int32
	FeedOffset  *api.FeedOffset
	Offset      int64
	LastGroup   int64
	HttpsUrlReq int32
	TabFrom     string
	CurrentTab  string
	OffsetStr   string
	Index       int64
	SortType    int64
	// 上下文数据
	PageRlyContext PageRlyContext
}

type ShareReq struct {
	ShareOrigin string
	TabID       int64
	TabModuleID int64
}

type PageRlyContext struct {
	HasNavigation bool
}

func NewSession(c context.Context) *Session {
	return &Session{ctx: c}
}

func (cs *Session) RawAuth() *auth.Auth {
	if cs.auth == nil {
		au, _ := auth.FromContext(cs.ctx)
		cs.auth = &au
	}
	return cs.auth
}

func (cs *Session) RawDevice() *device.Device {
	if cs.device == nil {
		dev, _ := device.FromContext(cs.ctx)
		cs.device = &dev
	}
	return cs.device
}

func (cs *Session) RawRestriction() *restriction.Restriction {
	if cs.restriction == nil {
		rest, _ := restriction.FromContext(cs.ctx)
		cs.restriction = &rest
	}
	return cs.restriction
}

func (cs *Session) RawNetwork() *network.Network {
	if cs.network == nil {
		net, _ := network.FromContext(cs.ctx)
		cs.network = &net
	}
	return cs.network
}

func (cs *Session) RawWebdevice() *webdevice.Webdevice {
	if cs.webdevice == nil {
		dev, _ := webdevice.FromContext(cs.ctx)
		cs.webdevice = &dev
	}
	return cs.webdevice
}

func (cs *Session) Mid() int64 {
	return cs.RawAuth().Mid
}

func (cs *Session) IsIOS() bool {
	return cs.RawDevice().IsIOS()
}

func (cs *Session) IsIPad() bool {
	return cs.RawDevice().Plat() == device.PlatIPad || cs.RawDevice().Plat() == device.PlatIPadI
}

func (cs *Session) IsAndroid() bool {
	return cs.RawDevice().IsAndroid()
}

func (cs *Session) IsH5() bool {
	return cs.RawWebdevice().Platform == _platformH5
}

// IsH5IOS h5只能区分iphone还是安卓.
func (cs *Session) IsH5IOS() bool {
	return cs.RawWebdevice().MobiApp == _mobiAppIphone
}

// IsH5Android h5只能区分iphone还是安卓.
func (cs *Session) IsH5Android() bool {
	return cs.RawWebdevice().MobiApp == _mobiAppAndroid
}

func (cs *Session) IsWeb() bool {
	return cs.RawWebdevice().Platform == _platformWeb
}

func (cs *Session) Platform() string {
	switch {
	case cs.IsIOS():
		// ios
		return cs.RawDevice().RawPlatform
	case cs.IsAndroid():
		// android
		return cs.RawDevice().RawPlatform
	case cs.IsH5():
		return _platformH5
	case cs.IsWeb():
		return _platformWeb
	default:
		log.Warn("invalid platform, device=%+v", cs.RawDevice())
	}
	return ""
}

func (cs *Session) MobiApp() string {
	switch {
	case cs.IsIOS(), cs.IsAndroid(), cs.IsIPad():
		return cs.RawDevice().MobiApp()
	case cs.IsH5(), cs.IsWeb():
		return cs.RawWebdevice().MobiApp
	default:
		return ""
	}
}

func (cs *Session) Buvid() string {
	switch {
	case cs.IsIOS(), cs.IsAndroid(), cs.IsIPad():
		return cs.RawDevice().Buvid
	case cs.IsH5(), cs.IsWeb():
		return cs.RawWebdevice().Buvid
	default:
		return ""
	}
}

func (cs *Session) UserAgent() string {
	switch {
	case cs.IsIOS(), cs.IsAndroid(), cs.IsIPad():
		return cs.RawDevice().UserAgent
	case cs.IsH5(), cs.IsWeb():
		return cs.RawWebdevice().UserAgent
	default:
		return ""
	}
}

func (cs *Session) TeenagerMode() int32 {
	if cs.RawRestriction().IsTeenagers {
		return 1
	}
	return 0
}

func (cs *Session) FormatInt(num int64) string {
	if num == 0 {
		return ""
	}
	return strconv.FormatInt(num, 10)
}

func (cs *Session) Ip() string {
	return metadata.String(cs.ctx, metadata.RemoteIP)
}

func (cs *Session) TraceId() string {
	if cs.trace == nil {
		t, ok := trace.FromContext(cs.ctx)
		if !ok {
			return ""
		}
		cs.trace = t
	}
	if reflect.ValueOf(cs.trace).IsNil() {
		return ""
	}
	return cs.trace.TraceID()
}

package cardschema

import (
	"time"
)

type Device interface {
	Sid() string          // cookie: sid
	Buvid3() string       // cookie: buvid
	Build() int64         // app: 构建号
	Buvid() string        // app: buvid
	Channel() string      // app: 市场渠道
	Device() string       // app: 运行设备
	RawPlatform() string  // app: 设备类型
	RawMobiApp() string   // app: 包类型
	Model() string        // app: 手机型号
	Brand() string        // app: 手机品牌
	Osver() string        // app: 系统版本
	UserAgent() string    // user-agent
	Network() string      // network wifi or mobile
	NetworkType() int32   // network type
	TfISP() string        // tf isp
	TfType() int32        // // tf type
	FawkesEnv() string    // fawkes 客户端版本管理工具 env if empty than prod
	FawkesAppKey() string // fawkes 客户端版本管理工具 app key

	Plat() int8
	IsAndroid() bool
	IsIOS() bool
	IsIOSNormal() bool
	IsWeb() bool
	IsOverseas() bool
	InvalidChannel(cfgCh string) bool
	//MobiApp() string
	MobiAPPBuleChange() string
	TrafficFree() (netType, tfType int32)

	String() string
}

type IndexParam interface {
	Idx() int64
	Pull() bool
	Column() int8
	LoginEvent() int
	OpenEvent() string
	BannerHash() string
	AdExtra() string
	Interest() string
	Flush() int
	AutoPlayCard() int
	DeviceType() int
	ParentMode() int
	ForceHost() int
	RecsysMode() int
	TeenagersMode() int
	LessonsMode() int
	DeviceName() string
	AccessKey() string
	ActionKey() string
	Statistics() string
	Appver() int
	Filtered() int
	AppKey() string
	HttpsUrlReq() int
	InterestV2() string
	SplashID() int64
	Guidance() int
	AppList() string
	DeviceInfo() string
	IsCloseRcmd() int
}

type UserSession interface {
	CurrentMid() int64
	IsAttentionTo(int64) bool
	IndexParam() IndexParam
}

type FlagFeature string
type StateFeature string

type FeatureGates interface {
	FeatureEnabled(FlagFeature) bool
	FeatureState(StateFeature) int64
	EnableFeature(FlagFeature)
	DisableFeature(FlagFeature)
	SetFeatureState(StateFeature, int64)
}

type VersionControl interface {
	Can(string) bool
}

type FeedContext interface {
	UserSession
	AtTime() time.Time
	Device() Device
	FeatureGates() FeatureGates
	VersionControl() VersionControl
}

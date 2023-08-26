package feedcard

import (
	"fmt"
	"time"

	"go-common/component/metadata/device"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	"go-gateway/app/app-svr/app-feed/interface/model"
)

type featureGates struct {
	enabledStore map[cardschema.FlagFeature]bool
	stateStore   map[cardschema.StateFeature]int64
}

type userSession struct {
	currentMid       int64
	isAttentionStore map[int64]int8
	param            ctxIndexParam
}

type ctxIndexParam struct {
	inner *IndexParam
}

type feedContext struct {
	*userSession
	atTime         time.Time
	featureGates   *featureGates
	device         *CtxDevice
	versionControl *ctxedVersionControl
}

type CtxDevice struct {
	inner *device.Device
}

func NewCtxDevice(dev *device.Device) *CtxDevice {
	return &CtxDevice{inner: dev}
}

func NewUserSession(currentMid int64, isAttentionStore map[int64]int8, param *IndexParam) *userSession {
	return &userSession{
		currentMid:       currentMid,
		isAttentionStore: isAttentionStore,
		param:            ctxIndexParam{inner: param},
	}
}

var _ cardschema.IndexParam = &ctxIndexParam{}
var _ cardschema.UserSession = &userSession{}
var _ cardschema.Device = &CtxDevice{}
var _ cardschema.FeatureGates = &featureGates{}
var _ cardschema.FeedContext = &feedContext{}

func (cd *CtxDevice) Sid() string         { return cd.inner.Sid }
func (cd *CtxDevice) Buvid3() string      { return cd.inner.Buvid3 }
func (cd *CtxDevice) Build() int64        { return cd.inner.Build }
func (cd *CtxDevice) Buvid() string       { return cd.inner.Buvid }
func (cd *CtxDevice) Channel() string     { return cd.inner.Channel }
func (cd *CtxDevice) Device() string      { return cd.inner.Device }
func (cd *CtxDevice) RawPlatform() string { return cd.inner.RawPlatform }
func (cd *CtxDevice) RawMobiApp() string  { return cd.inner.RawMobiApp }
func (cd *CtxDevice) Model() string       { return cd.inner.Model }
func (cd *CtxDevice) Brand() string       { return cd.inner.Brand }
func (cd *CtxDevice) Osver() string       { return cd.inner.Osver }
func (cd *CtxDevice) UserAgent() string   { return cd.inner.UserAgent }
func (cd *CtxDevice) Network() string     { return cd.inner.Network }
func (cd *CtxDevice) NetworkType() int32 {
	netType, _ := cd.inner.TrafficFree()
	return netType
}
func (cd *CtxDevice) TfISP() string { return cd.inner.TfISP }
func (cd *CtxDevice) TfType() int32 {
	_, tfType := cd.inner.TrafficFree()
	return tfType
}
func (cd *CtxDevice) FawkesEnv() string {
	if cd.inner.FawkesEnv == "" {
		return "prod"
	}
	return cd.inner.FawkesEnv
}
func (cd *CtxDevice) FawkesAppKey() string { return cd.inner.FawkesAppKey }
func (cd *CtxDevice) Plat() int8 {
	plat := cd.inner.Plat()
	if cd.inner.RawMobiApp == "ipad" {
		plat = model.PlatIPadHD
	}
	return plat
}
func (cd *CtxDevice) IsAndroid() bool                      { return cd.inner.IsAndroid() }
func (cd *CtxDevice) IsIOS() bool                          { return cd.inner.IsIOS() }
func (cd *CtxDevice) IsWeb() bool                          { return cd.inner.IsWeb() }
func (cd *CtxDevice) IsOverseas() bool                     { return cd.inner.IsOverseas() }
func (cd *CtxDevice) InvalidChannel(cfgCh string) bool     { return cd.inner.InvalidChannel(cfgCh) }
func (cd *CtxDevice) MobiApp() string                      { return cd.inner.MobiApp() }
func (cd *CtxDevice) MobiAPPBuleChange() string            { return cd.inner.MobiAPPBuleChange() }
func (cd *CtxDevice) TrafficFree() (netType, tfType int32) { return cd.inner.TrafficFree() }
func (cd *CtxDevice) String() string                       { return fmt.Sprintf("%+v", cd.inner) }
func (cd *CtxDevice) IsIOSNormal() bool {
	plat := cd.inner.Plat()
	return plat == device.PlatIPad ||
		plat == device.PlatIPhone ||
		plat == device.PlatIPadI ||
		plat == device.PlatIPhoneI
}

func (cip ctxIndexParam) Idx() int64         { return cip.inner.Idx }
func (cip ctxIndexParam) Pull() bool         { return cip.inner.Pull }
func (cip ctxIndexParam) Column() int8       { return int8(cip.inner.Column) }
func (cip ctxIndexParam) LoginEvent() int    { return cip.inner.LoginEvent }
func (cip ctxIndexParam) OpenEvent() string  { return cip.inner.OpenEvent }
func (cip ctxIndexParam) BannerHash() string { return cip.inner.BannerHash }
func (cip ctxIndexParam) AdExtra() string    { return cip.inner.AdExtra }
func (cip ctxIndexParam) Interest() string   { return cip.inner.Interest }
func (cip ctxIndexParam) Flush() int         { return cip.inner.Flush }
func (cip ctxIndexParam) AutoPlayCard() int  { return cip.inner.AutoPlayCard }
func (cip ctxIndexParam) DeviceType() int    { return cip.inner.DeviceType }
func (cip ctxIndexParam) ParentMode() int    { return cip.inner.ParentMode }
func (cip ctxIndexParam) ForceHost() int     { return cip.inner.ForceHost }
func (cip ctxIndexParam) RecsysMode() int    { return cip.inner.RecsysMode }
func (cip ctxIndexParam) TeenagersMode() int { return cip.inner.TeenagersMode }
func (cip ctxIndexParam) LessonsMode() int   { return cip.inner.LessonsMode }
func (cip ctxIndexParam) DeviceName() string { return cip.inner.DeviceName }
func (cip ctxIndexParam) AccessKey() string  { return cip.inner.AccessKey }
func (cip ctxIndexParam) ActionKey() string  { return cip.inner.ActionKey }
func (cip ctxIndexParam) Statistics() string { return cip.inner.Statistics }
func (cip ctxIndexParam) Appver() int        { return cip.inner.Appver }
func (cip ctxIndexParam) Filtered() int      { return cip.inner.Filtered }
func (cip ctxIndexParam) AppKey() string     { return cip.inner.AppKey }
func (cip ctxIndexParam) HttpsUrlReq() int   { return cip.inner.HttpsUrlReq }
func (cip ctxIndexParam) InterestV2() string { return cip.inner.InterestV2 }
func (cip ctxIndexParam) SplashID() int64    { return cip.inner.SplashID }
func (cip ctxIndexParam) Guidance() int      { return cip.inner.Guidance }
func (cip ctxIndexParam) AppList() string    { return cip.inner.AppList }
func (cip ctxIndexParam) DeviceInfo() string { return cip.inner.DeviceInfo }
func (cip ctxIndexParam) IsCloseRcmd() int   { return cip.inner.DisableRcmd }

func (us *userSession) CurrentMid() int64                 { return us.currentMid }
func (us *userSession) IsAttentionTo(in int64) bool       { return us.isAttentionStore[in] > 0 }
func (us *userSession) IndexParam() cardschema.IndexParam { return us.param }

func (fg *featureGates) FeatureEnabled(feature cardschema.FlagFeature) bool {
	return fg.enabledStore[feature]
}
func (fg *featureGates) FeatureState(feature cardschema.StateFeature) int64 {
	return fg.stateStore[feature]
}
func (fg *featureGates) EnableFeature(feature cardschema.FlagFeature) {
	fg.enabledStore[feature] = true
}
func (fg *featureGates) DisableFeature(feature cardschema.FlagFeature) {
	fg.enabledStore[feature] = false
}
func (fg *featureGates) SetFeatureState(feature cardschema.StateFeature, state int64) {
	fg.stateStore[feature] = state
}

func (fc *feedContext) Device() cardschema.Device                 { return fc.device }
func (fc *feedContext) FeatureGates() cardschema.FeatureGates     { return fc.featureGates }
func (fc *feedContext) VersionControl() cardschema.VersionControl { return fc.versionControl }
func (fc *feedContext) AtTime() time.Time                         { return fc.atTime }

func NewFeedContext(userSession *userSession, device *CtxDevice, atTime time.Time) cardschema.FeedContext {
	fCtx := &feedContext{
		userSession: userSession,
		atTime:      atTime,
		featureGates: &featureGates{
			enabledStore: map[cardschema.FlagFeature]bool{},
			stateStore:   map[cardschema.StateFeature]int64{},
		},
		device: device,
	}
	fCtx.versionControl = NewCtxedVersionControl(fCtx.device)
	return fCtx
}

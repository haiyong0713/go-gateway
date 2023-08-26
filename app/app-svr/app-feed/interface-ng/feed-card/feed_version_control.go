package feedcard

import (
	"fmt"

	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	"go-gateway/app/app-svr/app-feed/interface/model"

	"github.com/pkg/errors"
)

type versionFlagFunc func(cardschema.Device) bool

type versionControlStore struct {
	fnStore map[string]versionFlagFunc
}

var globalVersionControlStore = versionControlStore{
	fnStore: map[string]versionFlagFunc{},
}
var _ cardschema.VersionControl = &ctxedVersionControl{}

func init() {
	register("pic.activitySupported", activitySupported)
	register("pic.usingLikeText", usingLikeText)
	register("archive.storyPlayerSupported", storyPlayerSupported)
	register("archive.usingFeedIndexLike", usingFeedIndexLike)
	register("archvie.showCooperation", showCooperation)
	register("pgc.inlinePGCAutoPlaySupported", inlinePGCAutoPlaySupported)
	register("feed.usingNewThreePointV2", usingNewThreePointV2)
	register("feed.enableThreePointV2Feedback", enableThreePointV2Feedback)
	register("feed.usingNewRcmdReason", usingNewRcmdReason)
	register("feed.usingNewRcmdReasonV2", usingNewRcmdReasonV2)
	register("feed.usingChannelAsTag", UsingChannelAsTag)
	register("feed.usingOnePicV2", usingOnePicV2)
	register("feed.usingOnePicV3", usingOnePicV3)
	register("feed.usingThreePicV3", usingThreePicV3)
	register("feed.usingIpadColumn", usingIpadColumn)
	register("feed.usingInline2", usingInline2)
	register("feed.usingNewBanner", UsingNewBanner)
	register("feed.usingRemind", usingRemind)
	register("feed.compatibleWithGameID", compatibleWithGameID)
	register("feed.adAvHasPlayerArgs", adAvHasPlayerArgs)
	register("feed.isIOS617", IsIOS617)
	register("feed.inlineThreePointPanelSupported", inlineThreePointPanelSupported)
	register("feed.enableSwitchColumn", CanEnableSwitchColumn)
	register("feed.enableOGVFeedback", usingOGVFeedback)
	register("feed.enablePadNewCover", enablePadNewCover)
	register("banner.enableAdInline", enableAdInline)
	register("feed.enableInlineTunnel", enableInlineTunnel)
	register("feed.disableHDArticleS", disableHDArticleS)
	register("feed.disableInt64Mid", disableInt64Mid)
	register("feed.pgcScore", pgcScore)
	register("feed.enableLiveWatched", enableLiveWatched)
}

func register(identifier string, fn versionFlagFunc) {
	if err := globalVersionControlStore.register(identifier, fn); err != nil {
		panic(err)
	}
}

func (vcs *versionControlStore) register(identifier string, fn versionFlagFunc) error {
	if _, ok := vcs.fnStore[identifier]; ok {
		return errors.Errorf("conflicated version control function: %q", identifier)
	}
	vcs.fnStore[identifier] = fn
	return nil
}

func activitySupported(device cardschema.Device) bool {
	return (device.IsAndroid() && device.Build() >= 5510000) ||
		(device.IsIOS() && device.Build() > 8960)
}

func usingLikeText(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Device() == "phone" && device.Build() >= 8290) ||
		(device.RawMobiApp() == "android" && device.Build() >= 5360000)
}

func storyPlayerSupported(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 10030) ||
		(device.RawMobiApp() == "android" && device.Build() > 6025500) ||
		(device.RawMobiApp() == "android_i" && device.Build() >= 6790300) ||
		(device.RawMobiApp() == "iphone_i" && device.Build() >= 67900200)
}

func usingNewThreePointV2(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8470) ||
		(device.RawMobiApp() == "android" && device.Build() > 5405000) ||
		(device.RawMobiApp() == "ipad" && device.Build() > 12080) ||
		(device.RawMobiApp() == "iphone_b" && device.Build() >= 8000) ||
		(device.RawMobiApp() == "android_i" && device.Build() >= 3000500) ||
		(device.RawMobiApp() == "android_b" && device.Build() >= 5370100) ||
		(device.RawMobiApp() == "win")
}

func usingFeedIndexLike(device cardschema.Device) bool {
	// never using this now
	return false
}

func enableThreePointV2Feedback(device cardschema.Device) bool {
	return device.RawMobiApp() != "android_i"
}

func usingNewRcmdReason(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Device() == "phone" && device.Build() > 9150) ||
		(device.RawMobiApp() == "android" && device.Build() > 5525000)
}

func usingNewRcmdReasonV2(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() > 6025000) ||
		(device.RawMobiApp() == "iphone" && device.Build() > 10130) ||
		(device.RawMobiApp() == "ipad" && device.Build() >= 32100000) ||
		(device.RawMobiApp() == "android_hd" && device.Build() >= 1030000) ||
		(device.RawMobiApp() == "win")
}

func UsingChannelAsTag(device cardschema.Device) bool {
	return (device.Plat() == model.PlatIPhone && device.Build() > 8820) ||
		(device.Plat() == model.PlatAndroid && device.Build() > 5479999)
}

func usingOnePicV2(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8300) ||
		(device.Plat() == model.PlatAndroid && device.Build() > 5365000)
}

func usingOnePicV3(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8470) ||
		(device.Plat() == model.PlatAndroid && device.Build() > 5405000)
}

func usingThreePicV3(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8470) ||
		(device.Plat() == model.PlatAndroid && device.Build() > 5405000)
}

func usingIpadColumn(device cardschema.Device) bool {
	return device.RawMobiApp() == "ipad" || device.RawMobiApp() == "ipad_i"
}

func usingInline2(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 10130) ||
		(device.RawMobiApp() == "android" && device.Build() > 6045000)
}

func compatibleWithGameID(device cardschema.Device) bool {
	return (device.Plat() == model.PlatIPhone && device.Build() >= 10260 && device.Build() <= 10270) ||
		(device.Plat() == model.PlatAndroid && device.Build() >= 6090600 && device.Build() <= 6091000)
}

func usingRemind(device cardschema.Device) bool {
	return (model.IsIOS(device.Plat()) && device.Build() > 8330) || (model.IsAndroid(device.Plat()) && device.Build() > 5375000)
}

func UsingNewBanner(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8510) ||
		(device.RawMobiApp() == "android" && device.Build() > 5415000) ||
		(device.RawMobiApp() == "ipad" && device.Build() > 12110) ||
		(device.RawMobiApp() == "iphone_b" && device.Build() > 8110) ||
		(device.RawMobiApp() == "android_b" && device.Build() > 591240) ||
		(device.RawMobiApp() == "win")
}

func showCooperation(device cardschema.Device) bool {
	return (device.Plat() == model.PlatIPhone && device.Build() > 8290) ||
		(device.Plat() == model.PlatAndroid && device.Build() > 5365000) ||
		(device.RawMobiApp() == "ipad" && device.Build() > 12520) ||
		(device.Plat() == model.PlatIPad && device.Build() >= 63100000) ||
		(device.RawMobiApp() == "win")
}

func inlinePGCAutoPlaySupported(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8470) ||
		(device.RawMobiApp() == "android" && device.Build() > 5405000)
}

func IsIPad(device cardschema.Device) bool {
	return device.Plat() == model.PlatIPad || device.Plat() == model.PlatIPadI || device.Plat() == model.PlatIPadHD
}

func IsAndroidPad(device cardschema.Device) bool {
	return device.Plat() == model.PlatAndroidHD
}

func IsPad(device cardschema.Device) bool {
	return IsIPad(device) || IsAndroidPad(device)
}

func IsCmResource(device cardschema.Device) bool {
	return device.Plat() == model.PlatIPhone ||
		device.Plat() == model.PlatIPhoneB ||
		(device.Plat() == model.PlatAndroid && device.Build() >= 500001) ||
		device.Plat() == model.PlatIPad ||
		device.Plat() == model.PlatIPadHD
}

func IsIOSNewBlue(device cardschema.Device) bool {
	return device.RawMobiApp() == "iphone_b" && device.Build() > 8090
}

func adAvHasPlayerArgs(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() > 8430) || (device.RawMobiApp() == "android" && device.Build() > 5395000)
}

func IsIOS(device cardschema.Device) bool {
	plat := device.Plat()
	return plat == model.PlatIPad || plat == model.PlatIPhone || plat == model.PlatIPadI ||
		plat == model.PlatIPhoneI || plat == model.PlatIPhoneB || plat == model.PlatIPadHD
}

func CanEnable4GWiFiAutoPlay(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() >= 6140000) ||
		(device.RawMobiApp() == "iphone" && device.Build() > 10350)
}

func IsIOS617(device cardschema.Device) bool {
	return device.RawMobiApp() == "iphone" && device.Build() == 61700200
}

func inlineThreePointPanelSupported(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() >= 62600000) ||
		(device.RawMobiApp() == "android" && device.Build() >= 6260000)
}

func CanEnableSwitchColumn(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone" && device.Build() >= 62700000) ||
		(device.RawMobiApp() == "android" && device.Build() >= 6270000)
}

func usingOGVFeedback(device cardschema.Device) bool {
	return (device.RawMobiApp() == "iphone") || (device.RawMobiApp() == "android")
}

func enablePadNewCover(device cardschema.Device) bool {
	return (device.RawMobiApp() == "ipad" && device.Build() >= 32100000) ||
		(device.RawMobiApp() == "iphone" && device.Device() == "pad" && device.Build() >= 63300000) ||
		(device.RawMobiApp() == "android_hd" && device.Build() >= 1030000) ||
		(device.RawMobiApp() == "win")
}

func enableAdInline(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() >= 6290000) ||
		(device.RawMobiApp() == "iphone" && device.Build() >= 62900000)
}

func enableInlineTunnel(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() >= 6150000) ||
		(device.RawMobiApp() == "iphone" && device.Build() >= 61500000)
}

func disableHDArticleS(device cardschema.Device) bool {
	return (device.RawMobiApp() == "ipad" && device.Build() < 32500000) ||
		(device.RawMobiApp() == "iphone" && device.Device() == "pad" && device.Build() < 64200000)
}

func disableInt64Mid(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() < 6500000) ||
		(device.RawMobiApp() == "iphone" && device.Build() < 65000000) ||
		(device.RawMobiApp() == "ipad" && device.Build() < 33000000) ||
		(device.RawMobiApp() == "iphone_i" && device.Build() < 65000000) ||
		(device.RawMobiApp() == "android_i" && device.Build() < 6500000) ||
		(device.RawMobiApp() == "android_b" && device.Build() < 6500000) ||
		(device.RawMobiApp() == "iphone_b" && device.Build() < 65000000) ||
		(device.RawMobiApp() == "android_hd" && device.Build() < 1070000)
}

func pgcScore(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() >= 6560000) ||
		(device.RawMobiApp() == "iphone" && device.Build() >= 65600000)
}

func enableLiveWatched(device cardschema.Device) bool {
	return (device.RawMobiApp() == "android" && device.Build() >= 6610000) ||
		(device.RawMobiApp() == "iphone" && device.Device() == "phone" && device.Build() >= 66100000) ||
		(device.RawMobiApp() == "iphone" && device.Device() == "pad" && device.Build() >= 66200000) ||
		(device.RawMobiApp() == "ipad" && device.Build() >= 33600000)
}

type ctxedVersionControl struct {
	store  versionControlStore
	device cardschema.Device
}

func NewCtxedVersionControl(device cardschema.Device) *ctxedVersionControl {
	return &ctxedVersionControl{
		store:  globalVersionControlStore,
		device: device,
	}
}

func (cvc ctxedVersionControl) Can(identifier string) bool {
	fn, ok := cvc.store.fnStore[identifier]
	if !ok {
		panic(fmt.Sprintf("can not find version control identifier: %s", identifier))
	}
	return fn(cvc.device)
}

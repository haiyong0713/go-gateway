package adresource

var (
	BuildPegasusScene               = CurryingSceneBuilder("pegasus")
	BuildPegasusAdAvScene           = CurryingSceneBuilder("pegasus", "ad_av")
	BuildPegasusBannerScene         = CurryingSceneBuilder("pegasus", "banner")
	BuildPegasusInlineBannerScene   = CurryingSceneBuilder("pegasus", "inline_banner")
	BuildPegasusOldBannerScene      = CurryingSceneBuilder("pegasus", "old_banner")
	BuildPegasusBannerLessonScene   = CurryingSceneBuilder("pegasus", "banner_lesson")
	BuildPegasusBannerTeenagerScene = CurryingSceneBuilder("pegasus", "banner_teenager")

	PegasusAndroid = NewScene(BuildPegasusScene("android"))
	PegasusIOS     = NewScene(BuildPegasusScene("ios"))
	PegasusIPad    = NewScene(BuildPegasusScene("ipad"))

	PegasusAdAvAndroid = NewScene(BuildPegasusAdAvScene("android"))
	PegasusAdAvIOS     = NewScene(BuildPegasusAdAvScene("ios"))
	PegasusAdAvIPad    = NewScene(BuildPegasusAdAvScene("ipad"))

	PegasusInlineBannerAndroid  = NewScene(BuildPegasusInlineBannerScene("android"))
	PegasusInlineBannerAndroidI = NewScene(BuildPegasusInlineBannerScene("android_i"))
	PegasusInlineBannerIOS      = NewScene(BuildPegasusInlineBannerScene("ios"))
	PegasusInlineBannerIPad     = NewScene(BuildPegasusInlineBannerScene("ipad"))
	PegasusInlineBannerIPhoneI  = NewScene(BuildPegasusInlineBannerScene("iphone_i"))

	PegasusOldBannerAndroid  = NewScene(BuildPegasusOldBannerScene("android"))
	PegasusOldBannerAndroidI = NewScene(BuildPegasusOldBannerScene("android_i"))
	PegasusOldBannerAndroidG = NewScene(BuildPegasusOldBannerScene("android_g"))
	PegasusOldBannerIOS      = NewScene(BuildPegasusOldBannerScene("ios"))
	PegasusOldBannerIOSI     = NewScene(BuildPegasusOldBannerScene("iphone_i"))
	PegasusOldBannerIPad     = NewScene(BuildPegasusOldBannerScene("ipad"))
	PegasusOldBannerIPadI    = NewScene(BuildPegasusOldBannerScene("ipad_i"))

	PegasusBannerLessonAndroid = NewScene(BuildPegasusBannerLessonScene("android"))
	PegasusBannerLessonIOS     = NewScene(BuildPegasusBannerLessonScene("ios"))
	PegasusBannerLessonIPad    = NewScene(BuildPegasusBannerLessonScene("ipad"))

	PegasusBannerTeenagerAndroid = NewScene(BuildPegasusBannerTeenagerScene("android"))
	PegasusBannerTeenagerIOS     = NewScene(BuildPegasusBannerTeenagerScene("ios"))
	PegasusBannerTeenagerIPad    = NewScene(BuildPegasusBannerTeenagerScene("ipad"))

	PegasusAndroidID = NewResoueceID(1897)
	PegasusIOSID     = NewResoueceID(1890)
	PegasusIPadID    = NewResoueceID(1975)

	PegasusAdAvAndroidID = NewResoueceID(1690)
	PegasusAdAvIOSID     = NewResoueceID(1685)
	PegasusAdAvIPadID    = NewResoueceID(1974)

	PegasusInlineBannerAndroidID  = NewResoueceID(4336)
	PegasusInlineBannerAndroidIID = NewResoueceID(4592)
	PegasusInlineBannerIOSID      = NewResoueceID(4332)
	PegasusInlineBannerIPadID     = NewResoueceID(4340)
	PegasusInlineBannerIPhoneIID  = NewResoueceID(4948)

	PegasusOldBannerAndroidID  = NewResoueceID(631)
	PegasusOldBannerAndroidIID = NewResoueceID(1707)
	PegasusOldBannerIOSID      = NewResoueceID(467)
	PegasusOldBannerIOSIID     = NewResoueceID(947)
	PegasusOldBannerIPadID     = NewResoueceID(771)
	PegasusOldBannerIPadIID    = NewResoueceID(1117)
	PegasusOldBannerAndroidGID = NewResoueceID(1285)

	PegasusBannerLessonAndroidID = NewResoueceID(3852)
	PegasusBannerLessonIOSID     = NewResoueceID(3848)
	PegasusBannerLessonIPadID    = NewResoueceID(3856)

	PegasusBannerTeenagerAndroidID = NewResoueceID(4960)
	PegasusBannerTeenagerIOSID     = NewResoueceID(4953)
	PegasusBannerTeenagerIPadID    = NewResoueceID(4967)
)

func init() {
	RegisterPegasusResoueceID()
}

func RegisterPegasusResoueceID() {
	Register(PegasusAndroid, PegasusAndroidID)
	Register(PegasusIOS, PegasusIOSID)
	Register(PegasusIPad, PegasusIPadID)

	Register(PegasusAdAvAndroid, PegasusAdAvAndroidID)
	Register(PegasusAdAvIOS, PegasusAdAvIOSID)
	Register(PegasusAdAvIPad, PegasusAdAvIPadID)

	Register(PegasusInlineBannerAndroid, PegasusInlineBannerAndroidID)
	Register(PegasusInlineBannerAndroidI, PegasusInlineBannerAndroidIID)
	Register(PegasusInlineBannerIOS, PegasusInlineBannerIOSID)
	Register(PegasusInlineBannerIPad, PegasusInlineBannerIPadID)
	Register(PegasusInlineBannerIPhoneI, PegasusInlineBannerIPhoneIID)

	Register(PegasusOldBannerAndroid, PegasusOldBannerAndroidID)
	Register(PegasusOldBannerAndroidI, PegasusOldBannerAndroidIID)
	Register(PegasusOldBannerAndroidG, PegasusOldBannerAndroidGID)
	Register(PegasusOldBannerIOS, PegasusOldBannerIOSID)
	Register(PegasusOldBannerIOSI, PegasusOldBannerIOSIID)
	Register(PegasusOldBannerIPad, PegasusOldBannerIPadID)
	Register(PegasusOldBannerIPadI, PegasusOldBannerIPadIID)

	Register(PegasusBannerLessonAndroid, PegasusBannerLessonAndroidID)
	Register(PegasusBannerLessonIOS, PegasusBannerLessonIOSID)
	Register(PegasusBannerLessonIPad, PegasusBannerLessonIPadID)

	Register(PegasusBannerTeenagerAndroid, PegasusBannerTeenagerAndroidID)
	Register(PegasusBannerTeenagerIOS, PegasusBannerTeenagerIOSID)
	Register(PegasusBannerTeenagerIPad, PegasusBannerTeenagerIPadID)
}

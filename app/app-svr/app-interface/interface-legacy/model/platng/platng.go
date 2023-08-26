package platng

const (
	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatIPhone is int8 for iphone.
	PlatIPhone = int8(1)
	// PlatIPad is int8 for ipad.
	PlatIPad = int8(2)
	// PlatWPhone is int8 for wphone.
	PlatWPhone = int8(3)
	// PlatAndroidG is int8 for Android Googleplay.
	PlatAndroidG = int8(4)
	// PlatIPhoneI is int8 for Iphone Global.
	PlatIPhoneI = int8(5)
	// PlatIPadI is int8 for IPAD Global.
	PlatIPadI = int8(6)
	// PlatAndroidTV is int8 for AndroidTV Global.
	PlatAndroidTV = int8(7)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// PlatIpadHD is int8 for IpadHD
	PlatIpadHD = int8(20)
	// PlatAndroidB is int8 for Android Blue.
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for Android Blue.
	PlatIPhoneB = int8(10)
	// PlatAndroidHD is int8 for android_hd
	PlatAndroidHD = int8(90)
)

func Plat(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
		return PlatIPhone
	case "white":
		return PlatIPhone
	case "ipad":
		return PlatIpadHD
	case "android":
		return PlatAndroid
	case "win", "winphone":
		return PlatWPhone
	case "android_G":
		return PlatAndroidG
	case "android_i":
		return PlatAndroidI
	case "iphone_i":
		if device == "pad" {
			return PlatIPadI
		}
		return PlatIPhoneI
	case "ipad_i":
		return PlatIPadI
	case "android_tv":
		return PlatAndroidTV
	case "android_b":
		return PlatAndroidB
	case "iphone_b":
		return PlatIPhoneB
	case "android_hd":
		return PlatAndroidHD
	}
	return PlatIPhone
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
}

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG || plat == PlatAndroidI || plat == PlatAndroidB
}

// IsAndroidPick check plat is android pick
func IsAndroidPick(plat int8) bool {
	return plat == PlatAndroid
}

// IsIOSPick check plat is iphone or ipad pick
func IsIOSPick(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI || plat == PlatIpadHD
}

// IsIPhone check plat is iphone.
func IsIPhone(plat int8) bool {
	return plat == PlatIPhone
}

// IsIPad check plat is pad.
func IsIPad(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPadI || plat == PlatIpadHD
}

// IsPad check plat is pad.
func IsPad(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPadI || plat == PlatIpadHD || plat == PlatAndroidHD
}

// IsIPhoneB check plat is iphone_b.
func IsIPhoneB(plat int8) bool {
	return plat == PlatIPhoneB
}

func IsAndroidB(plat int8) bool {
	return plat == PlatAndroidB
}

func IsBlue(plat int8) bool {
	return plat == PlatIPhoneB || plat == PlatAndroidB
}

func IsAndroidHD(plat int8) bool {
	return plat == PlatAndroidHD
}

func IsIPadHD(plat int8) bool {
	return plat == PlatIpadHD
}

func IsPinkAndBlue(plat int8) bool {
	return plat == PlatIPhone || plat == PlatAndroid || plat == PlatIPhoneB || plat == PlatAndroidB
}

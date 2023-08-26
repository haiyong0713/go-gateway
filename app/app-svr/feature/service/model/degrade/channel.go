package degrade

type ChannelFeatrue struct {
	DecodeType int64 `json:"decode_type"`
	AutoLaunch int32 `json:"auto_launch"`
}

type TvSwitch struct {
	ID         int64               `json:"id"`
	Brand      string              `json:"brand"`
	Chil       string              `json:"chid"`
	Model      string              `json:"model"`
	SysVersion *TvSwitchSysVersion `json:"sys_version"`
}

type TvSwitchSysVersion struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

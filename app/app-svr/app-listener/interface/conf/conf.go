package conf

import (
	"go-common/component/metadata/device"

	"github.com/BurntSushi/toml"
)

// 保存一份package level配置，方便其他包引用
var C *AppConfig

type AppConfig struct {
	Switch  SwitchStatus
	Res     Resources
	Feature FeatureConf
}

func (c *AppConfig) Set(str string) error {
	if C == nil {
		C = c
	}
	return toml.Unmarshal([]byte(str), c)
}

type SwitchStatus struct {
	// 使用旧版HTTP music api
	LegacyMusicAPI bool
}

type Resources struct {
	Text        TextRes
	Icon        IconRes
	HistoryIcon IconHistoryRes
}

type TextRes struct {
	FavBatchAdd     string
	FavBatchDel     string
	FavBatch        string
	AddFav          string
	DelFav          string
	CreateFavFolder string
	DeleteFavFolder string

	EditOK      string
	ThumbUp     string
	ThumbCancel string
	CoinOK      string
	TripleLike  string // 一键三连

	PickHeaderBtn       string
	PickHeaderDetailBtn string
	PickHeaderDesc      string
	PickSeeMoreBtn      string

	MsgArchiveInvalid        string
	MsgUnsupportedSteinsGate string
	MsgCopyrightBanPlay      string
	MsgUnsupported           string
	MsgUnsupportedAudio      string

	TpcdHistory   string
	TpcdFavFolder string
	TpcdFavRecent string
	TpcdUpRecall  string
	TpcdPickToday string
}

type IconRes struct {
	PickHeaderBtn  string
	PickSeeMoreBtn string
}

type IconHistoryRes struct {
	Phone string
	Pad   string
	TV    string
	PC    string
	Car   string
	Iot   string
}

type FeatureConf struct {
	// 主站收藏夹是否出歌单tab
	MusicFavTabShow FeatureGate
	// // 推荐头部卡的Icon是否显示
	RcmdHeadCardIconShow FeatureGate
}

type FeatureGate struct {
	// 安卓手机
	Android BuildNum
	// ios手机
	IPhone BuildNum
}

type BuildNum int64

// 向后兼容逻辑 build号至少满足xxx才允许
// 要求build号必须配置，否则不生效
func (fg FeatureGate) Enabled(d *device.Device) bool {
	if d == nil {
		return false
	}
	switch d.RawMobiApp {
	case "android":
		return fg.Android.LtOrEq(d.Build)
	case "iphone":
		switch d.Device {
		case "phone":
			return fg.IPhone.LtOrEq(d.Build)
		}
	}
	return false
}

func (bd BuildNum) Lt(n int64) bool {
	if bd == 0 {
		return false
	}
	return int64(bd) < n
}

func (bd BuildNum) LtOrEq(n int64) bool {
	if bd == 0 {
		return false
	}
	return int64(bd) <= n
}

func (bd BuildNum) Eq(n int64) bool {
	return int64(bd) == n
}

func (bd BuildNum) Gt(n int64) bool {
	return int64(bd) > n
}

func (bd BuildNum) GtOrEq(n int64) bool {
	return int64(bd) >= n
}

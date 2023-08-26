package dynamic

import (
	"fmt"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	arcApi "go-gateway/app/app-svr/archive/service/api"
)

const (
	BgStyleFill = iota + 1

	// 服务端LBS功能是否打开
	CloseSLB = 0
	OpenSLB  = 1

	// 服务端LBS功能需要开关
	DontAsk = 0
	NeedAsk = 1

	// 客户端slb是否打开
	AppCloseSLB = 0
	AppOpenSLB  = 1

	// 样式 双列1，瀑布流2
	DoubleRow = 1
	WaterFall = 2

	// 封面图样式
	CoverStyle1610 = 1
	CoverStyle34   = 2
	CoverStyle11   = 3

	// 查看更多直播状态
	UplistMoreLiving = 1
)

const (
	LkIconPltAll     = 7
	AV               = "av"
	PGC              = "pgc"
	FOLD             = "fold"
	PerSecond        = 1
	PerMinute        = PerSecond * 60
	PerHour          = PerMinute * 60
	FoldMapTypeShow  = 1
	FoldMapTypeFirst = 2
	FoldTypePublish  = int32(1)
	FoldTypeFrequent = int32(2)
	FoldTypeUnite    = int32(3)
	FoldTypeLimit    = int32(4)
	CtrlTypeAite     = 1
	CtrlTypeLottery  = 2
	CtrlTypeVote     = 3
	CtrlTypeGoods    = 4
	DescTypeLottery  = "lottery"
	DescTypeText     = "text"
	DescTypeVote     = "vote"
	DynMdlFollowType = "followList"
	DynMdlUpList     = "upList"
	HasFold          = 1
)

var (
	// 合作角标
	CooperationBadge = &api.VideoBadge{
		Text:             "合作",
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
	// 付费角标
	PayBadge = &api.VideoBadge{
		Text:             "付费",
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FAAB4B",
		BgColorNight:     "#BA833F",
		BorderColor:      "#FAAB4B",
		BorderColorNight: "#BA833F",
		BgStyle:          BgStyleFill,
	}
	LkIconMap = map[string]int{
		"android": 1,
		"ios":     2,
		"mobile":  3,
		"web":     4,
	}
)

// nolint:gomnd
func DynCityTopLabel(noticeTyle int32, city string) (string, string) {
	switch noticeTyle {
	case 0:
		return "开启定位，获得更多附近精彩内容推荐", "去开启"
	case 2:
		return fmt.Sprintf("你当前所在的城市暂未开通本服务，已为你切换到：%s", city), ""
	}
	return "", ""
}

func MapToAids(m map[int64]struct{}) []*arcApi.PlayAv {
	var aids []*arcApi.PlayAv
	for k := range m {
		aids = append(aids, &arcApi.PlayAv{Aid: k})
	}
	return aids
}

func MapToInt64(m map[int64]struct{}) []int64 {
	var rsp []int64
	for k := range m {
		rsp = append(rsp, k)
	}
	return rsp
}

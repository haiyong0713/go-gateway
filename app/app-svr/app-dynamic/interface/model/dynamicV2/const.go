package dynamicV2

import (
	"fmt"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	// 角标填充类型
	BgStyleFill              = 1
	BgStyleStroke            = 2
	BgStyleFillAndStroke     = 3
	BgStyleNoFillAndNoStroke = 4

	// 服务端动态类型
	// https://info.bilibili.co/pages/viewpage.action?pageId=13358955
	DynTypeForward         = 1
	DynTypeDraw            = 2
	DynTypeWord            = 4
	DynTypeVideo           = 8
	DynTypeArticle         = 64
	DynTypeMusic           = 256
	DynTypeBangumi         = 512
	DynTypeCommonSquare    = 2048
	DynTypeCommonVertical  = 2049
	DynTypePGCBangumi      = 4097
	DynTypePGCMovie        = 4098
	DynTypePGCTv           = 4099
	DynTypePGCGuoChuang    = 4100
	DynTypePGCDocumentary  = 4101
	DynTypeLive            = 4200 // 转发中
	DynTypeMedialist       = 4300 // 转发中
	DynTypeAD              = 4301 // 透传卡
	DynTypeCheeseSeason    = 4302 // 转发中
	DynTypeCheeseBatch     = 4303
	DynTypeApplet          = 4305
	DynTypeSubscription    = 4306
	DynTypeLiveRcmd        = 4308
	DynTypeUGCSeason       = 4310
	DynTypeSubscriptionNew = 4311
	DynTypeBatch           = 4312
	DynTypeCourUp          = 4313
	DynTypeUGCSeasonShare  = 4314 // UGC合集分享卡（仅转发态）
	DynTypeNewTopicSet     = 4315 // 新话题-话题集订阅更新卡

	// 点赞bus类型
	BusTypeDraw    = "album"
	BusTypeDyn     = "dynamic"
	BusTypeVideo   = "archive"
	BusTypeArticle = "article"
	BusTypeAudio   = "audio"
	//BusTypePGC     = "bangumi"
	BusTypeCheese = "cheese"
	BusTypeAD     = "ad"

	// 高亮控制
	CtrlTypeAite    = 1
	CtrlTypeLottery = 2
	CtrlTypeVote    = 3
	CtrlTypeGoods   = 4

	PerSecond = 1
	PerMinute = PerSecond * 60
	PerHour   = PerMinute * 60

	// 评论业务类型
	CmtTypeAv        = 1
	CmtTypeDraw      = 11
	CmtTypeArticle   = 12
	CmtTypeMusic     = 14
	CmtTypeDynamic   = 17
	CmtTypeMedialist = 19
	CmtTypeAD        = 31
	CmtTypeCheese    = 33

	// 折叠类型
	FoldTypePublish  = int32(1)
	FoldTypeFrequent = int32(2)
	FoldTypeUnite    = int32(3)
	FoldTypeLimit    = int32(4)

	OrigInvisible   = 1
	OrigNoInvisible = 0

	// 图文标签
	DrawTagTypeCommon = 0
	DrawTagTypeGoods  = 1
	DrawTagTypeUser   = 2
	DrawTagTypeTopic  = 3
	DrawTagTypeLBS    = 4

	// 推荐关注类型
	NoFollow     = 1
	LowFollow    = 2
	RegionFollow = 3

	// 转发和分享类型
	DynShare   = 0
	DynForward = 1

	// 视频卡子类型
	VideoStypeDynamic      = 3
	VideoStypePlayback     = 2
	VideoStypeDynamicStory = 1

	// 附加底栏类型
	BottomBusinessGame    = 1
	BottomBusinessBiliCut = 2
	BottomBusinessBBQ     = 3
	BottomBusinessAutoPGC = 4

	// 查看更多排序类型
	UplistMoreSortTypeDefault = 0
	UplistMoreSortTypeRcmd    = 1
	UplistMoreSortTypeMore    = 2
	UplistMoreSortTypeNear    = 3

	// 角标颜色
	BgColorPink            = 1
	BgColorTransparentGray = 2
	BgColorGray            = 3
)

var (
	// 合作角标
	CooperationBadge = &api.VideoBadge{
		Text:             "合作",
		TextColor:        "#FFFFFF",
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
		TextColor:        "#FFFFFF",
		TextColorNight:   "#FFFFFF",
		BgColor:          "#FF6699",
		BgColorNight:     "#D44E7D",
		BorderColor:      "#FF6699",
		BorderColorNight: "#D44E7D",
		BgStyle:          BgStyleFill,
	}
	// 直播回放
	PlayBackBadge = &api.VideoBadge{
		Text:             "直播回放",
		TextColor:        "#FFFFFF",
		TextColorNight:   "#FFFFFF",
		BgColor:          "#FB7299",
		BgColorNight:     "#FB7299",
		BorderColor:      "#FB7299",
		BorderColorNight: "#FB7299",
		BgStyle:          BgStyleFill,
	}
	// story角标
	StoryBadge = &api.VideoBadge{
		Text:             "动态视频",
		TextColor:        "#FFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
)

// IsAv 视频卡
func (dyn *Dynamic) IsAv() bool {
	return dyn.Type == DynTypeVideo
}

// IsPGC PGC卡
func (dyn *Dynamic) IsPGC() bool {
	is := dyn.Type == DynTypePGCBangumi || dyn.Type == DynTypePGCMovie || dyn.Type == DynTypePGCTv ||
		dyn.Type == DynTypePGCGuoChuang || dyn.Type == DynTypePGCDocumentary || dyn.Type == DynTypeBangumi
	return is
}

// IsCourse 付费卡
func (dyn *Dynamic) IsCourse() bool {
	return dyn.Type == DynTypeCheeseSeason || dyn.Type == DynTypeCheeseBatch
}

// IsCheeseBatch 付费系列卡
func (dyn *Dynamic) IsCheeseBatch() bool {
	return dyn.Type == DynTypeCheeseBatch
}

func (dyn *Dynamic) IsForward() bool {
	return dyn.Type == DynTypeForward
}

func (dyn *Dynamic) IsWord() bool {
	return dyn.Type == DynTypeWord
}

func (dyn *Dynamic) IsDraw() bool {
	return dyn.Type == DynTypeDraw
}

func (dyn *Dynamic) IsArticle() bool {
	return dyn.Type == DynTypeArticle
}

func (dyn *Dynamic) IsMusic() bool {
	return dyn.Type == DynTypeMusic
}

func (dyn *Dynamic) IsCommon() bool {
	return dyn.Type == DynTypeCommonSquare || dyn.Type == DynTypeCommonVertical
}

func (dyn *Dynamic) IsCommonSquare() bool {
	return dyn.Type == DynTypeCommonSquare
}

func (dyn *Dynamic) IsCommonVertical() bool {
	return dyn.Type == DynTypeCommonVertical
}

// IsCheeseSeason 转发/分享 付费批次卡
func (dyn *Dynamic) IsCheeseSeason() bool {
	return dyn.Type == DynTypeCheeseSeason
}

// 转发/分享 直播间
func (dyn *Dynamic) IsLive() bool {
	return dyn.Type == DynTypeLive
}

// 转发/分享 播单收藏夹(仅转发态)
func (dyn *Dynamic) IsMedialist() bool {
	return dyn.Type == DynTypeMedialist
}

func (dyn *Dynamic) IsAD() bool {
	return dyn.Type == DynTypeAD
}

// 小程序卡
func (dyn *Dynamic) IsApplet() bool {
	return dyn.Type == DynTypeApplet
}

// 订阅卡
func (dyn *Dynamic) IsSubscription() bool {
	return dyn.Type == DynTypeSubscription
}

// 直播推荐卡
func (dyn *Dynamic) IsLiveRcmd() bool {
	return dyn.Type == DynTypeLiveRcmd
}

// 合集卡(更新)
func (dyn *Dynamic) IsUGCSeason() bool {
	return dyn.Type == DynTypeUGCSeason
}

// 新订阅卡
func (dyn *Dynamic) IsSubscriptionNew() bool {
	return dyn.Type == DynTypeSubscriptionNew
}

// 追漫卡
func (dyn *Dynamic) IsBatch() bool {
	return dyn.Type == DynTypeBatch
}

// 课堂 UP主主动触发更新卡
func (dyn *Dynamic) IsCourUp() bool {
	return dyn.Type == DynTypeCourUp
}

// UGC合集分享卡(仅转发态)
func (dyn *Dynamic) IsUGCSeasonShare() bool {
	return dyn.Type == DynTypeUGCSeasonShare
}

// 新话题-话题集订阅更新卡
func (dyn *Dynamic) IsNewTopicSet() bool {
	return dyn.Type == DynTypeNewTopicSet
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func BoolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func GetArchiveSType(s int64) api.VideoType {
	switch s {
	case VideoStypeDynamic:
		return api.VideoType_video_type_dynamic
	case VideoStypePlayback:
		return api.VideoType_video_type_playback
	case VideoStypeDynamicStory:
		return api.VideoType_video_type_story
	default:
		return api.VideoType_video_type_general
	}
}

func CanPlay(value int32) bool {
	switch value {
	case 1:
		return true
	default:
		return false
	}
}

func Int32ToBool(v int32) bool {
	return v != 0
}

func TranFoldType(foldType int) api.FoldType {
	switch int32(foldType) {
	case FoldTypePublish:
		return api.FoldType_FoldTypePublish
	case FoldTypeFrequent:
		return api.FoldType_FoldTypeFrequent
	case FoldTypeUnite:
		return api.FoldType_FoldTypeUnite
	case FoldTypeLimit:
		return api.FoldType_FoldTypeLimit
	default:
		return api.FoldType_FoldTypeZore
	}
}

func BadgeStyleFrom(style int8, text string) *api.VideoBadge {
	if text == "" {
		return nil
	}
	res := &api.VideoBadge{
		Text: text,
	}
	switch style {
	case BgColorPink:
		res.TextColor = "#FFFFFF"
		res.TextColorNight = "#E5E5E5"
		res.BgColor = "#FB7299"
		res.BgColorNight = "#BB5B76"
		res.BorderColor = "#FB7299"
		res.BorderColorNight = "#BB5B76"
		res.BgStyle = BgStyleFill
	case BgColorTransparentGray:
		res.TextColor = "#FFFFFF"
		res.TextColorNight = "#FFFFFF"
		res.BgColor = "#7F000000"
		res.BgColorNight = "#7F000000"
		res.BorderColor = "#A8FB7299"
		res.BorderColorNight = "#A8FB7299"
		res.BgStyle = BgStyleFill
	case BgColorGray:
		res.TextColor = "#FFFFFF"
		res.TextColorNight = "#FFFFFF"
		res.BgColor = "#B2000000"
		res.BgColorNight = "#B2000000"
		res.BorderColor = "#B2000000"
		res.BorderColorNight = "#B2000000"
		res.BgStyle = BgStyleFill
	}
	return res
}

func DynamicName(dynamicType int64) string {
	var name string
	switch dynamicType {
	case DynTypeForward:
		name = "转发卡"
	case DynTypeDraw:
		name = "图文卡"
	case DynTypeWord:
		name = "纯文字卡"
	case DynTypeVideo:
		name = "视频卡"
	case DynTypeArticle:
		name = "专栏卡"
	case DynTypeMusic:
		name = "音频卡"
	case DynTypeBangumi:
		name = "PGC卡"
	case DynTypeCommonSquare:
		name = "通用卡(方图)"
	case DynTypeCommonVertical:
		name = "通用卡(竖图)"
	case DynTypePGCBangumi:
		name = "PGC卡"
	case DynTypePGCMovie:
		name = "PGC卡"
	case DynTypePGCTv:
		name = "PGC卡"
	case DynTypePGCGuoChuang:
		name = "PGC卡"
	case DynTypePGCDocumentary:
		name = "PGC卡"
	case DynTypeLive: // 转发中
		name = "直播分享卡"
	case DynTypeMedialist: // 转发中
		name = "播单卡"
	case DynTypeAD: // 透传卡
		name = "广告卡"
	case DynTypeCheeseSeason: // 转发中
		name = "付费系列卡(仅转发)"
	case DynTypeCheeseBatch:
		name = "付费批次卡"
	case DynTypeApplet:
		name = "小程序卡"
	case DynTypeSubscription:
		name = "订阅卡(旧)"
	case DynTypeLiveRcmd:
		name = "直播推荐卡"
	case DynTypeUGCSeason:
		name = "合集卡"
	case DynTypeSubscriptionNew:
		name = "订阅卡(新)"
	case DynTypeBatch:
		name = "追漫卡"
	case DynTypeCourUp:
		name = "付费系列卡(UP主动触发)"
	case DynTypeUGCSeasonShare:
		name = "合集分享卡(仅转发)"
	case DynTypeNewTopicSet:
		name = "话题集卡"
	default:
		name = "未知卡"
	}
	return fmt.Sprintf("%s[%d]", name, dynamicType)
}

func PayAttrVal(a *archivegrpc.Arc) bool {
	if a.Pay == nil {
		return false
	}
	return a.AttrValV2(archivegrpc.AttrBitV2Pay) == archivegrpc.AttrYes && a.Pay.AttrVal(archivegrpc.PaySubTypeAttrBitSeason) == archivegrpc.AttrYes
}

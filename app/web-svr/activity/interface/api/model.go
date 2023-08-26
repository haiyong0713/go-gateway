package api

import (
	"encoding/json"
	"fmt"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	"strconv"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/rewards"
)

const (
	VIDEOLIKE  = 4
	VIDEO2     = 13
	PHONEVIDEO = 16
	ARTICLE    = 12

	TopicActType                 = 1 //话题活动页面
	_inlineActType               = 2 //页面inlinetab组件
	_menuType                    = 3 //首页menu页面
	_ogvType                     = 4 //ogv 页面
	_playerType                  = 6 //播放器内嵌活动页
	_spaceType                   = 7 //空间tab页面
	_ugcType                     = 8 //ugc播放页
	_topicActTypeStr             = "dynamic"
	WaitForCommit                = -3 //草稿箱
	CheckOffline                 = -2 //打回
	WaitForCheck                 = -1 //待审核
	WaitForOnline                = 0  //待上线
	OnlineState                  = 1  //page 上线
	OfflineState                 = 2  //page 下线
	_moduleOnlineState           = 1
	_moduleOfflineState          = 0
	_moduleClick                 = 1
	_moduleDynamic               = 2
	_moduleVideo                 = 3
	_moduleAct                   = 4
	_moduleVideoAvid             = 5
	_moduleVideoAct              = 6
	_moduleVideoDyn              = 7
	_moduleBanner                = 8
	_moduleStatement             = 9
	_moduleSingleDyn             = 10
	_moduleParticipation         = 11
	_moduleRecommend             = 12
	_moduleNavigation            = 13
	_moduleBaseHead              = 14 //版头组件
	_moduleResourceID            = 15 //资源小卡-id模式
	_moduleResourceAct           = 16 //资源小卡-act模式
	_moduleResourceDynamic       = 17 //资源小卡-动态模式
	_moduleInlineTab             = 18 //页面tab组件
	_moduleLive                  = 19 //直播卡
	_moduleCarouselImg           = 20 //轮播-图片模式
	_moduleIcon                  = 21 //图标
	_moduleNewVideoAvid          = 22 //新视频卡-id模式
	_moduleNewVideoAct           = 23 //新视频卡-act模式
	_moduleNewVideoDyn           = 24 //新视频卡-dyn模式
	_moduleEditor                = 25 //编辑推荐卡
	_moduleRcmdVertical          = 26 //推荐用户-竖卡
	_moduleSelect                = 27 //筛选组件
	_moduleProgress              = 28 //进度条
	_moduleResourceRole          = 29 //资源小卡-角色剧集模式
	_moduleCarouselWord          = 30 //轮播-文字模式
	_moduleTimelineIDs           = 31 //时间轴-ids
	_moduleTimelineSource        = 32 //时间轴-数据源模式
	_moduleOgvSeasonID           = 33 //ogv 剧集卡id模式
	_moduleOgvSeasonSource       = 34 //ogv 剧集卡source模式
	ActOrderLikeNum              = 1
	ActOrderCtimeNum             = 2
	ActOrderStochasticNum        = 3
	ActOrderEsLikeNum            = 4
	_moduleCardSingle            = 1
	_moduleCardDouble            = 2
	_moduleCardThree             = 3
	AttrIsLast                   = uint(0)
	AttrIsAutoPlay               = uint(1)
	AttrIsHideTitle              = uint(2)
	AttrIsHideMore               = uint(3)
	AttrIsDisplayUser            = uint(4)  //版头部分是否展示发起人
	AttrIsDisplayVideoIcon       = uint(5)  //资源小卡视频类型是否展示角标
	AttrIsDisplayArticleIcon     = uint(6)  //资源小卡专栏类型是否展示角标
	AttrIsDisplayPgcIcon         = uint(7)  //资源小卡pgc类型是否展示角标&ogv剧集卡是否展示付费角标
	AttrIsDisplayButton          = uint(8)  //inline tab 是否展示收起按钮
	AttrStatementIsDisplayButton = uint(9)  //文本组件是否展示收起按钮 0-展示 1-不展示
	AttrIsDisplayOp              = uint(10) //是否展示三点操作按钮
	AttrIsDisplayNum             = uint(11) //是否展示当前进度数值&ogv是否展示评分
	AttrIsDisplayNodeNum         = uint(12) //是否展示设置的节点数值
	AttrIsDisplayDesc            = uint(13) //是否展示节点描述&ogv卡是否展示副标题
	AttrIsDisplayRecommend       = uint(14) //是否展示推荐语

	//page.attr
	AttrIsNotNightModule = uint(5) //是否不需要支持夜间模式适配
	AttrModuleYes        = int64(1)
	_dyChoiceType        = 1
	_dyChoice            = "PICKED"
	// mix m_type
	MixTypeRcmd     = 0  //mid类型
	MixAvidType     = 1  //ugc-avid类型
	MixEpidType     = 2  //pgc-epid
	MixCvidType     = 3  //专栏-cvid
	MixInlineType   = 4  // inline tab page类型
	MixCarouselImg  = 5  //轮播-图片
	MixCarouselWord = 6  //轮播-文字
	MixIconImg      = 7  //图标-图片
	MixFolder       = 8  //播单
	MixRcmdVertical = 9  //推荐用户-竖卡
	MixTimelinePic  = 11 //时间轴-图片模式
	MixTimelineText = 12 //时间轴-文字模式
	MixTimeline     = 13 //时间轴-图文模式
	MixOgvSsid      = 14 //ogv ssid类型
	// native page class
	CommonPage       = 0 // 普通话题活动类型
	FeedPage         = 1 //天马落地页面
	BasePage         = 2 //基础组件-天马&普通话题页公共组件
	FeedBaseModule   = 3 // 天马-基础组件
	CommonBaseModule = 4 //普通话题 -基础组件
	//native page attr
	AttrForbid       = uint(1) //禁止上榜
	AttrDisplayCount = uint(2) //是否不隐藏浏览量、讨论量
	// 参与组件投稿类型 0:动态 1.视频 2.专栏
	_partDynamic = 0
	_partVideo   = 1
	_partArticle = 2
	_commonJump  = 0  //普通跳转
	_followWith  = 1  //关注
	_catchUp     = 2  //追番
	_reserve     = 3  //预约
	_redirect    = 5  //跳转链接
	_layerImage  = 10 //图片模式
	_layerLink   = 11 //链接模式
	_app         = 20 //拉起APP
	_progress    = 30 //进度数据
	// 打卡规则
	SubRuleBitCount = 0 //计数规则属性位
	SubRuleBitStart = 1 //统计开始规则属性位
	_pendant        = 4 //挂件领取
	// tab module category
	TabPageCategory = 1
	TabUrlCategory  = 2
	//page 发起类型
	// 运营配置活动
	PageFromSystem = 0
	// up主发起活动
	PageFromUid = 1
	//ts_page state
	TsWaitCheck    = 0 //待审核
	TsCheckOnline  = 1 //审核通过
	TsCheckOffline = 2 //审核不通过
	// ts_page audit_type
	TsAuditAuto   = "auto"   //自动审核
	TsAuditManual = "manual" //人工审核
	// 查看更多方式
	MoreJump        = 0
	MoreSupernatant = 1
	MoreExpand      = 2
	//时间轴 精确0:年 1:月 2: 日 3:时 4:分 5:秒
	TimeSortYear  = 0
	TimeSortMonth = 1
	TimeSortDay   = 2
	TimeSortHour  = 3
	TimeSortMin   = 4
	TimeSortSec   = 5
	//时间轴节点类型 0:文本 1:时间节点
	AxisText = 0
	AxisTime = 1
)

// isTopicAct .
func (nat *NativePage) IsTopicAct() bool {
	return nat.Type == TopicActType
}

func (nat *NativePage) IsUpTopicAct() bool {
	return nat.Type == TopicActType && nat.FromType == PageFromUid
}

// IsInlineAct.
func (nat *NativePage) IsInlineAct() bool {
	return nat.Type == _inlineActType
}

// _menuType
func (nat *NativePage) IsMenuAct() bool {
	return nat.Type == _menuType
}

func (nat *NativePage) IsOgvAct() bool {
	return nat.Type == _ogvType
}

func (nat *NativePage) IsSpaceAct() bool {
	return nat.Type == _spaceType
}

func (nat *NativePage) IsUgcAct() bool {
	return nat.Type == _ugcType
}

func (nat *NativePage) IsPlayerAct() bool {
	return nat.Type == _playerType
}

// IsOnline .
func (nat *NativePage) IsOnline() bool {
	return nat.State == OnlineState
}

func (nat *NativePage) IsWaitForCheck() bool {
	return nat.State == WaitForCheck
}

func (nat *NativePage) IsCheckOffline() bool {
	return nat.State == CheckOffline
}

func (nat *NativePage) IsWaitOnline() bool {
	return nat.State == WaitForOnline
}

func (nat *NativePage) TypeToString() string {
	if nat.IsTopicAct() {
		return _topicActTypeStr
	}
	return ""
}

// IsAttrDisplayCounty 是否展示浏览量、讨论量
func (nat *NativePage) IsAttrForbid() int64 {
	return (nat.Attribute >> AttrForbid) & int64(1)
}

// IsAttrDisplayCounty 是否展示浏览量、讨论量
func (nat *NativePage) IsAttrDisplayCounty() int64 {
	return (nat.Attribute >> AttrDisplayCount) & int64(1)
}

// IsOffline .
func (nat *NativePage) IsOffline() bool {
	return nat.State == OfflineState
}

// IsAttrNotNightModule .
func (nat *NativePage) IsAttrNotNightModule() int64 {
	return (nat.Attribute >> AttrIsNotNightModule) & int64(1)
}

// IsOnline .
func (mde *NativeModule) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOffline .
func (mde *NativeModule) IsOffline() bool {
	return mde.State == _moduleOfflineState
}

// IsClick .
func (mde *NativeModule) IsClick() bool {
	return mde.Category == _moduleClick
}

// IsDynamic .
func (mde *NativeModule) IsDynamic() bool {
	return mde.Category == _moduleDynamic
}

// IsVideo .
func (mde *NativeModule) IsVideo() bool {
	return mde.Category == _moduleVideo
}

// IsAct .
func (mde *NativeModule) IsAct() bool {
	return mde.Category == _moduleAct
}

// IsVideoAvid .
func (mde *NativeModule) IsVideoAvid() bool {
	return mde.Category == _moduleVideoAvid
}

// IsResourceID .
func (mde *NativeModule) IsResourceID() bool {
	return mde.Category == _moduleResourceID
}

// IsNewVideoID .
func (mde *NativeModule) IsNewVideoID() bool {
	return mde.Category == _moduleNewVideoAvid
}

// IsNewVideoAct .
func (mde *NativeModule) IsNewVideoAct() bool {
	return mde.Category == _moduleNewVideoAct
}

// IsNewVideoDyn .
func (mde *NativeModule) IsNewVideoDyn() bool {
	return mde.Category == _moduleNewVideoDyn
}

// IsResourceAct .
func (mde *NativeModule) IsResourceAct() bool {
	return mde.Category == _moduleResourceAct
}

// IsResourceDyn .
func (mde *NativeModule) IsResourceDyn() bool {
	return mde.Category == _moduleResourceDynamic
}

// IsInlineTab .
func (mde *NativeModule) IsInlineTab() bool {
	return mde.Category == _moduleInlineTab
}

// IsSelect
func (mde *NativeModule) IsSelect() bool {
	return mde.Category == _moduleSelect
}

func (mde *NativeModule) IsLive() bool {
	return mde.Category == _moduleLive
}

func (mde *NativeModule) IsCarouselImg() bool {
	return mde.Category == _moduleCarouselImg
}

func (mde *NativeModule) IsCarouselWord() bool {
	return mde.Category == _moduleCarouselWord
}

func (mde *NativeModule) IsOgvSeasonID() bool {
	return mde.Category == _moduleOgvSeasonID
}

func (mde *NativeModule) IsOgvSeasonSource() bool {
	return mde.Category == _moduleOgvSeasonSource
}

func (mde *NativeModule) IsIcon() bool {
	return mde.Category == _moduleIcon
}

// IsBanner .
func (mde *NativeModule) IsBanner() bool {
	return mde.Category == _moduleBanner
}

// IsStatement .
func (mde *NativeModule) IsStatement() bool {
	return mde.Category == _moduleStatement
}

// IsSingleDyn .
func (mde *NativeModule) IsSingleDyn() bool {
	return mde.Category == _moduleSingleDyn
}

// IsEditor .
func (mde *NativeModule) IsEditor() bool {
	return mde.Category == _moduleEditor
}

// IsResourceRole .
func (mde *NativeModule) IsResourceRole() bool {
	return mde.Category == _moduleResourceRole
}

// IsTimelineIDs .
func (mde *NativeModule) IsTimelineIDs() bool {
	return mde.Category == _moduleTimelineIDs
}

// IsTimelineSource .
func (mde *NativeModule) IsTimelineSource() bool {
	return mde.Category == _moduleTimelineSource
}

// IsCardSingle .
func (mde *NativeModule) IsCardSingle() bool {
	return mde.CardStyle == _moduleCardSingle
}

// IsCardDouble .
func (mde *NativeModule) IsCardDouble() bool {
	return mde.CardStyle == _moduleCardDouble
}

// IsCardThree .
func (mde *NativeModule) IsCardThree() bool {
	return mde.CardStyle == _moduleCardThree
}

// IsVideoAct .
func (mde *NativeModule) IsVideoAct() bool {
	return mde.Category == _moduleVideoAct
}

// IsVideoDyn .
func (mde *NativeModule) IsVideoDyn() bool {
	return mde.Category == _moduleVideoDyn
}

// IsPart .
func (mde *NativeModule) IsPart() bool {
	return mde.Category == _moduleParticipation
}

// IsRecommend .
func (mde *NativeModule) IsRecommend() bool {
	return mde.Category == _moduleRecommend
}

// IsRcmdVertical .
func (mde *NativeModule) IsRcmdVertical() bool {
	return mde.Category == _moduleRcmdVertical
}

// IsProgress .
func (mde *NativeModule) IsProgress() bool {
	return mde.Category == _moduleProgress
}

// IsNavigation
func (mde *NativeModule) IsNavigation() bool {
	return mde.Category == _moduleNavigation
}

// IsBaseHead .
func (mde *NativeModule) IsBaseHead() bool {
	return mde.Category == _moduleBaseHead
}

// IsAttrLast .
func (mde *NativeModule) IsAttrLast() int64 {
	return (mde.Attribute >> AttrIsLast) & int64(1)
}

// IsAttrAutoPlay .
func (mde *NativeModule) IsAttrAutoPlay() int64 {
	return (mde.Attribute >> AttrIsAutoPlay) & int64(1)
}

// IsAttrHideTitle .
func (mde *NativeModule) IsAttrHideTitle() int64 {
	return (mde.Attribute >> AttrIsHideTitle) & int64(1)
}

// IsAttrHideMore. 0:展示查看更多 1:隐藏查看更多
func (mde *NativeModule) IsAttrHideMore() int64 {
	return (mde.Attribute >> AttrIsHideMore) & int64(1)
}

// IsAttrDisplayUser. 0:不展示 1:展示
func (mde *NativeModule) IsAttrDisplayUser() int64 {
	return (mde.Attribute >> AttrIsDisplayUser) & int64(1)
}

// IsAttrDisplayVideoIcon .
func (mde *NativeModule) IsAttrDisplayVideoIcon() int64 {
	return (mde.Attribute >> AttrIsDisplayVideoIcon) & int64(1)
}

// IsAttrDisplayPgcIcon .
func (mde *NativeModule) IsAttrDisplayPgcIcon() int64 {
	return (mde.Attribute >> AttrIsDisplayPgcIcon) & int64(1)
}

// IsAttrDisplayButton .
func (mde *NativeModule) IsAttrDisplayButton() int64 {
	return (mde.Attribute >> AttrIsDisplayButton) & int64(1)
}

// IsAttrDisplayOp .
func (mde *NativeModule) IsAttrDisplayOp() int64 {
	return (mde.Attribute >> AttrIsDisplayOp) & int64(1)
}

func (mde *NativeModule) IsAttrDisplayRecommend() int64 {
	return (mde.Attribute >> AttrIsDisplayRecommend) & int64(1)
}

// IsAttrDisplayArticleIcon .
func (mde *NativeModule) IsAttrDisplayArticleIcon() int64 {
	return (mde.Attribute >> AttrIsDisplayArticleIcon) & int64(1)
}

// IsAttrStatementDisplayButton .
func (mde *NativeModule) IsAttrStatementDisplayButton() int64 {
	return (mde.Attribute >> AttrStatementIsDisplayButton) ^ int64(1)
}

// IsAttrDisplayNum 是否展示当前进度数值&ogv是否展示评分.
func (mde *NativeModule) IsAttrDisplayNum() int64 {
	return (mde.Attribute >> AttrIsDisplayNum) & int64(1)
}

func (mde *NativeModule) IsAttrDisplayNodeNum() int64 {
	return (mde.Attribute >> AttrIsDisplayNodeNum) & int64(1)
}

func (mde *NativeModule) IsAttrDisplayDesc() int64 {
	return (mde.Attribute >> AttrIsDisplayDesc) & int64(1)
}

func (mde *NativeModule) ColorsUnmarshal() *Colors {
	ry := &Colors{}
	if mde.Colors != "" {
		json.Unmarshal([]byte(mde.Colors), ry)
	}
	return ry
}

func (mde *NativeModule) ConfUnmarshal() *ConfSort {
	ry := &ConfSort{}
	if mde.ConfSort != "" {
		json.Unmarshal([]byte(mde.ConfSort), ry)
	}
	return ry
}

func (mde *NativeMixtureExt) RemarkUnmarshal() *MixReason {
	ry := &MixReason{}
	if mde.Reason != "" {
		json.Unmarshal([]byte(mde.Reason), ry)
	}
	return ry
}

func (item *MixReason) JoinCurrentTab() string {
	if item.Type != "" && item.LocationKey != "" {
		return fmt.Sprintf("%s-%s", item.Type, item.LocationKey)
	}
	return ""
}

// IsOffline .
func (mde *NativeClick) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOffline .
func (mde *NativeClick) IsOffline() bool {
	return mde.State == _moduleOfflineState
}

// IsOnline .
func (mde *NativeAct) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOffline .
func (mde *NativeAct) IsOffline() bool {
	return mde.State == _moduleOfflineState
}

// IsOnline .
func (mde *NativeDynamicExt) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOffline .
func (mde *NativeDynamicExt) IsOffline() bool {
	return mde.State == _moduleOfflineState
}

// IsOnline .
func (mde *NativeVideoExt) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOnline .
func (mde *NativeMixtureExt) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOnline m_type: 0动态，1视频，2专栏.
func (mde *NativeParticipationExt) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsOffline m_type: 0动态，1视频，2专栏.
func (mde *NativeParticipationExt) IsOffline() bool {
	return mde.State == _moduleOfflineState
}

// IsPartDynamic .
func (mde *NativeParticipationExt) IsPartDynamic() bool {
	return mde.MType == _partDynamic
}

// IsPartVideo .
func (mde *NativeParticipationExt) IsPartVideo() bool {
	return mde.MType == _partVideo
}

// IsPartArticle .
func (mde *NativeParticipationExt) IsPartArticle() bool {
	return mde.MType == _partArticle
}

// IsOffline .
func (mde *NativeVideoExt) IsOffline() bool {
	return mde.State == _moduleOfflineState
}

// IsCtimeType .
func (mde *NativeVideoExt) IsCtimeType() bool {
	return mde.SortType == ActOrderCtimeNum
}

// IsLikeType .
func (mde *NativeVideoExt) IsLikeType() bool {
	return mde.SortType == ActOrderLikeNum
}

// IsStochasticType .
func (mde *NativeVideoExt) IsStochasticType() bool {
	return mde.SortType == ActOrderStochasticNum
}

// IsEsLikesType .
func (mde *NativeVideoExt) IsEsLikesType() bool {
	return mde.SortType == ActOrderEsLikeNum
}

// JoinDyTypes is need to del.
func (mde *NativeDynamicExt) JoinDyTypes() (ty string) {
	if mde.ClassType == _dyChoiceType {
		ty = _dyChoice + "," + strconv.FormatInt(mde.ClassID, 10)
	} else {
		if mde.SelectType > 0 {
			ty = strconv.FormatInt(mde.SelectType, 10)
		}
	}
	return
}

// JoinMultiDyTypes
func (mde *NativeDynamicExt) JoinMultiDyTypes() (ty string, isSingle bool) {
	if mde.ClassType == _dyChoiceType {
		ty = _dyChoice + "," + strconv.FormatInt(mde.ClassID, 10)
		isSingle = true //精选只支持单选
	} else {
		if mde.SelectType > 0 {
			ty = strconv.FormatInt(mde.SelectType, 10)
		} else {
			isSingle = true //全选只支持单选
		}
	}
	return
}

func (sub *Subject) IsVideoCollection() bool {
	return sub.Type == VIDEO2 || sub.Type == PHONEVIDEO
}

func (sub *Subject) IsVideoLike() bool {
	return sub.Type == VIDEOLIKE
}

func (sub *Subject) IsArticle() bool {
	return sub.Type == ARTICLE
}

// 是否普通跳转
func (mde *NativeClick) IsCommonJump() bool {
	return mde.Type == _commonJump
}

// 是否关注
func (mde *NativeClick) IsFollow() bool {
	return mde.Type == _followWith
}

// 是否追番
func (mde *NativeClick) IsCatchUp() bool {
	return mde.Type == _catchUp
}

// 是否领取
func (mde *NativeClick) IsReserve() bool {
	return mde.Type == _reserve
}

// 是否领取
func (mde *NativeClick) IsRedirect() bool {
	return mde.Type == _redirect
}

// 是否预约
func (mde *NativeClick) IsPendant() bool {
	return mde.Type == _pendant
}

// 是否进度条
func (mde *NativeClick) IsProgress() bool {
	return mde.Type == _progress
}

// 是否浮层-图片模式
func (mde *NativeClick) IsLayerImage() bool {
	return mde.Type == _layerImage
}

// 是否浮层-链接模式
func (mde *NativeClick) IsLayerLink() bool {
	return mde.Type == _layerLink
}

// 是否配置链接-拉起APP
func (mde *NativeClick) IsAPP() bool {
	return mde.Type == _app
}

// tab是否有效
func (mde *NativeActTab) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// module是否有效
func (mde *NativeTabModule) IsOnline() bool {
	return mde.State == _moduleOnlineState
}

// IsTabPage .
func (mde *NativeTabModule) IsTabPage() bool {
	return mde.Category == TabPageCategory
}

// IsTabUrl .
func (mde *NativeTabModule) IsTabUrl() bool {
	return mde.Category == TabUrlCategory
}

func (out *ReserveRule) DeepCopyFromActSubjectProtocol(in *like.SubjectRule) {
	out.ID = in.ID
	out.Type = in.Type
	out.TypeIDs = in.TypeIds
	out.Tags = in.Tags
	out.Attribute = in.Attribute
	out.State = in.State
	out.Stime = in.Stime
	out.Etime = in.Etime
	return
}

func (m *RewardsSendAwardReply) ToSentInfo() *rewards.AwardSentInfo {
	res := &rewards.AwardSentInfo{
		Mid:          m.Mid,
		AwardId:      m.AwardId,
		AwardName:    m.Name,
		ActivityId:   m.ActivityId,
		ActivityName: m.ActivityName,
		Type:         m.Type,
		IconUrl:      m.Icon,
		SentTime:     xtime.Time(time.Now().Unix()),
		ExtraInfo:    m.ExtraInfo,
	}
	return res
}

func (m *Risk) ToBase(mid int64, action string) *riskmdl.Base {
	risk := &riskmdl.Base{
		Buvid:     m.Buvid,
		Origin:    m.Origin,
		Referer:   m.Referer,
		IP:        m.Ip,
		Ctime:     time.Now().Format("2006-01-02 15:04:05"),
		UserAgent: m.UserAgent,
		Build:     m.Build,
		Platform:  m.Platform,
		Action:    action,
		MID:       mid,
		API:       m.Api,
		EsTime:    time.Now().Unix(),
	}
	return risk
}

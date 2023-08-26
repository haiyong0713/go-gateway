package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	VIDEOLIKE  = 4
	VIDEO2     = 13
	PHONEVIDEO = 16
	ARTICLE    = 12

	TopicActType                 = 1  //话题活动页面
	_inlineActType               = 2  //页面inlinetab组件
	_menuType                    = 3  //首页menu页面
	_ogvType                     = 4  //ogv 页面
	_playerType                  = 6  //播放器内嵌活动页
	_spaceType                   = 7  //空间tab页面
	_ugcType                     = 8  //ugc播放页
	BottomType                   = 9  //首页底部tab页面
	LiveTabType                  = 10 //直播间tab页面
	NewactType                   = 11 //新活动页
	_topicActTypeStr             = "dynamic"
	WaitForCommit                = -3 //草稿箱
	CheckOffline                 = -2 //打回
	WaitForCheck                 = -1 //待审核
	WaitForOnline                = 0  //待上线
	OnlineState                  = 1  //page 上线
	OfflineState                 = 2  //page 下线
	_moduleOnlineState           = 1
	_moduleOfflineState          = 0
	ModuleClick                  = 1
	ModuleDynamic                = 2
	ModuleVideo                  = 3
	ModuleAct                    = 4
	_moduleVideoAvid             = 5
	_moduleVideoAct              = 6
	_moduleVideoDyn              = 7
	_moduleBanner                = 8
	ModuleStatement              = 9
	_moduleSingleDyn             = 10
	ModuleParticipation          = 11
	ModuleRecommend              = 12
	ModuleNavigation             = 13
	ModuleBaseHead               = 14 //版头组件
	ModuleResourceID             = 15 //资源小卡-id模式
	ModuleResourceAct            = 16 //资源小卡-act模式
	ModuleResourceDynamic        = 17 //资源小卡-动态模式
	ModuleInlineTab              = 18 //页面tab组件
	ModuleLive                   = 19 //直播卡
	ModuleCarouselImg            = 20 //轮播-图片模式
	ModuleIcon                   = 21 //图标
	ModuleNewVideoAvid           = 22 //新视频卡-id模式
	ModuleNewVideoAct            = 23 //新视频卡-act模式
	ModuleNewVideoDyn            = 24 //新视频卡-dyn模式
	ModuleEditor                 = 25 //编辑推荐卡
	ModuleRcmdVertical           = 26 //推荐用户-竖卡
	ModuleSelect                 = 27 //筛选组件
	ModuleProgress               = 28 //进度条
	ModuleResourceRole           = 29 //资源小卡-角色剧集模式
	ModuleCarouselWord           = 30 //轮播-文字模式
	ModuleTimelineIDs            = 31 //时间轴-ids
	ModuleTimelineSource         = 32 //时间轴-数据源模式
	ModuleOgvSeasonID            = 33 //ogv 剧集卡id模式
	ModuleOgvSeasonSource        = 34 //ogv 剧集卡source模式
	ModuleResourceOrigin         = 35 //资源小卡-外接数据源模式
	ModuleReply                  = 36 //评论组件
	ModuleEditorOrigin           = 37 //编辑卡-外接数据源
	ModuleActCapsule             = 38 //相关活动-胶囊样式
	ModuleCarouselSource         = 39 //轮播-数据源模式
	ModuleRcmdSource             = 40 //推荐用户-横卡-数据源模式
	ModuleRcmdVerticalSource     = 41 //推荐用户-竖卡-数据源模式
	ModuleGame                   = 42 //游戏组件
	ModuleBaseHoverButton        = 43 //自定义悬浮按钮
	ModuleReserve                = 44 //预约组件
	VoteModule                   = 45 //投票组件
	ModuleNewactHeader           = 46 //新活动页-版头
	ModuleNewactAward            = 47 //新活动页-活动奖励
	ModuleNewactStatement        = 48 //新活动页-文本
	BottomButtonModule           = 49 //吸底按钮
	MatchMedalModule             = 50 //奖牌榜
	MatchEventModule             = 51 //焦点赛事
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
	AttrIsDisplayNum             = uint(11) //是否展示当前进度数值&ogv是否展示评分&轮播图片：首页顶栏跟随图片变化&投票组件-是否展示选项得票
	AttrIsDisplayNodeNum         = uint(12) //是否展示设置的节点数值&资源小卡-只展示直播中
	AttrIsDisplayDesc            = uint(13) //是否展示节点描述&ogv卡是否展示副标题&轮播图片-轮播组件滑出屏幕后顶栏配置样式消失
	AttrIsDisplayRecommend       = uint(14) //是否展示推荐语&推荐用户是否展示推荐字段
	AttrIsShareImage             = uint(15) //是否开启长按保存图片
	AttrIsDisplayH5Header        = uint(16) //h5是否展示版头
	AttrIsDisplayUpIcon          = uint(17) //预约组件-是否展示up主头像昵称
	AttrIsCloseSubscribeBtn      = uint(18) //是否关闭展示订阅按钮
	AttrIsCloseViewNum           = uint(19) //是否关闭展示浏览量、讨论量

	AttrModuleYes = int64(1)
	_dyChoiceType = 1
	_dyChoice     = "PICKED"
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
	MixOgvFilm      = 15 //ogv 片单
	MixWeb          = 16 //网页
	MixProduct      = 17 //企业号-商品卡
	MixActivity     = 18 //相关活动-活动
	MixLive         = 19 //直播间id
	MixGame         = 20 //游戏id
	MixUpReserve    = 21 //up主预约id
	MixRankIcon     = 22 //排行榜icon
	MixMatchEvent   = 23 //焦点赛事-赛事id
	// native page class
	CommonPage       = 0 // 普通话题活动类型
	FeedPage         = 1 //天马落地页面
	BasePage         = 2 //基础组件-天马&普通话题页公共组件
	FeedBaseModule   = 3 // 天马-基础组件
	CommonBaseModule = 4 //普通话题 -基础组件
	//native page attr
	AttrForbid           = uint(1)   //禁止上榜
	AttrForbidNum        = int64(2)  //禁止上榜num
	AttrDisplayCount     = uint(2)   //是否不隐藏浏览量、讨论量
	AttrIsNotNightModule = uint(5)   //是否不需要支持夜间模式适配
	AttrIsNotNightNum    = int64(32) //是否不需要支持夜间模式适配对应num
	AttrIsWhiteSwitch    = uint(6)   //是否开启仅白名单用户可见
	AttrMaxNum           = 0xffffffff
	// 参与组件投稿类型 0:动态 1.视频 2.专栏
	PartDynamic         = 0
	PartVideo           = 1
	PartArticle         = 2
	_commonJump         = 0  //普通跳转
	_followWith         = 1  //关注
	_catchUp            = 2  //追番
	_reserve            = 3  //预约
	_pendant            = 4  //挂件领取
	Redirect            = 5  //跳转链接
	_actReserve         = 6  //活动项目
	_onlyImage          = 7  //仅展示图片
	_buyCoupon          = 8  //会员购票务想买
	_cartoon            = 9  //追漫
	_layerImage         = 10 //图片模式
	_layerLink          = 11 //链接模式
	_layerInterface     = 12 //接口模式
	_app                = 20 //拉起APP
	_interface          = 21 //点击区域-接口模式
	_progress           = 30 //进度数据
	_staticProcess      = 31 //进度数据-静态
	VoteButton          = 40 //投票组件-投票按钮
	VoteProcess         = 41 //投票组件-投票进度
	VoteUser            = 42 //投票组件-用户剩余票数
	_clickUpAppointment = 50 //自定义按钮-up预约
	ClickPublishBtn     = 60 //投稿按钮

	// 打卡规则
	SubRuleBitCount = 0 //计数规则属性位
	SubRuleBitStart = 1 //统计开始规则属性位
	// tab module category
	TabPageCategory = 1
	TabUrlCategory  = 2
	//page 发起类型
	PageFromSystem           = 0 // 运营配置活动
	PageFromUid              = 1 // up主发起活动
	PageFromUpgSourceHot     = 2 //话题自动升级-热度标准
	PageFromUpgSourceDiscuss = 3 //话题自动升级-讨论蹿升
	PageFromUpgSourceAI      = 4 //话题自动升级-AI召回
	PageFromUpgSourceVideo   = 5 //话题自动升级-视频数据源生成
	PageFromNewactCollect    = 6 //新活动页-视频收集
	PageFromNewactVote       = 7 //新活动页-视频投票
	PageFromNewactShoot      = 8 //新活动页-拍摄活动
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
	//资源小卡外接数源类型
	RDOBusinessCommodity = 1 //企业号-商品数据源
	RDOBusinessIDs       = 2 //企业号-稿件类数据源
	RDOOgvWid            = 3 //OGV运营后台-WID
	RDBMustsee           = 4 //编辑推荐卡-入站必刷
	RDBWeek              = 5 //每周必看数据源
	RDBLive              = 6 //资源小卡-直播间id
	RDBRank              = 7 //排行榜类型
	RDBChannel           = 8 //垂类id
	//首页tab背景类型
	BgTypeImage = 1
	BgTypeColor = 2
	//组件解锁类型 start
	NeedUnLock  = 1 //解锁后展示
	UnLockTime  = 1 //解锁类型	:时间
	UnLockOrder = 2 //解锁类型	:预约数据源
	//未解锁时 1:不展示 2:不可点
	NotDisplay = 1
	NotClick   = 2
	//组件解锁类型 end
	// progress stat_type
	StatTypeAllPoint    = "all_point"    //活动总积分
	StatTypeAllMember   = "all_member"   //活动总人数
	StatTypeUserRule    = "user_rule"    //用户单行为
	StatTypeUserFormula = "user_formula" //用户公式
	//select 定位类型
	SelectWeek = "week"
	SelectMiao = "miao"
	// ClickSort style
	ClickStyleBnj         = "bnj_2021_ar"
	ClickStyleBnjTaskGame = "bnj_2021_task_game"
	//投票组件圆角｜方角
	ClickStyleCircle = "circle"
	ClickStyleSquare = "square"
	//自定义点击组件-静态进度条数据源类型 0用户积分统计 1:活动报名量 2:任务统计 3:抽奖数量
	ProcessUserStatics = 0
	ProcessRegister    = 1
	ProcessTaskStatics = 2
	ProcessLottery     = 3
	ProcessScore       = 4
	//自定义点击组件-静态进度条数据源类型
	// user_space state
	USpaceOfflineNormal    = "offline_normal"     //下线-正常关闭
	USpaceOnline           = "online"             //上线
	USpaceWaitingOnline    = "waiting_online"     //待上线
	USpaceOfflineBindFail  = "offline_bind_fail"  //下线-绑定失败
	USpaceOfflineAuditFail = "offline_audit_fail" //下线-审核失败
	// ConfSort sort_type
	SortTypeCtime  = "ctime"  //创建时间排序
	SortTypeRandom = "random" //随机排序
	// ConfSort source_type
	SourceTypeActUp      = "act_up"   //活动的up主
	SourceTypeRank       = "rank"     //排行榜
	SourceTypeVoteAct    = "act_vote" //活动投票
	SourceTypeVoteUp     = "up_vote"  //UP主投票
	BizOgvType           = 1
	BizArtType           = 2
	BizUgcType           = 3
	BizBusinessCommodity = 4
	BizSeasonType        = 5
	BizWebType           = 6
	BizOgvFilmType       = 7
	BizLive              = 8
	// ConfSort button_type
	BtTypeAppoint    = "appoint_origin" //预约数据源
	BtTypeActProject = "act_project"    //活动项目
	BtTypeLink       = "link"           //跳转链接
	// 编辑推荐位置
	PosUp       = "up"
	PosView     = "view"
	PosPubTime  = "pub_time"
	PosLike     = "like"
	PosDanmaku  = "danmaku"
	PosDuration = "duration"
	PosFollow   = "follow"
	PosViewStat = "view_state"
	//排行版类型-综合
	PosComprehensive = "comprehensive"
	//排行版类型-投币
	PosCoin = "coin"
	//排行版类型-分享
	PosShare = "share"
	//inline-tab&select 默认tab生效类型1:默认生效 2:定时生效
	DefTypeTimely = 1
	DefTypeTiming = 2
	//文本类型
	StatementNewactTask        = 1 //新活动页-任务玩法
	StatementNewactRule        = 2 //新活动页-规则说明
	StatementNewactDeclaration = 3 //新活动页-平台声明
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

// IsTabAct内嵌类型.
func (nat *NativePage) IsTabAct() bool {
	switch nat.Type {
	case _inlineActType, _menuType, BottomType, _ogvType:
		return true
	default:
		return false
	}
}

// _menuType
func (nat *NativePage) IsMenuAct() bool {
	return nat.Type == _menuType
}

func (nat *NativePage) IsBottomAct() bool {
	return nat.Type == BottomType
}

func (nat *NativePage) IsLiveTabAct() bool {
	return nat.Type == LiveTabType
}

func (nat *NativePage) ConfSetUnmarshal() *ConfSet {
	if nat.ConfSet != "" {
		ry := &ConfSet{}
		if err := json.Unmarshal([]byte(nat.ConfSet), ry); err == nil {
			return ry
		}
	}
	return &ConfSet{}
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

func (nat *NativePage) IsNewact() bool {
	return nat.Type == NewactType
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

// IsAttrForbid 禁止上榜
func (nat *NativePage) IsAttrForbid() int64 {
	return (nat.Attribute >> AttrForbid) & int64(1)
}

// IsAttrDisplayCounty 是否展示浏览量、讨论量
func (nat *NativePage) IsAttrDisplayCounty() int64 {
	return (nat.Attribute >> AttrDisplayCount) & int64(1)
}

// IsAttrWhiteSwitch 是否开启白名单可见
func (nat *NativePage) IsAttrWhiteSwitch() int64 {
	return (nat.Attribute >> AttrIsWhiteSwitch) & int64(1)
}

// IsOffline
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
	return mde.Category == ModuleClick
}

// IsDynamic .
func (mde *NativeModule) IsDynamic() bool {
	return mde.Category == ModuleDynamic
}

// IsVideo .
func (mde *NativeModule) IsVideo() bool {
	return mde.Category == ModuleVideo
}

// IsAct .
func (mde *NativeModule) IsAct() bool {
	return mde.Category == ModuleAct
}

// IsVideoAvid .
func (mde *NativeModule) IsVideoAvid() bool {
	return mde.Category == _moduleVideoAvid
}

// IsResourceID .
func (mde *NativeModule) IsResourceID() bool {
	return mde.Category == ModuleResourceID
}

// IsNewVideoID .
func (mde *NativeModule) IsNewVideoID() bool {
	return mde.Category == ModuleNewVideoAvid
}

// IsNewVideoAct .
func (mde *NativeModule) IsNewVideoAct() bool {
	return mde.Category == ModuleNewVideoAct
}

// IsNewVideoDyn .
func (mde *NativeModule) IsNewVideoDyn() bool {
	return mde.Category == ModuleNewVideoDyn
}

// IsResourceAct .
func (mde *NativeModule) IsResourceAct() bool {
	return mde.Category == ModuleResourceAct
}

// IsResourceOrigin .
func (mde *NativeModule) IsResourceOrigin() bool {
	return mde.Category == ModuleResourceOrigin
}

// IsResourceDyn .
func (mde *NativeModule) IsResourceDyn() bool {
	return mde.Category == ModuleResourceDynamic
}

// IsInlineTab .
func (mde *NativeModule) IsInlineTab() bool {
	return mde.Category == ModuleInlineTab
}

// IsSelect
func (mde *NativeModule) IsSelect() bool {
	return mde.Category == ModuleSelect
}

func (mde *NativeModule) IsLive() bool {
	return mde.Category == ModuleLive
}

func (mde *NativeModule) IsCarouselImg() bool {
	return mde.Category == ModuleCarouselImg
}

func (mde *NativeModule) IsCarouselWord() bool {
	return mde.Category == ModuleCarouselWord
}

func (mde *NativeModule) IsOgvSeasonID() bool {
	return mde.Category == ModuleOgvSeasonID
}

func (mde *NativeModule) IsOgvSeasonSource() bool {
	return mde.Category == ModuleOgvSeasonSource
}

func (mde *NativeModule) IsReply() bool {
	return mde.Category == ModuleReply
}

func (mde *NativeModule) IsEditorOrigin() bool {
	return mde.Category == ModuleEditorOrigin
}

func (mde *NativeModule) IsActCapsule() bool {
	return mde.Category == ModuleActCapsule
}

func (mde *NativeModule) IsCarouselSource() bool {
	return mde.Category == ModuleCarouselSource
}

func (mde *NativeModule) IsRcmdSource() bool {
	return mde.Category == ModuleRcmdSource
}

func (mde *NativeModule) IsRcmdVerticalSource() bool {
	return mde.Category == ModuleRcmdVerticalSource
}

func (mde *NativeModule) IsGame() bool {
	return mde.Category == ModuleGame
}

func (mde *NativeModule) IsVote() bool {
	return mde.Category == VoteModule
}

func (mde *NativeModule) IsReserve() bool {
	return mde.Category == ModuleReserve
}

func (mde *NativeModule) IsIcon() bool {
	return mde.Category == ModuleIcon
}

// IsBanner .
func (mde *NativeModule) IsBanner() bool {
	return mde.Category == _moduleBanner
}

// IsStatement .
func (mde *NativeModule) IsStatement() bool {
	return mde.Category == ModuleStatement
}

// IsSingleDyn .
func (mde *NativeModule) IsSingleDyn() bool {
	return mde.Category == _moduleSingleDyn
}

// IsEditor .
func (mde *NativeModule) IsEditor() bool {
	return mde.Category == ModuleEditor
}

// IsResourceRole .
func (mde *NativeModule) IsResourceRole() bool {
	return mde.Category == ModuleResourceRole
}

// IsTimelineIDs .
func (mde *NativeModule) IsTimelineIDs() bool {
	return mde.Category == ModuleTimelineIDs
}

// IsTimelineSource .
func (mde *NativeModule) IsTimelineSource() bool {
	return mde.Category == ModuleTimelineSource
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
	return mde.Category == ModuleParticipation
}

// IsRecommend .
func (mde *NativeModule) IsRecommend() bool {
	return mde.Category == ModuleRecommend
}

// IsRcmdVertical .
func (mde *NativeModule) IsRcmdVertical() bool {
	return mde.Category == ModuleRcmdVertical
}

// IsProgress .
func (mde *NativeModule) IsProgress() bool {
	return mde.Category == ModuleProgress
}

// IsNavigation
func (mde *NativeModule) IsNavigation() bool {
	return mde.Category == ModuleNavigation
}

// IsBaseHead .
func (mde *NativeModule) IsBaseHead() bool {
	return mde.Category == ModuleBaseHead
}

func (mde *NativeModule) IsBaseHoverButton() bool {
	return mde.Category == ModuleBaseHoverButton
}

func (mde *NativeModule) IsNewactHeaderModule() bool {
	return mde.Category == ModuleNewactHeader
}

func (mde *NativeModule) IsNewactAwardModule() bool {
	return mde.Category == ModuleNewactAward
}

func (mde *NativeModule) IsNewactStatementModule() bool {
	return mde.Category == ModuleNewactStatement
}

func (mde *NativeModule) IsBaseBottomButton() bool {
	return mde.Category == BottomButtonModule
}

func (mde *NativeModule) IsMatchMedal() bool {
	return mde.Category == MatchMedalModule
}

func (mde *NativeModule) IsMatchEvent() bool {
	return mde.Category == MatchEventModule
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

func (mde *NativeModule) IsAttrShareImage() int64 {
	return (mde.Attribute >> AttrIsShareImage) & int64(1)
}

func (mde *NativeModule) IsAttrDisplayH5Header() int64 {
	return (mde.Attribute >> AttrIsDisplayH5Header) & int64(1)
}

func (mde *NativeModule) IsAttrIsDisplayUpIcon() int64 {
	return (mde.Attribute >> AttrIsDisplayUpIcon) & int64(1)
}

func (mde *NativeModule) IsAttrIsCloseSubscribeBtn() int64 {
	return (mde.Attribute >> AttrIsCloseSubscribeBtn) & int64(1)
}

func (mde *NativeModule) IsAttrIsCloseViewNum() int64 {
	return (mde.Attribute >> AttrIsCloseViewNum) & int64(1)
}

// IsAttrDisplayArticleIcon .
func (mde *NativeModule) IsAttrDisplayArticleIcon() int64 {
	return (mde.Attribute >> AttrIsDisplayArticleIcon) & int64(1)
}

// IsAttrStatementDisplayButton .
func (mde *NativeModule) IsAttrStatementDisplayButton() int64 {
	return (mde.Attribute >> AttrStatementIsDisplayButton) ^ int64(1)
}

// IsAttrDisplayNum 是否展示当前进度数值&ogv是否展示评分&投票组件-是否展示选项得票.
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
	if mde.Colors != "" {
		ry := &Colors{}
		if err := json.Unmarshal([]byte(mde.Colors), ry); err == nil {
			return ry
		}
	}
	return &Colors{}
}

func (mde *NativeModule) ConfUnmarshal() *ConfSort {
	if mde.ConfSort != "" {
		ry := &ConfSort{}
		if err := json.Unmarshal([]byte(mde.ConfSort), ry); err == nil {
			return ry
		}
	}
	return &ConfSort{}
}

func (mde *NativeMixtureExt) RemarkUnmarshal() *MixReason {
	if mde.Reason != "" {
		ry := &MixReason{}
		if err := json.Unmarshal([]byte(mde.Reason), ry); err == nil {
			return ry
		}
	}
	return &MixReason{}
}

func (mde *NativeClick) ExtUnmarshal() *ClickExt {
	if mde.Ext != "" {
		ry := &ClickExt{}
		if err := json.Unmarshal([]byte(mde.Ext), ry); err == nil {
			return ry
		}
	}
	return &ClickExt{}
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
	return mde.MType == PartDynamic
}

// IsPartVideo .
func (mde *NativeParticipationExt) IsPartVideo() bool {
	return mde.MType == PartVideo
}

// IsPartArticle .
func (mde *NativeParticipationExt) IsPartArticle() bool {
	return mde.MType == PartArticle
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

// 自定义按钮类型（0-10，50-59）
func (mde *NativeClick) IsCustom() bool {
	return (mde.Type > 0 && mde.Type < 10) || (mde.Type >= 50 && mde.Type <= 59)
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

// 是否是up主预约
func (mde *NativeClick) IsUpAppointment() bool {
	return mde.Type == _clickUpAppointment
}

// 是否活动预约
func (mde *NativeClick) IsActReserve() bool {
	return mde.Type == _actReserve
}

func (mde *NativeClick) IsOnlyImage() bool {
	return mde.Type == _onlyImage
}

func (mde *NativeClick) IsBuyCoupon() bool {
	return mde.Type == _buyCoupon
}

func (mde *NativeClick) IsCartoon() bool {
	return mde.Type == _cartoon
}

// 是否领取
func (mde *NativeClick) IsRedirect() bool {
	return mde.Type == Redirect
}

// 是否预约
func (mde *NativeClick) IsPendant() bool {
	return mde.Type == _pendant
}

// 点击区域-接口模式
func (mde *NativeClick) IsInterface() bool {
	return mde.Type == _interface
}

// 是否进度条
func (mde *NativeClick) IsProgress() bool {
	return mde.Type == _progress
}

// 是否是静态进度条
func (mde *NativeClick) IsStaticProgress() bool {
	return mde.Type == _staticProcess
}

func (mde *NativeClick) IsVoteButton() bool {
	return mde.Type == VoteButton
}

func (mde *NativeClick) IsVoteProcess() bool {
	return mde.Type == VoteProcess
}

func (mde *NativeClick) IsVoteUser() bool {
	return mde.Type == VoteUser
}

func (mde *NativeClick) IsPublishBtn() bool {
	return mde.Type == ClickPublishBtn
}

// 是否浮层-图片模式
func (mde *NativeClick) IsLayerImage() bool {
	return mde.Type == _layerImage
}

// 是否浮层-链接模式
func (mde *NativeClick) IsLayerLink() bool {
	return mde.Type == _layerLink
}

// 是否浮层-接口模式
func (mde *NativeClick) IsLayerInterface() bool {
	return mde.Type == _layerInterface
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

func IsFromTopicUpg(from int32) bool {
	return from == PageFromUpgSourceHot || from == PageFromUpgSourceDiscuss || from == PageFromUpgSourceAI || from == PageFromUpgSourceVideo
}

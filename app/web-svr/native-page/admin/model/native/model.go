package native

const (
	TopicType                = 1 //话题活动页page
	InLineType               = 2 //内部页面page
	MenuType                 = 3 //首页menupage
	OgvType                  = 4 //ogv page
	PlayerType               = 6 //播放器内嵌活动页
	SpaceType                = 7 //space page
	UgcType                  = 8 //ugc播放页
	ClickModule              = 1
	DynmaicModule            = 2
	VideoModule              = 3
	ActModule                = 4
	VideoAvidModule          = 5 //视频卡-id模式
	VideoActModule           = 6 //视频卡-act模式
	VideoDynModule           = 7 //视频卡-动态模式
	BannerModule             = 8
	StatementsModule         = 9
	SingleDynamic            = 10
	ParticipationModule      = 11
	RecommentModule          = 12
	NavigationModule         = 13
	HeadModule               = 14
	ResourceIDModule         = 15 //资源小卡-id模式
	ResourceActModule        = 16 //资源小卡-act模式
	ResourceDynamicModule    = 17 //资源小卡-动态模式
	InlineTabModule          = 18 //页面tab组件
	LiveModule               = 19 //直播卡
	CarouselImgModule        = 20 //轮播-图片模式
	IconModule               = 21 //图标
	NewVideoAvidModule       = 22 //新视频卡-id模式
	NewVideoActModule        = 23 //新视频卡-act模式
	NewVideoDynModule        = 24 //新视频卡-动态模式
	EditorModule             = 25 //编辑推荐卡
	RcmdVerticalModule       = 26 //推荐用户-竖卡
	SelectModule             = 27 //筛选组件
	ProgressModule           = 28 //进度条
	ResourceRoleModule       = 29 //资源小卡-角色剧集模式
	CarouselWordModule       = 30 //轮播-文字模式
	TimelineIDModule         = 31 //时间轴-id模式
	TimelineSourceModule     = 32 //时间轴-数据源模式
	OgvSeasonIDModule        = 33 //ogv 剧集卡-ids
	OgvSeasonSourceModule    = 34 //ogv 剧集卡-片单id
	ResourceDataOriginModule = 35 //资源小卡-外接数据源
	ReplyModule              = 36 //评论组件
	EditorResource           = 37 // 编辑推荐卡-外接数据源类型
	ActCapsuleModule         = 38 //相关活动-胶囊模式
	CarouselSourceModule     = 39 //轮播-数据源模式
	RcmdSourceModule         = 40 //推荐用户-横卡-数据源&排行榜
	RcmdVerticalSourceModule = 41 //推荐用户-竖卡-数据源模式
	HoverButtonModule        = 43 //自定义悬浮按钮
	MixtureUpType            = 0
	MixtureArcType           = 1  //ugc-avid类型
	MixtureEpidType          = 2  //pgc-epid
	MixtureCvidType          = 3  //专栏-cvid
	MixturePageType          = 4  // inline tab pageid
	MixtureCarouselImg       = 5  //轮播-图片
	MixtureCarouselWord      = 6  //轮播-文字
	MixtureIconImg           = 7  //图标-图片
	MixtureFolder            = 8  //播单
	MixtureRcmdVertical      = 9  //推荐用户-竖卡
	MixtureProgress          = 10 //进度条
	MixTimelinePic           = 11 //时间轴-图片模式
	MixTimelineText          = 12 //时间轴-文字模式
	MixTimeline              = 13 //时间轴-图文模式
	MixtureAct               = 18 //相关活动-活动
	TabStateInvalid          = 0
	TabStateValid            = 1
	TabModuleStateInValid    = 0
	TabModuleStateValid      = 1
	IconTypeCustom           = 1 //图标样式：自定义图标+文字
	IconTypeWord             = 2 //图标样式：纯文字
	CategoryPage             = 1 //跳转页面：活动页
	CategoryLink             = 2 //跳转页面：链接
	//page 发起类型
	// 运营配置活动
	PageFromSystem = 0
	// up主发起活动
	PageFromUid = 1
	// page state
	WaitForCommit = -3 //草稿箱
	CheckOffline  = -2 //打回
	WaitForCheck  = -1 //待审核
	WaitForOnline = 0  //待上线
	OnlineState   = 1  //page 上线
	OfflineState  = 2  //page 下线
	//ts state
	TsOffline    = 2
	TsOnline     = 1
	TsWaitOnline = 0
	// act_type
	ActTypeBiz = 9 //商业推广活动
	// ts audit_type
	TsAuditAuto   = "auto"   //自动审核
	TsAuditManual = "manual" //人工审核
	//首页tab背景类型
	BgTypeImage = 1
	BgTypeColor = 2
	// tab_type
	TabTypeUpAct  = "up_act"
	TabTypeNapage = "napage"
	// topic_upgrade source
	UpgSourceHot     = "hot"     //热度标准
	UpgSourceDiscuss = "discuss" //讨论蹿升
	UpgSourceAI      = "ai"      //AI召回
	UpgSourceVideo   = "video"   //视频数据源生成
	// action
	ActionOn  = "on"
	ActionOff = "off"
	//color length
	_colorLen = 2
	//table string length
	_colLen = 255
	MaxLen  = 2000
	// ukey prefix
	UkeyPrefixUpgrade = "UPG"
	UkeyPrefixUP      = "UP"
	UkeyPrefixNewact  = "ACT"
)

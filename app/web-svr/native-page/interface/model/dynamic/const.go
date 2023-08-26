package dynamic

const (
	VideoType                 = 8
	ARTICLETYPE               = 64
	GotoVote                  = "vote"
	GotoVoteArea              = "vote_area"
	GotoVoteBack              = "vote_back"
	GotoClick                 = "click"
	GotoClickArea             = "click_area"
	GotoClickButton           = "click_button"
	GotoClickButtonV2         = "click_button_v2"
	GotoClickButtonV3         = "click_button_v3"
	GotoClickProgress         = "click_progress"
	GotoClickStaticProgress   = "click_static_progress"
	GotoClickBack             = "click_back"
	GotoClickReserve          = "click_reserve"
	GotoClickURL              = "click_url"
	GotoClickUnable           = "click_unable"
	GotoClickAppointment      = "appointment"
	GotoClickAttention        = "act_relationInfo"
	GotoClickRedirect         = "redirect"
	GotoAct                   = "act_list"
	GotoVideoModule           = "video_module"
	GotoVideo                 = "video"
	GotoVideoMore             = "video_more"
	GotoDynamicModule         = "dynamic_module"
	GotoDynamic               = "dynamic"
	GotoDynamicMore           = "dynamic_more"
	GotoResource              = "resource"
	GotoNewUgcVideo           = "new_ugc_video"
	GotoNewPgcVideo           = "new_pgc_video"
	GotoOgvSeasoThree         = "ogv_season_three"
	GotoOgvSeasoOne           = "ogv_season_one"
	GotoInlineTabModule       = "inline_tab_module"
	GotoInlineTab             = "inline_tab"
	GotoReplyModule           = "reply"
	GotoSelect                = "select"
	GotoSelectModule          = "select_module"
	GotoLiveModule            = "live_module"
	GotoStatementModule       = "statement_module"
	GotoNewVideoModule        = "new_video_module" //新视频卡-act模式
	GotoResourceModule        = "resource_module"
	GotoEditorModule          = "editor_module"
	GotoNewEditorModule       = "new_editor_module"
	GotoIcon                  = "icon"
	GotoIconModule            = "icon_module"
	GotoRecommend             = "recommend"
	GotoCarouselImg           = "carousel_img"
	GotoCarouselImgModule     = "carousel_img_module"
	GotoEditor                = "editor"
	GotoNewEditor             = "new_editor"
	GotoTitleImage            = "display_image"
	GotoCarouselWord          = "carousel_word"
	GotoCarouselWordModule    = "carousel_word_module"
	GotoTimelineModule        = "timeline_module"
	GotoTimelineMore          = "timeline_more"
	GotoTimelineExpand        = "timeline_expand"
	GotoTimelineResource      = "timeline_event_resource"
	GotoTimelinePic           = "timeline_event_pic"
	GotoTimelineText          = "timeline_event_text"
	GotoTimelineMix           = "timeline_event_pic_text"
	GotoTitleName             = "display_name"
	GotoActModule             = "act_module"
	GotoActCapsuleModule      = "act_capsule_module"
	GotoActCapsule            = "act_capsule"
	GotoRecommendModule       = "recommend_module"
	GotoRcmdVerticalMou       = "recommend_vertical_module"
	GotoRcmdVertical          = "recommend_vertical"
	GotoNavigationModule      = "navigation_module"
	GotoNavigation            = "navigation"
	GotoOgvSeasonModule       = "ogv_season_module"
	GotoOgvSeasonMore         = "ogv_season_more"
	GotoProgressModule        = "progress_module"
	GotoProgress              = "progress"
	GotoGameModule            = "game_module"
	GotoGame                  = "game"
	GotoReserve               = "reserve"
	GotoReserveModule         = "reserve_module"
	GotoHoverButtonModule     = "hover_button_module"
	GotoNewactHeaderModule    = "newact_header_module"
	GotoNewactHeader          = "newact_header"
	GotoNewactAwardModule     = "newact_award_module"
	GotoNewactAward           = "newact_award"
	GotoNewactStatementModule = "newact_statement_module"
	GotoNewactStatement       = "newact_statement"
	GotoBottomButton          = "bottom_button"
	GotoMatchMedalModule      = "match_medal_module"
	GotoMatchMedal            = "match_medal"
	GotoMatchEventModule      = "match_event_module"
	GotoMatchEvent            = "match_event"
	//up主发起活动来源
	UpFromSquare = "square"
	UpArcsMax    = 21        //up主发起多选择的稿件个数
	UpSyncActMax = 3         //up主同步至空间的活动上限
	TsResFromAct = "act_arc" //活动稿件
	TsResFromMy  = "my_arc"  //我的稿件
	// icon地址
	IconPlay     = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/lM~lH4iu.png"
	IconDanmaku  = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/-udU-i01.png"
	IconLive     = "https://i0.hdslb.com/bfs/activity-plat/static/20210112/467746a96c68611c46194c29089d62f5/Th6JlIxPr.png"
	IconFavorite = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/vEWW131c.png"
	IconReserve  = "http://i0.hdslb.com/bfs/archive/f5b7dae25cce338e339a655ac0e4a7d20d66145c.png"
	// style
	StyleColor = "color"
	StyleImage = "image"
	// progress
	ProgressFromClick = "click"    //自定义组件
	ProgressFromProg  = "progress" //进度条组件
	// 同步至空间按钮
	SpaceBtPersonal  = "personal_page"  //同步到个人空间
	SpaceBtExclusive = "exclusive_page" //空间专属页设置
	// 空间配置保存的来源
	SpaceSaveFromUser   = "user_save" //用户保存操作
	SpaceSaveFromTsAdd  = "ts_add"    //up主发起-新建
	SpaceSaveFromTsSave = "ts_save"   //up主发起-编辑
	// 进度条样式
	ProgStyleRound     = 1 //圆角条
	ProgStyleRectangle = 2 //矩形条
	ProgStyleNode      = 3 //分节条
	// item.DisplayType
	DisTypeRound     = "round"
	DisTypeRectangle = "rectangle"
	DisTypeNode      = "node"
	//预约组件展示类型
	ReserveDisplayA    = int64(1) //直播和稿件类型
	ReserveDisplayC    = int64(2) //直播和稿件类型
	ReserveDisplayD    = int64(3) //直播类型，有回放
	ReserveDisplayE    = int64(4) //直播类型，无回放
	ReserveDisplayLive = int64(5)
	//直播态
	Living  = 1
	LiveEnd = 2
	//直播enterFrom
	LiveEnterFrom = "dynamic_activity_reserve"
	//limit
	MaxIDsLen     = 50
	MaxDynamicLen = 1000
	//内嵌页类型
	FormatModFromMenuSpace = "menu_space"
	FormatModFromMenuUp    = "menu_up"
	// up主发起模板
	UpTempDiscuss = "discuss" //讨论类
	UpTempCollect = "collect" //征稿类
	// 投票组件-选项数
	VoteOptionNum = 2
	// 表格对齐方式
	TextAlignCenter = "center" //居中
	TextAlignLeft   = "left"   //靠左
	TextAlignRight  = "right"  //靠右
)

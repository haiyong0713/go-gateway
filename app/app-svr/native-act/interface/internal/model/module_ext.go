package model

import (
	"encoding/json"

	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/api"
)

const (
	// 编辑推荐卡-位置属性
	PositionUp            = "up"
	PositionView          = "view"
	PositionPubTime       = "pub_time"
	PositionLike          = "like"
	PositionDanmaku       = "danmaku"
	PositionDuration      = "duration"
	PositionFollow        = "follow"
	PositionViewStat      = "view_state"    //用户观看状态
	PositionComprehensive = "comprehensive" //排行榜-综合
	PositionShare         = "share"
	PositionCoin          = "coin" //投币
	// 分享相关
	ShareOriginUGC             = "ugc"
	ShareOriginArticle         = "article"
	ShareOriginTab             = "activity_tab_share"
	ShareOriginInlineTab       = "activity_more_share"
	ShareOriginSimpleInlineTab = "activity_one_page_share"
	ShareOriginLongPress       = "activity_longpress_share"
	ShareTypeActivity          = 8
	// 上报相关
	ReportBizTypePGC          = "1"
	ReportBizTypeArticle      = "2"
	ReportBizTypeUGC          = "3"
	ReportBizTypeBizCommodity = "4"
	ReportBizTypeSeason       = "5"
	ReportBizTypeWeb          = "6"
	ReportBizTypeOgvFilm      = "7"
	ReportBizTypeLive         = "8"
	// 投稿按钮-投稿类型
	ParticipationDynamic     = "dynamic"
	ParticipationVideoChoose = "video-choose"
	ParticipationVideoShoot  = "video-shoot"
	ParticipationArticle     = "article"
	// 动态列表-ClassType
	DynClassChoice = 1 //精选
	// 二级列表透传字段
	SubParamsField = "grpc_params"
	// RDBType
	RDBBizCommodity = 1 //资源小卡-企业号-商品数据源
	RDBBizIds       = 2 //资源小卡-企业号-稿件类数据源
	RDBOgv          = 3 //资源小卡-OGV运营后台
	RDBMustsee      = 4 //编辑推荐卡-入站必刷
	RDBWeek         = 5 //编辑推荐卡-每周必看
	RDBLive         = 6 //资源小卡-直播活动
	RDBRank         = 7 //编辑推荐卡-排行榜
	RDBGAT          = 8 //编辑推荐卡-港澳台垂类数据
	// 资源类型
	ResourceTypeUGC       = "ugc"
	ResourceTypeOGV       = "ogv"
	ResourceTypeArticle   = "article"
	ResourceTypeLive      = "live"
	ResourceTypeSeason    = "season"
	ResourceTypeWeb       = "web"       //网页
	ResourceTypeOgvFilm   = "ogv_film"  //片单
	ResourceTypeCommodity = "commodity" //商品
	// 顶栏-背景配置模式
	TopTabBgImg   = 1
	TopTabBgColor = 2
	// 轮播-内容样式
	CarouselCSBanner     = 1 //banner模式
	CarouselCSSlide      = 2 //横滑模式
	CarouselCSSingleLine = 3 //单行
	CarouselCSMultiLine  = 4 //多行
	// 排序 ConfSort.SortType
	SortTypeCtime  = "ctime"  //创建时间
	SortTypeRandom = "random" //随机
	// 数据源类型
	SourceTypeActUp   = "act_up"   //活动的up主
	SourceTypeRank    = "rank"     //排行榜
	SourceTypeVoteAct = "act_vote" //活动投票
	SourceTypeVoteUp  = "up_vote"  //UP主投票
	// 投票组件-选项数
	VoteOptionNum = 2
	// 投票组件-进度条样式
	VPStyleCircle = "circle" //圆角
	VPStyleSquare = "square" //方角
	// UP主预约组件-展示类型
	ReserveDisplayA    = int64(1)
	ReserveDisplayC    = int64(2)
	ReserveDisplayD    = int64(3)
	ReserveDisplayE    = int64(4)
	ReserveDisplayLive = int64(5)
	// 时间轴组件-查看更多方式
	TimelineMoreByJump        = 0 //跳转至二级页
	TimelineMoreBySupernatant = 1 //浮层
	TimelineMoreByExpand      = 2 //下拉展示
	// 时间轴组件-节点类型
	TimelineNodeText = 0 //文本
	TimelineNodeTime = 1 //时间节点
	// 时间轴组件-时间精度
	TimelineTimeYear  = 0
	TimelineTimeMonth = 1
	TimelineTimeDay   = 2
	TimelineTimeHour  = 3
	TimelineTimeMin   = 4
	TimelineTimeSec   = 5
	// Tab组件-是否需解锁
	TabLockPass = 0 //无需解锁
	TabLockNeed = 1 //需解锁
	// Tab组件-解锁类型
	TabLockTypePass   = 0 //无需解锁
	TabLockTypeTime   = 1 //时间
	TabLockTypeSource = 2 //预约数据源
	// Tab组件-未解锁时
	TabNotUnlockDisableDisplay = 1
	TabNotUnlockDisableClick   = 2
	// Tab组件-生效类型
	EftTypeImmediately = 1 //立即生效
	EftTypeTiming      = 2 //定时生效
	// 文本类型
	StatementNewactTask        = 1 //新活动页-任务玩法
	StatementNewactRule        = 2 //新活动页-规则说明
	StatementNewactDeclaration = 3 //新活动页-平台声明
	// 进度条组件-样式配置
	PgStyleRound     = 1 //圆角条
	PgStyleRectangle = 2 //矩形条
	PgStyleNode      = 3 //分节条
	// 进度条组件-未达成态（进度槽）配置
	PgSlotOutline = "1" //描边
	PgSlotFill    = "2" //填充
	// 进度条组件-达成态（进度条）配置
	PgBarColor   = "1" //纯色填充
	PgBarTexture = "2" //纹理颜色填充
	// 进度条组件-纹理类型
	PgTexture1 = 1
	PgTexture2 = 2
	PgTexture3 = 3
	// 活动数据源-排序
	ActOrderTime   = 2 //时间
	ActOrderRandom = 3 //随机
	ActOrderHot    = 4 //热度
	// 自定义点击-区域类型
	ClickTypeRedirect       = 0  //点击区域-跳转链接
	ClickTypeFollowUser     = 1  //自定义按钮-关注
	ClickTypeFollowEpisode  = 2  //自定义按钮-追番/追剧
	ClickTypeReserve        = 3  //自定义按钮-预约数据源
	ClickTypeReceiveAward   = 4  //自定义按钮-奖励领取
	ClickTypeBtnRedirect    = 5  //自定义按钮-跳转链接
	ClickTypeActivity       = 6  //自定义按钮-活动项目（预约+其他操作）
	ClickTypeDisplayImage   = 7  //自定义按钮-仅展示图片
	ClickTypeMallWantGo     = 8  //自定义按钮-会员购票务「想去」
	ClickTypeFollowComic    = 9  //自定义按钮-追漫
	ClickTypeLayerImage     = 10 //浮层-图片模式
	ClickTypeLayerRedirect  = 11 //浮层-链接模式
	ClickTypeLayerInterface = 12 //浮层-接口模式（待有需要再重构）
	ClickTypeUrlPerDev      = 20 //点击区域-双端不同链接
	ClickTypeUrlInterface   = 21 //点击区域-接口模式（待有需要再重构）
	ClickTypeRTProgress     = 30 //进度数值-实时
	ClickTypeNRTProgress    = 31 //进度数值-非实时
	ClickTypeVoteButton     = 40 //投票组件-投票按钮
	ClickTypeVoteProcess    = 41 //投票组件-投票进度
	ClickTypeVoteUser       = 42 //投票组件-用户剩余票数
	ClickTypeUpReserve      = 50 //自定义按钮-UP主预约
	ClickTypeParticipation  = 60 //投稿按钮
	// 自定义点击-进度数值-数据来源
	ClickNRTProgUserStats   = 0 //用户积分统计
	ClickNRTProgActApplyNum = 1 //活动报名量
	ClickNRTProgTaskStats   = 2 //任务统计
	ClickNRTProgLotteryNum  = 3 //抽奖数量
	ClickNRTProgScore       = 4 //动态评分
	// 自定义悬浮按钮-按钮类型
	HoverBtnReserve  = "appoint_origin"
	HoverBtnActivity = "act_project"
	HoverBtnRedirect = "link"
)

type MixExtImage struct {
	Image  string `json:"image"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
	Size   int64  `json:"size"`
}

func (mi *MixExtImage) ToSizeImage() *api.SizeImage {
	if mi == nil {
		return nil
	}
	return &api.SizeImage{Image: mi.Image, Height: mi.Height, Width: mi.Width, Size_: mi.Size}
}

type MixExtEditor struct {
	Fid         int64       `json:"fid"`
	RcmdContent RcmdContent `json:"rcmd_content"` //编辑推荐内容
}

type RcmdContent struct {
	TopContent      string `json:"top_content"`       //顶部推荐语
	TopFontColor    string `json:"top_font_color"`    //顶部字体颜色
	BottomContent   string `json:"bottom_content"`    //底部推荐语
	BottomFontColor string `json:"bottom_font_color"` //底部字体颜色
	MiddleIcon      string `json:"middle_icon"`       //排行榜icon
}

func UnmarshalMixExtEditor(data string) (*MixExtEditor, error) {
	if data == "" {
		return &MixExtEditor{}, nil
	}
	mixEditor := &MixExtEditor{}
	if err := json.Unmarshal([]byte(data), mixEditor); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of Editor, data=%+v error=%+v", data, err)
		return &MixExtEditor{}, err
	}
	return mixEditor, nil
}

type MixExtCarouselImg struct {
	ImgUrl         string `json:"img_url"`
	RedirectUrl    string `json:"redirect_url"`
	Length         int64  `json:"length"`
	Width          int64  `json:"width"`
	BgType         int64  `json:"bg_type"`          //背景配置模式
	BgImage1       string `json:"bg_image_1"`       //背景图1
	BgImage2       string `json:"bg_image_2"`       //背景图2
	TabTopColor    string `json:"tab_top_color"`    //顶栏头部颜色
	TabMiddleColor string `json:"tab_middle_color"` //中间色值
	TabBottomColor string `json:"tab_bottom_color"` //tab栏底部色值
	FontColor      string `json:"font_color"`       //tab文本高亮色值
	BarType        int64  `json:"bar_type"`         //系统状态栏色值：0 黑色；1 白色
}

func UnmarshalMixExtCarouselImg(data string) (*MixExtCarouselImg, error) {
	if data == "" {
		return &MixExtCarouselImg{}, nil
	}
	img := &MixExtCarouselImg{}
	if err := json.Unmarshal([]byte(data), img); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of CarouselImg, data=%+v error=%+v", data, err)
		return &MixExtCarouselImg{}, err
	}
	return img, nil
}

type MixExtCarouselWord struct {
	Content string `json:"content"`
}

func UnmarshalMixExtCarouselWord(data string) (*MixExtCarouselWord, error) {
	if data == "" {
		return &MixExtCarouselWord{}, nil
	}
	word := &MixExtCarouselWord{}
	if err := json.Unmarshal([]byte(data), word); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of CarouselWord, data=%+v error=%+v", data, err)
		return &MixExtCarouselWord{}, err
	}
	return word, nil
}

type MixExtResourceFolder struct {
	Fid int64 `json:"fid"`
}

func UnmarshalMixExtResourceFolder(data string) (*MixExtResourceFolder, error) {
	if data == "" {
		return &MixExtResourceFolder{}, nil
	}
	folder := &MixExtResourceFolder{}
	if err := json.Unmarshal([]byte(data), folder); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of ResourceFolder, data=%+v error=%+v", data, err)
		return &MixExtResourceFolder{}, err
	}
	return folder, nil
}

type RcmdVerticalExt struct {
	Reason string `json:"reason"` //推荐理由
	Uri    string `json:"uri"`    //指定链接
}

func UnmarshalRcmdVerticalExt(data string) (*RcmdVerticalExt, error) {
	if data == "" {
		return &RcmdVerticalExt{}, nil
	}
	ext := &RcmdVerticalExt{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of RcmdVerticalExt, data=%s error=%+v", data, err)
		return &RcmdVerticalExt{}, err
	}
	return ext, nil
}

type IconExt struct {
	ImgUrl      string `json:"img_url"`
	RedirectUrl string `json:"redirect_url"`
	Content     string `json:"content"`
}

func UnmarshalIconExt(data string) (*IconExt, error) {
	if data == "" {
		return &IconExt{}, nil
	}
	ext := &IconExt{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of IconExt, data=%s error=%+v", data, err)
		return &IconExt{}, err
	}
	return ext, nil
}

type ParticipationExt struct {
	NewTid int64 `json:"new_tid"`
}

func UnmarshalParticipationExt(data string) (*ParticipationExt, error) {
	if data == "" {
		return &ParticipationExt{}, nil
	}
	ext := &ParticipationExt{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeParticipationExt.Ext, data=%+v error=%+v", data, err)
		return &ParticipationExt{}, err
	}
	return ext, nil
}

type TabExt struct {
	DefType     int64        `json:"def_type"`     //生效方式
	DStime      int64        `json:"d_stime"`      //生效开始时间
	DEtime      int64        `json:"d_etime"`      //生效结束时间
	UnI         *MixExtImage `json:"un_i"`         //未解锁态图片
	SI          *MixExtImage `json:"si"`           //选中态图片
	UnSI        *MixExtImage `json:"un_si"`        //未选中态图片
	Type        string       `json:"type"`         //业务方：week 每周必看
	LocationKey string       `json:"location_key"` //业务方唯一id
}

func UnmarshalTabExt(data string) (*TabExt, error) {
	if data == "" {
		return &TabExt{}, nil
	}
	ext := &TabExt{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of Tab, data=%+v error=%+v", data, err)
		return &TabExt{}, err
	}
	return ext, nil
}

type ClickExtProgress struct {
	GroupId     int64  `json:"group_id"`     //节点组id
	NodeId      int64  `json:"node_id"`      //节点id
	DisplayType string `json:"display_type"` //展示数值类型
	FontSize    int64  `json:"font_size"`    //字号
	FontType    string `json:"font_type"`    //字体
	FontColor   string `json:"font_color"`   //字体颜色
	PSort       int64  `json:"pSort"`        //非实时进度条-数据来源
	Activity    string `json:"activity"`     //非实时进度条-任务统计-活动名
	Counter     string `json:"counter"`      //非实时进度条-任务统计-counter名
	StatPc      string `json:"statPc"`       //非实时进度条-任务统计-统计周期
	LotteryID   string `json:"lotteryID"`    //非实时进度条-抽奖数量-抽奖ID
}

func UnmarshalClickExtProgress(data string) (*ClickExtProgress, error) {
	if data == "" {
		return &ClickExtProgress{}, nil
	}
	ext := &ClickExtProgress{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeClick.Tip of progress, data=%+v error=%+v", data, err)
		return nil, err
	}
	return ext, nil
}

type ClickExtLayer struct {
	ButtonImage string         `json:"button_image"` //浮层按钮
	Style       string         `json:"style"`        //浮层样式
	LayerImage  string         `json:"layer_image"`  //浮层图片标题
	ShareImage  *MixExtImage   `json:"share_image"`  //分享图片
	Images      []*MixExtImage `json:"images"`       //图片模式-图片
}

func UnmarshalClickExtLayer(data string) (*ClickExtLayer, error) {
	if data == "" {
		return &ClickExtLayer{}, nil
	}
	ext := &ClickExtLayer{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeClick.Ext of layer, data=%+v error=%+v", data, err)
		return nil, err
	}
	return ext, nil
}

type ClickExtLayerColor struct {
	Title      string `json:"title"`       //标题
	TopColor   string `json:"top_color"`   //顶栏颜色
	TitleColor string `json:"title_color"` //标题颜色
}

func UnmarshalClickExtLayerColor(data string) (*ClickExtLayerColor, error) {
	if data == "" {
		return &ClickExtLayerColor{}, nil
	}
	ext := &ClickExtLayerColor{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeClick.Tip of layer_color, data=%+v error=%+v", data, err)
		return nil, err
	}
	return ext, nil
}

type ClickExtCommon struct {
	SynHover        bool   `json:"syn_hover"`        //是否与悬浮按钮互通
	Ukey            string `json:"ukey"`             //点击区域ukey
	DisplayMode     int64  `json:"display_mode"`     //展示模式：0 无要求 1 解锁后展示
	UnlockCondition int64  `json:"unlock_condition"` //解锁条件：0 无要求；1 时间；2 预约/积分进度
	Stime           int64  `json:"stime"`            //时间解锁-开始时间
	Sid             int64  `json:"sid"`              //数据源id
	GroupId         int64  `json:"group_id"`         //节点组id
	NodeId          int64  `json:"node_id"`          //节点id
}

func UnmarshalClickExtCommon(data string) (*ClickExtCommon, error) {
	if data == "" {
		return &ClickExtCommon{}, nil
	}
	ext := &ClickExtCommon{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeClick.Ext of common, data=%+v error=%+v", data, err)
		return nil, err
	}
	return ext, nil
}

type HoverButtonExt struct {
	BtType string   `json:"bt_type"` //按钮类型
	Hint   string   `json:"hint"`    //成功提示
	MUkeys []string `json:"m_ukeys"` //当该组件划出屏幕后，悬浮按钮才会出现
}

func UnmarshalHoverButtonExt(data string) (*HoverButtonExt, error) {
	if data == "" {
		return &HoverButtonExt{}, nil
	}
	ext := &HoverButtonExt{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeModule.ConfSort of hover_button, data=%+v error=%+v", data, err)
		return nil, err
	}
	return ext, nil
}

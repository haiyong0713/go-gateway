package space

import (
	"time"

	ott "go-gateway/app/app-svr/app-interface/api-dependence/ott-service"

	accv1 "git.bilibili.co/bapis/bapis-go/account/service"
	vipclient "git.bilibili.co/bapis/bapis-go/vip/service"
)

// Mine my center struct
type Mine struct {
	Mid           int64   `json:"mid"`
	Name          string  `json:"name"`
	ShowNameGuide bool    `json:"show_name_guide"`
	Face          string  `json:"face"`
	ShowFaceGuide bool    `json:"show_face_guide"`
	Coin          float64 `json:"coin"`
	BCoin         float64 `json:"bcoin"`
	Sex           int32   `json:"sex"`
	Rank          int32   `json:"rank"`
	AnswerStatus  int32   `json:"answer_status,omitempty"`
	Silence       int32   `json:"silence"`
	EndTime       int64   `json:"end_time,omitempty"`
	ShowVideoup   int     `json:"show_videoup"`
	ShowCreative  int     `json:"show_creative"`
	Level         int32   `json:"level"`
	VipType       int32   `json:"vip_type"`
	AudioType     int     `json:"audio_type"`
	Dynamic       int64   `json:"dynamic"`
	Following     int64   `json:"following"`
	Follower      int64   `json:"follower"`
	NewFollowers  int64   `json:"new_followers"`
	//涨粉动态提醒时间
	NewFollowersRTime int64 `json:"new_followers_rtime"`
	Official          struct {
		Type int8   `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify"`
	Pendant           *Pendant         `json:"pendant,omitempty"`
	Sections          []*Section       `json:"sections,omitempty"`
	IpadSections      []*SectionItem   `json:"ipad_sections,omitempty"`
	IpadUpperSections []*SectionItem   `json:"ipad_upper_sections,omitempty"`
	VIPSection        *VIPSection      `json:"vip_section,omitempty"`
	VIPSectionV2      *VIPSectionV2    `json:"vip_section_v2,omitempty"`
	VIPSectionRight   *VIPSectionRight `json:"vip_section_right,omitempty"`
	Vip               VipInfo          `json:"vip,omitempty"`
	MallHome          *MallHome        `json:"mall_home,omitempty"`
	Answer            *Answer          `json:"answer,omitempty"`
	SectionsV2        []*SectionV2     `json:"sections_v2,omitempty"`
	InRegAudit        int32            `json:"in_reg_audit"`
	FirstLiveTime     int64            `json:"first_live_time"`
	LiveTip           *LiveTip         `json:"live_tip,omitempty"`
	Billboard         *BillBoard       `json:"billboard,omitempty"`
	//模块最后更新的时间(仅hd使用)
	SectionUpdateTime *SectionUpdateTime `json:"section_update_time,omitempty"`
	//ipad 新样式
	IpadSectionStyle      int64               `json:"ipad_section_style,omitempty"`
	IpadRecommendSections []*SectionItem      `json:"ipad_recommend_sections,omitempty"`
	IpadMoreSections      []*SectionItem      `json:"ipad_more_sections,omitempty"`
	EnableBiliLink        bool                `json:"enable_bili_link,omitempty"`
	BiliLinkBubble        *ott.BiliLinkBubble `json:"bili_link_bubble,omitempty"`
	FaceNftNew            int32               `json:"face_nft_new"`          // 当前头像是否为nft头像
	ShowNftFaceGuide      bool                `json:"show_nft_face_guide"`   // 是否展示nft头像引导
	GameTip               []*GameTip          `json:"game_tip,omitempty"`    //游戏中心提示条
	SeniorGate            *SeniorGate         `json:"senior_gate,omitempty"` //资深会员进阶测试
	NFT                   *NFT                `json:"nft,omitempty"`         //nft相关信息
	Achievement           Achievement         `json:"achievement"`
	Bubbles               []*Bubble           `json:"bubbles"` //我的页气泡引导相关
}

type BubbleType int64

const (
	SchoolBubble = BubbleType(1)
)

type Bubble struct {
	Icon string     `json:"icon"`
	Type BubbleType `json:"type"`
	ID   int64      `json:"id"`
}

// Achievement 相关成就达成
type Achievement struct {
	SeniorGateFlash *SeniorGateFlash `json:"senior_gate_flash,omitempty"`
	TopLevelFlash   *TopLevelFlash   `json:"top_level_flash,omitempty"`
}

type TopLevelFlash struct {
	Icon string `json:"icon"`
}

type SeniorGateFlash struct {
	Icon string `json:"icon"`
}

type NFT struct {
	//nft的地域类型 1-中国大陆发售 2-港澳台发售
	RegionType int64 `json:"region_type"`
	//角标信息
	NFTIcon `json:"icon"`
}

type NFTIcon struct {
	//角标链接
	Url string `json:"url"`
	//角标显示大小
	ShowStatus int64 `json:"show_status"`
}

type VipInfo struct {
	Type       int32 `json:"type"`
	Status     int32 `json:"status"`
	DueDate    int64 `json:"due_date"`
	VipPayType int32 `json:"vip_pay_type"`
	ThemeType  int32 `json:"theme_type"`
	//给pad端修复愚人节角标用
	ThemeType_ int32    `json:"themeType"`
	Label      VipLabel `json:"label"`
	// 大会员角标，0：无角标，1：粉色大会员角标，2：绿色小会员角标
	AvatarSubscript int32 `json:"avatar_subscript"`
	// 昵称色值，可能为空，色值示例：#FFFB9E60
	NicknameColor string `json:"nickname_color"`
	Role          int64  `json:"role"`
	// 大会员角标链接 仅pc、h5使用
	AvatarSubscriptUrl string `json:"avatar_subscript_url"`
}

type VipLabel struct {
	Path string `json:"path"`
	// 文本值
	Text string `json:"text"`
	// 对应颜色类型，在mod资源中通过：$app_theme_type.$label_theme获取对应标签的颜色配置信息
	LabelTheme string `json:"label_theme"`
	// 文本颜色, 仅pc、h5使用
	TextColor string `json:"text_color"`
	// 背景样式：1:填充 2:描边 3:填充 + 描边 4:背景不填充 + 背景不描边 仅pc、h5使用
	BgStyle int32 `json:"bg_style"`
	// 背景色：#FFFB9E60 仅pc、h5使用
	BgColor string `json:"bg_color"`
	// 边框：#FFFB9E60 仅pc、h5使用
	BorderColor string `json:"border_color"`
	// 大会员铭牌图片
	Image string `json:"image"`
}

type SeniorGate struct {
	// 0-普通注册会员 1-待试炼会员 2-资深会员
	Identity int64 `json:"identity,omitempty"`
	// 进阶测试文案
	Text string `json:"text,omitempty"`
	// 进阶测试跳转链接
	Url string `json:"url,omitempty"`
	// 纯色主题是否开启
	Mode       int64  `json:"mode,omitempty"`
	MemberText string `json:"member_text"`
	// 硬核会员生日祝福
	BirthdayConf *BirthdayConf `json:"birthday_conf"`
}

type BirthdayConf struct {
	Icon       string `json:"icon"`
	Url        string `json:"url"`
	BubbleText string `json:"bubble_text"`
}

type GameTip struct {
	ID      int64  `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Url     string `json:"url,omitempty"`
	//运营位排序字段
	PRank int64 `json:"prank,omitempty"`
	//是否使用定向人群包 1-否 2-是
	IsDirected int64 `json:"is_directed,omitempty"`
	//游戏人群包id
	DmpId int64 `json:"dmp_id,omitempty"`
}

type GameTips []*GameTip

func (gs GameTips) Iter(fn func(g *GameTip)) {
	for _, v := range gs {
		if v == nil {
			continue
		}
		fn(v)
	}
}

type GameDmpReply struct {
	//游戏人群包id
	ID int64 `json:"id"`
	//是否命中 1命中
	Val int64 `json:"val"`
}

type SectionUpdateTime struct {
	ToView    int64 `json:"to_view,omitempty"`
	Favorite  int64 `json:"favorite,omitempty"`
	History   int64 `json:"history,omitempty"`
	Following int64 `json:"following,omitempty"`
}

type BillBoard struct {
	Switch           bool   `json:"switch"`
	Guided           bool   `json:"guided"`
	CharacterUrl     string `json:"character_url"`
	BackgroundId     string `json:"background_id"`
	FullscreenSwitch bool   `json:"fullscreen_switch"`
}

type LiveTip struct {
	//提示条图标
	Icon string `json:"icon"`
	//提示条样式
	Mod int64 `json:"mod"`
	//跳链
	Url string `json:"url"`
	//文案
	Text string `json:"text"`
	//按钮文案
	ButtonText string `json:"button_text"`
	//按钮图案
	ButtonIcon string `json:"button_icon"`
	//跳链文案
	UrlText string `json:"url_text"`
	//配置id（预埋上报功能）
	Id int64 `json:"id"`
}

type ContributeTip struct {
	BeUpTitle  string
	TipIcon    string
	TipTitle   string
	ButtonText string
	ButtonUrl  string
}

type Answer struct {
	// 文案
	Text string `json:"text"`
	// 跳转链接
	URL string `json:"url"`
	// 进度 如10表示10%
	Progress string `json:"progress"`
}

// MallHome .
type MallHome struct {
	Icon  string `json:"icon,omitempty"`
	URI   string `json:"uri,omitempty"`
	Title string `json:"title,omitempty"`
}

// Section for mine page, like 【个人中心】【我的服务】
type Section struct {
	Tp    int            `json:"type,omitempty"`
	Title string         `json:"title,omitempty"`
	Items []*SectionItem `json:"items"`
}

// SectionV2 for mine page, include ios android
type SectionV2 struct {
	Title string         `json:"title,omitempty"`
	Items []*SectionItem `json:"items"`
	// 模块排列样式 1-普通横排 2-纵向列表
	Style int32 `json:"style"`
	// 右上角按钮
	Button *Button `json:"button,omitempty"`
	// 运营模块信息
	MngInfo *MngInfo `json:"mng_info,omitempty"`
	// 模块类型 创作中心-1 其他默认0不返回
	Type int32 `json:"type,omitempty"`
	// 创作中心用 up主标题
	UpTitle string `json:"up_title,omitempty"`
	// 创作中心用 非up主标题
	BeUpTitle string `json:"be_up_title,omitempty"`
	//创作中心提示条图标
	TipIcon string `json:"tip_icon,omitempty"`
	//创作中心提示条文案
	TipTitle string `json:"tip_title,omitempty"`
}

// SectionItem like 【离线缓存】 【历史记录】,a part of section
type SectionItem struct {
	ID           int64         `json:"id,omitempty"`
	Title        string        `json:"title"`
	URI          string        `json:"uri"`
	Icon         string        `json:"icon"`
	NeedLogin    int8          `json:"need_login,omitempty"`
	RedDot       int8          `json:"red_dot,omitempty"`
	GlobalRedDot int8          `json:"global_red_dot,omitempty"`
	Display      int32         `json:"display,omitempty"`
	MngRes       *MngRes       `json:"mng_resource,omitempty"`
	RedDotForNew bool          `json:"red_dot_for_new,omitempty"`
	CommonOpItem *CommonOpItem `json:"common_op_item,omitempty"`
}

type CommonOpItem struct {
	Title              string `json:"title,omitempty"`
	SubTitle           string `json:"sub_title,omitempty"`
	Text               string `json:"text,omitempty"`
	TitleIcon          string `json:"title_icon,omitempty"`
	LinkIcon           string `json:"link_icon,omitempty"`
	LinkType           int64  `json:"link_type,omitempty"`
	TitleColor         string `json:"title_color,omitempty"`
	BackgroundColor    string `json:"background_color,omitempty"`
	LinkContainerColor string `json:"link_container_color,omitempty"`
}

// MngRes is
type MngRes struct {
	IconID int64  `json:"icon_id"`
	Icon   string `json:"icon"`
}

// Button is
type Button struct {
	//按钮文字
	Text string `json:"text,omitempty"`
	//按钮链接
	URL string `json:"url,omitempty"`
	//按钮图标
	Icon string `json:"icon,omitempty"`
	//按钮样式 1-实心 2-空心
	Style int32 `json:"style,omitempty"`
}

// MngInfo is
type MngInfo struct {
	//标题字色
	TitleColor string `json:"title_color,omitempty"`
	//副标题
	Subtitle string `json:"subtitle,omitempty"`
	//副标题跳转链接
	SubtitleURL string `json:"subtitle_url,omitempty"`
	//副标题字色
	SubtitleColor string `json:"subtitle_color,omitempty"`
	//背景图
	Background string `json:"background,omitempty"`
	//背景颜色
	BackgroundColor string `json:"background_color,omitempty"`
}

// Myinfo myinfo
type Myinfo struct {
	Mid            int64              `json:"mid"`
	Name           string             `json:"name"`
	Sign           string             `json:"sign"`
	Coins          float64            `json:"coins"`
	Birthday       string             `json:"birthday"`
	Face           string             `json:"face"`
	FaceNftNew     int32              `json:"face_nft_new"`
	Sex            int                `json:"sex"`
	Level          int32              `json:"level"`
	Rank           int32              `json:"rank"`
	AnswerStatus   int32              `json:"answer_status,omitempty"`
	Silence        int32              `json:"silence"`
	EndTime        int64              `json:"end_time,omitempty"`
	Vip            accv1.VipInfo      `json:"vip"`
	EmailStatus    int32              `json:"email_status"`
	TelStatus      int32              `json:"tel_status"`
	Official       accv1.OfficialInfo `json:"official"`
	Identification int32              `json:"identification"`
	Pendant        *Pendant           `json:"pendant,omitempty"`
	Invite         *Invite            `json:"invite"`
	IsTourist      int32              `json:"is_tourist"`
	PinPrompting   int32              `json:"pin_prompting"`
	InRegAudit     int32              `json:"in_reg_audit"`
	HasFaceNft     bool               `json:"has_face_nft"`
}

// Invite .
type Invite struct {
	InviteRemind int64 `json:"invite_remind"`
	Display      bool  `json:"display"`
}

// MineParam struct
type MineParam struct {
	MobiApp       string `form:"mobi_app"`
	Device        string `form:"device"`
	Build         int    `form:"build"`
	Platform      string `form:"platform"`
	Mid           int64  `form:"mid"`
	Filtered      string `form:"filtered"`
	TeenagersMode int    `form:"teenagers_mode"`
	LessonsMode   int    `form:"lessons_mode"`
	Lang          string `form:"lang" default:"hans"`
	Channel       string `form:"channel"`
	SLocale       string `form:"s_locale"`
	CLocale       string `form:"c_locale"`
	BiliLinkNew   int64  `form:"bili_link_new"`
}

type ConfigSet struct {
	Buvid        string `json:"buvid"`
	Mid          int64  `json:"mid"`
	AdSpecial    int    `json:"ad_special"`
	SensorAccess int    `json:"sensor_access"`
}

// Pendant struct
type Pendant struct {
	Image        string `json:"image"`
	ImageEnhance string `json:"image_enhance"`
}

// VIPSection struct
type VIPSection struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
	STime int64  `json:"start_time,omitempty"`
	ETime int64  `json:"end_time,omitempty"`
}

// MineParam struct
type TeenagersPwdParam struct {
	Pwd         string `form:"pwd" validate:"required"`
	DeviceModel string `form:"device_model"`
}

// VIPSectionV2 is
type VIPSectionV2 struct {
	ID       int64  `json:"id,omitempty"`
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Desc     string `json:"desc,omitempty"`
	URL      string `json:"url,omitempty"`
}

type VIPSectionRight struct {
	ID    int64  `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Link  string `json:"link,omitempty"`
	Tip   string `json:"tip,omitempty"`
	Img   string `json:"img,omitempty"`
}

// FromVIPSectionV2 is
func (m *Mine) FromVIPSectionV2(vipTips *vipclient.TipsVipDetail, url string) {
	if vipTips == nil {
		return
	}
	m.VIPSectionV2 = &VIPSectionV2{
		Title:    "我的大会员",
		URL:      url,
		Desc:     vipTips.Tip,
		Subtitle: "了解更多权益",
	}
	if m.Vip.Status != 1 && (m.Vip.Status == 0 && m.Vip.DueDate > 0) {
		return
	}
	//nolint:gomnd
	switch m.Vip.Type {
	case 1, 2:
		tm := time.Unix(0, m.Vip.DueDate*int64(time.Millisecond))
		m.VIPSectionV2.Subtitle = tm.Format("2006-01-02") + "到期"
		//nolint:gomnd
		if m.Vip.Type == 2 {
			m.VIPSectionV2.Title = "我的年度大会员"
		}
	}
}

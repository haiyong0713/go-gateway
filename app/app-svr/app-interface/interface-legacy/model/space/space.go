package space

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	artm "go-gateway/app/app-svr/app-interface/interface-legacy/model/article"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/audio"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/comic"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/community"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/elec"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/favorite"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/game"
	mallmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/mall"
	"go-gateway/app/app-svr/archive/service/api"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"
	spaceclient "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/pkg/idsafe/bvid"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	cheeseGRPC "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcappcard "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	uparc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	digitalgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/digital/service"
)

const (
	_noticeTextColor      = "#FAAB4B"
	_giftTextColor        = "#FB7299"
	_noticeBgColor        = "#FFFBF0"
	_giftBgColor          = "#FFF0F1"
	_noticeTextColorNight = "#BA833F"
	_giftTextColorNight   = "#BB5B76"
	_noticeBgColorNight   = "#333029"
	_giftBgColorNight     = "#332929"
	_upArcTypeAddToView   = "addtoview"
	_upArcTypeShare       = "share"
)

// Space struct
type Space struct {
	Relation          int              `json:"relation"`
	RelSpecial        int32            `json:"rel_special,omitempty"`
	GuestRelation     int              `json:"guest_relation,omitempty"`
	GuestSpecial      int32            `json:"guest_special,omitempty"`
	Medal             int64            `json:"medal,omitempty"`
	Attention         uint32           `json:"attention,omitempty"`
	AnchorPoint       int              `json:"anchor_point,omitempty"`
	DefaultTab        string           `json:"default_tab,omitempty"`
	IsParams          bool             `json:"is_params,omitempty"`
	Setting           *Setting         `json:"setting,omitempty"`
	Tab               *Tab             `json:"tab,omitempty"`
	Card              *Card            `json:"card,omitempty"`
	Space             *Mob             `json:"images,omitempty"`
	Shop              *Shop            `json:"shop,omitempty"`
	Live              json.RawMessage  `json:"live,omitempty"`
	Elec              *elec.NewInfo    `json:"elec,omitempty"`
	Archive           *ArcList         `json:"archive,omitempty"`
	Series            *SeriesList      `json:"series,omitempty"`
	PlayGame          *GameList        `json:"play_game,omitempty"`
	Article           *ArticleList     `json:"article,omitempty"`
	Clip              *ClipList        `json:"clip,omitempty"`
	Album             *AlbumList       `json:"album,omitempty"`
	Favourite         *FavList         `json:"favourite,omitempty"`
	Season            *BangumiList     `json:"season,omitempty"`
	CoinArc           *ArcList         `json:"coin_archive,omitempty"`
	LikeArc           *ArcList         `json:"like_archive,omitempty"`
	Audios            *AudioList       `json:"audios,omitempty"`
	Community         *CommuList       `json:"community,omitempty"`
	Favourite2        *FavList2        `json:"favourite2,omitempty"`
	Mall              *MallItem        `json:"mall,omitempty"`
	Comic             *ComicList       `json:"comic,omitempty"`
	UGCSeason         *UGCSeasonList   `json:"ugc_season,omitempty"`
	AdSourceContent   json.RawMessage  `json:"ad_source_content,omitempty"`
	AdSourceContentV2 json.RawMessage  `json:"ad_source_content_v2,omitempty"` //新空间商品店铺信息
	AdShopType        int              `json:"ad_shop_type,omitempty"`         //商品tab类型1:店铺,2:橱窗
	Cheese            *CheeseList      `json:"cheese,omitempty"`
	Guard             *Guard           `json:"guard,omitempty"` //大航海信息
	SubComic          *SubComicList    `json:"sub_comic,omitempty"`
	LeadDownload      *OfficialItem    `json:"lead_download,omitempty"` //官号引导下载
	AttentionTip      *AttentionTip    `json:"attention_tip,omitempty"`
	FansDress         *GarbDressReply  `json:"fans_dress,omitempty"`
	FansEffect        *FansEffect      `json:"fans_effect,omitempty"`      //粉丝彩蛋
	DisableUpRcmd     bool             `json:"disable_up_rcmd,omitempty"`  // 相关用户推荐开关
	HiddenAttribute   *HiddenAttribute `json:"hidden_attribute,omitempty"` // 拉黑互不可见空间内容信息
	NftShowModule     *NftShowModule   `json:"nft_show_module,omitempty"`  // 数字艺术品空间展示
	VipSpaceLabel     *VipSpaceLabel   `json:"vip_space_label,omitempty"`  // 大会员标识动效
	// 5.57 new tab
	Tab2     []*TabItem `json:"tab2,omitempty"`
	Activity *Activity  `json:"activity,omitempty"`
	// created activity
	CreatedActivity []*CreatedActivity `json:"created_activity,omitempty"`
	// 空间预约卡信息
	ReservationCardInfo *UpActReserveRelationInfo   `json:"reservation_card_info,omitempty"`
	PreferSpaceTab      bool                        `json:"prefer_space_tab,omitempty"`
	ReservationCardList []*UpActReserveRelationInfo `json:"reservation_card_list,omitempty"`
	ContractResource    *ContractResource           `json:"contract_resource,omitempty"`
	// NFT头像按钮文案
	NftFaceButton *NftFaceButton `json:"nft_face_button"`
}

type VipSpaceLabel struct {
	ShowExpire     bool   `json:"show_expire"`
	ExpireTextFrom string `json:"expire_text_from,omitempty"`
	ExpireTextTo   string `json:"expire_text_to,omitempty"`
	LottieUri      string `json:"lottie_uri,omitempty"`
}

type HiddenAttribute struct {
	IsSpaceHidden bool   `json:"is_space_hidden,omitempty"`
	Text          string `json:"text,omitempty"`
}

type ReserveActSkin struct {
	Svga      string `json:"svga,omitempty"`
	LastImg   string `json:"last_img,omitempty"`
	PlayTimes int64  `json:"play_times,omitempty"`
}

type ReserveActExtra struct {
	Skin   *ReserveActSkin `json:"skin,omitempty"`
	ActUrl string          `json:"act_url,omitempty"`
}

func (i *ReserveActExtra) FromReserveActExtra(s *activitygrpc.ReserveDoveActRelationInfo) {
	i.ActUrl = s.ActUrl
	if s.Skin != nil {
		i.Skin.Svga = s.Skin.Svga
		i.Skin.LastImg = s.Skin.LastImg
		i.Skin.PlayTimes = s.Skin.PlayTimes
	}
}

type UpActReserveRelationInfo struct {
	Sid                int64                                        `json:"sid,omitempty"`
	Name               string                                       `json:"name,omitempty"`
	Total              int64                                        `json:"total,omitempty"`
	Stime              xtime.Time                                   `json:"stime,omitempty"`
	Etime              xtime.Time                                   `json:"etime,omitempty"`
	IsFollow           int64                                        `json:"is_follow,omitempty"`
	State              activitygrpc.UpActReserveRelationState       `json:"state,omitempty"`
	Oid                string                                       `json:"oid,omitempty"`
	Type               activitygrpc.UpActReserveRelationType        `json:"type,omitempty"`
	Upmid              int64                                        `json:"up_mid,omitempty"`
	ReserveRecordCtime xtime.Time                                   `json:"reserve_record_ctime,omitempty"`
	LivePlanStartTime  xtime.Time                                   `json:"live_plan_start_time,omitempty"`
	DescText1          HighlightText                                `json:"desc_text_1,omitempty"`
	DescText2          string                                       `json:"desc_text_2,omitempty"`
	ShowText2          bool                                         `json:"show_text_2"`
	AttachedBadgeText  string                                       `json:"attached_badge_text"`
	DynamicId          string                                       `json:"dynamic_id,omitempty"`
	IsDynamicValid     bool                                         `json:"is_dynamic_valid,omitempty"`
	ReserveActExtra    ReserveActExtra                              `json:"reserve_act_extra,omitempty"`
	LotteryType        activitygrpc.UpActReserveRelationLotteryType `json:"lottery_type,omitempty"`
	LotteryPrizeInfo   *LotteryPrizeInfo                            `json:"lottery_prize_info,omitempty"`
}

type HighlightText struct {
	// 展示文本
	Text string `json:"text,omitempty"`
	// 高亮类型
	TextStyle int8 `json:"text_style,omitempty"`
	// 跳转链接
	JumpUrl string `json:"jump_url,omitempty"`
	// icon
	Icon string `json:"icon,omitempty"`
}

func (i *UpActReserveRelationInfo) FromUpActReserveRelationInfo(s *activitygrpc.UpActReserveRelationInfo) {
	i.Sid = s.Sid
	i.Name = s.Title
	i.Total = s.Total
	i.Stime = s.Stime
	i.Etime = s.Etime
	i.IsFollow = s.IsFollow
	i.State = s.State
	i.Oid = s.Oid
	i.Type = s.Type
	i.Upmid = s.Upmid
	i.ReserveRecordCtime = s.ReserveRecordCtime
	i.LivePlanStartTime = s.LivePlanStartTime
	i.DescText1 = constructReserveDescText1(s.Type, s.LivePlanStartTime, s.Desc)
	i.DescText2 = constructReserveDescText2(s.Total)
	i.AttachedBadgeText = constructReserveAttatchedBadgeText(s.Type, s.Ext)
	i.ShowText2 = isReserveShowText2(s.Total, s.ReserveTotalShowLimit)
	i.DynamicId = s.DynamicId
}

func constructReserveAttatchedBadgeText(typ activitygrpc.UpActReserveRelationType, ext string) string {
	switch typ {
	case activitygrpc.UpActReserveRelationType_Live:
		if ext == "" {
			return ""
		}
		tmp := &activitygrpc.UpActReserveRelationInfoExtend{}
		// 大航海
		if err := json.Unmarshal([]byte(ext), &tmp); err == nil && tmp.SubType == 1 {
			return "大航海专属"
		}
	default:
	}
	return ""
}

func isReserveShowText2(total int64, limit int64) bool {
	return total >= limit
}

func constructReserveDescText1(typ activitygrpc.UpActReserveRelationType, time xtime.Time, desc string) HighlightText {
	if time <= 0 {
		return HighlightText{}
	}
	var res HighlightText
	switch typ {
	case activitygrpc.UpActReserveRelationType_Archive:
		res.Text = fmt.Sprintf("预计%s发布", model.PubTimeToString(time.Time()))
	case activitygrpc.UpActReserveRelationType_Live:
		res.Text = fmt.Sprintf("%s直播", model.PubTimeToString(time.Time()))
	case activitygrpc.UpActReserveRelationType_ESports:
		res.Text = desc
	case activitygrpc.UpActReserveRelationType_Premiere:
		res.Text = fmt.Sprintf("%s首映", model.PubTimeToString(time.Time()))
	case activitygrpc.UpActReserveRelationType_Course:
		res.Text = fmt.Sprintf("%s开售", model.PubTimeToString(time.Time()))
	default:
	}
	return res
}

func constructReserveDescText2(total int64) string {
	return model.StatNumberToString(total, "人预约")
}

func (i *UpActReserveRelationInfo) FromUpActReserveLotteryInfo(s *activitygrpc.UpActReserveRelationInfo) {
	const (
		_upActReserveRelationLotteryTypeCron = 1 // 预约抽奖类型定时抽奖
		// 预约抽奖预制icon
		_upActReserveRelationLotteryIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/rgHplMQyiX.png"
	)
	switch s.LotteryType {
	case _upActReserveRelationLotteryTypeCron:
		i.LotteryType = s.LotteryType
		if s.PrizeInfo != nil {
			i.LotteryPrizeInfo = &LotteryPrizeInfo{
				Text:        s.PrizeInfo.Text,
				LotteryIcon: _upActReserveRelationLotteryIcon,
				JumpUrl:     s.PrizeInfo.JumpUrl,
			}
		}
	default:
	}
}

type LotteryPrizeInfo struct {
	Text        string `json:"text,omitempty"`
	LotteryIcon string `json:"lottery_icon,omitempty"`
	JumpUrl     string `json:"jump_url,omitempty"`
}

type CreatedActivity struct {
	Name      string `json:"name"`
	TopicID   int64  `json:"topic_id"`
	View      int64  `json:"view"`
	Discuss   int64  `json:"discuss"`
	URI       string `json:"uri"`
	Cover     string `json:"cover"`
	CoverMd5  string `json:"cover_md5"`
	TopicType int64  `json:"topic_type,omitempty"`
}

type FansEffect struct {
	Show        bool   `json:"show,omitempty"`
	ResourceID  string `json:"resource_id,omitempty"`
	AchieveType int32  `json:"achieve_type,omitempty"` // 1 千万粉丝 2 百万粉丝
}

type AttentionTip struct {
	// 组内滑出X个视频卡片
	CardNum int `json:"card_num,omitempty"`
	// 提示文案
	Tip string `json:"tip,omitempty"`
}

// GuardList .
type Guard struct {
	URI       string       `json:"uri,omitempty"`
	Desc      string       `json:"desc,omitempty"`
	HighLight string       `json:"high_light,omitempty"`
	Item      []*GuardList `json:"item,omitempty"`
	ButtonMsg string       `json:"button_msg,omitempty"`
}

// GuardList .
type GuardList struct {
	Mid  int64  `json:"mid"`
	Face string `json:"face"`
}

// LikesTmp .
type LikesTmp struct {
	LikeNum int64  `json:"like_num"`
	SkrTip  string `json:"skr_tip,omitempty"`
}

// Card struct
type Card struct {
	Mid              string                `json:"mid"`
	Name             string                `json:"name"`
	Approve          bool                  `json:"approve"`
	Sex              string                `json:"sex"`
	Rank             string                `json:"rank"`
	Face             string                `json:"face"`
	DisplayRank      string                `json:"DisplayRank"`
	Regtime          int64                 `json:"regtime"`
	Spacesta         int                   `json:"spacesta"`
	Birthday         string                `json:"birthday"`
	Place            string                `json:"place"`
	Description      string                `json:"description"`
	Article          int                   `json:"article"`
	Attentions       []int64               `json:"attentions"`
	Fans             int                   `json:"fans"`
	Friend           int                   `json:"friend"`
	Attention        int                   `json:"attention"`
	Sign             string                `json:"sign"`
	LevelInfo        LevelInfo             `json:"level_info"`
	Pendant          account.PendantInfo   `json:"pendant"`
	Nameplate        account.NameplateInfo `json:"nameplate"`
	OfficialVerify   OfficialInfo          `json:"official_verify"`
	ProfessionVerify ProfessionVerify      `json:"profession_verify"`
	Vip              struct {
		Type          int          `json:"vipType"`
		DueDate       int64        `json:"vipDueDate"`
		DueRemark     string       `json:"dueRemark"`
		AccessStatus  int          `json:"accessStatus"`
		VipStatus     int          `json:"vipStatus"`
		VipStatusWarn string       `json:"vipStatusWarn"`
		ThemeType     int          `json:"themeType"`
		Label         VipLabelInfo `json:"label"`
	} `json:"vip"`
	FansGroup     int       `json:"fans_group,omitempty"`
	Audio         int       `json:"audio,omitempty"`
	FansUnread    bool      `json:"fans_unread,omitempty"`
	Silence       int32     `json:"silence"`
	EndTime       int64     `json:"end_time"`
	SilenceURL    string    `json:"silence_url"`
	Likes         *LikesTmp `json:"likes,omitempty"`
	Achieve       *Achieve  `json:"achieve,omitempty"`
	PendantURL    string    `json:"pendant_url,omitempty"`
	PendantTitle  string    `json:"pendant_title,omitempty"`
	PRInfo        *PRInfo   `json:"pr_info,omitempty"`
	BBQ           *BBQ      `json:"bbq,omitempty"`
	IsFakeAccount int32     `json:"-"`
	// 回粉
	Relation        *Relation        `json:"relation"`
	IsDeleted       int32            `json:"is_deleted"`
	Honours         Honours          `json:"honours"`
	LiveFansWearing *LiveFansWearing `json:"live_fans_wearing,omitempty"`
	Profession      Profession       `json:"profession"`
	School          School           `json:"school"`
	SpaceTag        []*SpaceTag      `json:"space_tag"`
	FaceNftNew      int32            `json:"face_nft_new"`
	HasFaceNft      bool             `json:"has_face_nft"`
	NftFaceJump     string           `json:"nft_face_jump,omitempty"`
	NftCertificate  *NftCertificate  `json:"nft_certificate,omitempty"`
	PickupEntrance  *PickupEntrance  `json:"entrance,omitempty"`
	NftId           string           `json:"nft_id"`
	// NFT头像icon
	NftFaceIcon *NftFaceIcon `json:"nft_face_icon"`
	// 空间底部tag
	SpaceTagBottom []*SpaceTagBottom `json:"space_tag_bottom,omitempty"`
}

type NftCertificate struct {
	DetailUrl string `json:"detail_url,omitempty"`
}

// SpaceTag is
type SpaceTag struct {
	Type                 string `json:"type"`
	Title                string `json:"title"`
	TextColor            string `json:"text_color"`
	NightTextColor       string `json:"night_text_color"`
	BackgroundColor      string `json:"background_color"`
	NightBackgroundColor string `json:"night_background_color"`
	URI                  string `json:"uri"`
	Schema               string `json:"schema,omitempty"` // 跳转到第三方app链接
	Icon                 string `json:"icon"`
}

type SpaceTagBottom struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	URI    string `json:"uri"`
	Schema string `json:"schema,omitempty"` // 跳转到第三方app链接
	Icon   string `json:"icon"`
}

type School struct {
	SchoolId int64  `json:"school_id,omitempty"`
	Name     string `json:"name,omitempty"`
}

type Profession struct {
	Id       int32  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	ShowName string `json:"show_name,omitempty"`
}

type LiveFansWearing struct {
	Level            int64  `json:"level,omitempty"`
	MedalName        string `json:"medal_name,omitempty"`
	MedalColorStart  int64  `json:"medal_color_start,omitempty"`
	MedalColorEnd    int64  `json:"medal_color_end,omitempty"`
	MedalColorBorder int64  `json:"medal_color_border,omitempty"`
	GuardIcon        string `json:"guard_icon,omitempty"`
	ShowDefaultIcon  bool   `json:"show_default_icon,omitempty"`
	MedalJumpUrl     string `json:"medal_jump_url,omitempty"`
}

type Honours struct {
	Colour HonourStyle          `json:"colour"`
	Tags   []*account.HonourTag `json:"tags"`
}

type HonourStyle struct {
	Dark   string `json:"dark"`
	Normal string `json:"normal"`
}

// BBQ .
type BBQ struct {
	URI    string `json:"uri,omitempty"`
	Schema string `json:"schema,omitempty"`
}

// PRInfo pr info.
type PRInfo struct {
	MID     int64  `json:"mid,omitempty"`
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
	// 公告配置类型，1-其他类型，2-去世公告
	NoticeType int `json:"notice_type,omitempty"`
	// 提示条icon
	Icon string `json:"icon,omitempty"`
	// 夜间提示条icon
	IconNight string `json:"icon_night,omitempty"`
	// 文字色
	TextColor string `json:"text_color,omitempty"`
	// 背景色
	BgColor string `json:"bg_color,omitempty"`
	// 夜间文字色
	TextColorNight string `json:"text_color_night,omitempty"`
	// 夜间背景色
	BgColorNight string `json:"bg_color_night,omitempty"`
}

// Achieve .
type Achieve struct {
	IsDefault  bool   `json:"is_default"`
	Image      string `json:"image,omitempty"`
	AchieveURL string `json:"achieve_url,omitempty"`
}

// Mob struct
type Mob struct {
	TopPhoto
	Archive        *TopPhotoArchive `json:"archive,omitempty"`
	GarbInfo       *GarbInfo        `json:"garb,omitempty"`
	CharacterInfo  *CharacterInfo   `json:"character,omitempty"`
	HasGarb        bool             `json:"has_garb,omitempty"`
	ShowReset      bool             `json:"show_reset,omitempty"`
	GoodsAvailable bool             `json:"goods_available,omitempty"`
	PurchaseButton *GarbButton      `json:"purchase_button,omitempty"`
	ShowSetArchive bool             `json:"show_set_archive,omitempty"`
	SetArchiveText string           `json:"set_archive_text,omitempty"`
	ShowCharacter  bool             `json:"show_character,omitempty"`
	DigitalInfo    *DigitalInfo     `json:"digital_info,omitempty"`
	ShowDigital    bool             `json:"show_digital,omitempty"`
}

type DigitalInfo struct {
	Active              bool                    `json:"active"`
	HeadUrl             string                  `json:"head_url,omitempty"`
	ItemId              int64                   `json:"item_id,omitempty"`
	NftId               string                  `json:"nft_id,omitempty"`
	JumpUrl             string                  `json:"jump_url,omitempty"`
	RegionType          int32                   `json:"region_type,omitempty"` // nft所属区域 0 默认 1 大陆 2 港澳台
	Icon                string                  `json:"icon,omitempty"`
	AnimationUrlList    []string                `json:"animation_url_list,omitempty"`
	NftType             int32                   `json:"nft_type"`
	BackgroundHandle    int32                   `json:"background_handle"`
	AnimationFirstFrame string                  `json:"animation_first_frame"`
	MusicAlbum          *digitalgrpc.MusicAlbum `json:"music_album"`
	Animation           *digitalgrpc.Animation  `json:"animation"`
	NftRegionTitle      string                  `json:"nft_region_title"`
	// NFT图片相关元数据
	NFTImage *digitalgrpc.NFTImage `json:"nft_image,omitempty"`
}

type CharacterInfo struct {
	IsActive bool `json:"is_active"`
}

type TopPhotoArchive struct {
	Aid int64 `json:"aid"`
	Cid int64 `json:"cid"`
	//	"uri": "bilibili://video/fullscreen/{aid}/{cid}/?player_preload=xxxxx&player_width=375&player_height=400&player_rotate=0&bvid=BV1F4411A7DB",
	URI      string `json:"uri"`
	ImageURL string `json:"image_url"`
}

// GarbButton def.
type GarbButton struct {
	URI   string `json:"uri"`
	Title string `json:"title"`
}

type GarbInfo struct {
	GarbID        int64  `json:"garb_id"`
	ImageID       int64  `json:"image_id"`
	SmallImage    string `json:"small_image"`
	LargeImage    string `json:"large_image"`
	FansLabel     string `json:"fans_label"`
	FansNumber    string `json:"fans_number,omitempty"`
	Mp4Vertical   string `json:"mp4_vertical,omitempty"`
	Mp4Horizontal string `json:"mp4_horizontal,omitempty"`
	Mp4PlayMode   string `json:"mp4_play_mode,omitempty"`
}

type OfficialItem struct {
	Uid    int64  `json:"uid"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Scheme string `json:"scheme"`
	Rcmd   string `json:"rcmd"`
	URL    string `json:"url"`
	Button string `json:"button"`
}

type McnInfo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// FromEquip def.
func (v *GarbInfo) FromEquip(equip *garbgrpc.SpaceBGUserEquipReply, fansNbr int64) (showGarb bool) {
	if equip == nil || equip.Item == nil || len(equip.Item.Images) <= int(equip.Index) ||
		equip.Item.Images[equip.Index] == nil { // 还是出头图
		return
	}
	currentImg := equip.Item.Images[equip.Index]
	v.ImageID = equip.Index
	v.GarbID = equip.Item.Id
	v.LargeImage = currentImg.Portrait
	v.SmallImage = currentImg.Landscape
	v.FansLabel = garbTitle(equip.Item.Name)
	if fansNbr > 0 {
		v.FansNumber = fansNbrLabel(fansNbr)
	}
	v.Mp4Vertical = currentImg.Mp4Vertical
	v.Mp4Horizontal = currentImg.Mp4Horizontal
	v.Mp4PlayMode = currentImg.Mp4PlayMode
	return true
}

type TopPhoto struct {
	ImgURL      string `json:"imgUrl"`
	NightImgURL string `json:"night_imgurl"`
}

// Shop struct
type Shop struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// LevelInfo struct
type LevelInfo struct {
	Cur           int32       `json:"current_level"`
	Min           int32       `json:"current_min"`
	NowExp        int32       `json:"current_exp"`
	NextExp       interface{} `json:"next_exp"`
	Identity      int64       `json:"identity"` // 0-普通注册会员 1-待试炼会员 2-资深会员
	SeniorInquiry struct {
		InquiryText string `json:"inquiry_text"`
		InquiryUrl  string `json:"inquiry_url"`
	} `json:"senior_inquiry"`
}

// PendantInfo struct
type PendantInfo struct {
	Pid    int    `json:"pid"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Expire int    `json:"expire"`
}

// NameplateInfo struct
type NameplateInfo struct {
	Nid        int    `json:"nid"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	ImageSmall string `json:"image_small"`
	Level      string `json:"level"`
	Condition  string `json:"condition"`
}

// OfficialInfo struct
type OfficialInfo struct {
	Type  int8   `json:"type"`
	Desc  string `json:"desc"`
	Role  int32  `json:"role"`
	Title string `json:"title"`
}

type ProfessionVerify struct {
	Icon     string `json:"icon"`
	ShowDesc string `json:"show_desc"`
}

// Setting struct
type Setting struct {
	Channel    int `json:"channel,omitempty"`
	FavVideo   int `json:"fav_video"`
	CoinsVideo int `json:"coins_video"`
	LikesVideo int `json:"likes_video"`
	Bangumi    int `json:"bangumi"`
	PlayedGame int `json:"played_game"`
	Groups     int `json:"groups"`
	Comic      int `json:"comic"`
	BBQ        int `json:"bbq"`
	DressUp    int `json:"dress_up"`
	// 客户端展示为：公开显示关注列表
	// DisableFollowing=0 公开关注列表
	// DisableFollowing=1 不公开关注列表
	// 由客户端取反展示
	DisableFollowing  int `json:"disable_following"`
	LivePlayback      int `json:"live_playback"`
	CloseSpaceMedal   int `json:"close_space_medal"`
	OnlyShowWearing   int `json:"only_show_wearing"`
	DisableShowSchool int `json:"disable_show_school"`
	DisableShowNft    int `json:"disable_show_nft"`
}

// ArcList struct
type ArcList struct {
	EpisodicButton *EpisodicButton `json:"episodic_button,omitempty"`
	Order          []*ArcOrder     `json:"order,omitempty"`
	Count          int             `json:"count"`
	Item           []*ArcItem      `json:"item"`
}

// SeriesList struct
type SeriesList struct {
	Item []*SeriesItem `json:"item"`
}

// SeriesArchiveList struct
type SeriesArchiveList struct {
	Next           int64           `json:"next"`
	Item           []*ArcItem      `json:"item"`
	EpisodicButton *EpisodicButton `json:"episodic_button,omitempty"`
	Order          []*ArcOrder     `json:"order,omitempty"`
}

type ArcOrder struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type EpisodicButton struct {
	Text string `json:"text"`
	Uri  string `json:"uri"`
}

type GameList struct {
	Count int         `json:"count"`
	Item  []*GameItem `json:"item"`
}

type GameListSub struct {
	Count int            `json:"count"`
	Item  []*GameItemSub `json:"item"`
	Uri   string         `json:"uri,omitempty"`
	Image string         `json:"image,omitempty"`
}

// UGCSeasonList struct
type UGCSeasonList struct {
	Count int64            `json:"count"`
	Item  []*UGCSeasonItem `json:"item"`
}

// CheeseList struct
type CheeseList struct {
	Count int64         `json:"count"`
	Item  []*CheeseItem `json:"item"`
}

// ComicList struct
type ComicList struct {
	Count int          `json:"count"`
	Item  []*ComicItem `json:"item"`
}

// SubComicList struct
type SubComicList struct {
	Count int             `json:"count"`
	Item  []*SubComicItem `json:"item"`
}

// ArticleList struct
type ArticleList struct {
	Count      int             `json:"count"`
	Item       []*ArticleItem  `json:"item"`
	ListsCount int             `json:"lists_count"`
	Lists      []*article.List `json:"lists"`
}

// CommuList struct
type CommuList struct {
	Count int         `json:"count"`
	Item  []*CommItem `json:"item"`
}

// FavList struct
type FavList struct {
	Count int                `json:"count"`
	Item  []*favorite.Folder `json:"item"`
}

// FavList2 struct
type FavList2 struct {
	Count int                 `json:"count"`
	Item  []*favorite.Folder2 `json:"item"`
}

// BangumiList struct
type BangumiList struct {
	Count int            `json:"count"`
	Item  []*BangumiItem `json:"item"`
}

// AudioList struct
type AudioList struct {
	Count int          `json:"count"`
	Item  []*AudioItem `json:"item"`
}

// ClipList struct
type ClipList struct {
	Count  int     `json:"count"`
	More   int     `json:"has_more"`
	Offset int     `json:"next_offset"`
	Item   []*Item `json:"item"`
}

// AlbumList struct
type AlbumList struct {
	Count  int     `json:"count"`
	More   int     `json:"has_more"`
	Offset int     `json:"next_offset"`
	Item   []*Item `json:"item"`
}

// GameItem .
type GameItem struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Icon  string  `json:"icon"`
	Grade float64 `json:"grade"`
	URI   string  `json:"uri"`
}

type GameItemSub struct {
	GameItem
	Tag            []string `json:"tag"`
	Title          string   `json:"title"`
	Content        string   `json:"content"`
	Button         string   `json:"button"`
	BgColor        string   `json:"bg_color"`
	TextColor      string   `json:"text_color"`
	BgColorNight   string   `json:"bg_color_night"`
	TextColorNight string   `json:"text_color_night"`
}

// SeriesItem struct
type SeriesItem struct {
	SeriesId       int64      `json:"series_id,omitempty"`
	Name           string     `json:"name"`
	IsLivePlayBack bool       `json:"-"`
	Mtime          xtime.Time `json:"-"`
}

// ArcItem struct
type ArcItem struct {
	Title          string `json:"title"`
	Subtitle       string `json:"subtitle"`
	TypeName       string `json:"tname"`
	Cover          string `json:"cover"`
	URI            string `json:"uri"`
	Param          string `json:"param"`
	Goto           string `json:"goto"`
	Length         string `json:"length"`
	Duration       int64  `json:"duration"`
	IsPopular      bool   `json:"is_popular"`
	IsSteins       bool   `json:"is_steins"`
	IsUGCPay       bool   `json:"is_ugcpay"`
	IsCooperation  bool   `json:"is_cooperation"`
	IsPGC          bool   `json:"is_pgc"`
	IsLivePlayBack bool   `json:"is_live_playback"`
	// av
	Play    int                  `json:"play"`
	Danmaku int                  `json:"danmaku"`
	CTime   xtime.Time           `json:"ctime"`
	UGCPay  int32                `json:"ugc_pay"`
	Badges  []*model.ReasonStyle `json:"badges,omitempty"`
	Author  string               `json:"author,omitempty"`
	State   bool                 `json:"state"`
	BvID    string               `json:"bvid,omitempty"`
	Videos  int64                `json:"videos"`
	// 三点
	ThreePoint []*ThreePoint `json:"three_point,omitempty"`
	FirstCid   int64         `json:"first_cid,omitempty"`
	CursorAttr *CursorAttr   `json:"cursor_attr,omitempty"`
	// 合集信息
	Season *SeasonInfo `json:"season,omitempty"`
	// 付费稿件
	IsPay bool `json:"-"`
}

type ThreePoint struct {
	Type           string `json:"type"`
	Icon           string `json:"icon"`
	Text           string `json:"text"`
	ShareSuccToast string `json:"share_succ_toast,omitempty"`
	ShareFailToast string `json:"share_fail_toast,omitempty"`
	SharePath      string `json:"share_path,omitempty"`
	ShortLink      string `json:"short_link,omitempty"`
	ShareSubtitle  string `json:"share_subtitle,omitempty"`
}

// UGCSeasonItem struct
type UGCSeasonItem struct {
	SeasonId  int64                `json:"season_id"`
	Title     string               `json:"title"`
	Cover     string               `json:"cover"`
	Param     string               `json:"param"`
	URI       string               `json:"uri"`
	Goto      string               `json:"goto"`
	Length    string               `json:"length"`
	Duration  int64                `json:"duration"`
	Play      int                  `json:"play"`
	Danmaku   int                  `json:"danmaku"`
	Count     int64                `json:"count"`
	MTime     xtime.Time           `json:"mtime"`
	Badges    []*model.ReasonStyle `json:"badges,omitempty"`
	IsPay     bool                 `json:"is_pay"`
	IsNoSpace bool                 `json:"is_no_space"`
}

// CheeseItem struct
type CheeseItem struct {
	Title      string         `json:"title"`
	Cover      string         `json:"cover"`
	CoverRight string         `json:"cover_right"`
	Param      string         `json:"param"`
	URI        string         `json:"uri"`
	Goto       string         `json:"goto"`
	Play       int32          `json:"play"`
	MTime      xtime.Time     `json:"mtime"`
	CTime      xtime.Time     `json:"ctime"`
	Badges     []*CheeseBadge `json:"badges,omitempty"`
}

type CheeseBadge struct {
	Mark string `json:"mark,omitempty"`
}

// ComicItem struct
type ComicItem struct {
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	Param  string `json:"param"`
	URI    string `json:"uri"`
	Goto   string `json:"goto"`
	Count  int    `json:"count"`
	Styles string `json:"styles,omitempty"`
	Label  string `json:"label,omitempty"`
}

// SubComicItem struct
type SubComicItem struct {
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	Param  string `json:"param"`
	URI    string `json:"uri"`
	Goto   string `json:"goto"`
	Styles string `json:"styles,omitempty"`
	Label  string `json:"label,omitempty"`
}

// ArticleItem struct
type ArticleItem struct {
	*article.Meta
	URI   string `json:"uri"`
	Param string `json:"param"`
	Goto  string `json:"goto"`
}

// BangumiItem struct
type BangumiItem struct {
	Title         string     `json:"title"`
	Cover         string     `json:"cover"`
	URI           string     `json:"uri"`
	Param         string     `json:"param"`
	Goto          string     `json:"goto"`
	Finish        int8       `json:"finish"`
	Index         string     `json:"index"`
	MTime         xtime.Time `json:"mtime"`
	NewestEpIndex string     `json:"newest_ep_index"`
	IsStarted     int        `json:"is_started"`
	IsFinish      string     `json:"is_finish"`
	NewestEpID    string     `json:"newest_ep_id"`
	TotalCount    string     `json:"total_count"`
	Attention     string     `json:"attention"`
}

// CommItem struct
type CommItem struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Desc           string `json:"desc"`
	Thumb          string `json:"thumb"`
	PostCount      int    `json:"post_count"`
	MemberCount    int    `json:"member_count"`
	PostNickname   string `json:"post_nickname"`
	MemberNickname string `json:"member_nickname"`
}

// AudioItem struct
type AudioItem struct {
	ID       int64      `json:"id"`
	Aid      int64      `json:"aid"`
	UID      int64      `json:"uid"`
	Title    string     `json:"title"`
	Cover    string     `json:"cover"`
	Author   string     `json:"author"`
	Schema   string     `json:"schema"`
	Duration int64      `json:"duration"`
	Play     int        `json:"play"`
	Reply    int        `json:"reply"`
	IsOff    int        `json:"isOff"`
	AuthType int        `json:"authType"`
	CTime    xtime.Time `json:"ctime"`
}

// MallItem struct
type MallItem struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
	Icon string `json:"icon"`
}

type PhotoMall struct {
	Title string           `json:"title"`
	List  []*PhotoMallItem `json:"list"`
}

type PhotoMallItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Img         string `json:"img"`
	NightImg    string `json:"night_img"`
	IsActivated int64  `json:"is_activated,omitempty"`
}

type PhotoTopParm struct {
	MobiApp   string `form:"mobi_app"`
	Device    string `form:"device"`
	ID        string `form:"id"`
	Platform  string `form:"platform"`
	AccessKey string `form:"access_key"`
	Type      int64  `form:"type" default:"1" validate:"min=1,max=2"`
	Oid       int64  `form:"-"`
}

type PickupEntrance struct {
	Icon           string `json:"icon"`
	JumpUrl        string `json:"jump_url"`
	IsShowEntrance bool   `json:"is_show_entrance"`
}

type SeasonInfo struct {
	SeasonId int64      `json:"season_id"`
	Cover    string     `json:"cover"`
	Title    string     `json:"title"`
	Count    int64      `json:"count"`
	Play     int32      `json:"play"`
	Danmaku  int32      `json:"danmaku"`
	Mtime    xtime.Time `json:"mtime"`
}

type VipLabelInfo struct {
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
	// 新版铭牌图片
	Image string `json:"image"`
}

func (i *PhotoMallItem) FromPhotoMallItem(s *spaceclient.PhotoMall) {
	i.ID = s.Id
	i.Name = s.Name
	i.Img = s.Img
	i.NightImg = s.NightImg
	i.IsActivated = s.IsActivated
}

// FromSeason func
func (i *BangumiItem) FromSeason(b *pgcappcard.CardSeasonProto) {
	i.Title = b.Title
	i.Cover = b.Cover
	i.Goto = model.GotoBangumi
	i.Param = strconv.FormatInt(int64(b.SeasonId), 10)
	i.URI = model.FillURI(model.GotoBangumiWeb, strconv.FormatInt(int64(b.SeasonId), 10), nil)
	i.IsStarted = int(b.IsStart)
	i.Finish = int8(b.IsFinish)
	i.TotalCount = strconv.FormatInt(int64(b.TotalCount), 10)
	i.NewestEpIndex = b.NewEp.Title
	if b.Follow {
		i.Attention = "1"
	}
}

// FromCoinArc func
func (i *ArcItem) FromCoinArc(a *api.Arc) {
	if a.IsNormal() {
		i.Title = a.Title
		i.Cover = a.Pic
	} else {
		i.Title = "已失效视频"
		i.Cover = "https://i0.hdslb.com/bfs/archive/be27fd62c99036dce67efface486fb0a88ffed06.jpg"
	}
	i.Param = strconv.FormatInt(int64(a.Aid), 10)
	if a.AttrVal(api.AttrBitIsPGC) == api.AttrYes && a.RedirectURL != "" {
		i.URI = a.RedirectURL
	} else {
		i.URI = model.FillURI(model.GotoAv, i.Param, model.AvHandler(a))
	}
	i.Goto = model.GotoAv
	i.Danmaku = int(a.Stat.Danmaku)
	i.Duration = a.Duration
	i.CTime = a.PubDate
	i.Play = int(a.Stat.View)
	i.State = a.IsNormal()
	if a.AttrValV2(api.AttrBitV2Pay) == api.AttrYes && a.Rights.ArcPayFreeWatch == 0 {
		i.Badges = append(i.Badges, model.NewPayBadge)
	}
	i.Author = a.Author.Name
	if a.Rights.IsCooperation == 1 {
		i.IsCooperation = true
		if i.Author != "" {
			i.Author += " 等联合创作"
		}
	}
}

// FromLikeArc fun
func (i *ArcItem) FromLikeArc(a *api.Arc) {
	if a.IsNormal() {
		i.Title = a.Title
		i.Cover = a.Pic
	} else {
		i.Title = "已失效视频"
		i.Cover = "https://i0.hdslb.com/bfs/archive/be27fd62c99036dce67efface486fb0a88ffed06.jpg"
	}
	i.Param = strconv.FormatInt(int64(a.Aid), 10)
	if a.AttrVal(api.AttrBitIsPGC) == api.AttrYes && a.RedirectURL != "" {
		i.URI = a.RedirectURL
	} else {
		i.URI = model.FillURI(model.GotoAv, i.Param, model.AvHandler(a))
	}
	i.Goto = model.GotoAv
	i.Danmaku = int(a.Stat.Danmaku)
	i.Duration = a.Duration
	i.CTime = a.PubDate
	i.Play = int(a.Stat.View)
	i.State = a.IsNormal()
	if a.AttrValV2(api.AttrBitV2Pay) == api.AttrYes && a.Rights.ArcPayFreeWatch == 0 {
		i.Badges = append(i.Badges, model.NewPayBadge)
	}
	i.Author = a.Author.Name
	if a.Rights.IsCooperation == 1 {
		i.IsCooperation = true
		if i.Author != "" {
			i.Author += " 等联合创作"
		}
	}
}

func (i *ArcItem) ConvertAsOGVEP(ep *pgccardgrpc.EpisodeCard) {
	i.IsPGC = true
	if ep.Season != nil {
		i.Title = ep.Season.Title
	}
	if ep.Meta != nil {
		i.Subtitle = ep.Meta.ShortLongTitle
	}
	i.URI = ep.Url
	i.Cover = ep.Cover
	i.Goto = model.GotoEP
	if ep.PubRealTime != nil {
		i.CTime = xtime.Time(ep.PubRealTime.Seconds)
	}
	if ep.Stat != nil {
		i.Play = int(ep.Stat.Play)
		i.Danmaku = int(ep.Stat.Danmaku)
	}
	i.Duration = ep.Duration
}

// FromArticle func
func (i *ArticleItem) FromArticle(ctx context.Context, a *article.Meta) {
	i.Meta = a
	i.Param = strconv.FormatInt(int64(a.ID), 10)
	articleInfo := artm.GetArticleInfo(ctx, int64(a.Type), a.ID, a.CoverAvid)
	i.URI = articleInfo.Uri
	i.Goto = model.GotoArticle
}

// FromArc func
func (i *ArcItem) FromArc(ap *api.ArcPlayer, popularAIDs map[int64]struct{}, hasShare, livePlaybackBadge bool, season *ugcSeasonGrpc.Season) {
	if ap.Arc == nil {
		return
	}
	c := ap.Arc
	i.Title = c.Title
	i.Cover = c.Pic
	i.TypeName = c.TypeName
	i.Param = strconv.FormatInt(c.Aid, 10)
	playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
	i.URI = model.FillURI(model.GotoAv, i.Param, model.AvPlayHandlerGRPC(ap.Arc, playInfo))
	if c.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
		i.IsPGC = true
		if c.RedirectURL != "" {
			i.URI = c.RedirectURL
		}
	}
	var isLivePlayback bool
	for _, val := range conf.Conf.LivePlayback.UpFrom {
		if c.GetUpFromV2() == val {
			isLivePlayback = true
			break
		}
	}
	i.IsLivePlayBack = isLivePlayback
	i.Goto = model.GotoAv
	i.Danmaku = int(c.Stat.Danmaku)
	i.CTime = c.PubDate
	i.Duration = c.Duration
	i.Play = int(c.Stat.View)
	i.UGCPay = c.Rights.UGCPay
	i.Videos = c.Videos
	i.Author = c.Author.Name
	if c.AttrValV2(api.AttrBitV2Pay) == api.AttrYes && c.Rights.ArcPayFreeWatch == 0 {
		i.IsPay = true
		i.Badges = append(i.Badges, model.NewPayBadge)
	}
	if c.Rights.UGCPay == 1 {
		i.IsUGCPay = true
		i.Badges = append(i.Badges, model.PayBadge)
	}
	if !i.IsPay && season != nil && season.AttrVal(ugcSeasonGrpc.SeasonAttrSnPay) == ugcSeasonGrpc.AttrSnYes {
		// 稿件不是付费稿件,但所属合集为付费,加上付费角标
		i.Badges = append(i.Badges, model.NewPayBadge)
	}
	if livePlaybackBadge && isLivePlayback {
		if season == nil {
			// 标签展示逻辑,如果需要展示合集信息不展示稿件独有的标签
			i.Badges = append(i.Badges, model.LivePlaybackBadge)
		}
	}
	if c.Rights.IsCooperation == 1 {
		i.IsCooperation = true
		if season == nil {
			i.Badges = append(i.Badges, model.CooperationBadge)
		}
		if i.Author != "" {
			i.Author += " 等联合创作"
		}
	}
	if c.AttrVal(api.AttrBitSteinsGate) == api.AttrYes {
		i.IsSteins = true
		if season == nil {
			i.Badges = append(i.Badges, model.SteinsBadge)
		}
	}
	if popularAIDs != nil {
		if _, ok := popularAIDs[c.Aid]; ok {
			i.IsPopular = true
			if season == nil {
				i.Badges = append(i.Badges, model.PopularBadge)
			}
		}
	}
	// 最多展示两个角标
	//nolint:gomnd
	if len(i.Badges) > 2 {
		i.Badges = i.Badges[:2]
	}
	i.BvID, _ = bvid.AvToBv(c.Aid)
	i.ThreePoint = append(i.ThreePoint, &ThreePoint{
		Type: _upArcTypeAddToView,
		Icon: conf.Conf.Custom.UpArcAddToViewIcon,
		Text: conf.Conf.Custom.UpArcAddToViewText,
	})
	if hasShare {
		tmp := &ThreePoint{
			Type:           _upArcTypeShare,
			Icon:           conf.Conf.Custom.UpArcShareIcon,
			Text:           conf.Conf.Custom.UpArcShareText,
			ShareSuccToast: conf.Conf.Custom.UpArcShareSuccToast,
			ShareFailToast: conf.Conf.Custom.UpArcShareFailToast,
			SharePath:      fmt.Sprintf("pages/video/video?avid=%d", c.Aid),
			ShortLink:      fmt.Sprintf("https://b23.tv/av%d", c.Aid),
		}
		if i.BvID != "" {
			tmp.ShortLink = fmt.Sprintf("https://b23.tv/%s", i.BvID)
		}
		//nolint:gomnd
		if c.Stat.View > 100000 {
			tmpView := strconv.FormatFloat(float64(c.Stat.View)/10000, 'f', 1, 64)
			tmp.ShareSubtitle = "已观看" + strings.TrimSuffix(tmpView, ".0") + "万次"
		}
		i.ThreePoint = append(i.ThreePoint, tmp)
	}
	i.FirstCid = c.FirstCid
	if season != nil {
		i.Season = &SeasonInfo{
			SeasonId: season.ID,
			Cover:    season.Cover,
			Title:    season.Title,
			Count:    season.EpCount,
			Play:     season.Stat.View,
			Danmaku:  season.Stat.Danmaku,
			Mtime:    season.Ptime,
		}
	}
}

// FromUGCSeason func
func (usi *UGCSeasonItem) FromUGCSeason(uc *ugcSeasonGrpc.Season) {
	usi.SeasonId = uc.ID
	usi.Title = uc.Title
	usi.Cover = uc.Cover
	usi.Param = strconv.FormatInt(int64(uc.ID), 10)
	usi.URI = model.FillURI(model.GotoAv, strconv.FormatInt(uc.FirstAid, 10), nil)
	usi.Goto = model.GotoAv
	usi.Play = int(uc.Stat.View)
	usi.Danmaku = int(uc.Stat.Danmaku)
	usi.MTime = uc.Ptime
	usi.Count = uc.EpCount
	if uc.SignState == model.UGCSeasonSole {
		usi.Badges = append(usi.Badges, model.UGCSeasonSoleBadge)
	} else if uc.SignState == model.UGCSeasonStarting {
		usi.Badges = append(usi.Badges, model.UGCSeasonStartingBadge)
	}
	if uc.AttrVal(ugcSeasonGrpc.SeasonAttrSnPay) == ugcSeasonGrpc.AttrSnYes {
		usi.IsPay = true
		usi.Badges = append(usi.Badges, model.NewPayBadge)
	}
	if uc.AttrVal(ugcSeasonGrpc.AttrSnNoSpace) == ugcSeasonGrpc.AttrSnYes {
		usi.IsNoSpace = true
	}
}

// FromComic comic
func (i *ComicItem) FromComic(c *comic.Comic, labelTime bool) {
	i.Title = c.Title
	i.Cover = c.VerticalCover
	i.Param = strconv.FormatInt(c.ID, 10)
	i.URI = c.URL
	i.Goto = model.GotoComic
	i.Count = c.Total
	var styles []string
	for _, style := range c.Styles {
		styles = append(styles, style.Name)
	}
	if len(styles) > 0 {
		i.Styles = strings.Join(styles, " ")
	}
	update, _ := strconv.ParseInt(c.LastUpdateTime, 10, 64)
	switch c.IsFinish {
	case ComicStatusSerialization:
		if update != 0 || c.LastShortTitle != "" {
			if labelTime {
				i.Label = fmt.Sprintf("%v更新至%v", time.Unix(update, 0).Format("01-02"), c.LastShortTitle)
			} else {
				i.Label = fmt.Sprintf("更新至%v", c.LastShortTitle)
			}
		}
	case ComicStatusFinished:
		i.Label = fmt.Sprintf("全%v话", c.Total)
	}
}

// FromSubComic sub comic
func (i *SubComicItem) FormSubComic(c *comic.FavComic) {
	i.Title = c.Title
	i.Cover = c.VCover
	i.Param = strconv.FormatInt(c.ComicID, 10)
	i.Goto = model.GotoComic
	i.URI = model.FillURI(model.GotoComic, i.Param, nil)
	//nolint:gomnd
	switch c.Status {
	case 2:
		i.Label = fmt.Sprintf("更新至%v话", c.OrdCount)
	case 3:
		i.Label = fmt.Sprintf("全%v话", c.OrdCount)
	}
}

// FromCommunity func
func (i *CommItem) FromCommunity(c *community.Community) {
	i.ID = c.ID
	i.Name = c.Name
	i.Desc = c.Desc
	i.Thumb = c.Thumb
	i.PostCount = c.PostCount
	i.MemberCount = c.MemberCount
	i.PostNickname = c.PostNickname
	i.MemberNickname = c.MemberNickname
}

// FromAudio func
func (i *AudioItem) FromAudio(a *audio.Audio) {
	i.ID = a.ID
	i.Aid = a.Aid
	i.UID = a.UID
	i.Title = a.Title
	i.Cover = a.Cover
	i.Author = a.Author
	i.Schema = a.Schema
	i.Duration = a.Duration
	i.Play = a.Play
	i.Reply = a.Reply
	i.IsOff = a.IsOff
	i.AuthType = a.AuthType
	i.CTime = a.CTime
}

// FormMall func
func (i *MallItem) FormMall(m *mallmdl.Mall) {
	i.Name = m.Name
	i.Icon = m.Logo
	i.URI = m.URL
}

// FromUGCSeason func
func (ci *CheeseItem) FromCheese(sc *cheeseGRPC.SeasonCard) {
	ci.Title = sc.Title
	ci.Cover = sc.Cover
	ci.CoverRight = sc.UpdateInfo1
	ci.Param = strconv.FormatInt(int64(sc.Id), 10)
	ci.URI = sc.Url
	ci.Goto = model.GotoCheese
	if sc.Stat != nil {
		ci.Play = sc.Stat.View
	}
	ci.MTime = xtime.Time(sc.Mtime)
	ci.CTime = xtime.Time(sc.Ctime)
	var badges []*CheeseBadge
	if sc.Cooperated {
		badges = append(badges, &CheeseBadge{Mark: sc.CooperationMark})
	}
	ci.Badges = badges
}

// FromGame .
func (i *GameItem) FromGame(a *game.PlayGame) {
	i.ID = a.GameBaseID
	i.URI = a.DetailURL
	i.Icon = a.GameIcon
	i.Grade = a.Grade
	i.Name = a.GameName
}

// FromGame .
func (i *GameItemSub) FromGameSub(a *game.PlayGameSub) {
	i.ID = a.GameBaseID
	i.URI = a.DetailURL
	i.Icon = a.GameIcon
	i.Grade = a.Grade
	i.Name = a.GameName
	i.Tag = a.GameTags
	if a.Notice != "" {
		i.Title = model.GameNotice
		i.Content = a.Notice
		i.TextColor = _noticeTextColor
		i.BgColor = _noticeBgColor
		i.TextColorNight = _noticeTextColorNight
		i.BgColorNight = _noticeBgColorNight
	} else if a.GiftTitle != "" {
		i.Title = model.GameGift
		i.Content = a.GiftTitle
		i.TextColor = _giftTextColor
		i.BgColor = _giftBgColor
		i.TextColorNight = _giftTextColorNight
		i.BgColorNight = _giftBgColorNight
	}
	switch a.GameStatus {
	case model.ButtonReserve:
		i.Button = model.ReserveName
	case model.ButtonDownload:
		i.Button = model.DownloadName
	case model.ButtonEnter:
		i.Button = model.EnterName
	}
}

type TabItem struct {
	Title    string     `json:"title"`
	Param    string     `json:"param"`
	SeriesId int64      `json:"series_id,omitempty"`
	Items    []*TabItem `json:"items,omitempty"`
	SeasonId int64      `json:"season_id,omitempty"`
	Mtime    xtime.Time `json:"-"`
}

type Relation struct {
	Status     int8 `json:"status,omitempty"`
	IsFollow   int8 `json:"is_follow,omitempty"`
	IsFollowed int8 `json:"is_followed,omitempty"`
}

// 空间直播粉丝佩戴勋章转换
func WearingInfoChange(wearingInfo *livexfans.WearingResp) *LiveFansWearing {
	const (
		_guardLevelGovernor = 1
		_guardLevelAdmiral  = 2
		_guardLevelCaptain  = 3
		// 直播粉丝勋章大航海icon
		_guardIconGovernor = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/FqYoOmgssP.png"
		_guardIconAdmiral  = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/J9nffR1Oah.png"
		_guardIconCaptain  = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/SREsu5SRPI.png"
	)
	if wearingInfo == nil || wearingInfo.GetWearing() == 0 || wearingInfo.GetMedal() == nil {
		return nil
	}
	var guardIcon string
	switch wearingInfo.GetMedal().GetGuardLevel() {
	// 大航海等级对应关系 1：总督，2：提督，3：舰长
	case _guardLevelGovernor:
		guardIcon = _guardIconGovernor
	case _guardLevelAdmiral:
		guardIcon = _guardIconAdmiral
	case _guardLevelCaptain:
		guardIcon = _guardIconCaptain
	default:
	}
	return &LiveFansWearing{
		Level:            wearingInfo.GetMedal().GetLevel(),
		MedalName:        wearingInfo.GetMedal().GetMedalName(),
		MedalColorStart:  wearingInfo.GetMedal().GetMedalColorStart(),
		MedalColorEnd:    wearingInfo.GetMedal().GetMedalColorEnd(),
		MedalColorBorder: wearingInfo.GetMedal().GetMedalColorBorder(),
		GuardIcon:        guardIcon,
		MedalJumpUrl:     fmt.Sprintf("https://live.bilibili.com/p/html/live-fansmedal-wall/index.html?is_live_webview=1&tId=%d#/medal", wearingInfo.GetMedal().GetUid()),
	}
}

// 互相关注关系转换
func RelationChange(upMid int64, relations map[int64]*relationgrpc.InterrelationReply) (r *Relation) {
	const (
		// state使用
		_statenofollow      = 1
		_statefollow        = 2
		_statefollowed      = 3
		_statemutualConcern = 4
		_specialFollow      = 5
		// 关注关系
		_follow = 1
	)
	r = &Relation{
		Status: _statenofollow,
	}
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	//nolint:gomnd
	switch rel.Attribute {
	case 2, 6: // 用户关注UP主
		r.Status = _statefollow
		r.IsFollow = _follow
	}
	if rel.IsFollowed { // UP主关注用户
		r.Status = _statefollowed
		r.IsFollowed = _follow
	}
	if r.IsFollow == _follow && r.IsFollowed == _follow { // 用户和UP主互相关注
		r.Status = _statemutualConcern
	}
	if rel.Special == 1 {
		r.Status = _specialFollow
	}
	return
}

type Activity struct {
	PageId int64  `json:"page_id,omitempty"`
	H5Link string `json:"h5_link,omitempty"`
}

type HistoryPosition struct {
	Offset int
	Desc   int
	Oid    int
}

type SeasonsRankInfo struct {
	Seasons         map[int64]*ugcSeasonGrpc.Season // 以aid为key的合集信息
	ArcPlayerCursor []*ArcPlayerCursor
	ArcPlayer       []*api.ArcPlayer
}

func FromUpArcToArc(from *uparc.Arc) *api.Arc {
	to := &api.Arc{
		Aid:         from.Aid,
		Videos:      from.Videos,
		TypeID:      from.TypeID,
		TypeName:    from.TypeName,
		Copyright:   from.Copyright,
		Pic:         from.Pic,
		Title:       from.Title,
		PubDate:     from.PubDate,
		Ctime:       from.Ctime,
		Desc:        from.Desc,
		State:       from.State,
		Access:      from.Access,
		Attribute:   from.Attribute,
		Tag:         from.Tag,
		Tags:        from.Tags,
		Duration:    from.Duration,
		MissionID:   from.MissionID,
		OrderID:     from.OrderID,
		RedirectURL: from.RedirectURL,
		Forward:     from.Forward,
		Rights: api.Rights{
			Bp:              from.Rights.Bp,
			Elec:            from.Rights.Elec,
			Download:        from.Rights.Download,
			Movie:           from.Rights.Movie,
			Pay:             from.Rights.Pay,
			HD5:             from.Rights.HD5,
			NoReprint:       from.Rights.NoReprint,
			Autoplay:        from.Rights.Autoplay,
			UGCPay:          from.Rights.UGCPay,
			IsCooperation:   from.Rights.IsCooperation,
			UGCPayPreview:   from.Rights.UGCPayPreview,
			NoBackground:    from.Rights.NoBackground,
			ArcPay:          from.Rights.ArcPay,
			ArcPayFreeWatch: from.Rights.ArcPayFreeWatch,
		},
		Author: api.Author{
			Mid:  from.Author.Mid,
			Name: from.Author.Name,
			Face: from.Author.Face,
		},
		Stat: api.Stat{
			Aid:     from.Stat.Aid,
			View:    from.Stat.View,
			Danmaku: from.Stat.Danmaku,
			Reply:   from.Stat.Reply,
			Fav:     from.Stat.Fav,
			Coin:    from.Stat.Coin,
			Share:   from.Stat.Share,
			NowRank: from.Stat.NowRank,
			HisRank: from.Stat.HisRank,
			Like:    from.Stat.Like,
			DisLike: from.Stat.DisLike,
			Follow:  from.Stat.Follow,
		},
		ReportResult: from.ReportResult,
		Dynamic:      from.Dynamic,
		FirstCid:     from.FirstCid,
		Dimension: api.Dimension{
			Width:  from.Dimension.Width,
			Height: from.Dimension.Height,
			Rotate: from.Dimension.Rotate,
		},
		SeasonID:    from.SeasonID,
		AttributeV2: from.AttributeV2,
	}
	for _, v := range from.StaffInfo {
		if v == nil {
			continue
		}
		to.StaffInfo = append(to.StaffInfo, &api.StaffInfo{
			Mid:       v.Mid,
			Title:     v.Title,
			Attribute: v.Attribute,
		})
	}
	return to
}

func FromVipLabelToVipSpaceLabel(vipLabel account.VipLabel, isHant bool) VipLabelInfo {
	out := VipLabelInfo{}
	out.Path = vipLabel.Path
	out.Text = vipLabel.Text
	out.LabelTheme = vipLabel.LabelTheme
	out.TextColor = vipLabel.TextColor
	out.BgStyle = vipLabel.BgStyle
	out.BgColor = vipLabel.BgColor
	out.BorderColor = vipLabel.BorderColor
	if vipLabel.UseImgLabel {
		out.Image = vipLabel.ImgLabelUriHansStatic
		if isHant {
			out.Image = vipLabel.ImgLabelUriHantStatic
		}
	}
	return out
}

func (m *McnInfo) FromUserExtraValues(u *memberAPI.UserExtraValues) {
	if m == nil {
		return
	}
	if data, ok := u.GetExtraInfo()["video_mcn_info"]; ok {
		if err := json.Unmarshal([]byte(data), m); err == nil {
			return
		}
	}
	if data, ok := u.GetExtraInfo()["live_mcn_info"]; ok {
		if err := json.Unmarshal([]byte(data), m); err == nil {
			return
		}
	}
	return
}

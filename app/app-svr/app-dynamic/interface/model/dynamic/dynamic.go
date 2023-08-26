package dynamic

import (
	"strconv"
	"time"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	pplApi "go-gateway/app/app-svr/app-show/interface/api"
	arcApi "go-gateway/app/app-svr/archive/service/api"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	articleMdl "git.bilibili.co/bapis/bapis-go/article/model"
	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
)

const (
	RefreshTypeUp   = 1
	RefreshTypeDown = 2
	// dynamic type
	DynTypeForward        = 1
	DynTypeDraw           = 2
	DynTypeVideo          = 8
	DynTypeBangumi        = 512
	DynTypePGCBangumi     = 4097
	DynTypePGCMovie       = 4098
	DynTypePGCTv          = 4099
	DynTypePGCGuoChuang   = 4100
	DynTypePGCDocumentary = 4101
	DynTypeCheeseSeason   = 4302
	DynTypeCheeseBatch    = 4303
	// attach topic
	ExtAttachTopicY = 1
	// like type businness
	BusTypeVideo  = "archive"
	BusTypePGC    = "bangumi"
	BusTypeCheese = "cheese"
	// tabRedType
	TabRedTypeCount = "count"
	TabRedTypePoint = "point"
	TabNoPoint      = "no_point"
	// tabAnchor
	AnchorAll   = "all"
	AnchorVideo = "video"

	// SVideoTypeDynamic
	SVideoTypeDynamic = "dynamic"
)

type Dynamic struct {
	Archive []*Archive  `json:"archive,omitempty"`
	Article []*Article  `json:"article,omitempty"`
	PGC     []*PGCShare `json:"pgc,omitempty"`
}

type Archive struct {
	BVID     string `json:"bvid"`
	AID      int64  `json:"aid"`
	Title    string `json:"title"`
	Pic      string `json:"pic"`
	Param    string `json:"param"`
	URI      string `json:"uri"`
	Goto     string `json:"goto"`
	Duration int64  `json:"duration"`
	UpName   string `json:"up_name"`
	View     int32  `json:"view"`
	Danmaku  int32  `json:"danmaku"`
}

type Article struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	TemplateID int32    `json:"template_id"`
	UpName     string   `json:"up_name"`
	ImgURLs    []string `json:"image_urls"`
	ViewNum    int64    `json:"view_num"`
	LikeNum    int64    `json:"like_num"`
	ReplyNum   int64    `json:"reply_num"`
}

type PGCShare struct {
	EpID     int32  `json:"ep_id"`
	Cover    string `json:"cover"`
	Title    string `json:"title"`
	Duration int32  `json:"duration"`
	View     int64  `json:"view"`
	Danmaku  int64  `json:"danmaku"`
	URL      string `json:"url"`
}

func (a *Archive) FormArc(arc *archivegrpc.Arc) {
	if arc == nil {
		return
	}
	a.AID = arc.Aid
	a.Title = arc.Title
	a.Pic = arc.Pic
	a.Param = strconv.FormatInt(a.AID, 10)
	a.Goto = model.GotoAv
	a.URI = model.FillURI(a.Goto, a.Param, model.AvPlayHandlerGRPC(arc, nil, nil))
	a.Duration = arc.Duration
	a.UpName = arc.Author.Name
	a.View = arc.Stat.View
	a.Danmaku = arc.Stat.Danmaku
	if !arc.IsNormal() {
		a.Title = model.InvalidTitle
		a.Pic = ""
		a.Duration = 0
	}
}

func (a *Article) FromArt(art *articleMdl.Meta) {
	if art == nil {
		return
	}
	a.ID = art.ID
	a.Title = art.Title
	a.Summary = art.Summary
	a.TemplateID = art.TemplateID
	if art.Author != nil {
		a.UpName = art.Author.Name
	}
	a.ImgURLs = art.ImageURLs
	if art.Stats != nil {
		a.ViewNum = art.Stats.View
		a.LikeNum = art.Stats.Like
		a.ReplyNum = art.Stats.Reply
	}
}

func (p *PGCShare) FromPgcShare(e *pgcShareGrpc.ShareMessageResBody) {
	if e == nil {
		return
	}
	p.EpID = e.EpId
	p.Cover = e.Cover
	p.Title = e.Title
	// PGC的播放时长是毫秒，需要和UGC统一转成秒
	// nolint:gomnd
	p.Duration = e.Duration / 1000
	p.Danmaku = e.Dm
	p.View = e.View
	p.URL = e.Url
}

// 动态Item容器
type DynContext struct {
	*api.DynamicItem
	DynInfo *Dynamics
	Interim *Interim
	Mid     int64
	// ugc,pgc,付费系列，付费批次，话题信息，账号，装扮卡片，好友点赞
	ResArcs      map[int64]*arcApi.ArcPlayer                            // ugc res: map[aid] info
	ResPGC       map[int64]*PGCInfo                                     // pgc res: map[epid]info
	ResUid       *accountgrpc.CardsReply                                // user res: map[uid] info
	ResTopic     map[int64]*TopicResItems                               // topic res: map[dynamic_id] info
	ResPGCSeason map[int64]*PGCSeason                                   // pgc season res: map[season] info
	ResPGCBatch  map[int64]*PGCBatch                                    // pgc batch res: map[batch_id] info
	ResDecorate  map[int64]*DecoCards                                   // DecoCards res: map[uid] info
	ResLikeIcon  map[int64]*LikeIconItems                               // like icon res
	ResEmoji     map[string]*EmojiItem                                  // emoji Res
	ResUserLike  map[string]*thumgrpc.ItemHasLikeRecentReply_MapRecords // user like Res
	ResThum      map[int64]*ThumbupItem                                 // 好友点赞信息（粉丝数、关注类型）
	ResThumStats *thumgrpc.MultiStatsReply                              // 用户点赞状态
	ResBottom    map[int64]*BottomItem
	// 需要第三次请求的资源信息
	Emoji map[string]struct{}
}

func (dyn *Dynamics) IsAv() bool {
	return dyn.Type == DynTypeVideo
}

func (dyn *Dynamics) IsPGC() bool {
	is := dyn.Type == DynTypePGCBangumi || dyn.Type == DynTypePGCMovie || dyn.Type == DynTypePGCTv ||
		dyn.Type == DynTypePGCGuoChuang || dyn.Type == DynTypePGCDocumentary || dyn.Type == DynTypeBangumi
	return is
}

func (dyn *Dynamics) IsCurr() bool {
	return dyn.Type == DynTypeCheeseSeason || dyn.Type == DynTypeCheeseBatch
}

func (dyn *Dynamics) IsCurrBatch() bool {
	return dyn.Type == DynTypeCheeseBatch
}

func (dyn *Dynamics) IsCurrSeason() bool {
	return dyn.Type == DynTypeCheeseSeason
}

func (dyn *Dynamics) GetPGCSubType() api.VideoSubType {
	switch dyn.Type {
	case DynTypeBangumi:
		return api.VideoSubType_VideoSubTypeBangumi
	case DynTypePGCBangumi:
		return api.VideoSubType_VideoSubTypeBangumi
	case DynTypePGCMovie:
		return api.VideoSubType_VideoSubTypeMovie
	case DynTypePGCTv:
		return api.VideoSubType_VideoSubTypeTeleplay
	case DynTypePGCGuoChuang:
		return api.VideoSubType_VideoSubTypeDomestic
	case DynTypePGCDocumentary:
		return api.VideoSubType_VideoSubTypeDocumentary
	default:
		return api.VideoSubType_VideoSubTypeNone
	}
}

type Interim struct {
	Face  string
	UName string
}

type DynBase struct {
	DynamicID int64
	DynType   int
	Rid       int64
	Uid       int64
	Acl       int
}

// 动态列表资源
type DynVideoListRes struct {
	UpdateNum      int         `json:"update_num"`
	HistoryOffset  string      `json:"history_offset"`
	UpdateBaseline string      `json:"update_baseline"`
	HasMore        int         `json:"has_more"`
	Dynamics       []*Dynamics `json:"dynamics"`
}
type Lott struct {
	LotteryID int    `json:"lottery_id"`
	Title     string `json:"title"`
}
type Vote struct {
	VoteID int `json:"vote_id"`
}
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
type Lbs struct {
	Address      string   `json:"address"`
	Distance     int      `json:"distance"`
	Location     Location `json:"location"`
	Poi          string   `json:"poi"`
	ShowTitle    string   `json:"show_title"`
	Title        string   `json:"title"`
	Type         int      `json:"type"`
	ShowDistance string   `json:"show_distance"`
}
type Ctrl struct {
	Length   int    `json:"length"`
	Location int    `json:"location"`
	Type     int    `json:"type"`
	Data     string `json:"data"`
	TypeID   string `json:"type_id"`
}
type TopicInfo struct {
	IsAttachTopic int `json:"is_attach_topic"`
}
type Dispute struct {
	Content string `json:"content"`
	Desc    string `json:"description"`
	Url     string `json:"jump_url"`
}
type Extend struct {
	Lott      *Lott     `json:"lott"`
	Vote      Vote      `json:"vote"`
	Lbs       *Lbs      `json:"lbs"`
	Ctrl      []*Ctrl   `json:"ctrl"`
	TopicInfo TopicInfo `json:"topic_info"`
	EmojiType int       `json:"emoji_type"`
	Dispute   *Dispute  `json:"dispute"`
	Bottom    *Bottom   `json:"bottom"`
}
type Dynamics struct {
	DynamicID int64       `json:"dynamic_id"`
	Type      int64       `json:"type"`
	Rid       int64       `json:"rid"`
	UID       int64       `json:"uid"`
	UIDType   int         `json:"uid_type"`
	Repost    int         `json:"repost"`
	ACL       Acl         `json:"acl"`
	Extend    Extend      `json:"extend"`
	Tips      string      `json:"tips"`
	Invisible int         `json:"invisible"`
	Timestamp int64       `json:"timestamp"`
	Origin    OrigDynamic `json:"Origin"`
	Display   Display     `json:"display"`
}
type OrigDynamic struct {
	DynamicID int64  `json:"dynamic_id"`
	Type      int    `json:"type"`
	Rid       int64  `json:"rid"`
	UID       int    `json:"uid"`
	UIDType   int    `json:"uid_type"`
	Repost    int    `json:"repost"`
	ACL       int    `json:"acl"`
	Extend    Extend `json:"extend"`
	Tips      string `json:"tips"`
	Invisible int    `json:"invisible"`
	Timestamp int64  `json:"timestamp"`
}
type Acl struct {
	RepostBan  int64 `json:"repost_banned"`
	CommentBan int64 `json:"comment_banned"`
	FoldLimit  int64 `json:"limit_display"`
}
type Bottom struct {
	Rid  int64 `json:"rid"`
	Type int   `json:"type"`
}

func (c *Ctrl) TranType() string {
	switch c.Type {
	case CtrlTypeLottery:
		return DescTypeLottery
	case CtrlTypeVote:
		return DescTypeVote
	}
	return DescTypeText
}

type CtrlSort []*Ctrl

func (t CtrlSort) Len() int {
	return len(t)
}

func (t CtrlSort) Less(i, j int) bool {
	return t[i].Location < t[j].Location
}

func (t CtrlSort) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type Display struct {
	LikeUsers []int64 `json:"like_users"`
}

// 动态ID获取动态详情
type DynDetailRsp struct {
	Dynamics []*Dynamics `json:"dynamics"`
}

// PGC资源
type PGCInfo struct {
	Aid        int64       `json:"aid"`
	Cid        int64       `json:"cid"`
	Cover      string      `json:"cover"`
	Dimension  Dimension   `json:"dimension"`
	Duration   int64       `json:"duration"`
	EpisodeID  int64       `json:"episode_id"`
	IndexTitle string      `json:"index_title"`
	IsFinish   int         `json:"is_finish"`
	IsPreview  int         `json:"is_preview"`
	NewDesc    string      `json:"new_desc"`
	PlayerInfo *PlayerInfo `json:"player_info"`
	Season     *Season     `json:"season"`
	ShortTitle string      `json:"short_title"`
	Stat       Stat        `json:"stat"`
	URL        string      `json:"url"`
}
type Dimension struct {
	Height int64 `json:"height"`
	Rotate int64 `json:"rotate"`
	Width  int64 `json:"width"`
}
type FileInfoItem struct {
	Ahead      string `json:"ahead"`
	Filesize   int    `json:"filesize"`
	Order      int    `json:"order"`
	Timelength int    `json:"timelength"`
	Vhead      string `json:"vhead"`
}
type FileInfo struct {
	Infos []*FileInfoItem `json:"infos"`
}
type PlayerInfo struct {
	Cid                int64               `json:"cid"`
	ExpireTime         int                 `json:"expire_time"`
	FileInfo           map[int64]*FileInfo `json:"file_info"`
	Fnval              int                 `json:"fnval"`
	Fnver              int                 `json:"fnver"`
	Quality            int                 `json:"quality"`
	SupportDescription []string            `json:"support_description"`
	SupportFormats     []string            `json:"support_formats"`
	SupportQuality     []int               `json:"support_quality"`
	URL                string              `json:"url"`
	VideoCodecid       int                 `json:"video_codecid"`
	VideoProject       bool                `json:"video_project"`
}
type Season struct {
	Cover       string `json:"cover"`
	IsFinish    int    `json:"is_finish"`
	SeasonID    int64  `json:"season_id"`
	SquareCover string `json:"square_cover"`
	Title       string `json:"title"`
	TotalCount  int    `json:"total_count"`
	Ts          int    `json:"ts"`
	Type        int    `json:"type"`
	TypeName    string `json:"type_name"`
}
type Stat struct {
	Danmaku int `json:"danmaku"`
	Play    int `json:"play"`
	Reply   int `json:"reply"`
}

// 话题资源
type TopicRes struct {
	Items []TopicResItems `json:"items"`
}
type FromContent struct {
	TopicID      int    `json:"topic_id"`
	TopicName    string `json:"topic_name"`
	IsActivity   int    `json:"is_activity"`
	TopicLink    string `json:"topic_link"`
	IsNewChannel int    `json:"is_new_channel"`
}
type TopicResItems struct {
	DynamicID     int64          `json:"dynamic_id"`
	FromContent   []*FromContent `json:"from_content"`
	TopicActivity *TopicActivity `json:"activity"`
}
type TopicActivity struct {
	TopicName string `json:"topic_name"`
	TopicLink string `json:"topic_link"`
}

// 付费系列资源
type PGCSeason struct {
	Badge       SeasonBadge  `json:"badge"`
	Cover       string       `json:"cover"`
	EpCount     int          `json:"ep_count"`
	ID          int          `json:"id"`
	Subtitle    string       `json:"subtitle"`
	Title       string       `json:"title"`
	UpID        int64        `json:"up_id"`
	UpInfo      SeasonUpInfo `json:"up_info"`
	UpdateCount int          `json:"update_count"`
	UpdateInfo  string       `json:"update_info"`
	URL         string       `json:"url"`
}
type SeasonBadge struct {
	BgColor       string `json:"bg_color"`
	BgDarkColor   string `json:"bg_dark_color"`
	Text          string `json:"text"`
	TextColor     string `json:"text_color"`
	TextDarkColor string `json:"text_dark_color"`
}
type SeasonUpInfo struct {
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

// 付费更新批次资源
type PGCBatch struct {
	Badge       BatchBadge  `json:"badge"`
	Cover       string      `json:"cover"`
	EpCount     int         `json:"ep_count"`
	ID          int         `json:"id"`
	Subtitle    string      `json:"subtitle"`
	Title       string      `json:"title"`
	UpID        int64       `json:"up_id"`
	UpInfo      BatchUpInfo `json:"up_info"`
	UpdateCount int         `json:"update_count"`
	URL         string      `json:"url"`
	NewEp       NewEp       `json:"new_ep"`
	UserProfile UserProfile `json:"user_profile"`
}
type BatchBadge struct {
	BgColor       string `json:"bg_color"`
	BgDarkColor   string `json:"bg_dark_color"`
	Text          string `json:"text"`
	TextColor     string `json:"text_color"`
	TextDarkColor string `json:"text_dark_color"`
}
type BatchUpInfo struct {
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}
type NewEp struct {
	Cover string `json:"cover"`
	ID    int    `json:"id"`
	Reply int    `json:"reply"`
	Title string `json:"title"`
}
type OfficialVerify struct {
	Desc string `json:"desc"`
	Type int    `json:"type"`
}
type Card struct {
	OfficialVerify OfficialVerify `json:"official_verify"`
}
type Info struct {
	Face  string `json:"face"`
	UID   int    `json:"uid"`
	Uname string `json:"uname"`
}
type Pendant struct {
	Expire int64  `json:"expire"`
	Image  string `json:"image"`
	Name   string `json:"name"`
	Pid    int64  `json:"pid"`
}
type Label struct {
	Path string `json:"path"`
}
type BatchVip struct {
	AccessStatus  int    `json:"accessStatus"`
	DueRemark     string `json:"dueRemark"`
	Label         Label  `json:"label"`
	ThemeType     int    `json:"themeType"`
	VipDueDate    int64  `json:"vipDueDate"`
	VipStatus     int    `json:"vipStatus"`
	VipStatusWarn string `json:"vipStatusWarn"`
	VipType       int    `json:"vipType"`
}
type UserProfile struct {
	Card    Card     `json:"card"`
	Info    Info     `json:"info"`
	Pendant Pendant  `json:"pendant"`
	Rank    string   `json:"rank"`
	Sign    string   `json:"sign"`
	Vip     BatchVip `json:"vip"`
}

// 装扮卡片资源
type DecoCards struct {
	ID           int64       `json:"id"`
	ItemID       int         `json:"item_id"`
	ItemType     int         `json:"item_type"`
	Name         string      `json:"name"`
	CardURL      string      `json:"card_url"`
	BigCardURL   string      `json:"big_card_url"`
	CardType     int         `json:"card_type"`
	ExpireTime   int         `json:"expire_time"`
	CardTypeName string      `json:"card_type_name"`
	JumpURL      string      `json:"jump_url"`
	Fan          DecorateFan `json:"fan"`
}
type DecorateFan struct {
	IsFan   int    `json:"is_fan"`
	Number  int    `json:"number"`
	Color   string `json:"color"`
	Name    string `json:"name"`
	NumDesc string `json:"num_desc"`
}

// 点赞图标信息
type LikeIcon struct {
	Items []*LikeIconItems `json:"items"`
}
type LikeIconItems struct {
	DynamicID  int64  `json:"dynamic_id"`
	OldIconID  int64  `json:"old_icon_id"`
	StartIcon  string `json:"start_icon"`
	ActionIcon string `json:"action_icon"`
	EndIcon    string `json:"end_icon"`
	NewIconID  int64  `json:"new_icon_id"`
	StartURL   string `json:"start_url"`
	ActionURL  string `json:"action_url"`
	EndURL     string `json:"end_url"`
}

// emoji表情信息
type Emoji struct {
	Emote map[string]*EmojiItem `json:"emote"`
}
type EmojiMeta struct {
	Size            int    `json:"size"`
	LabelText       string `json:"label_text"`
	LabelURL        string `json:"label_url"`
	LabelColor      string `json:"label_color"`
	LabelGuideTitle string `json:"label_guide_title"`
	LabelGuideText  string `json:"label_guide_text"`
}
type EmojiItem struct {
	ID        int       `json:"id"`
	PackageID int       `json:"package_id"`
	State     int       `json:"state"`
	Type      int       `json:"type"`
	Attr      int       `json:"attr"`
	Text      string    `json:"text"`
	URL       string    `json:"url"`
	Meta      EmojiMeta `json:"meta"`
	Mtime     int       `json:"mtime"`
}

// 点赞用户
type ThumbupItem struct {
	Mid     int64
	UName   string
	Fans    int64
	Special int
}
type ThumSort []*ThumbupItem

func (t ThumSort) Len() int {
	return len(t)
}
func (t ThumSort) Less(i, j int) bool {
	if (t[i].Special == 1) == (t[j].Special == 1) {
		return t[i].Fans > t[j].Fans
	}
	if t[i].Special == 1 {
		return true
	}
	return false
}
func (t ThumSort) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type LottURIParam struct {
	Uid        int64  `json:"uid"`
	Face       string `json:"face"`
	Name       string `json:"name"`
	CreateTime int64  `json:"create_time"`
	Content    string `json:"content"`
}

// 最近访问-个人feed流资源
type VideoPersonalRes struct {
	HasMore  int         `json:"has_more"`
	Offset   string      `json:"offset"`
	Dynamics []*Dynamics `json:"dynamics"`
}

// 最近访问-up主列表
type VdUpListRsp struct {
	Items       []UpListItem `json:"items"`
	ModuleTitle string       `json:"module_title"`
	ShowAll     string       `json:"show_all"`
}

type UpListItem struct {
	HasUpdate int   `json:"has_update"`
	UID       int64 `json:"uid"`
}

// 游戏小卡资源
type BottomRes struct {
	Items []*BottomItem `json:"items"`
}

type BottomInfo struct {
	Content string `json:"content"`
	JumpURL string `json:"jump_url"`
}

type BottomItem struct {
	DynamicID  int64       `json:"dynamic_id"`
	BottomInfo *BottomInfo `json:"bottom_info"`
}

// 折叠变量
type FoldList struct {
	List []*FoldItem
}

type FoldItem struct {
	*api.DynamicItem
	Rid       int64
	Uid       int64
	Acl       Acl
	Timestamp int64
	Type      int64
	FoldList  []string
	FoldType  api.FoldType
}

type FoldMapItem struct {
	DynItem    *FoldItem
	FoldType   int
	OrigDyn    *FoldItem
	InsertBase string
}

func (item *FoldItem) IsAv() bool {
	return item.Type == DynTypeVideo
}

func (item *FoldItem) CanFoldFrequent() bool {
	return item.IsAv()
}

func (item *FoldItem) GetLimitTime(conf *conf.Config) int64 {
	if item.Type == DynTypeForward {
		return int64(time.Duration(conf.Resource.FoldPublishForward) / time.Second)
	}
	return int64(time.Duration(conf.Resource.FoldPublishOther) / time.Second)
}

// DynSVideoList 动态列表资源
type DynSVideoList struct {
	MixVideoItem []*SVideoItem `json:"items"`
	HasMore      int32         `json:"has_more"`
	Offset       string        `json:"offset"`
	isForceAdd   bool          // 是否需要强行插入：focus_aid不在第一刷
}

// SVideoItem 联播小视频
type SVideoItem struct {
	RID   int64 `json:"rid"`
	UID   int64 `json:"uid"`
	DynID int64 `json:"dynamic_id"`
	Index int64 `json:"index"` // 游标
}

// SVideoMaterial 小视频素材
type SVideoMaterial struct {
	*api.SVideoItem
	Arc      *archivegrpc.ArcPlayer `json:"arc"`
	IsAtten  int32                  `json:"is_atten"`
	DynIdStr string                 `json:"dyn_id_str"`
	IsLike   int32                  `json:"is_like"`
}

// SVideoInfoc is
type SVideoInfoc struct {
	AID       int64
	UpID      int64
	Buvid     string
	MID       int64
	FromSpmid string
	// 当前是否关注 1是 0否
	Follow int32
	// 当前是否点赞 1是 0否
	Like int32
	// 卡片类型 视频卡-av
	CardType string
	// 卡片在一刷中出现的位置 从0开始
	CardIndex int32
	// 分页游标
	Offset string
	// 资源类型,如dynamic
	OType string
	// 资源内容id,如动态id
	OID int64
}

type DynTabResult struct {
	Switch     int32  `json:"switch"`
	NeedAsk    int32  `json:"need_ask"`
	RedPoint   int32  `json:"red_point"`
	BubbleDesc string `json:"bubble_desc"`
	TabName    string `json:"tab_name"`
	CityName   string `json:"city_name"`
	CityID     int64  `json:"city_id"`
}

type DynCityResult struct {
	NoticeType int               `json:"notice_type"`
	Dynamics   []*DynCityDynamic `json:"dynamics"`
	HasMore    int32             `json:"has_more"`
	Offset     string            `json:"offset"`
	UserGroup  int               `json:"user_group"`
}

type DynCityDynamic struct {
	DynID      int64   `json:"dynamic_id"`
	Type       int64   `json:"type"`
	RID        int64   `json:"rid"`
	UID        int64   `json:"uid"`
	CornerMark string  `json:"corner_mark"`
	Cover      string  `json:"cover_url"`
	Extend     *Extend `json:"extend"`
}

type DrawDetailRes struct {
	Item DrawItem `json:"item"`
	User *User    `json:"user"`
}

type DrawItem struct {
	ID            int64         `json:"id"`
	Pictures      []DrawPicture `json:"pictures"`
	PicturesCount int           `json:"pictures_count"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Reply         int           `json:"reply"`
	UploadTime    int64         `json:"upload_time"`
	AtControl     string        `json:"at_control"`
}

type DrawPicture struct {
	ImgSrc    string       `json:"img_src"`
	ImgHeight int64        `json:"img_height"`
	ImgWidth  int64        `json:"img_width"`
	ImgSize   float32      `json:"img_size"`
	ImgTags   []DrawImgTag `json:"img_tags"`
}

type DrawImgTag struct {
	Text        string `json:"text"`
	Type        int32  `json:"type"`
	Url         string `json:"url"`
	X           int64  `json:"x"`
	Y           int64  `json:"y"`
	Orientation int32  `json:"orientation"`
	SchemaURL   string `json:"schema_url"`
	ItemID      int64  `json:"item_id"`
	Source      int32  `json:"source_type"`
	Mid         int64  `json:"mid"`
	Tid         int64  `json:"tid"`
	Poi         string `json:"poi"`
}

type User struct {
	UID     int64  `json:"uid"`
	HeadURL string `json:"head_url"`
	Name    string `json:"name"`
}

func (t *DynSVideoList) FromPplIdx(val *pplApi.IndexSVideoReply, idx, focusAid int64) {
	t.HasMore = val.HasMore
	t.Offset = val.Offset
	if len(val.List) == 0 {
		return
	}
	if idx == 0 && focusAid != 0 {
		// 默认第一刷要强插
		t.isForceAdd = true
	}
	var lst []*SVideoItem
	for _, l := range val.List {
		sv := &SVideoItem{
			RID:   l.Rid,
			UID:   l.Uid,
			Index: l.Index,
		}
		// 锚点aid判断
		if l.Rid == focusAid {
			// 第一刷做置顶
			if idx == 0 && focusAid != 0 {
				t.isForceAdd = false // 第一刷命中不做强插
				lst = append(append([]*SVideoItem{}, sv), lst...)
			}
			// 非第一刷做排重
			continue
		}
		lst = append(lst, sv)
	}
	// 第一刷未命中 强插置顶
	if t.isForceAdd {
		lst = append(append([]*SVideoItem{}, &SVideoItem{RID: focusAid}), lst...)
	}
	t.MixVideoItem = lst
}

func (t *DynSVideoList) FromPplAggr(val *pplApi.AggrSVideoReply, idx, focusAid int64) {
	t.HasMore = val.HasMore
	t.Offset = val.Offset
	if len(val.List) == 0 {
		return
	}
	if idx == 0 && focusAid != 0 {
		// 默认第一刷要强插
		t.isForceAdd = true
	}
	var lst []*SVideoItem
	for _, l := range val.List {
		sv := &SVideoItem{
			RID:   l.Rid,
			UID:   l.Uid,
			Index: l.Index,
		}
		// 锚点aid判断
		if l.Rid == focusAid {
			// 第一刷做置顶
			if idx == 0 && focusAid != 0 {
				t.isForceAdd = false // 第一刷命中不做强插
				lst = append(append([]*SVideoItem{}, sv), lst...)
			}
			// 非第一刷做排重
			continue
		}
		lst = append(lst, sv)
	}
	// 第一刷未命中 强插置顶
	if t.isForceAdd {
		lst = append(append([]*SVideoItem{}, &SVideoItem{RID: focusAid}), lst...)
	}
	t.MixVideoItem = lst
}

func GetArcUpids(dynList *DynSVideoList, arcs map[int64]*archivegrpc.ArcPlayer) (upids []int64) {
	upMap := make(map[int64]int64)
	for _, arc := range arcs {
		if arc != nil && arc.Arc != nil {
			upids = append(upids, arc.Arc.Author.Mid)
			upMap[arc.Arc.Aid] = arc.Arc.Author.Mid
		}
	}
	for i := 0; i < len(dynList.MixVideoItem); i++ {
		if upid, ok := upMap[dynList.MixVideoItem[i].RID]; ok {
			dynList.MixVideoItem[i].UID = upid
		}
	}
	return
}

func ToTop(val *pplApi.SVideoTop) *api.SVideoTop {
	return &api.SVideoTop{
		Title: val.Title,
		Desc:  val.Desc,
	}
}

type GeoCoderReq struct {
	//纬度
	Lat float64 `form:"lat"`
	//经度
	Lng float64 `form:"lng"`
	//页面来源
	From string `form:"from"`
}

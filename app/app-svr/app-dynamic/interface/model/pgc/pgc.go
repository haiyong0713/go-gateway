package pgc

// DynPGCDetail 动态请求pgc详情
type DynPGCItem struct {
	Aid           int64       `json:"aid"`
	Cid           int64       `json:"cid"`
	Cover         string      `json:"cover"`
	Dimension     Dimension   `json:"dimension"`
	Duration      int64       `json:"duration"`
	EpisodeID     int64       `json:"episode_id"`
	IndexTitle    string      `json:"index_title"`
	IsFinish      int         `json:"is_finish"`
	IsPreview     int         `json:"is_preview"`
	NewDesc       string      `json:"new_desc"`
	PlayerInfo    *PlayerInfo `json:"player_info"`
	Season        *Season     `json:"season"`
	ShortTitle    string      `json:"short_title"`
	Stat          Stat        `json:"stat"`
	URL           string      `json:"url"`
	Tags          []*Tag      `json:"tags"`
	CardShowTitle string      `json:"card_show_title"`
	SectionType   int         `json:"section_type"`
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
	Danmaku    int    `json:"danmaku"`
	Play       int    `json:"play"`
	Reply      int    `json:"reply"`
	Follow     int    `json:"follow"`
	FollowDesc string `json:"follow_desc"`
}

type Tag struct {
	Name   string  `json:"name"`
	Icon   string  `json:"icon"`
	Link   string  `json:"link"`
	Report *Report `json:"report"`
}

type Report struct {
	SubModule string `json:"sub_module"`
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
	Stat        BatchStat   `json:"stat"`
	InlineVideo InlineVideo `json:"inline_video"`
	SeasonID    int64       `json:"season_id"`
	UpdateInfo  string      `json:"update_info"`
}

type BatchBadge struct {
	BgColor       string `json:"bg_color"`
	BgDarkColor   string `json:"bg_dark_color"`
	Text          string `json:"text"`
	TextColor     string `json:"text_color"`
	TextDarkColor string `json:"text_dark_color"`
}

type BatchStat struct {
	Duration  int64 `json:"duration"`
	PlayCount int64 `json:"play_count"`
	DmCount   int64 `json:"dm_count"`
	Reply     int64 `json:"reply"`
}

type InlineVideo struct {
	ExpireTime         int64    `json:"expire_time"`
	Cid                int64    `json:"cid"`
	SupportQuality     []int64  `json:"support_quality"`
	SupportFormats     []string `json:"support_formats"`
	SupportDescription []string `json:"support_description"`
	Quality            int64    `json:"quality"`
	Url                string   `json:"url"`
	Aid                int64    `json:"aid"`
	Epid               int64    `json:"ep_id"`
	Duration           int64    `json:"duration"`
	IsPreview          bool     `json:"preview"`
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

type UserProfile struct {
	Card    Card     `json:"card"`
	Info    Info     `json:"info"`
	Pendant Pendant  `json:"pendant"`
	Rank    string   `json:"rank"`
	Sign    string   `json:"sign"`
	Vip     BatchVip `json:"vip"`
}

type Card struct {
	OfficialVerify OfficialVerify `json:"official_verify"`
}

type OfficialVerify struct {
	Desc string `json:"desc"`
	Type int    `json:"type"`
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

type BatchVip struct {
	AccessStatus    int    `json:"accessStatus"`
	DueRemark       string `json:"dueRemark"`
	Label           Label  `json:"label"`
	ThemeType       int    `json:"themeType"`
	VipDueDate      int64  `json:"vipDueDate"`
	VipStatus       int    `json:"vipStatus"`
	VipStatusWarn   string `json:"vipStatusWarn"`
	VipType         int    `json:"vipType"`
	AvatarSubscript int32  `json:"avatar_subscript"`
	NicknameColor   string `json:"nickname_color"`
}

type Label struct {
	Path       string `json:"path"`
	Text       string `json:"text"`
	LabelTheme string `json:"label_theme"`
}

// 付费系列资源
type PGCSeason struct {
	Badge                 SeasonBadge  `json:"badge"`
	Cover                 string       `json:"cover"`
	EpCount               int          `json:"ep_count"`
	ID                    int          `json:"id"`
	Subtitle              string       `json:"subtitle"`
	Title                 string       `json:"title"`
	UpID                  int64        `json:"up_id"`
	UpInfo                SeasonUpInfo `json:"up_info"`
	UpdateCount           int          `json:"update_count"`
	UpdateInfo            string       `json:"update_info"`
	URL                   string       `json:"url"`
	InlineVideo           InlineVideo  `json:"inline_video"`
	NewEp                 NewEp        `json:"new_ep"`
	UserProfile           UserProfile  `json:"user_profile"`
	Stat                  BatchStat    `json:"stat"`
	DynamicShareContent   string       `json:"dynamic_share_content"`
	DynamicReserveContent string       `json:"dynamic_reserve_content"`
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

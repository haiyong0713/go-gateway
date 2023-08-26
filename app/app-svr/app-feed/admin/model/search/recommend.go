package search

import (
	xtime "go-common/library/time"

	model "go-gateway/app/app-svr/app-feed/admin/model/card"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"

	inlinegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	//PlatFromApp .
	PlatFromApp = 1
)

const (
	// CardVideo 视频卡
	CardVideo = 4
	//CardSpecial special big card
	CardSpecial = 3
	//CardSpecialSmall special card small
	CardSpecialSmall = 6
	//CardUnion card content union
	CardUnion = 7
	//CardGame card game
	CardGame = 8
	//CardPGC .
	CardPGC = 9
	//CardPGC fan
	CardPGCFan = 12
	//CardPGC move.
	CardPGCMove = 13
	//CardInlineUgc card inline ugc
	CardInlineUgc = 16
	//CardInlineLive card inline live
	CardInlineLive = 17
	//CardInlineOgv card inline ogv
	CardInlineOgv = 18
	// CardPGCEp
	CardPgcEp = 19
	// CardNavigation baike navigation card
	CardNavigation = 20
	// CardGameBig game card
	CardGameBig = 21

	// MainSearch 综搜
	MainSearch = "main_search"
	// MediaFt 影视垂搜
	MediaFt = "media_ft"
	// MediaBangumi 番剧垂搜
	MediaBangumi = "media_bangumi"
	// UpUser 用户垂搜
	UpUser = "up_user"
	// LiveAll 直播垂搜
	LiveAll = "live_all"
	// Article 专栏垂搜
	Article = "article"
)

var (
	SpecialGroupArray = []string{
		MediaFt,
		MediaBangumi,
		UpUser,
		LiveAll,
		Article,
	}
)

// SpreadConfig .
type SpreadConfig struct {
	ID           int64                     `gorm:"column:id" json:"id"`
	CardType     int64                     `gorm:"column:card_type" json:"card_type"`
	StartTime    xtime.Time                `gorm:"column:start_time" json:"start_time"`
	EndTime      xtime.Time                `gorm:"column:end_time" json:"end_time"`
	OperatorId   int64                     `gorm:"column:operator_id" json:"operator_id"`
	OperatorName string                    `gorm:"column:operator_name" json:"operator_name"`
	PlatVerStr   string                    `gorm:"column:plat_ver" json:"-"`
	ArticleId    int64                     `gorm:"column:article_id" json:"article_id"`
	Title        string                    `gorm:"column:title" json:"title"`
	ImgUrl       string                    `gorm:"column:img_url" json:"img_url"`
	RedirectUrl  string                    `gorm:"column:redirect_url" json:"redirect_url"`
	ValidStatus  int64                     `gorm:"column:valid_status" json:"valid_status"`
	DelStatus    int64                     `gorm:"column:del_status" json:"del_status"`
	ApplyReason  string                    `gorm:"column:apply_reason" json:"apply_reason"`
	Position     int64                     `gorm:"column:position" json:"position"`
	Check        int64                     `gorm:"column:check" json:"check"`
	Card         int64                     `gorm:"column:card" json:"card"`
	RecReason    string                    `gorm:"column:rec_reason" json:"rec_reason"`
	Platform     int64                     `gorm:"column:platform" json:"platform"`
	ExtraStr     string                    `gorm:"column:extra" json:"-"`
	SearchGroup  string                    `gorm:"column:search_group" json:"search_group" form:"search_group"`
	PlatVer      []*PlatVer                `json:"plat_ver"`
	Query        []*SpreadQuery            `json:"query"`
	Extra        *SpreadConfigExtra        `json:"extra_info"`
	PgcSeason    *seasongrpc.CardInfoProto `json:"pgc_card"`
	Special      *manager.SpecialCard      `json:"special_card"`
	Union        *manager.ContentCard      `json:"union_card"`
	Navigation   *NavigationCard           `json:"navigation_card"`
	Ogv          *inlinegrpc.EpisodeCard   `json:"ogv_card"`
}

// SpreadConfigExtra .
type SpreadConfigExtra struct {
	Title   string `json:"title"`
	ImgUrl  string `json:"img_url"`
	ReType  int32  `json:"re_type"`
	ReValue string `json:"re_value"`
	Wiki    *Wiki  `json:"wiki"`
	CardId  int64  `json:"card_id"`
}

type Wiki struct {
	CornerType     int32  `json:"corner_type"`
	CornerText     string `json:"corner_text"`
	CornerSunUrl   string `json:"corner_sun_url"`
	CornerNightUrl string `json:"corner_night_url"`
	CornerHeight   int32  `json:"corner_height"`
	CornerWidth    int32  `json:"corner_width"`
}

// RecomParam .
type RecomParam struct {
	Ts          int64    `form:"ts"` // 时间戳
	StartTs     int64    `form:"start_ts"`
	EndTs       int64    `form:"end_ts"`
	Ps          int      `form:"ps" default:"20"`    // 分页大小
	Pn          int      `form:"pn" default:"1"`     // 第几个分页
	Plat        int      `form:"plat"`               // Plat
	Pos         int      `form:"pos"`                // Pos
	CardType    []int    `form:"card_type,split"`    // 卡片类型
	SearchGroup []string `form:"search_group,split"` // 垂搜类型
}

// RecomRes .
type RecomRes struct {
	Item []*SpreadConfig `json:"spread_config"`
	Page common.Page     `json:"page"`
}

// PlatVer .
type PlatVer struct {
	Plat       string `json:"plat"`
	Conditions string `json:"conditions"`
	Build      string `json:"build"`
}

// SpreadQuery .
type SpreadQuery struct {
	ID        int64  `gorm:"column:id" json:"id"`
	SpreadId  int64  `gorm:"column:spread_id" json:"spread_id"`
	QueryName string `gorm:"column:query_name" json:"query_name"`
	DelStatus int64  `gorm:"column:del_status" json:"del_status"`
}

// NavigationCard provided to AI by OpenRecommend Api
type NavigationCard struct {
	CardId         int64             `json:"card_id"`
	Title          string            `json:"title"`
	Desc           string            `json:"desc"`
	CoverType      int32             `json:"cover_type"`
	CoverSunUrl    string            `json:"cover_sun_url"`
	CoverNightUrl  string            `json:"cover_night_url"`
	CoverWidth     int32             `json:"cover_width"`
	CoverHeight    int32             `json:"cover_height"`
	CornerType     int32             `json:"corner_type"`
	CornerText     string            `json:"corner_text"`
	CornerSunUrl   string            `json:"corner_sun_url"`
	CornerNightUrl string            `json:"corner_night_url"`
	CornerWidth    int32             `json:"corner_width"`
	CornerHeight   int32             `json:"corner_height"`
	ButtonType     int32             `json:"btn_type"`
	ButtonText     string            `json:"btn_text"`
	ButtonReType   int32             `json:"btn_re_type"`
	ButtonReValue  string            `json:"btn_re_value"`
	Navigation     *model.Navigation `json:"navigation"`
}

// TableName SearchSpreadConfig
func (a SpreadConfig) TableName() string {
	return "search_spread_config"
}

// TableName SearchSpreadConfig
func (a SpreadQuery) TableName() string {
	return "search_spread_query_associated"
}

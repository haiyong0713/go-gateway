package model

var (
	// ArticleSortType article list sort types.
	ArticleSortType = map[string]int{
		"publish_time": 0,
		"view":         5,
		"fav":          4,
	}
	// PrivacyFields privacy allowed field.
	PrivacyFields = []string{
		PcyBangumi,
		PcyTag,
		PcyFavVideo,
		PcyCoinVideo,
		PcyGroup,
		PcyGame,
		PcyChannel,
		PcyUserInfo,
		PcyLikeVideo,
		PcyBbq,
		PcyComic,
		PcyDressUp,
		LivePlayback,
	}
	// OuterPrivacyFields privacy out of fmtPrivacy.
	OuterPrivacyFields = []string{
		PcyDisableFollowing,
		PcyCloseSpaceMedal,
		PcyOnlyShowWearing,
		PcyDisableShowSchool,
		PcyDisableShowNft,
	}
	AllPrivacyFields = append(PrivacyFields, OuterPrivacyFields...)
	//ArcCheckType search arc check type.
	ArcCheckType = map[string]int{
		"channel": 1,
	}
)

// Page page return data struct.
type Page struct {
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
	Total int `json:"total"`
}

// SearchArg arc search param.
type SearchArg struct {
	Mid       int64  `form:"mid" validate:"gt=0"`
	Tid       int64  `form:"tid"`
	Order     string `form:"order"`
	Keyword   string `form:"keyword"`
	Pn        int    `form:"pn" default:"1" validate:"gt=0"`
	Ps        int    `form:"ps" default:"30" validate:"gt=0,lte=50"`
	CheckType string `form:"check_type"`
	CheckID   int64  `form:"check_id"`
	Index     int    `form:"index"`
	Token     string `form:"token"`
}

// WebIndex .
type WebIndex struct {
	Account *AccInfo `json:"account"`
	Setting *Setting `json:"setting"`
	Archive *WebArc  `json:"archive"`
}

// WebArc .
type WebArc struct {
	Page     WebPage    `json:"page"`
	Archives []*ArcItem `json:"archives"`
}

// WebPage .
type WebPage struct {
	Pn    int32 `json:"pn"`
	Ps    int32 `json:"ps"`
	Count int64 `json:"count"`
}

// DynamicSearchArg .
type DynamicSearchArg struct {
	Mid     int64  `form:"mid" validate:"min=1"`
	Keyword string `form:"keyword" validate:"required"`
	Pn      int    `form:"pn" validate:"min=1"`
	Ps      int    `form:"ps" validate:"min=1,max=50" default:"20"`
}

package search

type RecommendTagsReq struct {
	Style       int64  `form:"style"`
	Gt          string `form:"goto"`
	Id1st       string `form:"id_1st"`
	StartTs     int64  `form:"start_ts"`
	EndTs       int64  `form:"end_ts"`
	SLocale     string `form:"s_locale"`
	CLocale     string `form:"c_locale"`
	DisableRcmd int64  `form:"disable_rcmd"`
}

type RecommendTagsRsp struct {
	Title string          `json:"title"`
	Tags  []*RecommendTag `json:"tags"`
}

type RecommendTag struct {
	Query   string `json:"query"`
	JumpUrl string `json:"jump_url"`
}

type TagItemList struct {
	Query       string `json:"query"`
	SearchQuery string `json:"search_query"`
	ItemFeature string `json:"item_feature"`
}

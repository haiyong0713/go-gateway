package search

type SearchVideoCard struct {
	Author       string             `json:"author"`
	Cover        string             `json:"cover"`
	Description  string             `json:"description"`
	Favorites    int64              `json:"favorites"`
	HitColumns   []string           `json:"hit_columns"`
	ID           int64              `json:"id"`
	CateID       int64              `json:"cate_id"`
	CateName     string             `json:"cate_name"`
	Mid          int64              `json:"mid"`
	Play         int64              `json:"play"`
	RankSocre    int64              `json:"rank_socre"`
	SeasonList   []string           `json:"season_list"`
	MarkList     []string           `json:"mark_list"`
	Title        string             `json:"title"`
	Update       int64              `json:"update"`
	CardVideoNum int                `json:"card_video_num"`
	VideoList    []*SearchCardVideo `json:"video_list"`
}

type SearchCardVideo struct {
	Aid         int64    `json:"aid"`
	Avid        int64    `json:"avid"`
	Bvid        string   `json:"bvid"`
	Description string   `json:"description"`
	Episode     string   `json:"episode"`
	EpisodeNo   int64    `json:"episode_no"`
	Eptitle     string   `json:"eptitle"`
	HitColumns  []string `json:"hit_columns"`
	Mark        string   `json:"mark"`
	ID          int64    `json:"id"`
	RankScore   int64    `json:"rank_score"`
	Type        string   `json:"type"`
}

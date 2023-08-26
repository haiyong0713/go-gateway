package bangumi

type Bangumi struct {
	SeasonId     string `json:"season_id"`
	Spid         string `json:"spid"`
	Title        string `json:"title"`
	Brief        string `json:"brief"`
	Cover        string `json:"cover"`
	Evaluate     string `json:"evaluate"`
	TotalCount   string `json:"total_count"`
	PlayCount    string `json:"play_count"`
	DanmakuCount string `json:"danmaku_count"`
	Finish       string `json:"is_finish"`
	Badge        string `json:"badge"`
	SeasonStatus int    `json:"season_status"`
	Favorites    string `json:"favorites"`
	NewEp        struct {
		Aid    string `json:"av_id"`
		Cover  string `json:"cover"`
		Index  string `json:"index"`
		UpTime string `json:"update_time"`
	} `json:"new_ep"`
}

type SeasonInfo struct {
	SeasonID   int64 `json:"season_id"`
	SeasonType int   `json:"season_type"`
	EpisodeID  int   `json:"episode_id"`
}

type Banner struct {
	Title string `json:"title"`
	Image string `json:"img"`
	URI   string `json:"link"`
}

type EpPlayer struct {
	AID       int64  `json:"aid"`
	CID       int64  `json:"cid"`
	EpID      int64  `json:"episode_id"`
	Uri       string `json:"url"`
	Cover     string `json:"cover"`
	NewDesc   string `json:"new_desc"`
	ShowTitle string `json:"show_title"`
	Duration  int64  `json:"duration"`
	IsPreview int    `json:"is_preview"`
	Stat      struct {
		Play    int64 `json:"play"`
		Danmaku int64 `json:"danmaku"`
		Follow  int64 `json:"follow"`
	} `json:"stat"`
	Season struct {
		Type     int64  `json:"type"`
		Title    string `json:"title"`
		TypeName string `json:"type_name"`
		SeasonID int64  `json:"season_id"`
	} `json:"season"`
}

// CommonParam .
type CommonParam struct {
	MobiApp  string
	Build    int
	Fnval    int
	Fnver    int
	Device   string
	Platform string
	XTfIsp   string
}

package comic

// Comics get all comit.
type Comics struct {
	Total     int      `json:"total_count"`
	ComicList []*Comic `json:"comics"`
}

// Comic get from comit.
type Comic struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	VerticalCover string `json:"vertical_cover"`
	IsFinish      int8   `json:"is_finish"`
	Styles        []*struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"styles"`
	Total          int    `json:"total"`
	LastShortTitle string `json:"last_short_title"`
	LastUpdateTime string `json:"last_update_time"`
	URL            string `json:"url"`
}

// FavComic struct
type FavComic struct {
	ComicID  int64   `json:"comic_id"`
	Title    string  `json:"title"`
	Status   int     `json:"status"`       // 连载状态 1 未开刊, 2 连载中, 3 已完结
	LastOrd  float64 `json:"last_ord"`     // 看到话数
	OrdCount int     `json:"ord_count"`    // 新话数(总话数)
	HCover   string  `json:"hcover"`       // 横版封面
	SCover   string  `json:"scover"`       // 方版封面
	VCover   string  `json:"vcover"`       // 竖版封面
	PTime    string  `json:"publish_time"` // 漫画发布时间
	//nolint:govet
	LastEPPTime        string `json:"publish_time"`          // 漫画最新话发布时间
	LastEPID           int64  `json:"last_ep_id"`            // 看到话的编号
	LastEPShortTitle   string `json:"last_ep_short_title"`   // 看到话的标题
	LatestEPShortTitle string `json:"latest_ep_short_title"` // 当前漫画最新话
}

// FavComicCount struct
type FavComicCount struct {
	Count int `json:"count"`
}

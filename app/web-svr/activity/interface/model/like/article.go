package like

// ArticleGiant article giant.
type ArticleGiant struct {
	Articles int64 `json:"articles"`
	Likes    int64 `json:"likes"`
}

// ArticleList .
type ArticleList struct {
	ID           int64  `json:"id"`
	Mid          int64  `json:"mid"`
	Name         string `json:"name"`
	ImageURL     string `json:"image_url"`
	UpdateTime   int64  `json:"update_time"`
	Ctime        int64  `json:"ctime"`
	PublishTime  int64  `json:"publish_time"`
	Summary      string `json:"summary"`
	Words        int64  `json:"words"`
	Read         int64  `json:"read"`
	ArticleCount int64  `json:"article_count"`
}

package article

// Article .
type Article struct {
	Code int `json:"code"`
	Data struct {
		ID          int64    `json:"id"`
		Title       string   `json:"title"`
		ImageUrls   []string `json:"image_urls"`
		PublishTime int64    `json:"publish_time"`
		Author      struct {
			Mid int64 `json:"mid"`
		} `json:"author"`
	} `json:"Data"`
}

// ArticleInfo .
type ArticleInfo struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	ImageURLs []string `json:"image_urls"`
	Ctime     int64    `json:"ctime"`
}

// ArticleData .
type Articles struct {
	Code    int                     `json:"code"`
	Data    map[string]*ArticleInfo `json:"data"`
	Message string                  `json:"message"`
	TTL     int32                   `json:"ttl"`
}

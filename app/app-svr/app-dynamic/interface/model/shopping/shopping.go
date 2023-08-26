package shopping

type CardInfo struct {
	ItemsId      int64  `json:"itemsId"`
	Name         string `json:"name"`
	Img          string `json:"img"`
	PriceStr     string `json:"priceStr"`
	JumpLink     string `json:"jumpLink"`
	JumpLinkDesc string `json:"jumpLinkDesc"`
	SourceDesc   string `json:"sourceDesc"`
	AdMark       string `json:"adMark"`
}

type Item struct {
	Name  string `json:"name"`
	Price string `json:"price"`
	Img   string `json:"img"`
	URL   string `json:"url"`
	ID    int64  `json:"id"`
}

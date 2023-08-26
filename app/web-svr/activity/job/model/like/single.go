package like

type ImageUp struct {
	Mid   int64
	Score float64
}

type StupidVv struct {
	Mid int64 `json:"mid"`
	Vv  int64 `json:"vv"`
}

// ArcListData ...
type ArcListData struct {
	List []*ArcData `json:"list"`
}

// ArcData ...
type ArcData struct {
	ID   string    `json:"id"`
	Data *AidsData `json:"data"`
}

package model

// HisCursor .
type HisCursor struct {
	Max      int64  `json:"max"`
	ViewAt   int64  `json:"view_at"`
	Business string `json:"business"`
	Ps       int32  `json:"ps"`
}

// HisTab .
type HisTab struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// HisItem .
type HisItem struct {
	Title     string   `json:"title"`
	LongTitle string   `json:"long_title"`
	Cover     string   `json:"cover"`
	Covers    []string `json:"covers"`
	URI       string   `json:"uri"`
	History   struct {
		Oid      int64  `json:"oid"`
		Epid     int64  `json:"epid"`
		Bvid     string `json:"bvid"`
		Page     int32  `json:"page"`
		Cid      int64  `json:"cid"`
		Part     string `json:"part"`
		Business string `json:"business"`
		Dt       int32  `json:"dt"`
	} `json:"history"`
	Videos     int64  `json:"videos"`
	AuthorName string `json:"author_name"`
	AuthorFace string `json:"author_face"`
	AuthorMid  int64  `json:"author_mid"`
	ViewAt     int64  `json:"view_at"`
	Progress   int64  `json:"progress"`
	Badge      string `json:"badge"`
	ShowTitle  string `json:"show_title"`
	Duration   int64  `json:"duration"`
	Current    string `json:"current"`
	Total      int32  `json:"total"`
	NewDesc    string `json:"new_desc"`
	IsFinish   int32  `json:"is_finish"`
	IsFav      int32  `json:"is_fav"`
	Kid        int64  `json:"kid"`
	TagName    string `json:"tag_name"`
	LiveStatus int    `json:"live_status"`
}

// HisRes .
type HisRes struct {
	Cursor HisCursor  `json:"cursor"`
	Tab    []*HisTab  `json:"tab"`
	List   []*HisItem `json:"list"`
}

// WxHisRes .
type WxHisRes struct {
	Cursor HisCursor  `json:"cursor"`
	List   []*HisItem `json:"list"`
}

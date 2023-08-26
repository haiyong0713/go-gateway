package web

const (
	GotoLive = "live"
)

func FillURI(gt, param string) string {
	switch gt {
	case GotoLive:
		return "https://live.bilibili.com/" + param
	default:
	}
	return ""
}

type HisSearchReply struct {
	HasMore bool       `json:"has_more"`
	Page    *Page      `json:"page"`
	List    []*HisItem `json:"list"`
}

type Page struct {
	Pn    int64 `json:"pn"`
	Total int64 `json:"total"`
}

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
	Total      int32  `json:"total"`
	NewDesc    string `json:"new_desc"`
	IsFinish   int32  `json:"is_finish"`
	IsFav      int32  `json:"is_fav"`
	Kid        int64  `json:"kid"`
	TagName    string `json:"tag_name"`
	LiveStatus int    `json:"live_status"`
}

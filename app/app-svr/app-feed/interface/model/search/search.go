package search

type Upper struct {
	Mid       int64  `json:"up_id"`
	RecReason string `json:"rec_reason"`
}

// ArcSearchParam is
type ArcSearchParam struct {
	Mid       int64
	Tid       int64
	Order     string
	Keyword   string
	Pn        int64
	Ps        int64
	CheckType string
	CheckID   int64
	AttrNot   uint64
}

// ArcSearchReply is
type ArcSearchReply struct {
	TList map[string]*ArcSearchTList `json:"tlist"`
	VList []*ArcSearchVList          `json:"vlist"`
}

// ArcSearchTList is
type ArcSearchTList struct {
	Tid   int64  `json:"tid"`
	Count int64  `json:"count"`
	Name  string `json:"name"`
}

// ArcSearchVList is
type ArcSearchVList struct {
	Comment      int64       `json:"comment"`
	TypeID       int64       `json:"typeid"`
	Play         interface{} `json:"play"`
	Pic          string      `json:"pic"`
	SubTitle     string      `json:"subtitle"`
	Description  string      `json:"description"`
	Copyright    string      `json:"copyright"`
	Title        string      `json:"title"`
	Review       int64       `json:"review"`
	Author       string      `json:"author"`
	Mid          int64       `json:"mid"`
	Created      interface{} `json:"created"`
	Length       string      `json:"length"`
	VideoReview  int64       `json:"video_review"`
	Aid          int64       `json:"aid"`
	Bvid         string      `json:"bvid"`
	HideClick    bool        `json:"hide_click"`
	IsPay        int         `json:"is_pay"`
	IsUnionVideo int         `json:"is_union_video"`
}

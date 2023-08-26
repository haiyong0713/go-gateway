package model

import xtime "go-common/library/time"

type ReqTagTop struct {
	Tid    int64
	TName  string
	Mid    int64
	RealIP string
}

type TagTop struct {
	Tag      *TagTopTag    `json:"tag"`
	Similars []*SimilarTag `json:"similars"`
}

type TagTopTag struct {
	ID           int64      `json:"tag_id"`
	Name         string     `json:"tag_name"`
	Cover        string     `json:"cover"`
	HeadCover    string     `json:"head_cover"`
	Content      string     `json:"content"`
	ShortContent string     `json:"short_content"`
	Type         int8       `json:"type"`
	State        int8       `json:"state"`
	CTime        xtime.Time `json:"ctime"`
	MTime        xtime.Time `json:"-"`
	// tag count
	Count struct {
		View  int `json:"view"`
		Use   int `json:"use"`
		Atten int `json:"atten"`
	} `json:"count"`
	// subscriber
	IsAtten int8 `json:"is_atten"`
	// archive_tag
	Role      int8  `json:"-"`
	Likes     int64 `json:"likes"`
	Hates     int64 `json:"hates"`
	Attribute int8  `json:"attribute"`
	Liked     int8  `json:"liked"`
	Hated     int8  `json:"hated"`
	ExtraAttr int32 `json:"extra_attr"`
}

type SimilarTag struct {
	Rid    int64  `json:"rid"`
	Rname  string `json:"rname"`
	Tid    int64  `json:"tid"`
	TCover string `json:"cover"`
	Tatten int    `json:"atten"`
	Tname  string `json:"tname"`
}

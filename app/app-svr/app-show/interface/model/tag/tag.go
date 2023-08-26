package tag

import xtime "go-common/library/time"

type Tag struct {
	Tid     int64  `json:"tag_id"`
	Name    string `json:"tag_name"`
	IsAtten int8   `json:"is_atten"`
	Count   struct {
		Atten int `json:"atten,omitempty"`
	} `json:"count,omitempty"`
}

// TagOld .
type TagOld struct {
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

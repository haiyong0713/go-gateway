package favorite

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
)

// Article is
const (
	Article       = int8(1)
	TypeVideo     = int8(2)
	TypeMusic     = int8(3)
	TypeTopic     = int8(4)
	TypePlayVideo = int8(5)
	TypePlayList  = int8(6)
	TypeBangumi   = int8(7)
	TypeMoe       = int8(8)
	TypeComic     = int8(9)
	TypeEsports   = int8(10)
	TypeMediaList = int8(11)
	TypeMusicNew  = int8(12)
	TypeOgv       = int8(24)
)

// Folder struct
type Folder struct {
	MediaID    int64   `json:"media_id"`
	Fid        int     `json:"fid"`
	Mid        int     `json:"mid"`
	Name       string  `json:"name"`
	MaxCount   int     `json:"max_count"`
	CurCount   int     `json:"cur_count"`
	AttenCount int     `json:"atten_count"`
	State      int     `json:"state"`
	CTime      int     `json:"ctime"`
	MTime      int     `json:"mtime"`
	Cover      []Cover `json:"cover,omitempty"`
	Videos     []Cover `json:"videos,omitempty"` // NOTE: old favourite
}

// Folder2 from space.
type Folder2 struct {
	MediaID   int64      `json:"media_id"`
	ID        int64      `json:"id"`
	Mid       int64      `json:"mid"`
	Title     string     `json:"title"`
	Cover     string     `json:"cover"`
	Count     int        `json:"count"`
	Type      int32      `json:"type"`
	IsPublic  int8       `json:"is_public"`
	CTime     xtime.Time `json:"ctime"`
	MTime     xtime.Time `json:"mtime"`
	IsDefault bool       `json:"is_default,omitempty"`
}

// Cover struct
type Cover struct {
	Aid  int    `json:"aid"`
	Pic  string `json:"pic"`
	Type int32  `json:"type"`
}

// Video struct
type Video struct {
	Seid           string `json:"seid"`
	Page           int    `json:"page"`
	Pagesize       int    `json:"pagesize"`
	PageCount      int    `json:"pagecount"`
	Total          int    `json:"total"`
	SuggestKeyword string `json:"suggest_keyword"`
	Mid            int64  `json:"mid"`
	Fid            int64  `json:"fid"`
	Tid            int    `json:"tid"`
	Order          string `json:"order"`
	Keyword        string `json:"keyword"`
	Tlist          []struct {
		Tid   int16  `json:"tid"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tlist,omitempty"`
	Archives []*Archive `json:"archives"`
}

// Archive struct
type Archive struct {
	*api.Arc
	FavAt          int64  `json:"fav_at"`
	PlayNum        string `json:"play_num"`
	HighlightTitle string `json:"highlight_title"`
}

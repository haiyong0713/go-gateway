package space

import (
	xtime "go-common/library/time"
)

const (
	_gotoAv      = 0
	_gotoArticle = 1
	_gotoClip    = 2
	_gotoAlbum   = 3
	_gotoAudio   = 4
	_gotoComic   = 5

	GotoAv      = "av"
	GotoArticle = "article"
	GotoClip    = "clip"
	GotoAlbum   = "album"
	GotoAudio   = "audio"
	GotoComic   = "comic"

	AttrNo         = int32(0)
	AttrYes        = int32(1)
	AttrBitArchive = uint32(0)
	AttrBitArticle = uint32(1)
	AttrBitClip    = uint32(2)
	AttrBitAlbum   = uint32(3)
	AttrBitAudio   = uint32(34)
	AttrBitComic   = uint32(4)
)

type Attrs struct {
	Archive bool `json:"archive,omitempty"`
	Article bool `json:"article,omitempty"`
	Clip    bool `json:"clip,omitempty"`
	Album   bool `json:"album,omitempty"`
	Audio   bool `json:"audio,omitempty"`
	Comic   bool `json:"comic,omitempty"`
}

func (attrs *Attrs) Attr() (attr int32) {
	if attrs == nil {
		return
	}
	if attrs.Archive {
		attr = AttrSet(attr, AttrYes, AttrBitArchive)
	}
	if attrs.Article {
		attr = AttrSet(attr, AttrYes, AttrBitArticle)
	}
	if attrs.Clip {
		attr = AttrSet(attr, AttrYes, AttrBitClip)
	}
	if attrs.Album {
		attr = AttrSet(attr, AttrYes, AttrBitAlbum)
	}
	if attrs.Audio {
		attr = AttrSet(attr, AttrYes, AttrBitAudio)
	}
	if attrs.Comic {
		attr = AttrSet(attr, AttrYes, AttrBitComic)
	}
	return
}

type Item struct {
	ID     int64      `json:"id,omitempty"`
	Goto   string     `json:"goto,omitempty"`
	CTime  xtime.Time `json:"ctime,omitempty"`
	Member int64      `json:"member,omitempty"`
}

type Album struct {
	DocID int64      `json:"doc_id,omitempty"`
	CTime xtime.Time `json:"ctime,omitempty"`
}

type Audio struct {
	ID    int64      `json:"audioId,omitempty"`
	CTime xtime.Time `json:"cTime,omitempty"`
}

// Comics get all comit.
type Comics struct {
	Total     int      `json:"total_count"`
	ComicList []*Comic `json:"comics"`
}

// Comic get from comit.
type Comic struct {
	ID             int64  `json:"id"`
	LastUpdateTime string `json:"last_update_time"`
}

// FormatKey func
//
//nolint:gomnd
func (i *Item) FormatKey() {
	switch i.Goto {
	case GotoAv:
		i.Member = i.ID<<6 | _gotoAv
	case GotoArticle:
		i.Member = i.ID<<6 | _gotoArticle
	case GotoClip:
		i.Member = i.ID<<6 | _gotoClip
	case GotoAlbum:
		i.Member = i.ID<<6 | _gotoAlbum
	case GotoAudio:
		i.Member = i.ID<<6 | _gotoAudio
	case GotoComic:
		i.Member = i.ID<<6 | _gotoComic
	default:
		i.Member = i.ID
	}
}

func FormatKey(id int64, gt string) int64 {
	switch gt {
	case GotoAv:
		return id<<6 | _gotoAv
	case GotoArticle:
		return id<<6 | _gotoArticle
	case GotoClip:
		return id<<6 | _gotoClip
	case GotoAlbum:
		return id<<6 | _gotoAlbum
	case GotoAudio:
		return id<<6 | _gotoAudio
	}
	return id
}

// AttrSet set attribute value
func AttrSet(attr int32, v int32, bit uint32) int32 {
	return attr&(^(1 << bit)) | (v << bit)
}

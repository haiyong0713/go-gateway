package model

import (
	"database/sql/driver"
	"time"
)

// Archive archive result
type Archive struct {
	ID          int64     `json:"id"`
	AID         int64     `json:"aid"`
	Mid         int64     `json:"mid"`
	TypeID      int32     `json:"typeid"`
	Videos      int       `json:"videos"`
	Title       string    `json:"title"`
	Cover       string    `json:"cover"`
	Content     string    `json:"content"`
	Duration    int       `json:"duration"`
	Attribute   int32     `json:"attribute"`
	Copyright   int8      `json:"copyright"`
	Access      int       `json:"access"`
	PubTime     wocaoTime `json:"pubtime"`
	CTime       wocaoTime `json:"ctime"`
	MTime       wocaoTime `json:"mtime"`
	State       int       `json:"state"`
	MissionID   int64     `json:"mission_id"`
	OrderID     int64     `json:"order_id"`
	RedirectURL string    `json:"redirect_url"`
	Forward     int64     `json:"forward"`
	Dynamic     string    `json:"dynamic"`
	SeasonID    int64     `json:"season_id"`
	AttributeV2 int64     `json:"attribute_v2"`
	UpFromV2    int32     `json:"up_from"`
	FirstFrame  string    `json:"first_frame"`
}

// Video is
type Video struct {
	AID int64 `json:"aid"`
	CID int64 `json:"cid"`
}

type ArcExpand struct {
	Aid          int64     `json:"aid"`
	Mid          int64     `json:"mid"`
	ArcType      int64     `json:"arc_type"`
	RoomId       int64     `json:"room_id"`
	PremiereTime time.Time `json:"premiere_time"`
}

type SeasonEpisode struct {
	SeasonId  int64 `json:"season_id"`
	SectionId int64 `json:"section_id"`
	EpisodeId int64 `json:"episode_id"`
	Aid       int64 `json:"aid"`
	Attribute int64 `json:"attribute"`
}

func (sep *SeasonEpisode) AttrVal(bit uint) int32 {
	return int32((sep.Attribute >> bit) & int64(1))
}

type wocaoTime string

const _formatTime = "2006-01-02 15:04:05"

// Scan scan time.
func (jt *wocaoTime) Scan(src interface{}) (err error) {
	switch sc := src.(type) {
	case time.Time:
		*jt = wocaoTime(sc.Format(_formatTime))
	case string:
		*jt = wocaoTime(sc)
	}
	return
}

// Value get time value.
func (jt wocaoTime) Value() (driver.Value, error) {
	return time.Parse(_formatTime, string(jt))
}

func (jt wocaoTime) UnixValue() (int64, error) {
	t, err := time.ParseInLocation(_formatTime, string(jt), time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

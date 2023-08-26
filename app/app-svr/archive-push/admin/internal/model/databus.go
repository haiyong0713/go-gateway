package model

import (
	xdatabus "go-common/library/queue/databus"
	xtime "go-common/library/time"
)

type DatabusConfigs struct {
	ArchiveNotifySub     *xdatabus.Config
	UserAuthorizationSub *xdatabus.Config
}

type ArchiveNotify struct {
	Action string                `json:"action"`
	Table  string                `json:"table"`
	New    *ArchiveNotifyArchive `json:"new"`
	Old    *ArchiveNotifyArchive `json:"old"`
}

type ArchiveNotifyArchive struct {
	ID          int64      `json:"id"`
	AID         int64      `json:"aid"`
	MID         int64      `json:"mid"`
	Attribute   int32      `json:"attribute"`
	AttributeV2 int32      `json:"attribute_v2"`
	Content     string     `json:"content"`
	Copyright   int        `json:"copyright"`
	Cover       string     `json:"cover"`
	Duration    int64      `json:"duration"`
	Dynamic     string     `json:"dynamic"`
	Forward     int        `json:"forward"`
	MissionID   int64      `json:"mission_id"`
	OrderID     int64      `json:"order_id"`
	Pubtime     string     `json:"pubtime"`
	RedirectURL string     `json:"redirect_url"`
	SeasonID    int64      `json:"season_id"`
	State       int        `json:"state"`
	Title       string     `json:"title"`
	TypeID      int32      `json:"typeid"`
	Videos      int        `json:"videos"`
	CTimeStr    string     `json:"ctime"`
	CTimeTime   xtime.Time `json:"-"`
	MTimeStr    string     `json:"mtime"`
	MTimeTime   xtime.Time `json:"-"`
}

type UserAuthorizationContent struct {
	SID   int64 `json:"sid"`
	MID   int64 `json:"mid"`
	State int   `json:"state"`
}

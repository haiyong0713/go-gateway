package history

import cardm "go-gateway/app/app-svr/app-card/interface/model"

// HisParam fro history
type HisParam struct {
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Build    int64  `form:"build"`
	Platform string `form:"platform"`
	Pn       int32  `form:"pn"`
	Ps       int32  `form:"ps"`
	Mid      int64  `form:"mid"`
	Max      int64  `form:"max"`
	MaxTP    int32  `form:"max_tp"`
	Business string `form:"business" default:"all"`
	Buvid    string `form:"buvid"`
	Islocal  bool   `form:"is_local"`
}

// LiveParam statue param
type LiveParam struct {
	RoomIDs    string `form:"room_ids"`
	Uid        int64
	Platform   string
	ReqBiz     string
	DeviceName string
	NetWork    string
	Build      int64
}

// DelParam del param
type DelParam struct {
	Mid   int64    `form:"mid"`
	Boids []string `form:"boids,split" validate:"min=1"`
}

// ClearParam clear param
type ClearParam struct {
	Mid      int64  `form:"mid"`
	Business string `form:"business"`
}

// ListRes for history
type ListRes struct {
	Title   string   `json:"title"`
	Covers  []string `json:"covers,omitempty"`
	Cover   string   `json:"cover,omitempty"`
	URI     string   `json:"uri"`
	History struct {
		Oid      int64  `json:"oid"`
		Tp       int32  `json:"tp"`
		Cid      int64  `json:"cid,omitempty"`
		Page     int32  `json:"page,omitempty"`
		Part     string `json:"part,omitempty"`
		Business string `json:"business"`
		Kid      int64  `json:"-"`
		Dt       int32  `json:"-"`
		//产生该条历史记录的buvid
		WatchedBuvid string `json:"-"`
	} `json:"history"`
	Videos           int64           `json:"videos,omitempty"`
	View             int64           `json:"view,omitempty"`
	Name             string          `json:"name,omitempty"`
	Mid              int64           `json:"mid,omitempty"`
	Goto             string          `json:"goto"`
	Badge            string          `json:"badge,omitempty"`
	ViewAt           int64           `json:"view_at"`
	Progress         int64           `json:"progress,omitempty"`
	Duration         int64           `json:"duration,omitempty"`
	ShowTitle        string          `json:"show_title,omitempty"`
	TagName          string          `json:"tag_name,omitempty"`
	LiveStatus       int             `json:"live_status,omitempty"`
	LiveParentAreaId int64           `json:"live_parent_area_id,omitempty"`
	LiveAreaId       int64           `json:"live_area_id,omitempty"`
	DisAtten         int32           `json:"display_attention,omitempty"`
	Relation         *cardm.Relation `json:"-"`
	State            int64           `json:"-"`
}

// PGCRes for history
type PGCRes struct {
	EpID      int64  `json:"ep_id"`
	Cover     string `json:"cover"`
	URI       string `json:"uri"`
	Title     string `json:"title"`
	ShowTitle string `json:"show_title"`
	Season    struct {
		Title string `json:"title"`
	} `json:"season"`
}

// ListCursor for history
type ListCursor struct {
	Tab    []*BusTab  `json:"tab"`
	List   []*ListRes `json:"list"`
	Cursor *Cursor    `json:"cursor"`
}

// BusTab business tab
type BusTab struct {
	Business string `json:"business"`
	Name     string `json:"name"`
	Router   string `json:"router"`
}

// Cursor for history
type Cursor struct {
	Max   int64 `json:"max"`
	MaxTP int32 `json:"max_tp"`
	Ps    int32 `json:"ps"`
}

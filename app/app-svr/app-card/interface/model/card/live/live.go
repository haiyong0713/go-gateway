package live

type Room struct {
	UID                int64  `json:"uid,omitempty"`
	RoomID             int64  `json:"room_id,omitempty"`
	Title              string `json:"title,omitempty"`
	Cover              string `json:"cover,omitempty"`
	Uname              string `json:"uname,omitempty"`
	Face               string `json:"face,omitempty"`
	Online             int32  `json:"online,omitempty"`
	LiveStatus         int8   `json:"live_status,omitempty"`
	AreaV2ParentID     int64  `json:"area_v2_parent_id,omitempty"`
	AreaV2ParentName   string `json:"area_v2_parent_name,omitempty"`
	AreaV2ID           int64  `json:"area_v2_id,omitempty"`
	AreaV2Name         string `json:"area_v2_name,omitempty"`
	BroadcastType      int    `json:"broadcast_type,omitempty"`
	PlayurlH264        string `json:"playurl_h264,omitempty"`
	PlayurlH265        string `json:"playurl_h265,omitempty"`
	AcceptQuality      []int  `json:"accept_quality,omitempty"`
	CurrentQuality     int    `json:"current_quality,omitempty"`
	CurrentQn          int    `json:"current_qn,omitempty"`
	QualityDescription []*struct {
		Qn   int    `json:"qn,omitempty"`
		Desc string `json:"desc,omitempty"`
	} `json:"quality_description,omitempty"`
	ExtraParameter string       `json:"extra_parameter,omitempty"`
	PendentRu      string       `json:"pendent_ru,omitempty"`
	PendentRuColor string       `json:"pendent_ru_color,omitempty"`
	PendentRuPic   string       `json:"pendent_ru_pic,omitempty"`
	Link           string       `json:"link,omitempty"`
	AllPendants    []*Pendants  `json:"all_pendants,omitempty"`
	HotRank        int64        `json:"hot_rank,omitempty"`
	WatchedShow    *WatchedShow `json:"watched_show,omitempty"`
}

type WatchedShow struct {
	Switch bool  `json:"switch,omitempty"`
	Num    int64 `json:"num,omitempty"`
}

type Pendants struct {
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	Position  int64  `json:"position,omitempty"`
	Text      string `json:"text,omitempty"`
	BgColor   string `json:"bg_color,omitempty"`
	BgPic     string `json:"bg_pic,omitempty"`
	PendantID int64  `json:"pendant_id,omitempty"`
	Priority  int64  `json:"priority,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
}

type Card struct {
	RoomID        int64  `json:"roomid,omitempty"`
	UID           int64  `json:"uid,omitempty"`
	Title         string `json:"title,omitempty"`
	Uname         string `json:"uname,omitempty"`
	ShowCover     string `json:"show_cover,omitempty"`
	Online        int32  `json:"online,omitempty"`
	LiveStatus    int8   `json:"live_status,omitempty"`
	BroadcastType int    `json:"broadcast_type,omitempty"`
}

type TopicImage struct {
	ImageSrc    string `json:"image_src"`
	ImageWidth  int    `json:"image_width"`
	ImageHeight int    `json:"image_height"`
}

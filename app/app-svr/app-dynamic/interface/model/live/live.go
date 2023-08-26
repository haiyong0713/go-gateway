package live

const (
	CardTypeLiving   = 1
	CardTypePlayBack = 2
)

// 直播间资源
type LiveResItem struct {
	RoomID           int64  `json:"roomid"`
	Uid              int64  `json:"uid"`
	UName            string `json:"uname"`
	Verify           string `json:"verify"`
	Virtual          int    `json:"virtual"`
	Cover            string `json:"cover"`
	LiveTime         string `json:"live_time"`
	RoundStatus      int    `json:"round_status"`
	OnFlag           int    `json:"on_flag"`
	Title            string `json:"title"`
	Tags             string `json:"tags"`
	LockStatus       string `json:"lock_status"`
	HiddenStatus     string `json:"hidden_status"`
	UserCover        string `json:"user_cover"`
	ShortID          int64  `json:"sort_id"`
	Online           int64  `json:"online"`
	Area             int64  `json:"area"`
	AreaV2ID         int64  `json:"area_v2_id"`
	AreaV2ParentID   int64  `json:"area_v2_parent_id"`
	Attentions       int64  `json:"attentions"`
	Background       string `json:"background"`
	RoomSilent       int    `json:"room_silent"`
	RoomShield       int    `json:"room_shield"`
	TryTime          string `json:"try_time"`
	AreaV2Name       string `json:"area_v2_name"`
	FirstLiveTime    string `json:"first_live_time"`
	LiveStatus       int    `json:"live_status"`
	AreaV2ParentName string `json:"area_v2_parent_name"`
	BroadCastType    int    `json:"broadcast_type"`
	Face             string `json:"face"`
}

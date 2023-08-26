package live

// Room .
type Room struct {
	Title      string `json:"title,omitempty"`
	RoomID     int64  `json:"room_id,omitempty"`
	Cover      string `json:"cover,omitempty"`
	LiveStatus int8   `json:"live_status,omitempty"`
	UID        int64  `json:"uid,omitempty"`
	PlayURL    string `json:"play_url,omitempty"`
	Tips       string `json:"tips,omitempty"`
}

// Rooms .
type Rooms struct {
	Code int
	Data map[int64]*Room
}

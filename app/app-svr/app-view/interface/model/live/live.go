package live

type Live struct {
	Mid        int64  `json:"mid"`
	RoomID     int64  `json:"roomid"`
	URI        string `json:"uri,omitempty"`
	EndPageUri string `json:"endpage_uri"`
}

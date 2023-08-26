package model

type AppMRoomReq struct {
	RoomIds        []int64 `json:"room_ids"`
	Mid            int64   `json:"mid"`
	Platform       string  `json:"platform"`
	DeviceName     string  `json:"device_name"`
	AccessKey      string  `json:"access_key"`
	ActionKey      string  `json:"action_key"`
	Appkey         string  `json:"appkey"`
	Device         string  `json:"device"`
	MobiApp        string  `json:"mobi_app"`
	Statistics     string  `json:"statistics"`
	Buvid          string  `json:"buvid"`
	Network        string  `json:"network"`
	Build          int     `json:"build"`
	TeenagersMode  int     `json:"teenagers_mode"`
	Appver         int     `json:"appver"`
	Filtered       int     `json:"filtered"`
	HttpsUrlReq    int     `json:"https_url_req"`
	NeedRoomFilter int     `json:"need_room_filter"`
}

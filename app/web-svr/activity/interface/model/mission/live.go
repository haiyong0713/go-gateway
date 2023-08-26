package mission

import archive "go-gateway/app/app-svr/archive/service/api"

type LiveRoomInfo struct {
	RoomId         int64  `json:"room_id"`
	RoomMid        int64  `json:"room_mid"`
	RoomUserName   string `json:"room_user_name"`
	RoomUserAvatar string `json:"room_user_avatar"`
	RoomTitle      string `json:"room_title"`
	RoomCover      string `json:"room_cover"`
	JumpUrl        string `json:"jump_url"`
	RoomKeyFrame   string `json:"room_key_frame"`
	RoomStatus     int64  `json:"room_status"`
	Online         int64  `json:"online"`
}

type LiveRoomListOper struct {
	EntryFrom string `json:"entry_from"`
	RoomIds   string `json:"room_ids"`
}

type LiveRoomList struct {
	EntryFrom string
	RoomIds   []int64
}

type VideoAidListOper struct {
	RoomIds string `json:"video_ids"`
}

type VideoInfo struct {
	Id         int64          `json:"id"`
	Author     archive.Author `json:"author"`
	Stat       archive.Stat   `json:"stat"`
	VideoCover string         `json:"video_cover"`
	VideoTitle string         `json:"video_title"`
	VideoUrl   string         `json:"video_url"`
	Duration   int64          `json:"duration"`
}

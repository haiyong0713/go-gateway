package model

type BgmInfo struct {
	MusicId    string `json:"music_id"`
	MusicTitle string `json:"music_title"`
	JumpUrl    string `json:"jump_url"`
}

type BgmEntranceReply struct {
	State int      `json:"music_state"`
	Info  *BgmInfo `json:"music_info"`
}

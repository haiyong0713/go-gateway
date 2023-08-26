package music

type Entrance struct {
	MusicState int32      `json:"music_state"`
	MusicInfo  *MusicInfo `json:"music_info"`
}

type MusicInfo struct {
	JumpUrl    string `json:"jump_url"`
	MusicTitle string `json:"music_title"`
	MusicId    string `json:"music_id"`
}

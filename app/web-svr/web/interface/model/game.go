package model

// Game game info struct.
type Game struct {
	GameBaseID     int64   `json:"game_base_id"`
	GameName       string  `json:"game_name"`
	GameIcon       string  `json:"game_icon"`
	GameBookStatus int     `json:"game_book_status"`
	GameLink       string  `json:"game_link"`
	Grade          float64 `json:"grade"`
	BookNum        int64   `json:"book_num"`
}

type mediaScore struct {
	MediaName  string `json:"media_name"`
	MediaScore string `json:"media_score"`
}

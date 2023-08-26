package esports

import xtime "go-common/library/time"

// 用户的偏好游戏设置
type EsportsActFav struct {
	ID                int64      `json:"id"`
	Mid               int64      `json:"mid"`
	FirstFavGameId    int64      `json:"first_fav_game_id"`
	FirstFavGameName  string     `json:"first_fav_game_name"`
	SecondFavGameId   int64      `json:"second_fav_game_id"`
	SecondFavGameName string     `json:"second_fav_game_name"`
	ThirdFavGameId    int64      `json:"third_fav_game_id"`
	ThirdFavGameName  string     `json:"third_fav_game_name"`
	Ctime             xtime.Time `json:"ctime"`
	Mtime             xtime.Time `json:"mtime"`
}

type UserInfo struct {
	FavCompleted     bool `json:"fav_completed"`
	CollectCompleted bool `json:"collect_completed"`
	FavEsports       *EsportsActFav
}

package model

type ComicInfo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	FavStatus int64  `json:"fav_status"` //用户是否追漫：0 未追；1 已追
}

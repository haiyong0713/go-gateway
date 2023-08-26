package cartoon

type ComicItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// 用户是否追漫，0 未追；1 已追
	FavStatus int32 `json:"fav_status"`
}

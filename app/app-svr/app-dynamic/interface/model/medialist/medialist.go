package medialist

// 播单卡资源
type FavoriteRes struct {
	Cards map[int64]*FavoriteItem `json:"cards"`
}

type FavoriteItem struct {
	ID         int64     `json:"id"`  // 播单ID
	Fid        int64     `json:"fid"` // fav id
	Mid        int64     `json:"mid"` // 创建者ID
	Title      string    `json:"title"`
	Cover      string    `json:"cover"`
	Intro      string    `json:"intro"` // 播单描述
	MediaCount int       `json:"media_count"`
	Sharable   bool      `json:"sharable"` // 是否支持分享
	Upper      *FavUpper `json:"upper"`    // 创建者信息
	Type       int       `json:"type"`
	CoverType  int32     `json:"cover_type"` // 2 视频封面 12 音频封面
}

type FavUpper struct {
	Face     string `json:"face"`
	Followed int    `json:"followed"`
	Mid      int64  `json:"mid"`
	Name     string `json:"name"`
}

package archive

import go_common_library_time "go-common/library/time"

// Archive 稿件
type Archive struct {
	Bvid      string                      `json:"bvid"`
	Tname     string                      `json:"tname"`
	Title     string                      `json:"title"`
	Desc      string                      `json:"desc"`
	Duration  int64                       `json:"duration"`
	Pic       string                      `json:"pic"`
	ShortLink string                      `json:"short_link"`
	View      int64                       `json:"view"`
	Like      int64                       `json:"like"`
	Danmaku   int64                       `json:"danmaku"`
	Reply     int64                       `json:"reply"`
	Fav       int64                       `json:"fav"`
	Coin      int64                       `json:"coin"`
	Share     int64                       `json:"share"`
	Ctime     go_common_library_time.Time `json:"ctime"`
	Account   *Account                    `json:"account"`
}

// Account 账号
type Account struct {
	Name string `json:"name" `
	MID  int64  `json:"mid"`
	Face string `json:"face"`
}

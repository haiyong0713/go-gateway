package model

import "go-common/library/time"

// AlbumCount album count struct.
type AlbumCount struct {
	AllCount   int64 `json:"all_count"`
	DrawCount  int64 `json:"draw_count"`
	PhotoCount int64 `json:"photo_count"`
	DailyCount int64 `json:"daily_count"`
}

// Album album struct.
type Album struct {
	DocID       int64      `json:"doc_id"`
	PosterUID   int64      `json:"poster_uid"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Pictures    []*PicItem `json:"pictures"`
	Count       int64      `json:"count"`
	Ctime       time.Time  `json:"ctime"`
	View        int64      `json:"view"`
	Like        int64      `json:"like"`
}

// PicItem picture item struct.
type PicItem struct {
	ImgSrc    string  `json:"img_src"`
	ImgWidth  float32 `json:"img_width"`
	ImgHeight float32 `json:"img_height"`
	ImgSize   float32 `json:"img_size"`
}

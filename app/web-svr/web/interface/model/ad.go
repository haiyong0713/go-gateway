package model

import "encoding/json"

type CpmsRequestParam struct {
	Mid       int64
	Aid       int64
	UpID      int64
	Sid       string
	IP        string
	Country   string
	Province  string
	City      string
	Buvid     string
	UserAgent string
	FromSpmID string
	Ids       []int64
}

// Ad struct
type Ad struct {
	RequestID  string                         `json:"request_id"`
	AdsInfo    map[string]map[string]*AdsInfo `json:"ads_info"`
	AdsControl json.RawMessage                `json:"ads_control"`
}

// AdsInfo struct
type AdsInfo struct {
	Index  int64   `json:"index"`
	IsAd   bool    `json:"is_ad"`
	CmMark int8    `json:"cm_mark"`
	AdInfo *AdInfo `json:"ad_info"`
}

// AdInfo struct
type AdInfo struct {
	CreativeID      int64 `json:"creative_id"`
	CreativeType    int8  `json:"creative_type"`
	CreativeContent struct {
		Title        string `json:"title"`
		Desc         string `json:"description"`
		VideoID      int64  `json:"video_id"`
		UserName     string `json:"username"`
		ImageURL     string `json:"image_url"`
		ImageMD5     string `json:"image_md5"`
		LogURL       string `json:"log_url"`
		LogMD5       string `json:"log_md5"`
		URL          string `json:"url"`
		ClickURL     string `json:"click_url"`
		ShowURL      string `json:"show_url"`
		ThumbnailURL string `json:"thumbnail_url"`
	} `json:"creative_content"`
	AdCb  string `json:"ad_cb"`
	Extra struct {
		Card struct {
			AdverName    string          `json:"adver_name"`
			BusinessMark json.RawMessage `json:"ad_tag_style"`
		} `json:"card"`
	} `json:"extra"`
	CardType int64 `json:"card_type"`
}

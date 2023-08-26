package model

import (
	"time"
)

var (
	UploadTopPhotoWeb     = 4
	UploadTopPhotoIpad    = 3
	UploadTopPhotoAndroid = 2
	UploadTopPhotoIos     = 1
	UploadTopPhotoZero    = 0
	UploadTopPhotoPass    = 1
	UploadTopPhotoVerify  = 0
	MemTopPhotoMaxUpload  = 5 * 1024 * 1024
)

// MemberTopphoto member top photo
type MemberTopphoto struct {
	ID          int64 `json:"id"`
	Mid         int64 `json:"mid"`
	Sid         int64 `json:"sid"`
	Platfrom    int64 `json:"platfrom"`
	Expire      int64 `json:"expire"`
	IsActivated int64 `json:"is_activated"`
}

// MemberPhotoUpload .
type MemberPhotoUpload struct {
	ID         int64     `json:"id"`
	Mid        int64     `json:"mid"`
	ImgPath    string    `json:"img_path"`
	Platfrom   int       `json:"platform"`
	Status     int       `json:"status"`
	Deleted    int64     `json:"deleted"`
	UploadDate time.Time `json:"upload_date"`
	ModifyTime time.Time `json:"modify_time"`
}

// PhotoMallIndex
type PhotoMallIndex struct {
	ID           int64  `json:"id"`
	IsDisable    int64  `json:"is_disable"`
	Price        int64  `json:"price"`
	CoinType     int64  `json:"coin_type"`
	VipFree      int64  `json:"vip_free"`
	SortNum      int64  `json:"sort_num"`
	Expire       int64  `json:"expire,omitempty"`
	Had          int64  `json:"had,omitempty"`
	ProductName  string `json:"product_name"`
	SImg         string `json:"s_img"`
	LImg         string `json:"l_img"`
	ThumbnailImg string `json:"thumbnail_img"`
}

// MemberPhotoMallIndex
type MemberPhotoMallIndex struct {
	SID          int64  `json:"sid"`
	Expire       int64  `json:"expire,omitempty"`
	Platform     int64  `json:"platform"`
	SImg         string `json:"s_img"`
	LImg         string `json:"l_img"`
	AndroidImg   string `json:"android_img"`
	IphoneImg    string `json:"iphone_img"`
	IpadImg      string `json:"ipad_img"`
	ThumbnailImg string `json:"thumbnail_img"`
}

// PurgeCacheReq
type PurgeCacheParam struct {
	Mid          int64  `form:"mid"`
	Status       int64  `form:"status"`
	BuyMonth     int64  `form:"buyMonths"`
	Days         int64  `form:"days"`
	ModifiedAttr string `form:"modifiedAttr"`
}

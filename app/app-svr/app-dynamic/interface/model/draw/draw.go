package model

type LBSSearchItem struct {
	Pio string
}

type MallSearchItem struct {
	Title          string  `json:"title"`
	Cover          string  `json:"cover"`
	Url            string  `json:"url"`
	ItemId         int64   `json:"id"`
	RequiredNumber int     `json:"required_number"`
	Price          float64 `json:"price"`
	Brief          string  `json:"brief"`
	PriceEqual     *int    `json:"price_equal"`
}

type UserSearchItem struct {
	Face string `json:"upic"`
	Name string `json:"uname"`
	Mid  uint64 `json:"mid"`
}

type UserProfile struct {
	Info *UserProfileInfo `json:"info"`
}

type UserProfileInfo struct {
	Uid               uint64 `json:"uid"`
	UserName          string `json:"uname"`
	Face              string `json:"face"`
	Identification    string `json:"identification"`
	MobileVerify      string `json:"mobile_verify"`
	PlatformUserLevel string `json:"platform_user_level"`
}

type TopicHotItem struct {
	TopicName string `json:"topic_name"`
	TopicId   uint64 `json:"topic_id"`
}

type TopicSearchItem struct {
	TopicName string `json:"tag_name"`
	TopicId   uint64 `json:"tag_id"`
}

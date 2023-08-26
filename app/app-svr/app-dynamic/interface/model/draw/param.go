package model

// 入参
type SearchAllReq struct {
	Uid      uint64  `form:"uid" validate:"gt=0"`
	Keyword  string  `form:"keyword"`
	Page     int32   `form:"page" validate:"gte=0"`
	PageSize int32   `form:"page_size" default:"20" validate:"gte=0"`
	Lat      float64 `form:"lat" validate:"gte=0"`
	Lng      float64 `form:"lng" validate:"gte=0"`
}

type SearchUsersReq struct {
	Uid      uint64 `form:"uid" validate:"gt=0"`
	Keyword  string `form:"keyword"`
	Page     int32  `form:"page" validate:"gte=0"`
	PageSize int32  `form:"page_size" default:"20" validate:"gte=0"`
}
type SearchTopicsReq struct {
	Uid      uint64 `form:"uid" validate:"gt=0"`
	Keyword  string `form:"keyword"`
	Page     int32  `form:"page" validate:"gte=0"`
	PageSize int32  `form:"page_size" default:"20" validate:"gte=0"`
}
type SearchLocationsReq struct {
	Uid      uint64  `form:"uid" validate:"gt=0"`
	Keyword  string  `form:"keyword"`
	Page     int32   `form:"page" validate:"gte=0"`
	PageSize int32   `form:"page_size" default:"20" validate:"gte=0"`
	Lat      float64 `form:"lat" validate:"gte=0"`
	Lng      float64 `form:"lng" validate:"gte=0"`
}
type SearchItemsReq struct {
	Uid      uint64 `form:"uid" validate:"gt=0"`
	Keyword  string `form:"keyword"`
	Page     int32  `form:"page" validate:"gte=0"`
	PageSize int32  `form:"page_size" default:"20" validate:"gte=0"`
}

//
// 返回值
//

type UserReply struct {
	Profile string `json:"profile"`
	Name    string `json:"name"`
	Uid     uint64 `json:"uid"`
}

type TopicReply struct {
	TopicId   uint64 `json:"topic_id"`
	TopicName string `json:"topic_name"`
}

type LocationReply struct {
	Poi string `json:"poi"`
}
type SearchLocationsReply struct {
	Locations []*LocationReply `json:"locations"`
	HasMore   int32            `json:"has_more"`
}

type ItemReply struct {
	Name           string `json:"name"`
	Url            string `json:"url"`
	SchemaUrl      string `json:"schema_url"`
	ItemId         int64  `json:"item_id"`
	SourceType     int32  `json:"source_type"`
	RequiredNumber int    `json:"required_number"`
	Price          string `json:"price"`
	Cover          string `json:"cover"`
	Brief          string `json:"brief"`
	PriceEqual     int    `json:"price_equal"`
}

type SearchAllReply struct {
	Users     []*UserReply     `json:"users"`
	Topics    []*TopicReply    `json:"topics"`
	Locations []*LocationReply `json:"locations"`
	Items     []*ItemReply     `json:"items"`
}
type SearchUsersReply struct {
	Users   []*UserReply `json:"users"`
	HasMore int32        `json:"has_more"`
}
type SearchTopicsReply struct {
	Topics  []*TopicReply `json:"topics"`
	HasMore int32         `json:"has_more"`
}

type SearchItemsReply struct {
	Items   []*ItemReply `json:"items"`
	HasMore int32        `json:"has_more"`
}

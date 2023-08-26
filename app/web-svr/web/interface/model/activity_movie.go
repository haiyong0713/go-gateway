package model

import (
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
)

type ActivityMovieListReq struct {
	NewTopicId     int64                    `form:"new_topic_id"`
	TopicName      string                   `form:"topic_name"`
	PageSize       int32                    `form:"page_size" default:"30" validate:"gte=0"`
	Offset         *dyntopicgrpc.FeedOffset `form:"offset"`
	NewTopicOffset string                   `form:"new_topic_offset"`
}

type ActivityMovieListRsp struct {
	List           []*MovieReviewMeta       `json:"list"`
	HasMore        bool                     `json:"has_more"`
	Offset         *dyntopicgrpc.FeedOffset `json:"offset,omitempty"`
	NewTopicOffset string                   `json:"new_topic_offset,omitempty"`
}

type MovieReviewIntermediate struct {
	Rids           []int64
	Mids           []int64
	DynamicIdMap   map[int64]int64
	HasMore        bool
	Offset         *dyntopicgrpc.FeedOffset
	NewTopicOffset string
}

type MovieReviewMeta struct {
	Author         *MovieReviewAuthor `json:"author"`
	Content        string             `json:"content"`
	DynamicId      int64              `json:"dynamic_id"`
	Score          int                `json:"score"`
	PtimeLabelText string             `json:"ptime_label_text"`
}

type MovieReviewAuthor struct {
	Avatar   string               `json:"avatar"`
	Mid      int64                `json:"mid"`
	Uname    string               `json:"uname"`
	Level    int32                `json:"level"`
	Vip      *MovieReviewVip      `json:"vip"`
	VipLabel *MovieReviewVipLabel `json:"vip_label"`
}

type MovieReviewVipLabel struct {
	LabelTheme string `json:"label_theme"`
	Path       string `json:"path"`
	Text       string `json:"text"`
}

type MovieReviewVip struct {
	ThemeType int32 `json:"theme_type"`
	VipStatus int32 `json:"vip_status"`
	Type      int32 `json:"vip_type"`
}

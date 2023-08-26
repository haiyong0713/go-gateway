package model

import appcard "go-gateway/app/app-svr/app-card/interface/model/card"

type GeneralFeedListReq struct {
	TopicId          int64  `form:"topic_id"`
	PageSize         int64  `form:"page_size" default:"20"`
	Offset           string `form:"offset"`
	SortBy           int64  `form:"sort_by" default:"2"`
	FeedCardType     string `form:"feed_card_type"`
	ShowDynamicTypes string `form:"show_dynamic_types"`
	Business         string `form:"business"`
}

type VideoInlineCard struct {
	*appcard.LargeCoverInline
}

type GeneralFeedListRsp struct {
	TopicCards       []*TopicCardItem   `json:"topic_cards,omitempty"`
	VideoInlineCards []*VideoInlineCard `json:"video_inline_cards,omitempty"`
	VideoCards       []*VideoCard       `json:"video_cards,omitempty"`
	HasMore          bool               `json:"has_more"`
	Offset           string             `json:"offset"`
}

type TopicTimeLineReq struct {
	TopicId  int64 `form:"topic_id" validate:"required"`
	PageSize int32 `form:"ps" default:"20"`
	Offset   int32 `form:"offset"`
}

type TopicTimeLineRsp struct {
	TimeLineId     int64             `json:"time_line_id"`
	TimeLineTitle  string            `json:"time_line_title"`
	HasMore        bool              `json:"has_more"`
	Offset         int32             `json:"offset"`
	TimeLineEvents []*TimeLineEvents `json:"time_line_events"`
}

type TimeLineEvents struct {
	EventId  int64  `json:"event_id"`
	Title    string `json:"title"`
	TimeDesc string `json:"time_desc"`
	JumpLink string `json:"jump_link"`
}

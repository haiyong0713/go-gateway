package model

import (
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
)

type HotWordVideosReq struct {
	TopicId  int64  `form:"topic_id"`
	PageSize int64  `form:"page_size" default:"20"`
	Offset   string `form:"offset"`
}

type HotWordVideosRsp struct {
	VideoCards []*VideoCard `json:"video_cards"`
	HasMore    bool         `json:"has_more"`
	Offset     string       `json:"offset"`
}

type VideoCard struct {
	*appcard.SmallCoverV2
}

type HotWordDynamicReq struct {
	TopicId  int64  `form:"topic_id"`
	PageSize int32  `form:"page_size" default:"20"`
	Offset   string `form:"offset"`
}

type HotWordDynamicRsp struct {
	Items   []*TopicCardItem `json:"items"`
	HasMore bool             `json:"has_more"`
	Offset  string           `json:"offset"`
}

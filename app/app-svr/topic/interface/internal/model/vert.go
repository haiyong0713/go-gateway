package model

import (
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

type VertTopicCenterReq struct {
	PageSize int32  `form:"page_size" default:"20"`
	Offset   string `form:"offset"`
	Source   string `form:"source"` // 来源区分 填"H5"来自H5, 填"Web"来自pc web
}

type VertTopicOnlineReq struct {
	TopicId int64 `form:"topic_id" validate:"required"`
}

type VertSearchTopicsReq struct {
	Keywords string `form:"keywords"`
	PageNum  int32  `form:"page_num" validate:"gte=0"`
	PageSize int32  `form:"page_size" default:"20"`
	Offset   int64  `form:"offset"`
}

type EntranceButton struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
	Link string `json:"link,omitempty"`
}

func ConstructTopicEntranceButton() *EntranceButton {
	// EntranceButton结构目前只会出我的话题
	return &EntranceButton{
		Type: "myTopic",
		Text: "我的话题",
		Link: "https://www.bilibili.com/h5/topic-active/my-topic?navhide=1",
	}
}

type VertTopicCenterRsp struct {
	EntranceButton        *EntranceButton    `json:"entrance_button,omitempty"`
	TopicItems            []*TopicItem       `json:"topic_items"`
	HotTopics             *HotTopics         `json:"hot_topics,omitempty"`
	FavTopics             *FavTopics         `json:"fav_topics,omitempty"`
	PageInfo              *topicsvc.PageInfo `json:"page_info,omitempty"`
	HasCreateJurisdiction bool               `json:"has_create_jurisdiction"`
}

type VertTopicOnlineRsp struct {
	OnlineNum  int64  `json:"online_num,omitempty"`
	OnlineText string `json:"online_text,omitempty"`
}

type HotTopics struct {
	HotItems []*TopicItem `json:"hot_items,omitempty"`
}

type FavTopics struct {
	FavItems []*TopicItem `json:"fav_items,omitempty"`
	MoreLink string       `json:"more_link,omitempty"`
}

type VertSearchTopicsRsp struct {
	TopicItems []*TopicItem               `json:"topic_items"`
	PageInfo   *topiccommon.PaginationRsp `json:"page_info,omitempty"`
}

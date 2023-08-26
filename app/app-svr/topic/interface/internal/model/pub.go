package model

import (
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
)

type SearchPubTopicsReq struct {
	Keywords      string `form:"keywords"`
	PageNum       int32  `form:"page_num" validate:"gte=0"`
	PageSize      int32  `form:"page_size" default:"20"`
	Offset        int64  `form:"offset"`
	UploadId      string `form:"upload_id"`
	FromTopicId   int64  `form:"from_topic_id"`
	FromTopicName string `form:"from_topic_name"`
	Content       string `form:"content"`
	From          string `form:"from"`
}

type UsrPubTopicsReq struct {
	State    int32 `form:"state"` // 话题状态 0:已经上线 1:审核中 -1:已驳回 -2:已下线 -3:ALL
	PageNum  int32 `form:"page_num"`
	PageSize int32 `form:"page_size" default:"20"`
	Offset   int64 `form:"offset"`
}

type IsAlreadyExistedTopicReq struct {
	Topic       string `form:"topic" validate:"required"`
	Description string `form:"description"`
}

type IsAlreadyExistedTopicRsp struct {
	AlreadyExisted bool          `json:"already_existed"`
	SynonymTopic   *SynonymTopic `json:"synonym_topic"`
}

type TopicPubEventsReq struct {
	ShowMids string `form:"show_mids"`
	PubNum   int64  `form:"new_pub_num"`
}

type TopicPubEventsRsp struct {
	ShowText     string        `json:"show_text"`
	ShowMembers  []*ShowMember `json:"show_members"`
	ReqTimestamp int64         `json:"req_timestamp"`
}

type ShowMember struct {
	Mid    int64  `json:"mid"`
	Avatar string `json:"avatar"`
}

type SynonymTopic struct {
	TopicItems []*TopicItem `json:"topic_items,omitempty"`
}

type SearchRcmdPubTopicsReq struct {
	Keywords      string `form:"keywords"`
	UploadId      string `form:"upload_id"`
	FromTopicId   int64  `form:"from_topic_id"`
	FromTopicName string `form:"from_topic_name"`
}

type SearchRcmdPubTopicsRsp struct {
	TopicItems []*SeachRcmdTopicItem `json:"topic_items"`
	RequestId  string                `json:"request_id"`
}

type SeachRcmdTopicItem struct {
	TopicId   int64  `json:"topic_id"`
	TopicName string `json:"topic_name"`
}

type SearchPubTopicsRsp struct {
	NewTopic              SearchNewTopic             `json:"new_topic"`
	HasCreateJurisdiction bool                       `json:"has_create_jurisdiction"`
	TopicItems            []*TopicItem               `json:"topic_items"`
	RequestId             string                     `json:"request_id"`
	PageInfo              *topiccommon.PaginationRsp `json:"page_info"`
}

type SearchNewTopic struct {
	Name  string `json:"name"`
	IsNew bool   `json:"is_new,omitempty"`
}

type UsrPubTopicsRsp struct {
	HasCreateJurisdiction bool                       `json:"has_create_jurisdiction"`
	TopicItems            []*TopicItem               `json:"topic_items,omitempty"`
	PageInfo              *topiccommon.PaginationRsp `json:"page_info"`
}

type PubTopicEndpointReq struct {
	Scene string `form:"scene"` //场景:"dynamic"动态发布入口，"view"视频投稿入口，"topic"话题详情页&我的话题入口
}

type PubTopicEndpointRsp struct {
	MaxCnt                int64             `json:"max_cnt"`
	RemainCnt             int64             `json:"remain_cnt"`
	HasCreateJurisdiction bool              `json:"has_jurisdiction"`
	CenterPlusButton      *CenterPlusButton `json:"center_plus_button,omitempty"`
}

type CenterPlusButton struct {
	IsShow      bool   `json:"is_show"`
	Description string `json:"description"`
}

type PubTopicUploadReq struct {
	TopicId   string `form:"topic_id"`
	UploadId  string `form:"upload_id"`
	RequestId string `form:"request_id"`
}

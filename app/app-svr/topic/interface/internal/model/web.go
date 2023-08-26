package model

import (
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
)

type WebTopicInfoReq struct {
	TopicId int64  `form:"topic_id"`
	Source  string `form:"source"`
}

type WebTopicCardsReq struct {
	TopicId     int64  `form:"topic_id"`
	PageSize    int32  `form:"page_size" default:"20"`
	Offset      string `form:"offset"`
	SortBy      int64  `form:"sort_by"`
	NeedRefresh int32  `form:"need_refresh"`
	Source      string `form:"source"`
}

type WebTopicFoldCardsReq struct {
	TopicId    int64  `form:"topic_id"`
	PageSize   int32  `form:"page_size" default:"20"`
	Offset     string `form:"offset"`
	FromSortBy int64  `form:"from_sort_by"`
}

type WebTopicInfoRsp struct {
	TopDetails     *TopDetails     `json:"top_details,omitempty"`
	FunctionalCard *FunctionalCard `json:"functional_card,omitempty"`
}

type WebDynamicRcmdReq struct {
	PageSize int32 `form:"page_size" default:"6"` // web动态流默认请求6个
}

type WebDynamicRcmdRsp struct {
	TopicItems []*TopicItem `json:"topic_items"`
}

type FunctionalCard struct {
	Capsules    []*TopicCapsule `json:"capsules,omitempty"`
	TrafficCard *TrafficCard    `json:"traffic_card,omitempty"`
	GameCard    *GameCard       `json:"game_card,omitempty"`
}

type GameCard struct {
	GameId   int64  `json:"game_id,omitempty"`
	GameIcon string `json:"game_icon,omitempty"`
	GameName string `json:"game_name,omitempty"`
	Score    string `json:"score,omitempty"`
	GameTags string `json:"game_tags,omitempty"`
	Notice   string `json:"notice,omitempty"`
	GameLink string `json:"game_link,omitempty"`
}

type TopicCapsule struct {
	Name    string `json:"name,omitempty"`
	JumpUrl string `json:"jump_url,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type TrafficCard struct {
	Name         string `json:"name,omitempty"`
	JumpUrl      string `json:"jump_url,omitempty"`
	IconUrl      string `json:"icon_url,omitempty"`
	BasePic      string `json:"base_pic,omitempty"`
	BenefitPoint string `json:"benefit_point,omitempty"`
	CardDesc     string `json:"card_desc,omitempty"`
	JumpTitle    string `json:"jump_title,omitempty"`
}

type OperationContent struct {
	LargeCoverInline *LargeCoverInline `json:"large_cover_inline,omitempty"`
}

type LargeCoverInline struct {
	*appcard.LargeCoverInline
	LiveExtra struct {
		LiveStatus    int8   `json:"live_status,omitempty"`
		LiveStatsDesc string `json:"live_stats_desc,omitempty"`
	} `json:"live_extra,omitempty"`
}

type TopDetails struct {
	TopicItem             *TopicItem        `json:"topic_item"`
	TopicCreator          *TopicCreator     `json:"topic_creator"`
	OperationContent      *OperationContent `json:"operation_content"`
	HasCreateJurisdiction bool              `json:"has_create_jurisdiction"`
	HeadImgUrl            string            `json:"head_img_url,omitempty"`       // 头图图片url
	HeadImgBackcolor      string            `json:"head_img_backcolor,omitempty"` // 头图的主题色蒙层
	WordColor             int32             `json:"word_color"`                   // 文字颜色(0.深色 1.浅色)
	ClosePubLayerEntry    bool              `json:"close_pub_layer_entry"`
}

type TopicCreator struct {
	Uid  int64  `json:"uid"`
	Face string `json:"face"`
	Name string `json:"name"`
}

type WebTopicFoldCardsRsp struct {
	TopicCardList *TopicCardList `json:"topic_card_list"`
}

type WebTopicCardsRsp struct {
	TopicCardList *TopicCardList `json:"topic_card_list"`
	RelatedTopics *RelatedTopics `json:"related_topics"`
}

type RelatedTopics struct {
	TopicItems []*TopicItem `json:"topic_items,omitempty"`
}

type TopicCardList struct {
	Items           []*TopicCardItem `json:"items"`
	Offset          string           `json:"offset"`
	HasMore         bool             `json:"has_more"`
	TopicSortByConf *TopicSortByConf `json:"topic_sort_by_conf,omitempty"`
}

type TopicSortByConf struct {
	DefaultSortBy int64                   `json:"default_sort_by,omitempty"`
	AllSortBy     []*topicsvc.SortContent `json:"all_sort_by,omitempty"`
	ShowSortBy    int64                   `json:"show_sort_by,omitempty"` //当前需要显示的排序方式（1.推荐 2.热门 3.最新）
}

type SortContent struct {
	SortBy   int64  `json:"sort_by,omitempty"`
	SortName string `json:"sort_name,omitempty"`
}

type TopicCardItem struct {
	TopicType           string                    `json:"topic_type"`
	DynamicCardItem     *jsonwebcard.TopicCard    `json:"dynamic_card_item,omitempty"`
	FoldCardItem        *jsonwebcard.TopicCard    `json:"fold_card_item,omitempty"`
	VideoSmallCardItem  *appcard.SmallCoverV2     `json:"video_small_card_item,omitempty"`
	VideoInlineCardItem *appcard.LargeCoverInline `json:"video_inline_card_item,omitempty"`
}

type WebFavSubListReq struct {
	PageNum  int32 `form:"page_num" validate:"gte=0"`
	PageSize int32 `form:"page_size" default:"20"`
}

type WebFavSubListRsp struct {
	TopicList *TopicFavList `json:"topic_list,omitempty"`
}

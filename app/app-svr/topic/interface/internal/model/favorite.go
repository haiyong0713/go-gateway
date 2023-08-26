package model

type FavSubListReq struct {
	From     string `form:"from"`
	PageNum  int32  `form:"page_num" validate:"gte=0"`
	PageSize int32  `form:"page_size" default:"20"`
	Offset   int64  `form:"offset"`
}

type FavSubListRsp struct {
	FavTab    *FavTab       `json:"fav_tab,omitempty"`
	TopicList *TopicFavList `json:"topic_list,omitempty"`
	TagList   *TagFavList   `json:"tag_list,omitempty"`
}

type TopicFavList struct {
	TopicItems []*TopicItem   `json:"topic_items,omitempty"`
	PageInfo   *PaginationRsp `json:"page_info,omitempty"`
}

type TagFavList struct {
	TagItems []*TagItem     `json:"tag_items,omitempty"`
	PageInfo *PaginationRsp `json:"page_info,omitempty"`
}

type PaginationRsp struct {
	CurPageNum int32 `json:"cur_page_num,omitempty"`
	Offset     int64 `json:"offset,omitempty"`
	HasMore    bool  `json:"has_more,omitempty"`
	Total      int32 `json:"total,omitempty"`
}

type FavTab struct {
	Topic bool `json:"topic"`
	Tag   bool `json:"tag"`
}

type AddFavReq struct {
	TopicId    int64 `form:"topic_id"`
	TopicSetId int64 `form:"topic_set_id"`
}

type CancelFavReq struct {
	TopicId    int64 `form:"topic_id"`
	TopicSetId int64 `form:"topic_set_id"`
}

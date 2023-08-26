package model

type CreateTopicReq struct {
	TopicName       string `form:"topic_name" validate:"required"`  //话题名称
	Description     string `form:"description" validate:"required"` //说明
	Scene           string `form:"scene"`                           //场景:"dynamic"动态发布入口，"view"视频投稿入口，"topic"话题详情页&我的话题入口
	SubmitTopicType int32  `form:"submit_topic_type"`               //话题类型 0-普通类型 1-视频类型(跟我拍)
}

type CreateTopicRsp struct {
	TopicId     int64  `json:"topic_id"`               //话题id
	TopicName   string `json:"topic_name"`             //话题名称
	SuccessDesc string `json:"success_desc,omitempty"` //发布成功说明
}

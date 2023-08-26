package topic

import (
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

type NewHotTopicDetail struct {
	TopicDetail       *topicsvc.TopicDetail
	DynamicResourceId int64 // 动态资源id
}

package dynamicV2

import (
	"fmt"
	"strings"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicV2 "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (list *DynListRes) FromSearch(search *dyngrpc.SearchRsp, uid int64) {
	list.HasMore = search.HasMore
	var logs []string
	for _, item := range search.DynList {
		if item == nil || item.Type == 0 {
			log.Warn("FromSearch miss FromSearch mid %v, item %+v", uid, item)
			continue
		}
		if item.Type == 1 && item.Origin == nil {
			log.Warn("FromSearch miss forward origin nil mid %v, item %+v", uid, item)
			continue
		}
		logs = append(logs, fmt.Sprintf("dynid(%v) type(%v) rid(%v)", item.DynId, item.Type, item.Rid))
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
	log.Warn("FromSearch(new) origin mid(%d) list(%v)", uid, strings.Join(logs, "; "))
}

// 动态垂搜-话题搜索结果
type DynSearchTopicResult interface {
	ToDynV2SearchTopic(req *api.DynSearchReq) *api.SearchTopic
	ShowChannelSubCards() bool // 新话题有搜索结果时不展示频道小卡
}

type OldTopicSearchResImpl struct {
	Res []*dyntopicgrpc.SearchResult
}

func (otsi *OldTopicSearchResImpl) ShowChannelSubCards() bool {
	// 老话题固定展示
	return true
}

func (otsi *OldTopicSearchResImpl) ToDynV2SearchTopic(req *api.DynSearchReq) *api.SearchTopic {
	if otsi == nil || len(otsi.Res) == 0 {
		return nil
	}
	res := &api.SearchTopic{
		Title: "话题",
		MoreButton: &api.SearchTopicButton{
			Title:   "查看更多相关话题",
			JumpUri: model.FillURI(model.GotoTopicSearch, req.GetKeyword(), nil),
		},
	}
	for _, v := range otsi.Res {
		item := &api.SearchTopicItem{
			TopicId:   v.TopicId,
			TopicName: v.TopicName,
			Url:       v.TopicLink,
		}
		//是否是活动 0不是 1是
		if v.IsActivity == 1 {
			item.IsActivity = true
		}
		var labels []string
		labels = append(labels, model.StatString(v.DynCnt, "条动态"))
		labels = append(labels, model.StatString(v.FansCnt, "人关注"))
		if len(labels) > 0 {
			item.Desc = strings.Join(labels, "  ")
		}
		res.Items = append(res.Items, item)
	}
	return res
}

type NewTopicSearchResImpl struct {
	Res *topicV2.VertSearchTopicInfoV2Rsp
}

func (ntsi *NewTopicSearchResImpl) ToDynV2SearchTopic(_ *api.DynSearchReq) *api.SearchTopic {
	if ntsi == nil || ntsi.Res == nil || len(ntsi.Res.GetTopics()) == 0 {
		return nil
	}
	topics := ntsi.Res
	res := &api.SearchTopic{
		Title: topics.ModuleName,
	}
	if len(topics.MoreJumpUrl) > 0 {
		res.MoreButton = &api.SearchTopicButton{
			Title:   "查看更多相关话题",
			JumpUri: topics.MoreJumpUrl,
		}
		if len(topics.MoreTxt) > 0 {
			res.MoreButton.Title = topics.MoreTxt
		}
	}
	for idx, topic := range topics.Topics {
		res.Items = append(res.Items, &api.SearchTopicItem{
			TopicId:   topic.GetInfo().GetId(),
			TopicName: topic.GetInfo().GetName(),
			Desc:      "", // 留空 后续用作数量展示
			Url:       topic.GetInfo().GetJumpUrl(),
			TagIcon:   topic.GetIconUrl(),
			Cover:     topic.GetTopicCover(),
			DescLong:  topic.GetInfo().GetDescription(),
			TagText:   topic.GetIconText(),
		})
		// 有数据时展示
		if topic.GetInfo().GetView() != 0 && topic.GetInfo().GetDiscuss() != 0 {
			labels := make([]string, 0, 2)
			labels = append(labels, model.StatString(topic.GetInfo().GetView(), "浏览"))
			labels = append(labels, model.StatString(topic.GetInfo().GetDiscuss(), "讨论"))
			res.Items[idx].Desc += strings.Join(labels, "  ")
		}
	}

	return res
}

func (ntsi *NewTopicSearchResImpl) ShowChannelSubCards() bool {
	if ntsi == nil {
		return true
	}
	// 新话题有搜索结果时不展示频道小卡
	return len(ntsi.Res.GetTopics()) == 0
}

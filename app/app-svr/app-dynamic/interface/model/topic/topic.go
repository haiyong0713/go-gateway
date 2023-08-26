package topic

import (
	"strconv"

	"go-gateway/app/app-svr/app-dynamic/interface/model"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	activitygrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

type SquareReq struct {
	Build     string `form:"build"`
	Platform  string `form:"platform"`
	MobiApp   string `form:"mobi_app"`
	Device    string `form:"device"`
	FromSpmid string `form:"from_spmid"`
	Version   string `form:"version"`
}

type SquareReply struct {
	LaunchedActivity *LaunchedActivity `json:"launched_activity"` // 是否显示【发起活动】
	Subscription     *Subscription     `json:"subscription"`      // 我的关注
	Recommend        *Recommend        `json:"recommend"`         // 推荐活动
}

type LaunchedActivity struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// 我的关注
type Subscription struct {
	Title string              `json:"title"`
	Top   []*SubscriptionItem `json:"top"`
	Card  []*SubscriptionItem `json:"card"`
}

// 我的关注元素
type SubscriptionItem struct {
	Icon       string `json:"icon"`
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	IsActivity bool   `json:"is_activity"`
	SubType    string `json:"sub_type"`
}

func (s *SubscriptionItem) FormSubscriptionItem(cc *channelgrpc.ChannelCard, actInfos map[int64]*activitygrpc.NativePage) {
	s.Icon = ""
	s.ID = cc.GetChannelId()
	s.Name = cc.GetChannelName()
	switch cc.Ctype {
	case model.NewChannel:
		s.URL = model.FillURI(model.GotoChannel, strconv.FormatInt(cc.GetChannelId(), 10), nil)
	case model.OldChanne:
		// 优先级 活动跳链模式 > 活动普通模式 > 旧频道
		s.URL = model.FillURI(model.GotoTag, strconv.FormatInt(cc.GetChannelId(), 10), nil)
		// 如果是活动话题逻辑
		if cc.GetActAttr() == 1 {
			s.IsActivity = true
			if actInfo, ok := actInfos[cc.GetChannelId()]; ok && actInfo != nil {
				s.URL = model.FillURI(model.GotoActivity, strconv.FormatInt(actInfo.ID, 10), nil)
				if actInfo.SkipURL != "" {
					s.URL = actInfo.SkipURL
				}
			}
		}
	}
}

// 推荐活动
type Recommend struct {
	Title string           `json:"title"`
	List  []*RecommendItem `json:"list"`
}

type RecommendItem struct {
	Author     *RecommendItemAuthor `json:"author"`
	Cover      *RecommendItemCover  `json:"cover"`
	Topic      *RecommendItemTopic  `json:"topic"`
	DefauleURL string               `json:"default_url"`
	Type       int64                `json:"type"`
	Rid        int64                `json:"rid"`
	Mid        int64                `json:"mid"`
}

type RecommendItemAuthor struct {
	Name   string `json:"name"`
	Mid    int64  `json:"mid"`
	Face   string `json:"face"`
	URL    string `json:"url"`
	Suffix string `json:"suffix"`
}

type RecommendItemCover struct {
	Cover  string            `json:"cover"`
	URL    string            `json:"url"`
	Labels map[string]string `json:"labels"`
}

type RecommendItemTopic struct {
	Icon  string `json:"icon"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	IsSub bool   `json:"is_sub"`
	Desc  string `json:"desc"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

/*
热门推荐
*/
type HotListReq struct {
	Build       string `form:"build"`
	Platform    string `form:"platform"`
	MobiApp     string `form:"mobi_app"`
	Device      string `form:"device"`
	FromSpmid   string `form:"from_spmid"`
	Version     string `form:"version"`
	HotListType int32  `form:"hotlist_type" default:"1"` // 1.全部 2.推荐 3.投稿 4.话题讨论
	Offset      int64  `form:"offset"`
	PageSize    int32  `form:"page_size" default:"20" validate:"min=1"`
}

// 热门活动中心
type HotListReply struct {
	Tab     []*HotListTab  `json:"tab"`
	List    []*HotListItem `json:"list"`
	Offset  int64          `json:"offset"`
	HasMore bool           `json:"has_more"`
}

type HotListTab struct {
	Name   string `json:"name"`
	TypeID int32  `json:"type_id"` // 1.全部 2.推荐 3.投稿 4.话题讨论
}

type HotListItem struct {
	Icon  string `json:"icon"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Label string `json:"label"`
	URL   string `json:"url"`
	Cover string `json:"cover"`
}

// 保存排序
type SubscribeSaveReq struct {
	Top    string `form:"top"`
	Card   string `form:"card"`
	Action int32  `form:"action"`
}

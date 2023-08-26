package show

import (
	"fmt"

	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"

	listenerChannelgrpc "git.bilibili.co/bapis/bapis-go/car-channel/interface"
)

const (
	_chidHuawei = "huawei"
)

type ShowParam struct {
	model.DeviceInfo
	ChannelType     string `form:"channel_type"`
	FollowType      string `form:"follow_type"`
	LoginEvent      int    `form:"login_event"`
	DynUpdateNumber int64  `form:"dyn_update_number"`
}

type Item struct {
	Type         string          `json:"type,omitempty"`
	Title        string          `json:"title,omitempty"`
	RightButton  *RightButton    `json:"right_button,omitempty"`
	Items        []cardm.Handler `json:"items,omitempty"`
	UpdateNumber int64           `json:"update_number,omitempty"`
}

type Config struct {
	DynUpdateNumber int64 `json:"dyn_update_number,omitempty"`
}

type Tab struct {
	ChannelID   int64  `json:"channel_id,omitempty"`
	ChannelName string `json:"channel_name,omitempty"`
	URI         string `json:"uri,omitempty"`
	IsDefault   bool   `json:"is_default,omitempty"`
}

type RightButton struct {
	Title string     `json:"title,omitempty"`
	URI   string     `json:"uri,omitempty"`
	Icon  model.Icon `json:"icon,omitempty"`
}

type AudioShow struct {
	TabItems []*TabItem      `json:"tab_items,omitempty"`
	Items    []cardm.Handler `json:"items,omitempty"`
}

type TabItem struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
	Desc      string `json:"desc,omitempty"`
	Cover     string `json:"cover,omitempty"`
	Tabs      []*Tab `json:"tabs,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"`
	ChannelID int64  `json:"channel_id,omitempty"`
}

func (i *Item) FromItem() {
	switch i.Type {
	case "popular":
		i.Title = "热门"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntrancePopular),
		}
	case "my_bangumi":
		i.Title = "我的追番"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?followType=bangumi&sourceType=%s", model.EntranceMyAnmie),
		}
	case "bangumi":
		i.Title = "番剧推荐"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?followType=bangumi&sourceType=%s", model.EntrancePgcList),
		}
	case "domestic":
		i.Title = "国创推荐"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?followType=domestic&sourceType=%s", model.EntrancePgcList),
		}
	case "my_cinema":
		i.Title = "我的追剧"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?followType=cinema&sourceType=%s", model.EntranceMyAnmie),
		}
	case "cinema":
		i.Title = "电影热播"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?followType=cinema&sourceType=%s", model.EntrancePgcList),
		}
	case "cinema_doc":
		i.Title = "纪录片热播"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?followType=cinema_doc&sourceType=%s", model.EntrancePgcList),
		}
	case "dynamic_video":
		i.Title = "我的关注"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntranceDynamicVideo),
		}
	case "history":
		i.Title = "继续观看"
		i.RightButton = &RightButton{
			Title: "更多",
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntranceHistoryRecord),
		}
	}
	// 小于20不要更多的按钮
	// nolint:gomnd
	if len(i.Items) < 10 {
		i.RightButton = nil
	}
}

// 1.1版本新的文案
func (i *Item) FromItem2(chid string) {
	switch i.Type {
	case "feed":
		if chid == _chidHuawei { // show时判断
			i.Title = "精选"
		} else {
			i.Title = "为你推荐"
		}
		i.Type = "common-search"
	case "popular":
		i.Title = "热门"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntrancePopular),
		}
	case "my_bangumi":
		i.Title = "我的追番"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?followType=bangumi&sourceType=%s", model.EntranceMyAnmie),
		}
	case "bangumi":
		i.Title = "番剧推荐"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?followType=bangumi&sourceType=%s", model.EntrancePgcList),
		}
		i.Type = "partition"
	case "domestic":
		i.Title = "国创推荐"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?followType=domestic&sourceType=%s", model.EntrancePgcList),
		}
		i.Type = "partition"
	case "my_cinema":
		i.Title = "我的追剧"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?followType=cinema&sourceType=%s", model.EntranceMyAnmie),
		}
	case "cinema":
		i.Title = "电影热播"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?followType=cinema&sourceType=%s", model.EntrancePgcList),
		}
		i.Type = "partition"
	case "cinema_doc":
		i.Title = "纪录片热播"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?followType=cinema_doc&sourceType=%s", model.EntrancePgcList),
		}
		i.Type = "partition"
	case "dynamic_video":
		i.Title = "我的关注"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntranceDynamicVideo),
		}
	case "history":
		i.Title = "继续观看"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntranceHistoryRecord),
		}
	case "my_favorite":
		i.Title = "我的收藏夹"
		i.RightButton = &RightButton{
			Title: "更多",
			Icon:  model.IconShow,
			URI:   fmt.Sprintf("bilithings://favorite/second?sourceType=%s", model.EntranceMyFavorite),
		}
	case "region_3":
		i.Title = "音乐"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 3),
		}
		i.Type = "partition"
	case "region_129":
		i.Title = "舞蹈"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 129),
		}
		i.Type = "partition"
	case "region_4":
		i.Title = "游戏"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 4),
		}
		i.Type = "partition"
	case "region_36":
		i.Title = "知识"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 36),
		}
		i.Type = "partition"
	case "region_202":
		i.Title = "资讯"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 202),
		}
		i.Type = "partition"
	case "region_223":
		i.Title = "汽车"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 223),
		}
		i.Type = "partition"
	case "region_160":
		i.Title = "生活"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, 160),
		}
		i.Type = "partition"
	case fmt.Sprintf("region_%d", model.CustomModuleRid51):
		i.Title = "五一特辑"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, model.CustomModuleRid51),
		}
		i.Type = "partition"
	case fmt.Sprintf("region_%d", model.CustomModuleRid61Childhood):
		i.Title = "童年回来了"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, model.CustomModuleRid61Childhood),
		}
		i.Type = "partition"
	case fmt.Sprintf("region_%d", model.CustomModuleRid61Eden):
		i.Title = "小朋友乐园"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, model.CustomModuleRid61Eden),
		}
		i.Type = "partition"
	case fmt.Sprintf("region_%d", model.CustomModuleRidDW):
		i.Title = "“粽”有陪伴"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, model.CustomModuleRidDW),
		}
		i.Type = "partition"
	case "top_view":
		i.Title = "稍后再看"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntranceToView),
		}
	case "dynamic_video_new":
		i.Title = "最新更新"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", model.EntranceDynamicVideoNew),
		}
		i.Type = "new"
	}
	// 小于20不要播放全部的按钮
	// nolint:gomnd
	if len(i.Items) < 10 {
		i.RightButton = nil
	}
}

func (i *Item) FromFavItem(id, vmid int64) {
	switch i.Type {
	case model.EntranceToView:
		// 稍后再看
		i.Title = "稍后再看"
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s", i.Type),
		}
	default:
		// 否则播单列表
		i.RightButton = &RightButton{
			Title: "播放全部",
			Icon:  model.IconPlay,
			URI:   fmt.Sprintf("bilithings://player?sourceType=%s&fav_id=%d&vmid=%d", model.EntranceMediaList, id, vmid),
		}
	}
	// 小于10不要更多的按钮
	// nolint:gomnd
	if len(i.Items) < 10 {
		i.RightButton = nil
	}
}

func (i *TabItem) FromAudioItem() bool {
	switch i.Type {
	case "auido_history":
		i.Title = "最近播放"
		i.Desc = "Recently played"
		i.Cover = "http://i0.hdslb.com/bfs/feed-admin/dddf76c98468d2ec1025ae319f137e4dec0f8280.png"
	case "audio_feed":
		i.Title = "为你推荐"
		i.Desc = "Recommended"
		i.Cover = "http://i0.hdslb.com/bfs/feed-admin/e4b0069fb6c71b927ec98f820ec2e43efcd57a5b.png"
		i.IsDefault = true
	}
	return i.IsDefault
}

func (i *TabItem) FromChannel(isDefault bool) bool {
	switch i.Type {
	case "499":
		i.Desc = "Music"
		i.Cover = "http://i0.hdslb.com/bfs/feed-admin/85bee324b4f577308bbea47d0a76a7c6e1cbd669.png"
	case "10009":
		i.Desc = "Knowledge"
		i.Cover = "http://i0.hdslb.com/bfs/feed-admin/acca1bbba3ef968087ef5e54e40ce82f90a2818e.png"
	}
	if !isDefault {
		i.IsDefault = true
		return true
	}
	return isDefault
}

func (i *TabItem) FromAudioTabs(chls []*listenerChannelgrpc.ChannelRecommendInfo, topChannel *listenerChannelgrpc.ChannelRecommendInfo, isDefault bool) {
	// 默认第一个是一级tab数据
	tab := &Tab{ChannelID: topChannel.Id, ChannelName: "全部"}
	if !isDefault {
		tab.IsDefault = true
	}
	tabs := []*Tab{tab}
	for _, v := range chls {
		t := &Tab{
			ChannelID:   v.Id,
			ChannelName: v.Name,
		}
		tabs = append(tabs, t)
	}
	i.Tabs = tabs
}

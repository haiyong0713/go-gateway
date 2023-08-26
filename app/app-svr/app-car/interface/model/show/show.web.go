package show

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
)

type ItemWeb struct {
	Type           string          `json:"type,omitempty"`
	Title          string          `json:"title,omitempty"`
	RightButtonWeb *RightButtonWeb `json:"right_button,omitempty"`
	Items          []cardm.Handler `json:"items,omitempty"`
}

type RightButtonWeb struct {
	Title      string     `json:"title,omitempty"`
	SourceType string     `json:"source_type,omitempty"`
	Rid        int64      `json:"rid,omitempty"`
	Icon       model.Icon `json:"icon,omitempty"`
	FavID      int64      `json:"fav_id,omitempty"`
	Vmid       int64      `json:"vmid,omitempty"`
}

// 1.1版本新的文案
func (i *ItemWeb) FromItemWeb(entrance string, rid int64) {
	var isHide bool
	i.RightButtonWeb = &RightButtonWeb{
		Title:      "播放全部",
		Icon:       model.IconPlay,
		Rid:        rid,
		SourceType: entrance,
	}
	switch i.Type {
	case "feed":
		i.Title = "为你推荐"
		isHide = true
	case "popular":
		i.Title = "热门"
	case "my_bangumi":
		i.Title = "我的追番"
	case "bangumi":
		i.Title = "番剧推荐"
	case "domestic":
		i.Title = "国创推荐"
	case "my_cinema":
		i.Title = "我的追剧"
	case "cinema":
		i.Title = "电影热播"
	case "cinema_doc":
		i.Title = "纪录片热播"
	case "dynamic_video":
		i.Title = "我的关注"
	case "history":
		i.Title = "继续观看"
	case "region_3":
		i.Title = "音乐"
	case "region_129":
		i.Title = "舞蹈"
	case "region_4":
		i.Title = "游戏"
	case "region_36":
		i.Title = "知识"
	case "region_202":
		i.Title = "资讯"
	case "my_favorite":
		i.Title = "我的收藏夹"
		i.RightButtonWeb = &RightButtonWeb{
			Title: "更多",
			Icon:  model.IconShow,
		}
	case "top_view":
		i.Title = "稍后再看"
	}
	if i.RightButtonWeb != nil {
		i.RightButtonWeb.SourceType = entrance
	}

	// 小于20不要播放全部的按钮
	if len(i.Items) < 10 || isHide {
		i.RightButtonWeb = nil
	}
}

func (i *ItemWeb) FromFavItemWeb(id, vmid int64, entrance string) {
	switch i.Type {
	case model.EntranceToView:
		// 稍后再看
		i.Title = "稍后再看"
	}
	i.RightButtonWeb = &RightButtonWeb{
		Title:      "播放全部",
		Icon:       model.IconPlay,
		FavID:      id,
		Vmid:       vmid,
		SourceType: entrance,
	}
	// 小于10不要更多的按钮
	// nolint:gomnd
	if len(i.Items) < 10 {
		i.RightButtonWeb = nil
	}
}

package common

import (
	"strconv"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-car/interface/model"
	bangumimdl "go-gateway/app/app-svr/app-car/interface/model/bangumi"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

type Item struct {
	ItemType       ItemType    `json:"item_type"`
	ItemId         int64       `json:"item_id"` // 根据item_type不同，含义不同(合集 -> 合集id； 频道 -> 频道id)
	Otype          Otype       `json:"otype"`   // ugc/pgc/live  v2.4开始使用
	Oid            int64       `json:"oid"`
	Cid            int64       `json:"cid"`
	Url            string      `json:"url"`
	Title          string      `json:"title"`
	Cover          string      `json:"cover"`
	LandscapeCover string      `json:"landscape_cover,omitempty"`
	Author         *Author     `json:"author,omitempty"`
	Badge          *Badge      `json:"badge,omitempty"`
	PlayCount      int         `json:"play_count"`
	DanmakuCount   int         `json:"danmaku_count"`
	FavCount       int         `json:"fav_count"`
	ReplyCount     int         `json:"reply_count"`
	Duration       int64       `json:"duration"`
	Desc           string      `json:"desc,omitempty"`
	Playlist       []*Playlist `json:"play_list,omitempty"`
	Pubtime        xtime.Time  `json:"pubtime,omitempty"`
	// 是否收藏/追番
	IsFollow  bool   `json:"is_follow"`
	IndexShow string `json:"index_show,omitempty"`
	// 是否点赞
	IsLike bool `json:"is_like"`
	// 历史记录结构
	*ItemHistory
	// 详情结构
	*View
	SubTitle     string   `json:"sub_title"`
	Wrapper      *Wrapper `json:"wrapper,omitempty"`
	ArcCountShow string   `json:"arc_count_show"` // 集数展示，如："共5集"
	HotRate      int64    `json:"hot_rate"`       // 热度
	ServerInfo   string   `json:"server_info"`    // 推荐算法测透传的数据（包含 ab实验、排序分、召回策略等数据），客户端获取后，直接加到ubt埋点数据里即可
	ShowType     int      `json:"show_type"`      // FM金刚位露出类型：0-固定配置卡 1-内容透出卡
	Label        *Badge   `json:"label"`          // feed流"最近更新"标签
	Catalog      *Catalog `json:"catalog"`        // 分区信息
	Score        string   `json:"score"`          // ogv评分
}

type Catalog struct {
	CatalogId    int64  `json:"catalog_id"`
	CatalogName  string `json:"catalog_name"`
	CatalogPid   int64  `json:"catalog_pid"`
	CatalogPname string `json:"catalog_pname"`
}

type Wrapper struct {
	Id    int64  `json:"id"`
	Type  int    `json:"type"`
	Cover string `json:"cover"`
	Title string `json:"title"`
}

type FmItem struct {
	*Item
	MiniTitle string `json:"mini_title,omitempty"` // mini播控栏标题
}

type Badge struct {
	Text             string `json:"text"`
	TextColorDay     string `json:"text_color_day,omitempty"`
	TextColorNight   string `json:"text_color_night,omitempty"`
	BgColorDay       string `json:"bg_color_day,omitempty"`
	BgColorNight     string `json:"bg_color_night,omitempty"`
	BorderColorDay   string `json:"border_color_day,omitempty"`
	BorderColorNight string `json:"border_color_night,omitempty"`
	BgStyle          string `json:"bg_style"` // fill-填充、stroke-描边、fill_stroke-填充+描边、none-背景不填充 + 背景不描边
}

type Author struct {
	Mid       int64     `json:"mid"`
	Name      string    `json:"name"`
	Face      string    `json:"face"`
	FansCount int64     `json:"fans_count"`
	Relation  *Relation `json:"relation"`
}

type Playlist struct {
	Title      string     `json:"title"`
	Aid        int64      `json:"aid"`
	Cid        int64      `json:"cid"`
	Epid       int64      `json:"ep_id"`
	Duration   int64      `json:"duration"`
	LongTitle  string     `json:"long_title,omitempty"`
	Badge      *Badge     `json:"badge,omitempty"`
	ShareURL   string     `json:"share_url,omitempty"`
	ReplyCount int        `json:"reply_count,omitempty"`
	Dimension  *Dimension `json:"dimension,omitempty"`
	Cover      string     `json:"cover,omitempty"`
}

type Dimension struct {
	Height int64 `json:"height,omitempty"`
	Width  int64 `json:"width,omitempty"`
	Rotate int64 `json:"rotate,omitempty"`
}

type ItemHistory struct {
	Business string `json:"business"`
	ViewAt   int64  `json:"view_at"`
	Progress int64  `json:"progress"`
	Max      int64  `json:"max"`
}

type Introduction struct {
	Title  string `json:"title,omitempty"`
	Info   string `json:"info,omitempty"`
	Desc   string `json:"desc,omitempty"`
	Rating string `json:"rating,omitempty"`
	Cover  string `json:"cover,omitempty"`
}

type View struct {
	// 简介页面
	Introduction *Introduction `json:"introduction,omitempty"`
	SeasonType   int           `json:"season_type,omitempty"`
	PagesStyle   string        `json:"pages_style,omitempty"`
	History      *History      `json:"history,omitempty"`
	PayInfo      *PayInfo      `json:"pay_info,omitempty"`
}

type Relation struct {
	Status     int32 `json:"status,omitempty"`
	IsFollow   int32 `json:"is_follow,omitempty"`
	IsFollowed int32 `json:"is_followed,omitempty"`
}

func RelationChange(upMid int64, relations map[int64]*relationgrpc.InterrelationReply) (r *Relation) {
	const (
		// state使用
		_statenofollow      = 1
		_statefollow        = 2
		_statefollowed      = 3
		_statemutualConcern = 4
		// 关注关系
		_follow = 1
	)
	r = &Relation{
		Status: _statenofollow,
	}
	rel, ok := relations[upMid]
	if !ok {
		return
	}
	switch rel.Attribute {
	case 2, 6: // nolint:gomnd
		// 用户关注UP主
		r.Status = _statefollow
		r.IsFollow = _follow
	}
	if rel.IsFollowed { // UP主关注用户
		r.Status = _statefollowed
		r.IsFollowed = _follow
	}
	if r.IsFollow == _follow && r.IsFollowed == _follow { // 用户和UP主互相关注
		r.Status = _statemutualConcern
	}
	return
}

type History struct {
	Cid      int64 `json:"cid,omitempty"`
	Epid     int64 `json:"ep_id,omitempty"`
	Progress int64 `json:"progress,omitempty"`
	ViewAt   int64 `json:"view_at,omitempty"`
}

type PayInfo struct {
	Ptype      int    `json:"p_type"`
	Icon       string `json:"icon"`
	Text       string `json:"text"`
	ButtonText string `json:"button_text"`
	SubText    string `json:"sub_text"`
}

func (i *Introduction) FromOGVIntroduction(b *bangumimdl.View) {
	i.Title = b.Title
	i.Cover = b.Cover
	splice1 := "\n"
	splice2 := " "
	splice3 := " | "
	var (
		desc1, desc2, desc3 string
	)
	if b.Publish != nil {
		// 发布时间
		if pubData, err := model.TimeToUnix(b.Publish.PubTime); err == nil {
			desc1 = model.ViewInfo(desc1, pubData.Format("2006"), splice2)
		}
	}
	// PGC类型
	if pgcType, err := model.PGCTypeValue(b.Type); err == nil {
		desc1 = model.ViewInfo(desc1, pgcType, splice2)
	}
	// 地区
	if len(b.Areas) > 0 {
		desc1 = model.ViewInfo(desc1, b.Areas[0].Name, splice3)
	}
	// 是否完结
	if b.Publish != nil {
		if b.Publish.IsFinish == 1 {
			desc2 = model.ViewInfo(desc2, model.BangumiTotalCountString(strconv.Itoa(b.Total), b.Publish.IsFinish), splice2)
		} else {
			if b.NewEP != nil {
				desc2 = model.ViewInfo(desc2, model.BangumiTotalCountString(b.NewEP.Title, b.Publish.IsFinish), splice2)
			}
		}
	}
	// 播放信息
	if b.Stat != nil {
		desc3 = model.ViewInfo(desc3, model.StatString64(b.Stat.Views, "播放"), splice2)
		if b.Stat.Followers != "" {
			// 有系列时，返回系列追番量，没有，则显示season追番量
			desc3 = model.ViewInfo(desc3, b.Stat.Followers, splice2)
		}
	}
	// 第一行
	i.Info = model.ViewInfo(i.Info, desc1, splice1)
	// 第二行
	i.Info = model.ViewInfo(i.Info, desc2, splice1)
	// 第三行
	i.Info = model.ViewInfo(i.Info, desc3, splice1)
	if b.Rating != nil {
		i.Rating = model.BangumiRating(b.Rating.Score, "")
	}
}

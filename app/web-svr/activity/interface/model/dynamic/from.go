package dynamic

import (
	"fmt"
	"strconv"
	"strings"

	arccli "go-gateway/app/app-svr/archive/service/api"
	favmdl "go-main/app/community/favorite/service/model"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/web-svr/activity/interface/api"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
)

// Item .
type Item struct {
	Goto   string `json:"goto,omitempty"`
	Param  string `json:"param,omitempty"`
	ItemID int64  `json:"item_id,omitempty"`
	Ukey   string `json:"ukey,omitempty"`
	// click
	Width    int64   `json:"width,omitempty"`
	Length   int64   `json:"length,omitempty"`
	Image    string  `json:"image,omitempty"`
	Leftx    int64   `json:"leftx,omitempty"`
	Lefty    int64   `json:"lefty,omitempty"`
	URI      string  `json:"uri,omitempty"`
	Content  string  `json:"content,omitempty"`
	Subtitle string  `json:"subtitle,omitempty"`
	Title    string  `json:"title,omitempty"`
	Item     []*Item `json:"item,omitempty"`
	// 动态卡片接口 数据透传
	DyCard *DyCard `json:"dy_card,omitempty"`
	//点赞相关信息
	Liked          *Liked        `json:"liked,omitempty"`
	IsGap          int32         `json:"is_gap,omitempty"`
	IsFeed         int64         `json:"is_feed,omitempty"`
	Button         *Button       `json:"button,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 string        `json:"cover_left_icon_1,omitempty"`
	CoverLeftText2 string        `json:"cover_left_text_2,omitempty"`
	CoverLeftIcon2 string        `json:"cover_left_icon_2,omitempty"`
	CoverRightText string        `json:"cover_right_text,omitempty"`
	Badge          *ReasonStyle  `json:"badge,omitempty"`
	Repost         *Repost       `json:"repost,omitempty"`
	Fid            int64         `json:"fid,omitempty"`
	Duration       string        `json:"duration,omitempty"`
	Danmaku        string        `json:"danmaku,omitempty"`
	View           string        `json:"view,omitempty"`
	Dimension      *Dimension    `json:"dimension,omitempty"`
	ResourceInfo   *ResourceInfo `json:"resource_info,omitempty"`
	Stime          int64         `json:"stime,omitempty"`
}

type Button struct {
	FollowText    string `json:"follow_text,omitempty"`
	FollowIcon    string `json:"follow_icon,omitempty"`
	IsFollow      int32  `json:"is_follow,omitempty"`
	UnFollowText  string `json:"un_follow_text,omitempty"`
	UnFollowIcon  string `json:"un_follow_icon,omitempty"`
	FollowToast   string `json:"follow_toast,omitempty"`
	UnFollowToast string `json:"un_follow_toast,omitempty"`
}

type ResourceInfo struct {
	Up       string `json:"up,omitempty"`
	View     string `json:"view,omitempty"`
	PubTime  string `json:"pub_time,omitempty"`
	Like     string `json:"like,omitempty"`
	Danmaku  string `json:"danmaku,omitempty"`
	Duration string `json:"duration,omitempty"`
	Follow   string `json:"follow,omitempty"`
	Season   string `json:"season,omitempty"`
}

type Dimension struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
	Rotate int64 `json:"rotate"`
}

type Repost struct {
	BizType    string `json:"biz_type"`
	SeasonType string `json:"season_type"`
}

type ReasonStyle struct {
	Text         string `json:"text,omitempty"`
	BgColor      string `json:"bg_color,omitempty"`
	BgColorNight string `json:"bg_color_night,omitempty"`
}

type Liked struct {
	Sid          int64 `json:"sid"`
	Lid          int64 `json:"lid"`
	Score        int64 `json:"score"`
	HasLiked     int64 `json:"has_liked"`
	DisplayScore bool  `json:"display_score"`
}

func (i *Item) FromVideoLike(card *DyCard, itemObj *lmdl.ItemObj) {
	i.Goto = GotoVideoLike
	i.DyCard = card
	i.Liked = &Liked{Sid: itemObj.Item.Sid, Lid: itemObj.Item.ID, HasLiked: itemObj.HasLiked}
	if itemObj.Score == -1 {
		i.Liked.DisplayScore = false
	} else {
		i.Liked.DisplayScore = true
		i.Liked.Score = itemObj.Score
	}

}

func (i *Item) FromVideoMore() {
	i.Goto = GotoVideoMore
	i.Title = "查看更多"
}

func (i *Item) FromVideo(card *DyCard) {
	i.Goto = GotoVideo
	i.DyCard = card
}

func (i *Item) FromUgcVideo(c *arccli.Arc, bvid string) {
	i.Goto = GotoNewUgcVideo
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	i.Param = bvid
	i.URI = fmt.Sprintf("bilibili://video/%d", c.Aid)
	i.Duration = cardmdl.DurationString(c.Duration) //时长
	i.View = statString(int64(c.Stat.View), "观看")
	i.Danmaku = statString(int64(c.Stat.Danmaku), "弹幕")
	i.Dimension = &Dimension{
		Width:  c.Dimension.Width,
		Height: c.Dimension.Height,
		Rotate: c.Dimension.Rotate,
	}
}

func (i *Item) FromResourceArc(c *arccli.Arc, display bool, bvid string, f *favmdl.Folder) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	i.Image = c.Pic   //封面
	i.ItemID = c.Aid
	i.Param = bvid
	i.CoverRightText = cardmdl.DurationString(c.Duration)
	i.CoverLeftText1 = statString(int64(c.Stat.View), "")
	i.CoverLeftIcon1 = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/lM~lH4iu.png"
	i.CoverLeftText2 = statString(int64(c.Stat.Danmaku), "")
	i.CoverLeftIcon2 = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/-udU-i01.png"
	if display {
		i.Badge = &ReasonStyle{Text: "视频"}
	}
	if f == nil {
		i.URI = fmt.Sprintf("bilibili://video/%d", c.Aid)
		i.Repost = &Repost{BizType: strconv.Itoa(api.MixAvidType)}
	} else {
		i.URI = fmt.Sprintf("bilibili://music/playlist/playpage/%d?avid=%d", f.Mlid, c.Aid)
		i.Repost = &Repost{BizType: strconv.Itoa(api.MixFolder)}
		i.Fid = f.Mlid
	}
	i.ResourceInfo = &ResourceInfo{
		Up:      c.GetAuthor().Name,
		View:    statString(int64(c.GetStat().View), "观看"),
		PubTime: cardmdl.PubDataString(c.GetPubDate().Time()),
		Like:    statString(int64(c.GetStat().Like), "点赞"),
		Danmaku: statString(int64(c.GetStat().Danmaku), "弹幕"),
	}
}

// FromResourceArt .
func (i *Item) FromResourceArt(c *artmdl.Meta, artDisplay bool) {
	i.Goto = GotoResource
	i.Title = c.Title //标题
	if len(c.ImageURLs) >= 1 {
		i.Image = c.ImageURLs[0] //封面
	}
	i.ItemID = c.ID
	i.URI = fmt.Sprintf("https://www.bilibili.com/read/mobile/%d", c.ID)
	if c.Stats != nil {
		i.CoverLeftText1 = statString(int64(c.Stats.View), "")
		i.CoverLeftText2 = statString(int64(c.Stats.Reply), "")
	}
	i.CoverLeftIcon1 = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/UKWvn8PP.png"
	i.CoverLeftIcon2 = "https://i0.hdslb.com/bfs/activity-plat/static/20200317/467746a96c68611c46194c29089d62f5/Epgv08nd.png"
	if artDisplay {
		i.Badge = &ReasonStyle{Text: "文章"}
	}
	i.Repost = &Repost{BizType: strconv.Itoa(api.MixCvidType)}
	i.ResourceInfo = &ResourceInfo{
		Up:      c.Author.Name,
		View:    statString(c.Stats.View, "观看"),
		PubTime: cardmdl.PubDataString(c.PublishTime.Time()),
		Like:    statString(c.Stats.Like, "点赞"),
	}
}

func statString(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

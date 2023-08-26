package channel

import (
	"fmt"
	xtime "go-common/library/time"
	"strconv"

	"go-common/library/log"
	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	appchanmdl "go-gateway/app/app-svr/app-channel/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	cardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
)

const (
	CategoryChanListPS    = 6
	CategoryChanListCount = 6
	HotListCount          = 6
	DetailTabM            = "multiple"
	DetailTabF            = "featured"
	TabTypeS              = "season"
	TabTypeA              = "archive"
	TabTypeR              = "rank"
	MultiHot              = "hot"
	MultiView             = "view"
	MultiNew              = "new"
	SearchVideoCount      = 6
	ExtMore               = "more"
	ExtHot                = "hot"
)

var (
	_allSortHot = &TabSort{
		Title: "近期热门",
		Value: "hot",
		Icon:  "https://i0.hdslb.com/bfs/app/b65679acf2f791241a0612033a155259ccc2c4b9.png",
	}
	_allSortView = &TabSort{
		Title: "播放最多（近30天投稿）",
		Value: "view",
		Icon:  "https://i0.hdslb.com/bfs/app/7f812bdc4c36ebac1280ae9188970330a69b896c.png",
	}
	_allSortNew = &TabSort{
		Title: "最新投稿",
		Value: "new",
		Icon:  "https://i0.hdslb.com/bfs/app/4949c2c66b70c5e85a02c861835d9e22d2a9c287.png",
	}
	AllSortHotWeb = &DetailOption{
		Title: "近期热门",
		Value: MultiHot,
		Icon:  "https://i0.hdslb.com/bfs/app/b65679acf2f791241a0612033a155259ccc2c4b9.png",
	}
	AllSortVieWeb = &DetailOption{
		Title: "播放最多（近30天投稿）",
		Value: MultiView,
		Icon:  "https://i0.hdslb.com/bfs/app/7f812bdc4c36ebac1280ae9188970330a69b896c.png",
	}
	AllSortNewWeb = &DetailOption{
		Title: "最新投稿",
		Value: MultiNew,
		Icon:  "https://i0.hdslb.com/bfs/app/4949c2c66b70c5e85a02c861835d9e22d2a9c287.png",
	}
)

type Detail struct {
	ID            int64      `json:"id"`
	Param         string     `json:"param"`
	Title         string     `json:"title"`
	Cover         string     `json:"cover"`
	BgColor       string     `json:"theme_color"`
	Label1        string     `json:"label_1,omitempty"`
	Label2        string     `json:"label_2,omitempty"`
	Label3        string     `json:"label_3,omitempty"`
	Label4        string     `json:"label_4,omitempty"`
	OGVIcon       string     `json:"ogv_icon,omitempty"`
	IsAtten       int        `json:"is_atten"`
	Button        *Button    `json:"button,omitempty"`
	Tags          []*Tag     `json:"tags,omitempty"`
	Tabs          []*Tab     `json:"tabs"`
	DefaultTabIdx int64      `json:"default_tab_idx"`
	Parent        []*Channel `json:"parent,omitempty"`
	Child         []*Channel `json:"child,omitempty"`
	Alpha         int32      `json:"alpha,omitempty"`
	// 夜间模式颜色，服务端对明暗度做了调整
	BgColorNight string `json:"theme_color_night"`
}

type Channel struct {
	ID             int64   `json:"id,omitempty"`
	Name           string  `json:"name,omitempty"`
	Title          string  `json:"title,omitempty"`
	Cover          string  `json:"cover,omitempty"`
	Label          string  `json:"label,omitempty"`
	IsAtten        int     `json:"is_atten,omitempty"`
	OfficiaLVerify int32   `json:"official_verify,omitempty"`
	URI            string  `json:"uri,omitempty"`
	Goto           string  `json:"goto,omitempty"`
	Param          string  `json:"param,omitempty"`
	TypeIcon       string  `json:"type_icon,omitempty"`
	Button         *Button `json:"button,omitempty"`
	IsUpdate       int     `json:"is_update,omitempty"`
	SubType        string  `json:"sub_type,omitempty"`
	CoverLabel     string  `json:"cover_label,omitempty"`
	CoverLabel2    string  `json:"cover_label2,omitempty"`
	// 新增上报字段
	Position int64 `json:"position,omitempty"`
}

type Button struct {
	Param string `json:"param,omitempty"`
	Label string `json:"label,omitempty"`
	Text  string `json:"text,omitempty"`
	URI   string `json:"uri,omitempty"`
}

type Tag struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	URI   string `json:"uri,omitempty"`
}

type Tab struct {
	ID    string     `json:"id"`
	Title string     `json:"title"`
	URI   string     `json:"uri"`
	Sort  []*TabSort `json:"sort,omitempty"`
}

type TabSort struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Icon  string `json:"icon"`
}

func (d *Detail) FormDetail(cc *changrpc.ChannelCard, selectSort []*changrpc.FeaturedOption, seasonm map[int64]*cardgrpc.SeasonCards,
	tabs []*changrpc.ShowTab, labels []*changrpc.ShowLabel, defaultTabIdx int64) {
	if cc.GetChannelId() == 0 {
		return
	}
	d.ID = cc.GetChannelId()
	d.Param = strconv.FormatInt(cc.GetChannelId(), 10)
	d.Title = cc.GetChannelName()
	d.Cover = cc.GetBackground()
	d.BgColor = cc.GetColor()
	d.BgColorNight = cc.GetColorNight()
	d.Alpha = cc.GetAlpha()
	if showLabels, ok := makeChannelShowLabels(labels); ok {
		d.Label1 = showLabels[0]
		d.Label2 = showLabels[1]
		d.Label3 = showLabels[2]
		d.Label4 = showLabels[3]
	}
	if season, ok := seasonm[d.ID]; ok && season != nil {
		for _, card := range season.GetCards() {
			if card.GetCopyright() != "ugc" {
				d.OGVIcon = _ogvIconURL
				break
			}
		}
	}
	if cc.GetSubscribed() {
		d.IsAtten = 1
	}
	d.Button = &Button{
		Text: fmt.Sprintf("收藏 %s", StatString(cc.SubscribedCnt, "")),
	}
	for _, ac := range cc.GetAssocChannels() {
		if ac.GetChannelId() == 0 || ac.GetChannelName() == "" {
			continue
		}
		// 二期父子频道逻辑
		tag2 := &Channel{
			ID:    ac.GetChannelId(),
			Param: strconv.FormatInt(ac.GetChannelId(), 10),
			Name:  ac.GetChannelName(),
			Cover: ac.GetIcon(),
		}
		if ac.GetSubscribed() {
			tag2.IsAtten = 1
		}
		if ac.GetSubscribedCnt() != 0 {
			tag2.Label = Stat64String(ac.GetRCnt(), "投稿")
		}
		switch ac.GetAssocType() {
		case changrpc.AssocChannelType_PARENT:
			d.Parent = append(d.Parent, tag2)
		case changrpc.AssocChannelType_CHILD:
			d.Child = append(d.Child, tag2)
		default:
			log.Warn("AssocChannelType INVALID %v", ac.GetAssocType())
		}
	}
	d.Tabs = resolveChannelDetailTabs(tabs, selectSort)
	d.DefaultTabIdx = defaultTabIdx
}

func resolveChannelDetailTabs(tabs []*changrpc.ShowTab, selectSort []*changrpc.FeaturedOption) []*Tab {
	var res []*Tab
	for _, v := range tabs {
		tab := &Tab{
			ID:    v.Id,
			Title: v.Title,
			URI:   v.Url,
		}
		switch v.TabType {
		case changrpc.ShowTabType_SHOW_TAB_ALL:
			tab.Sort = append(tab.Sort, _allSortHot, _allSortView, _allSortNew)
		case changrpc.ShowTabType_SHOW_TAB_SELECT:
			var sortTmp []*TabSort
			for _, ss := range selectSort {
				if ss != nil {
					var title, year string
					if year = strconv.Itoa(int(ss.Year)); year == "0" {
						continue
					}
					title = fmt.Sprintf("%s年", year)
					sortTmp = append(sortTmp, &TabSort{Title: title, Value: year})
				}
			}
			// 筛选器中的"全部"需要网关层拼接,当服务端返回至少一个筛选项的时候才展示筛选器
			if len(sortTmp) > 0 {
				tab.Sort = append(tab.Sort, &TabSort{Title: "全部", Value: "0"})
				tab.Sort = append(tab.Sort, sortTmp...)
			}
		default:
		}
		res = append(res, tab)
	}
	return res
}

func makeChannelShowLabels(labels []*changrpc.ShowLabel) ([4]string, bool) {
	const (
		_channelLabelMaxLength = 4
	)
	if len(labels) == 0 {
		return [4]string{}, false
	}
	res := [4]string{}
	for i, v := range labels {
		if i >= _channelLabelMaxLength {
			break
		}
		res[i] = Stat64String(v.Count, v.Text)
	}
	return res, true
}

// ChannelListResult 频道详情 视频卡列表
type ChannelListResult struct {
	HasMore int            `json:"has_more"`
	Offset  string         `json:"offset"`
	Label   string         `json:"label"`
	Items   []card.Handler `json:"items"`
}

// 分类
type Category struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	ChannelCount string `json:"channel_count,omitempty"`
}

func (c *Category) FormChannelCategory(chanCategory *changrpc.ChannelCategory) {
	if chanCategory == nil {
		return
	}
	c.ID = chanCategory.GetCategoryType()
	c.Name = chanCategory.GetCategoryName()
	c.ChannelCount = appchanmdl.Stat64String(chanCategory.GetChannelCount(), "")
}

// 频道
type WebChannel struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Cover           string `json:"cover,omitempty"`
	Background      string `json:"background,omitempty"`
	SubscribedCount int32  `json:"subscribed_count,omitempty"`
	ArchiveCount    string `json:"archive_count,omitempty"`
	ViewCount       string `json:"view_count,omitempty"`
	FeaturedCount   int32  `json:"featured_count,omitempty"`
	SeasonCount     string `json:"season_count,omitempty"`
	IsSeason        bool   `json:"is_season,omitempty"`
	IsSubscribed    bool   `json:"is_subscribed,omitempty"`
	UpdatedCount    int32  `json:"updated_count,omitempty"`
	ThemeColor      string `json:"theme_color,omitempty"`
	Alpha           int32  `json:"alpha,omitempty"`
	CType           int32  `json:"ctype,omitempty"`
}

func (c *WebChannel) FormChannelCard(card *changrpc.ChannelCard) {
	if card == nil {
		return
	}
	c.ID = card.GetChannelId()
	c.Name = card.GetChannelName()
	c.Cover = card.GetIcon()
	c.Background = card.GetBackground()
	c.SubscribedCount = card.GetSubscribedCnt()
	c.ArchiveCount = appchanmdl.StatString(card.GetRCnt(), "")
	c.ViewCount = appchanmdl.Stat64String(card.GetViewCnt(), "")
	c.FeaturedCount = card.GetFeaturedCnt()
	c.IsSubscribed = card.GetSubscribed()
	c.UpdatedCount = card.GetUpdatedRsNum()
	c.ThemeColor = card.GetColor()
	c.Alpha = card.GetAlpha()
	c.CType = card.GetCtype()
}

func (c *WebChannel) FormViewChannelCard(card *changrpc.ViewChannelCard) {
	if card == nil {
		return
	}
	c.ID = card.GetCid()
	c.Name = card.GetCname()
	c.Cover = card.GetIcon()
	c.Background = card.GetBackground()
	c.SubscribedCount = card.GetSubscribedCnt()
	c.ArchiveCount = appchanmdl.StatString(card.GetResource(), "")
	c.FeaturedCount = card.GetFeaturedCnt()
	c.IsSubscribed = card.GetSubscribed()
	c.ThemeColor = card.GetColor()
	c.Alpha = card.GetAlpha()
}

func (c *WebChannel) FormAssocChannel(channel *changrpc.AssocChannel) {
	if channel == nil {
		return
	}
	c.ID = channel.GetChannelId()
	c.Name = channel.GetChannelName()
	c.Cover = channel.GetIcon()
	c.SubscribedCount = channel.GetSubscribedCnt()
	c.ArchiveCount = appchanmdl.Stat64String(channel.GetRCnt(), "")
	c.IsSubscribed = channel.GetSubscribed()
}

func (c *WebChannel) FormSeason(isPGC bool, seasonCards *cardgrpc.SeasonCards) {
	c.IsSeason = false
	c.SeasonCount = ""
	if !isPGC || seasonCards == nil || len(seasonCards.GetCards()) == 0 {
		return
	}
	c.IsSeason = true
	c.SeasonCount = appchanmdl.StatString(int32(len(seasonCards.GetCards())), "")
}

func (c *WebChannel) FormSearchChannelCard(card *changrpc.SearchChannelCard) {
	c.ID = card.GetCid()
	c.Name = card.GetCname()
	c.Cover = card.GetIcon()
	c.Background = card.GetBackground()
	c.SubscribedCount = int32(card.GetSubscribedCnt())
	c.ArchiveCount = appchanmdl.Stat64String(card.GetResourceCnt(), "")
	c.FeaturedCount = int32(card.GetFeaturedCnt())
	c.IsSubscribed = card.GetSubscribed()
	c.ThemeColor = card.GetColor()
	c.Alpha = card.Alpha
}

func (c *WebChannel) FormRelativeChannel(card *changrpc.RelativeChannel) {
	if card == nil {
		return
	}
	c.ID = card.GetCid()
	c.Name = card.GetCname()
	c.Cover = card.GetIcon()
	c.SubscribedCount = int32(card.GetSubscribedCnt())
	c.ArchiveCount = appchanmdl.Stat64String(card.GetResourceCnt(), "")
	c.FeaturedCount = int32(card.GetFeaturedCnt())
	c.IsSubscribed = card.GetSubscribed()
}

// 稿件
type Archive struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Cover           string `json:"cover,omitempty"`
	BadgeTitle      string `json:"badge_title,omitempty"`
	BadgeBackground string `json:"badge_background,omitempty"`
	ViewCount       string `json:"view_count,omitempty"`
	LikeCount       string `json:"like_count,omitempty"`
	Duration        string `json:"duration,omitempty"`
	AuthorName      string `json:"author_name,omitempty"`
	AuthorId        int64  `json:"author_id,omitempty"`
	Bvid            string `json:"bvid,omitempty"`
	Danmaku         int32  `json:"danmaku,omitempty"`
}

func (a *Archive) FormVideoCard(card *changrpc.VideoCard) {
	if card == nil {
		return
	}
	a.ID = card.GetRid()
	a.BadgeTitle = card.GetBadgeTitle()
	a.BadgeBackground = card.GetBadgeBackground()
}

func (a *Archive) FormArc(arc *arcgrpc.Arc) {
	if arc == nil {
		return
	}
	a.ID = arc.GetAid()
	a.Name = arc.GetTitle()
	a.Cover = arc.GetPic()
	a.ViewCount = appchanmdl.StatString(arc.GetStat().View, "")
	a.LikeCount = appchanmdl.StatString(arc.GetStat().Like, "")
	a.Duration = cardmdl.DurationString(arc.GetDuration())
	a.AuthorName = arc.GetAuthor().Name
	a.AuthorId = arc.GetAuthor().Mid
	a.Bvid, _ = bvid.AvToBv(arc.GetAid())
	a.Danmaku = arc.Stat.Danmaku
}

// 剧集
type Season struct {
	ID          int32  `json:"id"`
	BadgeTitle  string `json:"badge_title"`
	BadgeType   int32  `json:"badge_type"`
	Name        string `json:"name"`
	Cover       string `json:"cover"`
	Styles      string `json:"styles"`
	Actors      string `json:"actors"`
	ViewCount   string `json:"view_count"`
	FollowCount string `json:"follow_count"`
	Uri         string `json:"uri"`
}

func (s *Season) FormSeasonCard(season *cardgrpc.SeasonCard) {
	s.ID = season.GetSeasonId()
	s.BadgeTitle = season.GetBadge()
	s.BadgeType = season.GetBadgeType()
	s.Name = season.GetTitle()
	s.Cover = season.GetCover()
	s.Styles = season.GetStyles()
	s.Actors = season.GetActors()
	s.ViewCount = appchanmdl.Stat64String(season.GetStats().View, "")
	s.FollowCount = appchanmdl.Stat64String(season.GetStats().Follow, "")
	s.Uri = season.GetUri()
}

// 【频道】小红点
type RedReply struct {
	Cover           string `json:"cover"`
	ChannelID       int64  `json:"channel_id"`
	ChannelName     string `json:"channel_name"`
	Notify          bool   `json:"notify"`
	Ctype           int32  `json:"ctype"`
	SubscribedCount int32  `json:"subscribed_count"`
}

// 【分类】列表
type CategoryListReply struct {
	Categories []*Category `json:"categories"`
}

// 【频道】我的订阅列表
type SubscribedListReply struct {
	Total          int64         `json:"total"`
	StickChannels  []*WebChannel `json:"stick_channels"`
	NormalChannels []*WebChannel `json:"normal_channels"`
}

// 【频道】最近看过的频道列表
type ViewChannel struct {
	WebChannel
	Label string `json:"label"`
}

func (v *ViewChannel) FormViewChannelCard(card *changrpc.ViewChannelCard) {
	if card == nil {
		return
	}
	v.WebChannel.FormViewChannelCard(card)
	label := "看过该频道的视频"
	if card.ScanTs != 0 {
		label = cardmdl.PubDataString(card.ScanTs.Time()) + "浏览"
	}
	v.Label = label
}

type ViewListReply struct {
	Total    int64          `json:"total"`
	Channels []*ViewChannel `json:"channels"`
}

// 【分类】分类下的频道列表
type ArcChannel struct {
	WebChannel
	Archives []*Archive `json:"archives"`
}

type ChannelArcListReq struct {
	ID     int32  `json:"id" form:"id"`
	Offset string `json:"offset" form:"offset"`
}
type ChannelArcListReply struct {
	HasMore     bool          `json:"has_more"`
	Offset      string        `json:"offset"`
	Total       int32         `json:"total"`
	ArcChannels []*ArcChannel `json:"archive_channels"`
}

// 【频道】置顶
type StickReq struct {
	StickList  string `json:"stick_list" form:"stick_list"`
	NormalList string `json:"normal_list" form:"normal_list"`
}

// 【频道】热门频道列表-未登录
type HotListReq struct {
	Offset   string `json:"offset" form:"offset"`
	NeedArc  bool   `json:"need_archive" form:"need_archive"`
	PageSize int32  `json:"page_size" form:"page_size"`
}
type HotListReply struct {
	Offset      string        `json:"offset"`
	ArcChannels []*ArcChannel `json:"archive_channels"`
}

// 【频道】详情
type DetailOption struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Icon  string `json:"icon,omitempty"`
}

func (d *DetailOption) FormFeaturedOption(option *changrpc.FeaturedOption) {
	if option == nil {
		return
	}
	d.Title = fmt.Sprintf("%d年", option.GetYear())
	d.Value = strconv.Itoa(int(option.GetYear()))
}

type DetailTab struct {
	Type    string          `json:"type"`
	Options []*DetailOption `json:"options"`
}
type DetailReply struct {
	WebChannel
	TagChannels []*WebChannel `json:"tag_channels"`
	Tabs        []*DetailTab  `json:"tabs"`
}

// 【频道】精选列表
type FeaturedListReq struct {
	ChannelID  int64  `json:"channel_id" form:"channel_id" validate:"required"`
	Offset     string `json:"offset" form:"offset"`
	FilterType int32  `json:"filter_type" form:"filter_type"`
	PageSize   int32  `json:"page_size" form:"page_size" default:"20" validate:"min=1"`
}

type FeaturedListReply struct {
	Offset  string        `json:"offset"`
	HasMore bool          `json:"has_more"`
	List    []interface{} `json:"list"`
}

type TabListItem struct {
	CardType     string        `json:"card_type,omitempty"`
	PublishRange int32         `json:"publish_range,omitempty"`
	UpdateTime   int32         `json:"update_time,omitempty"`
	Title        string        `json:"title,omitempty"`
	Items        []interface{} `json:"items"`
}

// 【频道】综合列表
type MultipleListReq struct {
	ChannelID int64  `json:"channel_id" form:"channel_id" validate:"required"`
	Offset    string `json:"offset" form:"offset"`
	SortType  string `json:"sort_type" form:"sort_type" validate:"required"`
	PageSize  int32  `json:"page_size" form:"page_size" default:"20" validate:"min=1"`
}

type MultipleListReply struct {
	Offset  string        `json:"offset"`
	HasMore bool          `json:"has_more"`
	List    []interface{} `json:"list"`
}

// 【频道相关视频】精选列表->综合列表
type TopListReq struct {
	ChannelID int64  `json:"channel_id" form:"channel_id" validate:"required"`
	Offset    string `json:"offset" form:"offset"`
	PageSize  int32  `json:"page_size" form:"page_size" default:"20" validate:"min=1"`
}

type TopListReply struct {
	Offset  string        `json:"offset"`
	HasMore bool          `json:"has_more"`
	List    []interface{} `json:"list"`
}

// 【频道】搜索
type SearchReq struct {
	Keyword  string `json:"keyword" form:"keyword"`
	Page     int32  `json:"page" form:"page"`
	PageSize int32  `json:"page_size" form:"page_size"`
}

type SearchReply struct {
	Pages       int32         `json:"pages"`
	Total       int32         `json:"total"`
	ExtType     string        `json:"ext_type"`
	ArcChannels []*ArcChannel `json:"archive_channels"`
	ExtChannels []*WebChannel `json:"ext_channels,omitempty"`
}

// 【入口】视频详情页
type VideoTag struct {
	TagTopTag
	TagType         string `json:"tag_type"`
	IsActivity      bool   `json:"is_activity"`
	Color           string `json:"color"`
	Alpha           int32  `json:"alpha"`
	IsSeason        bool   `json:"is_season"`
	SubscribedCount int64  `json:"subscribed_count"`
	ArchiveCount    string `json:"archive_count"`
	FeaturedCount   int64  `json:"featured_count"`
	JumpUrl         string `json:"jump_url"`
}

type TagTopTag struct {
	ID           int64      `json:"tag_id"`
	Name         string     `json:"tag_name"`
	Cover        string     `json:"cover"`
	HeadCover    string     `json:"head_cover"`
	Content      string     `json:"content"`
	ShortContent string     `json:"short_content"`
	Type         int8       `json:"type"`
	State        int8       `json:"state"`
	CTime        xtime.Time `json:"ctime"`
	MTime        xtime.Time `json:"-"`
	// tag count
	Count struct {
		View  int `json:"view"`
		Use   int `json:"use"`
		Atten int `json:"atten"`
	} `json:"count"`
	// subscriber
	IsAtten int8 `json:"is_atten"`
	// archive_tag
	Role      int8  `json:"-"`
	Likes     int64 `json:"likes"`
	Hates     int64 `json:"hates"`
	Attribute int8  `json:"attribute"`
	Liked     int8  `json:"liked"`
	Hated     int8  `json:"hated"`
	ExtraAttr int32 `json:"extra_attr"`
	// bgm
	MusicId string `json:"music_id"`
}

// 【频道】订阅
type SubscribeReq struct {
	ID int64 `json:"id" form:"id" validate:"required"`
}

// 【频道】取消订阅
type UnsubscribeReq struct {
	ID int64 `json:"id" form:"id" validate:"required"`
}

// 【频道】详情
type WebDetailReq struct {
	ID int64 `json:"id" form:"channel_id" validate:"required"`
}

// 【频道】列表
type ChannelListReq struct {
	ID       int32  `json:"id" form:"id" validate:"required,min=1"`
	Offset   string `json:"offset" form:"offset"`
	PageSize int32  `json:"page_size" form:"page_size" validate:"required,min=1"`
}
type ChannelListReply struct {
	HasMore  bool          `json:"has_more"`
	Offset   string        `json:"offset"`
	Total    int64         `json:"total"`
	Channels []*WebChannel `json:"channels"`
}

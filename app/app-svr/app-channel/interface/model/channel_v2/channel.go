package channel_v2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/component/metadata/device"
	"go-common/library/log"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/conf"
	"go-gateway/app/app-svr/app-channel/interface/model"
	chmdl "go-gateway/app/app-svr/app-channel/interface/model/channel"
	tabmdl "go-gateway/app/app-svr/app-channel/interface/model/tab"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	natgrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	CardTypeVideo  = 0
	CardTypeCustom = 256
	CardTypeRank   = 257

	TypeIconHot     = 1
	_typeIconHotURI = "https://i0.hdslb.com/bfs/app/2d809eae3527525f5e7067500f186dc833de2e47.png"
	TypeIconNew     = 2
	_typeIconNewURI = "https://i0.hdslb.com/bfs/app/71e8c298f2d9790c27473cd05b714eb3a58e5f51.png"

	RankURL = "https://www.bilibili.com/h5/channel/rank?id=%v&theme=%v&navhide=1"

	RedTypePoint = 1
	RedTypeNum   = 2

	_ogvIconURL = "https://i0.hdslb.com/bfs/tag/dd3b2d6991e3dd8640a5412282f1e08800d1e0b6.png"

	TabAll    = 0
	TabMineID = 999
	TabSelect = 100

	TabTypeMine = 1
	TabTypeAll  = 2
	TabSubTitle = ` (<em class="count"></em>)`

	ModelTypeSearch    = "search"
	ModelNameSearch    = "搜索"
	ModelTypeSubscribe = "subscribe"
	ModelNameSubscribe = "我的订阅"
	ModelNameFav       = "我的收藏"
	ModelTypeNew       = "new"
	ModelNameNew       = "我订阅的更新"
	ModelNameFavNew    = "我收藏的更新"
	ModelTypeScaned    = "scaned"
	ModelNameScaned    = "我看过的频道"
	ModelTypeRcmd      = "rcmd"
	ModelNameRcmd      = "热门频道"
	ModelTypeHotTopic  = "topic_rcmd"
	ModelNameHotTopic  = "推荐话题"

	SquareAllChannelIcon = "https://i0.hdslb.com/bfs/app/e551e2b4350f85c955750ad6043143d9bb96ab21.png"
	SquareAddChannelIcon = "https://i0.hdslb.com/bfs/app/391349e0fe7c72fcfad18d7a5a7b0b1278d142c4.png"

	// https://www.tapd.bilibili.co/20064511/bugtrace/bugs/view/1120064511001212931?code=YmgEZ4foyX_HIURnq75DP0X-eDhw1ZMCtrCTCl3TnQA&state=TAPD_QY_WECHAT
	NewSubVersion = 1
	OldSubVersion = 0

	ActiveTag = 1

	DynamicTypeChannelOpen           = "channel_open"
	DynamicTypeChannelSub            = "channel_sub"
	DynamicTypeChannelFeaturedUpdate = "channel_featured_update"
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
)

type ChanelDetailExternalArgs struct {
	Args map[string]string `form:"args"`
}

type ChannelListTab struct {
	ID       int64  `json:"id"`
	TabType  int    `json:"tab_type,omitempty"`
	Title    string `json:"title"`
	Count    int64  `json:"count"`
	SubTitle string `json:"sub_title,omitempty"`
}

func (clt *ChannelListTab) FormChannelListTab(c *channelgrpc.ChannelCategory) {
	clt.ID = int64(c.GetCategoryType())
	clt.Title = c.GetCategoryName()
	clt.TabType = TabTypeAll
	clt.Count = c.GetChannelCount()
}

type ChannelResult struct {
	HasMore int        `json:"has_more"`
	Offset  string     `json:"offset"`
	Title   string     `json:"title"` // 页面顶部文案 与我的订阅相同
	Items   []*Channel `json:"items"`
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

func (c *Channel) FormChannel(cc *channelgrpc.ChannelCard, actInfos map[int64]*natgrpc.NativePage, mobiApp, spmid string, build int64, isHighBuild bool) {
	c.ID = cc.ChannelId
	c.Name = cc.ChannelName
	c.Title = cc.ChannelName
	c.Cover = cc.Icon
	c.OfficiaLVerify = cc.Verify
	c.Param = strconv.FormatInt(cc.ChannelId, 10)
	subscribeText := "订阅"
	if card.FavTextReplace(mobiApp, build) {
		subscribeText = "收藏"
	}
	var labels []string
	if cc.GetSubscribedCnt() != 0 {
		labels = append(labels, model.StatString(cc.GetSubscribedCnt(), subscribeText))
	}
	if cc.GetFeaturedCnt() != 0 {
		labels = append(labels, model.StatString(cc.GetFeaturedCnt(), "精选视频"))
	} else if cc.GetRCnt() != 0 {
		labels = append(labels, model.StatString(cc.GetRCnt(), "投稿"))
	}
	if len(labels) > 0 {
		c.Label = strings.Join(labels, "  ")
	}
	switch cc.Ctype {
	case model.NewChannel:
		if cc.BizType == channelgrpc.ChannelBizlType_MOVIE && isHighBuild {
			c.Goto = model.GotoChannelMedia
			c.URI = model.FillURI(c.Goto, fmt.Sprintf("%d", cc.ChannelId), 0, 0, 0, model.ChannelHandler(fmt.Sprintf("biz_id=%d&biz_type=0&source=%s", cc.ChannelId, spmid)))
		} else {
			c.Goto = model.GotoChannelNew
			c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
		}
	case model.OldChanne:
		// 优先级 活动跳链模式 > 活动普通模式 > 旧频道
		c.Goto = model.GotoTag
		c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
		// 如果是活动话题逻辑
		if cc.GetActAttr() == ActiveTag {
			if actInfo, ok := actInfos[c.ID]; ok && actInfo != nil {
				c.Goto = model.GotoActive
				c.URI = model.FillURI(c.Goto, strconv.FormatInt(actInfo.ID, 10), 0, 0, 0, nil)
				if actInfo.ShareImage != "" {
					c.Cover = actInfo.ShareImage
				}
				if actInfo.SkipURL != "" {
					c.Goto = model.GotoActivity
					c.URI = actInfo.SkipURL
				}
			}
		}
	}
	if cc.Subscribed {
		c.IsAtten = 1
	}
	if cc.SubscribedCnt == 0 {
		c.Button = &Button{
			Text: subscribeText,
		}
	} else {
		c.Button = &Button{
			Text: model.StatString(cc.SubscribedCnt, fmt.Sprintf(" %s", subscribeText)),
		}
	}
	switch cc.Class {
	case TypeIconHot:
		c.TypeIcon = _typeIconHotURI
	case TypeIconNew:
		c.TypeIcon = _typeIconNewURI
	}
	// 是否有更新,红圈逻辑
	if cc.HasNewerRs {
		c.IsUpdate = 1
	}
}

type ChannelMineResult struct {
	Config *SubscribeConfig `json:"config,omitempty"`
	Stick  []*Channel       `json:"stick,omitempty"`
	Normal []*Channel       `json:"normal,omitempty"`
	Scaned []*Channel       `json:"scaned,omitempty"`
}

type SubscribeConfig struct {
	Title        string  `json:"title,omitempty"` // 页面顶部文案 与 全部list相同
	Label        string  `json:"label,omitempty"`
	NoSubLabel   string  `json:"no_sub_label,omitempty"`
	SubLabel     string  `json:"sub_label,omitempty"`
	LoginButton  *Button `json:"login_button,omitempty"`
	NoSubButton  *Button `json:"no_sub_button,omitempty"`
	NoMoreButton *Button `json:"no_more_button,omitempty"`
}

func (c *Channel) FormChannelMine(cc *channelgrpc.ChannelCard, actInfos map[int64]*natgrpc.NativePage, isHighBuild bool, spmid string) {
	c.ID = cc.ChannelId
	c.Name = cc.ChannelName
	c.Cover = cc.Icon
	c.OfficiaLVerify = cc.Verify
	c.Param = strconv.FormatInt(cc.ChannelId, 10)
	switch cc.Ctype {
	case model.NewChannel:
		if cc.BizType == channelgrpc.ChannelBizlType_MOVIE && isHighBuild {
			c.Goto = model.GotoChannelMedia
			c.URI = model.FillURI(c.Goto, fmt.Sprintf("%d", cc.ChannelId), 0, 0, 0, model.ChannelHandler(fmt.Sprintf("biz_id=%d&biz_type=0&source=%s", cc.ChannelId, spmid)))
		} else {
			c.Goto = model.GotoChannelNew
			c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
		}
	case model.OldChanne:
		// 优先级 活动跳链模式 > 活动普通模式 > 旧频道
		c.Goto = model.GotoTag
		c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
		// 如果是活动话题逻辑
		if cc.GetActAttr() == ActiveTag {
			if actInfo, ok := actInfos[c.ID]; ok && actInfo != nil {
				c.Goto = model.GotoActive
				c.URI = model.FillURI(c.Goto, strconv.FormatInt(actInfo.ID, 10), 0, 0, 0, nil)
				if actInfo.ShareImage != "" {
					c.Cover = actInfo.ShareImage
				}
				if actInfo.SkipURL != "" {
					c.Goto = model.GotoActivity
					c.URI = actInfo.SkipURL
				}
			}
		}
	}
}

func (c *Channel) FormChannelMineScaned(cc *channelgrpc.ViewChannelCard, subscribeText string) {
	c.ID = cc.GetCid()
	c.Name = cc.GetCname()
	c.Cover = cc.Icon
	c.Param = strconv.FormatInt(cc.GetCid(), 10)
	c.Goto = model.GotoChannelNew
	c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
	var labels []string
	if cc.GetFeaturedCnt() > 0 {
		labels = append(labels, model.StatString(cc.GetFeaturedCnt(), "精选"))
	}
	label := "看过该频道的视频"
	if cc.ScanTs != 0 {
		label = cdm.PubDataString(cc.ScanTs.Time()) + "浏览"
	}
	labels = append(labels, label)
	if len(labels) > 0 {
		c.Label = strings.Join(labels, " | ")
	}
	c.Button = &Button{
		Text: subscribeText,
	}
	if cc.GetSubscribedCnt() != 0 {
		c.Button = &Button{
			Text: model.StatString(cc.GetSubscribedCnt(), fmt.Sprintf(" %s", subscribeText)),
		}
	}
}

type SquareResult struct {
	Region    []*chmdl.Region `json:"region,omitempty"`
	Subscribe *Subscribe      `json:"subscribe,omitempty"`
	Recent    []*Channel      `json:"recent,omitempty"`
	News      *New            `json:"new,omitempty"`
	Scaned    *Scaned         `json:"scaned,omitempty"`
	Rcmd      *Rcmd           `json:"rcmd,omitempty"`
}

type Subscribe struct {
	Count int64      `json:"count"`
	Items []*Channel `json:"items,omitempty"`
}

type New struct {
	HasMore int            `json:"has_more"`
	Label   string         `json:"label,omitempty"`
	Offset  string         `json:"offset"`
	Items   []card.Handler `json:"items,omitempty"`
}

type Scaned struct {
	Label string         `json:"label,omitempty"`
	Items []card.Handler `json:"items,omitempty"`
}

func (c *Channel) FormChannelRecent(cc *channelgrpc.ChannelCard) {
	c.ID = cc.ChannelId
	c.Name = cc.ChannelName
	c.Cover = cc.Icon
	c.OfficiaLVerify = cc.Verify
	c.Param = strconv.FormatInt(cc.ChannelId, 10)
	var labels []string
	if cc.RCnt != 0 {
		labels = append(labels, model.StatString(cc.RCnt, "投稿"))
	}
	if cc.FeaturedCnt != 0 {
		labels = append(labels, model.StatString(cc.FeaturedCnt, "个精选视频"))
	}
	if len(labels) > 0 {
		c.Label = strings.Join(labels, "  ")
	}
	if cc.Subscribed {
		c.IsAtten = 1
	}
	switch cc.Ctype {
	case model.NewChannel:
		c.Goto = model.GotoChannelNew
		c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
	case model.OldChanne:
		c.Goto = model.GotoTag
		c.URI = model.FillURI(c.Goto, c.Param, 0, 0, 0, nil)
	}
	if cc.SubscribedCnt == 0 {
		c.Button = &Button{
			Text: "订阅",
		}
	} else {
		c.Button = &Button{
			Text: model.StatString(cc.SubscribedCnt, " 订阅"),
		}
	}
}

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

// nolint:gocognit
func (d *Detail) FormDetail(ctx context.Context, cc *channelgrpc.ChannelCard, menus []*tabmdl.Menu, selectSort []*channelgrpc.FeaturedOption, seasonm map[int64]*appCardgrpc.SeasonCards, ogvSwitch, isOverSea, isHighBuild bool, prlimit *conf.PRLimit, build, defaultTabIdx int64,
	mobiApp, spmid string, tabs []*channelgrpc.ShowTab, labels []*channelgrpc.ShowLabel) {
	if cc == nil {
		return
	}
	if cc.ChannelId == 0 {
		return
	}
	d.ID = cc.ChannelId
	d.Param = strconv.FormatInt(cc.ChannelId, 10)
	d.Title = cc.ChannelName
	d.Cover = cc.Background
	d.BgColor = cc.Color
	d.BgColorNight = cc.ColorNight
	d.Alpha = cc.Alpha
	if showLabels, ok := makeChannelShowLabels(labels); ok {
		d.Label1 = showLabels[0]
		d.Label2 = showLabels[1]
		d.Label3 = showLabels[2]
		d.Label4 = showLabels[3]
	}
	if ogvSwitch {
		if season, ok := seasonm[d.ID]; ok && season != nil {
			for _, card := range season.GetCards() {
				if card == nil {
					continue
				}
				if card.Copyright != "ugc" {
					d.OGVIcon = _ogvIconURL
					break
				}
			}
		}
	}
	if cc.Subscribed {
		d.IsAtten = 1
	}
	text := fmt.Sprintf("订阅 %s", model.StatString(cc.SubscribedCnt, ""))
	if card.FavTextReplace(mobiApp, build) {
		text = fmt.Sprintf("收藏 %s", model.StatString(cc.SubscribedCnt, ""))
	}
	d.Button = &Button{
		Text: text,
	}
	for _, ac := range cc.AssocChannels {
		if ac.ChannelId == 0 || ac.ChannelName == "" {
			continue
		}
		var tUrl string
		if isHighBuild && ac.BizType == channelgrpc.ChannelBizlType_MOVIE {
			tUrl = model.FillURI(model.GotoChannelMedia, fmt.Sprintf("%d", ac.ChannelId), 0, 0, 0, model.ChannelHandler(fmt.Sprintf("biz_id=%d&biz_type=0&source=%s", ac.ChannelId, spmid)))
		} else {
			tUrl = model.FillURI(model.GotoChannelNew, strconv.FormatInt(ac.ChannelId, 10), 0, 0, 0, model.ChannelHandler("tab=all&sort=hot"))
		}
		tag := &Tag{
			ID:    ac.ChannelId,
			Title: ac.ChannelName,
			URI:   tUrl,
		}
		// 二期父子频道逻辑
		tag2 := &Channel{
			ID:    ac.ChannelId,
			Param: strconv.FormatInt(ac.ChannelId, 10),
			Goto:  model.GotoChannelNew,
			Name:  ac.ChannelName,
			Cover: ac.Icon,
			URI:   tUrl,
		}
		if ac.Subscribed {
			tag2.IsAtten = 1
		}
		if ac.SubscribedCnt != 0 {
			tag2.Label = model.Stat64String(ac.RCnt, "投稿")
		}
		switch ac.AssocType {
		case channelgrpc.AssocChannelType_PARENT:
			d.Parent = append(d.Parent, tag2)
		case channelgrpc.AssocChannelType_CHILD:
			d.Child = append(d.Child, tag2)
		default:
			log.Warn("AssocChannelType INVALID %v", ac.AssocType)
		}
		d.Tags = append(d.Tags, tag)
	}
	if CanShowNewVersionChannelTab(ctx) {
		d.Tabs = resolveChannelDetailTabs(tabs, selectSort)
		d.DefaultTabIdx = defaultTabIdx
		return
	}
	var tab *Tab
	// 精选tab
	if cc.FeaturedCnt != 0 {
		tab = &Tab{
			ID:    "select",
			Title: "精选",
			URI:   model.FillURI(model.GotoChannelNewSelect, d.Param, 0, 0, 0, nil),
		}
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
		d.Tabs = append(d.Tabs, tab)
	}
	// 综合tab
	tab = &Tab{
		ID:    "all",
		Title: "综合",
		URI:   model.FillURI(model.GotoChannelNewAll, d.Param, 0, 0, 0, nil),
	}
	// 被PR限制的频道不显示 自定义tab和话题tab
	var isPR bool
	for _, cid := range prlimit.ChannelList {
		if cid == cc.ChannelId {
			isPR = true
			break
		}
	}
	if isPR {
		tab.Sort = append(tab.Sort, _allSortHot)
		d.Tabs = append(d.Tabs, tab)
		return
	}
	tab.Sort = append(tab.Sort, _allSortHot, _allSortView, _allSortNew)
	d.Tabs = append(d.Tabs, tab)
	// 自定义tab
	for _, m := range menus {
		if m != nil {
			tab = &Tab{
				ID:    strconv.FormatInt(m.TabID, 10),
				Title: m.Name,
				URI:   model.FillURI(model.GotoChannelNewOP, strconv.FormatInt(m.TabID, 10), 0, 0, 0, model.PegasusHandler(m)),
			}
			d.Tabs = append(d.Tabs, tab)
			// only one tab
			break
		}
	}
	if isOverSea {
		return
	}
	// 话题tab
	tab = &Tab{
		ID:    "topic",
		Title: "话题",
		URI:   model.FillURI(model.GotoChannelNewTopic, d.Param, 0, 0, 0, model.NewChannelTopic(cc.ChannelName)),
	}
	d.Tabs = append(d.Tabs, tab)
}

func resolveChannelDetailTabs(tabs []*channelgrpc.ShowTab, selectSort []*channelgrpc.FeaturedOption) []*Tab {
	var res []*Tab
	for _, v := range tabs {
		tab := &Tab{
			ID:    v.Id,
			Title: v.Title,
			URI:   v.Url,
		}
		switch v.TabType {
		case channelgrpc.ShowTabType_SHOW_TAB_ALL:
			tab.Sort = append(tab.Sort, _allSortHot, _allSortView, _allSortNew)
		case channelgrpc.ShowTabType_SHOW_TAB_SELECT:
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

func makeChannelShowLabels(labels []*channelgrpc.ShowLabel) ([4]string, bool) {
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
		res[i] = model.Stat64String(v.Count, v.Text)
	}
	return res, true
}

func CanShowNewVersionChannelTab(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidB().Or().IsPlatAndroidI().And().Build(">", int64(6670000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().Or().IsPlatIPhoneB().Or().IsPlatIPhoneI().And().Build(">", int64(66700000))
	}).MustFinish()
}

// ChannelListResult 频道详情 视频卡列表
type ChannelListResult struct {
	HasMore int            `json:"has_more"`
	Offset  string         `json:"offset"`
	Label   string         `json:"label"`
	Items   []card.Handler `json:"items"`
}

func MarkRed(str string) (red string) {
	red = `<em class="keyword">` + str + `</em>`
	return
}

type ChannelInfoc struct {
	EventId     string       `json:"event_id"`
	Page        string       `json:"page,omitempty"`
	Sort        string       `json:"sort,omitempty"`
	Filt        string       `json:"filt,omitempty"`
	Items       []*InfocItem `json:"items"`
	CardNum     int          `json:"card_num"`
	RequestUrl  string       `json:"request_url"`
	TimeIso     int64        `json:"time_iso"`
	Ip          string       `json:"ip"`
	AppId       int32        `json:"app_id"`
	Platform    int32        `json:"platform"`
	Buvid       string       `json:"buvid"`
	Version     string       `json:"version"`
	VersionCode string       `json:"version_code"`
	Mid         string       `json:"mid"`
	Ctime       string       `json:"ctime"`
	Abtest      string       `json:"abtest"`
	AutoRefresh int          `json:"auto_refresh"`
	From        string       `json:"from"`
	Pos         string       `json:"pos"`
	CurRefresh  int          `json:"cur_refresh"`
}

type InfocItem struct {
	ChannelID int64        `json:"channle_id,omitempty"`
	CardType  string       `json:"card_type,omitempty"`
	OID       int64        `json:"oid,omitempty"`
	Corner    string       `json:"corner,omitempty"`
	Pos       int          `json:"pos,omitempty"`
	Items     []*InfocItem `json:"items,omitempty"`
}

type RankResult struct {
	HasMore int     `json:"has_more"`
	Offset  string  `json:"offset"`
	Items   []*Rank `json:"items"`
	Title   string  `json:"title,omitempty"`
	Label   string  `json:",omitempty"`
}

type Rank struct {
	ID       int64          `json:"id,omitempty"`
	Title    string         `json:"title,omitempty"`
	Cover    string         `json:"cover,omitempty"`
	Label    string         `json:"label,omitempty"`
	Duration int64          `json:"duration,omitempty"`
	Author   arcgrpc.Author `json:"author,omitempty"`
	Goto     string         `json:"goto,omitempty"`
	Param    string         `json:"param,omitempty"`
}

type Share struct {
	ID         int64      `json:"id,omitempty"`
	Share      *ShareItem `json:"share,omitempty"`
	ShareURI   string     `json:"share_uri,omitempty"`
	Title      string     `json:"title,omitempty"`
	Param      string     `json:"param,omitempty"`
	Desc       string     `json:"desc,omitempty"`
	Icon       string     `json:"icon,omitempty"`
	ChannelURI string     `json:"channel_uri,omitempty"`
}

type ShareItem struct {
	Weibo          bool `json:"weibo"`
	Wechart        bool `json:"wechat"`
	WechartMonment bool `json:"wechat_monment"`
	QQ             bool `json:"qq"`
	QZone          bool `json:"qzone"`
	Copy           bool `json:"copy"`
	More           bool `json:"more"`
}

type Red struct {
	Type   int `json:"type"`
	Number int `json:"number"`
}

type Rcmd struct {
	Label string         `json:"label,omitempty"`
	Items []card.Handler `json:"items,omitempty"`
}

type SquareScaned struct {
	CardType        string `json:"card_type,omitempty"`
	CardGoto        string `json:"card_goto,omitempty"`
	ID              int64  `json:"id,omitempty"`
	Param           string `json:"param,omitempty"`
	Title           string `json:"title,omitempty"`
	Cover           string `json:"cover,omitempty"`
	Background      string `json:"background,omitempty"`
	ThemeColor      string `json:"theme_color,omitempty"`
	ThemeColorNight string `json:"theme_color_night,omitempty"`
	Alpha           int32  `json:"alpha,omitempty"`
	Goto            string `json:"goto,omitempty"`
	URI             string `json:"uri,omitempty"`
	Desc            string `json:"desc,omitempty"`
	Position        int64  `json:"position,omitempty"`
	SType           int    `json:"s_type,omitempty"`
}

func (sq *SquareScaned) FormSquareScaned(c *channelgrpc.ViewChannelCard) {
	sq.ID = c.Cid
	sq.Param = strconv.FormatInt(c.Cid, 10)
	sq.Title = c.Cname
	sq.Cover = c.Icon
	sq.Background = c.Background
	sq.ThemeColor = c.Color
	sq.Alpha = c.Alpha
	sq.Goto = model.GotoChannelNew
	sq.URI = model.FillURI(sq.Goto, sq.Param, 0, 0, 0, model.ChannelHandler("tab=all"))
	sq.Desc = "看过该频道的视频"
	sq.SType = 2
	if c.ScanTs != 0 {
		sq.Desc = cdm.PubDataString(c.ScanTs.Time()) + "浏览"
		sq.SType = 1
	}
}

type SquareHot struct {
	List    []*HotList     `json:"list,omitempty"`
	Dynamic []*HotDynamic  `json:"dynamic,omitempty"`
	Rcmd    []card.Handler `json:"rcmd,omitempty"`
}

type HotList struct {
	ID       int64  `json:"id,omitempty"`
	Param    string `json:"param,omitempty"`
	Title    string `json:"title,omitempty"`
	Cover    string `json:"cover,omitempty"`
	Goto     string `json:"goto,omitempty"`
	URI      string `json:"uri,omitempty"`
	Position int64  `json:"position,omitempty"`
}

func (hl *HotList) FormHotList(h *channelgrpc.ViewChannelCard, position int64) {
	hl.ID = h.Cid
	hl.Param = strconv.FormatInt(hl.ID, 10)
	hl.Title = h.Cname
	hl.Cover = h.Icon
	hl.Goto = model.GotoChannelNew
	hl.URI = model.FillURI(hl.Goto, hl.Param, 0, 0, 0, model.ChannelHandler("tab=select"))
	hl.Position = position
}

type HotDynamic struct {
	ID     int64        `json:"id,omitempty"`
	Param  string       `json:"param,omitempty"`
	Title  string       `json:"title,omitempty"`
	Cover  string       `json:"cover,omitempty"`
	Goto   string       `json:"goto,omitempty"`
	URI    string       `json:"uri,omitempty"`
	Desc   string       `json:"desc"`
	Button *card.Button `json:"button,omitempty"`
}

func (hd *HotDynamic) FormHotDynamic(d *channelgrpc.DynamicCard, am map[int64]*arcgrpc.ArcPlayer) {
	hd.ID = d.Cid
	hd.Param = strconv.FormatInt(d.Cid, 10)
	hd.Title = d.Cname
	hd.Cover = d.Icon
	hd.Goto = model.GotoChannelNew
	hd.URI = model.FillURI(hd.Goto, hd.Param, 0, 0, 0, model.ChannelHandler("tab=select"))
	switch d.DynamicType {
	case DynamicTypeChannelOpen:
		hd.Desc = "新频道开放了～"
	case DynamicTypeChannelFeaturedUpdate:
		var arcTitle string
		if a, ok := am[d.Rid]; ok && a != nil && a.Arc != nil {
			arcTitle = a.Arc.Title
			hd.Desc = fmt.Sprintf("新增精选【%s】", arcTitle)
		}
	case DynamicTypeChannelSub:
		if d.SubscribedCnt > 0 {
			hd.Desc = fmt.Sprintf("订阅人数达到%s", model.Stat64String(d.SubscribedCnt, "～"))
		}
	}
	hd.Button = &card.Button{
		Text: "去看看",
		URI:  model.FillURI(hd.Goto, hd.Param, 0, 0, 0, model.ChannelHandler("tab=select")),
	}
}

type HotTopic struct {
	ID         int64        `json:"id,omitempty"`
	Title      string       `json:"title,omitempty"`
	Label      string       `json:"label,omitempty"`
	Cover      string       `json:"cover,omitempty"`
	URI        string       `json:"uri,omitempty"`
	SedID      int          `json:"sed_id,omitempty"`
	SedType    string       `json:"sed_type,omitempty"`
	RcmdReason *ReasonStyle `json:"rcmd_reason,omitempty"`
}

type ReasonStyle struct {
	Text             string `json:"text,omitempty"`
	TextColor        string `json:"text_color,omitempty"`
	TextColorNight   string `json:"text_color_night,omitempty"`
	BgColor          string `json:"bg_color,omitempty"`
	BgColorNight     string `json:"bg_color_night,omitempty"`
	BorderColor      string `json:"border_color,omitempty"`
	BorderColorNight string `json:"border_color_night,omitempty"`
	BgStyle          int8   `json:"bg_style,omitempty"`
}

func CanNewTopicOnline(ctx context.Context) bool {
	dev, _ := device.FromContext(ctx)
	// 与替换收藏字段的版本控制一致，以新话题上线时间为准
	return card.FavTextReplace(dev.MobiApp(), dev.Build)
}

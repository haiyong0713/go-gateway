package dynamicV2

import (
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	topicV2 "git.bilibili.co/bapis/bapis-go/topic/service"
)

/* ************************************
	综合feed上滑: FromMixNew
	综合feed下滑: FromMixHistory
	综合页快速消费: FromAllPersonal
*************************************** */

func (list *DynListRes) FromMixNew(new *dyngrpc.GeneralNewRsp, uid int64) {
	list.UpdateNum = new.UpdateNum
	list.HistoryOffset = new.HistoryOffset
	list.UpdateBaseline = new.UpdateBaseline
	list.HasMore = new.HasMore
	list.StoryUpCard = new.StoryUpCard
	for _, item := range new.Dyns {
		if item == nil || item.Type == 0 {
			continue
		}
		if item.Type == 1 && item.Origin == nil {
			continue
		}
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
	if new.FoldInfo != nil {
		fold := &FoldInfo{}
		fold.FromFold(new.FoldInfo)
		list.FoldInfo = fold
	}
	if rcmd := new.GetRcmdUps(); rcmd != nil {
		list.RegionUps = &dyngrpc.UnLoginRsp{
			Opts:       rcmd.Opts,
			RegionUps:  rcmd.RegionUps,
			ServerInfo: rcmd.ServerInfo,
		}
		rcmdUps := &RcmdUPCard{}
		rcmdUps.FromRcmdUPCard(rcmd)
		list.RcmdUps = rcmdUps
	}
}

func (list *DynListRes) FromMixHistory(history *dyngrpc.GeneralHistoryRsp, uid int64) {
	list.HistoryOffset = history.HistoryOffset
	list.HasMore = history.HasMore
	for _, item := range history.Dyns {
		if item == nil || item.Type == 0 {
			continue
		}
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
	if history.FoldInfo != nil {
		fold := &FoldInfo{}
		fold.FromFold(history.FoldInfo)
		list.FoldInfo = fold
	}
}

// 综合页个人feed流列表信息
type AllPersonal struct {
	HasMore    bool
	Offset     string
	Dynamics   []*Dynamic
	FoldInfo   *FoldInfo
	ReadOffset string
}

func (list *AllPersonal) FromAllPersonal(personal *dyngrpc.VideoPersonalRsp) {
	list.HasMore = personal.HasMore
	list.Offset = personal.Offset
	list.ReadOffset = personal.ReadOffset
	if personal.FoldInfo != nil {
		fo := &FoldInfo{}
		fo.FromFold(personal.FoldInfo)
		list.FoldInfo = fo
	}
	for _, item := range personal.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

/*
最常访问
*/
type MixUpList struct {
	ShowLiveNum int    `json:"show_live_num"` // 埋点用
	ModuleTitle string `json:"module_title"`
	ViewMore    *struct {
		Type               uint32 `json:"type"` // 服务端注释: 0-不展示，1-展示
		Text               string `json:"text"`
		MixFixedEntry      uint32 `json:"mix_fixed_entry"`      // 是否在综合页头像模块右上角加固定入口，1展示，0不展示
		PersonalFixedEntry uint32 `json:"personal_fixed_entry"` // 是否在综合快消页头像模块右上角加固定入口，1展示，0不展示
	} `json:"view_more"`
	List              []*MixUpListItem `json:"list"`
	Footprint         string           `json:"footprint"`
	ModuleTitleSwitch int32            `json:"module_title_switch"`
}

type MixUpListItem struct {
	Type int `json:"type"` // 服务端注释: 1-直播用户, 2-动态up主
	// 用户信息
	UserProfile *struct {
		Info *struct {
			UID   int64  `json:"uid"`
			Uname string `json:"uname"`
			Face  string `json:"face"`
		} `json:"info"`
	} `json:"user_profile"`
	// 直播用户
	LiveInfo *struct {
		RoomID         int64  `json:"room_id"`
		RUID           int64  `json:"ruid"`
		RUname         string `json:"runame"`
		Face           string `json:"face"`
		JumpURL        string `json:"jump_url"`
		AreaID         int64  `json:"area_v2_id"`
		AreaName       string `json:"area_v2_name"`
		AreaParentID   int    `json:"area_v2_parent_id"`
		AreaParentName string `json:"area_v2_parent_name"`
		LiveStart      int64  `json:"live_start"`
	} `json:"live_info"`
	HasPostSeparator   int                `json:"has_post_separator"`
	DisplayStyleNormal *api.UserItemStyle `json:"display_style_normal"`
	DisplayStyleDark   *api.UserItemStyle `json:"display_style_dark"`
	StyleID            int64              `json:"style_id"`
	// 动态up主
	HasUpdate       int32 `json:"has_update"`        // 服务端注释: 动态up主使用; 0-不显示小红点, 1-显示小红点
	IsReserveRecall bool  `json:"is_reserve_recall"` // 是否是预约召回
}

// 避免直接引用conf包
// 解决非本业务代码引用该model层误触误触conf包init函数的问题
type AppDynamicConfig interface {
	GetResDynMixTopicSquareMore() (icon, text string)
	GetIconModuleExtendNewTopic() (icon string)
	GetResModuleTitleForCampusTopic() (title, moreBtnText, moreBtnIcon string)
	GetPlusMarkIcon() (icon string)
}

// 动态综合页话题广场
type DynAllTopicSquare interface {
	ToDynV2TopicList(c AppDynamicConfig) *api.TopicList
}

// 老话题新鲜事
type OldTopicSquareImpl struct {
	Title string `json:"title"`
	List  []*struct {
		TopicID     int64  `json:"topic_id"`
		TopicName   string `json:"topic_name"`
		TopicLink   string `json:"topic_link"`
		IconDesc    string `json:"icon_desc"`
		IconURL     string `json:"icon_url"`
		IconType    string `json:"icon_type"`
		ServerInfo  string `json:"server_info"`
		HeadIconUrl string `json:"head_icon_url"`
	} `json:"list"`
	JumpURL   string `json:"jump_url"`
	ActButton *struct {
		ButtonDesc    string `json:"button_desc"`
		ButtonIcon    string `json:"button_icon"`
		ButtonJumpURL string `json:"button_jump_url"`
	} `json:"act_button"`
	ServerInfo string `json:"server_info"`
	RedDot     bool   `json:"red_dot"`
	JumpTitle  string `json:"jump_title"`
}

func (otsi *OldTopicSquareImpl) ToDynV2TopicList(c AppDynamicConfig) (res *api.TopicList) {
	if otsi == nil {
		return nil
	}
	var pos int64
	var list []*api.TopicListItem
	for _, topic := range otsi.List {
		pos++
		list = append(list, &api.TopicListItem{
			Icon:        topic.IconURL,
			IconTitle:   topic.IconDesc,
			TopicId:     topic.TopicID,
			TopicName:   topic.TopicName,
			Url:         topic.TopicLink,
			Pos:         pos,
			ServerInfo:  topic.ServerInfo,
			HeadIconUrl: topic.HeadIconUrl,
		})
	}
	if len(list) == 0 {
		return nil
	}
	res = &api.TopicList{
		Title:         otsi.Title,
		TopicListItem: list,
		ServerInfo:    otsi.ServerInfo,
	}
	if otsi.ActButton != nil {
		res.ActButton = &api.TopicButton{
			Icon:    otsi.ActButton.ButtonIcon,
			Title:   otsi.ActButton.ButtonDesc,
			JumpUri: otsi.ActButton.ButtonJumpURL,
		}
	}
	if otsi.JumpURL != "" {
		res.MoreButton = &api.TopicButton{
			JumpUri: otsi.JumpURL,
			RedDot:  otsi.RedDot,
		}
		res.MoreButton.Icon, res.MoreButton.Title = c.GetResDynMixTopicSquareMore()
		if otsi.JumpTitle != "" {
			res.MoreButton.Title = otsi.JumpTitle
		}
	}
	return
}

// 动态综合-新话题广场
type NewTopicSquareImpl struct {
	Resp *topicV2.RcmdNewTopicsRsp
}

func (ntsi *NewTopicSquareImpl) ToDynV2TopicList(c AppDynamicConfig) (res *api.TopicList) {
	if ntsi == nil || len(ntsi.Resp.GetTopicList()) == 0 {
		return nil
	}
	topics := ntsi.Resp
	var pos int64
	var list []*api.TopicListItem
	for _, t := range topics.GetTopicList() {
		pos++
		list = append(list, &api.TopicListItem{
			Icon:        t.IconUrl,                   // 尾标 "热" "新" 这种
			IconTitle:   t.GetRcmdReason().GetText(), // 上面图标的文案
			TopicId:     t.TopicId,
			TopicName:   t.TopicName,
			Url:         t.JumpUrl,
			Pos:         pos,
			ServerInfo:  t.ServerInfo,
			HeadIconUrl: c.GetIconModuleExtendNewTopic(),
			UpMid:       t.UpId,
			Extension:   t.LancerInfo,
			Position:    int64(t.Position),
		})
	}
	res = &api.TopicList{
		Title:         topics.GetRcmdHead().GetMainTitle(),
		SubTitle:      topics.RcmdHead.GetSubTitle(),
		TopicListItem: list,
	}
	if moreBtn := topics.GetRcmdHead().GetMoreButton(); moreBtn != nil {
		res.MoreButton = &api.TopicButton{
			JumpUri: moreBtn.GetUrl(),
			RedDot:  moreBtn.GetFlag(),
		}
		res.MoreButton.Icon, res.MoreButton.Title = c.GetResDynMixTopicSquareMore()
		if len(moreBtn.GetText()) > 0 {
			res.MoreButton.Title = moreBtn.GetText()
		}
	}
	if actBtn := topics.GetRcmdHead().GetActButton(); actBtn != nil {
		res.ActButton = &api.TopicButton{
			Icon:    actBtn.GetIcon(),
			Title:   actBtn.GetText(),
			JumpUri: actBtn.GetUrl(),
			RedDot:  actBtn.GetFlag(),
		}
	}
	return
}

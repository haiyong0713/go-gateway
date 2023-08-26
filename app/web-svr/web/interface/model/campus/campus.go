package campus

import (
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/app/web-svr/web/interface/model/rcmd"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"

	arcmdl "go-gateway/app/app-svr/archive/service/api"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	campusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

const (
	Campus_Dy_Draw_Type = 2 // 图文稿件
	Campus_Dy_Arc_Type  = 8 // 视频稿件
)

/******************—————————————————————————————— req struct ————————————————————————————————————******************/
type CampusRcmdReq struct {
	Mid        int64   `form:"-"`
	CampusId   int64   `form:"campus_id"`
	CampusName string  `form:"campus_name"`
	Lat        float64 `form:"lat"`
	Lng        float64 `form:"lng"`
}

// account/dynamics 通用请求参数
type CampusOfficialReq struct {
	Mid        int64  `form:"-"`
	CampusId   uint64 `form:"campus_id"`
	CampusName string `form:"campus_name"`
	Offset     uint64 `form:"offset"`
}

type CampusRedDotReq struct {
	Mid      int64  `form:"-"`
	CampusId uint64 `form:"campus_id"`
}

type CampusFeedbackReq struct {
	// 包含多种反馈信息
	Infos string `form:"infos"`
	// 反馈来源 0:校友圈 1:校园十大榜单 2:校园话题讨论
	From int32 `form:"from"`
	// 反馈用户Id
	Mid int64 `form:"-"`
	// 发送的内容
	List []*CampusFeedbackInfo `form:"-"`
}

type CampusFeedbackInfo struct {
	BizType  int64  `json:"biz_type"`
	BizId    string `json:"biz_id"`
	CampusId int64  `json:"campus_id"`
	Reason   string `json:"reason"`
}

type CampusNearbyRcmdReq struct {
	CampusId    int    `form:"campus_id"` //用户所在的学校id
	Pn          int    `form:"pn"`
	Ps          int    `form:"ps" default:"30"`
	FreshType   int    `form:"fresh_type"`    //1：用户进入tab页，自动刷新	2：顶部下拉刷新	3：正常下滑
	PreCampusId int    `form:"pre_campus_id"` //上一页面的校园id
	Mid         int64  `form:"-"`
	Buvid       string `form:"-"`
	Ip          string `form:"-"`
}

/******************———————————————————————————————— res struct ————————————————————————————————————******************/
type PagesReply struct {
	PageType      uint32                     `json:"page_type"`  //页面类型（1:已开通学校的主推荐页 major 2:学校未开通或者未选择学校的次推荐页 minor )
	MajorPageInfo *MajorPageInfo             `json:"major_page"` // 主页信息
	RcmdPageInfo  *campusgrpc.NearbyRcmdInfo `json:"rcmd_page"`  // 推荐页
}

type MajorPageInfo struct {
	CampusId         uint64                      `json:"campus_id"`         //校园ID
	CampusName       string                      `json:"campus_name"`       //校园名
	InviteDesc       string                      `json:"invite_desc"`       // 召唤校友文案
	CampusBadge      string                      `json:"campus_badge"`      // 校徽
	CampusBackground string                      `json:"campus_background"` // 学校背景图
	CampusMotto      string                      `json:"campus_motto"`      // 校训
	TabInfo          []*TabInfo                  `json:"tab_info"`          // Tab栏
	BannerInfo       []*BannerInfo               `json:"banner_info"`       // 顶部banner位
	TopicSquareInfo  *campusgrpc.TopicSquareInfo `json:"topic_square"`      //话题广场信息
}

type TabInfo struct {
	TabName   string `json:"tab_name"`   // 名称
	TabType   int64  `json:"tab_type"`   // 类型（1: 校友圈 2: 入校必看 3: 官方号 4：十大热榜 5：话题）用于区分不同的uri
	RedDot    int32  `json:"red_dot"`    // 是否有红点提示（0：没有 1：有）
	IconUrl   string `json:"icon_url"`   // icon链接）
	TabStatus int32  `json:"tab_status"` // tab灰度状态 （0：灰度开放 1：不开放）
}

type BannerInfo struct {
	PicUrl  string `json:"pic_url"`  // 图片链接
	JumpUrl string `json:"jump_url"` // 跳转链接
}

type TopicSquareInfo struct {
	Title      string         `json:"title"`       // 广场标题
	ButtonDesc string         `json:"button_desc"` // 按钮文案
	ButtonUrl  string         `json:"button_url"`  // 按钮链接
	RcmdCard   *TopicRcmdCard `json:"rcmd_card"`   // 校园话题推荐大卡
}

type TopicRcmdCard struct {
	TopicId    uint64 `json:"topic_id"`    // 话题ID
	TopicName  string `json:"topic_name"`  // 话题名
	TopicLink  string `json:"topic_link"`  // 话题跳链
	Rid        uint64 `json:"rid"`         // 业务ID
	Type       uint32 `json:"type"`        // 类型
	ButtonType uint32 `json:"button_type"` // 按钮类型：0：不展示 1：去拍摄 2：去投稿 3：去讨论
	ButtonDesc string `json:"button_desc"` // 按钮文案
	UpdateDesc string `json:"update_desc"` // 更新提示文案
}

type SchoolSearchRep struct {
	Results []*campusgrpc.CampusInfo `json:"results"`
	HasMore bool                     `json:"has_more"`
	Offset  uint64                   `json:"offset"`
}

type OfficialAccountInfo struct {
	AccountInfo *model.AccountCard `json:"account_info"` // 账号信息
	Follower    int64              `json:"follower"`     // 粉丝数
}

func (Info *OfficialAccountInfo) FromCard(card *accmdl.Card) {
	if Info.AccountInfo == nil {
		Info.AccountInfo = &model.AccountCard{}
	}
	Info.AccountInfo.FromCard(card)
}

type OfficialDynamicsReply struct {
	HasMore   int                     `json:"has_more"`   // 是否有更多
	Offset    int                     `json:"offset"`     // 偏移
	RcmdItems []*OfficialDynamicsItem `json:"rcmd_items"` // 推荐的稿件
}

type OfficialDynamicsItem struct {
	ArcInfo *rcmd.Item `json:"arc_info"` // 推荐的稿件信息
	DyId    int        `json:"dy_id"`    // 动态id
	Desc    string     `json:"desc"`     // 推荐理由
}

func (Info *OfficialDynamicsItem) FromArc(arc *arcmdl.Arc) {
	if arc == nil {
		return
	}
	if Info.ArcInfo == nil {
		Info.ArcInfo = &rcmd.Item{}
	}
	Info.ArcInfo.FromArc(arc, nil)
}

type CampusFeedbackReply struct {
	// 消息
	Message string `json:"message"` // 下发的消息
}

type CampusBillBoardReply struct {
	// 榜单标题文案（例如：bilibili校园热点）
	Title string `json:"title"`
	// 榜单标题右侧的问号icon 点击打开介绍校园热点的专栏
	HelpUri string `json:"help_uri"`
	// 校园名称
	CampusName string `json:"campus_name"`
	// 榜单生成时间 时间戳
	BuildTime int64 `json:"build_time"`
	// 当前榜单的版本标识 用于h5分享
	VersionCode string `json:"version_code"`
	// 榜单信息列表 榜单次序就是数组的顺序
	List []*CampusBillBoardRcmdItem `json:"list"`
	// 已经拼接好的榜单h5分享URL
	ShareUri string `json:"share_uri"`
	// 用于h5的学校绑定提醒banner 0:不显示 1:显示
	BindNotice int32 `json:"bind_notice"`
	// 用于端上的榜单更新toast
	UpdateToast string `json:"update_toast"`
	// 透传回去校园id 主要用于h5通过version code获取相关信息
	CampusId int64 `json:"campus_id"`
}

type Author struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type Stat struct {
	View    int32 `json:"view"`
	Like    int32 `json:"like"`
	Danmaku int32 `json:"danmaku"`
	Reply   int32 `json:"reply"`
}

type CampusBillBoardRcmdItem struct {
	CardType string  `json:"card_type"` // 类型
	ID       int64   `json:"id"`        // 稿件id
	Title    string  `json:"title"`     // 标题
	Cover    string  `json:"cover"`     // 封面
	Author   *Author `json:"author"`    // 作者
	Stat     *Stat   `json:"stat"`      // 状况
	Duration int64   `json:"duration"`  // 视频稿件的duration
	BVID     string  `json:"bvid"`      // bvid
	Cid      int64   `json:"cid"`       // 稿件cid
	Link     string  `json:"link"`      // 跳转链接
	Reason   string  `json:"reason"`    // 推荐理由
	DyId     string  `json:"dyid"`      // 动态id
}

func (item *CampusBillBoardRcmdItem) FromArc(arc *arcmdl.Arc) {
	if arc == nil {
		return
	}
	item.CardType = "archive"
	item.ID = arc.Aid
	bvid, err := bvid.AvToBv(arc.Aid)
	if err != nil {
		log.Error("日志告警 AvToBv aid:%v,error:%+v", arc.Aid, err)
	}
	item.Title = arc.GetTitle()
	item.Cover = arc.GetPic()
	item.Duration = arc.GetDuration()
	item.Stat = &Stat{View: arc.Stat.View, Danmaku: arc.Stat.Danmaku, Like: arc.Stat.Like, Reply: arc.Stat.Reply}
	item.Author = &Author{Mid: arc.Author.Mid, Face: arc.Author.Face, Name: arc.Author.Name}
	item.Cid = arc.GetFirstCid()
	item.BVID = bvid
	item.Link = "https://www.bilibili.com/video/" + bvid
}

func (item *CampusBillBoardRcmdItem) FromDraw(dynamic *model.DynamicCard, rid int64, detail *model.DrawDetail) {
	title := detail.Item.Description
	titleRune := []rune(title)
	titleRuneMaxLen := 50
	if len(titleRune) > titleRuneMaxLen {
		title = string(titleRune[:titleRuneMaxLen])
	}
	if len(title) <= 0 {
		title = "图文动态"
	}
	item.CardType = "dynamic"
	item.ID = rid
	item.Title = title
	item.Stat = &Stat{View: dynamic.Desc.View, Like: dynamic.Desc.Like, Reply: int32(detail.Item.Reply)}
	item.Author = &Author{Mid: dynamic.Desc.UserProfile.Info.UID, Face: dynamic.Desc.UserProfile.Info.Face, Name: dynamic.Desc.UserProfile.Info.UName}
	// item.DyId = dynamic.Desc.DynamicID
	for _, pic := range detail.Item.Pictures {
		item.Cover = pic.ImgSrc
		break
	}
	item.Link = "https://t.bilibili.com/" + strconv.FormatInt(dynamic.Desc.DynamicID, 10)
}

type CampusNearbyRcmdReply struct {
	Items       []*rcmd.Item `json:"item"`
	UserFeature string       `json:"user_feature"`
}

type CampusRedDotReply struct {
	RedDot bool `json:"red_dot"`
}

/******************——————————————————————————————  methods  ————————————————————————————————————******************/
func FromMajorPageInfo(in *campusgrpc.MajorPageInfo) (res *MajorPageInfo) {
	if in == nil {
		log.Warn("【@FromMajorPageInfo】MajorPage is null")
		return nil
	}
	res = &MajorPageInfo{}
	res.CampusId = in.CampusId
	res.CampusName = in.CampusName
	res.InviteDesc = in.InviteDesc
	res.CampusBackground = in.CampusBackground
	res.CampusBadge = in.CampusBadge
	res.CampusMotto = in.CampusMotto
	for _, banner := range in.Banner {
		bannerInfo := &BannerInfo{
			PicUrl:  banner.PicUrl,
			JumpUrl: banner.JumpUrl,
		}
		res.BannerInfo = append(res.BannerInfo, bannerInfo)
	}
	res.TopicSquareInfo = in.TopicSquare
	for _, tab := range in.Tabs {
		tabInfo := &TabInfo{
			TabName:   tab.TabName,
			TabType:   int64(tab.TabType),
			TabStatus: int32(tab.TabStatus),
			RedDot:    tab.RedDot,
			IconUrl:   tab.IconUrl,
		}
		res.TabInfo = append(res.TabInfo, tabInfo)
	}
	return
}

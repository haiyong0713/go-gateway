package dynamicV2

import (
	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"

	newtopicgrpc "git.bilibili.co/bapis/bapis-go/topic/service"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/gogo/protobuf/types"
)

// 动态列表资源
type DynListRes struct {
	UpdateNum        int64                       `json:"update_num"`
	HistoryOffset    string                      `json:"history_offset"`
	UpdateBaseline   string                      `json:"update_baseline"`
	HasMore          bool                        `json:"has_more"`
	Dynamics         []*Dynamic                  `json:"dynamics"`           // 动态核心
	FoldInfo         *FoldInfo                   `json:"fold_info"`          // 动态折叠
	RcmdUps          *RcmdUPCard                 `json:"rcmd_ups,omitempty"` // 推荐关注用户
	RegionUps        *dyngrpc.UnLoginRsp         `json:"-"`                  // 分区聚类推荐up+视频（空列表）
	Toast            string                      `json:"-"`
	OffsetInt        int64                       `json:"-"`
	GuideBar         *dyncampusgrpc.GuideBarInfo `json:"-"` // 校园feed流引导栏信息
	CampusFeedUpdate bool                        `json:"-"`
	StoryUpCard      *dyngrpc.StoryUPCard        `json:"-"` // story卡
	RcmdInfo         *RcmdInfo                   `json:"-"` // AI校园infoc
	CampusHotTopic   *CampusHotTopicInfo         `json:"-"` // 校友圈的校园热议卡片
	YellowBars       []*dyncampusgrpc.YellowBar  `json:"-"` // 校园tab小黄条热点提醒
}

/*
*************

		动态核心
	 *************
*/
type Dynamic struct {
	DynamicID       int64                           `json:"dynamic_id"`
	Type            int64                           `json:"type"`
	Rid             int64                           `json:"rid"`
	UID             int64                           `json:"uid"`
	UIDType         int                             `json:"uid_type"`
	Repost          int64                           `json:"repost"`
	ACL             *Acl                            `json:"acl"`    // 属性
	Extend          *Extend                         `json:"extend"` // 扩展
	Tips            string                          `json:"tips"`
	Timestamp       int64                           `json:"timestamp"`
	Origin          *Dynamic                        `json:"Origin"`  // 转发源动态
	Forward         *Dynamic                        `json:"Forward"` // 转发动态
	RType           int32                           `json:"r_type"`
	SType           int64                           `json:"stype"`
	PassThrough     *PassThrough                    `json:"pass_through"` // 透传
	Visible         bool                            `json:"visible"`
	AttachCardInfos []*dyncommongrpc.AttachCardInfo `json:"attachCardInfos"` // 附加大卡
	Tags            []*dyncommongrpc.Tag            `json:"tags"`            // 附加小卡
	ViewNum         int64                           `json:"repost_num"`      // 浏览数
	Property        *dyncommongrpc.Property         `json:"-"`
	// fake 假卡数据
	FakeContent         string               `json:"-"`
	FakeCover           string               `json:"-"`
	Duration            int64                `json:"-"` //稿件总时长 单位=秒
	FakeImages          []*FakeDynamicImages `json:"-"` // 兼容加卡
	AttachCardInfosFake []*AttachCardInfo    `json:"-"` // 附加大卡
	// ext
	Desc    string `json:"-"`
	TrackID string `json:"-"`
}

// 动态核心
func (dyn *Dynamic) FromDynamic(d *dyncommongrpc.DynBrief) {
	dyn.DynamicID = d.DynId
	dyn.Type = d.Type
	dyn.Rid = d.Rid
	dyn.UID = d.Uid
	dyn.UIDType = int(d.UidType)
	dyn.Repost = d.RepostNum
	dyn.Tips = d.Tips
	dyn.Timestamp = d.Ctime
	dyn.RType = d.RType
	dyn.Visible = d.Visible
	dyn.SType = d.SType
	if d.Ext != nil {
		ext := &Extend{}
		ext.FromExtend(d.Ext)
		dyn.Extend = ext
	}
	if d.Acl != nil {
		acl := &Acl{}
		acl.FromAcl(d.Acl)
		dyn.ACL = acl
	}
	if d.Origin != nil {
		origin := &Dynamic{}
		origin.FromDynamic(d.Origin)
		dyn.Origin = origin
	}
	if d.DynStat != nil {
		dyn.ViewNum = d.DynStat.ViewNum
		dyn.Repost = d.DynStat.RepostNum
	}
	if d.PassThrough != nil {
		dyn.PassThrough = &PassThrough{
			AdSourceContent: d.PassThrough.AdSourceContent,
			PgcBadge:        d.PassThrough.PgcBadge,
			FeedBack:        d.PassThrough.FeedBack,
		}
		if d.PassThrough.AdExtra != nil {
			dyn.PassThrough.AdverMid = d.PassThrough.AdExtra.AdverMid
			dyn.PassThrough.AdContentType = d.PassThrough.AdExtra.AdContentType
			dyn.PassThrough.AdAvid = d.PassThrough.AdExtra.AdAvid
			dyn.PassThrough.AdUrlExtra = d.PassThrough.AdExtra.AdUrlExtra
		}
	}
	dyn.AttachCardInfos = d.AttachCardInfos
	dyn.Tags = d.Tags
	dyn.Property = d.Property
}

/*
属性相关
*/
type Acl struct {
	RepostBan  int64 `json:"repost_banned"`  // 禁转
	CommentBan int64 `json:"comment_banned"` // 禁评
	FoldLimit  int64 `json:"limit_display"`  // 折叠
}

func (acl *Acl) FromAcl(v *dyncommongrpc.DynAcl) {
	acl.RepostBan = BoolToInt64(v.RepostBanned)
	acl.CommentBan = BoolToInt64(v.CommentBanned)
	acl.FoldLimit = BoolToInt64(v.LimitDisplay)
}

/*
属性:
抽奖、投票、lbs、高亮、话题/新话题顶部卡、争议警示信息、附加小卡、附加大卡、外露点赞、点赞动画、附加商品卡
*/
type Extend struct {
	Lott           *Lott             `json:"lott"`
	Vote           *Vote             `json:"vote"`
	Lbs            *Lbs              `json:"lbs"`
	Ctrl           []*Ctrl           `json:"ctrl"`
	TopicInfo      *TopicInfo        `json:"topic_info"`
	NewTopic       *NewTopicHeader   `json:"new_topic_header"` // 新话题动态顶部卡片
	EmojiType      int               `json:"emoji_type"`
	Dispute        *Dispute          `json:"dispute"`
	FlagCfg        *FlagCfg          `json:"flag_cfg"` // 附加大卡
	Display        *Display          `json:"dis_play"`
	BottomBusiness []*BottomBusiness `json:"buttom_business"` // 附加小卡
	LikeIcon       *LikeIcon         `json:"like_icon"`
	OpenGoods      *OpenGoods        `json:"open_goods"`
	VideoShare     *VideoShare       `json:"video_share"` // 分享视频卡(8)的cid与page
	CampusLike     *CampusLike       `json:"campus_like"` // 校园 同学点赞外露
	TopicSet       *NewTopicSet      `json:"topic_set"`   // 新话题 话题集
}

func (ext *Extend) FromExtend(v *dyncommongrpc.DynExt) {
	ext.EmojiType = int(v.EmojiType)
	if v.Lott != nil {
		lott := &Lott{}
		lott.FromLott(v.Lott)
		ext.Lott = lott
	}
	if v.Vote != nil {
		vote := &Vote{}
		vote.FromVote(v.Vote)
		ext.Vote = vote
	}
	if v.Lbs != nil {
		lbs := &Lbs{}
		lbs.FromLbs(v.Lbs)
		ext.Lbs = lbs
	}
	for _, ctrl := range v.HighLight {
		ctrlTmp := &Ctrl{}
		ctrlTmp.FromCtrl(ctrl)
		ext.Ctrl = append(ext.Ctrl, ctrlTmp)
	}
	if v.TopicInfo != nil {
		topic := &TopicInfo{}
		topic.FromTopicInfo(v.TopicInfo)
		ext.TopicInfo = topic
	}
	if v.NewTopic != nil {
		newTopic := &NewTopicHeader{}
		newTopic.FromNewTopic(v.NewTopic)
		ext.NewTopic = newTopic
	}
	if v.Dispute != nil {
		dispute := &Dispute{}
		dispute.FromDispute(v.Dispute)
		ext.Dispute = dispute
	}
	if v.Bottom != nil {
		for _, item := range v.Bottom.Business {
			bot := &BottomBusiness{}
			bot.FromBusiness(item)
			ext.BottomBusiness = append(ext.BottomBusiness, bot)
		}
	}
	if v.FlagCfg != nil {
		flag := &FlagCfg{}
		flag.FromFlagCfg(v.FlagCfg)
		ext.FlagCfg = flag
	}
	if len(v.LikeUsers) > 0 {
		ext.Display = &Display{
			LikeUsers: v.LikeUsers,
		}
	}
	if v.LikeShowIcon != nil {
		ext.LikeIcon = &LikeIcon{
			NewIconID: v.LikeShowIcon.GetNewIconId(),
			Begin:     v.LikeShowIcon.GetStartUrl(),
			Proc:      v.LikeShowIcon.GetActionUrl(),
			End:       v.LikeShowIcon.GetEndUrl(),
		}
	}
	if v.OpenGoods != nil {
		ext.OpenGoods = &OpenGoods{
			ItemsId:    v.GetOpenGoods().GetItemsId(),
			ShopId:     v.GetOpenGoods().GetShopId(),
			Type:       v.GetOpenGoods().GetType(),
			LinkItemId: v.GetOpenGoods().GetLinkItemId(),
			Version:    v.GetOpenGoods().GetVersion(),
		}
	}
	if v.VideoShare != nil {
		ext.VideoShare = &VideoShare{
			CID:  v.VideoShare.Cid,
			Part: v.VideoShare.Part,
		}
	}
	if v.CampusLike != nil {
		ext.CampusLike = &CampusLike{
			Users: v.CampusLike.Users,
			Total: v.CampusLike.Total,
		}
	}
	if v.TopicSet != nil && v.TopicSet.GetTopicSetPushId() != 0 && v.TopicSet.TopicSetId != 0 {
		ext.TopicSet = &NewTopicSet{
			TopicSetId: v.TopicSet.GetTopicSetId(),
			PushId:     v.TopicSet.GetTopicSetPushId(),
		}
	}
}

// 扩展：抽奖
type Lott struct {
	LotteryID   int64  `json:"lottery_id"`
	Title       string `json:"title"`
	LotteryTime int64  `json:"lottery_time"`
}

func (lott *Lott) FromLott(v *dyncommongrpc.ExtLottery) {
	lott.LotteryID = v.LotteryId
	lott.Title = v.Title
	lott.LotteryTime = v.LotteryTime
}

// 扩展投票
type Vote struct {
	VoteID int64  `json:"vote_id"`
	Title  string `json:"title"`
}

func (vote *Vote) FromVote(v *dyncommongrpc.ExtVote) {
	vote.VoteID = v.VoteId
	vote.Title = v.Title
}

// 扩展lbs
type Lbs struct {
	Address      string    `json:"address"`
	Distance     int64     `json:"distance"`
	Location     *Location `json:"location"`
	Poi          string    `json:"poi"`
	ShowTitle    string    `json:"show_title"`
	Title        string    `json:"title"`
	Type         int       `json:"type"`
	ShowDistance string    `json:"show_distance"`
}

func (lbs *Lbs) FromLbs(v *dyncommongrpc.ExtLbs) {
	lbs.Address = v.Address
	lbs.Distance = v.Distance
	lbs.Type = int(v.Type)
	lbs.Poi = v.Poi
	lbs.ShowDistance = v.ShowDistance
	lbs.ShowTitle = v.ShowTitle
	lbs.Title = v.Title
	if v.Location != nil {
		loc := &Location{}
		loc.FromLocation(v.Location)
		lbs.Location = loc
	}
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (loc *Location) FromLocation(v *dyncommongrpc.LbsLoc) {
	loc.Lat = v.Lat
	loc.Lng = v.Lng
}

type DrawTagLBS struct {
	PoiInfo *Lbs `json:"poi_info"`
}

// 扩展：高亮
type Ctrl struct {
	Length     int    `json:"length"`
	Location   int    `json:"location"`
	Type       int    `json:"type"`
	Data       string `json:"data"`
	TypeID     string `json:"type_id"`
	PrefixIcon string `json:"prefix_icon"`
}

func (c *Ctrl) FromCtrl(v *dyncommongrpc.ExtHighLight) {
	c.Length = int(v.Length)
	c.Location = int(v.Location)
	c.Type = int(v.Type)
	c.Data = v.Data
	c.TypeID = v.TypeId
	c.PrefixIcon = v.PrefixIcon
}

// 文案模块 高亮类型
func (c *Ctrl) TranType() api.DescType {
	switch c.Type {
	case CtrlTypeAite:
		return api.DescType_desc_type_aite
	case CtrlTypeLottery:
		return api.DescType_desc_type_lottery
	case CtrlTypeVote:
		return api.DescType_desc_type_vote
	case CtrlTypeGoods:
		return api.DescType_desc_type_goods
	}
	return api.DescType_desc_type_text
}

// 扩展：话题
type TopicInfo struct {
	IsAttachTopic int      `json:"is_attach_topic"`
	TopicInfos    []*Topic `json:"topic_infos"`
}

func (t *TopicInfo) FromTopicInfo(v *dyncommongrpc.ExtTopic) {
	t.IsAttachTopic = BoolToInt(v.IsAttachTopic)
	for _, item := range v.TopicInfos {
		top := &Topic{}
		top.FromTopic(item)
		t.TopicInfos = append(t.TopicInfos, top)
	}
}

type Topic struct {
	TopicID         int64
	TopicName       string
	Stat            int
	OriginTopicID   int64
	OriginTopicName string
	OriginType      int
	IsBigCard       bool
	TopicLink       string
	ShareTitle      string
	ShareImage      string
	ShareCaption    string
	TopicType       dyncommongrpc.TopicInfoType
}

func (t *Topic) FromTopic(v *dyncommongrpc.TopicInfo) {
	t.TopicID = v.TopicId
	t.TopicName = v.TopicName
	t.Stat = int(v.Stat)
	t.OriginTopicID = v.OriginTopicId
	t.OriginTopicName = v.OriginTopicName
	t.OriginType = int(v.OriginType)
	t.IsBigCard = v.IsBigCard
	t.TopicLink = v.TopicLink
	t.ShareTitle = v.ShareTitle
	t.ShareImage = v.ShareImage
	t.ShareCaption = v.ShareCaption
	t.TopicType = v.Type
}

// 扩展: 新话题动态顶部卡片
type NewTopicHeader struct {
	TopicID   int64
	TopicName string
	JumpURL   string
}

func (nth *NewTopicHeader) FromNewTopic(v *dyncommongrpc.ExtTopicV2) {
	if v.Topic == nil {
		return
	}
	nth.TopicID = v.Topic.Id
	nth.TopicName = v.Topic.Name
	nth.JumpURL = v.Topic.JumpUrl
}

type NewTopicSetDetail struct {
	SetInfo   *newtopicgrpc.TopicSetInfoRsp
	TopicList *newtopicgrpc.SetExposureTopicsRsp
}

func (ntsd *NewTopicSetDetail) IsValid() bool {
	if ntsd == nil || ntsd.SetInfo == nil || ntsd.TopicList == nil {
		return false
	}
	return true
}

func (ntsd *NewTopicSetDetail) FromSetInfo(v *newtopicgrpc.TopicSetInfoRsp) {
	ntsd.SetInfo = v
}

func (ntsd *NewTopicSetDetail) FromTopicList(v *newtopicgrpc.SetExposureTopicsRsp) {
	ntsd.TopicList = v
}

// 扩展：争议小黄条
type Dispute struct {
	Content string `json:"content"`
	Desc    string `json:"description"`
	Url     string `json:"jump_url"`
}

func (dispute *Dispute) FromDispute(v *dyncommongrpc.DynDispute) {
	dispute.Content = v.Content
	dispute.Desc = v.Description
	dispute.Url = v.JumpUrl
}

// 扩展：附加小卡
type BottomBusiness struct {
	Rid  int64 `json:"rid"`
	Type int   `json:"type"`
}

func (bot *BottomBusiness) FromBusiness(v *dyncommongrpc.BottomBusiness) {
	bot.Type = int(v.Type)
	bot.Rid = v.Rid
}

// 扩展：附加大卡
type FlagCfg struct {
	MangaID            int64 `json:"manga_id"`
	PugvID             int64 `json:"pugv_id"`
	MatchID            int64 `json:"match_id"`
	GameID             int64 `json:"game_id"`
	OGVID              int64 `json:"ogv_id"`
	DecorationID       int64 `json:"decoration_id"`
	OfficialActivityID int64 `json:"official_activity_id"`
	AvID               int64 `json:"ugc_id"`
}

func (f *FlagCfg) FromFlagCfg(v *dyncommongrpc.ExtFlagCfg) {
	if v.GetManga() != nil {
		f.MangaID = v.GetManga().GetMangaId()
	}
	if v.GetPugv() != nil {
		f.PugvID = v.GetPugv().GetPugvId()
	}
	if v.GetMatch() != nil {
		f.MatchID = v.GetMatch().GetMatchId()
	}
	if v.GetGame() != nil {
		f.GameID = v.GetGame().GetGameId()
	}
	if v.GetOgv() != nil {
		f.OGVID = v.GetOgv().GetOgvId()
	}
	if v.GetDecoration() != nil {
		f.DecorationID = v.GetDecoration().GetDecorationId()
	}
	if v.GetOfficialActivity() != nil {
		f.OfficialActivityID = v.GetOfficialActivity().GetOfficialActivityId()
	}
	if v.GetUgc() != nil {
		f.AvID = v.GetUgc().GetUgcId()
	}
}

// 扩展：点赞外露用户

type Display struct {
	LikeUsers []int64 `json:"like_users"`
}

// 扩展：点赞动画
type LikeIcon struct {
	NewIconID int64 `json:"new_icon_id"`
	// 开始动画
	Begin string `json:"begin,omitempty"`
	// 过程动画
	Proc string `json:"proc,omitempty"`
	// 结束动画
	End string `json:"end,omitempty"`
}

// 扩展：附加商品大卡
type OpenGoods struct {
	ItemsId    string `json:"itemsId,omitempty"`
	ShopId     int64  `json:"shopId,omitempty"`
	Type       int64  `json:"type,omitempty"`
	LinkItemId string `json:"linkItemId,omitempty"`
	Version    string `json:"version,omitempty"`
}

// 扩展：分享视频卡的cid和page
type VideoShare struct {
	CID  int64 `json:"cid"`
	Part int32 `json:"part"`
}

// 扩展：校园同学点赞外露
type CampusLike struct {
	Users []int64 `json:"users"`
	Total int64   `json:"total"`
}

// 扩展：新话题 话题集信息
type NewTopicSet struct {
	TopicSetId int64 `json:"topic_set_id"`
	PushId     int64 `json:"push_id"`
}

/*
*************

		动态折叠
	 *************
*/
type FoldInfo struct {
	FoldMgr     []*FoldMgr    `json:"fold_mgr"`
	InplaceFold []*FoldDetail `json:"inplace_fold"`
}

type FoldMgr struct {
	FoldType int     `json:"fold_type"`
	Folds    []*Fold `json:"Fold"`
}

type Fold struct {
	DynamicIDs []int64 `json:"dynamic_ids"`
}

type FoldDetail struct {
	Statement  string  `json:"statement"`
	DynamicIDs []int64 `json:"dynamic_ids"`
}

func (fold *FoldInfo) FromFold(v *dyncommongrpc.FoldInfo) {
	for _, item := range v.FoldMgr {
		foExt := &FoldMgr{}
		foExt.FoldType = int(item.FoldType)
		for _, item2 := range item.Folds {
			fo := &Fold{}
			fo.DynamicIDs = append(fo.DynamicIDs, item2.DynIds...)
			foExt.Folds = append(foExt.Folds, fo)
		}
		fold.FoldMgr = append(fold.FoldMgr, foExt)
	}
	for _, item := range v.InplaceFold {
		foDetail := &FoldDetail{}
		foDetail.Statement = item.Statement
		foDetail.DynamicIDs = append(foDetail.DynamicIDs, item.DynIds...)
		fold.InplaceFold = append(fold.InplaceFold, foDetail)
	}
}

/*
***************

		推荐关注用户
	 ***************
*/
type RcmdUPCard struct {
	// 接口返回，透传
	TrackId string `json:"-"`
	// 推荐up列表的类型； 	空动态列表 =1；低关注 =2
	Type int32 `json:"-"`
	// 出现在feed流列表的插入位置
	Pos int32 `json:"-"`
	// 推荐up主列表
	Users []*RcmdUser `json:"-"`
	// 监控日志
	Mids []int64 `json:"mids"`
	// 动态服务端上报
	ServerInfo string `json:"-"`
}

type RcmdUser struct {
	// 用户id
	Uid int64 `json:"uid,omitempty"`
	// 推荐信息
	Recommend *RecommendInfo `json:"recommend,omitempty"`
}

type RecommendInfo struct {
	//推荐理由
	Reason    string `json:"reason,omitempty"`
	Tid       int64  `json:"tid,omitempty"`
	SecondTid int64  `json:"second_tid,omitempty"`
}

func (rc *RcmdUPCard) FromRcmdUPCard(drc *dyngrpc.RcmdUPCard) {
	rc.TrackId = drc.TrackId
	rc.Type = drc.Type
	rc.Pos = drc.Pos
	rc.ServerInfo = drc.ServerInfo
	for _, rcmduser := range drc.Users {
		if rcmduser == nil || rcmduser.Uid == 0 {
			log.Warn("FromRcmdUPCard get error usrinfo %v", rcmduser)
			continue
		}
		rc.Mids = append(rc.Mids, rcmduser.Uid)
		user := &RcmdUser{
			Uid: rcmduser.Uid,
		}
		if rcmduser.GetRecommend() != nil {
			user.Recommend = &RecommendInfo{
				Reason:    rcmduser.GetRecommend().GetReason(),
				Tid:       rcmduser.GetRecommend().GetTid(),
				SecondTid: rcmduser.GetRecommend().GetSecondTid(),
			}
		}
		rc.Users = append(rc.Users, user)
	}
}

type PassThrough struct {
	AdSourceContent *types.Any                 `json:"ad_source_content"`
	PgcBadge        *dyncommongrpc.PgcBadge    `json:"pgc_badge"`
	FeedBack        *dyncommongrpc.DynFeedback `json:"feed_back"`
	AdverMid        int64                      `json:"-"`
	AdContentType   int32                      `json:"-"`
	AdAvid          int64                      `json:"-"`
	AdUrlExtra      string                     `json:"-"`
}

type DynDetailRes struct {
	Dynamic   *Dynamic                   `json:"-"` // 动态核心
	Recommend *dyncommongrpc.RelatedRcmd `json:"-"` // 相关推荐
}

type RepostListRes struct {
	Dynamics   []*Dynamic `json:"-"` // 动态核心
	Offset     string     `json:"-"`
	HasMore    bool       `json:"-"`
	TotalCount int64      `json:"-"`
}

func (dyn *RepostListRes) FromRepostList(rep *dyngrpc.RepostListRsp) {
	dyn.Offset = rep.Offset
	if rep.HasMore == 1 {
		dyn.HasMore = true
	}
	dyn.TotalCount = rep.TotalCount
	for _, v := range rep.Dyns {
		if v.Type == 0 {
			continue
		}
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(v)
		dyn.Dynamics = append(dyn.Dynamics, dynTmp)
	}
}

type AttachCardInfo struct {
	AttachCard *dyncommongrpc.AttachCardInfo
	Index      int `json:"-"`
}

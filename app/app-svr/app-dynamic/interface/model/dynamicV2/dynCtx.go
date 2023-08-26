package dynamicV2

import (
	"fmt"
	"strconv"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlaccount "go-gateway/app/app-svr/app-dynamic/interface/model/account"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	cheesemdl "go-gateway/app/app-svr/app-dynamic/interface/model/cheese"
	comicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/comic"
	gamemdl "go-gateway/app/app-svr/app-dynamic/interface/model/game"
	medialistmdl "go-gateway/app/app-svr/app-dynamic/interface/model/medialist"
	musicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/music"
	mdlpgc "go-gateway/app/app-svr/app-dynamic/interface/model/pgc"
	"go-gateway/app/app-svr/app-dynamic/interface/model/shopping"
	submdl "go-gateway/app/app-svr/app-dynamic/interface/model/subscription"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	ugcseasongrpc "go-gateway/app/app-svr/ugc-season/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	articleMdl "git.bilibili.co/bapis/bapis-go/article/model"
	bcgvo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	cmtGrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	shareApi "git.bilibili.co/bapis/bapis-go/community/interface/share"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicextgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic-ext"
	garbmdl "git.bilibili.co/bapis/bapis-go/garb/model"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomfeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	dramaseasongrpc "git.bilibili.co/bapis/bapis-go/maoer/drama/dramaseason"
	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	esportgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	pangugrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	passportgrpc "git.bilibili.co/bapis/bapis-go/passport/service/user"
	pgcInlineGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"
	pgcEpisodeGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	pgcSeasonGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	playurlgrpc "git.bilibili.co/bapis/bapis-go/playurl/service"
	videogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	// 起飞内容
	_adContentFly = 1
)

// DynamicContext 动态处理上下文
type DynamicContext struct {
	ForwardFrom       string                     // 转发/分享 入口标记
	From              string                     // 入口标记 处理某些特异化逻辑
	*api.DynamicItem                             // 半成品动态卡
	Dyn               *Dynamic                   // 动态原始数据
	Interim           *Interim                   // 内部通信
	ResFolds          map[string]*FoldResItem    // 折叠信息
	Grayscale         map[string]int             // 灰度实验
	Recmd             *dyncommongrpc.RelatedRcmd // 相关推荐
	HideThreeDecorate bool                       // 是否三点展示装扮
	SearchWords       []string                   // 搜索词
	SearchWordRed     bool                       // 搜索词飘红
	CampusID          int64                      // 校园id
	// Rcmd *dyncommongrpc.
	// 物料详情
	ResLike              map[string]*thumgrpc.MultiStatsReply_Records         // 点赞外露、点赞计数
	ResReply             map[string]*cmtGrpc.DynamicFeedReplyMeta             // 评论外露、评论计数
	ResUser              map[int64]*accountgrpc.Card                          // 用户信息
	ResUserFixedLocation map[int64]string                                     // 用户IP属地展示豁免信息（val不为空则为指定地址显示）
	ResManagerIpDisplay  map[int64]string                                     // 管理平台指定的IP属地信息
	ResIP2Loc            map[string]*locgrpc.InfoComplete                     // IP到属地转换信息
	ResUserFreqLocation  map[int64]*passportgrpc.UserActiveLocationReply      // 用户经常登录地址
	ResUserProfileStat   map[int64]*accountgrpc.ProfileStatReply              // 用户profile  with stat
	ResUserLive          map[int64]*livexroom.Infos                           // feed流用户直播态、0关注/低关注用户直播态
	ResUserLivePlayUrl   map[int64]*livexroom.LivePlayUrlData                 // 直播秒开地址
	ResArchive           map[int64]*archivegrpc.ArcPlayer                     // 稿件详情
	ResArcPart           map[int64]*archivegrpc.Page                          // 稿件分p信息（必要的情况下才获取）
	ResPGC               map[int32]*pgcInlineGrpc.EpisodeCard                 // OGV卡/转发OGV卡、附加小卡：OGV卡
	ResCheeseBatch       map[int64]*mdlpgc.PGCBatch                           // 付费批次卡
	ResCheeseSeason      map[int64]*mdlpgc.PGCSeason                          // 付费系列卡
	ResWords             map[int64]string                                     // 转发卡文案、纯文字卡文案
	ResDraw              map[int64]*DrawDetailRes                             // 图文卡
	ResArticle           map[int64]*articleMdl.Meta                           // 专栏卡
	ResMusic             map[int64]*musicmdl.MusicResItem                     // 音频卡
	ResCommon            map[int64]*DynamicCommonCard                         // 通用卡方/竖
	ResLive              map[int64]*livexroomgate.EntryRoomInfoResp_EntryList // 直播分享卡
	ResMedialist         map[int64]*medialistmdl.FavoriteItem                 // 播单卡
	ResAD                map[int64]*bcgvo.DynamicAdDto                        // 广告卡
	ResApple             map[int64]*dyncommongrpc.ProgramItem                 // 小程序卡
	ResSub               map[int64]*submdl.Subscription                       // 旧订阅卡
	ResLiveRcmd          map[int64]*livexroomfeed.HistoryCardInfo             // 开播卡
	ResUGCSeason         map[int64]*ugcseasongrpc.Season                      // 合集卡
	ResMyDecorate        map[int64]*mdlaccount.DecoCards                      // author模块装扮、三点逻辑
	ResBatch             map[int64]*comicmdl.Batch                            // 追漫卡
	ResBatchIsFav        map[int64]bool                                       // 追漫卡 是否追漫
	ResAdditionalTopic   map[int64][]*Topic
	ResNewTopic          map[int64]*NewTopicHeader                            // 新话题顶部卡
	ResNewTopicSet       map[int64]*NewTopicSetDetail                         // 新话题-话题集订阅更新卡 key是 push_id
	ResRelation          map[int64]int32                                      // 用户关注关系(不包括悄悄关注
	ResRelationUltima    map[int64]*relationgrpc.InterrelationReply           // 用户关注关系(包括悄悄关注)
	ResStat              map[int64]*relationgrpc.StatReply                    // 粉丝数
	ResGood              map[int64]map[int]map[string]*bcgmdl.GoodsItem       // 商品高亮、附加大卡商品
	ResExtendBBQ         map[int64]struct{}                                   // 服务端数据：是否出附加小卡BBQ
	ResSubNew            map[int64]*tunnelgrpc.DynamicCardMaterial            // 新订阅卡
	ResActivityRelation  map[int64]*activitygrpc.ActRelationInfoReply         // 通用附加卡普通活动
	ResNativePage        map[int64]*natpagegrpc.NativePageCard                // 附加大卡通用卡话题活动
	NativeAllPageCards   map[int64]*natpagegrpc.NativePageCard                // 附加大卡通用卡UP主发起活动
	ResVote              map[int64]*dyncommongrpc.VoteInfo                    // 附加大卡投票
	ResAdditionalOGV     map[int64]*pgcDynGrpc.FollowCardProto                // 附加大卡通用卡OGV、附加小卡自动OGV
	ResManga             map[int64]*comicmdl.Comic                            // 附加大卡通用卡漫画
	ResPUgv              map[int64]*cheesemdl.Cheese                          // 附加大卡通用卡付费课程
	ResMatch             map[int64]*esportgrpc.ContestDetail                  // 附加大卡通用卡电竞
	ResGame              map[int64]*gamemdl.Game                              // 附加大卡通用卡游戏
	ResDecorate          map[int64]*garbmdl.DynamicGarbInfo                   // 附加大卡通用卡装扮
	ResAttachedPromo     map[int64]int64                                      // 附加大卡通用卡帮推：topicid
	ResActivity          map[int64]*natpagegrpc.NativePage                    // 附加大卡通用卡帮推：活动详情
	ResTopicAdditiveCard map[int64]*dyntopicextgrpc.TopicAdditiveCard         // 附加大卡通用卡话题活动
	ResBiliCut           map[int64]*videogrpc.DynamicView                     // 附加小卡必剪
	ResUpActRelationInfo map[int64]*activitygrpc.UpActReserveRelationInfo     // 附加大卡UP主预约卡
	ResGameAct           map[int64]*gamemdl.Game                              // 附加大卡通用卡游戏
	ResUpActReserveDove  map[int64]*activitygrpc.ReserveDoveActRelationInfo   // 鸽子蛋
	ShareChannel         []*shareApi.ShareChannel                             // 分享组件
	ResLiveSessionInfo   map[string]*livexroomgate.SessionInfos               // 直播预约数据
	ResArcs              map[int64]*archivegrpc.Arc                           // 不要秒开信息
	ResEntryLiveUids     map[int64]*livexroomgate.EntryRoomInfoResp_EntryList // 直播推荐卡(召回)
	ResCreativeIDs       map[int64]int64                                      // 广告创意ID获取物料ID
	ResSearchChannels    map[int64]*channelgrpc.SearchChannelCard             // 频道垂搜
	ResSearchChannelMore []*channelgrpc.RelativeChannel
	ResSearchChannelHot  *channelgrpc.ChannelListReply
	ResDynSimpleInfos    map[int64]*dyngrpc.DynSimpleInfo
	ResFeedCardDramaInfo map[int64]*dramaseasongrpc.FeedCardDramaInfo // 猫儿
	ResPlayUrlCount      map[int64]*playurlgrpc.PlayOnlineReply
	ResManTianXinm       map[int64]*shopping.CardInfo
	ShoppingItems        map[int64]*shopping.Item
	ResNFTBatchInfo      map[int64]*membergrpc.NFTBatchInfoData // 批量获取nft信息
	ResNFTRegionInfo     map[string]*pangugrpc.NFTRegion
	// 回填的数据
	Emoji              map[string]struct{}                         // 正文和外露emoji缓存
	ResEmoji           map[string]*EmojiItem                       // emoji详情
	BackfillAvID       map[string]struct{}                         // 正文高亮id转title: avid缓存
	BackfillBvID       map[string]struct{}                         // 正文高亮id转title: bvid缓存
	ResBackfillArchive map[int64]*archivegrpc.ArcPlayer            // 正文高亮id转title: 稿件详情
	BackfillCvID       map[string]struct{}                         // 正文高亮id转title: cvid缓存
	ResBackfillArticle map[int64]*articleMdl.Meta                  // 正文高亮id转title: 专栏详情
	BackfillDescURL    map[string]*BackfillDescURLItem             // 正文高亮id转title: 长短链缓存
	ResBackfillSeason  map[int32]*pgcSeasonGrpc.CardInfoProto      // 正文高亮id转title: season详情
	ResBackfillEpisode map[int32]*pgcEpisodeGrpc.EpisodeCardsProto // 正文高亮id转title: ep详情
}

type BackfillDescURLItem struct {
	Type  api.DescType `json:"-"`
	Title string       `json:"-"`
	Rid   string       `json:"-"`
}

type Interim struct {
	VoteID                int64 // 记录当前卡片已露出的投票 用于保持投票唯一和互斥
	UName                 string
	Desc                  string
	DynTypeShell          int64  // 外层类型
	ShellRID              int64  // 外层物料ID
	DynTypeKernel         int64  // 内层类型
	KernelRID             int64  // 内层物料ID
	PromoURI              string // 帮推URL
	ForwardOrigFaild      bool   // 转发卡原卡失效
	HiddenAuthorLive      bool   // 隐藏author模块的直播标记
	IsPassCard            bool   // 跳过当前卡片
	IsPassAddition        bool   // 跳过附加大卡
	IsPassExtend          bool   // 跳过附加小卡
	IsPassExtendGameTopic bool   // 跳过附加游戏小卡
	IsPassExtendTopic     bool   // 跳过附加话题小卡
	CID                   int64  // 分享视频卡cid
}

/*
评论相关
*/
func (dyn *Dynamic) GetReplyID() string {
	if dyn.IsForward() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsAv() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeAv)
	}
	if dyn.IsCheeseBatch() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeCheese)
	}
	if dyn.IsWord() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsDraw() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeDraw)
	}
	if dyn.IsArticle() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeArticle)
	}
	if dyn.IsMusic() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeMusic)
	}
	if dyn.IsMedialist() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeMedialist)
	}
	if dyn.IsCommon() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsSubscription() || dyn.IsSubscriptionNew() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsApplet() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsLiveRcmd() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsAD() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeAD)
	}
	if dyn.IsUGCSeason() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeAv)
	}
	if dyn.IsBatch() {
		return fmt.Sprintf("%v,%v", dyn.DynamicID, CmtTypeDynamic)
	}
	if dyn.IsAD() && dyn.PassThrough != nil && dyn.PassThrough.AdContentType == _adContentFly && dyn.PassThrough.AdAvid > 0 {
		return fmt.Sprintf("%v,%v", dyn.PassThrough.AdAvid, CmtTypeAv)
	}
	if dyn.IsCourUp() {
		return fmt.Sprintf("%v,%v", dyn.Rid, CmtTypeCheese)
	}
	return ""
}

func GetPGCReplyID(pgc *pgcInlineGrpc.EpisodeCard) string {
	if pgc.Aid != 0 {
		return fmt.Sprintf("%v,%v", pgc.Aid, CmtTypeAv)
	}
	return ""
}

func (d *DynamicContext) GetReply() (*cmtGrpc.DynamicFeedReplyMeta, bool) {
	var replyID = d.Dyn.GetReplyID()
	if d.Dyn.IsPGC() {
		if pgc, ok := d.GetResPGC(int32(d.Dyn.Rid)); ok {
			replyID = GetPGCReplyID(pgc)
		}
	}
	if d.ResReply == nil || d.ResReply[replyID] == nil {
		return nil, false
	}
	return d.ResReply[replyID], true
}

/*
	点赞相关
*/

type ThumbsRecord struct {
	OrigID int64 `json:"origin_id"`
	MsgID  int64 `json:"message_id"`
}

func (dyn *Dynamic) GetLikeID() (*ThumbsRecord, string, bool) {
	if dyn.IsForward() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsAv() {
		item := &ThumbsRecord{MsgID: dyn.Rid}
		return item, BusTypeVideo, true
	}
	if dyn.IsCheeseBatch() {
		item := &ThumbsRecord{MsgID: dyn.Rid}
		item.OrigID = dyn.UID
		return item, BusTypeCheese, true
	}
	if dyn.IsWord() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsDraw() {
		item := &ThumbsRecord{MsgID: dyn.Rid}
		return item, BusTypeDraw, true
	}
	if dyn.IsArticle() {
		item := &ThumbsRecord{MsgID: dyn.Rid}
		return item, BusTypeArticle, true
	}
	if dyn.IsMusic() {
		item := &ThumbsRecord{MsgID: dyn.Rid, OrigID: 0}
		return item, BusTypeAudio, true
	}
	if dyn.IsCommon() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsCheeseSeason() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsLive() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsMedialist() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsAD() {
		item := &ThumbsRecord{MsgID: dyn.Rid}
		return item, BusTypeAD, true
	}
	if dyn.IsApplet() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsSubscription() || dyn.IsSubscriptionNew() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsLiveRcmd() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsUGCSeason() {
		item := &ThumbsRecord{MsgID: dyn.Rid}
		return item, BusTypeVideo, true
	}
	if dyn.IsBatch() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	if dyn.IsCourUp() {
		item := &ThumbsRecord{MsgID: dyn.DynamicID}
		return item, BusTypeDyn, true
	}
	return nil, "", false
}

func GetPGCLikeID(pgc *pgcInlineGrpc.EpisodeCard) (*ThumbsRecord, string, bool) {
	if pgc.Aid != 0 {
		return &ThumbsRecord{MsgID: pgc.Aid}, BusTypeVideo, true
	}
	return nil, "", false
}

/*
核心物料
*/
func (d *DynamicContext) GetUser(mid int64) (*accountgrpc.Card, bool) {
	if d.ResUser == nil || d.ResUser[mid] == nil {
		return nil, false
	}
	return d.ResUser[mid], true
}

func (d *DynamicContext) GetManagerIpDisplay(dynid int64) (string, bool) {
	if d.ResManagerIpDisplay == nil {
		return "", false
	}
	ipaddr, ok := d.ResManagerIpDisplay[dynid]
	return ipaddr, ok
}

func (d *DynamicContext) GetDynamicID() (int64, bool) {
	if d.Dyn == nil {
		return 0, false
	}
	return d.Dyn.DynamicID, true
}

func (d *DynamicContext) GetResUserLive(mid int64) (*livexroom.Infos, bool) {
	if d.ResUserLive == nil || d.ResUserLive[mid] == nil {
		return nil, false
	}
	return d.ResUserLive[mid], true
}

func (d *DynamicContext) GetResUserLivePlayURL(mid int64) (*livexroom.LivePlayUrlData, bool) {
	if d.ResUserLivePlayUrl == nil || d.ResUserLivePlayUrl[mid] == nil {
		return nil, false
	}
	return d.ResUserLivePlayUrl[mid], true
}

func (d *DynamicContext) GetArchive(aid int64) (*archivegrpc.ArcPlayer, bool) {
	if d.ResArchive == nil || d.ResArchive[aid] == nil || d.ResArchive[aid].Arc == nil {
		return nil, false
	}
	return d.ResArchive[aid], true
}

func (d *DynamicContext) GetResUpActRelationInfo(rid int64) (*activitygrpc.UpActReserveRelationInfo, bool) {
	if d.ResUpActRelationInfo == nil || d.ResUpActRelationInfo[rid] == nil {
		return nil, false
	}
	return d.ResUpActRelationInfo[rid], true
}

func (d *DynamicContext) IsVerticalArchive(ap *archivegrpc.Arc) bool {
	if ap == nil {
		return false
	}
	return ap.Dimension.Height > ap.Dimension.Width
}

func (d *DynamicContext) GetArchiveAutoPlayCid(ap *archivegrpc.ArcPlayer) int64 {
	if ap == nil || ap.Arc == nil || d == nil {
		return 0
	}
	var (
		playurl *archivegrpc.PlayerInfo
		ok      bool
	)
	if ap.PlayerInfo != nil {
		var interimCID int64
		if d.Interim != nil {
			interimCID = d.Interim.CID
		}
		if playurl, ok = ap.PlayerInfo[interimCID]; !ok {
			if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; !ok {
				playurl = ap.PlayerInfo[ap.Arc.FirstCid]
			}
		}
		if cid := playurl.GetPlayerExtra().GetCid(); cid > 0 {
			return cid
		}
	}
	return ap.Arc.FirstCid
}

func (d *DynamicContext) GetArcPart(cid int64) *archivegrpc.Page {
	if cid <= 0 || d == nil || d.ResArcPart == nil {
		return nil
	}
	return d.ResArcPart[cid]
}

func (d *DynamicContext) GetResPGC(rid int32) (*pgcInlineGrpc.EpisodeCard, bool) {
	if d.ResPGC == nil || d.ResPGC[rid] == nil {
		return nil, false
	}
	return d.ResPGC[rid], true
}

func (d *DynamicContext) GetResCheeseBatch(rid int64) (*mdlpgc.PGCBatch, bool) {
	if d.ResCheeseBatch == nil || d.ResCheeseBatch[rid] == nil {
		return nil, false
	}
	return d.ResCheeseBatch[rid], true
}

func (d *DynamicContext) GetResCheeseSeason(rid int64) (*mdlpgc.PGCSeason, bool) {
	if d.ResCheeseSeason == nil || d.ResCheeseSeason[rid] == nil {
		return nil, false
	}
	return d.ResCheeseSeason[rid], true
}

func (d *DynamicContext) GetResDraw(rid int64) (*DrawDetailRes, bool) {
	if d.ResDraw == nil || d.ResDraw[rid] == nil || d.ResDraw[rid].Item == nil {
		return nil, false
	}
	return d.ResDraw[rid], true
}

func (d *DynamicContext) GetResArticle(rid int64) (*articleMdl.Meta, bool) {
	if d.ResArticle == nil || d.ResArticle[rid] == nil {
		return nil, false
	}
	return d.ResArticle[rid], true
}

func (d *DynamicContext) GetResMusic(rid int64) (*musicmdl.MusicResItem, bool) {
	if d.ResMusic == nil || d.ResMusic[rid] == nil {
		return nil, false
	}
	return d.ResMusic[rid], true
}

func (d *DynamicContext) GetResCommon(rid int64) (*DynamicCommonCard, bool) {
	if d.ResCommon == nil || d.ResCommon[rid] == nil || d.ResCommon[rid].Sketch == nil {
		return nil, false
	}
	return d.ResCommon[rid], true
}

func (d *DynamicContext) GetResLive(rid int64) (*livexroomgate.EntryRoomInfoResp_EntryList, bool) {
	if d.ResLive == nil || d.ResLive[rid] == nil {
		return nil, false
	}
	return d.ResLive[rid], true
}

func (d *DynamicContext) GetResMedialist(rid int64) (*medialistmdl.FavoriteItem, bool) {
	if d.ResMedialist == nil || d.ResMedialist[rid] == nil {
		return nil, false
	}
	return d.ResMedialist[rid], true
}

func (d *DynamicContext) GetResApple(rid int64) (*dyncommongrpc.ProgramItem, bool) {
	if d.ResApple == nil || d.ResApple[rid] == nil {
		return nil, false
	}
	return d.ResApple[rid], true
}

func (d *DynamicContext) GetResSub(rid int64) (*submdl.Subscription, bool) {
	if d.ResSub == nil || d.ResSub[rid] == nil {
		return nil, false
	}
	return d.ResSub[rid], true
}

func (d *DynamicContext) GetResLiveRcmd(rid int64) (*livexroomfeed.HistoryCardInfo, bool) {
	if d.ResLiveRcmd == nil || d.ResLiveRcmd[rid] == nil || d.ResLiveRcmd[rid].LivePlayInfo == nil {
		return nil, false
	}
	return d.ResLiveRcmd[rid], true
}

func (d *DynamicContext) GetResUGCSeason(rid int64) (*ugcseasongrpc.Season, bool) {
	if d.ResUGCSeason == nil || d.ResUGCSeason[rid] == nil {
		return nil, false
	}
	return d.ResUGCSeason[rid], true
}

func (d *DynamicContext) GetResSubNew(rid int64) (*tunnelgrpc.DynamicCardMaterial, bool) {
	if d.ResSubNew == nil || d.ResSubNew[rid] == nil {
		return nil, false
	}
	return d.ResSubNew[rid], true
}

func (d *DynamicContext) GetResAdditionalOGV(aid int64) (*pgcDynGrpc.FollowCardProto, bool) {
	if d.ResAdditionalOGV == nil || d.ResAdditionalOGV[aid] == nil {
		return nil, false
	}
	return d.ResAdditionalOGV[aid], true
}

func (d *DynamicContext) GetAdditionalTopic() ([]*Topic, bool) {
	if len(d.ResAdditionalTopic[d.Dyn.DynamicID]) > 0 {
		return d.ResAdditionalTopic[d.Dyn.DynamicID], true
	}
	return nil, false
}

func (d *DynamicContext) GetResEntryLive(uid int64) (*livexroomgate.EntryRoomInfoResp_EntryList, bool) {
	if d.ResEntryLiveUids == nil || d.ResEntryLiveUids[uid] == nil {
		return nil, false
	}
	return d.ResEntryLiveUids[uid], true
}

func (d *DynamicContext) GetResNewTopicSet() *NewTopicSetDetail {
	if d.ResNewTopicSet == nil || len(d.ResNewTopicSet) == 0 {
		return nil
	}
	pushid := d.Dyn.GetNewTopicSetPushId()
	if pushid == 0 {
		return nil
	}
	detail, ok := d.ResNewTopicSet[pushid]
	if !ok || !detail.IsValid() {
		return nil
	}
	return detail
}

// 获取PGC卡子类型
func (dyn *Dynamic) GetPGCSubType() api.VideoSubType {
	switch dyn.Type {
	case DynTypeBangumi:
		return api.VideoSubType_VideoSubTypeBangumi
	case DynTypePGCBangumi:
		return api.VideoSubType_VideoSubTypeBangumi
	case DynTypePGCMovie:
		return api.VideoSubType_VideoSubTypeMovie
	case DynTypePGCTv:
		return api.VideoSubType_VideoSubTypeTeleplay
	case DynTypePGCGuoChuang:
		return api.VideoSubType_VideoSubTypeDomestic
	case DynTypePGCDocumentary:
		return api.VideoSubType_VideoSubTypeDocumentary
	default:
		return api.VideoSubType_VideoSubTypeNone
	}
}

func (dyn *Dynamic) GetLBS() (bool, *Lbs) {
	if dyn.Extend == nil || dyn.Extend.Lbs == nil || dyn.Extend.Lbs.Location == nil {
		return false, nil
	}
	return true, dyn.Extend.Lbs
}

func (dyn *Dynamic) GetTopicInfo() ([]*Topic, bool) {
	if dyn.Extend == nil || dyn.Extend.TopicInfo == nil || len(dyn.Extend.TopicInfo.TopicInfos) == 0 {
		return nil, false
	}
	return dyn.Extend.TopicInfo.TopicInfos, true
}

// 点赞外露
func (dyn *Dynamic) GetLikeUser() (bool, []int64) {
	if dyn.Extend == nil || dyn.Extend.Display == nil || len(dyn.Extend.Display.LikeUsers) == 0 {
		return false, nil
	}
	return true, dyn.Extend.Display.LikeUsers
}

// 校园 视频卡同学点赞外露
func (dyn *Dynamic) GetCampusLike() []int64 {
	if dyn.Extend == nil || dyn.Extend.CampusLike == nil || len(dyn.Extend.CampusLike.Users) == 0 {
		return nil
	}
	return dyn.Extend.CampusLike.Users
}

// 新话题 话题集订阅更新卡获取唯一的pushid
func (dyn *Dynamic) GetNewTopicSetPushId() int64 {
	if dyn.Rid != 0 {
		return dyn.Rid
	}
	if dyn.Extend == nil || dyn.Extend.TopicSet == nil ||
		dyn.Extend.TopicSet.PushId == 0 || dyn.Extend.TopicSet.TopicSetId == 0 ||
		dyn.Rid == 0 {
		return 0
	}
	return dyn.Extend.TopicSet.PushId
}

func (d *DynamicContext) GetBiliCut(rid int64, defaultTitle, defaultURI string) *videogrpc.DynamicView {
	biliCut, ok := d.ResBiliCut[rid]
	if !ok {
		return &videogrpc.DynamicView{
			Name:   defaultTitle,
			AppUrl: defaultURI,
		}
	}
	return biliCut
}

func (d *DynamicContext) GetPublishAddr() string {
	// 豁免/固定地址逻辑
	if d.ResUserFixedLocation != nil {
		fixedLoc, isHit := d.ResUserFixedLocation[d.Dyn.UID]
		if d.Dyn.UIDType == int(dyncommongrpc.DynUidType_DYNAMIC_UID_UP) && isHit {
			return fixedLoc
		}
	}
	if d.ResIP2Loc == nil {
		return "" // 可能依赖服务故障 不展示了
	}
	pubip := d.Dyn.Property.GetCreateIp()
	// 正常解析发布地址
	if len(pubip) > 0 {
		return d.ResIP2Loc[pubip].GetShow()
	} else {
		// 走用户空间兜底
		if d.ResUserFreqLocation != nil && d.Dyn.UIDType == int(dyncommongrpc.DynUidType_DYNAMIC_UID_UP) {
			userIP := d.ResUserFreqLocation[d.Dyn.UID]
			if userIP != nil {
				return userIP.Location
			}
		}
	}
	// 走地址解析兜底逻辑
	return d.ResIP2Loc[pubip].GetShow()
}

// 判断附加大卡稿件是否是首映前状态
func (d *DynamicContext) IsAttachCardPremiereBefore() bool {
	for _, aci := range d.Dyn.AttachCardInfos {
		if aci.CardType == dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE {
			up, ok := d.GetResUpActRelationInfo(aci.Rid)
			if !ok {
				continue
			}
			if up.Type != activitygrpc.UpActReserveRelationType_Premiere {
				continue
			}
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := d.GetArchive(aid)
			if !ok {
				continue
			}
			archive := ap.GetArc()
			premiere := archive.GetPremiere()
			if premiere != nil && premiere.State == archivegrpc.PremiereState_premiere_before {
				return true
			}
			break
		}
	}
	return false
}

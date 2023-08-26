package story

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-common/library/time"
	"go-gateway/app/app-svr/app-card/interface/model"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/report"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-feed/interface/common"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	liverankgrpc "git.bilibili.co/bapis/bapis-go/live/rankdb/v1"
	xroom "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	materialgrpc "git.bilibili.co/bapis/bapis-go/material/interface"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcFollowClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	vogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/pkg/errors"
)

const (
	EntryFromStoryFeedUpIcon             = "story_feed_upicon"
	EntryFromStoryFeedUpPanel            = "story_feed_uppanel"
	EntryFromStoryVideoUpIcon            = "story_video_upicon"
	EntryFromStoryVideoUpPanel           = "story_video_uppanel"
	EntryFromStoryDtUpicon               = "story_dt_upicon"
	EntryFromStoryDtUpPannel             = "story_dt_uppanel"
	EntryFromStoryLive                   = "story_live_card"
	EntryFromStoryLiveCloseEntryButton   = "story_live_action_close_entry_button"
	EntryFromStoryAdLive                 = "ad_story_live_card"
	EntryFromStoryAdLiveCloseEntryButton = "ad_story_live_action_close_entry_button"

	// 不感兴趣actionType
	_actionTypeNull = 0
	_actionTypeSkip = 1

	_dislikeVideoIcon   = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/e4mDdcgjh4.png"
	_dislikeUpIcon      = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/27guWatEiy.png"
	_dislikeStoryIcon   = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/vl6oCLzady.png"
	_dislikeChannelIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/M0k5Atr21S.png"
	_dislikeLiveArea    = "https://i0.hdslb.com/bfs/activity-plat/static/0977767b2e79d8ad0a36a731068a83d7/ul2yHpRe6B.png"
	_dislikeLiveIcon    = "https://i0.hdslb.com/bfs/activity-plat/static/20220307/2be2c5f696186bad80d4b452e4af2a76/Kx3nz0AFjw.png"
	_dislikeSeasonIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/20220620/0977767b2e79d8ad0a36a731068a83d7/N8RSrnT17M.png"
)

var (
	dislikeReasonV3ActionSheet = []*ActionSheet{
		{
			RecognizedName: "user-feedback",
			Icon:           "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/sDKSsbJzqi.png",
			Text:           "反馈建议",
		},
		{
			RecognizedName: "player-feedback",
			Icon:           "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/G3pMH8XOTl.png",
			Text:           "播放反馈",
		},
		{
			RecognizedName: "report",
			Icon:           "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/uwzHtpepw1.png",
			Text:           "举报",
		},
	}
	dislikeReasonV3ActionSheetFromPGC = []*ActionSheet{
		{
			RecognizedName: "user-feedback",
			Icon:           "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/sDKSsbJzqi.png",
			Text:           "反馈建议",
		},
		{
			RecognizedName: "player-feedback",
			Icon:           "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/G3pMH8XOTl.png",
			Text:           "播放反馈",
		},
	}
)

type Info struct {
	ShortLink         string                             `json:"short_link,omitempty"`
	Desc              string                             `json:"desc,omitempty"`
	PubDate           time.Time                          `json:"pubdate,omitempty"`
	Owner             *Owner                             `json:"owner,omitempty"`
	Rights            *api.Rights                        `json:"rights,omitempty"`
	Stat              *Stat                              `json:"stat,omitempty"`
	Dimension         *api.Dimension                     `json:"dimension,omitempty"`
	DislikeV2         *DislikeV2                         `json:"dislike_reasons_v2,omitempty"`
	Vip               accountgrpc.VipInfo                `json:"vip,omitempty"`
	ReqUser           ReqUser                            `json:"req_user,omitempty"`
	ShareSubtitle     string                             `json:"share_subtitle,omitempty"`
	Copyright         int32                              `json:"copyright,omitempty"`
	Label             *Label                             `json:"label,omitempty"`
	LiveRoom          *LiveRoom                          `json:"live_room,omitempty"`
	ThumbUpAnimation  string                             `json:"thumb_up_animation,omitempty"`
	ReservationInfo   *ReservationInfo                   `json:"reservation_info,omitempty"`
	DislikeReasonsV3  *DislikeReasonsV3                  `json:"dislike_reasons_v3,omitempty"`
	ArgueMsg          string                             `json:"argue_msg,omitempty"`
	ArgueIcon         string                             `json:"argue_icon,omitempty"`
	ThreePointButton  *ThreePointButton                  `json:"three_point_button,omitempty"`
	ShareBottomButton []*threePointMeta.FunctionalButton `json:"share_bottom_button,omitempty"`
	PGCInfo           *PGCInfo                           `json:"pgc_info,omitempty"`
	ContractResource  *ContractResource                  `json:"contract_resource,omitempty"`
	RcmdReason        string                             `json:"rcmd_reason,omitempty"`
}

func (i *Item) thumbupIcon(in *thumbupgrpc.LikeAnimation) {
	if in.StoryLikeIcon != "" && in.StoryLikedIcon != "" && in.LikedIcon != "" && in.LikeAnimation != "" {
		i.ThumbupIcon = &ThumbupIcon{
			HasIcon:         true,
			LikeIcon:        in.StoryLikeIcon,
			LikedIcon:       in.StoryLikedIcon,
			FullLikeIcon:    in.LikedIcon,
			TripleAnimation: in.LikeAnimation,
		}
	}
}

type ThumbupIcon struct {
	HasIcon         bool   `json:"has_icon,omitempty"`
	LikeIcon        string `json:"like_icon,omitempty"`
	LikedIcon       string `json:"liked_icon,omitempty"`
	FullLikeIcon    string `json:"full_like_icon,omitempty"`
	TripleAnimation string `json:"triple_animation,omitempty"`
}

type PGCInfo struct {
	Publisher      string `json:"publisher,omitempty"`
	EpUpdateDesc   string `json:"ep_update_desc,omitempty"`
	Ratio          string `json:"ratio,omitempty"`
	OGVStyle       int8   `json:"ogv_style,omitempty"`
	JumpURI        string `json:"jump_uri,omitempty"`
	ClipStart      int32  `json:"clip_start,omitempty"`
	ClipEnd        int32  `json:"clip_end,omitempty"`
	ShowSeasonType int32  `json:"show_season_type,omitempty"`
}

type ThreePointButton struct {
	Top    []*threePointMeta.FunctionalButton `json:"top,omitempty"`
	Bottom []*threePointMeta.FunctionalButton `json:"bottom,omitempty"`
}

type DislikeReasonsV3 struct {
	Dislike      *DislikeV3     `json:"dislike,omitempty"`
	Feedback     *FeedbackV3    `json:"feedback,omitempty"`
	ActionSheets []*ActionSheet `json:"action_sheets,omitempty"`
}

type DislikeV3 struct {
	Title        string         `json:"title,omitempty"`
	DislikeItems []*DislikeItem `json:"items,omitempty"`
}

type FeedbackV3 struct {
	Title         string          `json:"title,omitempty"`
	FeedbackItems []*FeedbackItem `json:"items,omitempty"`
}

type DislikeItem struct {
	ID         int    `json:"id,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Title      string `json:"title,omitempty"`
	SubTitle   string `json:"sub_title,omitempty"`
	UpperMid   int64  `json:"upper_mid,omitempty"`
	RID        int64  `json:"rid,omitempty"`
	TagID      int64  `json:"tag_id,omitempty"`
	Toast      string `json:"toast,omitempty"`
	ActionType int32  `json:"action_type"`
}

type FeedbackItem struct {
	ID         int    `json:"id,omitempty"`
	Icon       string `json:"icon,omitempty"`
	Title      string `json:"title,omitempty"`
	UpperMid   int64  `json:"upper_mid,omitempty"`
	RID        int32  `json:"rid,omitempty"`
	TagID      int64  `json:"tag_id,omitempty"`
	Toast      string `json:"toast,omitempty"`
	ActionType int32  `json:"action_type"`
}

type ActionSheet struct {
	RecognizedName string `json:"recognized_name,omitempty"`
	Icon           string `json:"icon,omitempty"`
	Text           string `json:"text,omitempty"`
}

type ReservationInfo struct {
	Sid               int64     `json:"sid,omitempty"`
	Name              string    `json:"name,omitempty"`
	IsFollow          int64     `json:"is_follow,omitempty"`
	LivePlanStartTime time.Time `json:"live_plan_start_time,omitempty"`
}

type LiveRoom struct {
	LiveStatus      int64  `json:"live_status,omitempty"`
	UpJumpURI       string `json:"up_jump_uri,omitempty"`
	UpPannelJumpURI string `json:"up_pannel_jump_uri,omitempty"`
	CloseButtonURI  string `json:"close_button_uri,omitempty"`
	AreaID          int64  `json:"area_id,omitempty"`
	ParentAreaID    int64  `json:"parent_area_id,omitempty"`
	LiveType        string `json:"live_type,omitempty"`
}

// DislikeV2 v2
type DislikeV2 struct {
	Title    string            `json:"title,omitempty"`
	Subtitle string            `json:"subtitle,omitempty"`
	Reasons  []*DislikeReasons `json:"reasons,omitempty"`
}

// DislikeReasons .
type DislikeReasons struct {
	ID    int    `json:"id,omitempty"`
	MID   int64  `json:"mid,omitempty"`
	RID   int32  `json:"rid,omitempty"`
	TagID int64  `json:"tag_id,omitempty"`
	Name  string `json:"name,omitempty"`
}

type Owner struct {
	Mid            int64           `json:"mid,omitempty"`
	Name           string          `json:"name,omitempty"`
	Face           string          `json:"face,omitempty"`
	OfficialVerify *OfficialInfo   `json:"official_verify,omitempty"`
	Fans           int64           `json:"fans,omitempty"`
	Attention      int64           `json:"attention,omitempty"`
	LikeNum        int64           `json:"like_num,omitempty"`
	Relation       *model.Relation `json:"relation,omitempty"`
	FaceType       model.Type      `json:"face_type,omitempty"`
}

type OfficialInfo struct {
	Role  int32  `json:"role,omitempty"`
	Title string `json:"title,omitempty"`
	Type  int32  `json:"type"`
}

type Item struct {
	Info
	CardGoto model.CardGt `json:"card_goto,omitempty"`
	Goto     model.Gt     `json:"goto,omitempty"`
	Param    string       `json:"param,omitempty"`
	Cover    string       `json:"cover,omitempty"`
	BvID     string       `json:"bvid,omitempty"`
	FfCover  string       `json:"ff_cover,omitempty"`
	Title    string       `json:"title,omitempty"`
	URI      string       `json:"uri,omitempty"`
	// app直播间背景图
	AppBackground string      `json:"app_background,omitempty"`
	PlayerArgs    *PlayerArgs `json:"player_args,omitempty"`
	// ai展现日志上报使用
	Rcmd                  *ai.SubItems      `json:"-"`
	AdInfo                *cm.AdInfo        `json:"ad_info,omitempty"`
	AdType                int32             `json:"ad_type,omitempty"`
	TrackID               string            `json:"track_id,omitempty"`
	ReportInfo            *ReportInfo       `json:"report_info,omitempty"`
	CreativeEntrance      *CreativeEntrance `json:"creative_entrance,omitempty"`
	StoryCartIcon         *CommonStoryCart  `json:"story_cart_icon,omitempty"`
	PosRecUniqueId        string            `json:"pos_rec_unique_id,omitempty"`
	LiveShowAttentionIcon bool              `json:"live_show_attention_icon,omitempty"`
	ScrollMessage         []string          `json:"scroll_message,omitempty"`
	DislikeReportData     string            `json:"dislike_report_data,omitempty"`
	ThumbupIcon           *ThumbupIcon      `json:"thumbup_icon,omitempty"`
	Duration              int64             `json:"duration,omitempty"`
}

type CommonStoryCart struct {
	IconURL   string `json:"icon_url,omitempty"`
	IconText  string `json:"icon_text,omitempty"`
	IconTitle string `json:"icon_title,omitempty"`
	URI       string `json:"uri,omitempty"`
	Goto_     string `json:"goto,omitempty"`
}

type ReportInfo struct {
	SubType  int32 `json:"sub_type,omitempty"`
	EpStatus int32 `json:"ep_status,omitempty"`
}

type StoryConfig struct {
	ProgressBar      *ProgressBar `json:"progress_bar,omitempty"`
	EnableRcmdGuide  bool         `json:"enable_rcmd_guide,omitempty"`
	SlideGuidanceAb  int8         `json:"slide_guidance_ab,omitempty"`
	ShowButton       []string     `json:"show_button"`
	EnableJumpToView bool         `json:"enable_jump_to_view,omitempty"`
	JumpToViewIcon   string       `json:"jump_to_view_icon,omitempty"`
	ReplyZoomExp     int8         `json:"reply_zoom_exp,omitempty"`
	ReplyNoDanmu     bool         `json:"reply_no_danmu,omitempty"`
	ReplyHighRaised  bool         `json:"reply_high_raised,omitempty"`
	SpeedPlayExp     bool         `json:"speed_play_exp,omitempty"`
}

type CreativeEntrance struct {
	Icon    string `json:"icon,omitempty"`
	JumpURI string `json:"jump_uri,omitempty"`
	Title   string `json:"title,omitempty"`
	Type    int64  `json:"type,omitempty"`
}

type ProgressBar struct {
	IconDrag     string `json:"icon_drag,omitempty"`
	IconDragHash string `json:"icon_drag_hash,omitempty"`
	IconStop     string `json:"icon_stop,omitempty"`
	IconStopHash string `json:"icon_stop_hash,omitempty"`
	IconZoom     string `json:"icon_zoom,omitempty"`
	IconZoomHash string `json:"icon_zoom_hash,omitempty"`
}

type PlayerArgs struct {
	Aid      int64    `json:"aid,omitempty"`
	Cid      int64    `json:"cid,omitempty"`
	Type     model.Gt `json:"type"`
	Rid      int64    `json:"rid,omitempty"`
	EpId     int32    `json:"ep_id,omitempty"`
	SeasonId int32    `json:"season_id,omitempty"`
	Duration int64    `json:"duration,omitempty"`
}

type ReqUser struct {
	Like     int8 `json:"like,omitempty"`
	Favorite int8 `json:"favorite,omitempty"`
	Coin     int8 `json:"coin,omitempty"`
	Follow   int8 `json:"follow,omitempty"`
}

type Label struct {
	Type int    `json:"type,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type IconMaterial struct {
	BcutStoryCart map[string]*materialgrpc.StoryRes
	Eppm          map[int32]*pgcinline.EpisodeCard
}

type ContractResource struct {
	IsFollowDisplay   int32         `json:"is_follow_display"`
	IsInteractDisplay int32         `json:"is_interact_display"`
	IsTripleDisplay   int32         `json:"is_triple_display"`
	ContractCard      *ContractCard `json:"contract_card"`
}

type ContractCard struct {
	Title    string `json:"title"`
	SubTitle string `json:"subtitle"`
}

func (i *Item) StoryFrom(rcmd *ai.SubItems, a *arcgrpc.ArcPlayer, cardm map[int64]*accountgrpc.Card,
	statm map[int64]*relationgrpc.StatReply, authorRelations map[int64]*relationgrpc.InterrelationReply, likeMap,
	coinMap map[int64]int64, haslike, isFav map[int64]int8, ts []*channelgrpc.Channel, hotAids map[int64]struct{},
	buildLiveRoom func() *LiveRoom, plat int8, build int, mobiApp, animation string, reservationInfo *ReservationInfo,
	arguement map[int64]*vogrpc.Argument, ffCoverFrom string, mid int64, contractResource map[int64]*ContractResource,
	likeAnimationIcon map[int64]*thumbupgrpc.LikeAnimation, fns ...StoryFn) {
	if rcmd == nil || a == nil {
		return
	}
	const (
		_hot = 1
	)
	i.TrackID = rcmd.TrackID
	i.PosRecUniqueId = rcmd.PosRecUniqueId
	i.DislikeReportData = report.BuildDislikeReportData(0, rcmd.PosRecUniqueId)
	i.CardGoto = model.CardGt(rcmd.Goto)
	i.Goto = model.GotoVerticalAv
	i.Param = strconv.FormatInt(a.Arc.Aid, 10)
	bvID, err := bvid.AvToBv(a.Arc.Aid)
	if err != nil {
		log.Error("bvid.AvToBv err(%+v)", err)
		//nolint:ineffassign
		err = nil
	}
	i.BvID = bvID
	i.Cover = a.Arc.Pic
	i.ShortLink = fmt.Sprintf(model.ShortLinkHost+"/%s", bvID)
	i.FfCover = common.Ffcover(a.Arc.FirstFrame, ffCoverFrom)
	i.Title = a.Arc.Title
	i.Rights = &a.Arc.Rights
	i.Stat = convertArchiveToStat(a.Arc.Stat)
	i.ShareSubtitle = model.ArchiveViewShareString(a.Arc.Stat.View)
	i.PubDate = a.Arc.PubDate
	i.Desc = a.Arc.Desc
	i.URI = model.FillURI(i.Goto, plat, build, i.Param, model.ArcPlayHandler(a.Arc, model.ArcPlayURL(a, 0), rcmd.TrackID, nil, build, mobiApp, true))
	i.Duration = a.Arc.Duration
	i.PlayerArgs = playerArgsFrom(a)
	i.Owner = &Owner{
		Mid:  a.Arc.Author.Mid,
		Name: a.Arc.Author.Name,
		Face: a.Arc.Author.Face,
	}
	i.Dimension = &a.Arc.Dimension
	authorCard, ok := cardm[TargetMid(rcmd, a.Arc.Author.Mid)]
	if ok {
		i.Owner.Mid = authorCard.Mid
		i.Owner.Name = authorCard.Name
		i.Owner.Face = authorCard.Face
		i.Owner.OfficialVerify = &OfficialInfo{
			Role:  authorCard.Official.Role,
			Title: authorCard.Official.Title,
			Type:  authorCard.Official.Type,
		}
		if i.Owner.OfficialVerify.Role == 7 {
			i.Owner.OfficialVerify.Role = 1
		}
		i.Vip = authorCard.Vip
		if i.Vip.Status == model.VipStatusExpire && authorCard.Vip.DueDate > 0 { //0-过期 (2-冻结 3-封禁不展示）
			i.Vip.Label.Path = model.VipLabelExpire
		}
	}
	i.LiveRoom = buildLiveRoom()
	if stat, ok := statm[TargetMid(rcmd, a.Arc.Author.Mid)]; ok {
		i.Owner.Fans = stat.Follower
		i.Owner.Attention = stat.Following
	}
	if allNum, ok := likeMap[TargetMid(rcmd, a.Arc.Author.Mid)]; ok {
		i.Owner.LikeNum = allNum
	}
	i.ReqUser.Like = haslike[a.Arc.Aid]
	i.ReqUser.Favorite = isFav[a.Arc.Aid]
	if coin, ok := coinMap[a.Arc.Aid]; ok && coin > 0 {
		i.ReqUser.Coin = 1
	}
	i.Owner.Relation = model.RelationChange(TargetMid(rcmd, a.Arc.Author.Mid), authorRelations)
	i.Copyright = a.Arc.Copyright
	if _, ok := hotAids[a.Arc.Aid]; ok {
		i.Label = &Label{
			Type: _hot,
			URI:  model.FillURI(model.GotoHotPage, plat, build, "", nil),
		}
	}
	// ai展现日志上报使用
	i.Rcmd = rcmd
	i.DislikeV2 = &DislikeV2{}
	i.DislikeV2.StoryDislike(a, ts)
	i.DislikeReasonsV3 = &DislikeReasonsV3{
		Dislike:      dislikeV3From(a, ts, rcmd, authorCard),
		Feedback:     feedbackV3From(a, ts, rcmd),
		ActionSheets: dislikeReasonV3ActionSheet,
	}
	// 填充点赞动画
	i.ThumbUpAnimation = animation
	i.ReservationInfo = reservationInfo
	if argue, ok := arguement[a.Arc.Aid]; ok {
		i.ArgueMsg = argue.ArgueMsg
		i.ArgueIcon = "https://i0.hdslb.com/bfs/activity-plat/static/0977767b2e79d8ad0a36a731068a83d7/story_argue_icon@3x.png"
	}
	i.compatibleUnloginThumbup(mid)
	// 老粉
	if resource, ok := contractResource[a.Arc.Aid]; ok {
		i.ContractResource = resource
	}
	if lm, haslm := likeAnimationIcon[a.Arc.Aid]; haslm && lm != nil {
		i.thumbupIcon(lm)
	}
	i.RcmdReason = rcmd.MutualReason
	for _, fn := range fns {
		fn(i)
	}
}

func TargetMid(item *ai.SubItems, arcUid int64) int64 {
	if item != nil && item.Uid() > 0 {
		return item.Uid()
	}
	return arcUid
}

func (i *Item) AdStoryWithMid(rcmd *ai.SubItems, cardm map[int64]*accountgrpc.Card,
	statm map[int64]*relationgrpc.StatReply, authorRelations map[int64]*relationgrpc.InterrelationReply,
	likeMap map[int64]int64, reservationInfo *ReservationInfo, buildLiveRoom func() *LiveRoom, fns ...StoryFn) {
	i.Owner = &Owner{}
	if cm, ok := cardm[rcmd.StoryUpMid]; ok {
		i.Owner.Mid = cm.Mid
		i.Owner.Name = cm.Name
		i.Owner.Face = cm.Face
		i.Owner.OfficialVerify = &OfficialInfo{
			Role:  cm.Official.Role,
			Title: cm.Official.Title,
			Type:  cm.Official.Type,
		}
		if i.Owner.OfficialVerify.Role == 7 {
			i.Owner.OfficialVerify.Role = 1
		}
		i.Vip = cm.Vip
		if i.Vip.Status == model.VipStatusExpire && cm.Vip.DueDate > 0 { //0-过期 (2-冻结 3-封禁不展示）
			i.Vip.Label.Path = model.VipLabelExpire
		}
	}
	if stat, ok := statm[rcmd.StoryUpMid]; ok {
		i.Owner.Fans = stat.Follower
		i.Owner.Attention = stat.Following
	}
	if allNum, ok := likeMap[rcmd.StoryUpMid]; ok {
		i.Owner.LikeNum = allNum
	}
	i.LiveRoom = buildLiveRoom()
	i.Owner.Relation = model.RelationChange(rcmd.StoryUpMid, authorRelations)
	i.ReservationInfo = reservationInfo
	for _, fn := range fns {
		fn(i)
	}
}

func (i *Item) AdStoryWithoutMid(fns ...StoryFn) {
	i.Owner.Mid = 0
	i.Owner.Name = ""
	i.Owner.Face = ""
	i.Owner.OfficialVerify = nil
	i.Owner.Relation = nil
	i.Owner.Attention = 0
	i.Owner.Fans = 0
	i.ReservationInfo = nil
	i.LiveRoom = nil
	i.Vip = accountgrpc.VipInfo{}
	for _, fn := range fns {
		fn(i)
	}
}

var (
	showLiveAttentionIconSet = sets.NewInt(3, 4)
	showLiveScrollMessageSet = sets.NewInt(2, 4)
)

func (i *Item) StoryFromLiveRoom(rcmd *ai.SubItems, room *xroom.EntryRoomInfoResp_EntryList,
	cardm map[int64]*accountgrpc.Card, statm map[int64]*relationgrpc.StatReply,
	relations map[int64]*relationgrpc.InterrelationReply, entryFrom string, liveRoom *LiveRoom,
	liveRankInfo map[int64]*liverankgrpc.IsInHotRankResp_HotRankData, fns ...StoryFn) {
	i.CardGoto = model.CardGt(rcmd.Goto)
	i.Goto = model.Gt(rcmd.Goto)
	i.Param = strconv.FormatInt(room.RoomId, 10)
	i.Cover = room.Cover
	i.Title = room.Title
	i.TrackID = rcmd.TrackID
	i.PosRecUniqueId = rcmd.PosRecUniqueId
	i.DislikeReportData = report.BuildDislikeReportData(0, rcmd.PosRecUniqueId)
	i.URI = model.FillURI("", 0, 0, room.JumpUrl[entryFrom], model.URLTrackIDHandler(rcmd))
	i.AppBackground = room.AppBackground
	i.Owner = &Owner{
		Mid: room.Uid,
	}
	if cm, ok := cardm[room.Uid]; ok {
		i.Owner.Name = cm.Name
		i.Owner.Face = cm.Face
		i.Owner.OfficialVerify = &OfficialInfo{
			Role:  cm.Official.Role,
			Title: cm.Official.Title,
			Type:  cm.Official.Type,
		}
		if i.Owner.OfficialVerify.Role == 7 {
			i.Owner.OfficialVerify.Role = 1
		}
		i.Vip = cm.Vip
		if i.Vip.Status == model.VipStatusExpire && cm.Vip.DueDate > 0 { //0-过期 (2-冻结 3-封禁不展示）
			i.Vip.Label.Path = model.VipLabelExpire
		}
	}
	if stat, ok := statm[room.Uid]; ok {
		i.Owner.Fans = stat.Follower
		i.Owner.Attention = stat.Following
	}
	i.Owner.Relation = model.RelationChange(room.Uid, relations)
	i.Rcmd = rcmd
	i.DislikeReasonsV3 = &DislikeReasonsV3{
		Dislike: dislikeV3FromLive(room, cardm, rcmd),
	}
	i.LiveRoom = liveRoom
	i.PlayerArgs = playerArgsFrom(room)
	if showLiveAttentionIconSet.Has(rcmd.LiveAttentionExp()) {
		i.LiveShowAttentionIcon = true
	}
	i.FillScrollMessage(rcmd, room, liveRankInfo)
	for _, fn := range fns {
		fn(i)
	}
}

func (i *Item) compatibleUnloginThumbup(mid int64) {
	if i.ReqUser.Like == 1 && mid == 0 && i.Stat != nil { // 未登录 && 已点赞 点赞数+1，但不落db
		i.Stat.Like++
	}
}

func (i *Item) StoryFromPGC(rcmd *ai.SubItems, ep *pgcinline.EpisodeCard, cardm map[int64]*accountgrpc.Card,
	statm map[int64]*relationgrpc.StatReply, relations map[int64]*relationgrpc.InterrelationReply,
	likeMap map[int64]int64, coinsMap map[int64]int64, haslike map[int64]int8, isFav map[int64]int8, animation string,
	mid int64, followMap map[int32]*pgcFollowClient.FollowStatusProto, fns ...StoryFn) {
	i.CardGoto = model.CardGt(rcmd.Goto)
	i.Goto = model.Gt(rcmd.Goto)
	i.Param = strconv.FormatInt(int64(ep.EpisodeId), 10)
	i.Cover = ep.Cover
	i.Title = func() string {
		if ep.GetNewDesc() != "" {
			return ep.GetNewDesc()
		}
		return ep.GetSeason().GetTitle()
	}()
	i.TrackID = rcmd.TrackID
	i.PosRecUniqueId = rcmd.PosRecUniqueId
	i.DislikeReportData = report.BuildDislikeReportData(0, rcmd.PosRecUniqueId)
	i.URI = ep.Url
	i.Stat = convertPGCToStat(ep.Stat, ep.Aid, ep.GetSeason())
	i.PubDate = time.Time(ep.GetPubRealTime().GetSeconds())
	i.Owner = &Owner{
		Mid: ep.GetContributeUpInfo().GetMid(),
	}
	authorCard, ok := cardm[ep.GetContributeUpInfo().GetMid()]
	if ok {
		i.Owner.Name = authorCard.Name
		i.Owner.Face = authorCard.Face
		i.Owner.OfficialVerify = &OfficialInfo{
			Role:  authorCard.Official.Role,
			Title: authorCard.Official.Title,
			Type:  authorCard.Official.Type,
		}
		if i.Owner.OfficialVerify.Role == 7 {
			i.Owner.OfficialVerify.Role = 1
		}
		i.Vip = authorCard.Vip
		if i.Vip.Status == model.VipStatusExpire && authorCard.Vip.DueDate > 0 { //0-过期 (2-冻结 3-封禁不展示）
			i.Vip.Label.Path = model.VipLabelExpire
		}
	}
	if stat, ok := statm[ep.GetContributeUpInfo().GetMid()]; ok {
		i.Owner.Fans = stat.Follower
		i.Owner.Attention = stat.Following
	}
	if allNum, ok := likeMap[ep.GetContributeUpInfo().GetMid()]; ok {
		i.Owner.LikeNum = allNum
	}
	i.Duration = ep.GetDuration()
	i.PlayerArgs = playerArgsFrom(ep)
	i.Dimension = &api.Dimension{
		Width:  int64(ep.GetDimension().GetWidth()),
		Height: int64(ep.GetDimension().GetHeight()),
		Rotate: int64(ep.GetDimension().GetRotate()),
	}
	i.ReqUser.Like = haslike[ep.Aid]
	i.ReqUser.Favorite = isFav[int64(ep.EpisodeId)]
	if follow, fok := followMap[ep.GetSeason().GetSeasonId()]; fok && follow.Follow == true {
		i.ReqUser.Follow = 1
	}
	if coin, ok := coinsMap[ep.Aid]; ok && coin > 0 {
		i.ReqUser.Coin = 1
	}
	i.Rcmd = rcmd
	i.Owner.Relation = model.RelationChange(ep.GetContributeUpInfo().GetMid(), relations)
	i.ThumbUpAnimation = animation
	i.ReportInfo = &ReportInfo{
		SubType:  ep.GetSeason().GetType(),
		EpStatus: ep.GetEpisodeStatus(),
	}
	i.DislikeReasonsV3 = &DislikeReasonsV3{
		Dislike:      dislikeV3FromPGC(ep, rcmd, authorCard),
		Feedback:     feedbackV3FromPGC(ep, rcmd),
		ActionSheets: dislikeReasonV3ActionSheetFromPGC,
	}
	i.Desc = ep.GetSeason().GetSummary()
	i.PGCInfo = &PGCInfo{
		Publisher:      ep.GetSeason().GetBadgeInfo().GetOptiText(),
		EpUpdateDesc:   ep.GetSeason().GetNewEp().GetDesc(),
		Ratio:          makeRatio(ep),
		OGVStyle:       rcmd.OGVStyle,
		JumpURI:        ep.GetUrl(),
		ClipStart:      ep.GetHeClipInfo().GetStart(),
		ClipEnd:        ep.GetHeClipInfo().GetEnd(),
		ShowSeasonType: ep.GetSeason().GetShowSeasonType(),
	}
	i.compatibleUnloginThumbup(mid)
	for _, fn := range fns {
		fn(i)
	}
}

func makeRatio(ep *pgcinline.EpisodeCard) string {
	if ep.GetSeason().GetRatingInfo().GetScore() == 0 {
		return "暂无评分"
	}
	return fmt.Sprintf("%.1f分", ep.GetSeason().GetRatingInfo().GetScore())
}

func (i *Item) FillStoryCartIcon(resource *cm.StoryAdResource) {
	storyCartIcon, hasStoryCartIcon := cm.AsStoryCartIcon(resource)
	if !hasStoryCartIcon {
		return
	}
	if storyCartIcon.IconTitle == "" && storyCartIcon.IconText == "" {
		log.Warn("Ad storyCartIcon is nil, %s", resource.RequestID)
		return
	}
	i.StoryCartIcon = &CommonStoryCart{
		IconURL:   storyCartIcon.IconURL,
		IconText:  storyCartIcon.IconText,
		IconTitle: storyCartIcon.IconTitle,
		Goto_:     "ad",
	}
}

func (i *Item) FillScrollMessage(rcmd *ai.SubItems, room *xroom.EntryRoomInfoResp_EntryList,
	rankInfos map[int64]*liverankgrpc.IsInHotRankResp_HotRankData) {
	if !showLiveScrollMessageSet.Has(rcmd.LiveAttentionExp()) {
		return
	}
	scrollMsg := []string{}
	if i.Owner.Relation.IsFollow == 1 && rcmd.LiveAttentionExp() == 2 {
		scrollMsg = append(scrollMsg, "你关注的主播")
	}
	if rank, ok := rankInfos[room.RoomId]; ok {
		msg := ""
		if rank.GetAreaData() != nil && rank.GetAreaData().GetRank() > 0 && rank.GetAreaData().GetRank() <= 10 {
			msg = "分区热门Top10"
		}
		if rank.GetTotalData() != nil && rank.GetTotalData().GetRank() > 0 && rank.GetTotalData().GetRank() <= 10 {
			msg = "直播热门Top10"
		}
		if msg != "" {
			scrollMsg = append(scrollMsg, msg)
		}
	}
	if room.WatchedShow == nil || !room.WatchedShow.Switch {
		scrollMsg = append(scrollMsg, appcardmodel.Stat64String(room.PopularityCount, "人气"))
		i.ScrollMessage = scrollMsg
		return
	}
	if room.WatchedShow.Num > 0 {
		scrollMsg = append(scrollMsg, appcardmodel.Stat64String(room.WatchedShow.Num, "人看过"))
	}
	i.ScrollMessage = scrollMsg
}

// Stat 稿件的所有计数信息
type Stat struct {
	Aid int64 `json:"aid"`
	// 播放数
	View int32 `json:"view"`
	// 弹幕数
	Danmaku int32 `json:"danmaku"`
	// 评论数
	Reply int32 `json:"reply"`
	// 收藏数
	Fav int32 `json:"favorite"`
	// 投币数
	Coin int32 `json:"coin"`
	// 分享数
	Share int32 `json:"share"`
	// 当前排名
	NowRank int32 `json:"now_rank"`
	// 历史最高排名
	HisRank int32 `json:"his_rank"`
	// 点赞数
	Like int32 `json:"like"`
	// 点踩数 已取消前台展示，现在均返回0
	DisLike int32 `json:"dislike"`
	// 追番数
	Follow int32 `json:"follow"`
}

func convertArchiveToStat(stat arcgrpc.Stat) *Stat {
	return &Stat{
		Aid:     stat.GetAid(),
		View:    stat.GetView(),
		Danmaku: stat.GetDanmaku(),
		Reply:   stat.GetReply(),
		Fav:     stat.GetFav(),
		Coin:    stat.GetCoin(),
		Share:   stat.GetShare(),
		NowRank: stat.GetNowRank(),
		HisRank: stat.GetHisRank(),
		Like:    stat.GetLike(),
		DisLike: stat.GetDisLike(),
		Follow:  stat.GetFollow(),
	}
}

func convertPGCToStat(stat *pgcinline.Stat, aid int64, seasonCard *pgcinline.SeasonCard) *Stat {
	share := int32(stat.GetShare())
	if seasonCard.GetTotalCount() > 0 {
		share = int32(stat.GetShare()) / seasonCard.GetTotalCount()
	}
	return &Stat{
		Aid:     aid,
		View:    int32(stat.GetPlay()),
		Danmaku: int32(stat.GetDanmaku()),
		Reply:   int32(stat.GetReply()),
		Fav:     int32(stat.GetFav()),
		Share:   share,
		Like:    int32(stat.GetLike()),
		Follow:  int32(seasonCard.GetStat().GetFollow()),
		Coin:    int32(stat.GetCoin()),
	}
}

func feedbackV3From(a *arcgrpc.ArcPlayer, ts []*channelgrpc.Channel, rcmd *ai.SubItems) *FeedbackV3 {
	const (
		_feedbackContentIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/RbFMYAPOiH.png"
		_feedbackPornIcon     = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/wMnnZQpics.png"
		_feedbackViolenceIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/9Xt1Io2r80.png"
		_feedbackTitleIcon    = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/lktjgzExnT.png"
		_feedbackToast        = "操作成功，将优化此类内容"
	)
	feedback := &FeedbackV3{
		Title: "内容反馈",
		FeedbackItems: []*FeedbackItem{
			{
				ID:    5,
				Icon:  _feedbackContentIcon,
				Title: "内容不适",
			},
			{
				ID:    2,
				Icon:  _feedbackPornIcon,
				Title: "色情低俗",
			},
			{
				ID:    1,
				Icon:  _feedbackViolenceIcon,
				Title: "恐怖血腥",
			},
			{
				ID:    4,
				Icon:  _feedbackTitleIcon,
				Title: "标题党",
			},
		},
	}
	var tagId int64
	if len(ts) > 0 {
		tagId = ts[0].ID
	}
	for _, item := range feedback.FeedbackItems {
		item.UpperMid = TargetMid(rcmd, a.Arc.Author.Mid)
		item.RID = a.Arc.TypeID
		item.TagID = tagId
		item.Toast = _feedbackToast
		item.ActionType = _actionTypeSkip
	}
	return feedback
}

var hideDislikeVerticalSet = sets.NewInt(2, 12)

func dislikeV3From(a *arcgrpc.ArcPlayer, ts []*channelgrpc.Channel, rcmd *ai.SubItems, card *accountgrpc.Card) *DislikeV3 {
	const (
		_tagMAX = 2
	)
	dislike := &DislikeV3{
		Title: "不感兴趣",
	}
	items := []*DislikeItem{
		{
			ID:         1,
			Icon:       _dislikeVideoIcon,
			Title:      "我不想看",
			SubTitle:   "该视频",
			UpperMid:   TargetMid(rcmd, a.Arc.Author.Mid),
			RID:        int64(a.Arc.TypeID),
			Toast:      dislikeToast(rcmd),
			ActionType: _actionTypeSkip,
		},
		{
			ID:         4,
			Icon:       _dislikeUpIcon,
			Title:      "不看UP主",
			SubTitle:   card.GetName(),
			UpperMid:   TargetMid(rcmd, a.Arc.Author.Mid),
			RID:        int64(a.Arc.TypeID),
			Toast:      dislikeToast(rcmd),
			ActionType: _actionTypeSkip,
		},
	}
	if !hideDislikeVerticalSet.Has(rcmd.VideoMode()) {
		items = append(items, &DislikeItem{
			ID:         9,
			Icon:       _dislikeStoryIcon,
			Title:      "我不想看",
			SubTitle:   "竖屏模式",
			UpperMid:   TargetMid(rcmd, a.Arc.Author.Mid),
			RID:        int64(a.Arc.TypeID),
			Toast:      dislikeStoryToast(rcmd),
			ActionType: _actionTypeNull,
		})
	}
	for i, t := range ts {
		if i == 0 {
			items[0].TagID = t.ID
			items[1].TagID = t.ID
		}
		items = append(items, &DislikeItem{
			ID:         3,
			Icon:       _dislikeChannelIcon,
			Title:      "不看频道",
			SubTitle:   "#" + t.Name + "#",
			TagID:      t.ID,
			UpperMid:   TargetMid(rcmd, a.Arc.Author.Mid),
			RID:        int64(a.Arc.TypeID),
			Toast:      dislikeToast(rcmd),
			ActionType: _actionTypeSkip,
		})
		if i+1 >= _tagMAX {
			break
		}
	}
	dislike.DislikeItems = items
	return dislike
}

func dislikeToast(rcmd *ai.SubItems) string {
	if rcmd.DisableRcmd() {
		return "将在开启个性化推荐后生效"
	}
	return "操作成功，将减少此类内容推荐"
}

func dislikeStoryToast(rcmd *ai.SubItems) string {
	if rcmd.DisableRcmd() {
		return "将在开启个性化推荐后生效"
	}
	return "操作成功，将减少竖屏模式推荐"
}

func dislikeV3FromLive(room *xroom.EntryRoomInfoResp_EntryList, cardm map[int64]*accountgrpc.Card, rcmd *ai.SubItems) *DislikeV3 {
	dislike := &DislikeV3{
		Title: "不感兴趣",
	}
	items := []*DislikeItem{}
	items = append(items, &DislikeItem{
		ID:         1,
		Icon:       _dislikeVideoIcon,
		Title:      "我不想看",
		SubTitle:   "该直播",
		RID:        room.AreaId,
		UpperMid:   room.Uid,
		Toast:      dislikeToast(rcmd),
		ActionType: _actionTypeSkip,
	})
	if card, ok := cardm[room.Uid]; ok {
		items = append(items, &DislikeItem{
			ID:         4,
			Icon:       _dislikeUpIcon,
			Title:      "不看UP主",
			SubTitle:   card.Name,
			UpperMid:   room.Uid,
			RID:        room.AreaId,
			Toast:      dislikeToast(rcmd),
			ActionType: _actionTypeSkip,
		})
	}
	items = append(items, &DislikeItem{
		ID:         2,
		Icon:       _dislikeLiveArea,
		Title:      "不看分区",
		SubTitle:   room.AreaName,
		RID:        room.AreaId,
		UpperMid:   room.Uid,
		Toast:      dislikeToast(rcmd),
		ActionType: _actionTypeSkip,
	}, &DislikeItem{
		ID:         14,
		Icon:       _dislikeLiveIcon,
		Title:      "不感兴趣",
		SubTitle:   "减少直播",
		RID:        room.AreaId,
		UpperMid:   room.Uid,
		Toast:      dislikeToast(rcmd),
		ActionType: _actionTypeSkip,
	})
	dislike.DislikeItems = items
	return dislike
}

func playerArgsFrom(v interface{}) (playerArgs *PlayerArgs) {
	//nolint:gosimple
	switch v.(type) {
	case *arcgrpc.Arc:
		a := v.(*arcgrpc.Arc)
		if a == nil || (a.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo && a.Rights.Autoplay != 1) || (a.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.AttrVal(arcgrpc.AttrBitBadgepay) == arcgrpc.AttrYes) {
			return
		}
		playerArgs = &PlayerArgs{Aid: a.Aid, Cid: a.FirstCid, Type: model.GotoAv}
	case *arcgrpc.ArcPlayer:
		a := v.(*arcgrpc.ArcPlayer)
		if a == nil || (a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrNo && a.Arc.Rights.Autoplay != 1) || (a.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.Arc.AttrVal(arcgrpc.AttrBitBadgepay) == arcgrpc.AttrYes) {
			return
		}
		playerArgs = &PlayerArgs{Aid: a.Arc.Aid, Cid: a.DefaultPlayerCid, Type: model.GotoAv}
	case *xroom.EntryRoomInfoResp_EntryList:
		r := v.(*xroom.EntryRoomInfoResp_EntryList)
		playerArgs = &PlayerArgs{Rid: r.RoomId, Type: model.GotoLive}
	case *pgcinline.EpisodeCard:
		epcard := v.(*pgcinline.EpisodeCard)
		playerArgs = &PlayerArgs{
			Aid:      epcard.Aid,
			Cid:      epcard.Cid,
			Type:     model.GotoBangumi,
			EpId:     epcard.EpisodeId,
			SeasonId: epcard.GetSeason().GetSeasonId(),
			Duration: epcard.GetDuration(),
		}
	default:
		log.Warn("playerArgsFrom: unexpected type %T", v)
	}
	return
}

func dislikeV3FromPGC(ep *pgcinline.EpisodeCard, rcmd *ai.SubItems, authorCard *accountgrpc.Card) *DislikeV3 {
	dislike := &DislikeV3{
		Title: "不感兴趣",
	}
	items := []*DislikeItem{
		{
			ID:         1,
			Icon:       _dislikeVideoIcon,
			Title:      "我不想看",
			SubTitle:   "该视频",
			UpperMid:   TargetMid(rcmd, ep.Season.GetUpInfo().GetMid()),
			Toast:      dislikeToast(rcmd),
			ActionType: _actionTypeSkip,
		},
	}
	if rcmd.DislikeStyle == 1 {
		items = append(items, &DislikeItem{
			ID:         15,
			Icon:       _dislikeSeasonIcon,
			Title:      "不看该剧",
			SubTitle:   ep.GetSeason().GetTitle(),
			UpperMid:   TargetMid(rcmd, ep.Season.GetUpInfo().GetMid()),
			Toast:      dislikeToast(rcmd),
			ActionType: _actionTypeSkip,
		})
	}
	items = append(items, &DislikeItem{
		ID:         4,
		Icon:       _dislikeUpIcon,
		Title:      authorCard.GetName(),
		SubTitle:   "不看UP主",
		UpperMid:   TargetMid(rcmd, ep.Season.GetUpInfo().GetMid()),
		Toast:      dislikeToast(rcmd),
		ActionType: _actionTypeSkip,
	})
	if !hideDislikeVerticalSet.Has(rcmd.VideoMode()) {
		items = append(items, &DislikeItem{
			ID:         9,
			Icon:       _dislikeStoryIcon,
			Title:      "我不想看",
			SubTitle:   "竖屏模式",
			UpperMid:   TargetMid(rcmd, ep.Season.GetUpInfo().GetMid()),
			Toast:      dislikeStoryToast(rcmd),
			ActionType: _actionTypeNull,
		})
	}
	dislike.DislikeItems = items
	return dislike
}

func feedbackV3FromPGC(ep *pgcinline.EpisodeCard, rcmd *ai.SubItems) *FeedbackV3 {
	const (
		_feedbackContentIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/RbFMYAPOiH.png"
		_feedbackPornIcon     = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/wMnnZQpics.png"
		_feedbackViolenceIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/9Xt1Io2r80.png"
		_feedbackTitleIcon    = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/lktjgzExnT.png"
		_feedbackToast        = "操作成功，将优化此类内容"
	)
	feedback := &FeedbackV3{
		Title: "内容反馈",
		FeedbackItems: []*FeedbackItem{
			{
				ID:    5,
				Icon:  _feedbackContentIcon,
				Title: "内容不适",
			},
			{
				ID:    2,
				Icon:  _feedbackPornIcon,
				Title: "色情低俗",
			},
			{
				ID:    1,
				Icon:  _feedbackViolenceIcon,
				Title: "恐怖血腥",
			},
			{
				ID:    4,
				Icon:  _feedbackTitleIcon,
				Title: "标题党",
			},
		},
	}
	for _, item := range feedback.FeedbackItems {
		item.UpperMid = TargetMid(rcmd, ep.GetSeason().GetUpInfo().GetMid())
		item.Toast = _feedbackToast
		item.ActionType = _actionTypeSkip
	}
	return feedback
}

func (d *DislikeV2) StoryDislike(a *arcgrpc.ArcPlayer, ts []*channelgrpc.Channel) {
	const (
		_noSeason = 1
		_channel  = 3
		_upper    = 4
		_tagMAX   = 2
	)
	var taginfo *channelgrpc.Channel
	d.Title = "选择不想看的原因，减少相似内容推荐"
	if a.Arc.Author.Name != "" {
		d.Reasons = append(d.Reasons, &DislikeReasons{ID: _upper, Name: "UP主:" + a.Arc.Author.Name, MID: a.Arc.Author.Mid})
	}
	for i, t := range ts {
		d.Reasons = append(d.Reasons, &DislikeReasons{ID: _channel, Name: "频道:" + t.Name, TagID: t.ID})
		if i == 0 {
			taginfo = t
		}
		if i+1 >= _tagMAX {
			break
		}
	}
	dislike := &DislikeReasons{ID: _noSeason, Name: "我不想看这个内容", MID: a.Arc.Author.Mid, RID: a.Arc.TypeID}
	if taginfo != nil {
		dislike.TagID = taginfo.ID
	}
	d.Reasons = append(d.Reasons, dislike)
	//nolint:gosimple
	return
}

// SpaceItem is
type SpaceItem struct {
	Item
}

// StoryFrom is
func (si *SpaceItem) StoryFrom(a *arcgrpc.ArcPlayer, cardm map[int64]*accountgrpc.Card, statm map[int64]*relationgrpc.StatReply,
	authorRelations map[int64]*relationgrpc.InterrelationReply, likeMap, coinMap map[int64]int64, haslike, isFav map[int64]int8,
	ts []*channelgrpc.Channel, hotAids map[int64]struct{}, buildLiveRoom func() *LiveRoom, plat int8, build int, mobiApp,
	animation string, mid int64) {
	fakeItem := &ai.SubItems{
		FfCover: "",
	}
	si.Item.StoryFrom(fakeItem, a, cardm, statm, authorRelations, likeMap, coinMap, haslike, isFav, ts, hotAids,
		buildLiveRoom, plat, build, mobiApp, animation, nil, nil, model.FfCoverFromSpaceStory,
		mid, nil, nil)
}

// SpaceCursorItem is
type SpaceCursorItem struct {
	Item
	Index   int64  `json:"index"`
	Postion string `json:"position,omitempty"`
	HasNext bool   `json:"has_next,omitempty"`
	HasPrev bool   `json:"has_prev,omitempty"`
}

// StoryFrom is
func (si *SpaceCursorItem) StoryFrom(a *arcgrpc.ArcPlayer, cardm map[int64]*accountgrpc.Card,
	statm map[int64]*relationgrpc.StatReply, authorRelations map[int64]*relationgrpc.InterrelationReply, likeMap,
	coinMap map[int64]int64, haslike, isFav map[int64]int8, ts []*channelgrpc.Channel, hotAids map[int64]struct{},
	buildLiveRoom func() *LiveRoom, storyArc *uparcgrpc.StoryArcs, total int64, plat int8,
	build int, mobiApp, animation string, argumentm map[int64]*vogrpc.Argument, mid int64, enableScreencast, buttonFilter,
	needCoin bool, likeAnimation map[int64]*thumbupgrpc.LikeAnimation) {
	fakeItem := &ai.SubItems{
		FfCover: "",
		Goto:    string(model.GotoVerticalAv),
	}
	fns := []StoryFn{
		OptThreePointButton(NeedDislike(false),
			NeedReport(true),
			NeedCoin(needCoin),
			NeedScreencast(enableScreencast),
			NoPlayBackground(a.Arc.Rights.NoBackground),
			ButtonFilter(buttonFilter),
		),
		OptShareBottomButton(NeedDislike(false),
			NeedReport(true),
			NeedCoin(needCoin),
			NeedScreencast(enableScreencast),
			CoinNum(appcardmodel.StatString(a.Arc.Stat.Coin, "")),
			NoPlayBackground(a.Arc.Rights.NoBackground),
			ButtonFilter(buttonFilter),
		),
	}
	si.Item.StoryFrom(fakeItem, a, cardm, statm, authorRelations, likeMap, coinMap, haslike, isFav, ts, hotAids,
		buildLiveRoom, plat, build, mobiApp, animation, nil, argumentm, model.FfCoverFromSpaceStory, mid,
		nil, likeAnimation, fns...)
	si.Index = storyArc.Rank

	si.HasNext = true
	si.HasPrev = true
	if storyArc.Rank <= 1 {
		si.HasPrev = false
	}
	if storyArc.Rank >= total {
		si.HasNext = false
	}
}

type DynamicStoryParam struct {
	Vmid   int64  `form:"vmid"`
	Scene  string `form:"scene" validate:"required"`
	Offset string `form:"offset"`
	Aid    int64  `form:"aid" validate:"min=1"`

	MobiApp    string `form:"mobi_app"`
	Platform   string `form:"platform"`
	Device     string `form:"device"`
	Build      int    `form:"build"`
	Qn         int    `form:"qn" default:"0"`
	Fnver      int    `form:"fnver" default:"0"`
	Fnval      int    `form:"fnval" default:"0"`
	ForceHost  int    `form:"force_host"`
	Fourk      int    `form:"fourk"`
	DeviceName string `form:"device_name"`
	TrackID    string `form:"trackid"`
	DisplayID  int64  `form:"display_id"`
	Pull       int    `form:"pull"`
	Network    string `form:"network"`
	From       int    `form:"from"`
	FromSpmid  string `form:"from_spmid"`
	Spmid      string `form:"spmid"`
	UidPos     int64  `form:"uid_pos"`
	NextUid    int64  `form:"next_uid"`
	TfType     int32
	NetType    int32

	Mid         int64
	Buvid       string
	Plat        int8
	TopicId     int64  `form:"topic_id"`
	TopicRid    int64  `form:"topic_rid"`
	TopicType   int64  `form:"topic_type"`
	TopicFrom   int64  `form:"topic_from"`
	DisableRcmd int    `form:"disable_rcmd"`
	SeasonID    int32  `form:"season_id"`
	OffsetType  string `form:"offset_type"`
	StoryParam  string `form:"story_param"`
}

type DynamicStoryReply struct {
	Items  []*Item           `json:"items"`
	Page   *DynamicStoryPage `json:"page"`
	Config *StoryConfig      `json:"config"`
}

type DynamicStoryPage struct {
	HasMore        bool   `json:"has_more"`
	Offset         string `json:"offset"`
	NextUid        int64  `json:"next_uid,omitempty"`
	OffsetType     string `json:"offset_type,omitempty"`
	SeasonID       int32  `json:"season_id,omitempty"`
	HasPrev        bool   `json:"has_prev,omitempty"`
	PrevOffset     string `json:"prev_offset,omitempty"`
	PrevOffsetType string `json:"prev_offset_type,omitempty"`
}

type StoryFn func(i *Item)

func OptAiStoryCommonJumpIcon(iconm *IconMaterial, rcmd *ai.SubItems, jumpToSeason bool) StoryFn {
	return func(i *Item) {
		if iconm == nil || rcmd.HasIcon == 0 {
			return
		}
		id := rcmd.IconID
		switch rcmd.IconType {
		case "ogv":
			ep, ok := iconm.Eppm[int32(id)]
			if !ok {
				return
			}
			uri := ep.Url
			if jumpToSeason == true && ep.GetFirstEpPlayUrl() != "" {
				uri = ep.GetFirstEpPlayUrl()
			}
			i.StoryCartIcon = &CommonStoryCart{
				IconURL:   "https://i0.hdslb.com/bfs/activity-plat/static/20220325/0977767b2e79d8ad0a36a731068a83d7/mxPHLManiG.png",
				IconText:  ep.GetSeason().GetTypeName(),
				IconTitle: ep.GetSeason().GetTitle(),
				URI:       uri,
				Goto_:     rcmd.IconType,
			}
		case "bcut":
			bcutIcon, ok := iconm.BcutStoryCart[fmt.Sprintf("%d:%d", rcmd.ID, id)]
			if !ok {
				return
			}
			i.StoryCartIcon = &CommonStoryCart{
				IconURL:   "https://i0.hdslb.com/bfs/activity-plat/static/20220325/0977767b2e79d8ad0a36a731068a83d7/5vfPBhvU7P.png",
				IconText:  "必剪",
				IconTitle: rcmd.IconTitle,
				Goto_:     rcmd.IconType,
				URI:       appcardmodel.FillURI("", 0, 0, bcutIcon.JumpUrl, appcardmodel.BcupURIHandler()),
			}
		case "cart":
			i.StoryCartIcon = &CommonStoryCart{
				IconURL:   "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/xWF5AJFFik.png",
				IconText:  "购物",
				IconTitle: rcmd.IconTitle,
				Goto_:     rcmd.IconType,
			}
		default:
		}
		return
	}
}

var (
	_noEntrance         = &empty{}
	_topicEntrance      = &topic{}
	_aspirationEntrance = &aspiration{}
)

func deriveEntranceLoader(rcmd *ai.SubItems) EntranceLoader {
	switch rcmd.HasTopic {
	case _topic:
		return _topicEntrance
	case _aspiration:
		return _aspirationEntrance
	case _tools:
		extra, ok := parseEntranceExtra(rcmd.ExtraJson)
		if !ok {
			return _noEntrance
		}
		return &tools{extra: extra}
	case _inspiration:
		extra, ok := parseEntranceExtra(rcmd.ExtraJson)
		if !ok {
			return _noEntrance
		}
		return &inspiration{extra: extra}
	case _music:
		extra, ok := parseEntranceExtra(rcmd.ExtraJson)
		if !ok {
			return _noEntrance
		}
		return &music{extra: extra}
	case _search:
		extra, ok := parseEntranceExtra(rcmd.ExtraJson)
		if !ok {
			return _noEntrance
		}
		return &search{extra: extra}
	default:
		return _noEntrance
	}
}

func parseEntranceExtra(extraJSON string) (EntranceExtra, bool) {
	extra := EntranceExtra{}
	if err := json.Unmarshal([]byte(extraJSON), &extra); err != nil {
		log.Error("Failed to unmarshal extra_json: %+v", errors.WithStack(err))
		return extra, false
	}
	return extra, true
}

func OptCreativeEntrance(rcmd *ai.SubItems) StoryFn {
	return func(i *Item) {
		entranceLoader := deriveEntranceLoader(rcmd)
		if !entranceLoader.Display() {
			return
		}
		uri := entranceLoader.JumpURI(rcmd)
		if uri == "" || rcmd.TopicTitle == "" {
			log.Warn("Unexpected entrance: id: %d, type: %d, title: %s, uri: %s", rcmd.TopicID, rcmd.HasTopic,
				rcmd.TopicTitle, uri)
			return
		}
		i.CreativeEntrance = &CreativeEntrance{
			Icon:    entranceLoader.Icon(),
			JumpURI: uri,
			Title:   entranceLoader.Title(rcmd),
			Type:    entranceLoader.Type(),
		}
	}
}

func OptPGCStyle(rcmd *ai.SubItems, ep *pgcinline.EpisodeCard) StoryFn {
	return func(i *Item) {
		if rcmd.OGVStyle == 1 || rcmd.OGVStyle == 2 {
			i.Owner.Face = ep.GetSeason().GetCover()
			i.Owner.FaceType = model.AvatarSquare
			i.Owner.Name = ep.GetSeason().GetTitle()
			i.Cover = ep.GetSeason().GetCover()
			i.Stat.View = int32(ep.GetSeason().GetStat().GetView())
			i.StoryCartIcon = nil
		}
	}
}

func OptPosRecTitle(rcmd *ai.SubItems) StoryFn {
	return func(i *Item) {
		if rcmd.PosRecUniqueId != "" && rcmd.PosRecTitle != "" {
			i.Title = rcmd.PosRecTitle
		}
	}
}

func OptThreePointButton(opts ...PanelButtonOption) StoryFn {
	return func(i *Item) {
		i.ThreePointButton = constructThreePointButton(opts...)
	}
}

func OptShareBottomButton(opts ...PanelButtonOption) StoryFn {
	return func(i *Item) {
		i.ShareBottomButton = constructShareBottomButton(opts...)
	}
}

type panelButtonConfig struct {
	needDislike       bool
	needReport        bool
	needCoin          bool
	needScreencast    bool
	noPlayBackground  int32
	coinNum           string
	buttonFilter      bool
	noWatchLater      bool
	smallWindowFilter bool
}

type PanelButtonOption func(*panelButtonConfig)

func (f *panelButtonConfig) Apply(opts ...PanelButtonOption) {
	for _, opt := range opts {
		opt(f)
	}
}

func NeedDislike(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.needDislike = in
	}
}

func NeedReport(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.needReport = in
	}
}

func NeedCoin(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.needCoin = in
	}
}

func NeedScreencast(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.needScreencast = in
	}
}

func CoinNum(in string) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.coinNum = in
	}
}

func NoPlayBackground(in int32) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.noPlayBackground = in
	}
}

func ButtonFilter(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.buttonFilter = in
	}
}

func NoWatchLater(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.noWatchLater = in
	}
}

func SmallWindowFilter(in bool) PanelButtonOption {
	return func(cfg *panelButtonConfig) {
		cfg.smallWindowFilter = in
	}
}

const (
	_dislikeType         = 1
	_coinType            = 2
	_speedPlayType       = 3
	_playModeType        = 4
	_watchLaterType      = 5
	_playBackgroundType  = 6
	_reportType          = 7
	_playFeedbackType    = 8
	_reportSuggestType   = 9
	_screencastType      = 10
	_danmuType           = 11
	_captionType         = 12
	_playSetType         = 13
	_mirrorFlipType      = 14
	_timedReminderType   = 15
	_smallWindowPlayType = 16
)

var (
	_dislikeFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _dislikeType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Text: "不感兴趣",
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/P7NhhyO3Jf.png",
			},
		},
	}
	_dislikeBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _dislikeType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Text:  "不感兴趣",
				Toast: "",
				Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/Kvjr0JsdP5.png",
			},
		},
	}
	_coinFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _coinType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Text:         "投币",
				Toast:        "",
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/nDe9pkEohr.png",
				ButtonStatus: "coined",
			},
			{
				Text:         "投币",
				Toast:        "",
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/SZOCvv7l4X.png",
				ButtonStatus: "coin",
			},
		},
	}
	_speedPlayFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _speedPlayType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/TC4bAHkgwc.png",
				Text:         "倍速播放",
				ButtonStatus: "0.5",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/Xeu9gkJLkJ.png",
				Text:         "倍速播放",
				ButtonStatus: "0.75",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/dZDKG3dA4b.png",
				Text:         "倍速播放",
				ButtonStatus: "1.0",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/aJtGPcdtWf.png",
				Text:         "倍速播放",
				ButtonStatus: "1.25",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/cRsuk6auaG.png",
				Text:         "倍速播放",
				ButtonStatus: "1.5",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/7lV7RiwUIb.png",
				Text:         "倍速播放",
				ButtonStatus: "2.0",
			},
		},
	}
	_speedPlayBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _speedPlayType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/zYqVnomYQ6.png",
				Text:         "倍速播放",
				ButtonStatus: "0.5",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/pahzfS3rse.png",
				Text:         "倍速播放",
				ButtonStatus: "0.75",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/KsWJp3rWOg.png",
				Text:         "倍速播放",
				ButtonStatus: "1.0",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/bRsjdEM3Yl.png",
				Text:         "倍速播放",
				ButtonStatus: "1.25",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/1sraELA9ih.png",
				Text:         "倍速播放",
				ButtonStatus: "1.5",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/4UPi4fcdXL.png",
				Text:         "倍速播放",
				ButtonStatus: "2.0",
			},
		},
	}
	_watchLaterFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _watchLaterType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/76iAc8fPXY.png",
				Text: "稍后再看",
			},
		},
	}
	_watchLaterBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _watchLaterType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/5XqGyNza4h.png",
				Text: "稍后再看",
			},
		},
	}
	_playModeBottomFunctionalButton = &threePointMeta.FunctionalButton{ // 6.83版本开始，客户端下线播放方式，该部分仅给老版本使用
		Type: _playModeType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Text:         "播放方式",
				ButtonStatus: "自动循环",
				Toast:        "自动循环",
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/jIf80Mv3JQ.png",
			},
			{
				Text:         "播放方式",
				ButtonStatus: "自动连播",
				Toast:        "自动连播",
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/jIf80Mv3JQ.png",
			},
		},
	}
	_playBackgroundFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _playBackgroundType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/CDSf9zKUqe.png",
				Text:         "后台播放",
				ButtonStatus: "open",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/kDzDfNsXzi.png",
				Text:         "后台播放",
				ButtonStatus: "close",
			},
		},
	}
	_playBackgroundBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _playBackgroundType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/reRrm2XG0n.png",
				Text:         "后台播放",
				ButtonStatus: "open",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/NPiRf8PYYX.png",
				Text:         "后台播放",
				ButtonStatus: "close",
			},
		},
	}
	_screencastFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _screencastType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/KRH8OwgWkp.png",
				Text:         "投屏中",
				ButtonStatus: "open",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/f599MlUZ24.png",
				Text:         "投屏",
				ButtonStatus: "close",
			},
		},
	}
	_screencastBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _screencastType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220606/0977767b2e79d8ad0a36a731068a83d7/Vnl6wgK081.png",
				Text:         "投屏中",
				ButtonStatus: "open",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220606/0977767b2e79d8ad0a36a731068a83d7/F5Qcn1292v.png",
				Text:         "投屏",
				ButtonStatus: "close",
			},
		},
	}
	_reportBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _reportType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/UkOZIfJsyT.png",
				Text: "举报",
			},
		},
	}
	_danmuBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _danmuType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/0977767b2e79d8ad0a36a731068a83d7/9q2ra6OYM2.png",
				Text: "弹幕设置",
			},
		},
	}
	_captionBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _captionType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/0977767b2e79d8ad0a36a731068a83d7/ppg0f4SGPY.png",
				Text: "字幕设置",
			},
		},
	}
	_playFeedbackBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _playFeedbackType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/iqt4EkQ7V1.png",
				Text: "播放反馈",
			},
		},
	}
	_reportSuggestBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _reportSuggestType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/heD1D9jUSQ.png",
				Text: "反馈建议",
			},
		},
	}
	_playSetBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _playSetType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220510/0977767b2e79d8ad0a36a731068a83d7/jIf80Mv3JQ.png",
				Text: "播放设置",
			},
		},
	}
	_mirrorFlipBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _mirrorFlipType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/WWOZsMcpr2.png",
				Text:         "镜像翻转",
				ButtonStatus: "open",
				Toast:        "镜像翻转已开启",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/NIj7LW1QFP.png",
				Text:         "镜像翻转",
				ButtonStatus: "close",
				Toast:        "镜像翻转已关闭",
			},
		},
	}
	_timedReminderBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _timedReminderType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/6t1GavFvUI.png",
				Text:         "定时关闭",
				ButtonStatus: "open",
				Toast:        "定时已开始",
			},
			{
				Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220801/0977767b2e79d8ad0a36a731068a83d7/Cl1lotjrC3.png",
				Text:         "定时关闭",
				ButtonStatus: "close",
				Toast:        "定时已关闭",
			},
		},
	}
	_smallWindowPlayBottomFunctionalButton = &threePointMeta.FunctionalButton{
		Type: _smallWindowPlayType,
		ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
			{
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220826/0977767b2e79d8ad0a36a731068a83d7/yMpnEPEypk.png",
				Text: "小窗播放",
			},
		},
	}
)

func constructThreePointButton(opts ...PanelButtonOption) *ThreePointButton {
	top := make([]*threePointMeta.FunctionalButton, 0, 6)
	cfg := &panelButtonConfig{}
	cfg.Apply(opts...)
	top = append(top, _speedPlayFunctionalButton)
	if cfg.needDislike {
		top = append(top, _dislikeFunctionalButton)
	}
	if cfg.needCoin {
		top = append(top, _coinFunctionalButton)
	}
	if !cfg.noWatchLater {
		top = append(top, _watchLaterFunctionalButton)
	}
	if cfg.noPlayBackground == 0 {
		top = append(top, _playBackgroundFunctionalButton)
	}
	if cfg.needScreencast {
		top = append(top, _screencastFunctionalButton)
	}
	bottom := make([]*threePointMeta.FunctionalButton, 0, 10)
	bottom = append(bottom, _playModeBottomFunctionalButton)
	if !cfg.buttonFilter { // 后续新增按钮均需此判断
		bottom = append(bottom, _mirrorFlipBottomFunctionalButton)
		bottom = append(bottom, _timedReminderBottomFunctionalButton)
		if !cfg.smallWindowFilter {
			bottom = append(bottom, _smallWindowPlayBottomFunctionalButton)
		}
		bottom = append(bottom, _playSetBottomFunctionalButton)
		bottom = append(bottom, _danmuBottomFunctionalButton)
		bottom = append(bottom, _captionBottomFunctionalButton)
	}
	if cfg.needReport {
		bottom = append(bottom, _reportBottomFunctionalButton)
	}
	bottom = append(bottom, _playFeedbackBottomFunctionalButton)
	bottom = append(bottom, _reportSuggestBottomFunctionalButton)
	return &ThreePointButton{
		Top:    top,
		Bottom: bottom,
	}
}

func constructShareBottomButton(opts ...PanelButtonOption) []*threePointMeta.FunctionalButton {
	out := make([]*threePointMeta.FunctionalButton, 0, 15)
	cfg := &panelButtonConfig{}
	cfg.Apply(opts...)
	out = append(out, _speedPlayBottomFunctionalButton)
	if cfg.needDislike {
		out = append(out, _dislikeBottomFunctionalButton)
	}
	if cfg.needCoin {
		out = append(out, &threePointMeta.FunctionalButton{
			Type: _coinType,
			ButtonMetas: []*threePointMeta.FunctionalButtonMeta{
				{
					Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/zEdwknqnpR.png",
					Text:         cfg.coinNum,
					ButtonStatus: "coin",
				},
				{
					Icon:         "https://i0.hdslb.com/bfs/activity-plat/static/20220513/0977767b2e79d8ad0a36a731068a83d7/FsaWVFC3ED.png",
					Text:         cfg.coinNum,
					ButtonStatus: "coined",
				},
			},
		})
	}
	if !cfg.noWatchLater {
		out = append(out, _watchLaterBottomFunctionalButton)
	}
	out = append(out, _playModeBottomFunctionalButton)
	if cfg.noPlayBackground == 0 {
		out = append(out, _playBackgroundBottomFunctionalButton)
	}
	if cfg.needScreencast {
		out = append(out, _screencastBottomFunctionalButton)
	}
	if !cfg.buttonFilter { // 后续新增按钮均需此判断
		out = append(out, _mirrorFlipBottomFunctionalButton)
		if !cfg.smallWindowFilter {
			out = append(out, _smallWindowPlayBottomFunctionalButton)
		}
		out = append(out, _playSetBottomFunctionalButton)
		out = append(out, _danmuBottomFunctionalButton)
		out = append(out, _captionBottomFunctionalButton)
	}
	if cfg.needReport {
		out = append(out, _reportBottomFunctionalButton)
	}
	out = append(out, _playFeedbackBottomFunctionalButton)
	out = append(out, _reportSuggestBottomFunctionalButton)
	return out
}

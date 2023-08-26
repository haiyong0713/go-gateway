package act

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	actapi "git.bilibili.co/bapis/bapis-go/activity/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	scoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	chagrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	playgrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgcClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	showecode "go-gateway/app/app-svr/app-show/ecode"
	pb "go-gateway/app/app-svr/app-show/interface/api"
	"go-gateway/app/app-svr/app-show/interface/conf"
	accdao "go-gateway/app/app-svr/app-show/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-show/interface/dao/act"
	arcdao "go-gateway/app/app-svr/app-show/interface/dao/archive"
	artdao "go-gateway/app/app-svr/app-show/interface/dao/article"
	bgmdao "go-gateway/app/app-svr/app-show/interface/dao/bangumi"
	busdao "go-gateway/app/app-svr/app-show/interface/dao/business"
	carddao "go-gateway/app/app-svr/app-show/interface/dao/card"
	cartdao "go-gateway/app/app-svr/app-show/interface/dao/cartoon"
	"go-gateway/app/app-svr/app-show/interface/dao/channel"
	dynvotedao "go-gateway/app/app-svr/app-show/interface/dao/dynamic-vote"
	"go-gateway/app/app-svr/app-show/interface/dao/dynamicsvr"
	"go-gateway/app/app-svr/app-show/interface/dao/esports"
	favdao "go-gateway/app/app-svr/app-show/interface/dao/favorite"
	gadao "go-gateway/app/app-svr/app-show/interface/dao/game"
	hmtchanneldao "go-gateway/app/app-svr/app-show/interface/dao/hmt-channel"
	livedao "go-gateway/app/app-svr/app-show/interface/dao/live"
	pgcdao "go-gateway/app/app-svr/app-show/interface/dao/pgc"
	platdao "go-gateway/app/app-svr/app-show/interface/dao/plat"
	populardao "go-gateway/app/app-svr/app-show/interface/dao/popular"
	reldao "go-gateway/app/app-svr/app-show/interface/dao/relation"
	scoredao "go-gateway/app/app-svr/app-show/interface/dao/score"
	shopdao "go-gateway/app/app-svr/app-show/interface/dao/shop"
	tagdao "go-gateway/app/app-svr/app-show/interface/dao/tag"
	showmdl "go-gateway/app/app-svr/app-show/interface/model"
	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	activitymdl "go-gateway/app/app-svr/app-show/interface/model/activity"
	bgmmdl "go-gateway/app/app-svr/app-show/interface/model/bangumi"
	cartmdl "go-gateway/app/app-svr/app-show/interface/model/cartoon"
	"go-gateway/app/app-svr/app-show/interface/model/dynamic"
	"go-gateway/app/app-svr/app-show/interface/model/selected"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arccli "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	natecode "go-gateway/app/web-svr/native-page/ecode"
	"go-gateway/app/web-svr/native-page/interface/api"
)

type Service struct {
	c             *conf.Config
	accDao        *accdao.Dao
	actDao        *actdao.Dao
	dynamicDao    *dynamicsvr.Dao
	tagDao        *tagdao.Dao
	reldao        *reldao.Dao
	pgcdao        *pgcdao.Dao
	arcdao        *arcdao.Dao
	artdao        *artdao.Dao
	bgmdao        *bgmdao.Dao
	livedao       *livedao.Dao
	favdao        *favdao.Dao
	populardao    *populardao.Dao
	businessdao   *busdao.Dao
	cdao          *carddao.Dao
	platDao       *platdao.Dao
	channelDao    *channel.Dao
	gameDao       *gadao.Dao
	shopDao       *shopdao.Dao
	cartdao       *cartdao.Dao
	hmtChannelDao *hmtchanneldao.Dao
	dynvoteDao    *dynvotedao.Dao
	scoreDao      *scoredao.Dao
	report        *Report
	esportsDao    *esports.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:             c,
		accDao:        accdao.New(c),
		actDao:        actdao.New(c),
		dynamicDao:    dynamicsvr.New(c),
		tagDao:        tagdao.New(c),
		reldao:        reldao.New(c),
		pgcdao:        pgcdao.New(c),
		arcdao:        arcdao.New(c),
		artdao:        artdao.New(c),
		bgmdao:        bgmdao.New(c),
		livedao:       livedao.New(c),
		favdao:        favdao.New(c),
		populardao:    populardao.New(c),
		businessdao:   busdao.New(c),
		cdao:          carddao.New(c),
		platDao:       platdao.New(c),
		channelDao:    channel.New(c),
		gameDao:       gadao.New(c),
		shopDao:       shopdao.New(c),
		cartdao:       cartdao.New(c),
		hmtChannelDao: hmtchanneldao.New(c),
		dynvoteDao:    dynvotedao.NewDao(c),
		report:        NewReport(c.NaInfoc.Infoc),
		scoreDao:      scoredao.NewDao(c),
		esportsDao:    esports.NewDao(c),
	}
	return
}

// ActFollow .
func (s *Service) ActFollow(c context.Context, arg *actmdl.ParamActFollow, mid int64) (*actmdl.FollowRly, error) {
	rly := &actmdl.FollowRly{}
	var err error
	switch arg.Goto {
	case actmdl.GotoClickCartoon:
		if arg.Type == actmdl.AddReserve {
			err = s.cartdao.AddFavorite(c, []int64{arg.FID}, mid)
		} else {
			err = s.cartdao.DelFavorite(c, []int64{arg.FID}, mid)
		}
	case actmdl.GotoClickVote:
		if arg.Type == actmdl.AddReserve {
			risk := &actapi.Risk{Buvid: arg.Buvid, Build: fmt.Sprintf("%d", arg.Build), Platform: arg.Platform, UserAgent: arg.UserAgent, Ip: metadata.String(c, metadata.RemoteIP), Api: "/x/v2/activity/follow"}
			rly.Num, rly.CanVoteNum, err = s.actDao.VoteUserDo(c, arg.FID, arg.GroupID, arg.ItemID, mid, 1, risk)
		} else {
			rly.Num, rly.CanVoteNum, err = s.actDao.VoteUserUndo(c, arg.FID, arg.GroupID, arg.ItemID, mid)
		}
	case actmdl.GotoClickVoteUp:
		if arg.Type == actmdl.AddReserve {
			_, err = s.dynvoteDao.DoVote(c, &dynvotegrpc.DoVoteReq{VoteId: arg.FID, Votes: []int32{int32(arg.ItemID)}, VoterUid: mid})
		} else {
			err = showecode.HasVoted
		}
	case actmdl.GotoClickBuy:
		if arg.Type == actmdl.AddReserve {
			err = s.shopDao.AddFav(c, arg.FID, mid)
		} else {
			err = s.shopDao.DelFav(c, arg.FID, mid)
		}
	case actmdl.GotoClickAttention:
		if arg.Type == actmdl.AddReserve {
			err = s.actDao.GRPCDoRelation(c, arg.FID, mid, "native_page", arg.FromSpmid, arg.Buvid, arg.Platform, arg.MobiApp)
		} else {
			err = s.actDao.RelationReserveCancel(c, arg.FID, mid, "native_page", arg.FromSpmid, arg.Buvid, arg.Platform, arg.MobiApp)
		}
	case actmdl.GotoClickAppointment, actmdl.GotoClickReserve:
		if arg.Type == actmdl.AddReserve {
			err = s.actDao.AddReserve(c, arg.FID, mid)
		} else {
			err = s.actDao.DelReserve(c, arg.FID, mid)
		}
	case actmdl.GotoClickPgc:
		if arg.Type == actmdl.AddReserve {
			err = s.pgcdao.AddFollow(c, int32(arg.FID), mid)
		} else {
			err = s.pgcdao.DeleteFollow(c, int32(arg.FID), mid)
		}
	case actmdl.GotoClickFollow:
		if arg.Type == actmdl.AddReserve {
			err = s.reldao.AddFollowing(c, arg.FID, mid, arg.FromSpmid)
		} else {
			err = s.reldao.DelFollowing(c, arg.FID, mid, arg.FromSpmid)
		}
	default:
		err = ecode.RequestErr
	}
	if err != nil {
		return nil, err
	}
	return rly, nil
}

// ActLiked .
func (s *Service) ActLiked(c context.Context, arg *actmdl.ParamActLike, mid int64) (res *actmdl.LikedReply, err error) {
	var (
		rely *actapi.ActLikedReply
	)
	if rely, err = s.actDao.ActLiked(c, &actapi.ActLikedReq{Sid: arg.Sid, Lid: arg.Lid, Score: arg.Score, Mid: mid}); err != nil {
		switch {
		case ecode.EqualError(natecode.ActivityLikeHasEnd, err):
			err = xecode.AppActHasEnd
		case ecode.EqualError(natecode.ActivityLikeNotStart, err):
			err = xecode.AppActNotStart
		case ecode.EqualError(natecode.ActivityOverLikeLimit, err):
			err = xecode.AppActOverLikeLimit
		case ecode.EqualError(natecode.ActivityLikeIPFrequence, err), ecode.EqualError(natecode.ActivityLikeScoreLower, err),
			ecode.EqualError(natecode.ActivityLikeRegisterLimit, err), ecode.EqualError(natecode.ActivityLikeBeforeRegister, err),
			ecode.EqualError(natecode.ActivityTelValid, err), ecode.EqualError(natecode.ActivityLikeLevelLimit, err):
			err = xecode.AppNoLikeCondition
		}
		log.Error("s.actDao.ActLiked(%v) error(%v)", arg, err)
		return
	}
	if rely == nil {
		return
	}
	res = &actmdl.LikedReply{Score: rely.Score, Toast: "投票成功，已为稿件增加" + strconv.FormatInt(rely.Score, 10) + "票"}
	return
}

func fromVideoGoto(mou *api.NativeModule) (gotoType string) {
	switch {
	case mou.IsVideo():
		gotoType = actmdl.GotoVideo
	case mou.IsVideoAct():
		if mou.IsCardSingle() {
			gotoType = actmdl.GotoActSingleModule
		} else {
			gotoType = actmdl.GotoActDoubleModule
		}
	case mou.IsVideoAvid():
		if mou.IsCardSingle() {
			gotoType = actmdl.GotoAvIDSingleModule
		} else {
			gotoType = actmdl.GotoAvIDDoubleModule
		}
	case mou.IsVideoDyn():
		if mou.IsCardSingle() {
			gotoType = actmdl.GotoDynSingleModule
		} else {
			gotoType = actmdl.GotoDynDoubleModule
		}
	case mou.IsResourceOrigin():
		gotoType = actmdl.GotoOriginResourceModule
	case mou.IsResourceAct():
		gotoType = actmdl.GotoActResourceModule
	case mou.IsResourceID():
		gotoType = actmdl.GotoIDResourceModule
	case mou.IsResourceDyn():
		gotoType = actmdl.GotoDynResourceModule
	case mou.IsNewVideoID():
		gotoType = actmdl.GotoNewIDVideoModule
	case mou.IsNewVideoAct():
		gotoType = actmdl.GotoNewActVideoModule
	case mou.IsNewVideoDyn():
		gotoType = actmdl.GotoNewDynVideoModule
	}
	return
}

// Supernatant 浮层相关数据
func (s *Service) Supernatant(c context.Context, arg *actmdl.ParamSupernatant, mid int64) (*actmdl.SupernatantReply, error) {
	eg := errgroup.WithContext(c)
	//获取配置信息
	var rely *api.ModuleConfigReply
	eg.Go(func(ctx context.Context) error {
		var e error
		if rely, e = s.actDao.ModuleConfig(ctx, &api.ModuleConfigReq{ModuleID: arg.ConfModuleID}); e != nil {
			log.Error("s.actDao.ModuleConfig(%d) error(%v)", arg.ConfModuleID, e)
			return showecode.ActivityNothingMore
		}
		if rely == nil || rely.NativePage == nil || rely.Module == nil || rely.Module.NativeModule == nil {
			return showecode.ActivityNothingMore
		}
		return nil
	})
	var natPage *api.NativePage
	//降级处理
	eg.Go(func(ctx context.Context) error {
		var e error
		if natPage, e = s.actDao.NativePage(ctx, arg.PageID); e != nil {
			log.Error("s.actDao.NativePage(%d) error(%v)", arg.PageID, e)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	itemRly := s.supernatantItem(c, rely.Module.NativeModule, arg, mid)
	rly := &actmdl.SupernatantReply{Supernatant: itemRly}
	if natPage != nil {
		rly.Color = &actmdl.Color{BgColor: natPage.BgColor}
		rly.AttrBit = &actmdl.AttrBit{NotNight: natPage.IsAttrNotNightModule() == api.AttrModuleYes}
	}
	return rly, nil
}

// supernatantItem .
// nolint:gomnd
func (s *Service) supernatantItem(c context.Context, mou *api.NativeModule, arg *actmdl.ParamSupernatant, mid int64) *actmdl.Supernatant {
	if arg.Offset == 0 && arg.LastIndex > 0 { //第一刷&有偏移量
		arg.Ps += arg.LastIndex
		if arg.Ps > 100 {
			arg.Ps = 100
		}
	}
	rly := &actmdl.Supernatant{UrlExt: &actmdl.UrlExt{LastIndex: -1, ConfModuleID: arg.ConfModuleID}}
	var (
		teRly   *actmdl.ResourceReply
		err     error
		ltIndex int64
	)
	switch {
	case mou.IsTimelineIDs():
		teRly, err = s.timelineIDs(c, mou, arg.Ps, arg.Offset, mid, arg.MobiApp, arg.Device)
		ltIndex = arg.LastIndex * 2 //+时间条+事件数
	case mou.IsTimelineSource():
		teRly, err = s.timelineResource(c, mou, arg.Ps, arg.Offset)
		ltIndex = arg.LastIndex * 2
	case mou.IsOgvSeasonID():
		teRly, err = s.ogvSeasonID(c, mou, arg.Ps, arg.Offset, mid)
		ltIndex = arg.LastIndex
	case mou.IsOgvSeasonSource():
		teRly, err = s.ogvSeasonResource(c, mou, arg.Ps, arg.Offset, mid)
		ltIndex = arg.LastIndex
	default:
		err = showecode.ActivityNothingMore
	}
	if err != nil || teRly == nil {
		log.Error("s.TimelineIDs(%d) error(%v) or is nil", arg.ConfModuleID, err)
		//降级处理
		return rly
	}
	rly.UrlExt.Offset = teRly.Offset
	rly.HasMore = teRly.HasMore
	if arg.Offset == 0 { //第一刷下发浮层标题
		rly.TitleConf = &actmdl.Item{}
		rly.TitleConf.FromTitleConf(mou)
	}
	tmpCd := &actmdl.Item{}
	switch {
	case mou.IsTimelineIDs(), mou.IsTimelineSource():
		tmpCd.FromTimelineModule(mou)
	case mou.IsOgvSeasonID(), mou.IsOgvSeasonSource():
		tmpCd.FromOgvSeasonModule(mou)
	}
	if arg.Offset == 0 { //第一刷
		rly.LastIndex = ltIndex // 浮层
	}
	tmpCd.Item = append(tmpCd.Item, teRly.List...)
	rly.Cards = append(rly.Cards, tmpCd)
	return rly
}

// LikeList 话题二级列表页面接口.
func (s *Service) LikeList(c context.Context, arg *actmdl.ParamLike, mid int64) (res *actmdl.LikeListRely, err error) {
	if isNeedFix(arg.MobiApp, arg.Device, arg.Build) && arg.RemoteFrom == "" {
		arg.RemoteFrom = actmdl.RemoteActivityPage
	}
	switch arg.Goto {
	case actmdl.GotoAvIDSingleModule: // 视频卡-avid-单列
		res, err = s.VideoAvidList(c, arg, mid, true)
	case actmdl.GotoAvIDDoubleModule: // 视频卡-avid-双列
		res, err = s.VideoAvidList(c, arg, mid, false)
	case actmdl.GotoDynSingleModule: //视频卡-动态-单列
		res, err = s.VideoDynList(c, arg, mid, true)
	case actmdl.GotoDynDoubleModule: //视频卡-动态-双列
		res, err = s.VideoDynList(c, arg, mid, false)
	case actmdl.GotoActSingleModule: // 视频卡-活动-单列
		res, err = s.VideoActList(c, arg, mid, true)
	case actmdl.GotoActDoubleModule: // 视频卡-活动-双列
		res, err = s.VideoActList(c, arg, mid, false)
	case actmdl.GotoIDResourceModule: //资源卡-id
		res, err = s.ResourceIDList(c, arg, mid)
	case actmdl.GotoDynResourceModule: //资源卡-dyn
		res, err = s.ResourceDynList(c, arg, mid)
	case actmdl.GotoActResourceModule: //资源卡-act
		res, err = s.ResourceActList(c, arg, mid)
	case actmdl.GotoOriginResourceModule: //资源卡-外接数据源模式
		res, err = s.ResourceOriginList(c, arg, mid)
	case actmdl.GotoNewIDVideoModule: //新视频卡-id
		res, err = s.newVideoAvidList(c, arg)
	case actmdl.GotoNewActVideoModule: //新视频卡-act
		res, err = s.newVideoActList(c, arg, mid)
	case actmdl.GotoNewDynVideoModule: //新视频卡-dyn
		res, err = s.newVideoDynList(c, arg, mid)
	case actmdl.GotoEditOriginModule: //编辑推荐卡-数据源模式
		res = s.editOriginList(c, arg, mid)
	default: //默认是活动二级列表
		res, err = s.VideoList(c, arg, mid)
	}
	// 降级处理,错误不抛出
	if err != nil {
		res = &actmdl.LikeListRely{Offset: arg.Offset, DyOffset: arg.DyOffset, HasMore: 0, Cards: make([]*actmdl.Item, 0), Page: &actmdl.Page{Pn: arg.Pn, Ps: arg.Ps}}
		err = nil
	}
	return
}

// BaseDetail .
func (s *Service) BaseDetail(c context.Context, arg *actmdl.ParamBaseDetail) (*actmdl.BaseReply, error) {
	// p_type=2参与组件等用户基础组件
	pageConf, err := s.actDao.BaseConfig(c, &api.BaseConfigReq{Pid: arg.PageID, PType: 2})
	if err != nil || pageConf == nil || pageConf.NativePage == nil {
		if ecode.EqualError(natecode.NativePageOffline, err) {
			err = xecode.AppPageOffline
		}
		log.Error("s.actDao.NatConfig(%d) error(%v)", arg.PageID, err)
		return nil, err
	}
	var partModule *api.Module
	if len(pageConf.Bases) > 0 {
		for _, vb := range pageConf.Bases {
			if vb.NativeModule == nil {
				continue
			}
			if vb.NativeModule.IsPart() && vb.Participation != nil && len(vb.Participation.List) > 0 {
				partModule = vb
				break
			}
		}
	}
	baseMs := &actmdl.BaseReply{PageID: arg.PageID}
	commonConf := pageConf.NativePage
	if partModule != nil {
		baseMs.Bases = &actmdl.Bases{}
		baseMs.Bases.Participation = s.FromPart(c, partModule.NativeModule, partModule.Participation, commonConf)
	}
	return baseMs, nil
}

// ActDetail .
// nolint:gocognit
func (s *Service) ActDetail(c context.Context, arg *actmdl.ParamActDetail) (res *actmdl.DetailReply, err error) {
	var (
		rely *api.ModuleConfigReply
	)
	if rely, err = s.actDao.ModuleConfig(c, &api.ModuleConfigReq{ModuleID: arg.ModuleID}); err != nil {
		log.Error("s.actDao.ModuleConfig(%d) error(%v)", arg.ModuleID, err)
		return
	}
	if rely == nil || rely.NativePage == nil || rely.Module == nil || rely.Module.NativeModule == nil {
		err = showecode.ActivityNothingMore
		return
	}
	v := rely.Module.NativeModule
	confSort := v.ConfUnmarshal()
	//549以前版本，只支持一种话题活动卡片类型
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynamicCard549, &feature.OriginResutl{
		BuildLimit: (arg.MobiApp == "iphone" && arg.Build <= 8910) || (arg.MobiApp == "android" && arg.Build < 5490000),
	}) {
		if !v.IsVideo() {
			err = showecode.DynamicBuildLimit
			return
		}
	}
	//557以下版本不支持资源小卡
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ResourceSmallCard557, &feature.OriginResutl{
		BuildLimit: (arg.MobiApp == "iphone" && arg.Build <= 9290) || (arg.MobiApp == "android" && arg.Build < 5570000),
	}) {
		if v.IsResourceAct() || v.IsResourceDyn() || v.IsResourceID() || v.IsResourceOrigin() {
			err = showecode.DynamicBuildLimit
			return
		}
	}
	res = &actmdl.DetailReply{
		PageID:      rely.NativePage.ID,
		ForeignID:   rely.NativePage.ForeignID,
		ForeignType: rely.NativePage.Type,
		Name:        rely.NativePage.Title,
		Title:       v.Title,
		Param:       &actmdl.PartParam{Goto: fromVideoGoto(v), Attr: v.Attribute},
	}
	if (v.IsAttrAutoPlay() == api.AttrModuleYes) || (v.IsAttrHideTitle() != api.AttrModuleYes) {
		res.Setting = &actmdl.Setting{AutoPlay: v.IsAttrAutoPlay() == api.AttrModuleYes, DisplayTitle: v.IsAttrHideTitle() != api.AttrModuleYes} //配置信息批量下发
	}
	switch {
	case v.IsVideo(), v.IsVideoAct(), v.IsResourceAct(), v.IsNewVideoAct():
		res.Sid = v.Fid
		sortTemp := &actmdl.Item{}
		sortTemp.FromSortModule(rely.Module.VideoAct)
		// 修护541版本下ios bug,只返回一个sort list
		if arg.MobiApp == "iphone" && (arg.Build == 8670 || arg.Build == 8680) {
			if len(sortTemp.Item) > 1 {
				sortTemp.Item = sortTemp.Item[:1]
			}
		}
		res.Cards = []*actmdl.Item{sortTemp}
	case v.IsResourceOrigin():
		// 下发老字段保证客户端可兼容
		if confSort.RdbType == api.RDBLive {
			sortTemp := &actmdl.Item{}
			sortTemp.FromSortModule(rely.Module.VideoAct)
			res.Cards = []*actmdl.Item{sortTemp}
		}
		res.Param.AvSort = confSort.RdbType //外接数据源类型
		res.Param.DyType = v.TName          //外接数据源id
	case v.IsVideoDyn(), v.IsNewVideoDyn():
		res.Param.TopicID = v.Fid
	case v.IsResourceDyn():
		res.Param.TopicID = v.Fid
		types := strconv.Itoa(dynamic.VideoType) //默认排序
		if rely.Module.Dynamic != nil && len(rely.Module.Dynamic.SelectList) > 0 {
			types = strconv.FormatInt(rely.Module.Dynamic.SelectList[0].SelectType, 10)
		}
		res.Param.DyType = types
	case v.IsVideoAvid(), v.IsResourceID():
		res.Param.AvSort = v.AvSort
	}
	return
}

// ogvSeasonResource .
func (s *Service) ogvSeasonResource(c context.Context, mou *api.NativeModule, ps, offset, mid int64) (*actmdl.ResourceReply, error) {
	if mou.Fid <= 0 {
		return nil, ecode.RequestErr
	}
	// 获取ogv 片单信息 平台信息
	treply, err := s.pgcdao.SeasonByPlayId(c, int32(mou.Fid), int32(offset), int32(ps), mid)
	if err != nil {
		log.Error("s.pgcdao.SeasonByPlayI(%d) error(%v)", mou.Fid, err)
		return nil, err

	}
	if treply == nil {
		return nil, showecode.ActivityNothingMore
	}
	rly := &actmdl.ResourceReply{Offset: int64(treply.NexOffset)}
	if treply.HasNext {
		rly.HasMore = 1
	}
	if len(treply.SeasonInfos) == 0 {
		return rly, nil
	}
	for _, v := range treply.SeasonInfos {
		if v == nil {
			continue
		}
		tmp := &actmdl.Item{}
		tmp.FromOgvSeason(mou, v, "")
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// timelineResource .
func (s *Service) timelineResource(c context.Context, mou *api.NativeModule, ps, offset int64) (*actmdl.ResourceReply, error) {
	// 获取资讯平台信息
	treply, err := s.populardao.TimeLine(c, mou.Fid, int32(offset), int32(ps))
	if err != nil {
		log.Error(" s.populardao.TimeLine(%d) error(%v)", mou.Fid, err)
		return nil, err
	}
	if treply == nil {
		return nil, showecode.ActivityNothingMore
	}
	rly := &actmdl.ResourceReply{Offset: int64(treply.Offset)}
	if treply.HasMore {
		rly.HasMore = 1
	}
	if len(treply.Events) == 0 {
		return rly, nil
	}
	confSort := mou.ConfUnmarshal()
	for _, v := range treply.Events {
		if v == nil {
			continue
		}
		tmp := &actmdl.Item{}
		//图文类型
		tmp.FromTimeline(v)
		head := &actmdl.Item{}
		head.FromTimelineFormatHead(v.Stime, confSort)
		rly.List = append(rly.List, head)
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// timelineIDs .
// nolint:gocognit
func (s *Service) timelineIDs(c context.Context, mou *api.NativeModule, ps, offset int64, mid int64, mobiApp, device string) (*actmdl.ResourceReply, error) {
	mixArg := &api.ModuleMixExtsReq{ModuleID: mou.ID, Ps: ps + 6, Offset: offset}
	likeList, err := s.actDao.ModuleMixExts(c, mixArg)
	if err != nil {
		log.Error(" s.actDao.ModuleMixExts(%v) error(%v)", mixArg, err)
		return nil, err
	}
	if likeList == nil {
		return nil, showecode.ActivityNothingMore
	}
	rly := &actmdl.ResourceReply{Offset: likeList.Offset, HasMore: likeList.HasMore}
	lg := len(likeList.List)
	if lg == 0 {
		return rly, nil
	}
	var aids, cvids []int64
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		switch v.MType {
		case api.MixAvidType:
			aids = append(aids, v.ForeignID)
		case api.MixCvidType:
			cvids = append(cvids, v.ForeignID)
		}
	}
	var (
		arcRly map[int64]*arccli.Arc
		artRly map[int64]*artmdl.Meta
	)
	eg := errgroup.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if arcRly, e = s.arcdao.ArchivesPB(ctx, aids, mid, mobiApp, device); e != nil {
				log.Error("s.arcdao.Arcs aids(%v) error(%v)", aids, e)
				e = nil
			}
			return
		})
	}
	if len(cvids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if artRly, e = s.artdao.ArticleMetas(ctx, cvids, 2); e != nil {
				log.Error("s.artdao.ArticleMeta cvids(%v) error(%v)", cvids, e)
				e = nil
			}
			return
		})
	}
	if err = eg.Wait(); err != nil { //发生错误降级处理
		log.Error("timelineIDs eg.Wait error(%v)", err)
		return rly, nil
	}
	toCount := 0
	confSort := mou.ConfUnmarshal()
	for _, v := range likeList.List {
		offset++
		if v == nil {
			continue
		}
		mixRemark := v.RemarkUnmarshal()
		tmp := &actmdl.Item{}
		switch v.MType {
		case api.MixAvidType:
			if v.ForeignID == 0 {
				continue
			}
			if va, ok := arcRly[v.ForeignID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromTimelineArc(arcRly[v.ForeignID])
		case api.MixCvidType:
			if v.ForeignID == 0 {
				continue
			}
			if va, ok := artRly[v.ForeignID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromTimelineArt(artRly[v.ForeignID])
		case api.MixTimelineText:
			tmp.FromTimelineText(mixRemark)
		case api.MixTimelinePic:
			tmp.FromTimelinePic(mixRemark)
		case api.MixTimeline:
			tmp.FromTimeline(mixRemark)
		default:
			continue
		}
		head := &actmdl.Item{}
		//时间轴节点类型 0:文本 1:时间节点
		if confSort.Axis == api.AxisText {
			head.FromTimelineHead(mixRemark)
		} else {
			head.FromTimelineFormatHead(xtime.Time(mixRemark.Stime), confSort)
		}
		rly.List = append(rly.List, head)
		rly.List = append(rly.List, tmp)
		toCount++
		if toCount >= int(ps) {
			break
		}
	}
	if likeList.HasMore == 0 && offset < likeList.Offset {
		rly.HasMore = 1
	}
	rly.Offset = offset
	return rly, nil
}

// ogvSeasonID .
func (s *Service) ogvSeasonID(c context.Context, mou *api.NativeModule, ps, offset, mid int64) (*actmdl.ResourceReply, error) {
	mixArg := &api.ModuleMixExtsReq{ModuleID: mou.ID, Ps: ps + 6, Offset: offset}
	likeList, err := s.actDao.ModuleMixExts(c, mixArg)
	if err != nil {
		log.Error(" s.actDao.ModuleMixExts(%v) error(%v)", mixArg, err)
		return nil, err
	}
	if likeList == nil {
		return nil, showecode.ActivityNothingMore
	}
	rly := &actmdl.ResourceReply{Offset: likeList.Offset, HasMore: likeList.HasMore}
	lg := len(likeList.List)
	if lg == 0 {
		return rly, nil
	}
	var ssids []int32
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		switch v.MType {
		case api.MixOgvSsid:
			ssids = append(ssids, int32(v.ForeignID))
		default:
			continue
		}
	}
	if len(ssids) == 0 {
		return rly, nil
	}
	//根据ssid获取ogv卡片信息
	seaRly, err := s.pgcdao.SeasonBySeasonId(c, ssids, mid)
	if err != nil {
		log.Error("s.pgcdao.SeasonBySeasonId %v,error(%v)", ssids, err)
		//降级处理，不返回错误
		return rly, nil
	}
	toCount := 0
	for _, v := range likeList.List {
		offset++
		if v == nil || v.ForeignID == 0 || v.MType != api.MixOgvSsid {
			continue
		}
		if sVal, ok := seaRly[int32(v.ForeignID)]; !ok || sVal == nil {
			continue
		}
		tmp := &actmdl.Item{}
		tmp.FromOgvSeason(mou, seaRly[int32(v.ForeignID)], v.RemarkUnmarshal().Title)
		rly.List = append(rly.List, tmp)
		toCount++
		if toCount >= int(ps) {
			break
		}
	}
	if likeList.HasMore == 0 && offset < likeList.Offset {
		rly.HasMore = 1
	}
	rly.Offset = offset
	return rly, nil
}

// resourceJoin 获取稿件信息.
func (s *Service) resourceJoin(c context.Context, aids, cvids, epids, fids []int64, roomids *actmdl.ParamLive, mid int64, mobiApp, device string) (map[int64]*arccli.Arc, map[int64]*artmdl.Meta, map[int64]*bgmmdl.EpPlayer, map[int64]*favmdl.Folder, map[int64]*playgrpc.RoomList) {
	var (
		arcRly  map[int64]*arccli.Arc
		artRly  map[int64]*artmdl.Meta
		epRly   map[int64]*bgmmdl.EpPlayer
		foldRly = make(map[int64]*favmdl.Folder)
		roomRly map[int64]*playgrpc.RoomList
	)
	eg := errgroup.WithContext(c)
	if roomids != nil && len(roomids.IDs) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if roomRly, e = s.livedao.GetListByRoomId(ctx, actmdl.RemoveDuplicates(roomids.IDs), roomids.IsLive); e != nil {
				log.Error("s.livedao.GetListByRoomId(%v) error(%v)", roomids.IDs, e)
				return nil
			}
			return nil
		})
	}
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if arcRly, e = s.arcdao.ArchivesPB(ctx, actmdl.RemoveDuplicates(aids), mid, mobiApp, device); e != nil {
				log.Error("s.arcdao.Arcs aids(%v) error(%v)", aids, e)
				return nil
			}
			return nil
		})
	}
	if len(cvids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if artRly, e = s.artdao.ArticleMetas(ctx, actmdl.RemoveDuplicates(cvids), 2); e != nil {
				log.Error("s.artdao.ArticleMeta cvids(%v) error(%v)", cvids, e)
				return nil
			}
			return nil
		})
	}
	if len(epids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			//资源小卡不需要inline播放
			if epRly, e = s.bgmdao.EpPlayer(ctx, actmdl.RemoveDuplicates(epids), nil); e != nil {
				log.Error("s.bgmdao.EpPlayer epids(%v) error(%v)", epids, e)
				e = nil
			}
			return
		})
	}
	if len(fids) > 0 {
		var mu sync.Mutex
		eg.Go(func(ctx context.Context) error {
			folders, err := s.favdao.Folders(ctx, actmdl.RemoveDuplicates(fids), int32(favmdl.TypeVideo))
			if err != nil {
				log.Error("s.favdao.Folders(%v) error(%v)", fids, err)
				return nil
			}
			for _, folder := range folders.GetRes() {
				// 私密播单
				if folder == nil || folder.Attr&1 == 1 {
					continue
				}
				mu.Lock()
				foldRly[folder.Mlid] = folder
				mu.Unlock()
			}
			return nil
		})
	}
	if e := eg.Wait(); e != nil {
		log.Error("resourceJoin eg.Wait() error(%v)", e) //可降级
	}
	return arcRly, artRly, epRly, foldRly, roomRly
}

// editWeekOrigin .
func (s *Service) editWeekOrigin(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	//获取每周必看数据,直接从缓存获取，缓存永不过期的
	weekRly, err := s.cdao.PickSerieCache(c, actmdl.WeekStyle, pas.FID)
	if err != nil {
		log.Error("s.cdao.PickSerieCache(%d,%s) error(%v)", pas.FID, actmdl.WeekStyle, err)
		return nil, err
	}
	rly := &actmdl.ResourceReply{}
	if weekRly == nil || len(weekRly.List) == 0 {
		return rly, nil
	}
	rly.Offset = int64(len(weekRly.List))
	var (
		aids []int64
		fid  int64
	)
	for _, v := range weekRly.List {
		if v == nil || v.RID == 0 {
			continue
		}
		if v.Rtype != "av" {
			continue
		}
		aids = append(aids, v.RID)
	}
	if weekRly.Config != nil && weekRly.Config.MediaID > 0 {
		fid = weekRly.Config.MediaID
	}
	viewedArcs := s.editorViewedArcs(c, mou, pas.Mid)
	arcRly, _, _, foldRly, _ := s.resourceJoin(c, aids, []int64{}, []int64{}, []int64{fid}, nil, pas.Mid, pas.MobiApp, pas.Device)
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	fold := foldRly[fid]
	for _, v := range weekRly.List {
		if v == nil || v.RID == 0 {
			continue
		}
		if v.Rtype != "av" {
			continue
		}
		if aVal, ok := arcRly[v.RID]; !ok || aVal == nil || !aVal.IsNormal() {
			continue
		}
		tmp := &actmdl.Item{}
		tmp.FromEditorArc(arcRly[v.RID], arcDisplay, fold, mou, &actmdl.RcmdContent{BottomContent: v.RcmdReason}, pas.MobiApp, pas.Device, viewedArcs)
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

func (s *Service) editRankOrigin(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	rly := &actmdl.ResourceReply{}
	if mou.Fid <= 0 {
		return rly, nil
	}
	ps := pas.Ps
	maxPs := int64(10) //最多下发10张卡片，产品逻辑
	if ps > maxPs {
		ps = maxPs
	}
	eg := errgroup.WithContext(c)
	rcmRly := make(map[int]*actmdl.RcmdContent)
	eg.Go(func(ctx context.Context) error {
		//获取icon
		mixArg := &api.ModuleMixExtReq{ModuleID: mou.ID, Ps: ps, Offset: pas.Offset, MType: api.MixRankIcon}
		mixIcon, e := s.actDao.ModuleMixExt(ctx, mixArg)
		if e != nil { //降级处理
			log.Error(" s.actDao.ModuleMixExt(%v) error(%v)", mixArg, e)
			return nil
		}
		if mixIcon == nil || len(mixIcon.List) == 0 {
			return nil
		}
		i := 0
		for _, v := range mixIcon.List {
			if v == nil {
				continue
			}
			mixFold := actmdl.MixFolderUnmarshal(v.Reason)
			if mixFold != nil && mixFold.RcmdContent != nil {
				rcmRly[i] = &actmdl.RcmdContent{MiddleIcon: mixFold.RcmdContent.MiddleIcon}
			}
			i++
		}
		return nil
	})
	//获取稿件信息
	var rankRly *actapi.RankResultResp
	eg.Go(func(ctx context.Context) (e error) {
		if rankRly, e = s.actDao.RankResult(ctx, mou.Fid, 1, ps); e != nil {
			log.Error("s.actDao.RankResult(%d,%d) error(%v)", mou.Fid, ps, e)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if rankRly == nil || len(rankRly.List) == 0 {
		return rly, nil
	}
	if rankRly.Page != nil {
		rly.Offset = rankRly.Page.Ps
	}
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	j := 0
	for _, v := range rankRly.List {
		if v == nil || v.ObjectType != 2 || len(v.Archive) < 1 { //稿件榜
			continue
		}
		arcVal := v.Archive[0]
		if arcVal == nil {
			continue
		}
		tmp := &actmdl.Item{}
		rcms := rcmRly[j]
		j++
		tmp.FromEditorRankArc(v, mou, arcVal, arcDisplay, rcms)
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// editChannelOrigin .
func (s *Service) editChannelOrigin(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	if mou.Fid == 0 {
		return &actmdl.ResourceReply{}, nil
	}
	//获取垂类id下对应的aid和epid
	chaRly, err := s.hmtChannelDao.ChannelFeed(c, mou.Fid, pas.Mid, pas.Buvid, int32(pas.Offset), int32(pas.Ps))
	if err != nil {
		log.Error("s.hmtChannelDao.ChannelFeed(%d,%d,%d) error(%v)", mou.Fid, pas.Offset, pas.Ps, err)
		return nil, err
	}
	if chaRly == nil || len(chaRly.List) == 0 {
		return &actmdl.ResourceReply{}, nil
	}
	rly := &actmdl.ResourceReply{Offset: int64(chaRly.GetOffset())}
	//是否还有更多数据
	if chaRly.GetHasMore() {
		rly.HasMore = 1
	}
	var (
		aids  []int64
		epids []int64
	)
	//拼接aids和epids
	for _, v := range chaRly.List {
		if v == nil || v.Id <= 0 {
			continue
		}
		switch v.Type {
		case chagrpc.ResourceType_UGC_RESOURCE:
			aids = append(aids, v.Id)
		case chagrpc.ResourceType_OGV_RESOURCE:
			epids = append(epids, v.Id)
		default:
		}
	}
	arcRly, _, epRly, _, _ := s.resourceJoin(c, aids, []int64{}, epids, []int64{}, nil, pas.Mid, pas.MobiApp, pas.Device)
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	pgcDisplay := mou.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	//拼接卡片信息
	for _, v := range chaRly.List {
		if v == nil || v.Id <= 0 {
			continue
		}
		tmp := &actmdl.Item{}
		switch v.Type {
		case chagrpc.ResourceType_UGC_RESOURCE:
			if aVal, ok := arcRly[v.Id]; !ok || aVal == nil || !aVal.IsNormal() {
				continue
			}
			tmp.FromEditorArc(arcRly[v.Id], arcDisplay, nil, mou, nil, pas.MobiApp, pas.Device, nil)
		case chagrpc.ResourceType_OGV_RESOURCE:
			if va, ok := epRly[v.Id]; !ok || va == nil {
				continue
			}
			// ugc确定每个position固定展示
			tmp.FromEditorEp(epRly[v.Id], pgcDisplay, mou, nil, `{"position2": "duration","position4": "view","position5": "follow"}`)
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// editMustseeOrigin .
func (s *Service) editMustseeOrigin(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	mustseeRly, err := s.populardao.PageArcs(c, pas.Offset, pas.Ps, pas.MustseeType)
	if err != nil {
		log.Error("s.populardao.PageArcs(%d,%d) error(%v)", pas.Offset, pas.Ps, err)
		return nil, err
	}
	rly := &actmdl.ResourceReply{}
	if mustseeRly == nil || len(mustseeRly.List) == 0 {
		return rly, nil
	}
	if mustseeRly.Page != nil {
		rly.Offset = mustseeRly.Page.Offset
		rly.HasMore = int32(mustseeRly.Page.HasMore)
	}
	var (
		aids []int64
		fid  int64
	)
	for _, v := range mustseeRly.List {
		if v == nil || v.Aid <= 0 {
			continue
		}
		aids = append(aids, v.Aid)
	}
	fid = mustseeRly.MediaId
	viewedArcs := s.editorViewedArcs(c, mou, pas.Mid)
	arcRly, _, _, foldRly, _ := s.resourceJoin(c, aids, []int64{}, []int64{}, []int64{fid}, nil, pas.Mid, pas.MobiApp, pas.Device)
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	fold := foldRly[fid]
	for _, v := range mustseeRly.List {
		if v == nil || v.Aid <= 0 {
			continue
		}
		if aVal, ok := arcRly[v.Aid]; !ok || aVal == nil || !aVal.IsNormal() {
			continue
		}
		tmp := &actmdl.Item{}
		tmp.FromEditorArc(arcRly[v.Aid], arcDisplay, fold, mou, &actmdl.RcmdContent{TopContent: v.Recommend}, pas.MobiApp, pas.Device, viewedArcs)
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// ResourceAvid .
// nolint:gocognit
func (s *Service) ResourceAvid(c context.Context, mou *api.NativeModule, ps, offset int64, param *actmdl.ParamFormatModule) (*actmdl.ResourceReply, error) {
	mixArg := &api.ModuleMixExtsReq{ModuleID: mou.ID, Ps: ps + 6, Offset: offset}
	likeList, err := s.actDao.ModuleMixExts(c, mixArg)
	if err != nil {
		log.Error(" s.actDao.ModuleMixExts(%v) error(%v)", mixArg, err)
		return nil, err
	}
	if likeList == nil {
		return nil, nil
	}
	rly := &actmdl.ResourceReply{Offset: likeList.Offset, HasMore: likeList.HasMore}
	lg := len(likeList.List)
	if lg == 0 {
		return rly, nil
	}
	edViewedArcs := s.editorViewedArcs(c, mou, param.Mid)
	var aids, cvids, epids, fids []int64
	roomids := &actmdl.ParamLive{}
	mixFolders := make(map[string]*actmdl.MixFolder)
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		mixFold := actmdl.MixFolderUnmarshal(v.Reason)
		if mixFold != nil {
			mixFolders[v.Reason] = mixFold
		}
		switch v.MType {
		case api.MixAvidType, api.MixFolder:
			aids = append(aids, v.ForeignID)
			if mixFold, ok := mixFolders[v.Reason]; ok && v.MType == api.MixFolder {
				fids = append(fids, mixFold.Fid)
			}
		case api.MixCvidType:
			cvids = append(cvids, v.ForeignID)
		case api.MixEpidType:
			epids = append(epids, v.ForeignID)
		case api.MixLive:
			roomids.IDs = append(roomids.IDs, v.ForeignID)
		}
	}
	if len(roomids.IDs) > 0 {
		roomids.IsLive = mou.IsAttrDisplayNodeNum()
	}
	arcRly, artRly, epRly, foldRly, roomRly := s.resourceJoin(c, aids, cvids, epids, fids, roomids, param.Mid, param.MobiApp, param.Device)
	artDisplay := mou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	pgcDisplay := mou.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	for _, v := range likeList.List {
		offset++
		if v == nil || v.ForeignID == 0 {
			continue
		}
		tmp := &actmdl.Item{}
		var rcmdContent *actmdl.RcmdContent
		if mou.IsEditor() && mixFolders[v.Reason] != nil {
			rcmdContent = mixFolders[v.Reason].RcmdContent
		}
		switch v.MType {
		case api.MixAvidType, api.MixFolder:
			if va, ok := arcRly[v.ForeignID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			if _, ok := mixFolders[v.Reason]; !ok && v.MType == api.MixFolder {
				continue
			}
			if v.MType == api.MixAvidType {
				if mou.IsEditor() {
					tmp.FromEditorArc(arcRly[v.ForeignID], arcDisplay, nil, mou, rcmdContent, param.MobiApp, param.Device, edViewedArcs)
				} else {
					tmp.FromResourceArc(arcRly[v.ForeignID], arcDisplay, nil)
				}
			} else {
				mixFold := mixFolders[v.Reason]
				if fold, ok := foldRly[mixFold.Fid]; !ok || fold == nil {
					continue
				}
				if mou.IsEditor() {
					tmp.FromEditorArc(arcRly[v.ForeignID], arcDisplay, foldRly[mixFold.Fid], mou, rcmdContent, param.MobiApp, param.Device, edViewedArcs)
				} else {
					tmp.FromResourceArc(arcRly[v.ForeignID], arcDisplay, foldRly[mixFold.Fid])
				}
			}
		case api.MixCvidType:
			if va, ok := artRly[v.ForeignID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			if mou.IsEditor() {
				tmp.FromEditorArt(artRly[v.ForeignID], artDisplay, mou, rcmdContent)
			} else {
				tmp.FromResourceArt(artRly[v.ForeignID], artDisplay)
			}
		case api.MixEpidType:
			if va, ok := epRly[v.ForeignID]; !ok || va == nil {
				continue
			}
			if mou.IsEditor() {
				tmp.FromEditorEp(epRly[v.ForeignID], pgcDisplay, mou, rcmdContent, mou.TName)
			} else {
				tmp.FromResourceEp(epRly[v.ForeignID], pgcDisplay)
			}
		case api.MixLive:
			if va, ok := roomRly[v.ForeignID]; !ok || va == nil {
				continue
			}
			tmp.FromResourceLive(roomRly[v.ForeignID], c, s.c.Feature)
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
		if len(rly.List) >= int(ps) {
			break
		}
	}
	if likeList.HasMore == 0 && offset < likeList.Offset {
		rly.HasMore = 1
	}
	rly.Offset = offset
	return rly, nil
}

func (s *Service) editorViewedArcs(c context.Context, module *api.NativeModule, mid int64) map[int64]struct{} {
	if !(module.IsEditor() || module.IsEditorOrigin()) {
		return nil
	}
	confSort := module.ConfUnmarshal()
	if confSort.Sid == 0 || confSort.Counter == "" {
		return nil
	}
	activity := strconv.FormatInt(confSort.Sid, 10)
	rly, err := s.platDao.GetHistory(c, activity, confSort.Counter, mid, nil)
	if err != nil {
		return nil
	}
	type historySource struct {
		Aid int64 `json:"aid"`
	}
	state := make(map[int64]struct{}, len(rly.GetHistory()))
	for _, his := range rly.GetHistory() {
		source := &historySource{}
		if err := json.Unmarshal([]byte(his.Source), source); err != nil {
			log.Error("Fail to unmarshal HistoryContent.Source, source=%+v error=%+v", source, err)
			continue
		}
		state[source.Aid] = struct{}{}
	}
	return state
}

func (s *Service) ResourceRole(c context.Context, mou *api.NativeModule, offset, ps int) (*actmdl.ResourceReply, error) {
	charID := int32(mou.Length)
	seasonID := int32(mou.Width)
	epIDs, err := s.actDao.GetCharacterEps(c, charID, seasonID)
	if err != nil {
		log.Error("Fail to get characterEps, charID=%d seasonID=%d", charID, seasonID)
		return nil, err
	}
	epIDs, _ = pagingList(epIDs, offset, ps)
	epList, err := s.bgmdao.EpPlayer(c, epIDs, nil)
	if err != nil {
		log.Error("Fail to get epPlayer, epIDs=%+v error=%+v", epIDs, err)
		return nil, err
	}
	pgcDisplay := mou.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	list := make([]*actmdl.Item, 0, len(epList))
	for _, v := range epList {
		if v == nil {
			continue
		}
		item := &actmdl.Item{}
		item.FromResourceEp(v, pgcDisplay)
		list = append(list, item)
	}
	reply := &actmdl.ResourceReply{
		List:    list,
		Offset:  int64(offset + len(list)),
		HasMore: 0, //不分页
	}
	return reply, nil
}

// newAvidInfo .
// nolint:gocognit
func (s *Service) newAvidInfo(c context.Context, arg *actmdl.NewAvidReq) (avidReply *actmdl.NewVideoReply, err error) {
	var (
		likeList *api.ModuleMixExtsReply
	)
	if arg == nil {
		return
	}
	mixArg := &api.ModuleMixExtsReq{ModuleID: arg.ModuleID, Ps: arg.Ps + 6, Offset: arg.Offset}
	if likeList, err = s.actDao.ModuleMixExts(c, mixArg); err != nil || likeList == nil {
		log.Error(" s.actDao.ModuleMixExts(%v) error(%v)", arg, err)
		return
	}
	avidReply = &actmdl.NewVideoReply{Total: likeList.Total, Offset: likeList.Offset, HasMore: likeList.HasMore}
	lg := len(likeList.List)
	if lg == 0 {
		return
	}
	var epids []int64
	var aids []*arccli.PlayAv
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		switch v.MType {
		case api.MixAvidType:
			// 不处理cid信息，不需要聚合aid处理
			aids = append(aids, &arccli.PlayAv{Aid: v.ForeignID})
		case api.MixEpidType:
			epids = append(epids, v.ForeignID)
		default:
			continue
		}
	}
	var (
		arcRly map[int64]*arccli.ArcPlayer
		epRly  map[int64]*bgmmdl.EpPlayer
	)
	eg := errgroup.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if arcRly, e = s.arcdao.ArcsPlayer(ctx, aids); e != nil {
				log.Error("s.arcdao.ArcsPlayer(%v) error(%v)", aids, e)
				e = nil
			}
			return
		})
	}
	if len(epids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			epParam := &bgmmdl.CommonParam{
				MobiApp:  arg.MobiApp,
				Build:    int(arg.Build),
				Device:   arg.Device,
				Platform: arg.Platform,
				XTfIsp:   arg.TfIsp,
			}
			if batchArg, ok := arcmid.FromContext(c); ok {
				epParam.Fnver = int(batchArg.Fnver)
				epParam.Fnval = int(batchArg.Fnval)
			}
			if epRly, e = s.bgmdao.EpPlayer(ctx, epids, epParam); e != nil {
				log.Error("s.bgmdao.EpPlayer epids(%v) error(%v)", epids, e)
				e = nil
			}
			return
		})
	}
	_ = eg.Wait()
	lastOffset := arg.Offset
	for _, v := range likeList.List {
		lastOffset++
		if v == nil || v.ForeignID == 0 {
			continue
		}
		tmp := &actmdl.Item{}
		switch v.MType {
		case api.MixAvidType:
			acVal, ok := arcRly[v.ForeignID]
			if !ok || acVal == nil || acVal.Arc == nil || !acVal.Arc.IsNormal() {
				continue
			}
			// 获取首p秒开地址即可
			firstPlay := acVal.PlayerInfo[acVal.DefaultPlayerCid]
			tmp.FromNewVideoCard(acVal.Arc, firstPlay, arg.Build, arg.MobiApp)
		case api.MixEpidType:
			if ep, ok := epRly[v.ForeignID]; !ok || ep == nil {
				continue
			}
			tmp.FromNewEPCard(epRly[v.ForeignID])
		default:
			continue
		}
		avidReply.Item = append(avidReply.Item, tmp)
		if len(avidReply.Item) >= int(arg.Ps) {
			break
		}
	}
	if likeList.HasMore == 0 && lastOffset < likeList.Offset {
		avidReply.HasMore = 1
	}
	// 重新计算的offset
	avidReply.Offset = lastOffset
	return
}

// AvidInfo .
func (s *Service) AvidInfo(c context.Context, arg *actmdl.AvidReq, mid int64) (avidReply *actmdl.VideoReply, err error) {
	var (
		likeList *api.ModuleMixExtReply
		rids     []*dynamic.RidInfo
		dyRes    *dynamic.DyResult
	)
	if arg == nil {
		return
	}
	mixArg := &api.ModuleMixExtReq{ModuleID: arg.ModuleID, Ps: arg.Ps + 6, Offset: arg.Offset, MType: api.MixAvidType}
	if likeList, err = s.actDao.ModuleMixExt(c, mixArg); err != nil || likeList == nil {
		log.Error(" s.actDao.ModuleMixExt(%v) error(%v)", arg, err)
		return
	}
	avidReply = &actmdl.VideoReply{Total: likeList.Total, Offset: likeList.Offset, HasMore: likeList.HasMore}
	lg := len(likeList.List)
	if lg == 0 {
		return
	}
	rids = make([]*dynamic.RidInfo, 0, lg)
	for _, v := range likeList.List {
		if v.ForeignID > 0 {
			rids = append(rids, &dynamic.RidInfo{Rid: v.ForeignID, Type: arg.AvSort})
		}
	}
	rous := &dynamic.Resources{Array: rids}
	if dyRes, err = s.dynamicDao.Dynamic(c, rous, arg.Platform, arg.RemoteFrom, arg.FromSpmid, mid, nil); err != nil || dyRes == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", rous, err)
		return
	}
	lastOffset := arg.Offset
	for _, v := range rids {
		lastOffset++
		if _, ok := dyRes.Cards[v.Rid]; !ok {
			continue
		}
		avidReply.DyReply = append(avidReply.DyReply, dyRes.Cards[v.Rid])
		if len(avidReply.DyReply) >= int(arg.Ps) {
			break
		}
	}
	if likeList.HasMore == 0 && lastOffset < likeList.Offset {
		avidReply.HasMore = 1
	}
	// 重新计算的offset
	avidReply.Offset = lastOffset
	return
}

func (s *Service) editOriginList(c context.Context, arg *actmdl.ParamLike, mid int64) *actmdl.LikeListRely {
	//获取module
	eg := errgroup.WithContext(c)
	var moduleConf *api.NativeModule
	eg.Go(func(ctx context.Context) error {
		modReply, err := s.actDao.ModuleConfig(ctx, &api.ModuleConfigReq{ModuleID: arg.ConfModuleID})
		if err != nil {
			log.Error("s.actDao.ModuleConfig(%d) error(%v)", arg.ConfModuleID, err) //降级处理
		}
		if modReply != nil && modReply.Module != nil && modReply.Module.NativeModule != nil {
			moduleConf = modReply.Module.NativeModule
		} else {
			moduleConf = &api.NativeModule{ID: arg.ConfModuleID}
		}
		return nil
	})
	var natPage *api.NativePage
	if arg.PrimaryPageID > 0 {
		eg.Go(func(ctx context.Context) error {
			var e error
			if natPage, e = s.actDao.NativePage(ctx, arg.PrimaryPageID); e != nil {
				log.Error("s.actDao.NativePage(%d) error(%v)", arg.PrimaryPageID, e)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		//降级处理
		log.Error("editOriginList wait error(%d,%d,%v)", arg.ConfModuleID, arg.PrimaryPageID, err)
	}
	confSort := moduleConf.ConfUnmarshal()
	rly := &actmdl.LikeListRely{}
	originParam := &actmdl.ResourceOriginReq{RdbType: int64(arg.SortType), Ps: int64(arg.Ps), Offset: arg.Offset, MobiApp: arg.MobiApp, Device: arg.Device, MustseeType: confSort.MseeType, Mid: mid, Buvid: arg.Buvid}
	reply, err := s.editOrigin(c, originParam, moduleConf)
	if err != nil {
		log.Error("s.editOrigin(%d) error(%v)", arg.SortType, err)
		//降级处理
		return rly
	}
	if reply == nil {
		return rly
	}
	if natPage != nil {
		rly.Color = &actmdl.Color{BgColor: natPage.BgColor}
		rly.AttrBit = &actmdl.AttrBit{NotNight: natPage.IsAttrNotNightModule() == api.AttrModuleYes}
	}
	rly.Offset = reply.Offset
	rly.HasMore = reply.HasMore
	if len(reply.List) > 0 {
		list := s.EditorJoin(c, reply, moduleConf)
		rly.Cards = []*actmdl.Item{list}
	}
	return rly
}

// newVideoDynList  新视频卡dyn类型.
func (s *Service) newVideoDynList(c context.Context, arg *actmdl.ParamLike, mid int64) (*actmdl.LikeListRely, error) {
	var (
		types = strconv.Itoa(dynamic.VideoType)
	)
	if arg == nil || arg.TopicID <= 0 {
		return nil, ecode.RequestErr
	}
	netType, tfType := showmdl.TrafficFree(arg.TfIsp)
	dynReq := &actmdl.NewDynReq{TopicID: arg.TopicID, Types: types, PageSize: int64(arg.Ps), Mid: mid,
		MobiApp: arg.MobiApp, Buvid: arg.Buvid, Build: arg.Build, Platform: arg.Platform, NetType: netType, TfType: tfType, DyOffset: arg.DyOffset}
	reply, err := s.newVideoDynamic(c, dynReq)
	if err != nil {
		log.Error("s.dynamicDao.FetchDynamics(%d) error(%v)", arg.TopicID, err)
		return nil, err
	}
	return s.videoSecondJoin(reply), nil
}

// newVideoActList 新视频卡act模式
func (s *Service) newVideoActList(c context.Context, arg *actmdl.ParamLike, mid int64) (*actmdl.LikeListRely, error) {
	if arg == nil || arg.Sid <= 0 || arg.SortType <= 0 {
		return nil, ecode.RequestErr
	}
	actReq := &actmdl.NewVideoActReq{Sid: arg.Sid, SortType: arg.SortType, Ps: int64(arg.Ps), Offset: arg.Offset,
		MobiApp: arg.MobiApp, Platform: arg.Platform, Build: arg.Build, Buvid: arg.Buvid, Device: arg.Device}
	actReply, err := s.newVideoAct(c, actReq, mid)
	if err != nil {
		log.Error(" s.newVideoAct(%v) error(%v)", actReq, err)
		return nil, err
	}
	return s.videoSecondJoin(actReply), nil
}

// VideoDynList  视频卡dyn类型.
func (s *Service) VideoDynList(c context.Context, arg *actmdl.ParamLike, mid int64, isSingle bool) (res *actmdl.LikeListRely, err error) {
	var (
		types = strconv.Itoa(dynamic.VideoType)
		reply *dynamic.DyReply
		list  *actmdl.Item
	)
	if arg == nil || arg.TopicID <= 0 {
		return
	}
	// dysort 使用默认0
	if reply, err = s.dynamicDao.FetchDynamics(c, arg.TopicID, mid, int64(arg.Ps), 0, arg.Device, types, arg.Platform, arg.DyOffset, arg.RemoteFrom, arg.FromSpmid, ""); err != nil || reply == nil {
		log.Error("s.dynamicDao.FetchDynamics(%d) error(%v)", arg.TopicID, err)
		return
	}
	if len(reply.Cards) == 0 {
		return
	}
	res = &actmdl.LikeListRely{DyOffset: reply.Offset, HasMore: int32(reply.HasMore)}
	list = &actmdl.Item{}
	list.FromVideoDynModule(nil, isSingle)
	for _, v := range reply.Cards {
		tmpAct := &actmdl.Item{}
		tmpAct.FromVideoCard(v, isSingle)
		list.Item = append(list.Item, tmpAct)
	}
	res.Cards = []*actmdl.Item{list}
	return
}

// VideoAvidList 视频卡act类型.
func (s *Service) VideoActList(c context.Context, arg *actmdl.ParamLike, mid int64, isSingle bool) (res *actmdl.LikeListRely, err error) {
	var (
		actReply *actmdl.VideoReply
		list     *actmdl.Item
	)
	if arg == nil || arg.Sid <= 0 || arg.SortType <= 0 {
		return
	}
	actReq := &actmdl.VideoActReq{Sid: arg.Sid, SortType: arg.SortType, Ps: int64(arg.Ps), Offset: arg.Offset,
		VideoMeta: arg.VideoMeta, MobiApp: arg.MobiApp, Platform: arg.Platform, Build: arg.Build, Buvid: arg.Buvid, RemoteFrom: arg.RemoteFrom, Device: arg.Device, TfIsp: arg.TfIsp, FromSpmid: arg.FromSpmid}
	if actReply, err = s.VideoAct(c, actReq, mid); err != nil || actReply == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", actReq, err)
		return
	}
	res = &actmdl.LikeListRely{Offset: actReply.Offset, HasMore: actReply.HasMore}
	list = &actmdl.Item{}
	list.FromVideoActModule(nil, isSingle)
	for _, v := range actReply.DyReply {
		temp := &actmdl.Item{}
		temp.FromVideoCard(v, isSingle)
		list.Item = append(list.Item, temp)
	}
	res.Cards = []*actmdl.Item{list}
	return
}

func (s *Service) ResourceOriginList(c context.Context, arg *actmdl.ParamLike, mid int64) (*actmdl.LikeListRely, error) {
	//参数校验
	if arg == nil || arg.ModuleID <= 0 || arg.AvSort <= 0 || arg.DyType == "" {
		return nil, ecode.RequestErr
	}
	dynReq := &actmdl.ResourceOriginReq{SourceID: arg.DyType, RdbType: int64(arg.AvSort), Ps: int64(arg.Ps), Offset: arg.Offset, Mid: mid, MobiApp: arg.MobiApp, Device: arg.Device, Platform: arg.Platform, Build: arg.Build}
	if dynReq.RdbType == api.RDBLive {
		dynReq.SortType = int64(arg.SortType)
	}
	avidReply, err := s.resourceOrigin(c, dynReq, &api.NativeModule{ID: arg.ModuleID, Attribute: arg.Attr})
	if err != nil {
		log.Error("s.resourceOrigin or nil error(%v)", err)
		return nil, err
	}
	if avidReply == nil {
		return &actmdl.LikeListRely{}, nil
	}
	rly := &actmdl.LikeListRely{Offset: avidReply.Offset, HasMore: avidReply.HasMore}
	if len(avidReply.List) == 0 {
		return rly, nil
	}
	list := &actmdl.Item{}
	list.FromResourceModule(c, s.c.Feature, nil, arg.MobiApp, arg.Build)
	list.Item = append(list.Item, avidReply.List...)
	rly.Cards = []*actmdl.Item{list}
	return rly, nil
}

// ResourceActList  资源卡act类型.
func (s *Service) ResourceActList(c context.Context, arg *actmdl.ParamLike, mid int64) (*actmdl.LikeListRely, error) {
	//参数校验
	if arg == nil || arg.ModuleID <= 0 || arg.SortType <= 0 || arg.Sid <= 0 {
		return nil, ecode.RequestErr
	}
	dynReq := &actmdl.ResourceActReq{Sid: arg.Sid, SortType: arg.SortType, Ps: int64(arg.Ps), Offset: arg.Offset, Mid: mid, MobiApp: arg.MobiApp, Device: arg.Device}
	avidReply, err := s.ResourceAct(c, dynReq, &api.NativeModule{ID: arg.ModuleID, Attribute: arg.Attr})
	if err != nil {
		log.Error("s.ResourceAct or nil error(%v)", err)
		return nil, err
	}
	if avidReply == nil {
		return &actmdl.LikeListRely{}, nil
	}
	rly := &actmdl.LikeListRely{Offset: avidReply.Offset, HasMore: avidReply.HasMore}
	list := &actmdl.Item{}
	list.FromResourceModule(c, s.c.Feature, nil, arg.MobiApp, arg.Build)
	list.Item = append(list.Item, avidReply.List...)
	rly.Cards = []*actmdl.Item{list}
	return rly, nil
}

// newVideoAvidList 新视频卡id模式
func (s *Service) newVideoAvidList(c context.Context, arg *actmdl.ParamLike) (*actmdl.LikeListRely, error) {
	var (
		err       error
		avidReply *actmdl.NewVideoReply
	)
	//参数校验
	if arg == nil || arg.ModuleID <= 0 {
		return nil, ecode.RequestErr
	}
	netType, tfType := showmdl.TrafficFree(arg.TfIsp)
	avidReq := &actmdl.NewAvidReq{ModuleID: arg.ModuleID, Offset: arg.Offset, Ps: int64(arg.Ps),
		MobiApp: arg.MobiApp, Platform: arg.Platform, Build: arg.Build, Buvid: arg.Buvid, Device: arg.Device, TfIsp: arg.TfIsp, TfType: tfType, NetType: netType}
	if avidReply, err = s.newAvidInfo(c, avidReq); err != nil {
		log.Error("s.newAvidInfo error(%v)", err)
		return nil, err
	}
	return s.videoSecondJoin(avidReply), nil
}

// ResourceDynList  资源卡idyn类型.
func (s *Service) ResourceDynList(c context.Context, arg *actmdl.ParamLike, mid int64) (*actmdl.LikeListRely, error) {
	//参数校验
	if arg == nil || arg.ModuleID <= 0 || arg.TopicID <= 0 {
		return nil, ecode.RequestErr
	}
	dynReq := &dynamic.ResourceDynReq{TopicID: arg.TopicID, Types: arg.DyType, PageSize: int64(arg.Ps), Offset: arg.DyOffset, Mid: mid, MobiApp: arg.MobiApp, Device: arg.Device}
	avidReply, err := s.ResourceDyn(c, dynReq, &api.NativeModule{ID: arg.ModuleID, Attribute: arg.Attr})
	if err != nil {
		log.Error("s.dynamicDao.Dynamic or nil error(%v)", err)
		return nil, err
	}
	if avidReply == nil {
		return &actmdl.LikeListRely{}, nil
	}
	rly := &actmdl.LikeListRely{DyOffset: avidReply.DyOffset, HasMore: avidReply.HasMore}
	list := &actmdl.Item{}
	list.FromResourceModule(c, s.c.Feature, nil, arg.MobiApp, arg.Build)
	list.Item = append(list.Item, avidReply.List...)
	rly.Cards = []*actmdl.Item{list}
	return rly, nil
}

// ResourceIDList 资源卡id类型.
func (s *Service) ResourceIDList(c context.Context, arg *actmdl.ParamLike, mid int64) (*actmdl.LikeListRely, error) {
	//参数校验
	if arg == nil || arg.ModuleID <= 0 {
		return nil, ecode.RequestErr
	}
	modParams := &actmdl.ParamFormatModule{Mid: mid, MobiApp: arg.MobiApp, Device: arg.Device}
	avidReply, err := s.ResourceAvid(c, &api.NativeModule{ID: arg.ModuleID, Attribute: arg.Attr}, int64(arg.Ps), arg.Offset, modParams)
	if err != nil {
		log.Error("s.dynamicDao.Dynamic or nil error(%v)", err)
		return nil, err
	}
	if avidReply == nil {
		return &actmdl.LikeListRely{}, nil
	}
	rly := &actmdl.LikeListRely{Offset: avidReply.Offset, HasMore: avidReply.HasMore}
	list := &actmdl.Item{}
	list.FromResourceModule(c, s.c.Feature, nil, arg.MobiApp, arg.Build)
	list.Item = append(list.Item, avidReply.List...)
	rly.Cards = []*actmdl.Item{list}
	return rly, nil
}

// VideoAvidList 视频卡avid类型.
func (s *Service) VideoAvidList(c context.Context, arg *actmdl.ParamLike, mid int64, isSingle bool) (res *actmdl.LikeListRely, err error) {
	var (
		list      *actmdl.Item
		avidReply *actmdl.VideoReply
	)
	//参数校验
	if arg == nil || arg.ModuleID <= 0 || arg.AvSort <= 0 {
		return
	}
	avidReq := &actmdl.AvidReq{ModuleID: arg.ModuleID, Offset: arg.Offset, Ps: int64(arg.Ps), VideoMeta: arg.VideoMeta,
		MobiApp: arg.MobiApp, Platform: arg.Platform, Build: arg.Build, AvSort: arg.AvSort, Buvid: arg.Buvid, RemoteFrom: arg.RemoteFrom, Device: arg.Device, TfIsp: arg.TfIsp, FromSpmid: arg.FromSpmid}
	if avidReply, err = s.AvidInfo(c, avidReq, mid); err != nil || avidReply == nil {
		log.Error("s.dynamicDao.Dynamic or nil error(%v)", err)
		return
	}
	res = &actmdl.LikeListRely{Offset: avidReply.Offset, HasMore: avidReply.HasMore}
	list = &actmdl.Item{}
	list.FromVideoAvidModule(nil, isSingle)
	for _, v := range avidReply.DyReply {
		temp := &actmdl.Item{}
		temp.FromVideoCard(v, isSingle)
		list.Item = append(list.Item, temp)
	}
	res.Cards = []*actmdl.Item{list}
	return
}

// VideoList 话题活动卡片.
// nolint:gocognit
func (s *Service) VideoList(c context.Context, arg *actmdl.ParamLike, mid int64) (res *actmdl.LikeListRely, err error) {
	var (
		likeList   *actapi.LikesReply
		rids       []*dynamic.RidInfo
		dyResult   *dynamic.DyResult
		itemObj    map[int64]*actapi.ItemObj
		attentions []int64
		list       *actmdl.Item
	)
	// 参数校验
	if arg.Sid == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	if mid > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			attes, e := s.reldao.Attentions(ctx, mid)
			if e != nil {
				log.Error("s.reldao.Attentions(%d) error(%v)", mid, e)
				return nil
			}
			attentions = make([]int64, 0, len(attes))
			for _, v := range attes {
				attentions = append(attentions, v.Mid)
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		var offset int64
		if arg.Offset >= 0 {
			offset = arg.Offset
		} else {
			offset = int64((arg.Pn - 1) * arg.Ps)
		}
		likeReq := &actapi.ActLikesReq{Sid: arg.Sid, Mid: mid, SortType: arg.SortType, Ps: arg.Ps, Offset: offset}
		// 默认default值
		if likeReq.SortType == 0 {
			likeReq.SortType = 1
		}
		if likeList, e = s.actDao.ActLikes(ctx, likeReq); e != nil {
			log.Error("s.actDao.ActLikes(%d) error(%v)", arg.Sid, e)
		}
		return
	})
	var moduleConf *api.NativeModule
	if arg.RemoteFrom == actmdl.RemoteActivity && arg.ConfModuleID > 0 {
		eg.Go(func(ctx context.Context) error {
			var (
				rely *api.ModuleConfigReply
				err  error
			)
			if rely, err = s.actDao.ModuleConfig(ctx, &api.ModuleConfigReq{ModuleID: arg.ConfModuleID}); err != nil {
				log.Error("s.actDao.ModuleConfig(%d) error(%v)", arg.ConfModuleID, err)
				return nil //降级处理
			}
			if rely != nil && rely.Module != nil && rely.Module.NativeModule != nil {
				moduleConf = rely.Module.NativeModule
				// fix: 下发is_feed后会命中安卓的一个容错逻辑：
				// 如果是无限feed组件（is_feed为true），这个组件会做为最后一个组件，且当前接口has_more为false。
				// 最后导致了现在线上安卓这边加载动态视频无限feed只会加载一页
				if arg.Platform == "android" {
					moduleConf.Attribute &= math.MaxInt64 - 1<<api.AttrIsLast
				}
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	if likeList == nil || likeList.Subject == nil {
		return
	}
	lg := len(likeList.List)
	if lg == 0 {
		return
	}
	rids = make([]*dynamic.RidInfo, 0, lg)
	itemObj = make(map[int64]*actapi.ItemObj, lg)
	widType := dynamic.VideoType
	if likeList.Subject.Type == activitymdl.Article {
		widType = dynamic.ArticleType
	}
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			rids = append(rids, &dynamic.RidInfo{Rid: v.Item.Wid, Type: int64(widType)})
			itemObj[v.Item.Wid] = v
		}
	}
	rous := &dynamic.Resources{Array: rids}
	if dyResult, err = s.dynamicDao.Dynamic(c, rous, arg.Platform, arg.RemoteFrom, arg.FromSpmid, mid, nil); err != nil || dyResult == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", rous, err)
		return
	}
	list = &actmdl.Item{}
	list.FromVideoModule(moduleConf, nil)
	list.Item = make([]*actmdl.Item, 0, lg)
	for _, v := range rids {
		if _, ok := dyResult.Cards[v.Rid]; !ok {
			continue
		}
		if _, k := itemObj[v.Rid]; !k {
			continue
		}
		temp := &actmdl.Item{}
		if likeList.Subject.Type == activitymdl.VideoLike {
			temp.FromVideoLike(dyResult.Cards[v.Rid], itemObj[v.Rid])
		} else {
			temp.FromVideo(dyResult.Cards[v.Rid])
		}
		list.Item = append(list.Item, temp)
	}
	res = &actmdl.LikeListRely{Cards: []*actmdl.Item{list}, Page: &actmdl.Page{Ps: arg.Ps, Pn: arg.Pn, Total: likeList.Total}, Offset: likeList.Offset, HasMore: likeList.HasMore}
	if len(attentions) > 0 {
		res.Attentions = &actmdl.Attentions{Uids: attentions}
	}
	return
}

func defaultModule(nid int64) (rly []*api.Module) {
	rly = []*api.Module{
		{
			NativeModule: &api.NativeModule{
				ID:       1,
				Category: 1,
				NativeID: nid,
				State:    1,
				Rank:     0,
				Meta:     "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/sImdqf73_w1125_h180.png",
				Width:    375,
				Length:   60,
			},
			Click: &api.Click{
				Areas: []*api.NativeClick{{State: 1, Leftx: 274, Lefty: 5, Width: 100, Length: 50, Link: "https://d.bilibili.com/download_app.html?schema=1"}},
			},
		},
		{
			NativeModule: &api.NativeModule{
				ID:       2,
				Category: 2,
				NativeID: nid,
				State:    1,
				Rank:     1,
				Meta:     "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/JdIs1OpI_w1125_h180.jpg",
				Num:      20,
				Title:    "话题讨论",
			},
			Dynamic: &api.Dynamic{
				SelectList: []*api.NativeDynamicExt{{State: 1}},
			},
		},
	}
	return
}

// InlineTab .
// nolint:gocognit
func (s *Service) InlineTab(c context.Context, a *actmdl.ParamInlineTab, mid int64) (*actmdl.InlineReply, error) {
	pageConf, e := s.actDao.BaseConfig(c, &api.BaseConfigReq{Pid: a.PageID, Offset: a.Offset, Ps: a.Ps, PType: api.CommonPage})
	if e != nil || pageConf == nil || pageConf.NativePage == nil {
		if ecode.EqualError(natecode.NativePageOffline, e) {
			e = xecode.AppPageOffline
		}
		log.Error("s.actDao.NatConfig(%d) error(%v)", a.PageID, e)
		return nil, e
	}
	commonConf := pageConf.NativePage
	if !commonConf.IsInlineAct() || (commonConf.SkipURL != "") {
		return nil, xecode.AppPageOffline
	}
	//白名单check
	if err := s.checkWhite(c, pageConf.FirstPage, mid); err != nil {
		return nil, err
	}
	_ = s.report.reportPageView(c, s.c.NaInfoc.PageViewLogID, &PageViewReport{
		Mid:      mid,
		PageID:   commonConf.ID,
		FromType: commonConf.FromType,
		Type:     commonConf.Type,
		MobiApp:  a.MobiApp,
		Build:    a.Build,
	})
	//是否锁定
	lockExt := commonConf.ConfSetUnmarshal()
	if lockExt.DT == api.NeedUnLock { //解锁模式
		var deblocking bool
		if lockExt.DC == api.UnLockTime && lockExt.Stime <= time.Now().Unix() { //时间模式&&到达解锁时间
			deblocking = true
		}
		//未解锁时
		if !deblocking {
			return nil, showecode.ActivityHasLock
		}
	}
	//是否锁定
	// 拼接组件信息
	modulesRly := &actmdl.ModulesReply{}
	eg := errgroup.WithContext(c)
	formatArg := &actmdl.ParamFormatModule{PageID: a.PageID, Device: a.Device, VideoMeta: a.VideoMeta, Platform: a.Platform, Buvid: a.Buvid, Build: a.Build, MobiApp: a.MobiApp, TfIsp: a.TfIsp, HttpsUrlReq: a.HttpsUrlReq, FromSpmid: a.FromSpmid, Mid: mid, UserAgent: a.UserAgent, Memory: a.Memory}
	s.formatModule(c, pageConf.Bases, eg, commonConf, mid, formatArg, modulesRly, actmdl.FormatModFromInline)
	var attentions []int64
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			attes, e := s.reldao.Attentions(ctx, mid)
			if e != nil {
				log.Error("s.reldao.Attentions(%d) error(%v)", mid, e)
				return nil
			}
			attentions = make([]int64, 0, len(attes))
			for _, v := range attes {
				attentions = append(attentions, v.Mid)
			}

			return nil
		})
	}
	_ = eg.Wait()
	reply := &actmdl.InlineReply{
		PageID:     commonConf.ID,
		Title:      commonConf.Title,
		VersionMsg: "当前版本较低，无法显示完全，请更新至最新版本后查看",
	}
	reply.Offset = pageConf.Offset
	reply.HasMore = pageConf.HasMore
	if len(modulesRly.Card) > 0 {
		//评论组件和动态无限feed流互斥
		var hasFeed bool
		for _, v := range pageConf.Bases {
			if v.NativeModule == nil {
				continue
			}
			val, k := modulesRly.Card[v.NativeModule.ID]
			if !k || val == nil {
				continue
			}
			switch val.Goto {
			case actmdl.GotoNavigationModule: // inline 页面与导航组件互斥
				continue
			case actmdl.GotoReplyModule:
				if hasFeed { // 如果有互斥组件，需要丢弃
					continue
				}
				hasFeed = true
			case actmdl.GotoDynamicModule:
				if val.IsFeed == 1 {
					if hasFeed {
						continue
					}
					hasFeed = true
				}
			}
			reply.Items = append(reply.Items, val)
		}
	}
	if len(attentions) > 0 {
		reply.Attentions = &actmdl.Attentions{Uids: attentions}
	}
	if a.PageID == 170466 && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.CompatIpadBPU2021, nil) {
		reply.Items = compatIpadBpu2021(reply.Items)
	}
	return reply, nil
}

// MenuTab .
// nolint:gocognit
func (s *Service) MenuTab(c context.Context, a *actmdl.ParamMenuTab, mid int64) (*actmdl.MenuReply, error) {
	pageConf, e := s.actDao.BaseConfig(c, &api.BaseConfigReq{Pid: a.PageID, Offset: a.Offset, Ps: a.Ps, PType: api.CommonPage})
	if e != nil || pageConf == nil || pageConf.NativePage == nil {
		if ecode.EqualError(natecode.NativePageOffline, e) {
			e = xecode.AppPageOffline
		}
		log.Error("s.actDao.NatConfig(%d) error(%v)", a.PageID, e)
		return nil, e
	}
	commonConf := pageConf.NativePage
	_ = s.report.reportPageView(c, s.c.NaInfoc.PageViewLogID, &PageViewReport{
		Mid:      mid,
		PageID:   commonConf.ID,
		FromType: commonConf.FromType,
		Type:     commonConf.Type,
		MobiApp:  a.MobiApp,
		Build:    a.Build,
	})
	var (
		needBase []*api.Module
		tabConf  *actmdl.TabConf
		opFrom   string
	)
	switch {
	case commonConf.IsBottomAct():
		opFrom = actmdl.FormatModFromMenuBottom
		needBase = actmdl.ChooseBottom(pageConf.Bases)
	case commonConf.IsMenuAct():
		opFrom = actmdl.FormatModFromMenuTab
		needBase = actmdl.ChooseMenu(pageConf.Bases)
		//首页配置信息
		tabConf = actmdl.TabConfJoin(commonConf.ConfSetUnmarshal())
	case commonConf.IsOgvAct():
		opFrom = actmdl.FormatModFromMenuOGV
		needBase = actmdl.ChooseOgv(pageConf.Bases)
	case commonConf.IsUgcAct():
		opFrom = actmdl.FormatModFromMenuUGC
		needBase = actmdl.ChooseUgc(pageConf.Bases)
	case commonConf.IsPlayerAct():
		opFrom = actmdl.FormatModFromMenuPlayer
		needBase = actmdl.ChoosePlayer(pageConf.Bases)
	case commonConf.IsSpaceAct():
		opFrom = actmdl.FormatModFromMenuSpace
		needBase = pageConf.Bases
	case commonConf.IsUpTopicAct():
		opFrom = actmdl.FormatModFromMenuUp
		actmdl.SetUpCurrentActPage(pageConf.Bases, a.PageID)
		needBase = pageConf.Bases
	case commonConf.IsLiveTabAct():
		opFrom = actmdl.FormatModFromMenuLive
		needBase = actmdl.ChooseLiveTab(pageConf.Bases)
	case commonConf.IsNewact():
		opFrom = actmdl.FormatModFromMenuNewact
		needBase = actmdl.ChooseNewact(pageConf.Bases)
	default:
		return nil, xecode.AppPageOffline
	}
	reply := &actmdl.MenuReply{
		PageID:     commonConf.ID,
		Title:      commonConf.Title,
		VersionMsg: "当前版本较低，无法显示完全，请更新至最新版本后查看",
		AttrBit:    &actmdl.AttrBit{NotNight: commonConf.IsAttrNotNightModule() == api.AttrModuleYes},
		Color:      &actmdl.Color{BgColor: commonConf.BgColor},
		TabConf:    tabConf,
		Bases:      &actmdl.Bases{},
	}
	if len(needBase) == 0 {
		return reply, nil
	}
	// 拼接组件信息
	modulesRly := &actmdl.ModulesReply{}
	eg := errgroup.WithContext(c)
	formatArg := &actmdl.ParamFormatModule{PageID: a.PageID, Device: a.Device, VideoMeta: a.VideoMeta, Platform: a.Platform, Buvid: a.Buvid, Build: a.Build,
		MobiApp: a.MobiApp, TfIsp: a.TfIsp, HttpsUrlReq: a.HttpsUrlReq, FromSpmid: a.FromSpmid, Mid: mid, UserAgent: a.UserAgent, Memory: a.Memory, TabFrom: a.TabFrom}
	s.formatModule(c, needBase, eg, commonConf, mid, formatArg, modulesRly, opFrom)
	var partiModule, bottomBtnModule *api.Module
	for _, module := range pageConf.BaseModules {
		if module.NativeModule == nil {
			continue
		}
		if module.NativeModule.IsPart() && module.Participation != nil && len(module.Participation.List) > 0 {
			partiModule = module
			break
		}
		if !commonConf.IsNewact() && module.NativeModule.IsBaseBottomButton() {
			bottomBtnModule = module
			break
		}
	}
	func() {
		if a.PageID == s.c.S11Cfg.PageID && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.LiveTabParticipation, nil) {
			reply.Bases.Participation = &actmdl.Item{
				Image:   "http://i0.hdslb.com/bfs/activity-plat/static/20210928/ad1ca28ff355085e142ad591df9c6f88/39zeSoTHKv.png",
				UnImage: "http://i0.hdslb.com/bfs/activity-plat/static/20210928/ad1ca28ff355085e142ad591df9c6f88/39zeSoTHKv.png",
			}
			return
		}
		if partiModule == nil {
			return
		}
		eg.Go(func(ctx context.Context) error {
			reply.Bases.Participation = s.FromPart(ctx, partiModule.NativeModule, partiModule.Participation, commonConf)
			return nil
		})
	}()
	if bottomBtnModule != nil && bottomBtnModule.NativeModule != nil {
		btnClick := bottomBtnModule.Click
		eg.Go(func(ctx context.Context) error {
			if bottomBtnModule.NativeModule.Meta == "" {
				return nil
			}
			reply.Bases.BottomButton = s.FormatClick(ctx, bottomBtnModule.NativeModule, btnClick, mid, commonConf, formatArg)
			return nil
		})
	} else if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.CompatIosHover, nil) {
		reply.Bases.BottomButton = compatIosHoverButton()
	}
	var attentions []int64
	if mid > 0 && commonConf.IsSpaceAct() {
		eg.Go(func(ctx context.Context) error {
			attes, err := s.reldao.Attentions(ctx, mid)
			if err != nil {
				log.Error("Failed to get attentions: mid: %d, error: %+v", mid, err)
				return nil
			}
			attentions = make([]int64, 0, len(attes))
			for _, v := range attes {
				attentions = append(attentions, v.Mid)
			}
			return nil
		})
	}
	_ = eg.Wait()
	reply.Offset = pageConf.Offset
	reply.HasMore = pageConf.HasMore
	if len(modulesRly.Card) > 0 {
		//评论组件和动态无限feed流互斥
		var hasFeed bool
		for _, v := range needBase {
			if v.NativeModule == nil {
				continue
			}
			val, k := modulesRly.Card[v.NativeModule.ID]
			if !k || val == nil {
				continue
			}
			switch val.Goto {
			case actmdl.GotoReplyModule:
				if hasFeed { // 如果有互斥组件，需要丢弃
					continue
				}
				hasFeed = true
			case actmdl.GotoDynamicModule, actmdl.GotoEditorModule:
				if val.IsFeed == 1 {
					if hasFeed {
						continue
					}
					hasFeed = true
				}
			}
			reply.Items = append(reply.Items, val)
		}
	}
	if len(attentions) > 0 {
		reply.Attentions = &actmdl.Attentions{Uids: attentions}
	}
	mItems := make([]*actmdl.Item, 0)
	naviItem := make([]*actmdl.Item, 0)
	//导航组件与inlinetab组件是互斥的,且组件本身互斥
	//评论组件和动态无限feed流互斥,且组件本身互斥
	var hasMutex, hasFeed bool
	for _, v := range pageConf.Bases {
		if v.NativeModule == nil {
			continue
		}
		mVal, ok := modulesRly.Card[v.NativeModule.ID]
		if !ok || mVal == nil {
			continue
		}
		switch mVal.Goto {
		case actmdl.GotoNavigationModule, actmdl.GotoInlineTabModule:
			if hasMutex { // 如果有互斥组件，需要丢弃
				continue
			}
			hasMutex = true
		case actmdl.GotoReplyModule:
			if hasFeed { // 如果有互斥组件，需要丢弃
				continue
			}
			hasFeed = true
		case actmdl.GotoDynamicModule:
			if mVal.IsFeed == 1 { // 如果有互斥组件，需要丢弃
				if hasFeed {
					continue
				}
				hasFeed = true
			}
		}
		mItems = append(mItems, mVal)
		// 拼接导航组件title
		if mVal.Bar != "" {
			naviItem = append(naviItem, &actmdl.Item{ItemID: v.NativeModule.ID, Title: mVal.Bar})
		}
	}
	reply.Items = make([]*actmdl.Item, 0)
	// 拼接导航组件信息
	for _, val := range mItems {
		switch val.Goto {
		case actmdl.GotoNavigationModule:
			if len(naviItem) > 0 { // 导航组件且导航title有数据才下发
				for _, iVal := range val.Item {
					if iVal.Goto == actmdl.GotoNavigation {
						iVal.Item = naviItem
					}
				}
			}
			reply.Items = append(reply.Items, val)
		case actmdl.GotoInlineTabModule:
			reply.Items = append(reply.Items, val)
			// 低版本兼容，下级页面组件往一级页面上提
			if actmdl.IsInlineLow(c, s.c.Feature, a.MobiApp, a.Build) && len(val.ChildItem) > 0 {
				reply.Items = append(reply.Items, val.ChildItem...)
				val.ChildItem = nil
			}
		case actmdl.GotoSelectModule:
			reply.Items = append(reply.Items, val)
			// 低版本兼容，下级页面组件往一级页面上提
			if actmdl.IsSelectLow(c, s.c.Feature, a.MobiApp, a.Build) && len(val.ChildItem) > 0 {
				reply.Items = append(reply.Items, val.ChildItem...)
				val.ChildItem = nil
			}
		default:
			reply.Items = append(reply.Items, val)
		}
	}
	return reply, nil
}

// 白名单check
func (s *Service) checkWhite(c context.Context, firstPage *api.FirstPage, mid int64) error {
	if firstPage == nil || firstPage.Item == nil { //历史数据无父page信息,没有白名单逻辑，直接校验通过
		return nil
	}
	if firstPage.Item.IsAttrWhiteSwitch() != api.AttrModuleYes { //没有开通白名单逻辑
		return nil
	}
	if mid <= 0 || firstPage.Ext == nil { //未登录用户不支持访问 || 开通了白名单逻辑，但是数据源获取失败
		return xecode.AppPageOffline
	}
	sid, ok := strconv.ParseInt(firstPage.Ext.WhiteValue, 10, 64)
	if ok != nil { //配置错误，页面不下发
		return xecode.AppPageOffline
	}
	upList, err := s.actDao.UpList(c, sid, 1, 50, 0, api.SortTypeCtime)
	if err != nil || upList == nil {
		log.Error("s.actDao.UpList(%d) error(%v)", sid, err)
		return xecode.AppPageOffline
	}
	for _, v := range upList.List {
		if v == nil || v.Item == nil {
			continue
		}
		if v.Item.Wid == mid { //是白名单mid
			return nil
		}
	}
	return xecode.AppPageOffline
}

// ActIndex .
// nolint:gocognit
func (s *Service) ActIndex(c context.Context, a *actmdl.ParamActIndex, mid int64) (reply *actmdl.IndexReply, err error) {
	var (
		pageConf   *api.NatConfigReply
		topicCount *dynamic.TopicCount
		tag        *taggrpc.Tag
		attentions []int64
		pType      int32
	)
	if a == nil {
		return
	}
	// 版本判断 553之后的一次取41个组件
	if pType == api.CommonPage && ((a.MobiApp == "iphone" && a.Build > 9150) || (a.MobiApp == "android" && a.Build >= 5530000) || (a.MobiApp == "ipad" && a.Build > 12350)) {
		a.Ps = 41
	}
	if pageConf, err = s.actDao.NatConfig(c, &api.NatConfigReq{Pid: a.PageID, Offset: a.Offset, Ps: a.Ps, PType: pType}); err != nil || pageConf == nil || pageConf.NativePage == nil {
		if ecode.EqualError(natecode.NativePageOffline, err) {
			err = xecode.AppPageOffline
		}
		log.Error("s.actDao.NatConfig(%d) error(%v)", a.PageID, err)
		return
	}
	commonConf := pageConf.NativePage
	if !(commonConf.IsTopicAct() || commonConf.IsNewact()) || (commonConf.SkipURL != "" && pType == api.CommonPage) {
		err = xecode.AppPageOffline
		return
	}
	//白名单check
	if err = s.checkWhite(c, pageConf.FirstPage, mid); err != nil {
		return
	}
	_ = s.report.reportPageView(c, s.c.NaInfoc.PageViewLogID, &PageViewReport{
		Mid:      mid,
		PageID:   commonConf.ID,
		FromType: commonConf.FromType,
		Type:     commonConf.Type,
		MobiApp:  a.MobiApp,
		Build:    a.Build,
	})
	// 低版本兼容(ios粉，安卓粉，ipad粉，安卓国际,ios蓝版)
	// 只展示一个自定义点击组件
	// 分页数据也重新定义
	// 需要过滤天马落地页
	var partModule, headModule, hoverModule, bottomBtnModule *api.Module
	if pType == api.CommonPage {
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.CommonPage, &feature.OriginResutl{
			BuildLimit: (a.MobiApp == "iphone" && a.Build >= 8580 && a.Build <= 8910) || (a.MobiApp == "android" && a.Build >= 5430400 && a.Build <= 5483000) ||
				(a.MobiApp == "android_i" && a.Build >= 2040200 && a.Build <= 2040300) || (a.MobiApp == "iphone_b" && a.Build >= 8000 && a.Build <= 8080),
		}) {
			pageConf.Modules = defaultModule(pageConf.NativePage.ID)
			pageConf.Page = &api.Page{Offset: 2, HasMore: 0}
		}
	}
	if len(pageConf.Bases) > 0 {
		for _, vb := range pageConf.Bases {
			if vb.NativeModule == nil {
				continue
			}
			if vb.NativeModule.IsPart() && vb.Participation != nil && len(vb.Participation.List) > 0 {
				partModule = vb
			}
			if vb.NativeModule.IsBaseHead() {
				headModule = vb
			}
			if vb.NativeModule.IsBaseHoverButton() {
				hoverModule = vb
			}
			if vb.NativeModule.IsBaseBottomButton() {
				bottomBtnModule = vb
			}
		}
	}
	eg := errgroup.WithContext(c)
	formatArg := &actmdl.ParamFormatModule{PageID: a.PageID, Device: a.Device, VideoMeta: a.VideoMeta, Platform: a.Platform,
		Buvid: a.Buvid, Build: a.Build, ActivityFrom: a.ActivityFrom, DynamicID: a.DynamicID, MobiApp: a.MobiApp, TfIsp: a.TfIsp,
		HttpsUrlReq: a.HttpsUrlReq, FromSpmid: a.FromSpmid, CurrentTab: a.CurrentTab, Mid: mid, ShareOrigin: a.ShareOrigin,
		TabID: a.TabID, TabModuleID: a.TabModuleID, FromPage: actmdl.PageFromIndex, UserAgent: a.UserAgent, Memory: a.Memory}
	baseMs := &actmdl.Bases{}
	if partModule != nil {
		eg.Go(func(ctx context.Context) error { //参与组件：参与动态，参与投稿，参与专栏
			baseMs.Participation = s.FromPart(ctx, partModule.NativeModule, partModule.Participation, commonConf)
			return nil
		})
	}
	var relatedInfo *actmdl.UserInfo
	if headModule != nil && headModule.NativeModule != nil && headModule.NativeModule.IsAttrDisplayUser() == api.AttrModuleYes && commonConf.RelatedUid > 0 {
		eg.Go(func(ctx context.Context) error { //版头信息用户信息
			if infoRep, e := s.accDao.Info3GRPC(ctx, commonConf.RelatedUid); e != nil {
				log.Error("s.accDao.Info3GRPC mid(%d) error(%v)", commonConf.RelatedUid, e)
			} else if infoRep != nil {
				relatedInfo = &actmdl.UserInfo{Mid: commonConf.RelatedUid, Name: infoRep.Info.Name, Face: infoRep.Info.Face}
			}
			return nil
		})
	}
	if hoverModule != nil && hoverModule.NativeModule != nil {
		eg.Go(func(ctx context.Context) error {
			baseMs.HoverButton = s.FormatHoverButton(ctx, hoverModule.NativeModule, mid)
			return nil
		})
	}
	if bottomBtnModule != nil && bottomBtnModule.NativeModule != nil {
		btnClick := bottomBtnModule.Click
		eg.Go(func(ctx context.Context) error {
			if bottomBtnModule.NativeModule.Meta == "" {
				return nil
			}
			baseMs.BottomButton = s.FormatClick(ctx, bottomBtnModule.NativeModule, btnClick, mid, commonConf, formatArg)
			return nil
		})
	}
	if (a.ActivityFrom == dynamic.FromFeed || a.ActivityFrom == dynamic.FromDt) && a.DynamicID > 0 && actmdl.IsNewFeed(c, s.c.Feature, a.MobiApp, a.Build) {
		eg.Go(func(ctx context.Context) error { //天马动态卡
			baseMs.SingleDynamic = s.FromSingleDynTm(ctx, a, mid)
			if a.ActivityFrom == dynamic.FromDt && baseMs.SingleDynamic != nil {
				baseMs.SingleDynamic.Title = "活动推荐"
			}
			return nil
		})
	}
	if commonConf.IsTopicAct() && (headModule == nil || headModule.NativeModule == nil || headModule.NativeModule.IsAttrIsCloseViewNum() != api.AttrModuleYes) {
		eg.Go(func(ctx context.Context) (e error) {
			if topicCount, e = s.dynamicDao.ActiveUsers(ctx, commonConf.ForeignID, commonConf.IsAttrDisplayCounty()); e != nil {
				log.Error("s.dynamicDao.ActiveUsers(%d) error(%v)", commonConf.ForeignID, e)
				e = nil
			}
			return
		})
	}
	if mid > 0 {
		if commonConf.IsTopicAct() && (headModule == nil || headModule.NativeModule == nil || headModule.NativeModule.IsAttrIsCloseSubscribeBtn() != api.AttrModuleYes) {
			eg.Go(func(ctx context.Context) (e error) {
				if tag, e = s.tagDao.TagMsg(ctx, commonConf.ForeignID, mid); e != nil {
					log.Error("s.tagDao.Tag(%d,%d) error(%v)", mid, commonConf.ForeignID, e)
					e = nil
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) error {
			attes, e := s.reldao.Attentions(ctx, mid)
			if e != nil {
				log.Error("s.reldao.Attentions(%d) error(%v)", mid, e)
				return nil
			}
			attentions = make([]int64, 0, len(attes))
			for _, v := range attes {
				attentions = append(attentions, v.Mid)
			}
			return nil
		})
	}
	// 暂时关闭解决go-main引用问题
	//var upInfo *dyngrpc.ActPromoIconVisibleRsp
	//帮推icon暂时不下发，等产品通知
	//if mid > 0 && commonConf.IsTopicAct() && commonConf.ForeignID > 0 {
	//	eg.Go(func(ctx context.Context) error {
	//		var e error
	//		if upInfo, e = s.dynamicDao.ActPromoIconVisible(ctx, mid, commonConf.ForeignID); e != nil {
	//			log.Error("s.dynamicDao.ActPromoIconVisible(%d,%d) error(%v)", mid, commonConf.ForeignID, e)
	//		}
	//		return nil
	//	})
	//}
	//帮推icon暂时不下发，等产品通知
	// 拼接组件信息
	modulesRly := &actmdl.ModulesReply{}
	s.formatModule(c, pageConf.Modules, eg, commonConf, mid, formatArg, modulesRly, actmdl.FormatModFromIndex)
	_ = eg.Wait()
	if headModule != nil {
		baseMs.Head = &actmdl.Item{}
		if headModule.NativeModule != nil {
			if headModule.NativeModule.IsAttrDisplayUser() == api.AttrModuleYes {
				baseMs.Head.UserInfo = relatedInfo
				baseMs.Head.Content = headModule.NativeModule.Title
				baseMs.Head.OptionalImage = "http://i0.hdslb.com/bfs/activity-plat/static/20200616/82ac2611e49c304c91fb79cc76b9b762/eDsMwRL6Y.png"
				baseMs.Head.OptionalImage2 = "http://i0.hdslb.com/bfs/activity-plat/static/20200616/82ac2611e49c304c91fb79cc76b9b762/IGGf9Vleq.png"
				if relatedInfo != nil {
					baseMs.Head.HeadURI = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", relatedInfo.Mid)
				}
			}
			// fix-客户端bug，556以下安卓版本不下发背景色
			if a.MobiApp != "android" || (a.MobiApp == "android" && a.Build >= 5560000) {
				baseMs.Head.Color = &actmdl.Color{BgColor: headModule.NativeModule.BgColor}
			}
		}
		//预埋功能，本期不下发（暂时关闭解决go-main引用问题）
		//if upInfo != nil {
		//	baseMs.Head.URI = upInfo.PromoUrl             //自定义功能按钮 (默认不下发)
		//	baseMs.Head.Image = upInfo.Img                //自定义功能按钮 (默认不下发)
		//	baseMs.Head.ImageType = upInfo.ImgType        //1:json 2:image
		//	baseMs.Head.UnImage = upInfo.NightImg         //自定义功能按钮(夜间) (默认不下发)
		//	baseMs.Head.UnImageType = upInfo.NightImgType //1:json 2:image
		//}
		baseMs.Head.ShareImage = "" //分享 (默认不下发)
		baseMs.Head.ShareType = 0   //1:json 2:image
		//预埋功能，本期不下发
	}
	reply = &actmdl.IndexReply{
		PageID:      commonConf.ID,
		Title:       commonConf.Title,
		ForeignID:   commonConf.ForeignID,
		ForeignType: commonConf.Type,
		ShareTitle:  commonConf.ShareTitle,
		ShareImage:  commonConf.ShareImage,
		ShareType:   actmdl.ShareTypeActivity,
		VersionMsg:  "当前版本较低，无法显示完全，请更新至最新版本后查看",
		Bases:       baseMs,
		Color:       &actmdl.Color{BgColor: commonConf.BgColor},
		AttrBit:     &actmdl.AttrBit{NotNight: commonConf.IsAttrNotNightModule() == api.AttrModuleYes},
		IsUpSponsor: commonConf.IsUpTopicAct() && commonConf.RelatedUid != 0,
		FromType:    commonConf.FromType,
	}
	if pageConf.Page != nil {
		reply.Offset = pageConf.Page.Offset
		reply.HasMore = pageConf.Page.HasMore
	}
	if a.ShareOrigin == actmdl.OriginTab && a.TabID > 0 && a.TabModuleID > 0 {
		reply.PageURL = fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d&ts=%d", a.PageID, a.TabID, a.TabModuleID, time.Now().Unix())
	} else {
		reply.PageURL = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", commonConf.ID, time.Now().Unix()) //分享动态增加时间戳参数
	}
	if commonConf.ShareURL != "" {
		reply.ShareURL = commonConf.ShareURL
	} else {
		reply.ShareURL = reply.PageURL
	}
	// 默认取话题名
	reply.ShareCaption = commonConf.Title
	if commonConf.ShareCaption != "" {
		reply.ShareCaption = commonConf.ShareCaption
	}
	if len(attentions) > 0 {
		reply.Attentions = &actmdl.Attentions{Uids: attentions}
	}
	reply.DynamicInfo = &actmdl.DynamicInfo{
		DisplaySubscribeBtn: tag != nil,
		DisplayViewNum:      topicCount != nil,
	}
	if topicCount != nil && topicCount.DiscussCount != nil {
		reply.DynamicInfo.DiscussCount = topicCount.DiscussCount
	}
	if topicCount != nil && topicCount.ViewCount != nil {
		reply.DynamicInfo.ViewCount = topicCount.ViewCount
	}
	if tag != nil {
		reply.DynamicInfo.IsFollowed = tag.Attention == 1
	}
	if len(modulesRly.Card) == 0 {
		return
	}
	mItems := make([]*actmdl.Item, 0)
	naviItem := make([]*actmdl.Item, 0)
	//导航组件与inlinetab组件是互斥的,且组件本身互斥
	//评论组件和动态无限feed流,编辑推荐卡无限feed流互斥,且组件本身互斥
	var hasMutex, hasFeed bool
	for _, v := range pageConf.Modules {
		if v.NativeModule == nil {
			continue
		}
		mVal, ok := modulesRly.Card[v.NativeModule.ID]
		if !ok || mVal == nil {
			continue
		}
		switch mVal.Goto {
		case actmdl.GotoNavigationModule, actmdl.GotoInlineTabModule:
			if hasMutex { // 如果有互斥组件，需要丢弃
				continue
			}
			hasMutex = true
		case actmdl.GotoReplyModule:
			if hasFeed { // 如果有互斥组件，需要丢弃
				continue
			}
			hasFeed = true
		case actmdl.GotoDynamicModule, actmdl.GotoEditorModule:
			if mVal.IsFeed == 1 { // 如果有互斥组件，需要丢弃
				if hasFeed {
					continue
				}
				hasFeed = true
			}
		}
		mItems = append(mItems, mVal)
		// 拼接导航组件title
		if mVal.Bar != "" {
			naviItem = append(naviItem, &actmdl.Item{ItemID: v.NativeModule.ID, Title: mVal.Bar})
		}
	}
	reply.Items = make([]*actmdl.Item, 0)
	// 拼接导航组件信息
	for _, val := range mItems {
		switch val.Goto {
		case actmdl.GotoNavigationModule:
			if len(naviItem) > 0 { // 导航组件且导航title有数据才下发
				for _, iVal := range val.Item {
					if iVal.Goto == actmdl.GotoNavigation {
						iVal.Item = naviItem
					}
				}
			}
			reply.Items = append(reply.Items, val)
		case actmdl.GotoInlineTabModule:
			reply.Items = append(reply.Items, val)
			// 低版本兼容，下级页面组件往一级页面上提
			if actmdl.IsInlineLow(c, s.c.Feature, a.MobiApp, a.Build) && len(val.ChildItem) > 0 {
				reply.Items = append(reply.Items, val.ChildItem...)
				val.ChildItem = nil
			}
		case actmdl.GotoSelectModule:
			reply.Items = append(reply.Items, val)
			// 低版本兼容，下级页面组件往一级页面上提
			if actmdl.IsSelectLow(c, s.c.Feature, a.MobiApp, a.Build) && len(val.ChildItem) > 0 {
				reply.Items = append(reply.Items, val.ChildItem...)
				val.ChildItem = nil
			}
		default:
			reply.Items = append(reply.Items, val)
		}
	}
	// up主空间相关
	reply.UpSpace = &actmdl.UpSpace{
		SpacePageURL:     fmt.Sprintf("https://www.bilibili.com/blackboard/up-sponsor.html?act_from=topic_sync_space&act_id=%d", a.PageID),
		ExclusivePageURL: fmt.Sprintf("https://www.bilibili.com/blackboard/up-sponsor.html?act_from=topic_set_space&act_id=%d", a.PageID),
	}
	return
}

// formatModule .
// nolint:gocognit
func (s *Service) formatModule(c context.Context, modules []*api.Module, eg *errgroup.Group, commonConf *api.NativePage, mid int64, a *actmdl.ParamFormatModule, reply *actmdl.ModulesReply, opFrom string) {
	var mu sync.Mutex
	if len(modules) == 0 {
		return
	}
	reply.Card = make(map[int64]*actmdl.Item)
	for _, v := range modules {
		if v.NativeModule == nil {
			continue
		}
		temModule := v.NativeModule
		_ = s.report.reportModuleView(c, s.c.NaInfoc.ModuleViewLogID, &ModuleViewReport{
			ModuleID: temModule.ID,
			Category: temModule.Category,
			PageID:   commonConf.ID,
			FromType: commonConf.FromType,
		})
		switch {
		case temModule.IsReserve(): //预约组件
			tempRev := v.Reserve
			eg.Go(func(ctx context.Context) error {
				ck := s.formatReserve(ctx, temModule, tempRev, mid, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsGame(): //游戏组件
			if a.MobiApp != "android" && (a.MobiApp != "iphone" || a.Device == "pad") { //仅支持粉版
				continue
			}
			tempGame := v.Game
			eg.Go(func(ctx context.Context) error {
				ck := s.formatGame(ctx, temModule, tempGame, mid, a.MobiApp)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsReply(): //评论组件
			// 低版本过滤该组件
			if actmdl.IsVersion615Low(c, s.c.Feature, a.MobiApp, a.Build) {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatReply(temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsOgvSeasonSource(): //ogv剧集-资源类型
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatOgvSeasonResource(ctx, temModule, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsOgvSeasonID(): //ogv剧集-ids
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatOgvSeasonID(ctx, temModule, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsTimelineSource(): //时间轴组件-资源类型
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatTimelineResource(ctx, temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsTimelineIDs(): //时间轴组件-ids
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatTimelineIDs(ctx, temModule, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsSelect(): // 筛选组件
			selects := v.Select
			eg.Go(func(ctx context.Context) error {
				ck := s.FormatSelect(ctx, temModule, commonConf, selects, a, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsLive(): //直播卡组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatLive(ctx, temModule, a, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsInlineTab(): //inline tab组件
			inlineTabs := v.InlineTab
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatInlineTab(ctx, temModule, inlineTabs, a, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsNavigation(): // 导航组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatNavigation(ctx, temModule, a.MobiApp, a.Build)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsRecommend(), temModule.IsRcmdSource(): //推荐用户组件
			tempRecom := v.Recommend
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatRecommend(ctx, temModule, tempRecom, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsRcmdVertical(), temModule.IsRcmdVerticalSource(): //推荐用户-竖卡组件
			rcmdVertical := v.Recommend
			eg.Go(func(ctx context.Context) error {
				ck := s.FormatRcmdVertical(ctx, temModule, rcmdVertical, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsVote(): //预约组件
			temClick := v.Click
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatVote(ctx, temModule, temClick, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsClick(): //自定义点击组件
			temClick := v.Click
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatClick(ctx, temModule, temClick, mid, commonConf, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsAct(): //相关活动组件
			tempAct := v.Act
			// 相关活动列表为空，则不下发组件
			if tempAct == nil {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatActCard(ctx, temModule, tempAct)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsActCapsule():
			actPage := v.ActPage
			if actPage == nil {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				item := s.FormatActCapsule(ctx, temModule, actPage, a, opFrom)
				if item != nil {
					mu.Lock()
					reply.Card[temModule.ID] = item
					mu.Unlock()
				}
				return nil
			})

		case temModule.IsNewVideoAct(): //新视频卡-活动数据源组件
			sortType := int32(0)
			if v.VideoAct != nil && len(v.VideoAct.SortList) > 0 {
				sortType = int32(v.VideoAct.SortList[0].SortType)
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatNewVideoAct(ctx, temModule, sortType, mid, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsNewVideoID(): // 新视频卡-avid模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatNewVideoAvid(ctx, temModule, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsNewVideoDyn(): //新视频卡-动态模式
			tempDynamic := v.Dynamic
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FromNewVideoDynamic(ctx, temModule, commonConf, mid, tempDynamic, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsVideo(): //动态列表-活动数据源组件
			sortType := int32(0)
			if v.VideoAct != nil && len(v.VideoAct.SortList) > 0 {
				sortType = int32(v.VideoAct.SortList[0].SortType)
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatVideo(ctx, temModule, sortType, mid, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsVideoAct(): //视频卡-活动数据源组件
			sortType := int32(0)
			if v.VideoAct != nil && len(v.VideoAct.SortList) > 0 {
				sortType = int32(v.VideoAct.SortList[0].SortType)
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatVideoAct(ctx, temModule, sortType, mid, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsVideoDyn(): // 视频卡-动态模式
			tempDynamic := v.Dynamic
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatVideoDynamic(ctx, temModule, commonConf.ID, mid, tempDynamic, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsVideoAvid(): // 视频卡-avid模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatVideoAvid(ctx, temModule, a, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsDynamic(): //动态组件
			tempDynamic := v.Dynamic
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatDynamic(ctx, temModule, commonConf, mid, tempDynamic, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsBanner(): //标准banner组件
			// 高版本过滤banner组件,与板头组件相冲
			if (a.MobiApp == "iphone" && a.Build > 9270) || (a.MobiApp == "ipad" && a.Build > 12380) || (a.MobiApp == "android" && a.Build >= 5560000) {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatBanner(ctx, temModule, commonConf)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsStatement(): // 文本组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatStatement(temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsResourceDyn(): //资源小卡，动态组件
			tempDynamic := v.Dynamic
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatResourceDyn(ctx, temModule, tempDynamic, mid, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsResourceOrigin(): //资源小卡-外接数据源
			sortType := int64(0)
			if v.VideoAct != nil && len(v.VideoAct.SortList) > 0 {
				sortType = v.VideoAct.SortList[0].SortType
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatResourceOrigin(ctx, temModule, a, sortType)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsResourceAct(): //资源小卡，活动组件
			sortType := int32(0)
			if v.VideoAct != nil && len(v.VideoAct.SortList) > 0 {
				sortType = int32(v.VideoAct.SortList[0].SortType)
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatResourceAct(ctx, temModule, sortType, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsResourceID(): // 资源小卡-avid模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatResourceAvid(ctx, temModule, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsResourceRole(): // 资源小卡-角色剧集模式
			eg.Go(func(ctx context.Context) error {
				ck := s.FormatResourceRole(ctx, temModule, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsSingleDyn(): //单个动态组件
			// 高版本过滤banner组件
			if actmdl.IsNewFeed(c, s.c.Feature, a.MobiApp, a.Build) {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatSingleDyn(ctx, temModule, a, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsEditor(): //编辑推荐卡
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatEditor(ctx, temModule, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsEditorOrigin(): //编辑推荐卡-外接数据源，每周必看，入站必刷,排行榜
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatEditorOrigin(ctx, temModule, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsCarouselImg(): //轮播-图片模式
			tempCarousel := v.Carousel
			if tempCarousel != nil && len(tempCarousel.List) > 0 {
				eg.Go(func(ctx context.Context) (e error) {
					ck := s.FormatCarouselImg(ctx, temModule, tempCarousel)
					if ck != nil {
						mu.Lock()
						reply.Card[temModule.ID] = ck
						mu.Unlock()
					}
					return nil
				})
			}
		case temModule.IsCarouselWord(): //轮播-文字模式
			tempCarousel := v.Carousel
			if tempCarousel != nil && len(tempCarousel.List) > 0 {
				eg.Go(func(ctx context.Context) (e error) {
					ck := s.FormatCarouselWord(ctx, temModule, tempCarousel)
					if ck != nil {
						mu.Lock()
						reply.Card[temModule.ID] = ck
						mu.Unlock()
					}
					return nil
				})
			}
		case temModule.IsCarouselSource(): //轮播-数据源模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatCarouselSource(ctx, temModule, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsIcon(): //图标组件
			tempIcon := v.Icon
			if tempIcon != nil && len(tempIcon.List) > 0 {
				eg.Go(func(ctx context.Context) (e error) {
					ck := s.FormatIcon(ctx, temModule, tempIcon)
					if ck != nil {
						mu.Lock()
						reply.Card[temModule.ID] = ck
						mu.Unlock()
					}
					return nil
				})
			}
		case temModule.IsProgress(): //进度条
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatProgress(ctx, temModule, mid)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsNewactHeaderModule():
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatNewactHeader(ctx, temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsNewactAwardModule():
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatNewactAward(ctx, temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsNewactStatementModule():
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatNewactStatement(ctx, temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsMatchMedal():
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatMatchMedal(ctx, temModule)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case temModule.IsMatchEvent():
			event := v.MatchEvent
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatMatchEvent(ctx, temModule, event, a)
				if ck != nil {
					mu.Lock()
					reply.Card[temModule.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		}
	}
}

// FormatClick .
// nolint:gocognit
func (s *Service) FormatClick(c context.Context, mou *api.NativeModule, acts *api.Click, mid int64, page *api.NativePage, params *actmdl.ParamFormatModule) (res *actmdl.Item) {
	if mou.AvSort == api.NeedUnLock && mou.DySort == api.UnLockTime && mou.Stime > time.Now().Unix() { //解锁后展示 &&时间限制 && 未到达解锁时间，不下发组件
		return
	}
	var (
		fids        []int64
		fidsReply   map[int64]*relationgrpc.FollowingReply
		sids        []int64
		sidReply    map[int64]*actapi.ReserveFollowingReply
		seasonIDs   []int32
		seasonReply map[int32]*pgcClient.FollowStatusProto
		progReqs    = make(map[int64][]int64) //sid=>[]groupID
		clickItem   []*actmdl.Item
		actIDs      []int64
		actReply    map[int64]*actapi.ActRelationInfoReserveItems
		// 挂件
		pendantIDs    []int64
		pendantStates map[int64]int
		//抽奖次数
		lotteryIDs   []string
		lotteryTimes map[string]int64
		//任务统计
		taskPoint []*actmdl.ParamPlat
		taskNums  map[string]int64
		//会员购票务-想买
		buyIDs     []int64
		buyReply   map[int64]bool
		cartIDs    []int64
		cartRly    map[int64]*cartmdl.ComicItem
		appointIDs []int64
		// 评分
		scoreIDs     []int64
		scoreTargets map[int64]*scoregrpc.ScoreTarget
	)
	menuPageCompat(c, s.c.Feature, acts, params, page)
	if acts != nil && len(acts.Areas) > 0 {
		interface2url := make(map[string]string)
		for _, v := range acts.Areas {
			switch {
			case v.IsUpAppointment():
				appointIDs = append(appointIDs, v.ForeignID)
			case v.IsCartoon():
				cartIDs = append(cartIDs, v.ForeignID)
			case v.IsBuyCoupon():
				buyIDs = append(buyIDs, v.ForeignID)
			case v.IsActReserve():
				actIDs = append(actIDs, v.ForeignID)
			case v.IsReserve():
				sids = append(sids, v.ForeignID)
			case v.IsFollow():
				fids = append(fids, v.ForeignID)
			case v.IsCatchUp():
				seasonIDs = append(seasonIDs, int32(v.ForeignID))
			case v.IsPendant():
				pendantIDs = append(pendantIDs, v.ForeignID)
			case v.IsProgress():
				sid, gid := extractProgressParamFromClick(v)
				if sid == 0 || gid == 0 {
					continue
				}
				progReqs[sid] = append(progReqs[sid], gid)
			case v.IsStaticProgress(): //静态-进度条
				areaTip := new(api.ClickTip)
				if err := json.Unmarshal([]byte(v.Tip), areaTip); err != nil {
					log.Error("Fail to unmarshal click ext=%+v error=%+v", v.Ext, err)
					continue
				}
				if areaTip.PSort == api.ProcessUserStatics {
					sid, gid := extractProgressParamFromClick(v)
					if sid == 0 || gid == 0 {
						continue
					}
					progReqs[sid] = append(progReqs[sid], gid)
				} else if areaTip.PSort == api.ProcessRegister { //老预约数据源
					sids = append(sids, v.ForeignID)
				} else if areaTip.PSort == api.ProcessTaskStatics {
					taskPoint = append(taskPoint, &actmdl.ParamPlat{Activity: areaTip.Activity, Counter: areaTip.Counter, StatPc: areaTip.StatPc})
				} else if areaTip.PSort == api.ProcessLottery { //抽奖数据源
					lotteryIDs = append(lotteryIDs, areaTip.LotteryID)
				} else if areaTip.PSort == api.ProcessScore {
					scoreIDs = append(scoreIDs, v.ForeignID)
				}
			case v.IsInterface():
				style, err := extractExt4ClickInterface(v.Ext)
				if err != nil || style == "" {
					continue
				}
				interface2url[style] = ""
			case v.IsLayerInterface():
				//低版本首页不下发接口类型浮层按钮
				if menuLayerInterface(c, s.c.Feature, params.MobiApp, params.Build, page) {
					interface2url[api.ClickStyleBnjTaskGame] = ""
				}
			}
			if v.IsCustom() {
				setUnlockProgReq(v, progReqs)
			}
		}
		eg := errgroup.WithContext(c)
		if len(taskPoint) > 0 {
			taskNums = make(map[string]int64)
			var taskMu sync.Mutex
			for _, v := range taskPoint {
				actStr := v.Activity
				counter := v.Counter
				statPc := v.StatPc
				if statPc == "daily" {
					eg.Go(func(ctx context.Context) error {
						num, e := s.platDao.GetCounterRes(ctx, counter, actStr, mid)
						if e != nil {
							log.Error("s.platDao.GetCounterRes(%s,%s,%d) error(%v)", counter, actStr, mid, e)
							//降级错误不抛出
							return nil
						}
						taskMu.Lock()
						taskNums[fmt.Sprintf("%s-%s-%s", counter, actStr, statPc)] = num
						taskMu.Unlock()
						return nil
					})
				} else {
					eg.Go(func(ctx context.Context) error {
						num, e := s.platDao.GetTotalRes(ctx, counter, actStr, mid)
						if e != nil {
							log.Error("s.platDao.GetTotalRes(%s,%s,%d) error(%v)", counter, actStr, mid, e)
							//降级错误不抛出
							return nil
						}
						taskMu.Lock()
						taskNums[fmt.Sprintf("%s-%s-%s", counter, actStr, statPc)] = num
						taskMu.Unlock()
						return nil
					})
				}
			}
		}
		var appointRly map[int64]*actapi.UpActReserveRelationInfo
		if len(appointIDs) > 0 {
			eg.Go(func(ctx context.Context) error {
				appointRly, _ = s.actDao.UpActReserveRelationInfo(ctx, mid, appointIDs)
				return nil
			})
		}
		if len(buyIDs) > 0 && mid > 0 {
			eg.Go(func(ctx context.Context) error {
				buyReply = s.shopDao.BatchMultiFavStat(ctx, buyIDs, mid)
				return nil
			})
		}
		if len(cartIDs) > 0 && mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if cartRly, e = s.cartdao.GetComicInfos(ctx, cartIDs, mid, ""); e != nil {
					log.Error("s.cartdao.GetComicInfos(%v,%d) error(%v)", cartIDs, mid, e)
					e = nil
				}
				return
			})
		}
		if len(lotteryIDs) > 0 && mid > 0 {
			lotteryTimes = make(map[string]int64)
			var lottMu sync.Mutex
			for _, v := range lotteryIDs {
				id := v
				eg.Go(func(ctx context.Context) error {
					reaRly, e := s.actDao.LotteryUnusedTimes(ctx, mid, id)
					if e != nil {
						log.Error("s.actDao.LotteryUnusedTimes(%d,%s) error(%v)", mid, id, e)
						//降级错误不抛出
						return nil
					}
					if reaRly != nil {
						lottMu.Lock()
						lotteryTimes[id] = reaRly.Times
						lottMu.Unlock()
					}
					return nil
				})
			}
		}
		if len(actIDs) > 0 && mid > 0 {
			actReply = make(map[int64]*actapi.ActRelationInfoReserveItems)
			var actMu sync.Mutex
			for _, v := range actIDs {
				id := v
				eg.Go(func(ctx context.Context) error {
					reaRly, e := s.actDao.ActRelationInfo(ctx, id, mid)
					if e != nil {
						log.Error("s.actDao.ActRelationInfo(%d,%d) error(%v)", mid, id, e)
						//降级错误不抛出
						return nil
					}
					if reaRly != nil && reaRly.ReserveItems != nil {
						actMu.Lock()
						actReply[id] = reaRly.ReserveItems
						actMu.Unlock()
					}
					return nil
				})
			}
		}
		if len(fids) > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if fidsReply, e = s.reldao.RelationsGRPC(ctx, mid, fids); e != nil {
					log.Error("s.reldao.RelationsGRPC(%d,%v) error(%v)", mid, fids, e)
					e = nil
				}
				return
			})
		}
		if len(seasonIDs) > 0 && mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if seasonReply, e = s.pgcdao.StatusByMid(ctx, mid, seasonIDs); e != nil {
					log.Error("s.pgcdao.StatusByMid(%d,%v) error(%v)", mid, seasonIDs, e)
					e = nil
				}
				return
			})
		}
		if len(sids) > 0 && mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if sidReply, e = s.actDao.ReserveFollowings(ctx, mid, sids); e != nil {
					log.Error("s.actDao.ReserveFollowings(%d,%v) error(%v)", mid, sids, e)
					e = nil
				}
				return
			})
		}
		if len(pendantIDs) > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if pendantStates, e = s.actDao.AwardSubjectStates(ctx, pendantIDs, mid); e != nil {
					log.Error("s.actDao.AwardSubjectStates(%d,%v) error(%v)", mid, pendantIDs, e)
					e = nil
				}
				return
			})
		}
		progRlys := make(map[int64]*actapi.ActivityProgressReply, len(progReqs))
		if len(progReqs) > 0 {
			lock := sync.Mutex{}
			for k, v := range progReqs {
				gids := v
				sid := k
				eg.Go(func(ctx context.Context) error {
					progress, err := s.actDao.ActivityProgress(ctx, sid, 2, mid, gids)
					if err != nil {
						return nil
					}
					lock.Lock()
					progRlys[sid] = progress
					lock.Unlock()
					return nil
				})
			}
		}
		if len(interface2url) > 0 {
			var mu sync.Mutex
			for v := range interface2url {
				inter := v
				eg.Go(func(ctx context.Context) error {
					switch inter {
					case api.ClickStyleBnj, api.ClickStyleBnjTaskGame:
						bnjBizType := actapi.AppJumpBizType_Type4Bnj2021AR
						if inter == api.ClickStyleBnjTaskGame {
							bnjBizType = actapi.AppJumpBizType_Type4Bnj2021TaskGame
						}
						jumpUrl, err := s.actDao.AppJumpUrl(ctx, bnjBizType, params.Memory, params.UserAgent)
						if err != nil || jumpUrl == "" {
							return nil
						}
						mu.Lock()
						interface2url[inter] = jumpUrl
						mu.Unlock()
					}
					return nil
				})
			}
		}
		if len(scoreIDs) > 0 {
			eg.Go(func(ctx context.Context) error {
				req := &scoregrpc.MultiGetTargetScoreReq{TntCode: 1, STargetType: 1, STargetIds: scoreIDs}
				if rly, err := s.scoreDao.MultiGetTargetScore(ctx, req); err == nil {
					scoreTargets = rly.GetTargets()
				}
				return nil
			})
		}
		_ = eg.Wait()
		for _, v := range acts.Areas {
			var (
				ext  *actmdl.ClickExt
				dTmp *actmdl.Item
			)
			switch {
			case v.IsCartoon():
				tipObj := &actmdl.TipCancel{}
				tipObj.FromTip(v.Tip)
				ext = &actmdl.ClickExt{FID: v.ForeignID, Goto: actmdl.GotoClickCartoon, Tip: tipObj}
				// 0:未收藏 1:已收藏
				if fval, ok := cartRly[v.ForeignID]; ok && fval.FavStatus == 1 {
					ext.IsFollow = true
				}
			case v.IsBuyCoupon():
				tipObj := &actmdl.TipCancel{}
				tipObj.FromTip(v.Tip)
				ext = &actmdl.ClickExt{FID: v.ForeignID, Goto: actmdl.GotoClickBuy, Tip: tipObj}
				// 0:未收藏 1:已收藏
				if fval, ok := buyReply[v.ForeignID]; ok && fval {
					ext.IsFollow = true
				}
			case v.IsActReserve():
				tipObj := &actmdl.TipCancel{}
				tipObj.FromTip(v.Tip)
				ext = &actmdl.ClickExt{FID: v.ForeignID, Goto: actmdl.GotoClickAttention, Tip: tipObj}
				// 0:未预约 1:已预约
				if fval, ok := actReply[v.ForeignID]; ok && fval != nil && fval.State == 1 {
					ext.IsFollow = true
				}
			case v.IsReserve():
				tipObj := &actmdl.TipCancel{}
				tipObj.FromTip(v.Tip)
				func() {
					if s.actDao.ClickSpecialTip == nil {
						return
					}
					if _, ok := s.actDao.ClickSpecialTip.Sid[strconv.FormatInt(v.ForeignID, 10)]; !ok {
						return
					}
					tipObj.Msg = s.actDao.ClickSpecialTip.Msg
					tipObj.SureMsg = s.actDao.ClickSpecialTip.SureMsg
					tipObj.ThinkMsg = s.actDao.ClickSpecialTip.ThinkMsg
					tipObj.CancelMsg = s.actDao.ClickSpecialTip.CancelMsg
				}()
				ext = &actmdl.ClickExt{FID: v.ForeignID, Goto: actmdl.GotoClickAppointment, Tip: tipObj}
				if sval, ok := sidReply[v.ForeignID]; ok && sval != nil && sval.IsFollow {
					ext.IsFollow = true
				}
			case v.IsFollow():
				tipObj := &actmdl.TipCancel{}
				tipObj.FromTip("关注")
				ext = &actmdl.ClickExt{FID: v.ForeignID, Goto: actmdl.GotoClickFollow, Tip: tipObj}
				// 1- 悄悄关注 2 关注  6-好友 128-拉黑
				if fval, ok := fidsReply[v.ForeignID]; ok && fval != nil && (fval.Attribute == 2 || fval.Attribute == 6) {
					ext.IsFollow = true
				}
			case v.IsCatchUp():
				tipObj := &actmdl.TipCancel{}
				tipObj.FromTip(v.Tip)
				ext = &actmdl.ClickExt{FID: v.ForeignID, Goto: actmdl.GotoClickPgc, Tip: tipObj}
				if seval, ok := seasonReply[int32(v.ForeignID)]; ok && seval != nil && seval.Follow {
					ext.IsFollow = true
				}
			case v.IsUpAppointment():
				if aVal, ok := appointRly[v.ForeignID]; !ok || aVal == nil { //没有返回值
					continue
				}
				//默认是不可点击状态
				var currentState int8
				if appointRly[v.ForeignID].UpActVisible == actapi.UpActVisible_DefaultVisible && (appointRly[v.ForeignID].State == actapi.UpActReserveRelationState_UpReserveRelated || appointRly[v.ForeignID].State == actapi.UpActReserveRelationState_UpReserveRelatedOnline) {
					if appointRly[v.ForeignID].IsFollow == 1 {
						currentState = 2
					} else {
						currentState = 1
					}
				}
				ext = &actmdl.ClickExt{}
				ext.FromClickReceive(v, currentState)
			case v.IsPendant():
				ext = &actmdl.ClickExt{}
				ext.FromClickReceive(v, int8(pendantStates[v.ForeignID]))
			case v.IsProgress():
				progRly, ok := progRlys[v.ForeignID]
				if !ok || len(progRly.Groups) == 0 {
					continue
				}
				areaTip := new(api.ClickTip)
				if err := json.Unmarshal([]byte(v.Tip), areaTip); err != nil {
					continue
				}
				group, ok := progRly.Groups[areaTip.GroupId]
				if !ok {
					continue
				}
				num, targetNum := extractProgNum(group, areaTip.NodeId)
				ext = &actmdl.ClickExt{Num: num, TargetNum: targetNum}
			case v.IsRedirect():
				ext = &actmdl.ClickExt{Goto: actmdl.GotoClickRedirect}
			case v.IsStaticProgress(): //静态-进度条
				areaTip := new(api.ClickTip)
				if err := json.Unmarshal([]byte(v.Tip), areaTip); err != nil {
					continue
				}
				if areaTip.PSort == api.ProcessUserStatics {
					progRly, ok := progRlys[v.ForeignID]
					if !ok || len(progRly.Groups) == 0 {
						continue
					}
					group, ok := progRly.Groups[areaTip.GroupId]
					if !ok {
						continue
					}
					num, targetNum := extractProgNum(group, areaTip.NodeId)
					ext = &actmdl.ClickExt{Num: num, TargetNum: targetNum}
				} else if areaTip.PSort == api.ProcessRegister { //老预约数据源
					ext = &actmdl.ClickExt{}
					if sval, ok := sidReply[v.ForeignID]; ok && sval != nil {
						dimension, _ := extractDimension(v)
						if dimension == actapi.GetReserveProgressDimension_Rule { //整体活动维度
							ext.Num = calculateProgress(sval.Total, areaTip.InterveNum)
						} else if sval.IsFollow { //用户维度&& 用户预约了
							ext.Num = 1
						}
					}
				} else if areaTip.PSort == api.ProcessTaskStatics {
					ext = &actmdl.ClickExt{}
					if tv, ok := taskNums[fmt.Sprintf("%s-%s-%s", areaTip.Counter, areaTip.Activity, areaTip.StatPc)]; ok {
						ext.Num = tv
					}
				} else if areaTip.PSort == api.ProcessLottery { //抽奖数据源
					ext = &actmdl.ClickExt{}
					if lv, ok := lotteryTimes[areaTip.LotteryID]; ok {
						ext.Num = lv
					}
				} else if areaTip.PSort == api.ProcessScore {
					st, ok := scoreTargets[v.ForeignID]
					if !ok {
						continue
					}
					ext = &actmdl.ClickExt{}
					ext.DisplayNum = finalScore(st)
				}
			case v.IsInterface():
				style, err := extractExt4ClickInterface(v.Ext)
				if err != nil || style == "" {
					continue
				}
				jumpUrl, ok := interface2url[style]
				if !ok || jumpUrl == "" {
					continue
				}
				v.Link = jumpUrl
			case v.IsLayerInterface():
				jumpUrl, ok := interface2url[api.ClickStyleBnjTaskGame]
				if !ok || jumpUrl == "" {
					continue
				}
				v.Link = jumpUrl
			}
			if v.IsCustom() && !reachUnlockCondition(v, progRlys) {
				continue
			}
			dTmp = &actmdl.Item{}
			dTmp.FromArea(c, s.c.Feature, v, ext, mou, params.MobiApp, params.Build)
			clickItem = append(clickItem, dTmp)
		}
	}
	res = &actmdl.Item{}
	res.FromClick(mou, clickItem)

	return
}

// FormatInlineTab .
// nolint:gocognit
func (s *Service) FormatInlineTab(c context.Context, mou *api.NativeModule, inline *api.InlineTab, arg *actmdl.ParamFormatModule, mid int64) *actmdl.Item {
	if inline == nil || len(inline.List) == 0 {
		return nil
	}
	var (
		pageIDs           []int64
		defTab, timingTab int64
		nowTime           = time.Now().Unix()
	)
	ext := make(map[int64]*api.MixReason)
	for _, v := range inline.List {
		if v == nil || v.MType != api.MixInlineType || v.ForeignID == 0 || !v.IsOnline() {
			continue
		}
		pageIDs = append(pageIDs, v.ForeignID)
		ext[v.ForeignID] = v.RemarkUnmarshal()
		//寻找默认tab
		if ext[v.ForeignID].DefType == api.DefTypeTimely { //立即生效的时间
			defTab = v.ForeignID
		} else if ext[v.ForeignID].DefType == api.DefTypeTiming { //定时生效的时间
			if ext[v.ForeignID].DStime <= nowTime && ext[v.ForeignID].DEtime > nowTime {
				timingTab = v.ForeignID
			}
		}
		//寻找默认tab
	}
	// 默认tab优先级,若立即生效的时间，与定时生效的时间一致，则优先以定时生效的为准
	if timingTab == 0 {
		timingTab = defTab
	}
	//寻找默认tab
	if len(pageIDs) == 0 {
		return nil
	}
	pagesInfo, e := s.actDao.NativePages(c, pageIDs)
	if e != nil {
		log.Error("s.actDao.NativePages %v error(%v)", pageIDs, e)
		return nil
	}
	tmpItem := &actmdl.Item{}
	tmpItem.FormatInline(c, s.c.Feature, mou, arg.MobiApp, arg.Build)
	var (
		currentIndex int32
		hasFind      bool
	)
	for _, v := range pageIDs {
		if val, ok := pagesInfo[v]; !ok || val == nil || !val.IsOnline() || val.Title == "" {
			continue
		}
		var (
			hasLock bool
		)
		eVal := ext[v]
		tmpID := &actmdl.Item{ItemID: v, Title: pagesInfo[v].Title}
		lockExt := pagesInfo[v].ConfSetUnmarshal()
		if lockExt.DT == api.NeedUnLock { //解锁模式
			var deblocking bool
			if lockExt.DC == api.UnLockTime && lockExt.Stime <= time.Now().Unix() { //时间模式&&到达解锁时间
				deblocking = true
			}
			//未解锁时
			if !deblocking {
				//高版本 && 不可点击
				if !actmdl.IsVersion615Low(c, s.c.Feature, arg.MobiApp, arg.Build) && lockExt.UnLock == api.NotClick {
					tmpID.ItemID = 0 //不下发pageid
					tmpID.Setting = &actmdl.Setting{UnAllowClick: true}
					tmpID.Content = "还未解锁，敬请期待"
					if lockExt.Tip != "" {
						tmpID.Content = lockExt.Tip //提示文案
					}
					hasLock = true //锁定
				} else { //不认识类型 || 未解锁：不展示 || 不可点击下,低版本
					continue
				}
			}
		}
		//组件是图片模式
		if mou.AvSort == 1 && eVal != nil {
			//锁定状态时图片展示未解锁态
			if hasLock {
				tmpID.ImagesUnion = &actmdl.ImagesUnion{
					UnSelect: actmdl.ImageChange(eVal.UnI), //未选中
				}
			} else {
				tmpID.ImagesUnion = &actmdl.ImagesUnion{
					Select:   actmdl.ImageChange(eVal.SI),   //选中
					UnSelect: actmdl.ImageChange(eVal.UnSI), //未选中
				}
			}
		}
		tmpItem.Item = append(tmpItem.Item, tmpID)
		//查找默认tab start
		//当【页面URL含有定位参数】与【页面设置默认tab】同时存在时，则优先以页面URL的定位参数为准
		if arg.CurrentTab != "" {
			if eVal != nil && eVal.JoinCurrentTab() == arg.CurrentTab && !hasLock {
				tmpItem.CurrentTabIndex = currentIndex
				hasFind = true
			}
		}
		// 没有指定定位 &&  没有锁定 && 有默认tab
		if !hasFind && !hasLock && v == timingTab {
			tmpItem.CurrentTabIndex = currentIndex
		}
		//查找默认tab end
		currentIndex++
	}
	if len(tmpItem.Item) == 0 {
		return nil
	}
	// 低版本兼容inline组件，取出第一个tab页面下对应的三个组件信息，拼接到一级页面上
	var child []*actmdl.Item
	if actmdl.IsInlineLow(c, s.c.Feature, arg.MobiApp, arg.Build) {
		func() {
			inlineReq := &actmdl.ParamInlineTab{
				PageID:      tmpItem.Item[0].ItemID,
				Device:      arg.Device,
				VideoMeta:   arg.VideoMeta,
				MobiApp:     arg.MobiApp,
				Platform:    arg.Platform,
				Build:       arg.Build,
				Buvid:       arg.Buvid,
				Offset:      0,
				Ps:          3,
				TfIsp:       arg.TfIsp,
				HttpsUrlReq: arg.HttpsUrlReq,
				FromSpmid:   arg.FromSpmid,
			}
			inlineRly, e := s.InlineTab(c, inlineReq, mid)
			if e != nil { //低版本兼容逻辑，错误不处理
				log.Error("s.InlineTa %d error(%v)", inlineReq.PageID, e)
				return
			}
			if inlineRly != nil {
				child = inlineRly.Items
			}
		}()
	}
	var items []*actmdl.Item
	items = append(items, tmpItem)
	// 低版本兼容inline组件
	first := &actmdl.Item{}
	first.FromInlineTabModule(mou, items, child)
	return first
}

// FormatActCard .
func (s *Service) FormatActCard(c context.Context, mou *api.NativeModule, acts *api.Act) (first *actmdl.Item) {
	if acts == nil || len(acts.List) == 0 {
		return
	}
	first = &actmdl.Item{}
	first.FromActModule(mou)
	first.Item = make([]*actmdl.Item, 0)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		first.Item = append(first.Item, tmpImage)
	}
	for _, v := range acts.List {
		tmpAct := &actmdl.Item{}
		tmpAct.FromActs(v)
		first.Item = append(first.Item, tmpAct)
	}
	return
}

func (s *Service) FormatActCapsule(c context.Context, mou *api.NativeModule, actPage *api.ActPage, params *actmdl.ParamFormatModule, opFrom string) *actmdl.Item {
	if opFrom != actmdl.FormatModFromMenuUp {
		actmdl.DelCurrentActPage(actPage, mou.NativeID)
	}
	cards := s.natCardsOfActPage(c, actPage, params)
	if len(cards) == 0 {
		return nil
	}
	capsule := &actmdl.Item{
		Goto:  actmdl.GotoActCapsule,
		Title: mou.Caption,
		Item:  make([]*actmdl.Item, 0, len(actPage.List)),
	}
	for _, v := range actPage.List {
		card, ok := cards[v.PageID]
		if !ok {
			continue
		}
		item := &actmdl.Item{}
		item.FromActCapsuleItem(card)
		capsule.Item = append(capsule.Item, item)
	}
	if len(capsule.Item) == 0 {
		return nil
	}
	capsuleMod := &actmdl.Item{}
	capsuleMod.FromActCapsuleModule(mou, []*actmdl.Item{capsule})
	return capsuleMod
}

func (s *Service) natCardsOfActPage(c context.Context, actPage *api.ActPage, params *actmdl.ParamFormatModule) map[int64]*api.NativePageCard {
	if actPage == nil || len(actPage.List) == 0 {
		return map[int64]*api.NativePageCard{}
	}
	// NativePageCards返回上线活动，跳转优先级为：配置跳转链接 > 活动聚合页 > 单个活动页
	pids := make([]int64, 0, len(actPage.List))
	for _, v := range actPage.List {
		if v == nil {
			continue
		}
		pids = append(pids, v.PageID)
	}
	cards, _ := s.actDao.NativePageCards(c, &api.NativePageCardsReq{
		Pids:     pids,
		Device:   params.Device,
		MobiApp:  params.MobiApp,
		Build:    int32(params.Build),
		Buvid:    params.Buvid,
		Platform: params.Platform,
	})
	// NativeAllPages返回剩余的活动（下线/NativePageCards失败），跳转优先级为 新频道页 > 旧频道普通话题页
	restPIDs := make([]int64, 0, len(actPage.List))
	for _, pid := range pids {
		if _, ok := cards[pid]; ok {
			continue
		}
		restPIDs = append(restPIDs, pid)
	}
	pages, _ := s.actDao.NativeAllPages(c, restPIDs)
	// 获取频道数据
	chanIDs := make([]int64, 0, len(pages))
	for _, v := range pages {
		if v == nil {
			continue
		}
		chanIDs = append(chanIDs, v.ForeignID)
	}
	chanInfos, _ := s.channelDao.Infos(c, chanIDs, 0)
	// 组装数据
	res := make(map[int64]*api.NativePageCard, len(pids))
	for _, pid := range pids {
		if card, ok := cards[pid]; ok {
			res[pid] = card
			continue
		}
		page, ok := pages[pid]
		if !ok || page == nil {
			continue
		}
		chanInfo, ok := chanInfos[page.ForeignID]
		if !ok || chanInfo == nil {
			continue
		}
		switch chanInfo.GetCType() {
		case showmdl.OldChannel:
			page.SkipURL = showmdl.FillURI(showmdl.GotoChannelTopic, strconv.FormatInt(page.ForeignID, 10), nil)
		case showmdl.NewChannel:
			page.SkipURL = showmdl.FillURI(showmdl.GotoChannelNewTopic, strconv.FormatInt(page.ForeignID, 10), nil)
		default:
			continue
		}
		res[pid] = &api.NativePageCard{
			Id:           page.ID,
			Title:        page.Title,
			Type:         page.Type,
			ForeignID:    page.ForeignID,
			ShareTitle:   page.ShareTitle,
			ShareImage:   page.ShareImage,
			ShareURL:     page.ShareURL,
			SkipURL:      page.SkipURL,
			RelatedUid:   page.RelatedUid,
			PcURL:        page.PcURL,
			ShareCaption: page.ShareCaption,
		}
	}
	return res
}

// FormatVideoDynamic .
func (s *Service) FormatVideoDynamic(c context.Context, mou *api.NativeModule, pageID, mid int64, dyn *api.Dynamic, pas *actmdl.ParamFormatModule) (dynamicReply *actmdl.Item) {
	var (
		types = strconv.Itoa(dynamic.VideoType)
		reply *dynamic.DyReply
		err   error
		psNum = mou.Num
		list  *actmdl.Item
	)
	// dysort 默认0
	if reply, err = s.dynamicDao.FetchDynamics(c, mou.Fid, mid, psNum, 0, pas.Device, types, pas.Platform, "", "", pas.FromSpmid, pas.TabFrom); err != nil || reply == nil {
		log.Error("s.dynamicDao.FetchDynamics(%d) error(%v)", mou.Fid, err)
		return
	}
	if len(reply.Cards) == 0 {
		return
	}
	list = &actmdl.Item{}
	list.FromVideoDynModule(mou, mou.IsCardSingle())
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	for _, v := range reply.Cards {
		tmpAct := &actmdl.Item{}
		tmpAct.FromVideoCard(v, mou.IsCardSingle())
		list.Item = append(list.Item, tmpAct)
	}
	if reply.HasMore > 0 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &actmdl.Item{}
		tmpMore.FromVideoMore(mou, 0, pageID, reply.Offset, true)
		list.Item = append(list.Item, tmpMore)
	}
	dynamicReply = list
	return
}

// newVideoDynamic .
func (s *Service) newVideoDynamic(c context.Context, arg *actmdl.NewDynReq) (*actmdl.NewVideoReply, error) {
	briRly, err := s.dynamicDao.BriefDynamics(c, arg.TopicID, arg.PageSize, arg.Mid, arg.Types, arg.DyOffset, 0)
	if err != nil {
		log.Error("s.dynamicDao.BriefDynamics(%d) error(%v)", arg.TopicID, err)
		return nil, err
	}
	if briRly == nil || len(briRly.Dynamics) == 0 {
		return nil, nil
	}
	var aids []*arccli.PlayAv
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		switch v.Type {
		case dynamic.VideoType:
			//只有aid
			aids = append(aids, &arccli.PlayAv{Aid: v.Rid})
		default:
			continue
		}
	}
	var (
		arcRly map[int64]*arccli.ArcPlayer
	)
	if len(aids) > 0 {
		if arcRly, err = s.arcdao.ArcsPlayer(c, aids); err != nil {
			log.Error("s.arcdao.ArcsPlayer aids(%v) error(%v)", aids, err)
			return nil, err
		}
	}
	rly := &actmdl.NewVideoReply{HasMore: int32(briRly.HasMore), DyOffset: briRly.Offset}
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		tmp := &actmdl.Item{}
		switch v.Type {
		case dynamic.VideoType:
			va, ok := arcRly[v.Rid]
			if !ok || va == nil || va.Arc == nil || !va.Arc.IsNormal() {
				continue
			}
			// 获取首p秒开地址即可
			firstPlay := va.PlayerInfo[va.DefaultPlayerCid]
			tmp.FromNewVideoCard(va.Arc, firstPlay, arg.Build, arg.MobiApp)
		default:
			continue
		}
		rly.Item = append(rly.Item, tmp)
	}
	return rly, nil
}

// FromNewVideoDynamic .
func (s *Service) FromNewVideoDynamic(c context.Context, mou *api.NativeModule, topicInfo *api.NativePage, mid int64, dyn *api.Dynamic, pas *actmdl.ParamFormatModule) *actmdl.Item {
	var (
		reply *actmdl.NewVideoReply
		err   error
	)
	netType, tfType := showmdl.TrafficFree(pas.TfIsp)
	arg := &actmdl.NewDynReq{TopicID: mou.Fid, Types: strconv.Itoa(dynamic.VideoType), PageSize: mou.Num, Mid: mid, MobiApp: pas.MobiApp, Buvid: pas.Buvid, Build: pas.Build, Platform: pas.Platform, NetType: netType, TfType: tfType}
	if reply, err = s.newVideoDynamic(c, arg); err != nil {
		log.Error("s.dynamicDao.FetchDynamics(%d) error(%v)", topicInfo.ForeignID, err)
		return nil
	}
	if reply == nil || len(reply.Item) == 0 {
		return nil
	}
	return s.videoJoin(reply, mou)
}

// isNeedFix 修护客户端二级列表页面inline播放bug ios build 9120,9150 ,9160 ,9170 ,9180
func isNeedFix(mobiApp, device string, build int64) bool {
	return mobiApp == "iphone" && device == "phone" && (build == 9120 || build == 9150 || build == 9160 || build == 9170 || build == 9180)
}

// fromFeedDynamic .
func (s *Service) fromFeedDynamic(c context.Context, mou *api.NativeModule, foreignID int64, types, tabFrom string) *actmdl.Item {
	rly, e := s.dynamicDao.HasFeed(c, foreignID, int64(mou.DySort), types)
	if e != nil {
		log.Error("s.dynamicDao.HasFeed(%d) error(%v)", foreignID, e)
		return nil
	}
	if rly != 1 {
		return nil
	}
	list := &actmdl.Item{}
	ext := &actmdl.UrlExt{TopicID: foreignID, Types: types, Sortby: mou.DySort, RemoteFrom: actmdl.RemoteActivity, ScenaryFrom: tabFrom}
	list.FromDynamicModule(mou, ext)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	return list
}

// FromDynamic .
func (s *Service) FormatDynamic(c context.Context, mou *api.NativeModule, topicInfo *api.NativePage, mid int64, dyn *api.Dynamic, pas *actmdl.ParamFormatModule) (dynamicReply *actmdl.Item) {
	var (
		types      string
		reply      *dynamic.DyReply
		err        error
		psNum      = mou.Num
		list       *actmdl.Item
		tys        []string
		tmpNum     int64
		ok         bool
		foreignID  = topicInfo.ForeignID
		topicTitle = topicInfo.Title
	)
	if mou.Fid > 0 {
		foreignID = mou.Fid
		// 不需要获取关注状态
		tagRly, _ := s.tagDao.TagMsg(c, foreignID, 0)
		if tagRly != nil {
			topicTitle = tagRly.Name
		}
	}
	if dyn != nil && len(dyn.SelectList) > 0 {
		for _, val := range dyn.SelectList {
			// 精选或者全选时，是不支持多选的
			if tempType, isSingle := val.JoinMultiDyTypes(); isSingle {
				types = tempType
				tys = []string{}
				break
			} else {
				tys = append(tys, tempType)
			}
		}
		if len(tys) > 0 {
			types = strings.Join(tys, ",")
		}
	}
	if mou.IsAttrLast() == api.AttrModuleYes {
		return s.fromFeedDynamic(c, mou, foreignID, types, pas.TabFrom)
	}
	if mou.DySort == dynamic.RandomSort {
		tmpNum = psNum + 1
	} else {
		tmpNum = psNum
	}
	if reply, err = s.dynamicDao.FetchDynamics(c, foreignID, mid, tmpNum, int64(mou.DySort), pas.Device, types, pas.Platform, "", "", pas.FromSpmid, pas.TabFrom); err != nil || reply == nil {
		log.Error("s.dynamicDao.FetchDynamics(%d) error(%v)", foreignID, err)
		return
	}
	if len(reply.Cards) == 0 {
		return
	}
	if mou.DySort == dynamic.RandomSort {
		dyCard := make([]*dynamic.DyCard, 0)
		for _, val := range reply.Cards {
			if val.Desc.DynamicID == pas.DynamicID {
				ok = true
				continue
			}
			dyCard = append(dyCard, val)
		}
		if ok { //有重复，过滤后的结果
			reply.Cards = dyCard
		}
		if len(dyCard) == int(tmpNum) {
			reply.Cards = dyCard[1:]
		}
	}
	list = &actmdl.Item{}
	list.FromDynamicModule(mou, nil)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	for _, v := range reply.Cards {
		tmpAct := &actmdl.Item{}
		tmpAct.FromDynamic(v)
		list.Item = append(list.Item, tmpAct)
	}
	if reply.HasMore > 0 && mou.DySort != dynamic.RandomSort {
		tmpMore := &actmdl.Item{}
		tmpMore.FromDynamicMore(mou, &actmdl.DynamicMore{TopicID: foreignID, PageID: topicInfo.ID, Sort: types, Name: topicTitle, Offset: reply.Offset})
		list.Item = append(list.Item, tmpMore)
	}
	// 组件数量大于1 才下发字段
	dynamicReply = list
	return
}

// ResourceDyn .
func (s *Service) ResourceDyn(c context.Context, arg *dynamic.ResourceDynReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	briRly, err := s.dynamicDao.BriefDynamics(c, arg.TopicID, arg.PageSize, arg.Mid, arg.Types, arg.Offset, 0)
	if err != nil {
		log.Error("s.dynamicDao.BriefDynamics(%d) error(%v)", arg.TopicID, err)
		return nil, err
	}
	if briRly == nil || len(briRly.Dynamics) == 0 {
		return nil, nil
	}
	var cvids, aids []int64
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		switch v.Type {
		case dynamic.ArticleType:
			cvids = append(cvids, v.Rid)
		case dynamic.VideoType:
			aids = append(aids, v.Rid)
		}
	}
	var (
		arcRly map[int64]*arccli.Arc
		artRly map[int64]*artmdl.Meta
	)
	eg := errgroup.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if arcRly, e = s.arcdao.ArchivesPB(ctx, aids, arg.Mid, arg.MobiApp, arg.Device); e != nil {
				log.Error("s.arcdao.Arcs aids(%v) error(%v)", aids, e)
				e = nil
			}
			return
		})
	}
	if len(cvids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if artRly, e = s.artdao.ArticleMetas(ctx, cvids, 2); e != nil {
				log.Error("s.artdao.ArticleMeta cvids(%v) error(%v)", cvids, e)
				e = nil
			}
			return
		})
	}
	_ = eg.Wait()
	rly := &actmdl.ResourceReply{HasMore: int32(briRly.HasMore), DyOffset: briRly.Offset}
	artDisplay := mou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		tmp := &actmdl.Item{}
		switch v.Type {
		case dynamic.VideoType:
			if va, ok := arcRly[v.Rid]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromResourceArc(arcRly[v.Rid], arcDisplay, nil)
		case dynamic.ArticleType:
			if va, ok := artRly[v.Rid]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromResourceArt(artRly[v.Rid], artDisplay)
		}
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// FormatResourceDyn .
func (s *Service) FormatResourceDyn(c context.Context, mou *api.NativeModule, dyn *api.Dynamic, mid int64, pas *actmdl.ParamFormatModule) (list *actmdl.Item) {
	types := strconv.Itoa(dynamic.VideoType) //默认排序
	if dyn != nil && len(dyn.SelectList) > 0 {
		types = strconv.FormatInt(dyn.SelectList[0].SelectType, 10)
	}
	dynArg := &dynamic.ResourceDynReq{TopicID: mou.Fid, PageSize: mou.Num, Types: types, Mid: mid, MobiApp: pas.MobiApp, Device: pas.Device}
	reply, err := s.ResourceDyn(c, dynArg, mou)
	if err != nil {
		log.Error("s.dynamicDao.BriefDynamics(%v) error(%v)", dynArg, err)
		return
	}
	if reply == nil || len(reply.List) == 0 {
		return
	}
	// 有卡片信息才下发组件
	list = s.ResourceJoin(c, reply, mou, pas)
	return
}

// FormatNewVideoAvid .
func (s *Service) FormatNewVideoAvid(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule) (list *actmdl.Item) {
	var (
		err      error
		psNum    = mou.Num
		actReply *actmdl.NewVideoReply
	)
	netType, tfType := showmdl.TrafficFree(pas.TfIsp)
	avidReq := &actmdl.NewAvidReq{ModuleID: mou.ID, Offset: 0, Ps: psNum,
		MobiApp: pas.MobiApp, Platform: pas.Platform, Build: pas.Build, Buvid: pas.Buvid, Device: pas.Device, TfIsp: pas.TfIsp, TfType: tfType, NetType: netType}
	if actReply, err = s.newAvidInfo(c, avidReq); err != nil {
		log.Error("s.newAvidInfo error(%v)", err)
		return
	}
	list = s.videoJoin(actReply, mou)
	return
}

// newVideoAct .
func (s *Service) newVideoAct(c context.Context, pas *actmdl.NewVideoActReq, mid int64) (res *actmdl.NewVideoReply, err error) {
	var (
		likeList *actapi.LikesReply
	)
	arg := &actapi.ActLikesReq{Sid: pas.Sid, Mid: mid, SortType: pas.SortType, Ps: int32(pas.Ps) + 6, Offset: pas.Offset}
	if likeList, err = s.actDao.ActLikes(c, arg); err != nil || likeList == nil {
		log.Error("s.actDao.ActLikes(%v) error(%v)", arg, err)
		return
	}
	lg := len(likeList.List)
	if lg == 0 || likeList.Subject == nil {
		return
	}
	res = &actmdl.NewVideoReply{Total: likeList.Total, Offset: likeList.Offset, HasMore: likeList.HasMore}
	var aids []*arccli.PlayAv
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			aids = append(aids, &arccli.PlayAv{Aid: v.Item.Wid})
		}
	}
	var arcs map[int64]*arccli.ArcPlayer
	if len(aids) > 0 {
		if arcs, err = s.arcdao.ArcsPlayer(c, aids); err != nil {
			log.Error("s.arcdao.ArcsPlayer(%v) error(%v)", aids, err)
			return
		}
	}
	lastOffset := arg.Offset
	for _, v := range likeList.List {
		lastOffset++
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		av, ok := arcs[v.Item.Wid]
		if !ok || av == nil || av.Arc == nil || !av.Arc.IsNormal() {
			continue
		}
		temp := &actmdl.Item{}
		// 获取首p秒开地址即可
		firstPlay := av.PlayerInfo[av.DefaultPlayerCid]
		temp.FromNewVideoCard(av.Arc, firstPlay, pas.Build, pas.MobiApp)
		res.Item = append(res.Item, temp)
		if len(res.Item) >= int(pas.Ps) {
			break
		}
	}
	if likeList.HasMore == 0 && lastOffset < likeList.Offset {
		res.HasMore = 1
	}
	res.Offset = lastOffset
	return
}

// FormatVideoAvid .
func (s *Service) FormatVideoAvid(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule, mid int64) (list *actmdl.Item) {
	var (
		err       error
		psNum     = mou.Num
		avidReply *actmdl.VideoReply
	)
	avidReq := &actmdl.AvidReq{ModuleID: mou.ID, Offset: 0, Ps: psNum, VideoMeta: pas.VideoMeta,
		MobiApp: pas.MobiApp, Platform: pas.Platform, Build: pas.Build, AvSort: mou.AvSort, Buvid: pas.Buvid, Device: pas.Device, TfIsp: pas.TfIsp, FromSpmid: pas.FromSpmid}
	if avidReply, err = s.AvidInfo(c, avidReq, mid); err != nil || avidReply == nil {
		log.Error("s.AvidInfo error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	if len(avidReply.DyReply) == 0 {
		return
	}
	list = &actmdl.Item{}
	list.FromVideoAvidModule(mou, mou.IsCardSingle())
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	for _, v := range avidReply.DyReply {
		temp := &actmdl.Item{}
		temp.FromVideoCard(v, mou.IsCardSingle())
		list.Item = append(list.Item, temp)
	}
	//查看更多标签
	if avidReply.HasMore >= 1 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &actmdl.Item{}
		tmpMore.FromVideoMore(mou, avidReply.Offset, pas.PageID, "", true)
		list.Item = append(list.Item, tmpMore)
	}
	return
}

// VideoAct .
func (s *Service) VideoAct(c context.Context, pas *actmdl.VideoActReq, mid int64) (res *actmdl.VideoReply, err error) {
	var (
		likeList *actapi.LikesReply
		rids     []*dynamic.RidInfo
		dyRes    *dynamic.DyResult
	)
	arg := &actapi.ActLikesReq{Sid: pas.Sid, Mid: mid, SortType: pas.SortType, Ps: int32(pas.Ps) + 6, Offset: pas.Offset}
	if likeList, err = s.actDao.ActLikes(c, arg); err != nil || likeList == nil {
		log.Error("s.actDao.ActLikes(%v) error(%v)", arg, err)
		return
	}
	lg := len(likeList.List)
	if lg == 0 || likeList.Subject == nil {
		return
	}
	res = &actmdl.VideoReply{Total: likeList.Total, Offset: likeList.Offset, HasMore: likeList.HasMore}
	rids = make([]*dynamic.RidInfo, 0, lg)
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			rids = append(rids, &dynamic.RidInfo{Rid: v.Item.Wid, Type: dynamic.VideoType})
		}
	}
	rous := &dynamic.Resources{Array: rids}
	if dyRes, err = s.dynamicDao.Dynamic(c, rous, pas.Platform, pas.RemoteFrom, pas.FromSpmid, mid, nil); err != nil || dyRes == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", rous, err)
		return
	}
	lastOffset := arg.Offset
	for _, v := range rids {
		lastOffset++
		if _, ok := dyRes.Cards[v.Rid]; !ok {
			continue
		}
		res.DyReply = append(res.DyReply, dyRes.Cards[v.Rid])
		if len(res.DyReply) >= int(pas.Ps) {
			break
		}
	}
	if likeList.HasMore == 0 && lastOffset < likeList.Offset {
		res.HasMore = 1
	}
	res.Offset = lastOffset
	return
}

func (s *Service) editOrigin(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	switch pas.RdbType {
	case api.RDBChannel: //编辑推荐卡-垂类id
		return s.editChannelOrigin(c, pas, mou)
	case api.RDBWeek: //编辑推荐卡-每周必看
		return s.editWeekOrigin(c, pas, mou)
	case api.RDBMustsee: //编辑推荐卡-入站必刷
		return s.editMustseeOrigin(c, pas, mou)
	case api.RDBRank: //编辑推荐卡-排行榜
		return s.editRankOrigin(c, pas, mou)
	}
	return &actmdl.ResourceReply{}, nil
}

func (s *Service) resourceOrigin(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	switch pas.RdbType {
	case api.RDBLive:
		return s.resourceLive(c, pas, mou)
	case api.RDOBusinessIDs:
		return s.businessIDs(c, pas, mou)
	case api.RDOBusinessCommodity:
		return s.businessCommodity(c, pas)
	case api.RDOOgvWid:
		return s.ogvWid(c, pas)
	case api.RDBWeek: //编辑推荐卡-每周必看
		return s.editWeekOrigin(c, pas, mou)
	}
	return &actmdl.ResourceReply{}, nil
}

// businessCommodity .
func (s *Service) businessCommodity(c context.Context, pas *actmdl.ResourceOriginReq) (*actmdl.ResourceReply, error) {
	pRes := &actmdl.ResourceReply{}
	// 根据商品id获取产品item
	rly, err := s.businessdao.ProductDetail(c, pas.SourceID, pas.Offset, pas.Ps)
	if err != nil {
		log.Error("s.businessdao.ProductDetail(%s,%d,%d) error(%v)", pas.SourceID, pas.Offset, pas.Ps, err)
		return pRes, nil
	}
	pRes.Offset = rly.Offset
	pRes.HasMore = rly.HasMore
	for _, v := range rly.ItemList {
		if v == nil {
			continue
		}
		tmp := &actmdl.Item{}
		tmp.FromResourceProduct(v)
		pRes.List = append(pRes.List, tmp)
	}
	return pRes, nil
}

func (s *Service) resourceLive(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	wid, err := strconv.ParseInt(pas.SourceID, 10, 64)
	if err != nil {
		log.Errorc(c, "Fail to parse wid, wid=%+v error=%+v", pas.SourceID, err)
		return nil, err
	}
	isLive := mou.IsAttrDisplayNodeNum()
	widItems, err := s.livedao.GetListByActId(c, wid, pas.SortType, isLive, pas.Ps, pas.Offset)
	if err != nil {
		log.Error("s.livedao.GetListByActId(%d,%d,%d) error(%v)", wid, pas.SortType, isLive, err)
		return nil, err
	}
	if widItems == nil {
		return &actmdl.ResourceReply{}, nil
	}
	list := make([]*actmdl.Item, 0)
	for _, v := range widItems.List {
		if v == nil {
			continue
		}
		item := &actmdl.Item{}
		item.FromResourceLive(v, c, s.c.Feature)
		list = append(list, item)
	}
	var hasMore int32
	if widItems.HasMore {
		hasMore = 1
	}
	return &actmdl.ResourceReply{
		List:    list,
		Offset:  widItems.Offset,
		HasMore: hasMore,
	}, nil
}

func (s *Service) ogvWid(c context.Context, pas *actmdl.ResourceOriginReq) (*actmdl.ResourceReply, error) {
	wid, err := strconv.ParseInt(pas.SourceID, 10, 64)
	if err != nil {
		log.Errorc(c, "Fail to parse wid, wid=%+v error=%+v", pas.SourceID, err)
		return nil, err
	}
	widItems, err := s.pgcdao.QueryWid(c, int32(wid), pas.Mid, pas.MobiApp, pas.Device, pas.Platform, int32(pas.Build))
	if err != nil {
		return nil, err
	}
	var hasMore int32
	// 首页返回 module.Num 条数据，二级页返回剩余的
	if pas.Offset > 0 {
		if len(widItems) > int(pas.Offset) {
			widItems = widItems[pas.Offset:]
		} else {
			// 返回空
			return &actmdl.ResourceReply{}, nil
		}
	} else if len(widItems) > int(pas.Ps) {
		hasMore = 1
		widItems = widItems[:pas.Ps]
	}
	offset := pas.Offset + int64(len(widItems))
	list := make([]*actmdl.Item, 0, len(widItems))
	for _, v := range widItems {
		item := &actmdl.Item{}
		item.FromResourceWidItem(v)
		list = append(list, item)
	}
	return &actmdl.ResourceReply{
		List:    list,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

// businessIDs .
func (s *Service) businessIDs(c context.Context, pas *actmdl.ResourceOriginReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	rly, err := s.businessdao.SourceDetail(c, pas.SourceID, pas.Offset, pas.Ps)
	if err != nil {
		log.Error("s.businessdao.SourceDetail(%v) error(%v)", pas, err)
		return nil, err
	}
	res := &actmdl.ResourceReply{Offset: rly.Offset, HasMore: rly.HasMore}
	var aids, cvids, epids, fids []int64
	for _, v := range rly.ItemList {
		if v == nil || v.ItemID == 0 {
			continue
		}
		switch v.Type {
		case api.MixAvidType:
			aids = append(aids, v.ItemID)
		case api.MixFolder:
			aids = append(aids, v.ItemID)
			fids = append(fids, v.FID)
		case api.MixCvidType:
			cvids = append(cvids, v.ItemID)
		case api.MixEpidType:
			epids = append(epids, v.ItemID)
		}
	}
	arcRly, artRly, epRly, foldRly, _ := s.resourceJoin(c, aids, cvids, epids, fids, nil, pas.Mid, pas.MobiApp, pas.Device)
	artDisplay := mou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	pgcDisplay := mou.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	for _, v := range rly.ItemList {
		if v == nil || v.ItemID == 0 {
			continue
		}
		tmp := &actmdl.Item{}
		switch v.Type {
		case api.MixAvidType:
			if va, ok := arcRly[v.ItemID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromResourceArc(arcRly[v.ItemID], arcDisplay, nil)
		case api.MixFolder:
			if va, ok := arcRly[v.ItemID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			if fold, ok := foldRly[v.FID]; !ok || fold == nil {
				continue
			}
			tmp.FromResourceArc(arcRly[v.ItemID], arcDisplay, foldRly[v.FID])
		case api.MixCvidType:
			if va, ok := artRly[v.ItemID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromResourceArt(artRly[v.ItemID], artDisplay)
		case api.MixEpidType:
			if va, ok := epRly[v.ItemID]; !ok || va == nil {
				continue
			}
			tmp.FromResourceEp(epRly[v.ItemID], pgcDisplay)
		default:
			continue
		}
		res.List = append(res.List, tmp)
	}
	return res, nil
}

// VideoAct .
func (s *Service) ResourceAct(c context.Context, pas *actmdl.ResourceActReq, mou *api.NativeModule) (*actmdl.ResourceReply, error) {
	arg := &actapi.ActLikesReq{Sid: pas.Sid, SortType: pas.SortType, Ps: int32(pas.Ps) + 6, Offset: pas.Offset}
	likeList, err := s.actDao.ActLikes(c, arg)
	if err != nil {
		log.Error("s.actDao.ActLikes(%v) error(%v)", arg, err)
		return nil, err
	}
	if likeList == nil || likeList.Subject == nil {
		return nil, nil
	}
	var isVideo, isArticle bool
	switch likeList.Subject.Type {
	case activitymdl.Article:
		isArticle = true
	case activitymdl.VideoLike, activitymdl.Video2, activitymdl.PhoneVideo:
		isVideo = true
	default:
		return nil, nil
	}
	res := &actmdl.ResourceReply{Offset: likeList.Offset, HasMore: likeList.HasMore}
	var ids []int64
	for _, v := range likeList.List {
		if v == nil || v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		ids = append(ids, v.Item.Wid)
	}
	if len(ids) == 0 {
		return res, nil
	}
	var (
		arcRly map[int64]*arccli.Arc
		artRly map[int64]*artmdl.Meta
		e      error
	)
	switch {
	case isVideo:
		arcRly, e = s.arcdao.ArchivesPB(c, ids, pas.Mid, pas.MobiApp, pas.Device)
	case isArticle:
		artRly, e = s.artdao.ArticleMetas(c, ids, 2)
	}
	if e != nil {
		return res, nil
	}
	artDisplay := mou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	lastOffset := arg.Offset
	for _, v := range likeList.List {
		lastOffset++
		if v == nil || v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		tmp := &actmdl.Item{}
		switch {
		case isVideo:
			if va, ok := arcRly[v.Item.Wid]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromResourceArc(arcRly[v.Item.Wid], arcDisplay, nil)
		case isArticle:
			if va, ok := artRly[v.Item.Wid]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromResourceArt(artRly[v.Item.Wid], artDisplay)
		}
		res.List = append(res.List, tmp)
		if len(res.List) >= int(pas.Ps) {
			break
		}
	}
	if res.HasMore == 0 && lastOffset < res.Offset {
		res.HasMore = 1
	}
	res.Offset = lastOffset
	return res, nil
}

// FormatResourceOrigin .
func (s *Service) FormatResourceOrigin(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule, sortType int64) (list *actmdl.Item) {
	confSort := mou.ConfUnmarshal()
	arg := &actmdl.ResourceOriginReq{SourceID: mou.TName, RdbType: confSort.RdbType, Ps: mou.Num, Offset: 0, Mid: pas.Mid, MobiApp: pas.MobiApp, Device: pas.Device, Platform: pas.Platform, Build: pas.Build}
	if confSort.RdbType == api.RDBLive {
		arg.SortType = sortType
	}
	actReply, err := s.resourceOrigin(c, arg, mou)
	if err != nil {
		log.Error("s.resourceOrigin(%v) error(%v)", arg, err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ResourceJoin(c, actReply, mou, pas)
	return
}

// FormatResourceAct .
func (s *Service) FormatResourceAct(c context.Context, mou *api.NativeModule, sortType int32, pas *actmdl.ParamFormatModule) (list *actmdl.Item) {
	arg := &actmdl.ResourceActReq{Sid: mou.Fid, SortType: sortType, Ps: mou.Num, Offset: 0, Mid: pas.Mid, MobiApp: pas.MobiApp, Device: pas.Device}
	actReply, err := s.ResourceAct(c, arg, mou)
	if err != nil {
		log.Error("s.VideoAct(%v) error(%v)", arg, err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ResourceJoin(c, actReply, mou, pas)
	return
}

// FormatVideoAct .
func (s *Service) FormatNewVideoAct(c context.Context, mou *api.NativeModule, sortType int32, mid int64, pas *actmdl.ParamFormatModule) (list *actmdl.Item) {
	var (
		err      error
		actReply *actmdl.NewVideoReply
	)
	netType, tfType := showmdl.TrafficFree(pas.TfIsp)
	arg := &actmdl.NewVideoActReq{Sid: mou.Fid, SortType: sortType, Ps: mou.Num, Offset: 0,
		MobiApp: pas.MobiApp, Platform: pas.Platform, Build: pas.Build, Buvid: pas.Buvid, Device: pas.Device, TfType: tfType, NetType: netType}
	if actReply, err = s.newVideoAct(c, arg, mid); err != nil || actReply == nil {
		log.Error("s.newVideoAct(%v) error(%v)", arg, err)
		return
	}
	list = s.videoJoin(actReply, mou)
	return
}

// ResourceJoin .
func (s *Service) videoJoin(req *actmdl.NewVideoReply, mou *api.NativeModule) *actmdl.Item {
	// 有卡片信息才下发组件
	if req == nil || len(req.Item) == 0 {
		return nil
	}
	list := &actmdl.Item{}
	list.FromNewVideoActModule(mou)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	list.Item = append(list.Item, req.Item...)
	//查看更多标签 && 开关控制
	if req.HasMore >= 1 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &actmdl.Item{}
		tmpMore.FromVideoMore(mou, req.Offset, mou.NativeID, req.DyOffset, true)
		list.Item = append(list.Item, tmpMore)
	}
	return list
}

// videoSecondJoin .
func (s *Service) videoSecondJoin(req *actmdl.NewVideoReply) *actmdl.LikeListRely {
	// 有卡片信息才下发组件
	if req == nil {
		return &actmdl.LikeListRely{}
	}
	rly := &actmdl.LikeListRely{DyOffset: req.DyOffset, HasMore: req.HasMore, Offset: req.Offset}
	list := &actmdl.Item{}
	list.FromNewVideoActModule(nil)
	list.Item = append(list.Item, req.Item...)
	rly.Cards = []*actmdl.Item{list}
	return rly
}

// FormatResourceAvid .
func (s *Service) FormatResourceAvid(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule) (list *actmdl.Item) {
	rly, err := s.ResourceAvid(c, mou, mou.Num, 0, pas)
	if err != nil || rly == nil {
		log.Error("s.AvidInfo error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ResourceJoin(c, rly, mou, pas)
	return
}

// FormatTimelineIDs .
func (s *Service) FormatTimelineIDs(c context.Context, mou *api.NativeModule, params *actmdl.ParamFormatModule) (list *actmdl.Item) {
	confSort := mou.ConfUnmarshal()
	//默认浮层
	ps := mou.Num
	if confSort.MoreSort == api.MoreExpand { // 下拉展示
		ps = 50
	}
	rly, err := s.timelineIDs(c, mou, ps, 0, params.Mid, params.MobiApp, params.Device)
	if err != nil || rly == nil {
		log.Error("s.TimelineIDs error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.timelineJoin(rly, mou)
	return
}

// formatReply .
func (s *Service) formatReply(mou *api.NativeModule) (list *actmdl.Item) {
	if mou.Fid == 0 {
		return
	}
	list = &actmdl.Item{}
	list.FromReplyModule(mou)
	return
}

func (s *Service) formatReserve(c context.Context, mou *api.NativeModule, rev *api.Reserve, mid int64, arg *actmdl.ParamFormatModule) (ck *actmdl.Item) {
	if rev == nil { //没有卡片，不下发组件
		return
	}
	var revIDs []int64
	//获取id
	for _, v := range rev.List {
		if v == nil || v.MType != api.MixUpReserve || v.ForeignID <= 0 {
			continue
		}
		revIDs = append(revIDs, v.ForeignID)
	}
	if len(revIDs) == 0 { //没有卡片，不下发组件
		return
	}
	//获取详情
	revRly := s.reserveInfo(c, revIDs, mid, mou.IsAttrIsDisplayUpIcon(), arg)
	var items []*actmdl.Item
	//拼接卡片信息
	for _, v := range rev.List {
		if v == nil || v.MType != api.MixUpReserve || v.ForeignID <= 0 {
			continue
		}
		if gv, ok := revRly[v.ForeignID]; !ok || gv == nil {
			continue
		}
		itemTep := &actmdl.Item{}
		itemTep.FromReserveExt(v, revRly[v.ForeignID], mou.IsAttrIsDisplayUpIcon(), arg.Build, mid, arg.MobiApp)
		items = append(items, itemTep)
	}
	if len(items) == 0 { //没有卡片，不下发组件
		return
	}
	ck = &actmdl.Item{}
	var lastItems []*actmdl.Item
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		lastItems = append(lastItems, tmpName)
	}
	lastItems = append(lastItems, items...)
	ck.FromReserve(mou, lastItems)
	return
}

// nolint:gocognit
func (s *Service) reserveInfo(c context.Context, ids []int64, mid, needIcon int64, arg *actmdl.ParamFormatModule) map[int64]*actmdl.ReserveRly {
	revRly, err := s.actDao.UpActReserveRelationInfo(c, mid, ids)
	if err != nil {
		log.Error("s.actDao.UpActReserveRelationInfo(%d,%v) error(%v)", mid, ids, err)
		return make(map[int64]*actmdl.ReserveRly)
	}
	var (
		aidStr    []string
		mids      []int64
		liveIDStr = make(map[int64][]string)
		nowtime   = time.Now().Unix()
	)
	rly := make(map[int64]*actmdl.ReserveRly)
	for _, v := range ids {
		if rval, ok := revRly[v]; !ok || rval == nil {
			continue
		}
		//话题活动页展示都为客态逻辑
		if revRly[v].UpActVisible != actapi.UpActVisible_DefaultVisible {
			continue
		}
		var changeType int64
		switch revRly[v].State {
		case actapi.UpActReserveRelationState_UpReserveRelated, actapi.UpActReserveRelationState_UpReserveRelatedOnline:
			changeType = actmdl.ReserveDisplayA
			if revRly[v].Type == actapi.UpActReserveRelationType_Course && int64(revRly[v].Etime) < nowtime {
				//预约结束未核销
				changeType = actmdl.ReserveDisplayE
			}
		case actapi.UpActReserveRelationState_UpReserveRelatedWaitCallBack, actapi.UpActReserveRelationState_UpReserveRelatedCallBackCancel, actapi.UpActReserveRelationState_UpReserveRelatedCallBackDone:
			if revRly[v].Type == actapi.UpActReserveRelationType_Archive {
				changeType = actmdl.ReserveDisplayC
				aidStr = append(aidStr, revRly[v].Oid)
			} else if revRly[v].Type == actapi.UpActReserveRelationType_Live {
				changeType = actmdl.ReserveDisplayLive
				liveIDStr[revRly[v].Upmid] = append(liveIDStr[revRly[v].Upmid], revRly[v].Oid)
			} else if revRly[v].Type == actapi.UpActReserveRelationType_Course {
				changeType = actmdl.ReserveDisplayC
			} else { //不认识的类型，不展示对应卡片
				continue
			}
		default: //不认识的类型，不展示对应卡片
			continue
		}
		mids = append(mids, revRly[v].Upmid)
		rly[v] = &actmdl.ReserveRly{
			ChangeType: changeType,
			Item:       revRly[v],
		}
	}
	var accRly map[int64]*accapi.Card
	eg := errgroup.WithContext(c)
	//获取账号信息
	if len(mids) > 0 && needIcon == 1 {
		eg.Go(func(ctx context.Context) error {
			var e error
			if accRly, e = s.accDao.Cards3GRPC(ctx, mids); e != nil { //获取账号信息失败，降级处理
				log.Error("s.accDao.Cards3GRPC(%v) error(%v)", mids, e)
			}
			return nil
		})
	}
	var (
		aids   []int64
		aidMap = make(map[int64]struct{})
	)
	for _, v := range aidStr {
		ak, err := strconv.ParseInt(v, 10, 64)
		if err != nil || ak <= 0 {
			continue
		}
		if _, ok := aidMap[ak]; !ok { //去重
			aids = append(aids, ak)
			aidMap[ak] = struct{}{}
		}
	}
	//获取稿件信息
	arcRly := make(map[string]*arccli.Arc)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcRes, e := s.arcdao.ArchivesPB(ctx, aids, mid, arg.MobiApp, arg.Device)
			if e != nil { //获取账号信息失败，降级处理
				log.Error("s.accDao.ArchivesPB(%v) error(%v)", mids, e)
				return nil
			}
			for k, v := range arcRes {
				arcRly[fmt.Sprintf("%d", k)] = v
			}
			return nil
		})
	}
	//获取直播信息
	var liveRly map[int64]*roomgategrpc.SessionInfos
	if len(liveIDStr) > 0 {
		//获取秒开参数
		var liveNetwork string
		batchArg, _ := arcmid.FromContext(c)
		//根据直播服务要求匹配对应关系
		switch batchArg.NetType {
		case arccli.NetworkType_NT_UNKNOWN:
			liveNetwork = "other"
		case arccli.NetworkType_WIFI:
			liveNetwork = "wifi"
		default:
			liveNetwork = "other"
		}
		playurlReq := &roomgategrpc.PlayUrlReq{
			Uid:         mid,
			Uipstr:      metadata.String(c, metadata.RemoteIP),
			HttpsUrlReq: arg.HttpsUrlReq == 1,
			Platform:    arg.Platform,
			Build:       arg.Build,
			DeviceName:  arg.Device,
			Network:     liveNetwork,
			ReqBiz:      "/x/v2/activity/index", //请求接口名
		}
		eg.Go(func(ctx context.Context) (e error) {
			if liveRly, e = s.livedao.SessionInfoBatch(ctx, liveIDStr, playurlReq, []string{actmdl.LiveEnterFrom}); e != nil {
				log.Error("s.livedao.SessionInfoBatch(%v) error(%v)", liveIDStr, e)
				e = nil
			}
			return
		})
	}
	_ = eg.Wait()
	lastRly := make(map[int64]*actmdl.ReserveRly)
	for k, v := range rly {
		if v == nil || v.Item == nil {
			continue
		}
		switch v.Item.Type {
		case actapi.UpActReserveRelationType_Archive:
			if aVal, ok := arcRly[v.Item.Oid]; ok && aVal.IsNormal() {
				v.Arc = aVal
			}
		case actapi.UpActReserveRelationType_Live:
			if acVal, ok := liveRly[v.Item.Upmid]; ok && acVal != nil {
				if seVal, k := acVal.SessionInfoPerLive[v.Item.Oid]; k && seVal != nil {
					v.Live = &actmdl.LiveInfos{
						RoomId:             acVal.RoomId,
						Uid:                acVal.Uid,
						JumpUrl:            acVal.JumpUrl,
						Title:              acVal.Title,
						SessionInfoPerLive: seVal,
					}
				}
			}
		default:
		}
		if needIcon == 1 {
			if actVal, ok := accRly[v.Item.Upmid]; ok {
				v.Account = actVal
			}
		}
		v.DisplayType = v.ChangeType
		//容错:类型CD 没有获取直播状态不下发
		if v.ChangeType == actmdl.ReserveDisplayLive {
			if v.Live == nil || v.Live.SessionInfoPerLive == nil {
				continue
			}
			switch v.Live.SessionInfoPerLive.Status {
			case actmdl.Living:
				v.DisplayType = actmdl.ReserveDisplayC
			case actmdl.LiveEnd:
				v.DisplayType = actmdl.ReserveDisplayD
			default:
				v.DisplayType = actmdl.ReserveDisplayE
			}
		}
		lastRly[k] = v
	}
	return lastRly
}

func (s *Service) formatGame(c context.Context, mou *api.NativeModule, games *api.Game, mid int64, mobiApp string) *actmdl.Item {
	if games == nil { //没有卡片，不下发组件
		return nil
	}
	var gameIDs []int64
	//获取游戏id
	for _, v := range games.List {
		if v == nil || v.MType != api.MixGame || v.ForeignID <= 0 {
			continue
		}
		gameIDs = append(gameIDs, v.ForeignID)
	}
	if len(gameIDs) == 0 { //没有卡片，不下发组件
		return nil
	}
	//获取游戏详情
	gamesInfo := s.gameDao.BatchMultiGameInfo(c, gameIDs, mid, mobiApp)
	var items []*actmdl.Item
	//拼接卡片信息
	for _, v := range games.List {
		if v == nil || v.MType != api.MixGame || v.ForeignID <= 0 {
			continue
		}
		if gv, ok := gamesInfo[v.ForeignID]; !ok || gv == nil {
			continue
		}
		itemTep := &actmdl.Item{}
		itemTep.FromGameExt(v, gamesInfo[v.ForeignID])
		items = append(items, itemTep)
	}
	if len(items) == 0 { //没有卡片，不下发组件
		return nil
	}
	var lastItems []*actmdl.Item
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		lastItems = append(lastItems, tmpName)
	}
	lastItems = append(lastItems, items...)
	ck := &actmdl.Item{}
	ck.FromGame(mou, lastItems)
	return ck
}

// FormatOgvSeasonID .
func (s *Service) FormatOgvSeasonID(c context.Context, mou *api.NativeModule, mid int64) (list *actmdl.Item) {
	rly, err := s.ogvSeasonID(c, mou, mou.Num, 0, mid)
	if err != nil || rly == nil {
		log.Error("s.TimelineIDs error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ogvSeasonJoin(rly, mou)
	return
}

// FormatOgvSeasonResource .
func (s *Service) FormatOgvSeasonResource(c context.Context, mou *api.NativeModule, mid int64) (list *actmdl.Item) {
	rly, err := s.ogvSeasonResource(c, mou, mou.Num, 0, mid)
	if err != nil || rly == nil {
		log.Error("s.ogvSeasonResource error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ogvSeasonJoin(rly, mou)
	return
}

// FormatTimelineResource .
func (s *Service) FormatTimelineResource(c context.Context, mou *api.NativeModule) (list *actmdl.Item) {
	//一次取50个
	confSort := mou.ConfUnmarshal()
	//默认浮层
	ps := mou.Num
	if confSort.MoreSort == api.MoreExpand { // 下拉展示
		ps = 50
	}
	rly, err := s.timelineResource(c, mou, ps, 0)
	if err != nil || rly == nil {
		log.Error("s.TimelineIDs error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.timelineJoin(rly, mou)
	return
}

// FormatResourceRole .
func (s *Service) FormatResourceRole(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule) *actmdl.Item {
	rly, err := s.ResourceRole(c, mou, 0, int(mou.Num))
	if err != nil || rly == nil {
		log.Error("Fail to get resource of role, module=%+v error=%+v", mou, err)
		return nil
	}
	// 有卡片信息才下发组件
	return s.ResourceJoin(c, rly, mou, pas)
}

// ResourceJoin .
func (s *Service) ResourceJoin(c context.Context, req *actmdl.ResourceReply, mou *api.NativeModule, pas *actmdl.ParamFormatModule) *actmdl.Item {
	// 有卡片信息才下发组件
	if req == nil || len(req.List) == 0 {
		return nil
	}
	list := &actmdl.Item{}
	list.FromResourceModule(c, s.c.Feature, mou, pas.MobiApp, pas.Build)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	list.Item = append(list.Item, req.List...)
	if req.HasMore > 0 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &actmdl.Item{}
		tmpMore.FromVideoMore(mou, req.Offset, mou.NativeID, req.DyOffset, false)
		list.Item = append(list.Item, tmpMore)
	}
	return list
}

// ogvSeasonJoin .
func (s *Service) ogvSeasonJoin(req *actmdl.ResourceReply, mou *api.NativeModule) *actmdl.Item {
	// 有卡片信息才下发组件
	if req == nil || len(req.List) == 0 {
		return nil
	}
	list := &actmdl.Item{}
	list.FromOgvSeasonModule(mou)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	list.Item = append(list.Item, req.List...)
	if req.HasMore > 0 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &actmdl.Item{}
		tmpMore.FromOgvSeasonMore(mou, req.Offset)
		list.Item = append(list.Item, tmpMore)
	}
	return list
}

// timelineJoin .
func (s *Service) timelineJoin(req *actmdl.ResourceReply, mou *api.NativeModule) *actmdl.Item {
	// 有卡片信息才下发组件
	if req == nil || len(req.List) == 0 {
		return nil
	}
	list := &actmdl.Item{}
	list.FromTimelineModule(mou)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	confSort := mou.ConfUnmarshal()
	if confSort.MoreSort == api.MoreExpand { // 下拉展示
		var before, after, con []*actmdl.Item
		if len(req.List) > int(mou.Num*2) {
			before = req.List[:mou.Num*2]
			after = req.List[mou.Num*2:]
			con = append(con, before...)
			moreTmp := &actmdl.Item{}
			moreTmp.FromTimelineExpand(mou)
			moreTmp.Item = append(moreTmp.Item, after...)
			con = append(con, moreTmp)
		} else {
			con = req.List
		}
		list.Item = append(list.Item, con...)
	} else { //默认浮层
		list.Item = append(list.Item, req.List...)
		if req.HasMore > 0 {
			tmpMore := &actmdl.Item{}
			tmpMore.FromTimelineMore(mou, req.Offset)
			list.Item = append(list.Item, tmpMore)
		}
	}
	return list
}

// FormatVideoAct .
func (s *Service) FormatVideoAct(c context.Context, mou *api.NativeModule, sortType int32, mid int64, pas *actmdl.ParamFormatModule) (list *actmdl.Item) {
	var (
		err      error
		actReply *actmdl.VideoReply
	)
	arg := &actmdl.VideoActReq{Sid: mou.Fid, SortType: sortType, Ps: mou.Num, Offset: 0, VideoMeta: pas.VideoMeta,
		MobiApp: pas.MobiApp, Platform: pas.Platform, Build: pas.Build, Buvid: pas.Buvid, Device: pas.Device, TfIsp: pas.TfIsp, FromSpmid: pas.FromSpmid}
	if actReply, err = s.VideoAct(c, arg, mid); err != nil || actReply == nil {
		log.Error("s.VideoAct(%v) error(%v)", arg, err)
		return
	}
	// 有卡片信息才下发组件
	if len(actReply.DyReply) == 0 {
		return
	}
	list = &actmdl.Item{}
	list.FromVideoActModule(mou, mou.IsCardSingle())
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	for _, v := range actReply.DyReply {
		temp := &actmdl.Item{}
		temp.FromVideoCard(v, mou.IsCardSingle())
		list.Item = append(list.Item, temp)
	}
	//查看更多标签 && 开关控制
	if actReply.HasMore >= 1 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &actmdl.Item{}
		tmpMore.FromVideoMore(mou, actReply.Offset, pas.PageID, "", true)
		list.Item = append(list.Item, tmpMore)
	}
	return
}

// FormatVideo .
func (s *Service) FormatVideo(c context.Context, mou *api.NativeModule, sortType int32, mid int64, pas *actmdl.ParamFormatModule) (videoReply *actmdl.Item) {
	var (
		likeList *actapi.LikesReply
		rids     []*dynamic.RidInfo
		err      error
		dyReply  *dynamic.DyResult
		itemObj  map[int64]*actapi.ItemObj
		ext      *actmdl.UrlExt
		psNum    = int32(mou.Num)
		list     *actmdl.Item
		moreItem bool
	)
	if mou.IsAttrLast() == api.AttrModuleYes {
		psNum = 5
	}
	arg := &actapi.ActLikesReq{Sid: mou.Fid, Mid: mid, SortType: sortType, Ps: psNum, Offset: 0}
	if likeList, err = s.actDao.ActLikes(c, arg); err != nil || likeList == nil {
		log.Error("s.actDao.ActLikes(%v) error(%v)", arg, err)
		return
	}
	lg := len(likeList.List)
	if lg == 0 || likeList.Subject == nil {
		return
	}
	rids = make([]*dynamic.RidInfo, 0, lg)
	widType := dynamic.VideoType
	if likeList.Subject.Type == activitymdl.Article {
		widType = dynamic.ArticleType
	}
	itemObj = make(map[int64]*actapi.ItemObj, lg)
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			rids = append(rids, &dynamic.RidInfo{Rid: v.Item.Wid, Type: int64(widType)})
			itemObj[v.Item.Wid] = v
		}
	}
	rous := &dynamic.Resources{Array: rids}
	if dyReply, err = s.dynamicDao.Dynamic(c, rous, pas.Platform, "", pas.FromSpmid, mid, nil); err != nil || dyReply == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", rous, err)
		return
	}
	list = &actmdl.Item{}
	if mou.IsAttrLast() == api.AttrModuleYes {
		ext = &actmdl.UrlExt{Sid: mou.Fid, SortType: sortType, RemoteFrom: actmdl.RemoteActivity, ConfModuleID: mou.ID}
	}
	list.FromVideoModule(mou, ext)
	list.Item = make([]*actmdl.Item, 0, lg+1)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	if mou.IsAttrLast() == api.AttrModuleYes {
		videoReply = list
		return
	}
	for _, v := range rids {
		if _, ok := dyReply.Cards[v.Rid]; !ok {
			continue
		}
		if _, k := itemObj[v.Rid]; !k {
			continue
		}
		temp := &actmdl.Item{}
		if likeList.Subject.Type == activitymdl.VideoLike {
			temp.FromVideoLike(dyReply.Cards[v.Rid], itemObj[v.Rid])
		} else {
			temp.FromVideo(dyReply.Cards[v.Rid])
		}
		list.Item = append(list.Item, temp)
		moreItem = true
	}
	// 视频组件 查看更多标签
	if moreItem && likeList.HasMore >= 1 {
		tmpMore := &actmdl.Item{}
		tmpMore.FromVideoMore(mou, likeList.Offset, pas.PageID, "", false)
		list.Item = append(list.Item, tmpMore)
	}
	// 视频组件数量大于1 才下发字段
	if moreItem {
		videoReply = list
	}
	return
}

func (s *Service) FormatBanner(c context.Context, mou *api.NativeModule, topicInfo *api.NativePage) (list *actmdl.Item) {
	var (
		mid     int64
		err     error
		infoRep *accapi.InfoReply
		user    *actmdl.UserInfo
	)
	mid = topicInfo.RelatedUid
	if mid > 0 {
		if infoRep, err = s.accDao.Info3GRPC(c, mid); err != nil {
			log.Error("FormatBanner s.accDao.Info3GRPC mid(%d) error(%v)", mid, err)
		}
		if infoRep != nil {
			user = &actmdl.UserInfo{Mid: mid, Name: infoRep.Info.Name, Face: infoRep.Info.Face}
		}
	}
	list = &actmdl.Item{}
	list.FromBannerModule(mou, topicInfo, user)
	return
}

func (s *Service) FormatStatement(mou *api.NativeModule) (list *actmdl.Item) {
	list = &actmdl.Item{}
	list.FromStatementModule(mou)
	return
}

func (s *Service) FormatNavigation(c context.Context, mou *api.NativeModule, mobiApp string, build int64) (list *actmdl.Item) {
	list = &actmdl.Item{}
	list.FromNavigation(c, s.c.Feature, mou, mobiApp, build)
	return
}

// FormatLive .
func (s *Service) FormatLive(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule, mid int64) *actmdl.Item {
	nowTime := time.Now().Unix()
	// 在设置的时间之内
	if mou.Stime > nowTime || mou.Etime < nowTime || mou.Fid == 0 {
		return nil
	}
	isHttps := pas.HttpsUrlReq == 1
	liveInfos, err := s.livedao.GetCardInfo(c, []int64{mou.Fid}, mid, pas.Build, pas.Platform, pas.Device, isHttps)
	if err != nil {
		log.Error("s.livedao.GetCardInfo(%d,%d) error(%v)", mou.Fid, mid, err)
		return nil
	}
	liveID := uint64(mou.Fid)
	if _, ok := liveInfos[liveID]; !ok || liveInfos[liveID] == nil {
		return nil
	}
	// 未开播
	if liveInfos[liveID].LiveStatus != 1 && mou.LiveType == 0 {
		return nil
	}
	list := &actmdl.Item{}
	list.FromLiveModule(mou)
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &actmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	liveCard := &actmdl.Item{}
	liveCard.FromLive(mou, liveInfos[liveID])
	list.Item = append(list.Item, liveCard)
	return list
}

func (s *Service) FormatSingleDyn(c context.Context, mou *api.NativeModule, pas *actmdl.ParamFormatModule, mid int64) (list *actmdl.Item) {
	var (
		reply *dynamic.DyResult
		err   error
		title string
		rid   int64
	)
	if reply, err = s.dynamicDao.Dynamic(c, nil, pas.Platform, "", pas.FromSpmid, mid, []int64{pas.DynamicID}); err != nil {
		log.Error("FormatSingleDyn s.dynamicDao.Dynamic() mid(%d) error(%v)", mid, err)
		return
	}
	if reply == nil {
		return
	}
	for k := range reply.Cards {
		rid = k
	}
	switch pas.ActivityFrom {
	case dynamic.FromFeed:
		title = dynamic.FeedStat
	case dynamic.FromRank:
		title = dynamic.AllActStat
	}
	if result, ok := reply.Cards[rid]; ok && result != nil {
		list = &actmdl.Item{}
		list.FromSingleDynModule(mou, title, result)
	}
	return
}

func (s *Service) FromSingleDynTm(c context.Context, pas *actmdl.ParamActIndex, mid int64) (list *actmdl.SingleDyn) {
	var (
		reply *dynamic.DyResult
		err   error
		rid   int64
	)
	if reply, err = s.dynamicDao.Dynamic(c, nil, pas.Platform, "", pas.FromSpmid, mid, []int64{pas.DynamicID}); err != nil {
		log.Error("FromSingleDyn s.dynamicDao.Dynamic(%d) mid(%d) error(%+v)", pas.DynamicID, mid, err)
		return
	}
	if reply == nil {
		return
	}
	for k := range reply.Cards {
		rid = k
	}
	if result, ok := reply.Cards[rid]; ok && result != nil {
		list = &actmdl.SingleDyn{}
		list.FromSingleDynTm(result)
	}
	return
}

// fromPart .
// nolint:gocognit
func (s *Service) FromPart(c context.Context, mou *api.NativeModule, part *api.Participation, page *api.NativePage) (ck *actmdl.Item) {
	var (
		sids    []int64
		subsRly map[int64]*actapi.ActSubProtocolReply
		cards   []*actmdl.Item
		e       error
	)
	if mou == nil || part == nil || len(part.List) == 0 {
		return
	}
	for _, v := range part.List {
		// 只会有一个活动投稿组件
		if (v.IsPartVideo() || v.IsPartArticle()) && v.ForeignID > 0 {
			sids = append(sids, v.ForeignID)
		}
	}
	if len(sids) > 0 {
		if subsRly, e = s.actDao.ActSubsProtocol(c, sids); e != nil {
			log.Error("s.actDao.ActSubsProtocol(%v) error(%v)", sids, e)
		}
	}
	useNewTopic := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.PartiUseNewTopic, nil)
	for _, v := range part.List {
		var (
			image, uri, joinType string
		)
		ext := &api.PartiExt{}
		if v.Ext != "" {
			if err := json.Unmarshal([]byte(v.Ext), ext); err != nil {
				log.Error("Fail to Unmarshal PartiExt, ext=%s error=%+v", v.Ext, err)
			}
		}
		switch {
		case v.IsPartArticle():
			image = "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/50d91hpX_w156_h156.png"
			// 专栏暂时不支持带活动信息
			uri = "https://member.bilibili.com/article-text/mobile"
			joinType = "article"
		case v.IsPartDynamic():
			image = "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/0-uIlgov_w156_h156.png"
			uri = func() string {
				if useNewTopic {
					var newTid string
					if ext.NewTid > 0 {
						newTid = strconv.FormatInt(ext.NewTid, 10)
					}
					return fmt.Sprintf("bilibili://following/publish?topicV2ID=%s", newTid)
				}
				titStr := fmt.Sprintf("#%s#", page.Title)
				infoDes := url.Values{}
				infoDes.Set("infoDescription", titStr)
				tmp := infoDes.Encode()
				if strings.IndexByte(tmp, '+') > -1 {
					tmp = strings.Replace(tmp, "+", "%20", -1)
				}
				return fmt.Sprintf("bilibili://following/publishInfo?%s", tmp)
			}()
			joinType = "dynamic"
		case v.IsPartVideo():
			image = "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/E-vXzW-~_w156_h156.png"
			var upFrom string
			//上传页面 0:上传 1.拍摄
			if v.UpType == 0 {
				joinType = "video-choose"
				upFrom = "0"
			} else {
				joinType = "video-shoot"
				upFrom = "1"
			}
			actDes := url.Values{}
			actDes.Set("copyright", "1")
			actDes.Set("from", upFrom)
			actDes.Set("relation_from", "NAactivityb")
			actDes.Set("is_new_ui", "1")
			// 拼接活动信息
			if ext.NewTid > 0 {
				actDes.Set("topic_id", strconv.FormatInt(ext.NewTid, 10))
			}
			if v.ForeignID > 0 {
				actDes.Set("mission_id", strconv.FormatInt(v.ForeignID, 10))
				if sVal, sok := subsRly[v.ForeignID]; sok && sVal != nil && sVal.Subject != nil && sVal.Protocol != nil {
					actDes.Set("mission_name", sVal.Protocol.Tags)
				}
			}
			uri = fmt.Sprintf("bilibili://uper/user_center/add_archive/?%s", actDes.Encode())
		default:
			continue
		}
		tmp := &actmdl.Item{}
		tmp.FromPartExt(v, image, uri, joinType)
		cards = append(cards, tmp)
	}
	if len(cards) > 0 {
		ck = &actmdl.Item{
			ItemID:  mou.ID,
			Param:   strconv.FormatInt(mou.ID, 10),
			Image:   s.c.Custom.PartImage,
			UnImage: s.c.Custom.PartUnImage,
			Item:    cards,
		}
	}
	return
}

// FormatRecommend .
func (s *Service) FormatRecommend(c context.Context, mou *api.NativeModule, recom *api.Recommend, mid int64) (ck *actmdl.Item) {
	var (
		fids      []int64
		followRly map[int64]*relationgrpc.FollowingReply
		cards     map[int64]*accountgrpc.Card
		items     []*actmdl.Item
	)
	if mou.IsRcmdSource() {
		confSort := mou.ConfUnmarshal()
		switch confSort.SourceType {
		case api.SourceTypeRank: //排行榜
			return s.rankListFromModule(c, mou, mid)
		default:
			fids, _ = s.rcmdSourceData(c, mou, mid)
			recom = trans2RecommendPB(fids)
		}
	}
	for _, v := range recom.List {
		if v.ForeignID > 0 {
			fids = append(fids, v.ForeignID)
		}
	}
	if len(fids) == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	// 获取关注关系
	if mid > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if followRly, e = s.reldao.RelationsGRPC(ctx, mid, fids); e != nil {
				log.Error(" s.reldao.RelationsGRPC(%d,%v) error(%v)", mid, fids, e)
				e = nil
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		//获取用户信息
		if cards, e = s.accDao.Cards3GRPC(ctx, fids); e != nil {
			log.Error("s.accDao.Cards3GRPC(%v) error(%v)", fids, e)
			e = nil
		}
		return
	})
	_ = eg.Wait()
	for _, reVal := range recom.List {
		if reVal.ForeignID == 0 {
			continue
		}
		if _, aok := cards[reVal.ForeignID]; !aok {
			continue
		}
		ext := &actmdl.ClickExt{}
		if rel, ok := followRly[reVal.ForeignID]; ok {
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			if rel.Attribute == 2 || rel.Attribute == 6 {
				ext.IsFollow = true
			}
		}
		itemTep := &actmdl.Item{}
		itemTep.FromRecommendExt(reVal, cards[reVal.ForeignID], ext)
		items = append(items, itemTep)
	}
	if len(items) == 0 {
		return
	}
	ck = &actmdl.Item{}
	var lastItems []*actmdl.Item
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	lastItems = append(lastItems, items...)
	ck.FromRecommend(mou, lastItems)
	return
}

func (s *Service) FormatRcmdVertical(c context.Context, mou *api.NativeModule, mixExt *api.Recommend, mid int64) *actmdl.Item {
	var uids []int64
	if mou.IsRcmdVerticalSource() {
		uids, _ = s.rcmdSourceData(c, mou, mid)
		mixExt = trans2RecommendPB(uids)
	}
	uids = getForeignIDsFromMix(mixExt.List)
	if len(uids) == 0 {
		return nil
	}
	eg := errgroup.WithContext(c)
	// 获取关注关系
	var relations map[int64]*relationgrpc.FollowingReply
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			var err error
			if relations, err = s.reldao.RelationsGRPC(ctx, mid, uids); err != nil {
				log.Error("Fail to get relations, mid=%d uids=%+v error=%+v", mid, uids, err)
			}
			return nil
		})
	}
	// 获取用户信息
	var accounts map[int64]*accountgrpc.Card
	eg.Go(func(ctx context.Context) error {
		var err error
		if accounts, err = s.accDao.Cards3GRPC(ctx, uids); err != nil {
			log.Error("Fail to get accounts, uids=%+v error=%+v", uids, err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to get user_info, err=%+v", err)
		return nil
	}
	var items []*actmdl.Item
	for _, v := range mixExt.List {
		if _, ok := accounts[v.ForeignID]; !ok {
			continue
		}
		clickExt := &actmdl.ClickExt{}
		if relation, ok := relations[v.ForeignID]; ok && relation != nil {
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			if relation.Attribute == 2 || relation.Attribute == 6 {
				clickExt.IsFollow = true
			}
		}
		item := &actmdl.Item{}
		item.FromRcmdVerticalItem(v, accounts[v.ForeignID], clickExt)
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	rcmd := &actmdl.Item{}
	rcmd.FromRcmdVertical(items)
	rcmdModule := &actmdl.Item{}
	rcmdModule.FromRcmdVerticalModule(mou)
	if mou.Meta != "" {
		titleImage := &actmdl.Item{}
		titleImage.FromTitleImage(mou)
		rcmdModule.Item = append(rcmdModule.Item, titleImage)
	}
	rcmdModule.Item = append(rcmdModule.Item, rcmd)
	return rcmdModule
}

// ActTab .
// nolint:gocognit
func (s *Service) ActTab(c context.Context, arg *actmdl.ParamActTab) (res *actmdl.TabReply, e error) {
	defer func() {
		if e != nil {
			// 兜底处理
			if arg.PageID > 0 {
				res = &actmdl.TabReply{ErrLimit: &actmdl.ErrLimit{Code: ecode.NothingFound.Code(), Message: "当前页面状态发生变化", Button: &actmdl.Button{Title: "前往活动页面", Link: "https://www.bilibili.com/blackboard/dynamic/" + strconv.FormatInt(arg.PageID, 10)}}}
				e = nil
			}
		}
	}()
	// 获取tab组件信息
	var rly *api.NatTabModulesReply
	if rly, e = s.actDao.NatTabModules(c, arg.TabID); e != nil {
		log.Error("s.actDao.NatTabModules(%d) error(%v) or nil", arg.TabID, e)
		return
	}
	if rly == nil || rly.Tab == nil || len(rly.List) == 0 {
		e = ecode.NothingFound
		return
	}
	// 获取pageinfo
	var pageIDs []int64
	for _, v := range rly.List {
		if v == nil || !v.IsOnline() {
			continue
		}
		if v.IsTabPage() && v.Pid > 0 {
			pageIDs = append(pageIDs, v.Pid)
		}
	}
	var pagesInfo map[int64]*api.NativePage
	var actErr error
	if len(pageIDs) > 0 {
		if pagesInfo, actErr = s.actDao.NativePages(c, pageIDs); actErr != nil {
			log.Error("s.actDao.NativePage(%v) error(%v)", pageIDs, actErr)
		}
	}
	//判断对应的module是否还存在绑定关系
	var defaultModule *actmdl.TabModule
	var mds []*actmdl.TabModule
	for _, v := range rly.List {
		if v == nil || !v.IsOnline() {
			continue
		}
		tmp := &actmdl.TabModule{}
		if !tmp.FormatTabModule(v) {
			continue
		}
		if canShowTab(arg.MobiApp, arg.Build) {
			tmp.ShareOrigin = actmdl.OriginTab
		}
		if tmp.TabModuleID == arg.TabModuleID {
			tmp.Select = true
			defaultModule = tmp
		}
		// 错误可以忽略
		var notOK bool
		if actErr == nil && v.IsTabPage() && v.Pid > 0 {
			notOK = func() bool {
				if pVal, ok := pagesInfo[v.Pid]; !ok || !pVal.IsOnline() {
					return true
				}
				tmp.TopicName = pagesInfo[v.Pid].Title
				tmp.ForeignID = pagesInfo[v.Pid].ForeignID
				return false
			}()
		}
		// page不存在且不是默认选中的module时跳过
		if notOK && !tmp.Select {
			continue
		}
		mds = append(mds, tmp)
	}
	if defaultModule == nil { //module不存在 errlimit处理，跳转对应话题活动页面
		e = ecode.NothingFound
		return
	}
	res = &actmdl.TabReply{
		Tab: &actmdl.Tab{
			BgType:        rly.Tab.BgType, //背景类型 1:图片 2:纯色
			BgImg:         rly.Tab.BgImg,
			BgColor:       rly.Tab.BgColor,
			IconType:      rly.Tab.IconType,      //图标样式 1:自定义图标+文字 2:文字
			ActiveColor:   rly.Tab.ActiveColor,   //选中态颜色
			InactiveColor: rly.Tab.InactiveColor, //未选中态颜色
		},
	}
	nowTime := time.Now().Unix()
	// tab在线 module 存在  下发整个tab
	if rly.Tab.IsOnline() && int64(rly.Tab.Stime) > 0 && int64(rly.Tab.Stime) <= nowTime && (int64(rly.Tab.Etime) >= nowTime || int64(rly.Tab.Etime) <= 0) {
		res.Tab.Items = mds
	} else {
		res.Tab.Items = []*actmdl.TabModule{defaultModule}
	}
	return
}

// ActNativeTab .
func (s *Service) ActNativeTab(c context.Context, arg *pb.ActNativeTabReq) (*pb.ActNativeTabReply, error) {
	// 获取tab信息
	rly, e := s.actDao.NativePagesTab(c, arg.Pids, arg.Category)
	if e != nil {
		return nil, e
	}
	res := make(map[int64]*pb.ActNativeTab, len(arg.Pids))
	for _, v := range arg.Pids {
		if val, ok := rly[v]; !ok || val == nil || val.TabID == 0 {
			continue
		}
		url := fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d", rly[v].PageID, rly[v].TabID, rly[v].TabModuleID)
		res[v] = &pb.ActNativeTab{Url: url}
	}
	return &pb.ActNativeTabReply{List: res}, nil
}

// nolint:gomnd
func (s *Service) ActReceive(c context.Context, arg *actmdl.ParamReceive, mid int64) (state int, msg string, err error) {
	switch arg.Goto {
	case actmdl.GotoClickUpAppointment:
		switch arg.State {
		case 1: //1 未预约
			err = s.actDao.AddReserve(c, arg.FID, mid)
			if err == nil {
				state = 2
			}
		case 2: //2 已预约
			err = s.actDao.DelReserve(c, arg.FID, mid)
			if err == nil {
				state = 1
			}
		default: //0 不可预约
			msg = "不可预约"
		}
	case actmdl.GotoClickPendant:
		var userState int
		if userState, err = s.actDao.AwardSubjectState(c, arg.FID, mid); err != nil {
			log.Error("%+v", err)
			return
		}
		// 0 无资格或无奖励，1 未领取，2 已领取
		switch userState {
		case 1:
			err = s.actDao.RewardSubject(c, arg.FID, mid)
			if err != nil {
				log.Error("%+v", err)
				err = xecode.AppReceiveErr
				return
			}
			msg = "奖励领取成功"
			state = 2
		case 2:
			state = 2
			msg = "你已经领取了该奖励"
		default:
			msg = "暂无领取资格"
		}
	default:
		err = ecode.RequestErr
	}
	return
}

func (s *Service) FormatEditorOrigin(c context.Context, mou *api.NativeModule, param *actmdl.ParamFormatModule) *actmdl.Item {
	confSort := mou.ConfUnmarshal()
	//无限feed流
	num := mou.Num
	if mou.IsAttrLast() == api.AttrModuleYes && (confSort.RdbType == api.RDBMustsee || confSort.RdbType == api.RDBChannel) {
		num = 5
		//低版本兼容，出40张卡片
		if actmdl.IsVersion618Low(c, s.c.Feature, param.MobiApp, param.Build) {
			num = 40
		}
	}
	arg := &actmdl.ResourceOriginReq{FID: mou.Fid, RdbType: confSort.RdbType, Ps: num, Offset: 0, MobiApp: param.MobiApp, Device: param.Device, MustseeType: confSort.MseeType, Mid: param.Mid, Buvid: param.Buvid}
	resources, err := s.editOrigin(c, arg, mou)
	if err != nil {
		log.Error("[FormatEditor] s.resourceOrigin(%+v,%d,%d) error(%+v)", mou, mou.Num, 0, err)
		return nil
	}
	if resources == nil || len(resources.List) == 0 {
		log.Error("[FormatEditor] s.resourceOrigin(%+v,%d,%d) is nil", mou, mou.Num, 0)
		return nil
	}
	//无限feed流
	if mou.IsAttrLast() == api.AttrModuleYes && (confSort.RdbType == api.RDBMustsee || confSort.RdbType == api.RDBChannel) && !actmdl.IsVersion618Low(c, s.c.Feature, param.MobiApp, param.Build) {
		return s.feedEditorJoin(mou, confSort.RdbType)
	}
	return s.EditorJoin(c, resources, mou)
}

func (s *Service) FormatEditor(c context.Context, mou *api.NativeModule, param *actmdl.ParamFormatModule) *actmdl.Item {
	resources, err := s.ResourceAvid(c, mou, mou.Num, 0, param)
	if err != nil {
		log.Error("[FormatEditor] s.ResourceAvid(%+v,%d,%d) error(%+v)", mou, mou.Num, 0, err)
		return nil
	}
	if resources == nil {
		log.Error("[FormatEditor] s.ResourceAvid(%+v,%d,%d) is nil", mou, mou.Num, 0)
		return nil
	}
	item := s.EditorJoin(c, resources, mou)
	return item
}

func (s *Service) FormatProgress(c context.Context, mou *api.NativeModule, mid int64) *actmdl.Item {
	groupID := mou.Width
	if mou.Fid == 0 || groupID == 0 {
		return nil
	}
	rly, err := s.actDao.ActivityProgress(c, mou.Fid, 2, mid, []int64{groupID})
	if err != nil {
		return nil
	}
	group, ok := rly.Groups[groupID]
	if !ok {
		log.Warn("node_group=%+v not found", groupID)
		return nil
	}
	if group == nil || len(group.Nodes) == 0 {
		log.Warn("node_group=%+v is empty", groupID)
		return nil
	}
	progress := &actmdl.Item{}
	progress.FromProgress(mou, group)
	progressModule := &actmdl.Item{}
	progressModule.FromProgressModule(mou, []*actmdl.Item{progress})
	return progressModule
}

func (s *Service) feedEditorJoin(mou *api.NativeModule, RdbType int64) *actmdl.Item {
	list := &actmdl.Item{}
	ext := &actmdl.UrlExt{SortType: int32(RdbType), ConfModuleID: mou.ID, Goto: actmdl.GotoEditOriginModule}
	list.FromEditorModule(mou, ext)
	return list
}

func (s *Service) EditorJoin(c context.Context, req *actmdl.ResourceReply, mou *api.NativeModule) *actmdl.Item {
	if req == nil || len(req.List) == 0 {
		return nil
	}
	list := &actmdl.Item{}
	list.FromEditorModule(mou, nil)
	list.Item = append(list.Item, req.List...)
	return list
}

// FormatSelect .
// nolint:gocognit
func (s *Service) FormatSelect(ctx context.Context, mou *api.NativeModule, commonConf *api.NativePage, select_ *api.Select, arg *actmdl.ParamFormatModule, mid int64) *actmdl.Item {
	if select_ == nil || len(select_.List) == 0 {
		return nil
	}
	var (
		pageIDs, weekIDs  []int64
		defTab, timingTab int64
		nowTime           = time.Now().Unix()
	)
	ext := make(map[int64]*api.MixReason)
	for _, v := range select_.List {
		if v == nil || v.MType != api.MixInlineType || v.ForeignID == 0 || !v.IsOnline() {
			continue
		}
		pageIDs = append(pageIDs, v.ForeignID)
		ext[v.ForeignID] = v.RemarkUnmarshal()
		//寻找默认tab
		if ext[v.ForeignID].DefType == api.DefTypeTimely { //立即生效的时间
			defTab = v.ForeignID
		} else if ext[v.ForeignID].DefType == api.DefTypeTiming { //定时生效的时间
			if ext[v.ForeignID].DStime <= nowTime && ext[v.ForeignID].DEtime > nowTime {
				timingTab = v.ForeignID
			}
		}
		//寻找默认tab
		// 只有index接口才需要下发定位能力
		if arg.FromPage == actmdl.PageFromIndex && ext[v.ForeignID].Type == api.SelectWeek && ext[v.ForeignID].LocationKey != "" {
			weekid, _ := strconv.ParseInt(ext[v.ForeignID].LocationKey, 10, 64)
			if weekid > 0 {
				weekIDs = append(weekIDs, weekid)
			}
		}
	}
	// 默认tab优先级,若立即生效的时间，与定时生效的时间一致，则优先以定时生效的为准
	if timingTab == 0 {
		timingTab = defTab
	}
	//寻找默认tab
	if len(pageIDs) == 0 {
		return nil
	}
	var (
		pagesInfo map[int64]*api.NativePage
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) (e error) {
		pagesInfo, e = s.actDao.NativePages(c, pageIDs)
		if e != nil {
			log.Error("s.actDao.NativePages %v error(%v)", pageIDs, e)
			return
		}
		return nil
	})
	var weekReply map[int64]*selected.SerieFull
	if len(weekIDs) > 0 {
		eg.Go(func(c context.Context) error {
			var e error
			if weekReply, e = s.cdao.BatchPickSerieCache(c, actmdl.WeekStyle, weekIDs); e != nil {
				//降级不下发
				log.Error("s.cdao.BatchPickSerieCache(%v) error(%v)", weekIDs, e)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil
	}
	tmpItem := &actmdl.Item{}
	tmpItem.FormatSelect(mou)
	var (
		currentIndex int32
		hasFind      bool
	)
	for _, v := range pageIDs {
		if val, ok := pagesInfo[v]; !ok || val == nil || !val.IsOnline() || val.Title == "" {
			continue
		}
		eVal := ext[v]
		tmpID := &actmdl.Item{ItemID: v, Title: pagesInfo[v].Title}
		func() {
			if eVal != nil && eVal.LocationKey != "" {
				weekid, e := strconv.ParseInt(eVal.LocationKey, 10, 64)
				if e != nil {
					return
				}
				var shareTitle, shareCaption, img string
				switch eVal.Type {
				case api.SelectWeek:
					if wkVal, k := weekReply[weekid]; !k || wkVal == nil || wkVal.Config == nil {
						return
					}
					img = s.c.Custom.ShareIcon
					shareTitle = weekReply[weekid].Config.ShareSubtitle //分享内容
					shareCaption = weekReply[weekid].Config.ShareTitle  //分享文案
				case api.SelectMiao:
					shareTitle = commonConf.ShareCaption //分享内容
					shareCaption = commonConf.ShareTitle //分享文案
					img = commonConf.ShareImage
				default:
					return
				}
				// 获取每周必看的share数据
				tmpID.Share = &actmdl.Share{
					ShareImage:   img,                      //分享图
					ShareType:    actmdl.ShareTypeActivity, //分享类型
					ShareTitle:   shareTitle,               //分享内容
					ShareCaption: shareCaption,             //分享文案
				}
				if arg.ShareOrigin == actmdl.OriginTab && arg.TabID > 0 && arg.TabModuleID > 0 {
					//"pageid,tabid,tabModuleId,type,id,current_tab"
					tmpID.Share.Sid = fmt.Sprintf("%d,%d,%d,%s,%d,%s", arg.PageID, arg.TabID, arg.TabModuleID, eVal.Type, weekid, eVal.JoinCurrentTab())
					tmpID.Share.ShareOrigin = actmdl.OriginInlineTab
					tmpID.Share.ShareURL = fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d&ts=%d&current_tab=%s", arg.PageID, arg.TabID, arg.TabModuleID, time.Now().Unix(), eVal.JoinCurrentTab())
				} else {
					//"pageid,type,id,current_tab"
					tmpID.Share.Sid = fmt.Sprintf("%d,%s,%d,%s", arg.PageID, eVal.Type, weekid, eVal.JoinCurrentTab())
					tmpID.Share.ShareOrigin = actmdl.SimpleInlineTab
					tmpID.Share.ShareURL = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d&current_tab=%s", arg.PageID, time.Now().Unix(), eVal.JoinCurrentTab()) //分享动态增加时间戳参数
				}
			}
		}()
		tmpItem.Item = append(tmpItem.Item, tmpID)
		//查找默认tab start
		//当【页面URL含有定位参数】与【页面设置默认tab】同时存在时，则优先以页面URL的定位参数为准
		if arg.CurrentTab != "" {
			if eVal != nil && eVal.JoinCurrentTab() == arg.CurrentTab {
				tmpItem.CurrentTabIndex = currentIndex
				hasFind = true
			}
		}
		// 没有指定定位  && 有默认tab
		if !hasFind && v == timingTab {
			tmpItem.CurrentTabIndex = currentIndex
		}
		//查找默认tab end
		currentIndex++
	}
	if len(tmpItem.Item) == 0 {
		return nil
	}
	// 低版本兼容select组件，取出第一个tab页面下对应的三个组件信息，拼接到一级页面上
	var child []*actmdl.Item
	if actmdl.IsSelectLow(ctx, s.c.Feature, arg.MobiApp, arg.Build) {
		func() {
			inlineReq := &actmdl.ParamInlineTab{
				PageID:      tmpItem.Item[0].ItemID,
				Device:      arg.Device,
				VideoMeta:   arg.VideoMeta,
				MobiApp:     arg.MobiApp,
				Platform:    arg.Platform,
				Build:       arg.Build,
				Buvid:       arg.Buvid,
				Offset:      0,
				Ps:          3,
				TfIsp:       arg.TfIsp,
				HttpsUrlReq: arg.HttpsUrlReq,
				FromSpmid:   arg.FromSpmid,
			}
			inlineRly, e := s.InlineTab(ctx, inlineReq, mid)
			if e != nil { //低版本兼容逻辑，错误不处理
				log.Error("s.InlineTa %d error(%v)", inlineReq.PageID, e)
				return
			}
			if inlineRly != nil {
				child = inlineRly.Items
			}
		}()
	}
	var items []*actmdl.Item
	items = append(items, tmpItem)
	// 低版本兼容select组件
	first := &actmdl.Item{}
	first.FromSelectModule(mou, items, child)
	return first
}

// ActShare .
// nolint:gomnd
func (s *Service) ActShare(c context.Context, arg *pb.ActShareReq) (*pb.ActShareReply, error) {
	var (
		stype, shareUrl, currentTab string
		id, pageID                  int64
	)
	strAry := strings.Split(arg.Sid, ",")
	if len(strAry) >= 1 {
		pageID, _ = strconv.ParseInt(strAry[0], 10, 64)
	}
	if arg.ShareOrigin == actmdl.OriginInlineTab {
		//"pageid,tabid,tabModuleId,type,id,current_tab"
		if len(strAry) >= 6 {
			tabid, _ := strconv.ParseInt(strAry[1], 10, 64)
			tabModuleId, _ := strconv.ParseInt(strAry[2], 10, 64)
			stype = strAry[3]
			id, _ = strconv.ParseInt(strAry[4], 10, 64)
			currentTab = strAry[5]
			if tabid > 0 && tabModuleId > 0 {
				shareUrl = fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d&ts=%d&current_tab=%s", pageID, tabid, tabModuleId, time.Now().Unix(), currentTab)
			}
		}
	} else if arg.ShareOrigin == actmdl.SimpleInlineTab {
		//"pageid,type,id,current_tab"
		if len(strAry) >= 4 {
			stype = strAry[1]
			id, _ = strconv.ParseInt(strAry[2], 10, 64)
			currentTab = strAry[3]
			shareUrl = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d&current_tab=%s", pageID, time.Now().Unix(), currentTab) //分享动态增加时间戳参数
		}
	}
	if shareUrl == "" && pageID > 0 {
		shareUrl = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", pageID, time.Now().Unix()) //分享动态增加时间戳参数
	}
	rly := &pb.ActShareReply{
		ShareImage: s.c.Custom.ShareIcon, //分享图
		ShareURL:   shareUrl,
	}
	if id > 0 && stype == api.SelectWeek {
		weekReply, e := s.cdao.PickSerieCache(c, actmdl.WeekStyle, id)
		if e == nil && weekReply != nil && weekReply.Config != nil {
			rly.ShareContent = weekReply.Config.ShareSubtitle
			rly.ShareCaption = weekReply.Config.ShareTitle
			return rly, nil
		}
	}
	//兜底走话题活动逻辑
	if pageID == 0 {
		return nil, ecode.NothingFound
	}
	pages, e := s.actDao.NativePage(c, pageID)
	if e != nil {
		log.Error("s.actDao.NativePage(%d) error(%v)", pageID, e)
		return nil, e
	}
	if pages == nil || !pages.IsTopicAct() {
		return nil, ecode.NothingFound
	}
	rly.ShareContent = pages.ShareTitle
	rly.ShareCaption = pages.Title
	if pages.ShareCaption != "" {
		rly.ShareCaption = pages.ShareCaption
	}
	rly.ShareImage = pages.ShareImage
	return rly, nil
}

func getForeignIDsFromMix(list []*api.NativeMixtureExt) []int64 {
	ids := make([]int64, 0, len(list))
	idsSet := make(map[int64]struct{}, len(list))
	for _, v := range list {
		if v == nil || v.ForeignID <= 0 {
			continue
		}
		if _, ok := idsSet[v.ForeignID]; ok {
			continue
		}
		ids = append(ids, v.ForeignID)
		idsSet[v.ForeignID] = struct{}{}
	}
	return ids
}

// offset: 从0开始
func pagingList(list []int64, offset, ps int) ([]int64, bool) {
	if offset > len(list) {
		offset = len(list)
	}
	end := offset + ps
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], end != len(list)
}

// FormatCarouselImg .
func (s *Service) FormatCarouselImg(c context.Context, mou *api.NativeModule, carousel *api.Carousel) *actmdl.Item {
	if carousel == nil || len(carousel.List) == 0 {
		return nil
	}
	items := make([]*actmdl.Item, 0, len(carousel.List)+1)
	if mou.Meta != "" {
		item := &actmdl.Item{}
		item.FromTitleImage(mou)
		items = append(items, item)
	}
	carouselItem := &actmdl.Item{}
	carouselItem.FromCarouselImg(mou)
	for _, v := range carousel.List {
		if v == nil {
			continue
		}
		ext := &actmdl.CarouselImage{}
		if err := json.Unmarshal([]byte(v.Reason), ext); err != nil {
			log.Error("Fail to unmarshal carouselImgExt, carouselImgExt=%s error=%+v", v.Reason, err)
			continue
		}
		item := &actmdl.Item{}
		item.FromCarouselImgItem(ext)
		carouselItem.Item = append(carouselItem.Item, item)
	}
	//当图片不存在时，不下发组件
	if len(carouselItem.Item) == 0 {
		return nil
	}
	items = append(items, carouselItem)
	ck := &actmdl.Item{}
	ck.FromCarouselImgModule(mou, items)
	return ck
}

// FormatCarouselWord .
func (s *Service) FormatCarouselWord(c context.Context, mou *api.NativeModule, carousel *api.Carousel) *actmdl.Item {
	if carousel == nil || len(carousel.List) == 0 {
		return nil
	}
	items := make([]*actmdl.Item, 0, len(carousel.List))
	carouselItem := &actmdl.Item{}
	carouselItem.FromCarouselWord(mou)
	for _, v := range carousel.List {
		if v == nil {
			continue
		}
		ext := new(struct {
			Content string `json:"content"`
		})
		if err := json.Unmarshal([]byte(v.Reason), ext); err != nil {
			log.Error("Fail to unmarshal carouselWordExt, carouselWordExt=%s error=%+v", v.Reason, err)
			continue
		}
		item := &actmdl.Item{}
		item.FromCarouselWordItem(ext.Content)
		carouselItem.Item = append(carouselItem.Item, item)
	}
	//没有数据不下发组件
	if len(carouselItem.Item) == 0 {
		return nil
	}
	items = append(items, carouselItem)
	ck := &actmdl.Item{}
	ck.FromCarouselWordModule(mou, items)
	return ck
}

func (s *Service) FormatCarouselSource(c context.Context, mou *api.NativeModule, mid int64) *actmdl.Item {
	list, err := s.carouselImgSourceData(c, mou, mid)
	if err != nil || len(list) == 0 {
		return nil
	}
	items := make([]*actmdl.Item, 0, len(list)+1)
	if mou.Meta != "" {
		item := &actmdl.Item{}
		item.FromTitleImage(mou)
		items = append(items, item)
	}
	carouselItem := &actmdl.Item{}
	carouselItem.FromCarouselImg(mou)
	for _, v := range list {
		if v == nil {
			continue
		}
		item := &actmdl.Item{}
		item.FromCarouselImgItem(v)
		carouselItem.Item = append(carouselItem.Item, item)
	}
	//当图片不存在时，不下发组件
	if len(carouselItem.Item) == 0 {
		return nil
	}
	items = append(items, carouselItem)
	ck := &actmdl.Item{}
	ck.FromCarouselImgModule(mou, items)
	return ck
}

// FormatIcon .
func (s *Service) FormatIcon(c context.Context, mou *api.NativeModule, icon *api.Icon) *actmdl.Item {
	if icon == nil || len(icon.List) == 0 {
		return nil
	}
	items := make([]*actmdl.Item, 0, len(icon.List))
	iconItem := &actmdl.Item{Goto: actmdl.GotoIcon}
	for _, v := range icon.List {
		if v == nil {
			continue
		}
		ext := &actmdl.IconRemark{}
		if err := json.Unmarshal([]byte(v.Reason), ext); err != nil {
			log.Error("Fail to unmarshal iconExt, iconExt=%s error=%+v", v.Reason, err)
			continue
		}
		item := &actmdl.Item{}
		item.FromIconExt(ext)
		iconItem.Item = append(iconItem.Item, item)
	}
	if len(iconItem.Item) == 0 {
		return nil
	}
	items = append(items, iconItem)
	ck := &actmdl.Item{}
	ck.FromIcon(mou, items)
	return ck
}

func extractDimension(click *api.NativeClick) (actapi.GetReserveProgressDimension, error) {
	tmpDimension, err := strconv.ParseInt(click.FinishedImage, 10, 64)
	if err != nil {
		log.Error("Fail to parse dimension, dimension=%+v err=%+v", click.FinishedImage, err)
		return 0, err
	}
	return actapi.GetReserveProgressDimension(tmpDimension), nil
}

func (s *Service) carouselImgSourceData(c context.Context, mou *api.NativeModule, mid int64) ([]*actmdl.CarouselImage, error) {
	_, sourceType, err := extractConf4SourcePattern(mou.ConfSort)
	if err != nil {
		return nil, err
	}
	var carouselImgs []*actmdl.CarouselImage
	switch sourceType {
	case api.SourceTypeActUp:
		upList, err := s.upListFromModule(c, mou, mid, 8)
		if err != nil {
			return nil, err
		}
		carouselImgs = upList2CarouselImg(upList, mou)
	}
	return carouselImgs, nil
}

func (s *Service) rcmdSourceData(c context.Context, mou *api.NativeModule, mid int64) ([]int64, error) {
	_, sourceType, err := extractConf4SourcePattern(mou.ConfSort)
	if err != nil {
		return nil, err
	}
	var mids []int64
	switch sourceType {
	case api.SourceTypeActUp:
		upList, err := s.upListFromModule(c, mou, mid, 40)
		if err != nil {
			return nil, err
		}
		mids = midsFromUpList(upList)
	}
	return mids, nil
}

func (s *Service) upListFromModule(c context.Context, mou *api.NativeModule, mid, pn int64) ([]*actapi.UpListItem, error) {
	if mou.Fid == 0 {
		return []*actapi.UpListItem{}, nil
	}
	sortType, _, err := extractConf4SourcePattern(mou.ConfSort)
	if err != nil {
		return nil, err
	}
	if sortType == "" {
		sortType = api.SortTypeCtime
	}
	upList, err := s.actDao.UpList(c, mou.Fid, 1, pn, mid, sortType)
	if err != nil {
		return nil, err
	}
	return upList.List, nil
}

func (s *Service) rankIcon(c context.Context, id, num int64) (rcmRly map[int]string) {
	rcmRly = make(map[int]string)
	mixArg := &api.ModuleMixExtReq{ModuleID: id, Ps: num, MType: api.MixRankIcon}
	mixIcon, e := s.actDao.ModuleMixExt(c, mixArg)
	if e != nil { //降级处理
		log.Error(" s.actDao.ModuleMixExt(%v) error(%v)", mixArg, e)
		return
	}
	if mixIcon == nil || len(mixIcon.List) == 0 {
		return
	}
	i := 0
	for _, v := range mixIcon.List {
		if v == nil {
			continue
		}
		remark := v.RemarkUnmarshal()
		if remark.Image == "" {
			continue
		}
		rcmRly[i] = remark.Image
		i++
	}
	return
}

func (s *Service) rankListFromModule(c context.Context, mou *api.NativeModule, mid int64) *actmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	num := mou.Num
	maxPs := int64(10)
	if num > maxPs {
		num = maxPs
	}
	eg := errgroup.WithContext(c)
	var rcmRly map[int]string
	eg.Go(func(ctx context.Context) error {
		//获取icon
		rcmRly = s.rankIcon(ctx, mou.ID, num)
		return nil
	})
	var rly *actapi.RankResultResp
	eg.Go(func(ctx context.Context) (e error) {
		if rly, e = s.actDao.RankResult(ctx, mou.Fid, 1, num); e != nil {
			log.Error("s.actDao.RankResult(%d) error(%v)", mou.Fid, e)

		}
		return
	})
	err := eg.Wait()
	if err != nil {
		return nil
	}
	if rly == nil || len(rly.List) == 0 {
		return nil
	}
	var fids []int64
	for _, v := range rly.List {
		if v == nil || v.Account == nil || v.ObjectType != 1 {
			continue
		}
		fids = append(fids, v.Account.MID)
	}
	//获取关注关系
	var followRly map[int64]*relationgrpc.FollowingReply
	if mid > 0 {
		if followRly, err = s.reldao.RelationsGRPC(c, mid, fids); err != nil { //错误降级
			log.Error(" s.reldao.RelationsGRPC(%d,%v) error(%v)", mid, fids, err)
		}
	}
	var (
		items []*actmdl.Item
		j     = 0
	)
	display := mou.IsAttrDisplayRecommend() == api.AttrModuleYes
	for _, reVal := range rly.List {
		if reVal == nil || reVal.Account == nil || reVal.ObjectType != 1 {
			continue
		}
		ext := &actmdl.ClickExt{}
		if rel, ok := followRly[reVal.Account.MID]; ok {
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			if rel.Attribute == 2 || rel.Attribute == 6 {
				ext.IsFollow = true
			}
		}
		itemTep := &actmdl.Item{}
		rcm := rcmRly[j]
		j++
		itemTep.FromRecommendRankExt(reVal, ext, display, rcm)
		items = append(items, itemTep)
	}
	if len(items) == 0 {
		return nil
	}
	ck := &actmdl.Item{}
	var lastItems []*actmdl.Item
	if mou.Meta != "" {
		tmpImage := &actmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	lastItems = append(lastItems, items...)
	ck.FromRecommend(mou, lastItems)
	return ck
}

func (s *Service) FormatHoverButton(c context.Context, mou *api.NativeModule, mid int64) *actmdl.Item {
	if mou.ConfSort == "" {
		return nil
	}
	confSort := &api.ConfSort{}
	if err := json.Unmarshal([]byte(mou.ConfSort), confSort); err != nil {
		log.Error("Fail to unmarshal confSort of hoverButton, confSort=%+v error=%+v", mou.ConfSort, err)
		return nil
	}
	var item *actmdl.Item
	switch confSort.BtType {
	case api.BtTypeAppoint:
		item = s.formatHoverAppointOrigin(c, mou, confSort, mid)
	case api.BtTypeActProject:
		item = s.formatHoverActProject(c, mou, confSort, mid)
	case api.BtTypeLink:
		item = s.formatHoverLink(mou)
	default:
		log.Warn("unknown button_type=%+v", confSort.BtType)
		return nil
	}
	hoverButton := &actmdl.Item{}
	hoverButton.FromHoverButton(mou, []*actmdl.Item{item}, confSort)
	return hoverButton
}

func (s *Service) formatHoverAppointOrigin(c context.Context, mou *api.NativeModule, confSort *api.ConfSort, mid int64) *actmdl.Item {
	tip := &actmdl.TipCancel{}
	tip.FromTip(confSort.Hint)
	ext := &actmdl.ClickExt{
		Goto:         actmdl.GotoClickAppointment,
		FID:          mou.Fid,
		Tip:          tip,
		ActionType:   actmdl.ActionTypeSub,
		UnactionType: actmdl.ActionTypeUnsub,
	}
	func() {
		if mid == 0 {
			return
		}
		rly, err := s.actDao.ReserveFollowings(c, mid, []int64{mou.Fid})
		if err != nil {
			return
		}
		if data, ok := rly[mou.Fid]; ok && data != nil {
			ext.IsFollow = data.IsFollow
		}
	}()
	return &actmdl.Item{
		Goto:     actmdl.GotoClickButton,
		Image:    mou.TitleColor,
		UnImage:  mou.FontColor,
		ClickExt: ext,
	}
}

func (s *Service) formatHoverActProject(c context.Context, mou *api.NativeModule, confSort *api.ConfSort, mid int64) *actmdl.Item {
	tip := &actmdl.TipCancel{}
	tip.FromTip(confSort.Hint)
	ext := &actmdl.ClickExt{
		Goto:         actmdl.GotoClickAttention,
		FID:          mou.Fid,
		Tip:          tip,
		ActionType:   actmdl.ActionTypeSub,
		UnactionType: actmdl.ActionTypeUnsub,
	}
	func() {
		if mid == 0 {
			return
		}
		rly, err := s.actDao.ActRelationInfo(c, mou.Fid, mid)
		if err != nil || rly.ReserveItems == nil {
			return
		}
		if rly.ReserveItems.State == 1 {
			ext.IsFollow = true
		}
	}()
	return &actmdl.Item{
		Goto:     actmdl.GotoClickButton,
		Image:    mou.TitleColor,
		UnImage:  mou.FontColor,
		ClickExt: ext,
	}
}

func (s *Service) formatHoverLink(mou *api.NativeModule) *actmdl.Item {
	return &actmdl.Item{
		Goto:     actmdl.GotoClickButtonV3,
		Image:    mou.MoreColor,
		URI:      mou.Colors,
		ClickExt: &actmdl.ClickExt{Goto: actmdl.GotoClickRedirect, ActionType: actmdl.ActionTypeJump},
	}
}

func extractProgressParamFromClick(click *api.NativeClick) (sid, groupID int64) {
	sid = click.ForeignID
	if click.Tip == "" {
		return
	}
	tip := &api.ClickTip{}
	if err := json.Unmarshal([]byte(click.Tip), tip); err != nil {
		log.Error("Fail to unmarshal clickTip, clickTip=%+v error=%+v", click.Tip, err)
		return
	}
	groupID = tip.GroupId
	return
}

func canShowTab(mobiAPP string, build int64) bool {
	if (mobiAPP == "android" && build >= 6020000) || (mobiAPP == "iphone" && build >= 10030) || (mobiAPP == "ipad" && build >= 10030) {
		return true
	}
	return false
}

// 首页tab低版本兼容-浮层接口不下发
func menuLayerInterface(c context.Context, featureCfg *conf.Feature, mobiApp string, build int64, page *api.NativePage) bool {
	if actmdl.IsVersion615Low(c, featureCfg, mobiApp, build) && page.IsMenuAct() {
		return false
	}
	return true
}

func menuPageCompat(c context.Context, featureCfg *conf.Feature, acts *api.Click, params *actmdl.ParamFormatModule, page *api.NativePage) {
	if !actmdl.IsVersion615Low(c, featureCfg, params.MobiApp, params.Build) || !page.IsMenuAct() {
		return
	}
	if acts == nil {
		return
	}
	areas := make([]*api.NativeClick, 0, len(acts.Areas))
	for _, v := range acts.Areas {
		if v.IsLayerImage() {
			continue
		}
		if v.IsLayerLink() {
			if v.Ext == "" {
				continue
			}
			v.Type = api.Redirect
			ext := &api.ClickExt{}
			if err := json.Unmarshal([]byte(v.Ext), ext); err != nil {
				log.Error("Fail to unmarshal clickExt, ext=%+v error=%+v", v.Ext, err)
				continue
			}
			v.OptionalImage = ext.ButtonImage
		}
		areas = append(areas, v)
	}
	acts.Areas = areas
}

func calculateProgress(progress, interveNum int64) int64 {
	total := progress + interveNum
	if total < 0 {
		total = 0
	}
	return total
}

func extractExt4ClickInterface(ext string) (string, error) {
	if ext == "" {
		return "", nil
	}
	confExt := &api.ClickExt{}
	if err := json.Unmarshal([]byte(ext), confExt); err != nil {
		log.Error("Fail to unmarshal confExt, ext=%+v error=%+v", ext, err)
		return "", err
	}
	return confExt.Style, nil
}

func extractProgNum(group *actapi.ActivityProgressGroup, nodeID int64) (total, targetNum int64) {
	total = group.Total
	for _, node := range group.Nodes {
		if node.Nid == nodeID {
			targetNum = node.Val
			return
		}
	}
	return
}

func extractConf4SourcePattern(raw string) (sort, source string, err error) {
	if raw == "" {
		return "", "", nil
	}
	confSort := &api.ConfSort{}
	if err := json.Unmarshal([]byte(raw), confSort); err != nil {
		log.Error("Fail to unmarshal confSort, confSort=%+v error=%+v", raw, err)
		return "", "", err
	}
	return confSort.SortType, confSort.SourceType, nil
}

func midsFromUpList(upList []*actapi.UpListItem) []int64 {
	mids := make([]int64, 0, len(upList))
	for _, v := range upList {
		if v == nil || v.Account == nil {
			continue
		}
		mids = append(mids, v.Account.Mid)
	}
	return mids
}

func upList2CarouselImg(upList []*actapi.UpListItem, mou *api.NativeModule) []*actmdl.CarouselImage {
	images := make([]*actmdl.CarouselImage, 0, len(upList))
	for _, v := range upList {
		if v == nil || v.Content == nil {
			continue
		}
		image := &actmdl.CarouselImage{
			ImgUrl:      v.Content.Image,
			RedirectUrl: v.Content.Link,
			Length:      mou.Length,
			Width:       mou.Width,
		}
		images = append(images, image)
	}
	return images
}

func trans2RecommendPB(mids []int64) *api.Recommend {
	recommend := &api.Recommend{
		List: make([]*api.NativeMixtureExt, 0, len(mids)),
	}
	for _, mid := range mids {
		recommend.List = append(recommend.List, &api.NativeMixtureExt{ForeignID: mid})
	}
	return recommend
}

func setUnlockProgReq(click *api.NativeClick, progReqs map[int64][]int64) {
	if click.Ext == "" {
		return
	}
	clickExt := &api.ClickExt{}
	if err := json.Unmarshal([]byte(click.Ext), clickExt); err != nil {
		log.Error("Fail to unmarshal clickExt, clickExt=%+v error=%+v", click.Ext, err)
		return
	}
	if clickExt.DisplayMode == api.NeedUnLock && clickExt.UnlockCondition == api.UnLockOrder {
		if clickExt.Sid == 0 || clickExt.GroupId == 0 {
			return
		}
		progReqs[clickExt.Sid] = append(progReqs[clickExt.Sid], clickExt.GroupId)
	}
}

func reachUnlockCondition(click *api.NativeClick, progRlys map[int64]*actapi.ActivityProgressReply) bool {
	if click.Ext == "" {
		return true
	}
	ext := &api.ClickExt{}
	if err := json.Unmarshal([]byte(click.Ext), ext); err != nil {
		log.Error("Fail to unmarshal clickExt, clickExt=%+v error=%+v", click.Ext, err)
		return false
	}
	if ext.DisplayMode != api.NeedUnLock {
		return true
	}
	if ext.UnlockCondition == api.UnLockTime {
		return time.Now().Unix() >= ext.Stime
	}
	if ext.UnlockCondition == api.UnLockOrder {
		progRly, ok := progRlys[ext.Sid]
		if !ok || progRly == nil || len(progRly.Groups) == 0 {
			return false
		}
		group, ok := progRly.Groups[ext.GroupId]
		if !ok || group == nil || len(group.Nodes) == 0 {
			return false
		}
		for _, node := range group.Nodes {
			if ext.NodeId == node.Nid {
				return group.Total >= node.Val
			}
		}
	}
	return false
}

func finalScore(target *scoregrpc.ScoreTarget) string {
	if target.GetShowFlag() == 1 {
		return "暂无评分"
	}
	if fs := target.GetFixScore(); fs != "" && fs != "0" && fs != "0.0" {
		return target.GetFixScore()
	}
	if target.GetTargetScore() == "0" || target.GetTargetScore() == "0.0" {
		return "暂无评分"
	}
	return target.GetTargetScore()
}

func compatIosHoverButton() *actmdl.Item {
	return &actmdl.Item{
		Goto: actmdl.GotoBottomButton,
		Ukey: "Base_bottom_click",
		Item: []*actmdl.Item{
			{
				Goto: actmdl.GotoClickBack,
			},
		},
	}
}

func compatIpadBpu2021(items []*actmdl.Item) []*actmdl.Item {
	list := make([]*actmdl.Item, 0, len(items))
	var cnt int64
	for _, item := range items {
		if item.Goto == actmdl.GotoCarouselImgModule {
			cnt++
			if cnt > 1 {
				continue
			}
		}
		list = append(list, item)
	}
	return list
}

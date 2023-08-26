package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/app/app-svr/topic/card/proto/dyn_handler"
	"go-gateway/app/app-svr/topic/interface/internal/model"
	"go-gateway/pkg/idsafe/bvid"

	"go-common/library/sync/errgroup.v2"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	cmtGrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynactivitygrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/activity"
	dyndrawrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/draw"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynamicrevs "git.bilibili.co/bapis/bapis-go/dynamic/service/revs"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"

	"github.com/pkg/errors"
)

func (s *Service) dynBriefs(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, params []*topiccardmodel.DynMetaCardListParam) (*model.DynListRes, error) {
	args := convertToDynBriefsParams(params, general)
	data, err := s.dynGRPC.DynBriefs(dynSchemaCtx.Ctx, args)
	if err != nil {
		return nil, err
	}
	ret := &model.DynListRes{}
	for _, item := range data.Dyns {
		dynTmp := &dynmdlV2.Dynamic{}
		dynTmp.FromDynamic(item)
		ret.Dynamics = append(ret.Dynamics, dynTmp)
	}
	reconstructDynSchemaContext(dynSchemaCtx, params)
	return ret, nil
}

// 取动态物料，逻辑暂时与app-dynamic保持一致
//
//nolint:gocognit
func (s *Service) getMaterial(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, dynList *model.DynListRes) (*topiccardmodel.DynSchemaCtx, error) {
	var (
		aidm                    = make(map[int64]map[int64]struct{})
		storyAidm               = make(map[int64]map[int64]struct{})
		midm                    = make(map[int64]struct{})
		epIDm                   = make(map[int64]struct{})
		dynIDm                  = make(map[int64]struct{})
		drawIDm                 = make(map[int64]struct{})
		articleIDm              = make(map[int64]struct{})
		commonIDm               = make(map[int64]struct{})
		topicIDm                = make(map[int64]struct{})
		attachedPromo           = make(map[int64]int64)
		likeIDm                 = map[string][]*dynmdlV2.ThumbsRecord{}
		replyIDm                = make(map[string]struct{})
		dynamicActivityArgs     = make(map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo)
		additionalTopic         = make(map[int64][]*dynmdlV2.Topic)
		storyArchive            = make(map[int64]*archivegrpc.ArcPlayer)
		dynArchive              = make(map[int64]*archivegrpc.ArcPlayer)
		voteIDm                 = make(map[int64]struct{})
		additionalUpActivityIDm = make(map[int64]struct{})
		// 预约参数
		reservesDynIDm   = make(map[int64]struct{})
		additionalUpIDm  = make(map[int64]struct{})
		uplivemid        = make(map[int64][]string)
		liveAidm         = make(map[int64]struct{})
		upAdditionalAids []int64
		relationReply    *activitygrpc.UpActReserveRelationInfoReply
		pugvIDm          = make(map[int64]struct{})           // 课程附加卡
		playCountIDs     = make(map[int64]int64)              // 在线人数
		premiereAidm     = make(map[int64]map[int64]struct{}) //首映aid
		goods            = make(map[int64]*bcgmdl.GoodsParams)
	)
	ctx := dynSchemaCtx.Ctx
	ret := &dynmdlV2.DynamicContext{}
	// 聚合物料ID
	for _, dyn := range dynList.Dynamics {
		s.getMaterialIDs(dyn, general, ret, aidm, storyAidm, midm, epIDm, dynIDm, drawIDm, articleIDm, commonIDm, additionalUpIDm, pugvIDm, likeIDm, replyIDm, dynamicActivityArgs, additionalTopic, voteIDm, additionalUpActivityIDm, premiereAidm, goods)
		if dyn.IsForward() {
			s.getMaterialIDs(dyn.Origin, general, ret, aidm, storyAidm, midm, epIDm, dynIDm, drawIDm, articleIDm, commonIDm, additionalUpIDm, pugvIDm, likeIDm, replyIDm, dynamicActivityArgs, additionalTopic, voteIDm, additionalUpActivityIDm, premiereAidm, goods)
		}
	}
	// 并发请求物料
	var midrw = sync.RWMutex{}
	eg := errgroup.WithCancel(ctx)
	if len(epIDm) != 0 {
		var (
			epIDsInt32 []int32
			epIDs      []int64
		)
		for epid := range epIDm {
			epIDsInt32 = append(epIDsInt32, int32(epid))
			epIDs = append(epIDs, epid)
		}
		// PGC详情
		eg.Go(func(ctx context.Context) error {
			res, err := s.getEpList(ctx, epIDsInt32, general)
			if err != nil {
				log.Error("getEpList mid(%v) getEpList(%v), err %v", general.Mid, epIDsInt32, err)
				return nil
			}
			for _, pgc := range res {
				if pgc == nil {
					continue
				}
				if replyID := dynmdlV2.GetPGCReplyID(pgc); replyID != "" {
					replyIDm[replyID] = struct{}{}
				}
				if likeParam, likeType, isLike := dynmdlV2.GetPGCLikeID(pgc); isLike {
					likeIDm[likeType] = append(likeIDm[likeType], likeParam)
				}
			}
			ret.ResPGC = res
			return nil
		})
		// 追番附加卡
		eg.Go(func(ctx context.Context) error {
			res, err := s.pgcDynGRPC.FollowCard(ctx, makeFollowCardParam(epIDs, general))
			if err != nil {
				log.Error("getMaterial mid(%v) FollowCard(%v), err %v", general.Mid, epIDs, err)
				return nil
			}
			ret.ResAdditionalOGV = res.Card
			return nil
		})
	}
	if len(midm) != 0 {
		var mids []int64
		for mid := range midm {
			mids = append(mids, mid)
		}
		// 装扮信息
		eg.Go(func(ctx context.Context) error {
			res, err := s.getDecorateCards(ctx, mids)
			if err != nil {
				log.Error("getMaterial mid(%v) getDecorateCards(%v), err %v", general.Mid, mids, err)
				return nil
			}
			ret.ResMyDecorate = res
			return nil
		})
		// 直播信息
		eg.Go(func(ctx context.Context) error {
			live, playURl, err := s.liveInfos(ctx, mids, general)
			if err != nil {
				log.Error("getMaterial mid(%v) liveInfos(%v), err %v", general.Mid, mids, err)
				return nil
			}
			ret.ResUserLive = live
			ret.ResUserLivePlayUrl = playURl
			return nil
		})
	}
	// 动态文案(纯文字卡、转发卡等)
	if len(dynIDm) > 0 {
		var dynIDs []int64
		for id := range dynIDm {
			dynIDs = append(dynIDs, id)
		}
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynGRPC.ListWordText(ctx, &dyngrpc.WordTextReq{Uid: general.Mid, Rids: dynIDs})
			if err != nil {
				log.Error("getMaterial mid(%v) ListWordText(%v), error %v", general.Mid, dynIDs, err)
				return nil
			}
			ret.ResWords = res.GetContent()
			return nil
		})
	}
	// 图文动态
	if len(drawIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			drawIDs := make([]int64, 0, len(drawIDm))
			for id := range drawIDm {
				drawIDs = append(drawIDs, id)
			}
			res, err := s.drawDetails(ctx, general, drawIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) drawDetails(%v), err %v", general.Mid, drawIDs, err)
				return nil
			}
			ret.ResDraw = res
			return nil
		})
	}
	// 专栏
	if len(articleIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var articleIDs []int64
			for id := range articleIDm {
				articleIDs = append(articleIDs, id)
			}
			res, err := s.articleGRPC.ArticleMetas(ctx, &articlegrpc.ArticleMetasReq{Ids: articleIDs})
			if err != nil {
				log.Error("s.articleGRPC.ArticleMetas mid(%+v) articleIDs(%+v), err=%+v", general.Mid, articleIDs, err)
				return nil
			}
			ret.ResArticle = res.GetRes()
			return nil
		})
	}
	// 通用卡
	if len(commonIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var commonIDs []int64
			for id := range commonIDm {
				commonIDs = append(commonIDs, id)
			}
			res, err := s.DynamicCommonInfos(ctx, commonIDs)
			if err != nil {
				log.Error("s.DynamicCommonInfos mid=%+v commonIDs=%+v, err=%+v", general.Mid, commonIDs, err)
				return nil
			}
			for _, re := range res {
				if re == nil || re.User == nil {
					continue
				}
				midrw.Lock()
				midm[re.User.UID] = struct{}{}
				midrw.Unlock()
			}
			ret.ResCommon = res
			return nil
		})
	}
	// 付费
	if len(pugvIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var cheeseIDs []int64
			for cheeseID := range pugvIDm {
				cheeseIDs = append(cheeseIDs, cheeseID)
			}
			cheeseRes, err := s.AdditionalCheese(ctx, cheeseIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) AdditionalCheese(%v), err %+v", general.Mid, cheeseIDs, err)
				return nil
			}
			ret.ResPUgv = cheeseRes
			return nil
		})
	}
	// 投票
	if len(voteIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var voteIDs []int64
			for id := range voteIDm {
				voteIDs = append(voteIDs, id)
			}
			res, err := s.votes(ctx, general.Mid, voteIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) voteIDs(%+v), err %+v", general.Mid, voteIDs, err)
				return nil
			}
			ret.ResVote = res
			return nil
		})
	}
	//商品
	if len(goods) > 0 {
		res := make(map[int64]map[int]map[string]*bcgmdl.GoodsItem)
		rw := sync.RWMutex{}
		for _, good := range goods {
			var goodParam = new(bcgmdl.GoodsParams)
			*goodParam = *good
			eg.Go(func(ctx context.Context) error {
				goodsDetail, err := s.goodsDetails(ctx, goodParam)
				if err != nil || goodsDetail == nil {
					log.Error("GoodsDetials(%v), error(%+v)", goodParam.DynamicID, err)
					return nil
				}
				rw.Lock()
				res[goodParam.DynamicID] = goodsDetail
				rw.Unlock()
				ret.ResGood = res
				return nil
			})
		}
	}
	// 帮推
	if len(dynamicActivityArgs) != 0 {
		eg.Go(func(ctx context.Context) error {
			// 动态接口拉取绑定tag
			var dynAttachedPromoInfos []*dynactivitygrpc.DynamicAttachedPromoInfo
			for _, tmpDynamicActivityArgs := range dynamicActivityArgs {
				dynAttachedPromoInfos = append(dynAttachedPromoInfos, tmpDynamicActivityArgs)
			}
			resTmp, err := s.dynamicAttachedPromo(ctx, dynAttachedPromoInfos)
			if err != nil {
				log.Error("getMaterialmid(%v) dynamicActivityArgs(%+v), err(%+v)", general.Mid, dynamicActivityArgs, err)
				return nil
			}
			for _, value := range resTmp {
				if value != nil && value.TagId != 0 {
					attachedPromo[value.DynamicId] = value.TagId
					topicIDm[value.TagId] = struct{}{}
				}
			}
			ret.ResAttachedPromo = attachedPromo
			return nil
		})
	}
	// UP主预约卡信息
	if len(additionalUpIDm) > 0 {
		var additionalUpIDs []int64
		for id := range additionalUpIDm {
			additionalUpIDs = append(additionalUpIDs, id)
		}
		eg.Go(func(ctx context.Context) (err error) {
			relationReply, err = s.actClient.UpActReserveRelationInfo(ctx, &activitygrpc.UpActReserveRelationInfoReq{
				Mid: general.Mid, Sids: additionalUpIDs, From: activitygrpc.UpCreateActReserveFrom_FromDynamic,
			})
			if err != nil {
				log.Error("s.actClient.UpActReserveRelationInfo err=%+v", err)
				return nil
			}
			for k, v := range relationReply.List {
				if ret.ResUpActRelationInfo == nil {
					ret.ResUpActRelationInfo = map[int64]*activitygrpc.UpActReserveRelationInfo{}
				}
				ret.ResUpActRelationInfo[k] = v
				// nolint:exhaustive
				switch v.Type {
				case activitygrpc.UpActReserveRelationType_Archive: // 稿件
					aid, _ := strconv.ParseInt(v.Oid, 10, 64)
					if aid > 0 {
						upAdditionalAids = append(upAdditionalAids, aid)
					}
				case activitygrpc.UpActReserveRelationType_Live: // 直播
					liveID, _ := strconv.ParseInt(v.Oid, 10, 64)
					if liveID > 0 {
						// 直播预约卡
						if uplivemid == nil {
							uplivemid = make(map[int64][]string)
						}
						uplivemid[v.Upmid] = append(uplivemid[v.Upmid], v.Oid)
					}
				case activitygrpc.UpActReserveRelationType_Premiere: // 首映
					aid, _ := strconv.ParseInt(v.Oid, 10, 64)
					if aid > 0 {
						premiereAidm[aid] = map[int64]struct{}{}
					}
				}
				dynID, _ := strconv.ParseInt(v.DynamicId, 10, 64)
				midrw.Lock()
				midm[v.Upmid] = struct{}{}
				reservesDynIDm[dynID] = struct{}{}
				midrw.Unlock()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// aids聚合到一个map做去重复
	for _, aid := range upAdditionalAids {
		if _, ok := aidm[aid]; !ok {
			aidm[aid] = make(map[int64]struct{})
		}
	}
	for k, v := range premiereAidm {
		aidm[k] = v
		playCountIDs[k] = 0
		for cid := range v {
			playCountIDs[k] = cid
		}
	}
	for _, channel := range ret.ResSearchChannels {
		if channel == nil {
			continue
		}
		for _, video := range channel.GetVideoCards() {
			if video.GetRid() == 0 {
				continue
			}
			aidm[video.GetRid()] = make(map[int64]struct{})
		}
	}
	/*
		第二级物料获取
	*/
	eg2 := errgroup.WithCancel(ctx)
	// 直播预约卡
	if len(uplivemid) > 0 {
		eg2.Go(func(ctx context.Context) error {
			res, err := s.SessionInfo(ctx, uplivemid, general)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResLiveSessionInfo = res
			// 回播aid
			for k, v := range res {
				if info, ok := v.SessionInfoPerLive[k]; ok {
					// 已关播有回放
					if info.Status == 2 && info.Bvid != "" {
						aid, _ := bvid.BvToAv(info.Bvid)
						if aid > 0 {
							liveAidm[aid] = struct{}{}
						}
					}
				}
			}
			return nil
		})
	}
	// 稿件详情
	if len(aidm) != 0 {
		var aids []*archivegrpc.PlayAv
		for aid, cidm := range aidm {
			var ap = &archivegrpc.PlayAv{Aid: aid}
			for cid := range cidm {
				ap.PlayVideos = append(ap.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
			}
			aids = append(aids, ap)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.arcsPlayer(ctx, aids, true, "")
			if err != nil {
				log.Error("getMaterial mid(%v) aids(%+v), err(%+v)", general.Mid, aids, err)
				return nil
			}
			dynArchive = res
			// 动态服务端没有返回合集UP信息 需要回填再获取
			for _, arc := range res {
				if arc != nil && arc.Arc != nil && arc.Arc.Author.Mid != 0 {
					midrw.Lock()
					midm[arc.Arc.Author.Mid] = struct{}{}
					midrw.Unlock()
				}
			}
			return nil
		})
	}
	// 预约卡的动态信息
	if len(reservesDynIDm) > 0 {
		var dynIDs []int64
		for id := range reservesDynIDm {
			dynIDs = append(dynIDs, id)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.dynGRPC.DynSimpleInfos(ctx, &dyngrpc.DynSimpleInfosReq{DynIds: dynIDs})
			if err != nil {
				log.Error("DynSimpleInfos mid(%+v) dynIDs(%+v) dynIDs, err(%+v)", general.Mid, dynIDs, err)
				return nil
			}
			ret.ResDynSimpleInfos = res.DynSimpleInfos
			return nil
		})
	}
	if len(storyAidm) != 0 {
		var aids []*archivegrpc.PlayAv
		for aid, cidm := range storyAidm {
			var ap = &archivegrpc.PlayAv{Aid: aid}
			for cid := range cidm {
				ap.PlayVideos = append(ap.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
			}
			aids = append(aids, ap)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.arcsPlayer(ctx, aids, true, "story")
			if err != nil {
				log.Error("getMaterial mid(%v) aids(%+v), err(%+v)", general.Mid, aids, err)
				return nil
			}
			storyArchive = res
			// 动态服务端没有返回合集UP信息 需要回填再获取
			for _, arc := range res {
				if arc != nil && arc.Arc != nil && arc.Arc.Author.Mid != 0 {
					midrw.Lock()
					midm[arc.Arc.Author.Mid] = struct{}{}
					midrw.Unlock()
				}
			}
			return nil
		})
	}
	// 评论外露和计数
	if len(replyIDm) != 0 {
		eg2.Go(func(ctx context.Context) error {
			var replyIDs []string
			for replyID := range replyIDm {
				replyIDs = append(replyIDs, replyID)
			}
			cmtRes, err := s.cmtGrpc.DynamicFeed(ctx, &cmtGrpc.DynamicFeedReq{
				Ids:                      replyIDs,
				Mid:                      general.Mid,
				Buvid:                    general.GetBuvid(),
				From:                     cmtGrpc.DynamicFeedReq_TOPIC,
				IdsNeedTopicCreatorReply: makeIdsNeedTopicCreatorReply(dynSchemaCtx.OwnerAppear, replyIDs),
			})
			if err != nil {
				log.Error("getMaterial mid(%v) DynamicFeed(%v), err %v", general.Mid, replyIDs, err)
				return nil
			}
			ret.ResReply = cmtRes.Feed
			for _, reply := range cmtRes.Feed {
				for _, item := range reply.Replies {
					midrw.Lock()
					midm[item.Mid] = struct{}{}
					midrw.Unlock()
				}
			}
			return nil
		})
	}
	// 用户与动态的点赞关系
	if len(likeIDm) != 0 {
		eg2.Go(func(ctx context.Context) error {
			res, err := s.multiStats(ctx, general.Mid, likeIDm)
			if err != nil {
				log.Error("getMaterial mid(%v) multiStats(%v), err %v", general.Mid, likeIDm, err)
				return nil
			}
			ret.ResLike = res.Business
			return nil
		})
		// 话题创建者点赞信息
		if dynSchemaCtx.TopicCreatorMid > 0 {
			eg2.Go(func(ctx context.Context) error {
				res, err := s.multiStats(ctx, dynSchemaCtx.TopicCreatorMid, likeIDm)
				if err != nil {
					log.Error("getMaterial mid(%v) multiStats(%v), err %v", dynSchemaCtx.TopicCreatorMid, likeIDm, err)
					return nil
				}
				dynSchemaCtx.TopicCreatorLike = res.Business
				return nil
			})
		}
	}
	// 附加卡-帮推-活动信息
	if len(topicIDm) > 0 {
		var tagIDs []int64
		for tagid := range topicIDm {
			tagIDs = append(tagIDs, tagid)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.natInfoFromForeign(ctx, tagIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) natInfoFromForeign(%v), err %v", general.Mid, tagIDs, err)
				return nil
			}
			ret.ResActivity = res
			return nil
		})
	}
	if err := eg2.Wait(); err != nil {
		return nil, err
	}
	// 稿件聚合
	if ret.ResArchive == nil {
		ret.ResArchive = map[int64]*archivegrpc.ArcPlayer{}
	}
	for k, v := range storyArchive {
		ret.ResArchive[k] = v
	}
	for k, v := range dynArchive {
		ret.ResArchive[k] = v
	}
	/*
		第三级调用
	*/
	eg3 := errgroup.WithCancel(ctx)
	if len(midm) > 0 {
		var mids []int64
		if dynSchemaCtx.TopicCreatorMid > 0 {
			mids = append(mids, dynSchemaCtx.TopicCreatorMid)
		}
		for mid := range midm {
			mids = append(mids, mid)
		}
		// 用户信息
		eg3.Go(func(ctx context.Context) error {
			res, err := s.cards3New(ctx, mids)
			if err != nil {
				log.Warn("getMaterial mid(%v) cards3New(%v) error(%v)", general.Mid, mids, err)
				return nil
			}
			ret.ResUser = res
			return nil
		})
		// 用户关注关系
		eg3.Go(func(ctx context.Context) error {
			ret.ResRelation = s.isAttention(ctx, mids, general.Mid)
			return nil
		})
		// 粉丝数
		eg3.Go(func(ctx context.Context) error {
			res, err := s.stats(ctx, mids)
			if err != nil {
				log.Warn("getMaterial mid(%v) stats(%v), error(%v)", general.Mid, mids, err)
				return nil
			}
			ret.ResStat = res
			return nil
		})
		// 用户关注关系(包括悄悄关注)
		eg3.Go(func(ctx context.Context) error {
			res, err := s.interrelations(ctx, general.Mid, mids)
			if err != nil {
				log.Error("getMaterial mid(%v) interrelations(%v), error %v", general.Mid, mids, err)
				return nil
			}
			ret.ResRelationUltima = res
			return nil
		})
	}
	// 附加UP发布的活动
	if len(additionalUpActivityIDm) > 0 {
		var additionalActivityIDs []int64
		for id := range additionalUpActivityIDm {
			additionalActivityIDs = append(additionalActivityIDs, id)
		}
		eg3.Go(func(ctx context.Context) error {
			res, err := s.natPageGrpcClient.NativeAllPageCards(ctx, &natpagegrpc.NativeAllPageCardsReq{
				Pids: additionalActivityIDs,
			})
			if err != nil {
				log.Error("getMaterial mid(%+v) NativeAllPageCards(%+v), err(%+v)", general.Mid, additionalActivityIDs, err)
				return nil
			}
			for _, v := range res.List {
				if v.RelatedUid != 0 {
					midrw.Lock()
					midm[v.RelatedUid] = struct{}{}
					midrw.Unlock()
				}
			}
			ret.NativeAllPageCards = res.List
			return nil
		})
	}
	// 直播回放稿件
	if len(liveAidm) > 0 {
		var aids []int64
		for v := range liveAidm {
			aids = append(aids, v)
		}
		eg3.Go(func(ctx context.Context) error {
			res, err := s.archiveGRPC.Arcs(ctx, &archivegrpc.ArcsRequest{Aids: aids, Device: general.GetDevice(), MobiApp: general.GetMobiApp(), Mid: general.Mid})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResArcs = res.Arcs
			return nil
		})
	}
	// 获取在线人数
	if len(playCountIDs) > 0 {
		eg3.Go(func(ctx context.Context) error {
			res, err := s.playOnline(ctx, playCountIDs)
			if err != nil {
				log.Error("s.playOnline %+v", err)
				return nil
			}
			ret.ResPlayUrlCount = res
			return nil
		})
	}
	if err := eg3.Wait(); err != nil {
		return nil, err
	}
	dynSchemaCtx.DynCtx = ret
	return dynSchemaCtx, nil
}

func makeIdsNeedTopicCreatorReply(appear int32, replyIds []string) []string {
	if appear == 0 {
		return nil
	}
	return replyIds
}

func (s *Service) procBackfill(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, rawList *topiccardmodel.DynRawList, schema *dynHandler.CardSchema) {
	// 聚合回填物料
	s.backfillGetMaterial(dynSchemaCtx, general)
	// 遍历回填
	for _, dynRawItem := range rawList.List {
		if dynRawItem == nil || dynRawItem.Item == nil {
			continue
		}
		schema.Backfill(dynSchemaCtx.DynCtx, dynRawItem.Item, general)
	}
}

// nolint:gocognit
func (s *Service) backfillGetMaterial(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) {
	ctx, dynCtx := dynSchemaCtx.Ctx, dynSchemaCtx.DynCtx
	// 正文高亮回填
	eg := errgroup.WithCancel(ctx)
	if len(dynCtx.Emoji) > 0 {
		eg.Go(func(ctx context.Context) error {
			var emoji []string
			for item := range dynCtx.Emoji {
				emoji = append(emoji, item)
			}
			resEmoji, err := s.getEmoji(ctx, emoji)
			if err != nil {
				log.Error("BackfillGetMaterial mid(%v) getEmoji(%v), error %v", general.Mid, emoji, err)
				return err
			}
			dynCtx.ResEmoji = resEmoji
			return nil
		})
	}
	var (
		aidm        = make(map[int64]map[int64]struct{})
		ssidm       = make(map[int32]struct{})
		epidm       = make(map[int32]struct{})
		cvidm       = make(map[int64]struct{})
		shortURLm   = make(map[string]struct{})
		shortToLong map[string]string
	)
	// 抽离短链
	for descURL := range dynCtx.BackfillDescURL {
		r := regexp.MustCompile(_shortURLRex)
		fIndex := r.FindStringIndex(descURL)
		if len(fIndex) == 0 {
			continue
		}
		shortURLm[descURL] = struct{}{}
	}
	// 短链转长链
	if len(shortURLm) > 0 {
		var shortURLs []string
		for shortURL := range shortURLm {
			shortURLs = append(shortURLs, shortURL)
		}
		// 短链转长链
		var err error
		shortToLong, err = s.shortUrls(ctx, shortURLs)
		if err != nil {
			log.Error("BackfillGetMaterial mid(%v) shortUrls(%v), error %v", general.Mid, shortURLs, err)
		}
	}
	// 聚合网页链接数据
	for descURL := range dynCtx.BackfillDescURL {
		var descURLTmp = descURL
		if stl, ok := shortToLong[descURL]; ok && stl != "" {
			descURLTmp = stl
		}
		// archive
		ugcr := regexp.MustCompile(_ugcURLReg)
		if ugcIndex := ugcr.FindStringIndex(descURLTmp); len(ugcIndex) > 0 {
			ugcURL := descURLTmp[ugcIndex[0]:ugcIndex[1]]
			// 拆bvid
			bvr := regexp.MustCompile(_bvRex)
			if bvIndex := bvr.FindStringIndex(ugcURL); len(bvIndex) > 0 {
				bv := ugcURL[bvIndex[0]:bvIndex[1]]
				if aid, _ := bvid.BvToAv(bv); aid != 0 {
					if _, ok := aidm[aid]; !ok {
						aidm[aid] = make(map[int64]struct{})
					}
					dynCtx.BackfillDescURL[descURL] = &dynmdlV2.BackfillDescURLItem{
						Type:  dynamicapi.DescType_desc_type_bv,
						Title: "",
						Rid:   bv,
					}
				}
				continue
			}
			// 拆avid
			avr := regexp.MustCompile(_avRex)
			if avIndex := avr.FindStringIndex(ugcURL); len(avIndex) > 0 {
				avid := ugcURL[avIndex[0]:avIndex[1]]
				// 拆id
				idr := regexp.MustCompile(_idReg)
				if idIndex := idr.FindStringIndex(avid); len(idIndex) > 0 {
					id := avid[idIndex[0]:idIndex[1]]
					if idInt64, _ := strconv.ParseInt(id, 10, 64); idInt64 != 0 {
						if _, ok := aidm[idInt64]; !ok {
							aidm[idInt64] = make(map[int64]struct{})
						}
						dynCtx.BackfillDescURL[descURL] = &dynmdlV2.BackfillDescURLItem{
							Type:  dynamicapi.DescType_desc_type_av,
							Title: "",
							Rid:   id,
						}
					}
				}
			}
			continue
		}
		// ogv
		ogvr := regexp.MustCompile(_ogvURLReg)
		if ogvIndex := ogvr.FindStringIndex(descURLTmp); len(ogvIndex) > 0 {
			ogvURL := descURLTmp[ogvIndex[0]:ogvIndex[1]]
			// 拆ssid
			ssidr := regexp.MustCompile(_ogvssRex)
			if ssidIndex := ssidr.FindStringIndex(ogvURL); len(ssidIndex) > 0 {
				ssid := ogvURL[ssidIndex[0]:ssidIndex[1]]
				idr := regexp.MustCompile(_idReg)
				if idIndex := idr.FindStringIndex(ssid); len(idIndex) > 0 {
					id := ssid[idIndex[0]:idIndex[1]]
					if idInt, _ := strconv.ParseInt(id, 10, 32); idInt != 0 {
						ssidm[int32(idInt)] = struct{}{}
						dynCtx.BackfillDescURL[descURL] = &dynmdlV2.BackfillDescURLItem{
							Type:  dynamicapi.DescType_desc_type_ogv_season,
							Title: "",
							Rid:   id,
						}
					}
				}
				continue
			}
			epidr := regexp.MustCompile(_ogvepRex)
			if epidIndex := epidr.FindStringIndex(ogvURL); len(epidIndex) > 0 {
				epid := ogvURL[epidIndex[0]:epidIndex[1]]
				idr := regexp.MustCompile(_idReg)
				if idIndex := idr.FindStringIndex(epid); len(idIndex) > 0 {
					id := epid[idIndex[0]:idIndex[1]]
					if idInt, _ := strconv.ParseInt(id, 10, 32); idInt != 0 {
						epidm[int32(idInt)] = struct{}{}
						dynCtx.BackfillDescURL[descURL] = &dynmdlV2.BackfillDescURLItem{
							Type:  dynamicapi.DescType_desc_type_ogv_ep,
							Title: "",
							Rid:   id,
						}
					}
				}
			}
			continue
		}
		// article
		cvr := regexp.MustCompile(_articleURLReg)
		if cvIndex := cvr.FindStringIndex(descURLTmp); len(cvIndex) > 0 {
			cvURL := descURLTmp[cvIndex[0]:cvIndex[1]]
			// 拆cvid
			cvidr := regexp.MustCompile(_cvRex)
			if cvidIndx := cvidr.FindStringIndex(cvURL); len(cvidIndx) > 0 {
				cvid := cvURL[cvidIndx[0]:cvidIndx[1]]
				// 拆id
				idr := regexp.MustCompile(_idReg)
				if idIndex := idr.FindStringIndex(cvid); len(idIndex) > 0 {
					cv := cvid[idIndex[0]:idIndex[1]]
					if cvidInt64, _ := strconv.ParseInt(cv, 10, 64); cvidInt64 != 0 {
						cvidm[cvidInt64] = struct{}{}
						dynCtx.BackfillDescURL[descURL] = &dynmdlV2.BackfillDescURLItem{
							Type:  dynamicapi.DescType_desc_type_cv,
							Title: "",
							Rid:   cv,
						}
					}
				}
			}
			continue
		}
	}
	// 聚合id
	for aid := range dynCtx.BackfillAvID {
		if aidInt64, _ := strconv.ParseInt(aid, 10, 64); aidInt64 != 0 {
			if _, ok := aidm[aidInt64]; !ok {
				aidm[aidInt64] = make(map[int64]struct{})
			}
		}
	}
	for bv := range dynCtx.BackfillBvID {
		if aid, _ := bvid.BvToAv(bv); aid != 0 {
			if _, ok := aidm[aid]; !ok {
				aidm[aid] = make(map[int64]struct{})
			}
		}
	}
	for cvid := range dynCtx.BackfillCvID {
		if cvidInt64, _ := strconv.ParseInt(cvid, 10, 64); cvidInt64 != 0 {
			cvidm[cvidInt64] = struct{}{}
		}
	}
	if len(aidm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var aids []*archivegrpc.PlayAv
			for aid, cidm := range aidm {
				var ap = &archivegrpc.PlayAv{Aid: aid, NoPlayer: true}
				for cid := range cidm {
					ap.PlayVideos = append(ap.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
				}
				aids = append(aids, ap)
			}
			// 假卡不要秒开
			res, err := s.arcsPlayer(ctx, aids, true, "")
			if err != nil {
				log.Error("BackfillGetMaterial mid(%v) ArcsWithPlayurl(%v), err %v", general.Mid, aids, err)
				return err
			}
			dynCtx.ResBackfillArchive = res
			return nil
		})
	}
	if len(cvidm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var cvids []int64
			for id := range cvidm {
				cvids = append(cvids, id)
			}
			rsp, err := s.articleGRPC.ArticleMetas(ctx, &articlegrpc.ArticleMetasReq{Ids: cvids})
			if err != nil {
				log.Error("BackfillGetMaterial mid(%v) ArticleMetasMc(%v), err %v", general.Mid, cvids, err)
				return err
			}
			dynCtx.ResBackfillArticle = rsp.GetRes()
			return nil
		})
	}
	if len(ssidm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var ssids []int32
			for id := range ssidm {
				ssids = append(ssids, id)
			}
			res, err := s.seasons(ctx, ssids)
			if err != nil {
				log.Error("BackfillGetMaterial mid(%v) seasons(%v), err %v", general.Mid, ssids, err)
				return err
			}
			dynCtx.ResBackfillSeason = res
			return nil
		})
	}
	if len(epidm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var epids []int32
			for id := range epidm {
				epids = append(epids, id)
			}
			res, err := s.episodes(ctx, epids)
			if err != nil {
				log.Error("BackfillGetMaterial mid(%v) episodes(%v), err %v", general.Mid, epids, err)
				return err
			}
			dynCtx.ResBackfillEpisode = res
			return nil
		})
	}
	_ = eg.Wait()
}

func makeFollowCardParam(aids []int64, general *topiccardmodel.GeneralParam) *pgcDynGrpc.FollowCardReq {
	return &pgcDynGrpc.FollowCardReq{
		Aid: aids,
		User: &pgcDynGrpc.UserProto{
			Mid:      general.Mid,
			MobiApp:  general.GetMobiApp(),
			Device:   general.GetDevice(),
			Platform: general.GetPlatform(),
			Build:    int32(general.GetBuild()),
		},
	}
}

// nolint:gocognit
func (s *Service) getMaterialIDs(dyn *dynmdlV2.Dynamic, general *topiccardmodel.GeneralParam, ret *dynmdlV2.DynamicContext, aidm, storyAidm map[int64]map[int64]struct{},
	midm, epIDm, dynIDm, drawIDm, articleIDm, commonIDm, additionalUpIDm, pugvIDm map[int64]struct{}, likeIDm map[string][]*dynmdlV2.ThumbsRecord, replyIDm map[string]struct{},
	dynamicActivityArgs map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo, additionalTopic map[int64][]*dynmdlV2.Topic,
	voteIDm map[int64]struct{}, additionalUpActivityIDm map[int64]struct{}, premiereAids map[int64]map[int64]struct{}, goods map[int64]*bcgmdl.GoodsParams) {
	// 评论业务ID聚合
	if dyn.GetReplyID() != "" {
		replyIDm[dyn.GetReplyID()] = struct{}{}
	}
	// 点赞业务ID聚合
	if busParam, busType, isThum := dyn.GetLikeID(); isThum {
		likeIDm[busType] = append(likeIDm[busType], busParam)
	}
	// 基础信息：用户
	if dyn.UID != 0 {
		midm[dyn.UID] = struct{}{}
	}
	// 核心物料ID
	if dyn.Rid != 0 {
		// 类型判断
		switch {
		case dyn.IsAv():
			// 首映召回
			if dyn.Property != nil && dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_PREMIERE_RESERVE {
				premiereAids[dyn.Rid] = make(map[int64]struct{})
			} else if dyn.SType == dynmdlV2.VideoStypeDynamicStory {
				if _, ok := aidm[dyn.Rid]; !ok {
					storyAidm[dyn.Rid] = make(map[int64]struct{})
				}
			} else {
				if _, ok := aidm[dyn.Rid]; !ok {
					aidm[dyn.Rid] = make(map[int64]struct{})
				}
			}
			// 获取当前UGC稿件已经收入到ogv视频的详细信息
			if dyn.PassThrough != nil && dyn.PassThrough.PgcBadge != nil && dyn.PassThrough.PgcBadge.EpisodeId > 0 {
				epIDm[dyn.PassThrough.PgcBadge.EpisodeId] = struct{}{}
			}
		case dyn.IsForward():
			dynIDm[dyn.Rid] = struct{}{}
		case dyn.IsDraw():
			drawIDm[dyn.Rid] = struct{}{}
		case dyn.IsWord():
			dynIDm[dyn.Rid] = struct{}{}
		case dyn.IsArticle():
			articleIDm[dyn.Rid] = struct{}{}
		case dyn.IsCommon():
			commonIDm[dyn.Rid] = struct{}{}
		case dyn.IsPGC():
			epIDm[dyn.Rid] = struct{}{}
		}
	}
	// 帮推
	tmpDynamicAttachedPromoInfo := &dynactivitygrpc.DynamicAttachedPromoInfo{
		DynamicId: dyn.DynamicID,
		Rid:       dyn.Rid,
		Type:      dyn.Type,
	}
	dynamicActivityArgs[dyn.DynamicID] = tmpDynamicAttachedPromoInfo
	// 附加大卡
	for _, v := range dyn.AttachCardInfos {
		if v.Rid == 0 {
			continue
		}
		// nolint:exhaustive
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:
			if _, ok := aidm[v.Rid]; !ok {
				aidm[v.Rid] = make(map[int64]struct{})
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:
			// 投票
			voteIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UP_ACTIVITY:
			// UP发布的活动
			additionalUpActivityIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			// UP主预约卡
			additionalUpIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV:
			// 课程
			pugvIDm[v.Rid] = struct{}{}
		}
	}
	// 扩展部分
	if dyn.Extend != nil {
		// 附加内容判断
		if ok, likeUser := dyn.GetLikeUser(); ok { // 好友点赞
			for _, uid := range likeUser {
				if uid != 0 {
					midm[uid] = struct{}{}
				}
			}
		}
		// 话题
		if dyn.Extend.TopicInfo != nil {
			for _, topic := range dyn.Extend.TopicInfo.TopicInfos {
				additionalTopic[topic.TopicID] = append(additionalTopic[topic.TopicID], topic)
			}
			ret.ResAdditionalTopic = additionalTopic
		}
		if dyn.Extend.VideoShare != nil && dyn.IsAv() && dyn.Rid != 0 {
			if dyn.SType == dynmdlV2.VideoStypeDynamicStory {
				storyAidm[dyn.Rid][dyn.Extend.VideoShare.CID] = struct{}{}
			} else {
				aidm[dyn.Rid][dyn.Extend.VideoShare.CID] = struct{}{}
			}
		}
		//商品卡
		if dyn.Extend.OpenGoods != nil {
			goodsCtx := &bcgmdl.GoodsCtx{
				Build:      general.GetBuild(),
				Buvid:      general.GetBuvid(),
				IP:         general.IP,
				PlatformID: bcgmdl.TranPlatformID(general.GetPlatform()),
			}
			goodsCtxStr, _ := json.Marshal(goodsCtx)
			inputStr, _ := json.Marshal(dyn.Extend.OpenGoods)
			tmp := &bcgmdl.GoodsParams{
				Uid:         general.Mid,
				UpUid:       dyn.UID,
				DynamicID:   dyn.DynamicID,
				Ctx:         string(goodsCtxStr),
				InputExtend: string(inputStr),
			}
			goods[dyn.DynamicID] = tmp
		}
	}
}

func (s *Service) drawDetails(c context.Context, general *topiccardmodel.GeneralParam, drawIds []int64) (map[int64]*dynmdlV2.DrawDetailRes, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*dynmdlV2.DrawDetailRes)
	for i := 0; i < len(drawIds); i += max50 {
		var tmpids []int64
		if i+max50 > len(drawIds) {
			tmpids = drawIds[i:]
		} else {
			tmpids = drawIds[i : i+max50]
		}
		g.Go(func(ctx context.Context) error {
			reply, err := s.drawDetailsGRPC(ctx, general, tmpids)
			if err != nil {
				log.Error("drawDetailsGRPC failed: %+v", err)
				return err
			}
			mu.Lock()
			for k, v := range reply {
				res[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("drawDetails error group=%+v", err)
		return nil, err
	}
	return res, nil
}

func (s *Service) drawDetailsGRPC(c context.Context, general *topiccardmodel.GeneralParam, drawIds []int64) (map[int64]*dynmdlV2.DrawDetailRes, error) {
	drawDetailReq := &dyndrawrpc.DrawDetailReq{
		Uid:      general.Mid,
		Ids:      drawIds,
		MetaData: general.ToDynCmnMetaData(),
	}
	resp, err := s.dynDrawGRPC.Detail(c, drawDetailReq)
	if err != nil {
		return nil, err
	}

	ret := make(map[int64]*dynmdlV2.DrawDetailRes)
	for _, item := range resp.GetDocItems() {
		itemSts := new(dynmdlV2.DrawDetailRes)
		err = json.Unmarshal([]byte(item.GetItem()), itemSts)
		if err != nil {
			log.Error("drawDetailsGRPC unmarshal resp item error: %+v", err)
			continue
		}
		ret[item.GetDocId()] = itemSts
	}
	return ret, nil
}

func (s *Service) goodsDetails(c context.Context, req *bcgmdl.GoodsParams) (map[int]map[string]*bcgmdl.GoodsItem, error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(req.Uid, 10))
	params.Set("up_uid", strconv.FormatInt(req.UpUid, 10))
	params.Set("dynamic_id", strconv.FormatInt(req.DynamicID, 10))
	params.Set("ctx", req.Ctx)
	params.Set("input_extend", req.InputExtend)
	var ret struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data *bcgmdl.GoodsRes `json:"data"`
	}
	if err := s.httpMgr.Get(c, s.goodsDetailsURL, "", params, &ret); err != nil {
		return nil, err
	}
	if ret.Code != 0 {
		return nil, errors.Wrap(ecode.Int(ret.Code), s.goodsDetailsURL+"?"+params.Encode())
	}
	if ret.Data == nil || ret.Data.OutputExtend.List == nil {
		return nil, errors.New("GoodsDetails ret.data is empty.")
	}
	rsp := make(map[int]map[string]*bcgmdl.GoodsItem, len(ret.Data.OutputExtend.List))
	for _, item := range ret.Data.OutputExtend.List {
		resMap, ok := rsp[item.Type]
		if !ok {
			resMap = make(map[string]*bcgmdl.GoodsItem, 1)
			rsp[item.Type] = resMap
		}
		resMap[strconv.FormatInt(item.ItemsID, 10)] = item
	}
	return rsp, nil
}

func (s *Service) votes(ctx context.Context, mid int64, voteIDs []int64) (map[int64]*dyncommongrpc.VoteInfo, error) {
	resTmp, err := s.dynVoteGRPC.ListFeedVotes(ctx, &dynvotegrpc.ListFeedVotesReq{Uid: mid, VoteIds: voteIDs})
	if err != nil {
		log.Error("votes %v, err %v", voteIDs, err)
		return nil, err
	}
	return resTmp.GetVoteInfos(), nil
}

func (s *Service) dynamicAttachedPromo(ctx context.Context, args []*dynactivitygrpc.DynamicAttachedPromoInfo) (map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo, error) {
	resTmp, err := s.dynamicActivityGRPC.DynamicAttachedPromo(ctx, &dynactivitygrpc.DynamicAttachedPromoReq{Dynamics: args})
	if err != nil {
		return nil, err
	}
	var res = make(map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo)
	if resTmp != nil {
		for _, promo := range resTmp.AttachedPromos {
			if promo == nil {
				continue
			}
			res[promo.DynamicId] = promo
		}
	}
	return res, nil
}

// 获取aids对应的动态id
func (s *Service) fetchRevs(ctx context.Context, aids []int64) (map[int64]int64, error) {
	const (
		_archiveDynamicType = 8
	)
	var devItems []*dynamicrevs.RevsItem
	for _, v := range aids {
		devItems = append(devItems, &dynamicrevs.RevsItem{Rid: v, Type: _archiveDynamicType})
	}
	fanoutRevs, err := s.dynRevGRPC.FetchRevs(ctx, &dynamicrevs.FetchRevsReq{Items: devItems})
	if err != nil {
		return nil, errors.Wrapf(err, "fetchRevs aids=%+v, err=%+v", aids, err)
	}
	res := map[int64]int64{}
	for _, v := range fanoutRevs.Items {
		res[v.Rid] = v.DynId
	}
	return res, nil
}

// 获取通用模板
func (s *Service) DynamicCommonInfos(ctx context.Context, ids []int64) (map[int64]*dynmdlV2.DynamicCommonCard, error) {
	params := new(struct {
		RIDs []int64 `json:"rid"`
	})
	params.RIDs = ids
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return nil, errors.Wrapf(err, "DynamicCommonInfos json.NewEncoder() params(%+v)", params)
	}
	req, err := http.NewRequest(http.MethodPost, s.dynCommonBiz, body)
	if err != nil {
		return nil, errors.Wrapf(err, "DynamicCommonInfos http.NewRequest() body(%s)", body)
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			Entry []*dynmdlV2.DynamicCommon `json:"entry"`
		} `json:"data"`
	}
	if err = s.httpMgr.Do(ctx, req, &ret); err != nil {
		return nil, errors.Wrapf(err, "DynamicCommonInfos http Post(%s) failed, req:(%+v)", s.dynCommonBiz, req)
	}
	if ret.Code != 0 {
		return nil, errors.Wrapf(ecode.Int(ret.Code), "DynamicCommonInfos url=%+v code=%+v msg=%+v", s.dynCommonBiz, ret.Code, ret.Msg)
	}
	if ret.Data == nil || len(ret.Data.Entry) == 0 {
		return nil, errors.New("DynamicCommonInfos get nothing")
	}
	var res = make(map[int64]*dynmdlV2.DynamicCommonCard)
	for _, entry := range ret.Data.Entry {
		if entry.RID == 0 || entry.Card == "" {
			log.Error("DynamicCommonInfos entry err=%+v", entry)
			continue
		}
		card := &dynmdlV2.DynamicCommonCard{}
		if err = json.Unmarshal([]byte(entry.Card), &card); err != nil {
			log.Error("DynamicCommonInfos json unmarshal entry err=%+v", err)
			continue
		}
		res[entry.RID] = card
	}
	return res, nil
}

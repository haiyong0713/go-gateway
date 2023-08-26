package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	articleMdl "git.bilibili.co/bapis/bapis-go/article/model"
	articlegrpc "git.bilibili.co/bapis/bapis-go/article/service"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"

	"github.com/pkg/errors"
)

// 话题详情页头部信息卡片物料装载器
type TopicFanoutLoader struct {
	General      *topiccardmodel.GeneralParam
	Service      *Service
	Archive      loaderArchiveSubset
	Live         loaderLiveSubset
	Account      loaderAccountSubset
	Bangumi      loaderBangumiSubset
	ThumbUp      loaderThumbUpSubset
	Dynamic      loaderDynamicSubset
	FavouriteAid []int64
}

type FanoutResult struct {
	Archive struct {
		Archive   map[int64]*archivegrpc.ArcPlayer
		DynamicId map[int64]int64
	}
	Live struct {
		InlineRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
	}
	Account struct {
		Card        map[int64]*accountgrpc.Card
		IsAttention map[int64]int32
	}
	Bangumi struct {
		InlinePGC map[int32]*pgcinline.EpisodeCard
	}
	ThumbUp struct {
		HasLikeArchive map[int64]int8
	}
	Favourite map[int64]int8
	Coin      map[int64]int64
	Dynamic   struct {
		ResUser          map[int64]*accountgrpc.Info        // 用户信息
		ResTopicUser     map[int64]*accountgrpc.Info        // 话题创建者用户信息
		ResArchive       map[int64]*archivegrpc.ArcPlayer   // 稿件详情
		ResWords         map[int64]string                   // 转发卡文案、纯文字卡文案
		ResDraw          map[int64]*dynamicV2.DrawDetailRes // 图文卡
		ResArticle       map[int64]*articleMdl.Meta         // 专栏卡
		ResDynSimpleInfo map[int64]*dyngrpc.DynSimpleInfo   // 动态简要信息
	}
}

func (loader *TopicFanoutLoader) doDynamicCardFanoutLoad(ctx context.Context) (*FanoutResult, error) {
	fanout := &FanoutResult{}
	if err := loader.doDynBriefs(ctx, fanout); err != nil {
		return nil, err
	}
	if len(loader.Dynamic.DynamicIds) == 0 || fanout.Dynamic.ResDynSimpleInfo == nil {
		return nil, errors.New("doDynamicCardFanoutLoad load nothing")
	}
	for _, v := range loader.Dynamic.DynamicIds {
		if info, ok := fanout.Dynamic.ResDynSimpleInfo[v]; ok {
			loader.Dynamic.AccountUIDs = append(loader.Dynamic.AccountUIDs, info.Uid)
			switch info.Type {
			case dynamicV2.DynTypeVideo:
				loader.Dynamic.Aids = append(loader.Dynamic.Aids, info.Rid)
			case dynamicV2.DynTypeForward, dynamicV2.DynTypeWord:
				loader.Dynamic.WordRids = append(loader.Dynamic.WordRids, info.Rid)
			case dynamicV2.DynTypeDraw:
				loader.Dynamic.DrawRids = append(loader.Dynamic.DrawRids, info.Rid)
			case dynamicV2.DynTypeArticle:
				loader.Dynamic.ArticleRids = append(loader.Dynamic.ArticleRids, info.Rid)
			}
		}
	}
	eg := errgroup.WithContext(ctx)
	loader.doDynUser(eg, fanout)
	loader.doDynArchive(eg, fanout)
	loader.doDynWords(eg, fanout)
	loader.doDynDraw(eg, fanout)
	loader.doDynArticle(eg, fanout)
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "doDynamicCardFanoutLoad eg.Wait() err=%+v", err)
		return nil, err
	}
	return fanout, nil
}

func (loader *TopicFanoutLoader) doDynWords(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Dynamic.WordRids) > 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := loader.Service.dynGRPC.ListWordText(ctx, &dyngrpc.WordTextReq{Uid: loader.General.Mid, Rids: loader.Dynamic.WordRids})
			if err != nil {
				log.Errorc(ctx, "doDynWords ListWordText loader.Dynamic.WordRids=%+v, err=%+v", loader.Dynamic.WordRids, err)
				return nil
			}
			fanout.Dynamic.ResWords = res.GetContent()
			return nil
		})
	}
}

func (loader *TopicFanoutLoader) doDynDraw(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Dynamic.DrawRids) > 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := loader.Service.drawDetails(ctx, loader.General, loader.Dynamic.DrawRids)
			if err != nil {
				log.Errorc(ctx, "doDynDraw drawDetails loader.Dynamic.DrawRids=%+v, err=%+v", loader.Dynamic.DrawRids, err)
				return nil
			}
			fanout.Dynamic.ResDraw = res
			return nil
		})
	}
}

func (loader *TopicFanoutLoader) doDynArticle(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Dynamic.ArticleRids) > 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := loader.Service.articleGRPC.ArticleMetas(ctx, &articlegrpc.ArticleMetasReq{Ids: loader.Dynamic.ArticleRids})
			if err != nil {
				log.Errorc(ctx, "doDynArticle ArticleMetas loader.Dynamic.ArticleRids=%+v, err=%+v", loader.Dynamic.ArticleRids, err)
				return nil
			}
			fanout.Dynamic.ResArticle = res.GetRes()
			return nil
		})
	}
}

func (loader *TopicFanoutLoader) doDynUser(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Dynamic.AccountUIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			// 动态作者信息
			res, err := loader.Service.accGRPC.Infos3(ctx, &accountgrpc.MidsReq{Mids: loader.Dynamic.AccountUIDs})
			if err != nil {
				return errors.Wrapf(err, "doDynUser Infos3 uids=%+v, error=%+v", loader.Dynamic.AccountUIDs, err)
			}
			fanout.Dynamic.ResUser = res.GetInfos()
			return nil
		})
	}
	if len(loader.Dynamic.TopicUpIds) > 0 {
		eg.Go(func(ctx context.Context) error {
			// 话题创建者信息
			res, err := loader.Service.accGRPC.Infos3(ctx, &accountgrpc.MidsReq{Mids: loader.Dynamic.TopicUpIds})
			if err != nil {
				log.Errorc(ctx, "doDynUser Infos3 topicCreator=%+v, error=%+v", loader.Dynamic.TopicUpIds, err)
				return nil
			}
			fanout.Dynamic.ResTopicUser = res.GetInfos()
			return nil
		})
	}
}

func (loader *TopicFanoutLoader) doDynArchive(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Dynamic.Aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var aids []*archivegrpc.PlayAv
			for _, v := range loader.Dynamic.Aids {
				aids = append(aids, &archivegrpc.PlayAv{Aid: v})
			}
			res, err := loader.Service.arcsPlayer(ctx, aids, false, "")
			if err != nil {
				log.Errorc(ctx, "TopicFanoutLoader doDynArchive aids=%+v, err=%+v", aids, err)
				return nil
			}
			fanout.Dynamic.ResArchive = res
			return nil
		})
	}
}

func (loader *TopicFanoutLoader) doDynBriefs(ctx context.Context, fanout *FanoutResult) error {
	if len(loader.Dynamic.DynamicIds) == 0 {
		return errors.Errorf(" no DynamicIds loader=%+v", loader)
	}
	reply, err := loader.Service.dynGRPC.DynSimpleInfos(ctx, &dyngrpc.DynSimpleInfosReq{
		DynIds: loader.Dynamic.DynamicIds,
		Uid:    loader.General.Mid,
	})
	if err != nil {
		return errors.Wrapf(err, "doDynBriefs DynSimpleInfos info: %+v: %+v", loader.Dynamic.DynamicIds, err)
	}
	fanout.Dynamic.ResDynSimpleInfo = reply.DynSimpleInfos
	return nil
}

func (loader *TopicFanoutLoader) doTopicCardFanoutLoad(ctx context.Context) (*FanoutResult, error) {
	fanout := &FanoutResult{}
	eg := errgroup.WithContext(ctx)
	loader.doArchive(eg, fanout)
	loader.doLive(eg, fanout)
	loader.doBangumi(eg, fanout)
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute doTopicCardFanoutLoad error group: %+v", err)
		return nil, err
	}
	if err := loader.doSecondLoader(ctx, fanout); err != nil {
		log.Error("Failed to execute doSecondLoader error=%+v", err)
		return nil, err
	}
	return fanout, nil
}

func (loader *TopicFanoutLoader) doArchive(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Archive.Aids) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		var aids []*archivegrpc.PlayAv
		for _, v := range loader.Archive.Aids {
			aids = append(aids, &archivegrpc.PlayAv{Aid: v})
		}
		res, err := loader.Service.arcsPlayer(ctx, aids, false, "story")
		if err != nil {
			return errors.Wrapf(err, "TopicFanoutLoader arcsPlayer aids=%+v, err=%+v", aids, err)
		}
		fanout.Archive.Archive = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		fanoutRevs, err := loader.Service.fetchRevs(ctx, loader.Archive.Aids)
		if err != nil {
			return errors.Wrapf(err, "TopicFanoutLoader fetchRevs aids=%+v, err=%+v", loader.Archive.Aids, err)
		}
		fanout.Archive.DynamicId = fanoutRevs
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if loader.General.Mid == 0 {
			return nil
		}
		fanoutcoins, err := loader.Service.archiveUserCoins(ctx, loader.Archive.Aids, loader.General.Mid)
		if err != nil {
			return errors.Wrapf(err, "TopicFanoutLoader archiveUserCoins mid=%d, aids=%+v, err=%+v", loader.General.Mid, loader.Archive.Aids, err)
		}
		fanout.Coin = fanoutcoins
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if loader.General.Mid == 0 {
			return nil
		}
		fanoutFavs, err := loader.Service.isFavVideos(ctx, loader.General.Mid, loader.Archive.Aids)
		if err != nil {
			return errors.Wrapf(err, "TopicFanoutLoader isFavVideos aids=%+v, err=%+v", loader.Archive.Aids, err)
		}
		fanout.Favourite = fanoutFavs
		return nil
	})
}

func (loader *TopicFanoutLoader) doLive(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Live.InlineRoomIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: []string{_newTopicLiveEntry},
				RoomIds:   loader.Live.InlineRoomIDs,
				Uid:       loader.General.Mid,
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
				Platform:  loader.General.GetPlatform(),
				Build:     loader.General.GetBuild(),
			}
			entryRoom, err := loader.Service.livexroomGateGRPC.EntryRoomInfo(ctx, req)
			if err != nil {
				return errors.Wrapf(err, "Failed to get entry room info: %+v: %+v", req, err)
			}
			fanout.Live.InlineRoom = entryRoom.List
			return nil
		})
	}
}

func (loader *TopicFanoutLoader) doAccount(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Account.AccountUIDs) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		res, err := loader.Service.cards3Slice(ctx, loader.Account.AccountUIDs)
		if err != nil {
			return errors.Wrapf(err, "doAccount loader.Service.cards3Slice uids=%+v, error= %+v", loader.Account.AccountUIDs, err)
		}
		fanout.Account.Card = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		fanout.Account.IsAttention = loader.Service.isAttention(ctx, loader.Account.AccountUIDs, loader.General.Mid)
		return nil
	})
}

func (loader *TopicFanoutLoader) doBangumi(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Bangumi.EPID) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		req := &pgcinline.EpReq{
			EpIds: loader.Bangumi.EPID,
			User: &pgcinline.UserReq{
				Mid:      loader.General.Mid,
				MobiApp:  loader.General.GetMobiApp(),
				Device:   loader.General.GetDevice(),
				Platform: loader.General.GetPlatform(),
				Ip:       metadata.String(ctx, metadata.RemoteIP),
				Build:    int32(loader.General.GetBuild()),
			},
		}
		if batchArg, ok := arcmid.FromContext(ctx); ok {
			req.User.Fnver = uint32(batchArg.Fnver)
			req.User.Fnval = uint32(batchArg.Fnval)
			req.User.Qn = uint32(batchArg.Qn)
			req.User.Fourk = int32(batchArg.Fourk)
			req.User.NetType = pgccard.NetworkType(batchArg.NetType)
			req.User.TfType = pgccard.TFType(batchArg.TfType)
		}
		res, err := loader.Service.pgcInlineGRPC.EpCard(ctx, req)
		if err != nil {
			return errors.Wrapf(err, "doBangumi loader.Service.pgcInlineGRPC.EpCard epids=%+v, error= %+v", loader.Bangumi.EPID, err)
		}
		fanout.Bangumi.InlinePGC = res.Infos
		return nil
	})
}

func (loader *TopicFanoutLoader) doSecondLoader(ctx context.Context, fanout *FanoutResult) error {
	for _, a := range fanout.Archive.Archive {
		loader.WithAccountProfile(a.Arc.Author.Mid)
		loader.WithThumbUpArchive(a.Arc.Aid)
	}
	for _, r := range fanout.Live.InlineRoom {
		loader.WithAccountProfile(r.Uid)
	}
	for _, p := range fanout.Bangumi.InlinePGC {
		loader.WithThumbUpArchive(p.Aid)
	}
	eg := errgroup.WithContext(ctx)
	loader.doAccount(eg, fanout)
	loader.doHasLike(eg, fanout)
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (loader *TopicFanoutLoader) doHasLike(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.ThumbUp.ArchiveAid) <= 0 || loader.General.Mid == 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		fanoutLikes, err := loader.Service.hasLike(ctx, loader.General.GetBuvid(), loader.General.Mid, loader.ThumbUp.ArchiveAid)
		if err != nil {
			return errors.Wrapf(err, "doSecondLoader hasLike aids=%+v, err=%+v", loader.Archive.Aids, err)
		}
		fanout.ThumbUp.HasLikeArchive = fanoutLikes
		return nil
	})
}

func (loader *TopicFanoutLoader) WithAccountProfile(mid ...int64) {
	loader.Account.AccountUIDs = append(loader.Account.AccountUIDs, mid...)
}

func (loader *TopicFanoutLoader) WithThumbUpArchive(aid ...int64) {
	loader.ThumbUp.ArchiveAid = append(loader.ThumbUp.ArchiveAid, aid...)
}

type loaderDynamicSubset struct {
	TopicUpIds  []int64
	DynamicIds  []int64
	AccountUIDs []int64
	Aids        []int64
	WordRids    []int64
	DrawRids    []int64
	ArticleRids []int64
}

type loaderArchiveSubset struct {
	Aids []int64
}

type loaderLiveSubset struct {
	InlineRoomIDs []int64
}

type loaderAccountSubset struct {
	AccountUIDs []int64
}

type loaderBangumiSubset struct {
	EPID []int32
}

type loaderThumbUpSubset struct {
	ArchiveAid []int64
}

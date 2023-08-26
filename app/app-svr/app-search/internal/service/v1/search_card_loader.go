package v1

import (
	"context"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xmetadata "go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-search/internal/model"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"

	"github.com/pkg/errors"
)

// 新搜索inline卡片物料装载器
type NewInlineCardFanoutLoader struct {
	General      *topiccardmodel.GeneralParam
	Service      *Service
	Archive      loaderArchiveSubset
	Live         loaderLiveSubset
	Account      loaderAccountSubset
	ThumbUp      loaderThumbUpSubset
	FavouriteAid []int64
}

type FanoutResult struct {
	Archive struct {
		Archive map[int64]*archivegrpc.ArcPlayer
	}
	Live struct {
		InlineRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
	}
	Account struct {
		Card        map[int64]*accountgrpc.Card
		IsAttention map[int64]bool
	}
	Bangumi struct {
		InlinePGC map[int32]*pgcinline.EpisodeCard
	}
	ThumbUp struct {
		HasLikeArchive map[int64]thumbupgrpc.State
	}
	Favourite map[int64]int8
	Coin      map[int64]int64
	NftRegion map[int64]*gallerygrpc.NFTRegion
}

func constructGeneralParamFromCtx(ctx context.Context) *topiccardmodel.GeneralParam {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	return &topiccardmodel.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
	}
}

func (loader *NewInlineCardFanoutLoader) doSearchRankingCardFanoutLoad(ctx context.Context) (*FanoutResult, error) {
	fanout := &FanoutResult{}
	eg := errgroup.WithContext(ctx)
	loader.doLive(eg, fanout)
	loader.doArchive(eg, fanout)
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute doSearchRankingCardFanoutLoad error: %+v", err)
		return nil, err
	}
	return fanout, nil
}

// nolint:ineffassign
func (loader *NewInlineCardFanoutLoader) doSearchCardFanoutLoad(ctx context.Context) (*FanoutResult, error) {
	fanout := &FanoutResult{}
	eg := errgroup.WithContext(ctx)
	loader.doLive(eg, fanout)
	loader.doArchive(eg, fanout)
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute doSearchCardFanoutLoad error group: %+v", err)
		err = nil
	}
	if err := loader.doSecondLoader(ctx, fanout); err != nil {
		log.Error("Failed to execute doSecondLoader error=%+v", err)
		return nil, err
	}
	return fanout, nil
}

func (loader *NewInlineCardFanoutLoader) doArchive(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Archive.Aids) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		var aids []*archivegrpc.PlayAv
		for _, v := range loader.Archive.Aids {
			aids = append(aids, &archivegrpc.PlayAv{Aid: v})
		}
		res, err := loader.Service.dao.ArcsPlayer(ctx, aids, false)
		if err != nil {
			return errors.Wrapf(err, "NewInlineCardFanoutLoader arcsPlayer aids=%+v, err=%+v", aids, err)
		}
		fanout.Archive.Archive = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if loader.General.Mid == 0 {
			return nil
		}
		fanoutcoins, err := loader.Service.dao.ArchiveUserCoins(ctx, loader.Archive.Aids, loader.General.Mid)
		if err != nil {
			return errors.Wrapf(err, "NewInlineCardFanoutLoader archiveUserCoins mid=%d, aids=%+v, err=%+v", loader.General.Mid, loader.Archive.Aids, err)
		}
		fanout.Coin = fanoutcoins
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if loader.General.Mid == 0 {
			return nil
		}
		fanoutFavs, err := loader.Service.dao.IsFavVideos(ctx, loader.General.Mid, loader.Archive.Aids)
		if err != nil {
			return errors.Wrapf(err, "NewInlineCardFanoutLoader isFavVideos aids=%+v, err=%+v", loader.Archive.Aids, err)
		}
		fanout.Favourite = fanoutFavs
		return nil
	})
}

func (loader *NewInlineCardFanoutLoader) doLive(eg *errgroup.Group, fanout *FanoutResult) {
	// 直播请求的特殊逻辑
	entryFrom := []string{model.SearchLiveInlineCard}
	if len(loader.Live.LiveEntryFrom) > 0 {
		entryFrom = loader.Live.LiveEntryFrom
	}
	if len(loader.Live.UpMids) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: entryFrom,
				Uids:      loader.Live.UpMids,
				Uid:       loader.General.Mid,
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
				Platform:  loader.General.GetPlatform(),
				Build:     loader.General.GetBuild(),
				ReqBiz:    "/x/v2/search",
			}
			entryRoom, err := loader.Service.dao.EntryRoomInfo(ctx, req)
			if err != nil {
				return errors.Wrapf(err, "Failed to get entry room info: %+v: %+v", req, err)
			}
			fanout.Live.InlineRoom = entryRoom
			return nil
		})
	}
	if len(loader.Live.InlineRoomIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: entryFrom,
				RoomIds:   loader.Live.InlineRoomIDs,
				Uid:       loader.General.Mid,
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
				Platform:  loader.General.GetPlatform(),
				Build:     loader.General.GetBuild(),
			}
			entryRoom, err := loader.Service.dao.EntryRoomInfo(ctx, req)
			if err != nil {
				return errors.Wrapf(err, "Failed to get entry room info: %+v: %+v", req, err)
			}
			fanout.Live.InlineRoom = entryRoom
			return nil
		})
	}
}

func (loader *NewInlineCardFanoutLoader) doAccount(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.Account.AccountUIDs) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		res, err := loader.Service.dao.Cards3(ctx, loader.Account.AccountUIDs)
		if err != nil {
			return errors.Wrapf(err, "doAccount loader.accDao.Cards3 uids=%+v, error= %+v", loader.Account.AccountUIDs, err)
		}
		fanout.Account.Card = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		fanout.Account.IsAttention = loader.Service.dao.Relations3(ctx, loader.Account.AccountUIDs, loader.General.Mid)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		res, err := loader.Service.getNFTIconInfo(ctx, loader.Account.AccountUIDs)
		if err != nil {
			log.Warn("doAccount loader.getNFTIconInfo uids=%+v, error= %+v", loader.Account.AccountUIDs, err)
			return nil
		}
		fanout.NftRegion = res
		return nil
	})
}

func (loader *NewInlineCardFanoutLoader) doSecondLoader(ctx context.Context, fanout *FanoutResult) error {
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

func (loader *NewInlineCardFanoutLoader) doHasLike(eg *errgroup.Group, fanout *FanoutResult) {
	if len(loader.ThumbUp.ArchiveAid) <= 0 || loader.General.Mid == 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		fanoutLikes, err := loader.Service.dao.HasLike(ctx, loader.General.GetBuvid(), loader.General.Mid, loader.ThumbUp.ArchiveAid)
		if err != nil {
			return errors.Wrapf(err, "doSecondLoader hasLike aids=%+v, err=%+v", loader.Archive.Aids, err)
		}
		fanout.ThumbUp.HasLikeArchive = fanoutLikes
		return nil
	})
}

func (loader *NewInlineCardFanoutLoader) WithAccountProfile(mid ...int64) {
	loader.Account.AccountUIDs = append(loader.Account.AccountUIDs, mid...)
}

func (loader *NewInlineCardFanoutLoader) WithThumbUpArchive(aid ...int64) {
	loader.ThumbUp.ArchiveAid = append(loader.ThumbUp.ArchiveAid, aid...)
}

type loaderArchiveSubset struct {
	Aids []int64
}

type loaderLiveSubset struct {
	UpMids        []int64
	InlineRoomIDs []int64
	LiveEntryFrom []string
}

type loaderAccountSubset struct {
	AccountUIDs []int64
}

type loaderThumbUpSubset struct {
	ArchiveAid []int64
}

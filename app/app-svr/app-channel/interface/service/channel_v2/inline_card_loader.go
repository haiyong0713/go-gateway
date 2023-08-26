package channel_v2

import (
	"context"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	xmetadata "go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	jsonlargecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	"github.com/pkg/errors"
)

// 新搜索inline卡片物料装载器
type NewInlineCardFanoutLoader struct {
	General *topiccardmodel.GeneralParam
	Service *Service
	Archive loaderArchiveSubset
	Account loaderAccountSubset
	ThumbUp loaderThumbUpSubset
}

func (loader *NewInlineCardFanoutLoader) setGeneralParamFromCtx(ctx context.Context) {
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	loader.General = &topiccardmodel.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
	}
}

func (loader *NewInlineCardFanoutLoader) doChannelInlineCardFanoutLoad(ctx context.Context) (*feedcard.FanoutResult, error) {
	fanout := &feedcard.FanoutResult{Inline: &jsonlargecover.Inline{
		LikeButtonShowCount:      true,
		LikeResource:             "http://i0.hdslb.com/bfs/archive/b9f49c9b33532c5d05f5ea701ecd063f81910e94.json",
		LikeResourceHash:         "c8b42c2a76890e703b15874175268b4b",
		DisLikeResource:          "http://i0.hdslb.com/bfs/archive/8aee6952487d118b4207c1afa2fd38616bd7545a.json",
		DisLikeResourceHash:      "bdbc35ebc88d178d1f409145dadec806",
		LikeNightResource:        "http://i0.hdslb.com/bfs/archive/3ed718f59e9e9cf1ce148105c9db9559951d5a7d.json",
		LikeNightResourceHash:    "bc9fecf2624a569c05cef8097e20eb37",
		DisLikeNightResource:     "http://i0.hdslb.com/bfs/archive/c9a20055b712068bfe293878639dc9066ba2690b.json",
		DisLikeNightResourceHash: "c370e8d031381f4716d7564956a8b182",
		IconDrag:                 "http://i0.hdslb.com/bfs/archive/c1461e2c6ca97783ac0298b6ebb2d85d94b8f37c.json",
		IconDragHash:             "31df8ce99de871afaa66a7a78f44deec",
		IconStop:                 "http://i0.hdslb.com/bfs/archive/6ee2f9b016f20714705cb5b8f15da1446587d172.json",
		IconStopHash:             "5648c2926c1c93eb2d30748994ba7b96",
		ThreePointPanelType:      1,
	}}
	eg := errgroup.WithContext(ctx)
	loader.doArchive(eg, fanout)
	if err := eg.Wait(); err != nil {
		return nil, errors.Wrapf(err, "Failed to execute doChannelInlineCardFanoutLoad error group: %+v", loader)
	}
	if err := loader.doSecondLoader(ctx, fanout); err != nil {
		return nil, errors.Wrapf(err, "Failed to execute doSecondLoader loader=%+v", loader)
	}
	return fanout, nil
}

func (loader *NewInlineCardFanoutLoader) doArchive(eg *errgroup.Group, fanout *feedcard.FanoutResult) {
	if len(loader.Archive.Aids) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		var aids []*archivegrpc.PlayAv
		for _, v := range loader.Archive.Aids {
			aids = append(aids, &archivegrpc.PlayAv{Aid: v})
		}
		res, err := loader.Service.arcDao.ArcsPlayer(ctx, aids, false)
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
		fanoutcoins, err := loader.Service.coinDao.ArchiveUserCoins(ctx, loader.Archive.Aids, loader.General.Mid)
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
		fanoutFavs, err := loader.Service.favDao.IsFavVideos(ctx, loader.General.Mid, loader.Archive.Aids)
		if err != nil {
			return errors.Wrapf(err, "NewInlineCardFanoutLoader isFavVideos aids=%+v, err=%+v", loader.Archive.Aids, err)
		}
		fanout.Favourite = fanoutFavs
		return nil
	})
}

func (loader *NewInlineCardFanoutLoader) doAccount(eg *errgroup.Group, fanout *feedcard.FanoutResult) {
	if len(loader.Account.AccountUIDs) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		res, err := loader.Service.accDao.Cards3GRPC(ctx, loader.Account.AccountUIDs)
		if err != nil {
			return errors.Wrapf(err, "doAccount loader.accDao.Cards3 uids=%+v, error= %+v", loader.Account.AccountUIDs, err)
		}
		fanout.Account.Card = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		fanout.Account.IsAttention = resolveFanoutIsAttention(loader.Service.accDao.Relations3GRPC(ctx, loader.Account.AccountUIDs, loader.General.Mid))
		return nil
	})
}

func resolveFanoutIsAttention(follows map[int64]bool) map[int64]int8 {
	m := make(map[int64]int8, len(follows))
	for k, v := range follows {
		var isAttention int8
		if v {
			isAttention = 1
		}
		m[k] = isAttention
	}
	return m
}

func (loader *NewInlineCardFanoutLoader) doSecondLoader(ctx context.Context, fanout *feedcard.FanoutResult) error {
	for _, a := range fanout.Archive.Archive {
		loader.WithAccountProfile(a.Arc.Author.Mid)
		loader.WithThumbUpArchive(a.Arc.Aid)
	}
	eg := errgroup.WithContext(ctx)
	loader.doAccount(eg, fanout)
	loader.doHasLike(eg, fanout)
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (loader *NewInlineCardFanoutLoader) doHasLike(eg *errgroup.Group, fanout *feedcard.FanoutResult) {
	if len(loader.ThumbUp.ArchiveAid) <= 0 || loader.General.Mid == 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		fanoutLikes, err := loader.Service.thumbupDao.HasLike(ctx, loader.General.Mid, loader.ThumbUp.ArchiveAid)
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

type loaderAccountSubset struct {
	AccountUIDs []int64
}

type loaderThumbUpSubset struct {
	ArchiveAid []int64
}

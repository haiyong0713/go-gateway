package service

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	bannermodel "go-gateway/app/app-svr/app-card/interface/model/card/banner"
	shopping "go-gateway/app/app-svr/app-card/interface/model/card/show"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/dao"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/model"
	appfeedmodel "go-gateway/app/app-svr/app-feed/interface/model"
	feedmdl "go-gateway/app/app-svr/app-feed/interface/model"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	resmodel "go-gateway/app/app-svr/resource/service/model"

	"go-common/library/sync/errgroup.v2"

	articlegrpc "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
)

const (
	PreferGIFTypeOperator      = "operator_gif"
	PreferGIFTypeAdvertisement = "advertisement_gif"
	// ad pk code
	_adPkGifCard = "gif"
	_adPkBigCard = "banner"
)

type fanoutDependency = dao.FanoutDependency

type fanoutCommon struct {
	CurrentMid    int64
	Device        cardschema.Device
	FeedParam     feedcard.IndexParam
	PreferGIFType string
}

type loaderArchiveSubset struct {
	Aid      []int64
	StoryAid []int64
}
type loaderLiveSubset struct {
	RoomID       []int64
	InlineRoomID []int64
}
type loaderDynamicSubset struct {
	Picture []int64
}
type loaderBangumiSubset struct {
	EPID        []int64
	SeasonID    []int32
	SeasonByAid []int32
	PlayerIDs   []int32
}
type loaderThumbUpSubset struct {
	ArchiveAid []int64
}
type loaderAccountSubset struct {
	CardMid         []int64
	RelationStatMid []int64
	IsAttentionMid  []int64
}

type feedFanoutLoader struct {
	fanoutCommon

	Archive          loaderArchiveSubset
	TagID            []int64
	Live             loaderLiveSubset
	ArticleID        []int64
	AudioID          []int64
	Dynamic          loaderDynamicSubset
	Bangumi          loaderBangumiSubset
	ChannelID        []int64
	TunnelGatherOids [][]int64
	ThumbUp          loaderThumbUpSubset
	Account          loaderAccountSubset
	BannerResourceID int64
	ShopID           []int64
	PosRecID         []int64
	FavouriteAid     []int64

	dep fanoutDependency
}

// DeriveEmpty will derive a loader with `fanoutCommon` and empty targets.
func (ffl *feedFanoutLoader) DeriveEmpty() *feedFanoutLoader {
	out := &feedFanoutLoader{
		fanoutCommon: ffl.fanoutCommon,
	}
	return out
}
func (ffl *feedFanoutLoader) WithArchive(aid ...int64) {
	ffl.Archive.Aid = append(ffl.Archive.Aid, aid...)
}
func (ffl *feedFanoutLoader) WithStoryArchive(storyAid ...int64) {
	ffl.Archive.StoryAid = append(ffl.Archive.StoryAid, storyAid...)
}
func (ffl *feedFanoutLoader) WithTag(tagID ...int64) {
	ffl.TagID = append(ffl.TagID, tagID...)
}
func (ffl *feedFanoutLoader) WithLiveRoom(roomID ...int64) {
	ffl.Live.RoomID = append(ffl.Live.RoomID, roomID...)
}
func (ffl *feedFanoutLoader) WithInlineLiveRoom(roomID ...int64) {
	ffl.Live.InlineRoomID = append(ffl.Live.InlineRoomID, roomID...)
}
func (ffl *feedFanoutLoader) WithArticle(articleID ...int64) {
	ffl.ArticleID = append(ffl.ArticleID, articleID...)
}
func (ffl *feedFanoutLoader) WithAudio(audioID ...int64) {
	ffl.AudioID = append(ffl.AudioID, audioID...)
}
func (ffl *feedFanoutLoader) WithPicture(dynamicID ...int64) {
	ffl.Dynamic.Picture = append(ffl.Dynamic.Picture, dynamicID...)
}
func (ffl *feedFanoutLoader) WithBangumiEP(epID ...int64) {
	ffl.Bangumi.EPID = append(ffl.Bangumi.EPID, epID...)
}
func (ffl *feedFanoutLoader) WithBangumiSeason(seasonID ...int32) {
	ffl.Bangumi.SeasonID = append(ffl.Bangumi.SeasonID, seasonID...)
}
func (ffl *feedFanoutLoader) WithBangumiSeasonAid(aid ...int32) {
	ffl.Bangumi.SeasonByAid = append(ffl.Bangumi.SeasonByAid, aid...)
}
func (ffl *feedFanoutLoader) WithBangumiPlayerIDs(pid ...int32) {
	ffl.Bangumi.PlayerIDs = append(ffl.Bangumi.PlayerIDs, pid...)
}
func (ffl *feedFanoutLoader) WithChannel(channelID ...int64) {
	ffl.ChannelID = append(ffl.ChannelID, channelID...)
}
func (ffl *feedFanoutLoader) WithTunnelFeed(tunnelID ...int64) {
	for _, id := range tunnelID {
		ffl.TunnelGatherOids = append(ffl.TunnelGatherOids, []int64{id})
	}
}
func (ffl *feedFanoutLoader) WithThumbUpArchive(aid ...int64) {
	ffl.ThumbUp.ArchiveAid = append(ffl.ThumbUp.ArchiveAid, aid...)
}
func (ffl *feedFanoutLoader) WithAccountProfile(mid ...int64) {
	ffl.WithAccountCard(mid...)
	ffl.WithAccountRelationStat(mid...)
	ffl.WithIsAttentionMid(mid...)
}
func (ffl *feedFanoutLoader) WithAccountCard(mid ...int64) {
	ffl.Account.CardMid = append(ffl.Account.CardMid, mid...)
}
func (ffl *feedFanoutLoader) WithAccountRelationStat(mid ...int64) {
	ffl.Account.RelationStatMid = append(ffl.Account.RelationStatMid, mid...)
}
func (ffl *feedFanoutLoader) WithIsAttentionMid(mid ...int64) {
	ffl.Account.IsAttentionMid = append(ffl.Account.IsAttentionMid, mid...)
}
func (ffl *feedFanoutLoader) WithPosRec(posRecID ...int64) {
	ffl.PosRecID = append(ffl.PosRecID, posRecID...)
}

func (ffl *feedFanoutLoader) WithShop(shopID ...int64) {
	ffl.ShopID = append(ffl.ShopID, shopID...)
}

func (ffl *feedFanoutLoader) WithBannerResourceID() {
	plat := ffl.Device.Plat()
	resourceID := _bannersResourceID[plat]
	if feedcard.UsingNewBanner(ffl.Device) {
		if tmpID, ok := _bannersResourceIDByABtest[plat]; ok {
			resourceID = tmpID
		}
	}
	if ffl.FeedParam.LessonsMode == 1 {
		if tmpID, ok := _bannersResourceIDByLesson[plat]; ok {
			resourceID = tmpID
		}
	}
	ffl.BannerResourceID = resourceID
}

func (ffl *feedFanoutLoader) WithFavourite(aid ...int64) {
	ffl.FavouriteAid = append(ffl.FavouriteAid, aid...)
}

func (ffl *feedFanoutLoader) Load(ctx context.Context, dep fanoutDependency) (*feedcard.FanoutResult, error) {
	ffl.dep = dep
	out := &feedcard.FanoutResult{}
	eg := errgroup.WithContext(ctx)
	ffl.doArchive(eg, out)
	ffl.doTag(eg, &out.Tag)
	ffl.doChannel(eg, &out.Channel)
	ffl.doThumbUp(eg, &out.ThumbUp.HasLikeArchive)
	ffl.doLive(eg, out)
	ffl.doArticle(eg, &out.Article)
	ffl.doAudio(eg, &out.Audio)
	ffl.doDynamic(eg, &out.Dynamic.Picture)
	ffl.doBangumi(eg, out)
	ffl.doAccount(eg, out)
	ffl.doBanner(eg, out)
	ffl.doInline(out)
	ffl.doFollowMode(out)
	ffl.doStoryIcon(out)
	ffl.doShop(eg, &out.Shop)
	ffl.doVip(eg, out)
	ffl.doTunnel(eg, &out.Tunnel)
	ffl.doFavourite(eg, out)
	ffl.doCoin(eg, out)

	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute feed fanout error group: %+v", err)
		return nil, err
	}
	return out, nil
}

func (ffl *feedFanoutLoader) doAccount(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if len(ffl.Account.CardMid) > 0 {
		eg.Go(func(ctx context.Context) error {
			reply, err := ffl.dep.Account().Cards3GRPC(ctx, ffl.Account.CardMid)
			if err != nil {
				log.Error("Failed to request account card: %+v", err)
				return nil
			}
			out.Account.Card = reply
			return nil
		})
	}
	if len(ffl.Account.RelationStatMid) > 0 {
		eg.Go(func(ctx context.Context) error {
			reply, err := ffl.dep.Relation().StatsGRPC(ctx, ffl.Account.RelationStatMid)
			if err != nil {
				log.Error("Failed to request relation stat: %+v", err)
				return nil
			}
			out.Account.RelationStatMid = reply
			return nil
		})
	}
	if len(ffl.Account.IsAttentionMid) > 0 {
		eg.Go(func(ctx context.Context) error {
			out.Account.IsAttention = ffl.dep.Account().IsAttentionGRPC(ctx, ffl.Account.IsAttentionMid, ffl.CurrentMid)
			return nil
		})
	}
}

var (
	_bannersResourceID = map[int8]int64{
		appfeedmodel.PlatIPhoneB:  467,
		appfeedmodel.PlatIPhone:   467,
		appfeedmodel.PlatAndroid:  631,
		appfeedmodel.PlatIPad:     771,
		appfeedmodel.PlatIPhoneI:  947,
		appfeedmodel.PlatAndroidG: 1285,
		appfeedmodel.PlatAndroidI: 1707,
		appfeedmodel.PlatIPadI:    1117,
	}
	// abtest
	_bannersResourceIDByABtest = map[int8]int64{
		appfeedmodel.PlatIPhone:  3143,
		appfeedmodel.PlatAndroid: 3150,
		appfeedmodel.PlatIPad:    3179,
	}
	_bannersResourceIDByLesson = map[int8]int64{
		appfeedmodel.PlatIPhone:  3848,
		appfeedmodel.PlatAndroid: 3852,
		appfeedmodel.PlatIPad:    3856,
	}
)

func (ffl *feedFanoutLoader) doBanner(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if ffl.BannerResourceID <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		hash := ffl.FeedParam.BannerHash
		if ffl.FeedParam.LoginEvent != 0 {
			hash = ""
		}
		data, version, err := ffl.dep.Resource().Banner(ctx, &resmodel.ArgBanner{
			Plat:      ffl.Device.Plat(),
			Build:     int(ffl.Device.Build()),
			MID:       ffl.CurrentMid,
			ResIDs:    strconv.FormatInt(ffl.BannerResourceID, 10),
			Buvid:     ffl.Device.Buvid(),
			Network:   ffl.Device.Network(),
			MobiApp:   ffl.Device.RawMobiApp(),
			Device:    ffl.Device.Device(),
			IsAd:      true,
			OpenEvent: ffl.FeedParam.OpenEvent,
			AdExtra:   ffl.FeedParam.AdExtra,
			Version:   hash,
			SplashID:  ffl.FeedParam.SplashID,
		})
		if err != nil {
			log.Error("Failed to request banner: %+v", err)
			return nil
		}
		banner := make([]*bannermodel.Banner, 0, len(data))
		for _, rb := range data[int(ffl.BannerResourceID)] {
			b := &bannermodel.Banner{}
			b.Change(rb)
			banner = append(banner, b)
		}
		out.Banner.Banners = banner
		out.Banner.Version = version
		return nil
	})
}

func (ffl *feedFanoutLoader) doBangumi(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if len(ffl.Bangumi.EPID) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &model.EpPlayerReq{
				EpIDs:    ffl.Bangumi.EPID,
				MobiApp:  ffl.Device.RawMobiApp(),
				Platform: ffl.Device.RawPlatform(),
				Device:   ffl.Device.Device(),
				Build:    int(ffl.Device.Build()),
			}
			if batchArg, ok := arcmid.FromContext(ctx); ok {
				req.Fnver = int(batchArg.Fnver)
				req.Fnval = int(batchArg.Fnval)
			}
			reply, err := ffl.dep.Bangumi().EpPlayer(ctx, req)
			if err != nil {
				log.Error("Failed to request bangumi ep player: %+v", err)
				return nil
			}
			out.Bangumi.EP = reply
			return nil
		})
	}
	if len(ffl.Bangumi.SeasonID) > 0 {
		eg.Go(func(ctx context.Context) error {
			reply, err := ffl.dep.Bangumi().CardsInfoReply(ctx, ffl.Bangumi.SeasonID)
			if err != nil {
				log.Error("Failed to request bangumi season: %+v", err)
				return nil
			}
			out.Bangumi.Season = reply
			return nil
		})
	}
	if len(ffl.Bangumi.SeasonByAid) > 0 {
		eg.Go(func(ctx context.Context) error {
			reply, err := ffl.dep.Bangumi().CardsByAids(ctx, ffl.Bangumi.SeasonByAid)
			if err != nil {
				log.Error("Failed to request bangumi season by aid: %+v", err)
				return nil
			}
			out.Bangumi.SeasonByAid = reply
			return nil
		})
	}
	if len(ffl.Bangumi.PlayerIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &pgcinline.EpReq{
				EpIds: ffl.Bangumi.PlayerIDs,
				User: &pgcinline.UserReq{
					Mid:      ffl.CurrentMid,
					MobiApp:  ffl.Device.RawMobiApp(),
					Device:   ffl.Device.Device(),
					Platform: ffl.Device.RawPlatform(),
					Ip:       metadata.String(ctx, metadata.RemoteIP),
					Build:    int32(ffl.Device.Build()),
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
			reply, err := ffl.dep.Bangumi().InlineCards(ctx, req)
			if err != nil {
				log.Error("Failed to request bangumi season by aid: %+v", err)
				return nil
			}
			out.Bangumi.InlinePGC = reply
			return nil
		})
	}
	if ffl.fanoutCommon.CurrentMid > 0 {
		eg.Go(func(ctx context.Context) error {
			reply, err := ffl.dep.Bangumi().Remind(ctx, ffl.fanoutCommon.CurrentMid)
			if err != nil {
				log.Error("Failed to request bangumi remind: %+v", err)
				return nil
			}
			out.Bangumi.Remind = reply
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			reply, err := ffl.dep.Bangumi().Updates(ctx, ffl.fanoutCommon.CurrentMid, time.Now())
			if err != nil {
				log.Error("Failed to request bangumi updates: %+v", err)
				return nil
			}
			out.Bangumi.Update = reply
			return nil
		})
	}
}

func (ffl *feedFanoutLoader) doDynamic(eg *errgroup.Group, out *map[int64]*bplus.Picture) {
	if len(ffl.Dynamic.Picture) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		req := &model.DynamicDetailReq{
			Platfrom: ffl.Device.RawPlatform(),
			MobiApp:  ffl.Device.RawMobiApp(),
			Device:   ffl.Device.Device(),
			Build:    strconv.Itoa(int(ffl.Device.Build())),
		}
		reply, err := ffl.dep.Dynamic().DynamicDetail(ctx, req, ffl.Dynamic.Picture...)
		if err != nil {
			log.Error("Failed to request dynamic: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doAudio(eg *errgroup.Group, out *map[int64]*audio.Audio) {
	if len(ffl.AudioID) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Audio().Audios(ctx, ffl.AudioID)
		if err != nil {
			log.Error("Failed to request audio: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doArticle(eg *errgroup.Group, out *map[int64]*articlegrpc.Meta) {
	if len(ffl.ArticleID) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Article().Articles(ctx, ffl.ArticleID)
		if err != nil {
			log.Error("Failed to request article: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doArchive(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if len(ffl.Archive.Aid) > 0 {
		eg.Go(func(ctx context.Context) error {
			ffl.arcsPlayer(ctx, "", ffl.Archive.Aid, &out.Archive.Archive)
			return nil
		})
	}
	if len(ffl.Archive.StoryAid) > 0 {
		eg.Go(func(ctx context.Context) error {
			ffl.arcsPlayer(ctx, "story", ffl.Archive.StoryAid, &out.Archive.StoryArchive)
			return nil
		})
	}
}

func (ffl *feedFanoutLoader) doChannel(eg *errgroup.Group, out *map[int64]*channelgrpc.ChannelCard) {
	if len(ffl.ChannelID) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Channel().Details(ctx, ffl.ChannelID)
		if err != nil {
			log.Error("Failed to request channel: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doTag(eg *errgroup.Group, out *map[int64]*taggrpc.Tag) {
	if len(ffl.TagID) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		tags, err := ffl.dep.Tag().Tags(ctx, ffl.CurrentMid, ffl.TagID)
		if err != nil {
			log.Error("Failed to request tag: %+v", err)
			return nil
		}
		*out = tags
		return nil
	})
}

func (ffl *feedFanoutLoader) doThumbUp(eg *errgroup.Group, out *map[int64]int8) {
	if len(ffl.ThumbUp.ArchiveAid) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.ThumbUp().HasLike(ctx, ffl.Device.Buvid(), ffl.CurrentMid, ffl.ThumbUp.ArchiveAid)
		if err != nil {
			log.Error("Failed to request thumbup: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doLive(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if len(ffl.Live.RoomID) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &model.AppMRoomReq{
				RoomIds:        ffl.Live.RoomID,
				Mid:            ffl.CurrentMid,
				Platform:       ffl.Device.RawPlatform(),
				DeviceName:     ffl.FeedParam.DeviceName,
				AccessKey:      ffl.FeedParam.AccessKey,
				ActionKey:      ffl.FeedParam.ActionKey,
				Appkey:         ffl.FeedParam.AppKey,
				Device:         ffl.Device.Device(),
				MobiApp:        ffl.Device.RawMobiApp(),
				Statistics:     ffl.FeedParam.Statistics,
				Buvid:          ffl.Device.Buvid(),
				Network:        ffl.Device.Network(),
				Build:          int(ffl.Device.Build()),
				TeenagersMode:  ffl.FeedParam.TeenagersMode,
				Appver:         ffl.FeedParam.Appver,
				Filtered:       ffl.FeedParam.Filtered,
				HttpsUrlReq:    ffl.FeedParam.HttpsUrlReq,
				NeedRoomFilter: 0,
			}
			reply, err := ffl.dep.Live().AppMRoom(ctx, req)
			if err != nil {
				log.Error("Failed to request live: %+v", err)
				return nil
			}
			out.Live.Room = reply
			return nil
		})
	}
	if len(ffl.Live.InlineRoomID) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &model.AppMRoomReq{
				RoomIds:        ffl.Live.InlineRoomID,
				Mid:            ffl.CurrentMid,
				Platform:       ffl.Device.RawPlatform(),
				DeviceName:     ffl.FeedParam.DeviceName,
				AccessKey:      ffl.FeedParam.AccessKey,
				ActionKey:      ffl.FeedParam.ActionKey,
				Appkey:         ffl.FeedParam.AppKey,
				Device:         ffl.Device.Device(),
				MobiApp:        ffl.Device.RawMobiApp(),
				Statistics:     ffl.FeedParam.Statistics,
				Buvid:          ffl.Device.Buvid(),
				Network:        ffl.Device.Network(),
				Build:          int(ffl.Device.Build()),
				TeenagersMode:  ffl.FeedParam.TeenagersMode,
				Appver:         ffl.FeedParam.Appver,
				Filtered:       ffl.FeedParam.Filtered,
				HttpsUrlReq:    ffl.FeedParam.HttpsUrlReq,
				NeedRoomFilter: 1,
			}
			reply, err := ffl.dep.Live().AppMRoom(ctx, req)
			if err != nil {
				log.Error("Failed to request live: %+v", err)
				return nil
			}
			out.Live.InlineRoom = reply
			return nil
		})
	}
}

func (ffl *feedFanoutLoader) arcsPlayer(ctx context.Context, from string, aids []int64, out *map[int64]*arcgrpc.ArcPlayer) {
	playAvs := make([]*arcgrpc.PlayAv, 0, len(aids))
	for _, aid := range aids {
		item := &arcgrpc.PlayAv{
			Aid: aid,
		}
		playAvs = append(playAvs, item)
	}
	arg := &arcgrpc.ArcsPlayerRequest{
		PlayAvs: playAvs,
	}
	if batchArg, ok := arcmid.FromContext(ctx); ok {
		duplicateBatchArg := *batchArg
		duplicateBatchArg.From = from
		arg.BatchPlayArg = &duplicateBatchArg
	}
	reply, err := ffl.dep.Archive().ArcsPlayer(ctx, arg)
	if err != nil {
		log.Error("Failed to request archvie: %+v", err)
		return
	}
	*out = reply
	//nolint:gosimple
	return
}

func (ffl *feedFanoutLoader) doInline(out *feedcard.FanoutResult) {
	out.Inline = ffl.dep.Inline()
}

func (ffl *feedFanoutLoader) doShop(eg *errgroup.Group, out *map[int64]*shopping.Shopping) {
	if len(ffl.ShopID) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Shop().Card(ctx, ffl.ShopID)
		if err != nil {
			log.Error("Failed to request shop: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doVip(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if ffl.CurrentMid <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		platform := int64(0)
		switch {
		case feedmdl.IsPad(ffl.Device.Plat()):
			platform = int64(2)
		case feedmdl.IsIOS(ffl.Device.Plat()):
			platform = int64(1)
		case feedmdl.IsAndroid(ffl.Device.Plat()):
			platform = int64(4)
		default:
			log.Warn("No match plat: %d", ffl.Device.Plat())
		}
		reply, err := ffl.dep.Vip().TipsRenew(ctx, ffl.Device.Build(), platform, ffl.CurrentMid)
		if err != nil {
			log.Error("Failed to request vip: %+v", err)
			return nil
		}
		out.Vip = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doFollowMode(out *feedcard.FanoutResult) {
	out.FollowMode = ffl.dep.FollowMode()
}

func (ffl *feedFanoutLoader) doStoryIcon(out *feedcard.FanoutResult) {
	out.StoryIcon = ffl.dep.StoryIcon()
}

func (ffl *feedFanoutLoader) doTunnel(eg *errgroup.Group, out *map[int64]*tunnelgrpc.FeedCard) {
	if len(ffl.TunnelGatherOids) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Tunnel().FeedCards(ctx, &tunnelgrpc.FeedCardsReq{
			Platform:   ffl.Device.RawMobiApp(),
			Mid:        ffl.CurrentMid,
			Build:      ffl.Device.Build(),
			MobiApp:    ffl.Device.RawMobiApp(),
			GatherOids: constructGatherOids(ffl.TunnelGatherOids),
		})
		if err != nil {
			log.Error("Failed to request tunnel: %+v", err)
			return nil
		}
		*out = reply
		return nil
	})
}

func constructGatherOids(gatherOids [][]int64) []*tunnelgrpc.FeedCardsReqGather {
	out := make([]*tunnelgrpc.FeedCardsReqGather, 0, len(gatherOids))
	for _, oids := range gatherOids {
		out = append(out, &tunnelgrpc.FeedCardsReqGather{Oids: oids})
	}
	return out
}

func (ffl *feedFanoutLoader) doFavourite(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if len(ffl.FavouriteAid) <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Favourite().IsFavVideos(ctx, ffl.CurrentMid, ffl.FavouriteAid)
		if err != nil {
			log.Error("Failed to request favourite: %+v", err)
			return nil
		}
		out.Favourite = reply
		return nil
	})
}

func (ffl *feedFanoutLoader) doCoin(eg *errgroup.Group, out *feedcard.FanoutResult) {
	if len(ffl.ThumbUp.ArchiveAid) <= 0 || ffl.CurrentMid <= 0 {
		return
	}
	eg.Go(func(ctx context.Context) error {
		reply, err := ffl.dep.Coin().ArchiveUserCoins(ctx, ffl.ThumbUp.ArchiveAid, ffl.CurrentMid)
		if err != nil {
			log.Error("Failed to request coin: %+v", err)
			return nil
		}
		out.Coin = reply
		return nil
	})
}

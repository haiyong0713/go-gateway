package view

import (
	"context"
	"strconv"
	"sync"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	appecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/dao/archive"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	grpcmetadata "google.golang.org/grpc/metadata"
)

type FeedItemParams struct {
	Now         time.Time
	FeedViewReq *viewApi.FeedViewReq
	CurrentAid  int64
	Restriction restriction.Restriction
	Device      device.Device
	AuthN       auth.Auth
	Plat        int8
	IP          string
	Network     network.Network
	Locale      locale.Locale
	IsMelloi    string
	UserZoneID  int64

	SlideItem *archive.SlidesItem
}

func (fip *FeedItemParams) DuplicateWith(slideItem *archive.SlidesItem) *FeedItemParams {
	return &FeedItemParams{
		Now:         fip.Now,
		FeedViewReq: fip.FeedViewReq,
		CurrentAid:  slideItem.ID,
		Restriction: fip.Restriction,
		Device:      fip.Device,
		AuthN:       fip.AuthN,
		Plat:        fip.Plat,
		IP:          fip.IP,
		Network:     fip.Network,
		Locale:      fip.Locale,
		IsMelloi:    fip.IsMelloi,
		UserZoneID:  fip.UserZoneID,
		SlideItem:   slideItem,
	}
}

func bool2int(in bool) int {
	if in {
		return 1
	}
	return 0
}

func (fip *FeedItemParams) DisableRcmdMode() int {
	return bool2int(fip.Restriction.DisableRcmd)
}

func (fip *FeedItemParams) TeenagersMode() int {
	return bool2int(fip.Restriction.IsTeenagers)
}

func (fip *FeedItemParams) LessonsMode() int {
	return bool2int(fip.Restriction.IsLessons)
}

func (fip *FeedItemParams) Filterd() string {
	return strconv.FormatInt(int64(bool2int(fip.Restriction.IsReview)), 10)
}

func (fip *FeedItemParams) LocalID() (string, string) {
	slocale, clocale := i18n.ConstructLocaleID(fip.Locale)
	return slocale, clocale
}

func (fip *FeedItemParams) Net() string {
	if fip.Network.Type == network.TypeCellular {
		return "mobile"
	} else if fip.Network.Type == network.TypeWIFI {
		return "wifi"
	}
	return ""
}

func wrapViewReplyItem(in *viewApi.ViewReply, slideItem *archive.SlidesItem) *viewApi.FeedViewItem {
	return &viewApi.FeedViewItem{
		View:    in,
		Goto:    slideItem.Goto,
		Uri:     "",
		TrackId: slideItem.TrackID,
	}
}

func (s *Service) requestFeedViewItem(ctx context.Context, args *FeedItemParams) (*viewApi.FeedViewItem, error) {
	vp, extra, err := s.ArcView(ctx, args.CurrentAid, 0, "", "", "", 0)
	if err != nil {
		return nil, err
	}
	slocale, clocale := args.LocalID()
	// 不在白名单内或获取合集配置失败 降级普通播放页
	if s.displayActSeason(vp, args.TeenagersMode(), args.LessonsMode(), int(args.Device.Build), args.Device.RawMobiApp, args.FeedViewReq.Spmid, args.AuthN.Mid, args.Plat) {
		res, err := s.ActivitySeason(ctx, args.AuthN.Mid, args.CurrentAid, args.Plat,
			int(args.Device.Build), 0, args.Device.RawMobiApp,
			args.Device.Device, args.Device.Buvid, args.IP, args.Network.WebcdnIP,
			args.Net(), "", args.FeedViewReq.From, args.FeedViewReq.Spmid,
			args.FeedViewReq.FromSpmid, args.Device.RawPlatform,
			args.Filterd(), args.IsMelloi, args.Device.Brand, slocale, clocale,
			args.FeedViewReq.FromTrackId, args.FeedViewReq.PageVersion, args.Now, vp, args.DisableRcmdMode(), extra)
		if err == nil {
			s.prom.Incr("展示活动合集")
			HideArcAttribute(res.GetActivitySeason().GetArc())
			return wrapViewReplyItem(res, args.SlideItem), nil
		}
		log.Error("ActivitySeason sid(%d) aid(%d) s.ActivitySeason err(%+v)", vp.SeasonID, vp.Aid, err)
		if err != appecode.AppActivitySeasonFallback {
			return nil, err
		}
		s.prom.Incr("降级普通合集")
	}
	v, err := s.ViewInfo(ctx, args.AuthN.Mid, args.CurrentAid, args.Plat, int(args.Device.Build), 0,
		0, args.TeenagersMode(), args.LessonsMode(), args.Device.RawMobiApp,
		args.Device.Device, args.Device.Buvid, args.Network.WebcdnIP, args.Net(), "", args.FeedViewReq.From,
		args.FeedViewReq.Spmid, args.FeedViewReq.FromSpmid, "", args.Device.RawPlatform, args.Filterd(), "1",
		true, args.IsMelloi, args.Device.Brand, slocale, clocale, args.FeedViewReq.PageVersion, vp, args.DisableRcmdMode(), 0, 0, "", "", 0, 0, 0, extra)
	if err != nil {
		return nil, err
	}
	v.DislikeReasons(ctx, s.c.Feature, args.Device.RawMobiApp, args.Device.Device, int(args.Device.Build), args.DisableRcmdMode())
	viewReply := &viewApi.ViewReply{
		Arc:               v.Arc,
		Pages:             view.FromPages(v.Pages),
		OwnerExt:          view.FromOwnerExt(v.OwnerExt),
		ReqUser:           v.ReqUser,
		Tag:               view.FromTag(v.Tag),
		TIcon:             v.TIcon,
		Season:            view.FromSeason(v.Season),
		ElecRank:          v.ElecRank,
		History:           v.History,
		Relates:           view.FromRelates(v.Relates),
		Dislike:           v.DislikeV2,
		PlayerIcon:        view.FromPlayerIcon(v.PlayerIcon),
		VipActive:         v.VIPActive,
		Bvid:              v.BvID,
		Honor:             v.Honor,
		RelateTab:         v.RelateTab,
		ActivityUrl:       v.ActivityURL,
		Bgm:               v.Bgm,
		Staff:             view.FromStaff(v.Staff),
		ArgueMsg:          v.ArgueMsg,
		ShortLink:         v.ShortLink,
		PlayParam:         int32(v.PlayParam),
		Label:             v.Label,
		UgcSeason:         view.FromUgcSeason(v.UgcSeason),
		Config:            view.FromConfig(v.Config),
		ShareSubtitle:     v.ShareSubtitle,
		Interaction:       v.Interaction,
		Cms:               v.CMSNew,
		CmConfig:          v.CMConfigNew,
		Rank:              v.Rank,
		TfPanelCustomized: v.TfPanelCustomized,
		UpAct:             v.UpAct,
		UserGarb:          v.UserGarb,
		BadgeUrl:          v.BadgeUrl,
		LiveOrderInfo:     v.LiveOrderInfo,
		DescV2:            v.DescV2,
		Sticker:           v.Sticker,
		CmIpad:            v.IPadCM,
		UpLikeImg:         v.UpLikeImg,
	}
	return wrapViewReplyItem(viewReply, args.SlideItem), nil
}

func calcIsMelloi(ctx context.Context) string {
	var isMelloi string
	if gmd, ok := grpcmetadata.FromIncomingContext(ctx); ok {
		if values := gmd.Get("x-melloi"); len(values) > 0 {
			isMelloi = values[0]
		}
	}
	return isMelloi
}

func (s *Service) makeSlideBackupReply(params *FeedItemParams, chunkSize int) *archive.SlidesReply {
	if len(s.slideBackupAids) <= 0 {
		reply := &archive.SlidesReply{}
		reply.MarkAsBackupReply()
		return reply
	}

	startAt := params.FeedViewReq.DisplayId * int64(chunkSize)
	backupSize := len(s.slideBackupAids)

	startAt = startAt % int64(backupSize)
	backupAids := make([]int64, 0, chunkSize)
	for _, aid := range s.slideBackupAids[startAt:] {
		backupAids = append(backupAids, aid)
		if len(backupAids) >= chunkSize {
			break
		}
	}
	remainSize := chunkSize - len(backupAids)
	backupAids = append(backupAids, s.slideBackupAids[0:remainSize]...)

	reply := &archive.SlidesReply{}
	for _, aid := range backupAids {
		reply.Data = append(reply.Data, &archive.SlidesItem{
			ID:   aid,
			Goto: model.GotoAv,
		})
	}
	reply.MarkAsBackupReply()

	return reply
}

//nolint:unparam
func (s *Service) slidesRecommend(ctx context.Context, params *FeedItemParams) (*archive.SlidesReply, error) {
	const (
		_200ms        = 200
		_request5item = 5
	)
	slidesReply, err := s.arcDao.SlidesRecommend(ctx, &archive.SlidesRequest{
		FromAV:      params.FeedViewReq.Aid,
		SessionID:   params.FeedViewReq.SessionId,
		DisplayID:   params.FeedViewReq.DisplayId,
		FromTrackID: params.FeedViewReq.FromTrackId,
		Mid:         params.AuthN.Mid,
		Buvid:       params.Device.Buvid,
		Timeout:     _200ms,
		Build:       params.Device.Build,
		MobiApp:     params.Device.RawMobiApp,
		Plat:        params.Plat,
		RequestCnt:  _request5item,
		ZoneID:      params.UserZoneID,
		IP:          params.IP,
		Network:     params.Net(),
	})
	if err != nil {
		log.Error("Failed to get slides recommend: %+v: %+v", params, err)
		backupReply := s.makeSlideBackupReply(params, _request5item)
		backupReply.StoreOriginCode(500) // cast as 500
		return backupReply, nil
	}
	if slidesReply.Code != 0 {
		log.Error("Failed to get slides recommend with code: %+v: %d: %+v", params, slidesReply.Code, slidesReply)
		backupReply := s.makeSlideBackupReply(params, _request5item)
		backupReply.StoreOriginCode(slidesReply.Code)
		return backupReply, nil
	}
	return slidesReply, nil
}

//nolint:gomnd
func (s *Service) userZoneID(ctx context.Context) int64 {
	loc, err := s.locDao.Info2(ctx)
	if err != nil {
		log.Error("Failed to get location info: %+v", err)
		return 0
	}
	if len(loc.ZoneId) >= 4 {
		return loc.ZoneId[3]
	}
	return 0
}

func (s *Service) FeedView(ctx context.Context, req *viewApi.FeedViewReq) (*viewApi.FeedViewReply, error) {
	params := &FeedItemParams{
		Now:         time.Now(),
		FeedViewReq: req,
	}
	params.Restriction, _ = restriction.FromContext(ctx)
	params.AuthN, _ = auth.FromContext(ctx)
	params.Device, _ = device.FromContext(ctx)
	params.Plat = model.PlatNew(params.Device.RawMobiApp, params.Device.Device)
	params.IP = metadata.String(ctx, metadata.RemoteIP)
	params.Network, _ = network.FromContext(ctx)
	params.Locale, _ = locale.FromContext(ctx)
	params.IsMelloi = calcIsMelloi(ctx)
	params.UserZoneID = s.userZoneID(ctx)

	reply := &viewApi.FeedViewReply{
		HasNext: true, // 以 true 开始
	}
	slidesMeta, err := s.slidesRecommend(ctx, params)
	if err != nil {
		reply.HasNext = false
		log.Error("Failed to get slides recommend: %+v: %+v", params, err)
		return reply, nil
	}
	defer func() {
		if err := s.infocV2Log.Info(ctx, makeSlideViewEventPayload(params, slidesMeta, reply)); err != nil {
			log.Error("Failed to send info event: %+v", err)
		}
	}()
	if len(slidesMeta.Data) <= 0 {
		reply.HasNext = false
		log.Error("Insuffcient slide reply data: %+v: %+v", params, slidesMeta)
		return reply, nil
	}

	slidesAids := make([]int64, 0, len(slidesMeta.Data))
	for _, item := range slidesMeta.Data {
		slidesAids = append(slidesAids, item.ID)
	}

	fanoutResult, err := s.doSlideViewFanout(ctx, &fanoutParams{
		feedItemParams: params,
		slideAids:      slidesAids,
	})
	if err != nil {
		return nil, err
	}

	cfg := s.defaultViewConfigCreater()()
	cfg.dep = ReplaceBatchDependency(cfg.dep,
		ReplaceArchiveDep(makeArchiveCacheStubImpl(cfg, fanoutResult)),
		ReplaceArchiveHonorDep(makeArchiveHonorCacheStubImpl(cfg, fanoutResult)),
		ReplaceAudioDep(makeAudioCacheStubImpl(cfg, fanoutResult)),
		ReplaceAssistDep(makeAssistCacheStubImpl(cfg, fanoutResult)),
		ReplaceReplyDep(makeReplyCacheStubImpl(cfg, fanoutResult)),
		ReplaceUgcpayRankDep(makeUgcpayRankCacheStubImpl(cfg, fanoutResult)),
		ReplaceThumbupDep(makeArchiveThumbupCacheStubImpl(cfg, fanoutResult)), //点赞
		ReplaceFavorDep(makeArchiveFavorCacheStubImpl(cfg, fanoutResult)),     //收藏
		ReplaceCoinDep(makeArchiveCoinCacheStubImpl(cfg, fanoutResult)),       //投币
		ReplaceDanmuDep(makeArchiveDanmuCacheStubImpl(cfg, fanoutResult)),     //弹幕
		ReplaceHistoryDep(makeArchiveHistoryCacheStubImpl(cfg, fanoutResult)), //历史记录
	)
	opts := []ViewOption{
		SkipRelate(true),
		WithPopupExp(s.buvidABTest(ctx, params.Device.Buvid, popupFlag)),
		WithAutoSwindowExp(s.buvidABTest(ctx, params.Device.Buvid, pipVal)),
		WithSmallWindowExp(s.SmallWindowConfig(ctx, params.Device.Buvid, smallWindowABtest)),
		SkipSpecialCell(true),
		WithAdTab(false), // 新详情页不出商业Tab
	}
	cfg.Apply(opts...)
	ctx = WithContext(ctx, cfg)

	lock := sync.Mutex{}
	viewItems := map[int64]*viewApi.FeedViewItem{}
	eg := errgroup.WithContext(ctx)
	for _, item := range slidesMeta.Data {
		dupParams := params.DuplicateWith(item)
		eg.Go(func(ctx context.Context) error {
			viewItem, err := s.requestFeedViewItem(ctx, dupParams)
			if err != nil {
				log.Error("Failed to request feed view: %+v: %+v", dupParams, err)
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			viewItems[dupParams.CurrentAid] = viewItem
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	for _, item := range slidesMeta.Data {
		vi, ok := viewItems[item.ID]
		if !ok {
			log.Error("Failed to get archive view: %+v", item)
			continue
		}
		ap, ok := fanoutResult.SlideArchivePlayer[item.ID]
		if ok {
			uriParam := strconv.FormatInt(item.ID, 10)
			vi.Uri = model.FillURI(item.Goto, uriParam, cdm.ArcPlayHandler(ap.Arc, ap.PlayerInfo[ap.DefaultPlayerCid], item.TrackID, nil, int(params.Device.Build), params.Device.RawMobiApp, true))
		}
		reply.List = append(reply.List, vi)
	}
	return reply, nil
}

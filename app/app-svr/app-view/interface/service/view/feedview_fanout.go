package view

import (
	"context"

	archivedao "go-gateway/app/app-svr/app-view/interface/dao/archive"
	"go-gateway/app/app-svr/app-view/interface/dao/dm"
	"go-gateway/app/app-svr/app-view/interface/dao/favorite"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
	archivecachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/archive"
	archivehonorcachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/archivehonor"
	assistcachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/assist"
	audiocachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/audio"
	archivecoincachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/coin"
	archivedanmucachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/danmu"
	archivefavorcachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/favor"
	archivehistorycachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/history"
	replycachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/reply"
	archivethumbupcachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/thumbup"
	rankcachestub "go-gateway/app/app-svr/app-view/interface/service/view/dependency/cachestub/ugcpayrank"
	archviehonorapi "go-gateway/app/app-svr/archive-honor/service/api"
	archiveapi "go-gateway/app/app-svr/archive/service/api"

	ugcpayrankapi "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	replyapi "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
)

type fanoutParams struct {
	feedItemParams *FeedItemParams

	slideAids []int64
}

type fanoutResult struct {
	SlideArchivePlayer         map[int64]*archiveapi.ArcPlayer
	SlideArchiveViews          map[int64]*archiveapi.ViewReply
	SlideArchiveDescriptionsV2 map[int64]*archivedao.DescriptionV2Reply
	SlideArchiveArguments      map[int64]string
	SlideArchiveHonors         map[int64]*archviehonorapi.HonorReply
	SlideAudios                map[int64]*view.Audio
	SlideAssists               map[int64][]int64
	SlideReplyPreface          map[string]*replyapi.ReplyListPrefaceReply
	SlideUgcpayRank            *ugcpayrankapi.RankElecUPListResp
	SlideArchiveHasLike        map[int64]thumbup.State
	SlideArchiveIsFavor        map[int32]map[int64]bool
	SlideArchiveUserCoins      map[int64]int64
	SlideArchiveSubjectInfos   map[int64]*dmApi.SubjectInfo
	SlideArchiveProgress       map[int64]*hisgrpc.ModelHistory
	SlideRedirectPolicy        map[int64]*archiveapi.RedirectPolicy
}

func makeDescrpitionV2Request(in map[int64]*archiveapi.ViewReply) []*archivedao.DescriptionV2Request {
	out := make([]*archivedao.DescriptionV2Request, 0, len(in))
	for _, view := range in {
		out = append(out, &archivedao.DescriptionV2Request{
			Aid: view.Aid,
		})
	}
	return out
}

func makeAudiosRequest(in map[int64]*archiveapi.ViewReply) []int64 {
	out := []int64{}
	for _, v := range in {
		pLen := len(v.Pages)
		if pLen == 0 || pLen > 100 {
			continue
		}
		maxPLen := 50
		if pLen > maxPLen {
			pLen = maxPLen
		}
		for _, p := range v.Pages[:pLen] {
			out = append(out, p.Cid)
		}
	}
	return out
}

func makeIsFavorRequest(in map[int64]*archiveapi.ViewReply) *favorite.BatchIsFavoredsResourcesReq {
	out := &favorite.BatchIsFavoredsResourcesReq{}
	for _, view := range in {
		out.Aids = append(out.Aids, view.Aid)
		if view.SeasonID > 0 {
			out.Sids = append(out.Sids, view.SeasonID)
		}
	}
	return out
}

func makeSubjectInfosRequest(in map[int64]*archiveapi.ViewReply) *dm.SubjectInfosReq {
	out := &dm.SubjectInfosReq{
		Typ:  1,
		Plat: 1,
	}
	limit := 50
	for _, view := range in {
		pLen := len(view.Pages)
		if pLen == 0 || pLen > 100 {
			continue
		}
		if pLen > limit {
			pLen = limit
		}
		for _, p := range view.Pages[:pLen] {
			out.Cids = append(out.Cids, p.Cid)
		}
	}
	return out
}

// nolint:gocognit
func (s *Service) doSlideViewFanout(ctx context.Context, params *fanoutParams) (*fanoutResult, error) {
	result := &fanoutResult{}

	eground := [2]*errgroup.Group{}
	eground[0] = errgroup.WithContext(ctx)
	eground[1] = errgroup.WithContext(ctx)

	eground[0].Go(func(ctx context.Context) error {
		playAV := []*archiveapi.PlayAv{}
		for _, aid := range params.slideAids {
			playAV = append(playAV, &archiveapi.PlayAv{
				Aid: aid,
			})
		}
		reply, err := s.arcDao.ArcsPlayer(ctx, playAV)
		if err != nil {
			log.Error("Failed to get archive player: %+v: %+v", playAV, err)
			return nil
		}
		result.SlideArchivePlayer = reply
		return nil
	})
	eground[0].Go(func(ctx context.Context) error {
		reply, err := s.arcDao.Views(ctx, params.slideAids)
		if err != nil {
			log.Error("Failed to get archive views: %+v: %+v", params.slideAids, err)
			return nil
		}
		result.SlideArchiveViews = reply
		return nil
	})
	eground[0].Go(func(ctx context.Context) error {
		req := &vuapi.MultiArchiveArgumentReq{
			Aids: params.slideAids,
		}
		reply, err := s.vuDao.MultiArchiveArgument(ctx, req)
		if err != nil {
			log.Error("Failed to get archive argument: %+v: %+v", req, err)
			return nil
		}
		arguments := map[int64]string{}
		for aid, a := range reply.Arguments {
			arguments[aid] = a.ArgueMsg
		}
		result.SlideArchiveArguments = arguments
		return nil
	})
	eground[0].Go(func(ctx context.Context) error {
		req := &archviehonorapi.HonorsRequest{
			Aids:    params.slideAids,
			Build:   params.feedItemParams.Device.Build,
			MobiApp: params.feedItemParams.Device.RawMobiApp,
			Device:  params.feedItemParams.Device.Device,
		}
		reply, err := s.ahDao.BatchHonors(ctx, req)
		if err != nil {
			log.Error("Failed to get archive honours: %+v: %+v", req, err)
			return nil
		}
		result.SlideArchiveHonors = reply.Honors
		return nil
	})
	eground[0].Go(func(ctx context.Context) error {
		var req []*replyapi.ReplyListPrefaceReq
		for _, aid := range params.slideAids {
			req = append(req, &replyapi.ReplyListPrefaceReq{
				Oid:   aid,
				Type:  model.ReplyTypeAv,
				Buvid: params.feedItemParams.Device.Buvid,
				Mid:   params.feedItemParams.AuthN.Mid,
			})
		}
		reply, err := s.replyDao.GetReplyListsPreface(ctx, &replyapi.ReplyListsPrefaceReq{Subjects: req})
		if err != nil {
			log.Error("Failed to get reply preface: %+v: %+v", req, err)
			return nil
		}
		result.SlideReplyPreface = reply.Prefaces
		return nil
	})
	eground[0].Go(func(ctx context.Context) error {
		req := &archiveapi.ArcsRedirectPolicyRequest{
			Aids: params.slideAids,
		}
		reply, err := s.arcDao.BatchArcRedirectUrls(ctx, req)
		if err != nil {
			log.Error("Failed to get ArcRedirectUrls is error: %+v: %+v", req, err)
			return nil
		}
		result.SlideRedirectPolicy = reply
		return nil
	})
	if params.feedItemParams.AuthN.Mid > 0 || params.feedItemParams.Device.Buvid != "" {
		//历史记录
		eground[0].Go(func(ctx context.Context) error {
			aids := params.slideAids
			mid := params.feedItemParams.AuthN.Mid
			buvid := params.feedItemParams.Device.Buvid
			reply, err := s.arcDao.BatchProgress(ctx, aids, mid, buvid)
			if err != nil {
				log.Error("Failed to get batchProgress is error: %+v: %+v", aids, err)
				return nil
			}
			result.SlideArchiveProgress = reply
			return nil
		})
		//是否点赞
		eground[0].Go(func(ctx context.Context) error {
			mid := params.feedItemParams.AuthN.Mid
			business := "archive"
			buvid := params.feedItemParams.Device.Buvid
			aids := params.slideAids
			reply, err := s.thumbupDao.BatchHasLike(ctx, mid, business, buvid, aids)
			if err != nil {
				log.Error("Failed to get hasLike is error: %+v: %+v", aids, err)
				return nil
			}
			result.SlideArchiveHasLike = reply
			return nil
		})
	}
	if params.feedItemParams.AuthN.Mid > 0 {
		//投币
		eground[0].Go(func(ctx context.Context) error {
			aids := params.slideAids
			mid := params.feedItemParams.AuthN.Mid
			reply, err := s.coinDao.BatchArchiveUserCoins(ctx, aids, mid, 1)
			if err != nil {
				log.Error("Failed to get userCoins is error: %+v: %+v", aids, err)
				return nil
			}
			result.SlideArchiveUserCoins = reply
			return nil
		})
	}
	if err := eground[0].Wait(); err != nil {
		return nil, err
	}
	var upIDs []int64
	for _, v := range result.SlideArchiveViews {
		upIDs = append(upIDs, v.Author.Mid)
	}
	// desc
	eground[1].Go(func(ctx context.Context) error {
		reply, err := s.arcDao.DescriptionsV2(ctx, makeDescrpitionV2Request(result.SlideArchiveViews))
		if err != nil {
			log.Error("Failed to get archive desc: %+v: %+v", params.slideAids, err)
			return nil
		}
		result.SlideArchiveDescriptionsV2 = reply
		return nil
	})
	// audio
	eground[1].Go(func(ctx context.Context) error {
		reply, err := s.audioDao.AudioByCids(ctx, makeAudiosRequest(result.SlideArchiveViews))
		if err != nil {
			log.Error("Failed to get archive audio: %+v: %+v", params.slideAids, err)
			return nil
		}
		result.SlideAudios = reply
		return nil
	})
	// assist
	eground[1].Go(func(ctx context.Context) error {
		reply, err := s.assDao.MultiAssist(ctx, upIDs)
		if err != nil {
			log.Error("Failed to get archive assist: %+v: %+v", params.slideAids, err)
			return nil
		}
		assists := make(map[int64][]int64)
		for upID, ass := range reply.AssistMids {
			if ass == nil {
				continue
			}
			assists[upID] = ass.AssistMids
		}
		result.SlideAssists = assists
		return nil
	})
	// elec rank
	eground[1].Go(func(ctx context.Context) error {
		reply, err := s.elcDao.RankElecMonthUPList(ctx, upIDs, params.feedItemParams.Device.Build, params.feedItemParams.Device.RawMobiApp, params.feedItemParams.Device.RawPlatform, params.feedItemParams.Device.Device)
		if err != nil {
			log.Error("Failed to get archive audio: %+v: %+v", params.slideAids, err)
			return nil
		}
		result.SlideUgcpayRank = reply
		return nil
	})
	//是否收藏
	if params.feedItemParams.AuthN.Mid > 0 {
		eground[1].Go(func(ctx context.Context) error {
			req := makeIsFavorRequest(result.SlideArchiveViews)
			req.Mid = params.feedItemParams.AuthN.Mid
			reply, err := s.favDao.BatchIsFavoredsResources(ctx, req)
			if err != nil {
				log.Error("Failed to get isFavor: %+v: %+v", req, err)
				return nil
			}
			result.SlideArchiveIsFavor = reply
			return nil
		})
	}
	//弹幕
	eground[1].Go(func(ctx context.Context) error {
		req := makeSubjectInfosRequest(result.SlideArchiveViews)
		reply, err := s.dmDao.SubjectInfos(ctx, req.Typ, req.Plat, req.Cids...)
		if err != nil {
			log.Error("Failed to get subjectInfos is error: %+v: %+v", req, err)
			return nil
		}
		result.SlideArchiveSubjectInfos = reply
		return nil
	})
	if err := eground[1].Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func makeArchiveCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivecachestub.Impl {
	out := &archivecachestub.Impl{
		Origin: cfg.dep.Archive,
	}
	out.Reply.Views = result.SlideArchiveViews
	out.Reply.DescriptionsV2 = result.SlideArchiveDescriptionsV2
	out.Reply.Arguments = result.SlideArchiveArguments
	out.Reply.RedirectPolicy = result.SlideRedirectPolicy
	return out
}

func makeArchiveHonorCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivehonorcachestub.Impl {
	out := &archivehonorcachestub.Impl{
		Origin: cfg.dep.ArchiveHornor,
	}
	out.Reply.ArchiveHonors = result.SlideArchiveHonors
	return out
}

func makeAudioCacheStubImpl(cfg viewConfig, result *fanoutResult) *audiocachestub.Impl {
	out := &audiocachestub.Impl{
		Origin: cfg.dep.Audio,
	}
	out.Reply.Audio = result.SlideAudios
	return out
}

func makeAssistCacheStubImpl(cfg viewConfig, result *fanoutResult) *assistcachestub.Impl {
	out := &assistcachestub.Impl{
		Origin: cfg.dep.Assist,
	}
	out.Reply.Assist = result.SlideAssists
	return out
}

func makeReplyCacheStubImpl(cfg viewConfig, result *fanoutResult) *replycachestub.Impl {
	out := &replycachestub.Impl{
		Origin: cfg.dep.Reply,
	}
	out.Reply.ReplyPreface = result.SlideReplyPreface
	return out
}

func makeUgcpayRankCacheStubImpl(cfg viewConfig, result *fanoutResult) *rankcachestub.Impl {
	out := &rankcachestub.Impl{
		Origin: cfg.dep.UGCPayRank,
	}
	out.Reply.RankElec = result.SlideUgcpayRank
	return out
}

func makeArchiveThumbupCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivethumbupcachestub.Impl {
	out := &archivethumbupcachestub.Impl{
		Origin: cfg.dep.ThumbUP,
	}
	out.Reply.ArchiveHasLike = result.SlideArchiveHasLike
	return out
}

func makeArchiveFavorCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivefavorcachestub.Impl {
	out := &archivefavorcachestub.Impl{
		Origin: cfg.dep.Fav,
	}
	out.Reply.ArchiveIsFavor = result.SlideArchiveIsFavor
	return out
}

func makeArchiveCoinCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivecoincachestub.Impl {
	out := &archivecoincachestub.Impl{
		Origin: cfg.dep.Coin,
	}
	out.Reply.ArchiveUserCoins = result.SlideArchiveUserCoins
	return out
}

func makeArchiveDanmuCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivedanmucachestub.Impl {
	out := &archivedanmucachestub.Impl{
		Origin: cfg.dep.Danmu,
	}
	out.Reply.ArchiveSubjectInfos = result.SlideArchiveSubjectInfos
	return out
}

func makeArchiveHistoryCacheStubImpl(cfg viewConfig, result *fanoutResult) *archivehistorycachestub.Impl {
	out := &archivehistorycachestub.Impl{
		Origin: cfg.dep.History,
	}
	out.Reply.ArchiveProgress = result.SlideArchiveProgress
	return out
}

func ReplaceBatchDependency(dep dependency.ViewDependency, opts ...ReplaceOption) dependency.ViewDependency {
	for _, opt := range opts {
		opt(&dep)
	}
	return dep
}

type ReplaceOption func(*dependency.ViewDependency)

func ReplaceArchiveDep(in dependency.ArchiveDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Archive = in
	}
}

func ReplaceArchiveHonorDep(in dependency.ArchiveHonorDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.ArchiveHornor = in
	}
}

func ReplaceAudioDep(in dependency.AudioDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Audio = in
	}
}

func ReplaceAssistDep(in dependency.AssistDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Assist = in
	}
}

func ReplaceReplyDep(in dependency.ReplyDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Reply = in
	}
}

func ReplaceUgcpayRankDep(in dependency.UgcpayRankDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.UGCPayRank = in
	}
}

func ReplaceThumbupDep(in dependency.ThumbupDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.ThumbUP = in
	}
}

func ReplaceFavorDep(in dependency.FavDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Fav = in
	}
}

func ReplaceCoinDep(in dependency.CoinDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Coin = in
	}
}

func ReplaceDanmuDep(in dependency.DanmuDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.Danmu = in
	}
}

func ReplaceHistoryDep(in dependency.HistoryDependency) ReplaceOption {
	return func(dep *dependency.ViewDependency) {
		dep.History = in
	}
}

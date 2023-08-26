package v1

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-search/internal/model"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	"go-gateway/app/app-svr/archive/service/api"

	livecommon "git.bilibili.co/bapis/bapis-go/live/xroom-gate/common"

	"github.com/pkg/errors"
)

// 热搜榜单
func (s *Service) Ranking(ctx context.Context, mid int64, req *search.TrendingRankingReq) (*search.TrendingRankingRsp, error) {
	zoneId := _defaultZoneID
	if zone, err := s.dao.LocationInfo(ctx, metadata.String(ctx, metadata.RemoteIP)); err == nil && zone != nil {
		zoneId = int(zone.ZoneId)
	}
	hot, err := s.dao.Trending(ctx, req.Buvid, mid, int(req.Build), int(req.Limit), zoneId, req.MobiApp, req.Device, req.Platform, time.Now(), true)
	if err != nil {
		log.Error("s.srchDao.Trending %+v", err)
		return nil, err
	}
	var fanout *FanoutResult
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		// 获取词条直播中状态
		liveTrending := liveTrendingRoomID(hot)
		if len(liveTrending) > 0 {
			inStreamingRoom := s.inStreamingRoom(ctx, liveTrending)
			setShowLiveIcon(hot, inStreamingRoom)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		fanout, err = s.makeRankingConfigCardFanout(ctx, hot)
		if err != nil {
			log.Error("makeRankingConfigCardFanout hot=%+v, error=%+v", hot, err)
			return nil
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &search.TrendingRankingRsp{
		Code:    hot.Code,
		TrackID: hot.SeID,
		List:    makeTrendingList(hot, fanout),
		ExpStr:  hot.ExpStr,
	}, nil
}

func (s *Service) makeRankingConfigCardFanout(ctx context.Context, hot *search.Hot) (*FanoutResult, error) {
	loader := NewInlineCardFanoutLoader{General: constructGeneralParamFromCtx(ctx), Service: s}
	for _, v := range hot.List {
		if len(v.Res) > 0 {
			for _, item := range v.Res {
				switch item.CardType {
				case _rankingConfigResTypeVideo:
					loader.Archive.Aids = append(loader.Archive.Aids, item.Id)
				case _rankingConfigResTypeLive:
					loader.Live.LiveEntryFrom = []string{model.HotSearchLiveCard}
					loader.Live.InlineRoomIDs = append(loader.Live.InlineRoomIDs, item.Id)
				default:
					log.Warn("Unexpected CardType in res(%+v)", item)
				}
			}
		}
	}
	fanout, err := loader.doSearchRankingCardFanoutLoad(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "doSearchRankingCardFanoutLoad loader=%+v", loader)
	}
	return fanout, nil
}

func makeTrendingList(hot *search.Hot, fanout *FanoutResult) []*search.TrendingList {
	var res []*search.TrendingList
	for _, v := range hot.List {
		meta := &search.TrendingList{
			Position:        v.Pos,
			Keyword:         v.Keyword,
			ShowName:        v.ShowName,
			WordType:        v.WordType,
			Icon:            v.Icon,
			HotId:           v.HotId,
			ResourceID:      v.ResourceID,
			ShowLiveIcon:    v.ShowLiveIcon,
			HeatValue:       v.HeatValue,
			ConfigCardItems: makeRankingConfigCardItems(v.Res, fanout),
		}
		switch v.GotoType {
		case search.HotTypeArchive:
			meta.Goto = model.GotoAv
			meta.URI = model.FillURI(v.Goto, v.GotoValue, nil)
		case search.HotTypeArticle:
			meta.Goto = model.GotoArticle
			meta.URI = model.FillURI(v.Goto, v.GotoValue, nil)
		case search.HotTypePGC:
			meta.Goto = model.GotoEP
			meta.URI = model.FillURI(v.Goto, v.GotoValue, nil)
		case search.HotTypeURL:
			meta.Goto = model.GotoWeb
			meta.URI = model.FillURI(v.Goto, v.GotoValue, nil)
		default:
		}
		res = append(res, meta)
	}
	return res
}

func makeRankingConfigCardItems(hotList []*search.HotListRes, fanout *FanoutResult) []*search.RankingConfigCardItem {
	if len(hotList) == 0 || fanout == nil {
		return nil
	}
	var res []*search.RankingConfigCardItem
	for _, v := range hotList {
		switch v.CardType {
		case _rankingConfigResTypeVideo:
			item, ok := fanout.Archive.Archive[v.Id]
			if !ok || item == nil || item.Arc == nil {
				continue
			}
			res = append(res, &search.RankingConfigCardItem{
				CardType:          _rankingConfigResTypeVideo,
				Cover:             item.Arc.Pic,
				CoverLeftShowDesc: cardmdl.StatString(item.Arc.Stat.View, ""),
				CoverLeftShowImg:  "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/iOKzjZdPCU.png",
				Title:             item.Arc.Title,
				JumpUrl:           resolveAvPlayerInfoJumpUrl(item),
				Param:             strconv.FormatInt(item.Arc.Aid, 10),
			})
		case _rankingConfigResTypeLive:
			item, ok := fanout.Live.InlineRoom[v.Id]
			if !ok {
				continue
			}
			liveCard := &search.RankingConfigCardItem{
				CardType:          _rankingConfigResTypeLive,
				Cover:             item.Cover,
				CoverLeftShowDesc: cardmdl.StatString(int32(item.PopularityCount), ""),
				CoverLeftShowImg:  "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/hIClEGTvtZ.png",
				Title:             item.Title,
				JumpUrl:           item.JumpUrl[model.HotSearchLiveCard],
				Param:             strconv.FormatInt(item.RoomId, 10),
				LiveConfigs: &search.LiveConfigs{
					ShowLiveIcon: true,
					LiveStatus:   item.LiveStatus,
				},
			}
			if desc, ok := updateLiveCoverLeftShow(item.WatchedShow); ok {
				liveCard.CoverLeftShowDesc = desc
			}
			res = append(res, liveCard)
		default:
			log.Warn("Unexpected makeRankingConfigCardItems CardType v=%+v", v)
		}
	}
	return res
}

func updateLiveCoverLeftShow(watchedShow *livecommon.WatchedShow) (string, bool) {
	if watchedShow == nil {
		return "", false
	}
	return watchedShow.TextSmall, true
}

func resolveAvPlayerInfoJumpUrl(ap *api.ArcPlayer) string {
	if ap.Arc.AttrVal(api.AttrBitIsPGC) == api.AttrYes && ap.Arc.RedirectURL != "" {
		// pgc视频
		return ap.Arc.RedirectURL
	}
	// ugc视频
	playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
	return model.FillURI(model.GotoAv, strconv.FormatInt(ap.Arc.Aid, 10), model.AvPlayHandlerGRPC(ap.Arc, playInfo))
}

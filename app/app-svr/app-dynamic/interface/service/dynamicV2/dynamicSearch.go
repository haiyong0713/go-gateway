package dynamicV2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	esmdl "go-gateway/app/app-svr/app-dynamic/interface/model/es"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (s *Service) DynSearch(c context.Context, general *mdlv2.GeneralParam, req *api.DynSearchReq) (*api.DynSearchReply, error) {
	// 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, err
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	var (
		_max      = 20 // 最大个数
		channlIDs []int64
		dynList   *mdlv2.DynListRes
		search    *dyngrpc.SearchRsp
		topics    mdlv2.DynSearchTopicResult
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if search, dynList, err = s.dynDao.Search(ctx, general, req, attentions, _max); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	if req.Page == 1 {
		eg.Go(func(ctx context.Context) (err error) {
			if channlIDs, err = s.searchChannel(ctx, general, req); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		// 粉板6.45以上使用新话题搜索
		if s.isDynNewTopicVerticalSearch(c, general) {
			eg.Go(func(ctx context.Context) (err error) {
				if topics, err = s.topDao.TopicSearchV2(ctx, req.Keyword, general); err != nil {
					log.Errorc(ctx, "TopicSearchV2 dao error: %v", err)
				}
				return nil
			})
		} else {
			eg.Go(func(ctx context.Context) (err error) {
				if topics, err = s.topDao.OldTopicSearch(ctx, req.Keyword, general); err != nil {
					log.Errorc(ctx, "OldTopicSearch dao error: %v", err)
				}
				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: dynList.Dynamics, rcmdUps: dynList.RcmdUps,
		upRegionRcmds: dynList.RegionUps, fold: dynList.FoldInfo, channelIDs: channlIDs,
	})
	if err != nil {
		xmetric.DynamicCoreAPI.Inc("垂搜页", "request_error")
		log.Error("DynSearch mid(%v) Search(), error %v", general.Mid, err)
		return nil, err
	}
	// 动态
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeSearch)
	s.procBackfill(c, dynCtx, general, foldList)
	// 频道
	channelInfo := s.searchChannelCard(dynCtx, req, channlIDs, topics)
	// 话题
	topicInfo := s.searchTopic(req, topics)
	res := &api.DynSearchReply{
		ChannelInfo: channelInfo,
		SearchTopic: topicInfo,
		SearchInfo: &api.SearchInfo{
			Title:   fmt.Sprintf("关于“%s”的动态", req.Keyword),
			List:    s.procFold(foldList, dynCtx, general),
			TrackId: search.TrackId,
			Total:   search.Total,
			Version: search.Version,
			HasMore: dynList.HasMore,
		},
	}
	return res, nil
}

func (s *Service) searchChannel(c context.Context, general *mdlv2.GeneralParam, req *api.DynSearchReq) ([]int64, error) {
	const (
		pn = 1
		ps = 3
	)
	var (
		channlids, hideTids []int64
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if channlids, err = s.esDao.SearchChannel(ctx, general.Mid, req.Keyword, pn, ps, esmdl.ChannelOK); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if hideTids, err = s.esDao.SearchChannel(ctx, general.Mid, req.Keyword, pn, ps, esmdl.ChannelHide); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	hideTids = append(hideTids, channlids...)
	return hideTids, nil
}

func (s *Service) searchChannelCard(dynCtx *mdlv2.DynamicContext, req *api.DynSearchReq, channelIDs []int64, topic mdlv2.DynSearchTopicResult) *api.SearchChannel {
	const (
		_itemlen = 3
		_bigCard = 1
	)
	showSubCard := true
	if topic != nil {
		showSubCard = topic.ShowChannelSubCards()
	}

	res := &api.SearchChannel{
		Title: "频道",
		MoreButton: &api.SearchTopicButton{
			Title:   "更多",
			JumpUri: model.FillURI(model.GotoChannelSearch, req.Keyword, nil),
		},
	}
	for _, channelID := range channelIDs {
		info, ok := dynCtx.ResSearchChannels[channelID]
		if !ok {
			continue
		}
		cInfo := &api.ChannelInfo{
			ChannelId:   info.Cid,
			ChannelName: info.Cname,
			Icon:        info.Icon,
			JumpUri:     model.FillURI(model.GotoChannel, strconv.FormatInt(info.Cid, 10), nil),
		}
		var labels []string
		if info.ResourceCnt > 0 {
			labels = append(labels, model.StatString(info.ResourceCnt, "视频"))
		}
		if info.FeaturedCnt > 0 {
			labels = append(labels, model.StatString(info.FeaturedCnt, "精选视频"))
		}
		if len(labels) > 0 {
			cInfo.Desc = strings.Join(labels, "  ")
		}
		cInfo.IsAtten = info.Subscribed
		cInfo.TypeIcon = "https://i0.hdslb.com/bfs/tag/3e82aab221dfccab444dafa9e3e95d2953cd4220.png"
		var items []*api.RcmdItem
		if showSubCard {
			for _, video := range info.VideoCards {
				ap, ok := dynCtx.GetArchive(video.Rid)
				if !ok {
					continue
				}
				var archive = ap.Arc
				cardArc := &api.RcmdArchive{
					Cover:           archive.Pic,
					CoverLeftIcon_1: api.CoverIcon_cover_icon_play,
					CoverLeftText_1: s.numTransfer(int(archive.Stat.View)),
					Uri:             model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, true)),
					Aid:             archive.Aid,
					Title:           archive.Title,
				}
				if video.BadgeTitle != "" && video.BadgeBackground != "" {
					cardArc.Badge = &api.IconBadge{
						IconBgUrl: video.BadgeBackground,
						Text:      video.BadgeTitle,
					}
				}
				// PGC特殊逻辑
				if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
					cardArc.Uri = archive.RedirectURL
					cardArc.IsPgc = true
				}
				item := &api.RcmdItem{
					Type: api.RcmdType_rcmd_archive,
					RcmdItem: &api.RcmdItem_RcmdArchive{
						RcmdArchive: cardArc,
					},
				}
				items = append(items, item)
			}
			if len(items) < _itemlen {
				continue
			}
		}
		cInfo.Items = items
		res.Channels = append(res.Channels, cInfo)
	}
	if len(res.Channels) == 0 {
		return nil
	}
	if len(res.Channels) > _itemlen {
		res.Channels = res.Channels[:_itemlen]
	}
	if len(res.Channels) > _bigCard {
		for _, v := range res.Channels {
			v.Items = nil
		}
	}
	return res
}

func (s *Service) searchTopic(req *api.DynSearchReq, topic mdlv2.DynSearchTopicResult) *api.SearchTopic {
	if topic == nil {
		return nil
	}
	return topic.ToDynV2SearchTopic(req)
}

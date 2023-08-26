package kernel

import (
	"context"
	"sync"
	"time"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	appshowgrpc "git.bilibili.co/bapis/bapis-go/app/show/v1"
	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	commscoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	xroomfeedgrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcfollowgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	chargrpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/sync/errgroup.v2"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcmidv1 "go-gateway/app/app-svr/archive/middleware/v1"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/util"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

// model.MaterialWeeks
func (ml *MaterialLoader) doWeekCards(eg *errgroup.Group, material *Material) {
	if len(ml.weekIDs) == 0 {
		return
	}
	if material.WeekCard == nil {
		material.WeekCard = make(map[int64]*appshowgrpc.SerieConfig)
	}
	mu := sync.Mutex{}
	groupRequest(ml.weekIDs, _weekReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Appshow().BatchSerie(ctx, &appshowgrpc.BatchSerieReq{Type: model.SelTypeWeek, Number: groupIDs})
			if err != nil {
				return err
			}
			mu.Lock()
			for k, val := range rly {
				if val == nil {
					continue
				}
				material.WeekCard[k] = val
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialGame
func (ml *MaterialLoader) doGameCard(eg *errgroup.Group, material *Material) {
	if len(ml.gameIDs) == 0 {
		return
	}
	if material.GameCard == nil {
		material.GameCard = make(map[int64]*model.GaItem, len(ml.gameIDs))
	}
	var (
		platformType string
	)
	//平台类型：0=PC，1=安卓，2=IOS
	switch {
	case ml.ss.IsIOS():
		platformType = "2"
	case ml.ss.IsAndroid():
		platformType = "1"
	case ml.ss.IsH5():
		if ml.ss.IsH5Android() {
			platformType = "1"
		} else {
			platformType = "2"
		}
	default:
		platformType = "0"
	}
	mu := sync.Mutex{}
	groupRequest(ml.gameIDs, _gameReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Game().MultiGameInfo(ctx, groupIDs, ml.ss.Mid(), platformType)
			if err != nil {
				return err
			}
			mu.Lock()
			for _, val := range rly {
				if val == nil {
					continue
				}
				material.GameCard[val.GameBaseId] = val
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialLive
func (ml *MaterialLoader) doLiveCard(eg *errgroup.Group, material *Material) {
	if len(ml.liveIDs) == 0 {
		return
	}
	if material.LiveCard == nil {
		material.LiveCard = make(map[uint64]*xroomfeedgrpc.LiveCardInfo, len(ml.liveIDs))
	}
	isHttps := ml.ss.HttpsUrlReq == 1
	var (
		build            int64
		platform, device string
	)
	if ml.ss.IsIOS() || ml.ss.IsAndroid() || ml.ss.IsIPad() {
		build = ml.ss.RawDevice().Build
		platform = ml.ss.RawDevice().RawPlatform
		device = ml.ss.RawDevice().Device
	}
	mu := sync.Mutex{}
	groupRequest(ml.liveIDs, _liveReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.LiveXRoomFeed().GetCardInfo(ctx, groupIDs, ml.ss.Mid(), build, platform, device, isHttps)
			if err != nil {
				return err
			}
			mu.Lock()
			for id, val := range rly {
				if val == nil {
					continue
				}
				material.LiveCard[id] = val
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialArchive
func (ml *MaterialLoader) doArcs(eg *errgroup.Group, material *Material) {
	if len(ml.aids) == 0 {
		return
	}
	if material.Arcs == nil {
		material.Arcs = make(map[int64]*arcgrpc.Arc, len(ml.aids))
	}
	mu := sync.Mutex{}
	groupRequest(ml.aids, _arcReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &arcgrpc.ArcsRequest{
				Aids:    groupIDs,
				Mid:     ml.ss.Mid(),
				MobiApp: ml.ss.RawDevice().RawMobiApp,
				Device:  ml.ss.RawDevice().Device,
			}
			rly, err := ml.dep.Archive().Arcs(ctx, req)
			if err != nil {
				return err
			}
			mu.Lock()
			for aid, arc := range rly {
				if arc == nil {
					continue
				}
				material.Arcs[aid] = arc
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialLiveRoom
func (ml *MaterialLoader) doLiveRooms(eg *errgroup.Group, material *Material) {
	if ml.roomIDs == nil {
		return
	}
	if material.LiveRooms == nil {
		material.LiveRooms = make(map[int64]map[int64]*liveplaygrpc.RoomList)
	}
	for k, roomIDs := range ml.roomIDs {
		isLive := k
		mu := sync.Mutex{}
		groupRequest(roomIDs, _roomReqMax, func(groupIDs []int64) {
			eg.Go(func(ctx context.Context) error {
				rly, err := ml.dep.LivePlay().GetListByRoomId(ctx, groupIDs, isLive)
				if err != nil {
					return err
				}
				mu.Lock()
				if _, ok := material.LiveRooms[isLive]; !ok {
					material.LiveRooms[isLive] = make(map[int64]*liveplaygrpc.RoomList, len(rly))
				}
				for rid, room := range rly {
					if room == nil {
						continue
					}
					material.LiveRooms[isLive][rid] = room
				}
				mu.Unlock()
				return nil
			})
		})
	}
}

// model.MaterialArticle
func (ml *MaterialLoader) doArticles(eg *errgroup.Group, material *Material) {
	if len(ml.cvids) == 0 {
		return
	}
	if material.Articles == nil {
		material.Articles = make(map[int64]*articlemdl.Meta, len(ml.cvids))
	}
	mu := sync.Mutex{}
	groupRequest(ml.cvids, _articleReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Article().ArticleMetas(ctx, groupIDs, 2)
			if err != nil {
				return err
			}
			mu.Lock()
			for cvid, article := range rly {
				if article == nil {
					continue
				}
				material.Articles[cvid] = article
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialEpisode
func (ml *MaterialLoader) doEpisodes(eg *errgroup.Group, material *Material) {
	if len(ml.epids) == 0 {
		return
	}
	if material.Episodes == nil {
		material.Episodes = make(map[int64]*model.EpPlayer, len(ml.epids))
	}
	mu := sync.Mutex{}
	groupRequest(ml.epids, _epReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			batchArgs, _ := arcmid.FromContext(ctx)
			rly, err := ml.dep.Bangumi().EpPlayer(ctx, groupIDs, ml.ss.RawDevice(), batchArgs)
			if err != nil {
				return err
			}
			mu.Lock()
			for epid, episode := range rly {
				if episode == nil {
					continue
				}
				material.Episodes[epid] = episode
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialFolder
func (ml *MaterialLoader) doFolders(eg *errgroup.Group, material *Material) {
	if ml.folderIDs == nil {
		return
	}
	if material.Folders == nil {
		material.Folders = make(map[int32]map[int64]*favmdl.Folder)
	}
	for k, fids := range ml.folderIDs {
		typ := k
		mu := sync.Mutex{}
		groupRequest(fids, _folderReqMax, func(groupIDs []int64) {
			eg.Go(func(ctx context.Context) error {
				rly, err := ml.dep.Favorite().Folders(ctx, groupIDs, typ)
				if err != nil {
					return err
				}
				mu.Lock()
				if material.Folders[typ] == nil {
					material.Folders[typ] = make(map[int64]*favmdl.Folder)
				}
				for fid, f := range rly {
					if f == nil {
						continue
					}
					material.Folders[typ][fid] = f
				}
				mu.Unlock()
				return nil
			})
		})
	}
}

// model.MaterialActSubProto
func (ml *MaterialLoader) doActSubProtos(eg *errgroup.Group, material *Material) {
	if len(ml.actSubProtoIDs) == 0 {
		return
	}
	if material.ActSubProtos == nil {
		material.ActSubProtos = make(map[int64]*activitygrpc.ActSubProtocolReply, len(ml.actSubProtoIDs))
	}
	mu := sync.Mutex{}
	groupRequest(ml.actSubProtoIDs, _actSubProtoReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Activity().ActSubsProtocol(ctx, groupIDs)
			if err != nil {
				return err
			}
			mu.Lock()
			for id, act := range rly {
				if act == nil {
					continue
				}
				material.ActSubProtos[id] = act
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialAccount
func (ml *MaterialLoader) doAccounts(eg *errgroup.Group, material *Material) {
	if len(ml.mids) == 0 {
		return
	}
	if material.Accounts == nil {
		material.Accounts = make(map[int64]*accountgrpc.Info, len(ml.mids))
	}
	mu := sync.Mutex{}
	groupRequest(ml.mids, _accountReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Account().Infos3(ctx, groupIDs)
			if err != nil {
				return err
			}
			mu.Lock()
			for mid, account := range rly {
				if account == nil {
					continue
				}
				material.Accounts[mid] = account
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialHasDynsRly
func (ml *MaterialLoader) doHasDynsRlys(eg *errgroup.Group, material *Material) {
	if len(ml.hasDynsReqs) == 0 {
		return
	}
	if material.HasDynsRlys == nil {
		material.HasDynsRlys = make(map[RequestID]*dyntopicgrpc.HasDynsRsp, len(ml.hasDynsReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.hasDynsReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Dyntopic().HasDyns(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.HasDynsRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialListDynsRly
func (ml *MaterialLoader) doListDynsRlys(eg *errgroup.Group, material *Material) {
	if len(ml.listDynsReqs) == 0 {
		return
	}
	if material.ListDynsRlys == nil {
		material.ListDynsRlys = make(map[RequestID]*dyntopicgrpc.ListDynsRsp, len(ml.listDynsReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.listDynsReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Dyntopic().ListDyns(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.ListDynsRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialDynDetail
func (ml *MaterialLoader) doDynDetails(eg *errgroup.Group, material *Material) {
	if len(ml.dynDetailReqs) == 0 {
		return
	}
	if material.DynDetails == nil {
		material.DynDetails = make(map[RequestID]map[int64]*appdyngrpc.DynamicItem, len(ml.dynDetailReqs))
	}
	mu := sync.Mutex{}
	var playArgs *arcmidv1.PlayerArgs
	if batchPlayArg, ok := arcmid.FromContext(ml.c); ok {
		playArgs = util.Trans2PlayerArgs(batchPlayArg)
	}
	for k, v := range ml.dynDetailReqs {
		reqID := k
		reqTmp := v
		groupRequest(reqTmp.DynamicIds, _dynDetailReqMax, func(groupIDs []int64) {
			eg.Go(func(ctx context.Context) error {
				req := &appdyngrpc.DynServerDetailsReq{
					DynamicIds:    groupIDs,
					LocalTime:     ml.ss.LocalTime,
					PlayerArgs:    playArgs,
					MobiApp:       ml.ss.RawDevice().RawMobiApp,
					Device:        ml.ss.RawDevice().Device,
					Buvid:         ml.ss.RawDevice().Buvid,
					Build:         ml.ss.RawDevice().Build,
					Mid:           ml.ss.Mid(),
					Platform:      ml.ss.RawDevice().RawPlatform,
					IsMaster:      reqTmp.IsMaster,
					TopDynamicIds: reqTmp.TopDynamicIds,
				}
				rly, err := ml.dep.Appdyn().DynServerDetails(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				if material.DynDetails[reqID] == nil {
					material.DynDetails[reqID] = make(map[int64]*appdyngrpc.DynamicItem, len(rly))
				}
				for dynID, detail := range rly {
					if detail == nil {
						continue
					}
					material.DynDetails[reqID][dynID] = detail
				}
				mu.Unlock()
				return nil
			})
		})
	}
}

// model.MaterialActLikesRly
func (ml *MaterialLoader) doActLikesRlys(eg *errgroup.Group, material *Material) {
	if len(ml.actLikesReqs) == 0 {
		return
	}
	if material.ActLikesRlys == nil {
		material.ActLikesRlys = make(map[RequestID]*activitygrpc.LikesReply, len(ml.actLikesReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.actLikesReqs {
		if v.Req == nil {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Activity().ActLikes(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.ActLikesRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialDynRevsRly
func (ml *MaterialLoader) doDynRevsRlys(eg *errgroup.Group, material *Material) {
	if len(ml.dynRevsReqs) == 0 {
		return
	}
	if material.DynRevsRlys == nil {
		material.DynRevsRlys = make(map[RequestID]*dynfeedgrpc.FetchDynIdByRevsRsp, len(ml.dynRevsReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.dynRevsReqs {
		if len(v.DynRevsIds) == 0 {
			continue
		}
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Dynfeed().FetchDynIdByRevs(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.DynRevsRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialTag
func (ml *MaterialLoader) doTags(eg *errgroup.Group, material *Material) {
	if len(ml.tagIDs) == 0 {
		return
	}
	if material.Tags == nil {
		material.Tags = make(map[int64]*taggrpc.Tag, len(ml.tagIDs))
	}
	mu := sync.Mutex{}
	groupRequest(ml.tagIDs, _tagReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Tag().Tags(ctx, groupIDs, ml.ss.Mid())
			if err != nil {
				return err
			}
			mu.Lock()
			for tagID, tag := range rly {
				if tag == nil {
					continue
				}
				material.Tags[tagID] = tag
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialMixExtsRly
func (ml *MaterialLoader) doMixExtsRlys(eg *errgroup.Group, material *Material) {
	if len(ml.mixExtsReqs) == 0 {
		return
	}
	if material.MixExtsRlys == nil {
		material.MixExtsRlys = make(map[RequestID]*natpagegrpc.ModuleMixExtsReply, len(ml.mixExtsReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.mixExtsReqs {
		if v.Req == nil || v.Req.ModuleID == 0 {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Natpage().ModuleMixExts(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.MixExtsRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialGetHisRly
func (ml *MaterialLoader) doGetHisRlys(eg *errgroup.Group, material *Material) {
	if len(ml.getHisReqs) == 0 {
		return
	}
	if material.GetHisRlys == nil {
		material.GetHisRlys = make(map[RequestID]*actplatv2grpc.GetHistoryResp, len(ml.getHisReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.getHisReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Actplatv2().GetHistory(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.GetHisRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialPageArcsRly
func (ml *MaterialLoader) doPageArcsRlys(eg *errgroup.Group, material *Material) {
	if len(ml.pageArcsReqs) == 0 {
		return
	}
	if material.PageArcsRlys == nil {
		material.PageArcsRlys = make(map[RequestID]*populargrpc.PageArcsResp, len(ml.pageArcsReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.pageArcsReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Popular().PageArcs(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.PageArcsRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialMixExtRly
func (ml *MaterialLoader) doMixExtRlys(eg *errgroup.Group, material *Material) {
	if len(ml.mixExtReqs) == 0 {
		return
	}
	if material.MixExtRlys == nil {
		material.MixExtRlys = make(map[RequestID]*natpagegrpc.ModuleMixExtReply, len(ml.mixExtReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.mixExtReqs {
		if v.ModuleID == 0 {
			continue
		}
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Natpage().ModuleMixExt(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.MixExtRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialRankRstRly
func (ml *MaterialLoader) doRankRstRlys(eg *errgroup.Group, material *Material) {
	if len(ml.rankRstReqs) == 0 {
		return
	}
	if material.RankRstRlys == nil {
		material.RankRstRlys = make(map[RequestID]*activitygrpc.RankResultResp, len(ml.rankRstReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.rankRstReqs {
		if v.Req == nil || v.Req.RankID == 0 {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Activity().RankResult(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.RankRstRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialSelSerieRly
func (ml *MaterialLoader) doSelSerieRlys(eg *errgroup.Group, material *Material) {
	if len(ml.selSerieReqs) == 0 {
		return
	}
	if material.SelSerieRlys == nil {
		material.SelSerieRlys = make(map[RequestID]*appshowgrpc.SelectedSerieRly, len(ml.selSerieReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.selSerieReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Appshow().SelectedSerie(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.SelSerieRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialUpListRly
func (ml *MaterialLoader) doUpListRlys(eg *errgroup.Group, material *Material) {
	if len(ml.upListReqs) == 0 {
		return
	}
	if material.UpListRlys == nil {
		material.UpListRlys = make(map[RequestID]*activitygrpc.UpListReply, len(ml.upListReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.upListReqs {
		if v.Req == nil {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Activity().UpList(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.UpListRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialRelInfosRly
func (ml *MaterialLoader) doRelInfosRlys(eg *errgroup.Group, material *Material) {
	if len(ml.relInfosReqs) == 0 {
		return
	}
	if material.RelInfosRlys == nil {
		material.RelInfosRlys = make(map[RequestID]*chargrpc.CharacterRelInfosReply, len(ml.relInfosReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.relInfosReqs {
		if v.Req == nil {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Character().RelInfos(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.RelInfosRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialBriefDynsRly
func (ml *MaterialLoader) doBriefDynsRlys(eg *errgroup.Group, material *Material) {
	if len(ml.briefDynsReqs) == 0 {
		return
	}
	if material.BriefDynsRlys == nil {
		material.BriefDynsRlys = make(map[RequestID]*model.BriefDynsRly, len(ml.briefDynsReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.briefDynsReqs {
		if v.Req == nil {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Dyntopic().BriefDyns(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.BriefDynsRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialQueryWidRly
func (ml *MaterialLoader) doQueryWidRlys(eg *errgroup.Group, material *Material) {
	if len(ml.wids) == 0 {
		return
	}
	if material.QueryWidRlys == nil {
		material.QueryWidRlys = make(map[int32]*pgcappgrpc.QueryWidReply, len(ml.wids))
	}
	user := &pgcappgrpc.UserReq{
		Mid:      ml.ss.Mid(),
		MobiApp:  ml.ss.RawDevice().RawMobiApp,
		Device:   ml.ss.RawDevice().Device,
		Platform: ml.ss.RawDevice().RawPlatform,
		Build:    int32(ml.ss.RawDevice().Build),
	}
	mu := sync.Mutex{}
	for _, v := range ml.wids {
		wid := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Pgcapp().QueryWid(ctx, &pgcappgrpc.QueryWidReq{Wid: wid, User: user})
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.QueryWidRlys[wid] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialRoomsByActIdRly
func (ml *MaterialLoader) doRoomsByActIdRlys(eg *errgroup.Group, material *Material) {
	if len(ml.roomsByActIdReqs) == 0 {
		return
	}
	if material.RoomsByActIdRlys == nil {
		material.RoomsByActIdRlys = make(map[RequestID]*liveplaygrpc.GetListByActIdResp, len(ml.roomsByActIdReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.roomsByActIdReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.LivePlay().GetListByActId(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.RoomsByActIdRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialRelation
func (ml *MaterialLoader) doRelations(eg *errgroup.Group, material *Material) {
	if len(ml.relFids) == 0 {
		return
	}
	if material.Relations == nil {
		material.Relations = make(map[int64]*relationgrpc.FollowingReply, len(ml.relFids))
	}
	mu := sync.Mutex{}
	groupRequest(ml.relFids, _relFidsMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &relationgrpc.RelationsReq{
				Mid:    ml.ss.Mid(),
				Fid:    groupIDs,
				RealIp: ml.ss.Ip(),
			}
			rly, err := ml.dep.Relation().Relations(ctx, req)
			if err != nil {
				return err
			}
			mu.Lock()
			for fid, relation := range rly {
				if relation == nil {
					continue
				}
				material.Relations[fid] = relation
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialAccountCard
func (ml *MaterialLoader) doAccountCards(eg *errgroup.Group, material *Material) {
	if len(ml.cardMids) == 0 {
		return
	}
	if material.AccountCards == nil {
		material.AccountCards = make(map[int64]*accountgrpc.Card, len(ml.cardMids))
	}
	mu := sync.Mutex{}
	ip := ml.ss.Ip()
	groupRequest(ml.cardMids, _accountReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Account().Cards3(ctx, &accountgrpc.MidsReq{Mids: groupIDs, RealIp: ip})
			if err != nil {
				return err
			}
			mu.Lock()
			for mid, card := range rly {
				if card == nil {
					continue
				}
				material.AccountCards[mid] = card
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialArcPlayer
func (ml *MaterialLoader) doArcsPlayer(eg *errgroup.Group, material *Material) {
	if len(ml.playAvs) == 0 {
		return
	}
	if material.ArcsPlayer == nil {
		material.ArcsPlayer = make(map[int64]*arcgrpc.ArcPlayer, len(ml.playAvs))
	}
	batchPlayArg, _ := arcmid.FromContext(ml.c)
	mu := sync.Mutex{}
	for i := 0; i < len(ml.playAvs); i += _playAvsMax {
		var groupIDs []*arcgrpc.PlayAv
		if i+_playAvsMax < len(ml.playAvs) {
			groupIDs = ml.playAvs[i : i+_playAvsMax]
		} else {
			groupIDs = ml.playAvs[i:]
		}
		eg.Go(func(ctx context.Context) error {
			req := &arcgrpc.ArcsPlayerRequest{PlayAvs: groupIDs, BatchPlayArg: batchPlayArg}
			rly, err := ml.dep.Archive().ArcsPlayer(ctx, req)
			if err != nil {
				return err
			}
			mu.Lock()
			for aid, arc := range rly {
				if arc == nil {
					continue
				}
				material.ArcsPlayer[aid] = arc
			}
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialChannelFeedRly
func (ml *MaterialLoader) doChannelFeedRlys(eg *errgroup.Group, material *Material) {
	if len(ml.channelFeedReqs) == 0 {
		return
	}
	if material.ChannelFeedRlys == nil {
		material.ChannelFeedRlys = make(map[RequestID]*hmtgrpc.ChannelFeedReply, len(ml.channelFeedReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.channelFeedReqs {
		if v.Req == nil {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Hmt().ChannelFeed(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.ChannelFeedRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialNativeCard
func (ml *MaterialLoader) doNativePageCards(eg *errgroup.Group, material *Material) {
	if len(ml.pidsOfNaCard) == 0 {
		return
	}
	if material.NativePageCards == nil {
		material.NativePageCards = make(map[int64]*natpagegrpc.NativePageCard, len(ml.pidsOfNaCard))
	}
	mu := sync.Mutex{}
	groupRequest(ml.pidsOfNaCard, _naCardMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &natpagegrpc.NativePageCardsReq{
				Pids:     groupIDs,
				Device:   ml.ss.RawDevice().Device,
				MobiApp:  ml.ss.MobiApp(),
				Build:    int32(ml.ss.RawDevice().Build),
				Buvid:    ml.ss.Buvid(),
				Platform: ml.ss.Platform(),
			}
			rly, err := ml.dep.Natpage().NativePageCards(ctx, req)
			if err != nil {
				return err
			}
			mu.Lock()
			for pid, page := range rly {
				if page == nil {
					continue
				}
				material.NativePageCards[pid] = page
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialNativeAllPage
func (ml *MaterialLoader) doNativeAllPages(eg *errgroup.Group, material *Material) {
	if len(ml.pidsOfNaAll) == 0 {
		return
	}
	if material.NativeAllPages == nil {
		material.NativeAllPages = make(map[int64]*natpagegrpc.NativePage, len(ml.pidsOfNaAll))
	}
	mu := sync.Mutex{}
	groupRequest(ml.pidsOfNaAll, _naAllMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Natpage().NativeAllPages(ctx, &natpagegrpc.NativeAllPagesReq{Pids: groupIDs})
			if err != nil {
				return err
			}
			mu.Lock()
			for pid, page := range rly {
				if page == nil {
					continue
				}
				material.NativeAllPages[pid] = page
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialNativePages
func (ml *MaterialLoader) doNativePages(eg *errgroup.Group, material *Material) {
	if len(ml.pidsOfNaPages) == 0 {
		return
	}
	if material.NativePages == nil {
		material.NativePages = make(map[int64]*natpagegrpc.NativePage, len(ml.pidsOfNaPages))
	}
	mu := sync.Mutex{}
	groupRequest(ml.pidsOfNaPages, _naPagesMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Natpage().NativePages(ctx, &natpagegrpc.NativePagesReq{Pids: groupIDs})
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for pid, page := range rly.List {
				if page == nil {
					continue
				}
				material.NativePages[pid] = page
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialChannel
func (ml *MaterialLoader) doChannels(eg *errgroup.Group, material *Material) {
	if len(ml.channelIDs) == 0 {
		return
	}
	if material.Channels == nil {
		material.Channels = make(map[int64]*channelgrpc.Channel, len(ml.channelIDs))
	}
	mu := sync.Mutex{}
	groupRequest(ml.channelIDs, _channelMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &channelgrpc.InfosReq{Mid: ml.ss.Mid(), Cids: groupIDs}
			rly, err := ml.dep.Channel().Infos(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for cid, channel := range rly.CidMap {
				if channel == nil {
					continue
				}
				material.Channels[cid] = channel
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialVoteRankRly
func (ml *MaterialLoader) doVoteRankRlys(eg *errgroup.Group, material *Material) {
	if len(ml.VoteRankReqs) == 0 {
		return
	}
	if material.VoteRankRlys == nil {
		material.VoteRankRlys = make(map[RequestID]*activitygrpc.GetVoteActivityRankResp, len(ml.VoteRankReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.VoteRankReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Activity().GetVoteActivityRank(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.VoteRankRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialUpRsvInfo
func (ml *MaterialLoader) doUpRsvInfos(eg *errgroup.Group, material *Material) {
	if len(ml.upRsvIDsReqs) == 0 {
		return
	}
	var (
		sids   []int64
		reqIDs = make(map[int64][]RequestID)
	)
	for reqID, req := range ml.upRsvIDsReqs {
		sids = append(sids, req.IDs...)
		for _, id := range req.IDs {
			reqIDs[id] = append(reqIDs[id], reqID)
		}
	}
	if material.UpRsvInfos == nil {
		material.UpRsvInfos = make(map[RequestID]map[int64]*activitygrpc.UpActReserveRelationInfo, len(ml.upRsvIDsReqs))
	}
	mu := sync.Mutex{}
	groupRequest(sids, _upRsvInfoMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &activitygrpc.UpActReserveRelationInfoReq{Mid: ml.ss.Mid(), Sids: groupIDs}
			rly, err := ml.dep.Activity().UpActReserveRelationInfo(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for sid, upRsvInfo := range rly {
				if upRsvInfo == nil {
					continue
				}
				for _, reqID := range reqIDs[sid] {
					if material.UpRsvInfos[reqID] == nil {
						material.UpRsvInfos[reqID] = make(map[int64]*activitygrpc.UpActReserveRelationInfo)
					}
					material.UpRsvInfos[reqID][sid] = upRsvInfo
				}
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialRoomSessionInfo
func (ml *MaterialLoader) doRoomSessionInfos(eg *errgroup.Group, material *Material) {
	if len(ml.uidLiveIDs) == 0 {
		return
	}
	if material.RoomSessionInfos == nil {
		material.RoomSessionInfos = make(map[int64]*roomgategrpc.SessionInfos, len(ml.uidLiveIDs))
	}
	uidLiveIDs := make(map[int64]*roomgategrpc.LiveIds, len(ml.uidLiveIDs))
	for mid, liveIDs := range ml.uidLiveIDs {
		if len(liveIDs) == 0 {
			continue
		}
		uidLiveIDs[mid] = &roomgategrpc.LiveIds{LiveIds: liveIDs}
	}
	network := "other"
	if batchPlayArg, ok := arcmid.FromContext(ml.c); ok {
		switch batchPlayArg.NetType {
		case arcgrpc.NetworkType_NT_UNKNOWN:
			network = "other"
		case arcgrpc.NetworkType_WIFI:
			network = "wifi"
		default:
		}
	}
	playurl := &roomgategrpc.PlayUrlReq{
		Uid:         ml.ss.Mid(),
		Uipstr:      ml.ss.Ip(),
		HttpsUrlReq: ml.ss.HttpsUrlReq == 1,
		Platform:    ml.ss.RawDevice().RawPlatform,
		Build:       ml.ss.RawDevice().Build,
		DeviceName:  ml.ss.RawDevice().Device,
		Network:     network,
		ReqBiz:      "/bilibili.app.nativeact.v1.NativeAct/Index",
	}
	mu := sync.Mutex{}
	eg.Go(func(ctx context.Context) error {
		req := &roomgategrpc.SessionInfoBatchReq{
			UidLiveIds: uidLiveIDs,
			EntryFrom:  []string{model.LiveEnteryFrom},
			Playurl:    playurl,
		}
		rly, err := ml.dep.RoomGate().SessionInfoBatch(ctx, req)
		if err != nil {
			return err
		}
		if rly == nil {
			return nil
		}
		if material.RoomSessionInfos == nil {
			material.RoomSessionInfos = make(map[int64]*roomgategrpc.SessionInfos)
		}
		mu.Lock()
		for mid, infos := range rly.List {
			material.RoomSessionInfos[mid] = infos
		}
		mu.Unlock()
		return nil
	})
}

// model.MaterialTimelineRly
func (ml *MaterialLoader) doTimelineRlys(eg *errgroup.Group, material *Material) {
	if len(ml.timelineReqs) == 0 {
		return
	}
	if material.TimelineRlys == nil {
		material.TimelineRlys = make(map[RequestID]*populargrpc.TimeLineReply, len(ml.timelineReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.timelineReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Popular().TimeLine(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.TimelineRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialSeasonCard
func (ml *MaterialLoader) doSeasonCards(eg *errgroup.Group, material *Material) {
	if len(ml.ssids) == 0 {
		return
	}
	if material.SeasonCards == nil {
		material.SeasonCards = make(map[int32]*pgcappgrpc.SeasonCardInfoProto, len(ml.ssids))
	}
	mu := sync.Mutex{}
	user := &pgcappgrpc.UserReq{Mid: ml.ss.Mid()}
	groupRequestWithInt32(ml.ssids, _ssidReqMax, func(groupIDs []int32) {
		eg.Go(func(ctx context.Context) error {
			req := &pgcappgrpc.SeasonBySeasonIdReq{SeasonIds: groupIDs, User: user}
			rly, err := ml.dep.Pgcapp().SeasonBySeasonId(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for _, seasonCard := range rly.GetSeasonInfos() {
				if seasonCard == nil {
					continue
				}
				material.SeasonCards[seasonCard.GetSeasonId()] = seasonCard
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialSeasonByPlayIdRly
func (ml *MaterialLoader) doSeasonByPlayIdRly(eg *errgroup.Group, material *Material) {
	if len(ml.seasonByPlayIdReqs) == 0 {
		return
	}
	if material.SeasonByPlayIdRlys == nil {
		material.SeasonByPlayIdRlys = make(map[RequestID]*pgcappgrpc.SeasonByPlayIdReply, len(ml.seasonByPlayIdReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.seasonByPlayIdReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Pgcapp().SeasonByPlayId(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.SeasonByPlayIdRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.ActiveUsersRly
func (ml *MaterialLoader) doActiveUsersRly(eg *errgroup.Group, material *Material) {
	if len(ml.activeUsersReqs) == 0 {
		return
	}
	if material.ActiveUsersRlys == nil {
		material.ActiveUsersRlys = make(map[RequestID]*model.ActiveUsersRly, len(ml.activeUsersReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.activeUsersReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Dyntopic().ActiveUsers(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.ActiveUsersRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialDynVoteInfo
func (ml *MaterialLoader) doDynVoteInfo(eg *errgroup.Group, material *Material) {
	if len(ml.dynVoteIDs) == 0 {
		return
	}
	if material.DynVoteInfos == nil {
		material.DynVoteInfos = make(map[int64]*dyncommongrpc.VoteInfo, len(ml.dynVoteIDs))
	}
	mu := sync.Mutex{}
	groupRequest(ml.dynVoteIDs, _dynVoteInfoReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &dynvotegrpc.ListFeedVotesReq{Uid: ml.ss.Mid(), VoteIds: groupIDs}
			rly, err := ml.dep.Dynvote().ListFeedVotes(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for id, info := range rly.GetVoteInfos() {
				if info == nil {
					continue
				}
				material.DynVoteInfos[id] = info
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialUpRsvInfo
func (ml *MaterialLoader) doActSubject(eg *errgroup.Group, material *Material) {
	if len(ml.actSidsReqs) == 0 {
		return
	}
	var (
		sids   []int64
		reqIDs = make(map[int64][]RequestID)
	)
	for reqID, req := range ml.actSidsReqs {
		sids = append(sids, req.IDs...)
		for _, id := range req.IDs {
			reqIDs[id] = append(reqIDs[id], reqID)
		}
	}
	if material.ActSubjects == nil {
		material.ActSubjects = make(map[RequestID]map[int64]*activitygrpc.Subject, len(ml.actSidsReqs))
	}
	mu := sync.Mutex{}
	groupRequest(sids, _actSubReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &activitygrpc.ActSubjectsReq{Sids: groupIDs}
			rly, err := ml.dep.Activity().ActSubjects(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for sid, subject := range rly.List {
				if subject == nil {
					continue
				}
				for _, reqID := range reqIDs[sid] {
					if material.ActSubjects[reqID] == nil {
						material.ActSubjects[reqID] = make(map[int64]*activitygrpc.Subject)
					}
					material.ActSubjects[reqID][sid] = subject
				}
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialActProgressGroup
func (ml *MaterialLoader) doActProgressGroups(eg *errgroup.Group, material *Material) {
	if ml.actSidGroupIDs == nil {
		return
	}
	if material.ActProgressGroups == nil {
		material.ActProgressGroups = make(map[int64]map[int64]*activitygrpc.ActivityProgressGroup)
	}
	for k, gids := range ml.actSidGroupIDs {
		sid := k
		mu := sync.Mutex{}
		groupRequest(gids, _actProgGroupMax, func(groupIDs []int64) {
			eg.Go(func(ctx context.Context) error {
				req := &activitygrpc.ActivityProgressReq{
					Sid:  sid,
					Gids: groupIDs,
					Type: 2,
					Mid:  ml.ss.Mid(),
					Time: time.Now().Unix(),
				}
				rly, err := ml.dep.Activity().ActivityProgress(ctx, req)
				if err != nil {
					return err
				}
				mu.Lock()
				if material.ActProgressGroups[sid] == nil {
					material.ActProgressGroups[sid] = make(map[int64]*activitygrpc.ActivityProgressGroup)
				}
				for gid, group := range rly.Groups {
					if group == nil {
						continue
					}
					material.ActProgressGroups[sid][gid] = group
				}
				mu.Unlock()
				return nil
			})
		})
	}
}

// model.SourceDetailRly
func (ml *MaterialLoader) doSourceDetailRly(eg *errgroup.Group, material *Material) {
	if len(ml.sourceDetailReqs) == 0 {
		return
	}
	if material.SourceDetailRlys == nil {
		material.SourceDetailRlys = make(map[RequestID]*model.SourceDetailRly, len(ml.sourceDetailReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.sourceDetailReqs {
		if v.Req == nil {
			continue
		}
		reqID := k
		req := v.Req
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Business().SourceDetail(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.SourceDetailRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.ProductDetailRly
func (ml *MaterialLoader) doProductDetailRly(eg *errgroup.Group, material *Material) {
	if len(ml.productDetailReqs) == 0 {
		return
	}
	if material.ProductDetailRlys == nil {
		material.ProductDetailRlys = make(map[RequestID]*model.ProductDetailRly, len(ml.productDetailReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.productDetailReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Business().ProductDetail(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.ProductDetailRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialPgcFollowStatus
func (ml *MaterialLoader) doPgcFollowStatuses(eg *errgroup.Group, material *Material) {
	if len(ml.pgcFollowSeasonIds) == 0 {
		return
	}
	if material.PgcFollowStatuses == nil {
		material.PgcFollowStatuses = make(map[int32]*pgcfollowgrpc.FollowStatusProto, len(ml.pgcFollowSeasonIds))
	}
	mu := sync.Mutex{}
	groupRequestWithInt32(ml.pgcFollowSeasonIds, _pgcFollowReqMax, func(groupIDs []int32) {
		eg.Go(func(ctx context.Context) error {
			req := &pgcfollowgrpc.FollowStatusByMidReq{Mid: ml.ss.Mid(), SeasonId: groupIDs}
			rly, err := ml.dep.Pgcfollow().StatusByMid(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for seasonId, status := range rly.Result {
				if status == nil {
					continue
				}
				material.PgcFollowStatuses[seasonId] = status
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialComicInfo
func (ml *MaterialLoader) doComicInfos(eg *errgroup.Group, material *Material) {
	if len(ml.comicIds) == 0 {
		return
	}
	if material.ComicInfos == nil {
		material.ComicInfos = make(map[int64]*model.ComicInfo, len(ml.comicIds))
	}
	mu := sync.Mutex{}
	groupRequest(ml.comicIds, _comicInfoReqMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Comic().GetComicInfos(ctx, groupIDs, ml.ss.Mid())
			if err != nil {
				return err
			}
			mu.Lock()
			for comicId, comicInfo := range rly {
				if comicInfo == nil {
					continue
				}
				material.ComicInfos[comicId] = comicInfo
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialActReserveFollow
func (ml *MaterialLoader) doActRsvFollows(eg *errgroup.Group, material *Material) {
	if len(ml.actRsvIds) == 0 {
		return
	}
	if material.ActRsvFollows == nil {
		material.ActRsvFollows = make(map[int64]*activitygrpc.ReserveFollowingReply, len(ml.actRsvIds))
	}
	mu := sync.Mutex{}
	groupRequest(ml.actRsvIds, _actRsvFollowMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &activitygrpc.ReserveFollowingsReq{Sids: groupIDs, Mid: ml.ss.Mid()}
			rly, err := ml.dep.Activity().ReserveFollowings(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for sid, info := range rly.List {
				if info == nil {
					continue
				}
				material.ActRsvFollows[sid] = info
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialActAwardState
func (ml *MaterialLoader) doActAwardStates(eg *errgroup.Group, material *Material) {
	if len(ml.awardIds) == 0 {
		return
	}
	if material.AwardStates == nil {
		material.AwardStates = make(map[int64]*activitygrpc.AwardSubjectStateReply, len(ml.awardIds))
	}
	mu := sync.Mutex{}
	for _, v := range ml.awardIds {
		id := v
		eg.Go(func(ctx context.Context) error {
			req := &activitygrpc.AwardSubjectStateReq{Mid: ml.ss.Mid(), Id: id}
			rly, err := ml.dep.Activity().AwardSubjectState(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.AwardStates[id] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialTicketFavState
func (ml *MaterialLoader) doTicketStates(eg *errgroup.Group, material *Material) {
	if len(ml.ticketFavIds) == 0 {
		return
	}
	if material.TicketFavStates == nil {
		material.TicketFavStates = make(map[int64]bool, len(ml.ticketFavIds))
	}
	mu := sync.Mutex{}
	groupRequest(ml.ticketFavIds, _ticketFavStateMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			states, err := ml.dep.Mallticket().FavStatuses(ctx, groupIDs, ml.ss.Mid())
			if err != nil {
				return err
			}
			mu.Lock()
			for id, state := range states {
				material.TicketFavStates[id] = state
			}
			mu.Unlock()
			return nil
		})
	})
}

// model.MaterialActRelationInfo
func (ml *MaterialLoader) doActRelationInfos(eg *errgroup.Group, material *Material) {
	if len(ml.actRelationIds) == 0 {
		return
	}
	if material.ActRelationInfos == nil {
		material.ActRelationInfos = make(map[int64]*activitygrpc.ActRelationInfoReply, len(ml.actRelationIds))
	}
	mu := sync.Mutex{}
	for _, v := range ml.actRelationIds {
		id := v
		eg.Go(func(ctx context.Context) error {
			req := &activitygrpc.ActRelationInfoReq{Mid: ml.ss.Mid(), Id: id, Specific: "reserve"}
			rly, err := ml.dep.Activity().ActRelationInfo(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.ActRelationInfos[id] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialPlatCounterRes
func (ml *MaterialLoader) doPlatCounterResRlys(eg *errgroup.Group, material *Material) {
	if len(ml.platCounterReqs) == 0 {
		return
	}
	if material.PlatCounterResRlys == nil {
		material.PlatCounterResRlys = make(map[RequestID]*actplatv2grpc.GetCounterResResp, len(ml.platCounterReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.platCounterReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Actplatv2().GetCounterRes(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.PlatCounterResRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialPlatTotalRes
func (ml *MaterialLoader) doPlatTotalResRlys(eg *errgroup.Group, material *Material) {
	if len(ml.platTotalReqs) == 0 {
		return
	}
	if material.PlatTotalResRlys == nil {
		material.PlatTotalResRlys = make(map[RequestID]*actplatv2grpc.GetTotalResResp, len(ml.platTotalReqs))
	}
	mu := sync.Mutex{}
	for k, v := range ml.platTotalReqs {
		reqID := k
		req := v
		eg.Go(func(ctx context.Context) error {
			rly, err := ml.dep.Actplatv2().GetTotalRes(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.PlatTotalResRlys[reqID] = rly
			mu.Unlock()
			return nil
		})
	}
}

// activitygrpc.LotteryUnusedTimesReply
func (ml *MaterialLoader) doLotUnusedRlys(eg *errgroup.Group, material *Material) {
	if len(ml.lotteryIds) == 0 {
		return
	}
	if material.LotUnusedRlys == nil {
		material.LotUnusedRlys = make(map[string]*activitygrpc.LotteryUnusedTimesReply, len(ml.lotteryIds))
	}
	mu := sync.Mutex{}
	for _, v := range ml.lotteryIds {
		id := v
		eg.Go(func(ctx context.Context) error {
			req := &activitygrpc.LotteryUnusedTimesdReq{Sid: id, Mid: ml.ss.Mid()}
			rly, err := ml.dep.Activity().LotteryUnusedTimes(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			material.LotUnusedRlys[id] = rly
			mu.Unlock()
			return nil
		})
	}
}

// model.MaterialTicketFavState
func (ml *MaterialLoader) doScoreTargets(eg *errgroup.Group, material *Material) {
	if len(ml.scoreIds) == 0 {
		return
	}
	if material.ScoreTargets == nil {
		material.ScoreTargets = make(map[int64]*commscoregrpc.ScoreTarget, len(ml.scoreIds))
	}
	mu := sync.Mutex{}
	groupRequest(ml.scoreIds, _scoreIdsMax, func(groupIDs []int64) {
		eg.Go(func(ctx context.Context) error {
			req := &commscoregrpc.MultiGetTargetScoreReq{TntCode: 1, STargetType: 1, STargetIds: groupIDs}
			rly, err := ml.dep.Commscore().MultiGetTargetScore(ctx, req)
			if err != nil {
				return err
			}
			if rly == nil {
				return nil
			}
			mu.Lock()
			for id, score := range rly.GetTargets() {
				if score == nil {
					continue
				}
				material.ScoreTargets[id] = score
			}
			mu.Unlock()
			return nil
		})
	})
}

func groupRequest(ids []int64, max int, egf func(groupIDs []int64)) {
	ids = util.UniqueArray(ids)
	for i := 0; i < len(ids); i += max {
		var groupIDs []int64
		if i+max < len(ids) {
			groupIDs = ids[i : i+max]
		} else {
			groupIDs = ids[i:]
		}
		egf(groupIDs)
	}
}

func groupRequestWithInt32(ids []int32, max int, egf func(groupIDs []int32)) {
	ids = util.UniqueArrayWithInt32(ids)
	for i := 0; i < len(ids); i += max {
		var groupIDs []int32
		if i+max < len(ids) {
			groupIDs = ids[i : i+max]
		} else {
			groupIDs = ids[i:]
		}
		egf(groupIDs)
	}
}

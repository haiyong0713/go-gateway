package dao

import (
	"context"
	"sync"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	avecode "go-gateway/app/app-svr/archive/ecode"
	arcMidV1 "go-gateway/app/app-svr/archive/middleware/v1"
	purlSvcV2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/pkg/idsafe/bvid"

	arcSvc "go-gateway/app/app-svr/archive/service/api"

	accSvc "git.bilibili.co/bapis/bapis-go/account/service"
	coinSvc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favSvc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbupSvc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

const (
	// UGC和OGV稿件共享一个投币business
	CoinBusinessUGCOGV = "archive"
	// 专栏投币
	CoinBusinessArticle = "article"
	// 老音频投币
	CoinBusinessAudio = "audio"

	// UGC和OGV稿件共享一个点赞business
	ThumbUpBusinessUGCOGV = "archive"
	// 老音频点赞
	ThumbUpBusinessAudio = "audio"
)

type ArcDetailsOpt struct {
	Aids           map[int64][]int64
	Mid            int64
	RemoteIP       string
	Dev            *device.Device
	FilterFn       func(arc *arcSvc.Arc) bool // true to filter out
	CheckCopyRight bool                       // 是否检查版权播控
}

//nolint:gocognit
func (d *dao) ArchiveDetails(ctx context.Context, opt ArcDetailsOpt) (arcs map[int64]model.ArchiveDetail, err error) {
	aids := make([]int64, 0, len(opt.Aids))
	for k := range opt.Aids {
		aids = append(aids, k)
	}
	req := &arcSvc.ViewsRequest{Aids: aids, Mid: opt.Mid, MobiApp: opt.Dev.RawMobiApp, Device: opt.Dev.Device}
	resp, err := d.arcGRPC.Views(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "arcGRPC.Views", req)
		return
	}
	uniqueSeasonIDs := make(map[int64]struct{})
	uniqueAuthorsIDs := make(map[int64]struct{})
	arcs = make(map[int64]model.ArchiveDetail)
	aids = aids[0:0]
	for aid, arc := range resp.GetViews() {
		if arc == nil {
			log.Warnc(ctx, "unexpected nil view for aid(%d)", aid)
			continue
		}
		// 重新统计所有的aids 以接口返回的有数据的为准
		aids = append(aids, aid)

		if opt.FilterFn != nil && opt.FilterFn(arc.Arc) {
			continue
		}
		bv, err := bvid.AvToBv(aid)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to convert avid(%d) to bvid", aid)
		}
		curArc := model.ArchiveDetail{
			AC:    &model.ArcControl{},
			Bvid:  bv,
			Arc:   resp.GetViews()[aid].Arc,
			Pages: resp.GetViews()[aid].GetPages(),
			UpInfo: &v1.Author{
				Name:     arc.Author.Name,
				Mid:      arc.Author.Mid,
				Avatar:   arc.Author.Face,
				Relation: &v1.FollowRelation{Status: v1.FollowRelation_NO_FOLLOW},
			},
			Stat: &v1.BKStat{
				Like:      arc.Stat.Like,
				Coin:      arc.Stat.Coin,
				Favourite: arc.Stat.Fav,
				Reply:     arc.Stat.Reply,
				Share:     arc.Stat.Share,
				View:      arc.Stat.View,
			},
			Season: &model.ArcUGCSeason{},
		}
		arcs[aid] = curArc
		// 暂存所有author mid 稍后批量获取关注关系
		uniqueAuthorsIDs[arc.Author.Mid] = struct{}{}
		if arc.SeasonID > 0 {
			uniqueSeasonIDs[arc.SeasonID] = struct{}{}
		}
	}

	eg := errgroup.WithContext(ctx)
	if opt.Mid > 0 {
		// 批量更新关注信息
		if len(uniqueAuthorsIDs) > 0 {
			eg.Go(func(ctx context.Context) error {
				authorIDs := make([]int64, 0, len(uniqueAuthorsIDs))
				for k := range uniqueAuthorsIDs {
					authorIDs = append(authorIDs, k)
				}
				req := &accSvc.RelationsReq{
					Mid: opt.Mid, Owners: authorIDs, RealIp: opt.RemoteIP,
				}
				rels, err := d.accGRPC.Relations3(ctx, req)
				if err != nil {
					return wrapDaoError(err, "accGRPC.Relations3", req)
				}
				if rels.GetRelations() != nil {
					for i := range arcs {
						if rel, ok := rels.GetRelations()[arcs[i].UpInfo.Mid]; ok {
							if rel.Following {
								arcs[i].UpInfo.Relation.Status = v1.FollowRelation_FOLLOWING
							}
						}
					}
				}
				return nil
			})
		}
		if len(aids) > 0 {
			// 批量更新投币状态
			eg.Go(func(ctx context.Context) error {
				req := &coinSvc.ItemsUserCoinsReq{
					Mid: opt.Mid, Aids: aids, Business: CoinBusinessUGCOGV,
				}
				coins, err := d.coinGRPC.ItemsUserCoins(ctx, req)
				if err != nil {
					return wrapDaoError(err, "coinGRPC.ItemsUserCoins", req)
				}
				for aid, coin := range coins.GetNumbers() {
					arcs[aid].Stat.HasCoin = coin > 0
				}
				return nil
			})
			// 批量更新稿件点赞状态
			eg.Go(func(ctx context.Context) error {
				req := &thumbupSvc.HasLikeReq{
					Mid: opt.Mid, Business: ThumbUpBusinessUGCOGV, MessageIds: aids, IP: opt.RemoteIP,
				}
				thumbs, err := d.thumbupGRPC.HasLike(ctx, req)
				if err != nil {
					return wrapDaoError(err, "thumbupGRPC.HasLike", req)
				}
				for aid, state := range thumbs.States {
					arcs[aid].Stat.HasLike = state.State == thumbupSvc.State_STATE_LIKE
				}
				return nil
			})
			// 批量更新稿件收藏状态
			eg.Go(func(ctx context.Context) error {
				req := &favSvc.IsFavoredsReq{
					// TODO: 这里type是收藏元素的类型
					// 对于ogv稿件需要可能需要额外处理
					Typ:  model.FavTypeVideo,
					Mid:  opt.Mid,
					Oids: aids,
				}
				favs, err := d.favGRPC.IsFavoreds(ctx, req)
				if err != nil {
					return wrapDaoError(err, "favGRPC.IsFavoreds", req)
				}
				for aid, state := range favs.GetFaveds() {
					arcs[aid].Stat.HasFav = state
				}
				return nil
			})
		}
	} else {
		// 未登录态
		if len(aids) > 0 {
			eg.Go(func(ctx context.Context) error {
				req := &thumbupSvc.BuvidHasLikeReq{
					Buvid: opt.Dev.Buvid, Business: ThumbUpBusinessUGCOGV, MessageIds: aids, IP: opt.RemoteIP,
				}
				thumbs, err := d.thumbupGRPC.BuvidHasLike(ctx, req)
				if err != nil {
					return wrapDaoError(err, "thumbupGRPC.BuvidHasLike", req)
				}
				for aid, state := range thumbs.States {
					arcs[aid].Stat.HasLike = state.State == thumbupSvc.State_STATE_LIKE
				}
				return nil
			})
		}
	}
	// 批量更新版权播控信息
	if opt.CheckCopyRight {
		eg.Go(func(ctx context.Context) error {
			banPlay, err := d.CopyrightBans(ctx, CopyrightBansOpt{Aids: aids})
			if err != nil {
				return err
			}
			for aid, banPlay := range banPlay {
				arcs[aid].AC.CopyRightBan = banPlay
			}
			return nil
		})
	}
	// 批量更新ugc合集状态
	if len(uniqueSeasonIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			metas := make([]model.FavFolderMeta, 0, len(uniqueSeasonIDs))
			for sid := range uniqueSeasonIDs {
				metas = append(metas, model.FavFolderMeta{
					Typ: model.FavTypeUgcSeason,
					Fid: sid,
				})
			}
			req := FavFoldersInfoOpt{
				Dev: opt.Dev, IP: opt.RemoteIP, Metas: metas, Mid: opt.Mid,
			}
			seasons, err := d.FavFoldersInfo(ctx, req)
			if err != nil {
				return wrapDaoError(err, "dao.FavFoldersInfo", req)
			}
			for _, arc := range arcs {
				if season, ok := seasons[arc.UGCSeasonMetaHash()]; ok {
					arc.Season.FavFolder = season
				}
			}
			return nil
		})
	}

	// 因为相关信息不是特别重要，所以只log不完全报错
	sideErr := eg.Wait()
	if sideErr != nil {
		log.Warnc(ctx, "failed to get associated info for ArchiveDetail: %v Discarded", sideErr)
	}

	return
}

type ArcPlayUrlOpt struct {
	Aid        int64
	Cids       []int64
	Mid        int64
	Dev        *device.Device
	Net        *network.Network
	PlayerArgs *arcMidV1.PlayerArgs
}

//nolint:gocognit
func (d *dao) ArcPlayUrl(ctx context.Context, opt ArcPlayUrlOpt) (map[int64]model.PlayUrlInfo, error) {
	var (
		simpleArc *arcSvc.SimpleArcReply
		banPlay   bool
	)

	eg1 := errgroup.WithContext(ctx)
	eg1.Go(func(c context.Context) (err error) {
		simpleArc, err = d.arcGRPC.SimpleArc(c, &arcSvc.SimpleArcRequest{Aid: opt.Aid})
		if err != nil {
			return wrapDaoError(err, "arcGRPC.SimpleArc", opt.Aid)
		}
		if simpleArc.GetArc() == nil {
			// 稿件完全不存在 大概率端上bug导致类型传错了 小概率是simpleArc接口的问题
			// 兼容一下端上老版本音频类型传递错误的问题 返回特定ecode帮助上层判断
			return errors.WithMessagef(avecode.ArchiveNotExist, "archive info not found. possible auid due to app bug aid(%v) cids(%v)", opt.Aid, opt.Cids)
		}
		// 如果没给cid  默认取第一p
		if len(opt.Cids) == 0 {
			opt.Cids = simpleArc.GetArc().GetCids()
			if len(opt.Cids) <= 0 {
				return errors.WithMessagef(ecode.NothingFound, "no cids found ArcPlayUrlOpt(%+v)", opt)
			}
			opt.Cids = opt.Cids[0:1]
		} else {
			// 检查 是否有请求的cid不存在
			if len(opt.Cids) > 0 {
				existingCids := make(map[int64]struct{})
				for _, cid := range simpleArc.GetArc().GetCids() {
					existingCids[cid] = struct{}{}
				}
				notFound := false
				for _, reqCid := range opt.Cids {
					if _, ok := existingCids[reqCid]; !ok {
						notFound = true
						break
					}
				}
				// 有想要的cid找不到，而且oid和cid相同（音频特殊设定）
				// 大概率是传错类型
				if notFound && opt.Aid == opt.Cids[0] {
					// 兼容一下端上老版本音频类型传递错误的问题 返回特定ecode帮助上层判断
					return errors.WithMessagef(avecode.ArchiveNotExist, "archive info not found. possible auid due to app bug aid(%v) cids(%v)", opt.Aid, opt.Cids)
				}
			}
			const _maxCids = 3
			if len(opt.Cids) > _maxCids {
				opt.Cids = opt.Cids[0:_maxCids]
			}
		}
		if simpleArc.GetArc().GetState() < 0 {
			return errors.WithMessagef(ecode.NothingFound, "archive state<0 detail(%+v)", simpleArc.GetArc())
		}
		return nil
	})
	eg1.Go(func(c context.Context) (err error) {
		banPlay, err = d.CopyrightBan(c, opt.Aid)
		if err != nil {
			log.Warnc(c, "failed to check copyright status for aid(%d). banPlay Discarded", opt.Aid)
		}
		return nil
	})
	err := eg1.Wait()
	if err != nil {
		return nil, err
	}

	resultCh := make(chan model.PlayUrlInfo, 3)
	eg2 := errgroup.WithContext(ctx)
	for _, cid := range opt.Cids {
		cidCopy := cid
		eg2.Go(func(ctx context.Context) error {
			req := &purlSvcV2.PlayURLReq{
				Aid:       opt.Aid,
				Cid:       cidCopy,
				Qn:        opt.PlayerArgs.Qn,
				Platform:  opt.Dev.RawPlatform,
				Fnver:     int32(opt.PlayerArgs.Fnver),
				Fnval:     int32(opt.PlayerArgs.Fnval),
				Mid:       opt.Mid,
				BackupNum: 2, // 客户端默认2个
				//Download:     0,
				ForceHost:    int32(opt.PlayerArgs.ForceHost),
				Device:       opt.Dev.Device,
				MobiApp:      opt.Dev.RawMobiApp,
				H5Hq:         false,
				Build:        int32(opt.Dev.Build),
				Buvid:        opt.Dev.Buvid,
				NetType:      purlSvcV2.NetworkType(opt.Net.Type),
				TfType:       purlSvcV2.TFType(opt.Net.TF),
				VoiceBalance: opt.PlayerArgs.VoiceBalance,
			}
			resp, err := d.playUrlV2GRPC.PlayURL(ctx, req)
			if err != nil {
				return wrapDaoError(err, "playUrlV2GRPC.PlayURL", req)
			}
			if resp.Playurl == nil {
				return errors.WithMessagef(ecode.NothingFound, "dao.playUrlV2GRPC.PlayURL unexpected nil Playurl for ArcPlayUrlOpt(%+v)", opt)
			}
			u := resp.Playurl
			ret := model.PlayUrlInfo{
				Arc:          simpleArc.Arc,
				CopyrightBan: banPlay,
				Cid:          cidCopy,
				Qn:           u.Quality,
				Format:       u.Format,
				QnType:       u.Type,
				FnVer:        u.Fnver,
				FnVal:        u.Fnval,
				Formats:      model.TransformFormats(u.SupportFormats),
				VideoCodecID: u.VideoCodecid,
				Length:       u.Timelength,
				Code:         u.Code,
				Message:      u.Message,
				Volume:       resp.Volume,
			}
			switch u.Type {
			case model.QnTypeFLV, model.QnTypeMP4:
				ret.PlayUrl = &v1.PlayInfo_PlayUrl{PlayUrl: model.TransformPlayUrl(u.Durl)}
			case model.QnTypeDASH:
				ret.PlayDash = &v1.PlayInfo_PlayDash{PlayDash: model.TransformPlayDash(u.Dash)}
			default:
				if u.Dash != nil {
					ret.QnType = model.QnTypeDASH // DASH
					ret.PlayDash = &v1.PlayInfo_PlayDash{PlayDash: model.TransformPlayDash(u.Dash)}
				} else if len(u.Durl) > 0 {
					ret.QnType = model.QnTypeFLV // default FLV
					ret.PlayUrl = &v1.PlayInfo_PlayUrl{PlayUrl: model.TransformPlayUrl(u.Durl)}
				} else {
					return errors.WithMessagef(ecode.ServerErr, "dao.playUrlV2GRPC.PlayURL unexpected QnType(%d) resp: %+v", u.Type, u)
				}
			}
			resultCh <- ret
			return nil
		})
	}
	err = eg2.Wait()
	close(resultCh)
	if err != nil {
		return nil, err
	}
	ret := make(map[int64]model.PlayUrlInfo)
	for info := range resultCh {
		ret[info.Cid] = info
	}
	return ret, nil
}

type ArchiveInfoOpt struct {
	Aids   []int64
	Mid    int64
	Device *device.Device
}

func (d *dao) ArchiveInfos(ctx context.Context, opt ArchiveInfoOpt) (map[int64]model.ArchiveInfo, error) {
	mu := sync.Mutex{}
	eg := errgroup.WithContext(ctx)
	maxBatch := 50
	ret := make(map[int64]model.ArchiveInfo)
	for i := 0; i < len(opt.Aids); i += maxBatch {
		var partial []int64
		if i+maxBatch > len(opt.Aids) {
			partial = opt.Aids[i:]
		} else {
			partial = opt.Aids[i : i+maxBatch]
		}
		eg.Go(func(c context.Context) error {
			arcs, err := d.archiveInfos(c, partial, opt.Mid, opt.Device)
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			for i, arc := range arcs {
				ret[arc.Arc.GetAid()] = arcs[i]
			}
			return nil
		})
	}
	return ret, eg.Wait()
}

func (d *dao) archiveInfos(ctx context.Context, aids []int64, mid int64, dev *device.Device) ([]model.ArchiveInfo, error) {
	req := &arcSvc.ArcsRequest{
		Aids:    aids,
		Mid:     mid,
		MobiApp: dev.RawMobiApp,
		Device:  dev.Device,
	}
	arcs, err := d.arcGRPC.Arcs(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "arcGRPC.Arcs", req)
	}
	ret := make([]model.ArchiveInfo, 0, len(arcs.GetArcs()))
	for _, arc := range arcs.GetArcs() {
		ret = append(ret, model.ArchiveInfo{
			Arc: arcs.GetArcs()[arc.Aid],
		})
	}
	return ret, nil
}

func (d *dao) FilterArchives(ctx context.Context, aids []int64) (map[int64]string, error) {
	resp, err := d.listenerGRPC.FilterArchives(ctx, &listenerSvc.FilterArchivesReq{Aids: aids})
	if err != nil {
		return nil, wrapDaoError(err, "listenerGRPC.FilterArchives", aids)
	}
	return resp.GetAids(), nil
}

package service

import (
	"context"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	avecode "go-gateway/app/app-svr/archive/ecode"
	archiveSvc "go-gateway/app/app-svr/archive/middleware/v1"

	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) PlayURL(ctx context.Context, req *v1.PlayURLReq) (ret *v1.PlayURLResp, err error) {
	defer func() { ret.AddExpireTime() }()
	if req.PlayerArgs == nil {
		return nil, errors.WithMessage(ecode.RequestErr, "unexpected empty player args for PlayURL")
	}
	err = validatePlayItem(ctx, req.Item, 0)
	if err != nil {
		return
	}

	dev, net, auth := DevNetAuthFromCtx(ctx)
	ret = &v1.PlayURLResp{
		Item:       req.Item,
		PlayerInfo: make(map[int64]*v1.PlayInfo),
	}
	isAudioRetry := false

REDO:
	switch req.Item.ItemType {
	case model.PlayItemUGC:
		infos, err := s.dao.ArcPlayUrl(ctx, dao.ArcPlayUrlOpt{
			Aid:        req.Item.Oid,
			Cids:       req.Item.SubId,
			Mid:        auth.Mid,
			Dev:        dev,
			Net:        net,
			PlayerArgs: req.PlayerArgs,
		})
		if err != nil {
			if ecode.EqualError(avecode.ArchiveNotExist, err) {
				// 如果不存在该稿件的任何信息，尝试一次音频稿件解析 兼容一下端上bug
				// TODO: remove when app ready
				req.Item.ItemType = model.PlayItemAudio
				isAudioRetry = true
				goto REDO
			}
			return nil, err
		}
		for cid, info := range infos {
			ret.Playable, ret.Message = info.CanPlay()
			ret.PlayerInfo[cid] = info.ToV1PlayInfo()
		}

	case model.PlayItemOGV:
		// TODO: ogv
		return nil, errors.WithMessage(ecode.RequestErr, "OGV PlayURL is not supported yet")

	case model.PlayItemAudio:
		songDetail, err := s.dao.SongPlayingDetailV1(ctx, dao.SongPlayingDetailOpt{
			SongId: req.Item.Oid, Mid: auth.Mid, PlayerArgs: req.PlayerArgs, Net: net, Dev: dev,
		})
		if err != nil {
			return nil, err
		}
		if isAudioRetry {
			log.Warnc(ctx, "audio resolve fallback succeed for item(%+v)", req.Item)
		}
		ret.Playable, ret.Message = songDetail.CanPlay()
		ret.PlayerInfo[req.Item.Oid] = songDetail.ToV1PlayInfo(req.PlayerArgs)

	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "unknown playItem type %d", req.Item.ItemType)
	}

	if ret.Playable != model.PlayableYES {
		ret.PlayerInfo = nil
	}
	return
}

//nolint:gocognit
func (s *Service) BKArcDetails(ctx context.Context, req *v1.BKArcDetailsReq) (reply *v1.BKArcDetailsResp, err error) {
	if len(req.Items) == 0 {
		return nil, errors.WithMessage(ecode.RequestErr, "unexpected req.Items.len=0")
	}
	// 限制批量数
	if len(req.Items) > MaxPageSize {
		return nil, errors.WithMessagef(ecode.RequestErr, "page size exceed, must<=%d", MaxPageSize)
	}
	ugcAids := make(map[int64][]int64)
	ogvSids := make(map[int64][]int64)
	auids := make(map[int64]struct{})

	for _, item := range req.Items {
		err = validatePlayItem(ctx, item, 0)
		if err != nil {
			return
		}
		switch item.GetItemType() {
		case model.PlayItemUGC:
			if item.GetOid() <= 0 {
				return nil, errors.WithMessagef(ecode.RequestErr, "unexpected UGC aid %d", item.GetOid())
			}
			if len(item.GetSubId()) > 0 {
				ugcAids[item.GetOid()] = append(ugcAids[item.GetOid()], item.GetSubId()...)
			} else {
				ugcAids[item.GetOid()] = nil
			}

		case model.PlayItemOGV:
			if item.GetOid() <= 0 {
				return nil, errors.WithMessagef(ecode.RequestErr, "unexpected OGV epid %d", item.GetOid())
			}
			if len(item.GetSubId()) > 0 {
				ogvSids[item.GetOid()] = append(ogvSids[item.GetOid()], item.GetSubId()...)
			} else {
				ogvSids[item.GetOid()] = nil
			}

		case model.PlayItemAudio:
			if item.GetOid() <= 0 {
				return nil, errors.WithMessagef(ecode.RequestErr, "unexpected auid %d", item.GetOid())
			}
			auids[item.GetOid()] = struct{}{}

		default:
			return nil, errors.WithMessagef(ecode.RequestErr, "unknown item_type %d", item.GetItemType())
		}
	}
	dev, net, auth := DevNetAuthFromCtx(ctx)
	sortedRes := make(map[int32]map[int64]*v1.DetailItem)
	filterControl := make(map[int64]string)
	mu := sync.Mutex{}

	eg := errgroup.WithContext(ctx)
	// ugc稿件
	if len(ugcAids) > 0 {
		eg.Go(func(ctx context.Context) error {
			ugcResps, err := s.dao.ArchiveDetails(ctx, dao.ArcDetailsOpt{
				Aids:           ugcAids,
				Mid:            auth.Mid,
				RemoteIP:       net.RemoteIP,
				CheckCopyRight: true,
				Dev:            dev,
			})
			if err != nil {
				return err
			}
			ugcDetails := make(map[int64]*v1.DetailItem)
			for aid, item := range ugcResps {
				ugcDetails[aid] = item.ToV1DetailItem(model.PlayItemUGC)
			}
			mu.Lock()
			sortedRes[model.PlayItemUGC] = ugcDetails
			mu.Unlock()
			return nil
		})
		// 服务端过滤
		if model.BkArchiveArgs(ctx).EnableServerFilter {
			eg.Go(func(ctx context.Context) (err error) {
				aids := make([]int64, 0, len(ugcAids))
				for aid := range ugcAids {
					aids = append(aids, aid)
				}
				filterControl, err = s.dao.FilterArchives(ctx, aids)
				if err != nil {
					log.Warnc(ctx, "error get filter ctrl while fetching BKArchiveDetails: %v", err)
				}
				return nil
			})
		}
	}
	// 老音频稿件
	if len(auids) > 0 {
		eg.Go(func(ctx context.Context) error {
			auResps, err := s.dao.SongDetailsV1(ctx, dao.SongDetailsOpt{
				SongIds: auids,
				Mid:     auth.Mid,
				Net:     net,
				Dev:     dev,
			})
			if err != nil {
				return err
			}
			auDetails := make(map[int64]*v1.DetailItem)
			for auid, item := range auResps {
				auDetails[auid] = item.ToV1DetailItem()
			}
			mu.Lock()
			sortedRes[model.PlayItemAudio] = auDetails
			mu.Unlock()
			return nil
		})
	}

	// TODO: unimplemented
	// ogv稿件
	//ogvResps, err := s.dao.SeasonDetails(ctx, dao.SeasonDetailsOpt{Sids: ogvSids})
	//if err != nil {
	//	log.Errorc(ctx, "dao.SeasonDetails failed: %+v", err)
	//	return
	//}
	//items = append(items, toDetailItems(ogvResps, PlayItemOGV)...)

	if err = eg.Wait(); err != nil {
		return
	}

	items := make([]*v1.DetailItem, 0, len(req.Items))
	var sorted map[int64]*v1.DetailItem
	for _, itm := range req.Items {
		// keep order
		sorted = sortedRes[itm.GetItemType()]
		if sorted == nil {
			log.Warnc(ctx, "no details sorted for playitem type(%d) item(%+v). Discarded", itm.GetItemType(), itm)
			continue
		}
		item := sorted[itm.GetOid()]
		if item == nil {
			log.Warnc(ctx, "no details item for playitem (%+v). Discarded", itm)
			continue
		}
		// UGC稿件服务端过滤，播单和历史均不展示
		if itm.GetItemType() == model.PlayItemUGC && len(filterControl[itm.Oid]) > 0 {
			continue
		}
		items = append(items, item)
	}
	reply = &v1.BKArcDetailsResp{List: items}
	return
}

func (s *Service) fillPlayerArgs(ctx context.Context, args *archiveSvc.PlayerArgs, items ...*v1.DetailItem) {
	if args == nil {
		return
	}
	eg := errgroup.WithContext(ctx)
	for _, it := range items {
		if it == nil {
			continue
		}
		if it.Item.GetItemType() == model.PlayItemAudio {
			continue
		}
		item := it
		eg.Go(func(ctx context.Context) error {
			// 优先秒开最后播放分p
			if item.LastPart != 0 && len(item.Item.SubId) <= 0 {
				item.Item.SubId = []int64{item.LastPart}
			}
			res, err := s.PlayURL(ctx, &v1.PlayURLReq{Item: item.Item, PlayerArgs: args})
			if err != nil {
				log.Errorc(ctx, "error handle player args for (%+v) Discarded: %v", item.Item, err)
				return nil
			}
			item.PlayerInfo = res.PlayerInfo
			item.Playable, item.Message = res.Playable, res.Message
			return nil
		})
	}
	_ = eg.Wait()
}

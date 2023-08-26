package service

import (
	"context"
	"strconv"

	"go-common/component/metadata/auth"
	"go-common/library/ecode"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

type playlistResult struct {
	ItemList []*v1.PlayItem
}

type mediaListResOpt struct {
	Id, Extra int64
	Anchor    *v1.PlayItem
	PageOpt   *v1.PageOption
	Auth      *auth.Auth
}

func (s *Service) mediaListResources(ctx context.Context, opt mediaListResOpt) (resp *playlistResult, err error) {
	resp = new(playlistResult)
	switch opt.Extra {
	// 收藏夹相关类型直接走收藏夹
	case model.MediaListTypeFav, model.MediaListTypeWeeklyRank:
		folderMeta := model.FavFolderMeta{Typ: model.FavTypeVideo}
		folderMeta.Fid, folderMeta.Mid = extractFidAndMid(opt.Id)
		var anchorMeta model.FavItemMeta
		if opt.Anchor != nil {
			anchorMeta.Otype, anchorMeta.Oid = model.Play2Fav[opt.Anchor.ItemType], opt.Anchor.Oid
		}
		folderDetail, err := s.dao.FavFolderDetail(ctx, dao.FavFolderDetailOpt{
			Mid:    opt.Auth.Mid,
			Folder: folderMeta,
			Anchor: anchorMeta,
		})
		if err != nil {
			if err == model.ErrAnchorNotFound {
				return nil, errors.WithMessagef(ecode.RequestErr, "MediaList: dao.FavFolderDetail: %v, anchor(%+v) page(%+v)", err, opt.Anchor, opt.PageOpt)
			}
			return nil, err
		}
		for _, dt := range folderDetail {
			pa := dt.ToV1PlayItem()
			if pa != nil {
				resp.ItemList = append(resp.ItemList, pa)
			}
		}

	default:
		// 其他场景走播单网关
		respList, err := s.dao.MediaListDetail(ctx, dao.MediaListDetailOpt{
			Typ: opt.Extra, BizId: opt.Id, Anchor: opt.Anchor,
		})
		if err != nil {
			return nil, err
		}
		for _, md := range respList {
			pa := md.ToV1PlayItem()
			if pa != nil {
				resp.ItemList = append(resp.ItemList, pa)
			}
		}
	}
	return
}

func (s *Service) Medialist(ctx context.Context, req *v1.MedialistReq) (resp *v1.MedialistResp, err error) {
	if req.BizId <= 0 || req.ListType <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "illegal bizId/lisType")
	}
	_, net, _ := DevNetAuthFromCtx(ctx)

	var data *dao.MediaListPagedResp
	var filterRes map[int64]string
	resp = new(v1.MedialistResp)

	eg := errgroup.WithContext(ctx)
	// 获取稿件信息
	eg.Go(func(ctx context.Context) (err error) {
		data, err = s.dao.MediaListPaged(ctx, dao.MediaListPagedOpt{
			Typ: req.ListType, BizId: req.BizId, Offset: req.Offset,
		})
		if err != nil {
			return err
		}
		resp.Total, resp.HasMore = data.Total, data.HasMore
		resp.Offset = data.Offset
		resp.Items = make([]*v1.MedialistItem, 0, len(data.Items))
		// 分拣aid出来过滤下
		aids := make([]int64, 0, len(data.Items))
		for _, m := range data.Items {
			if v1M := m.ToV1MedialistItem(); v1M != nil {
				resp.Items = append(resp.Items, v1M)
				if v1M.Item.ItemType == model.PlayItemUGC {
					aids = append(aids, m.Avid)
				}
			}

		}
		if len(aids) > 0 {
			filterRes, _ = s.dao.FilterArchives(ctx, aids)
		}
		markFn := func(item *v1.MedialistItem) {} // noop
		if filterRes != nil {
			markFn = func(item *v1.MedialistItem) {
				if item.Item.ItemType != model.PlayItemUGC {
					return
				}
				if reason := filterRes[item.Item.Oid]; len(reason) > 0 {
					item.State, item.Message = model.PlayableNO, reason
				}
			}
		}
		// 回填埋点
		eventTracking := func(et *v1.EventTracking) {
			et.TrackId, et.Batch = strconv.FormatInt(req.BizId, 10), strconv.FormatInt(req.ListType, 10)
		}
		// 标记稿件不可播状态
		for _, m := range resp.Items {
			markFn(m)
			m.Item.SetEventTracking(v1.OpMediaList, eventTracking)
		}
		return nil
	})

	if req.ListType == model.MediaListTypeSpace {
		eg.Go(func(ctx context.Context) error {
			res, err := s.dao.UpInfoStatByMid(ctx, req.BizId, net.RemoteIP)
			if err != nil {
				return nil
			}
			resp.UpInfo = res.ToV1MedialistUpInfo()
			return nil
		})
	}

	return resp, eg.Wait()
}

package dynamic

import (
	"context"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"

	egv2 "go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
	pplApi "go-gateway/app/app-svr/app-show/interface/api"

	arcApi "go-gateway/app/app-svr/archive/service/api"

	thumbApi "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

type SvHandlerFunc func(context.Context, *dynmdl.SVideoMaterial, *dynmdl.Header, *api.SVideoReq) error

// nolint:gocognit
func (s *Service) SVideo(c context.Context, req *api.SVideoReq, mid int64, header *dynmdl.Header, _ *dynmdl.VideoMate, idx int64) (res *api.SVideoReply, err error) {
	res = new(api.SVideoReply)
	offset := req.Offset
	needOffset := 0
	if offset == "" { //首次动态入口进来
		offset = strconv.FormatInt(req.Oid, 10)
		needOffset = 1
	}
	var dynList *dynmdl.DynSVideoList
	// nolint:exhaustive
	switch req.Type {
	case api.SVideoType_TypeDynamic: // 动态联播页
		dynList, err = s.dynDao.SVideo(c, offset, needOffset, mid)
	case api.SVideoType_TypePopularIndex: // 热门分类
		var pplIdxReply *pplApi.IndexSVideoReply
		if pplIdxReply, err = s.dynDao.PopularIndexSv(c, idx, req.Oid); err != nil {
			return nil, err
		}
		dynList = &dynmdl.DynSVideoList{}
		dynList.FromPplIdx(pplIdxReply, idx, req.FocusAid)
		if pplIdxReply.Top != nil {
			res.Top = dynmdl.ToTop(pplIdxReply.Top)
		}
	case api.SVideoType_TypePopularHotword: // 热点聚合
		var aggReply *pplApi.AggrSVideoReply
		if aggReply, err = s.dynDao.PopularAggrSv(c, idx, req.Oid); err != nil {
			return nil, err
		}
		dynList = &dynmdl.DynSVideoList{}
		dynList.FromPplAggr(aggReply, idx, req.FocusAid)
		if aggReply.Top != nil {
			res.Top = dynmdl.ToTop(aggReply.Top)
		}
	}
	if err != nil {
		log.Error("s.dynDao.SVideo req(%+v) err(%+v) offset(%s) needOffset(%d)", req, err, offset, needOffset)
		return nil, err
	}
	if dynList == nil {
		return nil, ecode.NothingFound
	}
	res.Offset = dynList.Offset
	res.HasMore = dynList.HasMore
	var (
		aids         []*arcApi.PlayAv
		upids        []int64
		thumBus      = make(map[string][]*dynmdl.LikeBusiItem)
		thumbBusItem []*dynmdl.LikeBusiItem
		thumbupm     *thumbApi.MultiStatsReply
		arcm         = make(map[int64]*arcApi.ArcPlayer)
		attm         = make(map[int64]int32, len(upids))
	)
	for _, v := range dynList.MixVideoItem {
		aids = append(aids, &arcApi.PlayAv{Aid: v.RID})
		upids = append(upids, v.UID)
		thumbBusItem = append(thumbBusItem, &dynmdl.LikeBusiItem{MsgID: v.RID})
	}
	thumBus[dynmdl.BusTypeVideo] = thumbBusItem
	eg := egv2.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			arcm, err = s.arcDao.ArcsPlayer(ctx, aids, false, "")
			if err != nil {
				log.Error("s.arcDao.ArcsWithPlayurl err(%v)", err)
				return nil
			}
			// 热门源数据不带uid, 获取upid
			if dynmdl.IsPopularSv(req) {
				upids = dynmdl.GetArcUpids(dynList, arcm)
				attm = s.accDao.IsAttention(ctx, upids, mid)
			}
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			if thumbupm, err = s.dynDao.MultiStats(ctx, mid, thumBus); err != nil {
				log.Error("s.dynDao.MultiStats err(%+v)", err)
			}
			return nil
		})
	}
	if !dynmdl.IsPopularSv(req) && len(upids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			attm = s.accDao.IsAttention(ctx, upids, mid)
			return nil
		})
	}
	_ = eg.Wait()
	for k, dyn := range dynList.MixVideoItem {
		arc, ok := arcm[dyn.RID]
		if !ok || arc == nil || arc.Arc == nil {
			continue
		}
		if s.checkMidMaxInt32(c, dyn.UID, header) {
			continue
		}
		isAtten := attm[dyn.UID]
		var isLike int32
		if thumbupm != nil && thumbupm.Business != nil && thumbupm.Business[dynmdl.BusTypeVideo] != nil {
			if thumbState, ok := thumbupm.Business[dynmdl.BusTypeVideo].Records[arc.Arc.Aid]; ok && thumbState != nil && thumbState.LikeState == thumbApi.State_STATE_LIKE {
				isLike = 1
			}
		}
		svm := &dynmdl.SVideoMaterial{
			Arc:        arc,
			IsAtten:    isAtten,
			SVideoItem: &api.SVideoItem{CardType: model.CardTypeAv, DynIdStr: strconv.FormatInt(dyn.DynID, 10), Index: dyn.Index},
			IsLike:     isLike,
		}
		err := s.svConveyer(c, svm, header, req, s.FromSVideoAuthor, s.FromSVideoPlayer, s.FromSVideoDesc, s.FromSVideoStat)
		if err != nil {
			log.Error("s.svConveyer() failed. error(%+v)", err)
			continue
		}
		res.List = append(res.List, svm.SVideoItem)
		s.infoc(dynmdl.SVideoInfoc{
			AID:       dyn.RID,
			UpID:      dyn.UID,
			Buvid:     header.Buvid,
			MID:       mid,
			FromSpmid: req.FromSpmid,
			Follow:    isAtten,
			Like:      isLike,
			CardType:  model.CardTypeAv,
			CardIndex: int32(k),
			Offset:    req.Offset,
			OType:     dynmdl.SVideoTypeDynamic,
			OID:       dyn.DynID,
		})
	}
	return
}

// 数据处理
func (s *Service) svConveyer(c context.Context, svm *dynmdl.SVideoMaterial, header *dynmdl.Header, req *api.SVideoReq, f ...SvHandlerFunc) error {
	for _, v := range f {
		err := v(c, svm, header, req)
		if err != nil {
			log.Errorc(c, "svConveyer failed. svm(%+v), err(%+v)", svm, err)
			return err
		}
	}
	return nil
}

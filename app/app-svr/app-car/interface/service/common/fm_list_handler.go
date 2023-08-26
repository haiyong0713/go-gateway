package common

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	errGroup "go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
	"go-gateway/app/app-svr/app-car/interface/model/recommend"
	arcApi "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"

	"github.com/pkg/errors"
)

const (
	_historyPs     = 1200
	_historyArcBiz = "archive"
	_historyArcTp  = 3

	_sourceFM  = "car-audio"
	_channelFM = "sound"

	_upFirstPs = 5 // 首次查询up主稿件的ps
)

var (
	oidListHandlerMap map[fm_v2.FmType]FmListHandler // 处理接口map
	forceOidFirst     map[fm_v2.FmType]bool          // 是否强制冷启Oid在播单列表首位
)

func initFmListHandler(s *Service) {
	oidListHandlerMap = map[fm_v2.FmType]FmListHandler{
		//fm_v2.AudioHistory:  &FmListHistory{s: s}, // 最近播单已下线
		fm_v2.AudioFeed:     &FmListFeed{s: s},
		fm_v2.AudioVertical: &FmListVertical{s: s},
		fm_v2.AudioUp:       &FmListUp{s: s},
		fm_v2.AudioRelate:   &FmListRelate{s: s},
		fm_v2.AudioSeason:   &FmListSeason{s: s},
		fm_v2.AudioSeasonUp: &FmListSeasonUp{s: s},
	}
	forceOidFirst = map[fm_v2.FmType]bool{
		//fm_v2.AudioHistory:  true,
		fm_v2.AudioFeed:     true,
		fm_v2.AudioVertical: true,
	}
}

func FmListParamStrategy(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	if _, ok := oidListHandlerMap[param.FmType]; !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "FmType Illegal: %s", param.FmType)
	}
	return oidListHandlerMap[param.FmType].HandleParam(c, param)
}

// OidsWithUpwardHandler 支持向前分页查找的函数
type OidsWithUpwardHandler func(ctx context.Context, req fm_v2.OidsWithUpwardReq) (oids []int64, hasMore bool, err error)

type FmListHandler interface {
	// HandleParam 依据FM播单请求，生成各类播单项id列表（aid列表）
	HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error)
}

type FmListHistory struct {
	s *Service
}

type FmListFeed struct {
	s *Service
}

type FmListVertical struct {
	s *Service
}

type FmListUp struct {
	s *Service
}

type FmListRelate struct {
	s *Service
}

type FmListSeason struct {
	s *Service
}

type FmListSeasonUp struct {
	s *Service
}

func (f *FmListHistory) HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	var (
		pageReq     *fm_v2.PageReq
		hisResp     []*hisApi.ModelResource
		lastHis     *hisApi.ModelResource // 分页范围内，最后一条FM历史记录
		arcReq      *common.ArchiveReq    // 公共方法ugc请求体
		arcIds      []int64               // ugc稿件id
		resPageNext *fm_v2.PageInfo
		resHasNext  bool
	)

	pageReq, err = extractPageReq(param.FmType, param.PageNext, param.PagePre, param.Ps, 0, param.DeviceInfo)
	if err != nil {
		log.Errorc(c, "FmListHistory HandleParam extractPageReq error:%+v, param:%+v", err, param)
		return nil, err
	}

	if pageReq.NextEmpty {
		// 首次请求
		hisResp, err = f.s.historyDao.HistoryCursorV2(c, param.Mid, 0, 0, _historyPs, _historyArcBiz, param.Buvid, nil)
	} else {
		// 后续请求
		hisResp, err = f.s.historyDao.HistoryCursorV2(c, param.Mid, pageReq.PageNext.Max, pageReq.PageNext.ViewAt, _historyPs, _historyArcBiz, param.Buvid, nil)
	}
	if err != nil {
		log.Errorc(c, "FmListHistory HandleParam s.his.HistoryCursorV2 error:%+v, param:%+v", err, param)
		return nil, err
	}
	// 过滤出ps条fm播放历史
	arcIds = make([]int64, 0)
	for _, v := range hisResp {
		if len(arcIds) >= pageReq.PageSize {
			break
		}
		if v.Tp != _historyArcTp || v.Source != _sourceFM {
			continue
		}
		arcIds = append(arcIds, v.Oid)
		lastHis = v
	}
	if lastHis != nil {
		resPageNext = &fm_v2.PageInfo{
			Ps:       pageReq.PageSize,
			Business: lastHis.Business,
			ViewAt:   lastHis.Unix,
			Max:      lastHis.Kid,
		}
		resHasNext = true
	} else {
		resHasNext = false
	}
	// 冷启Oid，校验并填充
	arcIds = addBootOid(param, arcIds)
	arcReq = generateArcReq(arcIds, map[int64]int64{param.BootOid: param.BootCid})
	return &fm_v2.HandlerResp{
		PageResp: fm_v2.PageResp{
			PageNext: resPageNext,
			HasNext:  resHasNext,
		},
		OidParam: &common.Params{
			ArchiveReq: arcReq,
			UGCViewReq: &common.UGCViewReq{Aids: arcIds},
		},
		OidList: arcIds,
	}, nil
}

func (f *FmListFeed) HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	var (
		pageReq     *fm_v2.PageReq
		arcIds      []int64
		arcReq      *common.ArchiveReq
		resPageNext *fm_v2.PageInfo
	)
	pageReq, err = extractPageReq(param.FmType, param.PageNext, param.PagePre, param.Ps, 0, param.DeviceInfo)
	if err != nil {
		log.Errorc(c, "FmListFeed HandleParam extractPageReq error:%+v, param:%+v", err, param)
		return nil, err
	}

	arcIds, err = f.s.dynDao.RecommendArchives(c, param.Mid, param.Buvid, param.Build, param.MobiApp, param.Platform, param.Device, _channelFM)
	if err != nil {
		log.Errorc(c, "FmListFeed HandleParam s.dyn.RecommendArchives error:%+v, param:%+v", err, param)
		return nil, err
	}
	arcIds = addBootOid(param, arcIds)
	arcReq = generateArcReq(arcIds, map[int64]int64{param.BootOid: param.BootCid})
	resPageNext = &fm_v2.PageInfo{
		Ps: pageReq.PageSize,
	}
	return &fm_v2.HandlerResp{
		PageResp: fm_v2.PageResp{
			PageNext: resPageNext,
			HasNext:  true,
		},
		OidParam: &common.Params{
			ArchiveReq: arcReq,
			UGCViewReq: &common.UGCViewReq{Aids: arcIds},
		},
		OidList: arcIds,
	}, nil
}

func (f *FmListVertical) HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	var (
		pageReq     *fm_v2.PageReq
		arcIds      []int64
		arcReq      *common.ArchiveReq
		resPageNext *fm_v2.PageInfo
		hasNext     = true
	)
	pageReq, err = extractPageReq(param.FmType, param.PageNext, param.PagePre, param.Ps, 0, param.DeviceInfo)
	if err != nil {
		log.Errorc(c, "FmListVertical HandleParam extractPageReq error:%+v, param:%+v", err, param)
		return nil, err
	}
	if pageReq.NextEmpty {
		pageReq.PageNext = new(fm_v2.PageInfo)
	}
	arcIds, err = f.s.fmDao.ChannelFeed(c, param.Mid, param.FmId, int64(pageReq.PageSize), pageReq.PageNext.Pn, param.Buvid, param.DeviceInfo)
	if err != nil {
		log.Errorc(c, "FmListVertical HandleParam s.dyn.ChannelFeedRecommend error:%+v, param:%+v", err, param)
		return nil, err
	}
	if len(arcIds) == 0 {
		log.Warnc(c, "FmListVertical HandleParam get no arc, param:%+v, pageReq:%+v", param, pageReq)
		return &fm_v2.HandlerResp{
			PageResp: fm_v2.PageResp{HasNext: false},
			OidParam: new(common.Params),
			OidList:  make([]int64, 0),
		}, nil
	}
	if len(arcIds) < pageReq.PageSize {
		hasNext = false
	}
	arcIds = addBootOid(param, arcIds)
	arcReq = generateArcReq(arcIds, map[int64]int64{param.BootOid: param.BootCid})
	resPageNext = &fm_v2.PageInfo{
		Ps: pageReq.PageSize,
		Pn: pageReq.PageNext.Pn + 1,
	}
	return &fm_v2.HandlerResp{
		PageResp: fm_v2.PageResp{
			PageNext: resPageNext,
			HasNext:  hasNext,
		},
		OidParam: &common.Params{
			ArchiveReq: arcReq,
			UGCViewReq: &common.UGCViewReq{Aids: arcIds},
		},
		OidList: arcIds,
	}, nil
}

func (f *FmListUp) HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	var (
		pageReq  *fm_v2.PageReq
		pageResp *fm_v2.PageResp
		arcIds   []int64
		arcReq   *common.ArchiveReq
	)

	pageReq, err = extractPageReq(param.FmType, param.PageNext, param.PagePre, param.Ps, 0, param.DeviceInfo)
	if err != nil {
		log.Errorc(c, "FmListUp HandleParam extractPageReq error:%+v, param:%+v", err, param)
		return nil, err
	}
	pageResp, arcIds, err = f.s.queryOidsWithUpward(c, param, pageReq, f.s.upDao.UpArcsWithUpward)
	if err != nil {
		log.Errorc(c, "FmListUp HandleParam queryOidsWithUpward error:%+v, param:%+v, pageReq:%+v", err, param, pageReq)
		return nil, err
	}
	arcReq = generateArcReq(arcIds, map[int64]int64{param.BootOid: param.BootCid})
	return &fm_v2.HandlerResp{
		PageResp: fm_v2.PageResp{
			PageNext:    pageResp.PageNext,
			PagePre:     pageResp.PagePre,
			HasNext:     pageResp.HasNext,
			HasPrevious: pageResp.HasPrevious,
		},
		OidParam: &common.Params{
			ArchiveReq: arcReq,
			UGCViewReq: &common.UGCViewReq{Aids: arcIds},
		},
		OidList: arcIds,
	}, nil
}

func (f *FmListRelate) HandleParam(c context.Context, param *fm_v2.FmListParam) (*fm_v2.HandlerResp, error) {
	var (
		arcIds  = make([]int64, 0)
		arcReq  *common.ArchiveReq
		recResp []*recommend.Item
		err     error
	)

	recResp, err = f.s.rcmdDao.Relate(c, param.Mid, param.FmId, param.Buvid)
	if err != nil {
		log.Errorc(c, "FmListRelate HandleParam s.rec.Relate error:%+v, param:%+v", err, param)
		return nil, err
	}
	if len(recResp) == 0 {
		log.Errorc(c, "FmListRelate HandleParam get no arc, param:%+v", param)
		return nil, ecode.NothingFound
	}
	for _, v := range recResp {
		if v.Goto != model.GotoAv {
			continue
		}
		arcIds = append(arcIds, v.ID)
	}
	arcIds = addBootOid(param, arcIds)
	arcReq = generateArcReq(arcIds, map[int64]int64{param.BootOid: param.BootCid})
	return &fm_v2.HandlerResp{
		OidParam: &common.Params{
			ArchiveReq: arcReq,
			UGCViewReq: &common.UGCViewReq{Aids: arcIds},
		},
		OidList: arcIds,
	}, nil
}

func (f *FmListSeason) HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	return f.s.HandleSeasonParam(c, param)
}

func (f *FmListSeasonUp) HandleParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	return f.s.HandleSeasonParam(c, param)
}

func (s *Service) HandleSeasonParam(c context.Context, param *fm_v2.FmListParam) (resp *fm_v2.HandlerResp, err error) {
	var (
		pageReq  *fm_v2.PageReq
		pageResp *fm_v2.PageResp
		arcIds   []int64
		arcReq   *common.ArchiveReq
	)
	pageReq, err = extractPageReq(param.FmType, param.PageNext, param.PagePre, param.Ps, 0, param.DeviceInfo)
	if err != nil {
		log.Errorc(c, "HandleSeasonParam extractPageReq error:%+v, param:%+v", err, param)
		return nil, err
	}
	pageResp, arcIds, err = s.queryOidsWithUpward(c, param, pageReq, s.getFmSeasonOid)
	if err != nil {
		log.Errorc(c, "HandleSeasonParam queryOidsWithUpward error:%+v, param:%+v, pageReq:%+v", err, param, pageReq)
		return nil, err
	}
	arcReq = generateArcReq(arcIds, map[int64]int64{param.BootOid: param.BootCid})
	return &fm_v2.HandlerResp{
		PageResp: fm_v2.PageResp{
			PageNext:    pageResp.PageNext,
			PagePre:     pageResp.PagePre,
			HasNext:     pageResp.HasNext,
			HasPrevious: pageResp.HasPrevious,
		},
		OidParam: &common.Params{
			ArchiveReq: arcReq,
			UGCViewReq: &common.UGCViewReq{Aids: arcIds},
		},
		OidList: arcIds,
	}, nil
}

func (s *Service) getFmSeasonOid(ctx context.Context, req fm_v2.OidsWithUpwardReq) (oids []int64, hasMore bool, err error) {
	return s.fmDao.GetSeasonOid(ctx, fm_v2.SeasonOidReq{
		Scene:       fm_v2.SceneFm,
		FmType:      req.FmType,
		SeasonId:    req.FmId,
		Cursor:      req.Cursor,
		Upward:      req.Upward,
		WithCurrent: req.WithCurrent,
		Ps:          req.Ps,
	})
}

// queryOidsWithUpward 支持向上分页的稿件查找方法，并返回下次请求的分页数据
func (s *Service) queryOidsWithUpward(c context.Context, param *fm_v2.FmListParam, pageReq *fm_v2.PageReq,
	handler OidsWithUpwardHandler) (resp *fm_v2.PageResp, aids []int64, err error) {
	var (
		preAids     []int64
		nextAids    []int64
		hasPre      bool
		hasNext     bool
		resPageNext *fm_v2.PageInfo
		resPagePre  *fm_v2.PageInfo
		arcIds      = make([]int64, 0)
	)
	// 首次请求，page参数均为空
	if pageReq.NextEmpty && pageReq.PreEmpty {
		if param.BootOid > 0 {
			// 前5个 + 后6个（加上了当前）
			eg := errGroup.WithContext(c)
			eg.Go(func(c context.Context) error {
				var (
					localErr error
					req      = fm_v2.OidsWithUpwardReq{
						DeviceInfo:  param.DeviceInfo,
						FmType:      param.FmType,
						FmId:        param.FmId,
						Cursor:      param.BootOid,
						Upward:      true,
						WithCurrent: false,
						Ps:          _upFirstPs,
					}
				)
				preAids, hasPre, localErr = handler(c, req)
				if localErr != nil {
					return errors.Wrap(localErr, "s.up.UpArcsWithUpward query pre error")
				}
				return nil
			})
			eg.Go(func(c context.Context) error {
				var (
					localErr error
					req      = fm_v2.OidsWithUpwardReq{
						DeviceInfo:  param.DeviceInfo,
						FmType:      param.FmType,
						FmId:        param.FmId,
						Cursor:      param.BootOid,
						Upward:      false,
						WithCurrent: true,
						Ps:          _upFirstPs + 1,
					}
				)
				nextAids, hasNext, localErr = handler(c, req)
				if localErr != nil {
					return errors.Wrap(localErr, "s.up.UpArcsWithUpward query next error")
				}
				return nil
			})
			if err = eg.Wait(); err != nil {
				return nil, nil, errors.Wrap(err, "first query with oid error")
			}
		} else {
			// 从头开始pageSize个
			req := fm_v2.OidsWithUpwardReq{
				DeviceInfo:  param.DeviceInfo,
				FmType:      param.FmType,
				FmId:        param.FmId,
				Upward:      false,
				WithCurrent: true,
				Ps:          pageReq.PageSize,
			}
			nextAids, hasNext, err = handler(c, req)
			if err != nil {
				return nil, nil, errors.Wrap(err, "first query no oid error")
			}
		}
	} else if !pageReq.NextEmpty {
		// 向下翻页
		req := fm_v2.OidsWithUpwardReq{
			DeviceInfo:  param.DeviceInfo,
			FmType:      param.FmType,
			FmId:        param.FmId,
			Cursor:      pageReq.PageNext.Oid,
			Upward:      false,
			WithCurrent: false,
			Ps:          pageReq.PageSize,
		}
		nextAids, hasNext, err = handler(c, req)
		if err != nil {
			return nil, nil, errors.Wrap(err, "query next error")
		}
	} else if !pageReq.PreEmpty {
		// 向上翻页
		req := fm_v2.OidsWithUpwardReq{
			DeviceInfo:  param.DeviceInfo,
			FmType:      param.FmType,
			FmId:        param.FmId,
			Cursor:      pageReq.PagePre.Oid,
			Upward:      true,
			WithCurrent: false,
			Ps:          pageReq.PageSize,
		}
		preAids, hasPre, err = handler(c, req)
		if err != nil {
			return nil, nil, errors.Wrap(err, "query pre error")
		}
	} else {
		return nil, nil, errors.Wrap(ecode.RequestErr, "both `page_next` and `page_previous` params exist!")
	}
	// 生成返回的pageInfo
	if hasNext && len(nextAids) > 0 {
		resPageNext = &fm_v2.PageInfo{
			Oid: nextAids[len(nextAids)-1], // 取最后一个视频作为游标
			Ps:  pageReq.PageSize,
		}
	}
	if hasPre && len(preAids) > 0 {
		resPagePre = &fm_v2.PageInfo{
			Oid: preAids[0], // 取首个视频作为游标
			Ps:  pageReq.PageSize,
		}
	}
	// 合并前序和后序的aid
	if len(preAids) > 0 {
		arcIds = append(arcIds, preAids...)
	}
	if len(nextAids) > 0 {
		arcIds = append(arcIds, nextAids...)
	}
	return &fm_v2.PageResp{
		PageNext:    resPageNext,
		PagePre:     resPagePre,
		HasNext:     hasNext,
		HasPrevious: hasPre,
	}, arcIds, nil
}

// addBootOid 冷启Oid，校验并注入
func addBootOid(param *fm_v2.FmListParam, arcIds []int64) []int64 {
	if param.BootOid <= 0 || param.BootCid <= 0 || len(arcIds) == 0 {
		return arcIds
	}
	// 冷启oid的index
	index := -1
	for i, v := range arcIds {
		if v == param.BootOid {
			index = i
			break
		}
	}
	if index == -1 {
		log.Warn("boot oid(%d) not exist in resp, req:%+v", param.BootOid, param)
		arcIds = append([]int64{param.BootOid}, arcIds...)
	} else if forceOidFirst[param.FmType] && arcIds[0] != param.BootOid {
		log.Warn("boot oid(%d) is not first in resp, req:%+v", param.BootOid, param)
		tmp := arcIds[0]
		arcIds[0] = arcIds[index]
		arcIds[index] = tmp
	}
	return arcIds
}

// generateArcReq 生成公共方法请求，可指定秒开的cid
func generateArcReq(arcIds []int64, cidMap map[int64]int64) *common.ArchiveReq {
	avs := make([]*arcApi.PlayAv, 0)
	for _, aid := range arcIds {
		av := &arcApi.PlayAv{Aid: aid}
		if cid, ok := cidMap[aid]; ok {
			av.PlayVideos = []*arcApi.PlayVideo{{Cid: cid}}
		}
		avs = append(avs, av)
	}
	return &common.ArchiveReq{PlayAvs: avs}
}

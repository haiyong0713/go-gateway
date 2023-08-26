// Package common
// FM和视频的合集业务
package common

import (
	"context"
	"go-common/library/log"
	"sync"

	"go-common/library/sync/errgroup.v2"
	comm "go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"

	"github.com/pkg/errors"
)

func (s *Service) serialInfoIntegrate(ctx context.Context, req comm.SerialInfosReq) (*comm.SerialInfosResp, error) {
	var (
		res = new(comm.SerialInfosResp)
		err error
	)
	eg := errgroup.WithContext(ctx)
	if len(req.FmCommonIds) > 0 {
		var lock sync.Mutex
		res.FmCommon = make(map[int64]*comm.SerialInfo)
		for _, id := range req.FmCommonIds {
			tmpReq := fm_v2.SeasonInfoReq{Scene: fm_v2.SceneFm, FmType: fm_v2.AudioSeason, SeasonId: id}
			eg.Go(func(ctx context.Context) error {
				resp, localErr := s.fmDao.GetSeasonInfo(ctx, tmpReq)
				if localErr != nil {
					return errors.Wrap(localErr, "s.fmDao.GetSeasonInfo FmCommon error")
				}
				count, localErr := s.fmDao.GetSeasonOidCount(ctx, tmpReq)
				if localErr != nil {
					return errors.Wrap(localErr, "s.fmDao.GetSeasonOidCount FmCommon error")
				}
				lock.Lock()
				res.FmCommon[tmpReq.SeasonId] = resp.ToSerialInfo()
				if res.FmCommon[tmpReq.SeasonId] != nil {
					res.FmCommon[tmpReq.SeasonId].Count = count
				}
				lock.Unlock()
				return nil
			})
		}
	}
	if len(req.VideoIds) > 0 {
		var lock sync.Mutex
		res.Video = make(map[int64]*comm.SerialInfo)
		for _, id := range req.VideoIds {
			tmpReq := fm_v2.SeasonInfoReq{Scene: fm_v2.SceneVideo, SeasonId: id}
			eg.Go(func(ctx context.Context) error {
				resp, localErr := s.fmDao.GetSeasonInfo(ctx, tmpReq)
				if localErr != nil {
					return errors.Wrap(localErr, "s.fmDao.GetSeasonInfo Video error")
				}
				count, localErr := s.fmDao.GetSeasonOidCount(ctx, tmpReq)
				if localErr != nil {
					return errors.Wrap(localErr, "s.fmDao.GetSeasonOidCount FmCommon error")
				}
				lock.Lock()
				res.Video[tmpReq.SeasonId] = resp.ToSerialInfo()
				if res.Video[tmpReq.SeasonId] != nil {
					res.Video[tmpReq.SeasonId].Count = count
				}
				lock.Unlock()
				return nil
			})
		}
	}
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "serialInfoIntegrate err:%+v, req:%+v", err, req)
		return nil, err
	}
	return res, nil
}

func (s *Service) serialOidsByPageIntegrate(ctx context.Context, req comm.SerialArcsReq) (*comm.SerialArcsResp, error) {
	var (
		resp = new(comm.SerialArcsResp)
		err  error
	)
	eg := errgroup.WithContext(ctx)
	if len(req.FmCommon) > 0 {
		resp.FmCommon = make(map[int64]*comm.SerialArcs)
		var lock sync.Mutex
		for _, _tmp := range req.FmCommon {
			tmpReq := _tmp
			eg.Go(func(ctx context.Context) error {
				arcs, localErr := s.serialOidsByPage(ctx, fm_v2.SceneFm, fm_v2.AudioSeason, *tmpReq)
				if localErr != nil {
					return errors.Wrap(err, "s.serialOidsByPage FmCommon err")
				}
				// 社区风险内容过滤
				arcs.Aids = s.SixLimitFilter(ctx, arcs.Aids)
				lock.Lock()
				resp.FmCommon[tmpReq.SerialId] = arcs
				lock.Unlock()
				return nil
			})
		}
	}
	if len(req.Video) > 0 {
		resp.Video = make(map[int64]*comm.SerialArcs)
		var lock sync.Mutex
		for _, _tmp := range req.Video {
			tmpReq := _tmp
			eg.Go(func(ctx context.Context) error {
				arcs, localErr := s.serialOidsByPage(ctx, fm_v2.SceneVideo, "", *tmpReq)
				if localErr != nil {
					return errors.Wrap(err, "s.serialOidsByPage Video err")
				}
				// 社区风险内容过滤
				arcs.Aids = s.SixLimitFilter(ctx, arcs.Aids)
				lock.Lock()
				resp.Video[tmpReq.SerialId] = arcs
				lock.Unlock()
				return nil
			})
		}
	}
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "serialOidsByPageIntegrate err:%+v, req:%+v", err, req)
		return nil, err
	}
	return resp, nil
}

// serialOidsByPage 分页查找合集内的稿件，相比 queryOidsWithUpward，支持同时向上和向下分页查找
func (s *Service) serialOidsByPage(ctx context.Context, scene fm_v2.Scene, fmType fm_v2.FmType, req comm.SerialArcReq) (*comm.SerialArcs, error) {
	var (
		preAids     []int64
		nextAids    []int64
		hasPre      bool
		hasNext     bool
		resPageNext *comm.SerialPageInfo
		resPagePre  *comm.SerialPageInfo
		arcIds      = make([]int64, 0)
		err         error
	)
	// 首次请求，page参数均为空
	if req.PageNext == nil && req.PagePre == nil {
		daoReq := fm_v2.SeasonOidReq{
			Scene:       scene,
			FmType:      fmType,
			SeasonId:    req.SerialId,
			Upward:      false,
			WithCurrent: true,
			Ps:          extractSerialPs(req.SerialPageReq, false, false),
		}
		nextAids, hasNext, err = s.fmDao.GetSeasonOid(ctx, daoReq)
		if err != nil {
			return nil, errors.Wrap(err, "first query no oid error")
		}
	}
	eg := errgroup.WithContext(ctx)
	// 向下翻页
	if req.PageNext != nil {
		eg.Go(func(ctx context.Context) error {
			daoReq := fm_v2.SeasonOidReq{
				Scene:       scene,
				FmType:      fmType,
				SeasonId:    req.SerialId,
				Cursor:      req.PageNext.Oid,
				Upward:      false,
				WithCurrent: req.PageNext.WithCurrent,
				Ps:          extractSerialPs(req.SerialPageReq, true, false),
			}
			nextAids, hasNext, err = s.fmDao.GetSeasonOid(ctx, daoReq)
			if err != nil {
				return errors.Wrap(err, "next query error")
			}
			return nil
		})
	}
	// 向上翻页
	if req.PagePre != nil {
		eg.Go(func(ctx context.Context) error {
			daoReq := fm_v2.SeasonOidReq{
				Scene:       scene,
				FmType:      fmType,
				SeasonId:    req.SerialId,
				Cursor:      req.PagePre.Oid,
				Upward:      true,
				WithCurrent: req.PagePre.WithCurrent,
				Ps:          extractSerialPs(req.SerialPageReq, false, true),
			}
			preAids, hasPre, err = s.fmDao.GetSeasonOid(ctx, daoReq)
			if err != nil {
				return errors.Wrap(err, "pre query error")
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	// 生成返回的pageInfo
	if hasNext && len(nextAids) > 0 {
		resPageNext = &comm.SerialPageInfo{
			Oid: nextAids[len(nextAids)-1], // 取最后一个视频作为游标
			Ps:  extractSerialPs(req.SerialPageReq, true, false),
		}
	}
	if hasPre && len(preAids) > 0 {
		resPagePre = &comm.SerialPageInfo{
			Oid: preAids[0], // 取首个视频作为游标
			Ps:  extractSerialPs(req.SerialPageReq, false, true),
		}
	}
	// 合并前序和后序的aid
	if len(preAids) > 0 {
		arcIds = append(arcIds, preAids...)
	}
	if len(nextAids) > 0 {
		arcIds = append(arcIds, nextAids...)
	}
	return &comm.SerialArcs{
		Aids: arcIds,
		SerialPageResp: comm.SerialPageResp{
			PageNext:    resPageNext,
			PagePre:     resPagePre,
			HasNext:     hasNext,
			HasPrevious: hasPre,
		},
	}, nil
}

// extractSerialPs 提取合集分页大小
func extractSerialPs(req comm.SerialPageReq, next bool, pre bool) int {
	if req.Ps > 0 {
		return req.Ps
	}
	if next && req.PageNext != nil && req.PageNext.Ps > 0 {
		return req.PageNext.Ps
	}
	if pre && req.PagePre != nil && req.PagePre.Ps > 0 {
		return req.PagePre.Ps
	}
	return _defaultPs
}

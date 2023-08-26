package service

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/native-page/job/internal/model"

	"go-common/library/log"
)

func (s *Service) broadcastProgress(c context.Context) {
	progressParams, err := s.dao.GetProgressParams(c)
	log.Info("Start to broadcastProgress, get %d to process", len(progressParams))
	if err != nil {
		return
	}
	s.setParentIDToProgressParams(progressParams)
	_ = s.pushFromProgressParams(c, progressParams)
}

func (s *Service) broadcastClickProgress(c context.Context) {
	progressParams, err := s.dao.GetProgressParamsFromClick(c)
	log.Info("Start to broadcastClickProgress, get %d to process", len(progressParams))
	if err != nil {
		return
	}
	s.setParentIDToProgressParams(progressParams)
	_ = s.pushFromProgressParams(c, progressParams)
}

func (s *Service) pushFromProgressParams(c context.Context, progressParams []*model.ProgressParam) error {
	rlys, err := s.dao.BatchActivityProgress(c, progressParams, false)
	if err != nil {
		return err
	}
	for _, v := range progressParams {
		rly, ok := rlys[v.Sid]
		if !ok || len(rly.Groups) == 0 {
			continue
		}
		group, ok := rly.Groups[v.GroupID]
		if !ok || group.Info == nil {
			continue
		}
		dimension := model.ProgressDimension(group.Info.Dim1)
		if !dimension.IsTotal() {
			continue
		}
		// 推送维度
		v.Dimension = group.Info.Dim1
		progress := group.Total
		s.progCacheMu.Lock()
		if cache, ok := s.progressCache[buildCacheKey(v)]; ok && cache == progress {
			s.progCacheMu.Unlock()
			continue
		}
		s.progressCache[buildCacheKey(v)] = progress
		s.progCacheMu.Unlock()
		param := v
		_ = s.progressWorker.Do(c, func(ctx context.Context) {
			_, _ = s.dao.PushProgress(ctx, param, progress, nil, dimension)
		})
	}
	return nil
}

func (s *Service) setParentIDToProgressParams(params []*model.ProgressParam) {
	for _, v := range params {
		parentID, ok := s.dao.GetParentPageID(v.PageID)
		if !ok {
			continue
		}
		v.PageID = parentID
	}
}

func buildCacheKey(param *model.ProgressParam) string {
	return fmt.Sprintf("%s_%d_%d", param.Type, param.Sid, param.GroupID)
}

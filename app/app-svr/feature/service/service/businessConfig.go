package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/xstr"

	featureAdminMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	feedadminMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	"go-gateway/app/app-svr/feature/service/api"
	businessConfigMdl "go-gateway/app/app-svr/feature/service/model/businessConfig"
)

func (s *Service) BusinessConfig(c context.Context, req *api.BusinessConfigReq) (*api.BusinessConfigReply, error) {
	var res = new(api.BusinessConfigReply)
	var tmps = make(map[string]*businessConfigMdl.BusinessConfig)
	for k, v := range s.businessConfigCache[req.TreeId] {
		tmps[k] = v
	}
	for k, v := range s.businessConfigCache[featureAdminMdl.Common] {
		tmps[k] = v
	}
	res.BusinessConfigs = make(map[string]*api.BusinessConfig)
	for _, tmp := range tmps {
		if tmp == nil {
			continue
		}
		var (
			RelationTreeIDs []int64
			errTmp          error
			isRelation      bool
		)
		if tmp.TreeID == feedadminMdl.Common {
			if RelationTreeIDs, errTmp = xstr.SplitInts(tmp.Relations); errTmp != nil {
				log.Error("%+v", errTmp)
				continue // 避免错误数据影响整体
			}
			for _, RelationTreeID := range RelationTreeIDs {
				if RelationTreeID == req.TreeId {
					isRelation = true
				}
			}
			if !isRelation {
				continue
			}
		}
		res.BusinessConfigs[tmp.KeyName] = &api.BusinessConfig{
			Id:          tmp.ID,
			TreeId:      req.TreeId,
			KeyName:     tmp.KeyName,
			Config:      tmp.Config,
			Description: tmp.Description,
			Relations:   RelationTreeIDs,
		}
	}
	return res, nil
}

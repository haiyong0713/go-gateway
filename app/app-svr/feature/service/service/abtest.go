package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-common/library/xstr"

	featureAdminMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	feedadminMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	"go-gateway/app/app-svr/feature/service/api"
	abtestMdl "go-gateway/app/app-svr/feature/service/model/abtest"
)

func (s *Service) ABTest(c context.Context, req *api.ABTestReq) (res *api.ABTestReply, err error) {
	res = new(api.ABTestReply)
	var tmps = make(map[string]*abtestMdl.ABTest)
	for k, v := range s.abtestCache[req.TreeId] {
		tmps[k] = v
	}
	for k, v := range s.abtestCache[featureAdminMdl.Common] {
		tmps[k] = v
	}
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
		var config []*api.ExpConfig
		if errTmp = json.Unmarshal([]byte(tmp.Config), &config); errTmp != nil {
			log.Error("%+v", errTmp)
			continue // 避免错误数据影响整体
		}
		res.AbtestItems = append(res.AbtestItems, &api.ABTestItem{
			Id:      tmp.ID,
			TreeId:  req.TreeId,
			KeyName: tmp.KeyName,
			Config:  config,
		})
	}
	return res, nil
}

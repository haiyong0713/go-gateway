package service

import (
	"context"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/api-gateway/api-manager/internal/model"
)

func (s *Service) DiscoveryList(_ context.Context) (res []string) {
	log.Info("all dis len dis:%v proto:%v", len(s.allDis), len(s.allProtos))
	res = s.allDis
	return
}

func (s *Service) AppList(ctx context.Context, key, discoveryID string, tp int8, pn, ps int64) (res *model.ApiList, err error) {
	res = &model.ApiList{
		Infos: make([]*model.ApiRawInfo, 0),
	}
	resTmp := make([]*model.ApiRawInfo, 0)
	switch tp {
	case model.ApiTypeGrpc:
		resTmp, err = s.dao.GetGrpcApis(ctx, discoveryID)
	case model.ApiTypeHttp:
		resTmp, err = s.dao.GetHttpApis(ctx)
	default:
		err = ecode.RequestErr
	}
	if err != nil || len(resTmp) == 0 {
		return nil, err
	}
	key = strings.ToLower(key)
	if key != "" {
		for _, r := range resTmp {
			if strings.Contains(strings.ToLower(r.ApiPath), key) || strings.Contains(strings.ToLower(r.DiscoveryID), key) {
				res.Infos = append(res.Infos, r)
			}
		}
	} else {
		res.Infos = resTmp
	}
	res.Count = int64(len(res.Infos))
	if pn > 0 && ps > 0 {
		start := (pn - 1) * ps
		end := start + ps - 1
		switch {
		case res.Count > start && res.Count > end:
			res.Infos = res.Infos[start : end+1]
		case res.Count > start && res.Count <= end:
			res.Infos = res.Infos[start:]
		default:
			res.Infos = make([]*model.ApiRawInfo, 0)
		}
	}
	return
}

func (s *Service) AddApi(ctx context.Context, req *model.AddApiReq) (err error) {
	return s.dao.AddApi(ctx, req.ToRawInfo())
}

func (s *Service) EditApi(ctx context.Context, req *model.AddApiReq) (err error) {
	return s.dao.AddApi(ctx, req.ToRawInfo())
}

func (s *Service) DelApi(ctx context.Context, id int64) (err error) {
	return s.dao.UpApi(ctx, id)
}

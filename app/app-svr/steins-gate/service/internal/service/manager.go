package service

import (
	"context"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// ManagerList .
func (s *Service) ManagerList(c context.Context, aid int64) (data *model.ManagerGraph, err error) {
	if env.DeployEnv == env.DeployEnvProd {
		err = ecode.AccessDenied
		return
	}
	return s.dao.ManagerList(c, aid)
}

// RecentArcs .
func (s *Service) RecentArcs(c context.Context, param *model.RecentArcReq) (data *model.RecentArcs, err error) {
	if env.DeployEnv == env.DeployEnvProd {
		err = ecode.AccessDenied
		return
	}
	var (
		aids    []int64
		aidsMap map[int64]*arcgrpc.Arc
	)
	if aids, err = s.dao.RecentArcs(c, param); err != nil {
		log.Error("RecentArcs Err %v", err)
		return
	}
	if aidsMap, err = s.arcDao.Arcs(c, aids); err != nil {
		log.Error("RecentArcs Aids %v, Err %v", aids, err)
		return
	}
	data = new(model.RecentArcs)
	for _, v := range aids {
		if arc, ok := aidsMap[v]; ok {
			data.Arcs = append(data.Arcs, arc)
		}
	}
	return

}

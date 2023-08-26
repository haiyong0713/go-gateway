package note

import (
	"context"
	"go-common/library/ecode"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/thoas/go-funk"
)

func (s *Service) ArcsForbid(c context.Context, req *api.ArcsForbidReq) (*api.ArcsForbidReply, error) {
	if !s.ArcsForbidAllower.Allow() {
		err := ecode.Error(ecode.LimitExceed, "ArcsForbid sent too fast, please try again later")
		return nil, err
	}
	arcs, err := s.dao.SimpleArcs(c, req.Aids)
	if err != nil {
		log.Warn("noteWarn ArcsForbid err(%+v)", err)
		return nil, err
	}
	res := make(map[int64]bool)
	for _, arc := range arcs {
		if funk.ContainsInt64(s.c.NoteCfg.ForbidCfg.ForbidTypeIds, int64(arc.TypeId)) {
			res[arc.Aid] = true
			continue
		}
		if _, ok := s.politicsUpMap[arc.Mid]; ok {
			res[arc.Aid] = true
		}
	}
	return &api.ArcsForbidReply{Items: res}, nil
}

func (s *Service) loadFeaPolitics() {
	list, err := s.dao.FeatureContList(context.Background())
	if err != nil {
		log.Error("noteError err(%+v)", err)
		return
	}
	pMap := make(map[int64]struct{})
	for _, l := range list {
		if l == nil {
			continue
		}
		mid, e := strconv.ParseInt(l.Oid, 10, 64)
		if e != nil {
			log.Warn("noteInfo loadFeaPolitics l(%+v) invalid,skip", l)
			continue
		}
		pMap[mid] = struct{}{}
	}
	s.politicsUpMap = pMap
}

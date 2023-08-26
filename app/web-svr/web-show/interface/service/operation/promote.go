package operation

import (
	"context"
	"go-gateway/app/web-svr/web-show/interface/model"
	"regexp"
	"strconv"

	"go-common/library/log"

	arcErr "go-gateway/app/app-svr/archive/ecode"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	opmdl "go-gateway/app/web-svr/web-show/interface/model/operation"
)

var (
	_emptyPromoteMap = make(map[string][]*opmdl.Promote)
	_avReg           = regexp.MustCompile(`video\/av[0-9]+`)
)

// Promote Service
func (s *Service) Promote(c context.Context, arg *opmdl.ArgPromote) (res map[string][]*opmdl.Promote, err error) {
	var (
		ok   bool
		arcs map[int64]*api.Arc
		arc  *api.Arc
		aid  int64
		aids []int64
	)
	opMap := s.operation(arg.Tp, arg.Rank, arg.Count)
	for _, ops := range opMap {
		for _, op := range ops {
			if aid, err = s.regAid(op.Link); err != nil {
				log.Error("service.regAid error(%v)", err)
				continue
			}
			op.Aid = aid
			aids = append(aids, aid)
		}
	}
	var (
		args    = &arcgrpc.ArcsRequest{Aids: aids}
		arcsTmp *arcgrpc.ArcsReply
	)
	if arcsTmp, err = s.arcGRPC.Arcs(c, args); err != nil {
		log.Error("%v", err)
		res = _emptyPromoteMap
		return
	}
	arcs = arcsTmp.GetArcs()
	res = make(map[string][]*opmdl.Promote)
	for rk, ops := range opMap {
		promotes := make([]*opmdl.Promote, 0, len(ops))
		for _, op := range ops {
			if arc, ok = arcs[op.Aid]; !ok {
				continue
			}
			model.ClearAttrAndAccess(arc)
			promote := &opmdl.Promote{
				IsAd:    int8(op.Ads),
				Archive: arc,
			}
			promotes = append(promotes, promote)
		}
		res[rk] = promotes
	}
	return
}

// regAid Service
func (s *Service) regAid(link string) (aid int64, err error) {
	avStr := _avReg.FindString(link)
	if avStr != "" {
		aidStr := avStr[8:]
		if aid, err = strconv.ParseInt(aidStr, 10, 64); err != nil {
			log.Error("strconv.ParseInt error(%v)", err)
			return
		}
	} else {
		err = arcErr.ArchiveNotExist
	}
	return
}

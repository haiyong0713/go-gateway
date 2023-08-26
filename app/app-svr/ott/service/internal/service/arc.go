package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/ott/service/api"
	"go-gateway/app/app-svr/ott/service/internal/model"

	"go-common/library/sync/errgroup.v2"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
)

func (s *Service) ArcsAllow(ctx context.Context, aids []int64) (res *pb.ArcsAllowReply, err error) {
	var (
		arcs     map[int64]*arcmdl.Arc
		smallMap = make(map[int64]bool)
	)
	res = &pb.ArcsAllowReply{
		Items: make(map[int64]bool),
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) (err error) {
		if arcs, err = s.dao.Arcs(ctx, aids); err != nil {
			return
		}
		return nil
	})
	eg.Go(func(c context.Context) (err error) {
		if smallMap, err = s.dao.SimpleArchives(ctx, aids); err != nil {
			return
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait() error(%+v)", err)
		return
	}
	for _, v := range aids {
		if isSmall, ok := smallMap[v]; ok && isSmall {
			res.Items[v] = false
			continue
		}
		if rpcArc, ok := arcs[v]; ok {
			if valid := s.CheckArc(ctx, rpcArc); !valid {
				res.Items[v] = false
				continue
			}
			res.Items[v] = true
			continue
		}
		res.Items[v] = false
	}
	return
}

func (s *Service) CheckArc(ctx context.Context, arc *arcmdl.Arc) (ok bool) {
	ok = true
	arcAllow := &model.ArcAllow{}
	arcAllow.FromArcReply(arc)
	if !s.arcAllowImport(arcAllow) {
		log.Warn("wrapSyncLic cAid %d Can't play", arc.Aid)
		ok = false
		return
	}
	return
}

func (s *Service) arcAllowImport(arc *model.ArcAllow) (allowed bool) {
	if !arc.CanPlay() {
		log.Warn("arcAllowImport Aid %d Not allowed Due to State %d", arc.Aid, arc.State)
		return
	}
	if arc.Ugcpay == arcmdl.AttrYes {
		log.Warn("arcAllowImport Aid %d Not allowed Due to Ugcpay %d", arc.Aid, arc.Ugcpay)
		return
	}
	if s.hitPGC(arc.Typeid) {
		log.Warn("arcAllowImport Aid %d Not allowed Due to HitPGC %d", arc.Aid, arc.Typeid)
		return
	}
	if arc.AttrIsPgc == arcmdl.AttrYes {
		log.Warn("arcAllowImport Aid %d Not allowed have pgc info", arc.Aid)
		return
	}
	if arc.AttrIsPugv == arcmdl.AttrYes {
		log.Warn("arcAllowImport Aid %d Not allowed have pugv info", arc.Aid)
		return
	}
	if arc.AttrIsStein == arcmdl.AttrYes {
		log.Warn("arcAllowImport Aid %d Not allowed have stein info", arc.Aid)
		return
	}
	if !arc.IsOrigin() {
		log.Warn("arcAllowImport Aid %d Not allowed Due to Not Origin copyright(%d)", arc.Aid, arc.Copyright)
	}
	allowed = true
	return
}

func (s *Service) hitPGC(tid int32) (hit bool) {
	_, hit = s.pgcTypes[s.getPTypeName(tid)]
	return
}

func (s *Service) getPTypeName(typeID int32) (name string) {
	var (
		second, first *arcmdl.Tp
		ok            bool
	)
	if second, ok = s.ArcTypes[typeID]; !ok {
		log.Error("can't find type for ID: %d ", typeID)
		return
	}
	if first, ok = s.ArcTypes[second.Pid]; !ok {
		log.Error("can't find type for ID: %d, second Info: %v", second.ID, second.Pid)
		return
	}
	return first.Name
}

func (s *Service) loadTypes() {
	var (
		res       map[int32]*arcmdl.Tp
		resRel    = make(map[int32][]int32)
		typeReply *arcmdl.TypesReply
		err       error
	)
	if typeReply, err = s.dao.ArcType(context.Background()); err != nil {
		log.Error("arcRPC loadType Error %v", err)
		return
	}
	if typeReply == nil || len(typeReply.Types) == 0 {
		err = ecode.NothingFound
		log.Error("arcRPC loadType Empty")
		return
	}
	res = typeReply.Types
	for _, tInfo := range res {
		if _, ok := resRel[tInfo.Pid]; !ok {
			resRel[tInfo.Pid] = []int32{tInfo.ID}
			continue
		}
		resRel[tInfo.Pid] = append(resRel[tInfo.Pid], tInfo.ID)
	}
	s.ArcTypes = res
	return
}

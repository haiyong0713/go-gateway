package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	xecode "go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

// TagAids gets avids by tagID from bigdata or backup cache,
// and updates the cache after getting bigdata's data.
func (s *Service) TagAids(c context.Context, tagID int64, pn, ps int) (total int, arcs []*model.BvArc, err error) {
	defer func() {
		if len(arcs) == 0 {
			arcs = _emptyBvArc
		}
	}()
	if err = s.checkTag(c, tagID); err != nil {
		err = nil
		return
	}
	return s.tagArcs(c, tagID, pn, ps)
}

func (s *Service) tagArcs(c context.Context, tagID int64, pn, ps int) (total int, arcs []*model.BvArc, err error) {
	var (
		start         = (pn - 1) * ps
		end           = start + ps - 1
		aids, allAids []int64
	)
	if allAids, err = s.dao.TagAids(c, tagID); err != nil {
		log.Error("s.dao.TagAids(%d) error(%v)", tagID, err)
		if allAids, err = s.dao.TagAidsBakCache(c, tagID); err != nil {
			log.Error("s.dao.TagAidsBakCache(%d) error(%v)", tagID, err)
			return
		}
	} else if len(allAids) > 0 {
		if err := s.cache.Do(c, func(c context.Context) {
			if err := s.dao.SetTagAidsBakCache(c, tagID, allAids); err != nil {
				log.Error("%+v", err)
			}
		}); err != nil {
			log.Error("%+v", err)
		}
	}
	total = len(allAids)
	if total < start {
		err = ecode.NothingFound
		return
	}
	if total > end {
		aids = allAids[start : end+1]
	} else {
		aids = allAids[start:]
	}
	if len(aids) > 0 {
		arcs, err = s.archives(c, aids)
	}
	return
}

// nolint:gomnd
func (s *Service) archives(c context.Context, aids []int64) (data []*model.BvArc, err error) {
	var (
		arg = &arcmdl.ArcsRequest{Aids: aids}
		res *arcmdl.ArcsReply
	)
	archivesArgLog("TagAids", aids)
	if res, err = s.arcGRPC.Arcs(c, arg); err != nil {
		log.Error("arcrpc.Archives3(%v) error(%v)", aids, err)
		return
	}
	for _, aid := range aids {
		arc, ok := res.Arcs[aid]
		if !ok {
			continue
		}
		if arc.Access >= 10000 {
			arc.Stat.View = -1
		}
		data = append(data, model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)))
	}
	return
}

func (s *Service) checkTag(ctx context.Context, tid int64) error {
	reply, err := s.tagGRPC.Tag(ctx, &taggrpc.TagReq{Tid: tid})
	if err != nil {
		log.Error("checkTag tid:%v,error:%+v", tid, err)
		return nil
	}
	state := reply.GetTag().GetState()
	if state == model.TagStateDeleted || state == model.TagStateBlocked {
		return xecode.TagIsSealing
	}
	return nil
}

// TagDetail group web tag data.
func (s *Service) TagDetail(c context.Context, tagID int64, ps int) (data *model.TagDetail, err error) {
	var tagInfo *model.TagTop
	if tagInfo, err = s.tag.TagTop(c, &model.ReqTagTop{Tid: tagID}); err != nil {
		return
	}
	data = &model.TagDetail{TagTop: tagInfo}
	data.Total, data.List, _ = s.tagArcs(c, tagID, _samplePn, ps)
	if len(data.List) == 0 {
		data.List = _emptyBvArc
	}
	if len(data.Similars) == 0 {
		data.Similars = make([]*model.SimilarTag, 0)
	}
	return
}

func (s *Service) TagArchives(c context.Context, req *model.TagArcsReq) (*model.TagArcsReply, error) {
	// 获取rids
	ridsReq := &taggrpc.RidsByTagReq{
		Tid:    req.TagID,
		Source: req.Source,
		Typ:    model.TagTypArc,
		Offset: req.Offset,
		Ps:     req.PS,
	}
	ridsReply, err := s.dao.RidsByTag(c, ridsReq)
	if err != nil {
		log.Error("s.dao.RidsByTag(%+v) (%+v)", req, err)
		return nil, err
	}
	// 获取arc详情
	var arcs []*model.BvArc
	if len(ridsReply.GetRids()) > 0 {
		if arcs, err = s.archives(c, ridsReply.GetRids()); err != nil {
			log.Error("s.archives(%+v) (%+v)", ridsReply.GetRids(), err)
			return nil, err
		}
	}
	list := make([]*model.TagArcItem, 0, len(arcs))
	for _, arc := range arcs {
		item := &model.TagArcItem{}
		item.FormArc(arc.Arc)
		list = append(list, item)
	}
	reply := &model.TagArcsReply{
		HasMore: ridsReply.GetHasmore(),
		Offset:  ridsReply.GetOffset(),
		List:    list,
	}
	return reply, nil
}

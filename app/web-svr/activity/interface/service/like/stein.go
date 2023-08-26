package like

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/archive/service/api"
	steinapi "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) SingleGroupWebData(c context.Context, mid int64) (res map[string]*like.SteinData, err error) {
	var (
		aids  []int64
		arcs  map[int64]*api.Arc
		evals map[int64]*steinapi.MarkEvaluations
	)
	tmp := s.steinData
	for _, v := range tmp {
		aids = append(aids, v.Aids...)
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if reply, e := client.ArchiveClient.Arcs(ctx, &api.ArcsRequest{Aids: aids}); e != nil {
			log.Error("SingleGroupWebData s.arcClient.Arcs aids(%v) error(%v)", aids, e)
		} else {
			arcs = reply.Arcs
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if reply, e := s.steinClient.MarkEvaluations(ctx, &steinapi.MarkEvaluationsReq{Mid: mid, Aids: aids}); e != nil {
			log.Error("SingleGroupWebData s.steinClient.MarkEvaluations aids(%v) mid(%d) error(%v)", aids, mid, e)
		} else {
			evals = reply.Items
		}
		return nil
	})
	group.Wait()
	res = make(map[string]*like.SteinData, len(tmp))
	for k, v := range tmp {
		item := &like.SteinData{Name: v.Name}
		for _, aid := range v.Aids {
			if arc, ok := arcs[aid]; ok && arc.IsNormal() {
				listArc := like.CopyFromArc(arc)
				if eval, ok := evals[aid]; ok && eval != nil {
					listArc.Stat.Evaluation = eval.Evaluation
					listArc.Stat.Mark = eval.Mark
				}
				item.List = append(item.List, listArc)
			}
		}
		res[k] = item
	}
	return
}

func (s *Service) loadSteinWebData() {
	res, err := s.dao.SourceItem(context.Background(), s.c.SteinV2.Vid)
	if err != nil {
		log.Error("loadSteinWebData s.dao.SourceItem(%d) error(%v)", s.c.Taaf.Vid, err)
		return
	}
	tmp := new(like.SteinWebData)
	if err = json.Unmarshal(res, tmp); err != nil {
		log.Error("loadSteinWebData s.dao.SourceItem(%d) error(%v)", s.c.Scholarship.ArcVid, err)
		return
	}
	if len(tmp.List) == 0 {
		log.Error("loadSteinWebData data len 0")
		return
	}
	tmpData := make(map[string]*like.SteinMemData, len(tmp.List))
	for _, v := range tmp.List {
		if v == nil || v.Data == nil {
			log.Error("loadSteinWebData v.Data(%v) is nil", v)
			continue
		}
		aids, err := xstr.SplitInts(v.Data.Aids)
		if err != nil {
			log.Error("loadSteinWebData xstr.SplitInts(%s) error(%v)", v.Data.Aids, err)
			continue
		}
		tmpData[v.Data.Name] = &like.SteinMemData{Name: v.Name, Aids: aids}
	}
	if len(tmpData) == 0 {
		log.Error("loadSteinWebData len(tmpData) == 0")
		return
	}
	s.steinData = tmpData
	log.Info("loadSteinWebData() success")
}

package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/job/internal/model"

	"go-common/library/sync/errgroup.v2"
)

const (
	_tableArchive = "archive"
)

func (s *Service) arcConsumeproc() {
	var err error
	defer s.waiter.Done()
	for {
		msg, ok := <-s.arcNotifySub.Messages()
		if !ok || s.daoClosed {
			log.Info("arc databus Consumer exit")
			break
		}
		//nolint:errcheck
		msg.Commit()
		var ms = &model.ArcMsg{}
		log.Info("arcConsumeproc New message: %v", msg.Value)
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		switch ms.Table {
		case _tableArchive:
			if ms.New != nil {
				s.arcHandle(ms.New)
			}
		}
	}
}

func (s *Service) arcHandle(arc *model.Archive) {
	if !arc.IsSteinsGate() || !arc.IsNormal() {
		log.Info("ArcHandle Aid %d is not SteinsGate or not open", arc.Aid)
		return
	}
	var (
		c           = context.Background()
		nodes       []*model.Node
		steinsViews map[int64]struct{}
		graph       *model.Graph
		viewReply   *api.SteinsGateViewReply
		gFirstCid   int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if viewReply, err = s.dao.ArcView(ctx, arc.Aid); err != nil {
			return
		}
		steinsViews = make(map[int64]struct{}, len(viewReply.Pages))
		for _, v := range viewReply.Pages {
			steinsViews[v.Cid] = struct{}{}
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if graph, err = s.dao.Graph(ctx, arc.Aid); err != nil { // if graph is not found, return the error
			return
		}
		if !graph.IsPass() {
			log.Warn("arcHandle Aid %d, Gid %d, Graph is not passed. ignore", arc.Aid, graph.ID)
			err = ecode.NothingFound
			return
		}
		if nodes, gFirstCid, err = s.dao.Nodes(ctx, graph.ID); err != nil || len(nodes) == 0 {
			log.Error("arcHandle Aid %d, Gid %d, Graph doesn't have nodes. ignore", arc.Aid, graph.ID)
			err = ecode.NothingFound
			return
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("arcHandler Aid %d Err %v", arc.Aid, err)
		return
	}
	for _, v := range nodes {
		if _, ok := steinsViews[v.CID]; !ok { // there is a cid used in the graph but not in the passed cids from ArcService, we need to put it under repair
			log.Warn("arcHandler Aid %d GraphID %d Cid %d invalid, return the graph", arc.Aid, graph.ID, v.CID)
			s.returnGraph(c, &model.ReqReturnGraph{
				Arc:     viewReply.Arc,
				GraphID: graph.ID,
			})
			return
		}
	}
	if gFirstCid > 0 && gFirstCid != viewReply.FirstCid {
		log.Warn("arcHandler FirstCid Not Equal! Aid %d GraphID %d Gcid %d Acid %d", arc.Aid, graph.ID, gFirstCid, viewReply.FirstCid)
		s.dao.UpArcFirstCid(c, arc.Aid, gFirstCid)
	}
	log.Info("arcHandle Aid %d, GraphID %d Len(nodes) %d Pass the check", arc.Aid, graph.ID, len(nodes))
}

func (s *Service) returnGraph(ctx context.Context, req *model.ReqReturnGraph) {
	if req.Arc == nil {
		return
	}
	s.dao.ReturnGraph(ctx, req.GraphID)
	s.dao.DelGraphCache(ctx, req.Arc.Aid)
	if mid := req.Arc.Author.Mid; mid != 0 {
		if err := s.dao.SendMessage(ctx, []int64{mid}, s.conf.Message.MC, s.conf.Message.Title,
			fmt.Sprintf(s.conf.Message.Content, req.Arc.Title, req.Arc.Aid)); err != nil { // for the error of sending, we don't retry
			log.Error("SendMessage Aid %d, GraphID %d, Err %v", req.Arc.Aid, req.GraphID, err)
			return
		}
	}

}

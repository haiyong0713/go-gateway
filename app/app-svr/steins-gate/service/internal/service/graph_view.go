package service

import (
	"context"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) Evaluation(c context.Context, aid int64) (evaluation string, err error) {
	var eval int64
	if eval, err = s.markDao.Evaluation(c, aid); err != nil {
		log.Error("View Evaluation Err %v", err)
		return
	}
	if eval != 0 {
		evaluation = model.Evaluation(eval)
	}
	return
}

func (s *Service) GraphView(ctx context.Context, aid int64) (page *api.Page, graphInfo *api.GraphInfo, evaluation string, err error) {
	var steinsView *arcgrpc.SteinsGateViewReply
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) (err error) {
		if steinsView, err = s.arcDao.ArcView(ctx, aid); err != nil { // archive-service Pages failed, just return error
			log.Error("%+v", err)
			return
		}
		if !steinsView.Arc.IsSteinsGate() { // not a steinsGate arc, just return
			log.Warn("View Aid %d, Not SteinsGate Arc", aid)
			err = ecode.NotSteinsGateArc
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if graphInfo, err = s.dao.GraphInfo(ctx, aid); err != nil { // if the latest tree is not valid, return an error
			log.Warn("View Aid %d, Non Valid Tree Err %v", aid, err)
			err = ecode.NonValidGraph
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		var eval int64
		if eval, err = s.markDao.Evaluation(ctx, aid); err != nil {
			log.Error("View Evaluation Err %v", err)
			return
		}
		if eval != 0 {
			evaluation = model.Evaluation(eval)
		}
		return
	})
	if err = g.Wait(); err != nil {
		return
	}
	page, err = matchPage(steinsView, graphInfo)
	return
}

func (s *Service) MarkEvaluations(c context.Context, mid int64, aid []int64) (res map[int64]*api.MarkEvaluations, err error) {
	var (
		evals = make(map[int64]int64)
		marks = make(map[int64]int64)
	)
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		if evals, err = s.markDao.Evaluations(ctx, aid); err != nil {
			log.Error("View Evaluations Err %v", err)
		}
		return
	})
	if mid > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if marks, err = s.markDao.Marks(ctx, aid, mid); err != nil {
				log.Error("View Marks Err %v", err)
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("Evaluations/Marks Err %v aids %v mid %d", err, aid, mid)
		return
	}
	res = make(map[int64]*api.MarkEvaluations)
	for _, item := range aid {
		rcd := new(api.MarkEvaluations)
		if v, ok := evals[item]; ok {
			rcd.Evaluation = model.Evaluation(v)
		}
		if v, ok := marks[item]; ok {
			rcd.Mark = v
		}
		res[item] = rcd
	}
	return

}

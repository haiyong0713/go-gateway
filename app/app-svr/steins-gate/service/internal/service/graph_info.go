package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// GraphInfo gets the latest and valid graph by the aid
func (s *Service) GraphInfo(c context.Context, aid int64) (a *api.GraphInfo, err error) {
	return s.dao.GraphInfo(c, aid)
}

// View gets the View Info and the Record Info
func (s *Service) View(ctx context.Context, req *api.ViewReq) (resp *api.ViewReply, err error) {
	resp = new(api.ViewReply)
	if resp.Page, resp.Graph, resp.Evaluation, err = s.GraphView(ctx, req.Aid); err != nil {
		return
	}
	if resp.Graph.NoEvaluation > 0 {
		resp.Evaluation = ""
	}
	if req.Mid == 0 { // if not logged in, just return the graph and the page
		if model.HasHvar(resp.Graph) && resp.Graph.GuestOverwriteRegionalVars > 0 {
			resp.ToastMsg = s.c.Rule.ToastMsg.NeedLogin
		}
		// 如果不覆盖本地变量的话就不用出需要登陆的提示了
		return
	}
	if req.Buvid == "" {
		s.IncrPromBusiness("rpc_no_buvid")
	} else {
		s.IncrPromBusiness("rpc_buvid")
	}
	var recordErr error
	eg := errgroup.WithContext(ctx)
	eg.Go(func(c context.Context) (err error) {
		var (
			record *api.GameRecords
			node   *api.GraphNode
			edge   *api.GraphEdge
		)
		if resp.Graph.NoBacktracking > 0 {
			return
		}
		if record, resp.RecordState, recordErr = s.getRecord(ctx, req, resp.Graph); recordErr != nil || resp.RecordState != 0 { // 如果有错，则不返回currentNode/Edge
			if resp.RecordState == model.RecordInvalid {
				resp.ToastMsg = s.c.Rule.ToastMsg.GraphUpdate
			}
			return
		}
		if model.IsEdgeGraph(resp.Graph) || model.IsInterventionGraph(resp.Graph) { // edge/中插 图需要获取edge信息，toNode信息，并且做兼容性拼接
			if node, edge, err = s.pickEdgeByRcd(ctx, record.CurrentEdge, resp.Graph); err != nil || node == nil {
				log.Warn("View Record Mid %d, GraphID %d, CurrentID %d, Not found Err %v", req.Mid, resp.Graph.Id, record.CurrentNode, err)
				err = ecode.NothingFound
				return
			}
			resp.CurrentNode = node
			resp.CurrentEdge = edge
			model.ClientCompatibility(resp.CurrentNode, record.CurrentEdge) // 拼接
			return
		}
		if node, err = s.dao.Node(ctx, record.CurrentNode); err != nil || node == nil { // node图还是走老逻辑
			log.Warn("View Record Mid %d, GraphID %d, CurrentID %d, Not found Err %v", req.Mid, resp.Graph.Id, record.CurrentNode, err)
			err = ecode.NothingFound
		}
		resp.CurrentNode = node
		return
	})
	eg.Go(func(c context.Context) (err error) {
		resp.Mark, err = s.GetMark(ctx, req.Aid, req.Mid)
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	return
}

func (s *Service) getRecord(ctx context.Context, req *api.ViewReq, graphInfo *api.GraphInfo) (record *api.GameRecords, state int32, err error) {
	if record, err = s.recordDao.Record(ctx, req.Mid, graphInfo.Id, req.Buvid); err != nil {
		log.Warn("View Record Mid %d, GraphID %d Err %v", req.Mid, graphInfo.Id, err)
		return
	}
	if record == nil {
		if record, err = s.recordDao.RecordByAid(ctx, req.Mid, req.Aid); err != nil { // db not found, return
			log.Warn("View Record RecordByAid err(%v) mid(%d) aid(%d)", err, req.Mid, req.Aid)
			return
		}
		if record == nil {
			err = ecode.NothingFound
			return
		}
		if record.GraphId != graphInfo.Id { // find the record from DB with AID, compare it!!
			state = model.RecordInvalid
			log.Warn("View Record Mid %d, GraphID %d, RecordGraphID %d Not Equal, State -1", req.Mid, graphInfo.Id, record.GraphId)
		}
		return
	}
	return
}

func (s *Service) pickEdgeByRcd(ctx context.Context, currentEdge int64, graphInfo *api.GraphInfo) (toNode *api.GraphNode, edge *api.GraphEdge, err error) {
	if edge, err = s.pickEdge(ctx, currentEdge, graphInfo); err != nil {
		return
	}
	toNode, err = s.dao.Node(ctx, edge.ToNode) // 这里相比较node增加了串行的to_node的查询
	return

}

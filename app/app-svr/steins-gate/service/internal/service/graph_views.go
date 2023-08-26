package service

import (
	"context"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func matchPage(view *arcgrpc.SteinsGateViewReply, graphInfo *api.GraphInfo) (page *api.Page, err error) {
	for _, v := range view.Pages { // find the starting point, put its page into the response
		if v.Cid == graphInfo.FirstCid {
			page = model.SteinsPage(v)
			return
		}
	}
	log.Warn("View Aid %d, Non Valid Tree %d Because no starting point is found in the View", graphInfo.Aid, graphInfo.Id)
	err = ecode.NonValidGraph
	return
}

func combineView(graphInfos map[int64]*api.GraphInfo, steinsViews map[int64]*arcgrpc.SteinsGateViewReply, aids []int64) (result map[int64]*api.ViewReply, err error) {
	result = make(map[int64]*api.ViewReply, len(aids))
	for _, aid := range aids {
		graphInfo, okG := graphInfos[aid]
		if !okG {
			continue
		}
		arcView, okA := steinsViews[aid]
		if !okA {
			continue
		}
		var viewReply = &api.ViewReply{Graph: graphInfo}
		if viewReply.Page, err = matchPage(arcView, graphInfo); err != nil { // 缺少page信息直接返回报错
			return
		}
		result[aid] = viewReply
	}
	return
}

// Views multi get the View Info and the Record Info
//
//nolint:gocognit
func (s *Service) Views(ctx context.Context, req *api.ViewsReq) (resp *api.ViewsReply, err error) {
	var (
		steinsViews = make(map[int64]*arcgrpc.SteinsGateViewReply)
		graphInfos  map[int64]*api.GraphInfo
		reqRecord   = &model.RecordReq{
			Buvid:        req.Buvid,
			MID:          req.Mid,
			GraphWithAID: make(map[int64]int64),
		}
		recordsByGraph map[int64]*api.GameRecords
		invalidRecords map[int64]struct{}
		historyNodes   map[int64]*api.GraphNode
		edges          map[int64]*api.GraphEdge
		missAIDs       []int64
	)
	if req.Buvid == "" {
		s.IncrPromBusiness("rpc_no_buvid")
	} else {
		s.IncrPromBusiness("rpc_buvid")
	}
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) (err error) {
		var arcViews map[int64]*arcgrpc.SteinsGateViewReply
		if arcViews, err = s.arcDao.ArcViews(ctx, append(req.Aids, req.AidsWithHistory...)); err != nil { // archive-service Pages failed, just return error
			log.Error("%+v", err)
			return
		}
		for aid, sv := range arcViews {
			if sv.Arc != nil && !sv.Arc.IsSteinsGate() { // not a steinsGate arc, just return
				log.Warn("View Aid %d, Not SteinsGate Arc", aid)
				continue
			}
			steinsViews[aid] = sv
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		graphInfos = s.dao.GraphInfos(ctx, append(req.Aids, req.AidsWithHistory...))
		return
	})
	if err = g.Wait(); err != nil {
		return nil, err
	}
	resp = new(api.ViewsReply)
	if resp.Views, err = combineView(graphInfos, steinsViews, req.Aids); err != nil {
		return
	}
	if len(req.AidsWithHistory) == 0 { // 不用处理存档逻辑
		return
	}
	if resp.ViewsWithHistory, err = combineView(graphInfos, steinsViews, req.AidsWithHistory); err != nil {
		return
	}
	if req.Mid == 0 { // if not logged in, just return the graph and the page
		for _, item := range resp.Views {
			if model.HasHvar(item.Graph) {
				item.ToastMsg = s.c.Rule.ToastMsg.NeedLogin
			}
		}
		for _, item := range resp.ViewsWithHistory {
			if model.HasHvar(item.Graph) {
				item.ToastMsg = s.c.Rule.ToastMsg.NeedLogin
			}
		}
		return
	}
	for _, v := range resp.ViewsWithHistory {
		if v.Graph == nil {
			continue
		}
		reqRecord.GraphWithAID[v.Graph.Id] = v.Graph.Aid
	}
	if recordsByGraph, missAIDs, err = s.recordDao.Records(ctx, reqRecord); err != nil {
		log.Error("Records Req %+v, err %v", req, err)
		return
	}
	eg := errgroup.WithContext(ctx)
	if len(recordsByGraph) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			var (
				nodeIDs     []int64 // node_ids是node_ids和edges的to_node的集合
				historyEIDs []int64
			)
			for _, v := range recordsByGraph {
				if item, ok := graphInfos[v.Aid]; ok {
					if model.IsEdgeGraph(item) || model.IsInterventionGraph(item) { // 如果是edge/中插 图，把CurrentEdge加入到edges里面
						if v.CurrentEdge == model.RootEdge {
							nodeIDs = append(nodeIDs, item.FirstNid) // edge图当为根节点的时候，需要把FirstNid加到nodes，以便后面可以拿到对应edge=rootEdge的ToNode信息
						}
						historyEIDs = append(historyEIDs, v.CurrentEdge) // 如果CurrentEdge为rootEdge也加进去，后面拿edges会过滤掉，最后会判断CurrentEdge是不是rootEdge，再特殊处理
						continue
					}
					nodeIDs = append(nodeIDs, v.CurrentNode) // 走node逻辑，把CurrentNode加到nodes
				}
			}
			if len(historyEIDs) > 0 { // 下面调用方法会忽略掉edgeId为root的数据
				if edges, err = s.dao.EdgesWithoutRoot(ctx, historyEIDs); err != nil {
					log.Error("edges %v err %v", historyEIDs, err)
				}
				for _, v := range edges {
					nodeIDs = append(nodeIDs, v.ToNode) // 因为edge图也需要拿对应的ToNode的信息，把ToNode加到nodes里面
				}
			}
			if historyNodes, err = s.dao.Nodes(ctx, nodeIDs); err != nil {
				log.Error("historyNodes %v err %v", nodeIDs, err)
			}
			return
		})
	}
	if len(missAIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if invalidRecords, err = s.recordDao.RecordByAids(ctx, req.Mid, missAIDs); err != nil {
				log.Error("RecordByAids mid %d aids %v, err %v", req.Mid, missAIDs, err)
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	for _, v := range resp.ViewsWithHistory {
		if v.Graph == nil {
			continue
		}
		var (
			record *api.GameRecords
			graph  *api.GraphInfo
			ok     bool
		)
		if _, ok = invalidRecords[v.Graph.Aid]; ok { // 存档失效的无需下面的处理
			v.RecordState = model.RecordInvalid
			v.ToastMsg = s.c.Rule.ToastMsg.GraphUpdate
			continue
		}
		if record, ok = recordsByGraph[v.Graph.Id]; !ok { // 找不到存档不处理
			continue
		}
		if graph, ok = graphInfos[v.Graph.Aid]; !ok { // 找不到graph不处理
			continue
		}
		if !model.IsEdgeGraph(graph) && !model.IsInterventionGraph(graph) { // node图，单层逻辑
			if node, okN := historyNodes[record.CurrentNode]; okN {
				v.CurrentNode = node
			}
			continue
		}
		// edge图，双层逻辑，先查edge再查node
		if record.CurrentEdge == model.RootEdge { // 根节点特殊逻辑
			if node, okN := historyNodes[graph.FirstNid]; okN {
				v.CurrentNode = node
			}
			v.CurrentEdge = model.GetFirstEdge(graph)
			model.ClientCompatibility(v.CurrentNode, v.CurrentEdge.Id) // 拼接
			continue
		}
		if edge, okE := edges[record.CurrentEdge]; okE { // 普通edge双层逻辑
			v.CurrentEdge = edge
			if node, okN := historyNodes[edge.ToNode]; okN {
				v.CurrentNode = node
				model.ClientCompatibility(v.CurrentNode, v.CurrentEdge.Id) // 拼接
			}
			continue
		}
	}
	return

}

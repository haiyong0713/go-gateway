package service

import (
	"context"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/admin/internal/model"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

func (s *Service) GraphShow(c context.Context, graphID int64) (res []*model.NodeShow, err error) {
	var (
		nodes        = make(map[int64]*api.GraphNode)
		edgesByFromN = make(map[int64][]*api.GraphEdge)
		graph        *model.GraphDB
		videos       = make(map[int64]*model.Video)
		rawNodes     []*api.GraphNode
	)
	if graph, err = s.dao.GraphAuditByID(c, graphID); err != nil {
		log.Error("GraphShow AuditGID %d Err %v", graphID, err)
		return
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(c context.Context) (err error) {
		var videoupView *model.VideoUpView
		if videoupView, err = s.dao.VideoUpView(c, graph.Aid); err != nil {
			log.Error("GraphShow Nodes GraphID %d, Err %v", graphID, err)
			return
		}
		for _, v := range videoupView.Videos {
			videos[v.Cid] = v
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		if rawNodes, err = s.dao.GraphNodeList(c, graphID); err != nil {
			log.Error("GraphShow Nodes GraphID %d, Err %v", graphID, err)
			return
		}
		for _, v := range rawNodes {
			nodes[v.Id] = v
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		var rawEdges []*api.GraphEdge
		if rawEdges, err = s.dao.GraphEdgeList(c, graphID); err != nil {
			log.Error("GraphShow Edges GraphID %d, Err %v", graphID, err)
			return
		}
		for _, v := range rawEdges {
			edgesByFromN[v.FromNode] = append(edgesByFromN[v.FromNode], v)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	for _, v := range rawNodes {
		var (
			node, toNode   *api.GraphNode
			video, toVideo *model.Video
			edges          []*api.GraphEdge
			ok             bool
		)
		if node, ok = nodes[v.Id]; !ok {
			continue
		}
		if video, ok = videos[v.Cid]; !ok {
			continue
		}
		if edges, ok = edgesByFromN[v.Id]; !ok { // 没有下叉剧情的节点不展示
			continue
		}
		nodeShow := new(model.NodeShow)
		nodeShow.FromProto(node, video, graph.Aid)
		for _, v := range edges {
			if toNode, ok = nodes[v.ToNode]; !ok {
				continue
			}
			if toVideo, ok = videos[v.ToNodeCid]; !ok {
				continue
			}
			toNodeShow := new(model.NodeCore) // 拼接to_node详情
			toNodeShow.FromProto(toNode, toVideo, graph.Aid)
			edgeShow := new(model.EdgeShow) // 拼接edge的相关信息
			edgeShow.FromProto(v, toNodeShow)
			nodeShow.Choices = append(nodeShow.Choices, edgeShow)
		}
		res = append(res, nodeShow)
	}
	return

}

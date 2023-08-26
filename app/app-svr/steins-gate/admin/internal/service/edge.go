package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/admin/internal/model"
	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

func (s *Service) EdgeInfoV2Audit(c context.Context, GraphVersion, edgeID int64) (res *model.EdgeInfoV2, err error) {
	var (
		nodeInfo         *api.GraphNode
		edgeInfo         *api.GraphEdge
		oriChoices       []*api.GraphEdge
		nodeID, firstCid int64
		questions        []*model.Question
		groups           map[int64]*api.EdgeGroup
		arcView          *arcgrpc.SteinsGateViewReply
		graphInfo        *model.GraphDB
	)
	if !model.IsRootEdge(edgeID) {
		if edgeInfo, err = s.dao.RawEdge(c, edgeID); err != nil { // 保证了edgeInfo一定不可能为nil
			log.Error("s.dao.Edge err(%v)", err)
			return
		}
		nodeID = edgeInfo.ToNode
	} else {
		if nodeID, firstCid, err = s.dao.StartingPoint(c, GraphVersion); err != nil {
			return
		}
		edgeInfo = &api.GraphEdge{
			Id:        model.RootEdge,
			GraphId:   GraphVersion,
			ToNode:    nodeID,
			ToNodeCid: firstCid,
		}
	}
	if graphInfo, err = s.dao.GraphAuditByID(c, GraphVersion); err != nil {
		return
	}
	eg := errgroup.WithContext(c) // 获取选项
	eg.Go(func(c context.Context) (err error) {
		nodeInfo, err = s.dao.Node(c, nodeID)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		oriChoices, err = s.dao.EdgeByNode(c, nodeID)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		arcView, err = s.dao.ArcView(c, graphInfo.Aid)
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if nodeInfo.GraphId != GraphVersion { // if node not match graphid, return 400
		err = ecode.RequestErr
		return
	}
	groups = model.BuildEdgeGroup(nodeInfo)
	if questions, err = model.BuildQuestions(oriChoices, groups, arcView, nodeInfo); err != nil {
		return
	}
	res = new(model.EdgeInfoV2)
	res.Edges = new(model.EdgeV2)
	res.Edges.BuildEdge(questions, nodeInfo)
	res.EdgeID = edgeInfo.Id
	res.Title = nodeInfo.Name
	return
}

func (s *Service) GetEdgeIdByNode(c context.Context, nodeId int64) (edgeId int64, err error) {
	var (
		rawEdges []*api.GraphEdge
		nodeInfo *api.GraphNode
	)
	nodeInfo, err = s.dao.Node(c, nodeId)
	if err != nil {
		log.Error("EdgeInfoV2Audit Node %d, Err %v", nodeId, err)
		return
	}
	if nodeInfo == nil {
		err = ecode.NothingFound
		return
	}
	if nodeInfo.IsStart == 1 {
		return 1, nil
	}
	if rawEdges, err = s.dao.GraphEdgeList(c, nodeInfo.GraphId); err != nil {
		log.Error("EdgeInfoV2Audit Edges GraphID %d, Err %v", nodeInfo.GraphId, err)
		return
	}
	for _, item := range rawEdges {
		if item.ToNode == nodeId {
			return item.Id, nil
		}
	}
	return 0, errors.New("Not Match NodeId to EdgeId")

}

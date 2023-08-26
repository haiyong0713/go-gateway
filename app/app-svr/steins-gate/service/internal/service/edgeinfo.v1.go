package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func (s *Service) EdgeInfo(c context.Context, params *model.NodeInfoParam, mid int64, graphInfo *api.GraphInfo) (res *model.NodeInfo, err error) {
	var (
		nodeInfo                    *api.GraphNode
		edgeInfo                    *api.GraphEdge
		showChoices                 []*model.Choice
		oriChoices, filteredChoices []*api.GraphEdge
		arcView                     *arcgrpc.SteinsGateViewReply
		hiddenVars                  []*model.HiddenVar
		hasHvar                     bool
		currentCursor               int64
		hvarRec                     *model.HiddenVarsRecord
		randomV                     *model.RandomVar
	)
	res = &model.NodeInfo{
		Preload: &model.Preload{Video: make([]*model.PreVideo, 0)},
	}
	if _, oriChoices, nodeInfo, edgeInfo, hasHvar, hvarRec, randomV, currentCursor, arcView, res.StoryList, _, err = s.edgeInfoProcess(c, params, mid, graphInfo, nil); err != nil {
		return
	}
	if hasHvar {
		if mid > 0 && params.Portal != 1 { // 回溯操作不修改隐藏变量存档
			//nolint:errcheck
			hvarRec.ApplyAttr(edgeInfo.Attribute) // 直接用当前edge进行隐藏变量操作
			//nolint:errcheck
			s.hvaDao.AddHiddenVarRecDBCache(c, mid, graphInfo.Id, params.EdgeID, currentCursor, params.Buvid, hvarRec, false) // set the new record into cache
		}
		if filteredChoices, err = model.FilterEdges(hvarRec, oriChoices, randomV); err != nil { // 条件过滤
			log.Error("Mid %d Aid %d Gid %d, FilterEdges Err %v", mid, params.AID, graphInfo.Id, err)
			return
		}
		hiddenVars = model.DisplayHvars(hvarRec, randomV)
	} else { // no need to filter choices
		filteredChoices = oriChoices
	}
	showChoices = model.BuildChoicesEdge(filteredChoices, s.skinInfo(graphInfo, nodeInfo))
	res.BuildEdgeReply(nodeInfo, edgeInfo, showChoices, arcView, hiddenVars)
	res.Preload.FromNode(nodeInfo, showChoices, params.AID)
	return
}

func (s *Service) pickEdge(c context.Context, edgeID int64, graphInfo *api.GraphInfo) (edgeInfo *api.GraphEdge, err error) {
	if edgeInfo, err = s.dao.Edge(c, edgeID, graphInfo); err != nil {
		log.Error("s.dao.Node err(%v) nodeID(%d)", err, edgeID)
		return
	}
	if edgeInfo == nil {
		err = ecode.NothingFound
	}
	return

}

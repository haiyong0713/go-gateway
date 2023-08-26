package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/stat/prom"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func (s *Service) NodeInfoV2(c context.Context, params *model.EdgeInfoV2Param, mid int64, graphInfo *api.GraphInfo) (res *model.EdgeInfoV2, err error) {
	var (
		nodeInfo   *api.GraphNode
		questions  []*model.Question
		oriEdges   []*api.GraphEdge
		hiddenVars []*model.HiddenVar
		groups     map[int64]*api.EdgeGroup
		hasHvar    bool
		hvarRec    *model.HiddenVarsRecord
		randomV    *model.RandomVar
		arcView    *arcgrpc.SteinsGateViewReply
	)
	res = &model.EdgeInfoV2{
		Preload: &model.Preload{Video: make([]*model.PreVideo, 0)},
	}
	if nodeInfo, oriEdges, arcView, hasHvar, hvarRec, randomV, res.StoryList, err =
		s.nodeInfoProcess(c, &params.NodeInfoParam, mid, graphInfo); err != nil {
		return
	}
	groups = model.BuildEdgeGroup(nodeInfo)
	if questions, res.IsLeaf, err = model.BuildQuestions(oriEdges, groups, model.ArcViewPicker(arcView), nodeInfo, graphInfo, nil); err != nil {
		return
	}
	if hasHvar {
		hiddenVars = model.DisplayHvars(hvarRec, randomV)
		if needProm, err := model.IsQuestionProm(hvarRec, randomV, questions); err == nil && needProm {
			log.Warn("NoInterventionGraphNoChoices: %+v ", params)
			prom.BusinessInfoCount.Incr("NoInterventionGraphNoChoices")
		}
	}
	// 结果拼接逻辑
	res.BuildReplyV2(nodeInfo, questions, hiddenVars, s.skinInfo(graphInfo, nodeInfo))
	res.Preload.FromNodeV2(nodeInfo, questions, params.AID)
	return

}

package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/stat/prom"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func (s *Service) EdgeInfoV2(c context.Context, params *model.EdgeInfoV2Param, mid int64, graphInfo *api.GraphInfo) (res *model.EdgeInfoV2, err error) {
	var (
		edges         map[int64]*api.GraphEdge
		nodeInfo      *api.GraphNode
		edgeInfo      *api.GraphEdge
		questions     []*model.Question
		oriEdges      []*api.GraphEdge
		currentCursor int64
		arcView       *arcgrpc.SteinsGateViewReply
		groups        map[int64]*api.EdgeGroup
		hasHvar       bool
		hvarRec       *model.HiddenVarsRecord
		randomV       *model.RandomVar
		attributes    []string
		edgeAnms      map[int64]*api.EdgeFrameAnimations
	)
	res = &model.EdgeInfoV2{
		Preload:        &model.Preload{Video: make([]*model.PreVideo, 0)},
		NoTutorial:     graphInfo.NoTutorial,
		NoBacktracking: graphInfo.NoBacktracking,
		NoEvaluation:   graphInfo.NoEvaluation,
	}
	if edges, oriEdges, nodeInfo, edgeInfo, hasHvar, hvarRec, randomV, currentCursor, arcView, res.StoryList, edgeAnms, err =
		s.edgeInfoProcess(c, &params.NodeInfoParam, mid, graphInfo, params.ChoiceIDs); err != nil {
		return
	}
	if model.IsInterventionGraph(graphInfo) {
		var groupIds []int64
		for _, item := range oriEdges {
			groupIds = append(groupIds, item.GroupId)
		}
		if groups, err = s.dao.EdgeGroups(c, groupIds); err != nil {
			return
		}
	} else { // 老的非中插树无edge group 数据
		groups = model.BuildEdgeGroup(nodeInfo)
	}
	if questions, res.IsLeaf, err = model.BuildQuestions(oriEdges, groups, model.ArcViewPicker(arcView), nodeInfo, graphInfo, edgeAnms); err != nil {
		return
	}
	if hasHvar {
		if mid > 0 && params.Portal != 1 && len(params.ChoiceIDs) > 0 { // 回溯操作不修改隐藏变量存档
			for _, item := range params.ChoiceIDs {
				if ed, ok := edges[item]; ok {
					attributes = append(attributes, ed.Attribute)
				}
			}
			//nolint:errcheck
			hvarRec.ApplyAttrs(attributes) // 直接用当前edge进行隐藏变量操作
			//nolint:errcheck
			s.hvaDao.AddHiddenVarRecDBCache(c, mid, graphInfo.Id, params.EdgeID, currentCursor, params.Buvid, hvarRec, graphInfo.NoBacktracking > 0) // set the new record into cache
		}
		res.HiddenVars = model.DisplayHvars(hvarRec, randomV)
		judgeSkipOverwrite(res.HiddenVars, mid, graphInfo)
		if !model.IsInterventionGraph(graphInfo) {
			if needProm, err := model.IsQuestionProm(hvarRec, randomV, questions); err == nil && needProm {
				log.Warn("NoInterventionGraphNoChoices: %+v ", params)
				prom.BusinessInfoCount.Incr("NoInterventionGraphNoChoices")
			}
		}
	}
	res.BuildEdgeV2Reply(nodeInfo, edgeInfo, questions, s.skinInfo(graphInfo, nodeInfo))
	res.Preload.FromNodeV2(nodeInfo, questions, params.AID)
	return
}

func judgeSkipOverwrite(vars []*model.HiddenVar, mid int64, graph *api.GraphInfo) {
	if mid > 0 {
		return
	}
	if graph.GuestOverwriteRegionalVars > 0 {
		return
	}
	// 目前仅游客可能跳过覆盖逻辑
	for _, v := range vars {
		v.SkipOverwrite = 1
	}

}

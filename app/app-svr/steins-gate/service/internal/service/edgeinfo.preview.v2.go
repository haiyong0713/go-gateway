package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) EdgeInfoV2Preview(c context.Context, params *model.EdgeInfoV2PreReq, mid int64, graphInfo *api.GraphInfo) (res *model.EdgeInfoV2PreReply, err error) {
	var (
		edgeArr, cursorArr        []int64
		nodes                     = make(map[int64]*api.GraphNode)
		edges                     = make(map[int64]*api.GraphEdge)
		nodeInfo                  *api.GraphNode
		edgeInfo                  *api.GraphEdge
		questions                 []*model.Question
		oriEdges                  []*api.GraphEdge
		inChoice, inCursorChoices string // 处理后要写入的最新存档信息，用于展示story_list
		hvarReq                   *model.HvarReq
		hvarRec                   *model.HiddenVarsRecord
		nodeID, currentCursor     int64
		groups                    map[int64]*api.EdgeGroup
		attributes                []string
		view                      *model.VideoUpView
	)
	if _, err = s.arcUpAuth(c, params.AID, mid); err != nil {
		return
	}
	if !model.IsRootEdge(params.EdgeID) {
		if edgeInfo, err = s.pickEdge(c, params.EdgeID, graphInfo); err != nil { // 保证了edgeInfo一定不可能为nil
			log.Error("s.dao.Edge err(%v)", err)
			return
		}
		nodeID = edgeInfo.ToNode
	} else {
		nodeID = graphInfo.FirstNid
		params.Portal = 1 // 有存档时请求根节点视为回溯，无存档时会走无存档逻辑
		edgeInfo = model.GetFirstEdge(graphInfo)
		params.RootEdgeCompatibility()
	}
	if nodeInfo, err = s.pickNode(c, nodeID); err != nil {
		return
	}
	hasHvar, _, randomV := s.pickRandom(graphInfo)
	eg := errgroup.WithContext(c) // 获取当前节点的选项 + 写入存档
	eg.Go(func(c context.Context) (err error) {
		view, err = s.arcDao.VideoUpView(c, params.AID)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		oriEdges, err = s.dao.EdgesByFromNode(c, nodeID)
		return
	})
	eg.Go(func(c context.Context) (err error) { // write preview record
		if inChoice, inCursorChoices, currentCursor, hvarReq, err = s.recordDao.WriteRecordPreview(c,
			mid, params.EdgeID, graphInfo.Id, params.AID, model.RootEdge, params.Cursor, params.Portal, model.NewEdgeRecord, model.PressEdge, model.FromRecEdge); err != nil {
			log.Error("s.dao.WriteRecord err(%v) aid(%d)", err, params.AID)
			return
		}
		if cursorArr, err = xstr.SplitInts(inCursorChoices); err != nil {
			return
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	eg = errgroup.WithContext(c) // fromNode依赖查存档，查到fromNode后查询上一条边和操作隐藏变量存档
	eg.Go(func(c context.Context) (err error) {
		if edgeArr, err = xstr.SplitInts(inChoice); err != nil {
			return
		}
		if edges, err = s.dao.Edges(c, edgeArr, graphInfo); err != nil {
			log.Error("s.dao.Edge err(%v)", err)
			return
		}
		var toNodeIds []int64
		for _, item := range edges {
			toNodeIds = append(toNodeIds, item.ToNode)
		}
		toNodeIds = append(toNodeIds, nodeID)
		nodes, err = s.dao.Nodes(c, toNodeIds)
		return
	})
	if hasHvar { // 如果没有隐藏变量，则无需判断
		eg.Go(func(c context.Context) (err error) {
			if hvarRec, err = s.hvaDao.HiddenVars(c, mid, graphInfo, hvarReq, ""); err != nil {
				log.Error("Mid %d Aid %d Gid %d, HiddenVars Err %v", mid, params.AID, graphInfo.Id, err)
				return
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	res = new(model.EdgeInfoV2PreReply)
	res.EdgeInfoV2 = new(model.EdgeInfoV2)
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
	if questions, res.IsLeaf, err = model.BuildQuestions(oriEdges, groups, model.VideoUpViewPicker(view), nodeInfo, graphInfo, nil); err != nil {
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
			s.hvaDao.AddHiddenVarRecDBCache(c, mid, graphInfo.Id, params.EdgeID, currentCursor, "", hvarRec, false) // set the new record into cache
		}
		res.HiddenVarsPreview = model.DisplayHvars(hvarRec, randomV)
	}
	res.EdgeInfoV2 = new(model.EdgeInfoV2)
	res.StoryList = model.GenEdgeStoryList(nil, s.c.Host.Bfs, currentCursor, edgeArr, cursorArr, nodes, edges)
	res.EdgeInfoV2.BuildEdgeV2Reply(nodeInfo, edgeInfo, questions, s.skinInfo(graphInfo, nodeInfo))
	return

}

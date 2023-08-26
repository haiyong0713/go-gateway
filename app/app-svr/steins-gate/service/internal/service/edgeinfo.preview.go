package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) EdgeInfoPreview(c context.Context, params *model.NodeinfoPreReq, mid int64, graphInfo *api.GraphInfo) (res *model.NodeInfoPreReply, err error) {
	var (
		edgeArr, cursorArr          []int64
		nodes                       = make(map[int64]*api.GraphNode)
		edges                       = make(map[int64]*api.GraphEdge)
		edgeInfo                    *api.GraphEdge
		oriChoices, filteredChoices []*api.GraphEdge
		showChoices                 []*model.Choice
		inChoice, inCursorChoices   string // 处理后要写入的最新存档信息，用于展示story_list
		hvarReq                     *model.HvarReq
		hvarRec                     *model.HiddenVarsRecord
		nodeID, currentCursor       int64
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
	hasHvar, _, randomV := s.pickRandom(graphInfo)
	eg := errgroup.WithContext(c) // 获取当前节点的选项 + 写入存档
	eg.Go(func(c context.Context) (err error) {
		oriChoices, err = s.dao.EdgesByFromNode(c, nodeID)
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
	res = new(model.NodeInfoPreReply)
	if hasHvar {
		if params.Portal != 1 { // 回溯操作不修改隐藏变量存档
			//nolint:errcheck
			hvarRec.ApplyAttr(edgeInfo.Attribute) // edgeInfo不会为空，所以不判断
			//nolint:errcheck
			s.hvaDao.AddHiddenVarRecDBCache(c, mid, graphInfo.Id, params.EdgeID, currentCursor, "", hvarRec, false) // set the new record into cache
		}
		if filteredChoices, err = model.FilterEdges(hvarRec, oriChoices, randomV); err != nil { // 条件过滤
			log.Error("Mid %d Aid %d Gid %d, FilterEdges Err %v", mid, params.AID, graphInfo.Id, err)
			return
		}
		res.HiddenVarsPreview = model.DisplayHvars(hvarRec, randomV)
	} else { // no need to filter choices
		filteredChoices = oriChoices
	}
	showChoices = model.BuildChoicesEdge(filteredChoices, s.skinInfo(graphInfo, nodes[nodeID]))
	res.NodeInfo = new(model.NodeInfo)
	res.StoryList = model.GenEdgeStoryList(nil, s.c.Host.Bfs, currentCursor, edgeArr, cursorArr, nodes, edges)
	res.NodeInfo.BuildEdgeReply(nodes[nodeID], edgeInfo, showChoices, nil, res.HiddenVarsPreview)
	return

}

package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	"go-common/library/sync/errgroup.v2"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	xecode "go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// edgeInfoProcess edgeinfo/edgeinfo_v2共用逻辑
//
//nolint:gocognit
func (s *Service) edgeInfoProcess(c context.Context, params *model.NodeInfoParam, mid int64, graphInfo *api.GraphInfo, choiceIDs []int64) (
	edges map[int64]*api.GraphEdge, oriEdges []*api.GraphEdge, nodeInfo *api.GraphNode, edgeInfo *api.GraphEdge,
	hasHvar bool, hvarRec *model.HiddenVarsRecord, randomV *model.RandomVar, currentCursor int64, arcView *arcgrpc.SteinsGateViewReply,
	storyList []*model.Story, edgeAnms map[int64]*api.EdgeFrameAnimations, err error) {
	var (
		inChoice, inCursorChoices string
		hvarReq                   *model.HvarReq
		nodeID, fromEdge          int64
		lastRec                   *api.GameRecords
		edgeArr, cursorArr        []int64
		nodes                     map[int64]*api.GraphNode
	)
	if !model.IsRootEdge(params.EdgeID) {
		if edgeInfo, err = s.pickEdge(c, params.EdgeID, graphInfo); err != nil {
			log.Error("s.dao.Edge err(%v)", err)
			return
		}
		nodeID = edgeInfo.ToNode
		if params.Cursor == model.RootCursor {
			params.Cursor = model.IllegalCursor
		}
	} else {
		nodeID = graphInfo.FirstNid
		params.Portal = 1 // 有存档时请求根节点视为回溯，无存档时会走无存档逻辑
		edgeInfo = model.GetFirstEdge(graphInfo)
		params.RootEdgeCompatibility()
	}
	if nodeInfo, err = s.pickNode(c, nodeID); err != nil {
		return
	}
	hasHvar, hvarRec, randomV = s.pickRandom(graphInfo) // 这里出一个默认的初始值的存档用于未登录的情况
	if nodeInfo.GraphId != graphInfo.Id {               // 校验node属于当前graph，避免写入脏数据进入record
		err = ecode.RequestErr
		log.Error("Node %d doesn't belong Graph %d, %d instead", nodeInfo.Id, graphInfo.Id, nodeInfo.GraphId)
		return
	}
	eg := errgroup.WithContext(c) // 获取选项
	eg.Go(func(c context.Context) (err error) {
		arcView, err = s.arcDao.ArcView(c, params.AID)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		oriEdges, err = s.dao.EdgesByFromNode(c, nodeID)
		return
	})
	if mid > 0 { // 登录用户记录存档并且将最新存档返回到story_list中
		eg.Go(func(c context.Context) (err error) {
			cacheOnly := graphInfo.NoBacktracking > int32(0)
			if cacheOnly {
				//nolint:govet
				log.Info("Skip to write record to DB: %d", graphInfo)
			}
			if !(cacheOnly && model.IsRootEdge(params.EdgeID)) {
				if lastRec, err = s.recordDao.GetLastRecord(c, params, params.EdgeID, mid, model.PullEdge); err != nil {
					log.Error("s.dao.GetLastRecord err(%v) aid(%d)", err, params.AID)
					return
				}
			}
			if inChoice, inCursorChoices, fromEdge, currentCursor, hvarReq, err = s.recordDao.WriteRecord(c,
				params, mid, params.EdgeID, model.RootEdge, model.NewEdgeRecord, model.PressEdge,
				model.FromRecEdge, lastRec, cacheOnly); err != nil {
				log.Error("s.dao.WriteRecord err(%v) aid(%d)", err, params.AID)
				return
			}
			if edgeArr, err = xstr.SplitInts(inChoice); err != nil {
				return
			}
			cursorArr, err = xstr.SplitInts(inCursorChoices)
			return
		})
	} else { // 未登录用户返回第1P
		edgeArr = []int64{model.RootEdge}
		cursorArr = []int64{model.RootCursor}
	}
	if err = eg.Wait(); err != nil {
		return
	}
	eg = errgroup.WithContext(c)                                  // fromEdge依赖查存档，查到fromEdge后查询上一条边和操作隐藏变量存档
	if arcView.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes { // ogv视频依赖grpc判断是否可播
		eg.Go(func(c context.Context) (err error) {
			var allowPlay bool
			if allowPlay, err = s.ogvDao.OgvPay(c, mid, graphInfo.Aid); err != nil {
				log.Error("allowPlay %v", err)
				return
			}
			if !allowPlay {
				err = xecode.GraphOgvNotAllowed
				log.Warn("aid %d mid %d not allowed", mid, graphInfo.Aid)
			}
			return
		})
	}
	eg.Go(func(c context.Context) (err error) {
		edgeIDs := edgeArr
		if len(choiceIDs) > 0 {
			edgeIDs = append(edgeIDs, choiceIDs...)
		}
		if edges, err = s.dao.Edges(c, edgeIDs, graphInfo); err != nil { // edgeArr含有0时取graphInfo拼接为根节点数据
			log.Error("s.dao.Edge err(%v)", err)
			return
		}
		oriEdgeIds := make([]int64, 0, len(oriEdges))
		for _, i := range oriEdges {
			oriEdgeIds = append(oriEdgeIds, i.Id)
		}
		if edgeAnms, err = s.dao.EdgeFrameAnimations(c, oriEdgeIds); err != nil {
			log.Error("s.dao.EdgeFrameAnimations err(%v)", err)
			return
		}
		var toNodeIds []int64
		for _, item := range edges {
			toNodeIds = append(toNodeIds, item.ToNode)
		}
		nodes, err = s.dao.Nodes(c, toNodeIds)
		return
	})
	if hasHvar && mid > 0 { // 如果没有隐藏变量或者未登录，则无需判断
		eg.Go(func(c context.Context) (err error) {
			if hvarRec, err = s.hvaDao.HiddenVars(c, mid, graphInfo, hvarReq, params.Buvid); err != nil {
				log.Error("Mid %d Aid %d Gid %d, HiddenVars Err %v", mid, params.AID, graphInfo.Id, err)
				return
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	arcPage := model.PickPagesMap(arcView)
	nodes[nodeInfo.Id] = nodeInfo // 未登录下，可能当前node不在列表里
	edges[edgeInfo.Id] = edgeInfo // 未登录下，可能当前edge不在列表里
	storyList = model.GenEdgeStoryList(arcPage, s.c.Host.Bfs, currentCursor, edgeArr, cursorArr, nodes, edges)
	s.infocAction(params, mid, params.EdgeID, fromEdge, nodeInfo.Otype)
	return

}

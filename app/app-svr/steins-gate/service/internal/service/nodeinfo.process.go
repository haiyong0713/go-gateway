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

// nodeinfoProcess 为nodeinfo和nodeinfoV2公用的核心方法
func (s *Service) nodeInfoProcess(c context.Context, params *model.NodeInfoParam, mid int64, graphInfo *api.GraphInfo) (
	nodeInfo *api.GraphNode, oriChoices []*api.GraphEdge, arcView *arcgrpc.SteinsGateViewReply,
	hasHvar bool, hvarRec *model.HiddenVarsRecord,
	randomV *model.RandomVar, storyList []*model.Story, err error) {
	var (
		inChoice, inCursorChoices string
		hvarReq                   *model.HvarReq
		lastRec                   *api.GameRecords
		fromNode, currentCursor   int64
		nodeArr, cursorArr        []int64
		nodes                     map[int64]*api.GraphNode
		walkedEdge                *api.GraphEdge
	)
	nodeID := params.NodeID // node info picking
	if nodeID == 0 {
		nodeID = graphInfo.FirstNid
		params.Portal = 1 // 有存档时请求根节点视为回溯，无存档时会走无存档逻辑
	}
	if nodeID != graphInfo.FirstNid && params.Cursor == model.RootCursor {
		params.Cursor = model.IllegalCursor
	}
	hasHvar, hvarRec, randomV = s.pickRandom(graphInfo) // 这里出一个默认的初始值的存档用于未登录的情况
	// 获取选项
	if nodeInfo, err = s.pickNode(c, nodeID); err != nil {
		return
	}
	if nodeInfo.GraphId != graphInfo.Id { // 校验node属于当前graph
		err = ecode.RequestErr
		log.Error("Node %d doesn't belong Graph %d, %d instead", nodeInfo.Id, graphInfo.Id, nodeInfo.GraphId)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		arcView, err = s.arcDao.ArcView(c, params.AID)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		oriChoices, err = s.dao.EdgesByFromNode(c, nodeID)
		return
	})
	if mid > 0 { // 登录用户记录存档并且将最新存档返回到story_list中
		eg.Go(func(c context.Context) (err error) {
			if lastRec, err = s.recordDao.GetLastRecord(c, params, nodeID, mid, model.PullNode); err != nil {
				log.Error("s.dao.GetLastRecord err(%v) aid(%d)", err, params.AID)
				return
			}
			if inChoice, inCursorChoices, fromNode, currentCursor, hvarReq, err = s.recordDao.WriteRecord(c,
				params, mid, nodeID, graphInfo.FirstNid, model.NewNodeRecord, model.PressNode, model.FromRecNode, lastRec, false); err != nil {
				log.Error("s.dao.WriteRecord err(%v) aid(%d)", err, params.AID)
				return
			}
			if nodeArr, err = xstr.SplitInts(inChoice); err != nil {
				return
			}
			cursorArr, err = xstr.SplitInts(inCursorChoices)
			return
		})
	} else { // 未登录用户返回第1P
		nodeArr = []int64{graphInfo.FirstNid}
		cursorArr = []int64{model.RootCursor}
	}
	if err = eg.Wait(); err != nil {
		return
	}
	eg = errgroup.WithContext(c)                                  // fromNode依赖查存档，查到fromNode后查询上一条边和操作隐藏变量存档
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
		nodes, err = s.dao.Nodes(c, nodeArr)
		return
	})
	if hasHvar && mid > 0 { // 如果没有隐藏变量或者未登录，则无需判断
		if params.Portal != 1 { // 回溯不需要计算,portal=2/0计算
			eg.Go(func(c context.Context) (err error) {
				walkedEdge, err = s.passingEdge(c, hvarReq.CurrentID, nodeID)
				return
			})
		}
		eg.Go(func(c context.Context) (err error) {
			if hvarRec, err = s.hvaDao.HiddenVars(c, mid, graphInfo, hvarReq, params.Buvid); err != nil {
				log.Error("Mid %d Aid %d Gid %d, HiddenVars Err %v", mid, params.AID, graphInfo.Id, err)
				return
			}
			return
		})
	}
	err = eg.Wait()
	arcPage := model.PickPagesMap(arcView)
	storyList = model.GenStoryList(arcPage, s.c.Host.Bfs, currentCursor, nodeArr, cursorArr, nodes)
	if hasHvar && mid > 0 && params.Portal != 1 { // 回溯操作不修改隐藏变量存档
		if walkedEdge != nil {
			//nolint:errcheck
			hvarRec.ApplyAttr(walkedEdge.Attribute)
		}
		//nolint:errcheck
		s.hvaDao.AddHiddenVarRecDBCache(c, mid, graphInfo.Id, nodeID, currentCursor, params.Buvid, hvarRec, false) // set the new record into cache
	}
	// infoc 逻辑
	s.infocAction(params, mid, nodeID, fromNode, nodeInfo.Otype)
	return

}

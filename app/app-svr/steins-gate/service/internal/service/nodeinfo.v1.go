package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const (
	_buvidRand = 10000
)

func (s *Service) pickRandom(graphInfo *api.GraphInfo) (hasHvar bool, hvarRec *model.HiddenVarsRecord, rand *model.RandomVar) {
	if len(graphInfo.RegionalVars) == 0 {
		return
	}
	var (
		variables map[string]*model.RegionalVal
		err       error
	)
	if variables, err = model.GetVarsMap(graphInfo); err != nil { // 如果当前图并无隐藏变量，无需继续回源
		log.Error("GraphID %d GetVarsMap Err %v", graphInfo.Id, err)
		return
	}
	if len(variables) > 0 {
		hasHvar = true
		hvarRec = new(model.HiddenVarsRecord)
		hvarRec.Vars = make(map[string]*model.HiddenVar)
	}
	for _, v := range variables { // 如果当前图有随机变量，则在区间范围内产生一个随机值
		if v.Type == model.RegionalVarTypeRandom {
			rand = &model.RandomVar{
				Name:  v.Name,
				ID:    v.ID,
				Value: model.RandInt(s.rand, v.InitMin, v.InitMax),
			}
		} else { // 准备一个初始值存档，用于未登录状态
			hvar := new(model.HiddenVar)
			hvar.FromRegionalVar(v)
			hvarRec.Vars[v.ID] = hvar
		}
	}
	return
}

func (s *Service) GenBuvid(mid, aid int64) string {
	return fmt.Sprintf("%d%d%d%d", mid, aid, time.Now().UnixNano(), s.rand.Int63n(_buvidRand))
}

func (s *Service) passingEdge(c context.Context, fromNID, toNID int64) (edge *api.GraphEdge, err error) {
	var edges []*api.GraphEdge
	if edges, err = s.dao.EdgesByFromNode(c, fromNID); err != nil {
		log.Error("s.dao.EdgesByFromNode err(%v) nodeID(%d)", err, fromNID)
		return
	}
	for _, v := range edges { // find the walked edge from last point
		if v.ToNode == toNID {
			edge = v
		}
	}
	return
}

func (s *Service) NodeInfo(c context.Context, params *model.NodeInfoParam, mid int64, graphInfo *api.GraphInfo) (res *model.NodeInfo, err error) {
	var (
		nodeInfo                    *api.GraphNode
		showChoices                 []*model.Choice
		oriChoices, filteredChoices []*api.GraphEdge
		arcView                     *arcgrpc.SteinsGateViewReply
		hiddenVars                  []*model.HiddenVar
		hasHvar                     bool
		hvarRec                     *model.HiddenVarsRecord
		randomV                     *model.RandomVar
	)
	res = &model.NodeInfo{
		Preload: &model.Preload{Video: make([]*model.PreVideo, 0)},
	}
	if nodeInfo, oriChoices, arcView, hasHvar, hvarRec, randomV, res.StoryList, err = s.nodeInfoProcess(c, params, mid, graphInfo); err != nil {
		return
	}
	if hasHvar {
		if filteredChoices, err = model.FilterEdges(hvarRec, oriChoices, randomV); err != nil { // 条件过滤
			log.Error("Mid %d Aid %d Gid %d, FilterEdges Err %v", mid, params.AID, graphInfo.Id, err)
			return
		}
		hiddenVars = model.DisplayHvars(hvarRec, randomV)
	} else { // no need to filter choices
		filteredChoices = oriChoices
	}
	showChoices = model.BuildChoicesNode(filteredChoices, s.skinInfo(graphInfo, nodeInfo))
	// 结果拼接逻辑
	res.BuildReply(nodeInfo, showChoices, arcView, hiddenVars)
	res.Preload.FromNode(nodeInfo, showChoices, params.AID)
	return
}

func (s *Service) pickNode(c context.Context, nodeID int64) (nodeInfo *api.GraphNode, err error) {
	if nodeInfo, err = s.dao.Node(c, nodeID); err != nil {
		log.Error("s.dao.Node err(%v) nodeID(%d)", err, nodeID)
		return
	}
	if nodeInfo == nil {
		err = ecode.NothingFound
	}
	return

}

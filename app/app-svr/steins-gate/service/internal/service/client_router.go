package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// RouterInfoPreview 预览逻辑，兼容客户端，根据graph类型判定走nodeinfo还是edgeinfo
func (s *Service) RouterInfoPreview(c context.Context, params *model.NodeinfoPreReq, mid int64) (res *model.NodeInfoPreReply, err error) {
	var graphInfo *api.GraphInfo
	if graphInfo, err = s.dao.GraphInfoPreview(c, params.AID); err != nil {
		log.Error("NodeInfoPreview s.dao.GraphInfoPreview aid(%d) error(%v)", params.AID, err)
		return
	}
	if model.IsEdgeGraph(graphInfo) { // edge剧情图走edgeInfo逻辑
		params.ClientCompatibility() // 兼容客户端
		res, err = s.EdgeInfoPreview(c, params, mid, graphInfo)
		return
	}
	log.Error("NodePreviewGraph Hit! Mid %d Req %+v", mid, params) // 日志报警
	return
}

// 返回新的V2结构体
func (s *Service) RouterInfoV2Preview(c context.Context, params *model.EdgeInfoV2PreReq, mid int64) (res *model.EdgeInfoV2PreReply, err error) {
	var graphInfo *api.GraphInfo
	if graphInfo, err = s.dao.GraphInfoPreview(c, params.AID); err != nil {
		log.Error("NodeInfoPreview s.dao.GraphInfoPreview aid(%d) error(%v)", params.AID, err)
		return
	}
	if model.IsEdgeGraph(graphInfo) { // edge剧情图走edgeInfo逻辑
		params.ChoiceHandler()
		res, err = s.EdgeInfoV2Preview(c, params, mid, graphInfo)
		return
	}
	return
}

// RouterInfo 兼容客户端逻辑，根据graph类型判定走nodeinfo还是edgeinfo
func (s *Service) RouterInfo(c context.Context, params *model.NodeInfoParam, mid int64) (res *model.NodeInfo, err error) {
	var graphInfo *api.GraphInfo
	if graphInfo, err = s.GraphInfo(c, params.AID); err != nil { // graph info picking
		log.Error("s.GraphInfo err(%v)", err)
		return
	}
	if graphInfo.Id != params.GraphVersion {
		err = ecode.GraphInvalid
		return
	}
	if model.IsEdgeGraph(graphInfo) { // edge剧情图走edgeInfo逻辑
		params.ClientCompatibility() // 兼容客户端
		res, err = s.EdgeInfo(c, params, mid, graphInfo)
		prom.BusinessInfoCount.Incr("EdgeInfo")
		return
	}
	if model.BuildRestrict(graphInfo) {
		err = ecode.GraphRestrictBuildErr
		log.Error("RouterInfoRestrict Param %+v", params)
		return
	}
	res, err = s.NodeInfo(c, params, mid, graphInfo) // node剧情图走nodeInfo逻辑
	prom.BusinessInfoCount.Incr("NodeInfo")
	return
}

// RouterInfoV2 根据graph类型判定走nodeinfo或者edgeinfo还是中插树
func (s *Service) RouterInfoV2(c context.Context, params *model.EdgeInfoV2Param, mid int64) (res *model.EdgeInfoV2, err error) {
	var graphInfo *api.GraphInfo
	if graphInfo, err = s.GraphInfo(c, params.AID); err != nil { // graph info picking
		log.Error("s.GraphInfo err(%v)", err)
		return
	}
	if graphInfo.Id != params.GraphVersion {
		err = ecode.GraphInvalid
		return
	}
	if model.IsEdgeGraph(graphInfo) || model.IsInterventionGraph(graphInfo) { // edge/中插剧情图走edgeInfo逻辑
		params.ChoiceHandler()
		res, err = s.EdgeInfoV2(c, params, mid, graphInfo)
		prom.BusinessInfoCount.Incr("EdgeInfoV2")
		return
	}
	params.ClientCompatibility()
	res, err = s.NodeInfoV2(c, params, mid, graphInfo) // node剧情图走nodeInfo逻辑
	prom.BusinessInfoCount.Incr("NodeInfoV2")
	return

}

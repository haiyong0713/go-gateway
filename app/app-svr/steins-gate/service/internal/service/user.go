package service

import (
	"context"
	"encoding/json"
	"strings"
	"unicode"
	"unicode/utf8"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"

	filterapi "git.bilibili.co/bapis/bapis-go/filter/service"

	"go-gateway/app/app-svr/archive/service/api"
	playurlapi "go-gateway/app/app-svr/playurl/service/api"
	xecode "go-gateway/app/app-svr/steins-gate/ecode"
	steinapi "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/sync/errgroup.v2"
)

const (
	_latestLimit          = 3
	_isStart              = 1
	_isDefault            = 1
	_filterArea           = "steins_gate"
	_filterLevel          = 20
	_nodeLength           = 600
	_nodeNameLength       = 30
	_nodeEdgeLength       = 4
	_nodeEdgeSpecLength   = 1
	_edgeTitleLength      = 80
	_playurlPlatform      = "pc"
	_defaultForceHost     = 2
	_videoDispatchFinish  = 6
	_graphRegionalVarLen  = 5
	_graphRegVarNormalLen = 4
	_graphRegVarRandomLen = 1
	_graphRegVarNameLen   = 20
	_normalInitMinBottom  = -100
	_normalInitMinTop     = 100
	_randomInitMinDft     = 1
	_randomInitMaxDft     = 100
	_edgeCondLen          = 10
	_edgeAttrLen          = 4
	_idStrLen             = 15
	_edgeOneVarCondLen    = 2
	_maxScriptLen         = 16777215
	_maxQn                = 112
	_defaultFnVer         = 0
	_defaultFnVal         = 1
	_cancelReason         = "用户提交新剧情图"
	_diffMsgPrefix        = "新增"
	_diffMsgMax           = 1000
	_errTypeNode          = 1
	_errTypeEdge          = 2
	_errTypeVars          = 3
)

func (s *Service) VideoInfo(c context.Context, param *model.VideoInfoParam, mid int64) (data *model.VideoInfo, err error) {
	var (
		playurl   *model.PlayurlRes
		dimension *model.Dimension
	)
	if err = s.videoUpAuth(c, param.Aid, param.Cid, mid); err != nil {
		return
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(c context.Context) (err error) { // get playurl info
		var reply *playurlapi.SteinsPreviewReply
		playurlArg := &playurlapi.SteinsPreviewReq{
			Aid:       param.Aid,
			Cid:       param.Cid,
			Qn:        _maxQn,
			Platform:  _playurlPlatform,
			Fnver:     _defaultFnVer,
			Fnval:     _defaultFnVal,
			Buvid:     param.Buvid,
			ForceHost: _defaultForceHost,
			Mid:       mid,
		}
		if reply, err = s.arcDao.Playurl(c, playurlArg); err != nil {
			log.Error("VideoInfo s.dao.Playurl(%+v) error(%v)", playurlArg, err)
			return
		}
		playurl = new(model.PlayurlRes)
		playurl.FromPlayurl(reply)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		var resp *model.DimensionInfo
		if resp, err = s.dao.BvcDimension(c, param.Cid); err != nil {
			log.Error("VideoInfo Dimension Aid %d, Cid %d Mid %d, err %+v", param.Aid, param.Cid, mid, err)
			err = xecode.GraphGetDimensionErr
			return
		}
		if resp != nil {
			dimension = new(model.Dimension)
			dimension.FromInfo(resp)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	data = &model.VideoInfo{
		Playurl:   playurl,
		Dimension: dimension,
	}
	return
}

// LatestGraphList .
func (s *Service) LatestGraphList(c context.Context, mid, aid int64) (list []*model.Graph, err error) {
	if _, err = s.arcUpAuth(c, aid, mid); err != nil {
		return
	}
	return s.dao.LatestGraphList(c, aid, _latestLimit)
}

// GraphShow .
func (s *Service) GraphShow(c context.Context, mid, aid, graphID int64) (data *model.GraphShow, err error) {
	var (
		graph *model.Graph
		nodes []*steinapi.GraphNode
	)
	if graph, err = s.dao.GraphShow(c, aid, graphID); err != nil {
		return
	}
	if nodes, err = s.dao.GraphNodeList(c, graph.ID); err != nil {
		return
	}
	var view *model.VideoUpView
	if view, err = s.arcDao.VideoUpView(c, aid); err != nil {
		log.Error("LatestGraphList s.dao.VideoUpView aid(%d) error(%v)", aid, err)
		return
	}
	if !s.isShowWhite(mid) {
		if err = checkInteractiveArchive(mid, view.Archive.Mid, view.Archive.Attribute); err != nil {
			return
		}
	}
	viewCidMap := make(map[int64]struct{}, len(view.Videos))
	data = &model.GraphShow{Graph: graph, DisabledCids: []int64{}}
	for _, v := range view.Videos {
		if v == nil {
			log.Error("VideoUpView Video Aid %d Nil", aid)
			continue
		}
		viewCidMap[v.Cid] = struct{}{}
	}
	for _, node := range nodes {
		if _, ok := viewCidMap[node.Cid]; !ok {
			data.DisabledCids = append(data.DisabledCids, node.Cid)
		}
	}
	return
}

// Playurl .
func (s *Service) Playurl(c context.Context, mid int64, arg *model.PlayurlParam) (data *model.PlayurlRes, err error) {
	var reply *playurlapi.SteinsPreviewReply
	if err = s.videoUpAuth(c, arg.Aid, arg.Cid, mid); err != nil {
		return
	}
	playurlArg := &playurlapi.SteinsPreviewReq{
		Aid:       arg.Aid,
		Cid:       arg.Cid,
		Qn:        arg.Qn,
		Platform:  _playurlPlatform,
		Fnver:     arg.Fnver,
		Fnval:     arg.Fnval,
		Buvid:     arg.Buvid,
		ForceHost: _defaultForceHost,
		Mid:       mid,
	}
	if reply, err = s.arcDao.Playurl(c, playurlArg); err != nil {
		log.Error("Playurl s.dao.Playurl(%+v) error(%v)", playurlArg, err)
		return
	}
	data = new(model.PlayurlRes)
	data.FromPlayurl(reply)
	return
}

// MsgCheck .
func (s *Service) MsgCheck(c context.Context, msg string) (err error) {
	var reply *filterapi.FilterReply
	if unicodeNameLen(msg) > _edgeTitleLength {
		err = ecode.RequestErr
		return
	}
	if reply, err = s.auditDao.FilterMsg(c, _filterArea, msg); err != nil {
		log.Error("s.dao.FilterMsg(%s) error(%v)", msg, err)
		err = xecode.GraphFilterErr
		return
	}
	if reply.Level >= _filterLevel {
		err = xecode.GraphFilterHitErr
	}
	return
}

// SaveGraph save node data.
func (s *Service) SaveGraph(c context.Context, mid int64, isPreview int, param *model.SaveGraphParam) (graphID int64, errInfo model.ErrInfo, err error) {
	var (
		dimensions map[int64]*model.DimensionInfo
		view       *model.VideoUpView
		diffMsg    string
		isAudit    bool
	)
	if dimensions, view, errInfo, err = s.checkGraphParam(c, mid, isPreview, param); err != nil {
		return
	}
	if param.Graph.State, diffMsg, err = s.graphDiff(c, param, isPreview); err != nil {
		return
	}
	if _, ok := model.GraphStateAudits[param.Graph.State]; ok {
		isAudit = true
	}
	if graphID, err = s.dao.SaveGraph(c, isPreview, isAudit, param, dimensions, mid); err != nil {
		return
	}
	if isAudit && graphID > 0 {
		//nolint:errcheck
		s.auditDao.AddAegisMsg(c, graphID, mid, param.Graph.Aid,
			int64(param.Graph.State), view.Archive.Title, diffMsg)
	}
	return
}

// GraphCheck .
func (s *Service) GraphCheck(c context.Context, aid, cid int64) (check *model.GraphCheck, err error) {
	var (
		graph *steinapi.GraphInfo
		nodes []*steinapi.GraphNode
	)
	check = new(model.GraphCheck)
	if graph, err = s.dao.GraphInfo(c, aid); err != nil {
		return
	}
	if graph == nil {
		log.Warn("GraphCheck aid(%d) graph is nil", aid)
		return
	}
	if nodes, err = s.dao.GraphNodeList(c, graph.Id); err != nil {
		return
	}
	for _, node := range nodes {
		if node.Cid == cid {
			check.HasCid = true
			break
		}
	}
	return
}

func (s *Service) graphDiff(c context.Context, param *model.SaveGraphParam, isPreview int) (state int, diffMsg string, err error) {
	var (
		preGraph   *model.GraphAuditDB
		noDelGraph *model.GraphAuditDB
		existPass  bool
	)
	if isPreview == model.GraphIsPreview {
		state = model.GraphStatePass
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(c context.Context) (err error) {
		if preGraph, err = s.auditDao.GraphAuditByAid(c, param.Graph.Aid); err != nil {
			log.Error("graphDiff s.dao.GraphAuditByAid(%d) error(%v)", param.Graph.Aid, err)
			return
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		if noDelGraph, err = s.auditDao.NoDelGraphAuditByAid(c, param.Graph.Aid); err != nil {
			log.Error("graphDiff s.dao.GraphAuditByAid(%d) error(%v)", param.Graph.Aid, err)
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		if existPass, err = s.dao.ExistGraph(c, param.Graph.Aid); err != nil {
			log.Error("graphDiff s.dao.ExistGraph(%d) error(%v)", param.Graph.Aid, err)
			return
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if preGraph == nil {
		state = model.GraphStateSubmit
		return
	}
	if preGraph.State == model.GraphStateSubmit || preGraph.State == model.GraphStateReSubmit { // 上个剧情未过审，需要取消上个剧情图的审核
		//nolint:errcheck
		s.auditDao.CancelAegisMsg(c, preGraph.Id, _cancelReason)
	}
	if noDelGraph == nil {
		state = model.GraphStateSubmit
		return
	}
	// 上个剧情被打回或者已经过审，需要对比修改项
	var (
		nodeNameMap map[string]*steinapi.GraphNode
		edgeNameMap map[string]*steinapi.GraphEdge
		nodes       []*steinapi.GraphNode
		edges       []*steinapi.GraphEdge
		edgeCnt     int
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (e error) {
		if nodes, e = s.dao.GraphNodeList(c, noDelGraph.Id, true); e != nil {
			log.Error("graphDiff s.dao.GraphNodeList(%d) error(%v)", noDelGraph.Id, e)
		}
		return e
	})
	group.Go(func(ctx context.Context) (e error) {
		if edges, e = s.dao.GraphEdgeList(c, noDelGraph.Id, true); e != nil {
			log.Error("graphDiff s.dao.GraphEdgeList(%d) error(%v)", noDelGraph.Id, e)
		}
		return e
	})
	if err = group.Wait(); err != nil {
		return
	}
	nodeNameMap = make(map[string]*steinapi.GraphNode, len(nodes))
	for _, v := range nodes {
		nodeNameMap[v.Name] = v
	}
	edgeNameMap = make(map[string]*steinapi.GraphEdge, len(edges))
	for _, v := range edges {
		edgeNameMap[v.Title] = v
	}
	var newNodeNames, newEdgeNames []string
	for _, v := range param.Graph.Nodes {
		if _, ok := nodeNameMap[v.Name]; !ok {
			newNodeNames = append(newNodeNames, v.Name)
		}
		for _, edge := range v.Edges {
			edgeCnt++
			if _, ok := edgeNameMap[edge.Title]; !ok {
				newEdgeNames = append(newEdgeNames, edge.Title)
			}
		}
	}
	hasNoVarsDiff := noDelGraph.HasNoDiffVarsName(param.Graph.RegionalVars)
	noNew := len(newNodeNames) == 0 && len(newEdgeNames) == 0
	noLess := len(param.Graph.Nodes) <= len(nodes) && edgeCnt <= len(edges)
	// 上个剧情打回，没有修改禁止提交
	if noDelGraph.State == model.GraphStateRepulse && noNew && noLess && hasNoVarsDiff {
		err = ecode.NotModified
		return
	}
	// 上个剧情图过审，这次没有新增节点和新增选项，不用审核
	if noDelGraph.State == model.GraphStatePass && noNew && hasNoVarsDiff {
		state = model.GraphStatePass
		return
	}
	diffMsg = s.getDiffMsg(newNodeNames, newEdgeNames, param.Graph.VarsNames())
	log.Warn("Metadata Aid %d, GraphID %d, newNodeNames %v, newEdgeNames %v, VarsNames %v, len %d", param.Graph.Aid, param.Graph.ID, newNodeNames, newEdgeNames, param.Graph.VarsNames(), utf8.RuneCountInString(diffMsg))
	if existPass { // 如果存在已经过审的剧情树状态为-10，否则状态默认为-30
		state = model.GraphStateReSubmit
	} else {
		state = model.GraphStateSubmit
	}
	return
}

func (s *Service) getDiffMsg(newNodeNames, newEdgeNames []string, varsNames string) (metaMsg string) {
	diffMsg := _diffMsgPrefix
	diffMsg = getDiffMsg(newNodeNames, diffMsg, " 剧情名称", varsNames)
	diffMsg = getDiffMsg(newEdgeNames, diffMsg, " 选项名称", varsNames)
	metaData, _ := json.Marshal(&model.AegisMetaData{DiffMsg: diffMsg, VarsName: varsNames})
	return string(metaData)
}

func getDiffMsg(names []string, originMsg, msgPrefix, varsNames string) (diffMsg string) {
	diffMsg = originMsg
	if len(names) == 0 {
		return
	}
	var index = 0
	for i := 1; i < len(names)+1 && model.AegisMetaLen(diffMsg, varsNames) < _diffMsgMax; i++ {
		if nameMsg := diffMsg + msgPrefix + ":【" + strings.Join(names[0:i], "、") + "】"; model.AegisMetaLen(nameMsg, varsNames) > _diffMsgMax {
			index = i - 1
			break
		}
		if i == len(names) { // 最后都没超过，直接加上所有的
			index = i
		}
	}
	if index > 1 {
		diffMsg += msgPrefix + ":【" + strings.Join(names[0:index], "、") + "】"
	}
	return
}

func (s *Service) checkGraphParam(c context.Context, mid int64, isPreview int, param *model.SaveGraphParam) (dimensions map[int64]*model.DimensionInfo, view *model.VideoUpView, errInfo model.ErrInfo, err error) {
	var (
		filterReply  *filterapi.MFilterReply
		regionVarMap map[string]*model.RegionalVal
		nodeParamMap map[string]*model.NodeParam
		bs           []byte
	)
	if err = s.checkGraph(param.Graph, isPreview); err != nil {
		return
	}
	filterMsgs := make(map[string]string)
	// regional vars check
	if regionVarMap, bs, errInfo, err = checkGraphRegVars(param.Graph.RegionalVars, filterMsgs); err != nil {
		return
	}
	param.Graph.RegionalStr = string(bs)
	// arc auth
	if view, err = s.arcUpAuth(c, param.Graph.Aid, mid); err != nil {
		return
	}
	// arc's videos' dimension check
	if dimensions, errInfo, err = s.checkDimension(c, view, param.Graph.Nodes); err != nil {
		log.Error("mid %d aid %d err %v", mid, param.Graph.Aid, err)
		return
	}
	// node check
	if nodeParamMap, errInfo, err = s.checkNodes(param.Graph.Nodes, filterMsgs); err != nil {
		return
	}
	// edge check
	if errInfo, err = checkEdges(param.Graph.Nodes, nodeParamMap, regionVarMap, filterMsgs); err != nil {
		return
	}
	// all titles filter check
	if filterReply, err = s.auditDao.MFilterMsg(c, _filterArea, filterMsgs); err != nil {
		log.Error("s.dao.MFilterMsg(%+v) error(%v)", filterMsgs, err)
		err = xecode.GraphFilterErr
		return
	}
	for _, v := range filterReply.RMap {
		if v.Level >= _filterLevel {
			err = xecode.GraphFilterHitErr
			return
		}
	}
	return
}

// checkDimension picks the dimension of the cids to find out whether there is a vertical video or not
func (s *Service) checkDimension(c context.Context, view *model.VideoUpView, nodes []*model.NodeParam) (dimensions map[int64]*model.DimensionInfo, errInfo model.ErrInfo, err error) {
	var (
		videoInfo      = make(map[int64]*model.Video)
		verticalTitles []string
		cids           []int64
	)
	for _, v := range view.Videos {
		videoInfo[v.Cid] = v
	}
	for _, node := range nodes {
		video, ok := videoInfo[node.Cid]
		if !ok {
			log.Error("VideoInfo Cid %d not found", node.Cid)
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphVideoupArcErr
			return
		}
		if video.XcodeState != _videoDispatchFinish {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphCidNotDispatched
			return
		}
		cids = append(cids, node.Cid)
	}
	if dimensions, err = s.dao.BvcDimensions(c, cids); err != nil {
		log.Error("BvcDimensions Cids %v, Err %+v", cids, err)
		err = xecode.GraphGetDimensionErr
		return
	}
	for _, node := range nodes {
		if dimension, ok := dimensions[node.Cid]; !ok {
			log.Error("Dimension Cid %d, Err %v", node.Cid, err)
			err = xecode.GraphVideoupArcErr // 返回获取数据失败
			return
		} else if dimension.IsVertical() {
			verticalTitles = append(verticalTitles, "【"+videoInfo[node.Cid].Title+"】")
		}
	}
	if len(verticalTitles) > 0 {
		err = ecode.Errorf(xecode.GraphPageWidthErr, xecode.GraphPageWidthErr.Message(), strings.Join(verticalTitles, "、"))
	}
	return
}

func (s *Service) checkGraph(graph *model.GraphParam, isPreview int) (err error) {
	if graph.Aid == 0 {
		err = xecode.GraphAidEmpty
		return
	}
	if graph.Script == "" && isPreview != model.GraphIsPreview {
		err = xecode.GraphScriptEmpty
		return
	}
	if len(graph.Script) > _maxScriptLen {
		err = xecode.GraphScriptTooLong
	}
	err = s.checkSkinID(graph.SkinID)
	return
}

func (s *Service) checkSkinID(skinID int64) (err error) {
	if skinID <= 0 {
		return
	}
	var check bool
	for _, v := range s.skinList {
		if v.ID == skinID {
			check = true
			break
		}
	}
	if !check {
		err = xecode.GraphSkinNotFound
	}
	return
}

func (s *Service) checkNodes(nodes []*model.NodeParam, filterMsgs map[string]string) (nodeParamMap map[string]*model.NodeParam, errInfo model.ErrInfo, err error) {
	var isStartCheck bool
	// check node length
	if nodeLen := len(nodes); nodeLen == 0 || nodeLen > _nodeLength {
		err = xecode.GraphNodeCntErr
		return
	}
	nodeNameMap := make(map[string]struct{}, len(nodes))
	nodeParamMap = make(map[string]*model.NodeParam, len(nodes))
	for _, node := range nodes {
		// check node id
		if node.ID == "" || len(node.ID) > _idStrLen {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphNodeIDErr
			return
		}
		if _, ok := nodeParamMap[node.ID]; ok {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphNodeIDExist
			return
		}
		nodeParamMap[node.ID] = node
		// check node name
		node.Name = strings.TrimSpace(node.Name)
		if node.Name == "" || unicodeNameLen(node.Name) > _nodeNameLength {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphNodeNameErr
			return
		}
		if _, ok := nodeNameMap[node.Name]; ok {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphNodeNameExist
			return
		}
		nodeNameMap[node.Name] = struct{}{}
		filterMsgs[node.Name] = node.Name
		// check node type
		if _, typeCheck := model.NodeOtypes[node.Otype]; !typeCheck {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphNodeOtypeErr
			return
		}
		// check cid
		if node.Cid == 0 {
			errInfo.BuildErrInfo(_errTypeNode, node.ID)
			err = xecode.GraphNodeCidEmpty
			return
		}
		// node only one is start
		if node.IsStart == _isStart {
			if isStartCheck {
				errInfo.BuildErrInfo(_errTypeNode, node.ID)
				err = xecode.GraphDefaultNodeErr
				return
			}
			isStartCheck = true
		}
		if err = s.checkSkinID(node.SkinID); err != nil {
			return
		}
	}
	if !isStartCheck {
		err = xecode.GraphLackStartNode
		return
	}
	return
}

//nolint:gocognit
func checkEdges(nodes []*model.NodeParam, nodeParamMap map[string]*model.NodeParam, regionVarMap map[string]*model.RegionalVal, filterMsgs map[string]string) (errInfo model.ErrInfo, err error) {
	var bs []byte
	for _, node := range nodes {
		var isDefaultCheck bool
		for _, edge := range node.Edges {
			// check to node id
			if _, ok := nodeParamMap[edge.ToNodeID]; !ok {
				errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
				err = xecode.GraphEdgeToNodeNotFound
				return
			}
			if _, ok := model.EdgeTextAligns[edge.TextAlign]; !ok && edge.TextAlign != 0 {
				errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
				err = xecode.GraphEdgeTextAlignErr
				return
			}
			// check edge length
			if len(node.Edges) > _nodeEdgeLength {
				errInfo.BuildErrInfo(_errTypeNode, node.ID)
				err = xecode.GraphEdgeCntErr
				return
			}
			// show time zero only one edge.
			if node.ShowTime == model.ShowTimeDefault && len(node.Edges) > _nodeEdgeSpecLength {
				errInfo.BuildErrInfo(_errTypeNode, node.ID)
				err = xecode.GraphShowTimeEdgeErr
				return
			}
			edge.Title = strings.TrimSpace(edge.Title)
			// show time zero no edge title empty limit.
			if node.ShowTime == model.ShowTimeDefault {
				if unicodeNameLen(edge.Title) > _edgeTitleLength {
					errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
					err = xecode.GraphEdgeNameErr
					return
				}
			} else {
				if edge.Title == "" || unicodeNameLen(edge.Title) > _edgeTitleLength {
					errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
					err = xecode.GraphEdgeNameErr
					return
				}
			}
			if edge.IsDefault == _isDefault {
				if isDefaultCheck {
					errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
					err = xecode.GraphDefaultEdgeErr
					return
				}
				isDefaultCheck = true
			}
			filterMsgs[edge.Title] = edge.Title
			// check condition
			if bs, err = checkEdgeCondition(edge.Condition, regionVarMap); err != nil {
				errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
				return
			}
			if len(edge.Condition) > 0 {
				edge.ConditionStr = string(bs)
			}
			// check attribute
			if bs, err = checkEdgeAttribute(edge.Attribute, regionVarMap); err != nil {
				errInfo.BuildErrInfo(_errTypeEdge, edge.ID)
				return
			}
			if len(edge.Attribute) > 0 {
				edge.AttributeStr = string(bs)
			}
		}
	}
	return
}

func checkInteractiveArchive(mid, authorMid int64, attr int32) (err error) {
	if authorMid != mid {
		err = xecode.GraphNotOwner
		return
	}
	if (attr>>api.AttrBitSteinsGate)&int32(1) != api.AttrYes {
		err = xecode.GraphAidAttrErr
	}
	return
}

func checkGraphRegVars(regionalVars []*model.RegionalVal, filterMsgs map[string]string) (regionVarMap map[string]*model.RegionalVal, bs []byte, errInfo model.ErrInfo, err error) {
	varLen := len(regionalVars)
	regionVarMap = make(map[string]*model.RegionalVal, varLen)
	if varLen == 0 {
		return
	}
	if varLen > _graphRegionalVarLen {
		err = xecode.GraphRegVarsLenErr
		return
	}
	var normalCnt, randomCnt int
	varNames := make(map[string]struct{}, varLen)
	for _, v := range regionalVars {
		if v.ID == "" || len(v.ID) > _idStrLen {
			errInfo.BuildErrInfo(_errTypeVars, v.ID)
			err = xecode.GraphVarIDErr
			return
		}
		v.Name = strings.TrimSpace(v.Name)
		if v.ID == "" {
			errInfo.BuildErrInfo(_errTypeVars, v.ID)
			err = xecode.GraphRegVarsErr
			return
		}
		if v.Name == "" || unicodeNameLen(v.Name) > _graphRegVarNameLen {
			errInfo.BuildErrInfo(_errTypeVars, v.ID)
			err = xecode.GraphRegVarsNameLenErr
			return
		}
		if _, idCheck := regionVarMap[v.ID]; idCheck {
			errInfo.BuildErrInfo(_errTypeVars, v.ID)
			err = xecode.GraphRegVarsIDRepeat
			return
		}
		regionVarMap[v.ID] = v
		if _, nameCheck := varNames[v.Name]; nameCheck {
			errInfo.BuildErrInfo(_errTypeVars, v.ID)
			err = xecode.GraphRegVarsNameRepeat
			return
		}
		varNames[v.Name] = struct{}{}
		filterMsgs[v.Name] = v.Name
		switch v.Type {
		case model.RegionalVarTypeNormal:
			if v.InitMin < _normalInitMinBottom || v.InitMax > _normalInitMinTop {
				errInfo.BuildErrInfo(_errTypeVars, v.ID)
				err = xecode.GraphRegVarsRangeErr
				return
			}
			normalCnt++
		case model.RegionalVarTypeRandom:
			v.InitMin = _randomInitMinDft
			v.InitMax = _randomInitMaxDft
			randomCnt++
		default:
			errInfo.BuildErrInfo(_errTypeVars, v.ID)
			err = xecode.GraphRegVarsTypeErr
			return
		}
	}
	if normalCnt > _graphRegVarNormalLen {
		err = xecode.GraphRegNormalVarsLenErr
		return
	}
	if randomCnt > _graphRegVarRandomLen {
		err = xecode.GraphRegRandomVarsLenErr
		return
	}
	if bs, err = json.Marshal(regionalVars); err != nil {
		log.Error("checkTreeParam vars json.Marshal(%+v) error(%v)", regionalVars, err)
		err = xecode.GraphRegVarsErr
	}
	return
}

func checkEdgeAttribute(attrs []*model.EdgeAttribute, regionVarMap map[string]*model.RegionalVal) (bs []byte, err error) {
	attrLen := len(attrs)
	if attrLen == 0 {
		return
	}
	if attrLen > _edgeAttrLen {
		err = xecode.GraphEdgeAttrLenErr
		return
	}
	attrMap := make(map[string]struct{}, attrLen)
	for _, v := range attrs {
		if v.VarID == "" {
			err = xecode.GraphEdgeAttrVarIDNone
			return
		}
		if _, ok := model.EdgeAttrActions[v.Action]; !ok {
			err = xecode.GraphEdgeAttrTypeNone
			return
		}
		if regVar, ok := regionVarMap[v.VarID]; !ok {
			err = xecode.GraphEdgeAttrVarIDErr
			return
		} else if regVar.Type != model.RegionalVarTypeNormal {
			err = xecode.GraphEdgeAttrTypeErr
			return
		}
		if _, ok := attrMap[v.VarID]; ok {
			err = xecode.GraphEdgeAttrRepeat
			return
		}
		attrMap[v.VarID] = struct{}{}
	}
	if bs, err = json.Marshal(attrs); err != nil {
		log.Error("checkTreeParam attr json.Marshal(%+v) error(%v)", attrs, err)
		err = xecode.GraphAttributeErr
	}
	return
}

func checkEdgeCondition(conds []*model.EdgeCondition, regionVarMap map[string]*model.RegionalVal) (bs []byte, err error) {
	if condLen := len(conds); condLen > _edgeCondLen {
		err = xecode.GraphEdgeCondLenErr
		return
	}
	condMap := make(map[string][]*model.EdgeCondition)
	for _, v := range conds {
		if v.VarID == "" {
			err = xecode.GraphEdgeCondVarIDNone
			return
		}
		if _, ok := model.EdgeConditionTypes[v.Condition]; !ok {
			err = xecode.GraphEdgeCondTypeNone
			return
		}
		if regVar, ok := regionVarMap[v.VarID]; !ok {
			err = xecode.GraphEdgeCondVarIDErr
			return
		} else if regVar.Type == model.RegionalVarTypeRandom {
			if v.Value < _randomInitMinDft || v.Value > _randomInitMaxDft {
				err = xecode.GraphEdgeCondRandRangeErr
				return
			}
		}
		condMap[v.VarID] = append(condMap[v.VarID], v)
	}
	for _, v := range condMap {
		if len(v) > _edgeOneVarCondLen {
			err = ecode.Errorf(xecode.GraphEdgeCondVarLenErr, xecode.GraphEdgeCondVarLenErr.Message(), _edgeOneVarCondLen)
			return
		}
		if len(v) == _edgeOneVarCondLen {
			condMap := map[string]int{
				v[0].Condition: v[0].Value,
				v[1].Condition: v[1].Value,
			}
			if exclusionCond(condMap) {
				err = xecode.GraphEdgeCondExclusion
				return
			}
		}
	}
	if bs, err = json.Marshal(conds); err != nil {
		log.Error("checkTreeParam cond json.Marshal(%+v) error(%v)", conds, err)
		err = xecode.GraphConditionErr
	}
	return
}

func exclusionCond(conds map[string]int) bool {
	if len(conds) == 1 {
		return true
	}
	gtVal, gtCheck := conds[model.EdgeConditionTypeGt]
	geVal, geCheck := conds[model.EdgeConditionTypeGe]
	ltVal, ltCheck := conds[model.EdgeConditionTypeLt]
	leVal, leCheck := conds[model.EdgeConditionTypeLe]
	if gtCheck && geCheck {
		return true
	}
	if ltCheck && leCheck {
		return true
	}
	if gtCheck {
		if ltCheck && ltVal <= gtVal {
			return true
		}
		if leCheck && leVal <= gtVal {
			return true
		}
	}
	if geCheck {
		if ltCheck && ltVal <= geVal {
			return true
		}
		if leCheck && leVal < geVal {
			return true
		}
	}
	return false
}

func unicodeNameLen(str string) (total int) {
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			total += 2
		} else {
			total += 1
		}
	}
	return
}

func (s *Service) isShowWhite(mid int64) bool {
	if env.DeployEnv != env.DeployEnvPre {
		return false
	}
	if _, ok := s.preGraphMids[mid]; ok {
		return true
	}
	return false

}

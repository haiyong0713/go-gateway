package service

import (
	"context"
	"strconv"
	"time"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	pgcfollowgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/net/metadata"

	appcardecode "go-gateway/app/app-svr/app-card/ecode"
	xecode "go-gateway/app/app-svr/native-act/ecode"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_modulePs = 50
)

// 白名单check
func (s *Service) checkWhite(c context.Context, firstPage *natpagegrpc.FirstPage, mid int64) error {
	if firstPage == nil || firstPage.Item == nil { //历史数据无父page信息,没有白名单逻辑，直接校验通过
		return nil
	}
	if firstPage.Item.IsAttrWhiteSwitch() != natpagegrpc.AttrModuleYes { //没有开通白名单逻辑
		return nil
	}
	if mid <= 0 || firstPage.Ext == nil { //未登录用户不支持访问 || 开通了白名单逻辑，但是数据源获取失败
		return ecode.ServerErr
	}
	sid, ok := strconv.ParseInt(firstPage.Ext.WhiteValue, 10, 64)
	if ok != nil { //配置错误，页面不下发
		return ecode.ServerErr
	}
	upList, err := s.dao.Activity().UpList(c, &activitygrpc.UpListReq{Sid: sid, Type: natpagegrpc.SortTypeCtime, Pn: 1, Ps: 50})
	if err != nil || upList == nil {
		return ecode.ServerErr
	}
	for _, v := range upList.List {
		if v == nil || v.Item == nil {
			continue
		}
		if v.Item.Wid == mid { //是白名单mid
			return nil
		}
	}
	return ecode.ServerErr
}

func (s *Service) Index(c context.Context, req *api.IndexReq) (*api.PageResp, error) {
	session := NewSessionOfIndex(c, req)
	page, err := s.dao.Natpage().NatConfig(c, req.PageId, _modulePs, 0, natpagegrpc.CommonPage)
	if page == nil {
		return &api.PageResp{IsOnline: false}, nil
	}
	natpage := page.NativePage
	if ok, resp, err := handleNatpageRly(natpage, err); !ok {
		return resp, err
	}
	//白名单check
	if err := s.checkWhite(c, page.FirstPage, session.Mid()); err != nil {
		return &api.PageResp{IsOnline: false}, nil
	}
	if natpage.IsInlineAct() {
		var (
			primaryPage *natpagegrpc.NativePage
			priErr      error
		)
		if natpage.FirstPid != 0 { //数据库中有最新的父子关系
			req.PrimaryPageId = natpage.FirstPid
			//grpc有返回父id页面info
			if page.FirstPage != nil && page.FirstPage.Item != nil {
				primaryPage = page.FirstPage.Item
			}
		}
		if req.PrimaryPageId <= 0 {
			return &api.PageResp{IsOnline: false}, nil
		}
		if primaryPage == nil {
			primaryPage, priErr = s.dao.Natpage().NativePage(c, req.PrimaryPageId)
		}
		if ok, resp, err := handleNatpageRly(primaryPage, priErr); !ok {
			return resp, err
		}
		if !checkTabLock(natpage) {
			return nil, xecode.ActivityHasLock
		}
		natpage = primaryPage
	}
	cardCfgs := GlobalCardResolver.Resolve(c, session, natpage, page.Modules)
	baseCfgs := GlobalCardResolver.Resolve(c, session, natpage, page.Bases)
	mlReqID, ml := newIndexML(c, s.dao, session, req)
	materials := loadMaterials(ml, append(cardCfgs, baseCfgs...))
	modules := GlobalCardBuilder.Build(c, session, s.dao, cardCfgs, materials)
	baseModules := GlobalCardBuilder.Build(c, session, s.dao, baseCfgs, materials)
	resp := buildPageResp(session, natpage, modules, baseModules)
	setExtraIndexResp(req, resp, materials, mlReqID)
	return resp, nil
}

func (s *Service) Dynamic(c context.Context, req *api.DynamicReq) (*api.DynamicResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.DynamicParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.DynamicParams)
			if !ok || params.ModuleId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	module, err := s.subpageModules(c, req.Params.ModuleId, req.PrimaryPageId, func() *kernel.Session { return NewSessionOfDynamic(c, req) })
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	return &api.DynamicResp{Module: module}, nil
}

func (s *Service) Editor(c context.Context, req *api.EditorReq) (*api.EditorResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.EditorParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.EditorParams)
			if !ok || params.ModuleId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	module, err := s.subpageModules(c, req.Params.ModuleId, req.PrimaryPageId, func() *kernel.Session { return NewSessionOfEditor(c, req) })
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	return &api.EditorResp{Module: module}, nil
}

func (s *Service) Resource(c context.Context, req *api.ResourceReq) (*api.ResourceResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.ResourceParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.ResourceParams)
			if !ok || params.ModuleId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	module, err := s.subpageModules(c, req.Params.ModuleId, req.PrimaryPageId, func() *kernel.Session { return NewSessionOfResource(c, req) })
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	return &api.ResourceResp{Module: module}, nil
}

func (s *Service) Video(c context.Context, req *api.VideoReq) (*api.VideoResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.VideoParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.VideoParams)
			if !ok || params.ModuleId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	module, err := s.subpageModules(c, req.Params.ModuleId, req.PrimaryPageId, func() *kernel.Session { return NewSessionOfVideo(c, req) })
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	return &api.VideoResp{Module: module}, nil
}

func (s *Service) subpageModules(c context.Context, moduleID int64, primaryID int64, ssf func() *kernel.Session) (*api.Module, error) {
	moduleRly, err := s.dao.Natpage().ModuleConfig(c, moduleID, primaryID)
	if err != nil {
		return nil, err
	}
	if moduleRly == nil {
		return nil, errors.WithMessage(ecode.ServerErr, "获取组件信息失败")
	}
	session := ssf()
	page := moduleRly.NativePage
	if moduleRly.PrimaryPage != nil {
		page = moduleRly.PrimaryPage
	}
	if page == nil {
		return nil, errors.WithMessage(ecode.ServerErr, "页面数据为空")
	}
	cardCfgs := GlobalCardResolver.Resolve(c, session, page, []*natpagegrpc.Module{moduleRly.Module})
	materials := loadMaterials(kernel.NewMaterialLoader(c, s.dao, session), cardCfgs)
	modules := GlobalCardBuilder.Build(c, session, s.dao, cardCfgs, materials)
	if len(modules) == 0 {
		return nil, ecode.NothingFound
	}
	return modules[0], nil
}

func (s *Service) Vote(c context.Context, req *api.VoteReq) (*api.VoteResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.VoteParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.VoteParams)
			if !ok || params.Sid <= 0 || params.SourceItemId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, errors.WithMessage(ecode.RequestErr, "入参错误")
	}
	ss := kernel.NewSession(c)
	switch req.Params.Type {
	case model.SourceTypeVoteAct:
		if req.Params.Gid <= 0 {
			return nil, errors.WithMessage(ecode.RequestErr, "gid为空")
		}
		return s.voteAct(c, req.Params, ss)
	case model.SourceTypeVoteUp:
		return s.voteUp(c, req.Params, ss)
	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "未知的数据源类型=%+v", req.Params.Type)
	}
}

func (s *Service) voteAct(c context.Context, params *api.VoteParams, ss *kernel.Session) (*api.VoteResp, error) {
	var leftNum, canVoteNum int64
	switch params.Action {
	case api.ActionType_Do:
		doRly, err := s.dao.Activity().VoteUserDo(c, &activitygrpc.VoteUserDoReq{
			ActivityId:    params.Sid,
			SourceGroupId: params.Gid,
			SourceItemId:  params.SourceItemId,
			VoteCount:     1,
			Risk: &activitygrpc.Risk{
				Buvid:     ss.Buvid(),
				UserAgent: ss.UserAgent(),
				Ip:        ss.Ip(),
				Build:     strconv.FormatInt(ss.RawDevice().Build, 10),
				Platform:  ss.RawDevice().RawPlatform,
				Api:       "/bilibili.app.nativeact.v1.NativeAct/Vote",
			},
			Mid: ss.Mid(),
		})
		if err != nil {
			return nil, err
		}
		leftNum = doRly.UserAvailVoteCount
		canVoteNum = doRly.UserCanVoteCountForItem
	case api.ActionType_Undo:
		undoRly, err := s.dao.Activity().VoteUserUndo(c, &activitygrpc.VoteUserUndoReq{
			ActivityId:    params.Sid,
			SourceGroupId: params.Gid,
			SourceItemId:  params.SourceItemId,
			Mid:           ss.Mid(),
		})
		if err != nil {
			return nil, err
		}
		leftNum = undoRly.UserAvailVoteCount
		canVoteNum = undoRly.UserCanVoteCountForItem
	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "未知操作=%+v", params.Action)
	}
	newAction := api.ActionType_Do
	if params.Action == api.ActionType_Do {
		newAction = api.ActionType_Undo
	}
	return &api.VoteResp{
		VoteParams: passthrough.Marshal(&api.VoteParams{
			Action:       newAction,
			Sid:          params.Sid,
			Gid:          params.Gid,
			SourceItemId: params.SourceItemId,
			Type:         params.Type,
		}),
		LeftNum:    leftNum,
		CanVoteNum: canVoteNum,
	}, nil
}

func (s *Service) voteUp(c context.Context, params *api.VoteParams, ss *kernel.Session) (*api.VoteResp, error) {
	switch params.Action {
	case api.ActionType_Do:
		voteReq := &dynvotegrpc.DoVoteReq{VoteId: params.Sid, Votes: []int32{int32(params.SourceItemId)}, VoterUid: ss.Mid()}
		if _, err := s.dao.Dynvote().DoVote(c, voteReq); err != nil {
			return nil, err
		}
	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "未知操作=%+v", params.Action)
	}
	return &api.VoteResp{
		VoteParams: passthrough.Marshal(&api.VoteParams{
			Action:       api.ActionType_Undo,
			Sid:          params.Sid,
			SourceItemId: params.SourceItemId,
			Type:         params.Type,
		}),
	}, nil
}

func (s *Service) Reserve(c context.Context, req *api.ReserveReq) (*api.ReserveRly, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.ReserveParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.ReserveParams)
			if !ok || params.Sid <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	var err error
	ss := kernel.NewSession(c)
	switch req.Params.Action {
	case api.ActionType_Do:
		err = s.dao.Activity().AddReserve(c, &activitygrpc.AddReserveReq{Sid: req.Params.Sid, Mid: ss.Mid()})
	case api.ActionType_Undo:
		err = s.dao.Activity().DelReserve(c, &activitygrpc.DelReserveReq{Sid: req.Params.Sid, Mid: ss.Mid()})
	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "未知操作=%+v", req.Params.Action)
	}
	if err != nil {
		return nil, ecode.ServerErr
	}
	newAction := api.ActionType_Do
	if req.Params.Action == api.ActionType_Do {
		newAction = api.ActionType_Undo
	}
	return &api.ReserveRly{
		ReserveParams: passthrough.Marshal(&api.ReserveParams{Action: newAction, Sid: req.Params.Sid}),
	}, nil
}

func (s *Service) TimelineSupernatant(c context.Context, req *api.TimelineSupernatantReq) (*api.TimelineSupernatantResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.TimelineSupernatantParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.TimelineSupernatantParams)
			if !ok || params.ModuleId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	module, err := s.subpageModules(c, req.Params.ModuleId, req.PrimaryPageId, func() *kernel.Session { return NewSessionOfTimelineSupernatant(c, req) })
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	var lastIndex int64
	const double = 2
	if req.Params.Offset == 0 {
		// 第一刷返回定位的卡片位置
		lastIndex = req.Params.LastIndex * double
	}
	return &api.TimelineSupernatantResp{Module: module, LastIndex: lastIndex}, nil
}

func (s *Service) OgvSupernatant(c context.Context, req *api.OgvSupernatantReq) (*api.OgvSupernatantResp, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.OgvSupernatantParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.OgvSupernatantParams)
			if !ok || params.ModuleId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	module, err := s.subpageModules(c, req.Params.ModuleId, req.PrimaryPageId, func() *kernel.Session { return NewSessionOfOgvSupernatant(c, req) })
	if err != nil && err != ecode.NothingFound {
		return nil, err
	}
	var lastIndex int64
	if req.Params.Offset == 0 {
		// 第一刷返回定位的卡片位置
		lastIndex = req.Params.LastIndex
	}
	return &api.OgvSupernatantResp{Module: module, LastIndex: lastIndex}, nil
}

func (s *Service) FollowOgv(c context.Context, req *api.FollowOgvReq) (*api.FollowOgvRly, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.FollowOgvParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.FollowOgvParams)
			if !ok || params.SeasonId <= 0 {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	var err error
	ss := kernel.NewSession(c)
	switch req.Params.Action {
	case api.ActionType_Do:
		err = s.dao.Pgcfollow().AddFollow(c, &pgcfollowgrpc.FollowReq{SeasonId: req.Params.SeasonId, Mid: ss.Mid()})
	case api.ActionType_Undo:
		err = s.dao.Pgcfollow().DeleteFollow(c, &pgcfollowgrpc.FollowReq{SeasonId: req.Params.SeasonId, Mid: ss.Mid()})
	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "未知操作=%+v", req.Params.Action)
	}
	if err != nil {
		return nil, ecode.ServerErr
	}
	newAction := api.ActionType_Do
	if req.Params.Action == api.ActionType_Do {
		newAction = api.ActionType_Undo
	}
	return &api.FollowOgvRly{
		FollowParams: passthrough.Marshal(&api.FollowOgvParams{Action: newAction, SeasonId: req.Params.SeasonId}),
	}, nil
}

func (s *Service) Progress(c context.Context, req *api.ProgressReq) (*api.ProgressRly, error) {
	if req.PageId <= 0 {
		return nil, errors.WithMessage(ecode.RequestErr, "page_id为空")
	}
	pgParams, err := s.dao.Natpage().GetNatProgressParams(c, req.PageId)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "获取进度条参数失败")
	}
	groups := s.progressGroups(c, pgParams)
	event := &api.ProgressEvent{
		PageID: req.PageId,
		Items:  make([]*api.ProgressEventItem, 0, len(groups)),
	}
	for _, param := range pgParams {
		group, ok := groups[param.Sid][param.GroupID]
		if !ok {
			continue
		}
		item := &api.ProgressEventItem{
			ItemID:     param.Id,
			Type:       param.Type,
			Num:        group.Total,
			DisplayNum: builder.StatString(group.Total),
			WebKey:     param.WebKey,
			Dimension:  param.Dimension,
		}
		event.Items = append(event.Items, item)
	}
	return &api.ProgressRly{Event: event}, nil
}

func (s *Service) progressGroups(c context.Context, pgParams []*natpagegrpc.ProgressParam) map[int64]map[int64]*activitygrpc.ActivityProgressGroup {
	ml := kernel.NewMaterialLoader(c, s.dao, kernel.NewSession(c))
	for _, param := range pgParams {
		if param.Sid <= 0 || param.GroupID <= 0 {
			continue
		}
		_, _ = ml.AddItem(model.MaterialActProgressGroup, param.Sid, []int64{param.GroupID})
	}
	material := &kernel.Material{}
	ml.Load(material)
	return material.ActProgressGroups
}

func (s *Service) HandleClick(c context.Context, req *api.HandleClickReq) (*api.HandleClickRly, error) {
	if ok := passthrough.ResolveRawParamsOfReq(req.RawParams, &api.ClickRequestParams{},
		func(message proto.Message) bool {
			params, ok := message.(*api.ClickRequestParams)
			if !ok || params.Id <= 0 || !(params.Action == api.ActionType_Do || params.Action == api.ActionType_Undo) {
				return false
			}
			req.Params = params
			return true
		},
	); !ok {
		return nil, ecode.RequestErr
	}
	var err error
	ss := kernel.NewSession(c)
	switch req.Params.ReqType {
	case api.ClickRequestType_CRTypeFollowUser:
		if req.Params.Action == api.ActionType_Do {
			err = s.dao.Relation().AddFollowing(c, &relationgrpc.FollowingReq{Mid: ss.Mid(), Fid: req.Params.Id, Spmid: req.Spmid})
		} else {
			err = s.dao.Relation().DelFollowing(c, &relationgrpc.FollowingReq{Mid: ss.Mid(), Fid: req.Params.Id, Spmid: req.Spmid})
		}
	case api.ClickRequestType_CRTypeFollowEpisode:
		if req.Params.Action == api.ActionType_Do {
			err = s.dao.Pgcfollow().AddFollow(c, &pgcfollowgrpc.FollowReq{SeasonId: int32(req.Params.Id), Mid: ss.Mid()})
		} else {
			err = s.dao.Pgcfollow().DeleteFollow(c, &pgcfollowgrpc.FollowReq{SeasonId: int32(req.Params.Id), Mid: ss.Mid()})
		}
	case api.ClickRequestType_CRTypeFollowComic:
		if req.Params.Action == api.ActionType_Do {
			err = s.dao.Comic().AddFavorite(c, []int64{req.Params.Id}, ss.Mid())
		} else {
			err = s.dao.Comic().DelFavorite(c, []int64{req.Params.Id}, ss.Mid())
		}
	case api.ClickRequestType_CRTypeReserve, api.ClickRequestType_CRTypeUpReserve:
		if req.Params.Action == api.ActionType_Do {
			err = s.dao.Activity().AddReserve(c, &activitygrpc.AddReserveReq{Sid: req.Params.Id, Mid: ss.Mid()})
		} else {
			err = s.dao.Activity().DelReserve(c, &activitygrpc.DelReserveReq{Sid: req.Params.Id, Mid: ss.Mid()})
		}
	case api.ClickRequestType_CRTypeReceiveAward:
		if req.Params.Action == api.ActionType_Do {
			if err = s.dao.Activity().RewardSubject(c, &activitygrpc.RewardSubjectReq{Mid: ss.Mid(), Id: req.Params.Id}); err != nil {
				err = appcardecode.AppReceiveErr
			}
		}
	case api.ClickRequestType_CRTypeMallWantGo:
		if req.Params.Action == api.ActionType_Do {
			err = s.dao.Mallticket().AddFav(c, req.Params.Id, ss.Mid())
		} else {
			err = s.dao.Mallticket().DelFav(c, req.Params.Id, ss.Mid())
		}
	case api.ClickRequestType_CRTypeActivity:
		if req.Params.Action == api.ActionType_Do {
			err = s.dao.Activity().GRPCDoRelation(c, &activitygrpc.GRPCDoRelationReq{
				Id:       req.Params.Id,
				Mid:      ss.Mid(),
				From:     "native_page",
				Ip:       metadata.String(c, metadata.RemoteIP),
				Platform: ss.RawDevice().RawPlatform,
				Mobiapp:  ss.RawDevice().MobiApp(),
				Buvid:    ss.RawDevice().Buvid,
				Spmid:    req.Spmid,
			})
		} else {
			err = s.dao.Activity().RelationReserveCancel(c, &activitygrpc.RelationReserveCancelReq{
				Id:       req.Params.Id,
				Mid:      ss.Mid(),
				From:     "native_page",
				Ip:       metadata.String(c, metadata.RemoteIP),
				Platform: ss.RawDevice().RawPlatform,
				Mobiapp:  ss.RawDevice().MobiApp(),
				Buvid:    ss.RawDevice().Buvid,
				Spmid:    req.Spmid,
			})
		}
	default:
		err = errors.WithMessagef(ecode.RequestErr, "未知类型=%+v", req.Params.ReqType)
	}
	if err != nil {
		return nil, err
	}
	state := api.ClickRequestState_CRSDone
	if req.Params.Action == api.ActionType_Undo {
		state = api.ClickRequestState_CRSUndone
	}
	return &api.HandleClickRly{State: state}, nil
}

func handleNatpageRly(nativePage *natpagegrpc.NativePage, err error) (bool, *api.PageResp, error) {
	if err != nil {
		if ecode.EqualError(xecode.OldNativePageOffline, err) {
			return false, &api.PageResp{IsOnline: false}, nil
		}
		return false, nil, ecode.ServerErr
	}
	if nativePage == nil || nativePage.IsOffline() {
		return false, &api.PageResp{IsOnline: false}, nil
	}
	return true, nil, nil
}

func checkTabLock(page *natpagegrpc.NativePage) bool {
	lockExt := page.ConfSetUnmarshal()
	if lockExt.DT == model.TabLockPass {
		return true
	}
	switch lockExt.DC {
	case model.TabLockTypeTime:
		if lockExt.Stime <= time.Now().Unix() {
			return true
		}
	}
	return false
}

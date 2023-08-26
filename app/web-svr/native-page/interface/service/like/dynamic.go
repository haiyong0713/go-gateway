package like

import (
	"context"
	"encoding/json"
	"fmt"
	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	"strconv"
	"strings"
	"sync"
	"time"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	channelapi "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	scoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
	hmtchagrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	playgrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	media "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	arccli "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/native-page/ecode"
	"go-gateway/app/web-svr/native-page/interface/api"
	carmdl "go-gateway/app/web-svr/native-page/interface/model/cartoon"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
	lmdl "go-gateway/app/web-svr/native-page/interface/model/like"
	"go-gateway/pkg/idsafe/bvid"
)

// InlineTab .
func (s *Service) InlineTab(c context.Context, a *dynmdl.ParamActInline, mid int64) (*dynmdl.InlineReply, error) {
	pageConf, e := s.BaseConfig(c, &api.BaseConfigReq{Pid: a.PageID, Offset: a.Offset, Ps: a.Ps, PType: api.CommonPage})
	if e != nil {
		return nil, e
	}
	if pageConf == nil || pageConf.NativePage == nil {
		return nil, ecode.NativePageOffline
	}
	commonConf := pageConf.NativePage
	if !commonConf.IsInlineAct() {
		return nil, ecode.NativePageOffline
	}
	//白名单checke
	if err := s.checkWhite(c, pageConf.FirstPage, mid); err != nil {
		return nil, err
	}
	//是否锁定
	lockExt := commonConf.ConfSetUnmarshal()
	if lockExt.DT == api.NeedUnLock { //解锁模式
		var deblocking bool
		if lockExt.DC == api.UnLockTime && lockExt.Stime <= time.Now().Unix() { //时间模式&&到达解锁时间
			deblocking = true
		}
		//未解锁时
		if !deblocking {
			return nil, ecode.NativePageOffline
		}
	}
	//是否锁定
	// 拼接组件信息
	modulesRly := &dynmdl.ModulesReply{}
	eg := errgroup.WithContext(c)
	s.fromParamModule(eg, pageConf.Bases, modulesRly, mid, a.PrimaryPageID, pageConf.NativePage, a.MobiApp, "", a.Buvid)
	if err := eg.Wait(); err != nil { //错误忽略降级处理
		log.Error("InlineTab eg.Wait() error(%v)", err)
	}
	reply := &dynmdl.InlineReply{
		PageID: commonConf.ID,
		Title:  commonConf.Title,
	}
	if len(modulesRly.Card) != 0 {
		for _, mv := range pageConf.Bases {
			if mv.NativeModule == nil {
				continue
			}
			val, k := modulesRly.Card[mv.NativeModule.ID]
			if !k || val == nil {
				continue
			}
			reply.Items = append(reply.Items, val)
		}
	}
	return reply, nil
}

// MenuTab .
func (s *Service) MenuTab(c context.Context, a *dynmdl.ParamMenuTab, mid int64) (*dynmdl.MenuReply, error) {
	pageConf, e := s.BaseConfig(c, &api.BaseConfigReq{Pid: a.PageID, Offset: 0, Ps: -1, PType: api.CommonPage})
	if e != nil || pageConf == nil || pageConf.NativePage == nil {
		log.Error("s.BaseConfig(%d) error(%v) or is nil", a.PageID, e)
		return nil, ecode.NativePageOffline
	}
	commonConf := pageConf.NativePage
	var (
		opFrom string
	)
	switch {
	case commonConf.IsSpaceAct():
		opFrom = dynmdl.FormatModFromMenuSpace
	case commonConf.IsUpTopicAct():
		opFrom = dynmdl.FormatModFromMenuUp
	default:
		return nil, ecode.NativePageOffline
	}
	reply := &dynmdl.MenuReply{
		PageID:  commonConf.ID,
		Title:   commonConf.Title,
		BgColor: commonConf.BgColor,
	}
	if len(pageConf.Bases) == 0 {
		return reply, nil
	}
	// 拼接组件信息
	modulesRly := &dynmdl.ModulesReply{}
	eg := errgroup.WithContext(c)
	s.fromParamModule(eg, pageConf.Bases, modulesRly, mid, a.PageID, pageConf.NativePage, a.MobiApp, opFrom, a.Buvid)
	var attentions []int64
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			attes, err := s.relDao.Attentions(ctx, mid)
			if err != nil {
				log.Error("Failed to get attentions: mid: %d, error: %+v", mid, err)
				return nil
			}
			for _, v := range attes.GetFollowingList() {
				attentions = append(attentions, v.Mid)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil { //错误忽略降级处理
		log.Error("InlineTab eg.Wait() error(%v)", err)
	}
	if len(attentions) > 0 {
		reply.Attentions = &dynmdl.Attentions{Uids: attentions}
	}
	//导航组件与inlinetab组件是互斥的,且组件本身互斥
	//评论组件和动态无限feed流互斥,且组件本身互斥
	var hasMutex, hasFeed bool
	for _, v := range pageConf.Bases {
		if v == nil || v.NativeModule == nil {
			continue
		}
		mVal, ok := modulesRly.Card[v.NativeModule.ID]
		if !ok || mVal == nil {
			continue
		}
		switch mVal.Goto {
		case dynmdl.GotoNavigationModule, dynmdl.GotoInlineTabModule:
			if hasMutex { // 如果有互斥组件，需要丢弃
				continue
			}
			hasMutex = true
		case dynmdl.GotoReplyModule:
			if hasFeed { // 如果有互斥组件，需要丢弃
				continue
			}
			hasFeed = true
		case dynmdl.GotoDynamicModule:
			if mVal.IsFeed == 1 { // 如果有互斥组件，需要丢弃
				if hasFeed {
					continue
				}
				hasFeed = true
			}
		}
		reply.Items = append(reply.Items, mVal)
	}
	return reply, nil
}

// 白名单check
func (s *Service) checkWhite(c context.Context, firstPage *api.FirstPage, mid int64) error {
	if firstPage == nil || firstPage.Item == nil { //历史数据无父page信息,没有白名单逻辑，直接校验通过
		return nil
	}
	if firstPage.Item.IsAttrWhiteSwitch() != api.AttrModuleYes { //没有开通白名单逻辑
		return nil
	}
	if mid <= 0 || firstPage.Ext == nil { //未登录用户不支持访问 || 开通了白名单逻辑，但是数据源获取失败
		return ecode.NativePageOffline
	}
	sid, ok := strconv.ParseInt(firstPage.Ext.WhiteValue, 10, 64)
	if ok != nil { //配置错误，页面不下发
		return ecode.NativePageOffline
	}
	upList, err := s.actDao.UpList(c, sid, 1, 50, 0, api.SortTypeCtime)
	if err != nil || upList == nil {
		log.Error("s.actDao.UpList(%d) error(%v)", sid, err)
		return ecode.NativePageOffline
	}
	for _, v := range upList.List {
		if v == nil || v.Item == nil {
			continue
		}
		if v.Item.Wid == mid { //是白名单mid
			return nil
		}
	}
	return ecode.NativePageOffline
}

// ActIndex .
func (s *Service) ActIndex(c context.Context, arg *dynmdl.ParamActIndex, mid int64) (reply *dynmdl.IndexReply, err error) {
	var (
		pageConf   *api.NatConfigReply
		attentions []int64
	)
	if pageConf, err = s.NatConfig(c, &api.NatConfigReq{Pid: arg.PageID, Offset: arg.Offset, Ps: arg.Ps, PType: arg.PType}); err != nil || pageConf == nil || pageConf.NativePage == nil {
		log.Error("s.NatConfig(%d) error(%v)", arg.PageID, err)
		return
	}
	//白名单checke
	if err = s.checkWhite(c, pageConf.FirstPage, mid); err != nil {
		return
	}
	commonConf := pageConf.NativePage
	reply = &dynmdl.IndexReply{
		PageID:       commonConf.ID,
		Title:        commonConf.Title,
		ForeignID:    commonConf.ForeignID,
		ForeignType:  commonConf.Type,
		ShareTitle:   commonConf.ShareTitle,
		ShareImage:   commonConf.ShareImage,
		Spmid:        commonConf.Spmid,
		SkipURL:      commonConf.SkipURL,
		ShareCaption: commonConf.Title,
		Uid:          commonConf.RelatedUid,
		BgColor:      commonConf.BgColor,
		FromType:     commonConf.FromType,
		Ver:          commonConf.Ver,
	}
	if commonConf.ShareCaption != "" {
		reply.ShareCaption = commonConf.ShareCaption
	}
	if !(commonConf.IsTopicAct() || commonConf.IsNewact()) || commonConf.SkipURL != "" {
		return
	}
	// 拼接组件信息
	modulesRly := &dynmdl.ModulesReply{}
	eg := errgroup.WithContext(c)
	s.fromParamModule(eg, pageConf.Modules, modulesRly, mid, arg.PageID, pageConf.NativePage, arg.MobiApp, "", arg.Buvid)
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			attes, e := s.relDao.Attentions(ctx, mid)
			if e != nil {
				log.Error("s.reldao.Attentions(%d) error(%v)", mid, e)
				return nil
			}
			if attes == nil {
				return nil
			}
			attentions = make([]int64, 0, len(attes.FollowingList))
			for _, v := range attes.FollowingList {
				attentions = append(attentions, v.Mid)
			}
			return nil
		})
	}
	_ = eg.Wait()
	pageURL := "https://www.bilibili.com/blackboard/dynamic/" + strconv.FormatInt(commonConf.ID, 10)
	reply.PageURL = pageURL
	if commonConf.ShareURL != "" {
		reply.ShareURL = commonConf.ShareURL
	} else {
		reply.ShareURL = pageURL
	}
	reply.PcURL = commonConf.PcURL
	if len(attentions) > 0 {
		reply.Attentions = &dynmdl.Attentions{Uids: attentions}
	}
	if len(modulesRly.Card) != 0 {
		for _, mv := range pageConf.Modules {
			if mv.NativeModule == nil {
				continue
			}
			val, k := modulesRly.Card[mv.NativeModule.ID]
			if !k || val == nil {
				continue
			}
			reply.Items = append(reply.Items, val)
		}
	}
	if len(pageConf.Bases) > 0 {
		baseHead, hoverButton := extractBaseModules(pageConf.Bases)
		reply.Bases = &dynmdl.Bases{}
		if baseHead != nil {
			reply.Bases.Head = s.formatBaseHead(baseHead.NativeModule)
		}
		if hoverButton != nil {
			reply.Bases.HoverButton = s.formatBaseHoverButton(c, hoverButton.NativeModule, mid)
		}
	}
	return
}

//nolint:gocognit
func (s *Service) fromParamModule(eg *errgroup.Group, arg []*api.Module, rly *dynmdl.ModulesReply, mid, primaryPageID int64, page *api.NativePage, mobiApp string, opFrom, buvid string) {
	var mu sync.Mutex
	if len(arg) == 0 {
		return
	}
	rly.Card = make(map[int64]*dynmdl.Item)
	for _, vk := range arg {
		if vk == nil || vk.NativeModule == nil {
			continue
		}
		tmpModu := vk.NativeModule
		switch {
		case tmpModu.IsReserve(): //预约组件
			tempRev := vk.Reserve
			eg.Go(func(ctx context.Context) error {
				ck := s.formatReserve(ctx, tmpModu, tempRev, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsGame(): //游戏组件
			if mobiApp != "android" && mobiApp != "iphone" { //h5页面下仅仅支持粉版
				continue
			}
			tempGame := vk.Game
			eg.Go(func(ctx context.Context) error {
				ck := s.formatGame(ctx, tmpModu, tempGame, mid, mobiApp)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsRecommend(), tmpModu.IsRcmdSource(): //推荐用户组件
			tempRecom := vk.Recommend
			eg.Go(func(ctx context.Context) error {
				ck := s.formatRecommend(ctx, tmpModu, tempRecom, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsRcmdVertical(), tmpModu.IsRcmdVerticalSource(): //推荐用户-竖卡组件
			rcmdVertical := vk.Recommend
			eg.Go(func(ctx context.Context) error {
				ck := s.formatRcmdVertical(ctx, tmpModu, rcmdVertical, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsAct(): //相关活动组件
			tempAct := vk.Act
			// 相关活动列表为空，则不下发组件
			if tempAct == nil {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatActCard(tmpModu, tempAct)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsActCapsule(): //相关活动组件 -胶囊
			actPage := vk.ActPage
			if actPage == nil {
				continue
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatCapsule(ctx, tmpModu, actPage, primaryPageID, opFrom)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsTimelineSource(): //时间轴组件-资源类型
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatTimelineResource(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsTimelineIDs(): //时间轴组件-ids
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatTimelineIDs(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsCarouselImg(): //轮播-图片模式
			tempCarousel := vk.Carousel
			if tempCarousel == nil || len(tempCarousel.List) == 0 {
				continue
			}
			eg.Go(func(ctx context.Context) error {
				ck := s.formatCarouselImg(tmpModu, tempCarousel)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsCarouselWord(): //轮播-文字模式
			tempCarousel := vk.Carousel
			if tempCarousel == nil || len(tempCarousel.List) == 0 {
				continue
			}
			eg.Go(func(ctx context.Context) error {
				ck := s.formatCarouselWord(tmpModu, tempCarousel)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsCarouselSource(): //轮播-数据源模式
			eg.Go(func(ctx context.Context) error {
				ck := s.formatCarouselSource(ctx, tmpModu, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsIcon(): //图标组件
			tempIcon := vk.Icon
			if tempIcon == nil || len(tempIcon.List) == 0 {
				continue
			}
			eg.Go(func(ctx context.Context) error {
				ck := s.formatIcon(tmpModu, tempIcon)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsEditorOrigin(): //编辑推荐卡数据源模式:每周必看，入站必刷,排行榜
			eg.Go(func(ctx context.Context) error {
				ck := s.formatEditorOrigin(ctx, tmpModu, mid, buvid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsEditor(): //编辑推荐卡
			eg.Go(func(ctx context.Context) error {
				ck := s.formatEditor(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsVote(): //投票组件
			temClick := vk.Click
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatVote(ctx, tmpModu, temClick, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsClick(): //自定义点击组件
			temClick := vk.Click
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatClick(ctx, tmpModu, temClick, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsInlineTab(): //inline tab组件
			inlineTabs := vk.InlineTab
			eg.Go(func(ctx context.Context) error {
				ck := s.formatInlineTab(ctx, tmpModu, inlineTabs)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsSelect(): // 筛选组件
			selects := vk.Select
			eg.Go(func(ctx context.Context) error {
				ck := s.formatSelect(ctx, tmpModu, selects)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsReply(): //评论组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatReply(tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsLive(): //直播卡组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatLive(tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsStatement(): // 文本组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatStatement(tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNewVideoAct(): //新视频卡-活动数据源组件
			sortType := int32(0)
			if vk.VideoAct != nil && len(vk.VideoAct.SortList) > 0 {
				sortType = int32(vk.VideoAct.SortList[0].SortType)
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatNewVideoAct(tmpModu, sortType)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNewVideoDyn(): //新视频卡-动态模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.fromNewVideoDynamic(tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsOgvSeasonSource(): //ogv剧集-资源类型
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatOgvSeasonResource(ctx, tmpModu, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsOgvSeasonID(): //ogv剧集-ids
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatOgvSeasonID(ctx, tmpModu, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNavigation(): // 导航组件
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatNavigation(tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNewVideoID(): // 新视频卡-avid模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatNewVideoAvid(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsResourceDyn(): //资源小卡，动态组件
			tempDynamic := vk.Dynamic
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatResourceDyn(tmpModu, tempDynamic)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsResourceOrigin(): //资源小卡-外接数据源
			sortType := int64(0)
			if vk.VideoAct != nil && len(vk.VideoAct.SortList) > 0 {
				sortType = vk.VideoAct.SortList[0].SortType
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatResourceOrigin(tmpModu, sortType)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsResourceAct(): //资源小卡，活动组件
			sortType := int32(0)
			if vk.VideoAct != nil && len(vk.VideoAct.SortList) > 0 {
				sortType = int32(vk.VideoAct.SortList[0].SortType)
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatResourceAct(tmpModu, sortType)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsResourceID(): // 资源小卡-avid模式
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatResourceAvid(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsResourceRole(): // 资源小卡-角色剧集模式
			eg.Go(func(ctx context.Context) error {
				ck := s.formatResourceRole(tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsDynamic():
			dyn := vk.Dynamic
			eg.Go(func(ctx context.Context) error {
				ck := s.formatDynamic(tmpModu, dyn, page)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsVideo():
			var sortType int64
			if vk.VideoAct != nil && len(vk.VideoAct.SortList) > 0 {
				sortType = vk.VideoAct.SortList[0].SortType
			}
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.formatVideo(tmpModu, sortType)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsProgress():
			eg.Go(func(ctx context.Context) error {
				ck := s.formatProgress(ctx, tmpModu, mid)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNewactHeaderModule():
			eg.Go(func(ctx context.Context) error {
				ck := s.formatNewactHeader(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNewactAwardModule():
			eg.Go(func(ctx context.Context) error {
				ck := s.formatNewactAward(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsNewactStatementModule():
			eg.Go(func(ctx context.Context) error {
				ck := s.formatNewactStatement(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsMatchMedal():
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatMatchMedal(ctx, tmpModu)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		case tmpModu.IsMatchEvent():
			event := vk.MatchEvent
			eg.Go(func(ctx context.Context) (e error) {
				ck := s.FormatMatchEvent(ctx, tmpModu, event)
				if ck != nil {
					mu.Lock()
					rly.Card[tmpModu.ID] = ck
					mu.Unlock()
				}
				return nil
			})
		default:
			continue
		}
	}
}

func (s *Service) formatResourceRole(mou *api.NativeModule) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromResourceRoleModule(mou)
	return list
}

func (s *Service) formatNavigation(mou *api.NativeModule) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromNavigation(mou)
	return list
}

func (s *Service) formatNewVideoAvid(c context.Context, mou *api.NativeModule) *dynmdl.Item {
	likeList, err := s.ModuleMixExts(c, mou.ID, 0, mou.Num+6)
	if err != nil || likeList == nil {
		log.Error("s.ModuleMixExts(%v) error(%v)", mou.ID, err)
		return nil
	}
	var aids []*dynmdl.ResourcesIDs
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		switch v.MType {
		case api.MixAvidType, api.MixEpidType:
			tmp := &dynmdl.ResourcesIDs{ID: v.ForeignID, Type: v.MType}
			aids = append(aids, tmp)
		default:
			continue
		}
	}
	if len(aids) == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromNewIDsModule(mou, aids, likeList.HasMore == 1)
	return list
}

func (s *Service) formatResourceAvid(c context.Context, mou *api.NativeModule) *dynmdl.Item {
	likeList, err := s.ModuleMixExts(c, mou.ID, 0, mou.Num+6)
	if err != nil || likeList == nil {
		log.Error("s.ModuleMixExts(%v) error(%v)", mou.ID, err)
		return nil
	}
	var aids []*dynmdl.ResourcesIDs
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		tmp := &dynmdl.ResourcesIDs{}
		switch v.MType {
		case api.MixAvidType, api.MixCvidType, api.MixEpidType, api.MixLive:
			tmp.ID = v.ForeignID
			tmp.Type = v.MType
		case api.MixFolder:
			tmp.ID = v.ForeignID
			tmp.Type = v.MType
			if v.Reason != "" {
				mixFold := &dynmdl.MixFolder{}
				if err := json.Unmarshal([]byte(v.Reason), mixFold); err == nil {
					tmp.FID = mixFold.Fid
				} else {
					continue
				}
			}
		default:
			continue
		}
		aids = append(aids, tmp)
	}

	if len(aids) == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromResourceIDsModule(mou, aids, likeList.HasMore == 1)
	return list
}

// ogvSeasonResource .
func (s *Service) ogvSeasonResource(c context.Context, mou *api.NativeModule, ps, offset, mid int64) (*dynmdl.IDsReply, error) {
	// 获取ogv 片单信息 平台信息
	treply, err := s.pgcDao.SeasonByPlayId(c, int32(mou.Fid), int32(offset), int32(ps), mid)
	if err != nil {
		log.Error("s.pgcdao.SeasonByPlayI(%d) error(%v)", mou.Fid, err)
		return nil, err

	}
	if treply == nil {
		return nil, xecode.NothingFound
	}
	rly := &dynmdl.IDsReply{Offset: int64(treply.NexOffset)}
	if treply.HasNext {
		rly.HasMore = 1
	}
	if len(treply.SeasonInfos) == 0 {
		return rly, nil
	}
	for _, v := range treply.SeasonInfos {
		if v == nil {
			continue
		}
		tmp := &dynmdl.Item{}
		tmp.FromOgvSeason(mou, v, "")
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// ogvSeasonID .
func (s *Service) ogvSeasonID(c context.Context, mou *api.NativeModule, ps, offset, mid int64) (*dynmdl.IDsReply, error) {
	likeList, err := s.ModuleMixExts(c, mou.ID, offset, ps+6)
	if err != nil {
		log.Error(" s.ModuleMixExts(%d) error(%v)", mou.ID, err)
		return nil, err
	}
	if likeList == nil {
		return nil, xecode.NothingFound
	}
	rly := &dynmdl.IDsReply{Offset: likeList.Offset, HasMore: likeList.HasMore}
	var ssids []int32
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 || v.MType != api.MixOgvSsid {
			continue
		}
		ssids = append(ssids, int32(v.ForeignID))
	}
	if len(ssids) == 0 {
		return rly, nil
	}
	//根据ssid获取ogv卡片信息
	seaRly, err := s.pgcDao.SeasonBySeasonId(c, ssids, mid)
	if err != nil {
		log.Error("s.pgcdao.SeasonBySeasonId %v,error(%v)", ssids, err)
		//降级处理，不返回错误
		return rly, nil
	}
	toCount := 0
	for _, v := range likeList.List {
		offset++
		if v == nil || v.ForeignID == 0 || v.MType != api.MixOgvSsid {
			continue
		}
		if sVal, ok := seaRly[int32(v.ForeignID)]; !ok || sVal == nil {
			continue
		}
		tmp := &dynmdl.Item{}
		tmp.FromOgvSeason(mou, seaRly[int32(v.ForeignID)], v.RemarkUnmarshal().Title)
		rly.List = append(rly.List, tmp)
		toCount++
		if toCount >= int(ps) {
			break
		}
	}
	if likeList.HasMore == 0 && offset < likeList.Offset {
		rly.HasMore = 1
	}
	rly.Offset = offset
	return rly, nil
}

// formatOgvSeasonID .
func (s *Service) formatOgvSeasonID(c context.Context, mou *api.NativeModule, mid int64) (list *dynmdl.Item) {
	rly, err := s.ogvSeasonID(c, mou, mou.Num, 0, mid)
	if err != nil || rly == nil {
		log.Error("s.TimelineIDs error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ogvSeasonJoin(rly, mou)
	return
}

// FormatOgvSeasonResource .
func (s *Service) formatOgvSeasonResource(c context.Context, mou *api.NativeModule, mid int64) (list *dynmdl.Item) {
	rly, err := s.ogvSeasonResource(c, mou, mou.Num, 0, mid)
	if err != nil || rly == nil {
		log.Error("s.ogvSeasonResource error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.ogvSeasonJoin(rly, mou)
	return
}

// ogvSeasonJoin .
func (s *Service) ogvSeasonJoin(req *dynmdl.IDsReply, mou *api.NativeModule) *dynmdl.Item {
	// 有卡片信息才下发组件
	if req == nil || len(req.List) == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromOgvSeasonModule(mou)
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &dynmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	list.Item = append(list.Item, req.List...)
	if req.HasMore > 0 && mou.IsAttrHideMore() != api.AttrModuleYes {
		tmpMore := &dynmdl.Item{}
		tmpMore.FromOgvSeasonMore(mou)
		list.Item = append(list.Item, tmpMore)
	}
	return list
}

func (s *Service) fromNewVideoDynamic(mou *api.NativeModule) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromNewDynModule(mou)
	return list
}

// FormatResourceDyn .
func (s *Service) formatResourceDyn(mou *api.NativeModule, dyn *api.Dynamic) *dynmdl.Item {
	types := int32(8) //默认排序
	if dyn != nil && len(dyn.SelectList) > 0 {
		types = int32(dyn.SelectList[0].SelectType)
	}
	list := &dynmdl.Item{}
	list.FromResourceModule(mou, types)
	return list
}

// FormatResourceAct .
func (s *Service) formatResourceAct(mou *api.NativeModule, sortType int32) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromResourceActModule(mou, sortType)
	return list
}

// FormatResourceOrigin .
func (s *Service) formatResourceOrigin(mou *api.NativeModule, sortType int64) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromResourceOriginModule(mou, sortType)
	return list
}

func (s *Service) formatNewVideoAct(mou *api.NativeModule, sortType int32) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromNewVideoActModule(mou, sortType)
	return list
}

func (s *Service) formatStatement(mou *api.NativeModule) *dynmdl.Item {
	list := &dynmdl.Item{}
	list.FromStatementModule(mou)
	return list
}

// formatReply .
func (s *Service) formatReply(mou *api.NativeModule) *dynmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromReplyModule(mou)
	return list
}

// formatLive .
func (s *Service) formatLive(mou *api.NativeModule) *dynmdl.Item {
	nowTime := time.Now().Unix()
	// 在设置的时间之内
	if mou.Stime > nowTime || mou.Etime < nowTime || mou.Fid == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromLiveModule(mou)
	return list
}

// formatSelect .
func (s *Service) formatSelect(ctx context.Context, mou *api.NativeModule, select_ *api.Select) *dynmdl.Item {
	if select_ == nil || len(select_.List) == 0 {
		return nil
	}
	var (
		pageIDs           []int64
		defTab, timingTab int64
		nowTime           = time.Now().Unix()
	)
	ext := make(map[int64]*api.MixReason)
	for _, v := range select_.List {
		if v == nil || v.MType != api.MixInlineType || v.ForeignID == 0 || !v.IsOnline() {
			continue
		}
		pageIDs = append(pageIDs, v.ForeignID)
		ext[v.ForeignID] = v.RemarkUnmarshal()
		//寻找默认tab
		if ext[v.ForeignID].DefType == api.DefTypeTimely { //立即生效的时间
			defTab = v.ForeignID
		} else if ext[v.ForeignID].DefType == api.DefTypeTiming { //定时生效的时间
			if ext[v.ForeignID].DStime <= nowTime && ext[v.ForeignID].DEtime > nowTime {
				timingTab = v.ForeignID
			}
		}
		//寻找默认tab
	}
	// 默认tab优先级,若立即生效的时间，与定时生效的时间一致，则优先以定时生效的为准
	if timingTab == 0 {
		timingTab = defTab
	}
	//寻找默认tab
	if len(pageIDs) == 0 {
		return nil
	}
	pagesInfo, e := s.natDao.NativePages(ctx, pageIDs)
	if e != nil {
		log.Error("s.actDao.NativePages %v error(%v)", pageIDs, e)
		return nil
	}
	var (
		currentIndex int32
	)
	tmpItem := &dynmdl.Item{}
	tmpItem.FormatSelect(mou)
	for _, v := range pageIDs {
		if val, ok := pagesInfo[v]; !ok || val == nil || !val.IsOnline() || val.Title == "" {
			continue
		}
		eVal := ext[v]
		tmpID := &dynmdl.Item{ItemID: v, Title: pagesInfo[v].Title, CurrentTab: eVal.JoinCurrentTab()}
		//  有默认tab
		if v == timingTab {
			tmpItem.CurrentTabIndex = currentIndex
		}
		currentIndex++
		//查找默认tab end
		tmpItem.Item = append(tmpItem.Item, tmpID)
	}
	if len(tmpItem.Item) == 0 {
		return nil
	}
	var items []*dynmdl.Item
	items = append(items, tmpItem)
	first := &dynmdl.Item{}
	first.FromSelectModule(mou, items)
	return first
}

// formatInlineTab .
// nolint:gocognit
func (s *Service) formatInlineTab(c context.Context, mou *api.NativeModule, inline *api.InlineTab) *dynmdl.Item {
	if inline == nil || len(inline.List) == 0 {
		return nil
	}
	var (
		pageIDs           []int64
		defTab, timingTab int64
		nowTime           = time.Now().Unix()
	)
	ext := make(map[int64]*api.MixReason)
	for _, v := range inline.List {
		if v == nil || v.MType != api.MixInlineType || v.ForeignID == 0 || !v.IsOnline() {
			continue
		}
		pageIDs = append(pageIDs, v.ForeignID)
		ext[v.ForeignID] = v.RemarkUnmarshal()
		//寻找默认tab
		if ext[v.ForeignID].DefType == api.DefTypeTimely { //立即生效的时间
			defTab = v.ForeignID
		} else if ext[v.ForeignID].DefType == api.DefTypeTiming { //定时生效的时间
			if ext[v.ForeignID].DStime <= nowTime && ext[v.ForeignID].DEtime > nowTime {
				timingTab = v.ForeignID
			}
		}
		//寻找默认tab
	}
	// 默认tab优先级,若立即生效的时间，与定时生效的时间一致，则优先以定时生效的为准
	if timingTab == 0 {
		timingTab = defTab
	}
	//寻找默认tab
	if len(pageIDs) == 0 {
		return nil
	}
	pagesInfo, e := s.natDao.NativePages(c, pageIDs)
	if e != nil {
		log.Error("s.actDao.NativePages %v error(%v)", pageIDs, e)
		return nil
	}
	tmpItem := &dynmdl.Item{}
	tmpItem.FormatInline(mou)
	var currentIndex int32
	for _, v := range pageIDs {
		if val, ok := pagesInfo[v]; !ok || val == nil || !val.IsOnline() || val.Title == "" {
			continue
		}
		var (
			hasLock bool
		)
		eVal := ext[v]
		tmpID := &dynmdl.Item{ItemID: v, Title: pagesInfo[v].Title, CurrentTab: eVal.JoinCurrentTab()}
		lockExt := pagesInfo[v].ConfSetUnmarshal()
		if lockExt.DT == api.NeedUnLock { //解锁模式
			var deblocking bool
			if lockExt.DC == api.UnLockTime && lockExt.Stime <= time.Now().Unix() { //时间模式&&到达解锁时间
				deblocking = true
			}
			//未解锁时
			if !deblocking {
				// 不可点击
				if lockExt.UnLock == api.NotClick {
					tmpID.ItemID = 0 //不下发pageid
					tmpID.Setting = &dynmdl.Setting{UnAllowClick: true}
					tmpID.Content = "还未解锁，敬请期待"
					if lockExt.Tip != "" {
						tmpID.Content = lockExt.Tip //提示文案
					}
					hasLock = true
				} else { //不认识类型 || 未解锁：不展示 || 不可点击下,低版本
					continue
				}
			}
		}
		//组件是图片模式
		if mou.AvSort == 1 && eVal != nil {
			//锁定状态时图片展示未解锁态
			if hasLock {
				tmpID.ImagesUnion = &dynmdl.ImagesUnion{
					UnSelect: dynmdl.ImageChange(eVal.UnI), //未选中
					Select:   dynmdl.ImageChange(eVal.SI),  //选中
				}
			} else {
				tmpID.ImagesUnion = &dynmdl.ImagesUnion{
					Select:   dynmdl.ImageChange(eVal.SI),   //选中
					UnSelect: dynmdl.ImageChange(eVal.UnSI), //未选中
				}
			}
		}
		//   没有锁定 && 有默认tab
		if !hasLock && v == timingTab {
			tmpItem.CurrentTabIndex = currentIndex
		}
		//查找默认tab end
		currentIndex++
		tmpItem.Item = append(tmpItem.Item, tmpID)
	}
	if len(tmpItem.Item) == 0 {
		return nil
	}
	var items []*dynmdl.Item
	items = append(items, tmpItem)
	first := &dynmdl.Item{}
	first.FromInlineTabModule(mou, items)
	return first
}

func (s *Service) formatEditor(c context.Context, mou *api.NativeModule) *dynmdl.Item {
	likeList, err := s.ModuleMixExts(c, mou.ID, 0, mou.Num+6)
	if err != nil || likeList == nil {
		log.Error("s.ModuleMixExts(%v) error(%v)", mou.ID, err)
		return nil
	}
	var aids []*dynmdl.ResourcesIDs
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		mixFold := &dynmdl.MixFolder{}
		if v.Reason != "" {
			//发生错误降级处理
			if e := json.Unmarshal([]byte(v.Reason), mixFold); e != nil {
				log.Error("formatEditor json.Unmarshal(%s) error(%v)", v.Reason, e)
			}
		}
		tmp := &dynmdl.ResourcesIDs{RcmdContent: mixFold.RcmdContent}
		switch v.MType {
		case api.MixAvidType, api.MixCvidType, api.MixEpidType:
			tmp.ID = v.ForeignID
			tmp.Type = v.MType
		case api.MixFolder:
			tmp.ID = v.ForeignID
			tmp.Type = v.MType
			tmp.FID = mixFold.Fid
		default:
			continue
		}
		aids = append(aids, tmp)
	}
	if len(aids) == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromEditorModule(mou, aids, likeList.HasMore == 1)
	return list
}

// formatTimelineIDs .
func (s *Service) formatTimelineIDs(c context.Context, mou *api.NativeModule) (list *dynmdl.Item) {
	confSort := mou.ConfUnmarshal()
	//默认浮层
	ps := mou.Num
	if confSort.MoreSort == api.MoreExpand { // 下拉展示
		ps = 50
	}
	rly, err := s.timelineIDs(c, mou, ps, 0)
	if err != nil || rly == nil {
		log.Error("s.TimelineIDs error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.timelineJoin(rly, mou)
	return
}

// timelineIDs .
func (s *Service) timelineIDs(c context.Context, mou *api.NativeModule, ps, offset int64) (*dynmdl.TimelineSourceReply, error) {
	likeList, err := s.ModuleMixExts(c, mou.ID, offset, ps+6)
	if err != nil {
		log.Error(" s.ModuleMixExts(%v) error(%v)", mou.ID, err)
		return nil, err
	}
	if likeList == nil {
		return nil, xecode.NothingFound
	}
	rly := &dynmdl.TimelineSourceReply{Offset: int32(likeList.Offset), HasMore: likeList.HasMore}
	lg := len(likeList.List)
	if lg == 0 {
		return rly, nil
	}
	var aids, cvids []int64
	for _, v := range likeList.List {
		if v == nil || v.ForeignID == 0 {
			continue
		}
		switch v.MType {
		case api.MixAvidType:
			aids = append(aids, v.ForeignID)
		case api.MixCvidType:
			cvids = append(cvids, v.ForeignID)
		}
	}
	arcRly, artRly, _, _, _ := s.getResource(c, aids, cvids, nil, nil, nil, mou.Attribute)
	toCount := 0
	confSort := mou.ConfUnmarshal()
	for _, v := range likeList.List {
		offset++
		if v == nil {
			continue
		}
		mixRemark := v.RemarkUnmarshal()
		var title string
		//时间轴节点类型 0:文本 1:时间节点
		if confSort.Axis == api.AxisText {
			title = mixRemark.Name
		} else {
			title = dynmdl.FromTimelineFormatHead(xtime.Time(mixRemark.Stime), confSort.TimeSort)
		}
		tmp := &dynmdl.Item{}
		switch v.MType {
		case api.MixAvidType:
			if v.ForeignID == 0 {
				continue
			}
			if va, ok := arcRly[v.ForeignID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromTimelineArc(arcRly[v.ForeignID], title)
		case api.MixCvidType:
			if v.ForeignID == 0 {
				continue
			}
			if va, ok := artRly[v.ForeignID]; !ok || va == nil || !va.IsNormal() {
				continue
			}
			tmp.FromTimelineArt(artRly[v.ForeignID], title)
		case api.MixTimelineText:
			tmp.FromTimelineText(mixRemark, title)
		case api.MixTimelinePic:
			tmp.FromTimelinePic(mixRemark, title)
		case api.MixTimeline:
			tmp.FromTimeline(mixRemark, title)
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
		toCount++
		if toCount >= int(ps) {
			break
		}
	}
	if likeList.HasMore == 0 && offset < likeList.Offset {
		rly.HasMore = 1
	}
	rly.Offset = int32(offset)
	return rly, nil
}

func (s *Service) natCardsOfActPage(c context.Context, pageIDs []int64) map[int64]*api.NativePageCard {
	if len(pageIDs) == 0 {
		return map[int64]*api.NativePageCard{}
	}
	pages, err := s.natDao.NativePages(c, pageIDs)
	if err != nil {
		log.Error("s.natDao.NativePages(%v) error(%v)", pageIDs, err)
		return map[int64]*api.NativePageCard{}
	}
	// NativePageCards返回上线活动，跳转优先级为：配置跳转链接 > 活动聚合页 > 单个活动页
	// NativeAllPages返回剩余的活动（下线/NativePageCards失败），跳转优先级为 新频道页 > 旧频道普通话题页
	// 获取频道数据
	chanIDs := make([]int64, 0)
	//获取低栏地址
	tabPageIDs := make([]int64, 0)
	for _, v := range pages {
		if v == nil {
			continue
		}
		if v.IsOnline() {
			if v.SkipURL == "" {
				tabPageIDs = append(tabPageIDs, v.ID)
			}
		} else {
			chanIDs = append(chanIDs, v.ForeignID)
		}
	}
	eg := errgroup.WithContext(c)
	var chanInfos map[int64]*channelapi.Channel
	if len(chanIDs) > 0 {
		eg.Go(func(c context.Context) error {
			chanRly, _ := s.channelClient.Infos(c, &channelapi.InfosReq{Cids: chanIDs})
			if chanRly != nil {
				chanInfos = chanRly.CidMap
			}
			return nil
		})
	}
	//没有需要处理的跳转地址
	var tabRly map[int64]*api.PagesTab
	if len(tabPageIDs) > 0 {
		eg.Go(func(c context.Context) error {
			var e error
			if tabRly, e = s.nativeTab(c, tabPageIDs, api.TopicActType); e != nil { //降级处理,错误忽略
				log.Error("s.nativeTab(%v) error(%v)", tabPageIDs, e)
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil { //降级处理,错误忽略
		log.Error("natCardsOfActPage eg.Wait() error(%v)", err)
	}
	// 组装数据
	res := make(map[int64]*api.NativePageCard, len(pageIDs))
	var (
		ctypeOne = int32(1)
		ctypeTwo = int32(2)
	)
	for _, pid := range pageIDs {
		page, ok := pages[pid]
		if !ok || page == nil {
			continue
		}
		skipUrl := page.SkipURL
		pcUrl := page.PcURL
		if page.IsOnline() {
			if page.SkipURL == "" {
				skipUrl = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", page.ID, time.Now().Unix())
				if tInfo, tok := tabRly[page.ID]; tok && tInfo != nil {
					skipUrl = tInfo.Url
				}
				page.ShareURL = skipUrl
			}
			if page.PcURL == "" {
				pcUrl = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", page.ID, time.Now().Unix())
			}
		} else { //下线的活动
			chanInfo, ok := chanInfos[page.ForeignID]
			if !ok || chanInfo == nil {
				continue
			}
			switch chanInfo.GetCType() {
			case ctypeOne:
				skipUrl = fmt.Sprintf("bilibili://pegasus/channel/%d?type=topic", page.ForeignID)
			case ctypeTwo:
				skipUrl = fmt.Sprintf("bilibili://pegasus/channel/v2/%d?tab=topic", page.ForeignID)
			default:
				continue
			}
		}
		// 分享title为空则取话题名
		if page.ShareCaption == "" {
			page.ShareCaption = page.Title
		}
		res[pid] = &api.NativePageCard{
			Id:           page.ID,
			Title:        page.Title,
			Type:         page.Type,
			ForeignID:    page.ForeignID,
			ShareTitle:   page.ShareTitle,
			ShareImage:   page.ShareImage,
			ShareURL:     page.ShareURL,
			SkipURL:      skipUrl,
			RelatedUid:   page.RelatedUid,
			PcURL:        pcUrl,
			ShareCaption: page.ShareCaption,
		}
	}
	return res
}

func (s *Service) formatCapsule(c context.Context, mou *api.NativeModule, actPage *api.ActPage, primaryPageID int64, opFrom string) *dynmdl.Item {
	// 过滤掉自己
	var pageIDs []int64
	for _, v := range actPage.List {
		if v.PageID == primaryPageID && opFrom != dynmdl.FormatModFromMenuUp {
			continue
		}
		pageIDs = append(pageIDs, v.PageID)
	}
	cards := s.natCardsOfActPage(c, pageIDs)
	if len(cards) == 0 {
		return nil
	}
	capsule := &dynmdl.Item{
		Goto:  dynmdl.GotoActCapsule,
		Title: mou.Caption,
	}
	for _, v := range pageIDs {
		card, ok := cards[v]
		if !ok {
			continue
		}
		item := &dynmdl.Item{}
		item.FromActCapsuleItem(card)
		capsule.Item = append(capsule.Item, item)
	}
	if len(capsule.Item) == 0 {
		return nil
	}
	capsuleMod := &dynmdl.Item{}
	capsuleMod.FromActCapsuleModule(mou, []*dynmdl.Item{capsule})
	return capsuleMod
}

// rcmdSourceData .
func (s *Service) rcmdSourceData(c context.Context, mou *api.NativeModule) []int64 {
	var (
		fids []int64
	)
	confSort := mou.ConfUnmarshal()
	switch confSort.SourceType {
	case api.SourceTypeActUp:
		sortType := confSort.SortType
		if sortType == "" {
			sortType = api.SortTypeCtime
		}
		upList, err := s.actDao.UpList(c, mou.Fid, 1, 40, 0, sortType)
		if err != nil {
			log.Error(" s.actDao.UpList(%d,%s) error(%v)", mou.Fid, sortType, err)
			return []int64{}
		}
		if upList == nil {
			return []int64{}
		}
		for _, v := range upList.List {
			if v.Item == nil {
				continue
			}
			if v.Item.Wid > 0 {
				fids = append(fids, v.Item.Wid)
			}
		}
	default:
		return []int64{}
	}
	return fids
}

func (s *Service) formatRcmdVertical(c context.Context, mou *api.NativeModule, recom *api.Recommend, mid int64) *dynmdl.Item {
	var (
		fids      []int64
		followRly map[int64]*relationapi.FollowingReply
		cards     map[int64]*accgrpc.Card
		recomList = make([]*api.NativeMixtureExt, 0)
	)
	if mou.IsRcmdVerticalSource() {
		fids = s.rcmdSourceData(c, mou)
		for _, v := range fids {
			recomList = append(recomList, &api.NativeMixtureExt{ForeignID: v})
		}
	} else {
		if recom == nil || len(recom.List) == 0 {
			return nil
		}
		recomList = recom.List
		for _, v := range recomList {
			if v.ForeignID > 0 {
				fids = append(fids, v.ForeignID)
			}
		}
	}
	if len(fids) == 0 {
		return nil
	}
	eg := errgroup.WithContext(c)
	// 获取关注关系
	if mid > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if followRly, e = s.relDao.RelationsGRPC(ctx, mid, fids); e != nil {
				log.Error(" s.reldao.RelationsGRPC(%d,%v) error(%v)", mid, fids, e)
				e = nil
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		//获取用户信息
		if cards, e = s.accDao.Cards3GRPC(ctx, fids); e != nil {
			log.Error("s.accDao.Cards3GRPC(%v) error(%v)", fids, e)
			e = nil
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("formatRcmdVertical eg.Wait() error(%v)", err)
		return nil
	}
	var items []*dynmdl.Item
	for _, v := range recomList {
		if _, ok := cards[v.ForeignID]; !ok {
			continue
		}
		clickExt := &dynmdl.ClickExt{}
		if relation, ok := followRly[v.ForeignID]; ok && relation != nil {
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			if relation.Attribute == 2 || relation.Attribute == 6 {
				clickExt.IsFollow = true
			}
		}
		item := &dynmdl.Item{}
		item.FromRcmdVerticalItem(v, cards[v.ForeignID], clickExt)
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	rcmd := &dynmdl.Item{}
	rcmd.FromRcmdVertical(items)
	rcmdModule := &dynmdl.Item{}
	rcmdModule.FromRcmdVerticalModule(mou)
	if mou.Meta != "" {
		titleImage := &dynmdl.Item{}
		titleImage.FromTitleImage(mou)
		rcmdModule.Item = append(rcmdModule.Item, titleImage)
	}
	rcmdModule.Item = append(rcmdModule.Item, rcmd)
	return rcmdModule
}

// nolint:gocognit
func (s *Service) reserveInfo(c context.Context, ids []int64, mid, needIcon int64) map[int64]*dynmdl.ReserveRly {
	revRly, err := s.actDao.UpActReserveRelationInfo(c, mid, ids)
	if err != nil {
		log.Error("s.actDao.UpActReserveRelationInfo(%d,%v) error(%v)", mid, ids, err)
		return make(map[int64]*dynmdl.ReserveRly)
	}
	var (
		aidStr    []string
		mids      []int64
		nowtime   = time.Now().Unix()
		liveIDStr = make(map[int64][]string)
	)
	rly := make(map[int64]*dynmdl.ReserveRly)
	for _, v := range ids {
		if rVal, ok := revRly[v]; !ok || rVal == nil {
			continue
		}
		//话题活动页展示都为客态逻辑
		if revRly[v].UpActVisible != actGRPC.UpActVisible_DefaultVisible {
			continue
		}
		var changeType int64
		switch revRly[v].State {
		case actGRPC.UpActReserveRelationState_UpReserveRelated, actGRPC.UpActReserveRelationState_UpReserveRelatedOnline:
			changeType = dynmdl.ReserveDisplayA
			if revRly[v].Type == actGRPC.UpActReserveRelationType_Course && int64(revRly[v].Etime) < nowtime {
				//预约结束未核销
				changeType = actmdl.ReserveDisplayE
			}
		case actGRPC.UpActReserveRelationState_UpReserveRelatedWaitCallBack, actGRPC.UpActReserveRelationState_UpReserveRelatedCallBackCancel, actGRPC.UpActReserveRelationState_UpReserveRelatedCallBackDone:
			switch revRly[v].Type {
			case actGRPC.UpActReserveRelationType_Archive:
				changeType = dynmdl.ReserveDisplayC
				aidStr = append(aidStr, revRly[v].Oid)
			case actGRPC.UpActReserveRelationType_Live:
				changeType = dynmdl.ReserveDisplayLive
				liveIDStr[revRly[v].Upmid] = append(liveIDStr[revRly[v].Upmid], revRly[v].Oid)
			case actGRPC.UpActReserveRelationType_Course:
				changeType = actmdl.ReserveDisplayC
			default: //不认识的类型，不展示对应卡片
				continue
			}
		default: //不认识的类型，不展示对应卡片
			continue
		}
		mids = append(mids, revRly[v].Upmid)
		rly[v] = &dynmdl.ReserveRly{
			ChangeType: changeType,
			Item:       revRly[v],
		}
	}
	var accRly map[int64]*accgrpc.Card
	eg := errgroup.WithContext(c)
	//获取账号信息
	if len(mids) > 0 && needIcon == 1 {
		eg.Go(func(ctx context.Context) error {
			var e error
			if accRly, e = s.accDao.Cards3GRPC(ctx, mids); e != nil { //获取账号信息失败，降级处理
				log.Error("s.accDao.Cards3GRPC(%v) error(%v)", mids, e)
			}
			return nil
		})
	}
	var (
		aids   []int64
		aidMap = make(map[int64]struct{})
	)
	for _, v := range aidStr {
		ak, err := strconv.ParseInt(v, 10, 64)
		if err != nil || ak <= 0 {
			continue
		}
		if _, ok := aidMap[ak]; !ok { //去重
			aids = append(aids, ak)
			aidMap[ak] = struct{}{}
		}
	}
	//获取稿件信息
	arcRly := make(map[string]*arccli.Arc)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcRes, e := s.arcClient.Arcs(ctx, &arccli.ArcsRequest{Aids: aids})
			if e != nil { //获取账号信息失败，降级处理
				log.Error("s.accDao.Cards3GRPC(%v) error(%v)", mids, e)
				return nil
			}
			for _, v := range arcRes.GetArcs() {
				arcRly[fmt.Sprintf("%d", v.Aid)] = v
			}
			return nil
		})
	}
	//获取直播信息
	var liveRly map[int64]*roomgategrpc.SessionInfos
	if len(liveIDStr) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if liveRly, e = s.liveDao.SessionInfoBatch(ctx, liveIDStr, []string{dynmdl.LiveEnterFrom}); e != nil {
				log.Error("s.liveDao.SessionInfoBatch(%v) error(%v)", liveIDStr, e)
				e = nil
			}
			return
		})
	}
	_ = eg.Wait()
	lastRly := make(map[int64]*dynmdl.ReserveRly)
	for k, v := range rly {
		if v == nil || v.Item == nil {
			continue
		}
		switch v.Item.Type {
		case actGRPC.UpActReserveRelationType_Archive:
			if aVal, ok := arcRly[v.Item.Oid]; ok && aVal.IsNormal() {
				v.Arc = aVal
			}
		case actGRPC.UpActReserveRelationType_Live:
			if acVal, ok := liveRly[v.Item.Upmid]; ok && acVal != nil {
				if seVal, k := acVal.SessionInfoPerLive[v.Item.Oid]; k && seVal != nil {
					v.Live = &dynmdl.LiveInfos{
						RoomId:             acVal.RoomId,
						Uid:                acVal.Uid,
						JumpUrl:            acVal.JumpUrl,
						Title:              acVal.Title,
						SessionInfoPerLive: seVal,
					}
				}
			}
		default:
		}
		if needIcon == 1 {
			if actVal, ok := accRly[v.Item.Upmid]; ok {
				v.Account = actVal
			}
		}
		v.DisplayType = v.ChangeType
		//容错:类型live
		if v.ChangeType == dynmdl.ReserveDisplayLive {
			//不可回放不可见 Status:0 不再直播也没有回放
			if v.Live == nil || v.Live.SessionInfoPerLive == nil {
				continue
			}
			switch v.Live.SessionInfoPerLive.Status {
			case dynmdl.Living:
				v.DisplayType = dynmdl.ReserveDisplayC
			case dynmdl.LiveEnd:
				v.DisplayType = dynmdl.ReserveDisplayD
			default:
				v.DisplayType = dynmdl.ReserveDisplayE
			}
		}
		lastRly[k] = v
	}
	return lastRly
}

func (s *Service) formatReserve(c context.Context, mou *api.NativeModule, rev *api.Reserve, mid int64) (ck *dynmdl.Item) {
	if rev == nil { //没有卡片，不下发组件
		return
	}
	var revIDs []int64
	//获取游戏id
	for _, v := range rev.List {
		if v == nil || v.MType != api.MixUpReserve || v.ForeignID <= 0 {
			continue
		}
		revIDs = append(revIDs, v.ForeignID)
	}
	if len(revIDs) == 0 { //没有卡片，不下发组件
		return
	}
	//获取游戏详情
	revRly := s.reserveInfo(c, revIDs, mid, mou.IsAttrIsDisplayUpIcon())
	var items []*dynmdl.Item
	//拼接卡片信息
	for _, v := range rev.List {
		if v == nil || v.MType != api.MixUpReserve || v.ForeignID <= 0 {
			continue
		}
		if gv, ok := revRly[v.ForeignID]; !ok || gv == nil {
			continue
		}
		itemTep := &dynmdl.Item{}
		itemTep.FromReserveExt(v, revRly[v.ForeignID], mou.IsAttrIsDisplayUpIcon(), mid)
		items = append(items, itemTep)
	}
	if len(items) == 0 { //没有卡片，不下发组件
		return
	}
	ck = &dynmdl.Item{}
	var lastItems []*dynmdl.Item
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &dynmdl.Item{}
		tmpName.FromTitleName(mou)
		lastItems = append(lastItems, tmpName)
	}
	lastItems = append(lastItems, items...)
	ck.FromReserve(mou, lastItems)
	return
}

func (s *Service) formatGame(c context.Context, mou *api.NativeModule, games *api.Game, mid int64, mobiApp string) (ck *dynmdl.Item) {
	if games == nil { //没有卡片，不下发组件
		return
	}
	var gameIDs []int64
	//获取游戏id
	for _, v := range games.List {
		if v == nil || v.MType != api.MixGame || v.ForeignID <= 0 {
			continue
		}
		gameIDs = append(gameIDs, v.ForeignID)
	}
	if len(gameIDs) == 0 { //没有卡片，不下发组件
		return
	}
	//获取游戏详情
	gamesInfo := s.gameDao.BatchMultiGameInfo(c, gameIDs, mid, mobiApp)
	var items []*dynmdl.Item
	//拼接卡片信息
	for _, v := range games.List {
		if v == nil || v.MType != api.MixGame || v.ForeignID <= 0 {
			continue
		}
		if gv, ok := gamesInfo[v.ForeignID]; !ok || gv == nil {
			continue
		}
		itemTep := &dynmdl.Item{}
		itemTep.FromGameExt(v, gamesInfo[v.ForeignID])
		items = append(items, itemTep)
	}
	if len(items) == 0 { //没有卡片，不下发组件
		return
	}
	ck = &dynmdl.Item{}
	var lastItems []*dynmdl.Item
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &dynmdl.Item{}
		tmpName.FromTitleName(mou)
		lastItems = append(lastItems, tmpName)
	}
	lastItems = append(lastItems, items...)
	ck.FromGame(mou, lastItems)
	return
}

func (s *Service) rankIcon(c context.Context, id, num int64) map[int]string {
	rcmRly := make(map[int]string)
	mixIcon, e := s.ModuleMixExt(c, id, 0, num, api.MixRankIcon)
	if e != nil { //降级处理
		log.Error(" s.ModuleMixExt(%v) error(%v)", id, e)
		return rcmRly
	}
	if mixIcon == nil || len(mixIcon.List) == 0 {
		return rcmRly
	}
	i := 0
	for _, v := range mixIcon.List {
		if v == nil {
			continue
		}
		remark := v.RemarkUnmarshal()
		if remark.Image == "" {
			continue
		}
		rcmRly[i] = remark.Image
		i++
	}
	return rcmRly
}

func (s *Service) rankListFromModule(c context.Context, mou *api.NativeModule, mid int64) *dynmdl.Item {
	if mou.Fid == 0 {
		return nil
	}
	num := mou.Num
	maxPs := int64(10)
	if num > maxPs {
		num = maxPs
	}
	eg := errgroup.WithContext(c)
	var rcmRly map[int]string
	eg.Go(func(ctx context.Context) error {
		//获取icon
		rcmRly = s.rankIcon(ctx, mou.ID, num)
		return nil
	})
	var rly *actGRPC.RankResultResp
	eg.Go(func(ctx context.Context) (e error) {
		if rly, e = s.actDao.RankResult(ctx, mou.Fid, 1, num); e != nil {
			log.Error("s.actDao.RankResult(%d) error(%v)", mou.Fid, e)
		}
		return
	})
	err := eg.Wait()
	if err != nil {
		return nil
	}
	if rly == nil || len(rly.List) == 0 {
		return nil
	}
	var fids []int64
	for _, v := range rly.List {
		if v == nil || v.Account == nil || v.ObjectType != 1 {
			continue
		}
		fids = append(fids, v.Account.MID)
	}
	//获取关注关系
	var followRly map[int64]*relationapi.FollowingReply
	if mid > 0 {
		if followRly, err = s.relDao.RelationsGRPC(c, mid, fids); err != nil { //错误降级
			log.Error(" s.relDao.RelationsGRPC(%d,%v) error(%v)", mid, fids, err)
		}
	}
	var (
		items []*dynmdl.Item
		j     = 0
	)
	display := mou.IsAttrDisplayRecommend() == api.AttrModuleYes
	for _, reVal := range rly.List {
		if reVal == nil || reVal.Account == nil || reVal.ObjectType != 1 {
			continue
		}
		ext := &dynmdl.ClickExt{}
		if rel, ok := followRly[reVal.Account.MID]; ok {
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			if rel.Attribute == 2 || rel.Attribute == 6 {
				ext.IsFollow = true
			}
		}
		itemTep := &dynmdl.Item{}
		rcm := rcmRly[j]
		j++
		itemTep.FromRecommendRankExt(reVal, ext, display, rcm)
		items = append(items, itemTep)
	}
	if len(items) == 0 {
		return nil
	}
	ck := &dynmdl.Item{}
	var lastItems []*dynmdl.Item
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	lastItems = append(lastItems, items...)
	ck.FromRecommend(mou, lastItems)
	return ck
}

// formatRecommend .
func (s *Service) formatRecommend(c context.Context, mou *api.NativeModule, recom *api.Recommend, mid int64) (ck *dynmdl.Item) {
	var (
		fids      []int64
		followRly map[int64]*relationapi.FollowingReply
		cards     map[int64]*accgrpc.Card
		items     []*dynmdl.Item
		recomList = make([]*api.NativeMixtureExt, 0)
	)
	if mou.IsRcmdSource() {
		confSort := mou.ConfUnmarshal()
		switch confSort.SourceType {
		case api.SourceTypeRank: //排行榜
			return s.rankListFromModule(c, mou, mid)
		default:
			fids = s.rcmdSourceData(c, mou)
			for _, v := range fids {
				recomList = append(recomList, &api.NativeMixtureExt{ForeignID: v})
			}
		}
	} else {
		if recom == nil || len(recom.List) == 0 {
			return
		}
		recomList = recom.List
		for _, v := range recomList {
			if v.ForeignID > 0 {
				fids = append(fids, v.ForeignID)
			}
		}
	}
	if len(fids) == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	// 获取关注关系
	if mid > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if followRly, e = s.relDao.RelationsGRPC(ctx, mid, fids); e != nil {
				log.Error(" s.reldao.RelationsGRPC(%d,%v) error(%v)", mid, fids, e)
				e = nil
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		//获取用户信息
		if cards, e = s.accDao.Cards3GRPC(ctx, fids); e != nil {
			log.Error("s.accDao.Cards3GRPC(%v) error(%v)", fids, e)
			e = nil
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("formatRecommend eg.Wait() error(%v)", err)
		return
	}
	for _, reVal := range recomList {
		if reVal.ForeignID == 0 {
			continue
		}
		if _, aok := cards[reVal.ForeignID]; !aok {
			continue
		}
		ext := &dynmdl.ClickExt{}
		if rel, ok := followRly[reVal.ForeignID]; ok {
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			if rel.Attribute == 2 || rel.Attribute == 6 {
				ext.IsFollow = true
			}
		}
		itemTep := &dynmdl.Item{}
		itemTep.FromRecommendExt(reVal, cards[reVal.ForeignID], ext)
		items = append(items, itemTep)
	}
	if len(items) == 0 {
		return
	}
	ck = &dynmdl.Item{}
	var lastItems []*dynmdl.Item
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		lastItems = append(lastItems, tmpImage)
	}
	lastItems = append(lastItems, items...)
	ck.FromRecommend(mou, lastItems)
	return
}

// formatActCard .
func (s *Service) formatActCard(mou *api.NativeModule, acts *api.Act) (first *dynmdl.Item) {
	if acts == nil || len(acts.List) == 0 {
		return
	}
	first = &dynmdl.Item{}
	first.FromActModule(mou)
	first.Item = make([]*dynmdl.Item, 0)
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		first.Item = append(first.Item, tmpImage)
	}
	for _, v := range acts.List {
		tmpAct := &dynmdl.Item{}
		tmpAct.FromActs(v)
		first.Item = append(first.Item, tmpAct)
	}
	return
}

// formatTimelineResource .
func (s *Service) formatTimelineResource(c context.Context, mou *api.NativeModule) (list *dynmdl.Item) {
	//一次取50个
	confSort := mou.ConfUnmarshal()
	//默认浮层
	ps := mou.Num
	if confSort.MoreSort == api.MoreExpand { // 下拉展示
		ps = 50
	}
	rly, err := s.TimelineSource(c, mou.Fid, confSort.TimeSort, 0, int32(ps))
	if err != nil || rly == nil {
		log.Error("s.TimelineIDs error(%v)", err)
		return
	}
	// 有卡片信息才下发组件
	list = s.timelineJoin(rly, mou)
	return
}

// timelineJoin .
func (s *Service) timelineJoin(req *dynmdl.TimelineSourceReply, mou *api.NativeModule) *dynmdl.Item {
	// 有卡片信息才下发组件
	if req == nil || len(req.List) == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromTimelineModule(mou)
	if mou.Meta != "" {
		tmpImage := &dynmdl.Item{}
		tmpImage.FromTitleImage(mou)
		list.Item = append(list.Item, tmpImage)
	}
	if mou.Caption != "" {
		tmpName := &dynmdl.Item{}
		tmpName.FromTitleName(mou)
		list.Item = append(list.Item, tmpName)
	}
	confSort := mou.ConfUnmarshal()
	if confSort.MoreSort == api.MoreExpand { // 下拉展示
		var before, after, con []*dynmdl.Item
		if len(req.List) > int(mou.Num) {
			before = req.List[:mou.Num]
			after = req.List[mou.Num:]
			con = append(con, before...)
			moreTmp := &dynmdl.Item{}
			moreTmp.FromTimelineExpand(mou)
			moreTmp.Item = append(moreTmp.Item, after...)
			con = append(con, moreTmp)
		} else {
			con = req.List
		}
		list.Item = append(list.Item, con...)
	} else { //默认浮层
		list.Item = append(list.Item, req.List...)
		if req.HasMore > 0 {
			tmpMore := &dynmdl.Item{}
			tmpMore.FromTimelineMore(mou)
			list.Item = append(list.Item, tmpMore)
		}
	}
	return list
}

// formatCarouselImg .
func (s *Service) formatCarouselImg(mou *api.NativeModule, carousel *api.Carousel) *dynmdl.Item {
	if carousel == nil || len(carousel.List) == 0 {
		return nil
	}
	items := make([]*dynmdl.Item, 0)
	if mou.Meta != "" {
		item := &dynmdl.Item{}
		item.FromTitleImage(mou)
		items = append(items, item)
	}
	carouselItem := &dynmdl.Item{}
	carouselItem.FromCarouselImg(mou)
	for _, v := range carousel.List {
		if v == nil {
			continue
		}
		ext := &dynmdl.CarouselImage{}
		if err := json.Unmarshal([]byte(v.Reason), ext); err != nil {
			log.Error("Fail to unmarshal carouselImgExt, carouselImgExt=%s error=%+v", v.Reason, err)
			continue
		}
		item := &dynmdl.Item{}
		item.FromCarouselImgItem(ext)
		carouselItem.Item = append(carouselItem.Item, item)
	}
	//当图片不存在时，不下发组件
	if len(carouselItem.Item) == 0 {
		return nil
	}
	items = append(items, carouselItem)
	ck := &dynmdl.Item{}
	ck.FromCarouselImgModule(mou, items)
	return ck
}

func (s *Service) upListFromModule(c context.Context, mou *api.NativeModule, mid, pn int64, sortType string) ([]*dynmdl.CarouselImage, error) {
	if sortType == "" {
		sortType = api.SortTypeCtime
	}
	upList, err := s.actDao.UpList(c, mou.Fid, 1, pn, mid, sortType)
	if err != nil {
		return nil, err
	}
	images := make([]*dynmdl.CarouselImage, 0)
	if upList == nil {
		return images, nil
	}
	for _, v := range upList.List {
		if v == nil || v.Content == nil {
			continue
		}
		image := &dynmdl.CarouselImage{
			ImgUrl:      v.Content.Image,
			RedirectUrl: v.Content.Link,
			Length:      mou.Length,
			Width:       mou.Width,
		}
		images = append(images, image)
	}
	return images, nil
}

func (s *Service) carouselImgSourceData(c context.Context, mou *api.NativeModule, mid int64) ([]*dynmdl.CarouselImage, error) {
	confSort := mou.ConfUnmarshal()
	var carouselImgs []*dynmdl.CarouselImage
	switch confSort.SourceType {
	case api.SourceTypeActUp:
		var err error
		if carouselImgs, err = s.upListFromModule(c, mou, mid, 8, confSort.SortType); err != nil {
			return nil, err
		}
	default:
		return make([]*dynmdl.CarouselImage, 0), nil
	}
	return carouselImgs, nil
}

func (s *Service) formatCarouselSource(c context.Context, mou *api.NativeModule, mid int64) *dynmdl.Item {
	if mou.Fid <= 0 {
		return nil
	}
	list, err := s.carouselImgSourceData(c, mou, mid)
	if err != nil || len(list) == 0 {
		return nil
	}
	items := make([]*dynmdl.Item, 0)
	if mou.Meta != "" {
		item := &dynmdl.Item{}
		item.FromTitleImage(mou)
		items = append(items, item)
	}
	carouselItem := &dynmdl.Item{}
	carouselItem.FromCarouselImg(mou)
	for _, v := range list {
		if v == nil {
			continue
		}
		item := &dynmdl.Item{}
		item.FromCarouselImgItem(v)
		carouselItem.Item = append(carouselItem.Item, item)
	}
	//当图片不存在时，不下发组件
	if len(carouselItem.Item) == 0 {
		return nil
	}
	items = append(items, carouselItem)
	ck := &dynmdl.Item{}
	ck.FromCarouselImgModule(mou, items)
	return ck
}

// formatCarouselWord .
func (s *Service) formatCarouselWord(mou *api.NativeModule, carousel *api.Carousel) *dynmdl.Item {
	if carousel == nil || len(carousel.List) == 0 {
		return nil
	}
	items := make([]*dynmdl.Item, 0)
	carouselItem := &dynmdl.Item{}
	carouselItem.FromCarouselWord(mou)
	for _, v := range carousel.List {
		if v == nil {
			continue
		}
		ext := new(struct {
			Content string `json:"content"`
		})
		if err := json.Unmarshal([]byte(v.Reason), ext); err != nil {
			log.Error("Fail to unmarshal carouselWordExt, carouselWordExt=%s error=%+v", v.Reason, err)
			continue
		}
		item := &dynmdl.Item{}
		item.FromCarouselWordItem(ext.Content)
		carouselItem.Item = append(carouselItem.Item, item)
	}
	//没有数据不下发组件
	if len(carouselItem.Item) == 0 {
		return nil
	}
	items = append(items, carouselItem)
	ck := &dynmdl.Item{}
	ck.FromCarouselWordModule(mou, items)
	return ck
}

// formatIcon .
func (s *Service) formatIcon(mou *api.NativeModule, icon *api.Icon) *dynmdl.Item {
	if icon == nil || len(icon.List) == 0 {
		return nil
	}
	items := make([]*dynmdl.Item, 0)
	iconItem := &dynmdl.Item{}
	iconItem.FromIconExt(icon.List)
	if len(iconItem.Item) == 0 {
		return nil
	}
	items = append(items, iconItem)
	ck := &dynmdl.Item{}
	ck.FromIcon(mou, items)
	return ck
}

func (s *Service) formatEditorOrigin(c context.Context, mou *api.NativeModule, mid int64, buvid string) *dynmdl.Item {
	confSort := mou.ConfUnmarshal()
	switch confSort.RdbType {
	case api.RDBRank: //编辑推荐卡-排行榜
		return s.editRankOrigin(c, mou)
	case api.RDBChannel: //编辑推荐卡-垂类id
		return s.editChannel(c, mou, mid, buvid)
	case api.RDBMustsee: //编辑推荐卡-入站必刷
		return s.editMustsee(mou)
	default:
		list := &dynmdl.Item{}
		list.FromEditorOriginModule(mou)
		return list
	}
}

// editMustsee 无限feed流模式.
func (s *Service) editMustsee(mou *api.NativeModule) *dynmdl.Item {
	confSort := mou.ConfUnmarshal()
	ext := &dynmdl.UrlExt{Category: mou.Category, Fid: mou.Fid, Type: int32(confSort.RdbType), ConfModuleID: mou.ID}
	list := &dynmdl.Item{}
	list.FromNewEditorModule(mou, ext)
	return list
}

// editChannel 无限feed流模式.
func (s *Service) editChannel(c context.Context, mou *api.NativeModule, mid int64, buvid string) *dynmdl.Item {
	//判断对应的垂类id是否有数据
	chaRly, err := s.hmtChannelDao.ChannelFeed(c, mou.Fid, mid, buvid, 0, 1)
	if err != nil {
		log.Error("s.hmtChannelDao.ChannelFeed(%d) error(%v)", mou.Fid, err)
		return nil
	}
	if chaRly == nil || len(chaRly.List) == 0 {
		return nil
	}
	confSort := mou.ConfUnmarshal()
	ext := &dynmdl.UrlExt{Category: mou.Category, Fid: mou.Fid, Type: int32(confSort.RdbType), ConfModuleID: mou.ID}
	list := &dynmdl.Item{}
	list.FromNewEditorModule(mou, ext)
	return list
}

func (s *Service) editRankOrigin(c context.Context, mou *api.NativeModule) *dynmdl.Item {
	if mou.Fid <= 0 {
		return nil
	}
	ps := mou.Num
	maxPs := int64(10) //最多下发10张卡片，产品逻辑
	if ps > maxPs {
		ps = maxPs
	}
	eg := errgroup.WithContext(c)
	rcmRly := make(map[int]string)
	eg.Go(func(ctx context.Context) error {
		//获取icon
		mixIcon, e := s.ModuleMixExt(ctx, mou.ID, 0, ps, api.MixRankIcon)
		if e != nil { //降级处理
			log.Error(" s.ModuleMixExt(%d) error(%v)", mou.ID, e)
			return nil
		}
		if mixIcon == nil || len(mixIcon.List) == 0 {
			return nil
		}
		i := 0
		for _, v := range mixIcon.List {
			if v == nil {
				continue
			}
			mixFold := dynmdl.MixFolderUnmarshal(v.Reason)
			if mixFold != nil && mixFold.RcmdContent != nil {
				rcmRly[i] = mixFold.RcmdContent.MiddleIcon
			}
			i++
		}
		return nil
	})
	//获取稿件信息
	var rankRly *actGRPC.RankResultResp
	eg.Go(func(ctx context.Context) (e error) {
		if rankRly, e = s.actDao.RankResult(ctx, mou.Fid, 1, ps); e != nil {
			log.Error("s.actDao.RankResult(%d,%d) error(%v)", mou.Fid, ps, e)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		return nil
	}
	if rankRly == nil || len(rankRly.List) == 0 {
		return nil
	}
	arcDisplay := mou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	j := 0
	rly := make([]*dynmdl.Item, 0)
	for _, v := range rankRly.List {
		if v == nil || v.ObjectType != 2 || len(v.Archive) < 1 { //稿件榜
			continue
		}
		arcVal := v.Archive[0]
		if arcVal == nil {
			continue
		}
		tmp := &dynmdl.Item{}
		rcms := rcmRly[j]
		j++
		tmp.FromEditorRankArc(v, mou, arcVal, arcDisplay, &dynmdl.RcmdContent{MiddleIcon: rcms})
		rly = append(rly, tmp)
	}
	if len(rly) == 0 {
		return nil
	}
	list := &dynmdl.Item{}
	list.FromNewEditorModule(mou, nil)
	list.Item = append(list.Item, rly...)
	return list
}

// formatClick .
// nolint:gocognit
func (s *Service) formatClick(c context.Context, mou *api.NativeModule, acts *api.Click, mid int64) *dynmdl.Item {
	if mou.AvSort == api.NeedUnLock && mou.DySort == api.UnLockTime && mou.Stime > time.Now().Unix() { //解锁后展示 &&时间限制 && 未到达解锁时间，不下发组件
		return nil
	}
	var (
		clickItem    []*dynmdl.Item
		sids         []int64
		sidReply     map[int64]*actGRPC.ReserveFollowingReply
		taskPoint    = make(map[int64]*dynmdl.ParamPlat)
		lotteryIDs   []string
		progReqs     = make(map[int64][]int64) //sid=>[]groupID
		taskNums     map[int64]int64
		lotteryTimes map[string]int64
		buyIDs       []int64
		buyReply     map[int64]bool
		cartoonIDs   []int64
		cartoonRly   map[int64]*carmdl.ComicItem
		appointIDs   []int64
		// 评分
		scoreIDs     []int64
		scoreTargets map[int64]*scoregrpc.ScoreTarget
	)
	if acts != nil && len(acts.Areas) > 0 {
		for _, v := range acts.Areas {
			switch {
			case v.IsUpAppointment(): //up主预约
				appointIDs = append(appointIDs, v.ForeignID)
			case v.IsCartoon():
				cartoonIDs = append(cartoonIDs, v.ForeignID)
			case v.IsBuyCoupon():
				buyIDs = append(buyIDs, v.ForeignID)
			case v.IsProgress():
				sid, gid := extractProgressParamFromClick(v)
				if sid == 0 || gid == 0 {
					continue
				}
				progReqs[sid] = append(progReqs[sid], gid)
			case v.IsStaticProgress(): //静态-进度条
				areaTip := new(api.ClickTip)
				if err := json.Unmarshal([]byte(v.Tip), areaTip); err != nil {
					log.Error("Fail to unmarshal click ext=%+v error=%+v", v.Ext, err)
					continue
				}
				if areaTip.PSort == api.ProcessUserStatics {
					sid, gid := extractProgressParamFromClick(v)
					if sid == 0 || gid == 0 {
						continue
					}
					progReqs[sid] = append(progReqs[sid], gid)
				} else if areaTip.PSort == api.ProcessRegister { //老预约数据源
					sids = append(sids, v.ForeignID)
				} else if areaTip.PSort == api.ProcessTaskStatics {
					taskPoint[v.ID] = &dynmdl.ParamPlat{Activity: areaTip.Activity, Counter: areaTip.Counter, StatPc: areaTip.StatPc}
				} else if areaTip.PSort == api.ProcessLottery { //抽奖数据源
					lotteryIDs = append(lotteryIDs, areaTip.LotteryID)
				} else if areaTip.PSort == api.ProcessScore {
					scoreIDs = append(scoreIDs, v.ForeignID)
				}
			}
			if v.IsCustom() {
				setUnlockProgReq(v, progReqs)
			}
		}
		eg := errgroup.WithContext(c)
		if len(taskPoint) > 0 {
			taskNums = make(map[int64]int64)
			var taskMu sync.Mutex
			for k, v := range taskPoint {
				actStr := v.Activity
				counter := v.Counter
				statPc := v.StatPc
				clickID := k
				if statPc == "daily" {
					eg.Go(func(ctx context.Context) error {
						num, e := s.platDao.GetCounterRes(ctx, counter, actStr, mid)
						if e != nil {
							log.Error("s.platDao.GetCounterRes(%s,%s,%d) error(%v)", counter, actStr, mid, e)
							//降级错误不抛出
							return nil
						}
						taskMu.Lock()
						taskNums[clickID] = num
						taskMu.Unlock()
						return nil
					})
				} else {
					eg.Go(func(ctx context.Context) error {
						num, e := s.platDao.GetTotalRes(ctx, counter, actStr, mid)
						if e != nil {
							log.Error("s.platDao.GetTotalRes(%s,%s,%d) error(%v)", counter, actStr, mid, e)
							//降级错误不抛出
							return nil
						}
						taskMu.Lock()
						taskNums[clickID] = num
						taskMu.Unlock()
						return nil
					})
				}
			}
		}
		var appointRly map[int64]*actGRPC.UpActReserveRelationInfo
		if len(appointIDs) > 0 {
			eg.Go(func(ctx context.Context) error {
				appointRly, _ = s.actDao.UpActReserveRelationInfo(ctx, mid, appointIDs)
				return nil
			})
		}
		if len(buyIDs) > 0 && mid > 0 { //获取当前buy状态
			eg.Go(func(ctx context.Context) error {
				buyReply = s.shopDao.BatchMultiFavStat(ctx, buyIDs, mid)
				return nil
			})
		}
		if len(cartoonIDs) > 0 && mid > 0 { //获取当前漫画的关注状态
			eg.Go(func(ctx context.Context) (e error) {
				if cartoonRly, e = s.cartoonDao.GetComicInfos(ctx, cartoonIDs, mid); e != nil {
					log.Error("s.cartoonDao.GetComicInfos(%v,%d) error(%v)", cartoonIDs, mid, e)
					e = nil
				}
				return
			})
		}
		if len(lotteryIDs) > 0 && mid > 0 {
			lotteryTimes = make(map[string]int64)
			var lottMu sync.Mutex
			for _, v := range lotteryIDs {
				id := v
				eg.Go(func(ctx context.Context) error {
					reaRly, e := s.actDao.LotteryUnusedTimes(ctx, mid, id)
					if e != nil {
						log.Error("s.actDao.LotteryUnusedTimes(%d,%s) error(%v)", mid, id, e)
						//降级错误不抛出
						return nil
					}
					if reaRly != nil {
						lottMu.Lock()
						lotteryTimes[id] = reaRly.Times
						lottMu.Unlock()
					}
					return nil
				})
			}
		}
		if len(sids) > 0 && mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if sidReply, e = s.actDao.ReserveFollowings(ctx, mid, sids); e != nil {
					log.Error("s.actDao.ReserveFollowings(%d,%v) error(%v)", mid, sids, e)
					e = nil
				}
				return
			})
		}
		progRlys := make(map[int64]*actGRPC.ActivityProgressReply, len(progReqs))
		if len(progReqs) > 0 {
			lock := sync.Mutex{}
			for k, v := range progReqs {
				gids := v
				sid := k
				eg.Go(func(ctx context.Context) error {
					progress, err := s.actDao.ActivityProgress(ctx, sid, 2, mid, gids)
					if err != nil {
						return nil
					}
					lock.Lock()
					progRlys[sid] = progress
					lock.Unlock()
					return nil
				})
			}
		}
		if len(scoreIDs) > 0 {
			eg.Go(func(ctx context.Context) error {
				req := &scoregrpc.MultiGetTargetScoreReq{TntCode: 1, STargetType: 1, STargetIds: scoreIDs}
				if rly, err := s.scoreDao.MultiGetTargetScore(ctx, req); err == nil {
					scoreTargets = rly.GetTargets()
				}
				return nil
			})
		}
		_ = eg.Wait()
		for _, v := range acts.Areas {
			var ext *dynmdl.ClickExt
			switch {
			case v.IsCartoon():
				ext = &dynmdl.ClickExt{FID: v.ForeignID}
				if cval, ok := cartoonRly[v.ForeignID]; ok && cval.FavStatus == 1 {
					ext.IsFollow = true
				}
			case v.IsBuyCoupon():
				ext = &dynmdl.ClickExt{FID: v.ForeignID, Tip: v.Tip}
				if buVal, ok := buyReply[v.ForeignID]; ok && buVal {
					ext.IsFollow = true
				}
			case v.IsActReserve(), v.IsReserve(), v.IsCatchUp():
				ext = &dynmdl.ClickExt{FID: v.ForeignID, Tip: v.Tip}
			case v.IsFollow():
				ext = &dynmdl.ClickExt{FID: v.ForeignID, Tip: "关注"}
			case v.IsUpAppointment():
				if aVal, ok := appointRly[v.ForeignID]; !ok || aVal == nil { //没有返回值
					continue
				}
				images := &dynmdl.ImagesUnion{
					FinishedImage:   &dynmdl.Image{Image: v.FinishedImage},
					OptionalImage:   &dynmdl.Image{Image: v.OptionalImage},
					UnfinishedImage: &dynmdl.Image{Image: v.UnfinishedImage},
				}
				ext = &dynmdl.ClickExt{FID: v.ForeignID, Images: images}
				// 默认CurrentState=0是不可点击状态
				if appointRly[v.ForeignID].UpActVisible == actGRPC.UpActVisible_DefaultVisible && (appointRly[v.ForeignID].State == actGRPC.UpActReserveRelationState_UpReserveRelated || appointRly[v.ForeignID].State == actGRPC.UpActReserveRelationState_UpReserveRelatedOnline) {
					if appointRly[v.ForeignID].IsFollow == 1 {
						ext.CurrentState = 2
					} else {
						ext.CurrentState = 1
					}
				}
			case v.IsPendant():
				images := &dynmdl.ImagesUnion{
					FinishedImage:   &dynmdl.Image{Image: v.FinishedImage},
					OptionalImage:   &dynmdl.Image{Image: v.OptionalImage},
					UnfinishedImage: &dynmdl.Image{Image: v.UnfinishedImage},
				}
				ext = &dynmdl.ClickExt{FID: v.ForeignID, Images: images}
			case v.IsInterface():
				style, err := extractExt4ClickInterface(v.Ext)
				if err != nil || style == "" {
					continue
				}
				ext = &dynmdl.ClickExt{Style: style}
			case v.IsProgress():
				progRly, ok := progRlys[v.ForeignID]
				if !ok || len(progRly.Groups) == 0 {
					continue
				}
				areaTip := new(api.ClickTip)
				if err := json.Unmarshal([]byte(v.Tip), areaTip); err != nil {
					continue
				}
				group, ok := progRly.Groups[areaTip.GroupId]
				if !ok {
					continue
				}
				num, targetNum := extractProgNum(group, areaTip.NodeId)
				ext = &dynmdl.ClickExt{CurrentNum: num, TargetNum: targetNum}
			case v.IsStaticProgress(): //静态-进度条
				areaTip := new(api.ClickTip)
				if err := json.Unmarshal([]byte(v.Tip), areaTip); err != nil {
					continue
				}
				ext = &dynmdl.ClickExt{}
				if areaTip.PSort == api.ProcessUserStatics {
					progRly, ok := progRlys[v.ForeignID]
					if !ok || len(progRly.Groups) == 0 {
						continue
					}
					group, ok := progRly.Groups[areaTip.GroupId]
					if !ok {
						continue
					}
					ext.CurrentNum, ext.TargetNum = extractProgNum(group, areaTip.NodeId)
				} else if areaTip.PSort == api.ProcessRegister { //老预约数据源
					if sval, ok := sidReply[v.ForeignID]; ok && sval != nil {
						dimension, _ := extractDimension(v)
						if dimension == actGRPC.GetReserveProgressDimension_Rule { //整体活动维度
							ext.CurrentNum = calculateProgress(sval.Total, areaTip.InterveNum)
						} else if sval.IsFollow { //用户维度&& 用户预约了
							ext.CurrentNum = 1
						}
					}
				} else if areaTip.PSort == api.ProcessTaskStatics {
					if tv, ok := taskNums[v.ID]; ok {
						ext.CurrentNum = tv
					}
				} else if areaTip.PSort == api.ProcessLottery { //抽奖数据源
					if lv, ok := lotteryTimes[areaTip.LotteryID]; ok {
						ext.CurrentNum = lv
					}
				} else if areaTip.PSort == api.ProcessScore {
					st, ok := scoreTargets[v.ForeignID]
					if !ok {
						continue
					}
					ext.DisplayNum = finalScore(st)
				}
			}
			if v.IsCustom() && !reachUnlockCondition(v, progRlys) {
				continue
			}
			dTmp := &dynmdl.Item{}
			dTmp.FromArea(v, ext, mou)
			clickItem = append(clickItem, dTmp)
		}
	}
	res := &dynmdl.Item{}
	res.FromClick(mou, clickItem)
	return res
}

// NatModule .
func (s *Service) NatModule(c context.Context, arg *dynmdl.ParamNatModule) (res *dynmdl.NatModuleReply, err error) {
	var (
		page *dynmdl.ModuleReply
	)
	if page, err = s.UkeyToModule(c, arg.PageID, arg.Ukey); err != nil {
		log.Error("s.UkeyToModule(%d,%s) error(%v)", arg.PageID, arg.Ukey, err)
		return
	}
	if page.Module == nil {
		return
	}
	res = &dynmdl.NatModuleReply{}
	switch {
	case page.Module.IsVideoAvid(), page.Module.IsVideoDyn(), page.Module.IsVideoAct(), page.Module.IsResourceID():
		res.MoreUrl = "bilibili://following/activity_detail/" + strconv.FormatInt(page.Module.ID, 10)
		res.MoreParam = &dynmdl.MoreParam{Offset: page.Module.Num, DyOffset: "", PageID: arg.PageID}
	case page.Module.IsLive():
		res.Stime = page.Module.Stime
		res.Etime = page.Module.Etime
	}
	return
}

// TimelineSource .
func (s *Service) TimelineSource(c context.Context, fid, timeSort int64, offset, ps int32) (*dynmdl.TimelineSourceReply, error) {
	// 获取配置信息
	rly, err := s.popularDao.TimeLine(c, fid, offset, ps)
	if err != nil {
		log.Error("s.popularDao.TimeLine(%d) error(%v)", fid, err)
		return nil, err
	}
	res := &dynmdl.TimelineSourceReply{Offset: rly.Offset}
	if rly.HasMore {
		res.HasMore = 1
	}
	for _, v := range rly.Events {
		if v == nil {
			continue
		}
		title := dynmdl.FromTimelineFormatHead(xtime.Time(v.Stime), timeSort)
		tmp := &dynmdl.Item{}
		tmp.FromTimeline(v, title)
		res.List = append(res.List, tmp)
	}
	return res, nil
}

// SeasonSource .
func (s *Service) SeasonSource(c context.Context, arg *dynmdl.ParamSeasonSource, mid int64) (*dynmdl.IDsReply, error) {
	tempMou := &api.NativeModule{Fid: arg.FID, Attribute: arg.Attribute, CardStyle: arg.CardStyle}
	rly, err := s.ogvSeasonResource(c, tempMou, arg.Ps, arg.Offset, mid)
	if err != nil {
		log.Error("s.pgcDao.SeasonByPlayId(%v) error(%v)", arg, err)
		return &dynmdl.IDsReply{}, nil
	}
	return rly, nil
}

// SeasonIDs .
func (s *Service) SeasonIDs(c context.Context, arg *dynmdl.ParamSeasonIDs, mid int64) (*dynmdl.IDsReply, error) {
	var listReq []*dynmdl.ResourceAid
	if err := json.Unmarshal([]byte(arg.IDs), &listReq); err != nil {
		return nil, xecode.RequestErr
	}
	if len(listReq) > dynmdl.MaxIDsLen {
		return nil, xecode.RequestErr
	}
	var ssids []int32
	for _, v := range listReq {
		if v.ID == 0 || v.Type != api.MixOgvSsid {
			continue
		}
		ssids = append(ssids, int32(v.ID))
	}
	if len(ssids) == 0 {
		return &dynmdl.IDsReply{}, nil
	}
	ep, err := s.pgcDao.SeasonBySeasonId(c, ssids, mid)
	if err != nil {
		log.Error("s.pgcDao.SeasonBySeasonId(%v) error(%v)", ssids, err)
		return &dynmdl.IDsReply{}, nil
	}
	tempMou := &api.NativeModule{Attribute: arg.Attribute, CardStyle: arg.CardStyle}
	rly := &dynmdl.IDsReply{}
	for _, v := range listReq {
		if v.ID == 0 || v.Type != api.MixOgvSsid {
			continue
		}
		// 获取ssid对应的ogv详情
		if _, ok := ep[int32(v.ID)]; !ok {
			continue
		}
		tmp := &dynmdl.Item{}
		tmp.FromOgvSeason(tempMou, ep[int32(v.ID)], "")
		// 拼接卡片信息
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// ResourceAid .
func (s *Service) ResourceAid(c context.Context, arg *dynmdl.ParamResourceAid) (*dynmdl.IDsReply, error) {
	var listReq []*dynmdl.ResourceAid
	if err := json.Unmarshal([]byte(arg.IDs), &listReq); err != nil {
		return nil, xecode.RequestErr
	}
	if len(listReq) > dynmdl.MaxIDsLen {
		return nil, xecode.RequestErr
	}
	return s.resourceJoin(c, listReq, arg.Attribute)
}

func (s *Service) getResource(c context.Context, aids, cvids, epids, fids, roomids []int64, attr int64) (map[int64]*arccli.Arc, map[int64]*artmdl.Meta, map[int64]*lmdl.EpPlayer, map[int64]*favmdl.Folder, map[int64]*playgrpc.RoomList) {
	var (
		arcRly  map[int64]*arccli.Arc
		artRly  map[int64]*artmdl.Meta
		epRly   map[int64]*lmdl.EpPlayer
		foldRly = make(map[int64]*favmdl.Folder)
		roomRly map[int64]*playgrpc.RoomList
	)
	tempMou := &api.NativeModule{Attribute: attr}
	isLive := tempMou.IsAttrDisplayNodeNum()
	eg := errgroup.WithContext(c)
	if len(roomids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if roomRly, e = s.liveDao.GetListByRoomId(ctx, roomids, isLive); e != nil {
				log.Error("s.liveDao.GetListByRoomId(%v) error(%v)", aids, e)
				e = nil
				return
			}
			return
		})
	}
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcRes, e := s.arcClient.Arcs(ctx, &arccli.ArcsRequest{Aids: aids})
			if e != nil {
				log.Error("s.arcClient.Arcs aids(%v) error(%v)", aids, e)
				return nil
			}
			if arcRes != nil {
				arcRly = arcRes.Arcs
			}
			return nil
		})
	}
	if len(cvids) > 0 {
		eg.Go(func(ctx context.Context) error {
			artRes, e := s.artClient.ArticleMetas(ctx, &artapi.ArticleMetasReq{Ids: cvids, From: 2})
			if e != nil {
				log.Error("s.dao.ArticleMeta cvids(%v) error(%v)", cvids, e)
				return nil
			}
			if artRes != nil {
				artRly = artRes.Res
			}
			return nil
		})
	}
	if len(epids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if epRly, e = s.dao.EpPlayer(ctx, epids); e != nil {
				log.Error("s.dao.EpPlayer epids(%v) error(%v)", epids, e)
				e = nil
			}
			return
		})
	}
	if len(fids) > 0 {
		var mu sync.Mutex
		eg.Go(func(ctx context.Context) error {
			folders, err := s.favDao.Folders(ctx, fids, int32(favmdl.TypeVideo))
			if err != nil {
				log.Error("s.favDao.Folders(%v) error(%v)", fids, err)
				return nil
			}
			for _, folder := range folders.GetRes() {
				if folder == nil || folder.Attr&1 == 1 {
					continue
				}
				mu.Lock()
				foldRly[folder.Mlid] = folder
				mu.Unlock()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("s.getResource  eg.Wait() error(%v)", err)
		//降级处理
	}
	return arcRly, artRly, epRly, foldRly, roomRly
}

// nolint:gocognit
func (s *Service) resourceJoin(c context.Context, ids []*dynmdl.ResourceAid, attr int64) (*dynmdl.IDsReply, error) {
	if len(ids) == 0 {
		return &dynmdl.IDsReply{}, nil
	}
	var aids, cvids, epids, fids, roomids []int64
	for _, v := range ids {
		switch v.Type {
		case api.MixAvidType, api.MixFolder:
			if v.Bvid != "" {
				if avid, err := bvid.BvToAv(v.Bvid); err == nil && avid > 0 {
					aids = append(aids, avid)
					v.ID = avid
				}
			} else if v.ID > 0 {
				if bvidStr, err := bvid.AvToBv(v.ID); err == nil && bvidStr != "" {
					v.Bvid = bvidStr
				}
				aids = append(aids, v.ID)
			}
			if v.Fid > 0 && v.Type == api.MixFolder {
				fids = append(fids, v.Fid)
			}
		case api.MixCvidType:
			if v.ID > 0 {
				cvids = append(cvids, v.ID)
			}
		case api.MixEpidType:
			if v.ID > 0 {
				epids = append(epids, v.ID)
			}
		case api.MixLive:
			if v.ID > 0 {
				roomids = append(roomids, v.ID)
			}
		default:
			continue
		}
	}
	arcRly, artRly, epRly, foldRly, roomRly := s.getResource(c, aids, cvids, epids, fids, roomids, attr)
	tempMou := &api.NativeModule{Attribute: attr}
	artDisplay := tempMou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	arcDisplay := tempMou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	pgcDisplay := tempMou.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	rly := &dynmdl.IDsReply{}
	for _, v := range ids {
		if v.ID == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		switch v.Type {
		case api.MixAvidType, api.MixFolder:
			if va, ok := arcRly[v.ID]; !ok || va == nil {
				continue
			}
			if v.Type == api.MixAvidType {
				tmp.FromResourceArc(arcRly[v.ID], arcDisplay, v.Bvid, nil)
			} else {
				if fold, ok := foldRly[v.Fid]; !ok || fold == nil {
					continue
				}
				tmp.FromResourceArc(arcRly[v.ID], arcDisplay, v.Bvid, foldRly[v.Fid])
			}
		case api.MixCvidType:
			if artRly == nil {
				continue
			}
			if va, ok := artRly[v.ID]; !ok || va == nil {
				continue
			}
			tmp.FromResourceArt(artRly[v.ID], artDisplay)
		case api.MixEpidType:
			if va, ok := epRly[v.ID]; !ok || va == nil {
				continue
			}
			tmp.FromResourceEp(epRly[v.ID], pgcDisplay)
		case api.MixLive:
			if va, ok := roomRly[v.ID]; !ok || va == nil {
				continue
			}
			tmp.FromResourceLive(roomRly[v.ID])
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// NewVideoAid 新视频卡id模式 .
// nolint:gocognit
func (s *Service) NewVideoAid(c context.Context, arg *dynmdl.ParamAid) (rly *dynmdl.IDsReply, err error) {
	var listReq []*dynmdl.ResourceAid
	if err = json.Unmarshal([]byte(arg.IDs), &listReq); err != nil {
		err = xecode.RequestErr
		return
	}
	if len(listReq) > dynmdl.MaxIDsLen {
		err = xecode.RequestErr
		return
	}
	var aids, epids []int64
	for _, v := range listReq {
		switch v.Type {
		case api.MixAvidType:
			if v.Bvid != "" {
				if avid, err := bvid.BvToAv(v.Bvid); err == nil && avid > 0 {
					aids = append(aids, avid)
					v.ID = avid
				}
			} else if v.ID > 0 {
				if bvidStr, err := bvid.AvToBv(v.ID); err == nil && bvidStr != "" {
					v.Bvid = bvidStr
				}
				aids = append(aids, v.ID)
			}
		case api.MixEpidType:
			if v.ID > 0 {
				epids = append(epids, v.ID)
			}
		default:
			continue
		}
	}
	var (
		arcRly map[int64]*arccli.Arc
		epRly  map[int64]*lmdl.EpPlayer
	)
	eg := errgroup.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcRes, e := s.arcClient.Arcs(ctx, &arccli.ArcsRequest{Aids: aids})
			if e != nil {
				log.Error("s.arcClient.Arcs aids(%v) error(%v)", aids, e)
				return nil
			}
			if arcRes != nil {
				arcRly = arcRes.Arcs
			}
			return nil
		})
	}
	if len(epids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if epRly, e = s.dao.EpPlayer(ctx, epids); e != nil {
				log.Error("s.dao.EpPlayer epids(%v) error(%v)", epids, e)
				e = nil
			}
			return
		})
	}
	_ = eg.Wait()
	rly = &dynmdl.IDsReply{}
	for _, v := range listReq {
		if v.ID == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		switch v.Type {
		case api.MixAvidType:
			if va, ok := arcRly[v.ID]; !ok || va == nil {
				continue
			}
			tmp.FromUgcVideo(arcRly[v.ID], v.Bvid)
		case api.MixEpidType:
			if va, ok := epRly[v.ID]; !ok || va == nil {
				continue
			}
			tmp.FromPgcVideo(epRly[v.ID])
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
	}
	return
}

// NatPages .
func (s *Service) NatPages(c context.Context, ids []int64) (res *dynmdl.NatReply, err error) {
	var (
		acts map[int64]*api.NativePage
	)
	if acts, err = s.natDao.NativePages(c, ids); err != nil {
		return
	}
	res = &dynmdl.NatReply{}
	if len(acts) == 0 {
		return
	}
	for _, v := range ids {
		if aVal, ok := acts[v]; !ok || aVal == nil || !aVal.IsOnline() {
			continue
		}
		tmpAct := &dynmdl.Item{}
		tmpAct.FromActs(acts[v])
		res.Items = append(res.Items, tmpAct)
	}
	return
}

// ActDynamic .
func (s *Service) ActDynamic(c context.Context, arg *dynmdl.ParamActDynamic, mid int64) (res *dynmdl.DynReply, err error) {
	var (
		reply *dynmdl.DyReply
		page  *dynmdl.ModuleReply
	)
	if reply, err = s.dynamicDao.FetchDynamics(c, arg.TopicID, mid, arg.Ps, arg.Types, "", arg.Sortby); err != nil || reply == nil {
		log.Error("s.dynamicDao.FetchDynamics(%d) error(%v)", arg.TopicID, err)
		return
	}
	// 拿module_id
	res = &dynmdl.DynReply{}
	if len(reply.Cards) > 0 {
		res.Display = true
	}
	tempMou := &api.NativeModule{Attribute: arg.Attribute}
	if tempMou.IsAttrLast() != api.AttrModuleYes && tempMou.IsAttrHideMore() != api.AttrModuleYes {
		if len(reply.Cards) > 0 && reply.HasMore > 0 {
			if arg.PageID <= 0 || arg.Ukey == "" {
				return
			}
			if page, err = s.UkeyToModule(c, arg.PageID, arg.Ukey); err != nil {
				log.Warn("s.UkeyToModule(%d,%s) error(%v)", arg.PageID, arg.Ukey, err)
				err = nil
				return
			}
			tmpMore := &dynmdl.Item{}
			tmpMore.FromDynamicMore(page.ForeignID, arg.PageID, page.Module, arg.Types, page.Title, reply.Offset)
			res.Items = append(res.Items, tmpMore)
		}
	}
	return
}

// NewVideoDyn .
func (s *Service) NewVideoDyn(c context.Context, arg *dynmdl.ParamVideoDyn, mid int64) (rly *dynmdl.IDsReply, err error) {
	var briRly *dynmdl.BriefReply
	types := fmt.Sprintf("%d", dynmdl.VideoType)
	if briRly, err = s.dynamicDao.BriefDynamics(c, arg.TopicID, arg.Ps, mid, types, "", arg.Sortby); err != nil || briRly == nil {
		log.Error(" s.dynamicDao.BriefDynamics(%d) error(%v)", arg.TopicID, err)
		return
	}
	var aids []int64
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		switch v.Type {
		case dynmdl.VideoType:
			aids = append(aids, v.Rid)
		}
	}
	rly = &dynmdl.IDsReply{HasMore: int32(briRly.HasMore), DyOffset: briRly.Offset}
	if len(aids) == 0 {
		return
	}
	arcRes, e := s.arcClient.Arcs(c, &arccli.ArcsRequest{Aids: aids})
	if e != nil {
		log.Error("s.arcClient.Arcs aids(%v) error(%v)", aids, e)
		return
	}
	if arcRes == nil {
		return
	}
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		switch v.Type {
		case dynmdl.VideoType:
			if va, ok := arcRes.Arcs[v.Rid]; !ok || va == nil {
				continue
			}
			bvidStr, _ := bvid.AvToBv(v.Rid)
			tmp.FromUgcVideo(arcRes.Arcs[v.Rid], bvidStr)
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
	}
	return
}

// ResourceDyn .
// nolint:gocognit
func (s *Service) ResourceDyn(c context.Context, arg *dynmdl.ParamResourceDyn, mid int64) (rly *dynmdl.IDsReply, err error) {
	var briRly *dynmdl.BriefReply
	if briRly, err = s.dynamicDao.BriefDynamics(c, arg.TopicID, arg.Ps, mid, arg.Types, "", arg.Sortby); err != nil || briRly == nil {
		log.Error(" s.dynamicDao.BriefDynamics(%d) error(%v)", arg.TopicID, err)
		return
	}
	var cvids, aids []int64
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		switch v.Type {
		case dynmdl.ARTICLETYPE:
			cvids = append(cvids, v.Rid)
		case dynmdl.VideoType:
			aids = append(aids, v.Rid)
		}
	}
	var (
		arcRly map[int64]*arccli.Arc
		artRly map[int64]*artmdl.Meta
	)
	eg := errgroup.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcRes, e := s.arcClient.Arcs(ctx, &arccli.ArcsRequest{Aids: aids})
			if e != nil {
				log.Error("s.arcClient.Arcs aids(%v) error(%v)", aids, e)
				return nil
			}
			if arcRes != nil {
				arcRly = arcRes.Arcs
			}
			return nil
		})
	}
	if len(cvids) > 0 {
		eg.Go(func(ctx context.Context) error {
			artRes, e := s.artClient.ArticleMetas(ctx, &artapi.ArticleMetasReq{Ids: cvids, From: 2})
			if e != nil {
				log.Error("s.dao.ArticleMeta cvids(%v) error(%v)", cvids, e)
				return nil
			}
			if artRes != nil {
				artRly = artRes.Res
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "ResourceDyn eg.Wait() failed, req=%+v error=%+v", arg, err)
	}
	rly = &dynmdl.IDsReply{HasMore: int32(briRly.HasMore), DyOffset: briRly.Offset}
	tempMou := &api.NativeModule{Attribute: arg.Attribute}
	artDisplay := tempMou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	arcDisplay := tempMou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	for _, v := range briRly.Dynamics {
		if v.Rid == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		switch v.Type {
		case dynmdl.VideoType:
			if va, ok := arcRly[v.Rid]; !ok || va == nil {
				continue
			}
			bvidStr, _ := bvid.AvToBv(v.Rid)
			tmp.FromResourceArc(arcRly[v.Rid], arcDisplay, bvidStr, nil)
		case dynmdl.ARTICLETYPE:
			if artRly == nil {
				continue
			}
			if va, ok := artRly[v.Rid]; !ok || va == nil {
				continue
			}
			tmp.FromResourceArt(artRly[v.Rid], artDisplay)
		}
		rly.List = append(rly.List, tmp)
	}
	if tempMou.IsAttrLast() != api.AttrModuleYes && tempMou.IsAttrHideMore() != api.AttrModuleYes {
		if len(rly.List) > 0 && rly.HasMore > 0 {
			if arg.PageID <= 0 || arg.Ukey == "" {
				return
			}
			page, e := s.UkeyToModule(c, arg.PageID, arg.Ukey)
			if e != nil {
				log.Warn("s.UkeyToModule(%d,%s) error(%v)", arg.PageID, arg.Ukey, e)
				return
			}
			if page.Module == nil {
				return
			}
			tmpMore := &dynmdl.Item{}
			tmpMore.FromVideoMore(page.Module, 0, arg.PageID, rly.DyOffset)
			rly.List = append(rly.List, tmpMore)
		}
	}
	return
}

// UkeyToModule .
func (s *Service) UkeyToModule(c context.Context, pid int64, ukey string) (res *dynmdl.ModuleReply, err error) {
	var (
		acts     map[int64]*api.NativePage
		module   map[int64]*api.NativeModule
		moduleID int64
	)
	if moduleID, err = s.natDao.NativeUkey(c, pid, ukey); err != nil {
		log.Error("s.natDao.NativeUkey(%d,%s) error(%v)", pid, ukey, err)
		return
	}
	if moduleID == 0 {
		err = ecode.NativePageOffline
		return
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (e error) {
		if acts, e = s.natDao.NativePages(ctx, []int64{pid}); e != nil {
			log.Error("s.natDao.NativePages(%d) error(%v)", pid, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if module, e = s.natDao.OnlineNativeModules(ctx, []int64{moduleID}); e != nil {
			log.Error("s.natDao.OnlineNativeModules(%d) error(%v)", moduleID, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if _, k := module[moduleID]; !k {
		err = ecode.NativePageOffline
		return
	}
	if _, ok := acts[pid]; !ok || module[moduleID].NativeID != pid {
		err = ecode.NativePageOffline
		return
	}
	res = &dynmdl.ModuleReply{NativePage: acts[pid], Module: module[moduleID]}
	return
}

// LiveDyn .
func (s *Service) LiveDyn(c context.Context, arg *dynmdl.ParamLiveDyn, mid int64) (*dynmdl.LiveDynRly, error) {
	rly, err := s.liveDao.GetCardInfo(c, arg.RoomIDs, mid, arg.IsHttps)
	if err != nil {
		log.Error("s.liveDao.GetCardInfo(%d,%d) error(%v)", mid, arg.RoomIDs, err)
		return nil, err
	}
	res := &dynmdl.LiveDynRly{Cards: make(map[int64]*livegrpc.LiveCardInfo)}
	for _, v := range rly {
		if v != nil {
			res.Cards[v.RoomId] = v
		}
	}
	return res, nil
}

func (s *Service) ResourceRole(c context.Context, arg *dynmdl.ParamResourceRole) (*dynmdl.ResourceRoleReply, error) {
	epIDs, err := s.GetCharacterEps(c, arg.RoleID, arg.SeasonID)
	if err != nil {
		log.Error("Fail to get characterEps, charID=%d seasonID=%d", arg.RoleID, arg.SeasonID)
		return nil, err
	}
	epIDs, _ = pagingList(epIDs, 0, arg.Ps)
	if len(epIDs) == 0 {
		return &dynmdl.ResourceRoleReply{}, nil
	}
	epList, err := s.dao.EpPlayer(c, epIDs)
	if err != nil {
		log.Error("Fail to get epPlayer, epIDs=%+v error=%+v", epIDs, err)
		return nil, err
	}
	module := &api.NativeModule{Attribute: arg.Attribute}
	pgcDisplay := module.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	list := make([]*dynmdl.Item, 0, len(epList))
	for _, v := range epList {
		if v == nil {
			continue
		}
		item := &dynmdl.Item{}
		item.FromResourceEp(v, pgcDisplay)
		list = append(list, item)
	}
	return &dynmdl.ResourceRoleReply{List: list}, nil
}

func (s *Service) GetCharacterEps(c context.Context, charID, seasonID int32) ([]int64, error) {
	req := &media.CharacterIdsOidsReq{
		CharacterIdOpusIds: map[int32]*media.OpusIdsReq{charID: {Ids: []int32{seasonID}}},
		Otype:              100,
	}
	reply, err := s.characterClient.RelInfos(c, req)
	if err != nil {
		log.Error("Fail to get RelInfos, req=%+v error=%+v", req, err)
		return nil, err
	}
	relInfo, ok := reply.GetInfos()[charID]
	if !ok || relInfo.GetCharacterEp() == nil {
		return []int64{}, nil
	}
	epList, ok := relInfo.GetCharacterEp()[seasonID]
	if !ok || epList.GetCharacterEp() == nil {
		return []int64{}, nil
	}
	epIDs := make([]int64, 0, len(epList.GetCharacterEp()))
	for _, ep := range epList.GetCharacterEp() {
		if ep == nil {
			continue
		}
		epIDs = append(epIDs, int64(ep.GetEpId()))
	}
	return epIDs, nil
}

// offset: 从0开始
func pagingList(list []int64, offset, ps int) ([]int64, bool) {
	if offset > len(list) {
		offset = len(list)
	}
	end := offset + ps
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], end != len(list)
}

// EditorOrigin -编辑推荐卡-垂类id.
func (s *Service) EditorOrigin(c context.Context, arg *dynmdl.ParamEditorOrigin, mid int64, buvid string) (*dynmdl.IDsReply, error) {
	if arg.ConfModuleID == 0 {
		return &dynmdl.IDsReply{}, nil
	}
	//根据moduleID获取配置信息
	confRly, err := s.natDao.NativeModules(c, []int64{arg.ConfModuleID})
	if err != nil {
		log.Error("s.natDao.NativeModules(%d) error(%v)", arg.ConfModuleID, err)
		return &dynmdl.IDsReply{}, nil
	}
	if cv, ok := confRly[arg.ConfModuleID]; !ok || cv == nil {
		return &dynmdl.IDsReply{}, nil
	}
	modu := confRly[arg.ConfModuleID]
	confSort := modu.ConfUnmarshal()
	switch confSort.RdbType {
	case api.RDBChannel:
		return s.editorChannel(c, arg, modu, mid, buvid)
	case api.RDBMustsee:
		return s.editorMustsee(c, arg, modu)
	}
	return &dynmdl.IDsReply{}, nil
}

func (s *Service) editorChannel(c context.Context, pas *dynmdl.ParamEditorOrigin, moud *api.NativeModule, mid int64, buvid string) (*dynmdl.IDsReply, error) {
	chaRly, err := s.hmtChannelDao.ChannelFeed(c, moud.Fid, mid, buvid, pas.Offset, pas.Ps)
	if err != nil {
		log.Error("s.hmtChannelDao.ChannelFeed(%d,%d) error(%v)", pas.Offset, pas.Ps, err)
		return nil, err
	}
	if chaRly == nil || len(chaRly.List) == 0 {
		return &dynmdl.IDsReply{}, nil
	}
	rly := &dynmdl.IDsReply{Offset: int64(chaRly.GetOffset())}
	if chaRly.GetHasMore() {
		rly.HasMore = 1
	}
	var (
		aids  []int64
		epids []int64
	)
	//拼接aids和epids
	for _, v := range chaRly.List {
		if v == nil || v.Id <= 0 {
			continue
		}
		switch v.Type {
		case hmtchagrpc.ResourceType_UGC_RESOURCE:
			aids = append(aids, v.Id)
		case hmtchagrpc.ResourceType_OGV_RESOURCE:
			epids = append(epids, v.Id)
		default:
		}
	}
	arcRly, _, epRly, _, _ := s.getResource(c, aids, []int64{}, epids, []int64{}, []int64{}, moud.Attribute)
	arcDisplay := moud.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	pgcDisplay := moud.IsAttrDisplayPgcIcon() == api.AttrModuleYes
	for _, v := range chaRly.List {
		if v == nil || v.Id <= 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		switch v.Type {
		case hmtchagrpc.ResourceType_UGC_RESOURCE:
			if aVal, ok := arcRly[v.Id]; !ok || aVal == nil || !aVal.IsNormal() {
				continue
			}
			tmp.FromNewEditorArc(moud, arcRly[v.Id], arcDisplay, nil, nil)
		case hmtchagrpc.ResourceType_OGV_RESOURCE:
			if va, ok := epRly[v.Id]; !ok || va == nil {
				continue
			}
			// ugc确认每个position固定展示
			tmp.FromEditorEp(epRly[v.Id], pgcDisplay, moud, nil, `{"position2": "duration","position4": "view","position5": "follow"}`)
		default:
			continue
		}
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

func (s *Service) editorMustsee(c context.Context, pas *dynmdl.ParamEditorOrigin, moud *api.NativeModule) (*dynmdl.IDsReply, error) {
	confSort := moud.ConfUnmarshal()
	mustseeRly, err := s.popularDao.PageArcs(c, int64(pas.Offset), int64(pas.Ps), confSort.MseeType)
	if err != nil {
		log.Error("s.populardao.PageArcs(%d,%d) error(%v)", pas.Offset, pas.Ps, err)
		return nil, err
	}
	rly := &dynmdl.IDsReply{}
	if mustseeRly == nil || len(mustseeRly.List) == 0 {
		return rly, nil
	}
	if mustseeRly.Page != nil {
		rly.Offset = mustseeRly.Page.Offset
		rly.HasMore = int32(mustseeRly.Page.HasMore)
	}
	var (
		aids []int64
		fid  int64
	)
	for _, v := range mustseeRly.List {
		if v == nil || v.Aid <= 0 {
			continue
		}
		aids = append(aids, v.Aid)
	}
	fid = mustseeRly.MediaId
	arcRly, _, _, foldRly, _ := s.getResource(c, aids, []int64{}, []int64{}, []int64{fid}, []int64{}, moud.Attribute)
	fold := foldRly[fid]
	arcDisplay := moud.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	for _, v := range mustseeRly.List {
		if v == nil || v.Aid <= 0 {
			continue
		}
		if aVal, ok := arcRly[v.Aid]; !ok || aVal == nil || !aVal.IsNormal() {
			continue
		}
		tmp := &dynmdl.Item{}
		tmp.FromNewEditorArc(moud, arcRly[v.Aid], arcDisplay, &dynmdl.RcmdContent{TopContent: v.Recommend}, fold)
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// ResourceOrigin -资源小卡外接数据源类型.
func (s *Service) ResourceOrigin(c context.Context, arg *dynmdl.ParamResourceOrigin, mid int64) (*dynmdl.IDsReply, error) {
	switch arg.RDBType {
	case api.RDOBusinessCommodity: //企业号-商品卡
		return s.businessProduct(c, arg)
	case api.RDOBusinessIDs: //企业号-id list
		return s.businessSource(c, arg)
	case api.RDOOgvWid: //OGV运营后台-WID
		return s.ogvWid(c, arg, mid)
	case api.RDBLive: //直播间
		return s.resourceLive(c, arg)
	}
	return &dynmdl.IDsReply{}, nil
}

// businessSource .
func (s *Service) businessSource(c context.Context, arg *dynmdl.ParamResourceOrigin) (*dynmdl.IDsReply, error) {
	rly, err := s.busDao.SourceDetail(c, arg.SourceID, arg.Offset, arg.Ps)
	if err != nil {
		log.Error("s.dynamicDao.ProductDetail(%s,%d,%d) error(%v)", arg.SourceID, arg.Offset, arg.Ps, err)
		return &dynmdl.IDsReply{}, nil
	}
	var joinParams []*dynmdl.ResourceAid
	for _, v := range rly.ItemList {
		if v == nil {
			continue
		}
		if v.ItemID > 0 && (v.Type == api.MixAvidType || v.Type == api.MixCvidType || v.Type == api.MixEpidType || v.Type == api.MixFolder) {
			joinParams = append(joinParams, &dynmdl.ResourceAid{ID: v.ItemID, Type: v.Type, Fid: v.FID})
		}
	}
	pRes := &dynmdl.IDsReply{}
	pRes.Offset = rly.Offset
	pRes.HasMore = rly.HasMore
	if len(joinParams) == 0 {
		return pRes, nil

	}
	listRly, _ := s.resourceJoin(c, joinParams, arg.Attribute)
	if listRly != nil {
		pRes.List = listRly.List
	}
	return pRes, nil
}

// businessProduct 企业号-商品卡 .
func (s *Service) businessProduct(c context.Context, arg *dynmdl.ParamResourceOrigin) (*dynmdl.IDsReply, error) {
	pRes := &dynmdl.IDsReply{}
	// 根据商品id获取产品item
	rly, err := s.busDao.ProductDetail(c, arg.SourceID, arg.Offset, arg.Ps)
	if err != nil {
		log.Error("s.dynamicDao.ProductDetail(%s,%d,%d) error(%v)", arg.SourceID, arg.Offset, arg.Ps, err)
		return pRes, nil
	}
	pRes.Offset = rly.Offset
	pRes.HasMore = rly.HasMore
	for _, v := range rly.ItemList {
		if v == nil {
			continue
		}
		tmp := &dynmdl.Item{}
		tmp.FromResourceProduct(v)
		pRes.List = append(pRes.List, tmp)
	}
	return pRes, nil
}

// NativeForbidList-供动态使用，更新禁止上榜attr位 .
func (s *Service) NativeForbidList(c context.Context, arg *api.NativeForbidListReq) error {
	if arg.AttrForbid != 1 && arg.AttrForbid != 2 {
		return nil
	}
	//获取attr
	list, err := s.natDao.RawNativePages(c, []int64{arg.Pid})
	if err != nil {
		return err
	}
	if lv, ok := list[arg.Pid]; !ok || lv == nil {
		return xecode.NothingFound
	}
	if !list[arg.Pid].IsTopicAct() {
		return xecode.NothingFound
	}
	//开启禁止上榜
	var nowAttr int64
	if arg.AttrForbid == 1 {
		//已经是禁止上榜，则直接返回
		if list[arg.Pid].IsAttrForbid() == 1 {
			return nil
		}
		nowAttr = list[arg.Pid].Attribute | api.AttrForbidNum
	} else {
		// 已经关闭禁止上榜
		if list[arg.Pid].IsAttrForbid() != 1 {
			return nil
		}
		nowAttr = list[arg.Pid].Attribute & (api.AttrMaxNum - api.AttrForbidNum)
	}
	//更新加锁 attr
	if err = s.natDao.PageAttrUpdate(c, arg.Pid, nowAttr, list[arg.Pid].Attribute); err != nil {
		log.Error("s.natDao.PageAttrUpdate(%d,%d,%d) error(%v)", arg.Pid, nowAttr, list[arg.Pid].Attribute, err)
		return err
	}
	return nil
}

func (s *Service) resourceLive(c context.Context, arg *dynmdl.ParamResourceOrigin) (*dynmdl.IDsReply, error) {
	wid, err := strconv.ParseInt(arg.SourceID, 10, 64)
	if err != nil {
		log.Errorc(c, "Fail to parse wid, wid=%+v error=%+v", arg.SourceID, err)
		return nil, err
	}
	tempMou := &api.NativeModule{Attribute: arg.Attribute}
	isLive := tempMou.IsAttrDisplayNodeNum()
	widItems, err := s.liveDao.GetListByActId(c, wid, arg.SortType, isLive, int64(arg.Ps), int64(arg.Offset))
	if err != nil {
		log.Error("s.liveDao.GetListByActId(%d,%d) error(%v)", wid, arg.SortType, err)
		return nil, err
	}
	if widItems == nil {
		return &dynmdl.IDsReply{}, nil
	}
	var list []*dynmdl.Item
	for _, v := range widItems.List {
		if v == nil {
			continue
		}
		item := &dynmdl.Item{}
		item.FromResourceLive(v)
		list = append(list, item)
	}
	var hasMore int32
	if widItems.HasMore {
		hasMore = 1
	}
	return &dynmdl.IDsReply{
		List:    list,
		Offset:  widItems.Offset,
		HasMore: hasMore,
	}, nil
}

func (s *Service) ogvWid(c context.Context, arg *dynmdl.ParamResourceOrigin, mid int64) (*dynmdl.IDsReply, error) {
	wid, err := strconv.ParseInt(arg.SourceID, 10, 64)
	if err != nil {
		log.Errorc(c, "Fail to parse wid, wid=%+v error=%+v", arg.SourceID, err)
		return nil, err
	}
	widItems, err := s.pgcDao.QueryWid(c, int32(wid), mid)
	if err != nil {
		return nil, err
	}
	var hasMore int32
	// 首页返回 module.Num 条数据，二级页返回剩余的
	if arg.Offset > 0 {
		widItems = widItems[arg.Offset:]
	} else if len(widItems) > arg.Ps {
		hasMore = 1
		widItems = widItems[:arg.Ps]
	}
	offset := int64(arg.Offset + len(widItems))
	list := make([]*dynmdl.Item, 0, len(widItems))
	for _, v := range widItems {
		item := &dynmdl.Item{}
		item.FromResourceWidItem(v)
		list = append(list, item)
	}
	return &dynmdl.IDsReply{
		List:    list,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

func (s *Service) formatDynamic(mou *api.NativeModule, dyn *api.Dynamic, page *api.NativePage) *dynmdl.Item {
	sourceID := page.ForeignID
	if mou.Fid > 0 {
		sourceID = mou.Fid
	}
	ext := &dynmdl.UrlExt{Fid: sourceID, Types: buildDynTypes(dyn), SortType: int64(mou.DySort)}
	list := &dynmdl.Item{}
	list.FromDynamicModule(mou, ext)
	return list
}

func (s *Service) formatProgress(c context.Context, mou *api.NativeModule, mid int64) *dynmdl.Item {
	groupID := mou.Width
	if mou.Fid == 0 || groupID == 0 {
		return nil
	}
	rly, err := s.actDao.ActivityProgress(c, mou.Fid, 2, mid, []int64{groupID})
	if err != nil {
		return nil
	}
	group, ok := rly.Groups[groupID]
	if !ok {
		log.Warn("node_group=%+v not found", groupID)
		return nil
	}
	if group == nil || len(group.Nodes) == 0 {
		log.Warn("node_group=%+v is empty", groupID)
		return nil
	}
	progress := &dynmdl.Item{}
	progress.FromProgress(mou, group)
	progressModule := &dynmdl.Item{}
	progressModule.FromProgressModule(mou, []*dynmdl.Item{progress})
	return progressModule
}

func (s *Service) formatVideo(mou *api.NativeModule, sortType int64) *dynmdl.Item {
	ext := &dynmdl.UrlExt{Fid: mou.Fid, SortType: sortType}
	list := &dynmdl.Item{}
	list.FromVideoModule(mou, ext)
	return list
}

func (s *Service) formatBaseHead(mou *api.NativeModule) *dynmdl.Item {
	item := &dynmdl.Item{}
	item.FromBaseHead(mou)
	return item
}

func (s *Service) EdViewedArcs(c context.Context, req *dynmdl.EdViewedArcsReq, mid int64) (*dynmdl.EdViewedArcsRly, error) {
	activity := strconv.FormatInt(req.Sid, 10)
	rly, err := s.platDao.GetHistory(c, activity, req.Counter, mid, nil)
	if err != nil {
		return nil, err
	}
	type historySource struct {
		Aid int64 `json:"aid"`
	}
	aids := make([]int64, 0, len(rly.GetHistory()))
	for _, his := range rly.GetHistory() {
		source := &historySource{}
		if err := json.Unmarshal([]byte(his.Source), source); err != nil {
			log.Error("Fail to unmarshal HistoryContent.Source, source=%+v error=%+v", source, err)
			continue
		}
		aids = append(aids, source.Aid)
	}
	return &dynmdl.EdViewedArcsRly{Aids: aids}, nil
}

func (s *Service) formatBaseHoverButton(c context.Context, mou *api.NativeModule, mid int64) *dynmdl.Item {
	if mou.ConfSort == "" {
		return nil
	}
	confSort := &api.ConfSort{}
	if err := json.Unmarshal([]byte(mou.ConfSort), confSort); err != nil {
		log.Error("Fail to unmarshal confSort of hoverButton, confSort=%+v error=%+v", mou.ConfSort, err)
		return nil
	}
	var item *dynmdl.Item
	switch confSort.BtType {
	case api.BtTypeAppoint:
		item = s.formatHoverAppointOrigin(c, mou, confSort, mid)
	case api.BtTypeActProject:
		item = s.formatHoverActProject(c, mou, confSort, mid)
	case api.BtTypeLink:
		item = s.formatHoverLink(mou)
	default:
		log.Warn("unknown button_type=%+v", confSort.BtType)
		return nil
	}
	hoverButton := &dynmdl.Item{}
	hoverButton.FromHoverButton(mou, []*dynmdl.Item{item}, confSort)
	return hoverButton
}

func (s *Service) formatHoverAppointOrigin(c context.Context, mou *api.NativeModule, confSort *api.ConfSort, mid int64) *dynmdl.Item {
	ext := &dynmdl.ClickExt{FID: mou.Fid, Tip: confSort.Hint}
	func() {
		if mid == 0 {
			return
		}
		rly, err := s.actDao.ReserveFollowings(c, mid, []int64{mou.Fid})
		if err != nil {
			return
		}
		if data, ok := rly[mou.Fid]; ok && data != nil {
			ext.IsFollow = data.IsFollow
		}
	}()
	return &dynmdl.Item{
		Goto:       dynmdl.GotoClickButton,
		ButtonType: api.BtTypeAppoint,
		ImagesUnion: &dynmdl.ImagesUnion{
			FinishedImage:   &dynmdl.Image{Image: mou.TitleColor},
			UnfinishedImage: &dynmdl.Image{Image: mou.FontColor},
		},
		ClickExt: ext,
	}
}

func (s *Service) formatHoverActProject(c context.Context, mou *api.NativeModule, confSort *api.ConfSort, mid int64) *dynmdl.Item {
	ext := &dynmdl.ClickExt{FID: mou.Fid, Tip: confSort.Hint}
	func() {
		if mid == 0 {
			return
		}
		rly, err := s.actDao.ActRelationInfo(c, mou.Fid, mid)
		if err != nil || rly.ReserveItems == nil {
			return
		}
		if rly.ReserveItems.State == 1 {
			ext.IsFollow = true
		}
	}()
	return &dynmdl.Item{
		Goto:       dynmdl.GotoClickButton,
		ButtonType: api.BtTypeActProject,
		ImagesUnion: &dynmdl.ImagesUnion{
			FinishedImage:   &dynmdl.Image{Image: mou.TitleColor},
			UnfinishedImage: &dynmdl.Image{Image: mou.FontColor},
		},
		ClickExt: ext,
	}
}

func (s *Service) formatHoverLink(mou *api.NativeModule) *dynmdl.Item {
	return &dynmdl.Item{
		Goto:       dynmdl.GotoClickButtonV3,
		ButtonType: api.BtTypeLink,
		Image:      mou.MoreColor,
		URI:        mou.Colors,
	}
}

func (s *Service) Partition(c context.Context) (*arccli.TypesReply, error) {
	rly, err := s.arcClient.Types(c, &arccli.NoArgRequest{})
	if err != nil {
		log.Errorc(c, "Fail to request Archive.Types, error=%+v", err)
		return nil, err
	}
	return rly, nil
}

func (s *Service) PartitionV2(c context.Context, mid int64) (*dynmdl.PartitionV2Rly, error) {
	typeList, err := s.dao.ArcTypeList(c, mid)
	if err != nil {
		return nil, err
	}
	rly := &dynmdl.PartitionV2Rly{Partitions: make([]*dynmdl.Partition, 0, len(typeList))}
	for _, v := range typeList {
		partition := &dynmdl.Partition{
			ID:       v.ID,
			Name:     v.Name,
			Children: make([]*dynmdl.Partition, 0, len(v.Children)),
		}
		if len(v.Children) > 0 {
			for _, child := range v.Children {
				partition.Children = append(partition.Children, &dynmdl.Partition{ID: child.ID, Name: child.Name})
			}
		}
		rly.Partitions = append(rly.Partitions, partition)
	}
	return rly, nil
}

func extractExt4ClickInterface(ext string) (string, error) {
	if ext == "" {
		return "", nil
	}
	confExt := &api.ClickExt{}
	if err := json.Unmarshal([]byte(ext), confExt); err != nil {
		log.Error("Fail to unmarshal confExt, ext=%+v error=%+v", ext, err)
		return "", err
	}
	return confExt.Style, nil
}

func extractBaseModules(bases []*api.Module) (head, hoverButton *api.Module) {
	for _, v := range bases {
		if v == nil {
			continue
		}
		if v.NativeModule.IsBaseHead() {
			head = v
		}
		if v.NativeModule.IsBaseHoverButton() {
			hoverButton = v
		}
	}
	return
}

func buildDynTypes(dyn *api.Dynamic) string {
	if dyn == nil || len(dyn.SelectList) == 0 {
		return ""
	}
	types := ""
	tys := make([]string, 0, len(dyn.SelectList))
	for _, val := range dyn.SelectList {
		// 精选或者全选时，是不支持多选的
		if tempType, isSingle := val.JoinMultiDyTypes(); isSingle {
			types = tempType
			tys = []string{}
			break
		} else {
			tys = append(tys, tempType)
		}
	}
	if len(tys) > 0 {
		types = strings.Join(tys, ",")
	}
	return types
}

func setUnlockProgReq(click *api.NativeClick, progReqs map[int64][]int64) {
	if click.Ext == "" {
		return
	}
	clickExt := &api.ClickExt{}
	if err := json.Unmarshal([]byte(click.Ext), clickExt); err != nil {
		log.Error("Fail to unmarshal clickExt, clickExt=%+v error=%+v", click.Ext, err)
		return
	}
	if clickExt.DisplayMode == api.NeedUnLock && clickExt.UnlockCondition == api.UnLockOrder {
		if clickExt.Sid == 0 || clickExt.GroupId == 0 {
			return
		}
		progReqs[clickExt.Sid] = append(progReqs[clickExt.Sid], clickExt.GroupId)
	}
}

func reachUnlockCondition(click *api.NativeClick, progRlys map[int64]*actGRPC.ActivityProgressReply) bool {
	if click.Ext == "" {
		return true
	}
	ext := &api.ClickExt{}
	if err := json.Unmarshal([]byte(click.Ext), ext); err != nil {
		log.Error("Fail to unmarshal clickExt, clickExt=%+v error=%+v", click.Ext, err)
		return false
	}
	if ext.DisplayMode != api.NeedUnLock {
		return true
	}
	if ext.UnlockCondition == api.UnLockTime {
		return time.Now().Unix() >= ext.Stime
	}
	if ext.UnlockCondition == api.UnLockOrder {
		progRly, ok := progRlys[ext.Sid]
		if !ok || progRly == nil || len(progRly.Groups) == 0 {
			return false
		}
		group, ok := progRly.Groups[ext.GroupId]
		if !ok || group == nil || len(group.Nodes) == 0 {
			return false
		}
		for _, node := range group.Nodes {
			if ext.NodeId == node.Nid {
				return group.Total >= node.Val
			}
		}
	}
	return false
}

func finalScore(target *scoregrpc.ScoreTarget) string {
	if target.GetShowFlag() == 1 {
		return "暂无评分"
	}
	if fs := target.GetFixScore(); fs != "" && fs != "0" && fs != "0.0" {
		return target.GetFixScore()
	}
	if target.GetTargetScore() == "0" || target.GetTargetScore() == "0.0" {
		return "暂无评分"
	}
	return target.GetTargetScore()
}

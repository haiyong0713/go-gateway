package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/newmont/service/api"
	"go-gateway/app/app-svr/newmont/service/internal/model/section"
	secmdl "go-gateway/app/app-svr/newmont/service/internal/model/section"

	locApi "git.bilibili.co/bapis/bapis-go/community/service/location"
	tus "git.bilibili.co/bapis/bapis-go/datacenter/service/titan"
	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	displayApi "git.bilibili.co/bapis/bapis-go/platform/interface/display"
	resV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"

	"github.com/pkg/errors"
)

const (
	_initSidebarKey   = "sidebar_%d_%d_%s"
	MType_MineSection = 1
	MType_HomeTab     = 2

	ModuleType_HomeTopIcon = 10
	// 港澳台+海外策略组id
	_areaPolicy = 1665

	EntranceTypeHome    = 0
	EntranceTypeSidebar = 2
	EntranceTypeModule  = 3
	EntranceTypeRegion  = 1
	EntranceTypeDynamic = 4

	//module style enum
	ModuleStyle_OP = 3
)

func (s *Service) HomeSections(ctx context.Context, req *api.HomeSectionsRequest) (*api.HomeSectionsReply, error) {
	var (
		res      = new(api.HomeSectionsReply)
		transReq = &secmdl.SectionReq{
			Plat:    req.Plat,
			Build:   req.Build,
			Mid:     req.Mid,
			Lang:    req.Lang,
			Channel: req.Channel,
			Ip:      req.Ip,
			Buvid:   req.Buvid,
		}
		bottomTabModuleId = int64(9)
		err               error
	)

	res.Sections, err = s.getSections(ctx, transReq, s.entryModuleCache, EntranceTypeHome, s.preIconCache)
	if err != nil {
		return nil, err
	}
	// TODO: 假如对于数量有限制的module，可以考虑在此处做操作
	res.Sections = s.mallTabFilter(res.Sections)
	// 此处需要有骚操作-对底tab的排序做调整，将dialog opener强制插入3个或者5个时的中间位置
	for i, m := range res.Sections {
		module := m
		index := i
		if module.Id != bottomTabModuleId {
			continue
		}
		normalItem := make([]*api.SectionItem, 0)
		var openerItem *api.SectionItem

		for _, ele := range module.Items {
			item := ele
			if item.OpLinkType == api.SectionItemOpLinkType_DIALOG_OPENER {
				if openerItem == nil {
					openerItem = item
				}
				continue
			}
			normalItem = append(normalItem, item)
		}
		// 有弹窗入口的时侯，普通icon数量够，就给到5个，产品说配错剁手
		if openerItem != nil && len(normalItem) >= 4 {
			res.Sections[index].Items = []*api.SectionItem{normalItem[0], normalItem[1], openerItem, normalItem[2], normalItem[3]}
			continue
		}
		// 普通icon数量不够，只有2个，就总共给到3个，产品说配错剁手
		if openerItem != nil && len(normalItem) == 2 {
			res.Sections[index].Items = []*api.SectionItem{normalItem[0], openerItem, normalItem[1]}
		}
		if openerItem == nil && ((req.Plat == 0 && req.Build > 6270199) || (req.Plat == 1 && req.Build > 62700099)) {
			tabIds := make([]int64, 0)
			for _, i := range res.Sections[index].Items {
				tabIds = append(tabIds, i.Id)
			}
			log.Warn("【HomeSection 告警】有下发未带大加号的底tab内容, req(%v+), ip(%s), res-ids(%+v)", req, metadata.String(ctx, metadata.RemoteIP), tabIds)
		}
	}
	return res, nil
}

func (s *Service) MineSections(ctx context.Context, req *api.MineSectionsRequest) (*api.MineSectionsReply, error) {
	transReq := &secmdl.SectionReq{
		Plat:       req.Plat,
		Build:      req.Build,
		Mid:        req.Mid,
		Lang:       req.Lang,
		Channel:    req.Channel,
		Ip:         req.Ip,
		IsUploader: req.IsUploader,
		IsLiveHost: req.IsLiveHost,
		FansCount:  req.FansCount,
		Buvid:      req.Buvid,
	}
	sections, err := s.getSections(ctx, transReq, s.mineModuleCache, EntranceTypeSidebar, s.IconCache)
	if err != nil {
		return nil, err
	}
	return &api.MineSectionsReply{
		Sections: sections,
	}, nil
}

func buildConditionForTus(tusValues []string) []string {
	var conditions []string
	for _, v := range tusValues {
		conditions = append(conditions, fmt.Sprintf("tag_%s==1", v))
	}
	return conditions
}

func (s *Service) getSections(c context.Context, req *section.SectionReq, modules map[int32][]*section.ModuleInfo, entranceType int64, iconCache map[int64]*api.MngIcon) (sections []*api.Section, err error) {
	sections = make([]*api.Section, 0)
	var (
		mWhite          = make(map[int64]bool) // module白名单
		sWhite          = make(map[int64]bool) // 模块白名单
		sRed            = make(map[int64]bool) // 模块红点
		mutex           sync.Mutex
		eg              = errgroup.WithContext(c)
		sids            []int64
		moduleIDs       []int64
		ehMap           = make(map[int64]bool)         // 二级模块入口屏蔽名单
		moduleHiddenMap = make(map[int64]bool)         //一级模块入口屏蔽名单
		icMap           = make(map[int64]*api.MngIcon) // 运营icon名单
		showAuthDef     = map[int64]*locApi.ZoneLimitAuth{}
		policiesAuth    map[int64]*locApi.ZoneLimitAuth
		dynamicConfigs  = make(map[int64]*section.DynamicConf)
		//推荐服务按顺序可出的tab
		recommendTopTabIDs []int64
		//港澳台垂类tab
		isRecommendTab  bool
		allHitTusValues = make(map[string]struct{})
	)
	module, ok := modules[req.Plat]
	if !ok {
		return nil, errors.Errorf("Failed to find module plat(%d), buvid(%+v)", req.Plat, req.Buvid)
	}
	if req.Mid > 0 && len(s.sidebarTusValues) > 0 {
		//数平人群包相关
		eg.Go(func(ctx context.Context) error {
			reply, err := s.tusClient.CheckTagBatch(ctx, &tus.TusBatchRequest{
				Uid:       strconv.FormatInt(req.Mid, 10),
				Condition: buildConditionForTus(s.sidebarTusValues),
				BizType:   "mgr",
				UidType:   "mid",
				Sign:      fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s%s", "c26de654b6fa", strconv.FormatInt(req.Mid, 10))))),
			})
			if err != nil {
				log.Error("s.tusClient.CheckTagBatch error(%+v), mid(%d)", err, req.Mid)
				return nil
			}
			if reply.Code != 0 {
				log.Error("s.tusClient.CheckTagBatch code(%d), mid(%d)", reply.Code, req.Mid)
				return nil
			}
			if len(reply.Hits) != len(s.sidebarTusValues) {
				log.Error("s.tusClient.CheckTagBatch incorrect length(%d) with sidebar tus values(%d)", len(reply.Hits), len(s.sidebarTusValues))
				return nil
			}
			for index, hit := range reply.Hits {
				if !hit {
					continue
				}
				allHitTusValues[s.sidebarTusValues[index]] = struct{}{}
			}
			return nil
		})
	}
	// 白名单+红点接口
	for _, m := range module {
		moduleIDs = append(moduleIDs, m.ID)
		if req.Mid > 0 && m.WhiteURL != "" {
			tmpID := m.ID
			tmpURL := m.WhiteURL
			eg.Go(func(ctx context.Context) (err error) {
				ok, err := s.sectionDao.WhiteCheck(ctx, tmpURL, req.Mid, req.Buvid)
				if err != nil {
					log.Error("s.show.WhiteCheck error(%+v) url(%s) mid(%d) buvid(%s)", err, tmpURL, req.Mid, req.Buvid)
					return nil
				}
				if ok {
					mutex.Lock()
					mWhite[tmpID] = ok
					mutex.Unlock()
				}
				return nil
			})
		}
		if m.IsRecommendTab() { //港澳台垂类tab，需要拿推荐服务接口数据进行过滤排序
			isRecommendTab = true
		}
		sidebars, ok := s.sideBarByModule[fmt.Sprintf(_initSidebarKey, m.Plat, m.ID, req.Lang)]
		if !ok {
			continue
		}
		for _, sd := range sidebars {
			sids = append(sids, sd.ID)
			tmpID := sd.ID
			tmpwURL := sd.WhiteURL
			tmprURL := sd.Red
			tmpGrayToken := sd.GrayToken
			tmpDynamicConfUrl := sd.DynamicConfUrl

			// 判断版本限制， 降低其他无效的判断数量
			if !CheckLimit(s.sideBarLimitCache[sd.ID], req.Build) {
				continue
			}

			// 检查二级模块的白名单和红点
			if req.Mid > 0 {
				if tmpwURL != "" {
					eg.Go(func(ctx context.Context) error {
						ok, err := s.sectionDao.WhiteCheck(ctx, tmpwURL, req.Mid, req.Buvid)
						if err != nil {
							log.Error("s.show.WhiteCheck error(%+v) url(%s) mid(%d) buvid(%s) ", err, tmpwURL, req.Mid, req.Buvid)
							return nil
						}
						if ok {
							mutex.Lock()
							sWhite[tmpID] = ok
							mutex.Unlock()
						}
						return nil
					})
				}
				// 单独去除首页顶tab icon的红点请求，因为游戏中心的数据返回和通常的返回不一致，由网关处理
				if tmprURL != "" && m.ID != ModuleType_HomeTopIcon {
					eg.Go(func(ctx context.Context) error {
						ok, err := s.sectionDao.RedDot(ctx, req.Mid, tmprURL)
						if err != nil {
							log.Error("s.show.RedDot error(%+v) url(%s) mid(%d) ", err, tmprURL, req.Mid)
							return nil
						}
						if ok {
							mutex.Lock()
							sRed[tmpID] = true
							mutex.Unlock()
						}
						return nil
					})
				}
			}

			if tmpDynamicConfUrl != "" {
				eg.Go(func(ctx context.Context) error {
					conf, err := s.sectionDao.FetchDynamicConf(ctx, tmpDynamicConfUrl, req.Mid, req.Buvid)
					if err != nil {
						log.Error("s.show.FetchDynamicConf error(%+v) url(%s) mid(%d) buvid(%s)", err, tmpwURL, req.Mid, req.Buvid)
						return nil
					}
					if conf != nil {
						mutex.Lock()
						dynamicConfigs[tmpID] = conf
						mutex.Unlock()
					}
					return nil
				})
			}

			// 检查二级模块的灰度分桶
			if tmpGrayToken != "" {
				eg.Go(func(ctx context.Context) (err error) {
					tokens := strings.Split(tmpGrayToken, "@")
					ip := req.Ip
					if ip == "" {
						ip = metadata.String(c, metadata.RemoteIP)
					}

					bwReq := &resV2.CheckCommonBWListReq{
						Oid:    strconv.FormatInt(req.Mid, 10),
						Token:  tokens[0],
						UserIp: ip,
						LargeOid: &resV2.LargeOidContent{
							Buvid: req.Buvid,
							Mid:   req.Mid,
						},
						LargeToken: tokens[1],
					}
					if rep, err := s.resClentV2.CheckCommonBWList(ctx, bwReq); err != nil {
						log.Error("CheckCommonBWList error(%+v) mid(%d) buvid(%s) token(%s)", err, req.Mid, req.Buvid, tmpGrayToken)
					} else {
						if rep.IsInList {
							mutex.Lock()
							sWhite[tmpID] = true
							mutex.Unlock()
						}
					}
					return nil
				})
			}

			if sd.AreaPolicy > 0 {
				showAuthDef[sd.AreaPolicy] = &locApi.ZoneLimitAuth{}
			}
		}
	}

	// 显隐服务调用
	itemsIdDict := make(map[int64]*displayApi.IconContent)
	if entranceType == EntranceTypeSidebar {
		eg.Go(func(ctx context.Context) (err error) {
			disReq := &displayApi.IconListReq{
				Mid: req.Mid,
			}
			var itemsInfo *displayApi.IconListResp
			if itemsInfo, err = s.displayClient.EverythingAboutIcon(ctx, disReq); err != nil {
				log.Error("Display interface error req:%+v, err: %v", req, err)
				return nil
			}
			for _, each := range itemsInfo.Item {
				itemsIdDict[each.Id] = each
			}
			return
		})
	}

	// 渠道入口隐藏，主要针对游戏入口
	eg.Go(func(ctx context.Context) (err error) {
		ehReq := &api.SectionIsHiddenRequest{
			Build:   int64(req.Build),
			Plat:    req.Plat,
			Channel: req.Channel,
			OidItems: map[int64]*api.OidList{
				entranceType: {
					Oids: sids,
				},
				EntranceTypeModule: {
					Oids: moduleIDs,
				},
			},
		}
		eh, err := s.SectionIsHidden(ctx, ehReq)
		if err != nil {
			log.Error("s.EntrancesIsHidden err(%+v) req(%+v)", err, req)
			return nil
		}
		ehMap = eh.Infos
		moduleHiddenMap = eh.ModuleInfos
		return nil
	})

	// 装扮icon
	eg.Go(func(ctx context.Context) (err error) {
		miReq := &api.MngIconRequest{
			Oids: sids,
			Plat: req.Plat,
			Mid:  req.Mid,
		}
		ic, err := s.mngIcon(ctx, miReq, iconCache)
		if err != nil {
			log.Error("s.MngIcon err(%+v) req(%+v)", err, req)
			return nil
		}
		if ic != nil {
			icMap = ic.Info
		}
		return nil
	})

	// 统一调用ip location服务
	if _, ok := showAuthDef[_areaPolicy]; !ok && isRecommendTab {
		showAuthDef[_areaPolicy] = &locApi.ZoneLimitAuth{}
	}
	if len(showAuthDef) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			userIp := req.Ip
			if userIp == "" {
				userIp = metadata.String(ctx, metadata.RemoteIP)
			}
			zoneReq := &locApi.ZoneLimitPoliciesReq{
				UserIp:       userIp,
				DefaultAuths: showAuthDef,
			}
			var reply *locApi.ZoneLimitPoliciesReply
			if reply, err = s.locClient.ZoneLimitPolicies(ctx, zoneReq); err != nil {
				log.Error("Failed to get ZoneLimitPolicies: %+v, ip: %s", err, userIp)
				err = nil
				return
			}
			//log.Error("get ZoneLimitPolicies(%+v), showAuthDef(%+v)", reply.Auths, showAuthDef)
			policiesAuth = reply.Auths
			if isRecommendTab {
				//港澳台垂类tab，需要拿服务接口数据进行过滤排序
				//过滤非港澳台+海外ip
				if auth, ok := policiesAuth[_areaPolicy]; !ok || auth.Play == locApi.Status_Unknown || auth.Play == locApi.Status_Forbidden {
					return nil
				}
				recommendTopTabIDs = s.fetchHMTChannelTab(ctx, req.Mid, int64(req.Plat), req.Buvid)
				return nil
			}
			return nil
		})
	}

	//nolint:errcheck
	eg.Wait()

	// 计算up主和主播对应的下发条件
	var conditions = []string{"0", "0"}
	if req.IsUploader {
		conditions[0] = "1"
	}
	if req.IsLiveHost {
		conditions[1] = "1"
	}
	opLoadCondition := strings.Join(conditions, "")

	for _, m := range module {
		if moduleHiddenMap[m.ID] {
			continue
		}
		if m.WhiteURL != "" && !mWhite[m.ID] {
			continue
		}
		// 当前模块是否只取一个item, 创作/直播运营位
		var isOnly bool
		if m.Style == ModuleStyle_OP {
			isOnly = true
		}
		items := make([]*api.SectionItem, 0)
		if side, ok := s.sideBarByModule[fmt.Sprintf(_initSidebarKey, m.Plat, m.ID, req.Lang)]; ok {
			for _, sd := range side {
				// 判断版本限制， 降低其他无效的判断数量
				if !CheckLimit(s.sideBarLimitCache[sd.ID], req.Build) {
					continue
				}
				// 判断白名单逻辑
				if !CheckWhite(req.Mid, sd, sWhite) {
					continue
				}
				// 判断是否在入口屏蔽配置里
				if isHidden, ok := ehMap[sd.ID]; ok && isHidden {
					continue
				}
				//人群包
				if pass := func() bool {
					if sd.TusValue == "" {
						return true
					}
					if _, ok := allHitTusValues[sd.TusValue]; ok {
						return true
					}
					return false
				}(); !pass {
					continue
				}
				if sd.AreaPolicy > 0 {
					auth, ok := policiesAuth[sd.AreaPolicy]

					//log.Error("sd.AreaPolicy check sid(%d), auth(%+v), ok(%+v), AreaPolicy(%+v), ShowPurposed(%+v)", sd.ID, auth.Play, ok, sd.AreaPolicy, sd.ShowPurposed)

					if ok && auth.Play != locApi.Status_Unknown {
						if auth.Play == locApi.Status_Forbidden {
							continue
						}
					} else if sd.ShowPurposed == 1 {
						continue
					}
				}
				// 判断显隐逻辑
				if len(itemsIdDict) > 0 {
					if _, ok := itemsIdDict[sd.ID]; !ok {
						continue
					}
				}
				var (
					mngIcon   *api.MngIcon
					redDot    int32
					globalRed = int32(sd.GlobalRed)
				)
				if sd.Red != "" && sRed[sd.ID] {
					redDot = 1
				}
				if ic, ok := icMap[sd.ID]; ok {
					mngIcon = ic
				}
				// 当模块为我的页，则判断是否为up主和主播
				if entranceType == EntranceTypeSidebar && sd.OpLoadCondition != "" {
					var isMatchCondition bool
					cs := strings.Split(sd.OpLoadCondition, ",")

					for _, c := range cs {
						if opLoadCondition == c {
							isMatchCondition = true
							break
						}
					}
					if !isMatchCondition {
						continue
					}
				}
				if (sd.OpFansLimit > 0 && req.FansCount < sd.OpFansLimit) || (sd.OpFansLimit < 0 && req.FansCount >= sd.OpFansLimit*-1) {
					continue
				}

				// 动态设置属性
				param := sd.Param
				name := sd.Name
				logo := sd.Logo
				logoSelected := sd.LogoSelected

				if conf, hasDynamicConf := dynamicConfigs[sd.ID]; hasDynamicConf {
					if conf.Param != "" {
						param = conf.Param
					}
					if conf.Name != "" {
						name = conf.Name
					}
					if opDefaultIcon, ok := s.operationIcon[conf.DefaultIcon]; ok {
						logo = opDefaultIcon
					}
					if opSelectedIcon, ok := s.operationIcon[conf.SelectedIcon]; ok {
						logoSelected = opSelectedIcon
					}
				}

				items = append(items, &api.SectionItem{
					Id:                   sd.ID,
					Title:                name,
					Uri:                  param,
					Icon:                 logo,
					NeedLogin:            int32(sd.NeedLogin),
					RedDot:               redDot,
					GlobalRedDot:         globalRed,
					MngIcon:              mngIcon,
					RedDotForNew:         sd.RedDotForNew,
					OpTitle:              sd.OpTitle,
					OpSubTitle:           sd.OpSubTitle,
					OpTitleIcon:          sd.OpTitleIcon,
					OpLinkType:           api.SectionItemOpLinkType_Enum(sd.OpLinkType),
					OpLinkText:           sd.OpLinkText,
					OpLinkIcon:           sd.OpLinkIcon,
					TabId:                sd.TabID,
					Animate:              sd.Animate,
					LogoSelected:         logoSelected,
					RedDotUrl:            sd.Red,
					OpTitleColor:         sd.OpTitleColor,
					OpBackgroundColor:    sd.OpBackgroundColor,
					OpLinkContainerColor: sd.OpLinkContainerColor,
				})
				// 假如当前模块只取一个，则直接跳出
				if isOnly {
					break
				}
			}
		}

		if m.MType == MType_HomeTab && len(items) == 0 {
			continue
		}

		if m.IsRecommendTab() {
			items = reorderRecommendTopTab(items, recommendTopTabIDs)
		}

		sections = append(sections, &api.Section{
			Id:              m.ID,
			Title:           m.Title,
			Style:           m.Style,
			ButtonName:      m.ButtonName,
			ButtonUrl:       m.ButtonURL,
			ButtonIcon:      m.ButtonIcon,
			ButtonStyle:     m.ButtonStyle,
			TitleColor:      m.TitleColor,
			Subtitle:        m.Subtitle,
			SubtitleUrl:     m.SubtitleURL,
			SubtitleColor:   m.SubtitleColor,
			Background:      m.Background,
			BackgroundColor: m.BackgroundColor,
			Items:           items,
			AuditShow:       m.AuditShow,
			IsMng:           m.IsMng,
			OpStyleType:     m.OpStyleType,
		})
	}
	return
}

// CheckLimit is
func CheckLimit(limit []*secmdl.SideBarLimit, build int32) bool {
	if len(limit) == 0 {
		return false
	}
	for _, l := range limit {
		if !secmdl.ValidBuild(int64(build), int64(l.Build), l.Condition) {
			return false
		}
	}
	return true
}

func (s *Service) fetchHMTChannelTab(c context.Context, mid, plat int64, buvid string) []int64 {
	req := &hmtchannelgrpc.ChannelTabReq{
		Mid:   mid,
		Buvid: buvid,
		Plat:  plat,
	}
	res, err := s.hmtChannelClient.ChannelTab(c, req)
	if err != nil {
		log.Error("s.fetchHMTChannelTab req:%+v, error:%+v", req, err)
		return nil
	}
	return res.Ids
}

// CheckWhite is
func CheckWhite(mid int64, sd *secmdl.SideBar, sWhite map[int64]bool) bool {
	_, ok := sWhite[sd.ID]
	hasWhite := sd.WhiteURL != ""
	hasGray := sd.GrayToken != ""
	if hasGray {
		return ok
	}
	if hasWhite {
		return ok || (mid == 0 && sd.WhiteURLShow == 1)
	}
	return true
}

func reorderRecommendTopTab(in []*api.SectionItem, recommendTabIDs []int64) []*api.SectionItem {
	var (
		itemsMap = make(map[int64]*api.SectionItem, len(in))
		out      []*api.SectionItem
	)
	//如果没有拿到推荐tab的数据则使用产品默认配置
	if len(recommendTabIDs) == 0 {
		return in
	}
	for _, v := range in {
		itemsMap[v.Id] = v
	}
	for _, id := range recommendTabIDs {
		item, ok := itemsMap[id]
		if !ok {
			continue
		}
		out = append(out, item)
	}
	return out
}

// 会员购底tab个性化需求，临时用来过滤有两个会员购底tab的情况
func (s *Service) mallTabFilter(sections []*api.Section) []*api.Section {
	var (
		bottomTabModuleId                     = int64(9)
		hasDefaultMallItem, hasCustomMallItem bool
	)

	for _, section := range sections {
		if section.Id != bottomTabModuleId {
			continue
		}
		for _, item := range section.Items {
			if item == nil {
				continue
			}

			if _, ok := s.c.MallDefaultIDMap[strconv.FormatInt(item.Id, 10)]; ok {
				hasDefaultMallItem = true
			}
			if _, ok := s.c.MallCustomIDMap[strconv.FormatInt(item.Id, 10)]; ok {
				hasCustomMallItem = true
			}
		}
		if !hasDefaultMallItem || !hasCustomMallItem { //没有默认的会员购底tab || 无定制化会员购底tab
			return sections
		}
		//如果有定制化的会员购底tab则过滤掉默认的会员购底tab
		section.Items = s.defaultMallItemFilter(section.Items)
	}
	return sections
}

func (s *Service) defaultMallItemFilter(in []*api.SectionItem) []*api.SectionItem {
	var out []*api.SectionItem
	for _, item := range in {
		if item == nil {
			continue
		}
		if _, ok := s.c.MallDefaultIDMap[strconv.FormatInt(item.Id, 10)]; ok {
			continue
		}
		out = append(out, item)
	}
	return out
}

func (s *Service) SectionIsHidden(ctx context.Context, req *api.SectionIsHiddenRequest) (*api.SectionIsHiddenReply, error) {
	res := new(api.SectionIsHiddenReply)
	if len(req.OidItems) == 0 { //使用老参数 oids,type
		if !legalEntranceType(int64(req.Otype)) {
			return nil, errors.Errorf("illegal entrance type")
		}
		res.Infos = s.hitEntranceHidden(ctx, req.Oids, int64(req.Otype), req.Build, req.Channel, req.Plat)
		return res, nil
	}

	for otype, item := range req.OidItems {
		if !legalEntranceType(otype) || item == nil {
			continue
		}
		switch otype {
		case EntranceTypeModule:
			res.ModuleInfos = s.hitEntranceHidden(ctx, item.Oids, otype, req.Build, req.Channel, req.Plat)
		case EntranceTypeDynamic:
			res.HideDynamic = s.hitDynamicHidden(ctx, req.Build, req.Channel, req.Plat)
		default:
			res.Infos = s.hitEntranceHidden(ctx, item.Oids, otype, req.Build, req.Channel, req.Plat)
		}
	}
	return res, nil
}

func legalEntranceType(entranceType int64) bool {
	return entranceType == EntranceTypeHome || entranceType == EntranceTypeSidebar ||
		entranceType == EntranceTypeRegion || entranceType == EntranceTypeModule || entranceType == EntranceTypeDynamic
}

func (s *Service) hitEntranceHidden(ctx context.Context, oids []int64, otype, build int64, channel string, plat int32) map[int64]bool {
	tmpHidden := make(map[int64]bool, len(oids))
	for _, oid := range oids {
		for _, v := range s.hiddenCache {
			if v == nil || v.Info == nil {
				continue
			}
			hasOidNeedHidden := func() bool {
				switch otype {
				case EntranceTypeHome:
					return v.Info.Sid == oid
				case EntranceTypeRegion:
					return v.Info.Rid == oid
				case EntranceTypeSidebar:
					return v.Info.Cid == oid
				case EntranceTypeModule:
					return v.Info.ModuleId == oid
				default:
					return false
				}
			}()

			if !hasOidNeedHidden {
				continue
			}
			if s.needHide(ctx, channel, v, plat, build) {
				tmpHidden[oid] = true
				break
			}
		}
	}
	return tmpHidden
}

func (s *Service) needHide(ctx context.Context, channel string, hiddenInfo *api.HiddenInfo, plat int32, build int64) bool {
	// 判断渠道
	if !hitHiddenChannel(channel, hiddenInfo.Info) {
		return false
	}
	// 判断版本限制
	var hitBuildLimit bool
	for _, b := range hiddenInfo.Limit {
		if plat != b.Plat {
			continue
		}
		if secmdl.ValidBuild(build, b.Build, b.Conditions) {
			hitBuildLimit = true // 命中了其中某条版本判断
			break
		}
	}
	if !hitBuildLimit {
		return false
	}
	// 判断地区限制
	ip := metadata.String(ctx, metadata.RemoteIP)
	pidStr := strconv.FormatInt(hiddenInfo.Info.Pid, 10)
	authRep, e := s.locClient.AuthPIDs(ctx, &locApi.AuthPIDsReq{Pids: pidStr, IpAddr: ip})
	if e != nil || authRep == nil {
		log.Error("s.locGRPC.AuthPIDs arg(%+v) err(%+v) or authRep=nil", &locApi.AuthPIDsReq{Pids: pidStr, IpAddr: ip}, e)
		return false
	}
	const forbidden = int64(1)
	if auth, ok := authRep.Auths[hiddenInfo.Info.Pid]; !ok || auth.Play != forbidden {
		return false
	}
	return true
}

func hitHiddenChannel(channel string, hidden *api.Hidden) bool {
	hit := matchChannelWithHiddenMetaInfo(channel, hidden.ChannelMap, hidden.ChannelFuzzy)
	switch hidden.HiddenCondition {
	case "exclude":
		return !hit
	default:
		return hit
	}
}

func matchChannelWithHiddenMetaInfo(channel string, channelMap map[string]string, channelFuzz []string) bool {
	if _, ok := channelMap[channel]; ok {
		return true
	}
	for _, v := range channelFuzz {
		vSubstr := strings.Trim(v, "%") //去掉%之后匹配
		if strings.Contains(channel, vSubstr) {
			return true
		}
	}
	return false
}

func (s *Service) hitDynamicHidden(ctx context.Context, build int64, channel string, plat int32) bool {
	hideDynamic := int64(1)
	for _, v := range s.hiddenCache {
		if v.Info.HideDynamic != hideDynamic {
			continue
		}
		if s.needHide(ctx, channel, v, plat, build) {
			return true
		}
	}
	return false
}

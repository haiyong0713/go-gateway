package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-resource/interface/model/location"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/model"

	articleApi "git.bilibili.co/bapis/bapis-go/article/service"
	locApi "git.bilibili.co/bapis/bapis-go/community/service/location"
	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	displayApi "git.bilibili.co/bapis/bapis-go/platform/interface/display"
	videoApi "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/pkg/errors"
)

// loadCardCache load all card cache
func (s *Service) loadCardCache() {
	now := time.Now()
	hdm, err := s.show.PosRecs(context.TODO(), now)
	if err != nil {
		log.Error("s.show.PosRecs error(%v)", err)
		return
	}
	itm, aids, err := s.show.RecContents(context.TODO(), now)
	if err != nil {
		log.Error("s.show.RecContents error(%v)", err)
		return
	}
	tmpItem := map[int]map[int64]*model.ShowItem{}
	for recid, aid := range aids {
		tmpItem[recid] = s.fromCardAids(context.TODO(), aid)
	}
	tmp := s.mergeCard(context.TODO(), hdm, itm, tmpItem, now)
	s.cardCache = tmp
}

const (
	MType_MineSection = 1
	MType_HomeTab     = 2

	ModuleType_HomeTopIcon = 10

	// 港澳台+海外策略组id
	_areaPolicy = 1665
)

func (s *Service) loadSideBarCache() {
	var (
		now      = time.Now()
		sidebar  []*model.SideBar
		sbm      = make(map[string][]*model.SideBar)
		limits   map[int64][]*model.SideBarLimit
		sModules map[int32][]*model.ModuleInfo
		eModules map[int32][]*model.ModuleInfo
	)
	eg := errgroup.WithCancel(context.Background())
	eg.Go(func(ctx context.Context) (err error) {
		sidebar, limits, err = s.show.SideBar(ctx, now)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取我的页模块属性配置
		sModules, err = s.show.SideBarModules(ctx, MType_MineSection)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取首页tab模块属性配置
		eModules, err = s.show.SideBarModules(ctx, MType_HomeTab)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("loadSideBarCache err(%+v)", err)
		return
	}
	for _, item := range sidebar {
		key := fmt.Sprintf(_initSidebarKey, item.Plat, item.Module, item.Language)
		sbm[key] = append(sbm[key], item)
	}
	s.sideBarByModule = sbm
	s.sideBarCache = sidebar
	s.sideBarLimitCache = limits
	s.mineModuleCache = sModules
	s.entryModuleCache = eModules
	//log.Warn("loadSideBarCache success")
}

// SideBars get side bars
func (s *Service) SideBars(c context.Context) (res *model.SideBars) {
	res = &model.SideBars{
		SideBar: s.sideBarCache,
		Limit:   s.sideBarLimitCache,
	}
	return res
}

// RegionCard get voice card.
func (s *Service) RegionCard(c context.Context, plat int8, build int) (res *model.Head, err error) {
	res = &model.Head{}
	sw := s.cardCache[plat]
	if sw == nil {
		return
	}
	if model.InvalidBuild(build, sw.Build, sw.Condition) {
		return
	}
	*res = *sw
	res.FillBuildURI(plat, build)
	return
}

// fromCardAids get Aids.
func (s *Service) fromCardAids(c context.Context, aids []int64) (data map[int64]*model.ShowItem) {
	var (
		args = &arcgrpc.ArcsRequest{Aids: aids}
		arcs *arcgrpc.ArcsReply
		as   map[int64]*arcgrpc.Arc
		err  error
	)
	if arcs, err = s.arcGRPC.Arcs(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	if as = arcs.GetArcs(); len(as) == 0 {
		//log.Warn("s.arcRPC.Archives3(%v) length is 0", aids)
		return
	}
	data = map[int64]*model.ShowItem{}
	for _, aid := range aids {
		if arc, ok := as[aid]; ok {
			if !arc.IsNormal() {
				continue
			}
			i := &model.ShowItem{}
			i.FromArchivePB(arc)
			data[aid] = i
		}
	}
	return
}

// mergeCard merge Card
func (s *Service) mergeCard(_ context.Context, hdm map[int8][]*model.Card, itm map[int][]*model.Content, tmpItems map[int]map[int64]*model.ShowItem, now time.Time) (res map[int8]*model.Head) {
	res = map[int8]*model.Head{}
	for plat, hds := range hdm {
		for _, hd := range hds {
			var (
				sis []*model.ShowItem
			)
			its, ok := itm[hd.ID]
			if !ok {
				its = []*model.Content{}
			}
			tmpItem, ok := tmpItems[hd.ID]
			if !ok {
				tmpItem = map[int64]*model.ShowItem{}
			}
			switch hd.Type {
			case 1:
				for _, ci := range its {
					si := s.fillCardItem(ci, tmpItem)
					if si.Title != "" {
						sis = append(sis, si)
					}
				}
			default:
				continue
			}
			if len(sis) == 0 {
				continue
			}
			sw := &model.Head{
				CardID:    hd.ID,
				Title:     hd.Title,
				Type:      hd.TypeStr,
				Build:     hd.Build,
				Condition: hd.Condition,
				Plat:      hd.Plat,
			}
			if hd.Cover != "" {
				sw.Cover = hd.Cover
			}
			switch sw.Type {
			case model.GotoDaily:
				sw.Date = now.Unix()
				sw.Param = hd.Rvalue
				sw.URI = hd.URI
				sw.Goto = hd.Goto
			}
			sw.Body = sis
			res[plat] = sw
		}
	}
	return
}

// fillCardItem fill card
func (s *Service) fillCardItem(csi *model.Content, tsi map[int64]*model.ShowItem) (si *model.ShowItem) {
	si = &model.ShowItem{}
	switch csi.Type {
	case model.CardGotoAv:
		si.Goto = model.GotoAv
		si.Param = csi.Value
	}
	si.URI = model.FillURI(si.Goto, si.Param)
	if si.Goto == model.GotoAv {
		aid, err := strconv.ParseInt(si.Param, 10, 64)
		if err != nil {
			log.Error("strconv.ParseInt(%s) error(%v)", si.Param, err)
		} else {
			if it, ok := tsi[aid]; ok {
				si = it
				if csi.Title != "" {
					si.Title = csi.Title
				}
			} else {
				si = &model.ShowItem{}
			}
		}
	}
	return
}

// Audit all audit config.
func (s *Service) Audit(c context.Context) map[string][]int {
	return s.auditCache
}

// WebRcmd all web_rcmd and card.
func (s *Service) WebRcmd(c context.Context, req *pb.NoArgRequest) (res *pb.WebRcmdReply, err error) {
	res = &pb.WebRcmdReply{
		Rcmd:     s.webRcmd,
		RcmdCard: s.webRcmdCard,
	}
	return
}

func (s *Service) loadHiddenCache() {
	hiddens, limits, err := s.show.Hiddens(context.Background(), time.Now())
	if err != nil {
		log.Error("s.show.Hiddens error(%v)", err)
		return
	}
	res := []*pb.HiddenInfo{}
	for _, h := range hiddens {
		if h == nil {
			continue
		}
		if hl, ok := limits[h.Id]; ok {
			res = append(res, &pb.HiddenInfo{Info: h, Limit: hl})
		}
	}
	s.HiddenCache = res
	log.Info("loadHiddenCache success")
}

func hitHiddenChannel(channel string, hidden *pb.Hidden) bool {
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

func legalEntranceType(entranceType int64) bool {
	return entranceType == model.EntranceTypeHome || entranceType == model.EntranceTypeSidebar ||
		entranceType == model.EntranceTypeRegion || entranceType == model.EntranceTypeModule || entranceType == model.EntranceTypeDynamic
}

func (s *Service) hitEntranceHidden(ctx context.Context, oids []int64, otype, build int64, channel string, plat int32) map[int64]bool {
	tmpHidden := make(map[int64]bool, len(oids))
	for _, oid := range oids {
		for _, v := range s.HiddenCache {
			if v == nil || v.Info == nil {
				continue
			}
			hasOidNeedHidden := func() bool {
				switch otype {
				case model.EntranceTypeHome:
					return v.Info.Sid == oid
				case model.EntranceTypeRegion:
					return v.Info.Rid == oid
				case model.EntranceTypeSidebar:
					return v.Info.Cid == oid
				case model.EntranceTypeModule:
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

func (s *Service) hitDynamicHidden(ctx context.Context, build int64, channel string, plat int32) bool {
	hideDynamic := int64(1)
	for _, v := range s.HiddenCache {
		if v.Info.HideDynamic != hideDynamic {
			continue
		}
		if s.needHide(ctx, channel, v, plat, build) {
			return true
		}
	}
	return false
}

func (s *Service) needHide(ctx context.Context, channel string, hiddenInfo *pb.HiddenInfo, plat int32, build int64) bool {
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
		if model.ValidBuild(build, b.Build, b.Conditions) {
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
	authRep, e := s.locGRPC.AuthPIDs(ctx, &locApi.AuthPIDsReq{Pids: pidStr, IpAddr: ip})
	if e != nil || authRep == nil {
		log.Error("s.locGRPC.AuthPIDs arg(%+v) err(%+v) or authRep=nil", &locApi.AuthPIDsReq{Pids: pidStr, IpAddr: ip}, e)
		return false
	}
	if auth, ok := authRep.Auths[hiddenInfo.Info.Pid]; !ok || auth.Play != location.Forbidden {
		return false
	}
	return true
}

// EntrancesIsHidden .
func (s *Service) EntrancesIsHidden(ctx context.Context, req *pb.EntrancesIsHiddenRequest) (*pb.EntrancesIsHiddenReply, error) {
	res := new(pb.EntrancesIsHiddenReply)
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
		case model.EntranceTypeModule:
			res.ModuleInfos = s.hitEntranceHidden(ctx, item.Oids, otype, req.Build, req.Channel, req.Plat)
		case model.EntranceTypeDynamic:
			res.HideDynamic = s.hitDynamicHidden(ctx, req.Build, req.Channel, req.Plat)
		default:
			res.Infos = s.hitEntranceHidden(ctx, item.Oids, otype, req.Build, req.Channel, req.Plat)
		}
	}
	return res, nil
}

type SectionReq struct {
	Plat       int32
	Build      int32
	Mid        int64
	Lang       string
	Channel    string
	Ip         string
	IsUploader bool
	IsLiveHost bool
	FansCount  int64
	Buvid      string
}

// 会员购底tab个性化需求，临时用来过滤有两个会员购底tab的情况
func (s *Service) mallTabFilter(sections []*pb.Section) []*pb.Section {
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

func (s *Service) defaultMallItemFilter(in []*pb.SectionItem) []*pb.SectionItem {
	var out []*pb.SectionItem
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

// 首页模块获取
func (s *Service) HomeSections(c context.Context, req *pb.HomeSectionsRequest) (res *pb.HomeSectionsReply, err error) {
	res = new(pb.HomeSectionsReply)
	transReq := &SectionReq{
		Plat:    req.Plat,
		Build:   req.Build,
		Mid:     req.Mid,
		Lang:    req.Lang,
		Channel: req.Channel,
		Ip:      req.Ip,
		Buvid:   req.Buvid,
	}
	var bottomTabModuleId = int64(9)

	res.Sections, err = s.getSections(c, transReq, s.entryModuleCache, model.EntranceTypeHome, s.PreIconCache)
	// TODO: 假如对于数量有限制的module，可以考虑在此处做操作
	res.Sections = s.mallTabFilter(res.Sections)
	// 此处需要有骚操作-对底tab的排序做调整，将dialog opener强制插入3个或者5个时的中间位置
	for i, m := range res.Sections {
		module := m
		index := i
		if module.Id != bottomTabModuleId {
			continue
		}
		normalItem := make([]*pb.SectionItem, 0)
		var openerItem *pb.SectionItem

		for _, ele := range module.Items {
			item := ele
			if item.OpLinkType == pb.SectionItemOpLinkType_DIALOG_OPENER {
				if openerItem == nil {
					openerItem = item
				}
				continue
			}
			normalItem = append(normalItem, item)
		}
		// 有弹窗入口的时侯，普通icon数量够，就给到5个，产品说配错剁手
		if openerItem != nil && len(normalItem) >= 4 {
			res.Sections[index].Items = []*pb.SectionItem{normalItem[0], normalItem[1], openerItem, normalItem[2], normalItem[3]}
			continue
		}
		// 普通icon数量不够，只有2个，就总共给到3个，产品说配错剁手
		if openerItem != nil && len(normalItem) == 2 {
			res.Sections[index].Items = []*pb.SectionItem{normalItem[0], openerItem, normalItem[1]}
		}
		if openerItem == nil && ((req.Plat == 0 && req.Build > 6270199) || (req.Plat == 1 && req.Build > 62700099)) {
			tabIds := make([]int64, 0)
			for _, i := range res.Sections[index].Items {
				tabIds = append(tabIds, i.Id)
			}
			log.Warn("【HomeSection 告警】有下发未带大加号的底tab内容, req(%v+), ip(%s), res-ids(%+v)", req, metadata.String(c, metadata.RemoteIP), tabIds)
		}
	}
	return res, err
}

// 我的页模块获取
func (s *Service) MineSections(c context.Context, req *pb.MineSectionsRequest) (res *pb.MineSectionsReply, err error) {
	res = new(pb.MineSectionsReply)
	transReq := &SectionReq{
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
	res.Sections, err = s.getSections(c, transReq, s.mineModuleCache, model.EntranceTypeSidebar, s.IconCache)
	return res, err
}

// MineSections .
//
//nolint:gocognit,unparam
func (s *Service) getSections(c context.Context, req *SectionReq, modules map[int32][]*model.ModuleInfo, entranceType int64, iconCache map[int64]*pb.MngIcon) (sections []*pb.Section, err error) {
	sections = make([]*pb.Section, 0)
	var (
		mWhite          = make(map[int64]bool) // module白名单
		sWhite          = make(map[int64]bool) // 模块白名单
		sRed            = make(map[int64]bool) // 模块红点
		mutex           sync.Mutex
		eg              = errgroup.WithContext(c)
		sids            []int64
		moduleIDs       []int64
		ehMap           = make(map[int64]bool)        // 二级模块入口屏蔽名单
		moduleHiddenMap = make(map[int64]bool)        //一级模块入口屏蔽名单
		icMap           = make(map[int64]*pb.MngIcon) // 运营icon名单
		showAuthDef     = map[int64]*locApi.ZoneLimitAuth{}
		policiesAuth    map[int64]*locApi.ZoneLimitAuth
		dynamicConfigs  = make(map[int64]*model.DynamicConf)
		//推荐服务按顺序可出的tab
		recommendTopTabIDs []int64
		//港澳台垂类tab
		isRecommendTab bool
	)
	module, ok := modules[req.Plat]
	if !ok {
		return
	}
	// 白名单+红点接口
	for _, m := range module {
		moduleIDs = append(moduleIDs, m.ID)
		if req.Mid > 0 && m.WhiteURL != "" {
			tmpID := m.ID
			tmpURL := m.WhiteURL
			eg.Go(func(ctx context.Context) (err error) {
				ok, err := s.show.WhiteCheck(ctx, tmpURL, req.Mid, req.Buvid)
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
						ok, err := s.show.WhiteCheck(ctx, tmpwURL, req.Mid, req.Buvid)
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
						ok, err := s.show.RedDot(ctx, req.Mid, tmprURL)
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
					conf, err := s.show.FetchDynamicConf(ctx, tmpDynamicConfUrl, req.Mid, req.Buvid)
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

					bwReq := &pb2.CheckCommonBWListReq{
						Oid:    strconv.FormatInt(req.Mid, 10),
						Token:  tokens[0],
						UserIp: ip,
						LargeOid: &pb2.LargeOidContent{
							Buvid: req.Buvid,
							Mid:   req.Mid,
						},
						LargeToken: tokens[1],
					}
					if rep, err := s.CheckCommonBWList(ctx, bwReq); err != nil {
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
	if entranceType == model.EntranceTypeSidebar {
		eg.Go(func(ctx context.Context) (err error) {
			disReq := &displayApi.IconListReq{
				Mid: req.Mid,
			}
			var itemsInfo *displayApi.IconListResp
			if itemsInfo, err = s.displayGRPC.EverythingAboutIcon(ctx, disReq); err != nil {
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
		ehReq := &pb.EntrancesIsHiddenRequest{
			Build:   int64(req.Build),
			Plat:    req.Plat,
			Channel: req.Channel,
			OidItems: map[int64]*pb.OidList{
				entranceType: {
					Oids: sids,
				},
				model.EntranceTypeModule: {
					Oids: moduleIDs,
				},
			},
		}
		eh, err := s.EntrancesIsHidden(ctx, ehReq)
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
		miReq := &pb.MngIconRequest{
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
			if reply, err = s.locGRPC.ZoneLimitPolicies(ctx, zoneReq); err != nil {
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
		if m.Style == model.ModuleStyle_OP {
			isOnly = true
		}
		items := make([]*pb.SectionItem, 0)
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
					mngIcon   *pb.MngIcon
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
				if entranceType == model.EntranceTypeSidebar && sd.OpLoadCondition != "" {
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

				items = append(items, &pb.SectionItem{
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
					OpLinkType:           pb.SectionItemOpLinkType_Enum(sd.OpLinkType),
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

		sections = append(sections, &pb.Section{
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

func reorderRecommendTopTab(in []*pb.SectionItem, recommendTabIDs []int64) []*pb.SectionItem {
	var (
		itemsMap = make(map[int64]*pb.SectionItem, len(in))
		out      []*pb.SectionItem
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

// CheckLimit is
func CheckLimit(limit []*model.SideBarLimit, build int32) bool {
	if len(limit) == 0 {
		return false
	}
	for _, l := range limit {
		if model.InvalidBuild(int(build), l.Build, l.Condition) {
			return false
		}
	}
	return true
}

// CheckWhite is
func CheckWhite(mid int64, sd *model.SideBar, sWhite map[int64]bool) bool {
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

func (s *Service) InformationRegionCard(c context.Context, req *pb.NoArgRequest) (res *pb.InformationRegionCardReply, err error) {
	var resTmp []*pb.InformationRegionCard
	if resTmp, err = s.show.InformationRegionCard(c, time.Now()); err != nil {
		log.Error("%v", err)
		return
	}
	res = new(pb.InformationRegionCardReply)
	res.InformationRegionCards = resTmp
	return
}

func (s *Service) loadTabExt() {
	menuExts, err := s.show.GetTabExtFromCache(context.Background(), model.TabExtCacheKey)
	if err != nil {
		log.Error("loadTabExt() fail:%+v", err)
		return
	}

	s.tabExt.Range(func(key, value interface{}) bool {
		tabKey := key.(string)
		if _, exist := menuExts[tabKey]; !exist {
			s.tabExt.Delete(tabKey)
		} else if menuExts[tabKey] == nil {
			s.tabExt.Delete(tabKey)
		}
		return true
	})

	for key, value := range menuExts {
		s.tabExt.Store(key, value)
	}
}

//nolint:gocognit
func (s *Service) GetTabExt(c context.Context, arg *pb.GetTabExtReq) (*pb.GetTabExtRep, error) {
	var (
		nowTime = xtime.Time(time.Now().Unix())
		err     error
		tabExts = make([]*pb.TabExt, 0)
		rep     = new(pb.GetTabExtRep)
	)
	// 只支持安卓，ios粉版和国际版本
	if platValid := s.validPlat(int8(arg.Plat)); !platValid {
		return rep, nil
	}

	// 批量查询当前设备tab的点击记录
	keys := make([]string, 0)
	for _, tab := range arg.Tabs {
		if menuExt, exist := s.tabExt.Load(fmt.Sprintf(model.InitTabExtKey, tab.TabId, tab.TType)); exist {
			ext := menuExt.(*model.MenuExt)
			key := s.cacheDao.KeyMenuVer(ext.ID, arg.Buvid, ext.Ver)
			keys = append(keys, key)
		}
	}
	clickMap, _ := s.cacheDao.MenuExtVers(c, keys)

	// 并发构建tab信息
	eg := errgroup.WithContext(context.TODO())
	mu := sync.Mutex{}
	for _, tab := range arg.Tabs {
		t := tab
		eg.Go(func(ctx context.Context) (err error) {
			menuExt, exist := s.tabExt.Load(fmt.Sprintf(model.InitTabExtKey, t.TabId, t.TType))
			if !exist {
				return
			}
			// 有效性检查
			sExt := menuExt.(*model.MenuExt)
			if nowTime < sExt.Stime || nowTime > sExt.Etime {
				return
			}
			// 版本检查
			buildCheck := false
			for _, vLt := range sExt.Limit {
				if vLt.Plat == arg.Plat {
					buildCheck = true
					if model.InvalidBuild(int(arg.Build), int(vLt.Build), vLt.Conditions) {
						return
					}
				}
			}
			// 需要至少包含一个匹配的版本
			if !buildCheck {
				return
			}
			// tabExt信息构建
			ext := pb.TabExt{}
			ext.TabId = sExt.TabID
			ext.TType = sExt.Type
			ext.Attribute = sExt.Attribute
			if sExt.AttrVal(model.AttrBitImage) == model.AttrYes {
				click := 0
				if sExt.Ver != "" {
					key := s.cacheDao.KeyMenuVer(sExt.ID, arg.Buvid, sExt.Ver)
					if clickMap != nil {
						click = clickMap[key]
					}
				}
				if clickMap == nil || click == 0 { // 获取数据错误降级处理，or没有点击过图片信息都需要下发
					ext.ActiveType = sExt.ActiveType
					ext.Active = sExt.Active
					ext.ActiveIcon = sExt.ActiveIcon
					ext.InactiveType = sExt.InactiveType
					ext.InactiveIcon = sExt.InactiveIcon
					ext.Inactive = sExt.Inactive
					if sExt.Ver != "" {
						ext.Click = &pb.Click{Id: sExt.ID, Ver: sExt.Ver, Type: model.ClearVer}
					}
				}
			}
			if sExt.AttrVal(model.AttrBitColor) == model.AttrYes {
				ext.FontColor = sExt.FontColor
				ext.BarColor = sExt.BarColor
				ext.TabTopColor = sExt.TabTopColor
				ext.TabMiddleColor = sExt.TabMiddleColor
				ext.TabBottomColor = sExt.TabBottomColor
			}
			if sExt.AttrVal(model.AttrBitBgImage) == model.AttrYes {
				ext.FontColor = sExt.FontColor
				ext.BarColor = sExt.BarColor
				ext.BgImage1 = sExt.BgImage1
				ext.BgImage2 = sExt.BgImage2
			}
			// 并发串行化
			mu.Lock()
			tabExts = append(tabExts, &ext)
			mu.Unlock()
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return rep, nil
	}

	if len(tabExts) <= 0 {
		return rep, nil
	}

	rep.TabExts = tabExts
	return rep, nil
}

func (s *Service) validPlat(plat int8) bool {
	return plat == model.PlatAndroid || plat == model.PlatIPhone || plat == model.PlatIPhoneI || plat == model.PlatAndroidI
}

func (s *Service) IsUploader(c context.Context, req *pb.IsUploaderReq) (reply *pb.IsUploaderReply, err error) {
	is, err := s.isUploader(c, req.Mid)
	if err != nil {
		log.Error("s.isUploader %v", err)
	}
	reply = &pb.IsUploaderReply{
		IsUploader: is,
	}
	return
}

func (s *Service) IsUploaderWhiteCheck(c context.Context, req *model.WhiteCheckForm) (reply *model.WhiteCheckStatus, err error) {
	var is bool
	is, err = s.isUploader(c, req.Uid)
	if err != nil {
		log.Error("s.isUploader %v", err)
		return nil, err
	}
	reply = &model.WhiteCheckStatus{
		Status: 0,
	}
	if is {
		reply.Status = 1
	}
	return
}

func (s *Service) IsNotUploaderWhiteCheck(c context.Context, req *model.WhiteCheckForm) (reply *model.WhiteCheckStatus, err error) {
	var is bool
	is, err = s.isUploader(c, req.Uid)
	if err != nil {
		log.Error("s.isUploader %v", err)
		return nil, err
	}
	reply = &model.WhiteCheckStatus{
		Status: 1,
	}
	if is {
		reply.Status = 0
	}
	return
}

//nolint:unparam
func (s *Service) isUploader(c context.Context, mid int64) (bool, error) {
	eg := errgroup.WithContext(c)
	var (
		isVideoUp, isArtUp, isSongUp bool
	)
	eg.Go(func(ctx context.Context) error {
		req, e := s.vedioGRPC.IsVideoUp(ctx, &videoApi.IsVideoUpReq{
			Mid: mid,
		})
		if e != nil {
			log.Error("s.vedioGRPC.IsVideoUp(%d) error(%v)", mid, e)
			return nil
		}
		isVideoUp = req.GetIsUp()
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		req, e := s.artGRPC.IsArticleUp(ctx, &articleApi.IsArticleUpReq{
			Mid: mid,
		})
		if e != nil {
			log.Error("s.artGRPC.IsArticleUp(%d) error(%v)", mid, e)
			return nil
		}
		isArtUp = req.GetIsUp()
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		song, e := s.show.IsSongUploader(ctx, mid, 1)
		if e != nil {
			log.Error("s.show.IsSongUploader(%d) error(%v)", mid, e)
			return nil
		}
		isSongUp = song
		return nil
	})
	if e := eg.Wait(); e != nil { //错误可降级
		log.Error("isUploader wait(%v)", e)
	}
	if isVideoUp || isSongUp || isArtUp {
		return true, nil
	}
	return false, nil
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

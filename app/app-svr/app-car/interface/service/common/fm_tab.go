package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"

	fmRec "git.bilibili.co/bapis/bapis-go/ott-recommend/automotive-channel"
	ab "git.bilibili.co/bapis/bapis-go/ott/ab"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	_thumbupBiz      = "archive"
	_defaultPs       = 10
	_searchMiniTitle = "搜索结果"

	_pinTypeNormal = 0 // 固定配置卡
	_pinTypeRoll   = 1 // 内容露出卡

	_v22 = 202 // 2.2版本(去除后四位)
)

// FmShow FM首页卡片，v2.0 - v2.2 使用
func (s *Service) FmShow(c context.Context, param *fm_v2.FmShowParam) (*fm_v2.FmShowResp, error) {
	var (
		tabs   *fm_v2.TabItemsAll
		tabRes []*fm_v2.TabItem
		err    error
	)
	// 1. 获取冷启播单卡片 + 为你推荐卡片 + 算法排序卡片
	tabs, err = s.fmTabAll(c, param)
	if err != nil {
		log.Error("FmShow s.fmTabAll error:%+v", err)
		return nil, err
	}
	// 2. 卡片去重与干预
	tabs, err = s.removeExtraTabs(param, tabs)
	if err != nil {
		log.Error("FmShow s.removeExtraTabs error:%+v", err)
		return nil, err
	}
	// 3. 排序后返回 冷启播单 -> 为你推荐 -> 算法排序播单
	tabRes = make([]*fm_v2.TabItem, 0)
	if tabs.BootTab != nil {
		tabRes = append(tabRes, tabs.BootTab)
	}
	if tabs.FeedTab != nil {
		tabRes = append(tabRes, tabs.FeedTab)
	}
	if tabs.HomeTab != nil {
		tabRes = append(tabRes, tabs.HomeTab...)
	}
	// 4. 填充干预参数server_extra
	tabRes = s.fillServerExtra(tabRes)
	// 5. 合集卡片标题ab实验
	tabRes = s.seasonTabAB(c, param, tabRes)
	return &fm_v2.FmShowResp{
		TabItems: tabRes,
		PageNext: tabs.PageResp.PageNext,
		HasNext:  tabs.PageResp.HasNext,
		BootInfo: &fm_v2.BootInfo{BootFmType: param.BootFmType, BootFmId: param.BootFmId},
	}, nil
}

func (s *Service) fmTabAll(ctx context.Context, param *fm_v2.FmShowParam) (*fm_v2.TabItemsAll, error) {
	var (
		bootTab *fm_v2.TabItem
		//feedTab  *fm_v2.TabItem
		homeTab  []*fm_v2.TabItem
		pageResp fm_v2.PageResp
	)

	eg := errgroup.WithContext(ctx)
	// 冷启播单卡片
	if param.BootFmType != "" {
		eg.Go(func(ctx context.Context) error {
			var (
				localErr error
				req      *fm_v2.HandleTabItemsReq
				resp     *fm_v2.HandleTabItemsResp
			)
			req = &fm_v2.HandleTabItemsReq{
				DeviceInfo: param.DeviceInfo,
				Mid:        param.Mid,
				Buvid:      param.Buvid,
				FmType:     param.BootFmType,
				FmId:       param.BootFmId,
			}
			resp, localErr = TabItemsStrategy(ctx, req)
			if localErr != nil {
				log.Error("fmTabAll query bootTab err:%+v, param:%+v", localErr, param)
				return nil
			}
			if len(resp.TabItems) == 0 {
				log.Error("fmTabAll query bootTab empty, param:%+v", param)
				return nil
			}
			bootTab = resp.TabItems[0]
			bootTab.IsBoot = 1
			return nil
		})
	}
	// 为你推荐卡片
	eg.Go(func(ctx context.Context) error {
		var (
			localErr error
			req      *fm_v2.HandleTabItemsReq
			resp     *fm_v2.HandleTabItemsResp
		)
		req = &fm_v2.HandleTabItemsReq{
			DeviceInfo: param.DeviceInfo,
			Mid:        param.Mid,
			Buvid:      param.Buvid,
			FmType:     fm_v2.AudioFeed,
		}
		resp, localErr = TabItemsStrategy(ctx, req)
		if localErr != nil {
			log.Error("fmTabAll query feedTab err:%+v, param:%+v", localErr, param)
			return nil
		}
		if len(resp.TabItems) == 0 {
			log.Error("fmTabAll query feedTab empty, param:%+v", param)
			return nil
		}
		//feedTab = resp.TabItems[0]
		return nil
	})
	// 算法排序FmHome播单卡片
	eg.Go(func(ctx context.Context) error {
		var (
			localErr error
			req      *fm_v2.HandleTabItemsReq
			pageReq  *fm_v2.PageReq
			resp     *fm_v2.HandleTabItemsResp
		)
		if pageReq, localErr = extractPageReq(fm_v2.AudioHome, param.PageNext, "", param.Ps, param.ManualRefresh, param.DeviceInfo); localErr != nil {
			log.Error("fmTabAll extractPageReq error:%+v, param:%+v", localErr, param)
			return nil
		}
		req = &fm_v2.HandleTabItemsReq{
			DeviceInfo: param.DeviceInfo,
			Mid:        param.Mid,
			Buvid:      param.Buvid,
			FmType:     fm_v2.AudioHome,
			PageReq:    pageReq,
		}
		resp, localErr = TabItemsStrategy(ctx, req)
		if localErr != nil {
			log.Error("fmTabAll query homeTab err:%+v, param:%+v", localErr, param)
			return nil
		}
		if len(resp.TabItems) == 0 {
			log.Error("fmTabAll query homeTab empty, param:%+v", param)
			return nil
		}
		homeTab = resp.TabItems
		pageResp = resp.PageResp
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &fm_v2.TabItemsAll{
		BootTab: bootTab,
		//FeedTab:  feedTab,
		HomeTab:  homeTab,
		PageResp: pageResp,
	}, nil
}

func (s *Service) removeExtraTabs(param *fm_v2.FmShowParam, tabs *fm_v2.TabItemsAll) (*fm_v2.TabItemsAll, error) {
	var (
		pageReq *fm_v2.PageReq
		err     error
	)
	if pageReq, err = extractPageReq(fm_v2.AudioHome, param.PageNext, "", param.Ps, param.ManualRefresh, param.DeviceInfo); err != nil {
		log.Error("fmTabAll extractPageReq error:%+v, param:%+v", err, param)
		return nil, err
	}
	// 冷启播单id若与其他播单重复，则需去除
	if tabs.BootTab != nil {
		switch tabs.BootTab.FmType {
		case fm_v2.AudioFeed:
			tabs.FeedTab = nil
		case fm_v2.AudioVertical, fm_v2.AudioSeason, fm_v2.AudioSeasonUp:
			if len(tabs.HomeTab) == 0 {
				break
			}
			bootIdx := -1
			for i, v := range tabs.HomeTab {
				if v.FmType == tabs.BootTab.FmType && v.FmId == tabs.BootTab.FmId {
					bootIdx = i
					break
				}
			}
			if bootIdx != -1 {
				tabs.HomeTab = append(tabs.HomeTab[:bootIdx], tabs.HomeTab[bootIdx+1:]...)
			}
		default:
			break
		}
	}
	// pn为0，保留冷启播单 + 为你推荐 + 算法
	// pn>0，只保留算法
	if pageReq.PageNext != nil {
		if pageReq.PageNext.Pn > 0 {
			tabs.FeedTab = nil
			tabs.BootTab = nil
		}
	}
	return tabs, nil
}

// seasonTabAB 合集卡片标题替换ab实验
func (s *Service) seasonTabAB(ctx context.Context, param *fm_v2.FmShowParam, tabs []*fm_v2.TabItem) []*fm_v2.TabItem {
	if len(tabs) == 0 {
		return tabs
	}
	if param.Build/10_000 != _v22 {
		return tabs
	}

	// 1. 请求实验grpc
	if s.c.ExpIds == nil || s.c.ExpIds.Season.ExpId == 0 || s.c.ExpIds.Season.ExpGroupId == 0 {
		return tabs
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "version", strconv.Itoa(param.Build))
	req := &ab.ExpGroupMatchReq{Buvid: param.Buvid, Mid: param.Mid, ExpIds: []int64{s.c.ExpIds.Season.ExpId}}
	reply, err := s.expDao.ExpGroupMatch(ctx, req)
	if err != nil {
		log.Errorc(ctx, "seasonTabAB s.expDao.ExpGroupMatch err:%+v, req:%+v", err, req)
		return tabs
	}
	log.Warnc(ctx, "【DEBUG】seasonTabAB DoExp expCfg:%+v, param:%s, res:%s", s.c.ExpIds.Season, toJson(req), toJson(reply))
	if reply.HitLongEmptyBucket || len(reply.GroupId) == 0 || reply.GroupId[s.c.ExpIds.Season.ExpId] <= 0 {
		log.Warnc(ctx, "seasonTabAB hit no exp group, req:%+v, reply:%+v", req, reply)
		return tabs
	}

	// 2. 实验组替换标题
	if reply.GroupId[s.c.ExpIds.Season.ExpId] == s.c.ExpIds.Season.ExpGroupId {
		for _, tab := range tabs {
			if tab.FmType != fm_v2.AudioSeason {
				continue
			}
			if tab.FirstArcTitle == "" {
				log.Warnc(ctx, "seasonTabAB tab FirstArcTitle empty, tab:%+v", tab)
				continue
			}
			tab.Title = tab.FirstArcTitle
		}
	}
	return tabs
}

func (s *Service) fillServerExtra(tabs []*fm_v2.TabItem) []*fm_v2.TabItem {
	for _, v := range tabs {
		extra := fm_v2.ServerExtra{FmTitle: v.Title}
		extraStr, _ := json.Marshal(extra)
		v.ServerExtra = string(extraStr)
	}
	return tabs
}

// FmShowV2 首页改版（v2.3及以上使用，包含金刚位与精选推荐）
func (s *Service) FmShowV2(ctx context.Context, param *fm_v2.ShowV2Param) (*fm_v2.ShowV2Resp, error) {
	pageInfo, _, err := validatePageInfo(fm_v2.AudioHomeV2, param.RecPageNext)
	if err != nil {
		log.Errorc(ctx, "FmShowV2 validatePageInfo err:%+v, pageNextStr:%+v", err, param.RecPageNext)
		return nil, err
	}

	pin, rec, err := s.fmDao.FmHomeV2(ctx, param, param.PinPs, &fm_v2.PageReq{PageNext: pageInfo, PageSize: param.RecPs})
	//log.Warnc(ctx, "FmShowV2 debug s.fmDao.FmHomeV2 param:%+v, pageInfo:%+v, pin:%s, rec:%s, err:%+v", param, pageInfo, toJson(pin), toJson(rec), err)
	if err != nil {
		log.Errorc(ctx, "FmShowV2 s.fmDao.FmHomeV2 err:%+v, param:%+v", err, param)
		return nil, err
	}

	var (
		hasPin   = true
		pinItems = make([]*common.Item, 0)
		recItems []*common.Item
	)
	if pin == nil || len(pin.Cards) == 0 {
		hasPin = false
	}
	eg := errgroup.WithContext(ctx)
	if hasPin {
		eg.Go(func(ctx context.Context) error {
			var localErr error
			if pinItems, localErr = s.AICardsToCommonItems(ctx, param.DeviceInfo, param.Mid, param.Buvid, pin.Cards); localErr != nil {
				return errors.Wrap(localErr, "s.AICardsToCommonItems pin err")
			}
			return nil
		})
	}

	eg.Go(func(ctx context.Context) error {
		var localErr error
		if recItems, localErr = s.AICardsToCommonItems(ctx, param.DeviceInfo, param.Mid, param.Buvid, rec.Cards); localErr != nil {
			return errors.Wrap(localErr, "s.AICardsToCommonItems rec err")
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "FmShowV2 errgroup:%+v, param:%+v", err, param)
		return nil, err
	}

	if s.v23debug(param.Mid, param.Build) {
		card := s.mockOgvEpHorizontalCard(ctx, param.Mid, param.DeviceInfo)
		if card != nil && len(recItems) > 0 {
			t := make([]*common.Item, 0)
			t = append(t, card)
			t = append(t, recItems...)
			recItems = t
		}
		for _, v := range recItems {
			if v == nil {
				continue
			}
			v.Label = &common.Badge{
				Text:             "有更新",
				TextColorDay:     "#FF6699",
				TextColorNight:   "#FF6699",
				BorderColorDay:   "#FF6699",
				BorderColorNight: "#FF6699",
				BgStyle:          model.BgStyleFill,
			}
		}
	}

	pageInfo.Pn = pageInfo.Pn + 1
	return &fm_v2.ShowV2Resp{
		PinItems:    pinItems,
		RecItems:    recItems,
		RecPageNext: pageInfo,
		RecHasNext:  rec.HasMore,
		HasPin:      hasPin && s.GetHasPin(param),
		PinMore:     hasPin && s.GetPinMore(param),
	}, nil
}

// PinPage 金刚位"更多"页
func (s *Service) PinPage(ctx context.Context, param *fm_v2.PinPageParam) (*fm_v2.PinPageResp, error) {
	pageInfo, _, err := validatePageInfo(fm_v2.AudioHomeV2, param.PinPageNext)
	if err != nil {
		log.Errorc(ctx, "PinPage validatePageInfo err:%+v, pageNextStr:%+v", err, param.PinPageNext)
		return nil, err
	}

	pin, err := s.fmDao.FmPinPage(ctx, param, &fm_v2.PageReq{PageNext: pageInfo, PageSize: param.PinPs})
	//log.Warnc(ctx, "PinPage debug s.fmDao.FmHomeV2 param:%+v, pageInfo:%+v, pin:%s, err:%+v", param, pageInfo, toJson(pin), err)
	if err != nil {
		log.Errorc(ctx, "PinPage s.fmDao.FmPinPage err:%+v, param:%+v", err, param)
		return nil, err
	}

	items, err := s.AICardsToCommonItems(ctx, param.DeviceInfo, param.Mid, param.Buvid, pin.Cards)
	if err != nil {
		log.Errorc(ctx, "PinPage s.AICardsToCommonItems err:%+v, param:%+v", err, param)
		return nil, err
	}

	pageInfo.Pn = pageInfo.Pn + 1
	return &fm_v2.PinPageResp{
		PinItems:    items,
		TopText:     s.GetTopText(),
		PinPageNext: pageInfo,
		PinHasNext:  pin.HasMore,
	}, nil
}

// AICardsToCommonItems 算法返回的卡片，转为Item返回值
func (s *Service) AICardsToCommonItems(ctx context.Context, dev model.DeviceInfo, mid int64, buvid string, cards []*fmRec.Card) ([]*common.Item, error) {
	var (
		aiCards  *fm_v2.AICardIds // 算法返回的物料ID
		itemsRaw []*common.Item   // 乱序卡片item
		itemMap  map[string]*common.Item
		itemsRes []*common.Item
		err      error
	)
	// 拼装公共方法请求，生成卡片
	aiCards = s.separateAICards(ctx, cards)
	// 社区风险内容过滤
	aiCards.Aids = s.SixLimitFilter(ctx, aiCards.Aids)
	itemsRaw, err = s.commonItems(ctx, dev, &commonItemsReq{
		Mid:               mid,
		Buvid:             buvid,
		Aids:              aiCards.Aids,
		FmSerialIds:       aiCards.FmSerialIds,
		FmChannelIds:      aiCards.FmChannelIds,
		FmChannelIdAidMap: aiCards.FmChannelShow,
	})
	if err != nil {
		return nil, err
	}
	s.fillCommonItemsRaw(ctx, aiCards, itemsRaw)

	// 卡片转为map，按 aiCards.Order 重新排序
	itemMap = make(map[string]*common.Item)
	for _, v := range itemsRaw {
		itemMap[aiCardKey(v.ItemType, v.ItemId, v.Oid)] = v
	}
	itemsRes = make([]*common.Item, 0)
	for _, key := range aiCards.Order {
		if _, ok := itemMap[key]; !ok {
			log.Errorc(ctx, "AICardsToCommonItems card not exist, key:%s", key)
			continue
		}
		itemsRes = append(itemsRes, itemMap[key])
	}
	// 补丁：干预内容透出卡
	s.modifyShowOutCard(ctx, itemsRes)
	return itemsRes, nil
}

func (s *Service) separateAICards(ctx context.Context, cards []*fmRec.Card) *fm_v2.AICardIds {
	var (
		resp           = &fm_v2.AICardIds{CardMap: make(map[string]*fm_v2.AICard)}
		showOutChanIds = make([]int64, 0)
	)
	for i, card := range cards {
		switch card.CardType {
		case fmRec.CardType_CARD_TYPE_SERIAL:
			if card.GetSerial().Id > 0 {
				key := aiCardKey(common.ItemTypeFmSerial, card.GetSerial().Id, 0)
				resp.FmSerialIds = append(resp.FmSerialIds, card.GetSerial().Id)
				resp.CardMap[key] = &fm_v2.AICard{Index: i, Card: card}
				resp.Order = append(resp.Order, key)
			}
		case fmRec.CardType_CARD_TYPE_CHANNEL:
			if card.GetChannel().ChannelId > 0 {
				key := aiCardKey(common.ItemTypeFmChannel, card.GetChannel().ChannelId, 0)
				resp.FmChannelIds = append(resp.FmChannelIds, card.GetChannel().ChannelId)
				resp.CardMap[key] = &fm_v2.AICard{Index: i, Card: card}
				resp.Order = append(resp.Order, key)
				if card.ShowContent {
					showOutChanIds = append(showOutChanIds, card.GetChannel().ChannelId)
				}
			}
		case fmRec.CardType_CARD_TYPE_UGC:
			if card.GetUgc().Avid > 0 {
				key := aiCardKey(common.ItemTypeUGCSingle, 0, card.GetUgc().Avid)
				resp.Aids = append(resp.Aids, card.GetUgc().Avid)
				resp.CardMap[key] = &fm_v2.AICard{Index: i, Card: card}
				resp.Order = append(resp.Order, key)
			}
		case fmRec.CardType_CARD_TYPE_UGC_MULTI_PART:
			if card.GetUgcMultiPart().Avid > 0 {
				key := aiCardKey(common.ItemTypeUGCMulti, 0, card.GetUgcMultiPart().Avid)
				resp.Aids = append(resp.Aids, card.GetUgcMultiPart().Avid)
				resp.CardMap[key] = &fm_v2.AICard{Index: i, Card: card}
				resp.Order = append(resp.Order, key)
			}
		case fmRec.CardType_CARD_TYPE_FOR_YOU_FEEDS:
			continue // 为你推荐 暂不处理
		default:
			log.Error("separateAICards unknown cardType:%d", card.CardType)
			continue
		}
	}
	// 如果是内容透出卡，需填充透出的稿件id
	resp.FmChannelShow = s.getChannelBootAid(ctx, showOutChanIds)
	return resp
}

// getChannelBootAid 频道填充透出的稿件id（如果存在）
func (s *Service) getChannelBootAid(ctx context.Context, chanIds []int64) map[int64]int64 {
	if len(chanIds) == 0 {
		return nil
	}
	infoAI, err := s.fmDao.FmChannelInfoAI(ctx, chanIds)
	if err != nil {
		log.Errorc(ctx, "getChannelBootAid s.fmDao.FmChannelInfoAI err:%+v, chanIds:%+v", err, chanIds)
		return nil
	}
	var fmChannelShow = make(map[int64]int64)
	for _, v := range chanIds {
		if _, ok := infoAI[v]; !ok {
			continue
		}
		if len(infoAI[v].Avids) == 0 {
			continue
		}
		fmChannelShow[v] = infoAI[v].Avids[0]
	}
	return fmChannelShow
}

// commonItems方法后置处理
func (s *Service) fillCommonItemsRaw(_ context.Context, cards *fm_v2.AICardIds, items []*common.Item) {
	if len(cards.CardMap) == 0 {
		return
	}
	// 填充server_info、卡片样式
	for _, v := range items {
		if card, ok := cards.CardMap[aiCardKey(v.ItemType, v.ItemId, v.Oid)]; ok {
			v.ServerInfo = card.ServerInfo
			if card.ShowContent {
				v.ShowType = _pinTypeRoll
			}
		}
	}
}

func aiCardKey(iType common.ItemType, itemId int64, oid int64) string {
	switch iType {
	case common.ItemTypeUGC, common.ItemTypeUGCSingle, common.ItemTypeUGCMulti:
		return fmt.Sprintf("%s:%d", iType, oid)
	case common.ItemTypeFmSerial, common.ItemTypeFmChannel:
		return fmt.Sprintf("%s:%d", iType, itemId)
	default:
		log.Error("aiCardKey unknown itemType:%s", iType)
	}
	return ""
}

func (s *Service) GetHasPin(param *fm_v2.ShowV2Param) bool {
	if s.c.PinPageCfgAll == nil || s.c.PinPageCfgAll.HasPin == nil {
		return true
	}
	for _, v := range s.c.PinPageCfgAll.HasPin.BlackChannel {
		if param.Channel == v {
			return false
		}
	}
	return true
}

func (s *Service) GetPinMore(param *fm_v2.ShowV2Param) bool {
	if s.c.PinPageCfgAll == nil || s.c.PinPageCfgAll.PinMore == nil {
		return true
	}
	for _, v := range s.c.PinPageCfgAll.PinMore.BlackChannel {
		if param.Channel == v {
			return false
		}
	}
	return true
}

func (s *Service) GetTopText() string {
	if s.c.PinPageCfgAll == nil || s.c.PinPageCfgAll.TopText == "" {
		return ""
	}
	return s.c.PinPageCfgAll.TopText
}

// modifyShowOutCard 干预内容透出卡，包括非首位不透出，非首位变更副标题：
//
//	若config配置中存在副标题，则取副标题；否则返回空
func (s *Service) modifyShowOutCard(_ context.Context, items []*common.Item) {
	if len(items) == 0 {
		return
	}

	// 1. 选出所有非首位的内容透出卡
	idxes := make([]int, 0)
	for i, v := range items {
		if v.ItemType != common.ItemTypeFmChannel {
			continue
		}
		if i != 0 && v.ShowType == _pinTypeRoll {
			idxes = append(idxes, i)
		}
	}
	if len(idxes) == 0 {
		return
	}
	// 2. 构建频道卡map
	channelMap := make(map[int64]*conf.FmTabConfig)
	for _, cfg := range s.c.Custom.FmTabConfigs {
		if cfg.FmType != string(fm_v2.AudioVertical) {
			continue
		}
		channelMap[cfg.FmId] = cfg
	}
	// 3. 干预非首位透出卡
	for _, i := range idxes {
		items[i].ShowType = _pinTypeNormal
		if channelMap[items[i].ItemId] == nil {
			items[i].SubTitle = ""
			continue
		}
		items[i].SubTitle = channelMap[items[i].ItemId].SubTitle
	}
}

func toJson(obj interface{}) string {
	bytes, _ := json.Marshal(obj)
	return string(bytes)
}

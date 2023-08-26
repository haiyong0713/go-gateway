package common

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	serialApi "git.bilibili.co/bapis/bapis-go/serial/service"
)

const (
	_tabTypeAll        = "all"
	_tabTypeSerial     = "serial"
	_tabTypeAllName    = "全部"
	_tabTypeSerialName = "剧集"
	_serialHisMaxPs    = 1000
	_serialHisGap      = 60
)

func (s *Service) ViewHistoryTab(ctx context.Context, req *commonmdl.HistoryTabReq, mid int64, buvid string) (*commonmdl.HistoryTabResp, error) {
	var (
		mainItems, serialItems       []*commonmdl.HisItem
		mainPageNext, serialPageNext *commonmdl.HistoryTabPageNext
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var mainErr error
		mainItems, mainPageNext, mainErr = s.viewHistoryTabAll(ctx, &commonmdl.HistoryTabMoreReq{
			DeviceInfo: req.DeviceInfo,
			Ps:         req.Ps,
			TabType:    _tabTypeAll,
		}, mid, buvid)
		if mainErr != nil {
			log.Errorc(ctx, "ViewHistoryTab tab all mid=%d buvid=%s error=%+v", mid, buvid, mainErr)
		}
		return nil
	})
	if req.IsWeb {
		eg.Go(func(ctx context.Context) error {
			var serialErr error
			serialItems, serialPageNext, serialErr = s.viewHistoryTabSerial(ctx, &commonmdl.HistoryTabMoreReq{
				DeviceInfo: req.DeviceInfo,
				Ps:         req.Ps,
				TabType:    _tabTypeSerial,
			}, mid, buvid)
			if serialErr != nil {
				log.Errorc(ctx, "ViewHistoryTab tab serial mid=%d buvid=%s error=%+v", mid, buvid, serialErr)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "ViewHistoryTab eg.Wait mid=%d buvid=%s req=%+v error=%+v", mid, buvid, req, err)
		return nil, err
	}
	serialSubTab := &commonmdl.HistoryTabSubTab{
		TabType:  _tabTypeSerial,
		TabName:  _tabTypeSerialName,
		ShowMore: len(serialItems) > 0,
		PageNext: serialPageNext,
		Items:    serialItems,
	}
	allSubTab := &commonmdl.HistoryTabSubTab{
		TabType:  _tabTypeAll,
		TabName:  _tabTypeAllName,
		ShowMore: len(mainItems) > 0,
		PageNext: mainPageNext,
		Items:    mainItems,
	}
	subTabs := func() []*commonmdl.HistoryTabSubTab {
		if req.IsWeb {
			return []*commonmdl.HistoryTabSubTab{serialSubTab, allSubTab}
		}
		return []*commonmdl.HistoryTabSubTab{allSubTab}
	}()
	return &commonmdl.HistoryTabResp{SubTab: subTabs}, nil
}

func (s *Service) ViewHistoryTabMore(ctx context.Context, req *commonmdl.HistoryTabMoreReq, mid int64, buvid string) (*commonmdl.HistoryTabMoreResp, error) {
	items, pageNextData, err := func() ([]*commonmdl.HisItem, *commonmdl.HistoryTabPageNext, error) {
		if req.TabType == _tabTypeAll {
			return s.viewHistoryTabAll(ctx, req, mid, buvid)
		}
		return s.viewHistoryTabSerial(ctx, req, mid, buvid)
	}()
	if err != nil {
		log.Errorc(ctx, "ViewHistoryTabMore mid=%d buvid=%s req=%+v error=%+v", mid, buvid, req, err)
		return nil, err
	}
	return &commonmdl.HistoryTabMoreResp{
		PageNext: pageNextData,
		Items:    items,
		ShowMore: len(items) > 0,
	}, nil
}

func (s *Service) viewHistoryTabAll(ctx context.Context, req *commonmdl.HistoryTabMoreReq, mid int64, buvid string) ([]*commonmdl.HisItem, *commonmdl.HistoryTabPageNext, error) { //nolint:gocognit
	var (
		mainHis   []*hisApi.ModelResource
		serialHis []*serialApi.SerialHistory
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var mainHisErr error
		mainHis, mainHisErr = s.historyDao.HistoryCursorV2(ctx, mid, req.Max, req.ViewAt, 2*int(req.Ps), "", buvid, []string{_historyBusinessUGC, _historyBusinessOGV})
		if mainHisErr != nil {
			log.Errorc(ctx, "viewHistoryTabAll mid=%d buvid=%s req=%+v mainHisErr=%+v", mid, buvid, req, mainHisErr)
			return mainHisErr
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var serialErr error
		serialHis, serialErr = s.serialDao.SerialHistory(ctx, mid, req.SerialID, req.SerialIDType, _serialHisMaxPs, buvid)
		if serialErr != nil {
			log.Errorc(ctx, "viewHistoryTabAll serialDao.SerialHistory mid=%d buvid=%s ps=%d error=%+v", mid, buvid, _serialHisMaxPs, serialErr)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "viewHistoryTabAll mid=%d buvid=%s eg.Wait error=%+v", mid, buvid, err)
		return nil, nil, err
	}
	var (
		mainHistoryTmpList []*hisApi.ModelResource
		aidm               = make(map[int64][]int64)
		epidm              = make(map[int32]struct{})
		serialm            = make(map[int64]*serialApi.SerialHistory)
		serialHisMap       = make(map[int64][]*serialApi.SerialHistory)
	)
	for _, serialItem := range serialHis {
		if serialItem == nil || serialItem.EpisodeType != serialApi.EpisodeType_EpisodeTypeUGC {
			continue
		}
		serialHisMap[serialItem.Episode] = append(serialHisMap[serialItem.Episode], serialItem)
	}
	for _, historyTmp := range mainHis {
		if historyTmp == nil {
			continue
		}
		switch historyTmp.Business {
		case _historyBusinessUGC:
			aidm[historyTmp.Oid] = append(aidm[historyTmp.Oid], historyTmp.Cid)
			mainHistoryTmpList = append(mainHistoryTmpList, historyTmp)
			if serialHisHit, ok := serialHisMap[historyTmp.Oid]; ok {
				for _, v := range serialHisHit {
					if v == nil {
						continue
					}
					if math.Abs(float64(v.Ctime-historyTmp.Unix)) < _serialHisGap {
						serialm[v.Episode] = v
						break
					}
				}
			}
		case _historyBusinessOGV:
			epidm[int32(historyTmp.Epid)] = struct{}{}
			mainHistoryTmpList = append(mainHistoryTmpList, historyTmp)
		}
		if int64(len(mainHistoryTmpList)) > 2*req.Ps { //2倍ps的id去获取数据
			break
		}
	}
	if len(mainHistoryTmpList) == 0 {
		log.Warnc(ctx, "viewHistoryTabAll mid=%d buvid=%s req=%+v nil", mid, buvid, req)
		return nil, nil, nil
	}
	// 获取物料
	materialParams := new(commonmdl.Params)
	if len(aidm) > 0 {
		materialParams.ArchivePlusReq = new(commonmdl.ArchivePlusReq)
		for aid, cids := range aidm {
			var playAv = &archivegrpc.PlayAv{Aid: aid}
			for _, cid := range cids {
				playAv.PlayVideos = append(playAv.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
			}
			materialParams.ArchivePlusReq.PlayAvs = append(materialParams.ArchivePlusReq.PlayAvs, playAv)
		}
	}
	if len(epidm) > 0 {
		var epids []int32
		for epid := range epidm {
			epids = append(epids, epid)
		}
		materialParams.EpisodeReq = new(commonmdl.EpisodeReq)
		materialParams.EpisodeReq.Epids = epids
	}
	if len(serialm) > 0 {
		materialParams.SerialInfosReq = new(commonmdl.SerialInfosReq)
		materialParams.ChannelInfosReq = new(commonmdl.ChannelInfosReq)
		for _, serialTmp := range serialm {
			if serialTmp == nil || serialTmp.SerialId <= 0 {
				continue
			}
			itemType, ok := commonmdl.SerialBusinessTypeToItemType[serialTmp.BusinessSerialType]
			if !ok {
				continue
			}
			switch itemType {
			case commonmdl.ItemTypeVideoSerial:
				materialParams.SerialInfosReq.VideoIds = append(materialParams.SerialInfosReq.VideoIds, serialTmp.SerialId)
			case commonmdl.ItemTypeVideoChannel:
				materialParams.ChannelInfosReq.Video = append(materialParams.ChannelInfosReq.Video, serialTmp.SerialId)
			case commonmdl.ItemTypeFmSerial:
				materialParams.SerialInfosReq.FmCommonIds = append(materialParams.SerialInfosReq.FmCommonIds, serialTmp.SerialId)
			case commonmdl.ItemTypeFmChannel:
				materialParams.ChannelInfosReq.Fm = append(materialParams.ChannelInfosReq.Fm, serialTmp.SerialId)
			default:
				continue
			}
			if materialParams.ArchiveReq == nil {
				materialParams.ArchiveReq = new(commonmdl.ArchiveReq)
			}
			materialParams.ArchiveReq.PlayAvs = append(materialParams.ArchiveReq.PlayAvs, &archivegrpc.PlayAv{Aid: serialTmp.Episode})
		}
	}
	carContext, err := s.material(ctx, materialParams, req.DeviceInfo)
	if err != nil {
		b, _ := json.Marshal(materialParams)
		log.Errorc(ctx, "viewHistoryTabAll material(%+v) error(%+v)", string(b), err)
		return nil, nil, err
	}
	// 聚合卡片
	var resItems []*commonmdl.HisItem
	for _, tmp := range mainHistoryTmpList {
		switch tmp.Business {
		case _historyBusinessUGC:
			carContext.OriginData = &commonmdl.OriginData{
				MaterialType: commonmdl.MaterialTypeUGCPlus,
				Oid:          tmp.Oid,
				Cid:          tmp.Cid,
			}
			serialHisItem, ok := serialm[tmp.Oid]
			if ok && serialHisItem != nil {
				itemType, typeOK := commonmdl.SerialBusinessTypeToItemType[serialHisItem.BusinessSerialType]
				if !typeOK {
					continue
				}
				carContext.OriginData = &commonmdl.OriginData{
					Oid: serialHisItem.SerialId,
					Cid: serialHisItem.Episode,
				}
				switch itemType {
				case commonmdl.ItemTypeVideoSerial:
					carContext.OriginData.MaterialType = commonmdl.MaterialTypeVideoSerial
				case commonmdl.ItemTypeVideoChannel:
					carContext.OriginData.MaterialType = commonmdl.MaterialTypeVideoChannel
				case commonmdl.ItemTypeFmSerial:
					carContext.OriginData.MaterialType = commonmdl.MaterialTypeFmSerial
				case commonmdl.ItemTypeFmChannel:
					carContext.OriginData.MaterialType = commonmdl.MaterialTypeFmChannel
				default:
					continue
				}
			}
		case _historyBusinessOGV:
			carContext.OriginData = &commonmdl.OriginData{
				MaterialType: commonmdl.MaterialTypeOGVEP,
				Oid:          tmp.Epid,
			}
		default:
			continue
		}
		item := s.formItem(carContext, req.DeviceInfo)
		if item == nil {
			log.Warnc(ctx, "viewHistoryTabAll mainHistoryTmpList item nil carContext=%+v", carContext)
			continue
		}
		if tmp.Business == _historyBusinessOGV {
			epData, ok := carContext.EpisodeResp[int32(tmp.Epid)]
			if ok && epData != nil {
				item.SubTitle = epShowTitle(epData)
				item.Title = item.Title + " " + item.SubTitle
			}
		}
		if tmp.Business == _historyBusinessUGC {
			if item.ItemType == commonmdl.ItemTypeVideoChannel || item.ItemType == commonmdl.ItemTypeFmChannel {
				arc := carContext.ArchiveResp[carContext.OriginData.Cid]
				if arc.GetArc().GetPic() != "" {
					item.Cover = arc.Arc.GetPic()
				}
				item.Duration = arc.GetArc().GetDuration()
			}
		}
		item.ItemHistory = &commonmdl.ItemHistory{
			Business: tmp.Business,
			ViewAt:   tmp.Unix,
			Progress: tmp.Pro,
			Max:      tmp.Kid,
		}
		hisItem := &commonmdl.HisItem{
			Item:     item,
			PlayType: commonmdl.FromHisSourceToPlayType(tmp.Source),
		}
		resItems = append(resItems, hisItem)
		if int64(len(resItems)) >= req.Ps {
			break
		}
	}
	return resItems, pageNext(resItems), nil
}

func (s *Service) viewHistoryTabSerial(ctx context.Context, req *commonmdl.HistoryTabMoreReq, mid int64, buvid string) ([]*commonmdl.HisItem, *commonmdl.HistoryTabPageNext, error) { //nolint:gocognit
	var (
		mainHis    []*hisApi.ModelResource
		serialHis  []*serialApi.SerialHistory
		serialAids []int64
	)
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var mainErr error
		mainHis, mainErr = s.historyDao.HistoryCursorV2(ctx, mid, req.Max, req.ViewAt, 2*int(req.Ps), "", buvid, []string{_historyBusinessOGV})
		if mainErr != nil {
			log.Errorc(ctx, "viewHistoryTab historyDao.HistoryCursorV2 mid=%d buvid=%s ps=%d error=%+v", mid, buvid, req.Ps, mainErr)
		}
		return mainErr
	})
	eg.Go(func(ctx context.Context) error {
		var serialErr error
		serialHis, serialErr = s.serialDao.SerialHistory(ctx, mid, req.SerialID, req.SerialIDType, _serialHisMaxPs, buvid)
		if serialErr != nil {
			log.Errorc(ctx, "viewHistoryTab serialDao.SerialHistory mid=%d ps=%d error=%+v", mid, _serialHisMaxPs, serialErr)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "viewHistoryTab mid=%d buvid=%s req=%+v error=%+v", mid, buvid, req, err)
		return nil, nil, err
	}
	// 剧集tab 合集历史用主站历史时间戳截断
	serialHis, serialAids = func() ([]*serialApi.SerialHistory, []int64) {
		var (
			tmpSerialHis   []*serialApi.SerialHistory
			tmpAids        []int64
			cutSerialIndex int
		)
		endViewAt := func() int64 {
			if len(mainHis) > 0 && mainHis[len(mainHis)-1] != nil {
				return mainHis[len(mainHis)-1].Unix
			}
			return 0
		}()
		for i, v := range serialHis {
			if req.ViewAt > 0 {
				if v.Ctime < req.ViewAt && v.Ctime >= endViewAt {
					tmpSerialHis = append(tmpSerialHis, v)
					tmpAids = append(tmpAids, v.Episode)
					cutSerialIndex = i
				}
			} else {
				if v.Ctime >= endViewAt {
					tmpSerialHis = append(tmpSerialHis, v)
					tmpAids = append(tmpAids, v.Episode)
					cutSerialIndex = i
				}
			}
			if int64(len(tmpSerialHis)) > req.Ps {
				break
			}
		}
		if req.TabType == _tabTypeSerial && req.ViewAt == 0 && int64(len(tmpSerialHis)) < req.Ps { // 第一页数量不够，多获取数据
			preSerialLen := len(tmpSerialHis)
			for i, v := range serialHis {
				if i > cutSerialIndex || preSerialLen == 0 {
					tmpSerialHis = append(tmpSerialHis, v)
					tmpAids = append(tmpAids, v.Episode)
					if int64(len(tmpSerialHis)) > req.Ps {
						break
					}
				}
			}
		}
		return tmpSerialHis, tmpAids
	}()
	var (
		mainHistoryTmpList []*hisApi.ModelResource
		epidm              = make(map[int32]struct{})
	)
	for _, historyTmp := range mainHis {
		if historyTmp == nil {
			continue
		}
		switch historyTmp.Business {
		case _historyBusinessOGV:
			epidm[int32(historyTmp.Epid)] = struct{}{}
			mainHistoryTmpList = append(mainHistoryTmpList, historyTmp)
		}
	}
	// 获取物料
	materialParams := new(commonmdl.Params)
	if len(epidm) > 0 {
		var epids []int32
		for epid := range epidm {
			epids = append(epids, epid)
		}
		materialParams.EpisodeReq = new(commonmdl.EpisodeReq)
		materialParams.EpisodeReq.Epids = epids
	}
	if len(serialHis) > 0 {
		materialParams.SerialInfosReq = new(commonmdl.SerialInfosReq)
		materialParams.ChannelInfosReq = new(commonmdl.ChannelInfosReq)
		for _, serialTmp := range serialHis {
			if serialTmp == nil || serialTmp.SerialId <= 0 {
				continue
			}
			itemType, ok := commonmdl.SerialBusinessTypeToItemType[serialTmp.BusinessSerialType]
			if !ok {
				continue
			}
			switch itemType {
			case commonmdl.ItemTypeVideoSerial:
				materialParams.SerialInfosReq.VideoIds = append(materialParams.SerialInfosReq.VideoIds, serialTmp.SerialId)
			case commonmdl.ItemTypeVideoChannel:
				materialParams.ChannelInfosReq.Video = append(materialParams.ChannelInfosReq.Video, serialTmp.SerialId)
			case commonmdl.ItemTypeFmSerial:
				materialParams.SerialInfosReq.FmCommonIds = append(materialParams.SerialInfosReq.FmCommonIds, serialTmp.SerialId)
			case commonmdl.ItemTypeFmChannel:
				materialParams.ChannelInfosReq.Fm = append(materialParams.ChannelInfosReq.Fm, serialTmp.SerialId)
			default:
				continue
			}
			if materialParams.ArchiveReq == nil {
				materialParams.ArchiveReq = new(commonmdl.ArchiveReq)
			}
			materialParams.ArchiveReq.PlayAvs = append(materialParams.ArchiveReq.PlayAvs, &archivegrpc.PlayAv{Aid: serialTmp.Episode})
		}
	}
	eg2 := errgroup.WithContext(ctx)
	var (
		carContext     *commonmdl.CarContext
		moreMainHisMap map[int64]*hisApi.ModelHistory
	)
	eg2.Go(func(ctx context.Context) error {
		var carErr error
		carContext, carErr = s.material(ctx, materialParams, req.DeviceInfo)
		return carErr
	})
	if req.TabType == _tabTypeSerial && len(serialAids) > 0 {
		eg2.Go(func(ctx context.Context) error {
			var moreHisErr error
			moreMainHisMap, _, moreHisErr = s.historyDao.BatchProgress(ctx, mid, buvid, serialAids, nil)
			if moreHisErr != nil {
				log.Errorc(ctx, "viewHistoryTab BatchProgress mid=%d buvid=%+v aids=%+v error=%+v", mid, buvid, serialAids, moreHisErr)
			}
			return nil
		})
	}
	if err := eg2.Wait(); err != nil {
		b, _ := json.Marshal(materialParams)
		log.Errorc(ctx, "viewHistoryTab material(%+v) error(%+v)", string(b), err)
		return nil, nil, err
	}
	// 聚合卡片
	var resItems []*commonmdl.HisItem
	for _, tmp := range mainHistoryTmpList {
		switch tmp.Business {
		case _historyBusinessOGV:
			carContext.OriginData = &commonmdl.OriginData{
				MaterialType: commonmdl.MaterialTypeOGVEP,
				Oid:          tmp.Epid,
			}
		default:
			continue
		}
		item := s.formItem(carContext, req.DeviceInfo)
		if item == nil {
			log.Warnc(ctx, "ViewHistoryTab mainHistoryTmpList item nil carContext=%+v", carContext)
			continue
		}
		if tmp.Business == _historyBusinessOGV {
			epData, ok := carContext.EpisodeResp[int32(tmp.Epid)]
			if ok && epData != nil {
				item.SubTitle = epShowTitle(epData)
			}
		}
		item.ItemHistory = &commonmdl.ItemHistory{
			Business: tmp.Business,
			ViewAt:   tmp.Unix,
			Progress: tmp.Pro,
			Max:      tmp.Kid,
		}
		hisItem := &commonmdl.HisItem{
			Item:     item,
			PlayType: commonmdl.FromHisSourceToPlayType(tmp.Source),
		}
		resItems = append(resItems, hisItem)
	}
	for _, serialItem := range serialHis {
		if serialItem == nil || serialItem.EpisodeType != serialApi.EpisodeType_EpisodeTypeUGC { // 合集只处理ugc内容
			continue
		}
		itemType, ok := commonmdl.SerialBusinessTypeToItemType[serialItem.BusinessSerialType]
		if !ok {
			continue
		}
		// 找主站历史中匹配的单ep历史
		hisItemTmp, ok := moreMainHisMap[serialItem.Episode]
		if !ok || hisItemTmp == nil {
			continue
		}
		carContext.OriginData = &commonmdl.OriginData{
			Oid: serialItem.SerialId,
			Cid: serialItem.Episode,
		}
		switch itemType {
		case commonmdl.ItemTypeVideoSerial:
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeVideoSerial
		case commonmdl.ItemTypeVideoChannel:
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeVideoChannel
		case commonmdl.ItemTypeFmSerial:
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeFmSerial
		case commonmdl.ItemTypeFmChannel:
			carContext.OriginData.MaterialType = commonmdl.MaterialTypeFmChannel
		default:
			continue
		}
		item := s.formItem(carContext, req.DeviceInfo)
		if item == nil {
			log.Warnc(ctx, "ViewHistoryTab serialHis item nil carContext=%+v", carContext)
			continue
		}
		if itemType == commonmdl.ItemTypeVideoChannel || itemType == commonmdl.ItemTypeFmChannel {
			item.SubTitle = item.Title
			arc := carContext.ArchiveResp[carContext.OriginData.Cid]
			if arc.GetArc().GetPic() != "" {
				item.Cover = arc.Arc.GetPic()
			}
			item.Title = arc.GetArc().GetTitle()
			item.Duration = arc.GetArc().GetDuration()
		}
		if itemType == commonmdl.ItemTypeVideoSerial || itemType == commonmdl.ItemTypeFmSerial {
			item.SubTitle = func() string {
				var data map[int64]*commonmdl.SerialInfo
				if carContext.OriginData.MaterialType == commonmdl.MaterialTypeVideoSerial {
					data = carContext.SerialInfosResp.Video
				} else {
					data = carContext.SerialInfosResp.FmCommon
				}
				serial := data[carContext.OriginData.Oid]
				if serial == nil {
					return ""
				}
				return serial.Title
			}()
		}
		item.ItemHistory = &commonmdl.ItemHistory{
			ViewAt:   serialItem.Ctime,
			Progress: serialItem.ViewAt,
		}
		hisItem := &commonmdl.HisItem{
			Item:     item,
			PlayType: commonmdl.FromHisSourceToPlayType(hisItemTmp.Source),
		}
		resItems = append(resItems, hisItem)
	}
	// 重排
	sort.Slice(resItems, func(i, j int) bool {
		return resItems[i].ItemHistory.ViewAt > resItems[j].ItemHistory.ViewAt
	})
	// 截断
	if int64(len(resItems)) > req.Ps {
		resItems = resItems[:req.Ps]
	}
	return resItems, pageNext(resItems), nil
}

func epShowTitle(epData *episodegrpc.EpisodeCardsProto) string {
	firstIndex := strings.Index(epData.ShowTitle, "第")
	if firstIndex != -1 {
		spaceIndex := strings.Index(epData.ShowTitle, " ")
		if spaceIndex != -1 {
			return epData.ShowTitle[:spaceIndex]
		}
	}
	return epData.ShowTitle
}

func pageNext(data []*commonmdl.HisItem) *commonmdl.HistoryTabPageNext {
	if len(data) == 0 {
		return nil
	}
	pageNextData := new(commonmdl.HistoryTabPageNext)
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] != nil && data[i].Item != nil {
			// 主历史游标
			if pageNextData.ViewAt == 0 {
				pageNextData.ViewAt = data[i].ViewAt
				pageNextData.Max = data[i].Oid
			}
			// 剧集历史游标
			if data[i].Item.ItemId > 0 && pageNextData.SerialID == 0 {
				pageNextData.SerialID = data[i].Item.ItemId
				pageNextData.SerialIDType = commonmdl.ItemTypeToSerialBusinessType[data[i].Item.ItemType]
			}
		}
	}
	if pageNextData.ViewAt == 0 && pageNextData.SerialID == 0 {
		return nil
	}
	return pageNextData
}

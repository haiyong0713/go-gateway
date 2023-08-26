package common

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	commonmdl "go-gateway/app/app-svr/app-car/interface/model/common"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

const (
	_videoHistoryPs     = 20
	_videoHistoryMax    = 1200
	_historyBusinessUGC = "archive"
	_historyBusinessOGV = "pgc"
	_audioLimitDay      = 86400 * 7
)

func (s *Service) ViewHistory(c context.Context, req *commonmdl.ViewContinueReq, mid int64, buvid string) (resp *commonmdl.ViewContinueResp, err error) { // nolint:gocognit
	// 分页前置逻辑
	var pageNext *commonmdl.ViewContinuePageNext
	if req.PageNext != "" {
		if errTmp := json.Unmarshal([]byte(req.PageNext), &pageNext); errTmp != nil {
			log.Error("ViewHistory json.Unmarshal() error(%v)", errTmp)
		}
	}
	businesses := []string{_historyBusinessUGC, _historyBusinessOGV}
	var (
		viewAt   int64
		business string
	)
	if pageNext != nil {
		viewAt = pageNext.ViewAt
		business = pageNext.Business
	}
	// 获取服务端数据
	historyTmps, err := s.historyDao.HistoryCursor(c, mid, viewAt, viewAt, int32(_videoHistoryMax), business, buvid, businesses)
	if err != nil {
		log.Error("ViewHistory HistoryCursor(%v, %v, %v, %v, %v, %v, %v) error(%v)", mid, viewAt, viewAt, int32(_videoHistoryMax), business, buvid, businesses, err)
		return
	}
	// 视频音频分离
	var (
		videoHistorys, audioHistorys []*hisApi.ModelResource
		firstAudio                   *hisApi.ModelResource
	)
	for _, historyTmp := range historyTmps {
		if historyTmp == nil {
			continue
		}
		if historyTmp.Source == "car-audio" {
			if len(audioHistorys) == 0 { // 取第一个音频
				firstAudio = new(hisApi.ModelResource)
				*firstAudio = *historyTmp
			}
			audioHistorys = append(audioHistorys, historyTmp)
		} else {
			videoHistorys = append(videoHistorys, historyTmp)
		}
	}
	// 物料ID分离
	var (
		aidm             = make(map[int64][]int64)
		epidm            = make(map[int32]struct{})
		historyTmpsSlice []*hisApi.ModelResource
		videoHistoryPs   = _videoHistoryPs
	)
	if pageNext != nil {
		videoHistoryPs = pageNext.Ps
	}
	if req.Ps != 0 {
		videoHistoryPs = req.Ps
	}
	for _, videoHistory := range videoHistorys {
		switch videoHistory.Business {
		case _historyBusinessUGC:
			aidm[videoHistory.Oid] = append(aidm[videoHistory.Oid], videoHistory.Cid)
			historyTmpsSlice = append(historyTmpsSlice, videoHistory)
		case _historyBusinessOGV:
			epidm[int32(videoHistory.Epid)] = struct{}{}
			historyTmpsSlice = append(historyTmpsSlice, videoHistory)
		}
		if len(historyTmpsSlice) == videoHistoryPs {
			break
		}
	}
	// 获取物料
	var materialParams = new(commonmdl.Params)
	if len(aidm) > 0 {
		materialParams.ArchiveReq = new(commonmdl.ArchiveReq)
		for aid, cids := range aidm {
			var playAv = &archivegrpc.PlayAv{Aid: aid}
			for _, cid := range cids {
				playAv.PlayVideos = append(playAv.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
			}
			materialParams.ArchiveReq.PlayAvs = append(materialParams.ArchiveReq.PlayAvs, playAv)
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
	carContext, err := s.material(c, materialParams, req.DeviceInfo)
	if err != nil {
		b, _ := json.Marshal(materialParams)
		log.Error("ViewHistory material(%+v) error(%v)", string(b), err)
		return
	}
	// 聚合卡片
	resp = new(commonmdl.ViewContinueResp)
	for _, tmp := range historyTmpsSlice {
		switch tmp.Business {
		case _historyBusinessUGC:
			carContext.OriginData = &commonmdl.OriginData{
				MaterialType: commonmdl.MaterialTypeUGC,
				Oid:          tmp.Oid,
				Cid:          tmp.Cid,
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
		if item != nil {
			item.ItemHistory = &commonmdl.ItemHistory{
				Business: tmp.Business,
				ViewAt:   tmp.Unix,
				Progress: tmp.Pro,
			}
			resp.Items = append(resp.Items, item)
			resp.PageNext = &commonmdl.ViewContinuePageNext{
				Business: item.ItemHistory.Business,
				ViewAt:   item.ItemHistory.ViewAt,
				Ps:       _videoHistoryPs,
			}
		}
	}
	// 后置逻辑
	if firstAudio != nil { // N天内 或者 在第一刷
		var (
			daylimit    = (time.Now().Unix() - firstAudio.Unix) < _audioLimitDay // N天内
			inFirstPage bool                                                     // 是否在第一刷
		)
		if req.PageNext == "" && resp.PageNext != nil { // 第一个音频 与 第一刷最后一个视频比较时间戳
			if firstAudio.Unix >= resp.PageNext.ViewAt {
				inFirstPage = true
			}
		}
		if daylimit || inFirstPage {
			var fmTitle, fmCover string
			for _, tmpConfig := range s.c.Custom.FmTabConfigs {
				if tmpConfig.FmType == "audio_history" {
					fmTitle = tmpConfig.Title
					fmCover = tmpConfig.Cover
					break
				}
			}
			if fmTitle != "" && fmCover != "" {
				resp.Fm = &commonmdl.ViewContinueFm{
					Title:  fmTitle,
					Cover:  fmCover,
					ViewAt: firstAudio.Unix,
				}
			}
		}
	}
	return
}

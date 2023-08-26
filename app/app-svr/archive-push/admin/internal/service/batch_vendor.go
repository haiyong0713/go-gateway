package service

import (
	"encoding/json"
	"fmt"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	"strconv"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	blizzardModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
)

func (s *Service) GenerateQQCMCPushings(details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory) (res interface{}, err error) {
	if len(details) == 0 || len(histories) == 0 {
		return
	}

	_res := make([]*qqModel.PushPGCAdminReq, 0)
	for _, detail := range details {
		_detail := detail
		if _detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL {
			continue
		}
		var (
			arc      *archiveGRPC.Arc
			tagNames = make([]string, 0)
		)
		if arc, err = s.dao.GetArcByAID(_detail.AID); err != nil {
			log.Error("Service: GenerateQQCMCPushings GetArcByAID %d error %v", _detail.AID, err)
			continue
		}
		if tags, _err := s.dao.GetTagsByAID(_detail.AID); _err != nil {
			log.Error("Service: GenerateQQCMCPushings GetTagsByAID %d error %v", _detail.AID, _err)
			continue
		} else if len(tags) > 0 {
			for _, tag := range tags {
				tagNames = append(tagNames, tag.Name)
			}
		}
		bvid, _ := util.AvToBv(detail.AID)
		// build cmc push model
		cmcReq := &qqModel.PushPGCAdminReq{
			SCreater:       strconv.FormatInt(arc.Author.Mid, 10),
			SCreaterHeader: arc.Author.Face,
			STitle:         arc.Title,
			SIMG:           arc.Pic,
			SDESC:          arc.Desc,
			SAuthor:        arc.Author.Name,
			IType:          "0",
			ISubType:       "0",
			SOriginID:      bvid,
			SURL:           "",
			SCreated:       arc.Ctime.Time().Format("2006-01-02 15:04:05"),
			SCreatedOther:  arc.Ctime.Time().Format("2006-01-02 15:04:05"),
			STagsOther:     strings.Join(tagNames, ","),
			SSource:        qqModel.DefaultSSource,
			SVID:           bvid,
			ITime:          strconv.FormatInt(arc.Duration, 10),
			IFrom:          "11",
			SVideoSize:     fmt.Sprintf("%d*%d", arc.Dimension.Width, arc.Dimension.Height),
		}
		_res = append(_res, cmcReq)
	}
	res = _res

	return
}

func (s *Service) TryQQCMCPushing(pushings []*qqModel.PushPGCAdminReq, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory, username string, uid int64) {
	if len(pushings) == 0 || len(details) == 0 {
		log.Warn("Service: TryQQCMCPushing no valid data to push")
		return
	}
	maxRetryCount := s.Cfg.Push.MaxRetryCount
	for _, push := range pushings {
		_push := push

		go func() {
			var detail model.ArchivePushBatchDetail
			var history model.ArchivePushBatchHistory
			for _, _detail := range details {
				currentDetail := _detail
				if currentDetail.AID == 0 && _push.SVID == currentDetail.ArchiveDetails {
					detail = *currentDetail
					for _, _history := range histories {
						if _history.BVID == detail.ArchiveDetails {
							history = *_history
							break
						}
					}
				} else if bvid, _ := util.AvToBv(currentDetail.AID); bvid == _push.SVID {
					detail = *currentDetail
					for _, _history := range histories {
						if _history.AID == detail.AID {
							history = *_history
							break
						}
					}
					break
				}
			}

			count := 0
			success := false
			var err error
			for count < maxRetryCount && !success {
				if success, err = s.DoQQCMCPushing(_push, &detail, &history); err != nil || !success {
					log.Error("Service: TryQQCMCPushing DoQQCMCPushing failed or Error %v", err)
					count++
					if xecode.EqualError(xecode.NothingFound, err) {
						count = maxRetryCount
					}
				}
				// 更新detail，新增history
				if err = s.AddBatchDetailActionLog([]*model.ArchivePushBatchHistory{&history}, model.ActionPushUp, username, uid); err != nil {
					log.Error("Service: TryQQCMCPushing AddBatchDetailActionLog Error (%v)", err)
					return
				}
				if _, err = s.UpdateBatchDetails([]*model.ArchivePushBatchDetail{&detail}); err != nil {
					log.Error("Service: TryQQCMCPushing UpdateBatchDetails Error (%v)", err)
					return
				}
				time.Sleep(5 * time.Second)
			}
			// 若达到重试上限，则记录修改为失败
			if count >= maxRetryCount {
				_history := &history
				if _history.NewPushStatus != api.ArchivePushDetailPushStatus_SUCCESS && _history.NewPushStatus != api.ArchivePushDetailPushStatus_INNER_FAIL {
					_history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
				}
				if err = s.AddBatchDetailActionLog([]*model.ArchivePushBatchHistory{_history}, model.ActionPushUp, username, uid); err != nil {
					log.Error("Service: TryQQCMCPushing AddBatchDetailActionLog Error (%v)", err)
					return
				}
				_detail := &detail
				if _detail.PushStatus != api.ArchivePushDetailPushStatus_SUCCESS && _detail.PushStatus != api.ArchivePushDetailPushStatus_INNER_FAIL {
					_detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
				}
				if _, err = s.UpdateBatchDetails([]*model.ArchivePushBatchDetail{_detail}); err != nil {
					log.Error("Service: TryQQCMCPushing UpdateBatchDetails Error (%v)", err)
					return
				}
			}
		}()
	}

	return
}

func (s *Service) DoQQCMCPushing(push *qqModel.PushPGCAdminReq, detail *model.ArchivePushBatchDetail, history *model.ArchivePushBatchHistory) (success bool, err error) {
	if push.SVID == "" || (detail.ArchiveDetails == "" && detail.AID == 0) || (history.AID == 0 && history.BVID == "") {
		log.Warn("Service: DoQQCMCPushing no valid data to push")
		return false, xecode.NothingFound
	} else if detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL || !strings.HasPrefix(detail.ArchiveDetails, "{") {
		log.Warn("Service: DoQQCMCPushing no valid data to push")
		return false, xecode.NothingFound
	}
	if history.AID == 0 && history.BVID == "" {
		return
	}
	if reply, perr := s.qqDAO.PushPGCAdmin(push); perr != nil {
		log.Error("Service: DoQQCMCPushing PushPGCAdmin(%+v) error %v", push, perr)
		return false, ecode.PushRequestError
	} else if reply == nil {
		// 推送失败
		log.Error("Service: DoQQCMCPushing.PushPGCAdmin 推送结果为空")
		detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		success = false
		err = ecode.QQCMCRequestError
	} else {
		if reply.Status == 0 && reply.Data != nil && reply.Data.DocID != "" {
			// 推送成功
			archiveDetails := &model.ArchiveMetadataAll{}
			if _err := json.Unmarshal([]byte(detail.ArchiveDetails), archiveDetails); _err != nil {
				log.Error("Service: DoQQCMCPushing PushPGCAdmin Unmarshal(%s) error %v", detail.ArchiveDetails, _err)
			}
			archiveDetails.DocID = reply.Data.DocID
			archiveDetailsBytes, _ := json.Marshal(archiveDetails)
			detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			detail.ArchiveDetails = string(archiveDetailsBytes)
			history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			history.ArchiveDetails = string(archiveDetailsBytes)
			history.Reason = reply.MSG
			return true, nil
		} else if qqModel.RegPushPGCExistingBVID.MatchString(reply.MSG) {
			// 推送稿件已存在
			archiveDetails := &model.ArchiveMetadataAll{}
			if _err := json.Unmarshal([]byte(detail.ArchiveDetails), archiveDetails); _err != nil {
				log.Error("Service: DoQQCMCPushing PushPGCAdmin Unmarshal(%s) error %v", detail.ArchiveDetails, _err)
			}
			if foundStrs := qqModel.RegPushPGCExistingBVID.FindStringSubmatch(reply.MSG); len(foundStrs) >= 2 {
				archiveDetails.DocID = foundStrs[1]
			}
			archiveDetailsBytes, _ := json.Marshal(archiveDetails)
			detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			detail.ArchiveDetails = string(archiveDetailsBytes)
			history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			history.ArchiveDetails = string(archiveDetailsBytes)
			history.Reason = reply.MSG
			return true, nil
		} else {
			// 推送失败
			log.Error("Service: DoQQCMCPushing.PushPGCAdmin failed or Error %v", perr)
			detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
			history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
			history.Reason = reply.MSG
			success = false
			if perr != nil {
				err = perr
			} else {
				err = ecode.QQCMCRequestError
			}
		}
	}
	return
}

// GenerateQQTGLPushings 生成TGL推送模型
func (s *Service) GenerateQQTGLPushings(details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory) (res interface{}, err error) {
	if len(details) == 0 || len(histories) == 0 {
		return
	}

	_res := make([]*qqModel.ContributeVideoReq, 0)
	for _, detail := range details {
		_detail := detail
		if _detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL {
			continue
		}
		var (
			arc      *archiveGRPC.Arc
			tagNames = make([]string, 0)
			openID   string
		)
		if arc, err = s.dao.GetArcByAID(_detail.AID); err != nil {
			log.Error("Service: GenerateQQTGLPushings GetArcByAID %d error %v", _detail.AID, err)
			continue
		}
		if tags, _err := s.dao.GetTagsByAID(_detail.AID); _err != nil {
			log.Error("Service: GenerateQQTGLPushings GetTagsByAID %d error %v", _detail.AID, _err)
			continue
		} else if len(tags) > 0 {
			for _, tag := range tags {
				tagNames = append(tagNames, tag.Name)
			}
		}
		if appkey, _err := s.GetOauthAppKeyByVendorID(model.DefaultVendors[1].ID); _err != nil {
			log.Error("Service: GenerateQQTGLPushings GetOauthAppKeyByVendorID %d error %v", model.DefaultVendors[1].ID, _err)
			continue
		} else {
			if openID, err = s.GetOpenIDByMID(arc.Author.Mid, appkey); err != nil {
				log.Error("Service: GenerateQQTGLPushings GetOpenIDByMID (%d, %s) error %v", arc.Author.Mid, appkey, err)
				continue
			}
		}
		bvid, _ := util.AvToBv(detail.AID)
		// build tgl push model
		pushReq := &qqModel.ContributeVideoReq{
			Title:     arc.Title,
			Summary:   arc.Desc,
			Cover:     arc.Pic,
			Author:    arc.Author.Name,
			Duration:  arc.Duration,
			OuterVID:  bvid,
			OuterUser: openID,
			ExtTags:   strings.Join(tagNames, ","),
		}
		_res = append(_res, pushReq)
	}
	res = _res

	return
}

func (s *Service) TryQQTGLPushing(pushings []*qqModel.ContributeVideoReq, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory, username string, uid int64) {
	if len(pushings) == 0 || len(details) == 0 {
		log.Warn("Service: TryQQCMCPushing no valid data to push")
		return
	}
	maxRetryCount := s.Cfg.Push.MaxRetryCount
	for _, _push := range pushings {
		push := *_push
		var detail model.ArchivePushBatchDetail
		var history model.ArchivePushBatchHistory
		for _, _detail := range details {
			currentDetail := _detail
			if currentDetail.AID == 0 && push.OuterVID == currentDetail.ArchiveDetails {
				detail = *currentDetail
				for _, _history := range histories {
					if _history.BVID == detail.ArchiveDetails {
						history = *_history
						break
					}
				}
			} else if bvid, _ := util.AvToBv(currentDetail.AID); bvid == push.OuterVID {
				detail = *currentDetail
				for _, _history := range histories {
					if _history.AID == detail.AID {
						history = *_history
						break
					}
				}
				break
			}
		}

		go func() {
			count := 0
			success := false
			var err error
			for count < maxRetryCount && !success {
				if success, err = s.DoQQTGLPushing(&push, &detail, &history); err != nil || !success {
					log.Error("Service: TryQQCMCPushing.DoQQCMCPushing failed or Error %v", err)
					count++
					if xecode.EqualError(xecode.NothingFound, err) {
						count = maxRetryCount
					}
				}
				// 更新detail，新增history
				if err = s.AddBatchDetailActionLog([]*model.ArchivePushBatchHistory{&history}, model.ActionPushUp, username, uid); err != nil {
					log.Error("Service: TryQQCMCPushing.AddBatchDetailActionLog Error (%v)", err)
					return
				}
				if _, err = s.UpdateBatchDetails([]*model.ArchivePushBatchDetail{&detail}); err != nil {
					log.Error("Service: TryQQCMCPushing.UpdateBatchDetails Error (%v)", err)
					return
				}
				time.Sleep(5 * time.Second)
			}
			// 若达到重试上限，则记录修改为失败
			if count == maxRetryCount {
				_detail := &detail
				if _detail.PushStatus != api.ArchivePushDetailPushStatus_SUCCESS && _detail.PushStatus != api.ArchivePushDetailPushStatus_INNER_FAIL {
					_detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
				}
				_history := &history
				if _history.NewPushStatus != api.ArchivePushDetailPushStatus_SUCCESS && _history.NewPushStatus != api.ArchivePushDetailPushStatus_INNER_FAIL {
					_history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
				}
				if err = s.AddBatchDetailActionLog([]*model.ArchivePushBatchHistory{_history}, model.ActionPushUp, username, uid); err != nil {
					log.Error("Service: TryQQCMCPushing.AddBatchDetailActionLog Error (%v)", err)
					return
				}
				if _, err = s.UpdateBatchDetails([]*model.ArchivePushBatchDetail{_detail}); err != nil {
					log.Error("Service: TryQQCMCPushing.UpdateBatchDetails Error (%v)", err)
					return
				}
			}
		}()
	}

	return
}

func (s *Service) DoQQTGLPushing(push *qqModel.ContributeVideoReq, detail *model.ArchivePushBatchDetail, history *model.ArchivePushBatchHistory) (success bool, err error) {
	if push.OuterVID == "" || (detail.ArchiveDetails == "" && detail.AID == 0) || (history.AID == 0 && history.BVID == "") {
		log.Warn("Service: DoQQTGLPushing no valid data to push")
		return false, xecode.NothingFound
	} else if detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL || !strings.HasPrefix(detail.ArchiveDetails, "{") {
		log.Warn("Service: DoQQTGLPushing no valid data to push")
		return false, xecode.NothingFound
	}
	if history.AID == 0 && history.BVID == "" {
		return
	}
	if reply, perr := s.qqDAO.ContributeVideo(push); perr != nil {
		log.Error("Service: DoQQTGLPushing ContributeVideo(%+v) error %v", push, perr)
		return false, ecode.PushRequestError
	} else if reply == nil {
		// 推送失败
		log.Error("Service: DoQQTGLPushing ContributeVideo 推送结果为空")
		detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		success = false
		err = ecode.QQTGLRequestError
	} else {
		if reply.Status == 200 {
			// 推送成功
			archiveDetails := &model.ArchiveMetadataAll{}
			if _err := json.Unmarshal([]byte(detail.ArchiveDetails), archiveDetails); _err != nil {
				log.Error("Service: DoQQTGLPushing ContributeVideo Unmarshal(%s) error %v", detail.ArchiveDetails, _err)
			}
			archiveDetailsBytes, _ := json.Marshal(archiveDetails)
			detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			detail.ArchiveDetails = string(archiveDetailsBytes)
			history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			history.ArchiveDetails = string(archiveDetailsBytes)
			history.Reason = reply.Message
			return true, nil
		} else {
			// 推送失败
			log.Error("Service: DoQQTGLPushing ContributeVideo failed or Error %v", perr)
			detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
			history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
			history.Reason = reply.Message
			success = false
			if perr != nil {
				err = perr
			} else {
				err = ecode.QQTGLRequestError
			}
		}
	}
	return
}

func (s *Service) GenerateBlizzardPushings(details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory) (res interface{}, err error) {
	if len(details) == 0 || len(histories) == 0 {
		return
	}

	_res := make([]*blizzardModel.VodAddReq, 0)
	for _, detail := range details {
		_detail := detail
		if _detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL {
			continue
		}
		arcMetaAll := &model.ArchiveMetadataAll{}
		if _err := json.Unmarshal([]byte(_detail.ArchiveDetails), arcMetaAll); _err != nil {
			log.Error("Service: GenerateQQCMCPushings unmarshal %s error %v", _detail.ArchiveDetails, _err)
			continue
		}
		arc := arcMetaAll.Arc
		tagName := ""
		for tagID := range arcMetaAll.Tags {
			tagName = arcMetaAll.Tags[tagID]
			break
		}
		bvid, _ := util.AvToBv(detail.AID)
		// build cmc push model
		req := &blizzardModel.VodAddReq{
			BVID:        bvid,
			Page:        1,
			Category:    blizzardModel.DefaultVodAddCategory,
			Stage:       tagName,
			Title:       arc.Title,
			Description: arc.Desc,
			Duration:    arc.Duration,
			Thumbnail:   arc.Pic,
			Status:      blizzardModel.VodAddStatusPushUp,
		}
		_res = append(_res, req)
	}
	res = _res

	return
}

func (s *Service) TryBlizzardPushing(pushings []*blizzardModel.VodAddReq, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory, username string, uid int64) {
	if len(pushings) == 0 || len(details) == 0 {
		log.Warn("Service: TryBlizzardPushing no valid data to push")
		return
	}
	maxRetryCount := s.Cfg.Push.MaxRetryCount
	for _, _push := range pushings {
		push := _push
		var detail *model.ArchivePushBatchDetail
		var history *model.ArchivePushBatchHistory
		for _, _detail := range details {
			currentDetail := _detail
			if currentDetail.AID == 0 && push.BVID == currentDetail.ArchiveDetails {
				detail = currentDetail
				for _, _history := range histories {
					if _history.BVID == detail.ArchiveDetails {
						history = _history
						break
					}
				}
			} else if bvid, _ := util.AvToBv(currentDetail.AID); bvid == push.BVID {
				detail = currentDetail
				for _, _history := range histories {
					if _history.AID == detail.AID {
						history = _history
						break
					}
				}
				break
			}
		}

		go func() {
			count := 0
			success := false
			var err error
			for count < maxRetryCount && !success {
				if success, err = s.DoBlizzardPushing(push, detail, history); err != nil || !success {
					log.Error("Service: TryBlizzardPushing DoBlizzardPushing failed or Error %v", err)
					count++
					if xecode.EqualError(xecode.NothingFound, err) {
						count = maxRetryCount
					}
				}
				time.Sleep(5 * time.Second)
			}
			// 若达到重试上限，则记录修改为失败
			if count == maxRetryCount {
				if detail.PushStatus != api.ArchivePushDetailPushStatus_SUCCESS && detail.PushStatus != api.ArchivePushDetailPushStatus_INNER_FAIL {
					detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
				}
				if history.NewPushStatus != api.ArchivePushDetailPushStatus_SUCCESS && history.NewPushStatus != api.ArchivePushDetailPushStatus_INNER_FAIL {
					history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
				}
			}

			if err = s.AddBatchDetailActionLog([]*model.ArchivePushBatchHistory{history}, model.ActionPushUp, username, uid); err != nil {
				log.Error("Service: TryBlizzardPushing AddBatchDetailActionLog Error (%v)", err)
				return
			}
			if _, err = s.UpdateBatchDetails([]*model.ArchivePushBatchDetail{detail}); err != nil {
				log.Error("Service: TryBlizzardPushing UpdateBatchDetails Error (%v)", err)
				return
			}
		}()
	}

	return
}

func (s *Service) DoBlizzardPushing(push *blizzardModel.VodAddReq, detail *model.ArchivePushBatchDetail, history *model.ArchivePushBatchHistory) (success bool, err error) {
	if push == nil || detail == nil {
		log.Error("Service: DoBlizzardPushing push or detail or history empty %v %v %v", push, detail, history)
		return false, ecode.PushRequestError
	}
	if push.BVID == "" || (detail.ArchiveDetails == "" && detail.AID == 0) || (history.AID == 0 && history.BVID == "") {
		log.Warn("Service: DoBlizzardPushing no valid data to push")
		return false, xecode.NothingFound
	} else if detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL || !strings.HasPrefix(detail.ArchiveDetails, "{") {
		log.Warn("Service: DoBlizzardPushing no valid data to push")
		return false, xecode.NothingFound
	}
	if history.AID == 0 && history.BVID == "" {
		return
	}
	if reply, perr := s.blizzardDAO.VodAdd(*push); perr != nil {
		log.Error("Service: DoBlizzardPushing VodAdd(%+v) error %v", push, perr)
		return false, ecode.PushRequestError
	} else if reply == nil {
		// 推送失败
		log.Error("Service: DoBlizzardPushing VodAdd 推送结果为空")
		detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		success = false
		err = ecode.BlizzardRequestError
	} else {
		if reply.Status == 0 {
			// 推送成功
			archiveDetails := &model.ArchiveMetadataAll{}
			if _err := json.Unmarshal([]byte(detail.ArchiveDetails), archiveDetails); _err != nil {
				log.Error("Service: DoBlizzardPushing VodAdd Unmarshal(%s) error %v", detail.ArchiveDetails, _err)
			}
			archiveDetailsBytes, _ := json.Marshal(archiveDetails)
			detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			detail.ArchiveDetails = string(archiveDetailsBytes)
			history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
			history.ArchiveDetails = string(archiveDetailsBytes)
			history.Reason = reply.MSG
			return true, nil
		} else {
			// 推送失败
			log.Error("Service: DoBlizzardPushing VodAdd failed or Error %v", perr)
			detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
			history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
			history.Reason = reply.MSG
			success = false
			if perr != nil {
				err = perr
			} else {
				err = ecode.BlizzardRequestError
			}
		}
	}
	return
}

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"github.com/pkg/errors"
	blizzardModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	"strings"
	"sync"
	"time"

	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
	archiveEcode "go-gateway/app/app-svr/archive/ecode"
)

func (s *Service) GetAllBatches() (list []*model.ArchivePushBatch, err error) {
	if list, err = s.dao.GetAllBatches(); err != nil {
		log.Error("Service: GetAllBatches Error (%v)", err)
		return nil, err
	}
	return
}

func (s *Service) GetBatchesByIDs(ids []int64) (list []*model.ArchivePushBatch, err error) {
	if list, err = s.dao.GetBatchesByIDs(ids); err != nil {
		log.Error("Service: GetBatchesByIDs Error (%v)", err)
		return nil, err
	}
	return
}

func (s *Service) GetBatchesByPage(ids []int64, pushVendorIDs []int64, cuser string, pn int, ps int) (list []*model.ArchivePushBatch, total int64, err error) {
	if list, total, err = s.dao.GetBatchesByPage(ids, pushVendorIDs, cuser, pn, ps); err != nil {
		log.Error("Service: GetAllBatches Error (%v)", err)
		return nil, 0, err
	}
	return
}

func (s *Service) GetBatchDetailsByBatchID(id int64) (batch *model.ArchivePushBatch, details []*model.ArchivePushBatchDetail, err error) {
	if id == 0 {
		return nil, nil, xecode.RequestErr
	}
	var _batch = model.ArchivePushBatch{}
	if _batch, err = s.dao.GetBatchByID(id); err != nil {
		log.Error("Service: GetBatchByID(%d) Error (%v)", id, err)
		return nil, nil, err
	}
	batch = &_batch
	if details, err = s.dao.GetBatchDetailsByBatchID(id); err != nil {
		log.Error("Service: GetBatchDetailsByBatchID(%d) Error (%v)", id, err)
		return nil, nil, err
	}
	return
}

func (s *Service) GetBatchDetailsByBVIDs(vendorIDs []int64, bvids []string, order string, desc bool) (resDetails []*model.ArchivePushBatchDetail, err error) {
	if len(bvids) == 0 {
		return
	}
	if resDetails, err = s.dao.GetBatchDetailsByBVIDs(vendorIDs, bvids, order, desc); err != nil {
		log.Error("Service: GetBatchDetailsByBVIDs(%v) Error (%v)", bvids, err)
		return nil, err
	}

	return
}

func (s *Service) GetBatchExportFilenameFormat() string {
	if s.Cfg.Export == nil || s.Cfg.Export.ArchivePushBatch == nil || s.Cfg.Export.ArchivePushBatch.FilenameFormat == "" {
		return "推送详情（%d） - %s"
	}
	return s.Cfg.Export.ArchivePushBatch.FilenameFormat
}

func (s *Service) GetBatchExportColumns() []string {
	if s.Cfg.Export == nil || s.Cfg.Export.ArchivePushBatch == nil || len(s.Cfg.Export.ArchivePushBatch.Columns) == 0 {
		return []string{}
	}
	return s.Cfg.Export.ArchivePushBatch.Columns
}

func (s *Service) GetBatchExportTitles() []string {
	if s.Cfg.Export == nil || s.Cfg.Export.ArchivePushBatch == nil || len(s.Cfg.Export.ArchivePushBatch.Titles) == 0 {
		return []string{}
	}
	return s.Cfg.Export.ArchivePushBatch.Titles
}

func (s *Service) GetBatchHistoryByTime(batchID int64, time int64) (res map[string]*model.ArchivePushBatchHistory, err error) {
	if batchID == 0 {
		return
	}
	res = make(map[string]*model.ArchivePushBatchHistory)
	queryParams := &model.AuditLogSearchParams{Business: model.BusinessIDBatch, Int1From: time}
	var rawRes *model.AuditLogSearchResRawData
	if rawRes, err = s.dao.SearchAuditLog(queryParams); err != nil {
		log.Error("Service: GetBatchHistory.SearchAuditLog Error (%v)", err)
		return nil, err
	} else if rawRes == nil || len(rawRes.Result) == 0 {
		return
	}
	for _, logObj := range rawRes.Result {
		if logObj.Str4 != api.ArchivePushDetailPushStatus_Enum_name[int32(api.ArchivePushDetailPushStatus_UNKNOWN)] {
			history := &model.ArchivePushBatchHistory{
				BatchID:          0,
				AID:              logObj.OID,
				BVID:             logObj.Str0,
				PushVendorID:     logObj.Int0,
				OldArchiveStatus: api.ArchiveStatus_Enum(api.ArchiveStatus_Enum_value[logObj.Str1]),
				OldPushStatus:    api.ArchivePushDetailPushStatus_Enum(api.ArchivePushDetailPushStatus_Enum_value[logObj.Str2]),
				NewArchiveStatus: api.ArchiveStatus_Enum(api.ArchiveStatus_Enum_value[logObj.Str3]),
				NewPushStatus:    api.ArchivePushDetailPushStatus_Enum(api.ArchivePushDetailPushStatus_Enum_value[logObj.Str4]),
				ArchiveDetails:   logObj.ExtraData,
				Reason:           logObj.Str5,
				CUser:            logObj.UName,
				CTime:            xtime.Time(logObj.Int1),
			}
			if history.BVID == "" {
				if history.BVID, err = util.AvToBv(history.AID); err != nil {
					log.Error("Service: GetBatchHistory.SearchAuditLog Error (%v)", err)
				}
			}
			if _, exists := res[logObj.Str0]; !exists {
				res[logObj.Str0] = history
			}
		}
	}
	return
}

// BatchPush 下载并读取文件，生成batch并推送
func (s *Service) BatchPush(ctx *bm.Context, batchToPush *model.ArchivePushBatch) (batchID int64, err error) {
	if batchToPush == nil {
		return
	}
	username, uid := util.UserInfo(ctx)
	batchToPush.CUser = username
	batchToPush.MUser = username
	now := time.Now().Unix()
	batchToPush.CTime = xtime.Time(now)
	batchToPush.MTime = xtime.Time(now)
	if batchToPush.FileURL == "" {
		err = errors.WithMessage(xecode.RequestErr, "没有下载链接")
		return
	}
	var (
		csvBuf   *bytes.Buffer
		closeBuf func()
		rawBVIDs = make([]string, 0)
	)
	csvBuf, closeBuf, err = s.dao.Download(batchToPush.FileURL)
	if err != nil {
		log.Error("Service: BatchPush.Download Error (%v)", err)
		return
	}
	defer closeBuf()
	r := csv.NewReader(strings.NewReader(string(csvBuf.Bytes())))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("Service: BatchPush.ReadAll Error (%v)", err)
		return
	}
	rawBVIDs = s.getAllBVIDsFromRecords(records)
	batchID, err = s.DoBatchPush(batchToPush, rawBVIDs, 0, username, uid)

	return
}

// DoBatchPush 推送batch
func (s *Service) DoBatchPush(batchToPush *model.ArchivePushBatch, bvids []string, delayMinutes int32, username string, uid int64) (batchID int64, err error) {
	if batchToPush == nil {
		return
	}
	var (
		validBVIDs     = make([]string, 0)
		batchDetails   = make([]*model.ArchivePushBatchDetail, 0)
		batchHistories = make([]*model.ArchivePushBatchHistory, 0)
		pushings       interface{}
	)

	if len(bvids) == 0 {
		err = errors.WithMessage(xecode.RequestErr, "没有合法的BVID")
		return
	}
	defer func() {
		if xecode.EqualError(ecode.BatchNotFound, err) {
			return
		}
		if err = s.AddBatchDetailActionLog(batchHistories, model.ActionPushUp, username, uid); err != nil {
			log.Error("Service: DoBatchPush AddBatchDetailActionLog Error (%v)", err)
			return
		}
		if batchDetails, err = s.UpdateBatchDetails(batchDetails); err != nil {
			log.Error("Service: DoBatchPush UpdateBatchDetails Error (%v)", err)
			return
		}
	}()
	// 创建batch记录
	logNow := time.Now()
	batchToPush.CUser = username
	batchToPush.CTime = xtime.Time(time.Now().Add(time.Duration(delayMinutes) * time.Minute).Unix())
	if batchToPush.ID == 0 {
		log.Info("Service: DoBatchPush CreateBatch start: %s", logNow.Format("2006-01-02 15:04:05"))
		if batchToPush, err = s.dao.CreateBatch(batchToPush); err != nil {
			log.Error("Service: DoBatchPush.CreateBatch Error (%v)", err)
			return
		}
		log.Info("Service: DoBatchPush CreateBatch used: %f", time.Since(logNow).Seconds())
	} else {
		if batches, _, _err := s.GetBatchesByPage([]int64{batchToPush.ID}, []int64{batchToPush.PushVendorID}, "", 1, 1); _err != nil || len(batches) == 0 {
			log.Error("Service: DoBatchPush GetBatchesByPage(%d) Error (%v) or not found", batchID, _err)
			err = ecode.BatchNotFound
			return
		} else {
			batchToPush = batches[0]
		}
	}
	batchID = batchToPush.ID

	// 若立即执行
	if delayMinutes == 0 {
		if batchDetails, batchHistories, validBVIDs, pushings, err = s.GenerateBatchPushings(batchToPush, bvids, username); err != nil {
			log.Error("Service: DoBatchPush GenerateBatchPushings error %+v", err)
		}

		// 将有效的BVIDs放入网关播放白名单池
		logNow = time.Now()
		log.Info("Service: DoBatchPush PutBVIDsForWhiteList start: %s", logNow.Format("2006-01-02 15:04:05"))
		if err = s.dao.PutBVIDsForWhiteList(batchToPush.PushVendorID, validBVIDs); err != nil {
			log.Error("Service: DoBatchPush PutBVIDsForWhiteList Error (%v)", err)
			return
		}
		log.Info("Service: DoBatchPush PutBVIDsForWhiteList used: %f", time.Since(logNow).Seconds())

		// 进行厂商推送
		s.TryPushing(batchToPush.PushVendorID, pushings, batchDetails, batchHistories, username, uid)
	} else {
		// 否则放入待执行池
		delayTime := time.Now().Add(time.Duration(delayMinutes) * time.Minute)
		if err = s.dao.PutBatchWithBVIDsForTodo(batchID, bvids, xtime.Time(delayTime.Unix())); err != nil {
			log.Error("Service: DoBatchPush PutBatchWithBVIDsForTodo (%d, %v, %d) error %v", batchID, bvids, delayTime.Format(model.DefaultTimeLayout))
		}
	}

	return
}

// GenerateBatchPushings 生成推送需要的detail和history和pushings
func (s *Service) GenerateBatchPushings(batchToPush *model.ArchivePushBatch, bvids []string, username string) (resBatchDetails []*model.ArchivePushBatchDetail, resBatchHistories []*model.ArchivePushBatchHistory, validBVIDs []string, pushings interface{}, err error) {
	// 生成对应的detail 和 history objects
	resBatchDetails = make([]*model.ArchivePushBatchDetail, 0)
	resBatchHistories = make([]*model.ArchivePushBatchHistory, 0)

	logNow := time.Now()
	log.Info("Service: GenerateBatchPushings GenerateBaseBatchDetailsByBVIDs start: %s", logNow.Format("2006-01-02 15:04:05"))
	if resBatchDetails, resBatchHistories, err = s.GenerateBaseBatchDetailsByBVIDs(batchToPush, bvids, username); err != nil {
		log.Error("Service: GenerateBatchPushings GenerateBatchDetailsByBVIDs Error (%v)", err)
		return
	}
	log.Info("Service: GenerateBatchPushings GenerateBaseBatchDetailsByBVIDs used: %f", time.Since(logNow).Seconds())
	// 创建并存储details
	logNow = time.Now()
	log.Info("Service: GenerateBatchPushings CreateBatchDetails start: %s", logNow.Format("2006-01-02 15:04:05"))
	if resBatchDetails, err = s.CreateBatchDetails(resBatchDetails); err != nil {
		log.Error("Service: GenerateBatchPushings CreateBatchDetails Error (%v)", err)
		return
	}
	log.Info("Service: GenerateBatchPushings CreateBatchDetails used: %f", time.Since(logNow).Seconds())
	// 检查details信息（包括视频格式等）并生成推送objects
	logNow = time.Now()
	log.Info("Service: GenerateBatchPushings ValidateAndFilterBatchDetails start: %s", logNow.Format("2006-01-02 15:04:05"))
	if resBatchDetails, resBatchHistories, validBVIDs, err = s.ValidateAndFilterBatchDetails(batchToPush, resBatchDetails, resBatchHistories, bvids); err != nil {
		log.Error("Service: GenerateBatchPushings ValidateAndFilterBatchDetails Error (%v)", err)
		return
	}
	log.Info("Service: GenerateBatchPushings ValidateAndFilterBatchDetails used: %f", time.Since(logNow).Seconds())
	// 生成厂商对应的推送模型
	logNow = time.Now()
	log.Info("Service: GenerateBatchPushings GenerateVendorPushings start: %s", logNow.Format("2006-01-02 15:04:05"))
	if pushings, err = s.GenerateVendorPushings(batchToPush.PushVendorID, resBatchDetails, resBatchHistories); err != nil {
		log.Error("Service: GenerateBatchPushings GenerateVendorPushings (%d, %v, %v) error %v", batchToPush.PushVendorID, resBatchDetails, resBatchHistories, err)
		return
	}
	log.Info("Service: GenerateBatchPushings GenerateVendorPushings used: %f", time.Since(logNow).Seconds())

	return
}

func (s *Service) getAllBVIDsFromRecords(records [][]string) (bvids []string) {
	if len(records) == 0 {
		return
	}
	bvidsMap := make(map[string]int)
	for _, rec := range records {
		if len(rec) > 0 {
			bvidsMap[strings.TrimSpace(rec[0])] = 1
		}
	}
	bvids = make([]string, 0)
	for bvid := range bvidsMap {
		bvids = append(bvids, bvid)
	}
	return
}

func (s *Service) GenerateBaseBatchDetailsByBVIDs(batch *model.ArchivePushBatch, bvids []string, username string) (res []*model.ArchivePushBatchDetail, resHistories []*model.ArchivePushBatchHistory, err error) {
	if batch == nil {
		err = ecode.BatchNotFound
		return
	}
	res = make([]*model.ArchivePushBatchDetail, 0)
	resHistories = make([]*model.ArchivePushBatchHistory, 0)
	if len(bvids) == 0 {
		return
	}
	existingBatchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if batch.ID > 0 {
		if batch, existingBatchDetails, err = s.GetBatchDetailsByBatchID(batch.ID); err != nil {
			log.Error("Service: GenerateBaseBatchDetailsByBVIDs GetBatchDetailsByBatchID Error (%v)", err)
			return
		}
	}
	for _, _bvid := range bvids {
		bvid := _bvid
		batchDetail := &model.ArchivePushBatchDetail{
			BatchID: batch.ID,
			CUser:   username,
			CTime:   batch.CTime,
			MUser:   username,
			MTime:   batch.CTime,
		}
		avid, _err := util.BvToAv(bvid)
		if _err != nil {
			log.Error("Service: GenerateBaseBatchDetailsByBVIDs BvToAv(%s) Error (%v)", bvid, _err)
			_err = ecode.AVBVIDConvertingError
			batchDetail.ArchiveDetails = bvid
		}
		existingDetailFlag := false
		for _, _existingDetail := range existingBatchDetails {
			existingDetail := _existingDetail
			if existingDetail.BatchID == batch.ID {
				if (existingDetail.AID != 0 && existingDetail.AID == avid) || (existingDetail.AID == 0 && existingDetail.ArchiveDetails == bvid) {
					existingDetailFlag = true
					batchDetail = existingDetail
				}
			}
		}
		batchDetail.AID = avid
		if xecode.EqualError(ecode.AVBVIDConvertingError, _err) {
			batchDetail.ArchiveStatus = api.ArchiveStatus_NOT_EXISTS
			batchDetail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
		}
		res = append(res, batchDetail)

		history := &model.ArchivePushBatchHistory{
			BatchID:          batch.ID,
			AID:              batchDetail.AID,
			BVID:             bvid,
			PushVendorID:     batch.PushVendorID,
			OldArchiveStatus: api.ArchiveStatus_UNKNOWN,
			OldPushStatus:    api.ArchivePushDetailPushStatus_UNKNOWN,
			NewArchiveStatus: batchDetail.ArchiveStatus,
			NewPushStatus:    batchDetail.PushStatus,
			ArchiveDetails:   batchDetail.ArchiveDetails,
			Reason:           "初始化推送",
			CUser:            batchDetail.CUser,
			CTime:            batchDetail.CTime,
		}
		if existingDetailFlag {
			history.OldArchiveStatus = batchDetail.ArchiveStatus
			history.OldPushStatus = batchDetail.PushStatus
		}
		if avid == 0 {
			history.NewArchiveStatus = api.ArchiveStatus_NOT_EXISTS
			history.NewPushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
		}
		resHistories = append(resHistories, history)
	}
	return
}

func (s *Service) ValidateAndFilterBatchDetails(batch *model.ArchivePushBatch, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory, bvids []string) (resDetails []*model.ArchivePushBatchDetail, resHistories []*model.ArchivePushBatchHistory, validBVIDs []string, err error) {
	if len(details) == 0 || len(bvids) == 0 {
		return
	}
	resDetails = make([]*model.ArchivePushBatchDetail, 0, len(details))
	resHistories = make([]*model.ArchivePushBatchHistory, 0)
	validBVIDs = make([]string, 0)
	var vendor model.ArchivePushVendor
	if vendor, err = s.GetVendorByID(batch.PushVendorID); err != nil {
		log.Error("Service: ValidateAndFilterBatchDetails GetVendorByID(%d) error %v", batch.PushVendorID, err)
		return
	} else if vendor.ID == 0 {
		err = ecode.VendorNotFound
		log.Error("Service: ValidateAndFilterBatchDetails GetVendorByID(%d) not found", batch.PushVendorID)
		return
	}
	var arcsMap map[int64]*archiveGRPC.Arc
	// 根据BVID获取稿件信息
	if arcsMap, err = s.dao.GetArcsByBVIDs(bvids); err != nil {
		log.Error("Service: ValidateAndFilterBatchDetails GetArcsByBVIDs(%v) Error (%v)", bvids, err)
		return
	}
	// 获取已推送成功的稿件，不重复推送
	pushedArchives := model.ArchivePushDetailByBVIDSlice{}
	if pushedArchives, err = s.GetArchivesByPushStatus(bvids, batch.PushVendorID, []api.ArchivePushDetailPushStatus_Enum{
		api.ArchivePushDetailPushStatus_SUCCESS,
		api.ArchivePushDetailPushStatus_UNKNOWN,
		api.ArchivePushDetailPushStatus_OUTER_FAIL,
		api.ArchivePushDetailPushStatus_INNER_FAIL,
	}, 0); err != nil {
		log.Error("archive-push-admin.service.GetPushedArchives.GetArchivesByPushStatus Error (%v)", err)
		return
	}
	lock := sync.RWMutex{}
	eg := errgroup.WithContext(context.Background())
	var appKey string
	if vendor.UserBindable {
		if appKey, err = s.GetOauthAppKeyByVendorID(batch.PushVendorID); err != nil {
			log.Error("Service: GenerateAndFillAuthors GetOauthAppKeyByVendorID(%d) error %v", batch.PushVendorID, err)
			return
		}
	}
	for _, detail := range details {
		_detail := detail
		if _detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL {
			continue
		}
		eg.Go(func(ctx context.Context) error {
			bvid, _ := util.AvToBv(_detail.AID)
			pushed := false
			// 找到history
			lock.Lock()
			var history *model.ArchivePushBatchHistory
			for historyI := range histories {
				if _detail.AID != 0 && histories[historyI].AID == _detail.AID {
					history = histories[historyI]
					break
				} else if _detail.AID == 0 && _detail.ArchiveDetails == histories[historyI].BVID {
					history = histories[historyI]
					break
				}
			}
			if history == nil {
				history = &model.ArchivePushBatchHistory{
					BatchID:          batch.ID,
					AID:              _detail.AID,
					BVID:             bvid,
					PushVendorID:     batch.PushVendorID,
					OldArchiveStatus: api.ArchiveStatus_UNKNOWN,
					OldPushStatus:    api.ArchivePushDetailPushStatus_UNKNOWN,
					NewArchiveStatus: _detail.ArchiveStatus,
					NewPushStatus:    _detail.PushStatus,
					ArchiveDetails:   _detail.ArchiveDetails,
					CUser:            _detail.CUser,
					CTime:            _detail.CTime,
				}
				if _detail.PushStatus != api.ArchivePushDetailPushStatus_UNKNOWN {
					history.OldArchiveStatus = _detail.ArchiveStatus
					history.OldPushStatus = _detail.PushStatus
					history.CTime = xtime.Time(time.Now().Unix())
				}
				histories = append(histories, history)
			}
			// 判断已推送成功的稿件，不重复推送
			if len(pushedArchives) > 0 {
				_pushedArchive := pushedArchives[0]
				if _pushedArchive.ArchiveStatus == api.ArchiveStatus_Enum_name[int32(api.ArchiveStatus_OPEN)] && _pushedArchive.PushStatus == api.ArchivePushDetailPushStatus_Enum_name[int32(api.ArchivePushDetailPushStatus_SUCCESS)] && _pushedArchive.BVID == bvid {
					pushed = true
					if _pushedDetail, _err := json.Marshal(_pushedArchive); _err != nil {
						log.Error("Service: ValidateBatchDetailsAndGeneratePushing Marshal %+v error %v", _pushedArchive, _err)
					} else {
						_detail.ArchiveDetails = string(_pushedDetail)
						history.ArchiveDetails = string(_pushedDetail)
					}
				}
			}
			if pushed {
				history.OldArchiveStatus = api.ArchiveStatus_OPEN
				history.OldPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				history.NewArchiveStatus = api.ArchiveStatus_OPEN
				history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				history.Reason = "稿件已推送"
				_detail.ArchiveStatus = api.ArchiveStatus_OPEN
				_detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				validBVIDs = append(validBVIDs, bvid)
			}
			lock.Unlock()

			arc, exists := arcsMap[_detail.AID]
			if !exists {
				lock.Lock()
				_detail.ArchiveStatus = api.ArchiveStatus_NOT_EXISTS
				_detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
				history.NewArchiveStatus = api.ArchiveStatus_NOT_EXISTS
				history.NewPushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
				lock.Unlock()
			} else if pushed {
				log.Warn("Service: ValidateBatchDetailsAndGeneratePushing %s 稿件已推送", bvid)
			} else {
				// 获取稿件对应tags
				tags, _err := s.dao.GetTagsByAID(_detail.AID)
				if _err != nil {
					log.Error("Service: ValidateAndFilterBatchDetails GetTagsByAID(%d) Error (%v)", _detail.AID, _err)
					lock.Lock()
					_detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
					history.NewPushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
					history.Reason = "调用Tag服务失败"
					lock.Unlock()
					return nil
				}
				tagNames := make([]string, 0)
				for _, tag := range tags {
					tagNames = append(tagNames, tag.Name)
				}
				if valid, _err := s.ValidateArchiveToPush(*arc, batch.PushVendorID, true); _err != nil {
					if xecode.EqualError(archiveEcode.ArchiveNotExist, _err) {
						// 未开放
						lock.Lock()
						_detail.ArchiveStatus = api.ArchiveStatus_NOT_OPEN
						_detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
						lock.Unlock()
					} else {
						log.Error("Service: ValidateAndFilterBatchDetails ValidateArchiveToPush Error (%v)", _err)
						return nil
					}
				} else if !valid {
					lock.Lock()
					_detail.ArchiveStatus = api.ArchiveStatus_FORMAT_INVALID
					_detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
					lock.Unlock()
				} else {
					// valid
					lock.Lock()
					_detail.ArchiveStatus = api.ArchiveStatus_OPEN
					bvid, _ := util.AvToBv(_detail.AID)

					validBVIDs = append(validBVIDs, bvid)
					lock.Unlock()
				}
				lock.Lock()
				history.NewArchiveStatus = _detail.ArchiveStatus
				if _detail.PushStatus == api.ArchivePushDetailPushStatus_INNER_FAIL {
					history.NewPushStatus = _detail.PushStatus
				}
				lock.Unlock()

				arcMetadata := &model.ArchiveMetadataAll{
					Arc:  arc,
					Tags: make(map[int64]string),
				}
				for _, tag := range tags {
					arcMetadata.Tags[tag.Id] = tag.Name
				}
				// 获取稿件作者的openid
				if vendor.ID != model.DefaultVendors[0].ID && vendor.UserBindable {
					if openID, _err := s.GetOpenIDByMID(arc.Author.Mid, appKey); _err != nil {
						log.Error("Service: ValidateAndFilterBatchDetails GetOpenIDByMID(%d, %s) error %v", arc.Author.Mid, appKey, _err)
						lock.Lock()
						_detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
						history.NewPushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
						history.Reason = "获取稿件作者OpenID失败"
						lock.Unlock()
						return nil
					} else {
						arcMetadata.OpenID = openID
					}
				}
				var arcMetaJSON []byte
				if arcMetaJSON, err = json.Marshal(arcMetadata); err != nil {
					log.Error("Service: ValidateAndFilterBatchDetails Marshal Error (%v)", err)
					return nil
				}
				lock.Lock()
				_detail.ArchiveDetails = string(arcMetaJSON)
				history.ArchiveDetails = string(arcMetaJSON)
				lock.Unlock()
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("Service: ValidateAndFilterBatchDetails.eg Error (%v)", err)
	}
	resDetails = details
	resHistories = histories

	return
}

// GenerateVendorPushings 根据不同厂商生成不同
func (s *Service) GenerateVendorPushings(vendorID int64, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory) (res interface{}, err error) {
	if vendorID == 0 || len(details) == 0 || len(histories) == 0 {
		log.Error("Service: GenerateVendorPushings (%d, %v, %v) 参数为空", vendorID, details, histories)
		err = ecode.PushRequestError
		return
	}
	switch vendorID {
	case model.DefaultVendors[0].ID:
		return s.GenerateQQCMCPushings(details, histories)
	case model.DefaultVendors[1].ID:
		return s.GenerateQQTGLPushings(details, histories)
	case model.DefaultVendors[2].ID:
		return s.GenerateBlizzardPushings(details, histories)
	default:
		err = ecode.VendorNotFound
		return
	}
}

// TryPushing 根据厂商进行不同的推送
func (s *Service) TryPushing(vendorId int64, pushings interface{}, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory, username string, uid int64) {
	if vendorId == 0 || pushings == nil || len(details) == 0 || len(histories) == 0 {
		return
	}
	log.Info("Service: TryPushing %d %v %v %v %s", vendorId, pushings, details, histories, username)

	switch vendorId {
	case model.DefaultVendors[0].ID:
		// 王者营地推送
		if qqCMCPushings, ok := pushings.([]*qqModel.PushPGCAdminReq); !ok {
			log.Error("Service: TryPushing (%d, %v) converting to []*qqModel.PushPGCAdminReq type converting error", vendorId, pushings)
		} else {
			s.TryQQCMCPushing(qqCMCPushings, details, histories, username, uid)
		}
		break
	case model.DefaultVendors[1].ID:
		// 王者营地TGL推送
		if qqTGLPushings, ok := pushings.([]*qqModel.ContributeVideoReq); !ok {
			log.Error("Service: TryPushing (%d, %v) converting to []*qqModel.PushPGCAdminReq type converting error", vendorId, pushings)
		} else {
			s.TryQQTGLPushing(qqTGLPushings, details, histories, username, uid)
		}
		break
	case model.DefaultVendors[2].ID:
		// 暴雪推送
		if pushings, ok := pushings.([]*blizzardModel.VodAddReq); !ok {
			log.Error("Service: TryPushing (%d, %v) converting to []*qqModel.PushPGCAdminReq type converting error", vendorId, pushings)
		} else {
			s.TryBlizzardPushing(pushings, details, histories, username, uid)
		}
		break
	default:
		log.Error("Service: TryPushing (%d) 不是有效推送厂商ID", vendorId)
		return
	}
	// 从待推送池移除
	batchIDs := make([]int64, 0)
	for _, detail := range details {
		batchIDs = append(batchIDs, detail.BatchID)
	}
	if len(batchIDs) > 0 {
		if err := s.dao.RemoveBatchesFromTodo(batchIDs); err != nil {
			log.Error("Service: TryPushing RemoveBatchesFromTodo (%v) error %v", batchIDs, err)
		} else {
			log.Info("Service: TryPushing RemoveBatchesFromTodo (%v)", batchIDs)
		}
	}
	return
}

func (s *Service) UpdateBatchDetails(details []*model.ArchivePushBatchDetail) (resDetails []*model.ArchivePushBatchDetail, err error) {
	if len(details) == 0 {
		resDetails = make([]*model.ArchivePushBatchDetail, 0)
		return
	}
	eg := errgroup.WithContext(context.Background())
	for _, _batchDetail := range details {
		batchDetail := _batchDetail
		eg.Go(func(ctx context.Context) error {
			if updateBatchDetail, err := s.dao.UpdateBatchDetail(*batchDetail); err != nil {
				log.Error("Service: UpdateBatchDetail Error (%v)", err)
				return err
			} else {
				resDetails = append(resDetails, updateBatchDetail)
			}
			return nil
		})
	}
	err = eg.Wait()
	return
}

func (s *Service) CreateBatchDetails(batchDetails []*model.ArchivePushBatchDetail) (res []*model.ArchivePushBatchDetail, err error) {
	if len(batchDetails) == 0 {
		res = make([]*model.ArchivePushBatchDetail, 0)
		return
	}
	lock := sync.Mutex{}
	eg := errgroup.WithContext(context.Background())
	for _, _batchDetail := range batchDetails {
		batchDetail := _batchDetail
		if batchDetail.ID == 0 {
			eg.Go(func(ctx context.Context) error {
				if createdBatchDetail, err := s.dao.CreateBatchDetail(*batchDetail); err != nil {
					log.Error("Service: CreateBatchDetails Error (%v)", err)
					return err
				} else {
					lock.Lock()
					res = append(res, createdBatchDetail)
					lock.Unlock()
				}
				return nil
			})
		} else {
			res = append(res, batchDetail)
		}
	}
	err = eg.Wait()
	return
}

func (s *Service) AddBatchDetailActionLog(histories []*model.ArchivePushBatchHistory, action string, username string, uid int64) (err error) {
	if len(histories) == 0 {
		return nil
	}
	eg := errgroup.WithContext(context.Background())
	for _, _history := range histories {
		history := _history
		eg.Go(func(ctx context.Context) error {
			now := time.Now()
			index := []interface{}{
				history.BatchID,
				now.Unix(),
				history.BVID,
				api.ArchiveStatus_Enum_name[int32(history.OldArchiveStatus)],
				api.ArchivePushDetailPushStatus_Enum_name[int32(history.OldPushStatus)],
				api.ArchiveStatus_Enum_name[int32(history.NewArchiveStatus)],
				api.ArchivePushDetailPushStatus_Enum_name[int32(history.NewPushStatus)],
				history.Reason,
			}

			params := &model.AuditLogInitParams{
				UName:    username,
				UID:      uid,
				Business: model.BusinessIDBatch,
				Type:     int(history.PushVendorID),
				OID:      history.AID,
				Action:   action,
				CTime:    now,
				Index:    index,
				Content:  history.ArchiveDetails,
			}
			if err := s.dao.AddAuditLog(params); err != nil {
				log.Error("Service: AddBatchDetailActionLog Error (%v)", err)
				return err
			}
			return nil
		})
	}
	err = eg.Wait()

	return
}

// CheckAndPushTodoArchives 检查待推送稿件池并推送
func (s *Service) CheckAndPushTodoArchives() {
	var (
		err                   error
		todoBatchIDs          = make([]int64, 0)
		todoBatchTimesMap     map[int64]time.Time
		needPushBatchBVIDsMap map[int64][]string
		needPushBatchIDs      = make([]int64, 0)
		batches               []*model.ArchivePushBatch
		authorBatches         = make([]*model.ArchivePushBatch, 0)
		bvidBatches           = make([]*model.ArchivePushBatch, 0)
		authorPushes          []*model.ArchivePushAuthorPushX
	)
	// 获取待推送batch ids
	if todoBatchIDs, err = s.dao.GetBatchesIDsFromTodo(); err != nil {
		log.Error("Service: CheckAndPushTodoArchives GetBatchesIDsFromTodo error %v", err)
		return
	} else if len(todoBatchIDs) == 0 {
		log.Info("Service: CheckAndPushTodoArchives GetBatchesIDsFromTodo 没有要推送的batch")
		return
	}
	// 获取待推送时间map
	if todoBatchTimesMap, err = s.dao.GetBatchesPushTimeFromTodo(todoBatchIDs); err != nil {
		log.Error("Service: CheckAndPushTodoArchives GetBatchesPushTimeFromTodo (%v) error %v", todoBatchIDs, err)
		return
	} else if len(todoBatchTimesMap) == 0 {
		log.Warn("Service: CheckAndPushTodoArchives GetBatchesPushTimeFromTodo %v 没有batch对应的待推送时间", todoBatchIDs)
		return
	}
	// 检查哪些需要推送
	now := time.Now()
	for _, batchID := range todoBatchIDs {
		if pushTime, exists := todoBatchTimesMap[batchID]; exists {
			if now.After(pushTime) {
				log.Info("Service: CheckAndPushTodoArchives batch %d 即将推送", batchID)
				needPushBatchIDs = append(needPushBatchIDs, batchID)
			}
		} else {
			log.Warn("Service: CheckAndPushTodoArchives batch %d 没有对应的推送时间", batchID)
		}
	}
	if len(needPushBatchIDs) == 0 {
		if s.Cfg.Debug {
			log.Warn("Service: CheckAndPushTodoArchives 没有要推送的batch")
		}
		return
	}
	// 获取待推送bvids map
	if needPushBatchBVIDsMap, err = s.dao.GetBatchesBVIDsFromTodo(needPushBatchIDs); err != nil {
		log.Error("Service: CheckAndPushTodoArchives GetBatchesBVIDsFromTodo (%v) error %v", needPushBatchIDs, err)
		return
	}

	// 根据batch ids获取batch信息
	if batches, err = s.GetBatchesByIDs(todoBatchIDs); err != nil {
		log.Error("Service: CheckAndPushTodoArchives GetBatchesByIDs (%v) error %v", todoBatchIDs, err)
		return
	} else if len(batches) == 0 {
		log.Error("Service: CheckAndPushTodoArchives GetBatchesByIDs (%v) error %v", todoBatchIDs, err)
		return
	}
	// 根据batches判断是否作者维度推送
	for _, batch := range batches {
		_batch := batch
		switch api.ArchivePushType_Enum(_batch.PushType) {
		case api.ArchivePushType_BVID:
			bvidBatches = append(bvidBatches, _batch)
			break
		case api.ArchivePushType_AUTHOR:
			authorBatches = append(authorBatches, _batch)
			break
		default:
			log.Warn("Service: CheckAndPushTodoArchives batch %d 推送类型不正确", _batch.ID)
		}
	}

	// 非作者维度推送则直接推送
	for _, batch := range bvidBatches {
		_batch := batch
		if details, histories, validBVIDs, pushings, _err := s.GenerateBatchPushings(_batch, needPushBatchBVIDsMap[_batch.ID], "system"); _err != nil {
			log.Error("Service: CheckAndPushTodoArchives batch %d GenerateBatchPushings error %v", _batch.ID, _err)
			continue
		} else {
			s.TryPushing(_batch.PushVendorID, pushings, details, histories, "system", 0)
			if _err := s.dao.PutBVIDsForWhiteList(_batch.PushVendorID, validBVIDs); _err != nil {
				log.Error("Service: CheckAndPushTodoArchives batch %d PutBVIDsForWhiteList %d %v error %v", _batch.ID, _batch.PushVendorID, validBVIDs, _err)
			}
		}
	}

	// 作者维度推送则检查是否仍符合推送条件，否则取消并检查作者推送条件重建batch
	if len(authorBatches) == 0 {
		return
	}
	// 获取作者推送作为推送条件
	if authorPushes, err = s.GetAllAuthorPushes(); err != nil {
		log.Error("Service: CheckAndPushTodoArchives GetAllAuthorPushes error %v", err)
		return
	} else if len(authorPushes) == 0 {
		log.Warn("Service: CheckAndPushTodoArchives GetAllAuthorPushes 没有找到作者推送数据")
		return
	}
	for _, batch := range authorBatches {
		_batch := batch
		for _, authorPush := range authorPushes {
			_authorPush := authorPush
			if _authorPush.VendorID == _batch.PushVendorID {
				if bvids, exists := needPushBatchBVIDsMap[_batch.ID]; exists {
					if author, _err := s.GetAuthorByBVID(_batch.PushVendorID, bvids[0]); _err != nil {
						log.Error("Service: CheckAndPushTodoArchives GetAuthorPushByBVID %s error %v", bvids[0], _err)
						continue
					} else if author == nil {
						log.Error("Service: CheckAndPushTodoArchives GetAuthorPushByBVID %s 获取作者为空", bvids[0])
						continue
					} else {
						_author := *author
						if valid := s.ValidateAuthorPushConditionsWithAuthor(_authorPush.PushConditions, _author); valid {
							if details, histories, validBVIDs, pushings, _err := s.GenerateBatchPushings(_batch, needPushBatchBVIDsMap[_batch.ID], "system"); _err != nil {
								log.Error("Service: CheckAndPushTodoArchives batch %d GenerateBatchPushings error %v", _batch.ID, _err)
								continue
							} else {
								s.TryPushing(_batch.PushVendorID, pushings, details, histories, "system", 0)
								if _err := s.dao.PutBVIDsForWhiteList(_batch.PushVendorID, validBVIDs); _err != nil {
									log.Error("Service: CheckAndPushTodoArchives batch %d PutBVIDsForWhiteList %d %v error %v", _batch.ID, _batch.PushVendorID, validBVIDs, _err)
								}
							}

						} else {
							log.Error("Service: CheckAndPushTodoArchives batch %d 作者状态已变更不符合推送条件，取消推送")
						}
					}
				}
				break
			}
		}
	}
}

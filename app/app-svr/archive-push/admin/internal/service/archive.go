package service

import (
	"context"
	blizzardModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	"sort"
	"sync"
	"time"

	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	v1 "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-main/app/ep/hassan/mock/support/slice"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
	archiveEcode "go-gateway/app/app-svr/archive/ecode"
)

func (s *Service) GetArchivesByPushStatus(bvids []string, vendorID int64, pushStatuses []api.ArchivePushDetailPushStatus_Enum, pushType api.ArchivePushType_Enum) (res model.ArchivePushDetailByBVIDSlice, err error) {
	res = make([]*model.ArchivePushDetailByBVID, 0)
	if len(pushStatuses) == 0 {
		return
	}
	var details []*model.ArchivePushBatchDetailWithVendor
	if details, err = s.dao.GetBatchDetailsByPushStatuses(vendorID, pushStatuses, pushType); err != nil {
		log.Error("Service: GetArchivesByPushStatus GetBatchDetailsByPushStatuses Error (%v)", err)
		return
	} else if len(details) == 0 {
		return
	}
	detailsBVIDMap := make(map[int64]map[string][]*model.ArchivePushBatchDetailWithVendor) // [vendorID][bvid]detail
	for _, _detail := range details {
		detail := _detail
		if detail.VendorID > 0 {
			if _, exists := detailsBVIDMap[detail.VendorID]; !exists {
				detailsBVIDMap[detail.VendorID] = make(map[string][]*model.ArchivePushBatchDetailWithVendor)
			}
		}
		bvid := ""
		if detail.AID == 0 {
			bvid = detail.ArchiveDetails
		} else {
			if bvid, err = util.AvToBv(detail.AID); err != nil {
				log.Error("Service: GetArchivesByPushStatus AvToBv(%d) Error (%v)", detail.AID, err)
				return nil, err
			}
		}
		if _, exists := detailsBVIDMap[detail.VendorID][bvid]; !exists {
			detailsBVIDMap[detail.VendorID][bvid] = make([]*model.ArchivePushBatchDetailWithVendor, 0)
		}
		detailsBVIDMap[detail.VendorID][bvid] = append(detailsBVIDMap[detail.VendorID][bvid], detail)
	}

	for _vendorID := range detailsBVIDMap {
		vendorID := _vendorID
		for bvid := range detailsBVIDMap[vendorID] {
			_details := detailsBVIDMap[vendorID][bvid]
			if len(bvids) > 0 && !slice.Contains(bvids, bvid) {
				continue
			}
			batchIDs := make([]int64, 0)
			for _, detail := range detailsBVIDMap[vendorID][bvid] {
				batchIDs = append(batchIDs, detail.BatchID)
			}
			_detailByBVID := &model.ArchivePushDetailByBVID{
				BVID:          bvid,
				VendorID:      vendorID,
				ArchiveStatus: api.ArchiveStatus_Enum_name[int32(_details[0].ArchiveStatus)],
				PushStatus:    api.ArchivePushDetailPushStatus_Enum_name[int32(_details[0].PushStatus)],
				PushType:      api.ArchivePushType_Enum_name[_details[0].PushType],
				BatchIDs:      batchIDs,
				CUser:         _details[0].CUser,
				CTime:         _details[0].CTime,
			}
			res = append(res, _detailByBVID)
		}
	}
	sort.Stable(res)

	return
}

func (s *Service) GetPushedArchives(vendorID int64, bvids []string, pushType int32, pn int, ps int) (res []*model.ArchivePushDetailByBVID, total int, err error) {
	pushedStatues := []api.ArchivePushDetailPushStatus_Enum{
		api.ArchivePushDetailPushStatus_UNKNOWN,
		api.ArchivePushDetailPushStatus_SUCCESS,
		api.ArchivePushDetailPushStatus_INNER_FAIL,
		api.ArchivePushDetailPushStatus_OUTER_FAIL,
	}
	var _res []*model.ArchivePushDetailByBVID
	if _res, err = s.GetArchivesByPushStatus(bvids, vendorID, pushedStatues, api.ArchivePushType_Enum(pushType)); err != nil {
		log.Error("archive-push-admin.service.GetPushedArchives.GetArchivesByPushStatus Error (%v)", err)
		return
	}
	res = make([]*model.ArchivePushDetailByBVID, 0)
	total = len(_res)
	if total == 0 {
		return
	}
	startI := (pn - 1) * ps
	endI := pn * ps
	if startI > total {
		startI = total
	}
	if endI > total {
		endI = total
	}
	res = _res[startI:endI]
	return
}

func (s *Service) WithdrawArchive(bvid string, reason string, vendorID int64, removeFromWhiteList bool, username string, uid int64) (err error) {
	if bvid == "" {
		return archiveEcode.ArchiveNotExist
	}
	var (
		batchDetails   []*model.ArchivePushBatchDetail
		batchHistories []*model.ArchivePushBatchHistory
		req            interface{}
	)
	if batchDetails, err = s.dao.GetBatchDetailsByBVID(vendorID, bvid); err != nil {
		log.Error("Service: WithdrawArchive GetBatchDetailsByBVID(%d, %s) error %v", vendorID, bvid, err)
		return
	} else if len(batchDetails) == 0 {
		err = archiveEcode.ArchiveNotExist
		return
	}
	detail := batchDetails[0]
	if able := s.ValidateDetailToWithdraw(*detail); !able {
		err = ecode.ArchiveCannotBeWithDrawn
		return
	}
	if req, err = s.GenerateWithdrawing(vendorID, bvid, reason, username, uid); err != nil {
		log.Error("Service: WithdrawArchive GenerateWithdrawing error %v", err)
		return
	}
	history := &model.ArchivePushBatchHistory{
		BatchID:          detail.BatchID,
		AID:              detail.AID,
		BVID:             bvid,
		PushVendorID:     vendorID,
		OldArchiveStatus: detail.ArchiveStatus,
		OldPushStatus:    detail.PushStatus,
		NewArchiveStatus: api.ArchiveStatus_WITHDRAW,
		NewPushStatus:    api.ArchivePushDetailPushStatus_UNKNOWN,
		ArchiveDetails:   detail.ArchiveDetails,
		Reason:           reason,
		CUser:            username,
		CTime:            xtime.Time(time.Now().Unix()),
	}
	detail.ArchiveStatus = api.ArchiveStatus_WITHDRAW
	batchHistories = []*model.ArchivePushBatchHistory{history}
	defer func() {
		if err = s.AddBatchDetailActionLog(batchHistories, model.ActionPushDown, username, uid); err != nil {
			log.Error("archive-push-admin.service.BatchPush.AddBatchDetailActionLog Error (%v)", err)
			return
		}
		if batchDetails, err = s.UpdateBatchDetails(batchDetails); err != nil {
			log.Error("archive-push-admin.service.BatchPush.UpdateBatchDetails Error (%v)", err)
			return
		}
	}()
	details, histories, _err := s.DoWithdrawArchive(vendorID, req, detail, history)
	if err = _err; _err != nil {
		log.Error("Service: WithdrawArchive DoWithdrawArchive(%d, %+v, %+v, %+v) error %v", vendorID, req, detail, history, err)
		return
	} else if len(details) == 0 || len(histories) == 0 {
		err = xecode.RequestErr
		return
	}
	if removeFromWhiteList {
		if err = s.dao.RemoveBVIDsForWhiteList(vendorID, []string{bvid}); err != nil {
			log.Error("Service: WithdrawArchive RemoveBVIDsForWhiteList Error (%v)", err)
			return
		}
	}

	return
}

// GenerateWithdrawing 生成下架models
func (s *Service) GenerateWithdrawing(vendorID int64, bvid string, reason string, username string, uid int64) (res interface{}, err error) {
	switch vendorID {
	case model.DefaultVendors[0].ID, model.DefaultVendors[1].ID:
		res, err = s.GenerateWithdrawingQQCMC(bvid)
		break
	case model.DefaultVendors[2].ID:
		res, err = s.GenerateWithdrawingBlizzard(bvid)
		break
	default:
		err = ecode.VendorNotFound
		return
	}
	return
}

func (s *Service) DoWithdrawArchive(vendorID int64, req interface{}, detail *model.ArchivePushBatchDetail, history *model.ArchivePushBatchHistory) (resDetails []*model.ArchivePushBatchDetail, resHistories []*model.ArchivePushBatchHistory, err error) {
	switch vendorID {
	case model.DefaultVendors[0].ID, model.DefaultVendors[1].ID:
		qqCMCReq, ok := req.(*qqModel.ModifyPGCAdminReq)
		if !ok {
			err = ecode.PushRequestError
			return
		}
		resDetails, resHistories, err = s.DoQQCMCWithdrawings([]*qqModel.ModifyPGCAdminReq{qqCMCReq}, []*model.ArchivePushBatchDetail{detail}, []*model.ArchivePushBatchHistory{history})
		break
	case model.DefaultVendors[2].ID:
		blizzardReq, ok := req.(*blizzardModel.VodAddReq)
		if !ok {
			err = ecode.PushRequestError
			return
		}
		resDetails, resHistories, err = s.DoBlizzardWithdrawings([]*blizzardModel.VodAddReq{blizzardReq}, []*model.ArchivePushBatchDetail{detail}, []*model.ArchivePushBatchHistory{history})
		break
	default:
		err = ecode.VendorNotFound
		return
	}
	return
}

// CheckIfBVIDInWhiteList 检查稿件BVID是否在白名单中（需要进行自动上下架操作）
func (s *Service) CheckIfBVIDInWhiteList(vendorID int64, bvid string, mid int64) (valid bool, err error) {
	if vendorID == 0 || bvid == "" {
		return false, xecode.RequestErr
	}
	valid = false
	for _, vendor := range model.DefaultVendors {
		_vendor := vendor
		if _vendor.ID == vendorID {
			if _vendor.UserBindable && mid != 0 {
				// 检查稿件作者是否在用户白名单中
				userArcWhiteList := make([]string, 0)
				if userArcWhiteList, err = s.dao.GetAuthorBVIDsWhiteList(vendorID, mid); err != nil {
					log.Error("Service: CheckIfBVIDInWhiteList GetAuthorBVIDsWhiteList(%d, %d) error %v", vendorID, mid, err)
					return
				}
				if slice.Contains(userArcWhiteList, bvid) {
					return true, nil
				}
			}
			// 检查稿件是否在稿件白名单中
			whiteList := make([]string, 0)
			if whiteList, err = s.dao.GetBVIDsWhiteList(vendorID); err != nil {
				log.Error("Service: CheckIfBVIDInWhiteList GetBVIDsWhiteList %d Error %v", vendorID, err)
				return
			}
			if slice.Contains(whiteList, bvid) {
				return true, nil
			}
			return false, nil
		}
	}

	// 若未找到vendor，返回错误
	return false, ecode.VendorNotFound
}

// CheckIfBVIDInWhiteList 检查稿件BVID是否在作者稿件白名单中（需要进行自动上下架操作）
func (s *Service) CheckIfBVIDInAuthorWhiteList(vendorID int64, bvid string) (exists bool, err error) {
	if vendorID == 0 || bvid == "" {
		return false, xecode.RequestErr
	}
	exists = false
	var (
		mid int64
		aid int64
		arc *archiveGRPC.Arc
	)
	aid, _ = util.BvToAv(bvid)
	if arc, err = s.dao.GetArcByAID(aid); err != nil {
		log.Error("Service: CheckIfBVIDInAuthorWhiteList(%d, %s) error %v", vendorID, bvid, err)
		return
	}
	mid = arc.Author.Mid

	for _, vendor := range model.DefaultVendors {
		_vendor := vendor
		if _vendor.UserBindable && _vendor.ID == vendorID {
			// 检查稿件BVID是否在作者稿件白名单中
			userArcWhiteList := make([]string, 0)
			if userArcWhiteList, err = s.dao.GetAuthorBVIDsWhiteList(vendorID, mid); err != nil {
				log.Error("Service: CheckIfBVIDInWhiteList GetAuthorBVIDsWhiteList(%d, %d) error %v", vendorID, mid, err)
				return
			}
			return slice.Contains(userArcWhiteList, bvid), nil
		}
	}

	// 若未找到vendor，返回错误
	return false, ecode.VendorNotFound
}

func (s *Service) ValidateDetailToWithdraw(detail model.ArchivePushBatchDetail) bool {
	if detail.ID == 0 {
		return false
	}
	if detail.PushStatus != api.ArchivePushDetailPushStatus_SUCCESS {
		return false
	}
	return true
}

func (s *Service) ValidateArchiveToPush(arc archiveGRPC.Arc, vendorID int64, initialPushFlag bool) (valid bool, err error) {
	valid = false

	if arc.Aid <= 0 {
		err = archiveEcode.ArchiveNotExist
		return
	} else if arc.State < 0 {
		// 未开放
		return
	}
	if vendorID == 0 || vendorID == model.DefaultVendors[0].ID || vendorID == model.DefaultVendors[1].ID {
		// 王者营地稿件要求
		if arc.Videos > 1 {
			// 多P视频
			return
		} else if len(arc.StaffInfo) > 0 || arc.Attribute>>24&int32(1) == 1 {
			// 联合投稿
			return
		} else if arc.Dimension.Width < arc.Dimension.Height || (arc.Dimension.Width > arc.Dimension.Height && arc.Dimension.Rotate == 1) {
			// 竖屏视频
			return
		} else if arc.Attribute>>29&int32(1) == 1 {
			// 互动视频
			return
		}
	}

	// 若为重新上架逻辑，先检查稿件是否在白名单中
	if !initialPushFlag {
		var bvid string
		if bvid, err = util.AvToBv(arc.Aid); err != nil {
			log.Error("archive-push-admin.service.ValidateArchiveToPush.AvToBv(%d) Error %v", arc.Aid, err)
			return false, nil
		}

		whiteList := make([]string, 0)
		if whiteList, err = s.dao.GetBVIDsWhiteList(vendorID); err != nil {
			log.Error("archive-push-admin.service.ValidateArchiveToPush.GetBVIDsWhiteList Error %v", err)
			return
		}
		if slice.Contains(whiteList, bvid) {
			return true, nil
		}
	}

	return true, nil
}

func (s *Service) GetArcByAID(aid int64) (res *archiveGRPC.Arc, err error) {
	if aid == 0 {
		return
	}
	if res, err = s.dao.GetArcByAID(aid); err != nil {
		log.Error("archive-push-admin.service.GetArcByAID(%d) Error(%v)", aid, err)
		return nil, err
	}
	return
}

func (s *Service) GetTagsByAID(aid int64) (res []*v1.Tag, err error) {
	res = make([]*v1.Tag, 0)
	if aid == 0 {
		return
	}
	if res, err = s.dao.GetTagsByAID(aid); err != nil {
		log.Error("archive-push-admin.service.GetTagsByAID(%d) Error(%v)", aid, err)
		return
	}
	return
}

func (s *Service) SyncArchiveStatus(sync model.SyncArchiveStatusReq) (err error) {
	var (
		batchDetail  *model.ArchivePushBatchDetail
		batchHistory *model.ArchivePushBatchHistory
	)
	if detailList, _err := s.GetBatchDetailsByBVIDs([]int64{sync.VendorID}, []string{sync.BVID}, "mtime", true); _err != nil || len(detailList) == 0 || detailList[0].ID == 0 {
		log.Error("Service: SyncArchiveStatus GetBatchDetailsByBVIDs(%d, %s) error %v", sync.VendorID, sync.BVID, _err)
		err = ecode.BatchDetailNotFound
		return
	} else {
		batchDetail = detailList[0]
	}

	switch sync.VendorID {
	case model.DefaultVendors[0].ID, model.DefaultVendors[1].ID:
		batchHistory, err = s.SyncArchiveStatusQQ(sync, batchDetail)
	default:
		err = ecode.SyncRequestError
	}

	// 行为日志
	if _err := s.AddBatchDetailActionLog([]*model.ArchivePushBatchHistory{batchHistory}, model.ActionBackflow, "system", 0); _err != nil {
		log.Error("Service: SyncArchiveStatus AddBatchDetailActionLog(%+v) error %v", sync, _err)
	}

	return
}

// GetArcsByAuthorsAndTags 根据作者及tags取稿件信息
//
// 若有tags限定，则会查询稿件tags，并过滤。
// 若无tags限定，直接取作者下所有稿件
func (s *Service) GetArcsByAuthorsAndTags(authors []*model.ArchivePushAuthor, tags []string) (res map[int64][]*archiveGRPC.Arc, err error) {
	if len(authors) == 0 {
		err = xecode.RequestErr
		log.Error("Service: GetArcsByAuthorsAndTags %v 没有要查询的作者", authors)
		return
	}
	res = make(map[int64][]*archiveGRPC.Arc)

	var (
		tempArcs []*archiveGRPC.Arc
	)
	for _, author := range authors {
		_author := author
		res[_author.ID] = make([]*archiveGRPC.Arc, 0)
		if tempArcs, err = s.dao.GetUpArcsByMID(_author.MID); err != nil {
			if xecode.EqualError(xecode.NothingFound, err) {
				log.Warn("Service: GetArcsByAuthorsAndTags GetUpArcsByMID(%d) 作者没有稿件", _author.MID)
				err = nil
				continue
			}
			log.Error("Service: GetArcsByAuthorsAndTags GetUpArcsByMID(%d) error %v", _author.MID, err)
			return
		}
		eg := errgroup.WithContext(context.Background())
		var lock sync.Mutex
		for _, arc := range tempArcs {
			_arc := arc
			if _arc == nil || _arc.State < 0 {
				continue
			}
			// 若有tag限定，取tags并判断
			if len(tags) > 0 {
				eg.Go(func(ctx context.Context) error {
					if arcTags, _err := s.dao.GetTagsByAID(_arc.Aid); _err != nil {
						return _err
					} else if len(arcTags) > 0 {
						for _, tag := range arcTags {
							_tag := tag
							if slice.Contains(tags, _tag.Name) {
								lock.Lock()
								res[_author.ID] = append(res[_author.ID], _arc)
								lock.Unlock()
							}
						}
					}

					return nil
				})
			} else {
				// 若无tag限定，直接取所有稿件
				res[_author.ID] = append(res[_author.ID], _arc)
			}
		}
		if err = eg.Wait(); err != nil {
			return
		}
	}

	return
}

// PutAuthorBVIDsForWhiteList 将
func (s *Service) PutAuthorBVIDsForWhiteList(vendorID int64, mid int64, bvids []string) (err error) {
	if vendorID == 0 || mid == 0 {
		return xecode.RequestErr
	} else if len(bvids) == 0 {
		return
	}

	if err = s.dao.PutAuthorBVIDsForWhiteList(vendorID, mid, bvids); err != nil {
		log.Error("Service: PutAuthorBVIDsForWhiteList PutAuthorBVIDsForWhiteList (%d, %d, %v) error %v", vendorID, mid, bvids, err)
	}

	return
}

package service

import (
	"context"
	"encoding/json"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	tagGRPC "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/pkg/errors"
	blizzardModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	"go-gateway/app/app-svr/archive-push/ecode"
	archiveEcode "go-gateway/app/app-svr/archive/ecode"
	"strconv"
	"strings"
	"sync"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
)

func (s *Service) SyncArchiveStatusQQ(sync model.SyncArchiveStatusReq, batchDetail *model.ArchivePushBatchDetail) (batchHistory *model.ArchivePushBatchHistory, err error) {
	if sync.BVID == "" || batchDetail == nil || batchDetail.ID == 0 {
		return
	}
	var bvid string
	if bvid, err = util.AvToBv(batchDetail.AID); err != nil {
		log.Error("Service: SyncArchiveStatusQQ AvToBv(%d) error %v", batchDetail.AID, err)
		return
	}
	batchHistory = &model.ArchivePushBatchHistory{
		BatchID:          batchDetail.BatchID,
		AID:              batchDetail.AID,
		BVID:             bvid,
		PushVendorID:     sync.VendorID,
		OldArchiveStatus: batchDetail.ArchiveStatus,
		OldPushStatus:    batchDetail.PushStatus,
		NewArchiveStatus: 0,
		NewPushStatus:    0,
		ArchiveDetails:   batchDetail.ArchiveDetails,
		Reason:           "",
		CUser:            "system",
		CTime:            0,
	}
	switch sync.Status {
	case string(qqModel.SyncArchiveStatusRejected):
		batchHistory.NewArchiveStatus = api.ArchiveStatus_WITHDRAW
		batchHistory.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		batchHistory.Reason = sync.StatusMsg
		batchDetail.ArchiveStatus = api.ArchiveStatus_WITHDRAW
		batchDetail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
		if _, _err := s.UpdateBatchDetails([]*model.ArchivePushBatchDetail{batchDetail}); _err != nil {
			log.Error("Service: SyncArchiveStatusQQ UpdateBatchDetails(%+v) error %v", batchDetail, _err)
			return
		}
		break
	case string(qqModel.SyncArchiveStatusReceived), string(qqModel.SyncArchiveStatusPassed):
		batchHistory.NewArchiveStatus = api.ArchiveStatus_OPEN
		batchHistory.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
		break
	}

	return
}

///////////////////
// QQ CMC Start

// GenerateWithdrawingQQCMC 生成游戏说推送下架models
func (s *Service) GenerateWithdrawingQQCMC(bvid string) (res *qqModel.ModifyPGCAdminReq, err error) {
	if bvid == "" {
		return
	}
	res = &qqModel.ModifyPGCAdminReq{
		SVID: bvid,
	}

	return
}

// DoQQCMCWithdrawings 进行游戏说稿件下架
func (s *Service) DoQQCMCWithdrawings(pushings []*qqModel.ModifyPGCAdminReq, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory) (resDetails []*model.ArchivePushBatchDetail, resHistories []*model.ArchivePushBatchHistory, err error) {
	if len(pushings) == 0 || len(details) == 0 {
		log.Warn("Service: DoQQCMCWithdrawings no valid data to push")
		return
	}
	lock := sync.RWMutex{}
	eg := errgroup.WithContext(context.Background())
	for _, _push := range pushings {
		push := _push
		eg.Go(func(_ context.Context) error {
			var detail *model.ArchivePushBatchDetail
			var history *model.ArchivePushBatchHistory
			var bvid string
			lock.Lock()
			for detailI := range details {
				if bvid, _ = util.AvToBv(details[detailI].AID); bvid == push.SVID {
					detail = details[detailI]
					detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
					for historyI := range histories {
						if histories[historyI].AID == details[detailI].AID {
							history = histories[historyI]
							history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
							break
						}
					}
					break
				}
			}
			lock.Unlock()
			// 获取 docid
			if detail == nil || len(detail.ArchiveDetails) == 0 {
				return ecode.BatchDetailNotFound
			}
			docidObj := &struct {
				DocID string `json:"docid"`
			}{}
			if perr := json.Unmarshal([]byte(detail.ArchiveDetails), docidObj); err != nil {
				log.Error("Service: DoQQCMCWithdrawings Unmarshal(%s) Error %v", detail.ArchiveDetails, perr)
				return perr
			} else if docidObj.DocID == "" {
				log.Warn("Service: DoQQCMCWithdrawings got no docid for bvid (%s) in detail", bvid)
				if docID, _err := s.GetQQCMCDocIDByBVID(bvid); _err != nil {
					log.Error("Service: DoQQCMCWithdrawings GetQQCMCDocIDByBVID (%s) error %v", bvid, _err)
					return ecode.QQCMCRequestError
				} else if docID == "" {
					log.Error("Service: DoQQCMCWithdrawings GetQQCMCDocIDByBVID (%s) 获取docid为空", bvid)
					return ecode.QQCMCRequestError
				} else {
					docidObj.DocID = docID
				}
			}
			if reply, perr := s.qqDAO.ModifyPGCAdmin(docidObj.DocID, qqModel.ModifyModeWithdraw, push); perr != nil {
				log.Error("Service: DoQQCMCWithdrawings ModifyPGCAdmin Error %v", perr)
				return perr
			} else if reply.Status == 0 && reply.Data.ID != "" {
				lock.Lock()
				detail.ArchiveStatus = api.ArchiveStatus_WITHDRAW
				detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				if history != nil {
					history.NewArchiveStatus = api.ArchiveStatus_WITHDRAW
					history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				}
				lock.Unlock()
				return ecode.QQCMCRequestError
			} else if reply.Status == -1 && strings.TrimSpace(reply.MSG) == "无法更新内容因为: 下架内容无法再次下架" {
				lock.Lock()
				detail.ArchiveStatus = api.ArchiveStatus_WITHDRAW
				detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				if history != nil {
					history.NewArchiveStatus = api.ArchiveStatus_WITHDRAW
					history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				}
				lock.Unlock()
				return nil
			} else {
				return errors.WithMessage(xecode.RequestErr, reply.MSG)
			}
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("Service: DoQQCMCWithdrawings eg Error (%v)", err)
	}
	resDetails = details
	resHistories = histories

	return
}

// GetQQCMCDocIDByBVID 根据BVID获取游戏说docid
func (s *Service) GetQQCMCDocIDByBVID(bvid string) (docid string, err error) {
	if bvid == "" {
		return "", xecode.RequestErr
	}
	var (
		aid int64
		arc *archiveGRPC.Arc
	)
	if aid, err = util.BvToAv(bvid); err != nil {
		log.Error("Service: GetQQCMCDocIDByBVID BvToAv %s error %v", bvid, err)
		return
	}
	if arc, err = s.GetArcByAID(aid); err != nil {
		log.Error("Service: GetQQCMCDocIDByBVID GetArcByAID %d error %v", aid, err)
		return
	} else if arc == nil {
		log.Error("Service: GetQQCMCDocIDByBVID GetArcByAID %d 获取稿件为空", aid)
		err = archiveEcode.ArchiveNotExist
		return
	}
	qqQuery := &qqModel.UserContentListAdminQuery{
		Creater: strconv.FormatInt(arc.Author.Mid, 10),
	}
	if qqReply, _err := s.qqDAO.UserContentListAdmin(qqQuery); _err != nil {
		log.Error("Service: GetQQCMCDocIDByBVID UserContentListAdmin %d error %v", arc.Author.Mid, _err)
		err = _err
		return
	} else if qqReply.Status != 0 {
		log.Error("Service: GetQQCMCDocIDByBVID UserContentListAdmin %d request error %s", arc.Author.Mid, qqReply.MSG)
		err = ecode.QQCMCRequestError
		return
	} else if len(qqReply.Data) == 0 {
		log.Error("Service: GetQQCMCDocIDByBVID UserContentListAdmin %d 获取稿件列表为空", arc.Author.Mid)
		err = ecode.QQCMCArchiveNotFound
		return
	} else {
		for _, content := range qqReply.Data {
			_content := content
			if _content.SVID == bvid {
				docid = _content.IDocID
				return
			}
		}
		log.Error("Service: GetQQCMCDocIDByBVID UserContentListAdmin %d 未找到作者对应已推送稿件 %s", arc.Author.Mid, bvid)
		err = ecode.QQCMCArchiveNotFound
	}

	return
}

// QQ CMC End
///////////////////

///////////////////
// Blizzard Start

// GenerateWithdrawingBlizzard 生成暴雪推送下架models
func (s *Service) GenerateWithdrawingBlizzard(bvid string) (res *blizzardModel.VodAddReq, err error) {
	if bvid == "" {
		return
	}
	var (
		aid  int64
		arc  *archiveGRPC.Arc
		tags []*tagGRPC.Tag
	)
	if aid, err = util.BvToAv(bvid); err != nil {
		log.Error("Service: GenerateWithdrawingBlizzard %s BvToAv error %v", bvid, err)
		return
	}
	if arc, err = s.dao.GetArcByAID(aid); err != nil {
		log.Error("Service: GenerateWithdrawingBlizzard GetArcByAID(%d) error %v", aid, err)
		return
	} else if arc == nil {
		err = archiveEcode.ArchiveNotExist
		log.Error("Service: GenerateWithdrawingBlizzard GetArcByAID(%d) 不存在稿件", aid)
		return
	}
	// 获取稿件对应tags
	if tags, err = s.dao.GetTagsByAID(aid); err != nil {
		log.Error("Service: GenerateWithdrawingBlizzard GetTagsByAID(%d) Error (%v)", aid, err)
		return
	}
	if len(tags) == 0 {
		log.Error("Service: GenerateWithdrawingBlizzard 稿件tag数量为0")
		return
	}
	res = &blizzardModel.VodAddReq{
		BVID:        bvid,
		Page:        1,
		Category:    blizzardModel.DefaultVodAddCategory,
		Title:       arc.Title,
		Description: arc.Desc,
		Duration:    arc.Duration,
		Thumbnail:   arc.Pic,
		Stage:       tags[0].Name,
		Status:      blizzardModel.VodAddStatusWithdraw,
	}

	return
}

// DoBlizzardWithdrawings 进行暴雪稿件下架
func (s *Service) DoBlizzardWithdrawings(pushings []*blizzardModel.VodAddReq, details []*model.ArchivePushBatchDetail, histories []*model.ArchivePushBatchHistory) (resDetails []*model.ArchivePushBatchDetail, resHistories []*model.ArchivePushBatchHistory, err error) {
	if len(pushings) == 0 || len(details) == 0 {
		log.Warn("Service: DoBlizzardWithdrawings no valid data to push")
		return
	}
	lock := sync.RWMutex{}
	eg := errgroup.WithContext(context.Background())
	for _, _push := range pushings {
		push := _push
		eg.Go(func(_ context.Context) error {
			var detail *model.ArchivePushBatchDetail
			var history *model.ArchivePushBatchHistory
			var bvid string
			lock.Lock()
			for detailI := range details {
				if bvid, _ = util.AvToBv(details[detailI].AID); bvid == push.BVID {
					detail = details[detailI]
					detail.PushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
					for historyI := range histories {
						if histories[historyI].AID == details[detailI].AID {
							history = histories[historyI]
							history.NewPushStatus = api.ArchivePushDetailPushStatus_OUTER_FAIL
							break
						}
					}
					break
				}
			}
			if detail == nil {
				err = ecode.BatchDetailNotFound
				return err
			}
			lock.Unlock()
			if reply, perr := s.blizzardDAO.VodAdd(*push); perr != nil {
				log.Error("Service: DoBlizzardWithdrawings VodAdd(%+v) Error %v", *push, perr)
				return perr
			} else if reply.Status == 0 {
				lock.Lock()
				detail.ArchiveStatus = api.ArchiveStatus_WITHDRAW
				detail.PushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				if history != nil {
					history.NewArchiveStatus = api.ArchiveStatus_WITHDRAW
					history.NewPushStatus = api.ArchivePushDetailPushStatus_SUCCESS
				}
				lock.Unlock()
				return ecode.BlizzardRequestError
			} else {
				return errors.WithMessage(xecode.RequestErr, reply.MSG)
			}
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("Service: DoBlizzardWithdrawings eg Error (%v)", err)
	}
	resDetails = details
	resHistories = histories

	return
}

// Blizzard End
///////////////////

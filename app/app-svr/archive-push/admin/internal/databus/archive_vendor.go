package databus

import (
	"encoding/json"
	"fmt"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	tagGRPC "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	blizzardModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
	archiveEcode "go-gateway/app/app-svr/archive/ecode"
	"go-main/app/ep/hassan/mock/support/slice"
	"strconv"
	"strings"
	"time"
)

/////////////////
// QQ CMC Start

// checkAndPushArchivesQQCMC QQCMC 检查并推送需要推送的稿件
func (d *IDatabus) checkAndPushArchivesQQCMC(archive *model.ArchiveNotifyArchive) (err error) {
	var (
		bvid        string
		inWhiteList = false
		detail      *model.ArchivePushBatchDetail
		batch       *model.ArchivePushBatch
		vendorID    = model.DefaultVendors[0].ID
	)
	if bvid, err = util.AvToBv(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchives Got wrong AID(%d) (%v)", archive.AID, err)
		err = nil
		return
	}
	if inWhiteList, err = d.svc.CheckIfBVIDInWhiteList(vendorID, bvid, 0); err != nil {
		log.Error("Databus: checkAndPushArchivesQQCMC %+v error %v", archive, err)
		return
	} else if !inWhiteList {
		if d.svc.Cfg.Debug {
			log.Warn("Databus: checkAndPushArchivesQQCMC %+v 稿件不在白名单中", archive)
		}
		// 检查并进行作者维度推送
		if err = d.checkAndPushArchivesForAuthor(archive, vendorID); err != nil {
			log.Error("Databus: checkAndPushArchivesQQCMC checkAndPushArchivesForAuthor %+v error %v", archive, err)
		}
		return
	}

	if archiveDetails, _err := d.svc.GetBatchDetailsByBVIDs([]int64{model.DefaultVendors[0].ID, model.DefaultVendors[1].ID}, []string{bvid}, "mtime", true); _err != nil {
		err = _err
		log.Error("Databus: checkAndPushArchivesQQCMC GetArchivesByPushStatus Error (%v)", _err)
		return
	} else if len(archiveDetails) == 0 {
		if d.svc.Cfg.Debug {
			log.Info("Databus: checkAndPushArchivesQQCMC GetArchivesByPushStatus cannot find BVID (%s) 找不到稿件详情，将以单batch形式推送", bvid)
		}
		// 检查并进行作者维度推送
		if err = d.checkAndPushArchivesForAuthor(archive, vendorID); err != nil {
			log.Error("Databus: checkAndPushArchivesQQCMC checkAndPushArchivesForAuthor %+v error %v", archive, err)
		}
		return
	} else {
		detail = archiveDetails[0]
		if batches, _, _err := d.svc.GetBatchesByPage([]int64{detail.BatchID}, []int64{model.DefaultVendors[0].ID, model.DefaultVendors[1].ID}, "", 1, 1); _err != nil {
			log.Error("Databus: checkAndPushArchivesQQCMC GetBatchesByPage(%d) Error %v", detail.BatchID, _err)
			return
		} else if len(batches) == 0 {
			err = ecode.BatchNotFound
			return
		} else {
			batch = batches[0]
		}
	}

	if detail != nil && !(detail.ArchiveStatus == api.ArchiveStatus_OPEN && detail.PushStatus == api.ArchivePushDetailPushStatus_SUCCESS) {
		log.Info("Databus: checkAndPushArchivesQQCMC Start pushing detail(%d) BatchID(%d) BVID(%s)", detail.ID, detail.BatchID, bvid)
		var rawArchive *archiveGRPC.Arc
		if rawArchive, err = d.svc.GetArcByAID(archive.AID); err != nil {
			log.Error("Databus: checkAndPushArchivesQQCMC GetArcByAID(%d) Error (%v)", archive.AID, err)
			return
		} else if rawArchive == nil {
			log.Error("Databus: checkAndPushArchivesQQCMC GetArcByAID(%d) not found")
			err = archiveEcode.ArchiveNotExist
			return
		}
		arcMetadata := &model.ArchiveMetadataAll{
			Arc:  rawArchive,
			Tags: make(map[int64]string),
		}

		if valid, _err := d.svc.ValidateArchiveToPush(*rawArchive, vendorID, false); _err != nil {
			if xecode.EqualError(archiveEcode.ArchiveNotExist, _err) {
				// 未开放
				detail.ArchiveStatus = api.ArchiveStatus_NOT_OPEN
				detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
				batchDetails := []*model.ArchivePushBatchDetail{detail}
				if batchDetails, err = d.svc.UpdateBatchDetails(batchDetails); err != nil {
					log.Error("Databus: checkAndPushArchivesQQCMC UpdateBatchDetails Error (%v)", err)
					return
				}
			} else {
				log.Error("Databus: checkAndPushArchivesQQCMC ValidateArchiveToPush Error (%v)", _err)
				return _err
			}
		} else if !valid {
			if d.svc.Cfg.Debug {
				log.Warn("Databus: checkAndPushArchivesQQCMC ValidateArchiveToPush(%s) invalid", bvid)
			}
		} else {
			// 符合要求稿件
			log.Info("Databus: checkAndPushArchivesQQCMC(%s) turned to be valid, going to push", bvid)
			detail.ArchiveStatus = api.ArchiveStatus_OPEN
			detail.PushStatus = api.ArchivePushDetailPushStatus_UNKNOWN
			username := "system"
			uid := int64(0)

			tagNames := rawArchive.Tags
			if tags, _err := d.svc.GetTagsByAID(detail.AID); _err != nil {
				log.Error("Databus: checkAndPushArchivesQQCMC GetTagsByAID(%d)", detail.AID)
				return _err
			} else if len(tags) > 0 {
				tagNames = make([]string, 0, len(tags))
				for _, tag := range tags {
					tagNames = append(tagNames, tag.Name)
					arcMetadata.Tags[tag.Id] = tag.Name
				}
			}
			var arcMetaJSON []byte
			if arcMetaJSON, err = json.Marshal(arcMetadata); err != nil {
				log.Error("Databus: checkAndPushArchivesQQCMC Marshal(arcMetadata) Error (%v)", detail.AID, err)
				return err
			}
			detail.ArchiveDetails = string(arcMetaJSON)

			cmcReq := &qqModel.PushPGCAdminReq{
				SCreater:       strconv.FormatInt(rawArchive.Author.Mid, 10),
				SCreaterHeader: rawArchive.Author.Face,
				STitle:         archive.Title,
				SIMG:           rawArchive.Pic,
				SDESC:          rawArchive.Desc,
				SAuthor:        rawArchive.Author.Name,
				IType:          "0",
				ISubType:       "0",
				SOriginID:      bvid,
				SURL:           "",
				SCreated:       rawArchive.Ctime.Time().Format(model.DefaultTimeLayout),
				SCreatedOther:  rawArchive.Ctime.Time().Format(model.DefaultTimeLayout),
				STagsOther:     strings.Join(tagNames, ","),
				SSource:        qqModel.DefaultSSource,
				SVID:           bvid,
				ITime:          strconv.FormatInt(archive.Duration, 10),
				IFrom:          "11",
				SVideoSize:     fmt.Sprintf("%d*%d", rawArchive.Dimension.Width, rawArchive.Dimension.Height),
			}
			cmcPushings := []*qqModel.PushPGCAdminReq{cmcReq}
			batchHistory := &model.ArchivePushBatchHistory{
				BatchID:          batch.ID,
				AID:              detail.AID,
				BVID:             bvid,
				PushVendorID:     batch.PushVendorID,
				OldArchiveStatus: detail.ArchiveStatus,
				OldPushStatus:    detail.PushStatus,
				NewArchiveStatus: api.ArchiveStatus_OPEN,
				NewPushStatus:    api.ArchivePushDetailPushStatus_UNKNOWN,
				ArchiveDetails:   string(arcMetaJSON),
				Reason:           "稿件状态变更为可浏览",
				CUser:            username,
				CTime:            xtime.Time(time.Now().Unix()),
			}
			batchHistories := []*model.ArchivePushBatchHistory{batchHistory}
			if err = d.svc.AddBatchDetailActionLog(batchHistories, model.ActionPushUp, username, uid); err != nil {
				log.Error("Databus: checkAndPushArchivesQQCMC AddBatchDetailActionLog Error (%v)", err)
				return
			}
			batchDetails := []*model.ArchivePushBatchDetail{detail}
			if batchDetails, err = d.svc.UpdateBatchDetails(batchDetails); err != nil {
				log.Error("Databus: checkAndPushArchivesQQCMC UpdateBatchDetails Error (%v)", err)
				return
			}
			d.svc.TryQQCMCPushing(cmcPushings, batchDetails, batchHistories, username, uid)
		}
	}

	return
}

// QQ CMC End
/////////////////

/////////////////
// QQ TGL Start

// checkAndPushArchivesQQTGL QQ TGL 检查并推送需要推送的稿件
func (d *IDatabus) checkAndPushArchivesQQTGL(archive *model.ArchiveNotifyArchive) (err error) {
	var (
		bvid        string
		inWhiteList = false
		arc         *archiveGRPC.Arc
		detail      *model.ArchivePushBatchDetail
		batch       *model.ArchivePushBatch
		vendorID    = model.DefaultVendors[1].ID
	)
	if bvid, err = util.AvToBv(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchivesQQTGL Got wrong AID(%d) (%v)", archive.AID, err)
		err = nil
		return
	}
	if arc, err = d.svc.GetArcByAID(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchivesQQTGL GetArcByAID (%d) error %v", archive.AID, err)
		return
	}
	if inWhiteList, err = d.svc.CheckIfBVIDInWhiteList(vendorID, bvid, arc.Author.Mid); err != nil {
		log.Error("Databus: checkAndPushArchivesQQTGL %+v error %v", archive, err)
		return
	} else if !inWhiteList {
		if d.svc.Cfg.Debug {
			log.Warn("Databus: checkAndPushArchivesQQTGL %+v 稿件不在白名单中", archive)
		}
		if err = d.checkAndPushArchivesForAuthor(archive, vendorID); err != nil {
			log.Error("Databus: checkAndPushArchivesQQTGL checkAndPushArchivesForAuthor %+v error %v", archive, err)
		}
		return
	}

	if archiveDetails, _err := d.svc.GetBatchDetailsByBVIDs([]int64{vendorID}, []string{bvid}, "mtime", true); _err != nil {
		err = _err
		log.Error("Databus: checkAndPushArchivesQQTGL GetArchivesByPushStatus Error (%v)", _err)
		return
	} else if len(archiveDetails) == 0 {
		if d.svc.Cfg.Debug {
			log.Info("Databus: checkAndPushArchivesQQTGL GetArchivesByPushStatus cannot find BVID (%s) 找不到稿件详情，将以单batch形式推送", bvid)
		}
		if err = d.checkAndPushArchivesForAuthor(archive, vendorID); err != nil {
			log.Error("Databus: checkAndPushArchivesQQTGL checkAndPushArchivesForAuthor %+v error %v", archive, err)
		}
		return
	} else {
		detail = archiveDetails[0]
		if batches, _, _err := d.svc.GetBatchesByPage([]int64{detail.BatchID}, []int64{model.DefaultVendors[0].ID, model.DefaultVendors[1].ID}, "", 1, 1); _err != nil {
			log.Error("Databus: checkAndPushArchivesQQTGL GetBatchesByPage(%d) Error %v", detail.BatchID, _err)
			return
		} else if len(batches) == 0 {
			err = ecode.BatchNotFound
			return
		} else {
			batch = batches[0]
		}
	}

	if detail != nil && !(detail.ArchiveStatus == api.ArchiveStatus_OPEN && detail.PushStatus == api.ArchivePushDetailPushStatus_SUCCESS) {
		log.Info("Databus: checkAndPushArchivesQQTGL Start pushing detail(%d) BatchID(%d) BVID(%s)", detail.ID, detail.BatchID, bvid)
		var rawArchive *archiveGRPC.Arc
		if rawArchive, err = d.svc.GetArcByAID(archive.AID); err != nil {
			log.Error("Databus: checkAndPushArchivesQQTGL GetArcByAID(%d) Error (%v)", archive.AID, err)
			return
		} else if rawArchive == nil {
			log.Error("Databus: checkAndPushArchivesQQTGL GetArcByAID(%d) not found")
			err = archiveEcode.ArchiveNotExist
			return
		}
		arcMetadata := &model.ArchiveMetadataAll{
			Arc:  rawArchive,
			Tags: make(map[int64]string),
		}

		if valid, _err := d.svc.ValidateArchiveToPush(*rawArchive, vendorID, false); _err != nil {
			if xecode.EqualError(archiveEcode.ArchiveNotExist, _err) {
				// 未开放
				detail.ArchiveStatus = api.ArchiveStatus_NOT_OPEN
				detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
				batchDetails := []*model.ArchivePushBatchDetail{detail}
				if batchDetails, err = d.svc.UpdateBatchDetails(batchDetails); err != nil {
					log.Error("Databus: checkAndPushArchives.UpdateBatchDetails Error (%v)", err)
					return
				}
			} else {
				log.Error("Databus: checkAndPushArchivesQQTGL ValidateArchiveToPush Error (%v)", _err)
				return _err
			}
		} else if !valid {
			if d.svc.Cfg.Debug {
				log.Warn("Databus: checkAndPushArchivesQQTGL ValidateArchiveToPush(%s) invalid", bvid)
			}
		} else {
			// 符合要求稿件
			log.Info("Databus: checkAndPushArchivesQQTGL(%s) turned to be valid, going to push", bvid)
			detail.ArchiveStatus = api.ArchiveStatus_OPEN
			detail.PushStatus = api.ArchivePushDetailPushStatus_UNKNOWN
			username := "system"
			uid := int64(0)

			tagNames := rawArchive.Tags
			if tags, _err := d.svc.GetTagsByAID(detail.AID); _err != nil {
				log.Error("Databus: checkAndPushArchivesQQTGL GetTagsByAID(%d)", detail.AID)
				return _err
			} else if len(tags) > 0 {
				tagNames = make([]string, 0, len(tags))
				for _, tag := range tags {
					tagNames = append(tagNames, tag.Name)
					arcMetadata.Tags[tag.Id] = tag.Name
				}
			}
			var arcMetaJSON []byte
			if arcMetaJSON, err = json.Marshal(arcMetadata); err != nil {
				log.Error("Databus: checkAndPushArchivesQQTGL Marshal(arcMetadata) Error (%v)", detail.AID, err)
				return err
			}
			detail.ArchiveDetails = string(arcMetaJSON)

			req := &qqModel.ContributeVideoReq{
				Title:    rawArchive.Title,
				Summary:  rawArchive.Desc,
				Cover:    rawArchive.Pic,
				Author:   rawArchive.Author.Name,
				Duration: rawArchive.Duration,
				OuterVID: bvid,
				ExtTags:  strings.Join(tagNames, ","),
			}
			// 获取作者open id
			appkey := ""
			if appkey, err = d.svc.GetOauthAppKeyByVendorID(vendorID); err != nil {
				log.Error("Databus: checkAndPushArchivesQQTGL GetOauthAppKeyByVendorID(%d) error %v", vendorID, err)
				return
			}
			if openID, _err := d.svc.GetOpenIDByMID(rawArchive.Author.Mid, appkey); _err != nil {
				log.Error("Databus: checkAndPushArchivesQQTGL GetOpenIDByMID(%d, %s) error %v", rawArchive.Author.Mid, appkey, _err)
				err = _err
				return
			} else {
				req.OuterUser = openID
			}
			tglPushings := []*qqModel.ContributeVideoReq{req}
			batchHistory := &model.ArchivePushBatchHistory{
				BatchID:          batch.ID,
				AID:              detail.AID,
				BVID:             bvid,
				PushVendorID:     batch.PushVendorID,
				OldArchiveStatus: detail.ArchiveStatus,
				OldPushStatus:    detail.PushStatus,
				NewArchiveStatus: api.ArchiveStatus_OPEN,
				NewPushStatus:    api.ArchivePushDetailPushStatus_UNKNOWN,
				ArchiveDetails:   string(arcMetaJSON),
				Reason:           "稿件状态变更为可浏览",
				CUser:            username,
				CTime:            xtime.Time(time.Now().Unix()),
			}
			batchHistories := []*model.ArchivePushBatchHistory{batchHistory}
			if err = d.svc.AddBatchDetailActionLog(batchHistories, model.ActionPushUp, username, uid); err != nil {
				log.Error("Databus: checkAndPushArchivesQQTGL AddBatchDetailActionLog Error (%v)", err)
				return
			}
			batchDetails := []*model.ArchivePushBatchDetail{detail}
			if batchDetails, err = d.svc.UpdateBatchDetails(batchDetails); err != nil {
				log.Error("Databus: checkAndPushArchivesQQTGL UpdateBatchDetails Error (%v)", err)
				return
			}
			d.svc.TryQQTGLPushing(tglPushings, batchDetails, batchHistories, username, uid)
		}
	}

	return
}

// QQ TGL End
/////////////////

// ///////////////
// 暴雪嘉年华 Start
// checkAndPushArchivesBlizzard 暴雪嘉年华 检查并推送需要推送的稿件
func (d *IDatabus) checkAndPushArchivesBlizzard(archive *model.ArchiveNotifyArchive) (err error) {
	var (
		bvid        string
		inWhiteList = false
		detail      *model.ArchivePushBatchDetail
		batch       *model.ArchivePushBatch
		vendorID    = model.DefaultVendors[2].ID
		tagName     string
	)
	if bvid, err = util.AvToBv(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchivesBlizzard Got wrong AID(%d) (%v)", archive.AID, err)
		err = nil
		return
	}
	if inWhiteList, err = d.svc.CheckIfBVIDInWhiteList(vendorID, bvid, 0); err != nil {
		log.Error("Databus: checkAndPushArchivesBlizzard %+v error %v", archive, err)
		return
	} else if !inWhiteList {
		if d.svc.Cfg.Debug {
			log.Warn("Databus: checkAndPushArchivesBlizzard %+v 稿件不在白名单中", archive)
		}
		return
	}

	if archiveDetails, _err := d.svc.GetBatchDetailsByBVIDs([]int64{vendorID}, []string{bvid}, "mtime", true); _err != nil {
		err = _err
		log.Error("Databus: checkAndPushArchivesBlizzard GetArchivesByPushStatus Error (%v)", _err)
		return
	} else if len(archiveDetails) == 0 {
		if d.svc.Cfg.Debug {
			log.Info("Databus: checkAndPushArchivesBlizzard GetArchivesByPushStatus cannot find BVID (%s)", bvid)
		}
		return
	} else {
		detail = archiveDetails[0]
		if batches, _, _err := d.svc.GetBatchesByPage([]int64{detail.BatchID}, []int64{model.DefaultVendors[2].ID}, "", 1, 1); _err != nil {
			log.Error("Databus: checkAndPushArchivesBlizzard GetBatchesByPage(%d) Error %v", detail.BatchID, _err)
			return
		} else if len(batches) == 0 {
			err = ecode.BatchNotFound
			return
		} else {
			batch = batches[0]
		}
	}

	if detail != nil && !(detail.ArchiveStatus == api.ArchiveStatus_OPEN && detail.PushStatus == api.ArchivePushDetailPushStatus_SUCCESS) {
		log.Info("Databus: checkAndPushArchivesBlizzard Start pushing detail(%d) BatchID(%d) BVID(%s)", detail.ID, detail.BatchID, bvid)
		var rawArchive *archiveGRPC.Arc
		if rawArchive, err = d.svc.GetArcByAID(archive.AID); err != nil {
			log.Error("Databus: checkAndPushArchivesBlizzard GetArcByAID(%d) Error (%v)", archive.AID, err)
			return
		} else if rawArchive == nil {
			log.Error("Databus: checkAndPushArchivesBlizzard GetArcByAID(%d) not found")
			err = archiveEcode.ArchiveNotExist
			return
		}
		arcMetadata := &model.ArchiveMetadataAll{
			Arc:  rawArchive,
			Tags: make(map[int64]string),
		}

		if valid, _err := d.svc.ValidateArchiveToPush(*rawArchive, vendorID, false); _err != nil {
			if xecode.EqualError(archiveEcode.ArchiveNotExist, _err) {
				// 未开放
				detail.ArchiveStatus = api.ArchiveStatus_NOT_OPEN
				detail.PushStatus = api.ArchivePushDetailPushStatus_INNER_FAIL
				batchDetails := []*model.ArchivePushBatchDetail{detail}
				if batchDetails, err = d.svc.UpdateBatchDetails(batchDetails); err != nil {
					log.Error("Databus: checkAndPushArchivesBlizzard UpdateBatchDetails Error (%v)", err)
					return
				}
			} else {
				log.Error("Databus: checkAndPushArchivesBlizzard ValidateArchiveToPush Error (%v)", _err)
				return _err
			}
		} else if !valid {
			if d.svc.Cfg.Debug {
				log.Warn("Databus: checkAndPushArchivesBlizzard ValidateArchiveToPush(%s) invalid", bvid)
			}
		} else {
			// 符合要求稿件
			log.Info("Databus: checkAndPushArchivesBlizzard(%s) turned to be valid, going to push", bvid)
			detail.ArchiveStatus = api.ArchiveStatus_OPEN
			detail.PushStatus = api.ArchivePushDetailPushStatus_UNKNOWN
			username := "system"
			uid := int64(0)

			if tags, _err := d.svc.GetTagsByAID(detail.AID); _err != nil {
				log.Error("Databus: checkAndPushArchivesBlizzard GetTagsByAID(%d)", detail.AID)
				return _err
			} else if len(tags) == 0 {
				log.Error("Databus: ")
			} else {
				tagName = tags[0].Name
				for _, tag := range tags {
					arcMetadata.Tags[tag.Id] = tag.Name
				}
			}
			var arcMetaJSON []byte
			if arcMetaJSON, err = json.Marshal(arcMetadata); err != nil {
				log.Error("Databus: checkAndPushArchivesBlizzard Marshal(arcMetadata) Error (%v)", detail.AID, err)
				return err
			}
			detail.ArchiveDetails = string(arcMetaJSON)

			req := &blizzardModel.VodAddReq{
				BVID:        bvid,
				Page:        1,
				Category:    blizzardModel.DefaultVodAddCategory,
				Title:       rawArchive.Title,
				Description: rawArchive.Desc,
				Duration:    rawArchive.Duration,
				Thumbnail:   rawArchive.Pic,
				Stage:       tagName,
				Status:      blizzardModel.VodAddStatusPushUp,
			}
			pushings := []*blizzardModel.VodAddReq{req}
			batchHistory := &model.ArchivePushBatchHistory{
				BatchID:          batch.ID,
				AID:              detail.AID,
				BVID:             bvid,
				PushVendorID:     batch.PushVendorID,
				OldArchiveStatus: detail.ArchiveStatus,
				OldPushStatus:    detail.PushStatus,
				NewArchiveStatus: api.ArchiveStatus_OPEN,
				NewPushStatus:    api.ArchivePushDetailPushStatus_UNKNOWN,
				ArchiveDetails:   string(arcMetaJSON),
				Reason:           "稿件状态变更为可浏览",
				CUser:            username,
				CTime:            xtime.Time(time.Now().Unix()),
			}
			batchHistories := []*model.ArchivePushBatchHistory{batchHistory}
			if err = d.svc.AddBatchDetailActionLog(batchHistories, model.ActionPushUp, username, uid); err != nil {
				log.Error("Databus: checkAndPushArchivesBlizzard AddBatchDetailActionLog Error (%v)", err)
				return
			}
			batchDetails := []*model.ArchivePushBatchDetail{detail}
			if batchDetails, err = d.svc.UpdateBatchDetails(batchDetails); err != nil {
				log.Error("Databus: checkAndPushArchivesBlizzard UpdateBatchDetails Error (%v)", err)
				return
			}
			d.svc.TryBlizzardPushing(pushings, batchDetails, batchHistories, username, uid)
		}
	}

	return
}

// 暴雪嘉年华 End
/////////////////

// checkAndPushArchivesForAuthor 检查签约作者并自动推送新的稿件上架
func (d *IDatabus) checkAndPushArchivesForAuthor(archive *model.ArchiveNotifyArchive, vendorID int64) (err error) {
	var (
		bvid             string
		pushable         = false
		authorPush       *model.ArchivePushAuthorPushWithAuthors
		pushTags         = make([]string, 0)
		archiveTags      []*tagGRPC.Tag
		authorsMap       = make(map[int64]*model.ArchivePushAuthor)
		pushableAuthors  []*model.ArchivePushAuthor
		archivesMap      = make(map[int64][]*archiveGRPC.Arc)
		arc              *archiveGRPC.Arc
		toPushBatchesMap map[int64]*model.ArchivePushBatch
		toPushBVIDsMap   map[int64][]string
		batchIDs         = make([]int64, 0)
	)
	if bvid, err = util.AvToBv(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor Got wrong AID (%d) (%v)", archive.AID, err)
		err = nil
		return
	}
	if pushable, err = d.svc.CheckIfAuthorPushable(vendorID, archive.MID); err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor CheckIfBVIDInAuthorWhiteList (%d, %s) error %v", vendorID, bvid, err)
		return
	} else if !pushable {
		if d.svc.Cfg.Debug {
			log.Warn("Databus: checkAndPushArchivesForAuthor CheckIfAuthorInWhiteList (%d, %d) 作者不可被推送", vendorID, archive.MID)
		}
		return
	}

	// 获取已存在的作者推送
	if authorPush, err = d.svc.GetActiveAuthorPushByMID(vendorID, archive.MID); err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor GetActiveAuthorPushByMID (%d, %d) error %v", vendorID, archive.MID, err)
		return
	} else if authorPush == nil || authorPush.Status != int(api.AuthorPushStatus_EFFECTIVE) {
		log.Warn("Databus: checkAndPushArchivesForAuthor GetActiveAuthorPushByMID (%d, %d) 作者推送未激活", vendorID, archive.MID)
		return
	}

	// 检查tag是否有要求
	if authorPush.Tags != "" {
		pushTags = strings.Split(authorPush.Tags, ",")
	}
	if len(pushTags) > 0 {
		if archiveTags, err = d.svc.GetTagsByAID(archive.AID); err != nil {
			log.Error("Databus: checkAndPushArchivesForAuthor GetTagsByAID (%d) error %v", archive.AID, err)
			return
		} else if len(archiveTags) == 0 {
			log.Error("Databus: checkAndPushArchivesForAuthor GetTagsByAID (%d) 获取tag为空", archive.AID)
			return
		} else {
			valid := false
			for _, archiveTag := range archiveTags {
				if slice.Contains(pushTags, archiveTag.Name) {
					valid = true
					break
				}
			}
			if !valid {
				log.Warn("Databus: checkAndPushArchivesForAuthor 稿件(%d) 的tag没有符合需要push的", archive.AID)
				return ecode.ArchiveCannotBePushed
			}
		}
	}

	// 获取作者
	if pushableAuthors, err = d.svc.GetRawAuthorsByUser(vendorID, archive.MID, ""); err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor GetRawAuthorsByUser (%d, %d, '') error %v", vendorID, archive.MID, err)
		return
	} else if len(pushableAuthors) == 0 {
		log.Warn("Databus: checkAndPushArchivesForAuthor GetRawAuthorsByUser (%d, %d, '') 没有找到有效作者", vendorID, archive.MID)
		err = ecode.AuthorNotFound
		return
	}
	authorsMap[pushableAuthors[0].ID] = pushableAuthors[0]
	archivesMap[pushableAuthors[0].ID] = make([]*archiveGRPC.Arc, 0)

	// 获取稿件
	if arc, err = d.svc.GetArcByAID(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor GetArcByAID (%d) error %v", archive.AID, err)
		return
	} else if arc == nil {
		log.Error("Databus: checkAndPushArchivesForAuthor GetArcByAID (%d) 获取稿件为空", archive.AID)
		err = archiveEcode.ArchiveNotExist
		return
	}
	archivesMap[pushableAuthors[0].ID] = append(archivesMap[pushableAuthors[0].ID], arc)

	// 检查稿件是否可上架
	var rawArchive *archiveGRPC.Arc
	if rawArchive, err = d.svc.GetArcByAID(archive.AID); err != nil {
		log.Error("Databus: checkAndPushArchivesQQTGL GetArcByAID (%d) Error (%v)", archive.AID, err)
		return
	} else if rawArchive == nil {
		log.Error("Databus: checkAndPushArchivesQQTGL GetArcByAID (%d) not found", archive.AID)
		err = archiveEcode.ArchiveNotExist
		return
	}
	if valid, _err := d.svc.ValidateArchiveToPush(*rawArchive, vendorID, true); _err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor ValidateArchiveToPush Error (%v)", _err)
		return _err
	} else if !valid {
		if d.svc.Cfg.Debug {
			log.Warn("Databus: checkAndPushArchivesForAuthor ValidateArchiveToPush (%s) invalid", bvid)
		}
		return ecode.ArchiveCannotBePushed
	}

	// 根据作者生成batch与bvid列表
	if toPushBatchesMap, toPushBVIDsMap, err = d.svc.GenerateAuthorPushingMaps(vendorID, archivesMap, "system"); err != nil {
		log.Error("Databus: checkAndPushArchivesForAuthor GenerateAuthorPushingMaps (%d, %+v) error %v", vendorID, archivesMap, err)
		return
	}

	// 将bvid放入作者稿件白名单
	for authorID := range toPushBVIDsMap {
		if len(toPushBVIDsMap[authorID]) > 0 {
			if author, exists := authorsMap[authorID]; exists {
				if _err := d.svc.PutAuthorBVIDsForWhiteList(vendorID, author.MID, toPushBVIDsMap[authorID]); _err != nil {
					log.Error("Service: checkAndPushArchivesForAuthor PutAuthorBVIDsForWhiteList (%d, %d, %v) error %v", vendorID, author.MID, toPushBVIDsMap[authorID], _err)
				}
			}
		}
	}

	// 根据作者推送稿件批次
	if batchIDs, err = d.svc.DoPushAuthorPushes(authorPush.ID, toPushBatchesMap, toPushBVIDsMap, authorPush.DelayMinutes, "system", 0); err != nil {
		log.Error("Service: checkAndPushArchivesForAuthor DoPushAuthorPushes (%d, %v, %v, %d) error %v", authorPush.ID, toPushBatchesMap, toPushBVIDsMap, authorPush.DelayMinutes, err)
		return
	}

	// 行为日志
	if _err := d.svc.AddAuthorPushAuditLog(vendorID, authorPush.Tags, authorPush.DelayMinutes, authorPush.ID, pushableAuthors, batchIDs, "system", 0); _err != nil {
		log.Error("Service: checkAndPushArchivesForAuthor AddAuthorPushAuditLog (%d, %s, %d, %d, %v, %v, %s, %d) error %v", vendorID, pushTags, authorPush.DelayMinutes, authorPush.ID, pushableAuthors, batchIDs, "system", 0, _err)
	}

	return
}

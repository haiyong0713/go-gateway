package http

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	xtime "go-common/library/time"
	"strconv"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
)

func batches(ctx *bm.Context) {
	batchesReq := &struct {
		IDs          string `form:"ids" json:"ids"`
		PushVendorID int64  `form:"pushVendorId" json:"pushVendorId"`
		CUser        string `form:"cuser" json:"cuser"`
		Pn           int    `form:"pn" json:"pn" validate:"min=1" default:"1"`
		Ps           int    `form:"ps" json:"ps" validate:"min=1" default:"20"`
	}{}
	if err := ctx.BindWith(batchesReq, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if batchesReq.Pn == 0 {
		batchesReq.Pn = 1
	}
	if batchesReq.Ps == 0 {
		batchesReq.Ps = 20
	}

	ids := make([]int64, 0)
	if len(batchesReq.IDs) > 0 {
		idStrs := strings.Split(batchesReq.IDs, ",")
		for _, idStr := range idStrs {
			if id, _err := strconv.ParseInt(idStr, 10, 64); _err == nil {
				ids = append(ids, id)
			}
		}
	}
	pushVendorIDs := make([]int64, 0)
	if batchesReq.PushVendorID != 0 {
		pushVendorIDs = append(pushVendorIDs, batchesReq.PushVendorID)
	}
	batchList, total, err := svc.GetBatchesByPage(ids, pushVendorIDs, batchesReq.CUser, batchesReq.Pn, batchesReq.Ps)
	type batchForRes struct {
		*model.ArchivePushBatchX
		PushVendorName string `json:"pushVendorName"`
	}
	res := struct {
		Items []*batchForRes `json:"items"`
		Pager *model.Page    `json:"pager"`
	}{
		Pager: &model.Page{
			Num:   batchesReq.Pn,
			Size:  batchesReq.Ps,
			Total: total,
		},
	}
	res.Items = make([]*batchForRes, 0)
	for _, batch := range batchList {
		toAppendBatch := &batchForRes{
			ArchivePushBatchX: model.ArchivePushBatchToX(*batch),
		}
		if vendor, err := svc.GetVendorByID(batch.PushVendorID); err != nil || vendor.ID == 0 {
			log.Error("HTTP: batchByID GetVendorByID(%d) not exists or error %v", batch.PushVendorID, err)
		} else {
			toAppendBatch.PushVendorName = vendor.Name
		}
		res.Items = append(res.Items, toAppendBatch)
	}
	ctx.JSON(res, err)
}

func batchPush(ctx *bm.Context) {
	batchPushReq := &struct {
		FileURL      string `form:"fileUrl" json:"fileUrl" validate:"required"`
		PushType     string `form:"pushType" json:"pushType" validate:"required"`
		PushVendorID int64  `form:"pushVendorId" json:"pushVendorId" validate:"required"`
	}{}
	if err := ctx.BindWith(batchPushReq, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	username, _ := util.UserInfo(ctx)
	batchToPush := &model.ArchivePushBatch{
		FileURL:      batchPushReq.FileURL,
		PushType:     1,
		PushVendorID: batchPushReq.PushVendorID,
		CUser:        username,
		CTime:        xtime.Time(time.Now().Unix()),
	}
	batchID, err := svc.BatchPush(ctx, batchToPush)
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	reply := &struct {
		ID int64 `json:"id"`
	}{ID: batchID}
	ctx.JSON(reply, err)
}

func batchByID(ctx *bm.Context) {
	var err error
	var batchID int64
	batchIDStr, exists := ctx.Params.Get("id")
	if !exists {
		err = errors.WithMessage(xecode.RequestErr, "Missing Batch ID")
	} else {
		batchID, err = strconv.ParseInt(batchIDStr, 10, 64)
	}
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}

	type batchByIDRes struct {
		*model.ArchivePushBatchX
		PushVendorName string   `json:"pushVendorName"`
		PushTotal      int      `json:"pushTotal"`
		PushSuccess    []string `json:"pushSuccess"`
		PushFail       []string `json:"pushFail"`
		PushPushing    []string `json:"pushPushing"`
	}

	res := batchByIDRes{
		PushVendorName: "",
		PushTotal:      0,
		PushSuccess:    make([]string, 0),
		PushFail:       make([]string, 0),
		PushPushing:    make([]string, 0),
	}
	var archivePushBatch *model.ArchivePushBatch
	batchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if archivePushBatch, batchDetails, err = svc.GetBatchDetailsByBatchID(batchID); err != nil {
		log.Error("archive-push-admin.http.batchByID.GetBatchDetailsByBatchID(%d) Error(%v)", batchID, err)
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	} else if archivePushBatch == nil || archivePushBatch.ID == 0 {
		err = ecode.BatchNotFound
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	res.ArchivePushBatchX = model.ArchivePushBatchToX(*archivePushBatch)
	if vendor, _err := svc.GetVendorByID(archivePushBatch.PushVendorID); _err != nil {
		err = _err
		log.Error("HTTP: batchByID GetVendorByID(%d) not exists or error %v", archivePushBatch.PushVendorID, err)
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	} else {
		res.PushVendorName = vendor.Name
	}
	for _, detail := range batchDetails {
		bvid := ""
		res.PushTotal++
		if detail.AID == 0 {
			res.PushFail = append(res.PushFail, detail.ArchiveDetails)
			continue
		}
		if bvid, err = util.AvToBv(detail.AID); err != nil {
			log.Error("archive-push-admin.http.batchByID.AvToBv(%d) Error (%v)", detail.AID, err)
			ctx.JSON(nil, ecode.AVBVIDConvertingError)
			ctx.Abort()
			return
		}
		switch detail.PushStatus {
		case api.ArchivePushDetailPushStatus_SUCCESS:
			res.PushSuccess = append(res.PushSuccess, bvid)
			break
		case api.ArchivePushDetailPushStatus_INNER_FAIL:
			res.PushFail = append(res.PushFail, bvid)
			break
		case api.ArchivePushDetailPushStatus_OUTER_FAIL:
			if detail.ArchiveStatus == api.ArchiveStatus_WITHDRAW {
				res.PushSuccess = append(res.PushSuccess, bvid)
			} else {
				res.PushFail = append(res.PushFail, bvid)
			}
		case api.ArchivePushDetailPushStatus_UNKNOWN:
			res.PushPushing = append(res.PushPushing, bvid)
			break
		}
	}
	if res.CTime.Time().After(time.Now()) {
		res.PushStatus = api.ArchivePushBatchPushStatus_Enum_name[int32(api.ArchivePushBatchPushStatus_TO_PUSH)]
	} else if len(res.PushPushing) > 0 {
		res.PushStatus = api.ArchivePushBatchPushStatus_Enum_name[int32(api.ArchivePushBatchPushStatus_PUSHING)]
	} else if len(res.PushSuccess) == 0 {
		res.PushStatus = api.ArchivePushBatchPushStatus_Enum_name[int32(api.ArchivePushBatchPushStatus_FAIL)]
	} else if len(res.PushFail) == 0 {
		res.PushStatus = api.ArchivePushBatchPushStatus_Enum_name[int32(api.ArchivePushBatchPushStatus_SUCCESS)]
	} else {
		res.PushStatus = api.ArchivePushBatchPushStatus_Enum_name[int32(api.ArchivePushBatchPushStatus_FAIL_PARTIAL)]
	}

	ctx.JSON(res, err)
}

// batchExportByID 导出当前的批次推送详情
func batchExportByID(ctx *bm.Context) {
	var err error
	var batchID int64
	batchIDStr, exists := ctx.Params.Get("id")
	if !exists {
		err = errors.WithMessage(xecode.RequestErr, "Missing Batch ID")
	} else {
		batchID, err = strconv.ParseInt(batchIDStr, 10, 64)
	}
	if batchID <= 0 {
		err = errors.WithMessage(xecode.RequestErr, "批次ID不合法")
	}
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}

	batch := &model.ArchivePushBatch{}
	batchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if batch, batchDetails, err = svc.GetBatchDetailsByBatchID(batchID); err != nil {
		log.Error("archive-push-admin.http.batchByID.GetBatchDetailsByBatchID(%d) Error(%v)", batchID, err)
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	} else if batch == nil || batch.ID == 0 {
		ctx.JSON(nil, ecode.BatchNotFound)
		ctx.Abort()
		return
	}

	// export CSV
	exportCSVForBatchDetails(ctx, batchID, batchDetails)
}

// initialBatchExportByID 导出批次初次推送详情
func initialBatchExportByID(ctx *bm.Context) {
	var err error
	var batchID int64
	batchIDStr, exists := ctx.Params.Get("id")
	if !exists {
		err = errors.WithMessage(xecode.RequestErr, "Missing Batch ID")
	} else {
		batchID, err = strconv.ParseInt(batchIDStr, 10, 64)
	}
	if batchID <= 0 {
		err = errors.WithMessage(xecode.RequestErr, "批次ID不合法")
	}
	if err != nil {
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}

	batch := &model.ArchivePushBatch{}
	batchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if batch, batchDetails, err = svc.GetBatchDetailsByBatchID(batchID); err != nil {
		log.Error("archive-push-admin.http.initialBatchExportByID.GetBatchDetailsByBatchID(%d) Error(%v)", batchID, err)
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	} else if batch == nil || batch.ID == 0 {
		ctx.JSON(nil, ecode.BatchNotFound)
		ctx.Abort()
		return
	}
	batchHistories := make(map[string]*model.ArchivePushBatchHistory, 0)
	if batchHistories, err = svc.GetBatchHistoryByTime(batchID, int64(batch.CTime)); err != nil {
		log.Error("archive-push-admin.http.initialBatchExportByID.GetBatchHistory(%d).GetBatchHistory Error(%v)", batchID, err)
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	for bvid := range batchHistories {
		if batchHistories[bvid] == nil {
			continue
		}
		history := batchHistories[bvid]
		for batchDetailI := range batchDetails {
			if batchDetails[batchDetailI].AID == 0 && batchDetails[batchDetailI].ArchiveDetails == history.BVID {
				batchDetails[batchDetailI].ArchiveStatus = history.NewArchiveStatus
				batchDetails[batchDetailI].PushStatus = history.NewPushStatus
				break
			} else if batchDetails[batchDetailI].AID != 0 && batchDetails[batchDetailI].AID == history.AID {
				batchDetails[batchDetailI].ArchiveStatus = history.NewArchiveStatus
				batchDetails[batchDetailI].PushStatus = history.NewPushStatus
				break
			}
		}
	}

	// export CSV
	exportCSVForBatchDetails(ctx, batchID, batchDetails)
}

// exportCSVForBatchDetails 导出推送详情的CSV文件
func exportCSVForBatchDetails(ctx *bm.Context, batchID int64, batchDetails []*model.ArchivePushBatchDetail) {
	var err error
	fileName := fmt.Sprintf(svc.GetBatchExportFilenameFormat(), batchID, time.Now().Format("2006-01-02 15:04:05"))
	columns := svc.GetBatchExportColumns()
	b := &bytes.Buffer{}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	if err = wr.Write(svc.GetBatchExportTitles()); err != nil {
		log.Error("archive-push-admin.http.exportCSVForBatchDetails.Write Error (%v)", err)
		ctx.JSON(nil, err)
		ctx.Abort()
		return
	}
	var bvid string
	for _, detail := range batchDetails {
		if detail.AID == 0 {
			bvid = detail.ArchiveDetails
		} else if bvid, err = util.AvToBv(detail.AID); err != nil {
			log.Error("archive-push-admin.http.exportCSVForBatchDetails.AvToBv(%d) Error (%v)", detail.AID, err)
			ctx.JSON(nil, err)
			ctx.Abort()
			return
		}
		exportRow := make([]string, 0)
		for _, col := range columns {
			switch col {
			case "BVID":
				exportRow = append(exportRow, bvid)
				break
			case "ID":
				exportRow = append(exportRow, strconv.FormatInt(detail.ID, 10))
				break
			case "PushStatus":
				exportRow = append(exportRow, api.ArchivePushDetailPushStatus_Enum_name[int32(detail.PushStatus)])
				break
			case "ArchiveStatus":
				exportRow = append(exportRow, api.ArchiveStatus_Enum_name[int32(detail.ArchiveStatus)])
				break
			case "Reason":
				exportRow = append(exportRow, model.GetReasonForPushFailure(detail.ArchiveStatus, detail.PushStatus))
			}
		}
		if err = wr.Write(exportRow); err != nil {
			log.Error("archive-push-admin.http.exportCSVForBatchDetails.Write Error (%v)", err)
		}
	}
	wr.Flush()
	ctx.Writer.Header().Set("Content-Type", "text/csv")
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	tet := b.String()
	ctx.String(200, tet)
}

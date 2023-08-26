package model

import (
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/archive-push/admin/api"
)

const (
	ActionPushUp   = "PUSH_UP"
	ActionPushDown = "PUSH_DOWN"
	ActionBackflow = "BACKFLOW"

	PushTypeBVID   = 1
	PushTypeAuthor = 2
)

// ArchivePushBatchX 稿件推送批次信息用于数据交换
type ArchivePushBatchX struct {
	ID           int64      `json:"id"`
	PushType     string     `json:"pushType" validate:"required"`
	PushVendorID int64      `json:"pushVendorId" validate:"required"`
	FileURL      string     `json:"fileUrl" validate:"required"`
	PushStatus   string     `json:"pushStatus"`
	CUser        string     `json:"cuser"`
	CTime        xtime.Time `json:"ctime"`
	MUser        string     `json:"muser"`
	MTime        xtime.Time `json:"mtime"`
}

// ArchivePushBatch 稿件推送批次信息ORM
type ArchivePushBatch struct {
	ID           int64                               `json:"id" gorm:"column:id"`
	PushType     int32                               `json:"pushType" gorm:"column:push_type"`
	PushVendorID int64                               `json:"pushVendorId" gorm:"column:push_vendor_id"`
	FileURL      string                              `json:"fileUrl" gorm:"column:file_url"`
	PushStatus   api.ArchivePushBatchPushStatus_Enum `json:"pushStatus" gorm:"-"`
	CUser        string                              `json:"cuser" gorm:"column:cuser"`
	CTime        xtime.Time                          `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	MUser        string                              `json:"muser" gorm:"column:muser"`
	MTime        xtime.Time                          `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (t *ArchivePushBatch) TableName() string {
	return "archive_push_batch"
}

func ArchivePushBatchToX(origin ArchivePushBatch) *ArchivePushBatchX {
	return &ArchivePushBatchX{
		ID:           origin.ID,
		PushType:     api.ArchivePushType_Enum_name[origin.PushType],
		PushVendorID: origin.PushVendorID,
		FileURL:      origin.FileURL,
		PushStatus:   api.ArchivePushBatchPushStatus_Enum_name[int32(origin.PushStatus)],
		CUser:        origin.CUser,
		CTime:        origin.CTime,
		MUser:        origin.MUser,
		MTime:        origin.MTime,
	}
}

func ArchivePushBatchFromX(origin ArchivePushBatchX) *ArchivePushBatch {
	return &ArchivePushBatch{
		ID:           origin.ID,
		PushType:     api.ArchivePushType_Enum_value[origin.PushType],
		PushVendorID: origin.PushVendorID,
		FileURL:      origin.FileURL,
		PushStatus:   api.ArchivePushBatchPushStatus_Enum(api.ArchivePushBatchPushStatus_Enum_value[origin.PushStatus]),
		CUser:        origin.CUser,
		CTime:        origin.CTime,
		MUser:        origin.MUser,
		MTime:        origin.MTime,
	}
}

// ArchivePushBatchDetailX 稿件推送批次详细信息用于数据交换
type ArchivePushBatchDetailX struct {
	ID             int64      `json:"id"`
	BatchID        int64      `json:"batchId"`
	AID            int64      `json:"aid"`
	ArchiveStatus  int        `json:"archiveStatus"`
	PushStatus     string     `json:"pushStatus"`
	ArchiveDetails string     `json:"archiveDetails"`
	VendorID       int64      `json:"vendorId"`
	CUser          string     `json:"cuser"`
	CTime          xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05"`
	MUser          string     `json:"muser"`
	MTime          xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05"`
}

// ArchivePushBatchDetail 稿件推送批次详细信息ORM
type ArchivePushBatchDetail struct {
	ID             int64                                `json:"id" gorm:"column:id"`
	BatchID        int64                                `json:"batchId" gorm:"column:batch_id"`
	AID            int64                                `json:"aid" gorm:"column:aid"`
	ArchiveStatus  api.ArchiveStatus_Enum               `json:"archiveStatus" gorm:"column:archive_status"`
	PushStatus     api.ArchivePushDetailPushStatus_Enum `json:"pushStatus" gorm:"column:push_status"`
	ArchiveDetails string                               `json:"archiveDetails" gorm:"column:archive_details"`
	CUser          string                               `json:"cuser" gorm:"column:cuser"`
	CTime          xtime.Time                           `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	MUser          string                               `json:"muser" gorm:"column:muser"`
	MTime          xtime.Time                           `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (t *ArchivePushBatchDetail) TableName() string {
	return "archive_push_batch_detail"
}

type ArchivePushBatchDetailWithVendor struct {
	*ArchivePushBatchDetail
	VendorID int64 `json:"vendorId" gorm:"column:push_vendor_id"`
	PushType int32 `json:"pushType" gorm:"column:push_type"`
}

// ArchivePushBatchHistory 稿件推送批次历史详细信息
type ArchivePushBatchHistory struct {
	BatchID          int64                                `json:"batchId"`
	AID              int64                                `json:"aid"`
	BVID             string                               `json:"bvid"`
	PushVendorID     int64                                `json:"pushVendorId"`
	OldArchiveStatus api.ArchiveStatus_Enum               `json:"oldArchiveStatus"`
	OldPushStatus    api.ArchivePushDetailPushStatus_Enum `json:"oldPushStatus"`
	NewArchiveStatus api.ArchiveStatus_Enum               `json:"newArchiveStatus"`
	NewPushStatus    api.ArchivePushDetailPushStatus_Enum `json:"newPushStatus"`
	ArchiveDetails   string                               `json:"archiveDetails"`
	Reason           string                               `json:"reason"`
	CUser            string                               `json:"cuser"`
	CTime            xtime.Time                           `json:"ctime"`
}

// GetReasonForPushFailure 获取失败情况下的状态说明
func GetReasonForPushFailure(archiveStatus api.ArchiveStatus_Enum, pushStatus api.ArchivePushDetailPushStatus_Enum) string {
	switch archiveStatus {
	case api.ArchiveStatus_NOT_EXISTS:
		return "稿件不存在"
	case api.ArchiveStatus_NOT_OPEN:
		return "稿件已失效"
	case api.ArchiveStatus_FORMAT_INVALID:
		return "稿件格式不合要求"
	case api.ArchiveStatus_WITHDRAW:
		if pushStatus == api.ArchivePushDetailPushStatus_SUCCESS {
			return "稿件推送下架成功"
		} else if pushStatus == api.ArchivePushDetailPushStatus_UNKNOWN {
			return "稿件推送下架中"
		} else {
			return "稿件推送下架失败"
		}
	}
	switch pushStatus {
	case api.ArchivePushDetailPushStatus_SUCCESS:
		return "推送成功"
	case api.ArchivePushDetailPushStatus_UNKNOWN:
		return "推送中"
	case api.ArchivePushDetailPushStatus_OUTER_FAIL:
		return "与外部内容中台交互失败"
	}
	return "其他"
}

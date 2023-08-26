package model

import (
	"fmt"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/ecode"
	"strconv"
)

// ArchivePushAuthor 稿件推送作者ORM
type ArchivePushAuthor struct {
	ID                  int64                              `json:"id" gorm:"column:id"`
	MID                 int64                              `json:"mid" gorm:"column:mid"`
	OpenID              string                             `json:"openId" gorm:"column:open_id"`
	Nickname            string                             `json:"nickname" gorm:"column:nickname"`
	PushVendorID        int64                              `json:"vendorId" gorm:"column:vendor_id"`
	OuterID             string                             `json:"outerId" gorm:"column:outer_id"`
	AuthorizationStatus api.AuthorAuthorizationStatus_Enum `json:"authorizationStatus" gorm:"column:authorization_status"`
	AuthorizationTime   xtime.Time                         `json:"authorizationTime" gorm:"column:authorization_time"`
	AuthorizationSID    int64                              `json:"authorizationSid" gorm:"column:authorization_sid"`
	BindStatus          api.AuthorBindStatus_Enum          `json:"bindStatus" gorm:"column:bind_status"`
	BindTime            xtime.Time                         `json:"bindTime" gorm:"column:bind_time"`
	VerificationStatus  api.AuthorVerificationStatus_Enum  `json:"verificationStatus" gorm:"column:verification_status"`
	VerificationTime    xtime.Time                         `json:"verificationTime" gorm:"column:verification_time"`
	CUser               string                             `json:"cuser" gorm:"column:cuser"`
	CTime               xtime.Time                         `json:"ctime" gorm:"column:ctime"`
	MUser               string                             `json:"muser" gorm:"column:muser"`
	MTime               xtime.Time                         `json:"mtime" gorm:"column:mtime"`
	IsDeprecated        int                                `json:"isDeprecated" gorm:"column:is_deprecated"`
}

func (t *ArchivePushAuthor) TableName() string {
	return "archive_push_author"
}

// ArchivePushAuthorX 稿件作者，用于数据交换
type ArchivePushAuthorX struct {
	ArchivePushAuthor
	AuthorizationStatus string `json:"authorizationStatus"`
	BindStatus          string `json:"bindStatus"`
	VerificationStatus  string `json:"verificationStatus"`
	PushVendorName      string `json:"pushVendorName"`
	Reason              string `json:"reason"`
}

// ArchivePushAuthorWithBVIDs 稿件作者与其稿件BVIDs，主要用于放到白名单
type ArchivePushAuthorWithBVIDs struct {
	ArchivePushAuthor
	BVIDs []string
}

type SyncAuthorBindingReq struct {
	VendorID   int64  `json:"vendorId" form:"vendorId" validate:"required"`
	BOpenID    string `json:"bOpenId" form:"bOpenId"`
	OOpenID    string `json:"oOpenId" form:"oOpenId"`
	Action     string `json:"action" form:"action"`
	ActionTime string `json:"actionTime" form:"actionTime"`
	ActionMsg  string `json:"actionMsg" form:"actionMsg"`
}

type SyncAuthorAuthorizationReq struct {
	VendorID int64 `json:"vendorId" form:"vendorId" validate:"required"`
	MID      int64 `json:"mid" form:"mid"`
}

// ArchivePushAuthorPush 稿件作者推送ORM
type ArchivePushAuthorPush struct {
	ID             int64      `json:"id" gorm:"column:id"`
	VendorID       int64      `json:"vendorId" gorm:"column:vendor_id"`
	Tags           string     `json:"tags" gorm:"column:tags"`
	DelayMinutes   int32      `json:"delayMinutes" gorm:"column:delay_minutes"`
	Status         int        `json:"status" gorm:"column:status"`
	PushConditions string     `json:"pushConditions" gorm:"column:push_conditions"`
	CUser          string     `json:"cuser" gorm:"column:cuser"`
	CTime          xtime.Time `json:"ctime" gorm:"column:ctime"`
	MUser          string     `json:"muser" gorm:"column:muser"`
	MTime          xtime.Time `json:"mtime" gorm:"column:mtime"`
	IsDeprecated   int        `json:"isDeprecated" gorm:"column:is_deprecated"`
}

func (t *ArchivePushAuthorPush) TableName() string {
	return "archive_push_author_push"
}

// ArchivePushAuthorPushX 稿件作者推送，用于数据交换
type ArchivePushAuthorPushX struct {
	ArchivePushAuthorPush
	Status         string                            `json:"status"`
	PushConditions []*ArchivePushAuthorPushCondition `json:"pushConditions"`
}

type ArchivePushAuthorPushWithAuthors struct {
	ArchivePushAuthorPush
	Authors []*ArchivePushAuthor `json:"authors"`
}

// AmisCondition 作者推送条件
type AmisCondition struct {
	Conjunction AmisConditionConjunction `json:"conjunction,omitempty"`
	Left        *AmisConditionLeft       `json:"left,omitempty"`
	Op          AmisConditionOp          `json:"op,omitempty"`
	Value       string                   `json:"right,omitempty"`
	Children    []*AmisCondition         `json:"children,omitempty"`
}

// AmisConditionLeft 作者推送条件子结构
type AmisConditionLeft struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Field string `json:"field"`
	Value string `json:"value"`
}

// AmisConditionConjunction 作者推送条件逻辑关系
type AmisConditionConjunction string

const AmisConditionConjunctionAnd AmisConditionConjunction = "and"
const AmisConditionConjunctionOr AmisConditionConjunction = "or"

// AmisConditionOp 作者推送条件操作符
type AmisConditionOp string

const ArchivePushAuthorPushConditionOpSelectEquals ArchivePushAuthorPushConditionOp = "select_equals"

type ArchivePushAuthorPushCondition struct {
	Type  ArchivePushAuthorPushConditionType `json:"type"`
	Op    ArchivePushAuthorPushConditionOp   `json:"op"`
	Value bool                               `json:"value"`
}

type ArchivePushAuthorPushConditionType string

const ArchivePushAuthorPushConditionTypeAuthorized ArchivePushAuthorPushConditionType = "authorized"
const ArchivePushAuthorPushConditionTypeBinded ArchivePushAuthorPushConditionType = "binded"
const ArchivePushAuthorPushConditionTypeVerified ArchivePushAuthorPushConditionType = "verified"

type ArchivePushAuthorPushConditionOp string

const ArchivePushAuthorPushConditionOpEquals ArchivePushAuthorPushConditionOp = "equals"
const ArchivePushAuthorPushConditionOpNotEquals ArchivePushAuthorPushConditionOp = "not_equals"

// ArchivePushBatchAuthorPushRel 稿件作者推送与推送批次关系ORM
type ArchivePushBatchAuthorPushRel struct {
	ID           int64      `json:"id" gorm:"column:id"`
	AuthorPushID int64      `json:"authorPushId" gorm:"column:author_push_id"`
	AuthorID     int64      `json:"authorId" gorm:"column:author_id"`
	BatchID      int64      `json:"batchId" gorm:"column:batch_id"`
	CUser        string     `json:"cuser" gorm:"column:cuser"`
	CTime        xtime.Time `json:"ctime" gorm:"column:ctime"`
	MUser        string     `json:"muser" gorm:"column:muser"`
	MTime        xtime.Time `json:"mtime" gorm:"column:mtime"`
	IsDeprecated int        `json:"isDeprecated" gorm:"column:is_deprecated"`
}

// ArchivePushBatchAuthorPushRelWithAuthorAndBatch 稿件作者推送与推送批次关系
type ArchivePushBatchAuthorPushRelWithAuthorAndBatch struct {
	ArchivePushBatchAuthorPushRel
	VendorID       int64  `json:"vendorId" gorm:"column:vendorId"`
	AuthorNickname string `json:"authorNickname" gorm:"column:authorNickname"`
	BatchPushType  int32  `json:"batchPushType" gorm:"column:batchPushType"`
}

func (t *ArchivePushBatchAuthorPushRel) TableName() string {
	return "archive_push_batch_author_push_rels"
}

// ArchivePushAuthorPushFull 稿件作者推送，用于前端展示
type ArchivePushAuthorPushFull struct {
	ArchivePushAuthorPush
	Status             string                     `json:"status"`
	AuthorsWithBatches map[int64]*AuthorWithBatch `json:"authorsWithBatches"`
	VendorName         string                     `json:"vendorName"`
	Authorized         bool                       `json:"authorized"`
	Binded             bool                       `json:"binded"`
	Verified           bool                       `json:"verified"`
}

type AuthorWithBatch struct {
	BatchAuthorPushRelID int64  `json:"batchAuthorPushRelId"`
	AuthorID             int64  `json:"authorId"`
	AuthorNickname       string `json:"authorNickname"`
	BatchID              int64  `json:"batchId"`
	BatchPushType        int32  `json:"batchPushType"`
}

// AuthorHistory 稿件推送批次历史详细信息
type AuthorHistory struct {
	AuthorID     int64      `json:"authorId"`
	MID          int64      `json:"mid"`
	PushVendorID int64      `json:"pushVendorId"`
	BOpenID      string     `json:"bOpenId"`
	OOpenID      string     `json:"oOpenId"`
	ActionTime   string     `json:"actionTime"`
	ActionMsg    string     `json:"actionMsg"`
	FileURL      string     `json:"fileUrl"`
	CUser        string     `json:"cuser"`
	CTime        xtime.Time `json:"ctime"`
}

// GetAuthorWhiteListKeyByAuthor 根据作者获取对应白名单key
func GetAuthorWhiteListKeyByAuthor(vendorID int64, mid int64) (string, error) {
	switch vendorID {
	case 0, DefaultVendors[0].ID, DefaultVendors[1].ID:
		return fmt.Sprintf(RedisAuthorWhiteListKey, strconv.FormatInt(mid, 10)), nil
	case DefaultVendors[2].ID:
		return fmt.Sprintf("%d_"+RedisAuthorWhiteListKey, vendorID, strconv.FormatInt(mid, 10)), nil
	default:
		return "", ecode.VendorNotFound
	}
}

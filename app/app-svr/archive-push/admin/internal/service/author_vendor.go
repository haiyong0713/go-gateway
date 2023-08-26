package service

import (
	"go-common/library/log"
	"time"

	xtime "go-common/library/time"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
	"go-gateway/app/app-svr/archive-push/ecode"
)

func (s *Service) SyncAuthorBindingQQTGL(sync model.SyncAuthorBindingReq, author *model.ArchivePushAuthor) (err error) {
	actionTime := time.Now()
	if sync.ActionTime != "" {
		actionTime, _ = time.Parse(model.DefaultTimeLayout, sync.ActionTime)
	}
	switch sync.Action {
	case string(qqModel.SyncUserActionBindStart):
		if actionTime.After(author.MTime.Time()) {
			author.OuterID = sync.OOpenID
			author.BindStatus = api.AuthorBindStatus_BINDED
			author.BindTime = xtime.Time(actionTime.Unix())
			author.VerificationStatus = api.AuthorVerificationStatus_VERIFYING
			author.MUser = "tgl"
			author.MTime = xtime.Time(actionTime.Unix())
		}
		break
	case string(qqModel.SyncUserActionBindSuccess):
		if actionTime.After(author.MTime.Time()) {
			author.OuterID = sync.OOpenID
			author.BindStatus = api.AuthorBindStatus_BINDED
			if author.BindTime.Time().Unix() == 0 {
				author.BindTime = xtime.Time(actionTime.Unix())
			}
			author.VerificationStatus = api.AuthorVerificationStatus_VERIFIED
			author.VerificationTime = xtime.Time(actionTime.Unix())
			author.MUser = "tgl"
			author.MTime = xtime.Time(actionTime.Unix())
		}
		break
	case string(qqModel.SyncUserActionBindReject):
		if actionTime.After(author.MTime.Time()) {
			author.OuterID = sync.OOpenID
			author.BindStatus = api.AuthorBindStatus_BINDED
			author.VerificationStatus = api.AuthorVerificationStatus_FAILED
			author.VerificationTime = xtime.Time(0)
			author.MUser = "tgl"
			author.MTime = xtime.Time(actionTime.Unix())
		}
		break
	case string(qqModel.SyncUserActionBindCancel):
		if actionTime.After(author.MTime.Time()) {
			author.OuterID = " "
			author.BindStatus = api.AuthorBindStatus_CANCELED
			author.BindTime = xtime.Time(0)
			author.VerificationStatus = api.AuthorVerificationStatus_CANCELED
			author.VerificationTime = xtime.Time(0)
			author.MUser = "tgl"
			author.MTime = xtime.Time(actionTime.Unix())
		}
		break
	default:
		err = ecode.SyncRequestError
		return
	}

	if err = s.dao.UpdateAuthorByID(author); err != nil {
		log.Error("Service: SyncAuthorBindingQQTGL UpdateAuthorByID %+v error %v", author, err)
		return
	}

	if err = s.HandleAuthorStatusChange(sync.VendorID, author.MID, author.AuthorizationStatus, author.BindStatus, author.VerificationStatus); err != nil {
		log.Error("Service: SyncAuthorAuthorizationQQTGL HandleAuthorStatusChange (%+v) error %v", sync, err)
	}

	return
}

// SyncAuthorAuthorizationQQTGL 同步作者授权状态
func (s *Service) SyncAuthorAuthorizationQQTGL(sync model.SyncAuthorAuthorizationReq) (err error) {
	if sync.MID == 0 {
		return
	}
	authorizationStatus := api.AuthorAuthorizationStatus_AUTHORIZED
	if err = s.HandleAuthorStatusChange(sync.VendorID, sync.MID, authorizationStatus, 0, 0); err != nil {
		log.Error("Service: SyncAuthorAuthorizationQQTGL HandleAuthorStatusChange (%+v) error %v", sync, err)
	}

	return
}

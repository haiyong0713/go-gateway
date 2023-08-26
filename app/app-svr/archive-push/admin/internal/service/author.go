package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
	archiveEcode "go-gateway/app/app-svr/archive/ecode"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Service) GetAuthorsByPage(ids []int64, mid int64, authorizationStatus int32, bindStatus int32, verificationStatus int32, vendorID int64, cuser string, pn int, ps int) (list []*model.ArchivePushAuthorX, total int64, err error) {
	if vendorID != 0 {
		if bindable, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
			log.Error("Service: GetAuthorsByUser CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
			err = _err
			return
		} else if !bindable {
			err = ecode.VendorNotAbleToBindAuthor
			return
		}
	}

	var rawList []*model.ArchivePushAuthor
	if rawList, total, err = s.dao.GetAuthorsByPage(ids, mid, authorizationStatus, bindStatus, verificationStatus, vendorID, 0, cuser, pn, ps); err != nil {
		log.Error("Service: GetAuthorsByPage error (%v)", err)
		return nil, 0, err
	}
	list = make([]*model.ArchivePushAuthorX, 0)
	if len(rawList) > 0 {
		for _, author := range rawList {
			_author := &model.ArchivePushAuthorX{
				ArchivePushAuthor:   *author,
				AuthorizationStatus: api.AuthorAuthorizationStatus_Enum_name[int32(author.AuthorizationStatus)],
				BindStatus:          api.AuthorBindStatus_Enum_name[int32(author.BindStatus)],
				VerificationStatus:  api.AuthorVerificationStatus_Enum_name[int32(author.VerificationStatus)],
			}
			if vendor, _err := s.GetVendorByID(_author.PushVendorID); _err != nil {
				log.Error("Service: GetAuthorsByPage GetVendorByID(%d) error %v", _author.PushVendorID, err)
			} else {
				_author.PushVendorName = vendor.Name
			}
			list = append(list, _author)
		}
	}
	return
}

// GetAuthorsByUser 根据MID或Open ID查询作者
func (s *Service) GetAuthorsByUser(vendorID int64, mid int64, openId string) (list []*model.ArchivePushAuthorX, err error) {
	if vendorID != 0 {
		if bindable, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
			log.Error("Service: GetAuthorsByUser CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
			err = _err
			return
		} else if !bindable {
			err = ecode.VendorNotAbleToBindAuthor
			return
		}
	}

	var (
		rawList   []*model.ArchivePushAuthor
		histories map[int64]*model.AuthorHistory
	)
	if rawList, err = s.dao.GetAuthorsByUser(vendorID, mid, openId); err != nil {
		log.Error("Service: GetAuthorsByUser GetAuthorsByUser error (%v)", err)
		return nil, err
	}
	if histories, err = s.GetAuthorHistory(vendorID); err != nil {
		log.Error("Service: GetAuthorsByUser GetAuthorHistory error %v", err)
		return
	}
	list = make([]*model.ArchivePushAuthorX, 0)
	if len(rawList) > 0 {
		for _, author := range rawList {
			_author := &model.ArchivePushAuthorX{
				ArchivePushAuthor:   *author,
				AuthorizationStatus: api.AuthorAuthorizationStatus_Enum_name[int32(author.AuthorizationStatus)],
				BindStatus:          api.AuthorBindStatus_Enum_name[int32(author.BindStatus)],
				VerificationStatus:  api.AuthorVerificationStatus_Enum_name[int32(author.VerificationStatus)],
			}
			// fill vendor name
			if vendor, _err := s.GetVendorByID(_author.PushVendorID); _err != nil {
				log.Error("Service: GetAuthorsByPage GetVendorByID(%d) error %v", _author.PushVendorID, err)
			} else {
				_author.PushVendorName = vendor.Name
			}
			// file reason
			if history, exists := histories[_author.ID]; exists {
				_author.Reason = history.ActionMsg
			}
			list = append(list, _author)
		}
	}
	return
}

// GetRawAuthorsByUser 根据MID或Open ID查询作者
func (s *Service) GetRawAuthorsByUser(vendorID int64, mid int64, openId string) (list []*model.ArchivePushAuthor, err error) {
	if bindable, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
		log.Error("Service: GetAuthorsByUser CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
		err = _err
		return
	} else if !bindable {
		err = ecode.VendorNotAbleToBindAuthor
		return
	}

	if list, err = s.dao.GetAuthorsByUser(vendorID, mid, openId); err != nil {
		log.Error("Service: GetAuthorsByPage error (%v)", err)
		return nil, err
	}
	return
}

// GetAuthorHistory 获取作者变更历史
func (s *Service) GetAuthorHistory(vendorID int64) (res map[int64]*model.AuthorHistory, err error) {
	if vendorID == 0 {
		return
	}
	if bindable, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
		log.Error("Service: GetAuthorHistory CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
		err = _err
		return
	} else if !bindable {
		err = ecode.VendorNotAbleToBindAuthor
		return
	}

	res = make(map[int64]*model.AuthorHistory)
	queryParams := &model.AuditLogSearchParams{Business: model.BusinessIDAuthor, Order: "ctime", Type: int(vendorID)}
	var rawRes *model.AuditLogSearchResRawData
	if rawRes, err = s.dao.SearchAuditLog(queryParams); err != nil {
		log.Error("Service: GetAuthorHistory SearchAuditLog Error (%v)", err)
		return nil, err
	} else if rawRes == nil || len(rawRes.Result) == 0 {
		return
	}
	for _, _logObj := range rawRes.Result {
		logObj := _logObj
		history := &model.AuthorHistory{
			PushVendorID: int64(logObj.Type),
			MID:          logObj.OID,
			AuthorID:     logObj.Int0,
			BOpenID:      logObj.Str0,
			OOpenID:      logObj.Str1,
			ActionTime:   logObj.Str2,
			ActionMsg:    logObj.Str3,
			CUser:        logObj.UName,
			CTime:        xtime.Time(logObj.Int1),
		}
		if _, exists := res[logObj.Int0]; !exists {
			res[logObj.Int0] = history
		}
	}
	return
}

// UploadAuthors 上传并添加作者
func (s *Service) UploadAuthors(vendorID int64, fileURL string, activityID int64, username string, uid int64) (res []*model.ArchivePushAuthor, err error) {
	if vendorID == 0 || fileURL == "" || activityID == 0 {
		err = ecode.PushRequestError
		return
	}
	if bindable, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
		log.Error("Service: GetAuthorHistory CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
		err = _err
		return
	} else if !bindable {
		err = ecode.VendorNotAbleToBindAuthor
		return
	}

	var (
		csvBuf   *bytes.Buffer
		closeBuf func()
		rawMIDs  = make([]string, 0)
		mids     = make([]int64, 0)
	)
	csvBuf, closeBuf, err = s.dao.Download(fileURL)
	if err != nil {
		log.Error("Service: UploadAuthors Download Error (%v)", err)
		return
	}
	defer closeBuf()
	r := csv.NewReader(strings.NewReader(string(csvBuf.Bytes())))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("Service: UploadAuthors ReadAll Error (%v)", err)
		return
	}
	rawMIDs = s.getAllMIDsFromRecords(records)
	for _, rawMID := range rawMIDs {
		_rawMID := rawMID
		if mid, _err := strconv.ParseInt(_rawMID, 10, 64); _err != nil {
			log.Error("Service: UploadAuthors ParseInt(%s) error %v", _rawMID, err)
			continue
		} else {
			mids = append(mids, mid)
		}
	}
	if len(mids) == 0 {
		log.Warn("Service: UploadAuthors(%s) no valid authors to add", fileURL)
		return
	}
	res, err = s.AddAuthors(mids, vendorID, activityID, username, uid, fileURL)

	return
}

// getAllMIDsFromRecords 从csv文件中读取的记录提取mid
func (s *Service) getAllMIDsFromRecords(records [][]string) (mids []string) {
	if len(records) == 0 {
		return
	}
	bvidsMap := make(map[string]int)
	for _, rec := range records {
		if len(rec) > 0 && strings.TrimSpace(rec[0]) != "" {
			bvidsMap[strings.TrimSpace(rec[0])] = 1
		}
	}
	mids = make([]string, 0)
	for bvid := range bvidsMap {
		mids = append(mids, bvid)
	}
	return
}

// AddAuthors 根据MIDs添加作者完整逻辑
func (s *Service) AddAuthors(mids []int64, vendorID int64, activityID int64, username string, uid int64, fileURL string) (res []*model.ArchivePushAuthor, err error) {
	if len(mids) == 0 || vendorID == 0 {
		err = xecode.RequestErr
		return
	}
	if able, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
		err = _err
		log.Error("Service: AddAuthors CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
		return
	} else if !able {
		err = ecode.VendorNotAbleToBindAuthor
		return
	}

	var (
		authors           = make([]*model.ArchivePushAuthor, 0)
		authorIDs         = make([]int64, 0)
		authorPush        *model.ArchivePushAuthorPush
		pushConditions    = make([]*model.ArchivePushAuthorPushCondition, 0)
		pushableAuthorIDs = make([]int64, 0)
	)

	// 查询Open ID并构建作者models
	if authors, err = s.GenerateAndFillAuthors(mids, vendorID, activityID, username); err != nil {
		log.Error("Service: AddAuthors GenerateAndFillAuthors(%v, %d) error %v", mids, vendorID, err)
		return
	}
	// 添加作者数据
	if res, err = s.AddRawAuthors(authors); err != nil {
		log.Error("Service: AddAuthors GenerateAndFillAuthors(%v, %d) error %v", mids, vendorID, err)
		return
	}
	for _, author := range res {
		authorIDs = append(authorIDs, author.ID)
	}
	// 添加作者白名单
	if err = s.AddWhiteListForAuthors(vendorID, authors); err != nil {
		log.Error("Service: AddAuthors AddWhiteListForAuthors(%d, %v) error %v", vendorID, authors, err)
	}
	// 恢复稿件推送关系记录
	if authorPushes, _err := s.dao.GetAuthorPushesByVendorIDs([]int64{vendorID}); _err != nil {
		log.Error("Service: AddAuthors GetAuthorPushesByVendorIDs %d error %v", vendorID, _err)
		err = _err
		return
	} else if len(authorPushes) == 0 {
		log.Error("Service: AddAuthors GetAuthorPushesByVendorIDs %d 未找到有效作者推送", vendorID)
	} else {
		authorPush = authorPushes[0]
		if err = json.Unmarshal([]byte(authorPush.PushConditions), &pushConditions); err != nil {
			log.Error("Service: AddAuthors GetAuthorPushesByVendorIDs Unmarshal %s error %v", authorPush.PushConditions, err)
			return
		}
		for _, author := range res {
			_author := author
			if valid := s.ValidateAuthorPushConditionsWithAuthor(pushConditions, *_author); valid {
				pushableAuthorIDs = append(pushableAuthorIDs, _author.ID)
			}
		}
		if len(pushableAuthorIDs) > 0 {
			for _, authorID := range pushableAuthorIDs {
				rel := &model.ArchivePushBatchAuthorPushRel{
					AuthorID:     authorID,
					AuthorPushID: authorPush.ID,
				}
				if _, _err := s.dao.AddBatchAuthorPushRel(rel); _err != nil {
					log.Error("Service: AddAuthors AddBatchAuthorPushRel (%+v) error %v", rel, _err)
				}
			}
			log.Info("Service: AddAuthors RestoreBatchAuthorPushRelsByAuthorIDs (%v)", authorIDs)
		}
	}

	// 行为日志
	go s.addAddAuthorsActionLog(res, username, uid, fileURL)

	return
}

// FilterExistingAuthors 过滤已存在的稿件作者信息
func (s *Service) FilterExistingAuthors(mids []int64, vendorID int64) (validaMIDs []int64, err error) {
	if len(mids) == 0 || vendorID == 0 {
		return
	}
	validaMIDs = make([]int64, 0, len(mids))
	var mutex sync.Mutex
	eg := errgroup.WithContext(context.Background())
	for _, mid := range mids {
		_mid := mid
		eg.Go(func(ctx context.Context) error {
			if existingAuthors, _err := s.GetAuthorsByUser(vendorID, _mid, ""); _err != nil {
				log.Error("Service: FilterExistingAuthors GetAuthorsByUser(%d, '', %d) error %v", _mid, vendorID, _err)
				return _err
			} else if len(existingAuthors) == 0 || existingAuthors[0].ID == 0 {
				mutex.Lock()
				validaMIDs = append(validaMIDs, _mid)
				mutex.Unlock()
			} else {
				log.Warn("Service: FilterExistingAuthors GetAuthorsByUser(%d, '', %d) existing", _mid, vendorID)
			}

			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		log.Error("Service: FilterExistingAuthors(%v, %d) error %v", mids, vendorID, err)
	}

	return
}

// GenerateAndFillAuthors 根据MIDs构建作者model并填充相关信息。如Open ID和昵称
func (s *Service) GenerateAndFillAuthors(mids []int64, vendorID int64, authorizationSID int64, username string) (res []*model.ArchivePushAuthor, err error) {
	res = make([]*model.ArchivePushAuthor, 0)
	eg := errgroup.WithContext(context.Background())
	var lock sync.Mutex
	var appKey string
	if appKey, err = s.GetOauthAppKeyByVendorID(vendorID); err != nil {
		log.Error("Service: GenerateAndFillAuthors GetOauthAppKeyByVendorID(%d) error %v", vendorID, err)
		return
	}
	for _, mid := range mids {
		_mid := mid
		eg.Go(func(ctx context.Context) error {
			author := &model.ArchivePushAuthor{
				MID:              _mid,
				OpenID:           "",
				Nickname:         "",
				PushVendorID:     vendorID,
				AuthorizationSID: authorizationSID,
				CUser:            username,
				CTime:            xtime.Time(time.Now().Unix()),
				MUser:            username,
				MTime:            xtime.Time(time.Now().Unix()),
			}

			// 填充用户昵称
			if account, _err := s.dao.GetAccountInfoByMID(_mid); _err != nil {
				log.Error("Service: GenerateAndFillAuthors GetAccountInfoByMID(%d) error %v", _mid, _err)
				return _err
			} else {
				author.Nickname = account.Name
			}
			// 填充Open ID
			if openID, _err := s.GetOpenIDByMID(_mid, appKey); _err != nil {
				log.Error("Service: GenerateAndFillAuthors GetOpenIDByMID(%d, %s) error %v", _mid, appKey, _err)
			} else {
				author.OpenID = openID
			}
			// 填充授权状态
			if authorized, authorizationTime, _err := s.dao.CheckIfAuthorizedByMID(authorizationSID, _mid); _err != nil {
				log.Error("Service: GenerateAndFillAuthors CheckIfAuthorizedByMID(%d, %d) error %v", authorizationSID, _mid, _err)
			} else {
				if authorized {
					author.AuthorizationStatus = api.AuthorAuthorizationStatus_AUTHORIZED
					// 若有授权时间则设置为授权时间，否则为当前时间
					if authorizationTime.Time().Unix() > 0 {
						author.AuthorizationTime = authorizationTime
					} else {
						author.AuthorizationTime = xtime.Time(time.Now().Unix())
					}
				}
			}
			// GICP填充认证状态
			if author.PushVendorID == model.DefaultVendors[0].ID {
				if cmcAuthor, _err := s.dao.GetAuthorByMID(model.DefaultVendors[1].ID, _mid); _err != nil {
					log.Error("Service: GenerateAndFillAuthors GetAuthorByMID (%d, %d) error %v", model.DefaultVendors[1].ID, _mid, _err)
				} else if cmcAuthor != nil {
					author.BindStatus = cmcAuthor.BindStatus
					author.BindTime = cmcAuthor.BindTime
					author.VerificationStatus = cmcAuthor.VerificationStatus
					author.VerificationTime = cmcAuthor.VerificationTime
				}
			}

			lock.Lock()
			res = append(res, author)
			lock.Unlock()

			return nil
		})
	}

	err = eg.Wait()

	return
}

// AddRawAuthors 添加作者数据
func (s *Service) AddRawAuthors(authors []*model.ArchivePushAuthor) (res []*model.ArchivePushAuthor, err error) {
	res = make([]*model.ArchivePushAuthor, 0, len(authors))
	for _, author := range authors {
		_author := author
		if err = s.dao.AddAuthor(_author); err != nil {
			log.Error("Service: AddRawAuthors AddAuthor(%+v) error %v", _author, err)
			continue
		}
		res = append(res, _author)
	}

	return
}

// AddWhiteListForAuthors 给作者添加白名单
func (s *Service) AddWhiteListForAuthors(vendorID int64, authors []*model.ArchivePushAuthor) (err error) {
	if vendorID == 0 || len(authors) == 0 {
		return
	}

	authorsWithBVIDs := make([]*model.ArchivePushAuthorWithBVIDs, 0)
	for _, author := range authors {
		_author := author
		authorWithBVIDs := &model.ArchivePushAuthorWithBVIDs{
			ArchivePushAuthor: *_author,
			BVIDs:             make([]string, 0),
		}
		authorsWithBVIDs = append(authorsWithBVIDs, authorWithBVIDs)
	}
	err = s.dao.PutAuthorsForWhiteList(vendorID, authorsWithBVIDs)

	return
}

// addAddAuthorsActionLog 新建作者数据行为日志
func (s *Service) addAddAuthorsActionLog(authors []*model.ArchivePushAuthor, username string, uid int64, fileURL string) {
	now := time.Now()
	for _, author := range authors {
		_author := author

		index := []interface{}{
			_author.ID,
			now.Unix(),
		}

		content := map[string]interface{}{
			"fileURL": fileURL,
		}
		params := &model.AuditLogInitParams{
			UName:    username,
			UID:      uid,
			Business: model.BusinessIDAuthor,
			Type:     int(_author.PushVendorID),
			OID:      _author.MID,
			Action:   "添加作者",
			Content:  content,
			CTime:    now,
			Index:    index,
		}
		if err := s.dao.AddAuditLog(params); err != nil {
			log.Error("Service: addAddAuthorsActionLog AddAuditLog(%+v) Error (%v)", params, err)
		}
	}
}

// RemoveAuthor 移除授权作者完整逻辑
func (s *Service) RemoveAuthor(vendorID int64, mid int64, reason string, needWithdrawHistoryArcs bool, username string, uid int64) (err error) {
	if vendorID == 0 || mid == 0 {
		return xecode.RequestErr
	}
	var (
		rawAuthor *model.ArchivePushAuthor
	)
	// 获取作者信息
	if rawAuthors, _err := s.GetRawAuthorsByUser(vendorID, mid, ""); _err != nil {
		err = _err
		log.Error("Service: RemoveAuthor GetRawAuthorsByUser(%d, %d, '') error %v", vendorID, mid, err)
		return
	} else if len(rawAuthors) == 0 {
		err = ecode.AuthorNotFound
		return
	} else {
		rawAuthor = rawAuthors[0]
	}

	// 移除作者白名单
	// 除了王者营地的
	if vendorID != model.DefaultVendors[0].ID {
		if err = s.dao.RemoveAuthorsFromWhiteList(vendorID, []int64{mid}); err != nil {
			log.Error("Service: RemoveAuthor RemoveAuthorsFromWhiteList(%d, %d) error %v", vendorID, mid, err)
			return
		}
	}

	if needWithdrawHistoryArcs {
		// 推送下架作者已推送稿件
		if err = s.WithdrawArchivesByAuthor(vendorID, mid, reason, username, uid); err != nil {
			log.Error("Service: RemoveAuthor WithdrawArchivesByAuthor(%d, %d)", vendorID, mid)
			return
		}
	}

	// 删除作者信息
	if err = s.RemoveRawAuthorByID(rawAuthor.ID, username); err != nil {
		log.Error("Service: RemoveAuthor RemoveRawAuthor(%d, %d, %s)", vendorID, mid, username)
		return
	}

	// 删除作者推送数据
	if err = s.RemoveBatchAuthorPushRelsByAuthor(rawAuthor.ID, username); err != nil {
		log.Error("Service: RemoveAuthor RemoveBatchAuthorPushRelsByAuthor (%d, %s) error %v", rawAuthor.ID, username, err)
	}

	// 行为日志
	go s.addRemoveAuthorActionLog([]*model.ArchivePushAuthor{rawAuthor}, reason, username, uid)

	return
}

// RemoveRawAuthorByID 删除作者数据
func (s *Service) RemoveRawAuthorByID(id int64, username string) (err error) {
	if err = s.dao.DeleteAuthorByID(id, username); err != nil {
		log.Error("Service: RemoveRawAuthorByID (%d) error %v", id, err)
	}
	return
}

// WithdrawArchivesByAuthor 推送下架作者已推送稿件
func (s *Service) WithdrawArchivesByAuthor(vendorID int64, mid int64, reason string, username string, uid int64) (err error) {
	if vendorID == 0 || mid == 0 {
		return xecode.RequestErr
	}
	// 获取作者已推送稿件batch ids
	var (
		batchDetails = make([]*model.ArchivePushBatchDetail, 0)
	)
	if batchDetails, err = s.dao.GetBatchDetailsByAuthor(vendorID, mid); err != nil {
		log.Error("Service: WithdrawArchivesByAuthor (%d, %d) GetBatchDetailsByBatchID error %v", vendorID, mid, err)
		return
	} else if len(batchDetails) == 0 {
		log.Warn("Service: WithdrawArchivesByAuthor (%d, %d) 没有要推送下架的稿件", vendorID, mid)
		return
	}

	// 下架batches
	eg := errgroup.WithContext(context.Background())
	for _, batchDetail := range batchDetails {
		_batchDetail := batchDetail
		if _batchDetail == nil || _batchDetail.ArchiveStatus != api.ArchiveStatus_OPEN || _batchDetail.PushStatus != api.ArchivePushDetailPushStatus_OUTER_FAIL || _batchDetail.AID == 0 {
			log.Warn("Service: WithdrawArchivesByAuthor (%d, %d) batch detail (%+v) 状态无法下架", vendorID, mid, _batchDetail)
			continue
		}
		eg.Go(func(ctx context.Context) error {
			bvid, err := util.AvToBv(_batchDetail.AID)
			if err != nil {
				log.Error("Service: WithdrawArchivesByAuthor (%d, %d) batch detail (%+v) AvToBv error %v", vendorID, mid, _batchDetail, err)
				return nil
			}
			return s.WithdrawArchive(bvid, reason, vendorID, true, username, uid)
		})
	}
	err = eg.Wait()

	return
}

// addAddAuthorsActionLog 移除作者数据行为日志
func (s *Service) addRemoveAuthorActionLog(authors []*model.ArchivePushAuthor, reason string, username string, uid int64) {
	if len(authors) == 0 {
		return
	}
	now := time.Now()
	for _, author := range authors {
		_author := author

		index := []interface{}{
			_author.ID,
			now.Unix(),
			_author.OpenID,
			_author.OuterID,
			now.Format(model.DefaultTimeLayout),
			reason,
		}

		params := &model.AuditLogInitParams{
			UName:    username,
			UID:      uid,
			Business: model.BusinessIDAuthor,
			Type:     int(_author.PushVendorID),
			OID:      _author.MID,
			Action:   "移除作者",
			CTime:    now,
			Index:    index,
		}
		if err := s.dao.AddAuditLog(params); err != nil {
			log.Error("Service: addRemoveAuthorActionLog AddAuditLog(%+v) Error (%v)", params, err)
		}
	}
}

// SyncAuthorBinding 更新作者绑定信息逻辑
func (s *Service) SyncAuthorBinding(sync model.SyncAuthorBindingReq) (err error) {
	if sync.BOpenID == "" || sync.VendorID == 0 {
		return ecode.SyncRequestError
	}
	if sync.ActionTime != "" {
		if _, _err := time.Parse(model.DefaultTimeLayout, sync.ActionTime); _err != nil {
			log.Error("Service: SyncAuthorBinding time.Parse(%s) error %v", sync.ActionTime, _err)
			return ecode.SyncRequestError
		}
	}
	var (
		author *model.ArchivePushAuthor
	)
	log.Info("Service: SyncAuthorBinding %+v Start", sync)
	defer func() {
		log.Info("Service: SyncAuthorBinding %+v End", sync)
	}()
	if valid, _err := s.CheckVendorAbleToBindUser(sync.VendorID); _err != nil {
		log.Error("Service: SyncAuthorBinding CheckVendorAbleToBindUser(%d) error %v", sync.VendorID, _err)
		return _err
	} else if !valid {
		return ecode.VendorNotAbleToSyncAuthorStatus
	}
	if authorList, _err := s.GetRawAuthorsByUser(sync.VendorID, 0, sync.BOpenID); _err != nil {
		log.Error("Service: SyncAuthorBinding GetAuthorsByUser(%d, 0, %s) error %v", sync.VendorID, sync.BOpenID, _err)
		return
	} else if len(authorList) == 0 || authorList[0].ID == 0 {
		log.Warn("Service: SyncAuthorBinding cannot find author by open id (%s)", sync.BOpenID)
		var (
			mid    int64
			appKey string
		)
		if appKey, err = s.GetOauthAppKeyByVendorID(sync.VendorID); err != nil {
			log.Error("Service: SyncAuthorBinding GetOauthAppKeyByVendorID(%d) error %v", sync.VendorID, err)
			return
		}
		if mid, err = s.GetMIDByOpenID(sync.BOpenID, appKey); err != nil {
			log.Error("Service: SyncAuthorBinding GetMIDByOpenID(%s) error %v", sync.BOpenID, err)
			return
		} else if mid == 0 {
			err = ecode.AuthorNotFound
			return
		}
		if authorList, _err = s.GetRawAuthorsByUser(sync.VendorID, mid, ""); _err != nil {
			log.Error("Service: SyncAuthorBinding GetRawAuthorsByUser(%d, %d, '') error %v", mid, sync.VendorID, err)
			return
		} else if len(authorList) == 0 || authorList[0].ID == 0 {
			log.Error("Service: SyncAuthorBinding GetRawAuthorsByUser(%d, %d, '') 没有找到对应MID作者", mid, sync.VendorID)
			err = ecode.AuthorNotFound
			return
		}
		author = authorList[0]
		author.OpenID = sync.BOpenID
		if err = s.dao.UpdateAuthorByID(author); err != nil {
			log.Error("Service: SyncAuthorBinding UpdateAuthorByID(%+v) error %v", author, err)
			return
		}
	} else {
		author = authorList[0]
	}

	switch sync.VendorID {
	case model.DefaultVendors[1].ID:
		err = s.SyncAuthorBindingQQTGL(sync, author)
	default:
		err = ecode.VendorNotAbleToSyncAuthorStatus
	}

	// 行为日志
	if _err := s.AddSyncAuthorBindingActionLog(sync, author); _err != nil {
		log.Error("Service: SyncAuthorBinding AddSyncAuthorBindingActionLog(%+v) error %v", sync, _err)
	}

	return
}

// AddSyncAuthorBindingActionLog 添加更新作者绑定信息行为日志
func (s *Service) AddSyncAuthorBindingActionLog(sync model.SyncAuthorBindingReq, author *model.ArchivePushAuthor) (err error) {
	now := time.Now()
	index := []interface{}{
		author.ID,
		now.Unix(),
		author.OpenID,
		author.OuterID,
		sync.ActionTime,
		sync.ActionMsg,
	}

	params := &model.AuditLogInitParams{
		UName:    "tgl",
		UID:      0,
		Business: model.BusinessIDAuthor,
		Type:     int(sync.VendorID),
		OID:      author.MID,
		Action:   sync.Action,
		CTime:    now,
		Index:    index,
	}
	if err := s.dao.AddAuditLog(params); err != nil {
		log.Error("Service: AddBatchDetailActionLog Error (%v)", err)
		return err
	}
	return
}

// SyncAuthorAuthorization 更新作者授权信息逻辑
func (s *Service) SyncAuthorAuthorization(sync model.SyncAuthorAuthorizationReq) (err error) {
	var (
		author *model.ArchivePushAuthor
	)
	if authorList, _err := s.GetRawAuthorsByUser(sync.VendorID, sync.MID, ""); _err != nil || len(authorList) == 0 || authorList[0].ID == 0 {
		log.Error("Service: SyncAuthorAuthorization GetAuthorsByUser(%d, %d, '') error %v", sync.VendorID, sync.MID, _err)
		err = ecode.AuthorNotFound
		return
	} else {
		author = authorList[0]
	}

	switch sync.VendorID {
	case model.DefaultVendors[1].ID:
		err = s.SyncAuthorAuthorizationQQTGL(sync)
	default:
		err = ecode.SyncRequestError
	}

	// 行为日志
	if _err := s.AddSyncAuthorAuthorizationActionLog(sync, author); _err != nil {
		log.Error("Service: SyncAuthorAuthorization AddSyncAuthorBindingActionLog(%+v) error %v", sync, _err)
	}

	return
}

// AddSyncAuthorAuthorizationActionLog 添加更新作者授权信息行为日志
func (s *Service) AddSyncAuthorAuthorizationActionLog(sync model.SyncAuthorAuthorizationReq, author *model.ArchivePushAuthor) (err error) {
	now := time.Now()
	index := []interface{}{
		author.ID,
		now.Unix(),
		author.OpenID,
		now.Format(model.DefaultTimeLayout),
	}

	params := &model.AuditLogInitParams{
		UName:    "system",
		UID:      0,
		Business: model.BusinessIDAuthor,
		Type:     int(sync.VendorID),
		OID:      author.MID,
		Action:   "authorization",
		CTime:    now,
		Index:    index,
	}
	if err := s.dao.AddAuditLog(params); err != nil {
		log.Error("Service: AddBatchDetailActionLog Error (%v)", err)
		return err
	}
	return
}

// GetPushableAuthors 获取所有可推送稿件的作者（已授权&已绑定&已认证）
func (s *Service) GetPushableAuthors(vendorID int64, conditions []*model.ArchivePushAuthorPushCondition) (res []*model.ArchivePushAuthor, err error) {
	if vendorID == 0 {
		err = xecode.RequestErr
		return
	}
	res = make([]*model.ArchivePushAuthor, 0)
	var (
		allAuthors []*model.ArchivePushAuthor
	)
	if allAuthors, err = s.dao.GetAuthorsByUser(vendorID, 0, ""); err != nil {
		log.Error("Service: GetPushableAuthors(%d) GetAuthorsByUser error %v", vendorID, err)
		return
	} else if len(allAuthors) == 0 {
		log.Warn("Service: GetPushableAuthors(%d) 下作者数量为0", vendorID)
		res = allAuthors
		return
	}

	for _, author := range allAuthors {
		_author := author
		if valid := s.ValidateAuthorPushConditionsWithAuthor(conditions, *_author); valid {
			res = append(res, _author)
		}
	}
	if len(res) == 0 {
		log.Warn("Service: GetPushableAuthors(%d) 可推送作者数量为0", vendorID)
	}

	return
}

// GetUserAuthorizationSIDByUser 根据推送厂商获取对应（授权）活动SID
func (s *Service) GetUserAuthorizationSIDByUser(vendorID int64, mid int64) (sid int64, err error) {
	if vendorID == 0 || mid == 0 {
		return 0, xecode.RequestErr
	}
	for _, vendor := range model.DefaultVendors {
		if vendor.UserBindable && vendorID == vendor.ID {
			return s.dao.GetAuthorizationSIDByVendorAndMID(vendorID, mid)
		}
	}
	return 0, ecode.VendorNotAbleToBindAuthor
}

// CheckVendorAbleToBindUser 检查推送厂商是否支持绑定用户并推送
func (s *Service) CheckVendorAbleToBindUser(vendorID int64) (bindable bool, err error) {
	bindable = false
	for _, vendor := range model.DefaultVendors {
		if vendor.ID == vendorID && vendor.UserBindable {
			bindable = true
			break
		}
	}
	return
}

// CheckIfAuthorInWhiteList 检查作者是否在授权白名单中
func (s *Service) CheckIfAuthorInWhiteList(vendorID int64, mid int64) (exists bool, err error) {
	if vendorID == 0 || mid == 0 {
		return false, xecode.RequestErr
	}
	if bindable, _err := s.CheckVendorAbleToBindUser(vendorID); _err != nil {
		log.Error("Service: CheckIfAuthorInWhiteList CheckVendorAbleToBindUser(%d) error %v", vendorID, _err)
		return false, _err
	} else if !bindable {
		return false, ecode.VendorNotAbleToBindAuthor
	}

	if authorMap, _err := s.dao.GetAuthorsWhiteList(vendorID, []int64{mid}); _err != nil {
		log.Error("Service: CheckIfAuthorInWhiteList GetAuthorsWhiteList(%d, %d) error %v", vendorID, mid, _err)
		return false, _err
	} else if len(authorMap) == 0 {
		return false, nil
	} else {
		if _, exists = authorMap[mid]; exists {
			return true, nil
		}
	}

	return false, nil
}

// CheckIfAuthorPushable 检查作者是否可推送
func (s *Service) CheckIfAuthorPushable(vendorID int64, mid int64) (pushable bool, err error) {
	if vendorID == 0 || mid == 0 {
		return false, xecode.RequestErr
	}
	pushable = false

	var (
		author         *model.ArchivePushAuthor
		authorPush     *model.ArchivePushAuthorPush
		pushConditions []*model.ArchivePushAuthorPushCondition
	)
	if exists, _err := s.CheckIfAuthorInWhiteList(vendorID, mid); _err != nil {
		log.Error("Service: CheckIfAuthorPushable CheckIfAuthorInWhiteList (%d, %d) error %v", vendorID, mid, _err)
		err = _err
		return
	} else if !exists {
		if s.Cfg.Debug {
			log.Warn("Service: CheckIfAuthorPushable CheckIfAuthorInWhiteList (%d, %d) 作者不在推送白名单中", vendorID, mid)
		}
		return
	}
	if authors, _err := s.dao.GetAuthorsByUser(vendorID, mid, ""); _err != nil {
		log.Error("Service: CheckIfAuthorPushable GetAuthorsByUser (%d, %d, '') error %v", vendorID, mid, _err)
		err = _err
		return
	} else if len(authors) == 0 {
		log.Error("Service: CheckIfAuthorPushable GetAuthorsByUser (%d, %d, '') 没找到对应作者", vendorID, mid)
		err = ecode.AuthorNotFound
		return
	} else {
		author = authors[0]
		if authorPushes, _err := s.dao.GetActiveAuthorPushesByVendorIDs([]int64{vendorID}); _err != nil {
			log.Error("Service: CheckIfAuthorPushable GetActiveAuthorPushesByVendorIDs %d error %v", vendorID, _err)
			err = _err
			return
		} else if len(authorPushes) == 0 {
			log.Error("Service: CheckIfAuthorPushable GetActiveAuthorPushesByVendorIDs %d 获取作者推送为空", vendorID)
			err = ecode.AuthorPushNotFound
			return
		} else {
			authorPush = authorPushes[0]
		}
		if err = json.Unmarshal([]byte(authorPush.PushConditions), &pushConditions); err != nil {
			log.Error("Service: CheckIfAuthorPushable Unmarshal %s error %v", authorPush.PushConditions, err)
			return
		}
		pushable = s.ValidateAuthorPushConditionsWithAuthor(pushConditions, *author)
	}

	return
}

// GetRawAuthorsByIDs 根据作者ID获取作者
func (s *Service) GetRawAuthorsByIDs(ids []int64) (res []*model.ArchivePushAuthor, err error) {
	if len(ids) == 0 {
		return
	}

	res = make([]*model.ArchivePushAuthor, 0)
	if res, err = s.dao.GetAuthorsByIDs(ids); err != nil {
		log.Error("Service: GetRawAuthorsByIDs GetAuthorsByIDs(%v) error %v", ids, err)
	}

	return
}

// GetRawAuthorByMID 根据MID获取作者
func (s *Service) GetRawAuthorByMID(vendorID int64, mid int64) (res *model.ArchivePushAuthor, err error) {
	if vendorID == 0 || mid == 0 {
		return nil, xecode.RequestErr
	}
	if res, err = s.dao.GetAuthorByMID(vendorID, mid); err != nil {
		log.Error("Service: GetRawAuthorByMID GetAuthorByMID (%d, %d) error %v", vendorID, mid, err)
	} else if res == nil || res.ID == 0 {
		err = ecode.AuthorNotFound
		log.Error("Service: GetRawAuthorByMID GetAuthorByMID (%d, %d) 无法找到对应绑定作者", vendorID, mid)
		return
	}

	return
}

// GetAuthorByBVID 根据稿件BVID获取作者
func (s *Service) GetAuthorByBVID(vendorID int64, bvid string) (author *model.ArchivePushAuthor, err error) {
	if bvid == "" {
		return nil, xecode.RequestErr
	}

	var (
		aid int64
		arc *archiveGRPC.Arc
	)
	if aid, err = util.BvToAv(bvid); err != nil {
		log.Error("Service: GetAuthorByBVID %s error %v", bvid, err)
		return
	}
	if arc, err = s.GetArcByAID(aid); err != nil {
		log.Error("Service: GetAuthorByBVID GetArcByAID %d error %v", aid, err)
		return
	} else if arc == nil {
		err = archiveEcode.ArchiveNotExist
		log.Error("Service: GetAuthorByBVID GetArcByAID %d 获取稿件为空", aid)
		return
	}
	author, err = s.GetRawAuthorByMID(vendorID, arc.Author.Mid)

	return
}

// HandleAuthorStatusChange 处理作者状态变更后的业务逻辑
func (s *Service) HandleAuthorStatusChange(vendorID int64, mid int64, authorizationStatus api.AuthorAuthorizationStatus_Enum, bindStatus api.AuthorBindStatus_Enum, verificationStatus api.AuthorVerificationStatus_Enum) (err error) {
	if vendorID == 0 || mid == 0 {
		return
	}
	var (
		author        *model.ArchivePushAuthor
		validVendorID int64
	)
	// 获取作者
	if author, err = s.dao.GetAuthorByMID(vendorID, mid); err != nil {
		log.Error("Service: HandleAuthorStatusChange GetAuthorByMID (%d, %d) error %v", vendorID, mid, err)
		return
	} else if author == nil {
		log.Error("Service: HandleAuthorStatusChange GetAuthorByMID (%d, %d) 获取作者为空", vendorID, mid)
		err = ecode.AuthorNotFound
		return
	}

	// 更新作者状态
	if authorizationStatus != 0 {
		author.AuthorizationStatus = authorizationStatus
	}
	if bindStatus != 0 {
		author.BindStatus = bindStatus
	}
	if verificationStatus != 0 {
		author.VerificationStatus = verificationStatus
	}
	if err = s.dao.UpdateAuthorByID(author); err != nil {
		log.Error("Service: HandleAuthorStatusChange UpdateAuthorByID (%+v) error %v", author, err)
		return
	}

	// TGL变更后同步CMC状态同步逻辑
	if author.PushVendorID == model.DefaultVendors[1].ID {
		if err = s.HandleAuthorStatusChange(model.DefaultVendors[0].ID, mid, authorizationStatus, bindStatus, verificationStatus); err != nil {
			log.Error("Service: HandleAuthorStatusChange 处理王者营地作者状态同步失败 %d %v", mid, err)
		}
	}
	// TGL&CMC状态处理
	if author.PushVendorID == model.DefaultVendors[0].ID || author.PushVendorID == model.DefaultVendors[1].ID {
		if validVendorID, err = s.handleAuthorStatusChangeTGLCMC(author); err != nil {
			log.Error("Service: HandleAuthorStatusChange handleAuthorStatusChangeTGLCMC (%+v) error %v", author, err)
			return
		}
	}
	if validVendorID > 0 && validVendorID != vendorID {
		if err = s.dao.DeleteBatchAuthorPushRelsByAuthorIDs([]int64{author.ID}, "system"); err != nil {
			log.Error("Service: HandleAuthorStatusChange DeleteBatchAuthorPushRelsByAuthorIDs %d error %v", author.ID, err)
		}
	}

	return
}

func (s *Service) handleAuthorStatusChangeTGLCMC(author *model.ArchivePushAuthor) (validVendorID int64, err error) {
	var (
		authorPushes            []*model.ArchivePushAuthorPush
		needCheckAuthorPushIDs  = make([]int64, 0)
		authorPushConditionsMap = make(map[int64][]*model.ArchivePushAuthorPushCondition)
		authorPushVendorIDMap   = make(map[int64]int64)
		batchAuthorPushRels     []*model.ArchivePushBatchAuthorPushRel
		needDeleteBatchIDs      = make([]int64, 0)
	)
	// 检查batch-author-push，待推送的batch，如果不符合作者推送条件了
	// 则统一更新成符合条件的vendor
	if authorPushes, err = s.dao.GetActiveAuthorPushesByVendorIDs([]int64{model.DefaultVendors[0].ID, model.DefaultVendors[1].ID}); err != nil {
		log.Error("Service: handleAuthorStatusChangeTGLCMC GetAllAuthorPushes error %v", err)
		return
	} else if len(authorPushes) == 0 {
		log.Warn("Service: handleAuthorStatusChangeTGLCMC GetAllAuthorPushes 没有作者推送数据")
		return
	}
	for _, authorPush := range authorPushes {
		_authorPush := authorPush
		if _authorPush.PushConditions != "" {
			authorPushConditions := make([]*model.ArchivePushAuthorPushCondition, 0)
			if err = json.Unmarshal([]byte(_authorPush.PushConditions), &authorPushConditions); err != nil {
				log.Error("Service: handleAuthorStatusChangeTGLCMC Unmarshal (%s) error %v", _authorPush.PushConditions, err)
				continue
			} else if len(authorPushConditions) == 0 {
				continue
			}
			authorPushConditionsMap[_authorPush.ID] = authorPushConditions
			needCheckAuthorPushIDs = append(needCheckAuthorPushIDs, _authorPush.ID)
		}
		authorPushVendorIDMap[_authorPush.ID] = _authorPush.VendorID
	}
	// 若都是无条件的，则不做变更
	if len(needCheckAuthorPushIDs) == 0 {
		validVendorID = author.PushVendorID
		return
	}
	if batchAuthorPushRels, err = s.dao.GetBatchAuthorPushRelsByPushIDsAndAuthorIDs(needCheckAuthorPushIDs, []int64{author.ID}); err != nil {
		log.Error("Service: handleAuthorStatusChangeTGLCMC GetBatchAuthorPushRelsByPushIDsAndAuthorIDs ")
		return
	}
	for pushID, conditions := range authorPushConditionsMap {
		if valid := s.ValidateAuthorPushConditionsWithAuthor(conditions, *author); !valid {
			for _, rel := range batchAuthorPushRels {
				_rel := rel
				if _rel.AuthorPushID == pushID {
					needDeleteBatchIDs = append(needDeleteBatchIDs, _rel.BatchID)
				}
			}
		} else {
			validVendorID = authorPushVendorIDMap[pushID]
		}
	}

	// 从待推送池中移除
	if err = s.dao.RemoveBatchesFromTodo(needDeleteBatchIDs); err != nil {
		log.Error("Service: handleAuthorStatusChangeTGLCMC RemoveBatchesFromTodo (%v) error %v", needDeleteBatchIDs, err)
		return
	}

	if validVendorID > 0 {
		// 更新batch对应vendor id
		if err = s.dao.UpdateBatchVendorIDsByBatchIDs(needDeleteBatchIDs, validVendorID); err != nil {
			log.Error("Service: handleAuthorStatusChangeTGLCMC UpdateBatchVendorIDsByBatchIDs (%v, %d) error %v", needDeleteBatchIDs, validVendorID, err)
			return
		}
		if toAddAuthor, _err := s.dao.GetAuthorByMID(validVendorID, author.MID); _err != nil {
			log.Error("Service: handleAuthorStatusChangeTGLCMC GetAuthorByMID (%d, %d) error %v", validVendorID, author.MID, _err)
			err = _err
			return
		} else if toAddAuthor == nil {
			log.Error("Service: handleAuthorStatusChangeTGLCMC GetAuthorByMID (%d, %d) 找不到对应作者")
			err = ecode.AuthorNotFound
		} else {
			if toAddAuthorPushes, _err := s.dao.GetAuthorPushesByVendorIDs([]int64{validVendorID}); _err != nil {
				log.Error("Service: handleAuthorStatusChangeTGLCMC GetAuthorPushesByVendorIDs %d error %v", validVendorID, _err)
				err = _err
				return
			} else if len(toAddAuthorPushes) == 0 {
				log.Warn("Service: handleAuthorStatusChangeTGLCMC GetAuthorPushesByVendorIDs %d 获取作者推送为空")
			} else {
				rel := &model.ArchivePushBatchAuthorPushRel{
					AuthorPushID: toAddAuthorPushes[0].ID,
					AuthorID:     toAddAuthor.ID,
				}
				if _, err = s.dao.AddBatchAuthorPushRel(rel); err != nil {
					log.Error("Service: GetAuthorPushesByVendorIDs AddBatchAuthorPushRel %+v error %v", rel, err)
				}
			}
		}
	}
	return
}

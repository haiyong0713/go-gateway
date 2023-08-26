package service

import (
	"context"
	"encoding/json"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"strings"
	"sync"
	"time"

	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/ecode"
)

// GetAllAuthorPushes 查询所有作者维度推送
func (s *Service) GetAllAuthorPushes() (list []*model.ArchivePushAuthorPushX, err error) {
	list = make([]*model.ArchivePushAuthorPushX, 0)
	var rawList []*model.ArchivePushAuthorPush
	if rawList, err = s.dao.GetAllAuthorPushes(); err != nil {
		log.Error("Service: GetAllAuthorPushes GetAllAuthorPushes Error (%v)", err)
		return nil, err
	}
	if len(rawList) == 0 {
		return
	}
	for _, rawPush := range rawList {
		_rawPush := rawPush
		push := &model.ArchivePushAuthorPushX{
			ArchivePushAuthorPush: *_rawPush,
			Status:                api.AuthorPushStatus_Enum_name[int32(_rawPush.Status)],
		}
		pushConditions := make([]*model.ArchivePushAuthorPushCondition, 0)
		if _err := json.Unmarshal([]byte(_rawPush.PushConditions), &pushConditions); _err != nil {
			log.Error("Service: GetAllAuthorPushes Unmarshal (%s) error %v", _rawPush.PushConditions, _err)
		} else {
			push.PushConditions = pushConditions
		}
		list = append(list, push)
	}
	return
}

// GetAuthorPushesByPage 查询作者维度推送
func (s *Service) GetAuthorPushesByPage(ids []int64, cuser string, pn int, ps int) (list []*model.ArchivePushAuthorPushX, total int64, err error) {
	list = make([]*model.ArchivePushAuthorPushX, 0)
	var rawList []*model.ArchivePushAuthorPush
	if rawList, total, err = s.dao.GetAuthorPushesByPage(ids, cuser, pn, ps); err != nil {
		log.Error("Service: GetAuthorPushesByPage GetAuthorPushesByPage Error (%v)", err)
		return nil, 0, err
	}
	if len(rawList) == 0 {
		return
	}
	for _, rawPush := range rawList {
		_rawPush := rawPush
		push := &model.ArchivePushAuthorPushX{
			ArchivePushAuthorPush: *_rawPush,
			Status:                api.AuthorPushStatus_Enum_name[int32(_rawPush.Status)],
		}
		list = append(list, push)
	}
	return
}

// GetAuthorPushFullsByPage 查询作者维度推送
func (s *Service) GetAuthorPushFullsByPage(ids []int64, cuser string, pn int, ps int) (list []*model.ArchivePushAuthorPushFull, total int64, err error) {
	list = make([]*model.ArchivePushAuthorPushFull, 0)
	var (
		rawList []*model.ArchivePushAuthorPush
		relList []*model.ArchivePushBatchAuthorPushRelWithAuthorAndBatch
		relMap  map[int64]map[int64]*model.AuthorWithBatch
	)
	if rawList, total, err = s.dao.GetAuthorPushesByPage(ids, cuser, pn, ps); err != nil {
		log.Error("Service: GetAuthorPushesByPage GetAuthorPushesByPage Error (%v)", err)
		return nil, 0, err
	}
	if len(rawList) == 0 {
		return
	}
	pushIDs := make([]int64, 0, len(rawList))
	for _, rawPush := range rawList {
		_rawPush := rawPush
		pushIDs = append(pushIDs, _rawPush.ID)
	}
	// 获取作者推送关系数据
	if relList, err = s.dao.GetBatchAuthorPushRelsWithAuthorAndBatchByPushIDs(pushIDs); err != nil {
		log.Error("Service: GetAuthorPushFullsByPage GetBatchAuthorPushRelsByPushIDs(%v) error %v", pushIDs, err)
		return
	}
	// 根据推送关系构造推送详情map
	if relMap, err = s.BuildBatchAuthorPushMap(relList); err != nil {
		log.Error("Service: GetAuthorPushFullsByPage BuildBatchAuthorPushMap error %v", err)
		return
	}
	for _, rawPush := range rawList {
		_rawPush := rawPush
		push := &model.ArchivePushAuthorPushFull{
			ArchivePushAuthorPush: *_rawPush,
			Status:                api.AuthorPushStatus_Enum_name[int32(_rawPush.Status)],
		}
		if vendor, _err := s.GetVendorByID(_rawPush.VendorID); _err != nil || vendor.ID == 0 {
			log.Error("Service: GetAuthorPushFullsByPage GetVendorByID(%d) error %v", _rawPush.VendorID, err)
		} else {
			push.VendorName = vendor.Name
		}
		if details, exists := relMap[_rawPush.ID]; exists {
			push.AuthorsWithBatches = details
		}
		pushConditions := make([]*model.ArchivePushAuthorPushCondition, 0)
		if _err := json.Unmarshal([]byte(_rawPush.PushConditions), &pushConditions); _err != nil {
			log.Error("Service: GetAuthorPushFullsByPage Unmarshal %s error %v", _rawPush.PushConditions, _err)
		} else {
			for _, condition := range pushConditions {
				switch condition.Type {
				case model.ArchivePushAuthorPushConditionTypeAuthorized:
					push.Authorized = condition.Value
					break
				case model.ArchivePushAuthorPushConditionTypeBinded:
					push.Binded = condition.Value
					break
				case model.ArchivePushAuthorPushConditionTypeVerified:
					push.Verified = condition.Value
					break
				}
			}
		}
		list = append(list, push)
	}
	return
}

// GetActiveAuthorPushByMID 根据MID和推送厂商获取作者推送数据
func (s *Service) GetActiveAuthorPushByMID(vendorID int64, mid int64) (res *model.ArchivePushAuthorPushWithAuthors, err error) {
	if vendorID == 0 || mid == 0 {
		return nil, xecode.RequestErr
	}

	var author *model.ArchivePushAuthorX
	if authors, _err := s.GetAuthorsByUser(vendorID, mid, ""); _err != nil {
		log.Error("Service: GetActiveAuthorPushByMID GetAuthorsByUser (%d, %d, '') error %v", vendorID, mid, _err)
		return
	} else if len(authors) == 0 {
		log.Warn("Service: GetActiveAuthorPushByMID GetAuthorsByUser (%d, %d, '') 未找到对应作者", vendorID, mid)
		err = ecode.AuthorNotFound
		return
	} else {
		author = authors[0]
	}
	if pushes, _err := s.dao.GetActiveAuthorPushesByAuthorIDs([]int64{author.ID}); _err != nil {
		log.Error("Service: GetActiveAuthorPushByMID GetActiveAuthorPushesByAuthorIDs (%d) error %v", author.ID)
		return
	} else if len(pushes) == 0 {
		log.Error("Service: GetActiveAuthorPushByMID GetActiveAuthorPushesByAuthorIDs (%d) 未找到对应推送")
	} else {
		res = pushes[0]
	}

	return
}

// BuildBatchAuthorPushMap 根据推送关系构造推送详情map
//
// map[authorPushID]map[authorID]*model.AuthorWithBatch
func (s *Service) BuildBatchAuthorPushMap(relList []*model.ArchivePushBatchAuthorPushRelWithAuthorAndBatch) (res map[int64]map[int64]*model.AuthorWithBatch, err error) {
	res = make(map[int64]map[int64]*model.AuthorWithBatch)
	if len(relList) == 0 {
		return
	}
	for _, rel := range relList {
		_rel := rel
		if _, exists := res[_rel.AuthorPushID]; !exists {
			res[_rel.AuthorPushID] = make(map[int64]*model.AuthorWithBatch)
		}
		detail := &model.AuthorWithBatch{
			BatchAuthorPushRelID: _rel.ID,
			AuthorID:             _rel.AuthorID,
			AuthorNickname:       _rel.AuthorNickname,
			BatchID:              _rel.BatchID,
			BatchPushType:        _rel.BatchPushType,
		}
		res[_rel.AuthorPushID][_rel.AuthorID] = detail
	}

	return
}

// AddAuthorPush 新增作者推送完整流程
func (s *Service) AddAuthorPush(vendorID int64, rawTags string, delayMinutes int32, pushConditions []*model.ArchivePushAuthorPushCondition, pushHistoryArchives bool, username string, uid int64) (id int64, err error) {
	if vendorID == 0 {
		err = xecode.RequestErr
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

	// 检查推送厂商的推送模型是否已建立，若已建立则不可重复建立
	if addedAuthorPushes, _err := s.dao.GetActiveAuthorPushesByVendorIDs([]int64{vendorID}); _err != nil {
		log.Error("Service: AddAuthorPush GetAuthorPushByVendorIDs %d error %v", vendorID, _err)
		err = _err
		return
	} else if len(addedAuthorPushes) > 0 {
		err = ecode.AuthorPushExisting
		return
	}

	var (
		tags               = make([]string, 0)
		authors            []*model.ArchivePushAuthor
		authorIDs          = make([]int64, 0)
		authorsMap         = make(map[int64]*model.ArchivePushAuthor)
		pushableAuthors    []*model.ArchivePushAuthor
		pushableAuthorsMap = make(map[int64]*model.ArchivePushAuthor)
		pushedAuthorPushes []*model.ArchivePushAuthorPushWithAuthors
		archivesMap        map[int64][]*archiveGRPC.Arc
		toPushBVIDsMap     map[int64][]string
		batchIDs           = make([]int64, 0)
	)

	// split tags
	if rawTags != "" {
		tags = strings.Split(rawTags, ",")
		// 去除空格
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// 获取vendor所有已完成绑定作者
	if len(pushConditions) == 0 {
		pushConditions = make([]*model.ArchivePushAuthorPushCondition, 0)
	}
	if authors, err = s.GetPushableAuthors(vendorID, pushConditions); err != nil {
		log.Error("Service: AddAuthorPush(%d, %s) GetPushableAuthors error %v", vendorID, tags, err)
		return
	}
	if len(authors) > 0 {
		for _, author := range authors {
			_author := author
			authorIDs = append(authorIDs, _author.ID)
			authorsMap[_author.ID] = _author
		}

		// 根据作者IDs获取已推送的作者
		if pushedAuthorPushes, err = s.dao.GetActiveAuthorPushesByAuthorIDs(authorIDs); err != nil {
			log.Error("Service: AddAuthorPush(%d, %s) GetActiveAuthorPushesByAuthorIDs(%v) error %v", vendorID, tags, authorIDs, err)
			return
		}
		// 过滤已推送的作者
		for _, author := range authors {
			_author := author
			exists := false
			for _, authorPush := range pushedAuthorPushes {
				_authorPush := authorPush
				if _authorPush.VendorID == _author.PushVendorID {
					for _, authorPushAuthor := range _authorPush.Authors {
						_authorPushAuthor := authorPushAuthor
						if _authorPushAuthor.MID == _author.MID {
							exists = true
							break
						}
					}
					if exists {
						break
					}
				}
			}
			if !exists {
				pushableAuthors = append(pushableAuthors, _author)
				pushableAuthorsMap[_author.ID] = _author
			}
		}
		if len(pushableAuthors) == 0 {
			err = ecode.NoValidAuthorsToPush
			return
		}
	}

	// 新增作者推送数据
	if id, err = s.AddRawAuthorPush(vendorID, rawTags, delayMinutes, api.AuthorPushStatus_EFFECTIVE, pushConditions, username); err != nil {
		log.Error("Service: AddAuthorPush AddRawAuthorPush(%d, %s) error %v", vendorID, rawTags, err)
		return
	}

	// 新增作者推送稿件推送rel数据
	for _, author := range pushableAuthors {
		_author := author
		rel := &model.ArchivePushBatchAuthorPushRel{
			AuthorPushID: id,
			AuthorID:     _author.ID,
			CUser:        username,
		}
		if rel, _err := s.dao.AddBatchAuthorPushRel(rel); _err != nil {
			log.Error("Service: AddAuthorPush AddBatchAuthorPushRel (%+v) error %v", rel, _err)
		}
	}

	// 是否需要推送作者历史稿件
	if pushHistoryArchives {
		// 获取相关所有稿件
		if archivesMap, err = s.GetArcsByAuthorsAndTags(pushableAuthors, tags); err != nil {
			log.Error("Service: AddAuthorPush(%d, %s) GetArcsByAuthorsAndTags error %v", vendorID, tags, err)
			return
		}

		// 根据作者生成batch与bvid列表
		if _, toPushBVIDsMap, err = s.GenerateAuthorPushingMaps(vendorID, archivesMap, username); err != nil {
			log.Error("Service: AddAuthorPush(%d, %s) GenerateAuthorPushingMaps error %v", vendorID, tags, err)
			return
		}

		// 将bvid放入作者稿件白名单
		for authorID := range toPushBVIDsMap {
			if len(toPushBVIDsMap[authorID]) > 0 {
				if author, exists := authorsMap[authorID]; exists {
					if _err := s.PutAuthorBVIDsForWhiteList(vendorID, author.MID, toPushBVIDsMap[authorID]); _err != nil {
						log.Error("Service: AddAuthorPush PutAuthorBVIDsForWhiteList(%d, %d, %v) error %v", vendorID, author.MID, toPushBVIDsMap[authorID], _err)
					}
				}
			}
		}
	}

	// 行为日志
	if _err := s.AddAuthorPushAuditLog(vendorID, rawTags, delayMinutes, id, pushableAuthors, batchIDs, username, uid); _err != nil {
		log.Error("Service: AddAuthorPush AddAuthorPushAuditLog(%d, %s, %d, %d, %v, %v, %s, %d) error %v", vendorID, tags, delayMinutes, id, pushableAuthors, batchIDs, username, uid, _err)
	}

	return
}

// AddRawAuthorPush 新增作者推送数据
func (s *Service) AddRawAuthorPush(vendorID int64, tags string, delayMinutes int32, status api.AuthorPushStatus_Enum, pushConditions []*model.ArchivePushAuthorPushCondition, username string) (id int64, err error) {
	if vendorID == 0 {
		err = xecode.RequestErr
		return
	}
	var (
		addedPush         *model.ArchivePushAuthorPush
		pushConditionsStr = "[]"
	)

	if pushConditionsBytes, _err := json.Marshal(pushConditions); _err != nil {
		log.Error("Service: AddAuthorPush Marshal %+v error %v", pushConditions, _err)
		err = _err
		return
	} else {
		pushConditionsStr = string(pushConditionsBytes)
	}

	toAddModel := model.ArchivePushAuthorPush{
		VendorID:       vendorID,
		Tags:           tags,
		DelayMinutes:   delayMinutes,
		Status:         int(status),
		PushConditions: pushConditionsStr,
		CUser:          username,
	}

	if addedPush, err = s.dao.AddAuthorPush(toAddModel); err != nil {
		log.Error("Service: AddRawAuthorPush(%d, %s, %d) error %v", vendorID, tags, delayMinutes, err)
		return
	}
	id = addedPush.ID

	return
}

// GenerateAuthorPushingMaps 根据稿件生成作者维度的推送batch
func (s *Service) GenerateAuthorPushingMaps(vendorID int64, archivesMap map[int64][]*archiveGRPC.Arc, username string) (resArchiveMap map[int64]*model.ArchivePushBatch, resBVIDMap map[int64][]string, err error) {
	if vendorID == 0 {
		err = xecode.RequestErr
		return
	}
	if len(archivesMap) == 0 {
		return
	}
	resArchiveMap = make(map[int64]*model.ArchivePushBatch)
	resBVIDMap = make(map[int64][]string)
	for authorID := range archivesMap {
		_authorID := authorID
		resArchiveMap[_authorID] = &model.ArchivePushBatch{
			PushType:     model.PushTypeAuthor,
			PushVendorID: vendorID,
			FileURL:      "作者批次推送",
			PushStatus:   api.ArchivePushBatchPushStatus_TO_PUSH,
			CUser:        username,
			CTime:        xtime.Time(time.Now().Unix()),
		}
		resBVIDMap[_authorID] = make([]string, 0)
		for _, arc := range archivesMap[_authorID] {
			_arc := arc
			var bvid string
			if bvid, err = util.AvToBv(_arc.Aid); err != nil {
				log.Error("Service: GenerateAuthorPushingMaps AvToBv(%d) error %v", _arc.Aid, err)
				return
			}
			resBVIDMap[_authorID] = append(resBVIDMap[_authorID], bvid)
		}
	}

	return
}

// AddRawBatchAuthorPushRel 添加作者稿件推送数据
func (s *Service) AddRawBatchAuthorPushRel(rel model.ArchivePushBatchAuthorPushRel) (id int64, err error) {
	if rel.AuthorID == 0 || rel.AuthorPushID == 0 {
		err = xecode.RequestErr
		return
	}
	var addedRel *model.ArchivePushBatchAuthorPushRel
	if addedRel, err = s.dao.AddBatchAuthorPushRel(&rel); err != nil {
		log.Error("Service: AddRawBatchAuthorPushRel %+v error %v", rel, err)
		return
	}
	id = addedRel.ID

	return
}

// DoPushAuthorPushes 进行稿件batch推送并返回推送的batch ids
func (s *Service) DoPushAuthorPushes(authorPushID int64, toPushBatchesMap map[int64]*model.ArchivePushBatch, toPushBVIDsMap map[int64][]string, delayMinutes int32, username string, uid int64) (batchIDs []int64, err error) {
	if authorPushID == 0 || len(toPushBatchesMap) == 0 || len(toPushBVIDsMap) == 0 {
		err = xecode.RequestErr
		return
	}
	batchIDs = make([]int64, 0)
	eg := errgroup.WithContext(context.Background())
	var lock sync.Mutex
	for authorID := range toPushBatchesMap {
		_authorID := authorID
		batch := toPushBatchesMap[_authorID]
		if bvids, exists := toPushBVIDsMap[_authorID]; !exists || len(toPushBVIDsMap[_authorID]) == 0 {
			log.Warn("Service: AddAuthorPush authorID(%d) 没有可推送的稿件", _authorID)
			rel := model.ArchivePushBatchAuthorPushRel{
				AuthorPushID: authorPushID,
				AuthorID:     _authorID,
				BatchID:      0,
				CUser:        username,
				CTime:        xtime.Time(time.Now().Unix()),
			}
			if _, err = s.AddRawBatchAuthorPushRel(rel); err != nil {
				log.Error("Service: AddAuthorPush AddRawBatchAuthorPushRel(%+v) error %v", rel, err)
			}
			continue
		} else {
			eg.Go(func(ctx context.Context) error {
				if batchID, err := s.DoBatchPush(batch, bvids, delayMinutes, username, uid); err != nil {
					log.Error("Service: AddAuthorPush DoBatchPush(%+v, %v) error %v", batch, bvids, err)
					return err
				} else {
					lock.Lock()
					batchIDs = append(batchIDs, batchID)
					lock.Unlock()
					rel := model.ArchivePushBatchAuthorPushRel{
						AuthorPushID: authorPushID,
						AuthorID:     _authorID,
						BatchID:      batchID,
						CUser:        username,
						CTime:        xtime.Time(time.Now().Unix()),
					}
					if _, err = s.AddRawBatchAuthorPushRel(rel); err != nil {
						log.Error("Service: AddAuthorPush AddRawBatchAuthorPushRel(%+v) error %v", rel, err)
						return err
					}
				}
				return nil
			})
		}
	}
	err = eg.Wait()
	return
}

// AddAuthorPushAuditLog 添加作者推送行为日志
func (s *Service) AddAuthorPushAuditLog(vendorID int64, tags string, delayMinutes int32, id int64, authors []*model.ArchivePushAuthor, batchIDs []int64, username string, uid int64) (err error) {
	now := time.Now()
	index := []interface{}{
		id,           // int0
		now.Unix(),   // int1
		delayMinutes, // int2
		0,            // int3
		tags,         // str0
	}

	content := map[string]interface{}{
		"authors":  authors,
		"batchIDs": batchIDs,
	}

	params := &model.AuditLogInitParams{
		UName:    username,
		UID:      uid,
		Business: model.BusinessIDAuthorPush,
		Type:     int(vendorID),
		OID:      id,
		Action:   "添加作者推送",
		Content:  content,
		CTime:    now,
		Index:    index,
	}
	if _err := s.dao.AddAuditLog(params); _err != nil {
		log.Error("Service: AddAuthorPushAuditLog Error (%v)", _err)
		return _err
	}
	return
}

// EditAuthorPush 更新作者推送完整流程
func (s *Service) EditAuthorPush(id int64, tags string, delayMinutes int32, pushConditions []*model.ArchivePushAuthorPushCondition, username string, uid int64) (err error) {
	if id == 0 {
		return xecode.RequestErr
	}

	var (
		toUpdatePush        *model.ArchivePushAuthorPushX
		batchAuthorPushRels = make([]*model.ArchivePushBatchAuthorPushRel, 0)
		toUpdateBatchIDs    = make([]int64, 0)
	)
	if gotPushes, _, _err := s.GetAuthorPushesByPage([]int64{id}, "", 1, 1); _err != nil {
		log.Error("Service: EditAuthorPush (%d, %s, %d) error %v", id, tags, delayMinutes, _err)
		err = _err
		return
	} else if len(gotPushes) == 0 {
		err = ecode.AuthorPushNotFound
		return
	} else {
		toUpdatePush = gotPushes[0]
	}

	// update作者推送数据
	if err = s.UpdateRawAuthorPush(id, tags, delayMinutes, 0, pushConditions, username); err != nil {
		log.Error("Service: EditAuthorPush (%d, %s, %d) error %v", id, tags, delayMinutes, err)
		return
	}

	// 获取所有对应作者推送的待推送batch ids
	if batchAuthorPushRels, err = s.dao.GetBatchAuthorPushRelsByPushIDs([]int64{id}); err != nil {
		log.Error("Service: EditAuthorPush GetBatchAuthorPushRelsByPushIDs %d error %v", id, err)
	} else if len(batchAuthorPushRels) == 0 {
		log.Warn("Service: EditAuthorPush GetBatchAuthorPushRelsByPushIDs %d 获取batch作者推送关系为空", id)
	}
	for _, rel := range batchAuthorPushRels {
		_rel := rel
		toUpdateBatchIDs = append(toUpdateBatchIDs, _rel.BatchID)
	}

	// 更新待推送时间统一更新成当前时间+delayMinutes
	if err = s.dao.UpdateBatchesPushTimeForTodo(toUpdateBatchIDs, xtime.Time(time.Now().Add(time.Duration(delayMinutes)*time.Minute).Unix())); err != nil {
		log.Error("Service: EditAuthorPush UpdateBatchExecTimeForTodo %v, %d error %v", toUpdateBatchIDs, delayMinutes, err)
	}

	// 行为日志
	if _err := s.AddAuthorEditAuditLog(toUpdatePush.VendorID, tags, delayMinutes, 0, id, "", nil, nil, username, uid); _err != nil {
		log.Error("Service: EditAuthorPush AddAuthorEditAuditLog error %v", _err)
	}

	return
}

// UpdateRawAuthorPush 更新作者推送数据
func (s *Service) UpdateRawAuthorPush(id int64, tags string, delayMinutes int32, status api.AuthorPushStatus_Enum, pushConditions []*model.ArchivePushAuthorPushCondition, username string) (err error) {
	if id == 0 {
		err = xecode.RequestErr
		return
	}
	var (
		updatedPush *model.ArchivePushAuthorPush
		pushes      []*model.ArchivePushAuthorPush
	)
	if pushes, _, err = s.dao.GetAuthorPushesByPage([]int64{id}, "", 1, 1); err != nil {
		log.Error("Service: UpdateRawAuthorPush GetAuthorPushesByPage(%d) error %v", id, err)
		return
	} else if len(pushes) == 0 {
		err = ecode.AuthorPushNotFound
		return
	}
	updatedPush = pushes[0]
	updatedPush.Tags = tags
	updatedPush.DelayMinutes = delayMinutes
	updatedPush.Status = int(status)
	updatedPush.MUser = username
	updatedPush.MTime = xtime.Time(time.Now().Unix())

	oldPushConditions := make([]*model.ArchivePushAuthorPushCondition, 0)
	if updatedPush.PushConditions != "" {
		if err = json.Unmarshal([]byte(updatedPush.PushConditions), &oldPushConditions); err != nil {
			log.Error("Service: UpdateRawAuthorPush Unmarshal %s error %v", updatedPush.PushConditions, err)
			return
		}
	}
	conditionsSameWithOld := true
	if len(oldPushConditions) != len(pushConditions) {
		conditionsSameWithOld = false
	} else {
		for _, oldCondition := range oldPushConditions {
			for _, condition := range pushConditions {
				if condition.Type == oldCondition.Type {
					if condition.Op != oldCondition.Op || condition.Value != oldCondition.Value {
						conditionsSameWithOld = false
						break
					}
				}
			}
			if !conditionsSameWithOld {
				break
			}
		}
	}
	if !conditionsSameWithOld {
		if err = s.UpdateAuthorPushConditions(id, updatedPush.VendorID, pushConditions, username); err != nil {
			log.Error("Service: UpdateRawAuthorPush UpdateAuthorPushConditions (%d, %d, %v, %s) error %v", id, updatedPush.VendorID, pushConditions, username)
			return
		}
	}

	if len(pushConditions) > 0 {
		if pushConditionsBytes, _err := json.Marshal(pushConditions); _err != nil {
			log.Error("Service: AddAuthorPush Marshal %+v error %v", pushConditions, _err)
			err = _err
			return
		} else {
			updatedPush.PushConditions = string(pushConditionsBytes)
		}
	}

	if err = s.dao.UpdateAuthorPushByID(updatedPush); err != nil {
		log.Error("Service: UpdateRawAuthorPush(%d, %s, %d) UpdateAuthorPushByID error %v", id, tags, delayMinutes, err)
		return
	}

	return
}

func (s *Service) UpdateAuthorPushConditions(id int64, vendorID int64, pushConditions []*model.ArchivePushAuthorPushCondition, username string) (err error) {
	var (
		authors                       []*model.ArchivePushAuthor
		authorMIDs                    = make([]int64, 0)
		toRemoveBatchAuthorPushRels   = make([]*model.ArchivePushBatchAuthorPushRel, 0)
		toRemoveBatchIDs              = make([]int64, 0)
		toRemoveBatchAuthorPushRelIDs = make([]int64, 0)
		toAddAuthorIDs                = make([]int64, 0)
		toAddAuthorsMap               = make(map[int64]*model.ArchivePushAuthor)
		pushableAuthors               []*model.ArchivePushAuthor
		pushableAuthorsMap            = make(map[int64]*model.ArchivePushAuthor)
		pushedAuthorPushes            []*model.ArchivePushAuthorPushWithAuthors
	)
	if authors, err = s.GetAuthorsByPushID(id); err != nil {
		log.Error("Service: InactivateAuthorPush(%d) GetAuthorsByPush error %v", id, err)
		return
	}
	if len(authors) > 0 {
		for _, author := range authors {
			_author := author
			authorMIDs = append(authorMIDs, _author.MID)
		}

		// 获取所有batch作者推送rels
		if toRemoveBatchAuthorPushRels, err = s.dao.GetBatchAuthorPushRelsByPushIDs([]int64{id}); err != nil {
			log.Error("Service: InactivateAuthorPush GetBatchAuthorPushRelsByPushIDs %d error %v", id, err)
			return
		} else if len(toRemoveBatchAuthorPushRels) == 0 {
			log.Warn("Service: InactivateAuthorPush GetBatchAuthorPushRelsByPushIDs %d 获取数据为空", id)
		}
		for _, rel := range toRemoveBatchAuthorPushRels {
			_rel := rel
			toRemoveBatchAuthorPushRelIDs = append(toRemoveBatchAuthorPushRelIDs, _rel.ID)
			toRemoveBatchIDs = append(toRemoveBatchIDs, _rel.BatchID)
		}

		// 移除所有batch作者推送rels
		if err = s.dao.DeleteBatchAuthorPushRelsByIDs(toRemoveBatchAuthorPushRelIDs, username); err != nil {
			log.Error("Service: InactivateAuthorPush DeleteBatchAuthorPushRelsByIDs %v error %v", toRemoveBatchAuthorPushRelIDs, err)
		}

		// 移除所有batch from 待推送池
		if err = s.dao.RemoveBatchesFromTodo(toRemoveBatchIDs); err != nil {
			log.Error("Service: InactivateAuthorPush RemoveBatchesFromTodo %v error %v", toRemoveBatchIDs, err)
		}
	}

	// 重新添加符合条件的作者
	if authors, err = s.GetPushableAuthors(vendorID, pushConditions); err != nil {
		log.Error("Service: AddAuthorPush(%d) GetPushableAuthors error %v", vendorID, err)
		return
	}
	if len(authors) > 0 {
		for _, author := range authors {
			_author := author
			toAddAuthorIDs = append(toAddAuthorIDs, _author.ID)
			toAddAuthorsMap[_author.ID] = _author
		}

		// 根据作者IDs获取已推送的作者
		if pushedAuthorPushes, err = s.dao.GetActiveAuthorPushesByAuthorIDs(toAddAuthorIDs); err != nil {
			log.Error("Service: AddAuthorPush(%d) GetActiveAuthorPushesByAuthorIDs(%v) error %v", vendorID, toAddAuthorIDs, err)
			return
		}
		// 过滤已推送的作者
		for _, author := range authors {
			_author := author
			exists := false
			for _, authorPush := range pushedAuthorPushes {
				_authorPush := authorPush
				if _authorPush.VendorID == _author.PushVendorID {
					for _, authorPushAuthor := range _authorPush.Authors {
						_authorPushAuthor := authorPushAuthor
						if _authorPushAuthor.MID == _author.MID {
							exists = true
							break
						}
					}
					if exists {
						break
					}
				}
			}
			if !exists {
				pushableAuthors = append(pushableAuthors, _author)
				pushableAuthorsMap[_author.ID] = _author
			}
		}

		if len(pushableAuthors) > 0 {
			// 新增作者推送稿件推送rel数据
			for _, author := range pushableAuthors {
				_author := author
				rel := &model.ArchivePushBatchAuthorPushRel{
					AuthorPushID: id,
					AuthorID:     _author.ID,
					CUser:        username,
				}
				if rel, _err := s.dao.AddBatchAuthorPushRel(rel); _err != nil {
					log.Error("Service: AddAuthorPush AddBatchAuthorPushRel (%+v) error %v", rel, _err)
				}
			}
		}
	}

	return
}

// AddAuthorPushAuditLog 添加作者推送行为日志
func (s *Service) AddAuthorEditAuditLog(vendorID int64, tags string, delayMinutes int32, status api.AuthorPushStatus_Enum, id int64, reason string, authors []*model.ArchivePushAuthor, batchIDs []int64, username string, uid int64) (err error) {
	now := time.Now()
	index := []interface{}{
		id,           // int0
		now.Unix(),   // int1
		delayMinutes, // int2
		status,       // int3
		tags,         // str0
		reason,       // str1
	}

	content := map[string]interface{}{
		"authors":  authors,
		"batchIDs": batchIDs,
	}

	params := &model.AuditLogInitParams{
		UName:    username,
		UID:      uid,
		Business: model.BusinessIDAuthorPush,
		Type:     int(vendorID),
		OID:      id,
		Action:   "修改作者推送",
		Content:  content,
		CTime:    now,
		Index:    index,
	}
	if _err := s.dao.AddAuditLog(params); _err != nil {
		log.Error("Service: AddAuthorPushAuditLog Error (%v)", err)
		return _err
	}
	return
}

// InactivateAuthorPush 作者推送失效
func (s *Service) InactivateAuthorPush(id int64, reason string, needWithdrawHistoryArcs bool, username string, uid int64) (err error) {
	if id == 0 {
		return xecode.RequestErr
	}

	var (
		toUpdatePush                  *model.ArchivePushAuthorPushX
		authors                       []*model.ArchivePushAuthor
		authorMIDs                    = make([]int64, 0)
		toRemoveBatchAuthorPushRels   = make([]*model.ArchivePushBatchAuthorPushRel, 0)
		toRemoveBatchIDs              = make([]int64, 0)
		toRemoveBatchAuthorPushRelIDs = make([]int64, 0)
	)
	// 检查并获取当前
	if gotPushes, _, _err := s.GetAuthorPushesByPage([]int64{id}, "", 1, 1); _err != nil {
		log.Error("Service: InactivateAuthorPush (%d, %s) error %v", id, reason, _err)
		err = _err
		return
	} else if len(gotPushes) == 0 {
		err = ecode.AuthorPushNotFound
		return
	} else {
		toUpdatePush = gotPushes[0]
	}
	if authors, err = s.GetAuthorsByPushID(id); err != nil {
		log.Error("Service: InactivateAuthorPush(%d, %s) GetAuthorsByPush error %v", id, reason, err)
		return
	} else if len(authors) == 0 {
		log.Warn("Service: InactivateAuthorPush(%d, %s) 没有关联的作者要处理", id, reason)
		return
	}
	for _, author := range authors {
		_author := author
		authorMIDs = append(authorMIDs, _author.MID)
	}

	if needWithdrawHistoryArcs {
		// 下架作者稿件
		eg := errgroup.WithContext(context.Background())
		for _, mid := range authorMIDs {
			_mid := mid
			eg.Go(func(_ context.Context) error {
				if _err := s.WithdrawArchivesByAuthor(toUpdatePush.VendorID, _mid, reason, username, uid); _err != nil {
					log.Error("Service: InactivateAuthorPush WithdrawArchivesByAuthor (%d, %d, %s ,%s) error %v", toUpdatePush.VendorID, _mid, reason, username, _err)
				}
				return nil
			})
		}
		if err = eg.Wait(); err != nil {
			log.Error("Service: InactivateAuthorPush(%d, %s) WithdrawArchivesByAuthor error %v", id, reason, err)
		}
	}

	// 修改push status
	if err = s.UpdateRawAuthorPush(id, "", 0, api.AuthorPushStatus_CANCELED, nil, username); err != nil {
		log.Error("Service: InactivateAuthorPush (%d, %s) error %v", id, reason, err)
		return
	}

	// 获取所有batch作者推送rels
	if toRemoveBatchAuthorPushRels, err = s.dao.GetBatchAuthorPushRelsByPushIDs([]int64{id}); err != nil {
		log.Error("Service: InactivateAuthorPush GetBatchAuthorPushRelsByPushIDs %d error %v", id, err)
		return
	} else if len(toRemoveBatchAuthorPushRels) == 0 {
		log.Warn("Service: InactivateAuthorPush GetBatchAuthorPushRelsByPushIDs %d 获取数据为空", id)
	}
	for _, rel := range toRemoveBatchAuthorPushRels {
		_rel := rel
		toRemoveBatchAuthorPushRelIDs = append(toRemoveBatchAuthorPushRelIDs, _rel.ID)
		toRemoveBatchIDs = append(toRemoveBatchIDs, _rel.BatchID)
	}

	// 移除所有batch作者推送rels
	if err = s.dao.DeleteBatchAuthorPushRelsByIDs(toRemoveBatchAuthorPushRelIDs, username); err != nil {
		log.Error("Service: InactivateAuthorPush DeleteBatchAuthorPushRelsByIDs %v error %v", toRemoveBatchAuthorPushRelIDs, err)
	}

	// 移除所有batch from 待推送池
	if err = s.dao.RemoveBatchesFromTodo(toRemoveBatchIDs); err != nil {
		log.Error("Service: InactivateAuthorPush RemoveBatchesFromTodo %v error %v", toRemoveBatchIDs, err)
	}

	// 行为日志
	if _err := s.AddAuthorEditAuditLog(toUpdatePush.VendorID, "", 0, api.AuthorPushStatus_CANCELED, id, reason, nil, nil, username, uid); _err != nil {
		log.Error("Service: EditAuthorPush AddAuthorEditAuditLog error %v", _err)
	}

	return
}

// GetAuthorsByPushID 根据author push id获取所有关联作者
func (s *Service) GetAuthorsByPushID(authorPushID int64) (resAuthors []*model.ArchivePushAuthor, err error) {
	if authorPushID == 0 {
		err = xecode.RequestErr
		return
	}

	var (
		authorIDs = make([]int64, 0)
	)

	resAuthors = make([]*model.ArchivePushAuthor, 0)
	if rels, _err := s.dao.GetBatchAuthorPushRelsByPushIDs([]int64{authorPushID}); _err != nil {
		log.Error("Service: GetAuthorsByPushID GetBatchAuthorPushRelsByPushIDs %v", _err)
		err = _err
		return
	} else if len(rels) == 0 {
		return
	} else {
		for _, rel := range rels {
			_rel := rel
			authorIDs = append(authorIDs, _rel.AuthorID)
		}
	}
	if resAuthors, err = s.GetRawAuthorsByIDs(authorIDs); err != nil {
		log.Error("Service: GetAuthorsByPushID %v error %v", authorPushID, err)
	}

	return
}

// RemoveBatchAuthorPushRelsByAuthor 删除作者推送数据
func (s *Service) RemoveBatchAuthorPushRelsByAuthor(authorID int64, username string) (err error) {
	if authorID == 0 {
		return xecode.RequestErr
	}

	if err = s.dao.DeleteBatchAuthorPushRelsByAuthorIDs([]int64{authorID}, username); err != nil {
		log.Error("Service: RemoveBatchAuthorPushRelsByAuthor (%d) DeleteBatchAuthorPushRelsByAuthorIDs error %v", authorID, err)
	}
	return
}

// ValidateAuthorPushConditionsWithAuthor 校验作者状态是否符合推送条件
func (s *Service) ValidateAuthorPushConditionsWithAuthor(conditions []*model.ArchivePushAuthorPushCondition, author model.ArchivePushAuthor) (valid bool) {
	if len(conditions) == 0 {
		return true
	}
	if author.ID == 0 {
		return false
	}
	valid = true
	falseCount := 0
	for _, condition := range conditions {
		_condition := condition
		switch _condition.Type {
		case model.ArchivePushAuthorPushConditionTypeAuthorized:
			switch _condition.Op {
			case model.ArchivePushAuthorPushConditionOpEquals:
				if _condition.Value && author.AuthorizationStatus != api.AuthorAuthorizationStatus_AUTHORIZED {
					return false
				} else if !_condition.Value {
					falseCount++
					if author.AuthorizationStatus == api.AuthorAuthorizationStatus_AUTHORIZED {
						valid = false
					}
				}
				break
			case model.ArchivePushAuthorPushConditionOpNotEquals:
				if _condition.Value && author.AuthorizationStatus == api.AuthorAuthorizationStatus_AUTHORIZED {
					return false
				} else if !_condition.Value && author.AuthorizationStatus != api.AuthorAuthorizationStatus_AUTHORIZED {
					return false
				}
				break
			}
			break
		case model.ArchivePushAuthorPushConditionTypeBinded:
			switch _condition.Op {
			case model.ArchivePushAuthorPushConditionOpEquals:
				if _condition.Value && author.BindStatus != api.AuthorBindStatus_BINDED {
					return false
				} else if !_condition.Value {
					falseCount++
					if author.BindStatus == api.AuthorBindStatus_BINDED {
						valid = false
					}
				}
				break
			case model.ArchivePushAuthorPushConditionOpNotEquals:
				if _condition.Value && author.BindStatus == api.AuthorBindStatus_BINDED {
					return false
				} else if !_condition.Value && author.BindStatus != api.AuthorBindStatus_BINDED {
					return false
				}
				break
			}
			break
		case model.ArchivePushAuthorPushConditionTypeVerified:
			switch _condition.Op {
			case model.ArchivePushAuthorPushConditionOpEquals:
				if _condition.Value && author.VerificationStatus != api.AuthorVerificationStatus_VERIFIED {
					return false
				} else if !_condition.Value {
					falseCount++
					if author.VerificationStatus == api.AuthorVerificationStatus_VERIFIED {
						valid = false
					}
				}
				break
			case model.ArchivePushAuthorPushConditionOpNotEquals:
				if _condition.Value && author.VerificationStatus == api.AuthorVerificationStatus_VERIFIED {
					return false
				} else if !_condition.Value && author.VerificationStatus != api.AuthorVerificationStatus_VERIFIED {
					return false
				}
				break
			}
			break
		}
	}
	if falseCount == len(conditions) {
		valid = true
	}

	return
}

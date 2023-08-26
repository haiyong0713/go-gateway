package dao

import (
	xecode "go-common/library/ecode"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"time"
)

// GetAllAuthorPushes 分页获取推送批次信息
func (d *Dao) GetAllAuthorPushes() (list []*model.ArchivePushAuthorPush, err error) {
	list = make([]*model.ArchivePushAuthorPush, 0)
	err = d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Where("is_deprecated = ?", model.NotDeprecated).Order("id DESC").Find(&list).Error

	return
}

// GetAuthorPushesByPage 分页获取推送批次信息
func (d *Dao) GetAuthorPushesByPage(ids []int64, cuser string, pn int, ps int) (list []*model.ArchivePushAuthorPush, total int64, err error) {
	list = make([]*model.ArchivePushAuthorPush, 0)
	query := d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Where("is_deprecated = ?", model.NotDeprecated)
	if len(ids) > 0 {
		query = query.Where("id IN (?)", ids)
	}
	if cuser != "" {
		query = query.Where("cuser = ?", cuser)
	}
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err = query.Order("id DESC").Limit(ps).Offset((pn - 1) * ps).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return
}

// GetAuthorPushesByVendorIDs 获取作者推送信息
func (d *Dao) GetAuthorPushesByVendorIDs(vendorIDs []int64) (list []*model.ArchivePushAuthorPush, err error) {
	if len(vendorIDs) == 0 {
		return
	}

	list = make([]*model.ArchivePushAuthorPush, 0)
	err = d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Where("vendor_id IN (?) AND is_deprecated = ?", vendorIDs, model.NotDeprecated).Order("mtime DESC").Find(&list).Error

	return
}

// GetActiveAuthorPushesByVendorIDs 获取作者推送信息
func (d *Dao) GetActiveAuthorPushesByVendorIDs(vendorIDs []int64) (list []*model.ArchivePushAuthorPush, err error) {
	if len(vendorIDs) == 0 {
		return
	}

	list = make([]*model.ArchivePushAuthorPush, 0)
	err = d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Where("vendor_id IN (?) AND is_deprecated = ? AND status = ?", vendorIDs, model.NotDeprecated, api.AuthorPushStatus_EFFECTIVE).Order("mtime DESC").Find(&list).Error

	return
}

// GetBatchAuthorPushRelsByPushIDs 根据推送IDs获取对应的推送关系数据
func (d *Dao) GetBatchAuthorPushRelsByPushIDs(pushIDs []int64) (res []*model.ArchivePushBatchAuthorPushRel, err error) {
	if len(pushIDs) == 0 {
		return
	}
	res = make([]*model.ArchivePushBatchAuthorPushRel, 0)

	err = d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Where("author_push_id IN (?) AND is_deprecated = ?", pushIDs, model.NotDeprecated).Find(&res).Error

	return
}

// GetBatchAuthorPushRelsByPushIDsAndAuthorIDs 根据推送IDs和作者IDs获取对应的推送关系数据
func (d *Dao) GetBatchAuthorPushRelsByPushIDsAndAuthorIDs(pushIDs []int64, authorIDs []int64) (res []*model.ArchivePushBatchAuthorPushRel, err error) {
	if len(pushIDs) == 0 && len(authorIDs) == 0 {
		return
	}
	res = make([]*model.ArchivePushBatchAuthorPushRel, 0)

	query := d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Where("is_deprecated = ?", model.NotDeprecated)
	if len(pushIDs) > 0 {
		query = query.Where("author_push_id IN (?)", pushIDs)
	}
	if len(authorIDs) > 0 {
		query = query.Where("author_id IN (?)", authorIDs)
	}
	err = query.Find(&res).Error

	return
}

// GetBatchAuthorPushRelsByPushIDs 根据推送IDs获取对应的推送关系数据
func (d *Dao) GetBatchAuthorPushRelsWithAuthorAndBatchByPushIDs(pushIDs []int64) (res []*model.ArchivePushBatchAuthorPushRelWithAuthorAndBatch, err error) {
	if len(pushIDs) == 0 {
		return
	}
	res = make([]*model.ArchivePushBatchAuthorPushRelWithAuthorAndBatch, 0)

	tx := d.ORM.Db.Debug()
	err = tx.Model(model.ArchivePushBatchAuthorPushRel{}).
		Where("archive_push_batch_author_push_rels.author_push_id IN (?) AND archive_push_batch_author_push_rels.is_deprecated = ?", pushIDs, model.NotDeprecated).
		Joins("LEFT JOIN archive_push_author ON archive_push_batch_author_push_rels.author_id=archive_push_author.id AND archive_push_author.is_deprecated = ?", model.NotDeprecated).
		Joins("LEFT JOIN archive_push_batch ON archive_push_batch_author_push_rels.batch_id=archive_push_batch.id AND archive_push_batch.is_deprecated = ?", model.NotDeprecated).
		Select("archive_push_batch_author_push_rels.*, archive_push_author.nickname as \"authorNickname\", archive_push_batch.push_type as \"batchPushType\", archive_push_author.vendor_id as \"vendorId\"").
		Find(&res).Error

	return
}

// AddAuthorPush 添加稿件作者推送信息
func (d *Dao) AddAuthorPush(toAddPush model.ArchivePushAuthorPush) (addedPush *model.ArchivePushAuthorPush, err error) {
	if toAddPush.VendorID == 0 {
		return
	}
	addedPush = &toAddPush
	addedPush.CTime = xtime.Time(time.Now().Unix())
	err = d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Create(addedPush).Error
	return
}

// UpdateAuthorPushByID 更新稿件作者推送信息
func (d *Dao) UpdateAuthorPushByID(toUpdatePush *model.ArchivePushAuthorPush) (err error) {
	if toUpdatePush == nil || toUpdatePush.ID == 0 {
		return
	}
	toUpdatePush.MTime = xtime.Time(time.Now().Unix())
	err = d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Where("id = ?", toUpdatePush.ID).UpdateColumn("tags", toUpdatePush.Tags).UpdateColumn("delay_minutes", toUpdatePush.DelayMinutes).Updates(toUpdatePush).Error
	return
}

func (d *Dao) GetActiveAuthorPushesByAuthorIDs(authorIDs []int64) (res []*model.ArchivePushAuthorPushWithAuthors, err error) {
	res = make([]*model.ArchivePushAuthorPushWithAuthors, 0)

	var (
		foundRels              = make([]*model.ArchivePushBatchAuthorPushRel, 0)
		authorsMap             = make(map[int64]*model.ArchivePushAuthor)
		authors                = make([]*model.ArchivePushAuthor, 0)
		authorPushIDs          = make([]int64, 0)
		authorPushesMap        = make(map[int64]*model.ArchivePushAuthorPush)
		authorPushes           = make([]*model.ArchivePushAuthorPush, 0)
		authorPushesAuthorsMap = make(map[int64][]*model.ArchivePushAuthor)
	)
	if err = d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Where("is_deprecated = ? AND author_id IN (?)", model.NotDeprecated, authorIDs).Find(&foundRels).Error; err != nil {
		return
	} else if len(foundRels) == 0 {
		return
	}
	for _, rel := range foundRels {
		_rel := rel
		authorPushIDs = append(authorPushIDs, _rel.AuthorPushID)
	}
	// 作者推送
	if err = d.ORM.Db.Model(model.ArchivePushAuthorPush{}).Where("is_deprecated = ? AND id in (?) AND status != ?", model.NotDeprecated, authorPushIDs, api.AuthorPushStatus_CANCELED).Find(&authorPushes).Error; err != nil {
		return
	} else if len(authorPushes) == 0 {
		return
	}
	for _, authorPush := range authorPushes {
		_authorPush := authorPush
		authorPushesMap[_authorPush.ID] = _authorPush
		authorPushesAuthorsMap[_authorPush.ID] = make([]*model.ArchivePushAuthor, 0)
	}
	// 作者
	if err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("is_deprecated = ? AND id in (?)", model.NotDeprecated, authorIDs).Find(&authors).Error; err != nil {
		return
	} else if len(authors) == 0 {
		return
	}
	for _, author := range authors {
		_author := author
		authorsMap[_author.ID] = _author
	}
	// 组成res
	for _, rel := range foundRels {
		_rel := rel
		if _, exists := authorPushesAuthorsMap[_rel.AuthorPushID]; exists {
			if author, authorExists := authorsMap[_rel.AuthorID]; authorExists {
				_author := author
				authorPushesAuthorsMap[_rel.AuthorPushID] = append(authorPushesAuthorsMap[_rel.AuthorPushID], _author)
			}
		}
	}
	for authorPushID := range authorPushesAuthorsMap {
		if authorPush, exists := authorPushesMap[authorPushID]; exists {
			_authorPush := authorPush
			authors := authorPushesAuthorsMap[authorPushID]
			resPush := &model.ArchivePushAuthorPushWithAuthors{
				ArchivePushAuthorPush: *_authorPush,
				Authors:               authors,
			}

			res = append(res, resPush)
		}
	}

	return
}

// AddBatchAuthorPushRel 添加稿件作者推送批次信息
func (d *Dao) AddBatchAuthorPushRel(rel *model.ArchivePushBatchAuthorPushRel) (addedRel *model.ArchivePushBatchAuthorPushRel, err error) {
	if rel.AuthorPushID == 0 || rel.AuthorID == 0 {
		err = xecode.RequestErr
		return
	}
	addedRel = rel
	addedRel.CTime = xtime.Time(time.Now().Unix())
	err = d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Create(addedRel).Error

	return
}

func (d *Dao) DeleteBatchAuthorPushRelsByAuthorIDs(authorIDs []int64, username string) (err error) {
	if len(authorIDs) == 0 {
		return
	}
	updateMap := map[string]interface{}{
		"is_deprecated": model.Deprecated,
		"muser":         username,
		"mtime":         xtime.Time(time.Now().Unix()),
	}
	err = d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Where("author_id IN (?) AND is_deprecated = ?", authorIDs, model.NotDeprecated).UpdateColumns(updateMap).Error

	return
}

// DeleteBatchAuthorPushRelsByIDs 根据IDs删除batch作者推送关系数据
func (d *Dao) DeleteBatchAuthorPushRelsByIDs(ids []int64, username string) (err error) {
	if len(ids) == 0 {
		return
	}
	updateMap := map[string]interface{}{
		"is_deprecated": model.Deprecated,
		"muser":         username,
		"mtime":         xtime.Time(time.Now().Unix()),
	}
	err = d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Where("id IN (?) AND is_deprecated = ?", ids, model.NotDeprecated).UpdateColumns(updateMap).Error

	return
}

func (d *Dao) RestoreBatchAuthorPushRelsByAuthorIDs(authorIDs []int64, username string) (err error) {
	if len(authorIDs) == 0 {
		return
	}
	updateMap := map[string]interface{}{
		"is_deprecated": model.NotDeprecated,
		"muser":         username,
		"mtime":         xtime.Time(time.Now().Unix()),
	}
	err = d.ORM.Db.Model(model.ArchivePushBatchAuthorPushRel{}).Where("author_id IN (?)", authorIDs).UpdateColumns(updateMap).Error

	return
}

package dao

import (
	"context"
	"sync"
	"time"

	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/ecode"
)

// GetAuthorsByPage 分页获取作者信息
func (d *Dao) GetAuthorsByPage(ids []int64, mid int64, authorizationStatus int32, bindStatus int32, verificationStatus int32, vendorID int64, authorizationSID int64, cuser string, pn int, ps int) (list []*model.ArchivePushAuthor, total int64, err error) {
	list = make([]*model.ArchivePushAuthor, 0)
	query := d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("is_deprecated = ?", model.NotDeprecated)
	if len(ids) > 0 {
		query = query.Where("id IN (?)", ids)
	}
	if mid != 0 {
		query = query.Where("mid = ?", mid)
	}
	if authorizationStatus != 0 {
		query = query.Where("authorization_status = ?", authorizationStatus)
	}
	if bindStatus != 0 {
		query = query.Where("bind_status = ?", bindStatus)
	}
	if verificationStatus != 0 {
		query = query.Where("verification_status = ?", verificationStatus)
	}
	if vendorID != 0 {
		query = query.Where("vendor_id = ?", vendorID)
	}
	if authorizationSID != 0 {
		query = query.Where("authorization_sid = ?", authorizationSID)
	}
	if cuser != "" {
		query = query.Where("cuser = ?", cuser)
	}
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err = query.Order("mtime DESC").Limit(ps).Offset((pn - 1) * ps).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return
}

// GetAuthorsByUser 分页获取作者信息
func (d *Dao) GetAuthorsByUser(vendorId int64, mid int64, openId string) (list []*model.ArchivePushAuthor, err error) {
	list = make([]*model.ArchivePushAuthor, 0)
	query := d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("is_deprecated = ?", model.NotDeprecated)
	if mid > 0 {
		query = query.Where("mid = ?", mid)
	}
	if openId != "" {
		query = query.Where("open_id = ?", openId)
	}
	if vendorId > 0 {
		query = query.Where("vendor_id = ?", vendorId)
	}
	if err = query.Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}

	return
}

// AddAuthor 添加稿件作者信息
func (d *Dao) AddAuthor(toAddAuthor *model.ArchivePushAuthor) (err error) {
	if toAddAuthor == nil || toAddAuthor.MID == 0 || toAddAuthor.PushVendorID == 0 {
		return
	}
	assignMap := map[string]interface{}{
		"is_deprecated":     model.NotDeprecated,
		"authorization_sid": toAddAuthor.AuthorizationSID,
		"mtime":             xtime.Time(time.Now().Unix()),
		"muser":             toAddAuthor.CUser,
	}
	err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("mid = ? AND vendor_id = ?", toAddAuthor.MID, toAddAuthor.PushVendorID).Assign(assignMap).FirstOrCreate(toAddAuthor).Error
	return
}

// UpdateAuthorByID 更新稿件作者信息
func (d *Dao) UpdateAuthorByID(toUpdateAuthor *model.ArchivePushAuthor) (err error) {
	if toUpdateAuthor == nil || toUpdateAuthor.ID == 0 {
		return
	}
	err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("id = ?", toUpdateAuthor.ID).Updates(toUpdateAuthor).Error
	return
}

// DeleteAuthorByID 删除稿件作者信息
func (d *Dao) DeleteAuthorByID(id int64, username string) (err error) {
	if id == 0 {
		return
	}
	updateMap := map[string]interface{}{
		"is_deprecated": model.Deprecated,
		"muser":         username,
		"mtime":         xtime.Time(time.Now().Unix()),
	}
	err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("id = ?", id).Updates(updateMap).Error
	return
}

// PutAuthorsForWhiteList 将作者放入白名单中
func (d *Dao) PutAuthorsForWhiteList(vendorID int64, authors []*model.ArchivePushAuthorWithBVIDs) (err error) {
	if vendorID == 0 || len(authors) == 0 {
		return
	}
	eg := errgroup.WithContext(context.Background())
	for _, author := range authors {
		_author := author
		eg.Go(func(ctx context.Context) error {
			args := redis.Args{}
			if redisKey, _err := model.GetAuthorWhiteListKeyByAuthor(vendorID, _author.ArchivePushAuthor.MID); _err != nil {
				return _err
			} else {
				args = args.Add(redisKey)
				args = args.Add("true")
			}
			args = args.AddFlat(_author.BVIDs)
			if _, err = d.redis.Do(context.Background(), "SADD", args...); err != nil {
				log.Error("Dao: PutAuthorsForWhiteList Error (%v)", err)
			}
			return nil
		})
	}
	err = eg.Wait()

	return
}

// RemoveAuthorsFromWhiteList 从白名单中移除作者
func (d *Dao) RemoveAuthorsFromWhiteList(vendorID int64, mids []int64) (err error) {
	if vendorID == 0 || len(mids) == 0 {
		return
	}

	eg := errgroup.WithContext(context.Background())
	for _, mid := range mids {
		_mid := mid
		eg.Go(func(ctx context.Context) error {
			args := redis.Args{}
			if redisKey, _err := model.GetAuthorWhiteListKeyByAuthor(vendorID, _mid); _err != nil {
				return _err
			} else {
				args = args.Add(redisKey)
			}
			if _, err = d.redis.Do(context.Background(), "DEL", args...); err != nil {
				log.Error("Dao: PutAuthorsForWhiteList Error (%v)", err)
			}
			return nil
		})
	}
	err = eg.Wait()
	return
}

// GetAuthorsWhiteList 根据MIDs查询作者白名单
func (d *Dao) GetAuthorsWhiteList(vendorID int64, mids []int64) (res map[int64][]string, err error) {
	if vendorID == 0 || len(mids) == 0 {
		return
	}
	res = make(map[int64][]string, 0)
	var (
		rawAuthors []*model.ArchivePushAuthor
		lock       sync.Mutex
	)
	eg := errgroup.WithContext(context.Background())
	if rawAuthors, err = d.GetAuthorsByUser(vendorID, 0, ""); err != nil {
		return
	} else if len(rawAuthors) == 0 {
		return
	}
	for _, rawAuthor := range rawAuthors {
		_rawAuthor := rawAuthor
		mid := _rawAuthor.MID
		eg.Go(func(ctx context.Context) error {
			lock.Lock()
			res[mid] = make([]string, 0)
			lock.Unlock()
			args := redis.Args{}
			if redisKey, _err := model.GetAuthorWhiteListKeyByAuthor(vendorID, mid); _err != nil {
				return _err
			} else {
				args = args.Add(redisKey)
			}
			if bvids, _err := redis.Strings(d.redis.Do(context.Background(), "SMEMBERS", args...)); _err != nil {
				log.Error("Dao: GetBVIDsWhiteList Error (%v)", _err)
				return _err
			} else if len(bvids) == 0 {
				return nil
			} else {
				lock.Lock()
				for _, bvid := range bvids {
					if bvid != "true" {
						res[mid] = append(res[mid], bvid)
					}
				}
				lock.Unlock()
			}
			return nil
		})

	}
	err = eg.Wait()
	return
}

// GetAuthorizationSIDByVendorAndMID 获取用户对应的活动的SID
func (d *Dao) GetAuthorizationSIDByVendorAndMID(vendorID int64, mid int64) (sid int64, err error) {
	if vendorID == 0 || mid == 0 {
		return
	}
	foundRow := &model.ArchivePushAuthor{}
	if err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("mid = ? AND vendor_id = ? AND is_deprecated = ?", mid, vendorID, model.NotDeprecated).Find(foundRow).Error; err != nil {
		return
	} else if foundRow.ID == 0 {
		err = ecode.AuthorNotFound
		return
	}
	sid = foundRow.AuthorizationSID

	return
}

// GetAuthorsByIDs 根据作者IDs获取作者
func (d *Dao) GetAuthorsByIDs(ids []int64) (res []*model.ArchivePushAuthor, err error) {
	if len(ids) == 0 {
		return
	}

	res = make([]*model.ArchivePushAuthor, 0)
	err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("is_deprecated = ? AND id IN (?)", model.NotDeprecated, ids).Order("id").Find(&res).Error

	return
}

// GetAuthorByMID 根据作者MID获取作者
func (d *Dao) GetAuthorByMID(vendorID int64, mid int64) (res *model.ArchivePushAuthor, err error) {
	if vendorID == 0 || mid == 0 {
		return nil, xecode.RequestErr
	}

	res = &model.ArchivePushAuthor{}
	err = d.ORM.Db.Model(model.ArchivePushAuthor{}).Where("is_deprecated = ? AND vendor_id = ? AND mid = ?", model.NotDeprecated, vendorID, mid).Order("id").Find(&res).Error

	return
}

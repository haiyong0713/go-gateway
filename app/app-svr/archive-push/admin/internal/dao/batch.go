package dao

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-main/app/ep/hassan/mock/support/slice"
	"time"

	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
)

var (
	batchDetailOrderableColumns = []string{"mtime", "id"}
)

func (d *Dao) GetAllBatches() (list []*model.ArchivePushBatch, err error) {
	list = make([]*model.ArchivePushBatch, 0)
	if err = d.ORM.Db.Model(model.ArchivePushBatch{}).Assign(model.ArchivePushBatch{PushVendorID: 1}).Find(&list).Error; err != nil {
		return
	}
	batchIDs := make([]int64, 0)
	for batchI, batch := range list {
		list[batchI].PushStatus = api.ArchivePushBatchPushStatus_SUCCESS
		batchIDs = append(batchIDs, batch.ID)
	}
	batchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).Where("batch_id IN (?)", batchIDs).Find(&batchDetails).Error; err != nil {
		return
	}
	successCountMap := make(map[int64]int)
	failCountMap := make(map[int64]int)
	for _, detail := range batchDetails {
		switch detail.PushStatus {
		case api.ArchivePushDetailPushStatus_SUCCESS:
			successCountMap[detail.BatchID]++
			break
		default:
			failCountMap[detail.BatchID]++
		}
	}
	for batchI := range list {
		batchID := list[batchI].ID
		if failCountMap[batchID] > 0 {
			list[batchI].PushStatus = api.ArchivePushBatchPushStatus_FAIL_PARTIAL
			if successCountMap[batchID] == 0 {
				list[batchI].PushStatus = api.ArchivePushBatchPushStatus_FAIL
			}
		}
	}

	return
}

// GetBatchesByIDs 根据batch ids获取batches
func (d *Dao) GetBatchesByIDs(ids []int64) (list []*model.ArchivePushBatch, err error) {
	if len(ids) == 0 {
		return
	}

	list = make([]*model.ArchivePushBatch, 0)
	if err = d.ORM.Db.Model(model.ArchivePushBatch{}).Where("id IN (?) AND is_deprecated = ?", ids, model.NotDeprecated).Find(&list).Error; err != nil {
		return
	}
	batchIDs := make([]int64, 0)
	for batchI, batch := range list {
		list[batchI].PushStatus = api.ArchivePushBatchPushStatus_SUCCESS
		batchIDs = append(batchIDs, batch.ID)
	}
	batchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).Where("batch_id IN (?)", batchIDs).Find(&batchDetails).Error; err != nil {
		return
	}
	successCountMap := make(map[int64]int)
	failCountMap := make(map[int64]int)
	for _, detail := range batchDetails {
		switch detail.PushStatus {
		case api.ArchivePushDetailPushStatus_SUCCESS:
			successCountMap[detail.BatchID]++
			break
		default:
			failCountMap[detail.BatchID]++
		}
	}
	for batchI := range list {
		batchID := list[batchI].ID
		if failCountMap[batchID] > 0 {
			list[batchI].PushStatus = api.ArchivePushBatchPushStatus_FAIL_PARTIAL
			if successCountMap[batchID] == 0 {
				list[batchI].PushStatus = api.ArchivePushBatchPushStatus_FAIL
			}
		}
	}

	return
}

// GetBatchesByPage 分页获取推送批次信息
func (d *Dao) GetBatchesByPage(ids []int64, pushVendorIDs []int64, cuser string, pn int, ps int) (list []*model.ArchivePushBatch, total int64, err error) {
	list = make([]*model.ArchivePushBatch, 0)
	query := d.ORM.Db.Model(model.ArchivePushBatch{})
	if len(ids) > 0 {
		query = query.Where("id IN (?)", ids)
	}
	if len(pushVendorIDs) > 0 {
		query = query.Where("push_vendor_id IN (?)", pushVendorIDs)
	}
	if cuser != "" {
		query = query.Where("cuser = ?", cuser)
	}
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err = query.Order("ctime DESC, mtime DESC").Limit(ps).Offset((pn - 1) * ps).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	batchIDs := make([]int64, 0)
	for batchI, batch := range list {
		list[batchI].PushStatus = api.ArchivePushBatchPushStatus_SUCCESS
		batchIDs = append(batchIDs, batch.ID)
	}
	batchDetails := make([]*model.ArchivePushBatchDetail, 0)
	if err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).Where("batch_id IN (?)", batchIDs).Find(&batchDetails).Error; err != nil {
		return
	}
	successCountMap := make(map[int64]int)
	failCountMap := make(map[int64]int)
	pushingCountMap := make(map[int64]int)
	for _, detail := range batchDetails {
		switch detail.PushStatus {
		case api.ArchivePushDetailPushStatus_SUCCESS:
			successCountMap[detail.BatchID]++
			break
		case api.ArchivePushDetailPushStatus_UNKNOWN:
			pushingCountMap[detail.BatchID]++
		default:
			// 手动推送下架的默认属于成功状态
			if detail.ArchiveStatus == api.ArchiveStatus_WITHDRAW {
				successCountMap[detail.BatchID]++
			} else {
				failCountMap[detail.BatchID]++
			}
		}
	}
	for batchI := range list {
		batchID := list[batchI].ID
		if list[batchI].CTime.Time().After(time.Now()) {
			list[batchI].PushStatus = api.ArchivePushBatchPushStatus_TO_PUSH
		} else if pushingCountMap[batchID] > 0 {
			list[batchI].PushStatus = api.ArchivePushBatchPushStatus_PUSHING
		} else if failCountMap[batchID] > 0 {
			list[batchI].PushStatus = api.ArchivePushBatchPushStatus_FAIL_PARTIAL
			if successCountMap[batchID] == 0 {
				list[batchI].PushStatus = api.ArchivePushBatchPushStatus_FAIL
			}
		}
	}

	return
}

func (d *Dao) GetBatchByID(id int64) (out model.ArchivePushBatch, err error) {
	out = model.ArchivePushBatch{}
	err = d.ORM.Db.Model(model.ArchivePushBatch{}).Where("id = ?", id).Find(&out).Error

	return
}

func (d *Dao) GetBatchDetailsByBatchID(batchID int64) (list []*model.ArchivePushBatchDetail, err error) {
	list = make([]*model.ArchivePushBatchDetail, 0)
	err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).Where("batch_id = ?", batchID).Find(&list).Error

	return
}

func (d *Dao) GetBatchDetailsByPushStatuses(vendorID int64, statuses []api.ArchivePushDetailPushStatus_Enum, pushType api.ArchivePushType_Enum) (res []*model.ArchivePushBatchDetailWithVendor, err error) {
	res = make([]*model.ArchivePushBatchDetailWithVendor, 0)
	joinStr := "JOIN archive_push_batch ON archive_push_batch_detail.batch_id=archive_push_batch.id AND archive_push_batch.is_deprecated = ?"
	joinParams := []interface{}{model.NotDeprecated}
	if vendorID != 0 {
		joinStr += " AND archive_push_batch.push_vendor_id = ? "
		joinParams = append(joinParams, vendorID)
	}
	if pushType != 0 {
		joinStr += " AND archive_push_batch.push_type = ? "
		joinParams = append(joinParams, pushType)
	}
	err = d.ORM.Db.Debug().Model(model.ArchivePushBatchDetail{}).
		Joins(joinStr, joinParams...).
		Where("archive_push_batch_detail.is_deprecated = ? AND archive_push_batch_detail.push_status IN (?)", model.NotDeprecated, statuses).Order("archive_push_batch_detail.mtime DESC").
		Select("archive_push_batch_detail.*, archive_push_batch.push_vendor_id, archive_push_batch.push_type").
		Order("archive_push_batch.push_vendor_id, archive_push_batch_detail.mtime").
		Find(&res).Error

	return
}

func (d *Dao) GetBatchDetailsByBVID(vendorID int64, bvid string) (res []*model.ArchivePushBatchDetail, err error) {
	res = make([]*model.ArchivePushBatchDetail, 0)
	var aid int64
	if aid, err = util.BvToAv(bvid); err != nil {
		err = ecode.AVBVIDConvertingError
		return
	}
	err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).
		Joins("JOIN archive_push_batch ON archive_push_batch.id=archive_push_batch_detail.batch_id AND archive_push_batch.push_vendor_id = ? AND archive_push_batch.is_deprecated = ?", vendorID, model.NotDeprecated).
		Where("archive_push_batch_detail.aid = ? AND archive_push_batch_detail.is_deprecated = ?", aid, model.NotDeprecated).
		Order("archive_push_batch_detail.mtime DESC").Find(&res).Error

	return
}

func (d *Dao) GetBatchDetailsByBVIDs(vendorIDs []int64, bvids []string, order string, desc bool) (res []*model.ArchivePushBatchDetail, err error) {
	res = make([]*model.ArchivePushBatchDetail, 0)
	aids := make([]int64, 0)
	for _, bvid := range bvids {
		if aid, _err := util.BvToAv(bvid); _err != nil {
			log.Error("archive-push-admin.dao.GetBatchDetailsByBVID.BvToAv(%s) Error (%v)", bvid, err)
			return
		} else {
			aids = append(aids, aid)
		}
	}
	if len(aids) == 0 {
		return
	}
	query := d.ORM.Db.Model(model.ArchivePushBatchDetail{}).
		Joins("JOIN archive_push_batch ON archive_push_batch.id=archive_push_batch_detail.batch_id AND archive_push_batch.push_vendor_id IN (?)", vendorIDs).
		Where("archive_push_batch_detail.aid IN (?)", aids).
		Select("archive_push_batch_detail.*")
	orderStr := "mtime DESC"
	if slice.Contains(batchDetailOrderableColumns, order) {
		orderStr = order
		if desc {
			orderStr += " DESC"
		} else {
			orderStr += " ASC"
		}
	}
	err = query.Order(orderStr).Find(&res).Error

	return
}

// GetBatchDetailsByAuthorPushes 根据作者获取所有对应稿件推送详情
func (d *Dao) GetBatchDetailsByAuthor(vendorID int64, mid int64) (list []*model.ArchivePushBatchDetail, err error) {
	if mid == 0 {
		return
	}
	list = make([]*model.ArchivePushBatchDetail, 0)
	query := d.ORM.Db.Model(model.ArchivePushAuthorPush{}).
		Joins("JOIN archive_push_batch_author_push_rels ON archive_push_author_push.id=archive_push_batch_author_push_rels.author_push_id AND archive_push_batch_author_push_rels.is_deprecated = ?", model.NotDeprecated).
		Joins("JOIN archive_push_author ON archive_push_batch_author_push_rels.author_id=archive_push_author.id AND archive_push_author.is_deprecated = ? AND archive_push_author.mid = ? AND archive_push_author.vendor_id = ?", model.NotDeprecated, mid, vendorID).
		Joins("JOIN archive_push_batch ON archive_push_batch_author_push_rels.batch_id=archive_push_batch.id AND archive_push_batch.is_deprecated = ? AND archive_push_batch.push_vendor_id = ? AND archive_push_batch.push_type = ?", model.NotDeprecated, vendorID, model.PushTypeAuthor).
		Joins("LEFT JOIN archive_push_batch_detail ON archive_push_batch_detail.batch_id=archive_push_batch.id AND archive_push_batch_detail.is_deprecated = ?", model.NotDeprecated).
		Where("archive_push_author_push.is_deprecated = ?", model.NotDeprecated).
		Select("archive_push_batch_detail.*")
	err = query.Order("archive_push_author_push.id ASC").Find(&list).Error

	return
}

func (d *Dao) CreateBatch(batch *model.ArchivePushBatch) (res *model.ArchivePushBatch, err error) {
	if err = d.ORM.Db.Model(model.ArchivePushBatch{}).Create(batch).Error; err != nil {
		return
	}
	res = batch

	return
}

// UpdateBatchVendorIDsByBatchIDs
func (d *Dao) UpdateBatchVendorIDsByBatchIDs(batchIDs []int64, vendorID int64) (err error) {
	if len(batchIDs) == 0 {
		return
	}
	if vendorID == 0 {
		return
	}
	err = d.ORM.Db.Model(model.ArchivePushBatch{}).Where("id in (?)", batchIDs).UpdateColumn("push_vendor_id", vendorID).Error

	return nil
}

func (d *Dao) CreateBatchDetail(detail model.ArchivePushBatchDetail) (res *model.ArchivePushBatchDetail, err error) {
	if err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).Create(&detail).Error; err != nil {
		return
	}
	res = &detail

	return
}

func (d *Dao) UpdateBatchDetail(detail model.ArchivePushBatchDetail) (res *model.ArchivePushBatchDetail, err error) {
	if detail.ID == 0 {
		err = xecode.NothingFound
		return
	}

	if err = d.ORM.Db.Model(model.ArchivePushBatchDetail{}).Where("id=?", detail.ID).Updates(detail).Error; err != nil {
		return
	}
	res = &detail

	return
}

// PutBatchWithBVIDsForTodo 将batch放入待推送池
func (d *Dao) PutBatchWithBVIDsForTodo(batchID int64, bvids []string, toPushTime xtime.Time) (err error) {
	if batchID == 0 || len(bvids) == 0 {
		return xecode.RequestErr
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if _, _err := conn.Do("SADD", model.RedisBatchToPushKey, batchID); _err != nil {
		log.Error("Dao: PutBatchWithBVIDsForTodo Error (%v)", _err)
	}
	if _, _err := conn.Do("SET", fmt.Sprintf(model.RedisBatchToPushTimeKey, batchID), toPushTime.Time().Format(model.DefaultTimeLayout)); _err != nil {
		log.Error("Dao: PutBatchWithBVIDsForTodo Error (%v)", _err)
	}
	args := redis.Args{}
	args = args.Add(fmt.Sprintf(model.RedisBatchToPushBVIDsKey, batchID))
	args = args.AddFlat(bvids)
	if _, _err := conn.Do("SADD", args...); _err != nil {
		log.Error("Dao: PutBatchWithBVIDsForTodo Error (%v)", _err)
	}
	return
}

// RemoveBatchFromTodo 将batch移出待推送池
func (d *Dao) RemoveBatchFromTodo(batchID int64) (err error) {
	if batchID == 0 {
		return xecode.RequestErr
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if _, _err := conn.Do("SREM", model.RedisBatchToPushKey, batchID); _err != nil {
		log.Error("Dao: RemoveBatchFromTodo Error (%v) 移除推送id错误", _err)
	}
	if _, _err := conn.Do("DEL", fmt.Sprintf(model.RedisBatchToPushTimeKey, batchID)); _err != nil {
		log.Error("Dao: RemoveBatchFromTodo Error (%v) 移除推送时间错误", _err)
	}
	if _, _err := conn.Do("DEL", fmt.Sprintf(model.RedisBatchToPushBVIDsKey, batchID)); _err != nil {
		log.Error("Dao: RemoveBatchFromTodo Error (%v) 移除推送BVIDs错误", _err)
	}
	return
}

// RemoveBatchesFromTodo 将batches移出待推送池
func (d *Dao) RemoveBatchesFromTodo(batchIDs []int64) (err error) {
	if len(batchIDs) == 0 {
		return
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	args := redis.Args{}
	args = args.Add(model.RedisBatchToPushKey)
	args = args.AddFlat(batchIDs)
	if _, _err := conn.Do("SREM", args...); _err != nil {
		log.Error("Dao: RemoveBatchFromTodo Error (%v) 移除推送id错误", _err)
	}
	for _, batchID := range batchIDs {
		if _, _err := conn.Do("DEL", fmt.Sprintf(model.RedisBatchToPushTimeKey, batchID)); _err != nil {
			log.Error("Dao: RemoveBatchFromTodo Error (%v) 移除推送时间错误", _err)
		}
		if _, _err := conn.Do("DEL", fmt.Sprintf(model.RedisBatchToPushBVIDsKey, batchID)); _err != nil {
			log.Error("Dao: RemoveBatchFromTodo Error (%v) 移除推送BVIDs错误", _err)
		}
	}
	return
}

// GetBatchesIDsFromTodo 获取所有待推送的batch IDs
func (d *Dao) GetBatchesIDsFromTodo() (batchIDs []int64, err error) {
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	batchIDs, err = redis.Int64s(conn.Do("SMEMBERS", model.RedisBatchToPushKey))

	return
}

// GetBatchesPushTimeFromTodo 获取待推送的batch的推送时间
func (d *Dao) GetBatchesPushTimeFromTodo(batchIDs []int64) (pushTimesMap map[int64]time.Time, err error) {
	if len(batchIDs) == 0 {
		return nil, xecode.RequestErr
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	pushTimesMap = make(map[int64]time.Time)
	for _, batchID := range batchIDs {
		if pushTimeStr, _err := redis.String(conn.Do("GET", fmt.Sprintf(model.RedisBatchToPushTimeKey, batchID))); _err != nil {
			err = _err
			return
		} else if pushTimeStr != "" {
			if pushTimesMap[batchID], _err = time.ParseInLocation(model.DefaultTimeLayout, pushTimeStr, time.Local); _err != nil {
				log.Error("Dao: GetBatchesPushTimeFromTodo %d time.Parse(%s) error %v", batchID, pushTimeStr, _err)
			}
		}
	}

	return
}

// GetBatchesBVIDsFromTodo 获取待推送的batch的BVIDs
func (d *Dao) GetBatchesBVIDsFromTodo(batchIDs []int64) (bvidsMap map[int64][]string, err error) {
	if len(batchIDs) == 0 {
		return nil, xecode.RequestErr
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	bvidsMap = make(map[int64][]string)
	for _, batchID := range batchIDs {
		if bvids, _err := redis.Strings(conn.Do("SMEMBERS", fmt.Sprintf(model.RedisBatchToPushBVIDsKey, batchID))); _err != nil {
			log.Error("Dao: GetBatchesBVIDsFromTodo %d SMEMBERS error %v", batchID, _err)
		} else {
			bvidsMap[batchID] = bvids
		}
	}

	return
}

// UpdateBatchesPushTimeForTodo 更新待推送的batch的推送时间
func (d *Dao) UpdateBatchesPushTimeForTodo(batchIDs []int64, toPushTime xtime.Time) (err error) {
	if len(batchIDs) == 0 {
		return
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	for _, batchID := range batchIDs {
		if _, _err := conn.Do("SET", fmt.Sprintf(model.RedisBatchToPushTimeKey, batchID), toPushTime.Time().Format(model.DefaultTimeLayout)); _err != nil {
			if _err != redis.ErrNil {
				err = errors.Wrap(_err, 1)
				return
			}
		}
	}

	return
}

// LockBatchTodo 锁定稿件检查推送操作
func (d *Dao) LockBatchTodo() (err error) {
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if locked, _err := redis.Bool(conn.Do("GET", model.RedisBatchToPushLockKey)); _err != nil {
		if _err == redis.ErrNil {
			_err = nil
		} else {
			err = _err
			return
		}
	} else if locked {
		err = ecode.BatchTodoAlreadyLocked
		return
	}
	_, err = conn.Do("SET", model.RedisBatchToPushLockKey, true)
	return
}

// CheckIfLockBatchTodo 检查稿件检查推送操作锁
func (d *Dao) CheckIfLockBatchTodo() (locked bool, err error) {
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if locked, err = redis.Bool(conn.Do("GET", model.RedisBatchToPushLockKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		}
	}
	return
}

// UnlockBatchTodo 解锁稿件检查推送操作
func (d *Dao) UnlockBatchTodo() (err error) {
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	_, err = conn.Do("DEL", model.RedisBatchToPushLockKey)
	return
}

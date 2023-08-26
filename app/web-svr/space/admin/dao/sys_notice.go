package dao

import (
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/space/admin/model"

	"github.com/jinzhu/gorm"
)

// SysNotice system notice
func (d *Dao) SysNotice(param *model.SysNoticeList) (total int, list []*model.SysNotice, err error) {
	query := d.DB.Model(&model.SysNotice{})
	if param.Uid != 0 {
		w := map[string]interface{}{
			"uid":        param.Uid,
			"is_deleted": model.NotDeleted,
		}
		value := &model.SysNoticeUid{}
		if err = d.DB.Model(&model.SysNoticeUid{}).Where(w).First(value).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				err = nil
				return
			}
			log.Error("dao.SysNotice error(%v)", err)
			return
		}
		w = map[string]interface{}{
			"id": value.SystemNoticeId,
		}
		query = query.Where(w)
	}
	if len(param.Scopes) > 0 {
		cond := ""
		for _, scope := range param.Scopes {
			if cond != "" {
				cond = cond + " OR "
			}
			cond = fmt.Sprintf("%s scopes LIKE '%%%d%%'", cond, scope)
		}
		query = query.Where(cond)
	}
	if param.Status != 0 {
		query = query.Where("status = ?", param.Status)
	}
	query = query.Count(&total).Order("id DESC")
	if param.Ps != 0 && param.Pn != 0 {
		query = query.Offset((param.Pn - 1) * param.Ps).Limit(param.Ps)
	}
	if err = query.Find(&list).Error; err != nil {
		log.Error("dao.SysNotice error(%v)", err)
		return
	}
	return
}

// SysNoticeInfo get system notice info by id
func (d *Dao) SysNoticeInfo(id int64) (info *model.SysNoticeInfo, err error) {
	var uidList []*model.SysNoticeUid
	info = &model.SysNoticeInfo{}

	if err = d.DB.Model(&model.SysNotice{}).Where("id = ?", id).Scan(info).Error; err != nil {
		return info, err
	}
	if err = d.DB.Model(&model.SysNoticeUid{}).
		Where("is_deleted = ? AND system_notice_id = ?", model.NotDeleted, id).
		Scan(&uidList).Error; err != nil {
		return info, err
	}

	info.Uids = make([]int64, len(uidList))
	for idx, uid := range uidList {
		info.Uids[idx] = uid.Uid
	}
	return
}

// SysNoticeAdd add system notice
func (d *Dao) SysNoticeAdd(param *model.SysNoticeAdd) (err error) {
	if err = d.DB.Model(&model.SysNoticeAdd{}).Create(param).Error; err != nil {
		log.Error("dao.SysNoticeAdd error(%v)", err)
		return
	}
	return
}

// SysNoticeUp add system update
func (d *Dao) SysNoticeUpdate(param *model.SysNoticeUp) (err error) {
	var (
		count int64
	)
	where := map[string]interface{}{
		"system_notice_id": param.ID,
		"is_deleted":       model.NotDeleted,
	}
	if err = d.DB.Model(&model.SysNoticeUid{}).Where(where).Count(&count).Error; err != nil {
		log.Error("dao.SysNoticeUpdate Count error(%v)", err)
		return
	}
	param.UidCount = count
	if err = d.DB.Model(&model.SysNoticeUp{}).Save(param).Error; err != nil {
		log.Error("dao.SysNoticeUp error(%v)", err)
		return
	}
	return
}

// SysNoticeOpt option system update
func (d *Dao) SysNoticeOpt(param *model.SysNoticeOpt) (err error) {
	up := map[string]interface{}{
		"status": param.Status,
	}
	if err = d.DB.Model(&model.SysNoticeOpt{}).Where("id = ?", param.ID).Update(up).Error; err != nil {
		log.Error("dao.SysNoticeOpt error(%v)", err)
		return
	}
	return
}

// SysNoticeUidAdd  add system notice uid
func (d *Dao) SysNoticeUidAdd(param *model.SysNotUidAddDel) (err error) {
	sql, sqlParam := model.BatchAddUIDSQL(param.ID, param.UIDs)
	if err = d.DB.Model(&model.SysNoticeUp{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("SysNoticeUidAdd Exec error(%v)", err)
		return
	}
	where := map[string]interface{}{
		"system_notice_id": param.ID,
		"is_deleted":       model.NotDeleted,
	}
	var count int
	if err = d.DB.Model(&model.SysNoticeUid{}).Where(where).Count(&count).Error; err != nil {
		log.Error("SysNoticeUidAdd Count error(%v)", err)
		return
	}
	up := map[string]interface{}{
		"uid_count": count,
	}
	if err = d.DB.Model(&model.SysNoticeUp{}).Where("id = ?", param.ID).Update(up).Error; err != nil {
		log.Error("SysNoticeUidAdd Update error(%v)", err)
		return
	}
	return
}

// SysNoticeUidFirst find one system notice uid
func (d *Dao) SysNoticeUidFirst(uid int64, noticeId int64, scopes []int) (err error) {
	if len(scopes) == 0 {
		log.Error("empty PR scopes")
		return ecode.Errorf(-1, "无PR公告场景")
	}
	for _, scope := range scopes {
		value := &model.SysNoticeUidWithScope{}
		db := d.DB.Model(&model.SysNoticeUid{}).
			Joins("LEFT JOIN system_notice AS n ON n.id = system_notice_uid.system_notice_id").
			Where("system_notice_uid.uid = ? AND system_notice_uid.is_deleted = ?", uid, model.NotDeleted).
			Where(fmt.Sprintf("n.scopes LIKE '%%%d%%'", scope))
		if noticeId > 0 {
			db = db.Where("n.id != ?", noticeId)
		}
		err = db.Select("n.id as system_notice_id, n.scopes as scopes, system_notice_uid.id as id, system_notice_uid.uid as uid").
			First(value).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				err = nil
				continue
			}
			log.Error("dao.SysNoticeUidFirst error(%v)", err)
			return
		}
		if value.ID != 0 {
			return ecode.Errorf(-1, "UID:%d, 在%s已有公告%d", uid, model.ScopeDict[scope], value.SystemNoticeId)
		}
	}
	return
}

// SysNoticeUidList returns notice related uid list in []int64
func (d *Dao) SysNoticeUidList(noticeId int64) (uids []int64, err error) {
	var obj []*model.SysNoticeUid
	if err = d.DB.Model(&model.SysNoticeUid{}).
		Where("system_notice_id = ? AND is_deleted = ?", noticeId, model.NotDeleted).
		Find(&obj).Error; err != nil {
		log.Error("dao.SysNoticeUidList noticeId(%v) error(%v)", noticeId, err)
		return
	}

	uids = make([]int64, len(obj))
	for idx, info := range obj {
		uids[idx] = info.Uid
	}
	return
}

// SysNoticeUidFind find all system notice uid
func (d *Dao) SysNoticeUidFind(param *model.SysNoticeUidParam) (value []*model.SysNoticeUid, err error) {
	w := map[string]interface{}{
		"system_notice_id": param.ID,
		"is_deleted":       model.NotDeleted,
	}
	if err = d.DB.Model(&model.SysNoticeUid{}).Where(w).Find(&value).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return value, nil
		}
		log.Error("dao.SysNoticeUidFind error(%v)", err)
		return
	}
	return
}

// SysNoticeUidDel  del system notice uid
func (d *Dao) SysNoticeUidDel(param *model.SysNotUidAddDel) (err error) {
	upUID := map[string]interface{}{
		"is_deleted": model.Deleted,
	}
	w := map[string]interface{}{
		"system_notice_id": param.ID,
	}
	if err = d.DB.Model(&model.SysNoticeUid{}).Where(w).Where("uid IN (?)", param.UIDs).Update(upUID).Error; err != nil {
		log.Error("SysNoticeUidDel uid Update error(%v)", err)
		return
	}
	where := map[string]interface{}{
		"system_notice_id": param.ID,
		"is_deleted":       model.NotDeleted,
	}
	var count int
	if err = d.DB.Model(&model.SysNoticeUid{}).Where(where).Count(&count).Error; err != nil {
		log.Error("SysNoticeUidDel count error(%v)", err)
		return
	}
	upCount := map[string]interface{}{
		"uid_count": count,
	}
	if err = d.DB.Model(&model.SysNoticeUp{}).Where("id = ?", param.ID).Update(upCount).Error; err != nil {
		log.Error("SysNoticeUidDel update count error(%v)", err)
		return
	}
	return
}

func (d *Dao) SysNoticeByUids(uids []int64, scopes []int, status int64) (ret map[int64][]*model.SysNotice, err error) {
	ret = make(map[int64][]*model.SysNotice)
	for _, uid := range uids {
		_, list, err1 := d.SysNotice(&model.SysNoticeList{
			Uid:    uid,
			Scopes: scopes,
			Status: status,
		})
		if err1 != nil {
			return ret, err1
		}
		if len(list) == 0 {
			continue
		}
		ret[uid] = list
	}
	return
}

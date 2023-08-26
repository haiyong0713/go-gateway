package show

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	_oneTimeRedDot          = 1
	_stateForbidden         = 1
	_allList                = 3
	_entranceOperateSQL     = "UPDATE popular_top_entrance SET state=? WHERE id=?"
	_updateRedDotSQL        = "UPDATE popular_top_entrance SET version=version+1,update_time=now() WHERE module_id=? AND state!=2 AND red_dot = 2"
	_updateHRedDotSQL       = "UPDATE popular_top_entrance SET version=?,update_time=now() WHERE id=? AND state!=2 AND red_dot = 2"
	_updateRedDotDisposable = "UPDATE popular_top_entrance SET version = version+1, update_time = now(), red_dot_text = ? WHERE id = ? AND state != 2 AND red_dot = 1"
	_entranceTopPhotoSQL    = "UPDATE popular_top_entrance SET top_photo=? WHERE id=?"
	_maxVersionSQL          = "SELECT MAX(version) AS ID FROM popular_top_entrance WHERE module_id='hot-channel'"
	_hotChannel             = "hot-channel"
)

type MaxVersion struct {
	ID int `gorm:"column:ID"`
}

// PopEntranceAdd .
func (d *Dao) PopEntranceAdd(ctx context.Context, param *show.EntranceSave) (id int64, err error) {
	maxVersion := MaxVersion{}
	if param.ModuleID == _hotChannel {
		if err = d.DB.Raw(_maxVersionSQL).Scan(&maxVersion).Error; err != nil {
			return
		}
	}
	if param.RedDot == _oneTimeRedDot { // 一次性红点有初始值
		param.Version = maxVersion.ID + 1
		log.Info("POPMaxVersion add have +1, id(%d) title(%s) moduleID(%s)", param.ID, param.Title, param.ModuleID)
	}
	param.State = _stateForbidden // 默认禁止
	if err = d.DB.Create(param).Error; err != nil {
		return
	}
	id = param.ID
	return
}

// PopEntranceEdit .
func (d *Dao) PopEntranceEdit(ctx context.Context, param *show.EntranceSave) (err error) {
	maxVersion := MaxVersion{}
	m := param.ToEntranceMap()
	isOneTimeRedDot := false
	oldVersion := 0
	if param.ModuleID == _hotChannel {
		if err = d.DB.Raw(_maxVersionSQL).Scan(&maxVersion).Error; err != nil {
			return err
		}
		var item show.EntranceSave
		db := d.DB.Model(&show.EntranceSave{})
		if err = db.Where("id=?", param.ID).First(&item).Error; err != nil {
			err = errors.Wrapf(err, "[GetEntrance] id %d", param.ID)
			return
		}
		if item.RedDot == _oneTimeRedDot {
			isOneTimeRedDot = true
			oldVersion = item.Version
		}
	}
	if param.RedDot == _oneTimeRedDot { // 一次性红点有初始值
		if isOneTimeRedDot {
			m["version"] = oldVersion
		} else {
			m["version"] = maxVersion.ID + 1
		}
		log.Info("POPMaxVersion edit have +1, id(%d) title(%s) moduleID(%s)", param.ID, param.Title, param.ModuleID)
	}
	if err = d.DB.Model(&show.EntranceSave{}).Where("id=?", param.ID).Where("state!=2").Update(m).Error; err != nil {
		err = errors.Wrapf(err, "id(%d)", param.ID)
		return
	}
	return
}

// PopEntrance list .
func (d *Dao) PopEntrance(ctx context.Context, state, pn, ps int) (res *show.EntranceListRes, err error) {
	var items []*show.EntranceList
	res = &show.EntranceListRes{}
	db := d.DB.Model(&show.EntranceSave{})
	if state != _allList {
		db = db.Where("state=?", state)
	} else {
		db = db.Where("state!=2")
	}
	db.Count(&res.Pager.Total)
	if err = db.Offset((pn - 1) * ps).Limit(ps).Order("rank ASC").Find(&items).Error; err != nil {
		err = errors.Wrapf(err, "[PopEntrance] state %d, pn %d, ps %d", state, pn, ps)
		return
	}
	res.Pager.Num = pn
	res.Pager.Size = ps
	res.Items = items
	return
}

// PopEntranceOperate .
func (d *Dao) PopEntranceOperate(ctx context.Context, id int64, state int) (err error) {
	if err = d.DB.Exec(_entranceOperateSQL, state, id).Error; err != nil {
		err = errors.Wrapf(err, "id(%d)", id)
	}
	return
}

// RedDotUpdate 更新红点 .
func (d *Dao) RedDotUpdate(ctx context.Context, moduleID string, id int64) (rows int64, err error) {
	var res *gorm.DB
	if moduleID == _hotChannel {
		if id == 0 {
			err = errors.Wrapf(ecode.RequestErr, "hot-channel id should > 0")
			return
		}
		maxVersion := MaxVersion{}
		if err = d.DB.Raw(_maxVersionSQL).Scan(&maxVersion).Error; err != nil {
			return
		}
		if res = d.DB.Exec(_updateHRedDotSQL, maxVersion.ID+1, id); res.Error != nil {
			err = errors.Wrapf(err, "id %d, version %d", id, maxVersion.ID+1)
			return
		}
	} else {
		if res = d.DB.Exec(_updateRedDotSQL, moduleID); res.Error != nil {
			err = errors.Wrapf(err, "moduleID %s", moduleID)
			return
		}
	}
	rows = res.RowsAffected
	return
}

// RedDotUpdate 更新一次性红点
func (d *Dao) RedDotUpdateDisposable(ctx context.Context, moduleID string, id int64, content string) (rows int64, err error) {
	var res *gorm.DB

	if res = d.DB.Exec(_updateRedDotDisposable, content, id); res.Error != nil {
		err = errors.Wrapf(err, "id %d, content %s", id, content)
		return
	}
	rows = res.RowsAffected
	return
}

// TopPhotoUpdate 更新头图 .
func (d *Dao) EntranceTopPhotoUpdate(ctx context.Context, id int64, topPhoto string) (err error) {
	if err = d.DB.Exec(_entranceTopPhotoSQL, topPhoto, id).Error; err != nil {
		err = errors.Wrapf(err, "id(%d)", id)
	}
	return
}

// GetTopPhoto 获取头图 .
func (d *Dao) EntranceGetTopPhoto(ctx context.Context, id int64) (topPhoto string, err error) {
	var item []*show.EntranceList
	db := d.DB.Model(&show.EntranceSave{})
	if err = db.Where("id=?", id).First(&item).Error; err != nil {
		err = errors.Wrapf(err, "[GetTopPhoto] id %d", id)
		return
	}
	if len(item) == 1 {
		return item[0].TopPhoto, nil
	}
	return "", err
}

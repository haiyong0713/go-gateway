package show

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/hidden"

	"github.com/jinzhu/gorm"
)

const (
	_addHiddenLimitSQL = "INSERT INTO `entrance_hidden_limit`(`oid`,`build`,`conditions`,`plat`) VALUES %s"
)

// HiddenSave .
func (d *Dao) HiddenSave(c context.Context, hd *hidden.Hidden, limits map[int8]*hidden.HiddenLimit) (id int64, err error) {
	if hd == nil {
		return
	}
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("HiddenSave %v", r)
		}
		if err != nil {
			if err1 := tx.Rollback().Error; err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
		id = hd.ID
	}()
	if hd.ID > 0 {
		upParam := map[string]interface{}{
			"sid":              hd.SID,
			"rid":              hd.RID,
			"cid":              hd.CID,
			"channel":          hd.Channel,
			"pid":              hd.PID,
			"stime":            hd.Stime,
			"etime":            hd.Etime,
			"hidden_condition": hd.HiddenCondition,
			"module_id":        hd.ModuleID,
			"hide_dynamic":     hd.HideDynamic,
		}
		if err = tx.Model(&hidden.Hidden{}).Where("id=?", hd.ID).Update(upParam).Error; err != nil {
			return
		}
		//删除limit的信息
		if err = tx.Model(&hidden.HiddenLimit{}).Where("oid=? AND `state`=?", hd.ID, 1).Update(map[string]int{"state": 0}).Error; err != nil {
			return
		}
	} else {
		if err = tx.Model(&hidden.Hidden{}).Create(hd).Error; err != nil {
			return
		}
	}
	var (
		rowStrings   []string
		paramStrings []interface{}
	)
	for _, v := range limits {
		rowStrings = append(rowStrings, "(?,?,?,?)")
		paramStrings = append(paramStrings, hd.ID, v.Build, v.Conditions, v.Plat)
	}
	err = tx.Exec(fmt.Sprintf(_addHiddenLimitSQL, strings.Join(rowStrings, ",")), paramStrings...).Error
	return
}

// HiddenLimits .
func (d *Dao) HiddenLimits(c context.Context, oIDs []int64) (hiddenLimits map[int64][]*hidden.HiddenLimit, err error) {
	tmpLimits := make([]*hidden.HiddenLimit, 0)
	hiddenLimits = make(map[int64][]*hidden.HiddenLimit)
	if err = d.DB.Model(&hidden.HiddenLimit{}).Where("oid in (?) AND `state`=?", oIDs, 1).Find(&tmpLimits).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("HiddenLimits Find oid(%+v) error(%v)", oIDs, err)
		}
		return
	}
	if len(tmpLimits) > 0 {
		for _, v := range tmpLimits {
			hiddenLimits[v.OID] = append(hiddenLimits[v.OID], v)
		}
	}
	return
}

// UpdateHiddenState .
func (d *Dao) UpdateHiddenState(c context.Context, id int64, whState, upState int) (rows int64, err error) {
	query := d.DB.Model(&hidden.Hidden{}).Where("id=? AND `state`=?", id, whState).Update(map[string]int{"state": upState})
	rows = query.RowsAffected
	err = query.Error
	return
}

// Hidden .
func (d *Dao) Hidden(c context.Context, id int64) (res *hidden.Hidden, err error) {
	res = new(hidden.Hidden)
	if err = d.DB.Model(&hidden.Hidden{}).Where("id=?", id).First(&res).Error; err != nil {
		log.Error("Hidden First id(%d) error(%v)", id, err)
	}
	return
}

// Hiddens .
func (d *Dao) Hiddens(c context.Context, pn, ps int) (res []*hidden.Hidden, total int64, err error) {
	query := d.DB.Model(&hidden.Hidden{}).Where("state != ?", hidden.StateDel)
	if err = query.Count(&total).Error; err != nil {
		log.Error("Hiddens count error(%v)", err)
		return
	}
	res = make([]*hidden.Hidden, 0)
	if err = query.Order("state DESC,id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&res).Error; err != nil {
		log.Error("Hiddens Find error(%v)", err)
	}
	return
}

// Region .
func (d *Dao) Region(c context.Context, oid int64) (region *hidden.Region, err error) {
	if oid <= 0 {
		return
	}
	region = new(hidden.Region)
	if err = d.DB.Model(&hidden.Region{}).Where("rid=?", oid).First(&region).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("Hidden Region Find error(%v)", err)
		}
		return
	}
	return
}

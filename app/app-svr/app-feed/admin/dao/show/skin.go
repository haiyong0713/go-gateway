package show

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/menu"

	"github.com/jinzhu/gorm"
)

const (
	_modifySkinExtSQL = "UPDATE `skin_ext` SET `skin_name` = ?, `attribute` = ?, `skin_id` = ?, `stime` = ?, `etime` = ?, `location_policy_gid` = ?,  `operator` = ?, `user_scope_type`=?, `user_scope_value`=? , `dress_up_type`=?, `dress_up_value`=? WHERE `id` = ?"
	_onOffSkinSQL     = "UPDATE `skin_ext` SET `state` = ? WHERE `id` = ? AND `state` = ?"
	_skinLimitAddSQL  = "INSERT INTO `skin_limit`(`s_id`,`plat`,`build`,`conditions`,`ctime`) VALUES %s"
	_delSkinLimitSQL  = "UPDATE `skin_limit` SET `state` = ? WHERE `s_id` = ? AND `state` = ?"
)

// SkinExtSave .
func (d *Dao) SkinExtSave(c context.Context, save *menu.SkinExt, limits map[int8]*menu.SkinBuildLimit) (id int64, err error) {
	if save == nil || len(limits) == 0 {
		return
	}
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("SkinExtSave %v", r)
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
		id = save.ID
	}()
	ctime := xtime.Time(time.Now().Unix())
	if save.ID > 0 {
		var argSql []interface{}
		argSql = append(argSql, save.SkinName, save.Attribute, save.SkinID, save.Stime, save.Etime, save.LocationPolicyGroupID, save.Operator, save.UserScopeType, save.UserScopeValue, save.DressUpType, save.DressUpValue, save.ID)
		if err = tx.Exec(_modifySkinExtSQL, argSql...).Error; err != nil {
			return
		}
		//删除limit的信息
		if err = tx.Exec(_delSkinLimitSQL, 0, save.ID, 1).Error; err != nil {
			return
		}
	} else {
		save.Ctime = ctime
		if err = tx.Create(save).Error; err != nil {
			return
		}
	}
	var (
		rowStrings   []string
		paramStrings []interface{}
	)
	for _, v := range limits {
		rowStrings = append(rowStrings, "(?,?,?,?,?)")
		paramStrings = append(paramStrings, save.ID, v.Plat, v.Build, v.Conditions, ctime)
	}
	err = tx.Exec(fmt.Sprintf(_skinLimitAddSQL, strings.Join(rowStrings, ",")), paramStrings...).Error
	return
}

// SkinLimits .
func (d *Dao) SkinLimits(c context.Context, sIDs []int64) (skinLimits map[int64][]*menu.SkinLimit, err error) {
	tempLimits := make([]*menu.SkinLimit, 0)
	if err = d.DB.Model(&menu.SkinLimit{}).Where("s_id in (?) AND `state` = ?", sIDs, 1).Find(&tempLimits).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error(" SkinLimits Find error(%v)", err)
		}
		return
	}
	if len(tempLimits) > 0 {
		skinLimits = make(map[int64][]*menu.SkinLimit)
		for _, v := range tempLimits {
			skinLimits[v.SID] = append(skinLimits[v.SID], v)
		}
	}
	return
}

// SkinModifyState .
func (d *Dao) SkinModifyState(c context.Context, id int64, whState, upState int) error {
	return d.DB.Exec(_onOffSkinSQL, upState, id, whState).Error
}

// RawSkinExts .
func (d *Dao) RawSkinExts(c context.Context) (tempExts []*menu.SkinExt, err error) {
	tempExts = make([]*menu.SkinExt, 0)
	if err = d.DB.Model(&menu.SkinExt{}).Where("`state` = ?", 1).Find(&tempExts).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("RawSkinExts menu.RawSkinExts{} Find error(%v)", err)
		}
	}
	return
}

// SkinExts .
func (d *Dao) SkinExts(c context.Context, id int64, pn, ps int) (rely *menu.SkinSearchReply, err error) {
	rely = &menu.SkinSearchReply{}
	query := d.DB.Model(&menu.SkinExt{})
	if id != 0 {
		query = query.Where("id = ?", id)
	}
	// 删除的不展示
	query = query.Where("state != ?", -1)
	if err = query.Count(&rely.Total).Error; err != nil {
		log.Error("SkinExts count error(%v)", err)
		return
	}
	rely.List = make([]*menu.SkinExt, 0)
	if err = query.Order("`id` DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rely.List).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("SkinExts Find error(%v)", err)
		}
	}
	return
}

// RawSkinExt .
func (d *Dao) RawSkinExt(c context.Context, id int64) (rly *menu.SkinExt, err error) {
	rly = &menu.SkinExt{}
	if err = d.DB.Model(&menu.SkinExt{}).Where("id = ?", id).First(rly).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

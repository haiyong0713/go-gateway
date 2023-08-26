package show

import (
	"context"
	"fmt"
	"strings"
	"time"

	xtime "go-common/library/time"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/menu"

	"github.com/jinzhu/gorm"
)

const (
	_modifyTabExtSQL = "UPDATE `tab_ext` SET `type` = ?,`tab_id` = ? ,`attribute` = ?,`inactive_icon` = ?,`inactive` = ?,`inactive_type` = ?,`active_icon` = ?,`active` = ?,`active_type` = ?,`font_color` = ?,`bar_color` = ?,`stime` = ?,`etime` = ?,`operator` = ?,`ver`=?,`tab_top_color`=?,`tab_middle_color`=?,`tab_bottom_color`=?,`bg_image1`=?,`bg_image2`=? WHERE `id` = ?"
	_limitAddSQL     = "INSERT INTO `tab_limit`(`t_id`,`type`,`plat`,`build`,`conditions`,`ctime`) VALUES %s"
)

// TabExtSave .
func (d *Dao) TabExtSave(c context.Context, save *menu.TabExt, limits map[int8]*menu.BuildLimit) (id int64, err error) {
	if save == nil || len(limits) == 0 {
		return
	}
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("TabExtSave %v", r)
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
		argSql = append(argSql, save.Type, save.TabID, save.Attribute, save.InactiveIcon, save.Inactive, save.InactiveType, save.ActiveIcon,
			save.Active, save.ActiveType, save.FontColor, save.BarColor, save.Stime, save.Etime, save.Operator, save.Ver,

			save.TabTopColor, save.TabMiddleColor, save.TabBottomColor, save.BgImage1, save.BgImage2, save.ID)
		if err = tx.Exec(_modifyTabExtSQL, argSql...).Error; err != nil {
			return
		}
		//删除limit的信息
		if err = tx.Model(&menu.TabLimit{}).Where("`t_id` = ? AND `type` = ? AND `state` = ?", save.ID, 0, 1).Update(map[string]int{"state": 0}).Error; err != nil {
			return
		}
	} else {
		save.Ctime = ctime
		if err = tx.Model(&menu.TabExt{}).Create(save).Error; err != nil {
			return
		}
	}
	var (
		rowStrings   []string
		paramStrings []interface{}
	)
	for _, v := range limits {
		rowStrings = append(rowStrings, "(?,?,?,?,?,?)")
		paramStrings = append(paramStrings, save.ID, 0, v.Plat, v.Build, v.Conditions, ctime)
	}
	err = tx.Exec(fmt.Sprintf(_limitAddSQL, strings.Join(rowStrings, ",")), paramStrings...).Error
	return
}

// TabLimits .
func (d *Dao) TabLimits(c context.Context, tIDs []int64, tType int) (tabLimits map[int64][]*menu.TabLimit, err error) {
	tempLimits := make([]*menu.TabLimit, 0)
	if err = d.DB.Model(&menu.TabLimit{}).Where("t_id in (?) AND `type` = ? AND `state` = ?", tIDs, tType, 1).Find(&tempLimits).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("MenuTabList menu.TabLimit{} Find error(%v)", err)
		}
		return
	}
	if len(tempLimits) > 0 {
		tabLimits = make(map[int64][]*menu.TabLimit)
		for _, v := range tempLimits {
			tabLimits[v.TID] = append(tabLimits[v.TID], v)
		}
	}
	return
}

// ModifyState .
func (d *Dao) ModifyState(c context.Context, id int64, whState, upState int) error {
	return d.DB.Model(&menu.TabExt{}).Where("id =? AND `state` = ?", id, whState).Update(map[string]int{"state": upState}).Error
}

// RawTabExts .
func (d *Dao) RawTabExts(c context.Context, tabID int64, tType int) (tempExts []*menu.TabExt, err error) {
	tempExts = make([]*menu.TabExt, 0)
	if err = d.DB.Model(&menu.TabExt{}).Where("tab_id = ? AND `type` = ? AND `state` = ?", tabID, tType, 1).Find(&tempExts).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("RawTabExts menu.TabExt{} Find error(%v)", err)
		}
	}
	return
}

// TabExts .
func (d *Dao) TabExts(c context.Context, id int64, pn, ps int) (rely *menu.SearchReply, err error) {
	rely = &menu.SearchReply{}
	query := d.DB.Model(&menu.TabExt{})
	if id != 0 {
		query = query.Where("id = ?", id)
	}
	// 删除的不展示
	query = query.Where("state != ?", menu.TabDel)
	if err = query.Count(&rely.Total).Error; err != nil {
		log.Error("TabExts count error(%v)", err)
		return
	}
	rely.List = make([]*menu.TabExt, 0)
	if err = query.Order("`id` DESC").Offset((pn - 1) * ps).Limit(ps).Find(&rely.List).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("TabExts Find error(%v)", err)
		}
	}
	return
}

// RawTabExt .
func (d *Dao) RawTabExt(c context.Context, id int64) (rly *menu.TabExt, err error) {
	rly = &menu.TabExt{}
	if err = d.DB.Model(&menu.TabExt{}).Where("id = ?", id).First(rly).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

// AppMenus .
func (d *Dao) AppMenus(c context.Context, ids []int64) (tabLimits map[int64]*menu.AppMenus, err error) {
	if len(ids) == 0 {
		return
	}
	tempLimits := make([]*menu.AppMenus, 0)
	if err = d.DB.Model(&menu.AppMenus{}).Where("id in (?)", ids).Find(&tempLimits).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("AppMenus menu.TabLimit{} Find error(%v)", err)
		}
		return
	}
	if len(tempLimits) > 0 {
		tabLimits = make(map[int64]*menu.AppMenus)
		for _, v := range tempLimits {
			tabLimits[v.ID] = v
		}
	}
	return
}

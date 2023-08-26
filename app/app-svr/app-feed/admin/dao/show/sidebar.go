package show

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/menu"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// SideBars .
func (d *Dao) SideBars(c context.Context, ids []int64) (res map[int64]*menu.Sidebar, err error) {
	if len(ids) == 0 {
		return
	}
	ss := make([]*menu.Sidebar, 0)
	if err = d.DB.Model(&menu.Sidebar{}).Where("id in (?)", ids).Find(&ss).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("SideBars Find error(%+v) ids(%+v)", err, ids)
		}
		return
	}
	if len(ss) > 0 {
		res = make(map[int64]*menu.Sidebar)
		for _, v := range ss {
			res[v.ID] = v
		}
	}
	return
}

// SideBarLimits .
func (d *Dao) SideBarLimits(c context.Context, sids []int64) (res map[int64][]*show.SidebarLimit, err error) {
	if len(sids) == 0 {
		return
	}
	sl := make([]*show.SidebarLimit, 0)
	if err = d.DB.Model(&show.SidebarLimit{}).Where("s_id in (?)", sids).Find(&sl).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("SideBars Find error(%+v) ids(%+v)", err, sids)
		}
		return
	}
	if len(sl) > 0 {
		res = make(map[int64][]*show.SidebarLimit)
		for _, v := range sl {
			res[v.SID] = append(res[v.SID], v)
		}
	}
	return
}

// SideBarsByPlat .
func (d *Dao) SideBarsByPlat(c context.Context, plat int32) (res []*show.SidebarWithLimit, err error) {
	ss := make([]*show.SidebarLimit, 0)
	if err = d.DB.Table("sidebar").
		Select("sidebar.id,sidebar.plat,sl.s_id,sidebar.name,sidebar.plat,sl.conditions,sl.build,sm.name as sm_name").
		Joins("LEFT JOIN sidebar_limit as sl ON sidebar.id = sl.s_id").
		Joins("LEFT JOIN sidebar_module as sm ON sidebar.module = sm.id").
		Where("sidebar.plat=? AND sidebar.state=1 AND sidebar.online_time<? AND (sm.mtype in (0,1) OR (sm.mtype=2 AND sm.style = 4))", plat, time.Now()).
		Order("id DESC").Find(&ss).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("SideBars Find error(%+v) plat(%d)", err, plat)
		}
		return
	}
	sidm := make(map[int64]int64)
	limits := make(map[int64][]*show.SidebarLimit)
	for _, v := range ss {
		limit := &show.SidebarLimit{Conditions: v.Conditions, Build: v.Build}
		limits[v.SID] = append(limits[v.SID], limit)
	}
	for _, v := range ss {
		if _, ok := sidm[v.SID]; ok || v.SID <= 0 {
			continue
		}
		sidm[v.SID] = v.SID
		res = append(res, &show.SidebarWithLimit{SID: v.SID, Name: v.Name, Plat: v.Plat, Limit: limits[v.SID], ModuleName: v.ModuleName})
	}
	return
}

// ModulesByPlat .id=8,9,10为首页tab和icon，兼容老数据处理
func (d *Dao) ModulesByPlat(c context.Context, plat int32) (res []*show.SidebarModuleSimple, err error) {
	res = make([]*show.SidebarModuleSimple, 0)
	err = d.DB.Table("sidebar_module").Select("id,name,mtype,style,op_load_condition,op_style_type,area_policy,show_purposed").
		Where("(plat=? OR (id in (8, 9, 10) AND mtype=2)) AND state=1", plat).Order("mtype DESC,rank DESC").
		Find(&res).Error
	return
}

// Module .
func (d *Dao) Module(c context.Context, id int64) (res *show.SidebarModule, err error) {
	res = new(show.SidebarModule)
	err = d.DB.Where("id=?", id).First(&res).Error
	return
}

// ModuleSave .
func (d *Dao) ModuleSave(_ context.Context, sm *show.SidebarModule) (int64, error) {
	if sm == nil {
		return 0, nil
	}
	if sm.Style == show.ModuleStyle_OP || sm.Style == show.ModuleStyle_Launcher || sm.Style == show.ModuleStyle_Classification || sm.Style == show.ModuleStyle_Recommend_tab {
		var count int
		query := d.DB.Model(&show.SidebarModule{}).
			Where("plat=?", sm.Plat).
			Where("state=?", 1).
			Where("style=?", sm.Style).
			Where("op_style_type=?", sm.OpStyleType).
			Count(&count)
		if err := query.Error; err != nil {
			err = errors.Wrapf(err, "db query module err param(%+v)", sm)
			return 0, err
		}
		if sm.ID == 0 && count >= 1 {
			var err error
			switch sm.Style {
			case show.ModuleStyle_OP:
				{
					err = ecode.Error(-400, "当前plat已存在创作/直播运营位")
					if sm.OpStyleType == 1 {
						err = ecode.Error(-400, "当前plat已存在投稿引导强化卡")
					}
				}
			case show.ModuleStyle_Launcher:
				{
					err = ecode.Error(-400, "当前plat已存在发布浮窗")
				}
			case show.ModuleStyle_Classification:
				{
					err = ecode.Error(-400, "当前plat已存在分区入口")
				}
			case show.ModuleStyle_Recommend_tab:
				{
					err = ecode.Error(-400, "当前plat已存在港澳台垂类tab")
				}
			}
			return 0, err
		}
	}
	if sm.ID > 0 {
		upParam := map[string]interface{}{
			"mtype":             sm.MType,
			"plat":              sm.Plat,
			"name":              sm.Name,
			"title":             sm.Title,
			"style":             sm.Style,
			"rank":              sm.Rank,
			"button_name":       sm.ButtonName,
			"button_url":        sm.ButtonURL,
			"button_icon":       sm.ButtonIcon,
			"button_style":      sm.ButtonStyle,
			"white_url":         sm.WhiteURL,
			"title_color":       sm.TitleColor,
			"subtitle":          sm.Subtitle,
			"subtitle_color":    sm.SubtitleColor,
			"subtitle_url":      sm.SubtitleURL,
			"background":        sm.Background,
			"background_color":  sm.BackgroundColor,
			"audit_show":        sm.AuditShow,
			"is_mng":            sm.IsMng,
			"op_load_condition": sm.OpLoadCondition,
			"op_style_type":     sm.OpStyleType,
			"area_policy":       sm.AreaPolicy,
			"show_purposed":     sm.ShowPurposed,
		}
		if err := d.DB.Model(&show.SidebarModule{}).Where("id=?", sm.ID).Update(upParam).Error; err != nil {
			err = errors.Wrapf(err, "db update err param(%+v)", upParam)
			return 0, err
		}
	} else {
		if err := d.DB.Model(&show.SidebarModule{}).Create(sm).Error; err != nil {
			err = errors.Wrapf(err, "db Create err param(%+v)", sm)
			return 0, err
		}
	}
	return sm.ID, nil
}

// UpdateModuleState .
func (d *Dao) UpdateModuleState(c context.Context, id int64, state int) (rows int64, err error) {
	query := d.DB.Model(&show.SidebarModule{}).Where("id=?", id).Update(map[string]int{"state": state})
	rows = query.RowsAffected
	err = query.Error
	return
}

// ModuleItemsByPlat .
func (d *Dao) ModuleItemsByPlat(_ context.Context, req *show.ModuleItemListReq) (res []*show.SidebarORM, err error) {
	res = make([]*show.SidebarORM, 0)

	err = d.DB.Model(&show.SidebarORM{}).
		Where("state != ?", show.SidebarORM_State_Deleted).
		Where("module = ?", req.Module).
		Where("plat = ?", req.Plat).
		Where("lang_id = ?", req.Lang).
		Order("state DESC").
		Order("rank DESC").
		Find(&res).Error

	return res, err
}

// ModuleItemsByPlat .
func (d *Dao) ModuleItemsDetail(_ context.Context, sidebarID int64) (res *show.SidebarORM, err error) {
	res = new(show.SidebarORM)

	err = d.DB.Model(&show.SidebarORM{}).
		Where("state != ?", show.SidebarORM_State_Deleted).
		Where("id = ?", sidebarID).
		Find(&res).Error

	return res, err
}

// UpdateModuleItemState .
func (d *Dao) UpdateModuleItemState(c context.Context, sidebarID int64, state int32) error {
	return d.DB.Model(&show.SidebarORM{}).Where("id=?", sidebarID).Update(map[string]int32{"state": state}).Error
}

// SaveModuleItem .
func (d *Dao) SaveModuleItem(_ context.Context, sidebar *show.SidebarORM) error {
	return d.DB.Model(&show.SidebarORM{}).Save(sidebar).Error
}

// SaveSidebarLimit .
func (d *Dao) SaveSidebarLimit(_ context.Context, limit *show.SidebarLimitORM) error {
	return d.DB.Model(&show.SidebarLimitORM{}).Save(limit).Error
}

// UpdateModuleItemState .
func (d *Dao) DeleteSidebarLimit(_ context.Context, sidebarID int64) error {
	return d.DB.Where("s_id = ?", sidebarID).Delete(&show.SidebarLimitORM{}).Error
}

// UpdateModuleItemState .
func (d *Dao) DisableExpiredSidebar(_ context.Context) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	return d.DB.Model(&show.SidebarORM{}).Where("state = 1 AND offline_time != '0000-00-00 00:00:00' AND offline_time < ?", now).
		Update(map[string]int32{"state": 0}).Error
}

package native

import (
	"context"
	"time"

	"go-common/library/log"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"

	"github.com/jinzhu/gorm"
)

const (
	TableTab       = "act_tab"
	TableTabModule = "act_tab_module"
)

func (d *Dao) GetTabById(c context.Context, id int32) (tab *natmdl.Tab, err error) {
	tab = &natmdl.Tab{}
	if err = d.DB.Table(TableTab).Where("id = ?", id).First(&tab).Error; err != nil {
		log.Error("[GetTabById] d.DB.First(%v), error(%v)", id, err)
	}
	return
}

func (d *Dao) SearchTab(c context.Context, req *natmdl.SearchTabReq) (list *natmdl.TabList, err error) {
	list = &natmdl.TabList{}
	db := d.DB.Table(TableTab)
	if req.ID != 0 {
		db = db.Where("id = ?", req.ID)
	}
	if req.Title != "" {
		db = db.Where("title like ?", req.Title+"%")
	}
	if req.Creator != "" {
		db = db.Where("creator = ?", req.Creator)
	}
	if req.CtimeStart != 0 {
		db = db.Where("ctime >= ?", time.Unix(req.CtimeStart, 0).Format("2006-01-02 15:04:05"))
	}
	if req.CtimeEnd != 0 {
		db = db.Where("ctime <= ?", time.Unix(req.CtimeEnd, 0).Format("2006-01-02 15:04:05"))
	}
	// 有效数据
	if req.State == 1 {
		db = db.Where("state = ?", req.State).Where("stime != 0").Where("stime <= ?", time.Now().Format("2006-01-02 15:04:05")).Where("etime >= ? OR etime = 0", time.Now().Format("2006-01-02 15:04:05"))
	} else if req.State == 0 {
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		defStr := "0000-00-00 00:00:00"
		db = db.Where("state = ? OR stime > ? OR stime = ? OR (etime != ? AND etime < ?)", 0, nowStr, defStr, defStr, nowStr)
	}
	if err = db.Count(&list.Total).Error; err != nil {
		log.Error("[SearchTab] d.DB.Count(%v), error(%v)", req, err)
	}
	db = db.Offset((req.Pn - 1) * req.Ps).Limit(req.Ps).Order("id DESC")
	if err = db.Find(&list.List).Error; err != nil {
		log.Error("[SearchTab] d.DB.Find(%v), error(%v)", req, err)
	}
	return
}

func (d *Dao) CreateTab(c context.Context, db *gorm.DB, tab *natmdl.Tab) (id int32, err error) {
	if db == nil {
		db = d.DB
	}
	if err = db.Table(TableTab).Create(tab).Error; err != nil {
		log.Error("[CreateTab] d.DB.Create(%v), error(%v)", tab, err)
		return
	}
	id = tab.ID
	return
}

func (d *Dao) UpdateTabById(c context.Context, db *gorm.DB, id int32, tabMap map[string]interface{}) (err error) {
	if db == nil {
		db = d.DB
	}
	if err = db.Table(TableTab).Where("id = ?", id).Update(tabMap).Error; err != nil {
		log.Error("[UpdateTabById] d.DB.Update(%v), error(%v)", tabMap, err)
	}
	return
}

func (d *Dao) GetTabModuleByTabIds(c context.Context, tabIds []int32) (tabModules []*natmdl.TabModule, err error) {
	db := d.DB.Table(TableTabModule).Where("tab_id in (?) and state = ?", tabIds, natmdl.TabModuleStateValid)
	if err = db.Find(&tabModules).Error; err != nil {
		log.Error("[GetTabModuleByTabIds] d.DB.Find(%v), error(%v)", tabIds, err)
	}
	return
}

func (d *Dao) GetTabModuleByPid(c context.Context, pid int32) (*natmdl.TabModule, error) {
	tabModule := new(natmdl.TabModule)
	db := d.DB.Table(TableTabModule).Where("pid = ? and state = ?", pid, natmdl.TabModuleStateValid)
	if err := db.First(&tabModule).Error; err != nil {
		log.Error("[GetTabModuleByPid] d.DB.Find(%v), error(%v)", pid, err)
		return nil, err
	}
	return tabModule, nil
}

func (d *Dao) GetTabModuleByPids(c context.Context, pids []int32, category int32) (tabModules []*natmdl.TabModule, err error) {
	db := d.DB.Table(TableTabModule).Where("pid in (?) and category = ? and  state = ?", pids, category, natmdl.TabModuleStateValid)
	if err = db.Find(&tabModules).Error; err != nil {
		log.Error("[GetTabModuleByTabIds] d.DB.Find(%v), error(%v)", pids, err)
	}
	return
}

func (d *Dao) CreateTabModule(c context.Context, db *gorm.DB, tabModule *natmdl.TabModule) (id int32, err error) {
	if db == nil {
		db = d.DB
	}
	if err = db.Table(TableTabModule).Create(tabModule).Error; err != nil {
		log.Error("[CreateTabModule] d.DB.Create(%v), error(%v)", tabModule, err)
		return
	}
	id = tabModule.ID
	return
}

func (d *Dao) UpdateTabModulesById(c context.Context, db *gorm.DB, id int32, tabModuleMap map[string]interface{}) (err error) {
	if db == nil {
		db = d.DB
	}
	if err = db.Table(TableTabModule).Where("id = ?", id).Update(tabModuleMap).Error; err != nil {
		log.Error("[UpdateTabModule] d.DB.Update(%v), error(%v)", tabModuleMap, err)
	}
	return
}

func (d *Dao) UpdateTabModulesByTabId(c context.Context, db *gorm.DB, tabId int32, tabModuleMap map[string]interface{}) (err error) {
	if db == nil {
		db = d.DB
	}
	if err = db.Table(TableTabModule).Where("tab_id = ?", tabId).Update(tabModuleMap).Error; err != nil {
		log.Error("[UpdateTabModule] d.DB.Update(%v), error(%v)", tabModuleMap, err)
	}
	return
}

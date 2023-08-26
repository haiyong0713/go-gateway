package manager

import (
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"

	"github.com/jinzhu/gorm"
)

// UserRole .
func (d *Dao) UserRole(name string, level int) (res []*manager.PosRecUserMgt, err error) {
	where := map[string]interface{}{
		"name": name,
		"type": level,
	}
	if err = d.DB.Where(where).Find(&res).Error; err != nil {
		return
	}
	return
}

// UserGroupByPids .
func (d *Dao) UserGroupByPids(pids []int) (res []*manager.PosRecUserMgt, err error) {
	where := map[string]interface{}{
		"type": common.RoleType,
	}
	query := d.DB
	if len(pids) > 0 {
		query = query.Where("id in (?)", pids)
	}
	if err = query.Where(where).Find(&res).Error; err != nil {
		return
	}
	return
}

// UserGroupByName .
func (d *Dao) UserGroupByName(name string) (*manager.PosRecUserMgt, error) {
	where := map[string]interface{}{
		"name": name,
	}
	res := &manager.PosRecUserMgt{}
	err := d.DB.Where("`type` != ?", common.RoleType).Where(where).Last(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			res = nil
		}
		return nil, err
	}
	return res, nil
}

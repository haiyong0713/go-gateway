package native

import (
	"context"

	"go-common/library/log"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
)

const (
	_tsModule = "native_ts_module"
)

func (d *Dao) GetTsModules(c context.Context, tsID int64) ([]*natmdl.NatTsModule, error) {
	var tabModules []*natmdl.NatTsModule
	db := d.DB.Table(_tsModule).Where("ts_id=? and state = ?", tsID, 1)
	if err := db.Find(&tabModules).Error; err != nil {
		log.Error("[GetTabModuleByTabIds] d.DB.Find(%v), error(%v)", tsID, err)
		return nil, err
	}
	return tabModules, nil
}

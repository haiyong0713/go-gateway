package native

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/admin/model/native"
)

const (
	_tsModResource = "native_ts_module_resource"
)

func (d *Dao) TsModResources(c context.Context, modIDs []int64) (map[int64][]*native.NatTsModuleResource, error) {
	if len(modIDs) == 0 {
		return map[int64][]*native.NatTsModuleResource{}, nil
	}
	resources := make([]*native.NatTsModuleResource, 0, len(modIDs))
	db := d.DB.Table(_tsModResource).Where("module_id in (?) and state = ?", modIDs, 1)
	if err := db.Find(&resources).Error; err != nil {
		log.Errorc(c, "Fail to get native_ts_module_resource, modIDs=%+v error=%+v", modIDs, err)
		return nil, err
	}
	list := make(map[int64][]*native.NatTsModuleResource, len(modIDs))
	for _, v := range resources {
		list[v.ModuleID] = append(list[v.ModuleID], v)
	}
	return list, nil
}

package show

import (
	"context"

	"go-gateway/app/app-svr/resource/service/api/v1"
)

const _paramSQL = "SELECT id,name,`value`,remark,plat,build,conditions,department FROM `param` WHERE `state` = 0"

// ParamList get param list
func (d *Dao) ParamList(c context.Context) ([]*v1.Param, error) {
	rows, err := d.db.Query(c, _paramSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*v1.Param
	for rows.Next() {
		p := new(v1.Param)
		if err = rows.Scan(&p.ID, &p.Name, &p.Value, &p.Remark, &p.Plat, &p.Build, &p.Conditions, &p.Department); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

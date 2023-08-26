package feature

import (
	"context"

	"go-common/library/log"

	buildLimitmdl "go-gateway/app/app-svr/feature/service/model/buildLimit"
)

const (
	_buildLimitSQL      = `SELECT key_name,config FROM build_limit WHERE state="on" AND tree_id=?`
	_buildLimitTreesSQL = `SELECT distinct(tree_id) FROM build_limit WHERE state="on"`
)

func (d *Dao) BuildLimit(c context.Context, treeID int64) (res map[string]*buildLimitmdl.BuildLimit, err error) {
	rows, err := d.db.Query(c, _buildLimitSQL, treeID)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]*buildLimitmdl.BuildLimit)
	for rows.Next() {
		var re = new(buildLimitmdl.BuildLimit)
		if err = rows.Scan(&re.KeyName, &re.Conditions); err != nil {
			log.Error("%+v", err)
			return
		}
		res[re.KeyName] = re
	}
	if err = rows.Err(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (d *Dao) BuildLimitTrees(c context.Context) (res []int64, err error) {
	rows, err := d.db.Query(c, _buildLimitTreesSQL)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var treeID int64
		if err = rows.Scan(&treeID); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, treeID)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
	}
	return
}

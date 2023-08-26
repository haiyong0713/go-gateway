package feature

import (
	"context"
	"go-common/library/log"

	businessConfigMdl "go-gateway/app/app-svr/feature/service/model/businessConfig"
)

const (
	_businessConfig         = `SELECT id,tree_id,key_name,config,description,relations FROM business_config WHERE state="on" AND tree_id=?`
	_businessConfigTreesSQL = `SELECT distinct(tree_id) FROM business_config WHERE state="on"`
)

func (d *Dao) BusinessConfig(c context.Context, treeID int64) (res map[string]*businessConfigMdl.BusinessConfig, err error) {
	rows, err := d.db.Query(c, _businessConfig, treeID)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]*businessConfigMdl.BusinessConfig)
	for rows.Next() {
		var re = new(businessConfigMdl.BusinessConfig)
		if err = rows.Scan(&re.ID, &re.TreeID, &re.KeyName, &re.Config, &re.Description, &re.Relations); err != nil {
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

func (d *Dao) BusinessConfigTrees(c context.Context) (res []int64, err error) {
	rows, err := d.db.Query(c, _businessConfigTreesSQL)
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

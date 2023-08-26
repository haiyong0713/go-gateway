package feature

import (
	"context"

	"go-common/library/log"

	abtestmdl "go-gateway/app/app-svr/feature/service/model/abtest"
)

const (
	_abTest     = `SELECT id,tree_id,key_name,ab_type,bucket,salt,config,relations FROM abtest WHERE state="on" AND tree_id=?`
	_abTreesSQL = `SELECT distinct(tree_id) FROM abtest WHERE state="on"`
)

func (d *Dao) ABTest(c context.Context, treeID int64) (res map[string]*abtestmdl.ABTest, err error) {
	rows, err := d.db.Query(c, _abTest, treeID)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]*abtestmdl.ABTest)
	for rows.Next() {
		re := new(abtestmdl.ABTest)
		if err = rows.Scan(&re.ID, &re.TreeID, &re.KeyName, &re.AbType, &re.Bucket, &re.Salt, &re.Config, &re.Relations); err != nil {
			log.Error("%+v", err)
			return
		}
		res[re.KeyName] = re
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return
}

func (d *Dao) ABTestTrees(c context.Context) (res []int64, err error) {
	rows, err := d.db.Query(c, _abTreesSQL)
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

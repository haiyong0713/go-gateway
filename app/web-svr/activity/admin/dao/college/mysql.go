package college

import (
	"context"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/college"
	"strings"

	"github.com/pkg/errors"
)

const (
	collegeName    = "act_college"
	collegeAidName = "act_college_aid"
)

const (
	insertCollegeSQL     = "INSERT INTO %s (college_name,province_id,province,province_initial,initial,tag_id,white,mid,relation_mid) VALUES %s"
	getCollegeByBatchSQL = "SELECT id,college_name,province_id,province,province_initial,initial,tag_id,white,mid,relation_mid,score,state,ctime,mtime FROM %s ORDER BY id LIMIT ?,?"
	updateCollege        = "INSERT INTO %s (`id`,`college_name`,`province_id`,`province`,`province_initial`,`initial`,`tag_id`,`white`,`mid`,`relation_mid`,`score`,`state`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id=VALUES(id), college_name=VALUES(college_name),province_id=VALUES(province_id),province=VALUES(province),province_initial=VALUES(province_initial),initial=VALUES(initial),tag_id=VALUES(tag_id),white=VALUES(white),mid=VALUES(mid),relation_mid=VALUES(relation_mid),score=VALUES(score),state=VALUES(state)"
	updateAid            = "INSERT INTO %s (`id`,`aid`,`score`,`state`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE id=VALUES(id), aid=VALUES(aid),score=VALUES(score),state=VALUES(state)"
)

// BatchAddCollege batch add college
func (d *dao) BatchAddCollege(c context.Context, tx *xsql.Tx, rank []*college.College) (err error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, r := range rank {
		rowsTmp = append(rowsTmp, "(?,?,?,?,?,?,?,?,?)")

		rows = append(rows, r.CollegeName, r.ProvinceID, r.Province, r.ProvinceInitial, r.Initial, r.TagID, r.White, r.Mid, r.RelationMid)
	}
	sql := fmt.Sprintf(insertCollegeSQL, collegeName, strings.Join(rowsTmp, ","))
	if _, err = tx.Exec(sql, rows...); err != nil {
		err = errors.Wrap(err, "BatchAddCollege: tx.Exec")
	}
	return
}

// UpdateCollegeByID ...
func (d *dao) BacthInsertOrUpdateCollege(c context.Context, collegeInfo *college.College) (err error) {
	if collegeInfo != nil {
		if _, err = d.db.Exec(c, fmt.Sprintf(updateCollege, collegeName), collegeInfo.ID, collegeInfo.CollegeName, collegeInfo.ProvinceID, collegeInfo.Province,
			collegeInfo.ProvinceInitial, collegeInfo.Initial,
			collegeInfo.TagID, collegeInfo.White, collegeInfo.Mid, collegeInfo.RelationMid, collegeInfo.Score, collegeInfo.State); err != nil {
			log.Errorc(c, "lottery@UpdateCollegeByID d.db.Exec() failed. error(%v)", err)
		}
	}
	return

}

// BacthInsertOrUpdateAidList ...
func (d *dao) BacthInsertOrUpdateAidList(c context.Context, aidInfo *college.AIDList) (err error) {
	if aidInfo != nil {
		if _, err = d.db.Exec(c, fmt.Sprintf(updateAid, collegeAidName), aidInfo.ID, aidInfo.Aid, aidInfo.Score, aidInfo.State); err != nil {
			log.Errorc(c, "BacthInsertOrUpdateAidList d.db.Exec() failed. error(%v)", err)
		}
	}
	return
}

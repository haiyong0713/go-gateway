package college

import (
	"context"
	xsql "database/sql"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/college"

	"github.com/pkg/errors"
)

const (
	collegeTableName    = "act_college"
	collegeAidTableName = "act_college_aid"
	collegeMidTableName = "act_college_mid"
)

const (
	allCollegeSQL              = "SELECT id,tag_id,college_name,province_id,province,white,mid,relation_mid,score FROM %s WHERE state = 1 limit ?,?"
	getCollegeMidByBatchSQL    = "SELECT mid,inviter,mid_type FROM %s WHERE state = 1 AND college_id = ? ORDER BY mid LIMIT ?,?"
	getCollegeArchiveAdjustSQL = "SELECT aid,score FROM %s WHERE state = 1 and score!=0 LIMIT ?,?"
	updateCollegeMidScoreSQL   = "UPDATE %s SET `score` = CASE %s END WHERE mid IN (%s)"
	updateCollegeScoreSQL      = "UPDATE %s SET score = score +? WHERE id IN (%s)"
)

// GetAllCollege get all college
func (d *dao) GetAllCollege(c context.Context, offset, limit int64) (rs []*college.College, err error) {
	rs = []*college.College{}
	rows, err := d.db.Query(c, fmt.Sprintf(allCollegeSQL, collegeTableName), offset, limit)
	if err != nil {
		err = errors.Wrap(err, "GetAllCollege:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &college.DB{}
		err = rows.Scan(&r.ID, &r.TagID, &r.Name, &r.ProvinceID, &r.Province, &r.White, &r.MID, &r.RelationMid, &r.Score)
		if err != nil {
			err = errors.Wrap(err, "GetAllCollege:rows.Scan error")
			return
		}
		data := &college.College{
			ID:         r.ID,
			TagID:      r.TagID,
			Name:       r.Name,
			ProvinceID: r.ProvinceID,
			MID:        r.MID,
			Province:   r.Province,
			Score:      r.Score,
		}
		if r.White != "" {
			white, err := xstr.SplitInts(r.White)
			if err != nil {
				log.Errorc(c, "white turn to ints error (%v)", err)
			} else {
				data.White = white
			}
		}
		if r.RelationMid != "" {
			relationMid, err := xstr.SplitInts(r.RelationMid)
			if err != nil {
				log.Errorc(c, "relationMid turn to ints error (%v)", err)
			} else {
				data.RelationMid = relationMid
			}
		}
		rs = append(rs, data)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetAllCollege:rows.Err")
	}
	return
}

// GetCollegeMidByBatch get all college mid by batch
func (d *dao) GetCollegeMidByBatch(c context.Context, collegeID int64, offset, limit int) (rs []*college.MidInfo, err error) {
	rs = make([]*college.MidInfo, 0)
	rows, err := d.db.Query(c, fmt.Sprintf(getCollegeMidByBatchSQL, collegeMidTableName), collegeID, offset, limit)
	if err != nil {
		err = errors.Wrap(err, "GetCollegeMidByBatch:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		midInfo := &college.MidInfo{}
		err = rows.Scan(&midInfo.MID, &midInfo.Inviter, &midInfo.MidType)
		if err != nil {
			err = errors.Wrap(err, "GetCollegeMidByBatch:rows.Scan error")
			return
		}
		rs = append(rs, midInfo)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetCollegeMidByBatch:rows.Err")
	}
	return
}

// GetCollegeAdjustArchive get all college mid by batch
func (d *dao) GetCollegeAdjustArchive(c context.Context, offset, limit int) (rs []*college.Archive, err error) {
	rs = make([]*college.Archive, 0)
	rows, err := d.db.Query(c, fmt.Sprintf(getCollegeArchiveAdjustSQL, collegeAidTableName), offset, limit)
	if err != nil {
		err = errors.Wrap(err, "GetCollegeAdjustArchive:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		archive := &college.Archive{}
		err = rows.Scan(&archive.AID, &archive.Score)
		if err != nil {
			err = errors.Wrap(err, "GetCollegeAdjustArchive:rows.Scan error")
			return
		}
		rs = append(rs, archive)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetCollegeAdjustArchive:rows.Err")
	}
	return
}

// UpdateCollegeMidScore 更新用户积分
func (d *dao) UpdateCollegeMidScore(c context.Context, personals []*college.Personal) (affected int64, err error) {
	var (
		caseStr string
		res     xsql.Result
	)
	mids := make([]int64, 0)
	for _, personal := range personals {
		caseStr = fmt.Sprintf("%s WHEN mid = %d THEN %d", caseStr, personal.MID, personal.Score)
		mids = append(mids, personal.MID)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(updateCollegeMidScoreSQL, collegeMidTableName, caseStr, xstr.JoinInts(mids))); err != nil {
		err = errors.Wrap(err, "UpdateCollegeMidScore:db.Exec error")
		return
	}
	return res.RowsAffected()
}

// UpdateCollegeScore 更新用户积分
func (d *dao) UpdateCollegeScore(c context.Context, collegeIDs string, score int64) (affected int64, err error) {
	var (
		res xsql.Result
	)
	if res, err = d.db.Exec(c, fmt.Sprintf(updateCollegeScoreSQL, collegeTableName, collegeIDs), score); err != nil {
		err = errors.Wrap(err, "UpdateCollegeScore:db.Exec error")
		return
	}
	return res.RowsAffected()
}

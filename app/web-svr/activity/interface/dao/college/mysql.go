package college

import (
	"context"
	xsql "database/sql"
	"fmt"

	sql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	"github.com/pkg/errors"

	"go-gateway/app/web-svr/activity/interface/model/college"
)

const (
	collegeTableName    = "act_college"
	collegeAidTableName = "act_college_aid"
	collegeMidTableName = "act_college_mid"
)

const (
	getMidBindSchoolSQL    = "SELECT mid,college_id FROM %s WHERE mid= ? and state = 1"
	insertMidBindSchoolSQL = "INSERT INTO %s (mid,mid_type,college_id,inviter,year) VALUES(?,?,?,?,?)"
	allProvinceSQL         = "SELECT distinct(province_id),province FROM %s where state = 1"
	allCollegeSQL          = "SELECT id,tag_id,college_name,province_id,province,white,mid,relation_mid,initial FROM %s WHERE state = 1 limit ?,?"
	countInviterSQL        = "SELECT COUNT(1) FROM %s WHERE `inviter`=? AND `mid_type`=? AND `state`=1"
)

// GetMidBindCollege 获得用户绑定的学校
func (d *dao) GetMidBindCollege(c context.Context, mid int64) (rs *college.PersonalCollege, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(getMidBindSchoolSQL, collegeMidTableName), mid)
	rs = &college.PersonalCollege{}
	if err = row.Scan(&rs.MID, &rs.CollegeID); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "GetMidBindCollege:row.Scan()")
		}
	}
	return
}

// MidBindCollege 用户绑定学校
func (d *dao) MidBindCollege(c context.Context, mid int64, midType int, collegeID int64, inviter int64, year int) (lastID int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(insertMidBindSchoolSQL, collegeMidTableName), mid, midType, collegeID, inviter, year); err != nil {
		log.Errorc(c, "AddUserAchieve error d.db.Exec(%d,%d,%d,%d,%d) error(%v)", mid, midType, collegeID, inviter, year, err)
		return
	}
	return res.LastInsertId()
}

// GetAllProvince ...
func (d *dao) GetAllProvince(c context.Context) (res []*college.Province, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, fmt.Sprintf(allProvinceSQL, collegeTableName)); err != nil {
		err = errors.Wrap(err, "GetAllProvince:d.db.Query()")
		return
	}
	defer rows.Close()
	res = make([]*college.Province, 0)
	for rows.Next() {
		var province = &college.Province{}
		if err = rows.Scan(&province.ID, &province.Name); err != nil {
			err = errors.Wrap(err, "GetAllProvince:rows.Scan()")
			return
		}
		res = append(res, province)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "GetAllProvince:rows.Err()")
	}
	return
}

// GetAllCollege get all college
func (d *dao) GetAllCollege(c context.Context, offset, limit int64) (rs []*college.Detail, err error) {
	rs = []*college.Detail{}
	rows, err := d.db.Query(c, fmt.Sprintf(allCollegeSQL, collegeTableName), offset, limit)
	if err != nil {
		err = errors.Wrap(err, "GetAllCollege:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &college.DB{}
		err = rows.Scan(&r.ID, &r.TagID, &r.Name, &r.ProvinceID, &r.Province, &r.White, &r.MID, &r.RelationMid, &r.Initial)
		if err != nil {
			err = errors.Wrap(err, "GetAllCollege:rows.Scan error")
			return
		}
		data := &college.Detail{
			ID:         r.ID,
			TagID:      r.TagID,
			Name:       r.Name,
			ProvinceID: r.ProvinceID,
			MID:        r.MID,
			Initial:    r.Initial,
			Province:   r.Province,
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

// CountInviterNum .
func (d *dao) CountInviterNum(c context.Context, inviter int64, inviterType int) (count int, err error) {
	row := d.db.QueryRow(c, fmt.Sprintf(countInviterSQL, collegeMidTableName), inviter, inviterType)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "CountInviterNum:row.Scan()")
		}
	}
	return
}

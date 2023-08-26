package summer_camp

import (
	"context"
	xsql "database/sql"
	"fmt"
	"go-common/library/log"
	mdlSC "go-gateway/app/web-svr/activity/interface/model/summer_camp"
	"strings"
	"time"
)

const (
	_selCourseListSQL      = "SELECT course_id,course_title,pic_cover,bodan_id,creator,status,ctime,mtime from camp_course where status=1 limit ?,?"
	_selUserCourseSQL      = "SELECT id,mid,course_id,course_title,status,join_time,ctime,mtime from user_course_info where mid = ? and status=1 order by join_time desc,id desc limit ?,? "
	_selUserCourseById     = "SELECT id,mid,course_id,course_title,status,join_time,ctime,mtime from user_course_info where status=1 and course_id =? and mid =?"
	_multiInsertUserCourse = "insert into user_course_info (`mid`, `course_id`, `course_title`, `status`, `join_time`) values %s"
)

// GetCourseList get all course.
func (d *dao) GetCourseList(ctx context.Context, offset, limit int) (res []*mdlSC.DBCourseCamp, err error) {
	res = []*mdlSC.DBCourseCamp{}

	rows, err := d.db.Query(ctx, _selCourseListSQL, offset, limit)
	if err != nil {
		log.Errorc(ctx, "GetCourseList:d.db.Query error.error detail is(%+v)", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		r := &mdlSC.DBCourseCamp{}
		err = rows.Scan(&r.CourseID, &r.CourseTitle, &r.PicCover, &r.BodanId, &r.Creator, &r.Status, &r.Ctime, &r.Mtime)
		if err != nil {
			log.Errorc(ctx, "GetCourseList:rows.Scan error.error detail is(%+v)", err)
			return
		}

		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "GetPlanList:rows.Err.error detail is(%+v)", err)
	}
	return
}

// GetUserCourse
func (d *dao) GetUserCourse(ctx context.Context, mid int64, offset, limit int) (res []*mdlSC.DBUserCourse, courseIds []int64, err error) {
	res = []*mdlSC.DBUserCourse{}
	courseIds = make([]int64, 0)
	// 读缓存
	res, err = d.CacheGetUserCourseList(ctx, mid)
	if err != nil || len(res) <= 0 {
		log.Errorc(ctx, "GetUserCourse get cache err or res is nil,err is (%v).", err)
	}
	if len(res) > 0 {
		for _, ci := range res {
			courseIds = append(courseIds, ci.CourseID)
		}
		return
	}
	// 回源
	rows, err := d.db.Query(ctx, _selUserCourseSQL, mid, offset, limit)
	if err != nil {
		log.Errorc(ctx, "GetUserCourse:d.db.Query error.error detail is(%+v)", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		r := &mdlSC.DBUserCourse{}
		err = rows.Scan(&r.ID, &r.Mid, &r.CourseID, &r.CourseTitle, &r.Status, &r.JoinTime, &r.Ctime, &r.Mtime)
		if err != nil {
			log.Errorc(ctx, "GetUserCourse:rows.Scan error.error detail is(%+v)", err)
			return
		}

		res = append(res, r)
		courseIds = append(courseIds, r.CourseID)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "GetUserCourse:rows.Err.error detail is(%+v)", err)
		return
	}
	// 塞缓存
	err = d.CacheSetUserCourseList(ctx, mid, res)
	if err != nil {
		log.Errorc(ctx, "GetUserCourse d.CacheSetUserCourseList err,err is (%v).", err)
		// 吞掉
		err = nil
	}
	return
}

// GetUserCourseById
func (d *dao) GetUserCourseById(ctx context.Context, mid int64, courseId int64) (r *mdlSC.DBUserCourse, err error) {
	// 读缓存
	r, err = d.CacheGetUserCourseByID(ctx, mid, courseId)
	if err != nil || r == nil {
		log.Errorc(ctx, "GetUserCourseById d.CacheGetUserCourseByID err or redis is nil,err is (%v).", err)
	}
	if r != nil {
		return
	}
	// 回源
	r = &mdlSC.DBUserCourse{}
	rows, err := d.db.Query(ctx, _selUserCourseById, courseId, mid)
	if err != nil {
		log.Errorc(ctx, "GetUserCourseById:d.db.Query error.error detail is(%+v)", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&r.ID, &r.Mid, &r.CourseID, &r.CourseTitle, &r.Status, &r.JoinTime, &r.Ctime, &r.Mtime)
		if err != nil {
			log.Errorc(ctx, "GetUserCourseById:rows.Scan error.error detail is(%+v)", err)
			return
		}

	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "GetUserCourseById:rows.Err.error detail is(%+v)", err)
		return
	}
	// 塞缓存
	err = d.CacheSetUserCourseByID(ctx, mid, courseId, r)
	if err != nil {
		log.Errorc(ctx, "GetUserCourseById d.CacheSetUserCourseByID err,err is (%v).", err)
		// 吞掉
		err = nil
	}
	return
}

// MultiInsertUserCourse
func (d *dao) MultiInsertUserCourse(ctx context.Context, records []*mdlSC.DBUserCourse) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}
	var (
		rowStrings []string
		param      []interface{}
	)
	now := time.Now()
	for _, v := range records {
		rowStrings = append(rowStrings, "(?,?,?,?,?)")
		param = append(param, v.Mid, v.CourseID, v.CourseTitle, v.Status, now)
	}
	sqlStr := fmt.Sprintf(_multiInsertUserCourse, strings.Join(rowStrings, ","))
	res, err := d.db.Exec(ctx, sqlStr, param...)
	if err != nil {
		log.Errorc(ctx, "MultiInsertUserCourse:d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.RowsAffected()

}

const sql4InsertUpdateUserCourse = `
INSERT INTO user_course_info(mid,course_id,course_title,status,join_time) 
VALUES %s ON DUPLICATE KEY UPDATE
course_title=values(course_title),
status=values(status),
join_time=values(join_time);
`

// MultiInsertOrUpdateUserCourse
func (d *dao) MultiInsertOrUpdateUserCourse(ctx context.Context, mid int64, records []*mdlSC.DBUserCourse) (int64, error) {
	var (
		rowsValue []interface{}
		rowsParam []string
	)
	now := time.Now()
	for _, r := range records {
		rowsParam = append(rowsParam, "(?,?,?,?,?)")
		rowsValue = append(rowsValue, r.Mid, r.CourseID, r.CourseTitle, r.Status, now)
	}
	sql := fmt.Sprintf(sql4InsertUpdateUserCourse, strings.Join(rowsParam, ","))
	res, err := d.db.Exec(ctx, sql, rowsValue...)
	if err != nil {
		log.Errorc(ctx, "MultiInsertOrUpdateUserCourse db.Exec() error(%+v)", err)
		return 0, err
	}
	// 删缓存
	err = d.CacheDelUserCourseList(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "MultiInsertOrUpdateUserCourse d.CacheDelUserCourseList err ,err is (%v)", err)
		// 吞掉
		err = nil
	}
	return res.RowsAffected()
}

const _updateUserRecordByMidCourse = "UPDATE user_course_info SET status = ? WHERE mid = ? and course_id = ?"

// SingleQuitJoin
func (d *dao) SingleQuitJoin(ctx context.Context, mid int64, record *mdlSC.DBUserCourse) (affected int64, err error) {
	if record == nil {
		return
	}
	var (
		res xsql.Result
	)
	if res, err = d.db.Exec(ctx, _updateUserRecordByMidCourse, record.Status, record.Mid, record.CourseID); err != nil {
		log.Errorc(ctx, "singleQuitJoin:db.Exec error is :(%v).", err)
		return
	}
	// 删列表缓存
	err = d.CacheDelUserCourseList(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "MultiInsertOrUpdateUserCourse d.CacheDelUserCourseList err ,err is (%v)", err)
		// 吞掉
		err = nil
	}
	// 删单个缓存
	err = d.CacheDelUserCourseByID(ctx, mid, record.CourseID)
	if err != nil {
		log.Errorc(ctx, "MultiInsertOrUpdateUserCourse d.CacheDelUserCourseByID err ,err is (%v)", err)
		// 吞掉
		err = nil
	}
	return res.RowsAffected()

}

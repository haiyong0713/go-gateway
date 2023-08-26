package summer_camp

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	mdlSC "go-gateway/app/web-svr/activity/interface/model/summer_camp"
	"strings"
	"time"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package fit

const (
	prefix            = "SummerCamp"
	separator         = ":"
	courseListKey     = "course_list"
	userCourseByIdKey = "user_course_%d_%d"
	userCourseKey     = "user_course_%d"
)

// Dao dao interface
type Dao interface {
	GetCourseList(ctx context.Context, offset, limit int) ([]*mdlSC.DBCourseCamp, error)
	GetUserCourse(ctx context.Context, mid int64, offset, limit int) (res []*mdlSC.DBUserCourse, courseIds []int64, err error)
	CacheGetCourseList(ctx context.Context) (res []*mdlSC.DBCourseCamp, err error)
	CacheSetCourseList(ctx context.Context, data []*mdlSC.DBCourseCamp) (err error)
	GetUserCourseById(ctx context.Context, mid int64, courseId int64) (r *mdlSC.DBUserCourse, err error)
	MultiInsertUserCourse(ctx context.Context, records []*mdlSC.DBUserCourse) (int64, error)
	MultiInsertOrUpdateUserCourse(ctx context.Context, mid int64, records []*mdlSC.DBUserCourse) (int64, error)
	SingleQuitJoin(ctx context.Context, mid int64, record *mdlSC.DBUserCourse) (affected int64, err error)
	CacheGetUserCourseByID(ctx context.Context, mid int64, courseId int64) (res *mdlSC.DBUserCourse, err error)
	CacheSetUserCourseByID(ctx context.Context, mid int64, courseId int64, data *mdlSC.DBUserCourse) (err error)
	CacheDelUserCourseByID(c context.Context, mid int64, courseId int64) (err error)
	CacheGetUserCourseList(ctx context.Context, mid int64) (res []*mdlSC.DBUserCourse, err error)
	CacheSetUserCourseList(ctx context.Context, mid int64, data []*mdlSC.DBUserCourse) (err error)
	CacheDelUserCourseList(c context.Context, mid int64) (err error)
}

// Dao dao.
type dao struct {
	c                *conf.Config
	redis            *redis.Redis
	db               *xsql.DB
	courseListExpire int32
}

// New init
func newDao(c *conf.Config) (nd Dao) {
	nd = &dao{
		c:                c,
		redis:            component.GlobalRedis,
		db:               component.GlobalDB,
		courseListExpire: int32(time.Duration(c.Redis.SummerCampCourseListExpire) / time.Second),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}

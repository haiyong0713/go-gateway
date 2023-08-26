package college

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/college"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package college

const (
	prefix            = "college2020"
	separator         = ":"
	midKey            = "mid"
	midCollegeKey     = "mid_college"
	collegeKey        = "college"
	idKey             = "id"
	scoreKey          = "score"
	nationwideRankKey = "nationwide_rank"
	provinceRankKey   = "province_rank"
	provinceKey       = "province"
	tabListKey        = "tab_list"
	relationKey       = "relation"
	nameKey           = "name"
	tabKey            = "college_tab"
	versionKey        = "version"
	inviterCountKey   = "inviter_count"
	followerKey       = "is_follower"
)

// Dao dao interface
type Dao interface {
	CacheSetMidCollege(c context.Context, mid int64, data *college.PersonalCollege) (err error)
	CacheGetMidCollege(c context.Context, mid int64) (res *college.PersonalCollege, err error)
	CacheGetCollegeDetail(c context.Context, collegeID int64, version int) (res *college.Detail, err error)
	GetCollegeVersion(c context.Context) (res *college.Version, err error)
	GetCollegePersonal(c context.Context, mid int64, version int) (res *college.Personal, err error)
	GetArchiveTabArchive(c context.Context, collegeID int64, tabID int) (res []int64, err error)
	DelCacheMidInviter(c context.Context, mid int64) (err error)
	CacheMidInviter(c context.Context, mid int64) (list map[string]int64, err error)
	AddCacheMidInviter(c context.Context, mid int64, list map[string]int64) (err error)
	MidFollow(c context.Context, mid int64) (err error)
	MidIsFollow(c context.Context, mid int64) (res int, err error)

	GetMidBindCollege(c context.Context, mid int64) (rs *college.PersonalCollege, err error)
	MidBindCollege(c context.Context, mid int64, midType int, collegeID int64, inviter int64, year int) (lastID int64, err error)
	GetAllProvince(c context.Context) (res []*college.Province, err error)
	GetAllCollege(c context.Context, offset, limit int64) (rs []*college.Detail, err error)
	CountInviterNum(c context.Context, inviter int64, inviterType int) (count int, err error)

	SendPoint(c context.Context, mid int64, data *college.ActPlatActivityPoints) (err error)

	Close()
}

// Dao dao.
type dao struct {
	c                       *conf.Config
	redis                   *redis.Pool
	db                      *xsql.DB
	collegeMidCollegeExpire int32
	actPlatPub              *databus.Databus
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:                       c,
		redis:                   redis.NewPool(c.Redis.Store),
		db:                      component.GlobalDB,
		actPlatPub:              databus.New(c.DataBus.ActPlatPub),
		collegeMidCollegeExpire: int32(time.Duration(c.Redis.CollegeMidCollegeExpire) / time.Second),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {
	if d.redis != nil {
		d.redis.Close()
	}
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}

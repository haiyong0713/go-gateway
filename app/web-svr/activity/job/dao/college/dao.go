package college

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/model/college"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package college

const (
	prefix            = "college2020"
	separator         = ":"
	midKey            = "mid"
	collegeKey        = "college"
	idKey             = "id"
	scoreKey          = "score"
	nationwideRankKey = "nationwide_rank"
	provinceKey       = "province"
	provinceRankKey   = "province_rank"
	tabListKey        = "tab_list"
	relationKey       = "relation"
	nameKey           = "name"
	tabKey            = "college_tab"
	versionKey        = "version"
	versionUpdateKey  = "version_update"
)

// Dao dao interface
type Dao interface {
	Close()
	GetAllCollege(c context.Context, offset, limit int64) (rs []*college.College, err error)
	GetCollegeMidByBatch(c context.Context, collegeID int64, offset, limit int) (rs []*college.MidInfo, err error)
	SetMidPersonal(c context.Context, midRank []*college.Personal, version int) (err error)
	SetCollegeDetail(c context.Context, college *college.Detail, version int) (err error)
	GetCollegeAdjustArchive(c context.Context, offset, limit int) (rs []*college.Archive, err error)
	SetArchiveTabArchive(c context.Context, collegeID int64, tabID int64, aids []int64) (err error)
	GetCollegeUpdateVersion(c context.Context) (version *college.Version, err error)
	SetCollegeUpdateVersion(c context.Context, version *college.Version) (err error)
	SetCollegeVersion(c context.Context, version *college.Version) (err error)
	UpdateCollegeMidScore(c context.Context, personals []*college.Personal) (affected int64, err error)
	UpdateCollegeScore(c context.Context, collegeIDs string, score int64) (affected int64, err error)

	SendPoint(c context.Context, mid int64, data *college.ActPlatActivityPoints) (err error)

	Ping(c context.Context) error
}

// Dao dao.
type dao struct {
	c          *conf.Config
	redis      *redis.Pool
	db         *xsql.DB
	actPlatPub *databus.Databus
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:          c,
		db:         sql.NewMySQL(c.MySQL.Like),
		redis:      redis.NewPool(c.Redis.Config),
		actPlatPub: initialize.NewDatabusV1(c.ActPlatPub),
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

// Ping ping
func (d *dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}

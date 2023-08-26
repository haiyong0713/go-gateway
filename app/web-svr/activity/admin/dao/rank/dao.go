package rank

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	xsql "go-common/library/database/sql"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/conf"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank"
	"strings"

	"github.com/jinzhu/gorm"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package rank

const (
	prefix    = "act_rank"
	separator = ":"
	rankKey   = "rank"
	midKey    = "mid"
)

// Dao dao interface
type Dao interface {
	Close()
	BeginTran(c context.Context) (tx *xsql.Tx, err error)

	Create(c context.Context, tx *xsql.Tx, sid int64, sidSource int, stime, etime xtime.Time) (id int64, err error)
	GetRankConfigBySid(c context.Context, sid int64, sidSource int) (rank *rankmdl.Rank, err error)
	GetRankConfigByID(c context.Context, id int64) (rank *rankmdl.Rank, err error)
	UpdateRankConfig(c context.Context, id int64, rank *rankmdl.Rank) (err error)
	BacthInsertOrUpdateBlackOrWhite(c context.Context, id int64, intervention []*rankmdl.Intervention) (err error)
	AllIntervention(c context.Context, id int64, objectType, interventionType, offset, limit int) (list []*rankmdl.Intervention, err error)
	AllInterventionTotal(c context.Context, id int64, objectType, interventionType int) (total int, err error)
	OidRankInRank(c context.Context, id int64, lastBatch int64, rankAttribute, offset, limit int) (list []*rankmdl.OidResult, err error)
	AllSnapshotByMids(c context.Context, id int64, mids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error)
	AllSnapshotByAids(c context.Context, id int64, aids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error)
	GetLastBatch(c context.Context, id int64, rankAttribute int) (rank *rankmdl.Log, err error)
	OidRankInRankTotal(c context.Context, id int64, lastBatch int64, rankAttribute int) (total int, err error)
	BacthInsertOrUpdateBlackOrWhiteTx(c context.Context, tx *xsql.Tx, id int64, intervention []*rankmdl.Intervention) (err error)
	BacthInsertOrUpdateSnapshotTx(c context.Context, tx *xsql.Tx, id int64, snapShot []*rankmdl.Snapshot) (err error)
	BacthInsertOrUpdateOidResultTx(c context.Context, tx *xsql.Tx, id int64, oidResult []*rankmdl.OidResult) (err error)
	AllOidRank(c context.Context, id int64, lastBatch int64, rankAttribute, offset, limit int) (list []*rankmdl.OidResult, err error)
	GetRankConfigByIDAll(c context.Context, id int64) (rank *rankmdl.Rank, err error)
	GetRank(c context.Context, rankActivityKey string) (res []*rankmdl.MidRank, err error)

	SetRank(c context.Context, rankNameKey string, rankBatch []*rankmdl.MidRank) (err error)
	SetMidRank(c context.Context, rankNameKey string, midRank []*rankmdl.MidRank) (err error)
}

// BeginTran begin transcation.
func (d *dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// Dao dao.
type dao struct {
	c     *conf.Config
	db    *xsql.DB
	redis *redis.Pool
	DB    *gorm.DB
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:     c,
		DB:    orm.NewMySQL(c.ORM),
		db:    xsql.NewMySQL(c.MySQL.Lottery),
		redis: redis.NewPool(c.Redis.Store),
	}
	return
}

func (d *dao) GetDB() *gorm.DB {
	return d.DB
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

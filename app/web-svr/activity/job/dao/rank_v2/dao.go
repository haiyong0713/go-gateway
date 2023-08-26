package rank

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
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
	Ping(c context.Context) error
	AllIntervention(c context.Context, id int64, objectType int, offset, limit int) (list []*rankmdl.Intervention, err error)
	AllOidRank(c context.Context, id int64, lastBatch, rankAttribute, offset, limit int) (list []*rankmdl.OidResult, err error)
	BatchAddOidRank(c context.Context, tx *xsql.Tx, id int64, rank []*rankmdl.OidResult) (err error)
	BatchAddSnapshotRank(c context.Context, tx *xsql.Tx, id int64, rank []*rankmdl.Snapshot) (err error)
	InsertRankLog(c context.Context, id int64, lastBatch, lastAttribute, state int) (int64, error)
	GetRankConfigOnline(c context.Context, now time.Time) (list []*rankmdl.Rank, err error)
	GetRankLog(c context.Context, rankID int64, batch, attributeType int) (rank *rankmdl.Log, err error)
	GetRankLogOrderByTime(c context.Context, rankID int64, attributeType int) (rank *rankmdl.Log, err error)
	UpdateRankLog(c context.Context, tx *xsql.Tx, id int64, state int) (err error)
	GetRankLogOrderByTimeAll(c context.Context, rankID int64, attributeType int) (rank *rankmdl.Log, err error)
	AllSnapshotByAids(c context.Context, id int64, aids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error)
	SnapshotByAllAids(c context.Context, id int64, aids []int64, lastBatch int64, rankAttribute int) (list []*rankmdl.Snapshot, err error)

	SetRank(c context.Context, rankNameKey string, rankBatch []*rankmdl.MidRank) (err error)
	SetMidRank(c context.Context, rankNameKey string, midRank []*rankmdl.MidRank) (err error)
	GetRankConfigByID(c context.Context, id int64) (rank *rankmdl.Rank, err error)

	SendWeChat(c context.Context, publicKey, title, msg, user string) (err error)
}

// BeginTran begin transcation.
func (d *dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// Dao dao.
type dao struct {
	c          *conf.Config
	redis      *redis.Pool
	db         *xsql.DB
	dataExpire int32
	client     *xhttp.Client
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:          c,
		db:         xsql.NewMySQL(c.MySQL.Like),
		redis:      redis.NewPool(c.Redis.Store),
		client:     xhttp.NewClient(c.HTTPClient),
		dataExpire: int32(time.Duration(c.Redis.RankDataExpire) / time.Second),
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

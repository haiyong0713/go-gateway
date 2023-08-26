package audit

import (
	"context"
	"encoding/json"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-job/job/conf"

	"github.com/pkg/errors"
)

const (
	_getSQL        = "SELECT mobi_app,build FROM audit"
	_auditRedisKey = "audit"
	_splitToken    = ":"
	_auditExpire   = 604800
)

// Dao is audit dao.
type Dao struct {
	db     *sql.DB
	audGet *sql.Stmt
	redis  *redis.Pool
}

// New new a audit dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:    sql.NewMySQL(c.MySQL.Show),
		redis: redis.NewPool(c.Redis.Recommend.Config),
	}
	d.audGet = d.db.Prepared(_getSQL)
	return
}

// Audits get all audit build.
func (d *Dao) Audits(ctx context.Context) (map[string]map[int]struct{}, error) {
	rows, err := d.audGet.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var (
		mobiApp string
		build   int
	)
	res := map[string]map[int]struct{}{}
	for rows.Next() {
		if err = rows.Scan(&mobiApp, &build); err != nil {
			return nil, err
		}
		if plat, ok := res[mobiApp]; ok {
			plat[build] = struct{}{}
		} else {
			res[mobiApp] = map[int]struct{}{
				build: {},
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheAudit(ctx context.Context, as map[string]map[int]struct{}) error {
	if len(as) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(as)
	if err != nil {
		return errors.WithStack(err)
	}
	key := auditActionKey("loadAuditCache", "struct")
	if _, err := conn.Do("SETEX", key, _auditExpire, bs); err != nil {
		return err
	}
	return nil
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		_ = d.db.Close()
	}
	if d.redis != nil {
		_ = d.redis.Close()
	}
}

func auditActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_auditRedisKey)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}

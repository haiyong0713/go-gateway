package audit

import (
	"context"
	"encoding/json"
	"go-gateway/app/app-svr/app-show/interface/component"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-show/interface/conf"
)

const (
	_auditRedisKeyPrefix = "audit"
	_splitToken          = ":"
)

// Dao is audit dao.
type Dao struct {
	db    *sql.DB
	redis *redis.Pool
}

// New new a audit dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:    component.GlobalShowDB,
		redis: redis.NewPool(c.Redis.Recommend.Config),
	}
	return
}

// Audits get all audit build.
func (d *Dao) Audits(ctx context.Context) (map[string]map[int]struct{}, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", auditActionKey("loadAuditCache", "struct")))
	if err != nil {
		return nil, err
	}
	var res map[string]map[int]struct{}
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Close close resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}

func auditActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_auditRedisKeyPrefix)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}

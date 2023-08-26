package database

import (
	"go-common/library/net/netutil/breaker"
	"go-common/library/time"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	// database driver
	_ "github.com/ClickHouse/clickhouse-go"
)

// Config mysql config.
type Config struct {
	Addr         string          // for trace
	DSN          string          // write data source name.
	ReadDSN      []string        // read data source name.
	Active       int             // pool
	Idle         int             // pool
	IdleTimeout  time.Duration   // connect max life time.
	QueryTimeout time.Duration   // query sql timeout
	ExecTimeout  time.Duration   // execute sql timeout
	TranTimeout  time.Duration   // transaction sql timeout
	Breaker      *breaker.Config // breaker
}

func NewClickhouse(c *Config) (db *DB) {
	if c.QueryTimeout == 0 || c.ExecTimeout == 0 || c.TranTimeout == 0 {
		panic("clickhouse must be set query/execute/transction timeout")
	}
	db, err := Open(c)
	if err != nil {
		log.Error("open clickhouse error(%v)", err)
		panic(err)
	}
	return
}

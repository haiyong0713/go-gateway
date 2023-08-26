package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-common/library/net/netutil/breaker"
	"go-common/library/net/trace"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

const (
	_family          = "clickhouse_client"
	_slowLogDuration = time.Millisecond * 250
)

var (
	// ErrStmtNil prepared stmt error
	ErrStmtNil = errors.New("sql: prepare failed and stmt nil")
	// ErrNoMaster is returned by Master when call master multiple times.
	ErrNoMaster = errors.New("sql: no master instance")
	// ErrNoRows is returned by Scan when QueryRow doesn't return a row.
	// In such a case, QueryRow returns a placeholder *Row value that defers
	// this error until a Scan.
	ErrNoRows = sql.ErrNoRows
	// ErrTxDone transaction done.
	ErrTxDone = sql.ErrTxDone
)

// DB database.
type DB struct {
	conn *conn
}

// conn database connection
type conn struct {
	*sql.DB
	breaker breaker.Breaker
	conf    *Config
	addr    string
}

// Row row.
type Row struct {
	err error
	*sql.Row
	db     *conn
	query  string
	args   []interface{}
	t      trace.Trace
	cancel func()
}

// Rows rows.
type Rows struct {
	*sql.Rows
	cancel func()
}

// Tx transaction.
type Tx struct {
	db     *conn
	tx     *sql.Tx
	t      trace.Trace
	c      context.Context
	cancel func()
}

func (db *conn) onBreaker(err *error) {
	if err != nil && *err != nil && *err != sql.ErrNoRows && *err != sql.ErrTxDone {
		db.breaker.MarkFailed()
	} else {
		db.breaker.MarkSuccess()
	}
}

func slowLog(statement string, now time.Time) {
	du := time.Since(now)
	if du > _slowLogDuration {
		log.Warn("%s slow log statement: %s time: %v", _family, statement, du)
	}
}

func Open(c *Config) (*DB, error) {
	db := new(DB)
	d, err := connect(c, c.DSN)
	if err != nil {
		return nil, err
	}
	brkGroup := breaker.NewGroup(c.Breaker)
	brk := brkGroup.Get(c.Addr)
	conn := &conn{DB: d, breaker: brk, conf: c, addr: c.Addr}
	db.conn = conn
	return db, nil
}

func connect(c *Config, dataSourceName string) (*sql.DB, error) {
	d, err := sql.Open("clickhouse", dataSourceName)
	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}
	d.SetMaxOpenConns(c.Active)
	d.SetMaxIdleConns(c.Idle)
	d.SetConnMaxLifetime(time.Duration(c.IdleTimeout))
	return d, nil
}

func (r *Row) Scan(dest ...interface{}) (err error) {
	defer slowLog(fmt.Sprintf("Scan query(%s) args(%+v)", r.query, r.args), time.Now())
	if r.t != nil {
		defer r.t.Finish(&err)
	}
	if r.err != nil {
		err = r.err
	} else if r.Row == nil {
		err = ErrStmtNil
	}
	if err != nil {
		return
	}
	err = r.Row.Scan(dest...)
	if r.cancel != nil {
		r.cancel()
	}
	r.db.onBreaker(&err)
	if err != ErrNoRows {
		err = errors.Wrapf(err, "query %s args %+v", r.query, r.args)
	}
	return
}

func (db *DB) GetConnect() (conn *conn) {
	return db.conn
}

func (db *DB) Begin(c context.Context) (tx *Tx, err error) {
	return db.conn.begin(c)
}

func (db *conn) begin(c context.Context) (tx *Tx, err error) {
	now := time.Now()
	defer slowLog("Begin", now)
	t, ok := trace.FromContext(c)
	if ok {
		t = t.Fork(_family, "begin")
		t.SetTag(trace.String(trace.TagAddress, db.addr), trace.String(trace.TagComment, ""))
		defer func() {
			if err != nil {
				t.Finish(&err)
			}
		}()
	}
	if err = db.breaker.Allow(); err != nil {
		_metricReqErr.Inc(db.addr, db.addr, "begin", "breaker")
		return
	}
	_, c, cancel := db.conf.TranTimeout.Shrink(c)
	rtx, err := db.BeginTx(c, nil)
	_metricReqDur.Observe(int64(time.Since(now)/time.Millisecond), db.addr, db.addr, "begin")
	if err != nil {
		err = errors.WithStack(err)
		cancel()
		return
	}
	tx = &Tx{tx: rtx, t: t, db: db, c: c, cancel: cancel}
	return
}

// Commit commits the transaction.
func (tx *Tx) Commit() (err error) {
	err = tx.tx.Commit()
	tx.cancel()
	tx.db.onBreaker(&err)
	if tx.t != nil {
		tx.t.Finish(&err)
	}
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() (err error) {
	err = tx.tx.Rollback()
	tx.cancel()
	tx.db.onBreaker(&err)
	if tx.t != nil {
		tx.t.Finish(&err)
	}
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

func (tx *Tx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	now := time.Now()
	defer slowLog(fmt.Sprintf("Exec query(%s) args(%+v)", query, args), now)
	if tx.t != nil {
		tx.t.SetTag(trace.String(trace.TagAnnotation, fmt.Sprintf("exec %s", query)))
	}
	res, err = tx.tx.ExecContext(tx.c, query, args...)
	_metricReqDur.Observe(int64(time.Since(now)/time.Millisecond), tx.db.addr, tx.db.addr, "tx:exec")
	if err != nil {
		err = errors.Wrapf(err, "exec:%s, args:%+v", query, args)
	}
	return
}

func (db *DB) Exec(c context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	return db.conn.exec(c, query, args...)
}

func (db *conn) exec(c context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	now := time.Now()
	defer slowLog(fmt.Sprintf("Exec query(%s) args(%+v)", query, args), now)
	if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "exec")
		t.SetTag(trace.String(trace.TagAddress, db.addr), trace.String(trace.TagComment, query))
		defer t.Finish(&err)
	}
	if err = db.breaker.Allow(); err != nil {
		_metricReqErr.Inc(db.addr, db.addr, "exec", "breaker")
		return
	}
	_, c, cancel := db.conf.ExecTimeout.Shrink(c)
	res, err = db.ExecContext(c, query, args...)
	cancel()
	db.onBreaker(&err)
	_metricReqDur.Observe(int64(time.Since(now)/time.Millisecond), db.addr, db.addr, "exec")
	if err != nil {
		err = errors.Wrapf(err, "exec:%s, args:%+v", query, args)
	}
	return
}

func (db *DB) QueryRow(c context.Context, query string, args ...interface{}) *Row {
	return db.conn.queryRow(c, query, args...)
}

func (db *conn) queryRow(c context.Context, query string, args ...interface{}) *Row {
	now := time.Now()
	defer slowLog(fmt.Sprintf("QueryRow query(%s) args(%+v)", query, args), now)
	t, ok := trace.FromContext(c)
	if ok {
		t = t.Fork(_family, "queryrow")
		t.SetTag(trace.String(trace.TagAddress, db.addr), trace.String(trace.TagComment, query))
	}
	if err := db.breaker.Allow(); err != nil {
		_metricReqErr.Inc(db.addr, db.addr, "queryRow", "breaker")
		return &Row{db: db, t: t, err: err}
	}
	_, c, cancel := db.conf.QueryTimeout.Shrink(c)
	r := db.DB.QueryRowContext(c, query, args...)
	_metricReqDur.Observe(int64(time.Since(now)/time.Millisecond), db.addr, db.addr, "queryrow")
	return &Row{db: db, Row: r, query: query, args: args, t: t, cancel: cancel}
}

func (db *DB) Query(c context.Context, query string, args ...interface{}) (rows *Rows, err error) {
	return db.conn.query(c, query, args...)
}

func (db *conn) query(c context.Context, query string, args ...interface{}) (rows *Rows, err error) {
	now := time.Now()
	defer slowLog(fmt.Sprintf("Query query(%s) args(%+v)", query, args), now)
	if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "query")
		t.SetTag(trace.String(trace.TagAddress, db.addr), trace.String(trace.TagComment, query))
		defer t.Finish(&err)
	}
	if err = db.breaker.Allow(); err != nil {
		_metricReqErr.Inc(db.addr, db.addr, "query", "breaker")
		return
	}
	_, c, cancel := db.conf.QueryTimeout.Shrink(c)
	// nolint:bilisqlclosecheck,rowserrcheck
	rs, err := db.DB.QueryContext(c, query, args...)
	db.onBreaker(&err)
	_metricReqDur.Observe(int64(time.Since(now)/time.Millisecond), db.addr, db.addr, "query")
	if err != nil {
		err = errors.Wrapf(err, "query:%s, args:%+v", query, args)
		cancel()
		return
	}
	rows = &Rows{Rows: rs, cancel: cancel}
	return
}

//nolint:unused
func (db *conn) ping(c context.Context) (err error) {
	now := time.Now()
	defer slowLog("Ping", now)
	if t, ok := trace.FromContext(c); ok {
		t = t.Fork(_family, "ping")
		t.SetTag(trace.String(trace.TagAddress, db.addr), trace.String(trace.TagComment, ""))
		defer t.Finish(&err)
	}
	if err = db.breaker.Allow(); err != nil {
		_metricReqErr.Inc(db.addr, db.addr, "ping", "breaker")
		return
	}
	_, c, cancel := db.conf.ExecTimeout.Shrink(c)
	err = db.PingContext(c)
	cancel()
	db.onBreaker(&err)
	_metricReqDur.Observe(int64(time.Since(now)/time.Millisecond), db.addr, db.addr, "ping")
	if err != nil {
		err = errors.WithStack(err)
	}
	return
}

func (rs *Rows) Close() (err error) {
	err = errors.WithStack(rs.Rows.Close())
	if rs.cancel != nil {
		rs.cancel()
	}
	return
}

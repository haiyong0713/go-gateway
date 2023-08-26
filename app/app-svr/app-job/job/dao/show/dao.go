package show

import (
	"context"
	"strings"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
	v1 "go-gateway/app/app-svr/app-show/interface/api"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const (
	// tran pub
	_ptimeSQL           = "SELECT plat FROM show_time WHERE state=1 AND ptime<?"
	_upStSQL            = "UPDATE show_time SET state=0 WHERE plat=?"
	_delHdSQL           = "DELETE FROM show_head WHERE plat=?"
	_delItSQL           = "DELETE FROM show_item WHERE plat=?"
	_cpHdSQL            = "INSERT INTO show_head(id,plat,title,type,style,param,rank,build,conditions,lang_id,ctime,mtime) SELECT id,plat,title,type,style,param,rank,build,conditions,lang_id,ctime,mtime FROM show_head_temp WHERE plat=?"
	_cpItSQL            = "INSERT INTO show_item(id,sid,plat,title,random,cover,param,ctime,mtime) SELECT id,sid,plat,title,random,cover,param,ctime,mtime FROM show_item_temp WHERE plat=?"
	_showRedisKeyPrefix = "show"
	_splitToken         = ":"
	_showExpire         = 604800
)

// Dao is show dao.
type Dao struct {
	conf                *conf.Config
	client              *httpx.Client
	db                  *sql.DB
	getPTime            *sql.Stmt
	mc                  *memcache.Pool
	showGrpc            v1.AppShowClient
	tagGRPC             taggrpc.TagRPCClient
	expireMC            int32
	redis               *redis.Pool
	selectedRedis       *redis.Pool
	entranceURL         string
	aggregationmc       *memcache.Pool
	aggURL              string
	hotHeTongtabcardURL string
	// database cron_job
	getHead    *sql.Stmt
	getItem    *sql.Stmt
	getHeadTmp *sql.Stmt
	getItemTmp *sql.Stmt
}

// New new a show dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf:                c,
		db:                  sql.NewMySQL(c.MySQL.Show),                // mysql
		client:              httpx.NewClient(c.HTTPClient),             // http client
		mc:                  memcache.NewPool(c.Memcache.Cards.Config), // cards cache for weekly selected
		expireMC:            int32(time.Duration(c.Memcache.Cards.ExpireAggregation) / time.Second),
		redis:               redis.NewPool(c.Redis.Entrance.Config),
		selectedRedis:       redis.NewPool(c.Redis.Recommend.Config),
		entranceURL:         c.Host.Data + _entranceURL,
		aggregationmc:       memcache.NewPool(c.Memcache.Aggregation.Config),
		aggURL:              c.Host.Data + _aggURL,
		hotHeTongtabcardURL: c.Host.Data + _hotHeTongtabcardURL,
	}
	d.getPTime = d.db.Prepared(_ptimeSQL)
	d.getHead = d.db.Prepared(_headSQL)
	d.getItem = d.db.Prepared(_itemSQL)
	d.getHeadTmp = d.db.Prepared(_headTmpSQL)
	d.getItemTmp = d.db.Prepared(_itemTmpSQL)
	var err error
	if d.showGrpc, err = v1.NewClient(c.ShowClient); err != nil {
		panic(err)
	}
	if d.tagGRPC, err = taggrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	return
}

// BeginTran begin a transacition
func (d *Dao) BeginTran(ctx context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(ctx)
}

// PTime get timing publis time.
func (d *Dao) PTime(ctx context.Context, now time.Time) (ps []int8, err error) {
	rows, err := d.getPTime.Query(ctx, now)
	if err != nil {
		return
	}
	defer rows.Close()
	var plat int8
	for rows.Next() {
		if err = rows.Scan(&plat); err != nil {
			return
		}
		ps = append(ps, plat)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// Pub check ptime and publish.
func (d *Dao) Pub(tx *sql.Tx, plat int8) (err error) {
	if _, err = tx.Exec(_delHdSQL, plat); err != nil {
		return
	}
	if _, err = tx.Exec(_delItSQL, plat); err != nil {
		return
	}
	if _, err = tx.Exec(_cpHdSQL, plat); err != nil {
		return
	}
	if _, err = tx.Exec(_cpItSQL, plat); err != nil {
		return
	}
	if _, err = tx.Exec(_upStSQL, plat); err != nil {
		return
	}
	return
}

func (d *Dao) PingDB(c context.Context) (err error) {
	return d.db.Ping(c)
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		_ = d.db.Close()
	}
	if d.redis != nil {
		_ = d.redis.Close()
	}
	if d.selectedRedis != nil {
		_ = d.selectedRedis.Close()
	}
}

func showActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_showRedisKeyPrefix)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}

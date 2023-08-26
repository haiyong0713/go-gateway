package record

// CREATE TABLE `free_record` (
//     `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
//     `ip_start` varchar(16) NOT NULL DEFAULT '' COMMENT 'IP段起始',
//     `ip_start_int` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'IP段起始int型',
//     `ip_end` varchar(16) NOT NULL DEFAULT '' COMMENT 'IP段结束',
//     `ip_end_int` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'IP段结束int型',
//     `isp` varchar(16) NOT NULL DEFAULT '' COMMENT '局数据备案所属运营商',
//     `is_bgp` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '局数据是否是BGP,0-否,1-是',
//     `business` varchar(16) NOT NULL DEFAULT '' COMMENT '业务',
//     `state` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'IP备案状态,0-备案中,1-备案成功,2-已下线',
//     `success_time`  timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '备案成功时间',
//     `cancel_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '备案下线时间',
//     `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//     `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
//     PRIMARY KEY (`id`),
//     KEY `ix_ip_start_int_ip_end_int` (`ip_start_int`,`ip_end_int`),
//     KEY `ix_mtime` (`mtime`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='免流局数据备案表';

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-free/admin/internal/model"
)

const (
	_getSQL     = "SELECT id,ip_start,ip_start_int,ip_end,ip_end_int,isp,is_bgp,business,state,success_time,cancel_time,ctime,mtime FROM free_record"
	_getSQLByIP = "SELECT id,ip_start,ip_start_int,ip_end,ip_end_int,isp,is_bgp,business,state,success_time,cancel_time,ctime,mtime FROM free_record WHERE %s"
	_insertSQL  = "INSERT INTO free_record (ip_start,ip_start_int,ip_end,ip_end_int,isp,is_bgp,business,state,success_time,cancel_time) VALUES %s"
	_deleteSQL  = "DELETE FROM free_record"
)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	InsertFreeRecords(ctx context.Context, rs []*model.FreeRecord) error
	AllFreeRecords(ctx context.Context) ([]*model.FreeRecord, error)
	FreeRecords(ctx context.Context, ips []int64) ([]*model.FreeRecord, error)
}

// dao dao.
type dao struct {
	db *sql.DB
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Close close the resource.
func (d *dao) Close() {
	d.db.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	err = d.db.Ping(ctx)
	return
}

// New new a dao and return.
func New() Dao {
	var (
		dc struct {
			Show *sql.Config
		}
	)
	checkErr(paladin.Get("mysql.toml").UnmarshalTOML(&dc))
	return &dao{
		// mysql
		db: sql.NewMySQL(dc.Show),
	}
}

func (d *dao) InsertFreeRecords(ctx context.Context, rs []*model.FreeRecord) error {
	const _count = 100
	if len(rs) == 0 {
		return nil
	}
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("%+v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%+v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%+v)", err)
		}
	}()
	if _, err = tx.Exec(_deleteSQL); err != nil {
		return err
	}
	var shard int
	if len(rs) < _count {
		shard = 1
	} else {
		shard = len(rs) / _count
		if len(rs)%(shard*_count) != 0 {
			shard++
		}
	}
	rss := make([][]*model.FreeRecord, shard)
	for i, aid := range rs {
		rss[i%shard] = append(rss[i%shard], aid)
	}
	for _, rs := range rss {
		var (
			sqls []string
			args []interface{}
		)
		for _, r := range rs {
			sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?)")
			args = append(args, r.IPStart, r.IPStartInt, r.IPEnd, r.IPEndInt, r.ISP, r.IsBGP, r.Business, r.State, r.SuccessTime, r.CancelTime)
		}
		if _, err = tx.Exec(fmt.Sprintf(_insertSQL, strings.Join(sqls, ",")), args...); err != nil {
			return err
		}
	}
	return nil
}

func (d *dao) AllFreeRecords(ctx context.Context) (res []*model.FreeRecord, err error) {
	rows, err := d.db.Query(ctx, _getSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.FreeRecord{}
		if err = rows.Scan(&r.ID, &r.IPStart, &r.IPStartInt, &r.IPEnd, &r.IPEndInt, &r.ISP, &r.IsBGP, &r.Business, &r.State, &r.SuccessTime, &r.CancelTime, &r.Ctime, &r.Mtime); err != nil {
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

func (d *dao) FreeRecords(ctx context.Context, ips []int64) ([]*model.FreeRecord, error) {
	if len(ips) == 0 {
		return nil, nil
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, ip := range ips {
		sqls = append(sqls, "(ip_start_int<=? AND ip_end_int>=?)")
		args = append(args, ip, ip)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_getSQLByIP, strings.Join(sqls, " OR ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rs []*model.FreeRecord
	for rows.Next() {
		r := &model.FreeRecord{}
		if err = rows.Scan(&r.ID, &r.IPStart, &r.IPStartInt, &r.IPEnd, &r.IPEndInt, &r.ISP, &r.IsBGP, &r.Business, &r.State, &r.SuccessTime, &r.CancelTime, &r.Ctime, &r.Mtime); err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rs, nil
}

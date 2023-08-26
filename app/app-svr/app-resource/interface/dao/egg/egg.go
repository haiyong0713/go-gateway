package egg

import (
	"context"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/static"
)

const (
	_eggSQL = `SELECT e.id,e.stime,e.etime,e.pre_time,e.mids,ep.plat,ep.conditions,ep.build,ep.url,ep.md5,ep.size,ep.mtime FROM egg AS e,egg_plat AS ep 
	WHERE e.id=ep.egg_id AND e.pre_time<? AND e.etime>? AND e.publish=1 AND e.delete=0 AND ep.deleted=0 AND e.type=1`
	_eggPicSQL = `SELECT e.id,e.stime,e.etime,e.pre_time,e.mids,ep.plat,ep.conditions,ep.build,epic.url,epic.md5,epic.size,epic.pic_type,epic.mtime FROM egg AS e,egg_plat AS ep,egg_pic AS epic 
	WHERE e.id=ep.egg_id AND e.id=epic.egg_id AND e.pre_time<? AND e.etime>? AND e.publish=1 AND e.delete=0 AND ep.deleted=0 AND epic.deleted=0 AND e.type=3`
)

const (
	_staticPic  = 1
	_dynamicPic = 2
)

type Dao struct {
	db *xsql.DB
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: component.GlobalDB,
	}
	return
}

// Egg select all egg
func (d *Dao) Egg(ctx context.Context, now time.Time) (res map[int8][]*static.Static, err error) {
	res = map[int8][]*static.Static{}
	rows, err := d.db.Query(ctx, _eggSQL, now, now)
	if err != nil {
		log.Error("mysqlDB.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		s := &static.Static{}
		if err = rows.Scan(&s.Sid, &s.Start, &s.End, &s.PreTime, &s.Mids, &s.Plat, &s.Condition, &s.Build, &s.URL, &s.Hash, &s.Size, &s.Mtime); err != nil {
			log.Error("egg rows.Scan error(%v)", err)
			return
		}
		if s.URL == "" {
			continue
		}
		//处理白名单
		if len(s.Mids) > 0 {
			wl := make(map[int64]int64)
			mids := strings.Split(s.Mids, ",")
			for _, m := range mids {
				mid, parseErr := strconv.ParseInt(m, 10, 64)
				if parseErr != nil {
					log.Error("egg whitelist parse s.Mids(%s) error(%+v)", s.Mids, parseErr)
					continue
				}
				wl[mid] = 1
			}
			s.Whitelist = wl
		}
		s.StaticChange("mov")
		res[s.Plat] = append(res[s.Plat], s)
	}
	err = rows.Err()
	return
}

// EggPic select all pic egg
func (d *Dao) EggPic(ctx context.Context, now time.Time) (res map[int8][]*static.Static, err error) {
	res = map[int8][]*static.Static{}
	rows, err := d.db.Query(ctx, _eggPicSQL, now, now)
	if err != nil {
		log.Error("mysqlDB.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var picType int
		s := &static.Static{}
		if err = rows.Scan(&s.Sid, &s.Start, &s.End, &s.PreTime, &s.Mids, &s.Plat, &s.Condition, &s.Build, &s.URL, &s.Hash, &s.Size, &picType, &s.Mtime); err != nil {
			log.Error("egg rows.Scan error(%v)", err)
			return
		}
		if s.URL == "" {
			continue
		}
		//处理白名单
		if len(s.Mids) > 0 {
			wl := make(map[int64]int64)
			mids := strings.Split(s.Mids, ",")
			for _, m := range mids {
				mid, parseErr := strconv.ParseInt(m, 10, 64)
				if parseErr != nil {
					log.Error("egg whitelist parse s.Mids(%s) error(%+v)", s.Mids, parseErr)
					continue
				}
				wl[mid] = 1
			}
			s.Whitelist = wl
		}
		var eggType string
		switch picType {
		case _staticPic:
			eggType = "static"
		case _dynamicPic:
			eggType = "dynamic"
		}
		s.StaticChange(eggType)
		res[s.Plat] = append(res[s.Plat], s)
	}
	err = rows.Err()
	return
}

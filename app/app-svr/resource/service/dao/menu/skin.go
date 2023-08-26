package menu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	api "go-gateway/app/app-svr/resource/service/api/v1"
)

const (
	_skinExtSQL   = "SELECT `id`,`skin_id`,`skin_name`,`attribute`,`state`,`ctime`,`mtime`,`stime`,`etime`,`user_scope_type`,`user_scope_value` FROM `skin_ext` WHERE `stime` <= ? AND `etime` >= ? AND `state` = 1"
	_skinLimitSQL = "SELECT `id`,`s_id`,`plat`,`build`,`conditions`,`state`,`ctime`,`mtime` FROM `skin_limit` WHERE `s_id` in (%s) AND `state` = 1"
)

// RawSkinExts .
func (d *Dao) RawSkinExts(c context.Context, now time.Time) (rly []*api.SkinExt, err error) {
	var rows *sql.Rows
	stime := now.Format("2006-01-02 15:04:05")
	if rows, err = d.db.Query(c, _skinExtSQL, stime, stime); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		m := &api.SkinExt{}
		if err = rows.Scan(&m.ID, &m.SkinID, &m.SkinName, &m.Attribute, &m.State, &m.Ctime, &m.Mtime, &m.Stime, &m.Etime, &m.UserScopeType, &m.UserScopeValue); err != nil {
			return
		}
		rly = append(rly, m)
	}
	err = rows.Err()
	return

}

// RawSkinLimits .
func (d *Dao) RawSkinLimits(c context.Context, sids []int64) (rly map[int64][]*api.SkinLimit, err error) {
	var (
		rows      *sql.Rows
		sqlString []string
		params    []interface{}
	)
	if len(sids) == 0 {
		return
	}
	for _, v := range sids {
		sqlString = append(sqlString, "?")
		params = append(params, v)
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_skinLimitSQL, strings.Join(sqlString, ",")), params...); err != nil {
		return
	}
	defer rows.Close()
	rly = make(map[int64][]*api.SkinLimit)
	for rows.Next() {
		m := &api.SkinLimit{}
		if err = rows.Scan(&m.ID, &m.SID, &m.Plat, &m.Build, &m.Conditions, &m.State, &m.Ctime, &m.Mtime); err != nil {
			return
		}
		rly[m.SID] = append(rly[m.SID], m)
	}
	err = rows.Err()
	return
}

// GetSkinInfoFromRedis 从Redis获取缓存的在线主题数据
func (d *Dao) GetSkinInfosFromRedis(ctx context.Context, key string) (skinInfo []*api.SkinInfo, err error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	skinInfoJSON, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			//log.Warn("resource-service.menuDao.GetSkinInfosFromRedis(%s) get empty body", key)
			err = nil
		} else {
			log.Error("resource-service.menuDao.GetSkinInfoFromRedis(%s) Error (%+v)", key, err)
		}
		return
	}
	if skinInfoJSON == "" {
		return
	}
	if err = json.Unmarshal([]byte(skinInfoJSON), &skinInfo); err != nil {
		log.Error("resource-service.menuDao.GetSkinInfoFromRedis(%s) json parsing Error (%+v)", key, err)
	}
	return
}

package bubble

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"

	bubblemdl "go-gateway/app/app-svr/app-feed/admin/model/bubble"
)

const (
	_seletSidebarSQL      = `SELECT id,plat,logo,logo_selected,name FROM sidebar WHERE module=9 AND state=1 AND lang_id=1 AND plat IN(0,1) ORDER BY id ASC`
	_seletSidebarLimitSQL = `SELECT id,conditions,build,s_id FROM sidebar_limit WHERE s_id IN (%s)`

	_bubbleSQL            = "SELECT id,position,icon,`desc`,url,stime,etime,operator,state,white_list FROM bubble WHERE id=?"
	_listSQL              = "SELECT id,position,icon,`desc`,url,stime,etime,operator,state,white_list FROM bubble ORDER BY id DESC LIMIT ? OFFSET ?"
	_clashSQL             = "SELECT id,position,icon,`desc`,url,stime,etime,operator,state,white_list FROM bubble WHERE ((stime>=? AND stime<=?) OR (stime<=? AND etime>=?) OR (etime>=? AND etime<=?)) AND state=1"
	_addBubbleSQL         = "INSERT INTO bubble (position,icon,`desc`,url,stime,etime,operator,state,white_list) VALUES (?,?,?,?,?,?,?,?,?)"
	_updateBubbleSQL      = "UPDATE bubble SET position=?,icon=?,`desc`=?,url=?,stime=?,etime=?,operator=?,white_list=? WHERE id=?"
	_updateBubbleStateSQL = `UPDATE bubble SET state=? WHERE id=?`
)

const (
	_bubbleKey = "bub_%d_%d"
)

func BubbleKey(bid, mid int64) (key string) {
	return fmt.Sprintf(_bubbleKey, bid, mid)
}

func (d *Dao) SetBubbleConfig(c context.Context, bid, mid int64, state int, expire int32) (err error) {
	var (
		key  = BubbleKey(bid, mid)
		conn = d.bubbleMc.Get(c)
	)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: state, Flags: memcache.FlagJSON, Expiration: expire}
	if err = conn.Set(item); err != nil {
		log.Error("SetBubbleConfig conn.Set() error(%v)", err)
	}
	return
}

func (d *Dao) Bubble(c context.Context, id int64) (re *bubblemdl.Bubble, err error) {
	row := d.db.QueryRow(c, _bubbleSQL, id)
	var position string
	re = &bubblemdl.Bubble{}
	if err = row.Scan(&re.ID, &position, &re.Icon, &re.Desc, &re.URL, &re.STime, &re.ETime, &re.Operator, &re.State, &re.WhiteList); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("Bubble %v", err)
		}
		return
	}
	if err = json.Unmarshal([]byte(position), &re.Position); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) List(c context.Context, pn, ps int) (res []*bubblemdl.Bubble, err error) {
	rows, err := d.db.Query(c, _listSQL, ps, (pn-1)*ps)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var position string
		re := &bubblemdl.Bubble{}
		if err = rows.Scan(&re.ID, &position, &re.Icon, &re.Desc, &re.URL, &re.STime, &re.ETime, &re.Operator, &re.State, &re.WhiteList); err != nil {
			log.Error("%v", err)
			return
		}
		if err = json.Unmarshal([]byte(position), &re.Position); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) Clash(c context.Context, stime, etime xtime.Time) (res []*bubblemdl.Bubble, err error) {
	rows, err := d.db.Query(c, _clashSQL, stime, etime, stime, etime, stime, etime)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var position string
		re := &bubblemdl.Bubble{}
		if err = rows.Scan(&re.ID, &position, &re.Icon, &re.Desc, &re.URL, &re.STime, &re.ETime, &re.Operator, &re.State, &re.WhiteList); err != nil {
			log.Error("%v", err)
			return
		}
		if err = json.Unmarshal([]byte(position), &re.Position); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxAddBubble(tx *sql.Tx, positionID, icon, desc, url string, stime, etime xtime.Time, whiteList, username string, state int) (row int64, err error) {
	res, err := tx.Exec(_addBubbleSQL, positionID, icon, desc, url, stime, etime, username, state, whiteList)
	if err != nil {
		log.Error("TxAddBubble %v", err)
		return
	}
	return res.LastInsertId()
}

func (d *Dao) TxUpdateBubble(tx *sql.Tx, id int64, positionID, icon, desc, url string, stime, etime xtime.Time, whiteList, username string) (row int64, err error) {
	res, err := tx.Exec(_updateBubbleSQL, positionID, icon, desc, url, stime, etime, username, whiteList, id)
	if err != nil {
		log.Error("TxUpdateBubble %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxUpdateBubbleState(tx *sql.Tx, id int64, state int) (row int64, err error) {
	res, err := tx.Exec(_updateBubbleStateSQL, state, id)
	if err != nil {
		log.Error("TxUpdateBubbleState %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) Siderbar(c context.Context) (res []*bubblemdl.Sidebar, err error) {
	rows, err := d.db.Query(c, _seletSidebarSQL)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bubblemdl.Sidebar{}
		if err = rows.Scan(&re.ID, &re.Plat, &re.Logo, &re.LogoSelected, &re.Name); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) SiderbarLimit(c context.Context, sids []int64) (res map[int64][]*bubblemdl.SidebarLimit, err error) {
	var (
		args []string
		sqls []interface{}
	)
	for _, sid := range sids {
		args = append(args, "?")
		sqls = append(sqls, sid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_seletSidebarLimitSQL, strings.Join(args, ",")), sqls...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64][]*bubblemdl.SidebarLimit)
	for rows.Next() {
		var sid int64
		re := &bubblemdl.SidebarLimit{}
		if err = rows.Scan(&re.ID, &re.Conditions, &re.Build, &sid); err != nil {
			log.Error("%v", err)
			return
		}
		res[sid] = append(res[sid], re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
	}
	return
}

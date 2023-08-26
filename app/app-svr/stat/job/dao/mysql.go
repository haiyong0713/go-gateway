package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_sharding   = 100
	_statSQL    = "SELECT aid,click,fav,share,reply,coin,dm,now_rank,his_rank,likes,dislike FROM archive_stat_%s WHERE aid=?"
	_updateStat = `UPDATE archive_stat_%s SET click=?,fav=?,share=?,reply=?,coin=?,dm=?,now_rank=?,his_rank=?,mtime=?,likes=?,follow=? WHERE aid=?`
	_insertStat = `INSERT INTO archive_stat_%s (aid,click,fav,share,reply,coin,dm,now_rank,his_rank,ctime,mtime,likes,follow) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`
)

func (d *Dao) hit(id int64) string {
	return fmt.Sprintf("%02d", id%_sharding)
}

// Stat returns stat info
func (d *Dao) Stat(c context.Context, aid int64) (stat *api.Stat, err error) {
	stat = &api.Stat{}
	err = d.db.QueryRow(c, fmt.Sprintf(_statSQL, d.hit(aid)), aid).Scan(&stat.Aid, &stat.View, &stat.Fav, &stat.Share, &stat.Reply, &stat.Coin, &stat.Danmaku, &stat.NowRank, &stat.HisRank, &stat.Like, &stat.DisLike)
	if err == sql.ErrNoRows {
		err = nil
		stat = nil
	} else if err != nil {
		log.Error("Stat(%v) error(%v)", aid, err)
	}
	return
}

// Update update stat's fields
func (d *Dao) Update(c context.Context, stat *api.Stat) (int64, error) {
	now := time.Now()
	rows, err := d.db.Exec(c, fmt.Sprintf(_updateStat, d.hit(stat.Aid)), stat.View, stat.Fav, stat.Share, stat.Reply, stat.Coin,
		stat.Danmaku, stat.NowRank, stat.HisRank, now, stat.Like, stat.Follow, stat.Aid)
	if err != nil {
		log.Error("UpdateStat(%d,%v) error(%+v)", stat.Aid, stat, err)
		return 0, err
	}
	return rows.RowsAffected()
}

func (d *Dao) Insert(c context.Context, stat *api.Stat) error {
	now := time.Now()
	_, err := d.db.Exec(c, fmt.Sprintf(_insertStat, d.hit(stat.Aid)), stat.Aid, stat.View, stat.Fav, stat.Share, stat.Reply, stat.Coin,
		stat.Danmaku, stat.NowRank, stat.HisRank, now, now, stat.Like, stat.Follow)
	if err != nil {
		log.Error("InsertStat(%d,%v) error(%+v)", stat.Aid, stat, err)
		return err
	}
	return nil
}

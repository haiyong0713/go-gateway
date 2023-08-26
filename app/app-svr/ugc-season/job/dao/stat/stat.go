package stat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_snStatSQL   = "SELECT season_id,click,fav,share,reply,coin,dm,likes,mtime FROM season_stat WHERE season_id=%d"
	_upSnStatSQL = `INSERT INTO season_stat (season_id,click,fav,share,reply,coin,dm,ctime,mtime,likes) VALUES (?,?,?,?,?,?,?,?,?,?)
				ON DUPLICATE KEY UPDATE click=?,fav=?,share=?,reply=?,coin=?,dm=?,mtime=?,likes=?`
	_upSnMStatSQL = `REPLACE INTO season_stat (season_id,click,fav,share,reply,coin,dm,now_rank,his_rank,mtime,likes) VALUES %s`
)

// SnStat returns stat info
func (d *Dao) SnStat(c context.Context, SID int64) (stat *api.Stat, err error) {
	stat = &api.Stat{}
	err = d.db.QueryRow(c, fmt.Sprintf(_snStatSQL, SID)).Scan(&stat.SeasonID, &stat.View, &stat.Fav, &stat.Share, &stat.Reply, &stat.Coin, &stat.Danmaku, &stat.Like, &stat.Mtime)
	if err == sql.ErrNoRows {
		err = nil
		stat = nil
	} else if err != nil {
		log.Error("SnStat(%v) error(%v)", SID, err)
	}
	return
}

// UpdateSnStat update stat's fields
func (d *Dao) UpdateSnStat(c context.Context, stat *api.Stat) (rows int64, err error) {
	now := time.Now()
	res, err := d.db.Exec(c, _upSnStatSQL, stat.SeasonID, stat.View, stat.Fav, stat.Share, stat.Reply, stat.Coin, stat.Danmaku, now, now, stat.Like,
		stat.View, stat.Fav, stat.Share, stat.Reply, stat.Coin, stat.Danmaku, now, stat.Like)
	if err != nil {
		log.Error("UpdateSnStat(%d,%v) error(%v)", stat.SeasonID, stat, err)
		return
	}
	rows, err = res.RowsAffected()
	return
}

// MultiSnUpdate update some stat's fields
func (d *Dao) MultiSnUpdate(c context.Context, stats ...*api.Stat) (rows int64, err error) {
	if len(stats) == 0 {
		return
	}
	const field = `(%d,%d,%d,%d,%d,%d,%d,%d,%d,'%s',%d)`
	var (
		fsqls = make([]string, 0, len(stats))
		now   = time.Now().Format("2006-01-02 15:04:05")
	)
	for _, stat := range stats {
		fsqls = append(fsqls, fmt.Sprintf(field, stat.SeasonID, stat.View, stat.Fav, stat.Share, stat.Reply, stat.Coin, stat.Danmaku, stat.NowRank, stat.HisRank, now, stat.Like))
	}
	res, err := d.db.Exec(c, fmt.Sprintf(_upSnMStatSQL, strings.Join(fsqls, ",")))
	if err != nil {
		log.Error("upMstat error(%v)", err)
		return
	}
	rows, err = res.RowsAffected()
	return
}

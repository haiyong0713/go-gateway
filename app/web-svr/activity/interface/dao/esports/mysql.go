package esports

import (
	"context"
	xsql "database/sql"
	"fmt"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/esports"
	"strings"

	"github.com/pkg/errors"
)

const (
	_insertFavSQL = "INSERT INTO act_esports_arena_fav(mid,first_fav_game_id,second_fav_game_id,third_fav_game_id) VALUES(?,?,?,?)"
	_getFavSQL    = "SELECT id,mid,first_fav_game_id,second_fav_game_id,third_fav_game_id,ctime,mtime FROM act_esports_arena_fav WHERE mid = ?"
)

// InsertEsportsArenaFav ...
func (d *Dao) InsertEsportsArenaFav(c context.Context, mid, fav1Id, fav2Id, fav3Id int64) (id int64, err error) {
	var res xsql.Result
	if res, err = component.GlobalDB.Exec(c, fmt.Sprintf(_insertFavSQL), mid, fav1Id, fav2Id, fav3Id); err != nil {

		if strings.Contains(err.Error(), "Duplicate entry") {
			return 0, ecode.ActivityTaskHasFinish
		}
		err = errors.Wrap(err, "InsertEsportsArenaFav:dao.db.Exec")
		return
	}
	return res.LastInsertId()
}

// ArchiveNumsDB 获取用户投稿数
func (d *Dao) EsportsArenaFavDB(c context.Context, mid int64) (favs []*esports.EsportsActFav, err error) {
	var rows *sql.Rows
	if rows, err = component.GlobalDB.Query(c, _getFavSQL, mid); err != nil {
		err = errors.Wrap(err, "EsportsArenaFavDB:dao.db.Query()")
		return
	}
	defer rows.Close()

	favs = make([]*esports.EsportsActFav, 0)
	for rows.Next() {
		l := &esports.EsportsActFav{}
		if err = rows.Scan(&l.ID, &l.Mid, &l.FirstFavGameId, &l.SecondFavGameId, &l.ThirdFavGameId, &l.Ctime, &l.Mtime); err != nil {
			err = errors.Wrap(err, "RawTaskList:rows.Scan()")
			return
		}
		favs = append(favs, l)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawTaskList:rows.Err()")
	}
	return
}

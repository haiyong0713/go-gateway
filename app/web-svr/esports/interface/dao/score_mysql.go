package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/esports/interface/model"

	"github.com/pkg/errors"
)

const sql4LolDataHero2 = `SELECT id,tournament_id,hero_id,hero_name,hero_image,appear_count,prohibit_count,victory_count,game_count FROM es_lol_data_hero2 WHERE tournament_id=?`

// FetchLolDataHero2 lol data hero2.
func (d *Dao) FetchLolDataHero2(c context.Context, leidaSID int64) (res []*model.LolDataHero2, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, sql4LolDataHero2, leidaSID); err != nil {
		err = errors.Wrapf(err, "FetchLolDataHero2:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LolDataHero2)
		if err = rows.Scan(&r.ID, &r.TournamentID, &r.HeroID, &r.HeroName, &r.HeroImage,
			&r.AppearCount, &r.ProhibitCount, &r.VictoryCount, &r.GameCount); err != nil {
			err = errors.Wrapf(err, "LolPlayers:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "FetchLolDataHero2:rows.Err() error")
	}
	return
}

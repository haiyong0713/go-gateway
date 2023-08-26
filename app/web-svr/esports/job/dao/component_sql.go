package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	mdlesp "go-gateway/app/web-svr/esports/job/model"
	innerSql "go-gateway/app/web-svr/esports/job/sql"
)

const sql4GoingSeasons = `SELECT id, leida_sid, serie_type, title, stime, etime,season_type FROM es_seasons WHERE stime<=? AND etime>=? limit 1000`

func GoingSeasons(ctx context.Context, before, after int64) (list []*mdlesp.Season, err error) {
	var rows *xsql.Rows
	rows, err = innerSql.GlobalDB.Query(ctx, sql4GoingSeasons, before, after)
	if err != nil {
		return
	}
	list = make([]*mdlesp.Season, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "GoingSeasons rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		season := new(mdlesp.Season)
		scanErr := rows.Scan(
			&season.ID,
			&season.LeidaSid,
			&season.SerieType,
			&season.Title,
			&season.Stime,
			&season.Etime,
			&season.SeasonType)
		if scanErr == nil {
			list = append(list, season)
		} else {
			log.Errorc(ctx, "contest component GoingSeasons scan error(%+v)", scanErr)
		}
	}
	return
}

const _sql4ContestStatus = `SELECT id FROM es_contests WHERE status=0 AND contest_status in (0,1,2) AND stime<=?`

func ContestIDsByTime(ctx context.Context, nowTime int64) (contestIDList []int64, err error) {
	var rows *xsql.Rows
	if rows, err = innerSql.GlobalDB.Query(ctx, _sql4ContestStatus, nowTime); err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "ContestStatusByTime rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		var contestID int64
		scanErr := rows.Scan(&contestID)
		if scanErr == nil {
			contestIDList = append(contestIDList, contestID)
		} else {
			log.Errorc(ctx, "ContestStatusByTime  error(%+v)", scanErr)
		}
	}
	return
}

const _sql4FreezeContestStatus = `SELECT id,stime,etime FROM es_contests WHERE contest_status in (0,1,2) AND stime<=?`

func GetContestListByTime(ctx context.Context, nowTime int64) (contestList []*mdlesp.Contest, err error) {
	var rows *xsql.Rows
	if rows, err = innerSql.GlobalDB.Query(ctx, _sql4FreezeContestStatus, nowTime); err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "GetContestListByTime rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		r := new(mdlesp.Contest)
		if err = rows.Scan(&r.ID, &r.Stime, &r.Etime); err != nil {
			log.Error("GetContestListByTime:row.Scan() error(%v)", err)
			return
		}
		contestList = append(contestList, r)
	}
	return
}

const _upContestStatusSQL = "Update es_contests set contest_status=? where id=?"

func UpContestStatus(c context.Context, contestStatus, contestID int64) (err error) {
	if _, err = innerSql.GlobalDB.Exec(c, _upContestStatusSQL, contestStatus, contestID); err != nil {
		log.Error("UpContestStatus:Exec() contestID(%d) error(%v)", contestID, err)
	}
	return
}

const _upContestStatusIngSQL = "Update es_contests set contest_status=1 where contest_status=0"

func UpContestStatusDoIng(c context.Context) (err error) {
	if _, err = innerSql.GlobalDB.Exec(c, _upContestStatusIngSQL); err != nil {
		log.Error("UpContestStatusDoIng:Exec() error(%v)", err)
	}
	return
}

const (
	sqlOfInsertUpdate4PlayerDataHero2 = `
INSERT INTO es_lol_data_hero2 (tournament_id, hero_id, hero_name
    , hero_image, appear_count, prohibit_count,victory_count,game_count)
VALUES (?, ?, ?, ?
	, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tournament_id = ?, hero_id = ?, hero_name = ?, hero_image = ?
, appear_count = ?, prohibit_count = ?, victory_count = ?, game_count = ?
`
)

func InsertUpdatePlayerDataHero2(ctx context.Context, serieID int64, hero *mdlesp.LolDataHero2Data) (err error) {
	if _, err = innerSql.GlobalDB.Exec(ctx, sqlOfInsertUpdate4PlayerDataHero2,
		serieID, hero.HeroID, hero.HeroName, hero.HeroImage, hero.AppearCount,
		hero.ProhibitCount, hero.VictoryCount, hero.GameCount,
		serieID, hero.HeroID, hero.HeroName, hero.HeroImage, hero.AppearCount,
		hero.ProhibitCount, hero.VictoryCount, hero.GameCount); err != nil {
		log.Errorc(ctx, "InsertUpdatePlayerDataHero2:Exec()  serieID(%+d) hero(%+v) error(%v)", serieID, hero, err)
	}
	return
}

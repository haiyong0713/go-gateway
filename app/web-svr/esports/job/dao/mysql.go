package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	mdlesp "go-gateway/app/web-svr/esports/job/model"
	innerSql "go-gateway/app/web-svr/esports/job/sql"

	"github.com/pkg/errors"
)

const (
	_empJson              = "[]"
	_autoSoure            = 1
	_arcType              = 5
	_contestsSQL          = "SELECT c.id,c.stime,c.live_room,c.home_id,c.away_id,c.success_team,c.special,c.special_name,c.special_tips,s.title,s.sub_title FROM `es_contests` as c INNER JOIN `es_seasons` as s ON c.sid=s.id  WHERE c.status = 0  AND  c.stime >= ? and c.stime < ? "
	_contestPushSQL       = "SELECT id,stime,etime  FROM `es_contests`  WHERE status=0 AND push_switch=1 AND active_push=0 AND stime>=? and stime<?"
	_contestPushHandSQL   = "SELECT id,stime,etime  FROM `es_contests`  WHERE id=?"
	_upContestPushSQL     = "Update es_contests set active_push=1 where id=?"
	_teamSQL              = "SELECT id,title,sub_title FROM `es_teams`  WHERE  is_deleted = 0  AND (id = ? or id = ?)"
	_teamAllSQL           = "SELECT id,title,sub_title FROM `es_teams`  WHERE  is_deleted = 0"
	_gameAllSQL           = "SELECT id,title,sub_title FROM `es_games`  WHERE  status = 0"
	_matchAllSQL          = "SELECT id,title,sub_title FROM `es_matchs`  WHERE status = 0"
	_gameTeamSQL          = "SELECT oid FROM es_gid_map WHERE type=4 AND is_deleted = 0 AND gid=?"
	_arcSQL               = "SELECT id,aid,score,is_deleted FROM `es_archives`  WHERE  is_deleted != 1  AND id > ? ORDER BY id ASC LIMIT ? "
	_arcEditSQL           = "UPDATE es_archives SET score = CASE %s END WHERE aid IN (%s)"
	_contPointSQL         = "SELECT id,match_id,stime,etime FROM es_contests  WHERE status = 0 AND  match_id > 0"
	_seriesSQL            = "SELECT distinct s.id,s.leida_sid,s.stime,s.etime,s.serie_type FROM  es_seasons as s inner join es_gid_map as g on s.id=g.oid WHERE leida_sid > 0 and g.type=2 and g.gid = ? ORDER BY etime"
	_lolPlayerSerieSQL    = "SELECT id,team_image,leida_team_image,image_url,leida_image,champions_image,leida_champions_image  FROM es_lol_player   WHERE leida_sid = ? AND player_id = ?"
	_lolTeamSerieSQL      = "SELECT id,image_url,leida_image  FROM es_lol_team     WHERE leida_sid = ?  AND team_id = ?"
	_dotaPlayerSerieSQL   = "SELECT id,team_image,leida_team_image,image_url,leida_image,heroes_image,leida_heroes_image  FROM es_dota_player  WHERE leida_sid = ? AND player_id = ?"
	_dotaTeamSerieSQL     = "SELECT id,image_url,leida_image  FROM es_dota_team    WHERE leida_sid = ? AND team_id = ?"
	_lolPlayersInsertSQL  = "INSERT INTO es_lol_player(player_id,team_id,team_acronym,team_image,leida_sid,name,image_url,champions_image,role,win,kda,kills,deaths,assists,minions_killed,wards_placed,games_count,leida_team_image,leida_image,leida_champions_image,position_id,position,mvp) VALUES %s"
	_dotaPlayersInsertSQL = "INSERT INTO es_dota_player(player_id,team_id,team_acronym,team_image,leida_sid,name,image_url,heroes_image,role,win,kda,kills,deaths,assists,wards_placed,last_hits,observer_wards_placed,sentry_wards_placed,xp_per_minute,gold_per_minute,games_count,leida_team_image,leida_image,leida_heroes_image) VALUES %s"
	_lolTeamsInsertSQL    = "INSERT INTO es_lol_team(team_id,leida_sid,name,acronym,image_url,win,kda,kills,deaths,assists,tower_kills,total_minions_killed,first_tower,first_inhibitor,first_dragon,first_baron,first_blood,wards_placed,inhibitor_kills,baron_kills,gold_earned,games_count,players,leida_image,baron_rate,dragon_rate,hits,lose_num,money,total_damage,win_num,image_thumb,new_data) VALUES %s"
	_dotaTeamsInsertSQL   = "INSERT INTO es_dota_team(team_id,leida_sid,name,acronym,image_url,win,kda,kills,deaths,assists,tower_kills,last_hits,observer_used,sentry_used,xp_per_minute,first_blood,heal,gold_spent,gold_per_min,denies,damage_taken,camps_stacked,games_count,players,leida_image) VALUES %s"
	_lolPlayersUpdateSQL  = "Update es_lol_player set team_id=?,team_acronym=?,team_image=?,name=?,image_url=?,role=?,win=?,kda=?,kills=?,deaths=?,assists=?,minions_killed=?,wards_placed=?,games_count=?,leida_team_image=?,leida_image=?,leida_champions_image=?,champions_image=?,position_id=?,position=?,mvp=? where player_id = ? and leida_sid = ?"
	_dotaPlayersUpdateSQL = "Update es_dota_player set team_id=?,team_acronym=?,team_image=?,name=?,image_url=?,role=?,win=?,kda=?,kills=?,deaths=?,assists=?,wards_placed=?,last_hits=?,observer_wards_placed=?,sentry_wards_placed=?,xp_per_minute=?,gold_per_minute=?,games_count=?,leida_team_image=?,leida_image=?,leida_heroes_image=?,heroes_image=? where player_id = ? and leida_sid = ?"
	_lolTeamUpdateSQL     = "Update es_lol_team set name=?,acronym=?,image_url=?,win=?,kda=?,kills=?,deaths=?,assists=?,tower_kills=?,total_minions_killed=?,first_tower=?,first_inhibitor=?,first_dragon=?,first_baron=?,first_blood=?,wards_placed=?,inhibitor_kills=?,baron_kills=?,gold_earned=?,games_count=?,leida_image=?,players=?,baron_rate=?,dragon_rate=?,hits=?,lose_num=?,money=?,total_damage=?,win_num=?,image_thumb=?,new_data=?  where team_id = ? and leida_sid = ?"
	_dotaTeamUpdateSQL    = "Update es_dota_team set name=?,acronym=?,image_url=?,win=?,kda=?,kills=?,deaths=?,assists=?,tower_kills=?,last_hits=?,observer_used=?,sentry_used=?,xp_per_minute=?,first_blood=?,heal=?,gold_spent=?,gold_per_min=?,denies=?,damage_taken=?,camps_stacked=?,games_count=?,leida_image=?,players=?  where team_id = ? and leida_sid = ?"
	_contestLeidaSQL      = "SELECT id,stime,etime,data_type,match_id,sid FROM `es_contests` WHERE match_id > 0 and status = 0"
	_lolGameSQL           = "SELECT game_id FROM `es_lol_game` where match_id = ?"
	_dotaGameSQL          = "SELECT game_id FROM `es_dota_game` where match_id = ?"
	_owGameSQL            = "SELECT game_id FROM `es_overwatch_game` where match_id = ?"
	_lolGameInsertSQL     = "INSERT INTO `es_lol_game` (game_id,match_id,teams,players,position,begin_at,end_at,finished)  VALUES %s"
	_dotaGameInsertSQL    = "INSERT INTO `es_dota_game` (game_id,match_id,teams,players,position,begin_at,end_at,finished)  VALUES %s"
	_owGameInsertSQL      = "INSERT INTO `es_overwatch_game` (game_id,match_id,win_team,teams,map,position,begin_at,end_at,finished)  VALUES %s"
	_lolGameUPdateSQL     = "UPDATE `es_lol_game`  set teams=?,players=?,position=?,begin_at=?,end_at=?,finished=?  where match_id = ? AND game_id = ?"
	_dotaGameUPdateSQL    = "UPDATE `es_dota_game`  set teams=?,players=?,position=?,begin_at=?,end_at=?,finished=?  where match_id = ? AND game_id = ?"
	_owGameUPdateSQL      = "UPDATE `es_overwatch_game`  set win_team=?,teams=?,map=?,position=?,begin_at=?,end_at=?,finished=?  where match_id = ? AND game_id = ?"
	_lolChamSQL           = "REPLACE INTO `es_lol_champion` (hero_id,name,image_url)  VALUES  %s"
	_dotaHeroSQL          = "REPLACE INTO `es_dota_hero` (hero_id,name,image_url)  VALUES  %s"
	_owHeroSQL            = "REPLACE INTO `es_overwatch_hero` (hero_id,name,image_url)  VALUES  %s"
	_lolItemSQL           = "REPLACE INTO `es_lol_item` (item_id,name,image_url)  VALUES  %s"
	_dotaItemSQL          = "REPLACE INTO `es_dota_item` (item_id,name,image_url)  VALUES  %s"
	_owMapSQL             = "REPLACE INTO `es_overwatch_map` (item_id,name,image_url)  VALUES  %s"
	_lolPlayerSQL         = "REPLACE INTO `es_lol_match_player` (player_id,name,image_url)  VALUES  %s"
	_dotaPlayerSQL        = "REPLACE INTO `es_dota_match_player` (player_id,name,image_url)  VALUES  %s"
	_owPlayerSQL          = "REPLACE INTO `es_overwatch_match_player` (player_id,name,image_url)  VALUES  %s"
	_lolAbilitySQL        = "REPLACE INTO `es_lol_ability` (ability_id,name,image_url)  VALUES  %s"
	_dotaAbilitySQL       = "REPLACE INTO `es_dota_ability` (ability_id,name,image_url)  VALUES  %s"
	_lolTeamsSQL          = "REPLACE INTO `es_lol_match_team` (team_id,name,image_url)  VALUES  %s"
	_dotaTeamsSQL         = "REPLACE INTO `es_dota_match_team` (team_id,name,image_url)  VALUES  %s"
	_owTeamsSQL           = "REPLACE INTO `es_overwatch_match_team` (team_id,name,image_url)  VALUES  %s"
	_autoWhiteSQL         = "SELECT mid,game_ids,match_ids FROM es_archive_whites WHERE is_deleted=0"
	_autoTagSQL           = "SELECT id,tag,game_ids,match_ids FROM es_archive_tags WHERE is_deleted=0"
	_autoKeywordSQL       = "SELECT id,keyword,game_ids,match_ids FROM es_archive_keywords WHERE is_deleted=0"
	_autoInsertArcSQL     = "INSERT INTO es_archives(aid,source,is_deleted) VALUES(?,?,?)"
	_autoInsertHitSQL     = "INSERT INTO es_archive_hits(arcs_id,white_mid,tag_ids,keyword_ids) VALUES(?,?,?,?)"
	_autoUpdateHitSQL     = "UPDATE es_archive_hits SET white_mid=?,tag_ids=?,keyword_ids=? WHERE arcs_id=?"
	_autoGameInsertSQL    = "INSERT INTO es_gid_map(type,oid,gid) VALUES %s"
	_autoGameDelSQL       = "UPDATE es_gid_map SET is_deleted=1 WHERE oid=? AND type=?"
	_autoMatchInsertSQL   = "INSERT INTO es_matchs_map(mid,aid) VALUES %s"
	_autoMatchDelSQL      = "UPDATE es_matchs_map SET is_deleted=1 WHERE aid=?"
	_autoTeamInsertSQL    = "INSERT INTO es_teams_map(tid,aid) VALUES %s"
	_autoTeamDelSQL       = "UPDATE es_teams_map SET is_deleted=1 WHERE aid=?"
	_autoYearInsertSQL    = "INSERT INTO es_year_map(year,aid) VALUES(?,?)"
	_autoYearDelSQL       = "UPDATE es_year_map SET is_deleted=1 WHERE aid=?"
	_autoTagsInsertSQL    = "INSERT INTO es_tags_map(tid,aid) VALUES(?,?)"
	_autoTagDelSQL        = "UPDATE es_tags_map SET is_deleted=1 WHERE aid=?"
	_autoArcSQL           = "SELECT id FROM `es_archives` WHERE aid=?"
	_autoArcPassSQL       = "UPDATE es_archives SET is_deleted=4 WHERE aid=?"
	_arcAutoPassSQL       = "SELECT id,aid,score,is_deleted FROM `es_archives` WHERE id > ? AND is_deleted = ? ORDER BY id ASC LIMIT ?"

	sql4FetchSeasonByID    = "SELECT id, stime, etime FROM es_seasons WHERE id = %v"
	sql4FetchSeasonTeamIDs = `
SELECT home_id, away_id
FROM es_contests
WHERE sid = ?
`
	sql4FetchAvCid = `
SELECT cid, av_cid
FROM es_contests_data
WHERE cid IN (%v) and is_deleted = 0 and av_cid != 0
`
	sql4FetchContestSeries = `
SELECT id, parent_title, child_title, score_id, start_time, end_time
FROM contest_series
WHERE id IN (%v) and is_deleted = 0
`

	sql4FetchContestsOfAll = `
SELECT id, UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(stime), '%Y-%m-%d')) AS date_unix
	, stime, etime, game_stage, live_room, playback
	, collection_url, home_id, home_score, away_id, away_score, match_id, series_id
FROM es_contests
WHERE sid = ? AND status = 0
ORDER BY stime ASC
`
	sql4FetchContestsFromNowDayOn = `
SELECT id, DATE_FORMAT(FROM_UNIXTIME(stime), '%Y-%m-%d') AS date_str
	, stime, etime, game_stage, live_room, playback
	, collection_url, home_id, home_score, away_id, away_score, match_id, series_id
FROM es_contests
WHERE sid = ? AND stime > ? AND status = 0
ORDER BY stime ASC
`
	sql4FetchTeamsByIDs = `
SELECT id, title, sub_title, logo, region_id, leida_tid
FROM es_teams
WHERE id IN (%v) and is_deleted = 0
`

	sql4FetchPostersOfAll = `
SELECT bg_image, contest_id, is_centeral
FROM match_poster
WHERE online_status = 1 and is_deprecated = 0 and contest_id > 0
ORDER BY position_order asc
`

	sqlLimitOf1              = 1
	sqlLimitOf100            = 100
	sqlOfSeasonQuery         = "SELECT title, stime, etime FROM es_seasons WHERE id = %v limit 1"
	sqlOfInPlayContestsQuery = "SELECT away_id, home_id FROM es_contests WHERE sid = %v AND stime <= UNIX_TIMESTAMP() AND etime >= UNIX_TIMESTAMP() ORDER BY id DESC LIMIT %v, %v"

	_sqlContestById = "SELECT c.id,c.stime,c.live_room,c.home_id,c.away_id,c.success_team,c.special,c.special_name,c.special_tips,s.title,s.sub_title,c.contest_status,s.message_senduid FROM `es_contests` as c INNER JOIN `es_seasons` as s ON c.sid=s.id  WHERE c.status = 0  AND  c.id = ?"
)

func (d *Dao) InPlaySeasonByID(ctx context.Context, id int64) (season mdlesp.Season, err error) {
	err = d.db.QueryRow(ctx, fmt.Sprintf(sqlOfSeasonQuery, id)).Scan(&season.Title, &season.Stime, &season.Etime)
	if err != nil {
		return
	}

	season.ID = id

	return
}

func (d *Dao) HasInPlayContest(ctx context.Context, seasonID int64) (has bool) {
	contest := mdlesp.Contest{}
	err := d.db.QueryRow(ctx, fmt.Sprintf(sqlOfInPlayContestsQuery, seasonID, 0, sqlLimitOf1)).Scan(&contest.AwayID, &contest.HomeID)
	if err == nil {
		has = true
	}

	return
}

// SeriesSeason serie ids season  list.
func (d *Dao) SeriesSeason(c context.Context, gameID int) (res []*mdlesp.Season, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _seriesSQL, gameID); err != nil {
		err = errors.Wrapf(err, "SeriesSeason:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Season)
		if err = rows.Scan(&r.ID, &r.LeidaSid, &r.Stime, &r.Etime, &r.SerieType); err != nil {
			err = errors.Wrapf(err, "SeriesSeason:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rows.Err() error")
	}
	return
}

// LolPlayerSerie lol player  exists.
func (d *Dao) LolPlayerSerie(c context.Context, serieID, id int64) (res *mdlesp.LolPlayer, err error) {
	res = &mdlesp.LolPlayer{}
	row := d.db.QueryRow(c, _lolPlayerSerieSQL, serieID, id)
	if err = row.Scan(&res.ID, &res.TeamImage, &res.LeidaTeamImage, &res.ImageURL, &res.LeidaImage, &res.ChampionsImage, &res.LeidaChampionsImage); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("LolPlayerSerie:row.Scan error(%v)", err)
		}
	}
	return
}

// LolTeamSerie lol team exists.
func (d *Dao) LolTeamSerie(c context.Context, serieID, id int64) (res *mdlesp.LolTeam, err error) {
	res = &mdlesp.LolTeam{}
	row := d.db.QueryRow(c, _lolTeamSerieSQL, serieID, id)
	if err = row.Scan(&res.ID, &res.ImageURL, &res.LeidaImage); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("LolTeamSerie:row.Scan error(%v)", err)
		}
	}
	return
}

// DotaPlayerSerie lol player exists.
func (d *Dao) DotaPlayerSerie(c context.Context, serieID, id int64) (res *mdlesp.DotaPlayer, err error) {
	res = &mdlesp.DotaPlayer{}
	row := d.db.QueryRow(c, _dotaPlayerSerieSQL, serieID, id)
	if err = row.Scan(&res.ID, &res.TeamImage, &res.LeidaTeamImage, &res.ImageURL, &res.LeidaImage, &res.HeroesImage, &res.LeidaHeroesImage); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("DotaPlayerSerie:row.Scan error(%v)", err)
		}
	}
	return
}

// DotaTeamSerie lol team exists.
func (d *Dao) DotaTeamSerie(c context.Context, serieID, id int64) (res *mdlesp.DotaTeam, err error) {
	res = &mdlesp.DotaTeam{}
	row := d.db.QueryRow(c, _dotaTeamSerieSQL, serieID, id)
	if err = row.Scan(&res.ID, &res.ImageURL, &res.LeidaImage); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("DotaTeamSerie:row.Scan error(%v)", err)
		}
	}
	return
}

// AddLolPlayer add lol player to mysql.
func (d *Dao) AddLolPlayer(c context.Context, data []*mdlesp.LolPlayer) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%s','%s',%d,'%s','%s','%s','%s',%f,%f,%f,%f,%f,%f,%f,%d,'%s','%s','%s','%d','%s','%d')",
			v.PlayerID, v.TeamID, v.TeamAcronym, v.TeamImage, v.LeidaSID, v.Name, v.ImageURL, v.ChampionsImage, v.Role, v.Win, v.KDA, v.Kills, v.Deaths,
			v.Assists, v.MinionsKilled, v.WardsPlaced, v.GamesCount, v.LeidaTeamImage, v.LeidaImage, v.LeidaChampionsImage,
			v.PositionID, v.Position, v.MVP))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolPlayersInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolPlayer db.Exec error")
	}
	return
}

// UpLolPlayer update lol player.
func (d *Dao) UpLolPlayer(c context.Context, data *mdlesp.LolPlayer) (err error) {
	if _, err = d.db.Exec(c, _lolPlayersUpdateSQL, data.TeamID, data.TeamAcronym, data.TeamImage, data.Name, data.ImageURL, data.Role, data.Win, data.KDA, data.Kills, data.Deaths, data.Assists, data.MinionsKilled,
		data.WardsPlaced, data.GamesCount, data.LeidaTeamImage, data.LeidaImage, data.LeidaChampionsImage, data.ChampionsImage, data.PositionID, data.Position, data.MVP, data.PlayerID, data.LeidaSID); err != nil {
		err = errors.Wrapf(err, "UpLolPlayer db.Exec error")
	}
	return
}

// AddLolTeam add lol team to mysql.
func (d *Dao) AddLolTeam(c context.Context, data []*mdlesp.LolTeam) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%s','%s','%s',%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%d,'%s','%s',%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,'%s',%d)",
			v.TeamID, v.LeidaSID, v.Name, v.Acronym, v.ImageURL, v.Win, v.KDA, v.Kills, v.Deaths, v.Assists, v.TowerKills, v.TotalMinionsKilled, v.FirstTower,
			v.FirstInhibitor, v.FirstDragon, v.FirstBaron, v.FirstBlood, v.WardsPlaced, v.InhibitorKills, v.BaronKills, v.GoldEarned, v.GamesCount, v.Players, v.LeidaImage,
			v.BaronRate, v.DragonRate, v.Hits, v.LoseNum, v.Money, v.TotalDamage, v.WinNum, v.ImageThumb, v.NewData))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolTeamsInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolTeam db.Exec error")
	}
	return
}

// AddDotaTeam add dota team to mysql.
func (d *Dao) AddDotaTeam(c context.Context, data []mdlesp.DotaTeam) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%s','%s','%s',%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%d,'%s','%s')",
			v.TeamID, v.LeidaSID, v.Name, v.Acronym, v.ImageURL, v.Win, v.KDA, v.Kills, v.Deaths, v.Assists, v.TowerKills, v.LastHits, v.ObserverUsed,
			v.SentryUsed, v.XpPerMinute, v.FirstBlood, v.Heal, v.GoldSpent, v.GoldPerMin, v.Denies, v.DamageTaken, v.CampsStacked, v.GamesCount, v.Players, v.LeidaImage))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaTeamsInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotaTeam db.Exec error")
	}
	return
}

// UpLolTeam update lol team.
func (d *Dao) UpLolTeam(c context.Context, data *mdlesp.LolTeam) (err error) {
	if _, err = d.db.Exec(c, _lolTeamUpdateSQL, data.Name, data.Acronym, data.ImageURL, data.Win, data.KDA, data.Kills, data.Deaths, data.Assists, data.TowerKills, data.TotalMinionsKilled, data.FirstTower,
		data.FirstInhibitor, data.FirstDragon, data.FirstBaron, data.FirstBlood, data.WardsPlaced, data.InhibitorKills, data.BaronKills, data.GoldEarned, data.GamesCount, data.LeidaImage, data.Players,
		data.BaronRate, data.DragonRate, data.Hits, data.LoseNum, data.Money, data.TotalDamage, data.WinNum, data.ImageThumb, data.NewData, data.TeamID, data.LeidaSID); err != nil {
		err = errors.Wrapf(err, "UpLolTeam db.Exec error")
	}
	return
}

// UpDotaTeam update dota team.
func (d *Dao) UpDotaTeam(c context.Context, data mdlesp.DotaTeam) (err error) {
	if _, err = d.db.Exec(c, _dotaTeamUpdateSQL, data.Name, data.Acronym, data.ImageURL, data.Win, data.KDA, data.Kills, data.Deaths, data.Assists, data.TowerKills,
		data.LastHits, data.ObserverUsed, data.SentryUsed, data.XpPerMinute, data.FirstBlood, data.Heal, data.GoldSpent, data.GoldPerMin,
		data.Denies, data.DamageTaken, data.CampsStacked, data.GamesCount, data.LeidaImage, data.Players, data.TeamID, data.LeidaSID); err != nil {
		err = errors.Wrapf(err, "UpDotaTeam db.Exec error")
	}
	return
}

// AddDotaPlayer add dota player to mysql.
func (d *Dao) AddDotaPlayer(c context.Context, data []mdlesp.DotaPlayer) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%s','%s',%d,'%s','%s','%s','%s',%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%.1f,%d,'%s','%s','%s')",
			v.PlayerID, v.TeamID, v.TeamAcronym, v.TeamImage, v.LeidaSID, v.Name, v.ImageURL, v.HeroesImage, v.Role, v.Win, v.KDA, v.Kills, v.Deaths, v.Assists, v.WardsPlaced, v.LastHits, v.ObserverWardsPlaced, v.SentryWardsPlaced,
			v.XpPerMinute, v.GoldPerMinute, v.GamesCount, v.LeidaTeamImage, v.LeidaImage, v.LeidaHeroesImage))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaPlayersInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotaPlayer db.Exec error")
	}
	return
}

// UpDotaPlayer update dota player.
func (d *Dao) UpDotaPlayer(c context.Context, data mdlesp.DotaPlayer) (err error) {
	if _, err = d.db.Exec(c, _dotaPlayersUpdateSQL, data.TeamID, data.TeamAcronym, data.TeamImage, data.Name, data.ImageURL, data.Role, data.Win, data.KDA, data.Kills, data.Deaths, data.Assists, data.WardsPlaced, data.LastHits, data.ObserverWardsPlaced, data.SentryWardsPlaced,
		data.XpPerMinute, data.GoldPerMinute, data.GamesCount, data.LeidaTeamImage, data.LeidaImage, data.LeidaHeroesImage, data.HeroesImage, data.PlayerID, data.LeidaSID); err != nil {
		err = errors.Wrapf(err, "UpDotaPlayer db.Exec error")
	}
	return
}

func SeasonByID(ctx context.Context, seasonID int64) (season *mdlesp.Season, err error) {
	season = new(mdlesp.Season)
	err = innerSql.GlobalDB.QueryRow(
		ctx,
		fmt.Sprintf(sql4FetchSeasonByID, seasonID)).Scan(
		&season.ID, &season.Stime, &season.Etime)

	return
}

func SeasonTeamIDList(ctx context.Context, seasonID int64) (list []int64, err error) {
	var rows *xsql.Rows
	rows, err = innerSql.GlobalDB.Query(ctx, sql4FetchSeasonTeamIDs, seasonID)
	if err != nil {
		return
	}

	var homeID, awayID int64
	m := make(map[int64]int64, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "SeasonTeamIDList rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		scanErr := rows.Scan(&homeID, &awayID)
		if scanErr == nil && homeID != 0 && awayID != 0 {
			m[homeID] = homeID
			m[awayID] = awayID
		}
	}
	for k := range m {
		list = append(list, k)
	}

	return
}

func AvCIDMap(ctx context.Context, contestIDList []int64) (m map[int64]int64, err error) {
	var rows *xsql.Rows
	rows, err = innerSql.GlobalDB.Query(ctx, fmt.Sprintf(sql4FetchAvCid, xstr.JoinInts(contestIDList)))
	if err != nil {
		return
	}

	var cid, avCID int64
	m = make(map[int64]int64, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "AvCIDMap rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		scanErr := rows.Scan(&cid, &avCID)
		if scanErr == nil && cid != 0 && avCID != 0 {
			m[avCID] = cid
		}
	}

	return
}

func FetchContestSeriesList(ctx context.Context, idList []int64) (list map[int64]*mdlesp.ContestSeries, err error) {
	var rows *xsql.Rows
	list = make(map[int64]*mdlesp.ContestSeries, 0)
	rows, err = innerSql.GlobalDB.Query(ctx, fmt.Sprintf(sql4FetchContestSeries, xstr.JoinInts(idList)))
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "FetchContestSeriesList rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		series := new(mdlesp.ContestSeries)
		scanErr := rows.Scan(
			&series.ID,
			&series.ParentTitle,
			&series.ChildTitle,
			&series.ScoreID,
			&series.StartTime,
			&series.EndTime)
		if scanErr == nil {
			list[series.ID] = series
		} else {
			fmt.Println("FetchContestSeriesList scan err: ", err)
		}
	}
	return
}

func S10Poster(ctx context.Context) (list []*mdlesp.Poster4S10, err error) {
	var rows *xsql.Rows
	rows, err = innerSql.GlobalDB.Query(ctx, sql4FetchPostersOfAll)
	if err != nil {
		return
	}

	list = make([]*mdlesp.Poster4S10, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "S10Poster rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		poster := new(mdlesp.Poster4S10)
		err = rows.Scan(
			&poster.BackGround,
			&poster.ContestID,
			&poster.InCenter)
		if err == nil {
			list = append(list, poster)
		}
	}
	return
}

func FetchContestsBySeasonID(ctx context.Context, seasonID int64, fetchAll bool) (list []*mdlesp.Contest2Tab, err error) {
	var rows *xsql.Rows
	if fetchAll {
		rows, err = innerSql.GlobalDB.Query(ctx, sql4FetchContestsOfAll, seasonID)
	} else {
		year, month, day := time.Now().Date()
		dayUnix := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
		rows, err = innerSql.GlobalDB.Query(ctx, sql4FetchContestsFromNowDayOn, seasonID, dayUnix)
	}

	if err != nil {
		return
	}

	list = make([]*mdlesp.Contest2Tab, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "FetchContestsBySeasonID rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		contest := new(mdlesp.Contest2Tab)
		err = rows.Scan(
			&contest.ID,
			&contest.StimeDate,
			&contest.Stime,
			&contest.Etime,
			&contest.GameStage,
			&contest.LiveRoom,
			&contest.PlayBack,
			&contest.CollectionUrl,
			&contest.HomeID,
			&contest.HomeScore,
			&contest.AwayID,
			&contest.AwayScore,
			&contest.MatchID,
			&contest.SeriesID)
		if err == nil {
			list = append(list, contest)
		}
	}
	return
}

func FetchTeamsByIDs(ctx context.Context, teamIDs []int64) (list []*mdlesp.Team2Tab, err error) {
	var rows *xsql.Rows
	rows, err = innerSql.GlobalDB.Query(ctx, fmt.Sprintf(sql4FetchTeamsByIDs, xstr.JoinInts(teamIDs)))
	if err != nil {
		return
	}

	list = make([]*mdlesp.Team2Tab, 0)
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "FetchTeamsByIDs rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		team := new(mdlesp.Team2Tab)
		err = rows.Scan(&team.ID, &team.Title, &team.SubTitle, &team.Logo, &team.RegionID, &team.ScoreTeamID)
		if err == nil {
			list = append(list, team)
		} else {
			fmt.Println("FetchTeamsByIDs occur err: ", err)
		}
	}
	return
}

// Contests  contests by time.
func (d *Dao) Contests(c context.Context, stime, etime int64) (res []*mdlesp.Contest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _contestsSQL, stime, etime); err != nil {
		log.Error("Contests:d.db.Query(%d) error(%v)", stime, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Contest)
		if err = rows.Scan(&r.ID, &r.Stime, &r.LiveRoom, &r.HomeID, &r.AwayID, &r.SuccessTeam, &r.Special, &r.SpecialName, &r.SpecialTips, &r.SeasonTitle, &r.SeasonSubTitle); err != nil {
			log.Error("Contests:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// ContestsPush  contests by time.
func (d *Dao) ContestsPush(c context.Context, stime, etime int64) (res []*mdlesp.Contest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _contestPushSQL, stime, etime); err != nil {
		log.Error("ContestsPush:d.db.Query(%d) error(%v)", stime, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Contest)
		if err = rows.Scan(&r.ID, &r.Stime, &r.Etime); err != nil {
			log.Error("ContestsPush:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// UpContestPush update push active event.
func (d *Dao) UpContestPush(c context.Context, contestID int64) (err error) {
	if _, err = d.db.Exec(c, _upContestPushSQL, contestID); err != nil {
		err = errors.Wrapf(err, "UpContestPush db.Exec error")
	}
	return
}

// Teams  teams by id.
func (d *Dao) Teams(c context.Context, homeID, awayID int64) (res []*mdlesp.Team, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _teamSQL, homeID, awayID); err != nil {
		log.Error("Teams:d.db.Query homeID(%d) awayID(%d) error(%v)", homeID, awayID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Team)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle); err != nil {
			log.Error("Teams:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// AllTeams .
func (d *Dao) AllTeams(c context.Context) (res map[int64]*mdlesp.Team, err error) {
	var (
		rows *xsql.Rows
	)
	res = make(map[int64]*mdlesp.Team)
	if rows, err = d.db.Query(c, _teamAllSQL); err != nil {
		log.Error("AllTeams:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Team)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle); err != nil {
			log.Error("AllTeams:row.Scan() error(%v)", err)
			return
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("AllTeams rows.Err() error(%v)", err)
	}
	return
}

// AllGames .
func (d *Dao) AllGames(c context.Context) (res map[string]*mdlesp.BaseInfo, err error) {
	var (
		rows *xsql.Rows
	)
	res = make(map[string]*mdlesp.BaseInfo)
	if rows, err = d.db.Query(c, _gameAllSQL); err != nil {
		log.Error("AllGames:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.BaseInfo)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle); err != nil {
			log.Error("AllGames:row.Scan() error(%v)", err)
			return
		}
		res[strings.ToLower(r.Title)] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("AllGames rows.Err() error(%v)", err)
	}
	return
}

// AllMatchs .
func (d *Dao) AllMatchs(c context.Context) (res map[int64]*mdlesp.BaseInfo, err error) {
	var (
		rows *xsql.Rows
	)
	res = make(map[int64]*mdlesp.BaseInfo)
	if rows, err = d.db.Query(c, _matchAllSQL); err != nil {
		log.Error("AllMatchs:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.BaseInfo)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle); err != nil {
			log.Error("AllMatchs:row.Scan() error(%v)", err)
			return
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("AllMatchs rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) TeamsByGame(ctx context.Context, gid int64) (res []int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(ctx, _gameTeamSQL, gid); err != nil {
		log.Error("TeamsByGame: db.Exec(%d) error(%v)", gid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.PointsAward)
		if err = rows.Scan(&r.ID); err != nil {
			log.Error("TeamsByGame:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r.ID)
	}
	err = rows.Err()
	return
}

// Arcs archives by ids.
func (d *Dao) Arcs(c context.Context, id int64, limit int) (res []*mdlesp.Arc, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _arcSQL, id, limit); err != nil {
		log.Error("Arcs:d.db.Query id(%d) limit(%d) error(%v)", id, limit, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Arc)
		if err = rows.Scan(&r.ID, &r.Aid, &r.Score, &r.IsDeleted); err != nil {
			log.Error("Arcs:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Arcs archives auto.
func (d *Dao) ArcsAuto(c context.Context, id, checkTp int64, limit int) (res []*mdlesp.Arc, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _arcAutoPassSQL, id, checkTp, limit); err != nil {
		log.Error("ArcsAuto:d.db.Query id(%d) checkTp(%d) limit(%d) error(%v)", id, checkTp, limit, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Arc)
		if err = rows.Scan(&r.ID, &r.Aid, &r.Score, &r.IsDeleted); err != nil {
			log.Error("ArcsAuto:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("ArcsAuto rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) UpdateScoreByArchiveMap(ctx context.Context, archiveMap map[int64]int64) {
	caseStr := ""
	index := 0
	archiveAIDs := make([]int64, 0)

	for k, v := range archiveMap {
		if index == 1000 {
			execSql := fmt.Sprintf(_arcEditSQL, caseStr, xstr.JoinInts(archiveAIDs))
			_, _ = d.db.Exec(ctx, execSql)

			index = 0
			caseStr = ""
			archiveAIDs = archiveAIDs[0:0]
		}

		caseStr = fmt.Sprintf("%s WHEN aid = %d THEN %d", caseStr, k, v)
		archiveAIDs = append(archiveAIDs, k)
		index++
	}

	if len(archiveAIDs) > 0 {
		execSql := fmt.Sprintf(_arcEditSQL, caseStr, xstr.JoinInts(archiveAIDs))
		_, _ = d.db.Exec(ctx, execSql)
	}
}

// UpArcScore  update  archive score.
func (d *Dao) UpArcScore(c context.Context, partArcs []*mdlesp.Arc, arcs map[int64]*arcmdl.Arc) (err error) {
	var (
		caseStr string
		aids    []int64
		score   int64
	)
	for _, v := range partArcs {
		if arc, ok := arcs[v.Aid]; ok {
			score = d.score(arc)
		} else {
			continue
		}
		caseStr = fmt.Sprintf("%s WHEN aid = %d THEN %d", caseStr, v.Aid, score)
		aids = append(aids, v.Aid)
	}
	if len(aids) == 0 {
		return
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_arcEditSQL, caseStr, xstr.JoinInts(aids))); err != nil {
		err = errors.Wrapf(err, "UpArcScore  d.db.Exec")
	}
	return
}

func (d *Dao) score(arc *arcmdl.Arc) (res int64) {
	tmpRs := float64(arc.Stat.Coin)*d.c.Rule.CoinPercent +
		float64(arc.Stat.Fav)*d.c.Rule.FavPercent + float64(arc.Stat.Danmaku)*d.c.Rule.DmPercent +
		float64(arc.Stat.Reply)*d.c.Rule.ReplyPercent + float64(arc.Stat.View)*d.c.Rule.ViewPercent +
		float64(arc.Stat.Like)*d.c.Rule.LikePercent + float64(arc.Stat.Share)*d.c.Rule.SharePercent
	now := time.Now()
	hours := now.Sub(arc.PubDate.Time()).Hours()
	if hours/24 <= d.c.Rule.NewDay {
		tmpRs = tmpRs * 1.5
	}
	decimal, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", tmpRs), 64)
	res = int64(decimal * 100)
	return
}

// ContPoints  contests point data by time.
func (d *Dao) ContPoints(c context.Context) (res []*mdlesp.ContestData, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _contPointSQL); err != nil {
		log.Error("ContPoints:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.ContestData)
		if err = rows.Scan(&r.CID, &r.MatchID, &r.Stime, &r.Etime); err != nil {
			log.Error("ContPoints:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// ContestDatas contest datas.
func (d *Dao) ContestDatas(c context.Context) (res []*mdlesp.Contest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _contestLeidaSQL); err != nil {
		log.Error("Contests: db.Exec(%s) error(%v)", _contestLeidaSQL, err)
		return
	}
	defer rows.Close()
	res = make([]*mdlesp.Contest, 0)
	for rows.Next() {
		r := new(mdlesp.Contest)
		if err = rows.Scan(&r.ID, &r.Stime, &r.Etime, &r.DataType, &r.MatchID, &r.SeasonID); err != nil {
			log.Error("Contests:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// LolGames lol games  list.
func (d *Dao) LolGames(c context.Context, matchID int64) (res []*mdlesp.Oid, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _lolGameSQL, matchID); err != nil {
		err = errors.Wrapf(err, "LolGames:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Oid)
		if err = rows.Scan(&r.ID); err != nil {
			err = errors.Wrapf(err, "LolGames:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rows.Err() error")
	}
	return
}

// DotaGames dota games  list.
func (d *Dao) DotaGames(c context.Context, matchID int64) (res []*mdlesp.Oid, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _dotaGameSQL, matchID); err != nil {
		err = errors.Wrapf(err, "DotaGames:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Oid)
		if err = rows.Scan(&r.ID); err != nil {
			err = errors.Wrapf(err, "DotaGames:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rows.Err() error")
	}
	return
}

// OwGames overwatch games  list.
func (d *Dao) OwGames(c context.Context, matchID int64) (res []*mdlesp.Oid, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _owGameSQL, matchID); err != nil {
		err = errors.Wrapf(err, "OwGames:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.Oid)
		if err = rows.Scan(&r.ID); err != nil {
			err = errors.Wrapf(err, "OwGames:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rows.Err() error")
	}
	return
}

// AddLolGame add lol game to mysql.
func (d *Dao) AddLolGame(c context.Context, data []*mdlesp.LolGame) (err error) {
	var inTeams, inPlayers string
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		if inTeams, err = d.escapeStr(v.Teams, true); err != nil {
			err = errors.Wrapf(err, "AddLolGame  inTeams json.Marshal error")
			return
		}
		if inPlayers, err = d.escapeStr(v.Players, true); err != nil {
			err = errors.Wrapf(err, "AddLolGame  inPlayers json.Marshal error")
			return
		}
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%s','%s',%d,'%s','%s',%t)",
			v.ID, v.MatchID, inTeams, inPlayers, v.Position, v.BeginAt, v.EndAt, v.Finished))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolGameInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolGame db.Exec error")
	}
	return
}

// AddDotaGame add dota game to mysql.
func (d *Dao) AddDotaGame(c context.Context, data []*mdlesp.DotaGame) (err error) {
	var inTeams, inPlayers string
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		if inTeams, err = d.escapeStr(v.Teams, true); err != nil {
			err = errors.Wrapf(err, "AddDotaGame  inTeams json.Marshal error")
			return
		}
		if inPlayers, err = d.escapeStr(v.Players, true); err != nil {
			err = errors.Wrapf(err, "AddDotaGame  inPlayers json.Marshal error")
			return
		}
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%s','%s',%d,'%s','%s',%t)",
			v.ID, v.MatchID, inTeams, inPlayers, v.Position, v.BeginAt, v.EndAt, v.Finished))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaGameInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotaGame db.Exec error")
	}
	return
}

// AddOwGame add overwatch game to mysql.
func (d *Dao) AddOwGame(c context.Context, data []*mdlesp.OwGame) (err error) {
	var inTeams, inMap string
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		if inTeams, err = d.escapeStr(v.Teams, true); err != nil {
			err = errors.Wrapf(err, "AddOwGame  inRounds json.Marshal error")
			return
		}
		if inMap, err = d.escapeStr(v.Map, true); err != nil {
			err = errors.Wrapf(err, "AddOwGame  inMap json.Marshal error")
			return
		}
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,'%d','%s','%s',%d,'%s','%s',%t)",
			v.ID, v.MatchID, v.WinTeam, inTeams, inMap, v.Position, v.BeginAt, v.EndAt, v.Finished))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_owGameInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddOwGame db.Exec error")
	}
	return
}

// UpLolGame update lol game.
func (d *Dao) UpLolGame(c context.Context, data *mdlesp.LolGame) (err error) {
	var inTeams, inPlayers string
	if inTeams, err = d.escapeStr(data.Teams, false); err != nil {
		err = errors.Wrapf(err, "UpLOLGame  inTeams json.Marshal error")
		return
	}
	if inPlayers, err = d.escapeStr(data.Players, false); err != nil {
		err = errors.Wrapf(err, "UpLOLGame  inPlayers json.Marshal error")
		return
	}
	if _, err = d.db.Exec(c, _lolGameUPdateSQL, inTeams, inPlayers, data.Position, data.BeginAt, data.EndAt, data.Finished, data.MatchID, data.ID); err != nil {
		err = errors.Wrapf(err, "UpLOLGame db.Exec error")
	}
	return
}

// UpDotaGame update dota game.
func (d *Dao) UpDotaGame(c context.Context, data *mdlesp.DotaGame) (err error) {
	var inTeams, inPlayers string
	if inTeams, err = d.escapeStr(data.Teams, false); err != nil {
		err = errors.Wrapf(err, "UpDotaGame  inTeams json.Marshal error")
		return
	}
	if inPlayers, err = d.escapeStr(data.Players, false); err != nil {
		err = errors.Wrapf(err, "UpDotaGame  inPlayers json.Marshal error")
		return
	}
	if _, err = d.db.Exec(c, _dotaGameUPdateSQL, inTeams, inPlayers, data.Position, data.BeginAt, data.EndAt, data.Finished, data.MatchID, data.ID); err != nil {
		err = errors.Wrapf(err, "UpDotaGame db.Exec error")
	}
	return
}

// UpOwGame update dota game.
func (d *Dao) UpOwGame(c context.Context, data *mdlesp.OwGame) (err error) {
	var inTeams, inMap string
	if inTeams, err = d.escapeStr(data.Teams, false); err != nil {
		err = errors.Wrapf(err, "UpOwGame  inRounds json.Marshal error")
		return
	}
	if inMap, err = d.escapeStr(data.Map, false); err != nil {
		err = errors.Wrapf(err, "UpOwGame  inMap json.Marshal error")
		return
	}
	if _, err = d.db.Exec(c, _owGameUPdateSQL, data.WinTeam, inTeams, inMap, data.Position, data.BeginAt, data.EndAt, data.Finished, data.MatchID, data.ID); err != nil {
		err = errors.Wrapf(err, "UpOwGame db.Exec error")
	}
	return
}

func (d *Dao) escapeStr(data interface{}, isInsert bool) (rs string, err error) {
	var bs []byte
	if bs, err = json.Marshal(data); err != nil {
		rs = _empJson
		err = nil
		return
	}
	rs = string(bs)
	if rs == "null" {
		rs = _empJson
		return
	}
	if isInsert {
		rs = strings.Replace(rs, "'", "\\'", -1)
	}
	if rs == "" {
		rs = _empJson
	}
	return
}

// AddLolCham  lol cham.
func (d *Dao) AddLolCham(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolChamSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolCham db.Exec error")
	}
	return
}

// AddDotaHero dota hero.
func (d *Dao) AddDotaHero(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaHeroSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotaHero db.Exec error")
	}
	return
}

// AddOwHero overwatch hero.
func (d *Dao) AddOwHero(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_owHeroSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddOwHero db.Exec error")
	}
	return
}

// AddLolitem  lol item.
func (d *Dao) AddLolItem(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolItemSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolitem db.Exec error")
	}
	return
}

// AddDotaitem  dota item.
func (d *Dao) AddDotaItem(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaItemSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotaitem db.Exec error")
	}
	return
}

// AddOwMap  overwatch map.
func (d *Dao) AddOwMap(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_owMapSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddOwMap db.Exec error")
	}
	return
}

// AddLolMatchPlayer add lol player.
func (d *Dao) AddLolMatchPlayer(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolPlayerSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolMatchPlayer db.Exec error")
	}
	return
}

// AddDotaMatchPlayer add dota player.
func (d *Dao) AddDotaMatchPlayer(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaPlayerSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotaMatchPlayer db.Exec error")
	}
	return
}

// AddOwMatchPlayer add ow player.
func (d *Dao) AddOwMatchPlayer(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_owPlayerSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddOwMatchPlayer db.Exec error")
	}
	return
}

// AddLolAbility add lol spells.
func (d *Dao) AddLolAbility(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolAbilitySQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolAbility db.Exec error")
	}
	return
}

// AddDotaAbility add dota ability.
func (d *Dao) AddDotaAbility(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaAbilitySQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolAbility db.Exec error")
	}
	return
}

// AddLolTeams add lol teams.
func (d *Dao) AddLolTeams(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_lolTeamsSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddLolTeams db.Exec error")
	}
	return
}

// AddDotateams add dota teams.
func (d *Dao) AddDotateams(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_dotaTeamsSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddDotateams db.Exec error")
	}
	return
}

// AddOwTeams add ow teams.
func (d *Dao) AddOwTeams(c context.Context, data []*mdlesp.LdInfo) (err error) {
	if len(data) == 0 {
		return
	}
	var rowStrings []string
	for _, v := range data {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,'%s','%s')",
			v.ID, v.Name, v.ImageURL))
	}
	if _, err = d.db.Exec(c, fmt.Sprintf(_owTeamsSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "AddOwTeams db.Exec error")
	}
	return
}

// RuleWhite auto arc white rule.
func (d *Dao) RuleWhite(c context.Context) (rs map[int64]*mdlesp.RuleRs, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _autoWhiteSQL)
	if err != nil {
		err = errors.Wrap(err, "RuleWhite:d.db.Query")
		return
	}
	defer rows.Close()
	rs = make(map[int64]*mdlesp.RuleRs)
	for rows.Next() {
		r := new(mdlesp.RuleRs)
		if err = rows.Scan(&r.ID, &r.GameIDs, &r.MatchIDs); err != nil {
			err = errors.Wrap(err, "RuleWhite:rows.Scan() error")
			return
		}
		rs[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RuleWhite:rows.Err")
	}
	return
}

// RuleTag auto arc tag rule.
func (d *Dao) RuleTag(c context.Context) (rs map[string]*mdlesp.RuleRs, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _autoTagSQL)
	if err != nil {
		err = errors.Wrap(err, "RuleTag:d.db.Query")
		return
	}
	defer rows.Close()
	rs = make(map[string]*mdlesp.RuleRs)
	for rows.Next() {
		r := new(mdlesp.RuleRs)
		if err = rows.Scan(&r.ID, &r.Name, &r.GameIDs, &r.MatchIDs); err != nil {
			err = errors.Wrap(err, "RuleTag:rows.Scan() error")
			return
		}
		rs[strings.ToLower(r.Name)] = r
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RuleTag:rows.Err")
	}
	return
}

// RuleKeyword auto arc keyword rule.
func (d *Dao) RuleKeyword(c context.Context) (rs map[string]*mdlesp.RuleRs, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _autoKeywordSQL)
	if err != nil {
		err = errors.Wrap(err, "RuleKeyword:d.db.Query")
		return
	}
	rs = make(map[string]*mdlesp.RuleRs)
	defer rows.Close()
	for rows.Next() {
		r := new(mdlesp.RuleRs)
		if err = rows.Scan(&r.ID, &r.Name, &r.GameIDs, &r.MatchIDs); err != nil {
			err = errors.Wrap(err, "RuleKeyword:rows.Scan() error")
			return
		}
		rs[r.Name] = r
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RuleKeyword:rows.Err")
	}
	return
}

func (d *Dao) autoAddArc(tx *xsql.Tx, aid, checkTp int64) (lastID int64, err error) {
	var sqlRes sql.Result
	sqlRes, err = tx.Exec(_autoInsertArcSQL, aid, _autoSoure, checkTp)
	if err != nil {
		err = errors.Wrap(err, "AutoAddArc:rows.Err")
		return
	}
	return sqlRes.LastInsertId()
}

func (d *Dao) autoAddGame(tx *xsql.Tx, aid int64, gameIDs []int64) (err error) {
	if aid == 0 || len(gameIDs) == 0 {
		return
	}
	var rowStrings []string
	for _, gid := range gameIDs {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d,%d)", _arcType, aid, gid))
	}
	if _, err = tx.Exec(fmt.Sprintf(_autoGameInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "autoAddGame db.Exec error")
	}
	return
}

func (d *Dao) autoDelGame(tx *xsql.Tx, aid int64) (err error) {
	if aid == 0 {
		return
	}
	if _, err = tx.Exec(_autoGameDelSQL, aid, _arcType); err != nil {
		err = errors.Wrapf(err, "autoDelGame db.Exec error")
	}
	return
}

func (d *Dao) autoAddMatch(tx *xsql.Tx, aid int64, matchIDs []int64) (err error) {
	if aid == 0 || len(matchIDs) == 0 {
		return
	}
	var rowStrings []string
	for _, mid := range matchIDs {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d)", mid, aid))
	}
	if _, err = tx.Exec(fmt.Sprintf(_autoMatchInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "autoAddGame db.Exec error")
	}
	return
}

func (d *Dao) autoDelMatch(tx *xsql.Tx, aid int64) (err error) {
	if aid == 0 {
		return
	}
	if _, err = tx.Exec(_autoMatchDelSQL, aid); err != nil {
		err = errors.Wrapf(err, "autoDelMatch db.Exec error")
	}
	return
}

func (d *Dao) AutoArcPass(ctx context.Context, aid int64) (err error) {
	if aid == 0 {
		return
	}
	if _, err = d.db.Exec(ctx, _autoArcPassSQL, aid); err != nil {
		err = errors.Wrapf(err, "AutoArcPass db.Exec error")
	}
	return
}

func (d *Dao) autoDelTeam(tx *xsql.Tx, aid int64) (err error) {
	if aid == 0 {
		return
	}
	if _, err = tx.Exec(_autoTeamDelSQL, aid); err != nil {
		err = errors.Wrapf(err, "autoDelTeam db.Exec error")
	}
	return
}

func (d *Dao) autoAddTeam(tx *xsql.Tx, aid int64, teamIDs []int64) (err error) {
	if aid == 0 || len(teamIDs) == 0 {
		return
	}
	var rowStrings []string
	for _, tid := range teamIDs {
		rowStrings = append(rowStrings, fmt.Sprintf("(%d,%d)", tid, aid))
	}
	if _, err = tx.Exec(fmt.Sprintf(_autoTeamInsertSQL, strings.Join(rowStrings, ","))); err != nil {
		err = errors.Wrapf(err, "autoAddTeam db.Exec error")
	}
	return
}

func (d *Dao) autoDelYear(tx *xsql.Tx, aid int64) (err error) {
	if aid == 0 {
		return
	}
	if _, err = tx.Exec(_autoYearDelSQL, aid); err != nil {
		err = errors.Wrapf(err, "autoDelYear db.Exec error")
	}
	return
}

func (d *Dao) autoAddYear(tx *xsql.Tx, aid, pubYear int64) (lastID int64, err error) {
	if pubYear == 0 {
		return
	}
	var sqlRes sql.Result
	sqlRes, err = tx.Exec(_autoYearInsertSQL, pubYear, aid)
	if err != nil {
		err = errors.Wrap(err, "autoAddYear:rows.Err")
		return
	}
	return sqlRes.LastInsertId()
}

func (d *Dao) autoDelTag(tx *xsql.Tx, aid int64) (err error) {
	if aid == 0 {
		return
	}
	if _, err = tx.Exec(_autoTagDelSQL, aid); err != nil {
		err = errors.Wrapf(err, "autoDelTag db.Exec error")
	}
	return
}

func (d *Dao) autoAddTag(tx *xsql.Tx, aid, tid int64) (lastID int64, err error) {
	if tid == 0 {
		return
	}
	var sqlRes sql.Result
	sqlRes, err = tx.Exec(_autoTagsInsertSQL, tid, aid)
	if err != nil {
		err = errors.Wrap(err, "autoAddTag:rows.Err")
		return
	}
	return sqlRes.LastInsertId()
}

func (d *Dao) autoAddHit(tx *xsql.Tx, arcID, mid int64, tags, keywords string) (lastID int64, err error) {
	var sqlRes sql.Result
	sqlRes, err = tx.Exec(_autoInsertHitSQL, arcID, mid, tags, keywords)
	if err != nil {
		err = errors.Wrap(err, "autoAddHit:rows.Err")
		return
	}
	return sqlRes.LastInsertId()
}

func (d *Dao) autoUpdateHit(tx *xsql.Tx, arcID, mid int64, tags, keywords string) (err error) {
	if arcID == 0 {
		return
	}
	if _, err = tx.Exec(_autoUpdateHitSQL, mid, tags, keywords, arcID); err != nil {
		err = errors.Wrapf(err, "autoUpdateHit db.Exec error")
	}
	return
}

// AutoAdd.
func (d *Dao) AutoAdd(c context.Context, aid, mid, tid int64, tags, keywords string, gameIDs, matchIDs, teamIDs []int64, pubYear, checkTp int64) (err error) {
	var (
		tx    *xsql.Tx
		arcID int64
	)
	if tx, err = d.db.Begin(c); err != nil {
		err = errors.Wrap(err, "AutoAdd:d.db.Begin().Err")
		return
	}
	defer func() {
		if err != nil {
			log.Error("AutoAdd aid(%d) error(%+v)", aid, err)
			if err1 := tx.Rollback(); err1 != nil {
				err = errors.Wrap(err1, "AutoAdd:tx.Rollback().Err")
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "AutoAdd:tx.Commit().Err")
		}
	}()
	if arcID, err = d.autoAddArc(tx, aid, checkTp); err != nil {
		return
	}
	if err = d.autoAddGame(tx, aid, gameIDs); err != nil {
		return
	}
	if err = d.autoAddMatch(tx, aid, matchIDs); err != nil {
		return
	}
	if err = d.autoAddTeam(tx, aid, teamIDs); err != nil {
		return
	}
	if _, err = d.autoAddYear(tx, aid, pubYear); err != nil {
		return
	}
	if _, err = d.autoAddTag(tx, aid, tid); err != nil {
		return
	}
	_, err = d.autoAddHit(tx, arcID, mid, tags, keywords)
	return
}

// AutoUpdate.
func (d *Dao) AutoUpdate(c context.Context, aid, mid, tid int64, tags, keywords string, gameIDs, matchIDs, teamIDs []int64, pubYear, checkTp, arcID int64) (err error) {
	var tx *xsql.Tx
	if tx, err = d.db.Begin(c); err != nil {
		err = errors.Wrap(err, "AutoUpdate:d.db.Begin().Err")
		return
	}
	defer func() {
		if err != nil {
			log.Error("AutoUpdate aid(%d) error(%+v)", aid, err)
			if err1 := tx.Rollback(); err1 != nil {
				err = errors.Wrap(err1, "AutoUpdate:tx.Rollback().Err")
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "AutoUpdate:tx.Commit().Err")
		}
	}()
	if arcID == 0 {
		if arcID, err = d.autoAddHit(tx, arcID, mid, tags, keywords); err != nil {
			return
		}
	} else {
		if err = d.autoUpdateHit(tx, arcID, mid, tags, keywords); err != nil {
			return
		}
	}
	if err = d.autoDelGame(tx, aid); err != nil {
		return
	}
	if err = d.autoAddGame(tx, aid, gameIDs); err != nil {
		return
	}
	if err = d.autoDelMatch(tx, aid); err != nil {
		return
	}
	if err = d.autoAddMatch(tx, aid, matchIDs); err != nil {
		return
	}
	if err = d.autoDelTeam(tx, aid); err != nil {
		return
	}
	if err = d.autoAddTeam(tx, aid, teamIDs); err != nil {
		return
	}
	if err = d.autoDelYear(tx, aid); err != nil {
		return
	}
	if _, err = d.autoAddYear(tx, aid, pubYear); err != nil {
		return
	}
	if err = d.autoDelTag(tx, aid); err != nil {
		return
	}
	if _, err = d.autoAddTag(tx, aid, tid); err != nil {
		return
	}
	return
}

// AutoArc.
func (d *Dao) AutoArc(c context.Context, aid int64) (rs int64, err error) {
	row := d.db.QueryRow(c, _autoArcSQL, aid)
	if err = row.Scan(&rs); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "AutoArc row.Scan().Err")
		}
	}
	return
}

// PushHandUse.
func (d *Dao) PushHandUse(c context.Context, contestID int64) (rs *mdlesp.Contest, err error) {
	rs = new(mdlesp.Contest)
	row := d.db.QueryRow(c, _contestPushHandSQL, contestID)
	if err = row.Scan(&rs.ID, &rs.Stime, &rs.Etime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "PushHandUse row.Scan().Err")
		}
	}
	return
}

func (d *Dao) ContestById(c context.Context, contestId int64) (r *mdlesp.Contest, err error) {
	r = new(mdlesp.Contest)
	if err = d.db.QueryRow(c, _sqlContestById, contestId).
		Scan(&r.ID, &r.Stime, &r.LiveRoom, &r.HomeID, &r.AwayID, &r.SuccessTeam, &r.Special, &r.SpecialName, &r.SpecialTips, &r.SeasonTitle, &r.SeasonSubTitle, &r.ContestStatus, &r.MessageSendUid); err != nil {
		log.Error("[ContestsById]Contests:row.Scan() error(%v)", err)
		return
	}
	return
}

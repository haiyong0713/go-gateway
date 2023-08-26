package dao

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	ftpModel "go-gateway/app/web-svr/esports/job/model"
)

const (
	_seasonWhere      = "FROM es_seasons WHERE `status` = 0"
	_seasonCount      = "SELECT count(*) " + _seasonWhere
	_seasonSQL        = "SELECT id,mid,title,sub_title,stime,etime,sponsor,logo,dic,ctime,mtime,status,rank,is_app,url,data_focus,focus_url,leida_sid " + _seasonWhere + " AND id > ? ORDER BY id ASC limit ?"
	_teamsWhere       = "FROM es_teams WHERE is_deleted=0"
	_teamsCount       = "SELECT count(*) " + _teamsWhere
	_teamsSQL         = "SELECT id,title,sub_title,logo,team_type,e_title " + _teamsWhere + " AND id > ? ORDER BY id ASC limit ?"
	_ftpContestsWhere = "FROM es_contests WHERE `status` = 0"
	_ftpContestsCount = "SELECT count(*) " + _ftpContestsWhere
	_ftpContestsSQL   = "SELECT id,game_stage,stime,etime,home_id,away_id,home_score,away_score,live_room,aid,collection,game_state,dic,ctime,mtime,status,sid,mid,special,special_name,special_tips,success_team,special_image,playback,collection_url,live_url,data_type,match_id,guess_type " + _ftpContestsWhere + " AND id > ? ORDER BY id ASC limit ?"
	_matchsWhere      = "FROM es_matchs WHERE status=0"
	_matchsSQL        = "SELECT id,title,sub_title,logo,rank FROM es_matchs WHERE status=0 " + " AND id > ? ORDER BY id ASC limit ?"
	_matchsCount      = "SELECT count(*) " + _matchsWhere
)

// FtpMatchsCount .
func (d *Dao) FtpMatchsCount(c context.Context) (rs int, err error) {
	row := d.db.QueryRow(c, _matchsCount)
	if err = row.Scan(&rs); err != nil {
		log.Error("FtpMatchsCount: %s error(%v)", _matchsCount, err)
	}
	return
}

// FtpMatchs .
func (d *Dao) FtpMatchs(c context.Context, id, limit int64) (res []*ftpModel.FtpMatchs, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _matchsSQL, id, limit); err != nil {
		log.Error("Match:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(ftpModel.FtpMatchs)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle, &r.Logo, &r.Rank); err != nil {
			log.Error("Match:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// FtpContestsCount .
func (d *Dao) FtpContestsCount(c context.Context) (contestCount int, err error) {
	row := d.db.QueryRow(c, _ftpContestsCount)
	if err = row.Scan(&contestCount); err != nil {
		log.Error("FtpTeamsCount.Query: %s error(%v)", _seasonCount, err)
	}
	return
}

// FtpContests .
func (d *Dao) FtpContests(c context.Context, id, limit int64) (res []*ftpModel.FtpContest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _ftpContestsSQL, id, limit); err != nil {
		log.Error("FtpContests: db.Query(%s) id(%d) limit(%d) error(%v)", _ftpContestsSQL, id, limit, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(ftpModel.FtpContest)
		if err = rows.Scan(&r.ID, &r.GameStage, &r.Stime, &r.Etime, &r.HomeID, &r.AwayID, &r.HomeScore, &r.AwayScore,
			&r.LiveRoom, &r.Aid, &r.Collection, &r.GameState, &r.Dic, &r.Ctime, &r.Mtime, &r.Status, &r.Sid, &r.Mid,
			&r.Special, &r.SpecialName, &r.SpecialTips, &r.SuccessTeam, &r.SpecialImage, &r.Playback, &r.CollectionURL,
			&r.LiveURL, &r.DataType, &r.MatchID, &r.GuessType); err != nil {
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

// FtpTeamsCount .
func (d *Dao) FtpTeamsCount(c context.Context) (seasonCount int, err error) {
	row := d.db.QueryRow(c, _teamsCount)
	if err = row.Scan(&seasonCount); err != nil {
		log.Error("FtpTeamsCount.Query: %s error(%v)", _seasonCount, err)
	}
	return
}

// FtpTeams .
func (d *Dao) FtpTeams(c context.Context, id, limit int64) (res []*ftpModel.FtpTeams, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _teamsSQL, id, limit); err != nil {
		log.Error("FtpTeams:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		team := new(ftpModel.FtpTeams)
		if err = rows.Scan(&team.ID, &team.Title, &team.SubTitle, &team.Logo, &team.TeamType, &team.ETitle); err != nil {
			log.Error("FtpTeams:row.Scan() error(%v)", err)
			return
		}
		res = append(res, team)
	}
	if err = rows.Err(); err != nil {
		log.Error("FtpTeams rows.Err() error(%v)", err)
	}
	return
}

// SeasonCount season count
func (d *Dao) SeasonCount(c context.Context) (seasonCount int, err error) {
	row := d.db.QueryRow(c, _seasonCount)
	if err = row.Scan(&seasonCount); err != nil {
		log.Error("d.SeasonCount.Query: %s error(%v)", _seasonCount, err)
	}
	return
}

// Season .
func (d *Dao) Season(c context.Context, id, limit int64) (res []*ftpModel.FtpSeason, err error) {
	var (
		rows *xsql.Rows
	)

	if rows, err = d.db.Query(c, _seasonSQL, id, limit); err != nil {
		log.Error("Season d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(ftpModel.FtpSeason)
		if err = rows.Scan(&r.ID, &r.Mid, &r.Title, &r.SubTitle, &r.Stime, &r.Etime, &r.Sponsor, &r.Logo, &r.Dic, &r.Ctime,
			&r.Mtime, &r.Status, &r.Rank, &r.IsApp, &r.URL, &r.DataFocus, &r.FocusURL, &r.LeidaSID); err != nil {
			log.Error("Season row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("Season rows.Err() error(%v)", err)
	}
	return
}

package dao

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	mdlEp "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"

	"github.com/pkg/errors"
)

const (
	_decimalOne        = 1
	_matchsSQL         = "SELECT id,title,sub_title,logo,rank FROM es_matchs WHERE status=0 ORDER BY rank DESC, ID ASC"
	_gamesSQL          = "SELECT id,title,sub_title,logo FROM es_games WHERE status=0 ORDER BY id ASC"
	_teamsSQL          = "SELECT id,title,sub_title,logo FROM es_teams WHERE is_deleted=0 ORDER BY id ASC"
	_tagsSQL           = "SELECT id,name FROM es_tags WHERE status=0 ORDER BY id ASC"
	_yearsSQL          = "SELECT distinct year as id, year FROM es_year_map WHERE is_deleted=0 ORDER BY id ASC"
	_dayContestSQL     = "SELECT FROM_UNIXTIME(stime, '%Y-%m-%d') as s,count(1) as c FROM `es_contests` WHERE status=0 AND stime>=? AND stime<=? GROUP BY s ORDER BY stime"
	_seasonSQL         = "SELECT id,mid,title,sub_title,stime,etime,sponsor,logo,dic,ctime,mtime,status,rank,is_app,url,data_focus,focus_url,leida_sid,search_image,sync_platform FROM es_seasons WHERE  status=0  ORDER BY stime DESC"
	_epSeasonSQL       = "SELECT id,mid,title,sub_title,stime,etime,sponsor,logo,dic,ctime,mtime,status,rank,is_app,url,data_focus,focus_url,leida_sid,search_image,sync_platform FROM es_seasons WHERE  status=0 AND id in (%s) ORDER BY stime DESC"
	_epGameSQL         = "SELECT id,title,sub_title,e_title,plat,type,logo,publisher,operations,pb_time,dic FROM es_games WHERE status=0 AND id IN (%s)"
	_epGameMapSQL      = "SELECT oid,gid FROM es_gid_map WHERE is_deleted=0 AND type=? AND oid IN (%s)"
	_seasonMSQL        = "SELECT id,mid,title,sub_title,stime,etime,sponsor,logo,dic,ctime,mtime,status,rank,is_app,url,data_focus,focus_url,leida_sid,search_image,sync_platform FROM es_seasons WHERE  status=0 AND is_app=1  ORDER BY rank DESC,stime DESC"
	_seasonLdSQL       = "SELECT id,mid,title,sub_title,stime,etime,sponsor,logo,dic,ctime,mtime,status,rank,is_app,url,data_focus,focus_url,leida_sid,search_image,sync_platform FROM es_seasons WHERE  leida_sid=?"
	_contestSQL        = "SELECT id,game_stage,stime,etime,home_id,away_id,home_score,away_score,live_room,aid,collection,game_state,dic,ctime,mtime,status,sid,mid,special,special_name,special_tips,success_team,special_image,playback,collection_url,live_url,data_type,match_id,guess_type,series_id,contest_status FROM `es_contests` WHERE id=?"
	_contestsSQL       = "SELECT id,game_stage,stime,etime,home_id,away_id,home_score,away_score,live_room,aid,collection,game_state,dic,ctime,mtime,status,sid,mid,special,special_name,special_tips,success_team,special_image,playback,collection_url,live_url,data_type,match_id,guess_type,game_stage1,game_stage2,push_switch,series_id,contest_status FROM `es_contests` WHERE id IN (%s) ORDER BY ID ASC"
	_s9contestsSQL     = "SELECT id,game_stage,stime,etime,home_id,away_id,home_score,away_score,live_room,aid,collection,game_state,dic,ctime,mtime,status,sid,mid,special,special_name,special_tips,success_team,special_image,playback,collection_url,live_url,data_type,match_id,guess_type,contest_status FROM `es_contests` WHERE status=0 AND sid=? AND stime>=? AND etime<=? ORDER BY stime ASC LIMIT 50"
	_moduleSQL         = "SELECT id,ma_id,name,oids FROM `es_matchs_module` WHERE id=? AND status=0"
	_activeSQL         = "SELECT id,mid,sid,background,live_id,intr,focus,url,back_color,color_step,h5_background,h5_back_color,intr_logo,intr_title,intr_text,h5_focus,h5_url,sids,is_live FROM es_matchs_active WHERE id=? AND `status`=0"
	_modulesSQL        = "SELECT id,ma_id,name,oids FROM `es_matchs_module` WHERE ma_id=? AND status=0 ORDER BY ID ASC"
	_pDetailSQL        = "SELECT ma_id,game_type,stime,etime FROM es_matchs_detail WHERE id=? AND `status`=0"
	_actDetail         = "SELECT id,ma_id,game_type,stime,etime,score_id,game_stage,knockout_type,winner_type,online FROM es_matchs_detail WHERE ma_id=? AND status=0"
	_actLive           = "SELECT id,ma_id,live_id,title FROM es_active_live WHERE ma_id=? AND is_deleted=0"
	_treeSQL           = "SELECT id,ma_id,mad_id,pid,root_id,game_rank,mid FROM es_matchs_tree WHERE mad_id=? AND is_deleted=0 ORDER BY root_id ASC,pid ASC,game_rank ASC"
	_teamsInSQL        = "SELECT id,title,sub_title,logo,video_url,profile,leida_tid,reply_id,team_type,region_id FROM es_teams WHERE is_deleted=0 AND id in (%s)"
	_teamLdSQL         = "SELECT id,title,sub_title,logo,video_url,profile,leida_tid,reply_id FROM es_teams WHERE leida_tid=?"
	_kDetailsSQL       = "SELECT id,ma_id,game_type,stime,etime,online FROM es_matchs_detail WHERE `status`=0 AND game_type=2"
	_contestDataSQL    = "SELECT id,cid,url,point_data FROM `es_contests_data` WHERE cid=? AND is_deleted=0"
	_contestRecent     = "SELECT id,game_stage,stime,etime,home_id,away_id,home_score,away_score,live_room,aid,collection,game_state,dic,ctime,mtime,status,sid,mid,special,special_name,special_tips,success_team,special_image,playback,collection_url,live_url,data_type,guess_type,contest_status FROM es_contests WHERE (`status`=0 AND home_id=? AND away_id=?) OR (`status`=0 AND home_id=? AND away_id=?) ORDER BY stime DESC LIMIT ?"
	_liveSQL           = "SELECT ma_id,live_id,title FROM es_active_live WHERE live_id=? AND is_deleted=0 ORDER BY ma_id limit 1"
	_lolPlayerSQL      = "SELECT id,player_id,team_id,team_acronym,team_image,leida_sid,name,image_url,champions_image,role,win,kda,kills,deaths,assists,minions_killed,wards_placed,games_count,ctime,mtime,position_id,position,mvp FROM es_lol_player WHERE leida_sid=?"
	_lolTeamSQL        = "SELECT id,team_id,leida_sid,name,acronym,image_url,win,kda,kills,deaths,assists,tower_kills,total_minions_killed,first_tower,first_inhibitor,first_dragon,first_baron,first_blood,wards_placed,inhibitor_kills,baron_kills,gold_earned,games_count,players,ctime,mtime,baron_rate,dragon_rate,hits,lose_num,money,total_damage,win_num,image_thumb,new_data FROM es_lol_team WHERE leida_sid=?"
	_dotaPlayerSQL     = "SELECT id,player_id,team_id,team_acronym,team_image,leida_sid,name,image_url,heroes_image,role,win,kda,kills,deaths,assists,wards_placed,last_hits,observer_wards_placed,sentry_wards_placed,xp_per_minute,gold_per_minute,games_count,ctime,mtime FROM es_dota_player WHERE leida_sid=?"
	_dotaTeamSQL       = "SELECT id,team_id,leida_sid,name,acronym,image_url,win,kda,kills,deaths,assists,tower_kills,last_hits,observer_used,sentry_used,xp_per_minute,first_blood,heal,gold_spent,gold_per_min,denies,damage_taken,camps_stacked,games_count,players,ctime,mtime FROM es_dota_team WHERE leida_sid=?"
	_gameSeasonSQL     = "SELECT distinct s.id,s.mid,s.title,s.sub_title,s.stime,s.etime,s.sponsor,s.logo,s.dic,s.ctime,s.mtime,s.status,s.rank,s.is_app,s.url,s.data_focus,s.focus_url,s.leida_sid FROM  es_seasons as s inner join es_gid_map as g on s.id=g.oid WHERE  leida_sid>0 AND g.type=2 AND g.gid=? ORDER BY stime DESC"
	_lolGameSQL        = "SELECT id,game_id,teams,players,position,match_id,begin_at,end_at,finished FROM es_lol_game WHERE match_id=?"
	_dotaGameSQL       = "SELECT id,game_id,teams,players,position,match_id,begin_at,end_at,finished FROM es_dota_game WHERE match_id=?"
	_owGameSQL         = "SELECT id,game_id,win_team,teams,map,position,match_id,begin_at,end_at,finished FROM es_overwatch_game WHERE match_id=?"
	_lolItemSQL        = "SELECT item_id,name,image_url FROM es_lol_item"
	_dotaItemSQL       = "SELECT item_id,name,image_url FROM es_dota_item"
	_owMapSQL          = "SELECT item_id,name,image_url FROM es_overwatch_map"
	_lolAbilitySQL     = "SELECT ability_id,name,image_url FROM es_lol_ability"
	_dotaAbilitySQL    = "SELECT ability_id,name,image_url FROM es_dota_ability"
	_lolchamSQL        = "SELECT hero_id,name,image_url FROM es_lol_champion"
	_dotaHeroSQL       = "SELECT hero_id,name,image_url FROM es_dota_hero"
	_owHeroSQL         = "SELECT hero_id,name,image_url FROM es_overwatch_hero"
	_lolPlaysSQL       = "SELECT player_id,name,image_url FROM es_lol_match_player"
	_dotaPlaysSQL      = "SELECT player_id,name,image_url FROM es_dota_match_player"
	_owPlaysSQL        = "SELECT player_id,name,image_url FROM es_overwatch_match_player"
	_lolTeamsSQL       = "SELECT team_id,name,image_url FROM es_lol_match_team"
	_dotaTeamsSQL      = "SELECT team_id,name,image_url FROM es_dota_match_team"
	_owTeamsSQL        = "SELECT team_id,name,image_url FROM es_overwatch_match_team"
	_guessGameSQL      = "SELECT DISTINCT(gid) FROM es_gid_map LEFT JOIN es_contests on es_contests.id=es_gid_map.oid WHERE (type=3 AND is_deleted=0) AND (guess_type=1 AND stime>? AND `status`=0)"
	_guessSeasonSQL    = "SELECT sid FROM es_contests LEFT JOIN es_gid_map ON es_gid_map.oid=es_contests.id WHERE es_contests.guess_type=1 AND es_contests.`status`=0 AND es_contests.stime>? AND es_gid_map.type=3 AND es_gid_map.is_deleted=0 AND es_gid_map.gid=?"
	_guessSeasonAllSQL = "SELECT sid FROM es_contests WHERE es_contests.guess_type=1 AND es_contests.`status`=0 AND es_contests.stime>?"
	_guessCCalenSQL    = "SELECT FROM_UNIXTIME(stime, '%Y-%m-%d') as stime,count(*) as count FROM es_contests WHERE guess_type=1 AND `status`=0 AND stime>? GROUP BY stime"
	_searchMainSQL     = "SELECT id FROM es_search_card WHERE status=0 ORDER BY id ASC"
	_searchCardSQL     = "SELECT m.id,m.query_name,m.stime,m.etime,d.cid FROM es_search_card AS m INNER JOIN es_search_contest AS d ON m.id=d.mid WHERE m.status=0 AND d.is_deleted=0 AND m.id in (%s) ORDER BY m.id ASC,d.id ASC"
	_selSeasonGidSQL   = "SELECT DISTINCT gid FROM es_season_ranks WHERE is_deleted=0 ORDER BY ID ASC LIMIT 50"
	_selGameRankSQL    = "SELECT id,title,sub_title,rank FROM es_games WHERE status=0 AND id IN (%s) ORDER BY rank DESC"
	_selSeasonRankSQL  = "SELECT id,sid,rank FROM es_season_ranks WHERE is_deleted=0 AND gid=? ORDER BY rank DESC,id DESC LIMIT 20"

	_sqlMatchSeasonSQL    = `SELECT id,title,logo,mid,stime,etime FROM es_seasons WHERE status=0 AND mid IN(%v) ORDER BY mid ASC,stime ASC`
	_sqlSeasonsByMatchSQL = `SELECT id,title,logo,mid,stime,etime FROM es_seasons WHERE status=0 AND mid=? ORDER BY stime ASC`
	_sqlBatchSeasonsSQL   = `SELECT id,title,logo,mid,stime,etime FROM es_seasons WHERE id IN(%v) ORDER BY stime ASC`

	limit4EveryFetch      = 5000
	sql4ContestOfCanGuess = `SELECT id FROM es_contests WHERE stime >= ? ORDER BY stime asc LIMIT ?, ?`
	_sqlTeamsInSeasonSQL  = `SELECT t.id, 
       t.title, 
       t.region_id, 
       r.sid, 
       r.rank,
       t.logo,
       t.uid,
       t.leida_tid
FROM   es_teams t, 
       es_team_in_seasons r 
WHERE  r.sid  IN (%v) 
       AND r.tid = t.id 
       AND t.is_deleted = 0 
ORDER  BY r.rank DESC, 
          t.id ASC `
	_sqlOngoingSeasonSQL = "select id from es_seasons where etime > ?"

	sql4EffectiveTeamList = `
SELECT id, title, sub_title, logo, video_url
	, profile, leida_tid, reply_id, team_type, region_id
FROM es_teams
WHERE is_deleted = 0
	AND id > ?
ORDER BY id ASC
LIMIT 1000
`
	sql4EffectiveSeasonList = `
SELECT id, mid, title, sub_title, stime
	, etime, sponsor, logo, dic, ctime
	, mtime, status, rank, is_app, url
	, data_focus, focus_url, leida_sid, search_image, sync_platform
FROM es_seasons
WHERE status = 0
	AND id > ?
ORDER BY id ASC
LIMIT 1000
`
)

const (
	_limitKey2FetchSeasonsByMatch = "match_seasons_limit"
	_limitKey2BatchSeasonsTeams   = "season_teams_limit"
	_limitKey2FetchTeamsInSeason  = "teams_in_season_limit"
)

// logoURL convert logo url to full url.
func logoURL(uri string) (logo string) {
	if uri == "" {
		return
	}
	logo = uri
	if strings.Index(uri, "http://") == 0 || strings.Index(uri, "//") == 0 {
		return
	}
	if len(uri) >= 10 && uri[:10] == "/templets/" {
		return
	}
	if strings.HasPrefix(uri, "group1") {
		logo = "//i0.hdslb.com/" + uri
		return
	}
	if pos := strings.Index(uri, "/uploads/"); pos != -1 && (pos == 0 || pos == 3) {
		logo = uri[pos+8:]
	}
	logo = strings.Replace(logo, "{IMG}", "", -1)
	logo = "//i0.hdslb.com" + logo
	return
}

// Matchs filter matchs.
func (d *Dao) Matchs(c context.Context) (res []*model.Filter, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _matchsSQL); err != nil {
		log.Error("Match:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Filter)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle, &r.Logo, &r.Rank); err != nil {
			log.Error("Match:row.Scan() error(%v)", err)
			return
		}
		r.Logo = logoURL(r.Logo)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Games filter games.
func (d *Dao) Games(c context.Context) (res []*model.Filter, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _gamesSQL); err != nil {
		log.Error("Games:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Filter)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle, &r.Logo); err != nil {
			log.Error("Games:row.Scan() error(%v)", err)
			return
		}
		r.Logo = logoURL(r.Logo)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Teams filter teams.
func (d *Dao) Teams(c context.Context) (res []*model.Filter, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _teamsSQL); err != nil {
		log.Error("Teams:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Filter)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle, &r.Logo); err != nil {
			log.Error("Teams:row.Scan() error(%v)", err)
			return
		}
		r.Logo = logoURL(r.Logo)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Tags filter Tags.
func (d *Dao) Tags(c context.Context) (res []*model.Filter, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _tagsSQL); err != nil {
		log.Error("Tags:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Filter)
		if err = rows.Scan(&r.ID, &r.Title); err != nil {
			log.Error("Tags:row.Scan() error(%v)", err)
			return
		}
		r.Logo = logoURL(r.Logo)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Years filter years.
func (d *Dao) Years(c context.Context) (res []*model.Filter, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _yearsSQL); err != nil {
		log.Error("Years:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Year)
		if err = rows.Scan(&r.ID, &r.Year); err != nil {
			log.Error("Years:row.Scan() error(%v)", err)
			return
		}
		res = append(res, &model.Filter{ID: r.ID, Title: strconv.FormatInt(r.Year, 10)})
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Calendar  calendar count.
func (d *Dao) Calendar(c context.Context, stime, etime int64) (res []*model.Calendar, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _dayContestSQL, stime, etime); err != nil {
		log.Error("Calendar:d.db.Query(%d,%d) error(%v)", stime, etime, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Calendar)
		if err = rows.Scan(&r.Stime, &r.Count); err != nil {
			log.Error("Calendar:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("Calendar rows.Err() error(%v)", err)
	}
	return
}

// Season season list.
func (d *Dao) Season(c context.Context) (res []*model.Season, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _seasonSQL); err != nil {
		log.Error("Contest:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Season)
		if err = rows.Scan(&r.ID, &r.Mid, &r.Title, &r.SubTitle, &r.Stime, &r.Etime, &r.Sponsor, &r.Logo, &r.Dic, &r.Ctime,
			&r.Mtime, &r.Status, &r.Rank, &r.IsApp, &r.URL, &r.DataFocus, &r.FocusURL, &r.LeidaSID, &r.SearchImage, &r.SyncPlatform); err != nil {
			log.Error("Contest:row.Scan() error(%v)", err)
			return
		}
		r.Logo = logoURL(r.Logo)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// Module active module
func (d *Dao) Module(c context.Context, mmid int64) (mod *model.Module, err error) {
	mod = &model.Module{}
	row := d.db.QueryRow(c, _moduleSQL, mmid)
	if err = row.Scan(&mod.ID, &mod.MAid, &mod.Name, &mod.Oids); err != nil {
		if err == sql.ErrNoRows {
			mod = nil
			err = nil
		} else {
			log.Error("Esport dao Module:row.Scan error(%v)", err)
		}
	}
	return
}

// Modules active module
func (d *Dao) Modules(c context.Context, aid int64) (mods []*model.Module, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _modulesSQL, aid); err != nil {
		log.Error("Esport dao Modules:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Module)
		if err = rows.Scan(&r.ID, &r.MAid, &r.Name, &r.Oids); err != nil {
			log.Error("Esport dao Modules:row.Scan() error(%v)", err)
			return
		}
		mods = append(mods, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("Esport dao Modules.Err() error(%v)", err)
	}
	return
}

// Trees match tree
func (d *Dao) Trees(c context.Context, madID int64) (mods []*model.Tree, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _treeSQL, madID); err != nil {
		log.Error("Esport dao Trees:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Tree)
		if err = rows.Scan(&r.ID, &r.MaID, &r.MadID, &r.Pid, &r.RootID, &r.GameRank, &r.Mid); err != nil {
			log.Error("Esport dao Trees:row.Scan() error(%v)", err)
			return
		}
		mods = append(mods, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("Esport dao Trees.Err() error(%v)", err)
	}
	return
}

// Active matchs active
func (d *Dao) Active(c context.Context, aid int64) (mod *model.Active, err error) {
	mod = &model.Active{}
	row := d.db.QueryRow(c, _activeSQL, aid)
	if err = row.Scan(&mod.ID, &mod.Mid, &mod.Sid, &mod.Background, &mod.Liveid, &mod.Intr, &mod.Focus, &mod.URL,
		&mod.BackColor, &mod.ColorStep, &mod.H5Background, &mod.H5BackColor, &mod.IntrLogo, &mod.IntrTitle, &mod.IntrText,
		&mod.H5Focus, &mod.H5Url, &mod.Sids, &mod.IsLive); err != nil {
		if err == sql.ErrNoRows {
			mod = nil
			err = nil
		} else {
			log.Error("Esport dao Active:row.Scan error(%v)", err)
		}
	}
	return
}

// PActDetail poin match detail
func (d *Dao) PActDetail(c context.Context, id int64) (mod *model.ActiveDetail, err error) {
	mod = &model.ActiveDetail{}
	row := d.db.QueryRow(c, _pDetailSQL, id)
	if err = row.Scan(&mod.Maid, &mod.GameType, &mod.STime, &mod.ETime); err != nil {
		if err == sql.ErrNoRows {
			mod = nil
			err = nil
		} else {
			log.Error("Esport dao Contest:row.Scan error(%v)", err)
		}
	}
	return
}

// ActDetail data module
func (d *Dao) ActDetail(c context.Context, aid int64) (actDetail []*model.ActiveDetail, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _actDetail, aid); err != nil {
		log.Error("Esport dao ActDetail:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.ActiveDetail)
		if err = rows.Scan(&r.ID, &r.Maid, &r.GameType, &r.STime, &r.ETime, &r.ScoreID, &r.GameStage, &r.KnockoutType, &r.WinnerType, &r.Online); err != nil {
			log.Error("Esport dao ActDetail:row.Scan() error(%v)", err)
			return
		}
		actDetail = append(actDetail, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("Esport dao ActDetail.Err() error(%v)", err)
	}
	return
}

// ActLives lives data.
func (d *Dao) ActLives(c context.Context, aid int64) (actLives []*model.ActiveLives, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _actLive, aid); err != nil {
		log.Error("Esport dao ActLives:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.ActiveLives)
		if err = rows.Scan(&r.ID, &r.Maid, &r.LiveID, &r.Title); err != nil {
			log.Error("Esport dao ActLives:row.Scan() error(%v)", err)
			return
		}
		actLives = append(actLives, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("Esport dao ActLives.Err() error(%v)", err)
	}
	return
}

// AppSeason season match list.
func (d *Dao) AppSeason(c context.Context) (res []*model.Season, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _seasonMSQL); err != nil {
		log.Error("Contest:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Season)
		if err = rows.Scan(&r.ID, &r.Mid, &r.Title, &r.SubTitle, &r.Stime, &r.Etime, &r.Sponsor, &r.Logo,
			&r.Dic, &r.Ctime, &r.Mtime, &r.Status, &r.Rank, &r.IsApp, &r.URL, &r.DataFocus, &r.FocusURL, &r.LeidaSID, &r.SearchImage, &r.SyncPlatform); err != nil {
			log.Error("Contest:row.Scan() error(%v)", err)
			return
		}
		r.Logo = logoURL(r.Logo)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// LdSeason leida season.
func (d *Dao) LdSeason(c context.Context, ldSid int64) (mod *model.Season, err error) {
	mod = &model.Season{}
	row := d.db.QueryRow(c, _seasonLdSQL, ldSid)
	if err = row.Scan(&mod.ID, &mod.Mid, &mod.Title, &mod.SubTitle, &mod.Stime, &mod.Etime, &mod.Sponsor, &mod.Logo,
		&mod.Dic, &mod.Ctime, &mod.Mtime, &mod.Status, &mod.Rank, &mod.IsApp, &mod.URL, &mod.DataFocus, &mod.FocusURL, &mod.LeidaSID, &mod.SearchImage, &mod.SyncPlatform); err != nil {
		if err == sql.ErrNoRows {
			mod = nil
			err = nil
		} else {
			err = errors.Wrapf(err, "LdSeason:row.Scan() error")
		}
	}
	return
}

// Contest get contest by id.
func (d *Dao) Contest(c context.Context, cid int64) (res *model.Contest, err error) {
	res = &model.Contest{}
	row := d.db.QueryRow(c, _contestSQL, cid)
	if err = row.Scan(&res.ID, &res.GameStage, &res.Stime, &res.Etime, &res.HomeID, &res.AwayID, &res.HomeScore, &res.AwayScore,
		&res.LiveRoom, &res.Aid, &res.Collection, &res.GameState, &res.Dic, &res.Ctime, &res.Mtime, &res.Status, &res.Sid, &res.Mid,
		&res.Special, &res.SpecialName, &res.SpecialTips, &res.SuccessTeam, &res.SpecialImage, &res.Playback, &res.CollectionURL,
		&res.LiveURL, &res.DataType, &res.MatchID, &res.GuessType, &res.SeriesID, &res.ContestStatus); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("Contest:row.Scan error(%v)", err)
		}
	}
	return
}

// ContestRecent get recent contest
func (d *Dao) ContestRecent(c context.Context, homeid, awayid, contestid, ps int64) (res []*model.Contest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _contestRecent, homeid, awayid, awayid, homeid, ps+1); err != nil {
		log.Error("ContestRecent: db.Exec(%s) error(%v)", _contestRecent, err)
		return
	}
	defer rows.Close()
	res = make([]*model.Contest, 0)
	for rows.Next() {
		r := new(model.Contest)
		if err = rows.Scan(&r.ID, &r.GameStage, &r.Stime, &r.Etime, &r.HomeID, &r.AwayID, &r.HomeScore, &r.AwayScore,
			&r.LiveRoom, &r.Aid, &r.Collection, &r.GameState, &r.Dic, &r.Ctime, &r.Mtime, &r.Status, &r.Sid, &r.Mid,
			&r.Special, &r.SpecialName, &r.SpecialTips, &r.SuccessTeam, &r.SpecialImage, &r.Playback, &r.CollectionURL,
			&r.LiveURL, &r.DataType, &r.GuessType, &r.ContestStatus); err != nil {
			log.Error("ContestRecent:row.Scan() error(%v)", err)
			return
		}
		if r.ID != contestid && len(res) != int(ps) {
			res = append(res, r)
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("ContestRecent rows.Err() error(%v)", err)
	}
	return
}

//func (d *Dao) FetchContestList4CanGuess(ctx context.Context) (list []int64) {
//	list = make([]int64, 0)
//	startTime := time.Now().Unix() + 600
//	var startIndex, count4EveryLoop int64
//
//	for {
//		rows, err := d.db.Query(ctx, sql4ContestOfCanGuess, startTime, startIndex, limit4EveryFetch)
//		if err != nil {
//			return
//		}
//		for rows.Next() {
//			count4EveryLoop++
//
//			var contestID int64
//			if err = rows.Scan(&contestID); err == nil {
//				list = append(list, contestID)
//			}
//		}
//
//		if err = rows.Close();err!=nil{
//			log.Error("ContestRecent rows.Err() error(%v)", err)
//			return
//		}
//		if count4EveryLoop < limit4EveryFetch {
//			break
//		}
//
//		startIndex = startIndex + limit4EveryFetch
//		count4EveryLoop = 0
//	}
//
//	return
//}

// ContestData get contest by id.
func (d *Dao) ContestData(c context.Context, cid int64) (res []*model.ContestsData, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _contestDataSQL, cid); err != nil {
		log.Error("ContestsData: db.Exec(%s) error(%v)", _contestDataSQL, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		data := new(model.ContestsData)
		if err = rows.Scan(&data.ID, &data.Cid, &data.URL, &data.PointData); err != nil {
			log.Error("ContestsData:row.Scan() error(%v)", err)
			return
		}
		res = append(res, data)
	}
	if err = rows.Err(); err != nil {
		log.Error("ContestssData rows.Err() error(%v)", err)
	}
	return
}

// RawEpContests get contests by ids.
func (d *Dao) RawEpContests(c context.Context, cids []int64) (res map[int64]*model.Contest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_contestsSQL, xstr.JoinInts(cids))); err != nil {
		log.Error("RawEpContests: db.Exec(%s) error(%v)", xstr.JoinInts(cids), err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*model.Contest, len(cids))
	for rows.Next() {
		r := new(model.Contest)
		if err = rows.Scan(&r.ID, &r.GameStage, &r.Stime, &r.Etime, &r.HomeID, &r.AwayID, &r.HomeScore, &r.AwayScore,
			&r.LiveRoom, &r.Aid, &r.Collection, &r.GameState, &r.Dic, &r.Ctime, &r.Mtime, &r.Status, &r.Sid, &r.Mid,
			&r.Special, &r.SpecialName, &r.SpecialTips, &r.SuccessTeam, &r.SpecialImage, &r.Playback, &r.CollectionURL,
			&r.LiveURL, &r.DataType, &r.MatchID, &r.GuessType, &r.GameStage1, &r.GameStage2, &r.PushSwitch,
			&r.SeriesID, &r.ContestStatus); err != nil {
			log.Error("RawEpContests:row.Scan() error(%v)", err)
			return
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("RawEpContests rows.Err() error(%v)", err)
	}
	return
}

// S9Contests get s9 contests.
func (d *Dao) S9Contests(c context.Context, sid, stime, etime int64) (res []*model.Contest, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _s9contestsSQL, sid, stime, etime); err != nil {
		log.Error("S9Contests: db.Exec(%s) sid(%d) stime(%d) etime(%d) error(%v)", _s9contestsSQL, sid, stime, etime, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Contest)
		if err = rows.Scan(&r.ID, &r.GameStage, &r.Stime, &r.Etime, &r.HomeID, &r.AwayID, &r.HomeScore, &r.AwayScore,
			&r.LiveRoom, &r.Aid, &r.Collection, &r.GameState, &r.Dic, &r.Ctime, &r.Mtime, &r.Status, &r.Sid, &r.Mid,
			&r.Special, &r.SpecialName, &r.SpecialTips, &r.SuccessTeam, &r.SpecialImage, &r.Playback, &r.CollectionURL,
			&r.LiveURL, &r.DataType, &r.MatchID, &r.GuessType, &r.ContestStatus); err != nil {
			log.Error("S9Contests:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("S9Contests rows.Err() error(%v)", err)
	}
	return
}

// RawEpSeasons get seasons by ids.
func (d *Dao) RawEpSeasons(c context.Context, sids []int64) (res map[int64]*model.Season, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_epSeasonSQL, xstr.JoinInts(sids))); err != nil {
		log.Error("RawEpSeasons: db.Exec(%s) error(%v)", xstr.JoinInts(sids), err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*model.Season, len(sids))
	for rows.Next() {
		r := new(model.Season)
		if err = rows.Scan(&r.ID, &r.Mid, &r.Title, &r.SubTitle, &r.Stime, &r.Etime, &r.Sponsor, &r.Logo, &r.Dic, &r.Ctime,
			&r.Mtime, &r.Status, &r.Rank, &r.IsApp, &r.URL, &r.DataFocus, &r.FocusURL, &r.LeidaSID, &r.SearchImage, &r.SyncPlatform); err != nil {
			log.Error("RawEpSeasons:row.Scan() error(%v)", err)
			return
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("RawEpSeasons rows.Err() error(%v)", err)
	}
	return
}

// RawEpTeams get seasons by ids.
func (d *Dao) RawEpTeams(c context.Context, tids []int64) (res map[int64]*model.Team, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_teamsInSQL, xstr.JoinInts(tids))); err != nil {
		log.Error("RawEpTeams: db.Exec(%s) error(%v)", xstr.JoinInts(tids), err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*model.Team, len(tids))
	for rows.Next() {
		r := new(model.Team)
		if err = rows.Scan(
			&r.ID,
			&r.Title,
			&r.SubTitle,
			&r.Logo,
			&r.VideoURL,
			&r.Profile,
			&r.LeidaTID,
			&r.ReplyID,
			&r.TeamType,
			&r.RegionID); err != nil {
			log.Error("RawEpTeams:row.Scan() error(%v)", err)
			return
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("RawEpTeams.Err() error(%v)", err)
	}
	return
}

// LdTeam leida team.
func (d *Dao) LdTeam(c context.Context, ldTeamID int64) (mod *model.Team, err error) {
	mod = &model.Team{}
	row := d.db.QueryRow(c, _teamLdSQL, ldTeamID)
	if err = row.Scan(&mod.ID, &mod.Title, &mod.SubTitle, &mod.Logo, &mod.VideoURL, &mod.Profile, &mod.LeidaTID, &mod.ReplyID); err != nil {
		if err == sql.ErrNoRows {
			mod = nil
			err = nil
		} else {
			err = errors.Wrapf(err, "LdTeam:row.Scan() error")
		}
	}
	return
}

// KDetails knockout detail
func (d *Dao) KDetails(c context.Context) (res []*model.ActiveDetail, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _kDetailsSQL); err != nil {
		log.Error("ActPDetails: db.Exec(%s) error(%v)", _kDetailsSQL, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		detail := new(model.ActiveDetail)
		if err = rows.Scan(&detail.ID, &detail.Maid, &detail.GameType, &detail.STime, &detail.ETime, &detail.Online); err != nil {
			log.Error("KDetails:row.Scan() error(%v)", err)
			return
		}
		res = append(res, detail)
	}
	if err = rows.Err(); err != nil {
		log.Error("KDetails rows.Err() error(%v)", err)
	}
	return
}

// LiveInfo active live.
func (d *Dao) LiveInfo(c context.Context, liveID int64) (mod *model.ActiveLive, err error) {
	mod = &model.ActiveLive{}
	row := d.db.QueryRow(c, _liveSQL, liveID)
	if err = row.Scan(&mod.MaID, &mod.LiveID, &mod.Title); err != nil {
		if err == sql.ErrNoRows {
			mod = nil
			err = nil
		} else {
			log.Error("LiveInfo row.Scan error(%v)", err)
		}
	}
	return
}

// LolPlayers lol players.
func (d *Dao) LolPlayers(c context.Context, sid int64) (res []*model.LolPlayer, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _lolPlayerSQL, sid); err != nil {
		err = errors.Wrapf(err, "LolPlayers:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LolPlayer)
		if err = rows.Scan(&r.ID, &r.PlayerID, &r.TeamID, &r.TeamAcronym, &r.TeamImage, &r.LeidaSID,
			&r.Name, &r.ImageURL, &r.ChampionsImage, &r.Role, &r.Win, &r.KDA, &r.Kills, &r.Deaths,
			&r.Assists, &r.MinionsKilled, &r.WardsPlaced, &r.GamesCount, &r.Ctime, &r.Mtime,
			&r.PositionID, &r.Position, &r.MVP); err != nil {
			err = errors.Wrapf(err, "LolPlayers:row.Scan() error")
			return
		}
		r.RoleName = lolRole[r.Role]
		r.KDA = decimal(r.KDA, _decimalOne)
		r.Kills = decimal(r.Kills, _decimalOne)
		r.Deaths = decimal(r.Deaths, _decimalOne)
		r.Assists = decimal(r.Assists, _decimalOne)
		r.MinionsKilled = decimal(r.MinionsKilled, _decimalOne)
		r.WardsPlaced = decimal(r.WardsPlaced, _decimalOne)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolPlayers:rows.Err() error")
	}
	return
}

// LolTeams lol teams.
func (d *Dao) LolTeams(c context.Context, sid int64) (res []*model.LolTeam, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _lolTeamSQL, sid); err != nil {
		err = errors.Wrapf(err, "LolTeams:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LolTeam)
		if err = rows.Scan(&r.ID, &r.TeamID, &r.LeidaSID, &r.Name, &r.Acronym, &r.ImageURL, &r.Win, &r.KDA, &r.Kills, &r.Deaths, &r.Assists, &r.TowerKills,
			&r.TotalMinionsKilled, &r.FirstTower, &r.FirstInhibitor, &r.FirstDragon, &r.FirstBaron, &r.FirstBlood, &r.WardsPlaced, &r.InhibitorKills,
			&r.BaronKills, &r.GoldEarned, &r.GamesCount, &r.Players, &r.Ctime, &r.Mtime, &r.BaronRate, &r.DragonRate, &r.Hits, &r.LoseNum, &r.Money,
			&r.TotalDamage, &r.WinNum, &r.ImageThumb, &r.NewData); err != nil {
			err = errors.Wrapf(err, "LolTeams:row.Scan() error")
			return
		}
		r.KDA = decimal(r.KDA, _decimalOne)
		r.Kills = decimal(r.Kills, _decimalOne)
		r.Deaths = decimal(r.Deaths, _decimalOne)
		r.Assists = decimal(r.Assists, _decimalOne)
		r.TowerKills = decimal(r.TowerKills, _decimalOne)
		r.TotalMinionsKilled = decimal(r.TotalMinionsKilled, _decimalOne)
		r.FirstTower = decimal(r.FirstTower, _decimalOne)
		r.FirstInhibitor = decimal(r.FirstInhibitor, _decimalOne)
		r.FirstDragon = decimal(r.FirstDragon, _decimalOne)
		r.FirstBaron = decimal(r.FirstBaron, _decimalOne)
		r.FirstBlood = decimal(r.FirstBlood, _decimalOne)
		r.WardsPlaced = decimal(r.WardsPlaced, _decimalOne)
		r.InhibitorKills = decimal(r.InhibitorKills, _decimalOne)
		r.BaronKills = decimal(r.BaronKills, _decimalOne)
		r.GoldEarned = decimal(r.GoldEarned, _decimalOne)
		r.BaronRate = decimal(r.BaronRate, _decimalOne)
		r.DragonRate = decimal(r.DragonRate, _decimalOne)
		r.Hits = decimal(r.Hits, _decimalOne)
		r.Money = decimal(r.Money, _decimalOne)
		r.TotalDamage = decimal(r.TotalDamage, _decimalOne)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolTeams:rows.Err() error")
	}
	return
}

// DotaPlayers dota players.
func (d *Dao) DotaPlayers(c context.Context, sid int64) (res []*model.DotaPlayer, err error) {
	var (
		rows  *xsql.Rows
		roles []string
	)
	if rows, err = d.db.Query(c, _dotaPlayerSQL, sid); err != nil {
		err = errors.Wrapf(err, "DotaPlayers:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.DotaPlayer)
		if err = rows.Scan(&r.ID, &r.PlayerID, &r.TeamID, &r.TeamAcronym, &r.TeamImage, &r.LeidaSID, &r.Name, &r.ImageURL, &r.HeroesImage, &r.Role,
			&r.Win, &r.KDA, &r.Kills, &r.Deaths, &r.Assists, &r.WardsPlaced, &r.LastHits, &r.ObserverWardsPlaced, &r.SentryWardsPlaced,
			&r.XpPerMinute, &r.GoldPerMinute, &r.GamesCount, &r.Ctime, &r.Mtime); err != nil {
			err = errors.Wrapf(err, "DotaPlayers:row.Scan() error")
			return
		}
		roles = nil
		roleIDs := strings.Split(r.Role, "/")
		for _, roleID := range roleIDs {
			roles = append(roles, dotaRole[roleID])
		}
		r.RoleName = strings.Join(roles, "/")
		r.KDA = decimal(r.KDA, _decimalOne)
		r.Kills = decimal(r.Kills, _decimalOne)
		r.Deaths = decimal(r.Deaths, _decimalOne)
		r.Assists = decimal(r.Assists, _decimalOne)
		r.WardsPlaced = decimal(r.WardsPlaced, _decimalOne)
		r.LastHits = decimal(r.LastHits, _decimalOne)
		r.ObserverWardsPlaced = decimal(r.ObserverWardsPlaced, _decimalOne)
		r.SentryWardsPlaced = decimal(r.SentryWardsPlaced, _decimalOne)
		r.XpPerMinute = decimal(r.XpPerMinute, _decimalOne)
		r.GoldPerMinute = decimal(r.GoldPerMinute, _decimalOne)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaPlayers:rows.Err() error")
	}
	return
}

// DotaTeams dota teams.
func (d *Dao) DotaTeams(c context.Context, sid int64) (res []*model.DotaTeam, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _dotaTeamSQL, sid); err != nil {
		err = errors.Wrapf(err, "DotaTeams:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.DotaTeam)
		if err = rows.Scan(&r.ID, &r.TeamID, &r.LeidaSID, &r.Name, &r.Acronym, &r.ImageURL, &r.Win, &r.KDA, &r.Kills, &r.Deaths, &r.Assists, &r.TowerKills,
			&r.LastHits, &r.ObserverUsed, &r.SentryUsed, &r.XpPerMinute, &r.FirstBlood, &r.Heal, &r.GoldSpent, &r.GoldPerMin, &r.Denies, &r.DamageTaken,
			&r.CampsStacked, &r.GamesCount, &r.Players, &r.Ctime, &r.Mtime); err != nil {
			err = errors.Wrapf(err, "DotaTeams:row.Scan() error")
			return
		}
		r.KDA = decimal(r.KDA, _decimalOne)
		r.Kills = decimal(r.Kills, _decimalOne)
		r.Deaths = decimal(r.Deaths, _decimalOne)
		r.Assists = decimal(r.Assists, _decimalOne)
		r.TowerKills = decimal(r.TowerKills, _decimalOne)
		r.LastHits = decimal(r.LastHits, _decimalOne)
		r.ObserverUsed = decimal(r.ObserverUsed, _decimalOne)
		r.SentryUsed = decimal(r.SentryUsed, _decimalOne)
		r.XpPerMinute = decimal(r.XpPerMinute, _decimalOne)
		r.FirstBlood = decimal(r.FirstBlood, _decimalOne)
		r.Heal = decimal(r.Heal, _decimalOne)
		r.GoldSpent = decimal(r.GoldSpent, _decimalOne)
		r.GoldPerMin = decimal(r.GoldPerMin, _decimalOne)
		r.Denies = decimal(r.Denies, _decimalOne)
		r.DamageTaken = decimal(r.DamageTaken, _decimalOne)
		r.CampsStacked = decimal(r.CampsStacked, _decimalOne)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaTeams:rows.Err() error")
	}
	return
}

// GameSeason game season  list.
func (d *Dao) GameSeason(c context.Context, gameID int64) (res []*model.Season, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _gameSeasonSQL, gameID); err != nil {
		err = errors.Wrapf(err, "GameSeason:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Season)
		if err = rows.Scan(&r.ID, &r.Mid, &r.Title, &r.SubTitle, &r.Stime, &r.Etime, &r.Sponsor, &r.Logo, &r.Dic, &r.Ctime,
			&r.Mtime, &r.Status, &r.Rank, &r.IsApp, &r.URL, &r.DataFocus, &r.FocusURL, &r.LeidaSID); err != nil {
			err = errors.Wrapf(err, "Contest:row.Scan() error")
			return
		}
		r.GameType = d.gameType(gameID)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rows.Err() error")
	}
	return
}

// RawLolGames lol game list.
func (d *Dao) RawLolGames(c context.Context, matchID int64) (res []*model.LolGame, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _lolGameSQL, matchID); err != nil {
		err = errors.Wrapf(err, "LolGames:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LolGame)
		if err = rows.Scan(&r.ID, &r.GameID, &r.Teams, &r.Players, &r.Position, &r.MatchID, &r.BeginAt, &r.EndAt, &r.Finished); err != nil {
			err = errors.Wrapf(err, "LolGames:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolGames:rows.Err() error")
	}
	return
}

// RawDotaGames dota game list.
func (d *Dao) RawDotaGames(c context.Context, matchID int64) (res []*model.LolGame, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _dotaGameSQL, matchID); err != nil {
		err = errors.Wrapf(err, "DotaGames:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LolGame)
		if err = rows.Scan(&r.ID, &r.GameID, &r.Teams, &r.Players, &r.Position, &r.MatchID, &r.BeginAt, &r.EndAt, &r.Finished); err != nil {
			err = errors.Wrapf(err, "DotaGames:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaGames:rows.Err() error")
	}
	return
}

// RawOwGames overwatch game list.
func (d *Dao) RawOwGames(c context.Context, matchID int64) (res []*model.OwGame, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _owGameSQL, matchID); err != nil {
		err = errors.Wrapf(err, "OwGames:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.OwGame)
		if err = rows.Scan(&r.ID, &r.GameID, &r.WinTeam, &r.Teams, &r.Map, &r.Position, &r.MatchID, &r.BeginAt, &r.EndAt, &r.Finished); err != nil {
			err = errors.Wrapf(err, "OwGames:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "OwGames:rows.Err() error")
	}
	return
}

// LolItems  lol item list.
func (d *Dao) LolItems(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _lolItemSQL); err != nil {
		err = errors.Wrapf(err, "LolItems:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "LolItems:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolItems:rows.Err() error")
	}
	return
}

// DotaItems  dota item list.
func (d *Dao) DotaItems(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _dotaItemSQL); err != nil {
		err = errors.Wrapf(err, "DotaItems:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "DotaItems:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaItems:rows.Err() error")
	}
	return
}

// OwMaps  overitem map list.
func (d *Dao) OwMaps(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _owMapSQL); err != nil {
		err = errors.Wrapf(err, "OwMaps:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "OwMaps:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "OwMaps:rows.Err() error")
	}
	return
}

// LolSpells  lol spells list.
func (d *Dao) LolSpells(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _lolAbilitySQL); err != nil {
		err = errors.Wrapf(err, "LolSpells:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "LolSpells:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolSpells:rows.Err() error")
	}
	return
}

// DotaAbility  dota ability list.
func (d *Dao) DotaAbility(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _dotaAbilitySQL); err != nil {
		err = errors.Wrapf(err, "DotaAbility:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "DotaAbility:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaAbility:rows.Err() error")
	}
	return
}

// LolCham lol Champions list.
func (d *Dao) LolCham(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _lolchamSQL); err != nil {
		err = errors.Wrapf(err, "LolCham:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "LolCham:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolCham:rows.Err() error")
	}
	return
}

// DotaHero dota heroes list.
func (d *Dao) DotaHero(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _dotaHeroSQL); err != nil {
		err = errors.Wrapf(err, "DotaHero:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "DotaHero:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaHero:rows.Err() error")
	}
	return
}

// OwHero ow heroes list.
func (d *Dao) OwHero(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _owHeroSQL); err != nil {
		err = errors.Wrapf(err, "OwHero:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "OwHero:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "OwHero:rows.Err() error")
	}
	return
}

// LolMatchPlayer lol players list.
func (d *Dao) LolMatchPlayer(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _lolPlaysSQL); err != nil {
		err = errors.Wrapf(err, "LolMatchPlayer:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "LolMatchPlayer:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolMatchPlayer:rows.Err() error")
	}
	return
}

// DotaMatchPlayer dota players list.
func (d *Dao) DotaMatchPlayer(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _dotaPlaysSQL); err != nil {
		err = errors.Wrapf(err, "DotaMatchPlayer:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "DotaMatchPlayer:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaMatchPlayer:rows.Err() error")
	}
	return
}

// OwMatchPlayer overplay players list.
func (d *Dao) OwMatchPlayer(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _owPlaysSQL); err != nil {
		err = errors.Wrapf(err, "OwMatchPlayer:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "OwMatchPlayer:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "OwMatchPlayer:rows.Err() error")
	}
	return
}

// LolMatchTeam lol team list.
func (d *Dao) LolMatchTeam(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _lolTeamsSQL); err != nil {
		err = errors.Wrapf(err, "LolMatchTeam:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "LolMatchTeam:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "LolMatchTeam:rows.Err() error")
	}
	return
}

// DotaMatchTeam dota teams list.
func (d *Dao) DotaMatchTeam(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _dotaTeamsSQL); err != nil {
		err = errors.Wrapf(err, "DotaMatchTeam:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "DotaMatchTeam:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "DotaMatchTeam:rows.Err() error")
	}
	return
}

// OwMatchTeam overplay teams list.
func (d *Dao) OwMatchTeam(c context.Context) (res []*model.LdInfo, err error) {
	var rows *xsql.Rows

	if rows, err = d.db.Query(c, _owTeamsSQL); err != nil {
		err = errors.Wrapf(err, "OwMatchTeam:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.LdInfo)
		if err = rows.Scan(&r.ID, &r.Name, &r.ImageURL); err != nil {
			err = errors.Wrapf(err, "OwMatchTeam:row.Scan() error")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "OwMatchTeam:rows.Err() error")
	}
	return
}

func (d *Dao) gameType(gid int64) (rs int64) {
	for _, tp := range d.c.GameTypes {
		if tp.DbGameID == gid {
			rs = tp.ID
			break
		}
	}
	return
}

// GuessCollGame guess collection game
func (d *Dao) GuessCollGame(c context.Context) (res []int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _guessGameSQL, time.Now().Unix()); err != nil {
		err = errors.Wrapf(err, "GuessCollGame:d.db.Query(%s) error(%v)", _guessGameSQL, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var gid int64
		if err = rows.Scan(&gid); err != nil {
			err = errors.Wrapf(err, "GuessCollGame:row.Scan error(%v)", err)
			return
		}
		res = append(res, gid)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GuessCollGame:rows.Err error(%v)", err)
		return
	}
	return
}

// GuessCollSeason guess collection season
func (d *Dao) GuessCollSeason(c context.Context, gid int64) (res []int64, err error) {
	var rows *xsql.Rows
	t := time.Now().Unix()
	if gid == 0 {
		if rows, err = d.db.Query(c, _guessSeasonAllSQL, t); err != nil {
			err = errors.Wrapf(err, "GuessCollSeason:d.db.Query(%s) Time(%d) Gid(%d) error(%v)", _guessSeasonAllSQL, t, gid, err)
			return
		}
	} else {
		if rows, err = d.db.Query(c, _guessSeasonSQL, t, gid); err != nil {
			err = errors.Wrapf(err, "GuessCollSeason:d.db.Query(%s) Time(%d) Gid(%d) error(%v)", _guessSeasonSQL, t, gid, err)
			return
		}
	}
	defer rows.Close()
	for rows.Next() {
		var sid int64
		if err = rows.Scan(&sid); err != nil {
			err = errors.Wrapf(err, "GuessCollSeason:row.Scan error(%v)", err)
			return
		}
		res = append(res, sid)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GuessCollSeason:rows.Err error(%v)", err)
		return
	}
	return
}

// GuessCCalen guess game collection calendar
func (d *Dao) GuessCCalen(c context.Context) (res []*model.Calendar, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _guessCCalenSQL, time.Now().Unix()); err != nil {
		err = errors.Wrapf(err, "GuessCCalen:d.db.Query(%s) error(%v)", _guessCCalenSQL, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		cal := &model.Calendar{}
		if err = rows.Scan(&cal.Stime, &cal.Count); err != nil {
			err = errors.Wrapf(err, "GuessCCalen:row.Scan error(%v)", err)
			return
		}
		res = append(res, cal)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "GuessCCalen:rows.Err error(%v)", err)
		return
	}
	return
}

// decimal .
func decimal(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

// RawSearchMainIDs get search main ids.
func (d *Dao) RawSearchMainIDs(c context.Context) (rs []int64, err error) {
	var (
		rows   *xsql.Rows
		mainID int64
	)
	rows, err = d.db.Query(c, _searchMainSQL)
	if err != nil {
		err = errors.Wrap(err, "RawSearchMainIDs:d.db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&mainID); err != nil {
			err = errors.Wrap(err, "RawSearchMainIDs:rows.Scan() error")
			return
		}
		rs = append(rs, mainID)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawSearchMainIDs:rows.Err")
	}
	return
}

// RawSearchMD search card lists.
func (d *Dao) RawSearchMD(c context.Context, ids []int64) (res map[int64]*model.SearchRes, err error) {
	var (
		rows *xsql.Rows
		rs   []*model.SearchMD
		item *model.SearchRes
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_searchCardSQL, xstr.JoinInts(ids))); err != nil {
		log.Error("RawSearchMD:d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.SearchMD)
		if err = rows.Scan(&r.ID, &r.QueryName, &r.Stime, &r.Etime, &r.Cid); err != nil {
			err = errors.Wrap(err, "RawSearchMD:rows.Scan() error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawSearchMD:rows.Err")
		return
	}
	res = make(map[int64]*model.SearchRes, len(ids))
	for _, v := range rs {
		if _, ok := res[v.ID]; !ok {
			item = &model.SearchRes{
				SearchMain: &model.SearchMain{
					ID:        v.ID,
					QueryName: v.QueryName,
					Stime:     v.Stime,
					Etime:     v.Etime,
				},
				ContestIDs: []int64{
					v.Cid,
				},
			}
			res[v.ID] = item
			continue
		}
		res[v.ID].ContestIDs = append(res[v.ID].ContestIDs, v.Cid)
	}
	return
}

// RawEpGames get game by ids.
func (d *Dao) RawEpGames(c context.Context, gids []int64) (res map[int64]*mdlEp.Game, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_epGameSQL, xstr.JoinInts(gids))); err != nil {
		log.Error("RawEpGames: db.Exec(%s) error(%v)", xstr.JoinInts(gids), err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*mdlEp.Game, len(gids))
	for rows.Next() {
		r := new(mdlEp.Game)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle, &r.ETitle, &r.Plat, &r.GameType, &r.Logo, &r.Publisher, &r.Operations, &r.PbTime, &r.Dic); err != nil {
			log.Error("RawEpGames:row.Scan() error(%v)", err)
			return
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) RawEpGameMap(c context.Context, oids []int64, tp int64) (res map[int64]int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_epGameMapSQL, xstr.JoinInts(oids)), tp); err != nil {
		log.Error("RawEpGames: db.Exec tp(%d) oids(%s) error(%v)", tp, xstr.JoinInts(oids), err)
		return
	}
	defer rows.Close()
	res = make(map[int64]int64, len(oids))
	for rows.Next() {
		var oid, gid int64
		if err = rows.Scan(&oid, &gid); err != nil {
			log.Error("RawEpGames:row.Scan() error(%v)", err)
			return
		}
		res[oid] = gid
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// RawSeasonGames .
func (d *Dao) RawSeasonGames(c context.Context) (ids []int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _selSeasonGidSQL); err != nil {
		log.Error("RawSeasonGames: db.Exec() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var gid int64
		if err = rows.Scan(&gid); err != nil {
			log.Error("RawSeasonGames:row.Scan() error(%v)", err)
			return
		}
		ids = append(ids, gid)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// RawH5Games game rank list.
func (d *Dao) RawH5Games(c context.Context, gids []int64) (res []*model.GameRank, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_selGameRankSQL, xstr.JoinInts(gids))); err != nil {
		log.Error("RawH5Games: db.Exec(%s) error(%v)", xstr.JoinInts(gids), err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.GameRank)
		if err = rows.Scan(&r.ID, &r.Title, &r.SubTitle, &r.Rank); err != nil {
			log.Error("RawH5Games:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// SeasonRank .
func (d *Dao) SeasonRank(c context.Context, gid int64) (res []*model.SeasonRank, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _selSeasonRankSQL, gid); err != nil {
		log.Error("SeasonRank: db.Exec() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.SeasonRank)
		if err = rows.Scan(&r.ID, &r.Sid, &r.Rank); err != nil {
			log.Error("SeasonRank:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) EffectiveTeamList(ctx context.Context, startID int64) (list []*model.Team, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(ctx, sql4EffectiveTeamList, startID); err != nil {
		log.Errorc(ctx, "EffectiveTeamList d.db.Query error(%+v)", err)
		return
	}

	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "EffectiveTeamList rows.Err() error(%v)", err)
		}
	}()
	list = make([]*model.Team, 0)
	for rows.Next() {
		r := new(model.Team)
		err = rows.Scan(
			&r.ID,
			&r.Title,
			&r.SubTitle,
			&r.Logo,
			&r.VideoURL,
			&r.Profile,
			&r.LeidaTID,
			&r.ReplyID,
			&r.TeamType,
			&r.RegionID)
		if err != nil {
			return
		}
		list = append(list, r)
	}
	return
}

func (d *Dao) EffectiveSeasonList(ctx context.Context, startID int64) (list []*model.Season, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(ctx, sql4EffectiveSeasonList, startID); err != nil {
		log.Errorc(ctx, "EffectiveSeasonList d.db.Query error(%+v)", err)
		return
	}

	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "EffectiveSeasonList rows.Err() error(%v)", err)
		}
	}()
	list = make([]*model.Season, 0)
	for rows.Next() {
		r := new(model.Season)
		if err = rows.Scan(&r.ID, &r.Mid, &r.Title, &r.SubTitle, &r.Stime, &r.Etime, &r.Sponsor, &r.Logo, &r.Dic, &r.Ctime,
			&r.Mtime, &r.Status, &r.Rank, &r.IsApp, &r.URL, &r.DataFocus, &r.FocusURL, &r.LeidaSID, &r.SearchImage, &r.SyncPlatform); err != nil {
			return
		}
		list = append(list, r)
	}
	return
}

func (d *Dao) GetTeamsInSeasonFromDB(c context.Context, seasonIds []int64) (map[int64] /*seasonId*/ []*model.TeamInSeason, error) {
	res := make(map[int64][]*model.TeamInSeason, 0)
	if len(seasonIds) == 0 {
		return res, nil
	}
	if tool.IsLimiterAllowedByUniqBizKey(_limitKey2FetchTeamsInSeason, _limitKey2FetchTeamsInSeason) {
		rows, err := d.db.Query(c, fmt.Sprintf(_sqlTeamsInSeasonSQL, xstr.JoinInts(seasonIds)))
		if err != nil {
			log.Errorc(c, "query teams_in_season from db error: %v", err)
			return res, err
		}
		defer func() {
			_ = rows.Close()
			if err = rows.Err(); err != nil {
				log.Errorc(c, "GetTeamsInSeasonFromDB rows error: %v", err)
				return
			}
		}()
		for rows.Next() {
			tmp := &model.TeamInSeason{}
			if err = rows.Scan(&tmp.TeamId, &tmp.TeamTitle, &tmp.RegionId, &tmp.SeasonId, &tmp.Rank, &tmp.Logo, &tmp.Uid, &tmp.LeidaID); err != nil {
				log.Errorc(c, "scan teams_in_season error: %v", err)
				return res, err
			}
			if _, ok := res[tmp.SeasonId]; !ok {
				res[tmp.SeasonId] = make([]*model.TeamInSeason, 0)
			}
			res[tmp.SeasonId] = append(res[tmp.SeasonId], tmp)
		}
		return res, nil
	} else {
		return nil, xecode.LimitExceed
	}
}

func (d *Dao) GetOngoingSeasonIDFromDB(c context.Context) ([]int64, error) {
	ctx := context.Background()
	res := make([]int64, 0)
	rows, err := d.db.Query(ctx, _sqlOngoingSeasonSQL, time.Now().Unix())
	if err != nil {
		log.Errorc(c, "get ongoing season from db error: %v", err)
		return res, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Errorc(ctx, "scan ongoing season id error: %v", err)
			continue
		}
		if id != 0 {
			res = append(res, id)
		}
	}
	if err := rows.Err(); err != nil {
		log.Errorc(c, "rows error: %v", err)
		return res, err
	}
	return res, nil

}

func (d *Dao) MatchSeasonsFromDB(c context.Context, matchIDs []int64) (res map[int64][]*model.MatchSeason, err error) {
	var (
		rows *xsql.Rows
	)
	res = make(map[int64][]*model.MatchSeason)
	if len(matchIDs) == 0 {
		return
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_sqlMatchSeasonSQL, xstr.JoinInts(matchIDs))); err != nil {
		log.Errorc(c, "MatchSeasonsFromDB query from db error: %v", err)
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(c, "MatchSeasonsFromDB rows error: %v", err)
			return
		}
	}()
	for rows.Next() {
		tmp := &model.MatchSeason{}
		if err = rows.Scan(&tmp.SeasonID, &tmp.SeasonTitle, &tmp.Logo, &tmp.MatchID, &tmp.Stime, &tmp.Etime); err != nil {
			log.Errorc(c, "MatchSeasonsFromDB scan error: %v", err)
			return
		}
		if _, ok := res[tmp.MatchID]; !ok {
			res[tmp.MatchID] = make([]*model.MatchSeason, 0)
		}
		res[tmp.MatchID] = append(res[tmp.MatchID], tmp)
	}
	return
}

func (d *Dao) RawFetchSeasonsByMatchId(c context.Context, matchId int64) (res []*model.MatchSeason, err error) {
	if tool.IsLimiterAllowedByUniqBizKey(_limitKey2FetchSeasonsByMatch, _limitKey2FetchSeasonsByMatch) {
		res, err = d.FetchSeasonsByMatchIdFromDB(c, matchId)
		if err != nil {
			log.Errorc(c, "RawFetchSeasonsByMatchId d.FetchSeasonsByMatchIdFromDB() matchID(%d) error(%+v)", matchId, err)
			return
		}
	} else {
		err = xecode.LimitExceed
		return
	}
	return
}

func (d *Dao) FetchSeasonsByMatchIdFromDB(c context.Context, matchId int64) ([]*model.MatchSeason, error) {
	res := make([]*model.MatchSeason, 0)
	rows, err := d.db.Query(c, _sqlSeasonsByMatchSQL, matchId)
	if err != nil {
		log.Errorc(c, "SeasonsByMatchIdFromDB query from db error: %v", err)
		return res, err
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(c, "SeasonsByMatchIdFromDB rows error: %v", err)
			return
		}
	}()
	for rows.Next() {
		tmp := &model.MatchSeason{}
		if err := rows.Scan(&tmp.SeasonID, &tmp.SeasonTitle, &tmp.Logo, &tmp.MatchID, &tmp.Stime, &tmp.Etime); err != nil {
			log.Errorc(c, "SeasonsByMatchIdFromDB scan error: %v", err)
			return res, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

func (d *Dao) RawFetchSeasonsInfoMap(c context.Context, sids []int64) (res map[int64]*model.MatchSeason, err error) {
	var (
		rows *xsql.Rows
	)
	res = make(map[int64]*model.MatchSeason)
	if len(sids) == 0 {
		return res, nil
	}
	if tool.IsLimiterAllowedByUniqBizKey(_limitKey2BatchSeasonsTeams, _limitKey2BatchSeasonsTeams) {
		if rows, err = d.db.Query(c, fmt.Sprintf(_sqlBatchSeasonsSQL, xstr.JoinInts(sids))); err != nil {
			log.Errorc(c, "MatchSeasonsFromDB query from db error: %v", err)
			return
		}
		defer func() {
			_ = rows.Close()
			if err = rows.Err(); err != nil {
				log.Errorc(c, "MatchSeasonsFromDB rows error: %v", err)
				return
			}
		}()
		for rows.Next() {
			tmp := &model.MatchSeason{}
			if err := rows.Scan(&tmp.SeasonID, &tmp.SeasonTitle, &tmp.Logo, &tmp.MatchID, &tmp.Stime, &tmp.Etime); err != nil {
				log.Errorc(c, "MatchSeasonsFromDB scan error: %v", err)
				return res, err
			}
			res[tmp.SeasonID] = tmp
		}
		return
	} else {
		err = xecode.LimitExceed
	}
	return
}

const _videoListComponent = `SELECT id,ugc_aids,game_id,match_id,year_id FROM es_video_lists WHERE is_deleted=0 and id=? `

// RawVideoList .
func (d *Dao) RawVideoList(ctx context.Context, id int64) (res *model.VideoListInfo, err error) {
	res = &model.VideoListInfo{}
	row := d.db.QueryRow(ctx, _videoListComponent, id)
	if err = row.Scan(&res.ID, &res.UgcAids, &res.GameID, &res.MatchID, &res.YearID); err != nil {
		if err == sql.ErrNoRows {
			res = nil
			err = nil
		} else {
			log.Errorc(ctx, "contest component RawVideoList id(%d) error(%+v)", id, err)
		}
	}
	return
}

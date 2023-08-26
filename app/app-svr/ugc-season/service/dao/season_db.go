package dao

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_seasonSQL       = "SELECT season_id,title,`desc`,cover,mid,attribute,sign_state,first_aid,ptime,ep_count,ep_num FROM season WHERE season_id=?"
	_seasonsSQL      = "SELECT season_id,title,`desc`,cover,mid,attribute,sign_state,first_aid,ptime,ep_count,ep_num FROM season WHERE season_id IN (%s)"
	_secSQL          = "SELECT section_id,season_id,title,`type` FROM season_section WHERE season_id=? ORDER BY `order` ASC"
	_epSQL           = "SELECT episode_id,section_id,season_id,title,aid,cid,attribute FROM season_episode WHERE season_id=? ORDER BY `order` ASC"
	_secsSQL         = "SELECT section_id,season_id,title,`type` FROM season_section WHERE season_id IN (%s) ORDER BY `order` ASC"
	_epsSQL          = "SELECT episode_id,section_id,season_id,title,aid,cid,attribute FROM season_episode WHERE season_id IN (%s) ORDER BY `order` ASC"
	_seasonStatSQL   = "SELECT season_id,fav,`share`,reply,coin,dm,click,likes,mtime FROM season_stat WHERE season_id=?"
	_seasonStatsSQL  = "SELECT season_id,fav,`share`,reply,coin,dm,click,likes,mtime FROM season_stat WHERE season_id IN (%s)"
	_upperSeasonsSQL = "SELECT season_id,ptime FROM season WHERE mid=? ORDER BY ptime DESC"
)

// SeasonInfo get a season by sid.
func (d *Dao) SeasonInfo(c context.Context, sid int64) (s *api.Season, err error) {
	row := d.season.QueryRow(c, _seasonSQL, sid)
	s = &api.Season{}
	if err = row.Scan(&s.ID, &s.Title, &s.Intro, &s.Cover, &s.Mid, &s.Attribute, &s.SignState, &s.FirstAid, &s.Ptime, &s.EpCount, &s.EpNum); err != nil {
		if err == sql.ErrNoRows {
			s = nil
			err = nil
		} else {
			log.Error("row.Scan error(%+v)", err)
		}
		return
	}
	return
}

// SeasonsInfo get seasons by sids.
func (d *Dao) SeasonsInfo(c context.Context, sids []int64) (ss map[int64]*api.Season, err error) {
	ss = make(map[int64]*api.Season, len(sids))
	var rows *xsql.Rows
	if rows, err = d.season.Query(c, fmt.Sprintf(_seasonsSQL, xstr.JoinInts(sids))); err != nil {
		log.Error("d.season.Query(%s) error(%+v)", fmt.Sprintf(_seasonsSQL, xstr.JoinInts(sids)), err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		s := &api.Season{}
		if err = rows.Scan(&s.ID, &s.Title, &s.Intro, &s.Cover, &s.Mid, &s.Attribute, &s.SignState, &s.FirstAid, &s.Ptime, &s.EpCount, &s.EpNum); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return
		}
		ss[s.ID] = s
	}
	err = rows.Err()
	return
}

// SectionsInfo get a season_sec by sid.
func (d *Dao) SectionsInfo(c context.Context, sid int64) (res []*api.Section, err error) {
	rows, err := d.season.Query(c, _secSQL, sid)
	if err != nil {
		log.Error("d.resultDB.Query seasonSec(%d) error(%+v)", sid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		ss := &api.Section{}
		if err = rows.Scan(&ss.ID, &ss.SeasonID, &ss.Title, &ss.Type); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return
		}
		res = append(res, ss)
	}
	err = rows.Err()
	return
}

// EpisodesInfo get season_ep by sid.
func (d *Dao) EpisodesInfo(c context.Context, sid int64) (res map[int64][]*api.Episode, err error) {
	rows, err := d.season.Query(c, _epSQL, sid)
	if err != nil {
		log.Error("d.resultDB.Query seasonEp(%d) error(%+v)", sid, err)
		return
	}
	defer rows.Close()
	res = make(map[int64][]*api.Episode)
	for rows.Next() {
		se := &api.Episode{}
		if err = rows.Scan(&se.ID, &se.SectionID, &se.SeasonID, &se.Title, &se.Aid, &se.Cid, &se.Attribute); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return
		}
		res[se.SectionID] = append(res[se.SectionID], se)
	}
	err = rows.Err()
	return
}

// StatInfo get season stat.
func (d *Dao) StatInfo(c context.Context, sid int64) (st *api.Stat, err error) {
	row := d.stat.QueryRow(c, _seasonStatSQL, sid)
	st = &api.Stat{}
	if err = row.Scan(&st.SeasonID, &st.Fav, &st.Share, &st.Reply, &st.Coin, &st.Danmaku, &st.View, &st.Like, &st.Mtime); err != nil {
		if err == sql.ErrNoRows {
			st = nil
			err = nil
		} else {
			log.Error("row.Scan error(%+v)", err)
		}
		return
	}
	return
}

// StatsInfo archive stats.
func (d *Dao) StatsInfo(c context.Context, sids []int64) (sts map[int64]*api.Stat, err error) {
	sts = make(map[int64]*api.Stat, len(sids))
	var rows *xsql.Rows
	if rows, err = d.stat.Query(c, fmt.Sprintf(_seasonStatsSQL, xstr.JoinInts(sids))); err != nil {
		log.Error("d.seasonStats.Query(%s) error(%+v)", fmt.Sprintf(_seasonStatsSQL, xstr.JoinInts(sids)), err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		st := &api.Stat{}
		if err = rows.Scan(&st.SeasonID, &st.Fav, &st.Share, &st.Reply, &st.Coin, &st.Danmaku, &st.View, &st.Like); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return
		}
		sts[st.SeasonID] = st
	}
	err = rows.Err()
	return
}

// UpperSeasonInfo is
func (d *Dao) UpperSeasonInfo(c context.Context, mid int64) (sids []int64, ptimes []time.Time, err error) {
	var rows *xsql.Rows
	if rows, err = d.season.Query(c, _upperSeasonsSQL, mid); err != nil {
		log.Error("d.season.Query(%s) mid(%d) error(%+v)", _upperSeasonsSQL, mid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			sid   int64
			ptime time.Time
		)
		if err = rows.Scan(&sid, &ptime); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			return
		}
		sids = append(sids, sid)
		ptimes = append(ptimes, ptime)
	}
	err = rows.Err()
	return
}

// SectionsInfos get a season_sec by sids.
func (d *Dao) SectionsInfos(c context.Context, sids []int64) (map[int64][]*api.Section, error) {
	rows, err := d.season.Query(c, fmt.Sprintf(_secsSQL, xstr.JoinInts(sids)))
	if err != nil {
		log.Error("d.resultDB.Query seasonsSec(%+v) error(%+v)", sids, err)
		return nil, err
	}
	defer rows.Close()
	var res = make(map[int64][]*api.Section, len(sids))
	for rows.Next() {
		ss := &api.Section{}
		if err := rows.Scan(&ss.ID, &ss.SeasonID, &ss.Title, &ss.Type); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			continue
		}
		if ss.SeasonID > 0 {
			res[ss.SeasonID] = append(res[ss.SeasonID], ss)
		}
	}
	if err := rows.Err(); err != nil {
		log.Error("rows.Err error(%+v)", err)
		return nil, err
	}
	return res, nil
}

// EpisodesInfos get season_ep by sids.
func (d *Dao) EpisodesInfos(c context.Context, sids []int64) (map[int64][]*api.Episode, error) {
	rows, err := d.season.Query(c, fmt.Sprintf(_epsSQL, xstr.JoinInts(sids)))
	if err != nil {
		log.Error("d.resultDB.Query seasonsEp(%+v) error(%+v)", sids, err)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64][]*api.Episode)
	for rows.Next() {
		se := &api.Episode{}
		if err = rows.Scan(&se.ID, &se.SectionID, &se.SeasonID, &se.Title, &se.Aid, &se.Cid, &se.Attribute); err != nil {
			log.Error("rows.Scan error(%+v)", err)
			continue
		}
		res[se.SectionID] = append(res[se.SectionID], se)
	}
	if err := rows.Err(); err != nil {
		log.Error("rows.Err error(%+v)", err)
		return nil, err
	}
	return res, nil
}

package result

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/ugc-season/job/model/archive"
)

const (
	_insertSeasonSQL  = "INSERT INTO season (season_id,title,`desc`,cover,mid,attribute,sign_state,ptime,first_aid,ep_count,ep_num) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE title=values(title),`desc`=values(`desc`),cover=values(cover),mid=values(mid),attribute=values(attribute),sign_state=values(sign_state),ptime=values(ptime),first_aid=values(first_aid),ep_count=values(ep_count),ep_num=values(ep_num)"
	_inSectionSQL     = "INSERT INTO season_section (section_id,`type`,season_id,title,`order`) VALUES %s ON DUPLICATE KEY UPDATE section_id=values(section_id),`type`=values(`type`),season_id=values(season_id),title=values(title),`order`=values(`order`)"
	_inEpSQL          = "INSERT INTO season_episode (episode_id,title,aid,cid,season_id,section_id,`order`,attribute) VALUES %s ON DUPLICATE KEY UPDATE episode_id=values(episode_id),title=values(title),aid=values(aid),cid=values(cid),season_id=values(season_id),section_id=values(section_id),`order`=values(`order`),attribute=values(attribute)"
	_delSeasonByIDSQL = "DELETE FROM season WHERE season_id=?"
	_delSecBySIDSQL   = "DELETE FROM season_section WHERE season_id=?"
	_delSecByIDSQL    = "DELETE FROM season_section WHERE section_id IN (%s)"
	_delEpBySIDSQL    = "DELETE FROM season_episode WHERE season_id=?"
	_delEpByIDSQL     = "DELETE FROM season_episode WHERE episode_id IN (%s)"
	_upSeasonMtimeSQL = "UPDATE season SET mtime=? WHERE season_id=?"
	_maxSeasonIDSQL   = "SELECT MAX(season_id) FROM season"
)

// TxAddSeason add archive season
func (d *Dao) TxAddSeason(c context.Context, tx *sql.Tx, season *archive.Season, maxPtime xtime.Time, firstAid, epCnt int64) (err error) {
	_, err = tx.Exec(_insertSeasonSQL, season.SeasonID, season.Title, season.Desc, season.Cover, season.Mid, season.Attribute, season.SignState, maxPtime, firstAid, epCnt, season.EpNum)
	if err != nil {
		log.Error("tx.Exec(%s) error(%v)", _insertSeasonSQL, err)
		return
	}
	return
}

// TxAddSection add archive section
func (d *Dao) TxAddSection(c context.Context, tx *sql.Tx, sections []*archive.SeasonSection) (err error) {
	var (
		valSQL []string
		values []interface{}
	)
	for _, s := range sections {
		valSQL = append(valSQL, "(?,?,?,?,?)")
		values = append(values, s.SectionID, s.Type, s.SeasonID, s.Title, s.Order)
	}
	valSQLStr := strings.Join(valSQL, ",")
	_, err = tx.Exec(fmt.Sprintf(_inSectionSQL, valSQLStr), values...)
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxAddEp add archive Ep
func (d *Dao) TxAddEp(c context.Context, tx *sql.Tx, eps []*archive.SeasonEp) (err error) {
	var (
		valSQL []string
		values []interface{}
	)
	for _, e := range eps {
		valSQL = append(valSQL, "(?,?,?,?,?,?,?,?)")
		values = append(values, e.EpID, e.Title, e.AID, e.CID, e.SeasonID, e.SectionID, e.Order, e.Attribute)
	}
	valSQLStr := strings.Join(valSQL, ",")
	_, err = tx.Exec(fmt.Sprintf(_inEpSQL, valSQLStr), values...)
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxDelSeasonByID del season by season_id
func (d *Dao) TxDelSeasonByID(c context.Context, tx *sql.Tx, sid int64) (err error) {
	_, err = tx.Exec(_delSeasonByIDSQL, sid)
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxDelSecByID del section by section_id
func (d *Dao) TxDelSecByID(c context.Context, tx *sql.Tx, ids []int64) (err error) {
	_, err = tx.Exec(fmt.Sprintf(_delSecByIDSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxDelEpByID del ep by ep_id
func (d *Dao) TxDelEpByID(c context.Context, tx *sql.Tx, ids []int64) (err error) {
	_, err = tx.Exec(fmt.Sprintf(_delEpByIDSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxDelSecBySID del section by season_id
func (d *Dao) TxDelSecBySID(c context.Context, tx *sql.Tx, sid int64) (err error) {
	_, err = tx.Exec(_delSecBySIDSQL, sid)
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxDelEpBySID del ep by season_id
func (d *Dao) TxDelEpBySID(c context.Context, tx *sql.Tx, sid int64) (err error) {
	_, err = tx.Exec(_delEpBySIDSQL, sid)
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// TxUpSeasonMtime update season.mtime
func (d *Dao) TxUpSeasonMtime(c context.Context, tx *sql.Tx, sid int64) (err error) {
	_, err = tx.Exec(_upSeasonMtimeSQL, time.Now(), sid)
	if err != nil {
		log.Error("tx.Exec error(%v)", err)
		return
	}
	return
}

// MaxSeasonID get max season id
func (d *Dao) MaxSeasonID(c context.Context) (sid int64, err error) {
	row := d.db.QueryRow(c, _maxSeasonIDSQL)
	if err = row.Scan(&sid); err != nil {
		log.Error("rows.Scan error(%v)", err)
		return
	}
	return
}

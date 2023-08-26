package show

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"
	v1 "go-gateway/app/app-svr/app-show/interface/api"

	"github.com/siddontang/go-mysql/mysql"
)

// sql and status
const (
	_newSerieSQL       = "INSERT INTO selected_serie (`type`,number, stime, etime, pubtime) VALUES (?,?,?,?,?)"
	_AIDataSelectedSQL = "INSERT INTO selected_resource (rid,rtype, serie_id, source, creator, position, status) VALUES (?,?,?,?,?,?,1)"                                                                                                       // 1 means passed by default
	_PickSerie         = "SELECT id, type, number, status, stime, etime, subject, share_subtitle FROM selected_serie WHERE (type = ? AND date_sub(now(), INTERVAL 1 DAY) BETWEEN stime AND etime AND deleted = 0) ORDER BY ctime DESC LIMIT 1" // insertion on Friday means last period's data
	_PickSerieID       = "SELECT id, type, number, status, stime, etime, subject, share_subtitle, push_title, push_subtitle FROM selected_serie WHERE id = ?"
	_AIAlert           = "SELECT COUNT(1) FROM selected_resource WHERE serie_id = ? AND source = 1"                               // source = 1 means from AI
	_MaxPosition       = "SELECT IFNULL(MAX(position),0)FROM selected_resource WHERE serie_id = ? AND status = 1 AND deleted = 0" // pick the max position of the valid cards
	_MaxNumber         = "SELECT IFNULL(MAX(number),0) FROM selected_serie WHERE type = ? AND deleted = 0"
	_PickRes           = "SELECT rid, rtype FROM selected_resource WHERE serie_id = ? AND status = 1 AND deleted = 0 ORDER BY POSITION ASC"
	_UpdateSerie       = "UPDATE selected_serie SET status = ? WHERE id = ? AND deleted = 0"
	_AddMedialistID    = "UPDATE selected_serie SET media_id = ? WHERE id = ? AND deleted = 0"
	_FetchSerieID      = "SELECT id FROM selected_serie WHERE number = ? AND deleted = 0 AND `status` IN (2,4) AND pubtime <= NOW()" //2=审核通过, 4=灾备数据, pubtime控制发布时间
	_sourceAI          = 1
	_creatorAI         = "AI"
	_statusRecovery    = 4
)

// SerieRecovery updates the status of the given serie
func (d *Dao) SerieRecovery(c context.Context, sid int64) (err error) {
	_, err = d.db.Exec(c, _UpdateSerie, _statusRecovery, sid)
	return
}

// WriteMedialist updates the mediaID of the given serie
func (d *Dao) WriteMedialist(c context.Context, sid, mediaList int64) (err error) {
	_, err = d.db.Exec(c, _AddMedialistID, mediaList, sid)
	return
}

// AICount picks AI data to distinguish whether we need to alert AI colleagues
func (d *Dao) AICount(c context.Context, sid int64) (count int, err error) {
	err = d.db.QueryRow(c, _AIAlert, sid).Scan(&count)
	return
}

// SerieRes picks the resources of the given serie
func (d *Dao) SerieRes(c context.Context, sid int64) (res []*show.SerieRes, err error) {
	rows, err := d.db.Query(c, _PickRes, sid)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		idx := new(show.SerieRes)
		if err = rows.Scan(&idx.RID, &idx.Rtype); err != nil {
			log.Error("row.Scan() error(%v)", err)
			return
		}
		res = append(res, idx)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) SerieID(c context.Context, number int64) (id int64, err error) {
	err = d.db.QueryRow(c, _FetchSerieID, number).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// NewSerie adds a new serie of "xxx" selected into DB
func (d *Dao) NewSerie(c context.Context, serie *show.Serie) (err error) {
	if serie == nil || serie.Number == 0 {
		return ecode.RequestErr
	}
	_, err = d.db.Exec(c, _newSerieSQL,
		serie.Type, serie.Number,
		serie.Stime.Time().Format(mysql.TimeFormat),
		serie.Etime.Time().Format(mysql.TimeFormat),
		serie.Pubtime.Time().Format(mysql.TimeFormat))
	return
}

// PickSerie picks the correct serie_id according to the insertion time
func (d *Dao) PickSerie(c context.Context, sType string) (res *show.Serie, err error) {
	res = &show.Serie{}
	err = d.db.QueryRow(c, _PickSerie, sType).Scan(&res.ID, &res.Type, &res.Number, &res.Status, &res.Stime, &res.Etime, &res.Subject, &res.ShareSubtitle)
	return
}

// PickSerie picks the correct serie_id according to the insertion time
func (d *Dao) PickSerieID(c context.Context, sid int64) (res *show.Serie, err error) {
	res = &show.Serie{}
	err = d.db.QueryRow(c, _PickSerieID, sid).Scan(&res.ID, &res.Type, &res.Number, &res.Status, &res.Stime, &res.Etime, &res.Subject, &res.ShareSubtitle, &res.PushTitle, &res.PushSubtitle)
	return
}

// MaxPosition picks the max position of one given serie's resources
func (d *Dao) MaxPosition(c context.Context, sid int64) (max int64, err error) {
	err = d.db.QueryRow(c, _MaxPosition, sid).Scan(&max)
	return
}

// MaxNumber picks the current max number to the selected
func (d *Dao) MaxNumber(c context.Context, sType string) (number int64, err error) {
	err = d.db.QueryRow(c, _MaxNumber, sType).Scan(&number)
	return
}

// AIInsertion def.
func (d *Dao) AIInsertion(c context.Context, serieID int64, maxPos int64, aiData []*show.SerieRes) (err error) {
	for _, v := range aiData {
		maxPos++
		if _, err = d.db.Exec(c, _AIDataSelectedSQL, v.RID, v.Rtype, serieID, _sourceAI, _creatorAI, maxPos); err != nil {
			return
		}
	}
	return
}

// RefreshSeries def.
func (d *Dao) RefreshSeries(c context.Context, sType string) (err error) {
	_, err = d.showGrpc.RefreshSeriesList(c, &v1.RefreshSeriesListReq{Type: sType})
	return
}

// RefreshSingleSerie def
func (d *Dao) RefreshSingleSerie(c context.Context, sType string, number int64) (err error) {
	_, err = d.showGrpc.RefreshSerie(c, &v1.RefreshSerieReq{Type: sType, Number: number})
	return
}

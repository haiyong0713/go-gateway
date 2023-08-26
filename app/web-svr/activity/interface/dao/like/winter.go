package like

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_selWinterStudySQL   = "SELECT id,mid,order_no,season_id,real_price,duration,ep_count,season_title,cover,is_end,total_progress,clock_in,watch_progress,share_progress,upload_progress,watch_duration,ctime FROM act_summer_study WHERE mid=?"
	_addWinterStudySQL   = "INSERT IGNORE INTO act_summer_study(mid,order_no,season_id,real_price,duration,ep_count,season_title,cover,is_notice) VALUES(?,?,?,?,?,?,?,?,?)"
	_selWinterMidSQL     = "SELECT mid FROM act_summer_study limit 100000"
	_upWinterProgressSQL = "UPDATE act_summer_study SET is_end=1,total_progress=?,clock_in=?,watch_progress=?,share_progress=?,upload_progress=?,watch_duration=? WHERE mid=?"
)

// RawWinterStudy.
func (d *Dao) RawWinterStudy(ctx context.Context, mid int64) (data *likemdl.WinterStudy, err error) {
	data = new(likemdl.WinterStudy)
	row := d.db.QueryRow(ctx, _selWinterStudySQL, mid)
	if err = row.Scan(&data.ID, &data.Mid, &data.OrderNo, &data.SeasonID, &data.RealPrice, &data.Duration, &data.EpCount, &data.SeasonTitle, &data.Cover,
		&data.IsEnd, &data.TotalProgress, &data.ClockIn, &data.WatchProgress, &data.ShareProgress, &data.UploadProgress, &data.WatchDuration, &data.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(ctx, "RawWinterStudy QueryRow mid(%d) error(%v)", mid, err)
		}
	}
	return
}

// JoinWinterStudy.
func (d *Dao) JoinWinterStudy(ctx context.Context, mid, IsNotice int64, course *likemdl.CourseOrder) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _addWinterStudySQL, mid, course.OrderNo, course.SeasonID, course.RealPrice, course.Duration, course.EpCount, course.SeasonTitle, course.Cover, IsNotice); err != nil {
		log.Errorc(ctx, "JoinWinterStudy error d.db.Exec(%d) error(%v)", mid, err)
		return
	}
	return res.LastInsertId()
}

func (d *Dao) RawWinterMids(ctx context.Context) ([]int64, error) {
	rows, err := d.db.Query(ctx, _selWinterMidSQL)
	if err != nil {
		log.Errorc(ctx, "RawWinterMids d.db.Query(%s) error(%+v)", _selWinterMidSQL, err)
		return nil, err
	}
	defer rows.Close()
	var mids []int64
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			log.Errorc(ctx, "RawWinterMids scan() error(%+v)", err)
			return nil, err
		}
		mids = append(mids, mid)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return mids, nil
}

func (d *Dao) UpWinterProgress(ctx context.Context, mid int64, progress *likemdl.WinterProgress) (int64, error) {
	row, err := d.db.Exec(ctx, _upWinterProgressSQL, progress.TotalProgress, progress.ClockIn, progress.WatchProgress, progress.ShareProgress, progress.UploadProgress, progress.WatchDuration, mid)
	if err != nil {
		return 0, err
	}
	return row.RowsAffected()
}

func winterStudyKey(mid int64) string {
	return fmt.Sprintf("summer_study_%d", mid)
}

// CacheWinterStudy.
func (d *Dao) CacheWinterStudy(ctx context.Context, mid int64) (res *likemdl.WinterStudy, err error) {
	var (
		key = winterStudyKey(mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warnc(ctx, "CacheWinterStudy key(%s) return nil", key)
		} else {
			log.Errorc(ctx, "CacheWinterStudy conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "CacheWinterStudy key(%s) json.Unmarshal(%v) error(%v)", key, bs, err)
	}
	return
}

func (d *Dao) AddCacheWinterStudy(ctx context.Context, mid int64, data *likemdl.WinterStudy) (err error) {
	var (
		key = winterStudyKey(mid)
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "AddCacheWinterStudy key(%s) json.Marshal(%v) error (%+v)", key, data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 8640000, bs); err != nil {
		log.Errorc(ctx, "AddCacheWinterStudy conn.Send(SETEX, %s, %v, %s) error(%v)", key, 8640000, string(bs), err)
	}
	return
}

func (d *Dao) DelCacheWinterStudy(ctx context.Context, mid int64) error {
	key := winterStudyKey(mid)
	_, err := component.GlobalRedis.Do(ctx, "DEL", key)
	return err
}

func (d *Dao) WinterInfo(c context.Context, mid int64) (res *likemdl.WinterStudy, err error) {
	if res, err = d.CacheWinterStudy(c, mid); err != nil {
		err = nil
	}
	if res != nil {
		return
	}
	if res, err = d.RawWinterStudy(c, mid); err != nil {
		return
	}
	if res != nil && res.ID > 0 {
		d.AddCacheWinterStudy(c, mid, res)
	}
	return
}

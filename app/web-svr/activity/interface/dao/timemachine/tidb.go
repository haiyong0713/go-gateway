package timemachine

import (
	"context"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"
)

func (d *Dao) RawUserYearReport2020(c context.Context, mid int64) (res *timemachine.UserYearReport2020, err error) {
	var sqlstr = `SELECT ` +
		`mid,visit_days,play_videos,play_minutes_real,play_minutes,hour_visit_days,fav_type,fav_tag,top6_tid_score,is_show_p4,latest_play_time,latest_play_up,latest_play_avid,is_show_p5,longest_play_day,longest_play_hours_real,longest_play_hours,longest_play_tag,longest_play_subtid,is_show_p6,max_vv,max_vv_up,max_vv_avid,is_show_p7,sum_like,sum_coin,sum_fav,coin_time,coin_users,coin_avid,is_show_p8,recommand_avid,is_show_p9,fav_up_type,fav_up,fav_up_vv,fav_up_oid,is_show_p10,create_avs,create_reads,av_vv,read_vv,best_create_type,best_create,is_show_p11,play_comic,play_movie,play_drama,play_documentary,play_variety,fav_season_id,fav_season_type,is_show_p12,live_hours,live_beyond_percent,fav_live_up,fav_live_up_play,max_vv_highlight,coin_highlight,ctime,mtime,vip_days,vip_av_count,vip_av_play,is_show_p13,latest_play_highlight,fav_up_highlight ` +
		`FROM  ads_user_year_report_2020_1y_y` +
		` WHERE mid = ? AND log_date = ?`
	a := timemachine.UserYearReport2020{}
	err = component.GlobalTiDB.QueryRow(c, sqlstr, mid, d.c.Timemachine.LogDate).Scan(&a.Mid, &a.VisitDays, &a.PlayVideos, &a.PlayMinutesReal, &a.PlayMinutes, &a.HourVisitDays, &a.FavType, &a.FavTag, &a.Top6TidScore, &a.IsShowP4, &a.LatestPlayTime, &a.LatestPlayUp, &a.LatestPlayAvid, &a.IsShowP5, &a.LongestPlayDay, &a.LongestPlayHoursReal, &a.LongestPlayHours, &a.LongestPlayTag, &a.LongestPlaySubtid, &a.IsShowP6, &a.MaxVv, &a.MaxVvUp, &a.MaxVvAvid, &a.IsShowP7, &a.SumLike, &a.SumCoin, &a.SumFav, &a.CoinTime, &a.CoinUsers, &a.CoinAvid, &a.IsShowP8, &a.RecommandAvid, &a.IsShowP9, &a.FavUpType, &a.FavUp, &a.FavUpVv, &a.FavUpOid, &a.IsShowP10, &a.CreateAvs, &a.CreateReads, &a.AvVv, &a.ReadVv, &a.BestCreateType, &a.BestCreate, &a.IsShowP11, &a.PlayComic, &a.PlayMovie, &a.PlayDrama, &a.PlayDocumentary, &a.PlayVariety, &a.FavSeasonID, &a.FavSeasonType, &a.IsShowP12, &a.LiveHours, &a.LiveBeyondPercent, &a.FavLiveUp, &a.FavLiveUpPlay, &a.MaxVvHighlight, &a.CoinHighlight, &a.Ctime, &a.Mtime, &a.VipDays, &a.VipAvCount, &a.VipAvPlay, &a.IsShowP13, &a.LatestPlayHighlight, &a.FavUpHighlight)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorc(c, "RawUserYearReport2020 QueryRow err: %v mid: %v", err, mid)
		return nil, err
	}
	return &a, nil
}

func (d *Dao) RawUserReport2020TagInfo(c context.Context) (res []*timemachine.UserReport2020TagInfo, err error) {
	const sqlstr = `SELECT ` +
		`id,tag_name,display,description,img,ctime,mtime ` +
		`FROM  user_report_2020_tag_info `
	var q *sql.Rows
	q, err = component.GlobalTiDB.Query(c, sqlstr)
	if err != nil {
		log.Errorc(c, "RawUserReport2020TagInfo Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res = []*timemachine.UserReport2020TagInfo{}
	for q.Next() {
		a := timemachine.UserReport2020TagInfo{}

		err = q.Scan(&a.ID, &a.TagName, &a.Display, &a.Description, &a.Img, &a.Ctime, &a.Mtime)
		if err != nil {
			log.Errorc(c, "RawUserReport2020TagInfo Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Errorc(c, "RawUserReport2020TagInfo Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

func (d *Dao) RawUserReport2020TypeInfo(c context.Context) (res []*timemachine.UserReport2020TypeInfo, err error) {
	const sqlstr = `SELECT ` +
		`tid,tid_name,sub_tid_name,display,description,img,ctime,mtime,pid ` +
		`FROM  user_report_2020_type_info `
	var q *sql.Rows
	q, err = component.GlobalTiDB.Query(c, sqlstr)
	if err != nil {
		log.Errorc(c, "RawUserReport2020TypeInfo Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res = []*timemachine.UserReport2020TypeInfo{}
	for q.Next() {
		a := timemachine.UserReport2020TypeInfo{}

		err = q.Scan(&a.Tid, &a.TidName, &a.SubTidName, &a.Display, &a.Description, &a.Img, &a.Ctime, &a.Mtime, &a.Pid)
		if err != nil {
			log.Errorc(c, "RawUserReport2020TypeInfo Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Errorc(c, "RawUserReport2020TypeInfo Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

func (d *Dao) InsertUserInfo(ctx context.Context, a *timemachine.UserInfo) error {
	var err error
	const sqlstr = `INSERT INTO  user_report_2020_user_info (` +
		` mid,aid,is_new,lottery_id` +
		`) VALUES (` +
		` ?,?,?,?` +
		`)`

	_, err = component.GlobalTiDB.Exec(ctx, sqlstr, a.Mid, a.Aid, a.IsNew, a.LotteryID)
	if err != nil {
		log.Errorc(ctx, "UserReport2020UserInfo Insert Exec err: %v", err)
		return err
	}
	return nil
}

func (d *Dao) RawUserInfoByMid(ctx context.Context, mid int64) (*timemachine.UserInfo, error) {
	var err error

	const sqlstr = `SELECT ` +
		`mid,aid,is_new,lottery_id,ctime,mtime ` +
		`FROM  user_report_2020_user_info ` +
		`WHERE mid = ?`

	a := timemachine.UserInfo{}

	err = component.GlobalTiDB.QueryRow(ctx, sqlstr, mid).Scan(&a.Mid, &a.Aid, &a.IsNew, &a.LotteryID, &a.Ctime, &a.Mtime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("UserReport2020UserInfoByMid QueryRow err: %v mid: %v", err, mid)
		return nil, err
	}

	return &a, nil
}

func (d *Dao) UpdateUserInfo(ctx context.Context, a *timemachine.UserInfo) error {
	var err error

	const sqlstr = `UPDATE user_report_2020_user_info SET ` +
		` aid = ? ,is_new = ?, lottery_id = ? ` +
		` WHERE mid = ?`

	_, err = component.GlobalTiDB.Exec(ctx, sqlstr, a.Aid, a.IsNew, a.LotteryID, a.Mid)
	if err != nil {
		log.Error("UserReport2020UserInfo Update err: %v", err)
	}
	return err
}

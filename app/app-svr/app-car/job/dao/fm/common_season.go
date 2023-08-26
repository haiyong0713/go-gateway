package fm

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/model/fm"

	"github.com/pkg/errors"
)

const (
	_fmSeasonInfo = "SELECT `id`, `fm_type`, `fm_id`, `title`, `cover`, `subtitle`, `fm_state`, `ctime`, `mtime` " +
		" FROM `fm_season_info` WHERE `fm_type` = ? AND `fm_id` = ? AND `is_deleted` = 0"
	_fmSeasonOid = "SELECT `id`, `fm_type`, `fm_id`, `oid`, `seq`, `ctime`, `mtime` " +
		" FROM `fm_season_oid` WHERE `fm_type` = ? AND `fm_id` = ? AND `is_deleted` = 0"
	_insertFmSeasonInfo = "INSERT INTO `fm_season_info` (`fm_type`, `fm_id`, `title`, `cover`) VALUES (?, ?, ?, ?)"
	_insertFmSeasonOids = "INSERT INTO `fm_season_oid` (`fm_type`, `fm_id`, `oid`, `seq`) VALUES (?, ?, ?, ?)"
	_delFmSeasonInfo    = "DELETE FROM `fm_season_info` WHERE `fm_type` = ? AND `fm_id` = ?"
	_delFmSeasonOids    = "DELETE FROM `fm_season_oid` WHERE `fm_type` = ? AND `fm_id` = ?"

	_videoSeasonInfo = "SELECT `id`, `season_id`, `title`, `cover`, `subtitle`, `season_state`, `ctime`, `mtime` " +
		" FROM `video_season_info` WHERE `season_id` = ? AND `is_deleted` = 0"
	_videoSeasonOid = "SELECT `id`, `season_id`, `oid`, `seq`, `ctime`, `mtime` " +
		" FROM `video_season_oid` WHERE `season_id` = ? AND `is_deleted` = 0"
	_insertVideoSeasonInfo = "INSERT INTO `video_season_info` (`season_id`, `title`, `cover`) VALUES (?, ?, ?)"
	_insertVideoSeasonOids = "INSERT INTO `video_season_oid` (`season_id`, `oid`, `seq`) VALUES (?, ?, ?)"
	_delVideoSeasonInfo    = "DELETE FROM `video_season_info` WHERE `season_id` = ?"
	_delVideoSeasonOids    = "DELETE FROM `video_season_oid` WHERE `season_id` = ?"
)

func (d *Dao) UpsertSeasonWithLock(ctx context.Context, season *fm.CommonSeason, handler SeasonHandler) (railgun.MsgPolicy, error) {
	conn := d.redisCli.Conn(ctx)
	defer conn.Close()
	switch season.Scene {
	case fm.SceneFm:
		infoLock, err := SetLock(conn, SeasonInfoLock(fm.SeasonInfoReq{Scene: season.Scene, FmType: season.Fm.FmType, SeasonId: season.Fm.FmId}))
		if !infoLock || err != nil {
			log.Error("UpsertSeasonWithLock SetLock err:%+v, lockKey:%s, infoLock:%t", err,
				SeasonInfoLock(fm.SeasonInfoReq{Scene: season.Scene, FmType: season.Fm.FmType, SeasonId: season.Fm.FmId}), infoLock)
			return railgun.MsgPolicyAttempts, ecode.ServiceUpdate
		}
		defer DelLock(conn, SeasonInfoLock(fm.SeasonInfoReq{Scene: season.Scene, FmType: season.Fm.FmType, SeasonId: season.Fm.FmId})) // nolint:errcheck

		oidLock, err := SetLock(conn, SeasonOidLock(season.Scene, season.Fm.FmType, season.Fm.FmId))
		if !oidLock || err != nil {
			log.Error("UpsertSeasonWithLock d.SetLock err:%+v, lockKey:%s, oidLock:%t", err, SeasonOidLock(season.Scene, season.Fm.FmType, season.Fm.FmId), oidLock)
			return railgun.MsgPolicyAttempts, ecode.ServiceUpdate
		}
		defer DelLock(conn, SeasonOidLock(season.Scene, season.Fm.FmType, season.Fm.FmId)) // nolint:errcheck
	case fm.SceneVideo:
		infoLock, err := SetLock(conn, SeasonInfoLock(fm.SeasonInfoReq{Scene: season.Scene, FmType: "", SeasonId: season.Video.SeasonId}))
		if !infoLock || err != nil {
			log.Error("UpsertSeasonWithLock SetLock err:%+v, lockKey:%s, infoLock:%t", err,
				SeasonInfoLock(fm.SeasonInfoReq{Scene: season.Scene, FmType: "", SeasonId: season.Video.SeasonId}), infoLock)
			return railgun.MsgPolicyAttempts, ecode.ServiceUpdate
		}
		defer DelLock(conn, SeasonInfoLock(fm.SeasonInfoReq{Scene: season.Scene, FmType: "", SeasonId: season.Video.SeasonId})) // nolint:errcheck

		oidLock, err := SetLock(conn, SeasonOidLock(season.Scene, "", season.Video.SeasonId))
		if !oidLock || err != nil {
			log.Error("UpsertSeasonWithLock d.SetLock err:%+v, lockKey:%s, oidLock:%t", err, SeasonOidLock(season.Scene, "", season.Video.SeasonId), oidLock)
			return railgun.MsgPolicyAttempts, ecode.ServiceUpdate
		}
		defer DelLock(conn, SeasonOidLock(season.Scene, "", season.Video.SeasonId)) // nolint:errcheck
	default:
		return railgun.MsgPolicyIgnore, ecode.RequestErr
	}
	return handler(ctx, season)
}

// ModifySeasonWithTx 开启事务修改FM合集
func (d *Dao) ModifySeasonWithTx(c context.Context, season *fm.CommonSeason, modType fm.ModifyType) error {
	tx, err := d.db.Begin(c)
	if err != nil {
		log.Error("ModifySeasonWithTx begin transaction failed, err=%+v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("ModifySeasonWithTx tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("ModifySeasonWithTx tx.Commit() error(%v)", err)
		}
	}()

	switch season.Scene {
	case fm.SceneFm:
		if modType == fm.TypeUpdate {
			_, err = tx.Exec(_delFmSeasonInfo, string(season.Fm.FmType), season.Fm.FmId)
			if err != nil {
				return err
			}
			_, err = tx.Exec(_delFmSeasonOids, string(season.Fm.FmType), season.Fm.FmId)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(_insertFmSeasonInfo, string(season.Fm.FmType), season.Fm.FmId, season.Fm.Title, season.Fm.Cover)
		if err != nil {
			return err
		}
		for i, item := range season.Fm.FmList {
			_, err = tx.Exec(_insertFmSeasonOids, season.Fm.FmType, season.Fm.FmId, item.Aid, i)
			if err != nil {
				return err
			}
		}
	case fm.SceneVideo:
		if modType == fm.TypeUpdate {
			_, err = tx.Exec(_delVideoSeasonInfo, season.Video.SeasonId)
			if err != nil {
				return err
			}
			_, err = tx.Exec(_delVideoSeasonOids, season.Video.SeasonId)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(_insertVideoSeasonInfo, season.Video.SeasonId, season.Video.Title, season.Video.Cover)
		if err != nil {
			return err
		}
		for i, item := range season.Video.SeasonList {
			_, err = tx.Exec(_insertVideoSeasonOids, season.Video.SeasonId, item.Aid, i)
			if err != nil {
				return err
			}
		}
	default:
		log.Errorc(c, "ModifySeasonWithTx unknown scene:%s", season.Scene)
		return ecode.RequestErr
	}
	return nil
}

func (d *Dao) QuerySeasonInfo(ctx context.Context, req fm.SeasonInfoReq) (*fm.SeasonInfoResp, error) {
	if req.Scene == fm.SceneFm {
		po, err := d.QueryFmSeasonInfo(ctx, req.FmType, req.SeasonId)
		if err != nil {
			return nil, err
		}
		return &fm.SeasonInfoResp{Scene: fm.SceneFm, Fm: po}, nil
	} else if req.Scene == fm.SceneVideo {
		po, err := d.QueryVideoSeasonInfo(ctx, req.SeasonId)
		if err != nil {
			return nil, err
		}
		return &fm.SeasonInfoResp{Scene: fm.SceneVideo, Video: po}, nil
	}
	return nil, ecode.RequestErr
}

func (d *Dao) QueryFmSeasonInfo(ctx context.Context, fmType fm.FmType, fmId int64) (*fm.FmSeasonInfoPo, error) {
	var (
		pos []*fm.FmSeasonInfoPo
	)
	rows, err := d.db.Query(ctx, _fmSeasonInfo, fmType, fmId)
	if err != nil {
		return nil, errors.Wrap(err, "QueryFmSeasonInfo d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm.FmSeasonInfoPo)
		if err := rows.Scan(&po.Id, &po.FmType, &po.FmId, &po.Title, &po.Cover, &po.Subtitle, &po.FmState,
			&po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "QueryFmSeasonInfo scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "QueryFmSeasonInfo rows.Err error")
	}
	if len(pos) == 0 {
		return nil, ecode.NothingFound
	} else if len(pos) > 1 {
		// 存在多个相同合集，则取最近插入的，并告警
		log.Warn("【P2】QueryFmSeasonInfo get multiple seasons, fmType:%s, fmId:%d", fmType, fmId)
		return pos[len(pos)-1], nil
	}
	return pos[0], nil
}

func (d *Dao) QueryVideoSeasonInfo(ctx context.Context, seasonId int64) (*fm.VideoSeasonInfoPo, error) {
	var (
		pos []*fm.VideoSeasonInfoPo
	)
	rows, err := d.db.Query(ctx, _videoSeasonInfo, seasonId)
	if err != nil {
		return nil, errors.Wrap(err, "QueryVideoSeasonInfo d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm.VideoSeasonInfoPo)
		if err := rows.Scan(&po.Id, &po.SeasonId, &po.Title, &po.Cover, &po.Subtitle, &po.SeasonState,
			&po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "QueryVideoSeasonInfo scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "QueryVideoSeasonInfo rows.Err error")
	}
	if len(pos) == 0 {
		return nil, ecode.NothingFound
	} else if len(pos) > 1 {
		// 存在多个相同合集，则取最近插入的，并告警
		log.Warnc(ctx, "【P2】QueryVideoSeasonInfo get multiple seasons, seasonId:%d", seasonId)
		return pos[len(pos)-1], nil
	}
	return pos[0], nil
}

func (d *Dao) QuerySeasonOid(ctx context.Context, scene fm.Scene, fmType fm.FmType, seasonId int64) ([]*fm.FmSeasonOidPo, []*fm.VideoSeasonOidPo, error) {
	if scene == fm.SceneFm {
		pos, err := d.QueryFmSeasonOid(ctx, fmType, seasonId)
		if err != nil {
			return nil, nil, err
		}
		return pos, nil, nil
	} else if scene == fm.SceneVideo {
		pos, err := d.QueryVideoSeasonOid(ctx, seasonId)
		if err != nil {
			return nil, nil, err
		}
		return nil, pos, nil
	}
	return nil, nil, ecode.RequestErr
}

func (d *Dao) QueryFmSeasonOid(ctx context.Context, fmType fm.FmType, fmId int64) ([]*fm.FmSeasonOidPo, error) {
	var (
		pos = make([]*fm.FmSeasonOidPo, 0)
	)
	rows, err := d.db.Query(ctx, _fmSeasonOid, fmType, fmId)
	if err != nil {
		return nil, errors.Wrap(err, "seasonOidByDB d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm.FmSeasonOidPo)
		if err := rows.Scan(&po.Id, &po.FmType, &po.FmId, &po.Oid, &po.Seq, &po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "seasonOidByDB scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "seasonOidByDB rows.Err error")
	}
	return pos, nil
}

func (d *Dao) QueryVideoSeasonOid(ctx context.Context, seasonId int64) ([]*fm.VideoSeasonOidPo, error) {
	var (
		pos = make([]*fm.VideoSeasonOidPo, 0)
	)
	rows, err := d.db.Query(ctx, _videoSeasonOid, seasonId)
	if err != nil {
		return nil, errors.Wrap(err, "QueryVideoSeasonOid d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm.VideoSeasonOidPo)
		if err := rows.Scan(&po.Id, &po.SeasonId, &po.Oid, &po.Seq, &po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "QueryVideoSeasonOid scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "QueryVideoSeasonOid rows.Err error")
	}
	return pos, nil
}

func (d *Dao) DelSeasonCache(c context.Context, scene fm.Scene, fmType fm.FmType, seasonId int64) error {
	conn1 := d.redisCli.Conn(c)
	defer conn1.Close()
	_, err := conn1.Do("DEL", SeasonInfoKey(fm.SeasonInfoReq{Scene: scene, FmType: fmType, SeasonId: seasonId}))
	if err != nil {
		return err
	}
	_, err = conn1.Do("DEL", SeasonOidKey(scene, fmType, seasonId))
	if err != nil {
		return err
	}
	conn2 := d.redisCliJd.Conn(c)
	defer conn2.Close()
	_, err = conn2.Do("DEL", SeasonInfoKey(fm.SeasonInfoReq{Scene: scene, FmType: fmType, SeasonId: seasonId}))
	if err != nil {
		return err
	}
	_, err = conn2.Do("DEL", SeasonOidKey(scene, fmType, seasonId))
	if err != nil {
		return err
	}
	return nil
}

// 音视频合集通用dao查询方法，从fm合集改造过来所以放fm包下；TODO 重构调整包路径

package fm

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"

	"github.com/pkg/errors"
)

const (
	_fmSeasonInfo = "SELECT `id`, `fm_type`, `fm_id`, `title`, `cover`, `subtitle`, `fm_state`, `ctime`, `mtime` " +
		" FROM `fm_season_info` WHERE `fm_type` = ? AND `fm_id` = ? AND `is_deleted` = 0"
	_fmSeasonOid = "SELECT `id`, `fm_type`, `fm_id`, `oid`, `seq`, `ctime`, `mtime` " +
		" FROM `fm_season_oid` WHERE `fm_type` = ? AND `fm_id` = ? AND `is_deleted` = 0"
	_videoSeasonInfo = "SELECT `id`, `season_id`, `title`, `cover`, `subtitle`, `season_state`, `ctime`, `mtime` " +
		" FROM `video_season_info` WHERE `season_id` = ? AND `is_deleted` = 0"
	_videoSeasonOid = "SELECT `id`, `season_id`, `oid`, `seq`, `ctime`, `mtime` " +
		" FROM `video_season_oid` WHERE `season_id` = ? AND `is_deleted` = 0"
)

// GetSeasonInfo 获取合集播单的基础信息（可能为空结构体）
func (d *Dao) GetSeasonInfo(ctx context.Context, req fm_v2.SeasonInfoReq) (*fm_v2.SeasonInfoResp, error) {
	conn := d.redisCli.Conn(ctx)
	defer conn.Close()
	// 1. 查询缓存
	cache, err := d.seasonInfoCache(conn, req)
	if err == nil {
		return cache, nil
	}
	// 2. 缓存miss，则加锁
	isLock, err := SetLock(conn, SeasonInfoLock(req))
	if !isLock || err != nil {
		log.Error("GetSeasonInfo d.SetLock err:%+v, lockKey:%s, isLock:%t", err, SeasonInfoLock(req), isLock)
		return d.defaultSeasonInfo(req), nil // redis加锁失败降级为空结构返回
	}
	defer DelLock(conn, SeasonInfoLock(req)) // nolint:errcheck
	// 3. 查MySQL，存入缓存
	infoByDB, err := d.seasonInfoByDB(ctx, req)
	if err != nil {
		return nil, err
	}
	err = d.setSeasonInfoCache(conn, req, infoByDB)
	if err != nil {
		return nil, err
	}
	return infoByDB, nil
}

func (d *Dao) seasonInfoCache(conn redis.Conn, req fm_v2.SeasonInfoReq) (*fm_v2.SeasonInfoResp, error) {
	reply, err := redis.String(conn.Do("GET", SeasonInfoKey(req)))
	if err != nil {
		return nil, err
	}
	if req.Scene == fm_v2.SceneFm {
		info := new(fm_v2.FmSeasonInfoPo)
		if err = json.Unmarshal([]byte(reply), info); err != nil {
			return nil, err
		}
		return &fm_v2.SeasonInfoResp{Scene: fm_v2.SceneFm, Fm: info}, nil
	} else if req.Scene == fm_v2.SceneVideo {
		info := new(fm_v2.VideoSeasonInfoPo)
		if err = json.Unmarshal([]byte(reply), info); err != nil {
			return nil, err
		}
		return &fm_v2.SeasonInfoResp{Scene: fm_v2.SceneVideo, Video: info}, nil
	} else {
		return nil, ecode.RequestErr
	}
}

func (d *Dao) setSeasonInfoCache(conn redis.Conn, req fm_v2.SeasonInfoReq, resp *fm_v2.SeasonInfoResp) error {
	var (
		bytes []byte
		err   error
	)
	if resp == nil {
		return errors.Wrap(ecode.NothingFound, "DB入缓存数据为空")
	}
	if req.Scene == fm_v2.SceneFm {
		bytes, err = json.Marshal(resp.Fm)
	} else if req.Scene == fm_v2.SceneVideo {
		bytes, err = json.Marshal(resp.Video)
	} else {
		err = ecode.RequestErr
	}
	if err != nil {
		return err
	}
	reply, err := redis.String(conn.Do("SET", SeasonInfoKey(req), string(bytes)))
	// todo 增加重试机制
	if err != nil {
		return err
	}
	if reply != "OK" {
		return errors.New("setSeasonInfoCache reply !ok")
	}
	return nil
}

func (d *Dao) seasonInfoByDB(ctx context.Context, req fm_v2.SeasonInfoReq) (*fm_v2.SeasonInfoResp, error) {
	if req.Scene == fm_v2.SceneFm {
		po, err := d.fmSeasonInfoDB(ctx, req.FmType, req.SeasonId)
		if err != nil {
			return nil, err
		}
		return &fm_v2.SeasonInfoResp{Scene: fm_v2.SceneFm, Fm: po}, nil
	} else if req.Scene == fm_v2.SceneVideo {
		po, err := d.videoSeasonInfoDB(ctx, req.SeasonId)
		if err != nil {
			return nil, err
		}
		return &fm_v2.SeasonInfoResp{Scene: fm_v2.SceneVideo, Video: po}, nil
	}
	return nil, ecode.RequestErr
}

func (d *Dao) fmSeasonInfoDB(ctx context.Context, fmType fm_v2.FmType, fmId int64) (*fm_v2.FmSeasonInfoPo, error) {
	var (
		pos []*fm_v2.FmSeasonInfoPo
	)
	rows, err := d.db.Query(ctx, _fmSeasonInfo, fmType, fmId)
	if err != nil {
		return nil, errors.Wrap(err, "fmSeasonInfoDB d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm_v2.FmSeasonInfoPo)
		if err := rows.Scan(&po.Id, &po.FmType, &po.FmId, &po.Title, &po.Cover, &po.Subtitle, &po.FmState,
			&po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "fmSeasonInfoDB scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "fmSeasonInfoDB rows.Err error")
	}
	if len(pos) == 0 {
		// 传空结构规避缓存穿透
		return new(fm_v2.FmSeasonInfoPo), nil
	} else if len(pos) > 1 {
		// 存在多个相同合集，则取最近插入的，并告警
		log.Warnc(ctx, "【P2】fmSeasonInfoDB get multiple seasons, fmType:%s, fmId:%d", fmType, fmId)
		return pos[len(pos)-1], nil
	}
	return pos[0], nil
}

func (d *Dao) videoSeasonInfoDB(ctx context.Context, seasonId int64) (*fm_v2.VideoSeasonInfoPo, error) {
	var (
		pos []*fm_v2.VideoSeasonInfoPo
	)
	rows, err := d.db.Query(ctx, _videoSeasonInfo, seasonId)
	if err != nil {
		return nil, errors.Wrap(err, "videoSeasonInfoDB d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm_v2.VideoSeasonInfoPo)
		if err := rows.Scan(&po.Id, &po.SeasonId, &po.Title, &po.Cover, &po.Subtitle, &po.SeasonState,
			&po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "videoSeasonInfoDB scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "videoSeasonInfoDB rows.Err error")
	}
	if len(pos) == 0 {
		// 传空结构规避缓存穿透
		return new(fm_v2.VideoSeasonInfoPo), nil
	} else if len(pos) > 1 {
		// 存在多个相同合集，则取最近插入的，并告警
		log.Warnc(ctx, "【P2】videoSeasonInfoDB get multiple seasons, seasonId:%d", seasonId)
		return pos[len(pos)-1], nil
	}
	return pos[0], nil
}

func (d *Dao) defaultSeasonInfo(req fm_v2.SeasonInfoReq) *fm_v2.SeasonInfoResp {
	if req.Scene == fm_v2.SceneFm {
		return &fm_v2.SeasonInfoResp{Scene: fm_v2.SceneFm, Fm: new(fm_v2.FmSeasonInfoPo)}
	} else if req.Scene == fm_v2.SceneVideo {
		return &fm_v2.SeasonInfoResp{Scene: fm_v2.SceneVideo, Video: new(fm_v2.VideoSeasonInfoPo)}
	}
	return nil
}

// GetSeasonOid 查询合集稿件oid（可能为空数组）
func (d *Dao) GetSeasonOid(ctx context.Context, req fm_v2.SeasonOidReq) (oids []int64, hasMore bool, err error) {
	conn := d.redisCli.Conn(ctx)
	defer conn.Close()
	// 1. 查询缓存
	oids, hasMore, err = d.seasonOidCache(conn, req)
	if err == nil {
		return oids, hasMore, nil
	}
	if err != ecode.NothingFound {
		return nil, false, err
	}
	// 2. miss后加载缓存
	err = d.loadSeasonOids(ctx, conn, req.Scene, req.FmType, req.SeasonId)
	if err != nil {
		return nil, false, err
	}
	// 3. 分页查询后返回
	oids, hasMore, err = d.seasonOidCache(conn, req)
	if err != nil {
		return nil, false, err
	}
	return oids, hasMore, nil
}

func (d *Dao) loadSeasonOids(ctx context.Context, conn redis.Conn, scene fm_v2.Scene, fmType fm_v2.FmType, seasonId int64) error {
	// 加分布式锁
	isLock, err := SetLock(conn, SeasonOidLock(scene, fmType, seasonId))
	if !isLock || err != nil {
		log.Errorc(ctx, "GetSeasonOid d.SetLock err:%+v, lockKey:%s, isLock:%t", err, SeasonOidLock(scene, fmType, seasonId), isLock)
		return err
	}
	defer DelLock(conn, SeasonOidLock(scene, fmType, seasonId)) // nolint:errcheck
	// 查MySQL，存入缓存
	fmPos, videoPos, err := d.seasonOidByDB(ctx, scene, fmType, seasonId)
	if err != nil {
		return err
	}
	err = d.setSeasonOidCache(ctx, scene, fmType, seasonId, fmPos, videoPos)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) seasonOidCache(conn redis.Conn, req fm_v2.SeasonOidReq) (oids []int64, hasMore bool, err error) {
	count, err := redis.Int(conn.Do("ZCARD", SeasonOidKey(req.Scene, req.FmType, req.SeasonId)))
	if err != nil {
		return nil, false, errors.Wrapf(err, "ZCARD err, key:%s", SeasonOidKey(req.Scene, req.FmType, req.SeasonId))
	}
	if count == _notExist {
		return nil, false, ecode.NothingFound
	}
	var (
		min      int // zset range min
		max      int // zset range max
		fromHead = false
	)
	if !req.Upward {
		// 向下翻页
		rank, err := redis.Int(conn.Do("ZRANK", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), req.Cursor))
		if err != nil {
			if err == redis.ErrNil {
				fromHead = true
			} else {
				return nil, false, errors.Wrapf(err, "ZRANK err, key:%s, member:%d", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), req.Cursor)
			}
		}
		min, max = seasonOidRange(req.WithCurrent, rank, req.Ps, fromHead)
		slices, err := redis.ByteSlices(conn.Do("ZRANGE", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), min, max))
		if err != nil {
			return nil, false, errors.Wrapf(err, "ZRANGE err, key:%s, min:%d, max:%d", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), min, max)
		}
		for _, v := range slices {
			oid, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				log.Error("seasonOidCache ZRANGE strconv.ParseInt err:%+v, str:%s", err, string(v))
				continue
			}
			oids = append(oids, oid)
		}
	} else {
		// 向上翻页
		rank, err := redis.Int(conn.Do("ZREVRANK", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), req.Cursor))
		if err != nil {
			if err == redis.ErrNil {
				fromHead = true
			} else {
				return nil, false, errors.Wrapf(err, "ZREVRANK err, key:%s, member:%d", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), req.Cursor)
			}
		}
		min, max = seasonOidRange(req.WithCurrent, rank, req.Ps, fromHead)
		slices, err := redis.ByteSlices(conn.Do("ZREVRANGE", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), min, max))
		if err != nil {
			return nil, false, errors.Wrapf(err, "ZREVRANGE err, key:%s, min:%d, max:%d", SeasonOidKey(req.Scene, req.FmType, req.SeasonId), min, max)
		}
		if len(slices) == 0 {
			return make([]int64, 0), false, nil
		}
		// 逆序后返回
		for i := len(slices) - 1; i >= 0; i-- {
			oid, err := strconv.ParseInt(string(slices[i]), 10, 64)
			if err != nil {
				log.Error("seasonOidCache ZREVRANGE strconv.ParseInt err:%+v, str:%s", err, string(slices[i]))
				continue
			}
			oids = append(oids, oid)
		}
	}
	// 当此轮查询最大下标 小于 zset最大下标时，有下一页
	hasMore = max < count-1
	return oids, hasMore, nil
}

func (d *Dao) setSeasonOidCache(ctx context.Context, scene fm_v2.Scene, fmType fm_v2.FmType, seasonId int64, fm []*fm_v2.FmSeasonOidPo, video []*fm_v2.VideoSeasonOidPo) error {
	pipeline := d.redisCli.Pipeline()
	if scene == fm_v2.SceneFm && len(fm) == 0 || scene == fm_v2.SceneVideo && len(video) == 0 {
		// 规避缓存穿透
		pipeline.Send("ZADD", SeasonOidKey(scene, fmType, seasonId), -1, -1)
		pipeline.Send("EXPIRE", SeasonOidKey(scene, fmType, seasonId), _oneHourSec)
	} else {
		if scene == fm_v2.SceneFm {
			for _, po := range fm {
				pipeline.Send("ZADD", SeasonOidKey(scene, fmType, seasonId), po.Seq, po.Oid)
			}
		} else if scene == fm_v2.SceneVideo {
			for _, po := range video {
				pipeline.Send("ZADD", SeasonOidKey(scene, fmType, seasonId), po.Seq, po.Oid)
			}
		}
	}
	replies, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}
	var idx = 0
	for replies.Next() {
		reply, err := redis.Int(replies.Scan())
		if err != nil {
			log.Error("setSeasonOidCache reply err:%+v, scene:%+v, fmType:%s, seasonId:%d, fm:%+v, video:%+v, idx:%d", err, scene, fmType, seasonId, fm, video, idx)
			continue
		}
		if reply != 1 {
			log.Error("setSeasonOidCache zadd/expire fail, scene:%+v, fmType:%s, seasonId:%d, fm:%+v, video:%+v, idx:%d", scene, fmType, seasonId, fm, video, idx)
			continue
		}
		idx++
	}
	return nil
}

func (d *Dao) seasonOidByDB(ctx context.Context, scene fm_v2.Scene, fmType fm_v2.FmType, seasonId int64) ([]*fm_v2.FmSeasonOidPo, []*fm_v2.VideoSeasonOidPo, error) {
	if scene == fm_v2.SceneFm {
		pos, err := d.fmSeasonOidByDB(ctx, fmType, seasonId)
		if err != nil {
			return nil, nil, err
		}
		return pos, nil, nil
	} else if scene == fm_v2.SceneVideo {
		pos, err := d.videoSeasonOidByDB(ctx, seasonId)
		if err != nil {
			return nil, nil, err
		}
		return nil, pos, nil
	}
	return nil, nil, ecode.RequestErr
}

func (d *Dao) fmSeasonOidByDB(ctx context.Context, fmType fm_v2.FmType, seasonId int64) ([]*fm_v2.FmSeasonOidPo, error) {
	var (
		pos = make([]*fm_v2.FmSeasonOidPo, 0)
	)
	rows, err := d.db.Query(ctx, _fmSeasonOid, fmType, seasonId)
	if err != nil {
		return nil, errors.Wrap(err, "fmSeasonOidByDB d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm_v2.FmSeasonOidPo)
		if err := rows.Scan(&po.Id, &po.FmType, &po.FmId, &po.Oid, &po.Seq, &po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "fmSeasonOidByDB scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "fmSeasonOidByDB rows.Err error")
	}
	return pos, nil
}

func (d *Dao) videoSeasonOidByDB(ctx context.Context, seasonId int64) ([]*fm_v2.VideoSeasonOidPo, error) {
	var (
		pos = make([]*fm_v2.VideoSeasonOidPo, 0)
	)
	rows, err := d.db.Query(ctx, _videoSeasonOid, seasonId)
	if err != nil {
		return nil, errors.Wrap(err, "videoSeasonOidByDB d.db.Query error")
	}
	defer rows.Close()
	for rows.Next() {
		po := new(fm_v2.VideoSeasonOidPo)
		if err := rows.Scan(&po.Id, &po.SeasonId, &po.Oid, &po.Seq, &po.Ctime, &po.Mtime); err != nil {
			return nil, errors.Wrap(err, "videoSeasonOidByDB scan error")
		}
		pos = append(pos, po)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "videoSeasonOidByDB rows.Err error")
	}
	return pos, nil
}

// GetSeasonOidCount 查询合集中稿件数量
func (d *Dao) GetSeasonOidCount(ctx context.Context, req fm_v2.SeasonInfoReq) (int, error) {
	conn := d.redisCli.Conn(ctx)
	defer conn.Close()
	// 查询合集缓存
	count, err := d.seasonOidCountCache(conn, req)
	if err == nil {
		return count, nil
	}
	if err != ecode.NothingFound {
		return 0, err
	}
	// miss则加载缓存
	err = d.loadSeasonOids(ctx, conn, req.Scene, req.FmType, req.SeasonId)
	if err != nil {
		return 0, err
	}
	// 重新查缓存
	return d.seasonOidCountCache(conn, req)
}

func (d *Dao) seasonOidCountCache(conn redis.Conn, req fm_v2.SeasonInfoReq) (int, error) {
	count, err := redis.Int(conn.Do("ZCARD", SeasonOidKey(req.Scene, req.FmType, req.SeasonId)))
	if err != nil {
		return 0, errors.Wrapf(err, "ZCARD err, key:%s", SeasonOidKey(req.Scene, req.FmType, req.SeasonId))
	}
	if count == _notExist {
		return 0, ecode.NothingFound
	}
	return count, nil
}

func seasonOidRange(withCurrent bool, rank int, ps int, fromHead bool) (min, max int) {
	if fromHead {
		min = 0
		max = ps - 1
	} else if withCurrent {
		min = rank
		max = rank + ps - 1
	} else {
		min = rank + 1
		max = rank + ps
	}
	return
}

package steins

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/stat/prom"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_rawRankListSQL     = "SELECT mid,score,mtime FROM rank_list_%02d WHERE aid=? and cid=? ORDER BY score DESC LIMIT 50"
	_rankScoreSQL       = "SELECT mid,score,mtime FROM rank_list_%02d WHERE mid=? AND aid=? AND cid=?"
	_rankScoreInsertSQL = "INSERT IGNORE INTO rank_list_%02d (mid,aid,cid,score) VALUES (?,?,?,?)"
	_rankScoreUpdateSQL = "UPDATE rank_list_%02d SET score=? WHERE mid=? AND aid=? AND cid=? AND score<?"
)

// RawRankList is
func (d *Dao) RawRankList(ctx context.Context, req *api.RankListReq) ([]*model.RankItem, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_rawRankListSQL, req.Aid%20), req.Aid, req.Cid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.RankItem{}
	for rows.Next() {
		item := &model.RankItem{}
		if err := rows.Scan(&item.Mid, &item.Score, &item.MTime); err != nil {
			log.Error("Failed to scan rank item: %+v", err)
			continue
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].MTime < out[j].MTime
		}
		return out[i].Score > out[j].Score
	})
	return out, nil
}

// RawRankScoreUpdate is
func (d *Dao) RawRankScoreUpdate(ctx context.Context, req *api.RankScoreSubmitReq) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_rankScoreInsertSQL, req.Aid%20), req.CurrentMid, req.Aid, req.Cid, req.Score)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected > 0 {
		return affected, nil
	}
	res, err = d.db.Exec(ctx, fmt.Sprintf(_rankScoreUpdateSQL, req.Aid%20), req.Score, req.CurrentMid, req.Aid, req.Cid, req.Score)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) RawGetScore(ctx context.Context, mid, aid, cid int64) (*model.RankItem, error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_rankScoreSQL, aid%20), mid, aid, cid)
	out := &model.RankItem{}
	if err := row.Scan(&out.Mid, &out.Score, &out.MTime); err != nil {
		return nil, err
	}
	return out, nil
}

// CacheScore is
func (d *Dao) CacheScore(ctx context.Context, mid, aid, cid int64) (*model.RankItem, error) {
	key := keyRankScore(mid, aid, cid)
	conn := d.rds.Get(ctx)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	out := &model.RankItem{}
	if err := json.Unmarshal(bs, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RankScoreUpdate is
func (d *Dao) RankScoreUpdate(ctx context.Context, req *api.RankScoreSubmitReq) error {
	affected, err := d.RawRankScoreUpdate(ctx, req)
	if err != nil {
		return err
	}
	if affected <= 0 {
		return nil
	}
	if err := d.AddCacheScoreUpdate(ctx, req); err != nil {
		return err
	}
	if err := d.DelCacheScore(ctx, req.CurrentMid, req.Aid, req.Cid); err != nil {
		log.Error("Failed to delete score cache: mid: %d aid: %d cid: %d: %+v", req.CurrentMid, req.Aid, req.Cid, err)
	}
	if err := d.reduceScoreList(ctx, req.Aid, req.Cid, 50); err != nil {
		log.Error("Failed to reduce score list: aid: %d cid: %d: %+v", req.Aid, req.Cid, err)
	}
	return nil
}

// 前 4 位为分数，后 4 位为时间戳
func encodeScore(score int32, at time.Time) uint64 {
	return uint64(score)<<32 | uint64(uint32(at.Unix()&0xFFFFFFFF))
}

func decodeScore(v uint64) (int32, time.Time) {
	at := time.Unix(int64(v)&0xFFFFFFFF, 0)
	score := int32(v >> 32)
	return score, at
}

func keyRankScoreList(aid, cid int64) string {
	return fmt.Sprintf("rank_score_list_%d_%d", aid, cid)
}

func keyRankScore(mid, aid, cid int64) string {
	return fmt.Sprintf("rank_score_%d_%d_%d", mid, aid, cid)
}

// AddCacheScoreUpdate is
func (d *Dao) AddCacheScoreUpdate(ctx context.Context, req *api.RankScoreSubmitReq) error {
	key := keyRankScoreList(req.Aid, req.Cid)
	encodedScore := encodeScore(req.Score, time.Now())
	conn := d.rds.Get(ctx)
	defer conn.Close()
	exist, err := redis.Bool(conn.Do("EXPIRE", key, 12*3600))
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	if _, err := conn.Do("ZADD", key, encodedScore, strconv.FormatInt(req.CurrentMid, 10)); err != nil {
		return err
	}
	return nil
}

func (d *Dao) reduceScoreList(ctx context.Context, aid, cid int64, toSize int64) error {
	key := keyRankScoreList(aid, cid)
	conn := d.rds.Get(ctx)
	defer conn.Close()
	count, err := redis.Int64(conn.Do("ZCOUNT", key, "-inf", "+inf"))
	if err != nil {
		return err
	}
	if count < toSize*2 {
		return nil
	}
	if _, err := conn.Do("ZREMRANGEBYRANK", key, 0, toSize-1); err != nil {
		return err
	}
	return nil
}

// CacheRankList is
func (d *Dao) CacheRankList(ctx context.Context, req *api.RankListReq) ([]*model.RankItem, error) {
	key := keyRankScoreList(req.Aid, req.Cid)
	conn := d.rds.Get(ctx)
	defer conn.Close()
	exist, err := redis.Bool(conn.Do("EXPIRE", key, 12*3600))
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	values, err := redis.Int64Map(conn.Do("ZREVRANGE", key, 0, req.Size_-1, "WITHSCORES"))
	if err != nil {
		return nil, err
	}
	out := []*model.RankItem{}
	for midStr, v := range values {
		mid, err := strconv.ParseInt(midStr, 10, 64)
		if err != nil {
			log.Error("Failed to parse mid: %q: %+v", midStr, errors.WithStack(err))
			continue
		}
		score, atTime := decodeScore(uint64(v))
		out = append(out, &model.RankItem{
			Mid:   mid,
			Score: score,
			MTime: xtime.Time(atTime.Unix()),
		})
	}
	return out, nil
}

// AddCacheRankList is
func (d *Dao) AddCacheRankList(ctx context.Context, req *api.RankListReq, values []*model.RankItem) error {
	key := keyRankScoreList(req.Aid, req.Cid)
	conn := d.rds.Get(ctx)
	defer conn.Close()
	for _, v := range values {
		encodedScore := encodeScore(v.Score, v.MTime.Time())
		if err := conn.Send("ZADD", key, encodedScore, strconv.FormatInt(v.Mid, 10)); err != nil {
			log.Error("Failed to send zadd: %q: %+v", key, err)
			continue
		}
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	return nil
}

// AddCacheScore is
func (d *Dao) AddCacheScore(ctx context.Context, mid, aid, cid int64, value *model.RankItem) error {
	key := keyRankScore(mid, aid, cid)
	conn := d.rds.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := conn.Do("SET", key, bs); err != nil {
		return err
	}
	return nil
}

func (d *Dao) DelCacheScore(ctx context.Context, mid, aid, cid int64) error {
	key := keyRankScore(mid, aid, cid)
	conn := d.rds.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		return err
	}
	return nil
}

// RankList is
func (d *Dao) RankList(ctx context.Context, req *api.RankListReq) ([]*model.RankItem, error) {
	addCache := true
	res, err := d.CacheRankList(ctx, req)
	if err != nil {
		addCache = false
		//nolint:ineffassign
		err = nil
	}
	if res != nil {
		prom.CacheHit.Incr("RankList")
		return res, nil
	}
	prom.CacheMiss.Incr("RankList")
	res, err = d.RawRankList(ctx, req)
	if err != nil {
		//nolint:errcheck
		errors.Wrapf(err, "RankList %+v", req)
		return res, nil
	}
	miss := res
	if !addCache {
		return res, nil
	}
	d.cache.Do(ctx, func(ctx context.Context) {
		//nolint:errcheck
		d.AddCacheRankList(ctx, req, miss)
	})
	return res, nil
}

// GetScore is
func (d *Dao) GetScore(ctx context.Context, mid, aid, cid int64) (*model.RankItem, error) {
	addCache := true
	res, err := d.CacheScore(ctx, mid, aid, cid)
	if err != nil {
		addCache = false
		//nolint:ineffassign
		err = nil
	}
	if res != nil {
		prom.CacheHit.Incr("GetScore")
		return res, nil
	}
	prom.CacheMiss.Incr("GetScore")
	res, err = d.RawGetScore(ctx, mid, aid, cid)
	if err == sql.ErrNoRows {
		res = &model.RankItem{
			Mid: mid,
		}
		err = nil
	}
	if err != nil {
		//nolint:govet
		return nil, errors.Wrapf(err, "GetScore %+v", mid, aid, cid)
	}
	miss := res
	if !addCache {
		return res, nil
	}
	d.cache.Do(ctx, func(ctx context.Context) {
		//nolint:errcheck
		d.AddCacheScore(ctx, mid, aid, cid, miss)
	})
	return res, nil

}

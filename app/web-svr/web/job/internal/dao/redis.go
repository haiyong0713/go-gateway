package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/web/job/internal/model"
)

func NewRedis() (r *redis.Redis, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func webTopKey() string {
	return "cache_new_web_top"
}

func (d *dao) AddCacheWebTop(ctx context.Context, aids []int64) error {
	return d.addCache(ctx, webTopKey(), xstr.JoinInts(aids))
}

func rankIndexKey(day int64) string {
	return fmt.Sprintf("cache_rank_index_%d", day)
}

func (d *dao) AddCacheRankIndex(ctx context.Context, day int64, aids []int64) error {
	return d.addCache(ctx, rankIndexKey(day), xstr.JoinInts(aids))
}

func rankRcmdKey(rid int64) string {
	return fmt.Sprintf("cache_rank_rcmd_%d", rid)
}

func (d *dao) AddCacheRankRecommend(ctx context.Context, rid int64, aids []int64) error {
	return d.addCache(ctx, rankRcmdKey(rid), xstr.JoinInts(aids))
}

func lpRankRcmdKey(business string) string {
	return fmt.Sprintf("cache_lp_rank_rcmd_%s", business)
}

func (d *dao) AddCacheLpRankRecommend(ctx context.Context, business string, aids []int64) error {
	return d.addCache(ctx, lpRankRcmdKey(business), xstr.JoinInts(aids))
}

func rankRegionKey(rid, day, original int64) string {
	return fmt.Sprintf("cache_rank_region_%d_%d_%d", rid, day, original)
}

func (d *dao) AddCacheRankRegion(ctx context.Context, rid, day, original int64, list []*model.RankAid) error {
	bs, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return d.addCache(ctx, rankRegionKey(rid, day, original), string(bs))
}

func rankTagKey(rid, tagID int64) string {
	return fmt.Sprintf("cache_rank_tag_%d_%d", rid, tagID)
}

func (d *dao) AddCacheRankTag(ctx context.Context, rid, tagID int64, list []*model.RankAid) error {
	bs, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return d.addCache(ctx, rankTagKey(rid, tagID), string(bs))
}

func onlineAidsKey() string {
	return "cache_online_aids"
}

func (d *dao) AddCacheOnlineAids(ctx context.Context, list []*model.OnlineAid) error {
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return d.addCache(ctx, onlineAidsKey(), string(data))
}

func rankListKey(typ model.RankListType, rid int64) string {
	return fmt.Sprintf("cache_rank_list_%d_%d", typ, rid)
}

func (d *dao) AddCacheRankList(ctx context.Context, typ model.RankListType, rid int64, data *model.RankList) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return d.addCache(ctx, rankListKey(typ, rid), string(bs))
}

func (d *dao) addCache(ctx context.Context, key, data string) error {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if _, err := conn.Do("SET", key, data); err != nil {
		return err
	}
	return nil
}

func newListKey(rid, typ int64) string {
	return fmt.Sprintf("newlist_%d_%d", rid, typ)
}

func (d *dao) AddCacheNewList(ctx context.Context, rid, typ int64, list []*model.BvArc, total int) error {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := newListKey(rid, typ)
	count := 0
	if err := conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return err
	}
	count++
	for _, arc := range list {
		bs, _ := json.Marshal(arc)
		if err := conn.Send("ZADD", key, combine(arc.PubDate, total), bs); err != nil {
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			return err
		}
	}
	return nil
}

func combine(pubdate time.Time, count int) int64 {
	return pubdate.Time().Unix()<<24 | int64(count)
}

func regionListKey() string {
	return "region_list"
}

func (d *dao) AddCacheRegionList(ctx context.Context, data map[string][]*model.Region) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return d.addCache(ctx, regionListKey(), string(bs))
}

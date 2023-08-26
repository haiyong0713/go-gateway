package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive-shjd/job/model"
	"go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) statPBKey(aid int64) (key string) {
	return _prefixStatPB + strconv.FormatInt(aid, 10)
}

// toFieldValue 根据Type类型获取到对应的值
func toFieldValue(msg *model.StatMsg) (int32, error) {
	switch msg.Type {
	case model.TypeForCoin:
		return msg.Coin, nil
	case model.TypeForFav:
		return msg.Fav, nil
	case model.TypeForRank:
		return msg.HisRank, nil
	case model.TypeForLike:
		return msg.Like, nil
	case model.TypeForDm:
		return msg.DM, nil
	case model.TypeForReply:
		return msg.Reply, nil
	case model.TypeForShare:
		return msg.Share, nil
	case model.TypeForView:
		return msg.Click, nil
	case model.TypeForFollow:
		return msg.Follow, nil
	default:
		return 0, fmt.Errorf("does not support type %s", msg.Type)
	}
}

// 保存数据到redis中
func (s *Service) saveStatToRedis(c context.Context, stat *api.Stat) (err error) {
	conn := s.statRedis.Get(c)
	defer conn.Close()
	args := []interface{}{s.statPBKey(stat.Aid)}
	statMap := model.ConvertStatToMap(stat)
	for k, v := range statMap {
		args = append(args, k, v)
	}
	if err = conn.Send("HSET", args...); err != nil {
		log.Error("conn.Send(HSET) key(%d) error(%+v)", stat.Aid, err)
		return
	}
	// field不能设置expire time。对整个键设置expire时间
	if err = conn.Send("EXPIRE", s.statPBKey(stat.Aid), s.c.Custom.RedisAvExpireTime); err != nil {
		log.Error("conn.Send(EXPIRE) av(%d) error(%+v)", stat.Aid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%+v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive error(%+v)", err)
			return
		}
	}
	return
}

func (s *Service) getStatFromRedis(ctx context.Context, aid int64) (stat *api.Stat, err error) {
	conn := s.statRedis.Get(ctx)
	defer conn.Close()
	key := s.statPBKey(aid)
	stat = &api.Stat{}
	resMap, err := redis.Int64Map(conn.Do("HGETALL", key))
	if err != nil {
		log.Error("conn.Do('HGETALL', %d) error(%+v)", aid, err)
		return
	}
	if len(resMap) == 0 {
		err = fmt.Errorf("redis doesn't exsit key %s when execute HGETALL", key)
		return
	}
	model.MapToStat(resMap, stat)
	return
}

// 更新archive-service的5个redis集群
func (s *Service) updateArcRedis(ctx context.Context, stat *api.Stat, aid int64) (err error) {
	var item []byte
	if item, err = stat.Marshal(); err != nil {
		log.Error("stat.Marshal error，aid(%d), err(%+v)", aid, err)
		return
	}
	eg := errgroup.WithContext(ctx)
	for _, pool := range s.arcRedises {
		thisPool := pool
		eg.Go(func(ctx context.Context) error {
			arcConn := thisPool.Get(ctx)
			if _, err = arcConn.Do("SET", s.statPBKey(aid), item); err != nil {
				log.Error("update arc-service redis cluster error, aid = %d, err = %+v, pool = %+v", aid, err, thisPool)
			}
			arcConn.Close()
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait() error, with aid = %d", aid)
		return
	}
	return
}

// 从 archive-service 的集群中获取 stat
func (s *Service) getStatFromArcRedis(c context.Context, aid int64) (stat *api.Stat) {
	for k, pool := range s.arcRedises {
		func() {
			conn := pool.Get(c)
			defer conn.Close()
			bs, pErr := redis.Bytes(conn.Do("GET", s.statPBKey(aid)))
			if pErr != nil {
				if pErr == redis.ErrNil {
					return
				}
				log.Error("s.getStatFromArcRedis idx(%d) (%d) error(%+v)", k, aid, pErr)
				return
			}
			stat = &api.Stat{}
			if pErr = stat.Unmarshal(bs); pErr != nil {
				log.Error("s.getStatFromArcRedis Unmarshal(%d) error(%+v)", aid, pErr)
				stat = nil
				return
			}
		}()
		if stat != nil {
			return stat
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

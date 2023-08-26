package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/stat/job/model"

	"go-common/library/sync/errgroup.v2"
)

const (
	_prefixStatPB     = "stpr_"
	_babyKey          = "%d:%s"
	_avLastUpdateTime = "l:%d"
	_BabyGroup        = int64(100) // 每天没有处理的稿件数据分成100个set

)

// 格式化这个aid，格式是stp_100010这样（如果在配置文件中没有定义av号的格式时)
func (s *Service) statPBKey(aid int64) (key string) {
	return _prefixStatPB + strconv.FormatInt(aid, 10)
}

func babyKey(aid int64, now time.Time) string {
	return fmt.Sprintf(_babyKey, aid%_BabyGroup, now.Format("20060102"))
}

func avLastUpdateTimeKey(aid int64) string {
	return fmt.Sprintf(_avLastUpdateTime, aid)
}

// toFieldValue 根据Type类型获取到对应的值
func toFieldValue(msg *model.StatMsg) (int32, error) {
	switch msg.Type {
	case model.TypeForCoin:
		return int32(msg.Coin), nil
	case model.TypeForFav:
		return int32(msg.Fav), nil
	case model.TypeForRank:
		return int32(msg.HisRank), nil
	case model.TypeForLike:
		return int32(msg.Like), nil
	case model.TypeForDm:
		return int32(msg.DM), nil
	case model.TypeForReply:
		return int32(msg.Reply), nil
	case model.TypeForShare:
		return int32(msg.Share), nil
	case model.TypeForView:
		return int32(msg.Click), nil
	case model.TypeForFollow:
		return int32(msg.Follow), nil
	default:
		return 0, fmt.Errorf("does not support type %s", msg.Type)
	}
}

// 保存数据到redis中
func (s *Service) saveStatToRedis(c context.Context, stat *api.Stat) (err error) {
	conn := s.statRedis.Get(c)
	defer conn.Close()
	n := 0
	args := []interface{}{s.statPBKey(stat.Aid)}
	statMap := model.ConvertStatToMap(stat)
	for k, v := range statMap {
		args = append(args, k, v)
	}
	n++
	if err = conn.Send("HSET", args...); err != nil {
		log.Error("conn.Send(HSET) key(%d) error(%v)", stat.Aid, err)
		return
	}
	// field不能设置expire time。对整个键设置expire时间
	if err = conn.Send("EXPIRE", s.statPBKey(stat.Aid), s.c.Custom.RedisAvExpireTime); err != nil {
		log.Error("conn.Send(EXPIRE) av(%d) error(%v)", stat.Aid, err)
		return
	}
	n++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < n; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive error(%v)", err)
			return
		}
	}
	log.Info("save stat to stat-job's redis %v", stat)
	return
}

func (s *Service) flushRedisToDB(ctx context.Context, aid int64) error {
	conn := s.statRedis.Get(ctx)
	defer conn.Close()
	stat, err := s.getStatFromRedis(ctx, aid) // 此处ErrNil也算出错
	if err != nil {
		log.Error("s.getStatFromRedis with aid %d err(%+v)", aid, err)
		return err
	}
	rows, err := s.dao.Update(ctx, stat)
	if err != nil {
		log.Error("s.dao.Update with aid %d err(%+v)", aid, err)
		return err
	}
	if rows > 0 { //更新操作成功
		log.Info("flush Redis Update To DB, aid=%d, stat=%+v", aid, stat)
		return nil
	}
	if err = s.dao.Insert(ctx, stat); err != nil {
		log.Error("s.dao.Insert with aid %d err(%+v)", aid, err)
		return err
	}
	log.Info("flush Redis Insert To DB, aid=%d, stat=%+v", aid, stat)
	return nil
}

func (s *Service) getStatFromRedis(ctx context.Context, aid int64) (stat *api.Stat, err error) {
	conn := s.statRedis.Get(ctx)
	defer conn.Close()
	key := s.statPBKey(aid)
	stat = &api.Stat{}
	resMap, err := redis.Int64Map(conn.Do("HGETALL", key))
	if err != nil {
		log.Error("conn.Do('HGETALL', %d) error(%v)", aid, err)
		return
	}
	if len(resMap) == 0 {
		err = fmt.Errorf("redis doesn't exsit key %s when execute HGETALL", key)
		return
	}
	mapToStat(resMap, stat)
	return
}

func mapToStat(resMap map[string]int64, stat *api.Stat) {
	stat.Aid = resMap["aid"]
	// 播放数
	stat.View = int32(resMap[model.TypeForView])
	// 弹幕数
	stat.Danmaku = int32(resMap[model.TypeForDm])
	// 评论数
	stat.Reply = int32(resMap[model.TypeForReply])
	// 收藏数
	stat.Fav = int32(resMap[model.TypeForFav])
	// 投币数
	stat.Coin = int32(resMap[model.TypeForCoin])
	// 分享数
	stat.Share = int32(resMap[model.TypeForShare])
	// 当前排名
	stat.NowRank = int32(resMap[model.TypeForNowRank])
	// 历史最高排名
	stat.HisRank = int32(resMap[model.TypeForHisRank])
	// 点赞数
	stat.Like = int32(resMap[model.TypeForLike])
	// 追番数
	stat.Follow = int32(resMap[model.TypeForFollow])
}

// 更新archive-service的5个redis集群
func (s *Service) updateArcRedis(ctx context.Context, stat *api.Stat, aid int64) (err error) {
	var item []byte
	if item, err = stat.Marshal(); err != nil {
		log.Error("stat.Marshal error，aid(%d), err(%v)", aid, err)
		return
	}
	eg := errgroup.WithContext(ctx)
	for _, pool := range s.arcRedises {
		thisPool := pool
		eg.Go(func(ctx context.Context) (err error) {
			err = retry.WithAttempts(ctx, "updateArcRedis-retry", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
				arcConn := thisPool.Get(c)
				defer arcConn.Close()
				if _, redisErr := arcConn.Do("SET", s.statPBKey(aid), item); err != nil {
					return redisErr
				}
				return nil
			})
			return err
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait() err(%+v) with aid = %d", err, aid)
	}
	return
}

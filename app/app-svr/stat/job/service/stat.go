package service

import (
	"context"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/stat/job/model"
)

func (s *Service) statDealproc() {
	defer s.waiter.Done()
	var (
		ch  = s.statChan
		ctx = context.TODO()
		err error
	)
	for {
		ms, ok := <-ch
		if !ok {
			log.Warn("ArcStat statDealproc quit")
			return
		}
		if err = s.initRedisAndDB(ctx, ms); err != nil {
			continue
		}
		if err = s.updateRedisAndDB(ctx, ms); err != nil {
			continue
		}
		if err = s.arcRedisSync(ctx, ms.Aid); err != nil {
			log.Error("redis sync error aid = %d", ms.Aid)
			continue
		}
		log.Info("statDealproc handle msg succeeded, ms %+v", ms)
	}
}

func (s *Service) initRedisAndDB(ctx context.Context, ms *model.StatMsg) (err error) {
	var (
		aid    = ms.Aid
		exists bool
		conn   = s.statRedis.Get(ctx)
	)
	defer conn.Close()
	if exists, err = redis.Bool(conn.Do("EXPIRE", s.statPBKey(aid), s.c.Custom.RedisAvExpireTime)); err != nil {
		log.Error("conn.Do(EXPIRE) key(%s) error(%v)", s.statPBKey(aid), err)
		return
	}
	if exists {
		return
	}
	// 此处应该从arc service里面读取数据
	var statReply *api.StatReply
	if statReply, err = s.arcClient.Stat(ctx, &api.StatRequest{Aid: ms.Aid}); err != nil || statReply == nil {
		log.Error("s.arcClient.Stat(%d) error(%v)", ms.Aid, err)
		return
	}
	if ms.Type == model.TypeForView && int32(ms.Click) < statReply.Stat.View {
		log.Error("日志告警 消息中的播放数小于grpc接口中的播放数!!!!! aid(%d), msClick(%d), statView(%d)", ms.Aid, ms.Click, statReply.Stat.View)
	}
	// 初始化缓存的内容
	if err = s.saveStatToRedis(ctx, statReply.Stat); err != nil {
		log.Error("s.setAvhash(%+v) error(%v)", statReply.Stat, err)
		return
	}
	return
}

func (s *Service) updateRedisAndDB(ctx context.Context, ms *model.StatMsg) (err error) {
	var fieldValue int32
	if fieldValue, err = toFieldValue(ms); err != nil {
		log.Error("to field value error, aid = %d, ms = %+v", ms.Aid, ms)
		return
	}
	now := time.Now()
	conn := s.statRedis.Get(ctx)
	defer conn.Close()
	receiveCount := 5 // 初始化就有5个SEND：HSET EXPIRE SADD EXPIRE SETLOCK
	if err = conn.Send("HSET", s.statPBKey(ms.Aid), toFieldValueName(ms), fieldValue); err != nil {
		log.Error("conn.Send(HSET) key(%s) field(%s) error(%v)", s.statPBKey(ms.Aid), ms.Type, err)
		return
	}
	// 设置avKey expire
	// field不能设置expire time。对整个键设置expire时间
	if err = conn.Send("EXPIRE", s.statPBKey(ms.Aid), s.c.Custom.RedisAvExpireTime); err != nil {
		log.Error("conn.Send(EXPIRE) av(%d) error(%v)", ms.Aid, err)
		return
	}
	// 添加之后将这个稿件加入到set中去（表示需要落入到DB中）SADD 10:20191029 ms.Aid为100010
	if err = conn.Send("SADD", babyKey(ms.Aid, now), ms.Aid); err != nil {
		log.Error("conn.Send(SADD) key(%s) value(%d) error(%v)", babyKey(ms.Aid, now), ms.Aid, err)
		return
	}
	// expire 10:20191029 整个set的过期时间设置为2天
	if err = conn.Send("EXPIRE", babyKey(ms.Aid, now), s.c.Custom.BabyExpire); err != nil {
		log.Error("conn.Send(EXPIRE) key(%s) (%d)", babyKey(ms.Aid, now), s.c.Custom.BabyExpire)
		return
	}
	// l:100010 --> 当前时间  如果不存在上次更新的情况，那么就把当次更新的数据存入。如果存在上次更新的情况，就要考虑是否超过了120s了，相当于拿一个分布式锁
	if err = conn.Send("SET", avLastUpdateTimeKey(ms.Aid), now.Unix(), "EX", s.c.Custom.LastChangeTime, "NX"); err != nil {
		log.Error("conn.Send(SETNX) ms.Aid(%d) key(%s) error(%v)", ms.Aid, avLastUpdateTimeKey(ms.Aid), err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v) with ms.Aid(%d)", err, ms.Aid)
		return
	}
	for i := 0; i < receiveCount-1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn receive error(%v) with ms.Aid(%d)", err, ms.Aid)
			return
		}
	}
	reply, err := redis.String(conn.Receive())
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.receive() ms.Aid(%d) error(%v)", ms.Aid, err)
		return
	}
	if reply != "OK" {
		return
	}
	if err = s.flushRedisToDB(ctx, ms.Aid); err != nil {
		log.Error("s.flushRedisToDB:  ms.Aid = %d, err(%v)", ms.Aid, err)
		return
	}
	if _, err = conn.Do("SREM", babyKey(ms.Aid, now), ms.Aid); err != nil {
		log.Error("conn.Send(SREM, %s, %d) error(%v)", babyKey(ms.Aid, now), ms.Aid, err)
		return
	}
	log.Info("update stat-job redis succeeded. ms(%+v)", ms)
	return
}

// 处理Rank到HisRank的兼容
func toFieldValueName(msg *model.StatMsg) string {
	if msg.Type == model.TypeForRank {
		return model.TypeForHisRank
	}
	return msg.Type
}

func (s *Service) arcRedisSync(ctx context.Context, aid int64) (err error) {
	stat, err := s.getStatFromRedis(ctx, aid) // ErrNil也算出错，因为此处stat job redis理应有内容
	if err != nil {
		log.Error("stat get error, aid = %d, err(%v)", aid, err)
		return
	}
	if err = s.updateArcRedis(ctx, stat, aid); err != nil {
		log.Error("s.updateArcRedis: aid = %d, stat = %+v, err = %v", aid, stat, err)
		return
	}
	return
}

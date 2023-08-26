package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/archive-shjd/job/model"
	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_prefixStatPB = "stpr_"
)

// consumerproc consumer all topic
func (s *Service) consumerproc(k string, d *databus.Databus) {
	defer s.waiter.Done()
	var msgs = d.Messages()
	for {
		var (
			err error
			ok  bool
			msg *databus.Message
			now = time.Now().Unix()
		)
		msg, ok = <-msgs
		if !ok || s.close {
			log.Info("databus(%s) consumer exit", k)
			return
		}
		_ = msg.Commit()
		var ms = &model.StatCount{}
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			log.Error("json.Unmarshal(%s) error(%+v)", string(msg.Value), err)
			continue
		}
		if ms.Aid <= 0 || (ms.Type != "archive" && ms.Type != "archive_his") {
			continue
		}
		// nolint:gomnd
		if now-ms.TimeStamp > 60 {
			log.Error("日志告警 topic(%s) message(%s) too early", msg.Topic, msg.Value)
		}
		stat := &model.StatMsg{Aid: ms.Aid, Type: k, Ts: ms.TimeStamp}
		switch k {
		case model.TypeForView:
			stat.Click = ms.Count
		case model.TypeForDm:
			stat.DM = ms.Count
		case model.TypeForReply:
			stat.Reply = ms.Count
		case model.TypeForFav:
			stat.Fav = ms.Count
		case model.TypeForCoin:
			stat.Coin = ms.Count
		case model.TypeForShare:
			stat.Share = ms.Count
		case model.TypeForRank:
			stat.HisRank = ms.Count
		case model.TypeForLike:
			//在实验白名单+灰度内，不处理，交由新消息处理
			_, inWhitelist := s.c.LikeRailgunWhitelist[strconv.FormatInt(ms.Aid, 10)]
			if inWhitelist || ms.Aid%10000 < s.c.LikeRailgunGray {
				continue
			}
			stat.Like = ms.Count
		case model.TypeForFollow:
			stat.Follow = ms.Count
		default:
			log.Error("unknown type(%s) message(%s)", k, msg.Value)
			continue
		}
		s.statChan <- stat
	}
}

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
		if err = s.initRedis(ctx, ms); err != nil {
			log.Error("s.initRedis msg(%+v) error(%+v)", ms, err)
			continue
		}
		if err = s.updateRedis(ctx, ms); err != nil {
			log.Error("s.updateRedis err aid = %d", ms.Aid)
			continue
		}
		if err = s.arcRedisSync(ctx, ms.Aid); err != nil {
			log.Error("redis sync error aid = %d", ms.Aid)
			continue
		}
		log.Info("statDealproc handle msg succeeded, ms %+v", ms)
	}
}

func (s *Service) initRedis(ctx context.Context, ms *model.StatMsg) (err error) {
	var (
		aid    = ms.Aid
		exists bool
		conn   = s.statRedis.Get(ctx)
	)
	defer conn.Close()
	if exists, err = redis.Bool(conn.Do("EXPIRE", s.statPBKey(aid), s.c.Custom.RedisAvExpireTime)); err != nil {
		log.Error("conn.Do(EXPIRE) key(%s) error(%+v)", s.statPBKey(aid), err)
		return
	}
	if !exists {
		var stat *api.Stat
		// 此处应该从arc service里面读取数据
		if stat = s.getStatFromArcRedis(ctx, aid); stat == nil {
			if stat, err = s.dao.Stat(ctx, aid); err != nil {
				if err != sql.ErrNoRows {
					log.Error("s.dao.Stat(%d) error(%+v)", aid, err)
					return err
				}
				err = nil
			}
		}
		// 如果 arc-service 的redis和db中都没有这个aid的统计数据，说明这是一个全新的稿件，则初始化aid到缓存中即可
		if stat == nil {
			stat = &api.Stat{Aid: aid}
		}
		// 初始化缓存的内容
		if err = s.saveStatToRedis(ctx, stat); err != nil {
			log.Error("s.setAvhash(%+v) error(%+v)", stat, err)
			return
		}
	}
	return
}

func (s *Service) updateRedis(ctx context.Context, ms *model.StatMsg) (err error) {
	var fieldValue int32
	if fieldValue, err = toFieldValue(ms); err != nil {
		log.Error("to field value error, aid = %d, ms = %+v", ms.Aid, ms)
		return
	}
	conn := s.statRedis.Get(ctx)
	defer conn.Close()
	if err = conn.Send("HSET", s.statPBKey(ms.Aid), toFieldValueName(ms), fieldValue); err != nil {
		log.Error("conn.Send(HSET) key(%s) field(%s) error(%+v)", s.statPBKey(ms.Aid), ms.Type, err)
		return
	}
	// 设置avKey expire
	// field不能设置expire time。对整个键设置expire时间
	if err = conn.Send("EXPIRE", s.statPBKey(ms.Aid), s.c.Custom.RedisAvExpireTime); err != nil {
		log.Error("conn.Send(EXPIRE) av(%d) error(%+v)", ms.Aid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%+v) with ms.Aid(%d)", err, ms.Aid)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn receive error(%+v) with ms.Aid(%d)", err, ms.Aid)
			return
		}
	}
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
	stat, err := s.getStatFromRedis(ctx, aid) // 如果stat redis里面为空，也会报错
	if err != nil {
		log.Error("stat get error, aid = %d, err(%+v)", aid, err)
		return
	}
	err = s.updateArcRedis(ctx, stat, aid)
	if err != nil {
		log.Error("s.updateArcRedis: aid = %d, stat = %+v, err = %+v", aid, stat, err)
		return
	}
	return
}

package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/archive-shjd/job/conf"
	"go-gateway/app/app-svr/archive-shjd/job/model"
)

const _messageBeforeTimeNotHandle = 60

type Railgun struct {
	cfg  *conf.SingleRailgun
	r    *railgun.Railgun
	tp   string
	name string
}

func (s *Service) initStatRailgun() {
	s.statRgs = []*Railgun{
		{name: "点赞数-云立方", cfg: s.c.LikeYLFRailgun, tp: model.TypeForLikeYLF},
		{name: "点赞数-嘉定", cfg: s.c.LikeJDRailgun, tp: model.TypeForLikeJD},
	}
	for _, r := range s.statRgs {
		r.r = railgun.NewRailGun(r.name, r.cfg.Cfg,
			railgun.NewDatabusV1Inputer(r.cfg.Databus),
			railgun.NewSingleProcessor(r.cfg.Single, s.statUnpackGen(r.tp), s.statRailgunDo),
		)
		r.r.Start()
	}
}

func (s *Service) closeStatRailgun() {
	for _, statRg := range s.statRgs {
		statRg.r.Close()
	}
}

// statUnpackGen 消费
func (s *Service) statUnpackGen(tp string) func(msg railgun.Message) (res *railgun.SingleUnpackMsg, err error) {
	return func(msg railgun.Message) (res *railgun.SingleUnpackMsg, err error) {
		// 处理逻辑 与旧的consumerproc方法 保持一致
		var ms = &model.StatCount{}
		if err = json.Unmarshal(msg.Payload(), ms); err != nil {
			log.Error("statUnpackGen json.Unmarshal(%s) error(%v)", msg.Payload(), err)
			return
		}
		if ms.Aid <= 0 || (ms.Type != "archive" && ms.Type != "archive_his") {
			log.Warn("statUnpackGen message(%s) type is not archive nor archive_his, abort", msg.Payload())
			return
		}
		//在实验白名单+灰度内处理
		_, inWhitelist := s.c.LikeRailgunWhitelist[strconv.FormatInt(ms.Aid, 10)]
		if !(inWhitelist || ms.Aid%10000 < s.c.LikeRailgunGray) {
			return
		}
		now := time.Now().Unix()
		if now-ms.TimeStamp > _messageBeforeTimeNotHandle { // 太老的消息就不处理了，只处理60s以内的消息
			log.Warn("statUnpackGen tp(%s) message(%s) too early", tp, msg.Payload())
		}

		stat := &model.StatMsg{Aid: ms.Aid, Type: tp, Ts: ms.TimeStamp}
		switch tp {
		case model.TypeForLikeYLF, model.TypeForLikeJD:
			stat.Like = ms.Count
		default:
			log.Error("statUnpackGen unknown type(%s) message(%s)", tp, msg.Payload())
			return
		}
		log.Info("statUnpackGen got message(%+v)", stat)
		return &railgun.SingleUnpackMsg{
			Group: stat.Aid,
			Item:  stat,
		}, nil
	}
}

// statRailgunDo 处理
func (s *Service) statRailgunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	// 处理逻辑 与旧的statDealproc方法 保持一致
	stat := item.(*model.StatMsg)

	// 处理点赞多机房消息, 异常情况直接返回
	if stat.Type == model.TypeForLikeYLF || stat.Type == model.TypeForLikeJD {
		info, err := s.thumbupAurora.CalculateUnitFromInt(stat.Aid)
		if err != nil {
			log.Error("statRailgunDo CalculateUnitFromInt(%v) err: %+v", stat.Aid, err)
			return railgun.MsgPolicyFailure
		}
		if (stat.Type == model.TypeForLikeYLF && info.Zone != "sh001") || (stat.Type == model.TypeForLikeJD && info.Zone != "sh004") {
			log.Warn("statRailgunDo ignore like message(%+v), info(%+v)", stat, info)
			return railgun.MsgPolicyIgnore
		}
		//修正为点赞，处理逻辑按原点赞处理
		stat.Type = model.TypeForLike
	}

	var err error
	if err = s.initRedis(ctx, stat); err != nil {
		log.Error("statRailgunDo s.initRedis msg(%+v) error(%+v)", stat, err)
		return railgun.MsgPolicyFailure
	}
	if err = s.updateRedis(ctx, stat); err != nil {
		log.Error("statRailgunDo s.updateRedis err aid = %d error(%+v)", stat.Aid, err)
		return railgun.MsgPolicyFailure
	}
	if err = s.arcRedisSync(ctx, stat.Aid); err != nil {
		log.Error("statRailgunDo redis sync error aid = %d error(%+v)", stat.Aid, err)
		return railgun.MsgPolicyFailure
	}

	log.Info("statRailgunDo handle msg succeeded, ms %+v", stat)
	return railgun.MsgPolicyNormal
}

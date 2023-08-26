package unicom

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"
)

func (s *Service) initComicRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.comicUnpack, s.comicDo)
	g := railgun.NewRailGun("漫画点击日志", nil, inputer, processor)
	s.comicRailGun = g
	g.Start()
}

func (s *Service) comicUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	var v *unicom.ComicView
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || v.Uid == 0 {
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  v,
	}, nil
}

func (s *Service) comicDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	v := item.(*unicom.ComicView)
	if err := s.treatComic(ctx, v); err != nil {
		log.Error("treatComic view:%+v,error:%+v", v, err)
		return railgun.MsgPolicyIgnore
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) treatComic(ctx context.Context, view *unicom.ComicView) error {
	now := time.Now()
	// 过滤未绑定的用户
	ub, err := s.dao.UserBindCache(ctx, view.Uid)
	if err != nil {
		return err
	}
	if ub == nil {
		return nil
	}
	// 过滤非免流卡的用户
	orderm, err := s.orders(ctx, ub.Usermob, now)
	if err != nil {
		return err
	}
	order, err := s.orderState(orderm, unicom.CardProduct, now)
	if err != nil {
		return nil
	}
	// 这边有并发问题，需要加锁
	// 更新 mc db 失败后del锁
	key := comicScoreLockKey(now, ub.Mid)
	locked, err := s.lockdao.TryLock(ctx, key, s.lockExpire)
	if err != nil {
		return err
	}
	if !locked {
		log.Warn("TryLock fail key(%s)", key)
		return nil
	}
	integral := 10
	row, err := s.dao.AddUserBindIntegral(ctx, ub.Mid, strconv.Itoa(ub.Phone), integral)
	if err != nil || row == 0 {
		// row == 0的情况是用户解绑了手机号，此时del锁
		// 尝试删除锁
		if err1 := retry.WithAttempts(ctx, "unlock", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.lockdao.UnLock(ctx, key)
		}); err1 != nil {
			log.Error("%+v", err1)
		}
		return err
	}
	log.Info("观看漫画增加福利点:%+v,order:%+v", ub, order)
	s.addUserIntegralLog(&unicom.UserPackLog{Phone: ub.Phone, Mid: ub.Mid, Integral: 10, UserDesc: "每日礼包"})
	return nil
}

func comicScoreLockKey(now time.Time, mid int64) string {
	// key的前缀是当日在这一个月中的哪一天，按月循环
	// key的超时时间设置为略大于一天，25小时，满足当前按天设限的场景
	return fmt.Sprintf("comic_score_lock_%d_%d", now.Day(), mid)
}

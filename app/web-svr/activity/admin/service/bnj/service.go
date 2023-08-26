package bnj

import (
	"context"
	"fmt"
	"time"

	api "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao"
	"go-gateway/app/web-svr/activity/admin/dao/bnj"
	model "go-gateway/app/web-svr/activity/admin/model/bnj"
	currmdl "go-gateway/app/web-svr/activity/admin/model/currency"

	"go-common/library/sync/errgroup.v2"
)

// Service struct
type Service struct {
	c             *conf.Config
	dao           *bnj.Dao
	like          *dao.Dao
	thumbupClient api.ThumbupClient
	bnjpub        *databus.Databus
}

// New init bnj service.
func New(c *conf.Config) *Service {
	s := &Service{
		c:      c,
		dao:    bnj.New(c),
		like:   dao.New(c),
		bnjpub: databus.New(c.Bnj.Pub),
	}
	var err error
	if s.thumbupClient, err = api.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	return s
}

// PendantCheck checkout user pendant.
func (s *Service) PendantCheck(c context.Context, mid int64) (data *model.PendantCheck, err error) {
	data = new(model.PendantCheck)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		reply, subCheckErr := s.like.SearchReserve(ctx, s.c.Bnj.SidNew, mid, 0, 1)
		if subCheckErr != nil {
			log.Error("PendantCheck like check(%d,%d) error(%v)", mid, s.c.Bnj.Lid, subCheckErr)
			return subCheckErr
		}
		if reply != nil && len(reply.List) > 0 {
			data.SubCheck = true
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		live, e := s.dao.LiveGift(ctx, mid, s.c.Bnj.RoomID, s.c.Bnj.Indexes, s.c.Bnj.Start, s.c.Bnj.End)
		if e != nil {
			log.Error("PendantCheck s.dao.LiveGift(%d,%d) error(%v)", mid, s.c.Bnj.RoomID, e)
			return e
		}
		for _, v := range live.Result {
			if v.Mid == mid {
				data.LiveCheck = true
				break
			}
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, e := s.thumbupClient.HasLike(ctx, &api.HasLikeReq{Business: "archive", MessageIds: s.c.Bnj.Aids, Mid: mid})
		if e != nil {
			log.Error("PendantCheck s.thumbupClient.HasLike(%v,%d) error(%v)", s.c.Bnj.Aids, mid, e)
			return e
		}
		for _, aid := range s.c.Bnj.Aids {
			if stat, ok := reply.States[aid]; ok {
				if stat.State == api.State_STATE_LIKE {
					ts := stat.Time.Time().Unix()
					if ts >= s.c.Bnj.Start.Unix() && ts <= s.c.Bnj.End.Unix() {
						data.LikeCheck = true
						break
					}
				}
			}
		}
		return nil
	})
	err = group.Wait()
	return
}

func (s *Service) AddARSetting(ctx context.Context, setting string) (err error) {
	err = bnj.AddARSetting(ctx, setting)

	return
}

func (s *Service) UpsertScore2CouponRule(ctx context.Context, rule *model.Score2CouponRule) (err error) {
	err = bnj.UpsertARScore2Coupon(ctx, rule)

	return
}

func (s *Service) DelScore2CouponRule(ctx context.Context, rule *model.Score2CouponRule) (err error) {
	err = bnj.DeleteARScore2Coupon(ctx, rule)

	return
}

func (s *Service) ValueChange(c context.Context, value int64) (err error) {
	type Action struct {
		Type int   `json:"type"`
		Num  int64 `json:"num"`
		Ts   int64 `json:"ts"`
	}
	act := &Action{Num: value, Ts: time.Now().Unix()}
	if value > 0 {
		act.Type = 1
	} else {
		act.Num = -value
		act.Type = 2
	}
	err = s.bnjpub.Send(c, fmt.Sprintf("admin"), act)
	if err != nil {
		log.Error("s.bnjpub.Send (%+v) error(%v)", act, err)
	}
	return
}

func (s *Service) Value(c context.Context) (currency *currmdl.CurrencyUser, err error) {
	currency = new(currmdl.CurrencyUser)
	err = s.like.DB.Model(currency).Where("mid = ?", s.c.Bnj.SidNew).First(currency).Error
	if err != nil {
		log.Error("Value sid(%d) error(%v)", s.c.Bnj.SidNew, err)
	}
	return
}

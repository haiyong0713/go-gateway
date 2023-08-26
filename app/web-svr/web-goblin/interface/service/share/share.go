package share

import (
	"context"
	"strconv"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	suit "git.bilibili.co/bapis/bapis-go/account/service/usersuit"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	shamdl "go-gateway/app/web-svr/web-goblin/interface/model/share"
)

// Encourage  share encourage.
func (s *Service) Encourage(c context.Context, mid int64) (res *shamdl.Encourage, err error) {
	var (
		mcShare map[string]int64
		shares  []*shamdl.Share
		key     string
		gps     *suit.GroupPendantMidReply
		shaPend []*shamdl.GroupPendant
		group   = errgroup.Group{}
		accData *accapi.InfoReply
	)
	group.Go(func() (e error) {
		if accData, e = s.accGRPC.Info3(c, &accapi.MidReq{Mid: mid}); e != nil {
			log.Error("Encourage s.accGRPC.Info3(%d) error(%v)", mid, e)
		}
		return
	})
	group.Go(func() (e error) {
		if mcShare, e = s.dao.SharesCache(context.Background(), mid); e != nil {
			log.Error("s.dao.SharesCache mid(%d) error(%v)", mid, e)
			if shares, e = s.dao.Shares(context.Background(), mid); e != nil {
				log.Error("s.dao.Shares mid(%d) error(%v)", mid, e)
				return
			}
			count := len(shares)
			if count > 0 {
				mcShare = make(map[string]int64, count)
				for _, share := range shares {
					key = strconv.FormatInt(share.ShareDate, 10)
					mcShare[key] = share.DayCount
				}
				s.cache.Do(c, func(c context.Context) {
					expire := monthShare()
					_ = s.dao.SetSharesCache(context.Background(), expire, mid, mcShare)
				})
			}
		}
		return
	})
	group.Go(func() (e error) {
		if gps, e = s.suit.GroupPendantMid(context.Background(), &suit.GPMIDReq{MID: mid, GID: s.c.Rule.Gid}); e != nil {
			log.Error("s.suit.GroupPendantMid  mid(%d) error(%v)", mid, e)
		}
		return
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	res = new(shamdl.Encourage)
	if len(gps.GetGroupPendants()) > 0 {
		for _, gp := range gps.GetGroupPendants() {
			shaPend = append(shaPend, &shamdl.GroupPendant{NeedDays: s.Pendants[gp.ID], Pendant: gp})
		}
	}
	if accData == nil || accData.Info == nil || accData.Info.Mid == 0 {
		res.UserInfo = struct{}{}
	} else {
		res.UserInfo = accData.Info
	}
	days := int64(len(mcShare))
	if days > 0 {
		res.TodayShare = mcShare[time.Now().Format("20060102")]
		res.ShareDays = int64(days)
	}
	if len(shaPend) == 0 {
		res.Pendants = struct{}{}
	} else {
		res.Pendants = shaPend
	}
	return
}

func monthShare() int {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 23, 59, 59, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	return int(lastOfMonth.Sub(now).Seconds())
}

func (s *Service) loadPendant() {
	for _, p := range s.c.Pendants {
		s.Pendants[p.Pid] = p.Level
	}
}

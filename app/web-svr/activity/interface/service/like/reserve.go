package like

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/trace"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/tool"
	"go-main/app/account/usersuit/service/api"
	fav "go-main/app/community/favorite/service/api"
	favmdl "go-main/app/community/favorite/service/model"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	actapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/task"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	relApi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	relaapi "git.bilibili.co/bapis/bapis-go/account/service/relation"

	pgcAct "git.bilibili.co/bapis/bapis-go/pgc/service/activity"
)

const (
	_awardSubjectSuccess = "1"
	_awardSubjectFail    = "2"
	_reportBusinessID    = 250
)

func (s *Service) ReserveCancel(c context.Context, sid, mid int64) (err error) {
	var (
		subject *likemdl.SubjectItem
		nowTime = time.Now().Unix()
		repet   *likemdl.HasReserve
	)

	reserveMap4NoCancel := conf.LoadNoCancelReserveMap()
	sIDStr := strconv.FormatInt(sid, 10)
	if d, ok := reserveMap4NoCancel[sIDStr]; ok && d {
		return ecode.ActivityReserveOfNoCancel
	}

	if sid == s.c.StarSpring.ReserveSid {
		err = xecode.RequestErr
		return
	}
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		log.Errorc(c, "LikeAddText:s.dao.ActSubject(%d) error(%+v)", sid, err)
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if int64(subject.Stime) > nowTime {
		err = ecode.ActivityNotStart
		return
	}
	if int64(subject.Etime) < nowTime {
		err = ecode.ActivityOverEnd
		return
	}
	if !subject.CacheReserve() {
		return
	}
	if subject.IsForbidCancel() {
		err = ecode.ActivityReserveCancelForbidden
		return
	}
	if repet, err = s.dao.ReserveOnly(c, sid, mid); err != nil {
		log.Errorc(c, "s.dao.ReserveOnly(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	if repet == nil || repet.ID < 0 || repet.State != 1 {
		return
	}
	// 插入数据库
	item := &likemdl.ActReserve{
		Sid:   sid,
		Mid:   mid,
		State: 0,
		IPv6:  []byte{},
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		item.IPv6 = IPv6
	}
	if err = s.dao.CancelReserve(c, item); err != nil {
		log.Errorc(c, "s.dao.CancelReserve(%d,%d) error(%v)", sid, mid, err)
		return
	}
	log.Infoc(c, "SaveReserve CancelReserve item[%v]", *item)
	if err = s.dao.DelCacheReserveOnly(c, sid, mid); err != nil {
		log.Errorc(c, "DelCacheReserveOnly Err err(%v)", err)
	}
	log.Infoc(c, "SaveReserve DelCacheReserveOnly item[%v]", *item)
	s.cache.SyncDo(c, func(ctx context.Context) {
		num := 0 - repet.Num
		s.dao.IncrCacheReserveTotal(ctx, sid, num)
		s.dao.IncrSubjectStat(ctx, sid, num)
	})
	return
}

// InterReserve for internal only.
func (s *Service) InterReserve(c context.Context, sid, mid int64, num int32) (res int64, err error) {
	// 是否重复预约
	var (
		repet  *likemdl.HasReserve
		incrID int64
		addNum int32
	)
	if repet, err = s.dao.ReserveOnly(c, sid, mid); err != nil {
		log.Errorc(c, "s.dao.ReserveOnly(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	item := &likemdl.ActReserve{
		Num:   num,
		Sid:   sid,
		Mid:   mid,
		State: 1,
		IPv6:  []byte{},
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		item.IPv6 = IPv6
	}
	addNum = num
	if repet != nil && repet.ID > 0 {
		incrID = repet.ID
		if repet.State == 1 {
			if repet.Num >= num {
				return
			}
			addNum = num - repet.Num
		}
		if err = s.dao.UpReserve(c, item); err != nil {
			log.Errorc(c, " s.dao.UpReserve(%v) error(%v)", item, err)
			return
		}
	} else {
		// 并发则直接报错，不好计算直接增加的num
		if incrID, err = s.dao.AddReserve(c, item); err != nil {
			log.Errorc(c, "s.dao.AddReserve(%v) error(%v)", item, err)
			return
		}
	}
	res = incrID
	s.dao.AddCacheReserveOnly(c, sid, &likemdl.HasReserve{ID: incrID, Num: num, State: 1}, mid)
	s.cache.Do(c, func(ctx context.Context) {
		s.dao.IncrCacheReserveTotal(ctx, sid, addNum)
		s.dao.IncrSubjectStat(ctx, sid, addNum)
	})
	return
}

// DelCacheReserveOnly ...
func (s *Service) DelCacheReserveOnly(c context.Context, id, mid int64) (err error) {
	err = s.dao.DelCacheReserveOnly(c, id, mid)
	if err != nil {
		log.Errorc(c, "s.dao.DelCacheReserveOnly(%d,%d)", id, mid)
	}
	return err
}

func (s *Service) AsyncReserve(ctx context.Context, sid, mid int64, num int32, report *likemdl.ReserveReport) (err error) {
	var (
		nowTime = time.Now().Unix()
		repet   *likemdl.HasReserve
		subject *likemdl.SubjectItem
	)
	if subject, err = s.GetActSubjectInfoByOptimization(ctx, sid); err != nil {
		return
	}
	if subject.ID == 0 || subject.State != likemdl.ActSubjectStateNormal {
		err = ecode.ActivityHasOffLine
		return
	}
	if int64(subject.Stime) > nowTime {
		err = ecode.ActivityNotStart
		return
	}
	if int64(subject.Etime) < nowTime {
		err = ecode.ActivityOverEnd
		return
	}

	if repet, err = s.dao.ReserveOnly(ctx, sid, mid); err != nil {
		log.Errorc(ctx, "s.dao.ReserveOnly(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	if repet != nil && repet.ID > 0 && repet.State == 1 {
		err = ecode.ActivityRepeatSubmit
		return
	}
	// star spring start
	if sid == s.c.StarSpring.ReserveSid {
		var stat *relaapi.StatReply
		stat, err = s.relClient.Stat(ctx, &relaapi.MidReq{Mid: mid})
		if err != nil {
			log.Errorc(ctx, "StarSpring s.relClient.Stat mid:%d error(%v)", mid, err)
			return
		}
		if stat.GetFollower() > s.c.StarSpring.FollowerLimit {
			err = ecode.ActivityUpFanLimit
			return
		}
		if err = s.currDao.UpUserAmount(ctx, s.c.StarSpring.CurrID, 0, mid, stat.Follower, ""); err != nil {
			log.Errorc(ctx, "StarSpring UpUserAmount sid:%d mid:%d error(%v)", s.c.StarSpring.CurrID, mid, err)
			return
		}
	}

	// 获取预约总数
	totals, err := s.dao.ReservesTotal(ctx, []int64{sid})
	if err != nil {
		log.Errorc(ctx, "s.dao.ReservesTotal error(%+v) sid(%+v)", err, sid)
		return
	}
	var total int64
	if _, ok := totals[sid]; ok {
		total = totals[sid]
	}

	item := &likemdl.ActReserve{
		Num:   num,
		Sid:   sid,
		Mid:   mid,
		State: 1,
		IPv6:  []byte{},
		Order: total + 1,
	}

	if report.Ip == "" {
		if IPv6 := net.ParseIP(metadata.String(ctx, metadata.RemoteIP)); IPv6 != nil {
			item.IPv6 = IPv6
		}
	} else {
		if IPv6 := net.ParseIP(report.Ip); IPv6 != nil {
			item.IPv6 = IPv6
		}
	}

	asyncReserve := likemdl.AsyncReserve{}
	{
		asyncReserve.OpType = likemdl.AsyncReserveTypeOfInsert
		asyncReserve.Timestamp = time.Now().UnixNano()
		asyncReserve.ActReserve = item
		asyncReserve.Report = likemdl.ReserveReport{
			From:     report.From,
			Typ:      report.Typ,
			Oid:      report.Oid,
			Platform: report.Platform,
			Mobiapp:  report.Mobiapp,
			Buvid:    report.Buvid,
			Spmid:    report.Spmid,
		}
	}
	if bmCtx, ok := ctx.(*bm.Context); ok {
		var carrier interface{} = bmCtx.Request.Header
		if cr, ok := carrier.(trace.Carrier); ok {
			asyncReserve.TraceID = cr.Get(trace.BiliTraceID)
		}
	}

	if repet != nil && repet.ID > 0 {
		{
			asyncReserve.PrimaryKey = repet.ID
			asyncReserve.OpType = likemdl.AsyncReserveTypeOfUpdate
			asyncReserve.Ctime = repet.Ctime
			asyncReserve.Order = repet.Order
		}
	}

	return s.dbOperationOrAsync(ctx, asyncReserve)
}

func (s *Service) asyncReserveConsume() {
	log.Info("start asyncReserveConsume")
	wg := new(sync.WaitGroup)
	wg.Add(s.c.AsyncReserveConfig.Concurrency)

	for i := 0; i < s.c.AsyncReserveConfig.Concurrency; i++ {
		consumer, err := component.DatabusV2ActivityClient.NewConsumer(s.c.AsyncReserveConfig.Topic)

		if err == nil {
			go func() {
				defer func() {
					consumer.Close()
					wg.Done()
				}()

				msgCh := consumer.MessageChan()
				for {
					select {
					case msg, ok := <-msgCh:
						if !ok {
							return
						} else {
							asyncReserve := likemdl.AsyncReserve{}
							ctx := trace.NewContext(context.Background(), trace.New("asyncReserveConsume"))
							if asyncReserve.TraceID != "" {
								header := http.Header{}
								header.Set(trace.BiliTraceID, asyncReserve.TraceID)
								tr, _ := trace.Extract(trace.HTTPFormat, header)
								if tr != nil {
									ctx = trace.NewContext(ctx, tr)
								}
							}
							if err := json.Unmarshal(msg.Payload(), &asyncReserve); err == nil {
								_ = s.SaveReserve(ctx, asyncReserve)
							}

							_ = msg.Ack()
						}
					}
				}
			}()
		}

		wg.Wait()
	}

	// if consumer exit chan is closed, return now, do not loop
	if _, ok := <-consumerExitChan; !ok {
		return
	}

	// avoid frequently restart
	time.Sleep(time.Second * 5)
	s.asyncReserveConsume()
}

// store in db or send into kafka
//  1. if producer is not ready, should store in db
//  2. if producer is ready now, send into kafka
//     2.1 will store in db by consumer
func (s *Service) dbOperationOrAsync(ctx context.Context, asyncReserve likemdl.AsyncReserve) (err error) {
	// 开启异步消费数据
	if s.c.AsyncReserveConfig.Switch == 1 {
		b, _ := json.Marshal(asyncReserve)
		if err = component.AsyncReserveProducer.Send(
			ctx,
			fmt.Sprintf("async_reserve_%v_%v", asyncReserve.Sid, asyncReserve.Mid),
			b); err != nil {
			log.Errorc(ctx, "dbOperationOrAsync sync failed:%v", err)
			return s.SaveReserve(ctx, asyncReserve)
		}
		return
	}

	return s.SaveReserve(ctx, asyncReserve)
}

func (s *Service) SaveReserve(ctx context.Context, asyncReserve likemdl.AsyncReserve) (err error) {
	log.Infoc(ctx, "SaveReserve receive message [%v] reserve[%v]", asyncReserve, *asyncReserve.ActReserve)
	tool.IncrAsyncReserveCount(asyncReserve.Sid, asyncReserve.State)
	tool.SetAsyncReserveDelay(asyncReserve.Sid, asyncReserve.State, asyncReserve.Timestamp)
	item := asyncReserve.ActReserve
	switch asyncReserve.OpType {
	case likemdl.AsyncReserveTypeOfInsert:
		var primaryKey int64

		if primaryKey, err = s.dao.AddReserve(ctx, item); err != nil {
			log.Errorc(ctx, "s.dao.AddReserve(%v) error(%v)", item, err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = ecode.ActivityRepeatSubmit
			}
			// 潜在缓存不一致，尝试清理缓存恢复，有两种case存在风险
			// 1. 超时，不排除数据已经真的更新了
			// 2. 冲突，不排除从其他途径触发了更新
			if err == ecode.ActivityRepeatSubmit || strings.Contains(err.Error(), "context deadline exceeded") {
				s.dao.DelCacheReserveOnly(ctx, item.Sid, item.Mid)
			}
			return
		}

		asyncReserve.PrimaryKey = primaryKey
		asyncReserve.Ctime = xtime.Time(time.Now().Unix())
	case likemdl.AsyncReserveTypeOfUpdate:
		if err = s.dao.UpReserve(ctx, item); err != nil {
			log.Errorc(ctx, " s.dao.UpReserve(%v) error(%v)", item, err)
			return
		}
	}
	log.Infoc(ctx, "SaveReserve save to db finish [%v] reserve[%v]", asyncReserve, *asyncReserve.ActReserve)

	if err = s.dao.AddCacheReserveOnly(
		ctx,
		item.Sid,
		&likemdl.HasReserve{
			ID:    asyncReserve.PrimaryKey,
			Num:   item.Num,
			State: item.State,
			Mtime: xtime.Time(time.Now().Unix()),
			Ctime: asyncReserve.Ctime,
			Order: item.Order,
		},
		item.Mid); err != nil {
		log.Errorc(ctx, "AddCacheReserveOnly Err (%v)", err)
	}
	log.Infoc(ctx, "SaveReserve reset cache finish [%v] reserve[%v]", asyncReserve, *asyncReserve.ActReserve)
	_ = s.dao.IncrCacheReserveTotal(ctx, item.Sid, item.Num)
	_ = s.dao.IncrSubjectStat(ctx, item.Sid, item.Num)
	_ = s.AwardSubject(ctx, item.Sid, item.Mid)

	return
}

// Reserve .
func (s *Service) Reserve(c context.Context, sid, mid int64, num int32) (res int64, err error) {
	// 是否重复预约
	var (
		repet  *likemdl.HasReserve
		incrID int64
	)
	if repet, err = s.dao.ReserveOnly(c, sid, mid); err != nil {
		log.Errorc(c, "s.dao.ReserveOnly(%d,%d) error(%+v)", sid, mid, err)
		return
	}
	if repet != nil && repet.ID > 0 && repet.State == 1 {
		err = ecode.ActivityRepeatSubmit
		return
	}
	// star spring start
	if sid == s.c.StarSpring.ReserveSid {
		var stat *relaapi.StatReply
		stat, err = s.relClient.Stat(c, &relaapi.MidReq{Mid: mid})
		if err != nil {
			log.Errorc(c, "StarSpring s.relClient.Stat mid:%d error(%v)", mid, err)
			return
		}
		if stat.GetFollower() > s.c.StarSpring.FollowerLimit {
			err = ecode.ActivityUpFanLimit
			return
		}
		if err = s.currDao.UpUserAmount(c, s.c.StarSpring.CurrID, 0, mid, stat.Follower, ""); err != nil {
			log.Errorc(c, "StarSpring UpUserAmount sid:%d mid:%d error(%v)", s.c.StarSpring.CurrID, mid, err)
			return
		}
	}
	// star spring ent
	// 插入数据库
	item := &likemdl.ActReserve{
		Num:   num,
		Sid:   sid,
		Mid:   mid,
		State: 1,
		IPv6:  []byte{},
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		item.IPv6 = IPv6
	}
	if repet != nil && repet.ID > 0 {
		incrID = repet.ID
		if err = s.dao.UpReserve(c, item); err != nil {
			log.Errorc(c, " s.dao.UpReserve(%v) error(%v)", item, err)
			return
		}
	} else {
		if incrID, err = s.dao.AddReserve(c, item); err != nil {
			log.Errorc(c, "s.dao.AddReserve(%v) error(%v)", item, err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = ecode.ActivityRepeatSubmit
			}
			return
		}
	}
	res = incrID
	s.cache.Do(c, func(ctx context.Context) {
		s.dao.AddCacheReserveOnly(c, sid, &likemdl.HasReserve{ID: incrID, Num: num, State: 1, Mtime: xtime.Time(time.Now().Unix())}, mid)
		s.dao.IncrCacheReserveTotal(ctx, sid, num)
		s.dao.IncrSubjectStat(ctx, sid, num)
		s.AwardSubject(ctx, sid, mid)
	})
	return
}

// ReserveFollowing .
func (s *Service) ReserveFollowing(c context.Context, sid, mid int64) (res *likemdl.ActFollowingReply, err error) {
	var (
		subject                       *likemdl.SubjectItem
		isFollow                      bool
		total, ticketTotal, reserveID int64
		mtime                         xtime.Time
		ctime                         xtime.Time
		order                         int64
	)
	if subject, err = s.GetActSubjectInfoByOptimization(c, sid); err != nil {
		log.Errorc(c, "LikeAddText:s.dao.ActSubject(%d) error(%+v)", sid, err)
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if !subject.CacheOnly() {
		err = ecode.ActivityNotExist
		return
	}
	eg := errgroup.WithContext(c)
	if subject.CacheReserve() {
		if mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				var repet *likemdl.HasReserve
				if repet, e = s.dao.ReserveOnly(ctx, sid, mid); e != nil {
					log.Errorc(ctx, "s.dao.ReserveOnly(%d,%d) error(%+v)", sid, mid, e)
					return
				}
				if repet != nil && repet.ID > 0 {
					mtime = repet.Mtime
					ctime = repet.Ctime
					if repet.State == 1 {
						isFollow = true
						reserveID = repet.ID
					}
					order = repet.Order
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) error {
			rly, e := s.GetActSubjectsReserveIDsFollowTotalByOptimization(ctx, []int64{sid})
			if e != nil {
				log.Errorc(ctx, "s.dao.ReservesTotal(%d) error(%v)", sid, e)
				return nil
			}
			if _, ok := rly[sid]; ok {
				total = rly[sid]
			}
			return nil
		})
	} else {
		if mid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				var repetNum int
				if repetNum, e = s.dao.TextOnly(ctx, sid, mid); e != nil {
					log.Errorc(ctx, "s.dao.CacheTextOnly(%d,%d) error(%+v)", sid, mid, e)
					return
				}
				if repetNum > 0 {
					isFollow = true
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) (e error) {
			if total, e = s.dao.LikeTotal(c, sid); e != nil {
				log.Errorc(ctx, "s.dao.LikeTotal(%d) error(%v)", sid, e)
				e = nil
			}
			return
		})
	}
	cfg := func() *conf.Bml20 {
		if s.c.Bml20 == nil {
			return nil
		}
		for _, v := range s.c.Bml20 {
			if v != nil && v.Sid == sid {
				return v
			}
		}
		return nil
	}()
	if cfg != nil && cfg.ShopID > 0 {
		eg.Go(func(ctx context.Context) error {
			if cnt, e := s.dao.TicketFavCount(c, cfg.ShopID); e != nil {
				log.Errorc(ctx, "ReserveFollowing TicketFavCount ticketID:%d error(%v)", cfg.ShopID, e)
			} else {
				ticketTotal = cnt
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	if ticketTotal > 0 {
		total = ticketTotal
	}
	res = &likemdl.ActFollowingReply{IsFollowing: isFollow, Total: total, ReserveID: reserveID, Mtime: mtime, Ctime: ctime, Order: order}
	return
}

// ReserveFollowings only for new sid.
func (s *Service) ReserveFollowings(c context.Context, sids []int64, mid int64) (res map[int64]*likemdl.ActFollowingReply, err error) {
	var (
		subjects map[int64]*likemdl.SubjectItem
		newSids  []int64
		totalRly map[int64]int64
	)
	if subjects, err = s.GetActSubjectsInfoByOptimization(c, sids); err != nil {
		log.Errorc(c, "LikeAddText:s.dao.ActSubject(%d) error(%+v)", sids, err)
		return
	}
	for _, v := range sids {
		if _, ok := subjects[v]; !ok {
			continue
		}
		if subjects[v].ID == 0 {
			continue
		}
		if subjects[v].CacheReserve() {
			newSids = append(newSids, v)
		}
	}
	if len(newSids) == 0 {
		return
	}
	mu := sync.Mutex{}
	newReply := make(map[int64]*likemdl.HasReserve)
	eg := errgroup.WithContext(c)
	if mid > 0 {
		for _, vSid := range newSids {
			tmpID := vSid
			eg.Go(func(ctx context.Context) error {
				newR, e := s.dao.ReserveOnly(ctx, tmpID, mid)
				if e != nil {
					log.Errorc(ctx, "s.dao.ReserveOnly(%v,%d) error(%v)", tmpID, mid, e)
					return nil
				}
				mu.Lock()
				newReply[tmpID] = newR
				mu.Unlock()
				return nil
			})
		}
	}
	eg.Go(func(ctx context.Context) (e error) {
		if totalRly, e = s.GetActSubjectsReserveIDsFollowTotalByOptimization(ctx, newSids); e != nil {
			log.Errorc(ctx, "s.dao.ReservesTotal(%v) error(%v)", newSids, e)
			e = nil
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res = make(map[int64]*likemdl.ActFollowingReply)
	for _, v := range newSids {
		tmp := &likemdl.ActFollowingReply{}
		if repet, ok := newReply[v]; ok && repet != nil && repet.ID > 0 {
			tmp.Ctime = repet.Ctime
			tmp.Mtime = repet.Mtime
			if repet.State == 1 {
				tmp.IsFollowing = true
			}
			tmp.Order = repet.Order
		}
		if _, tok := totalRly[v]; tok {
			tmp.Total = totalRly[v]
		}
		res[v] = tmp
	}
	return
}

func (s *Service) AwardSubjectState(c context.Context, sid, mid int64) (state int, err error) {
	var (
		awardData *likemdl.AwardSubject
		hasAward  string
	)
	if awardData, err = s.dao.AwardSubject(c, sid); err != nil {
		log.Errorc(c, "AwardSubjectState:s.dao.ActSubject sid:%d error(%+v)", sid, err)
		return
	}
	nowTs := time.Now().Unix()
	if awardData == nil || awardData.ID == 0 || awardData.State != likemdl.AwardOnline || int64(awardData.Etime) <= nowTs {
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		var awardErr error
		hasAward, awardErr = s.dao.RsGet(ctx, subjectAwardKey(mid, sid))
		if awardErr != nil {
			log.Errorc(ctx, "AwardSubjectState s.dao.RsGet mid:%d sid:%d error(%v)", mid, sid, err)
			return awardErr
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		return s.subjectJoinCheck(ctx, sid, mid)
	})
	if err = group.Wait(); err != nil {
		log.Errorc(c, "AwardSubjectState group.Wait() mid:%d sid:%d error(%v)", mid, sid, err)
		err = nil
		state = likemdl.AwardNotAllow
		return
	}
	if hasAward != "" {
		state = likemdl.AwardReward
	} else {
		state = likemdl.AwardAllowed
	}
	return
}

func (s *Service) AwardSubjectReward(c context.Context, sid, mid int64) (err error) {
	if err = s.subjectJoinCheck(c, sid, mid); err != nil {
		return
	}
	return s.AwardSubject(c, sid, mid)
}

func (s *Service) AwardSubjectStateByID(c context.Context, id, mid int64) (state int, err error) {
	var (
		awardData *likemdl.AwardSubject
		hasAward  string
	)
	if awardData, err = s.dao.AwardSubjectByID(c, id); err != nil {
		log.Errorc(c, "AwardSubjectStateByID:s.dao.AwardSubjectByID id:%d error(%+v)", id, err)
		return
	}
	nowTs := time.Now().Unix()
	if awardData == nil || awardData.ID == 0 || awardData.State != likemdl.AwardOnline || int64(awardData.Etime) <= nowTs {
		return
	}
	if awardData.SidType == likemdl.AwardSidTypeSingle {
		group := errgroup.WithContext(c)
		group.Go(func(ctx context.Context) error {
			var awardErr error
			hasAward, awardErr = s.dao.RsGet(ctx, subjectAwardKey(mid, awardData.Sid))
			if awardErr != nil {
				log.Errorc(ctx, "AwardSubjectState s.dao.RsGet mid:%d sid:%d error(%v)", mid, awardData.Sid, err)
				return awardErr
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			return s.subjectJoinCheck(ctx, awardData.Sid, mid)
		})
		if err = group.Wait(); err != nil {
			log.Errorc(c, "AwardSubjectState group.Wait() mid:%d sid:%d error(%v)", mid, awardData.Sid, err)
			err = nil
			state = likemdl.AwardNotAllow
			return
		}
		if hasAward != "" {
			state = likemdl.AwardReward
		} else {
			state = likemdl.AwardAllowed
		}
		return
	}
	if awardData.SidType == likemdl.AwardSidTypeMulti {
		var taskList []*task.TaskItem
		taskList, err = s.taskList(c, mid, task.BusinessAct, awardData.Sid)
		if err != nil {
			log.Errorc(c, "AwardSubjectState taskList mid:%d sid:%d error(%v)", mid, awardData.Sid, err)
			err = nil
			state = likemdl.AwardNotAllow
			return
		}
		for _, v := range taskList {
			if v != nil && v.ID == awardData.TaskID {
				if v.UserFinish != task.HasFinish {
					state = likemdl.AwardNotAllow
					return
				}
				if v.UserAward == task.NotAward {
					state = likemdl.AwardAllowed
					return
				}
				if v.UserAward == task.HasAward {
					state = likemdl.AwardReward
					return
				}
			}
		}
	}
	return
}

func (s *Service) AwardSubjectRewardByID(c context.Context, id, mid int64) (err error) {
	awardData, err := s.dao.AwardSubjectByID(c, id)
	if err != nil || awardData == nil {
		log.Errorc(c, "AwardSubjectRewardByID s.dao.awardSubject sid(%d) mid(%d) error(%v)", id, mid, err)
		return
	}
	nowTs := time.Now().Unix()
	if awardData.State != likemdl.AwardOnline || int64(awardData.Etime) <= nowTs {
		err = ecode.ActivityOverEnd
		return
	}
	if awardData.SidType == likemdl.AwardSidTypeMulti {
		_, err = s.AwardTask(c, mid, awardData.TaskID)
		return
	}
	if err = s.subjectJoinCheck(c, id, mid); err != nil {
		return
	}
	return s.sendSubjectAward(c, mid, nowTs, awardData)
}

func (s *Service) InnerAwardSubject(c context.Context, sid, mid int64) (err error) {
	err = s.AwardSubject(c, sid, mid)
	s.cache.Do(c, func(ctx context.Context) {
		s.dao.AddCacheTextOnly(ctx, sid, 1, mid)
	})
	return
}

func (s *Service) AwardSubject(c context.Context, sid, mid int64) (err error) {
	awardData, err := s.dao.AwardSubject(c, sid)
	if err != nil {
		log.Errorc(c, "awardSubject s.dao.awardSubject sid(%d) mid(%d) error(%v)", sid, mid, err)
		return
	}
	if awardData == nil {
		return
	}
	nowTs := time.Now().Unix()
	if awardData.State != likemdl.AwardOnline || int64(awardData.Etime) <= nowTs {
		err = ecode.ActivityOverEnd
		return
	}
	if awardData.SidType != likemdl.AwardSidTypeSingle {
		err = ecode.ActivityNoAward
		return
	}
	return s.sendSubjectAward(c, mid, nowTs, awardData)
}

func (s *Service) sendSubjectAward(c context.Context, mid, nowTs int64, awardData *likemdl.AwardSubject) (err error) {
	pids, err := xstr.SplitInts(awardData.SourceId)
	if err != nil {
		log.Errorc(c, "awardSubject xstr.SplitInts sid:%d mid:%d sourceID(%s) error(%v)", awardData.Sid, mid, awardData.SourceId, err)
		return
	}
	var hasAward bool
	hasAward, err = s.dao.RsSetNX(c, subjectAwardKey(mid, awardData.Sid), int32(int64(awardData.Etime)-nowTs))
	if err != nil {
		log.Errorc(c, "awardSubject RsSetNx mid:%d sid:%d error(%v)", mid, awardData.Sid, err)
		return
	}
	if !hasAward {
		err = ecode.ActivityHasAward
		return
	}
	expires := make([]int64, 0, len(pids))
	for range pids {
		expires = append(expires, awardData.SourceExpire)
	}
	_, err = s.suitClient.GrantByPids(c, &api.GrantByPidsReq{Mid: mid, Pids: pids, Expires: expires})
	action := _awardSubjectSuccess
	if err != nil {
		log.Errorc(c, "awardSubject s.suitClient.GrantByMids mid(%d) pid(%v) error(%v)", mid, pids, err)
		action = _awardSubjectFail
		s.cache.Do(c, func(ctx context.Context) {
			s.dao.RsDelNX(ctx, subjectAwardKey(mid, awardData.Sid))
		})
	}
	logErr := s.dao.AddSubAwardLog(c, _reportBusinessID, action, awardData.ID, mid)
	if logErr != nil {
		log.Errorc(c, "awardSubject AddSubAwardLog mid(%d) action(%v) sid(%d) error(%v)", mid, action, awardData.Sid, err)
	}
	return
}

func (s *Service) subjectJoinCheck(c context.Context, sid, mid int64) error {
	subject, err := s.dao.ActSubject(c, sid)
	if err != nil {
		log.Errorc(c, "subjectJoinCheck:s.dao.ActSubject(%d) error(%+v)", sid, err)
		return err
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return err
	}
	if subject.CacheReserve() {
		var reserve *likemdl.HasReserve
		if reserve, err = s.dao.ReserveOnly(c, sid, mid); err != nil {
			log.Errorc(c, "subjectJoinCheck s.dao.ReserveOnly(%d,%d) error(%+v)", sid, mid, err)
			return err
		}
		if reserve == nil || reserve.ID < 0 || reserve.State != 1 {
			err = ecode.ActivityNotJoin
			return err
		}
	} else {
		var subjectState int
		if subjectState, err = s.dao.TextOnly(c, sid, mid); err != nil {
			log.Errorc(c, "subjectJoinCheck s.dao.TextOnly sid:%d mid:%d error(%v)", sid, mid, err)
			return err
		}
		if subjectState <= 0 {
			err = ecode.ActivityNotJoin
			return err
		}
	}
	return nil
}

func subjectAwardKey(mid, sid int64) string {
	return fmt.Sprintf("sub_awd_%d_%d", mid, sid)
}

func (s *Service) GetReserveProgress(ctx context.Context, req *pb.GetReserveProgressReq) (*pb.GetReserveProgressRes, error) {
	if req.Sid <= 0 {
		return nil, ecode.ActivityNotExist
	}
	// 检查活动是否存在
	subject, err := s.GetActSubjectInfoByOptimization(ctx, req.Sid)
	if err != nil {
		log.Errorc(ctx, "GetReserveProgress:s.dao.ActSubject(%d) error(%+v)", req.Sid, err)
		return nil, err
	}
	// 检查活动类型
	if !subject.CacheReserve() {
		return nil, ecode.ActivityNotExist
	}
	if req.Mid == 0 {
		// 检查是否有查询用户维度的
		for _, r := range req.Rules {
			if r.Dimension == pb.GetReserveProgressDimension_User {
				return nil, ecode.ActivityReserveProgressNeedMid
			}
		}
	}

	// 限制请求批量
	if len(req.Rules) > 20 {
		return nil, xecode.RequestErr
	}

	mRules := make(map[int64]*likemdl.SubjectRule)

	if subject.Type == likemdl.CLOCKIN || subject.Type == likemdl.USERACTIONSTAT {
		// 打卡和积分查询规则维度需要传rule_id
		for _, r := range req.Rules {
			if r.RuleId == 0 {
				return nil, ecode.ActivityReserveProgressNeedRuleID
			}
		}

		// 加载规则列表
		rules, err := s.dao.MemorySubjectRulesBySid(ctx, req.Sid)
		if err != nil {
			return nil, err
		}
		for _, r := range rules {
			mRules[r.ID] = r
		}

		for _, r := range req.Rules {
			if _, ok := mRules[r.RuleId]; !ok {
				return nil, ecode.ActivityReserveProgressNeedRuleID
			}
		}
	}

	res := new(pb.GetReserveProgressRes)
	res.Data = make([]*pb.OneReserveProgressRes, 0, len(req.Rules))
	// 根据活动类型查数据
	switch subject.Type {
	case likemdl.RESERVATION:
		{
			// 普通预约
			for _, r := range req.Rules {
				switch r.Dimension {
				case pb.GetReserveProgressDimension_User:
					{
						if reserve, err := s.dao.ReserveOnly(ctx, req.Sid, req.Mid); err != nil {
							log.Errorc(ctx, "s.dao.ReserveOnly(%d,%d) error(%+v)", req.Sid, req.Mid, err)
						} else {
							if reserve != nil && reserve.State == 1 {
								// 判断已预约
								res.Data = append(res.Data, &pb.OneReserveProgressRes{
									Progress: int64(reserve.Num),
									Rule:     r,
								})
							} else {
								res.Data = append(res.Data, &pb.OneReserveProgressRes{
									Progress: 0,
									Rule:     r,
								})
							}
						}
						break
					}
				default:
					{
						rly, err := s.GetActSubjectsReserveIDsFollowTotalByOptimization(ctx, []int64{req.Sid})
						if err != nil {
							log.Errorc(ctx, "s.dao.ReservesTotal(%d) error(%v)", req.Sid, err)
							return nil, err
						}
						if _, ok := rly[req.Sid]; ok {
							res.Data = append(res.Data, &pb.OneReserveProgressRes{
								Progress: rly[req.Sid],
								Rule:     r,
							})
						} else {
							res.Data = append(res.Data, &pb.OneReserveProgressRes{
								Progress: 0,
								Rule:     r,
							})
						}
					}
				}
			}
			break
		}
	case likemdl.CLOCKIN:
		{
			nowTs := time.Now().Unix()
			for _, r := range req.Rules {
				switch r.Dimension {
				case pb.GetReserveProgressDimension_User:
					{
						var data *task.Task
						if data, err = s.taskDao.Task(ctx, mRules[r.RuleId].TaskID); err != nil {
							log.Errorc(ctx, "s.taskDao.Task(%d) error(%v)", mRules[r.RuleId].TaskID, err)
							return nil, err
						}
						userTask, err := s.taskDao.UserTaskState(ctx, map[int64]*task.Task{data.ID: data}, req.Mid, data.BusinessID, data.ForeignID, nowTs)
						if err != nil {
							log.Errorc(ctx, "s.taskDao.UserTaskState(%d) error(%v)", mRules[r.RuleId].TaskID, err)
							return nil, err
						}
						if len(userTask) > 0 {
							if mRules[r.RuleId].Attribute&1 == 1 {
								// 按天统计处理
								res.Data = append(res.Data, &pb.OneReserveProgressRes{
									Progress: userTask[fmt.Sprintf("%d_%d", data.ID, data.Round(nowTs))].RoundCount,
									Rule:     r,
								})
							} else {
								res.Data = append(res.Data, &pb.OneReserveProgressRes{
									Progress: userTask[fmt.Sprintf("%d_%d", data.ID, data.Round(nowTs))].Count,
									Rule:     r,
								})
							}
						} else {
							res.Data = append(res.Data, &pb.OneReserveProgressRes{
								Progress: 0,
								Rule:     r,
							})
						}
						break
					}
				case pb.GetReserveProgressDimension_Rule:
					{
						taskStats, err := s.taskDao.TaskStats(ctx, []int64{mRules[r.RuleId].TaskID}, req.Sid, task.BusinessAct)
						if err != nil {
							log.Errorc(ctx, "s.taskDao.TaskStats(%d) error(%v)", mRules[r.RuleId].TaskID, err)
							return nil, err
						}
						if val, ok := taskStats[mRules[r.RuleId].TaskID]; ok {
							res.Data = append(res.Data, &pb.OneReserveProgressRes{
								Progress: val,
								Rule:     r,
							})
						} else {
							res.Data = append(res.Data, &pb.OneReserveProgressRes{
								Progress: 0,
								Rule:     r,
							})
						}
						break
					}
				}
			}
			break
		}
	case likemdl.USERACTIONSTAT:
		{
			for _, r := range req.Rules {
				var err error
				var counterRes *actapi.GetCounterResResp
				var counterReq *actapi.GetCounterResReq
				switch r.Dimension {
				case pb.GetReserveProgressDimension_User:
					{
						counterReq = &actapi.GetCounterResReq{
							Counter:  mRules[r.RuleId].RuleName,
							Activity: fmt.Sprint(req.Sid),
							Mid:      req.Mid,
						}
						counterRes, err = client.ActPlatClient.GetCounterRes(ctx, counterReq)
						break
					}
				case pb.GetReserveProgressDimension_Rule:
					{
						counterReq = &actapi.GetCounterResReq{
							Counter:  "SUM_" + mRules[r.RuleId].RuleName,
							Activity: fmt.Sprint(req.Sid),
							Mid:      -1, // 本来不需要mid，counter要求必传，等他们优化
						}
						counterRes, err = client.ActPlatClient.GetCounterRes(ctx, counterReq)
						break
					}
				}
				if err != nil {
					log.Errorc(ctx, "s.actPlatClient.GetCounterRes(%v) error(%v)", *counterReq, err)
					return nil, err
				}
				var total int64
				for _, c := range counterRes.CounterList {
					total += c.Val
				}
				res.Data = append(res.Data, &pb.OneReserveProgressRes{
					Progress: total,
					Rule:     r,
				})
			}
			break
		}
	}
	return res, nil
}

func (s *Service) RelationReserveInfo(c context.Context, id, mid int64) (res *likemdl.RelationReserveInfo, err error) {
	ts := time.Now().Unix()
	var (
		actRelationSubject *likemdl.ActRelationInfo
	)
	actRelationSubject, err = s.GetActRelationInfoByOptimization(c, id)
	if err != nil {
		log.Errorc(c, "RelationReserveInfo GetActRelationInfoByOptimization Err %v", err)
		return
	}
	if actRelationSubject.ID <= 0 {
		log.Errorc(c, "ActRelationReserveInfo ActRelationInfo No Exist id:%+v reply:%+v", id, actRelationSubject)
		return nil, ecode.ActivityRelationIDNoExistErr
	}
	if actRelationSubject.ReserveIDs == "" {
		log.Errorc(c, "ActRelationReserveInfo res.ReserveIDs empty id:%+v reply:%+v", id, actRelationSubject)
		return nil, errors.New("no exist reserveIDs")
	}
	reserveIDs := strings.Split(actRelationSubject.ReserveIDs, ",")
	if len(reserveIDs) == 0 {
		log.Errorc(c, "ActRelationReserveInfo res.ReserveIDs split reserveIDs empty id:%+v reply:%+v", id, actRelationSubject)
		return nil, errors.New("split reserveIDs empty")
	}

	storeReserveIDs := make([]int64, 0)
	for _, v := range reserveIDs {
		reserveID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(c, "ActRelationReserveInfo Format reserveIDs Err reserveID:%v err:%v", reserveID, err)
			return nil, errors.New("get info failed")
		}
		storeReserveIDs = append(storeReserveIDs, reserveID)
	}

	var (
		actSubject map[int64]*likemdl.SubjectItem
	)
	// 获取预约sid信息
	actSubject, err = s.dao.ActSubjects(c, storeReserveIDs)
	if err != nil {
		log.Errorc(c, "RelationReserveInfo.dao.ActSubjects Err req:%+v reply:%+v err:%v", storeReserveIDs, actSubject, err)
		return nil, err
	}

	data, err := s.ReserveFollowings(c, storeReserveIDs, mid)
	if err != nil {
		log.Errorc(c, "ActRelationReserveInfo ReserveFollowings Err reserveID:%v mid:%v err:%v", storeReserveIDs, mid, err)
		return nil, errors.New("get info failed")
	}

	if len(data) > 0 {
		res = &likemdl.RelationReserveInfo{
			State: 1, // 默认预约成功
		}
		for _, sid := range storeReserveIDs {
			if v, ok := data[sid]; ok {
				// 从subject中获取活动信息
				item := new(likemdl.RelationReserveInfoItem)
				if sub, ok := actSubject[sid]; ok {
					item.Sid = sid
					item.StartTime = sub.Stime.Time().Unix()
					item.EndTime = sub.Etime.Time().Unix()
					item.Total = v.Total

					item.ActStatus = 0 // 0 活动未开始 1 活动已开始 2 活动已结束
					if ts >= item.StartTime && ts <= item.EndTime {
						item.ActStatus = 1
					}
					if ts > item.EndTime {
						item.ActStatus = 2
					}

					item.State = 1 // 默认已预约
					if !v.IsFollowing {
						item.State = int64(0) // 未预约
						res.State = 0         // 有一个未预约 最外层就是未预约
					}

					res.List = append(res.List, item)
				}
			}

			// 去除第一个数据 来填充到最外层
			outSide := (res.List)[0]
			res.SID = outSide.Sid
			res.StartTime = outSide.StartTime
			res.EndTime = outSide.EndTime
			res.Total = outSide.Total
			res.ActStatus = outSide.ActStatus
		}
	}

	return res, nil
}

func (s *Service) DoRelation(c context.Context, id, mid int64, report *likemdl.ReserveReport) (res int, err error) {
	var (
		actRelationSubject *likemdl.ActRelationInfo
	)
	actRelationSubject, err = s.GetActRelationInfoByOptimization(c, id)
	if err != nil {
		log.Errorc(c, "DoRelation GetActRelationInfoByOptimization Err %v", err)
		return 0, err
	}
	if actRelationSubject.ID <= 0 {
		log.Errorc(c, "Get ActRelationInfo No Exist id:%+v reply:%+v", id, actRelationSubject)
		err = ecode.ActivityRelationIDNoExistErr
		return
	}

	ts := time.Now().Unix()
	// 分化
	if actRelationSubject.ReserveIDs != "" {
		config := new(likemdl.RelationReserveConfig)
		if actRelationSubject.ReserveConfig != "" {
			if err = json.Unmarshal([]byte(actRelationSubject.ReserveConfig), config); err != nil {
				log.Errorc(c, "Do-Relation JSON Unmarshal ReserveConfig Err subject:%+v", actRelationSubject)
				return
			}
		}
		if config.StartTime != 0 || config.EndTime != 0 {
			if ts >= config.StartTime && ts <= config.EndTime {
				if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
					s.DoRelationReserve(ctx, id, mid, actRelationSubject, report)
				}); err != nil {
					log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationReserve err[%v]", err)
				}
			}
		} else {
			if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
				s.DoRelationReserve(ctx, id, mid, actRelationSubject, report)
			}); err != nil {
				log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationReserve err[%v]", err)
			}
		}
	}

	if actRelationSubject.FollowIDs != "" {
		config := new(likemdl.RelationFollowConfig)
		if actRelationSubject.FollowConfig != "" {
			if err = json.Unmarshal([]byte(actRelationSubject.FollowConfig), config); err != nil {
				log.Errorc(c, "Do-Relation JSON Unmarshal FollowConfig Err subject:%+v", actRelationSubject)
				return
			}
		}
		if config.StartTime != 0 || config.EndTime != 0 {
			if ts >= config.StartTime && ts <= config.EndTime {
				if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
					s.DoRelationFollow(ctx, id, mid, actRelationSubject, report)
				}); err != nil {
					log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationFollow err[%v]", err)
				}
			}
		} else {
			if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
				s.DoRelationFollow(ctx, id, mid, actRelationSubject, report)
			}); err != nil {
				log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationFollow err[%v]", err)
			}
		}
	}

	if actRelationSubject.SeasonIDs != "" {
		config := new(likemdl.RelationSeasonConfig)
		if actRelationSubject.SeasonConfig != "" {
			if err = json.Unmarshal([]byte(actRelationSubject.SeasonConfig), config); err != nil {
				log.Errorc(c, "Do-Relation JSON Unmarshal SeasonConfig Err subject:%+v", actRelationSubject)
				return
			}
		}
		if config.StartTime != 0 || config.EndTime != 0 {
			if ts >= config.StartTime && ts <= config.EndTime {
				if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
					s.DoRelationSeason(ctx, id, mid, actRelationSubject)
				}); err != nil {
					log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationSeason err[%v]", err)
				}
			}
		} else {
			if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
				s.DoRelationSeason(ctx, id, mid, actRelationSubject)
			}); err != nil {
				log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationSeason err[%v]", err)
			}
		}
	}
	if actRelationSubject.TopicIDs != "" {
		config := new(likemdl.RelationTopicConfig)
		if actRelationSubject.TopicConfig != "" {
			if err = json.Unmarshal([]byte(actRelationSubject.TopicConfig), config); err != nil {
				log.Errorc(c, "Do-Relation JSON Unmarshal TopicConfig Err subject:%+v", actRelationSubject)
				return
			}
		}
		if config.StartTime != 0 || config.EndTime != 0 {
			if ts >= config.StartTime && ts <= config.EndTime {
				if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
					s.DoRelationTopic(ctx, id, mid, actRelationSubject)
				}); err != nil {
					log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationTopic err[%v]", err)
				}
			}
		} else {
			if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
				s.DoRelationTopic(ctx, id, mid, actRelationSubject)
			}); err != nil {
				log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationTopic err[%v]", err)
			}
		}
	}

	if actRelationSubject.MallIDs != "" {
		config := new(likemdl.RelationTopicConfig)
		if actRelationSubject.MallConfig != "" {
			if err = json.Unmarshal([]byte(actRelationSubject.MallConfig), config); err != nil {
				log.Errorc(c, "Do-Relation JSON Unmarshal MallConfig Err subject:%+v", actRelationSubject)
				return
			}
		}
		if config.StartTime != 0 || config.EndTime != 0 {
			if ts >= config.StartTime && ts <= config.EndTime {
				if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
					s.DoRelationMall(ctx, id, mid, actRelationSubject)
				}); err != nil {
					log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationMall err[%v]", err)
				}
			}
		} else {
			if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
				s.DoRelationMall(ctx, id, mid, actRelationSubject)
			}); err != nil {
				log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationMall err[%v]", err)
			}
		}
	}

	if actRelationSubject.FavoriteInfo != "" {
		config := new(likemdl.RelationFavoriteConfig)
		if actRelationSubject.FavoriteConfig != "" {
			if err = json.Unmarshal([]byte(actRelationSubject.FavoriteConfig), config); err != nil {
				log.Errorc(c, "Do-Relation JSON Unmarshal FavoriteConfig Err subject:%+v", actRelationSubject)
				return
			}
		}
		if config.StartTime != 0 || config.EndTime != 0 {
			if ts >= config.StartTime && ts <= config.EndTime {
				if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
					s.DoRelationFavorite(ctx, mid, actRelationSubject)
				}); err != nil {
					log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationFavorite err[%v]", err)
				}
			}
		} else {
			if err := component.ReserveFanout.SyncDo(c, func(ctx context.Context) {
				s.DoRelationFavorite(ctx, mid, actRelationSubject)
			}); err != nil {
				log.Errorc(c, "Do-Relation component.ReserveFanout.SyncDo.DoRelationFavorite err[%v]", err)
			}
		}
	}

	res = 1
	return res, nil
}

func (s *Service) DoRelationReserve(bg context.Context, id, mid int64, subject *likemdl.ActRelationInfo, report *likemdl.ReserveReport) {
	reserveIDs := strings.Split(subject.ReserveIDs, ",")
	if len(reserveIDs) == 0 {
		log.Errorc(bg, "[Do-Relation Reserve Error] Split ReserveIDs Empty id:%v mid:%v subject:%+v", id, mid, subject)
		return
	}

	storeReserveIDs := make([]int64, 0)
	for _, v := range reserveIDs {
		reserveID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(bg, "[Do-Relation Reserve Error] Format reserveIDs Err id:%v mid:%v subject:%+v", id, mid, subject)
			return
		}
		storeReserveIDs = append(storeReserveIDs, reserveID)
	}

	for _, reserveID := range storeReserveIDs {
		sid := reserveID
		if err := component.ReserveFanout.SyncDo(bg, func(ctx context.Context) {
			s.DoRelationAsyncReserve(ctx, sid, mid, 1, report, subject)
		}); err != nil {
			log.Errorc(bg, "Do-Relation component.ReserveFanout.SyncDo.DoRelationAsyncReserve err[%v]", err)
		}
	}
}

func (s *Service) DoRelationFollow(bg context.Context, id, mid int64, subject *likemdl.ActRelationInfo, report *likemdl.ReserveReport) {
	followIDs := strings.Split(subject.FollowIDs, ",")
	if len(followIDs) == 0 {
		log.Errorc(bg, "[Do-Relation follow Error] Split FollowIDs Empty id:%v mid:%v subject:%+v", id, mid, subject)
		return
	}
	storeFollowIDs := make([]int64, 0)
	for _, v := range followIDs {
		followID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(bg, "[Do-Relation Follow Error] Format followIDs Err id:%v mid:%v subject:%+v", id, mid, subject)
			return
		}
		storeFollowIDs = append(storeFollowIDs, followID)
	}
	req := &relApi.BatchAddFollowingsReq{
		Mid:   mid,
		Fid:   storeFollowIDs,
		Spmid: report.Spmid,
	}
	var (
		err   error
		reply *relApi.BatchAddFollowingsReply
	)
	for i := 0; i < 3; i++ {
		reply, err = s.relClient.BatchAddFollowingAsync(bg, req)
		if err == nil && reply.AllSucceed == true {
			break
		}
	}
	if err != nil || reply.AllSucceed != true {
		log.Errorc(bg, "[Do-Relation Follow GRPC BatchAddFollowingAsync Error] subject:%+v req:%+v reply:%+v err:%v", subject, req, reply, err)
	}
}

func (s *Service) DoRelationSeason(bg context.Context, id, mid int64, subject *likemdl.ActRelationInfo) {
	seasonIDs := strings.Split(subject.SeasonIDs, ",")
	if len(seasonIDs) == 0 {
		log.Errorc(bg, "[Do-Relation Season Error] Split SeasonIDs Empty id:%v mid:%v subject:%+v", id, mid, subject)
		return
	}
	storeSeasonIDs := make([]int32, 0)
	for _, v := range seasonIDs {
		seasonID, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			log.Errorc(bg, "[Do-Relation Season Error] Format SeasonIDs Err id:%v mid:%v subject:%+v err:%v", id, mid, subject, err)
			return
		}
		storeSeasonIDs = append(storeSeasonIDs, int32(seasonID))
	}
	for _, v := range storeSeasonIDs {
		seasonID := v
		if err := component.ReserveFanout.SyncDo(bg, func(ctx context.Context) {
			s.DoRelationSeasonFollow(ctx, seasonID, mid, subject)
		}); err != nil {
			log.Errorc(bg, "Do-Relation component.ReserveFanout.SyncDo.DoRelationSeasonFollow err[%v]", err)
		}
	}
}

func (s *Service) DoRelationTopic(bg context.Context, id, mid int64, subject *likemdl.ActRelationInfo) {
	topicIDs := strings.Split(subject.TopicIDs, ",")
	storeTopicIDs := make([]int64, 0)
	for _, v := range topicIDs {
		topicID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(bg, "[Do-Relation Topic Error] Format TopicIDs Err id:%v mid:%v subject:%+v err:%v", id, mid, subject, err)
			return
		}
		storeTopicIDs = append(storeTopicIDs, topicID)
	}

	var err error
	for i := 0; i < 3; i++ {
		_, err = client.TagClient.AddSub(bg, &tag.AddSubReq{Tids: storeTopicIDs, Mid: mid})
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(bg, "[Do-Relation Topic GRPC Tags Error] subject:%+v err:%+v", subject, err)
	}
}

func (s *Service) DoRelationMall(bg context.Context, id, mid int64, subject *likemdl.ActRelationInfo) {
	mallIDs := strings.Split(subject.MallIDs, ",")
	storeMallIDs := make([]int64, 0)
	for _, v := range mallIDs {
		mallID, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorc(bg, "[Do-Relation Mall Error] Format MallIDs Err id:%v mid:%v subject:%+v err:%v", id, mid, subject, err)
			return
		}
		storeMallIDs = append(storeMallIDs, mallID)
	}
	for _, v := range storeMallIDs {
		mallID := v
		if err := component.ReserveFanout.SyncDo(bg, func(ctx context.Context) {
			if e := s.dao.TicketAddFavInner(ctx, mallID, mid); e != nil {
				log.Error("TicketAddFavInner mid:%d mallID:%d error(%v)", mid, mallID, e)
			}
		}); err != nil {
			log.Errorc(bg, "Do-Relation component.ReserveFanout.SyncDo.DoRelationMall err[%v]", err)
		}
	}
}

func (s *Service) DoRelationFavorite(bg context.Context, mid int64, subject *likemdl.ActRelationInfo) {
	var err error
	favoriteInfoItem := likemdl.RelationFavoriteInfoItem{}
	if err = json.Unmarshal([]byte(subject.FavoriteInfo), &favoriteInfoItem); err != nil {
		log.Errorc(bg, "Do-Relation JSON Unmarshal FavoriteInfo Err subject:%+v err:%v", subject, err)
		return
	}
	req := &fav.MultiAddAllReq{
		Typ: 2,
		Fid: 0,
		Mid: mid,
	}

	IDs := strings.Split(favoriteInfoItem.Content, ",")
	if len(IDs) == 0 {
		log.Errorc(bg, "[Do-Relation Split Content IDs Error] Split Content Err content:%v mid:%v subject:%+v", favoriteInfoItem.Content, mid, subject)
		return
	}
	// 根据id和type组装数据
	for _, oID := range IDs {
		id, err := strconv.ParseInt(oID, 10, 64)
		if err != nil {
			log.Errorc(bg, "[Do-Relation Split ParseInt IDs Error] ParseInt Content Err Oid:%v mid:%v subject:%+v err:%v", oID, mid, subject, err)
			return
		}
		req.Resources = append(req.Resources, &favmdl.Resource{
			Typ: int32(favoriteInfoItem.Type),
			Oid: id,
		})
	}

	for i := 0; i < 3; i++ {
		_, err = s.favDao.FavClient.MultiAddAll(bg, req)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Errorc(bg, "[Do-Relation Favorite GRPC MultiAddAll Error] subject:%+v req:%+v err:%+v", subject, req, err)
	}
}

func (s *Service) DoRelationAsyncReserve(bg context.Context, sid, mid int64, num int32, report *likemdl.ReserveReport, subject *likemdl.ActRelationInfo) {
	var err error
	for i := 0; i < 3; i++ {
		err = s.AsyncReserve(bg, sid, mid, num, report)
		if err == nil || err == ecode.ActivityRepeatSubmit {
			break
		}
	}
	if err != ecode.ActivityRepeatSubmit && err != nil {
		log.Errorc(bg, "[Do-Relation Reserve Error] AsyncReserve reserveID Err sid:%v mid:%v subject:%+v err:%v", sid, mid, subject, err)
	}
}

func (s *Service) DoRelationSeasonFollow(bg context.Context, seasonID int32, mid int64, subject *likemdl.ActRelationInfo) {
	var err error
	req := &pgcAct.AddFollowReq{
		SeasonId: seasonID,
		Mid:      mid,
	}
	for i := 0; i < 3; i++ {
		_, err = s.pgcActClient.AddFollow(bg, req)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(bg, "[Do-Relation Season Error] SeasonFollow seasonID Err ssid:%v mid:%v subject:%+v err:%v", seasonID, mid, subject, err)
	}
}

func (s *Service) GetActRelationInfo(c context.Context, id int64, mid int64, specific string) (*likemdl.ActRelationInfoReply, error) {
	ts := time.Now().Unix()
	reply := new(likemdl.ActRelationInfoReply)
	needModules := strings.Split(specific, ",")
	needModulesSet := make(map[string]interface{})
	var err error
	for _, needModule := range needModules {
		switch needModule {
		case "reserve":
			needModulesSet["reserve"] = true
		case "lottery":
			needModulesSet["lottery"] = true
		case "native":
			needModulesSet["native"] = true
		case "h5":
			needModulesSet["h5"] = true
		case "web":
			needModulesSet["web"] = true
		case "videoSource":
			needModulesSet["videoSource"] = true
		}
	}
	if len(needModulesSet) == 0 {
		err = ecode.ActivityRelationParamsErr
		return nil, err
	}
	var actRelationSubject *likemdl.ActRelationInfo
	actRelationSubject, err = s.GetActRelationInfoByOptimization(c, id)
	if err != nil {
		log.Errorc(c, "GetActRelationInfo GetActRelationInfoByOptimization Err %v", err)
		return nil, err
	}
	if actRelationSubject.ID <= 0 {
		log.Warnc(c, "ActRelationInfo No Exist ID:%v reply:%+v", id, actRelationSubject)
		return nil, ecode.ActivityRelationIDNoExistErr
	}

	reply.Name = actRelationSubject.Name

	_, nativeModule := needModulesSet["native"]
	if actRelationSubject.NativeIDs != "" && nativeModule {
		tmp := strings.Split(actRelationSubject.NativeIDs, ",")
		if len(tmp) > 0 {
			for _, nativeID := range tmp {
				val, err := strconv.ParseInt(nativeID, 10, 64)
				if err != nil {
					continue
				}
				reply.NativeIDs = append(reply.NativeIDs, val)
			}
		}
		if len(reply.NativeIDs) > 0 {
			reply.NativeID = reply.NativeIDs[0]
		}
	}

	_, h5Module := needModulesSet["h5"]
	if actRelationSubject.H5IDs != "" && h5Module {
		tmp := strings.Split(actRelationSubject.H5IDs, ",")
		if len(tmp) > 0 {
			for _, h5ID := range tmp {
				val, err := strconv.ParseInt(h5ID, 10, 64)
				if err != nil {
					continue
				}
				reply.H5IDs = append(reply.H5IDs, val)
			}
		}
	}

	_, webModule := needModulesSet["web"]
	if actRelationSubject.WebIDs != "" && webModule {
		tmp := strings.Split(actRelationSubject.WebIDs, ",")
		if len(tmp) > 0 {
			for _, webID := range tmp {
				val, err := strconv.ParseInt(webID, 10, 64)
				if err != nil {
					continue
				}
				reply.WebIDs = append(reply.WebIDs, val)
			}
		}
	}

	_, lotteryModule := needModulesSet["lottery"]
	if actRelationSubject.LotteryIDs != "" && lotteryModule {
		tmp := strings.Split(actRelationSubject.LotteryIDs, ",")
		if len(tmp) > 0 {
			for _, val := range tmp {
				reply.LotteryIDs = append(reply.LotteryIDs, val)
			}
		}
	}

	_, reserveModule := needModulesSet["reserve"]
	if actRelationSubject.ReserveIDs != "" && reserveModule {
		tmp := strings.Split(actRelationSubject.ReserveIDs, ",")
		if len(tmp) > 0 {
			for _, reserveID := range tmp {
				val, err := strconv.ParseInt(reserveID, 10, 64)
				if err != nil {
					continue
				}
				reply.ReserveIDs = append(reply.ReserveIDs, val)
			}
		}
		if len(reply.ReserveIDs) > 0 {
			reply.ReserveID = reply.ReserveIDs[0]
		}
	}

	_, videoSourceModule := needModulesSet["videoSource"]
	if actRelationSubject.VideoSourceIDs != "" && videoSourceModule {
		tmp := strings.Split(actRelationSubject.VideoSourceIDs, ",")
		if len(tmp) > 0 {
			for _, videoSourceID := range tmp {
				val, err := strconv.ParseInt(videoSourceID, 10, 64)
				if err != nil {
					continue
				}
				reply.VideoSourceIDs = append(reply.VideoSourceIDs, val)
			}
		}
	}

	// 获取预约sid信息
	var actSubjects map[int64]*likemdl.SubjectItem
	actSubjects, err = s.GetActSubjectsInfoByOptimization(c, reply.ReserveIDs)
	if err != nil {
		log.Errorc(c, "GetActRelationInfo GetActSubjectsInfoByOptimization Err %v", err)
		return nil, err
	}
	// 存在预约ID
	if len(reply.ReserveIDs) > 0 && reserveModule {
		// 获取所有预约ID信息
		reserveInfos, err := s.ReserveFollowings(c, reply.ReserveIDs, mid)
		if err != nil {
			return nil, errors.New("get info failed")
		}

		reply.ReserveItem = new(likemdl.ActRelationInfoReserveItem)
		reply.ReserveItems = new(likemdl.ActRelationInfoReserveItems)
		reply.ReserveItems.ReserveList = make([]*likemdl.ActRelationInfoReserveItem, 0)

		// items里面 默认预约状态为1 总体total拿第一个预约的数字
		reply.ReserveItems.State = 1
		if mid == 0 {
			reply.ReserveItems.State = 0
		}
		// 将所有信息按顺序展现
		for _, orderReserveID := range reply.ReserveIDs {
			if reserveFollow, ok := reserveInfos[orderReserveID]; ok {
				if subject, ok := actSubjects[orderReserveID]; ok {
					item := new(likemdl.ActRelationInfoReserveItem)

					item.Sid = orderReserveID
					item.Name = subject.Name
					item.Total = reserveFollow.Total
					item.StartTime = subject.Stime.Time().Unix()
					item.EndTime = subject.Etime.Time().Unix()

					item.State = 1
					if mid == 0 {
						item.State = 0
					}
					if reserveFollow.IsFollowing == false {
						item.State = 0
						// 一旦有一个为非预约 items整体为非预约
						reply.ReserveItems.State = 0
					}

					item.ActStatus = 0 // 0 活动未开始 1 活动已开始 2 活动已结束
					if ts >= item.StartTime && ts <= item.EndTime {
						item.ActStatus = 1
					}
					if ts > item.EndTime {
						item.ActStatus = 2
					}

					// 如果这个id是平摊的reserveID
					if reply.ReserveID == orderReserveID {
						reply.ReserveItem = item
						reply.ReserveItems.Total = reserveFollow.Total
					}

					reply.ReserveItems.ReserveList = append(reply.ReserveItems.ReserveList, item)
				}
			}
		}
	}

	return reply, nil
}

func (s *Service) DoRelationReserveCancel(c context.Context, id, mid int64, report *likemdl.ReserveReport) (err error) {
	var (
		actRelationSubject *likemdl.ActRelationInfo
	)
	actRelationSubject, err = s.GetActRelationInfoByOptimization(c, id)
	if err != nil {
		log.Errorc(c, "DoRelationReserveCancel GetActRelationInfoByOptimization Err %v", err)
		return err
	}
	if actRelationSubject.ID <= 0 {
		log.Errorc(c, "Get ActRelationInfo No Exist id:%+v reply:%+v", id, actRelationSubject)
		err = ecode.ActivityRelationIDNoExistErr
		return
	}

	if actRelationSubject.ReserveIDs != "" {
		reserveIDs := strings.Split(actRelationSubject.ReserveIDs, ",")
		if len(reserveIDs) > 0 {
			for _, v := range reserveIDs {
				var sid int64
				sid, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return
				}
				err = s.ReserveCancel(c, sid, mid)
			}
		}
	}

	return err
}

func (s *Service) GetActReserveTotal(ctx context.Context, sid int64) (reply *pb.GetActReserveTotalReply, err error) {
	reply = new(pb.GetActReserveTotalReply)
	res, err := s.GetActSubjectsReserveIDsFollowTotalByOptimization(ctx, []int64{sid})
	if err != nil {
		return
	}
	if v, ok := res[sid]; ok {
		reply.Total = v
	}
	return
}

func (s *Service) CheckReserveDoveAct(ctx context.Context, req *pb.CheckReserveDoveActReq) (res *pb.CheckReserveDoveActReply, err error) {
	log.Infoc(ctx, "CheckReserveDoveAct params : %v", req)
	res = new(pb.CheckReserveDoveActReply)

	if req == nil || len(req.Relations.List) == 0 {
		return
	}

	var playTimes = int64(1)
	if req.Source > 0 && int(req.Source) < len(conf.Conf.ReserveDoveAct.PlayTimes) {
		playTimes = conf.Conf.ReserveDoveAct.PlayTimes[int(req.Source)]
	}

	res.List = make(map[int64]*pb.ReserveDoveActRelationInfo)
	for sid, relationInfo := range req.Relations.List {
		res.List[sid] = &pb.ReserveDoveActRelationInfo{
			IsValid: false,
		}

		nowTime := time.Now()
		// 鸽子蛋活动在有效展示时间内
		if conf.Conf.ReserveDoveAct.Stime > nowTime.Unix() ||
			conf.Conf.ReserveDoveAct.Etime < nowTime.Unix() {
			continue
		}

		if conf.Conf.ReserveDoveAct.BlackList != nil && len(conf.Conf.ReserveDoveAct.BlackList) > 0 {
			blackListFlag := false
			log.Infoc(ctx, "CheckReserveDoveAct check blackList , len:%v", len(conf.Conf.ReserveDoveAct.BlackList))
			for _, blackUpId := range conf.Conf.ReserveDoveAct.BlackList {
				log.Infoc(ctx, "CheckReserveDoveAct check blackList blackId:%v , Upmid:%v", blackUpId, relationInfo.Upmid)
				if blackUpId == relationInfo.Upmid {
					blackListFlag = true
					break
				}
			}
			if blackListFlag {
				log.Infoc(ctx, "CheckReserveDoveAct check blackList Hit !!! Upmid:%v ", relationInfo.Upmid)
				continue
			}
		}

		log.Infoc(ctx, "CheckReserveDoveAct  %v", s.dao.DynamicArc[relationInfo.Upmid])
		// 1  稿件预约目前在有效期内
		// 2、UP主白名单过滤
		// 3、客态用户首次参与活动，主态用户不限制
		if relationInfo.Stime.Time().Before(nowTime) &&
			relationInfo.Etime.Time().After(nowTime) &&
			s.dao.DynamicArc[relationInfo.Upmid] &&
			(relationInfo.IsFollow == 0 && int64(relationInfo.ReserveRecordCtime) == 0 || req.Mid <= 0 || req.Mid == relationInfo.Upmid) &&
			relationInfo.Type == pb.UpActReserveRelationType_Archive &&
			(relationInfo.State == pb.UpActReserveRelationState_UpReserveRelated ||
				relationInfo.State == pb.UpActReserveRelationState_UpReserveRelatedAudit ||
				relationInfo.State == pb.UpActReserveRelationState_UpReserveRelatedOnline) {
			param := url.Values{"sid": {strconv.FormatInt(relationInfo.Sid, 10)}, "up_mid": {strconv.FormatInt(relationInfo.Upmid, 10)}}
			res.List[sid] = &pb.ReserveDoveActRelationInfo{
				IsValid: true,
				Skin: &pb.ReserveDoveActSkin{
					Svga:      conf.Conf.ReserveDoveAct.Svga,
					LastImg:   conf.Conf.ReserveDoveAct.LastImg,
					PlayTimes: playTimes,
				},
				ActUrl: conf.Conf.ReserveDoveAct.ActUrl + param.Encode(),
			}
			continue
		}
	}
	return
}

func (s *Service) GetReserveUpinfo(ctx context.Context, upMid, sid int64) (*accapi.InfoReply, error) {
	if sid <= 0 {
		return nil, errors.New("empty sid")
	}
	if upMid <= 0 {
		relations, err := s.dao.GetUpActReserveRelationInfoBySid(ctx, []int64{sid})
		if err != nil || relations == nil || relations[sid] == nil {
			return nil, errors.New(fmt.Sprintf("can not find related upmid of  sid[%d]", sid))
		}
		upMid = relations[sid].Mid
	}

	if upMid > 0 {
		return s.accClient.Info3(ctx, &accapi.MidReq{Mid: upMid})
	}

	return nil, errors.New("params error")
}

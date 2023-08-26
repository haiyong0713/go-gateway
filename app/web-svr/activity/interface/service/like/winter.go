package like

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	rty "go-common/library/retry"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/like"

	cheesePayApi "git.bilibili.co/bapis/bapis-go/cheese/service/pay"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

const (
	_upMaxCount = 2
	_maxCount   = 1
	_maxPercent = 100
)

var _emptyCourse = make([]*like.CourseOrder, 0)

func (s *Service) WinterCourse(ctx context.Context, mid int64) (res *like.CourseInfo, err error) {
	var (
		payReply   *cheesePayApi.PeriodAssignReply
		winterInfo *like.WinterStudy
	)
	res = &like.CourseInfo{List: _emptyCourse}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (e error) {
		arg := &cheesePayApi.PeriodAssignReq{
			Mid:       mid,
			StartTime: s.c.WinterStudy.BeginTime,
			EndTime:   s.c.WinterStudy.EndTime,
		}
		if payReply, e = client.CheesePayClient.PeriodAssignPaid(ctx, arg); e != nil {
			log.Errorc(ctx, "WinterCourse s.cheesePayClient.PeriodAssignPaid() arg(%+v) error(%+v)", arg, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if winterInfo, e = s.dao.WinterInfo(ctx, mid); e != nil {
			log.Errorc(ctx, "WinterCourse s.dao.WinterInfo() mid(%d) error(%+v)", mid, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if winterInfo != nil && winterInfo.SeasonID > 0 {
		res.IsJoin = true
		res.BuySeason = winterInfo.SeasonID
	}
	if payReply == nil {
		return
	}
	for _, payInfo := range payReply.OrderInfos {
		isBuy := payInfo.SeasonId == res.BuySeason
		res.List = append(res.List, &like.CourseOrder{
			SeasonID:    payInfo.SeasonId,
			SeasonTitle: payInfo.SeasonTitle,
			EpCount:     payInfo.EpCount,
			Cover:       payInfo.Cover,
			Duration:    payInfo.Duration,
			IsBuy:       isBuy,
			OrderNo:     payInfo.OrderNo,
			RealPrice:   payInfo.RealPrice,
		},
		)
	}
	return
}

func (s *Service) checkActWinter(mid int64) error {
	nowTime := time.Now().Unix()
	for _, whiteMid := range s.c.Rule.ActWhiteList {
		if mid == whiteMid {
			return nil
		}
	}
	if nowTime < s.c.WinterStudy.BeginTime {
		return ecode.ActivityNotStart
	}
	if nowTime > s.c.WinterStudy.EndTime {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) WinterJoin(ctx context.Context, mid int64, params *like.ParamWinterJoin) (err error) {
	var (
		courseInfo *like.CourseInfo
		buySeason  *like.CourseOrder
	)
	if err = s.checkActWinter(mid); err != nil {
		return
	}
	if courseInfo, err = s.WinterCourse(ctx, mid); err != nil {
		log.Errorc(ctx, "WinterJoin s.WinterCourse() mid(%d) courseInfo params(%+v) error(%+v)", mid, params, err)
		return
	}
	if courseInfo == nil {
		err = ecode.ActivityLotteryNetWorkError
		log.Errorc(ctx, "WinterJoin s.WinterCourse() mid(%d) course is nil params(%+v) error(%+v)", mid, params, err)
		return
	}
	if courseInfo.BuySeason > 0 {
		err = ecode.ActivityWinterAlreadyErr
		log.Errorc(ctx, "WinterJoin s.WinterCourse() mid(%d) join already params(%+v) error(%+v)", mid, params, err)
		return
	}
	if len(courseInfo.List) == 0 {
		err = ecode.ActivityWinterNoPayErr
		log.Errorc(ctx, "WinterJoin s.WinterCourse() mid(%d) not have course params(%+v) error(%+v)", mid, params, err)
		return
	}
	for _, payInfo := range courseInfo.List {
		if payInfo.SeasonID == int32(params.SeasonID) {
			buySeason = payInfo
			break
		}
	}
	if buySeason == nil {
		err = ecode.ActivityWinterNoPayErr
		log.Errorc(ctx, "WinterJoin s.WinterCourse() mid(%d) course is wrong params(%+v) error(%+v)", mid, params, err)
		return
	}
	if buySeason.Duration <= 0 {
		err = ecode.ActivityWinterNoPayErr
		log.Errorc(ctx, "WinterJoin s.WinterCourse() mid(%d) course Duration zero params(%+v) error(%+v)", mid, params, err)
		return
	}
	// 上报开始统计
	if err = rty.WithAttempts(ctx, "winterJoin_report_start", _retryTime, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.reportStart(ctx, mid)
	}); err != nil {
		log.Errorc(ctx, "WinterJoin s.reportStart() mid(%d) params(%+v) error(%+v)", mid, params, err)
		err = ecode.ActivityWinterJoinErr
		return
	}
	// 保存数据库
	if err = rty.WithAttempts(ctx, "winterJoin_join_winter", _retryTime, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		_, err = s.dao.JoinWinterStudy(ctx, mid, params.IsNotice, buySeason)
		return err
	}); err != nil {
		log.Errorc(ctx, "WinterJoin s.dao.JoinWinterStudy() mid(%d) params(%+v) error(%+v)", mid, params, err)
		err = ecode.ActivityWinterJoinErr
		return
	}
	if params.IsNotice == 1 {
		if err = s.reserveDo(ctx, mid, params); err != nil {
			log.Errorc(ctx, "WinterJoin s.reserveDo() mid(%d) params(%+v) error(%+v)", mid, params, err)
		}
	}
	return
}

func (s *Service) reportStart(ctx context.Context, mid int64) (err error) {
	_, err = client.ActPlatClient.AddSetMemberInt(ctx, &actPlat.SetMemberIntReq{
		Activity: s.c.WinterStudy.ActPlatActivity,
		Name:     "mid",
		Values:   []*actPlat.SetMemberInt{{Value: mid, ExpireTime: 86400 * 60}},
	})
	return
}

func (s *Service) reserveDo(ctx context.Context, mid int64, params *like.ParamWinterJoin) (err error) {
	report := &like.ReserveReport{
		From:     params.From,
		Typ:      params.Typ,
		Oid:      params.Oid,
		Ip:       params.IP,
		Platform: params.Platform,
		Mobiapp:  params.Mobiapp,
		Buvid:    params.Buvid,
		Spmid:    params.Spmid,
	}
	return retry(func() error {
		return s.AsyncReserve(ctx, s.c.WinterStudy.PushSid, mid, 1, report)
	})
}

func (s *Service) WinterProgress(ctx context.Context, mid int64) (res *like.WinterProgress, err error) {
	var (
		winterInfo *like.WinterStudy
		viewTotal  int64
	)
	res = &like.WinterProgress{}
	if winterInfo, err = s.dao.WinterInfo(ctx, mid); err != nil {
		log.Errorc(ctx, "WinterProgress s.dao.WinterInfo() mid(%d) error(%+v)", mid, err)
		return
	}
	if winterInfo == nil || winterInfo.SeasonID <= 0 {
		log.Infoc(ctx, "WinterProgress s.dao.WinterInfo() mid(%d) not buy course", mid)
		return
	}
	// 活动结束直接读取数据库
	if s.c.WinterStudy.ProgressUseDB == 1 && winterInfo.IsEnd == 1 {
		res = &like.WinterProgress{
			IsJoin:         true,
			RealPrice:      winterInfo.RealPrice,
			SeasonID:       winterInfo.SeasonID,
			TotalProgress:  winterInfo.TotalProgress,
			ClockIn:        winterInfo.ClockIn,
			WatchProgress:  winterInfo.WatchProgress,
			ShareProgress:  winterInfo.ShareProgress,
			UploadProgress: winterInfo.UploadProgress,
			WatchDuration:  winterInfo.WatchDuration,
		}
		return
	}
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) (e error) {
		if res.ClockIn, viewTotal, e = s.historyProgress(ctx, mid, winterInfo); e != nil {
			log.Errorc(ctx, "WinterProgress s.getVideoProgress() mid(%d) error(%+v)", mid, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if res.ShareProgress, e = s.shareProgress(ctx, mid, winterInfo.Ctime.Time().Unix()); e != nil {
			log.Errorc(ctx, "WinterProgress s.shareProgress() mid(%d) error(%+v)", mid, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if res.UploadProgress, e = s.uploadProgress(ctx, mid, winterInfo.Ctime.Time().Unix()); e != nil {
			log.Errorc(ctx, "WinterProgress s.uploadProgress() mid(%d) error(%+v)", mid, e)
		}
		return
	})
	eg.Wait()
	res.IsJoin = true
	res.RealPrice = winterInfo.RealPrice
	res.SeasonID = winterInfo.SeasonID
	if res.ClockIn > _maxCount {
		res.ClockIn = _maxCount
	}
	if viewTotal > 0 && winterInfo.Duration > 0 {
		res.WatchProgress = int64(math.Floor((float64(viewTotal) / (float64(winterInfo.Duration) * 0.9)) * 100))
		res.WatchDuration = viewTotal
	}
	if res.WatchProgress > _maxPercent {
		res.WatchProgress = _maxPercent
	}
	if res.ShareProgress > _maxCount {
		res.ShareProgress = _maxCount
	}
	if res.UploadProgress > _upMaxCount {
		res.UploadProgress = _upMaxCount
	}
	res.TotalProgress = getTotalProgress(res.ClockIn, res.WatchProgress, res.ShareProgress, res.UploadProgress)
	if res.TotalProgress > _maxPercent {
		res.TotalProgress = 100
	}
	return
}

func getTotalProgress(clockIn, watch, share, upload int64) int64 {
	clockInRate := int64(math.Ceil(float64(clockIn) / float64(_maxCount) * float64(100) * 0.1))
	watchRate := int64(math.Floor(float64(watch) * 0.7))
	shareRate := int64(math.Ceil(float64(share) / float64(_maxCount) * float64(100) * 0.1))
	uploadRate := int64(math.Ceil(float64(upload) / float64(_upMaxCount) * float64(100) * 0.1))
	return clockInRate + watchRate + shareRate + uploadRate
}

func (s *Service) historyProgress(ctx context.Context, mid int64, winterInfo *like.WinterStudy) (clockInDays, totalCount int64, err error) {
	var start []byte
	for {
		var (
			historyReply      *actPlat.GetHistoryResp
			tmpDays, tmpTotal int64
		)
		if historyReply, err = client.ActPlatClient.GetHistory(ctx, &actPlat.GetHistoryReq{
			Activity: s.c.WinterStudy.ActPlatActivity,
			Counter:  s.c.WinterStudy.HistoryCounter,
			Mid:      mid,
			Start:    start,
		}); err != nil {
			log.Errorc(ctx, "WinterProgress historyDayMap client.ActPlatClient.GetHistory() mid(%d) error(%+v)", mid, err)
			return
		}
		if historyReply == nil {
			log.Warnc(ctx, "WinterProgress historyDayMap client.ActPlatClient.GetHistory() mid(%d) historyReply is nil", mid)
			return
		}
		log.Infoc(ctx, "WinterProgress historyDayMap d.actPlatClient.GetCounterRes(), mid(%d) resp count(%d)", mid, len(historyReply.History))
		if tmpDays, tmpTotal, err = s.getVideoProgress(ctx, historyReply, winterInfo); err != nil {
			log.Errorc(ctx, "WinterProgress historyDayMap client.ActPlatClient.GetHistory mid(%d) error(%+v)", mid, err)
			return
		}
		clockInDays += tmpDays
		totalCount += tmpTotal
		start = historyReply.Next
		if len(start) == 0 {
			break
		}
	}
	return
}

func (s *Service) getVideoProgress(ctx context.Context, historyReply *actPlat.GetHistoryResp, winterInfo *like.WinterStudy) (clockInDays, totalCount int64, err error) {
	historyMap := make(map[string]int64)
	for _, history := range historyReply.History {
		var (
			historySource *like.ProgressHistory
		)
		if history == nil {
			continue
		}
		if history.Source == "" {
			log.Warnc(ctx, "WinterProgress history.Source is empty source(%s)", history.Source)
			continue
		}
		if err = json.Unmarshal([]byte(history.Source), &historySource); err != nil {
			log.Errorc(ctx, "WinterProgress json.Unmarshal source(%s) error(%+v)", history.Source, err)
			return
		}
		if historySource == nil {
			continue
		}
		if historySource.Sid != int64(winterInfo.SeasonID) {
			continue
		}
		totalCount += history.Count
		t := time.Unix(history.Timestamp, 0)
		dateStr := t.Format("20060102")
		if _, ok := historyMap[dateStr]; !ok {
			historyMap[dateStr] = history.Count
			continue
		}
		historyMap[dateStr] += history.Count
	}
	log.Infoc(ctx, "WinterProgress historyMap(%+v) mid(%d)", historyMap, winterInfo.Mid)
	for _, viewCount := range historyMap {
		if viewCount >= s.c.WinterStudy.ClockInCount {
			clockInDays++
		}
	}
	return
}

func (s *Service) shareProgress(ctx context.Context, mid int64, joinTime int64) (shareDays int64, err error) {
	var start []byte
	for {
		var (
			shareReply *actPlat.GetCounterResResp
			tmpDays    int64
		)
		if shareReply, err = client.ActPlatClient.GetCounterRes(ctx, &actPlat.GetCounterResReq{
			Activity: s.c.WinterStudy.ActPlatActivity,
			Counter:  s.c.WinterStudy.ShareCounter,
			Mid:      mid,
			Start:    start,
		}); err != nil {
			log.Errorc(ctx, "WinterProgress  getShareProgress client.ActPlatClient.GetCounterRes() ShareCounter mid(%d) error(%+v)", mid, err)
			return
		}
		if shareReply == nil {
			log.Warnc(ctx, "WinterProgress getShareProgress client.GetCounterRes.GetHistory() mid(%d) shareReply is nil", mid)
			return
		}
		tmpDays = s.getShareDays(ctx, shareReply)
		shareDays += tmpDays
		start = shareReply.Next
		if len(start) == 0 {
			break
		}
	}
	return
}

func (s *Service) getShareDays(ctx context.Context, shareReply *actPlat.GetCounterResResp) (res int64) {
	shareMap := make(map[string]struct{})
	for _, share := range shareReply.CounterList {
		if share == nil {
			continue
		}
		t := time.Unix(share.Time, 0)
		dateStr := t.Format("20060102")
		if _, ok := shareMap[dateStr]; !ok {
			shareMap[dateStr] = struct{}{}
			res++
		}
	}
	return
}

func (s *Service) getUploadDays(upReply *actPlat.GetCounterResResp) (res int64) {
	for _, upload := range upReply.CounterList {
		if upload == nil {
			continue
		}
		res += upload.Val
	}
	return
}

func (s *Service) uploadProgress(ctx context.Context, mid int64, joinTime int64) (uploadDays int64, err error) {
	var start []byte
	for {
		var (
			uploadReply *actPlat.GetCounterResResp
			tmpDays     int64
		)
		if uploadReply, err = client.ActPlatClient.GetCounterRes(ctx, &actPlat.GetCounterResReq{
			Activity: s.c.WinterStudy.ActPlatActivity,
			Counter:  s.c.WinterStudy.UploadCounter,
			Mid:      mid,
			Start:    start,
		}); err != nil {
			log.Errorc(ctx, "WinterProgress  uploadProgress client.ActPlatClient.GetCounterRes() ShareCounter mid(%d) error(%+v)", mid, err)
			return
		}
		if uploadReply == nil {
			log.Warnc(ctx, "WinterProgress uploadProgress client.GetCounterRes.GetHistory() mid(%d) shareReply is nil", mid)
			return
		}
		tmpDays = s.getUploadDays(uploadReply)
		uploadDays += tmpDays
		start = uploadReply.Next
		if len(start) == 0 {
			break
		}
	}
	return
}

func (s *Service) UpWinterProgress(c context.Context) {
	ctx := context.Background()
	mids, err := s.dao.RawWinterMids(ctx)
	if err != nil {
		log.Errorc(ctx, "UpWinterProgress s.dao.RawWinterMids() error(%+v)", err)
		return
	}
	for _, mid := range mids {
		progress, e := s.WinterProgress(ctx, mid)
		if e != nil {
			log.Errorc(ctx, "UpWinterProgress s.WinterProgress() mid(%d) error(%+v)", mid, e)
			continue
		}
		if _, e = s.dao.UpWinterProgress(ctx, mid, progress); e != nil {
			log.Errorc(ctx, "UpWinterProgress s.dao.UpWinterProgress() mid(%d) error(%+v)", mid, e)
			return
		}
		if e = rty.WithAttempts(ctx, "upWinterProgress_del_mid_cache", _retryTime, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.dao.DelCacheWinterStudy(ctx, mid)
		}); e != nil {
			log.Errorc(ctx, "UpWinterProgress s.dao.DelCacheWinterStudy() mid(%d) error(%+v)", mid, e)
		}
	}
}

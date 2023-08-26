package knowledge

import (
	"context"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	ecodex "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	model "go-gateway/app/web-svr/activity/interface/model/knowledge"
	"time"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	livedataapi "git.bilibili.co/bapis/bapis-go/live/data-guru/v1"
)

func (s *Service) IsActiveAid(sid int64) bool {
	for _, v := range s.c.Knowledge.ActiveSid {
		if sid == v {
			return true
		}
	}
	return false
}

// UserInfo 获取用户看板信息
func (s *Service) UserInfo(ctx context.Context, mid, sid int64) (res *model.UserInfoRes, err error) {
	res = &model.UserInfoRes{}
	if s.c.Knowledge.Mid != 0 {
		mid = s.c.Knowledge.Mid
	}
	if !s.IsActiveAid(sid) {
		err = ecode.RequestErr
		return
	}
	r := &model.UserInfo{}
	var toEnd int64
	activity := fmt.Sprintf("%d", sid)
	end, ok := s.c.Knowledge.ActivityEnd[activity]
	if ok {
		now := time.Now().Unix()
		if end-now < 0 {
			toEnd = 0
		} else {
			toEnd = (end - now) / 86400

		}
	}
	res.ActivityEnd = toEnd
	if mid == 0 {
		return
	}
	var t int64
	var infosReply *accountapi.InfoReply
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		r, err = s.dao.UsersMid(ctx, mid, sid)
		if err != nil {
			log.Errorc(ctx, "s.dao.UsersMid err(%v)", err)
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		infosReply, err = client.AccountClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
		if err != nil || infosReply == nil || infosReply.Info == nil {
			log.Errorc(ctx, "client.AccountClient.Info3: error(%v) batch(%d)", err, mid)
			return ecodex.ActivityWriteHandMemberInfoErr
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		t = s.LiveInfo(ctx, mid, sid)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return res, err
	}

	if r != nil {
		res.ArchiveCount = r.ArchiveCount
		res.SingleView = r.SingleView
		res.AllView = r.AllView
		res.Live = t
	}
	res.Account = &model.Account{
		Mid:  infosReply.Info.Mid,
		Face: infosReply.Info.Face,
		Name: infosReply.Info.Name,
	}
	return
}

// LiveInfo 直播信息
func (s *Service) LiveInfo(ctx context.Context, mid int64, sid int64) (r int64) {
	activity := fmt.Sprintf("%d", sid)
	start, ok := s.c.Knowledge.LiveStartTime[activity]
	if !ok {
		return 0
	}
	end, ok := s.c.Knowledge.LiveEndTime[activity]
	if !ok {
		return 0
	}

	//日期当天0点时间戳(拼接字符串)
	t := time.Now()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local) // 创建本月活动
	if today.Unix() < end {
		end = today.Unix()
	}
	if today.Unix() < start {
		start = today.Unix()
	}
	st := time.Unix(start, 0).Format("2006-01-02 15:04:05")
	et := time.Unix(end, 0).Format("2006-01-02 15:04:05")
	entityId := mid
	res, err := client.LiveDataClient.BatchGetFeatureWindowValues(ctx, &livedataapi.BatchGetFeatureWindowValuesReq{
		FeatureId:    200000,
		EntityIds:    []int64{entityId},
		DimId:        0,
		DimValue:     0,
		BatchStart:   st,
		BatchEnd:     et,
		Op:           1,
		DetailedData: false,
	})
	if err != nil {
		log.Errorc(ctx, "client.LiveDataClient.BatchGetFeatureWindowValues err(%v)", err)
		return
	}
	if res != nil {
		if t, ok := res.Values[entityId]; ok {
			return t
		}
	}
	return
}
